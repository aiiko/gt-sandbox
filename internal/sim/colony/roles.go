package colony

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/world"
)

// Oficios preferidos de un miembro de colonia (propuesta, 18.2).
// Un oficio nunca prohíbe otras tareas: da preferencia y habilita suplencias
// cuando falta un especialista, vence una breve espera o escasea la comida.
const (
	RoleGeneralist = ""
	RoleGatherer   = "gatherer"
	RolePorter     = "porter"
	RoleBuilder    = "builder"
	RoleLookout    = "lookout"
	RoleScout      = "scout"
)

// Parámetros de afinidad y suplencias.
const (
	// RoleBonus es la ventaja de un especialista sobre el trabajo de su oficio.
	RoleBonus = 0.08
	// PantryUrgentBonus sube la prioridad de acarrear comida con despensa crítica.
	PantryUrgentBonus = 0.06
	// StockedFiberBonus inclina al porteador hacia las fibras cuando la despensa está llena.
	StockedFiberBonus = 0.04
	// CoverDelay es cuánto espera un trabajo antes de que otro oficio lo cubra
	// si hay un especialista disponible en la colonia.
	CoverDelay clock.Tick = 2 * clock.TPS
)

// AliveMembers cuenta los miembros vivos (y no en despawn) de la colonia col.
func AliveMembers(w *world.World, col ids.ID) int32 {
	r := w.ECS()
	n := int32(0)
	ecs.Get[Member](r).Each(func(e ids.ID, m *Member) {
		if m.Colony == col && !r.Despawning(e) {
			n++
		}
	})
	return n
}

// CriticalFood indica si la despensa de col está por debajo de una porción por
// miembro vivo (con meta de despensa configurada). En ese estado, el
// abastecimiento gana prioridad sobre la obra y las salidas.
func CriticalFood(w *world.World, col ids.ID) bool {
	r := w.ECS()
	c, ok := ecs.Get[Colony](r).Get(col)
	if !ok || c.Stock <= 0 {
		return false
	}
	n := AliveMembers(w, col)
	if n <= 0 {
		return false
	}
	p, _ := ecs.Get[ai.Portions](r).Get(col)
	return p.Left < n
}

// MissingPortions devuelve cuántas porciones faltan en el depósito de col para
// alcanzar su meta (Colony.Stock * miembros vivos).
func MissingPortions(w *world.World, col ids.ID) int32 {
	r := w.ECS()
	c, ok := ecs.Get[Colony](r).Get(col)
	if !ok || c.Stock <= 0 {
		return 0
	}
	goal := c.Stock * AliveMembers(w, col)
	p, _ := ecs.Get[ai.Portions](r).Get(col)
	if p.Left >= goal {
		return 0
	}
	return goal - max(p.Left, 0)
}

// PendingFoodPortions suma las porciones de comida que ya tienen un acarreo
// publicado (abierto o en curso) hacia el nido de col.
func PendingFoodPortions(w *world.World, col ids.ID) int32 {
	r := w.ECS()
	food := ecs.Get[Food](r)
	portions := ecs.Get[ai.Portions](r)
	total := int32(0)
	ecs.Get[jobs.Job](r).Each(func(id ids.ID, j *jobs.Job) {
		if r.Despawning(id) || j.Type != "haul" || !r.Alive(j.Target) || r.Despawning(j.Target) || !food.Has(j.Target) {
			return
		}
		n := int32(1)
		if p, ok := portions.Get(j.Target); ok && p.Left > 0 {
			n = p.Left
		}
		total += n
	})
	return total
}

// HasRoles indica si algún miembro vivo de col tiene un oficio asignado.
// Cuando nadie tiene oficio, el selector conserva el comportamiento previo.
func HasRoles(w *world.World, col ids.ID) bool {
	r := w.ECS()
	found := false
	ecs.Get[Member](r).Each(func(e ids.ID, m *Member) {
		if !found && m.Colony == col && m.Role != RoleGeneralist && !r.Despawning(e) {
			found = true
		}
	})
	return found
}

// specialistAvailable indica si otro miembro vivo de col con el oficio role
// está libre y en condiciones de tomar un trabajo.
func specialistAvailable(w *world.World, col ids.ID, role string, exclude ids.ID) bool {
	if role == RoleGeneralist {
		return false
	}
	r := w.ECS()
	agents := ecs.Get[ai.Agent](r)
	found := false
	ecs.Get[Member](r).Each(func(e ids.ID, m *Member) {
		if found || e == exclude || m.Colony != col || m.Role != role || r.Despawning(e) {
			return
		}
		a := agents.Ptr(e)
		if a == nil || w.Tick() < a.RecoverUntil || a.Trapped {
			return
		}
		if a.Doing != nil && (a.Doing.Action == ai.FleeAction || a.Doing.Action == ExpeditionActionName ||
			(role == RoleGatherer && a.Doing.Action == SearchAction) ||
			(role == RoleLookout && a.Doing.Action == WatchAction) ||
			(role == RoleScout && a.Doing.Action == ScoutAction)) {
			return
		}
		if !jobs.AssignedTo(w, e).IsNil() {
			return
		}
		found = true
	})
	return found
}

// jobRole devuelve el oficio especialista para un tipo de trabajo.
func jobRole(jobType string) string {
	switch jobType {
	case "haul":
		return RolePorter
	case JobPluck, JobBuild:
		return RoleBuilder
	default:
		return RoleGeneralist
	}
}

// scoreJob ajusta la utilidad de un trabajo según el oficio del actor, el
// estado de la despensa y la disponibilidad de especialistas (suplencias).
func scoreJob(c *ai.Ctx, _ ids.ID, j jobs.Job, base float64) (float64, bool) {
	r := c.W.ECS()
	m, ok := ecs.Get[Member](r).Get(c.Actor)
	if !ok || !HasRoles(c.W, m.Colony) {
		return base, true
	}
	isFoodHaul := j.Type == "haul" && ecs.Get[Food](r).Has(j.Target)
	isFiberHaul := j.Type == "haul" && ecs.Get[Material](r).Has(j.Target)
	isBuildWork := j.Type == JobPluck || j.Type == JobBuild
	wantRole := jobRole(j.Type)
	critical := CriticalFood(c.W, m.Colony)
	def := Deficit(c.W, m.Colony)

	if critical {
		if isFoodHaul {
			if m.Role != RolePorter && j.Assignee.IsNil() && c.W.Tick()-j.Since < ai.DecidePeriod &&
				specialistAvailable(c.W, m.Colony, RolePorter, c.Actor) {
				return 0, false
			}
			u := base + PantryUrgentBonus
			if m.Role == RolePorter {
				u += RoleBonus
			}
			return u, true
		}
		if isFiberHaul || isBuildWork {
			// Con comida crítica, la obra y sus fibras ceden ante el abastecimiento.
			return base - 0.12, true
		}
	}

	// Especialista del trabajo.
	if wantRole != RoleGeneralist && m.Role == wantRole {
		u := base + RoleBonus
		if m.Role == RolePorter {
			if def > 0 && isFoodHaul {
				u += 0.04
			} else if def == 0 && isFiberHaul {
				u += StockedFiberBonus
			}
		}
		return u, true
	}

	// Suplencia: si el trabajo está libre hace menos de CoverDelay y hay un
	// especialista disponible, dejar que el especialista lo tome primero.
	if wantRole != RoleGeneralist && j.Assignee.IsNil() && c.W.Tick()-j.Since < CoverDelay &&
		specialistAvailable(c.W, m.Colony, wantRole, c.Actor) {
		return 0, false
	}

	// Constructor cubriendo acarreo de fibra cuando el porteador no está libre.
	if m.Role == RoleBuilder && isFiberHaul && def == 0 {
		return base + 0.03, true
	}
	// Recolector ayudando a llevar una carga con la despensa en la meta.
	if m.Role == RoleGatherer && def == 0 && (isFoodHaul || isFiberHaul) {
		return base + 0.02, true
	}
	return base, true
}

// JobReason resume el motivo por el que un miembro está haciendo su tarea
// actual, para la ficha en la interfaz (propuesta, 18.5).
func JobReason(w *world.World, actor ids.ID) string {
	r := w.ECS()
	m, ok := ecs.Get[Member](r).Get(actor)
	if !ok {
		return ""
	}
	a, ok := ecs.Get[ai.Agent](r).Get(actor)
	if !ok || a.Doing == nil {
		return ""
	}
	switch a.Doing.Action {
	case "work":
		j, ok := ecs.Get[jobs.Job](r).Get(a.Doing.Target)
		if !ok {
			return ""
		}
		want := jobRole(j.Type)
		if CriticalFood(w, m.Colony) && ecs.Get[Food](r).Has(j.Target) {
			return "despensa baja"
		}
		if want != RoleGeneralist && m.Role != RoleGeneralist && m.Role != want {
			if want == RolePorter {
				return "cubre el acarreo"
			}
			return "cubre la obra"
		}
		if Deficit(w, m.Colony) > 0 && ecs.Get[Food](r).Has(j.Target) {
			return "despensa baja"
		}
		if ecs.Get[Material](r).Has(j.Target) || j.Type == JobPluck || j.Type == JobBuild {
			return "obra en marcha"
		}
	case SearchAction, "forage":
		if Deficit(w, m.Colony) > 0 {
			return "despensa baja"
		}
	case ScoutAction:
		return "abre ruta"
	}
	return ""
}

// registerRoles conecta el puntaje de oficios en el selector de trabajos.
func registerRoles(w *world.World) {
	ai.SetJobScorer(w, scoreJob)
}
