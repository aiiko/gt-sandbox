package colony

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Expediciones de la colonia (propuesta, 18.3 y 18.4).
//
// Una expedición es una entidad temporal de la colonia con fase explícita:
// gathering (convocar) -> traveling (viajar/reagruparse) -> surveying
// (reconocer) -> returning (regresar y contar). Cada integrante conserva su
// ubicación, necesidades y huida individual. Los hallazgos hechos durante la
// salida se guardan en la expedición y se comunican a la colonia al volver al
// nido. Quien queda aislado conserva el componente Stranded con sus hallazgos
// y provisiones hasta poder regresar.

const (
	ExpeditionActionName = "expedition"

	ExpMinReadyNeed int32 = 600 // saciedad y energía mínimas al partir
	ExpLowEnergy    int32 = 350 // umbral de energía para conformarse en un waypoint y volver
	ExpMaxCompany         = 4
	ExpFunGain      int32 = 250

	DefaultGatherTicks   clock.Tick = 12 * clock.TPS
	DefaultRegroupTicks  clock.Tick = 20 * clock.TPS
	DefaultSurveyTicks   clock.Tick = 4 * clock.TPS
	DefaultMaxDuration   clock.Tick = 180 * clock.TPS
	DefaultExpCooldown   clock.Tick = 60 * clock.TPS
	DefaultReturnTimeout clock.Tick = 180 * clock.TPS
	StrandedRetryTicks   clock.Tick = 15 * clock.TPS

	GatherRadius = 8 * units.Centimeter
	ArriveRadius = 6 * units.Centimeter
)

// ExpPhase es la fase de una expedición.
type ExpPhase string

const (
	ExpGathering ExpPhase = "gathering"
	ExpTraveling ExpPhase = "traveling"
	ExpSurveying ExpPhase = "surveying"
	ExpReturning ExpPhase = "returning"
)

// ExpeditionTarget es el componente "expedition_target" de un destino de
// expedición (mirador del respaldo, pasacables del zócalo, bolsillo del sofá).
type ExpeditionTarget struct {
	Colony      ids.ID     `json:"colony"`
	Clue        ids.ID     `json:"clue"`
	MinMembers  int        `json:"min_members"`
	MaxMembers  int        `json:"max_members,omitempty"`
	SurveyTicks clock.Tick `json:"survey_ticks,omitempty"`
	MaxDuration clock.Tick `json:"max_duration,omitempty"`
	Cooldown    clock.Tick `json:"cooldown,omitempty"`
	Next        clock.Tick `json:"next,omitempty"`
	Visited     bool       `json:"visited,omitempty"`
	Waypoints   []ids.ID   `json:"waypoints,omitempty"`
	ExtraClues  []ids.ID   `json:"extra_clues,omitempty"`
}

func (t ExpeditionTarget) minMembers() int { return max(1, t.MinMembers) }
func (t ExpeditionTarget) maxMembers() int {
	m := t.MaxMembers
	if m <= 0 {
		m = ExpMaxCompany
	}
	return min(max(m, t.minMembers()), ExpMaxCompany)
}
func (t ExpeditionTarget) survey() clock.Tick      { return cmpOr(t.SurveyTicks, DefaultSurveyTicks) }
func (t ExpeditionTarget) maxDuration() clock.Tick { return cmpOr(t.MaxDuration, DefaultMaxDuration) }
func (t ExpeditionTarget) cooldown() clock.Tick    { return cmpOr(t.Cooldown, DefaultExpCooldown) }

// ExpMember es el estado de un integrante dentro de una expedición.
type ExpMember struct {
	Entity   ids.ID `json:"entity"`
	Arrived  bool   `json:"arrived,omitempty"`
	Stranded bool   `json:"stranded,omitempty"`
	Returned bool   `json:"returned,omitempty"`
}

// Expedition es el componente "expedition" de una salida en preparación o en
// marcha (como máximo una por colonia).
type Expedition struct {
	Colony       ids.ID        `json:"colony"`
	Target       ids.ID        `json:"target"`
	Phase        ExpPhase      `json:"phase"`
	Guide        ids.ID        `json:"guide"`
	Members      []ExpMember   `json:"members"`
	WaypointIdx  int           `json:"waypoint_idx,omitempty"`
	Rendezvous   ids.ID        `json:"rendezvous,omitempty"`
	PhaseUntil   clock.Tick    `json:"phase_until,omitempty"`
	DepartTick   clock.Tick    `json:"depart_tick,omitempty"`
	MaxUntil     clock.Tick    `json:"max_until,omitempty"`
	Findings     []ids.ID      `json:"findings,omitempty"`
	SupportNotes []SupportNote `json:"support_notes,omitempty"`
	Reason       string        `json:"reason,omitempty"`
}

// Stranded es el componente "stranded" de un expedicionario que quedó aislado
// tras cerrarse un paso: conserva sus hallazgos y provisiones, busca refugio y
// reintenta volver al nido cuando cambia un portal o vence RetryAt.
type Stranded struct {
	Colony       ids.ID        `json:"colony"`
	Findings     []ids.ID      `json:"findings,omitempty"`
	SupportNotes []SupportNote `json:"support_notes,omitempty"`
	RetryAt      clock.Tick    `json:"retry_at,omitempty"`
	PortalSig    uint64        `json:"portal_sig,omitempty"`
}

// Eventos de expedición.
type (
	ExpeditionProposed struct {
		Expedition ids.ID `json:"expedition"`
		Colony     ids.ID `json:"colony"`
		Target     ids.ID `json:"target"`
		By         ids.ID `json:"by"`
	}
	ExpeditionJoined struct {
		Expedition ids.ID `json:"expedition"`
		Entity     ids.ID `json:"entity"`
	}
	ExpeditionDeparted struct {
		Expedition ids.ID   `json:"expedition"`
		Colony     ids.ID   `json:"colony"`
		Target     ids.ID   `json:"target"`
		Members    []ids.ID `json:"members"`
	}
	ExpeditionGuideChanged struct {
		Expedition ids.ID `json:"expedition"`
		Guide      ids.ID `json:"guide"`
	}
	ExpeditionSurveyed struct {
		Expedition ids.ID `json:"expedition"`
		Target     ids.ID `json:"target"`
		Findings   int    `json:"findings"`
	}
	ExpeditionReported struct {
		Expedition ids.ID `json:"expedition"`
		Colony     ids.ID `json:"colony"`
		Entity     ids.ID `json:"entity"`
		Findings   int    `json:"findings"`
	}
	ExpeditionCompleted struct {
		Expedition ids.ID `json:"expedition"`
		Colony     ids.ID `json:"colony"`
		Target     ids.ID `json:"target"`
		Early      bool   `json:"early,omitempty"`
		Reason     string `json:"reason,omitempty"`
	}
	ExpeditionCanceled struct {
		Expedition ids.ID `json:"expedition"`
		Colony     ids.ID `json:"colony"`
		Target     ids.ID `json:"target"`
		Reason     string `json:"reason"`
	}
	ExpeditionStranded struct {
		Expedition ids.ID `json:"expedition"`
		Entity     ids.ID `json:"entity"`
	}
	StrandedReturned struct {
		Colony   ids.ID `json:"colony"`
		Entity   ids.ID `json:"entity"`
		Findings int    `json:"findings"`
	}
)

func (ExpeditionProposed) Kind() string     { return "expedition_proposed" }
func (ExpeditionJoined) Kind() string       { return "expedition_joined" }
func (ExpeditionDeparted) Kind() string     { return "expedition_departed" }
func (ExpeditionGuideChanged) Kind() string { return "expedition_guide_changed" }
func (ExpeditionSurveyed) Kind() string     { return "expedition_surveyed" }
func (ExpeditionReported) Kind() string     { return "expedition_reported" }
func (ExpeditionCompleted) Kind() string    { return "expedition_completed" }
func (ExpeditionCanceled) Kind() string     { return "expedition_canceled" }
func (ExpeditionStranded) Kind() string     { return "expedition_stranded" }
func (StrandedReturned) Kind() string       { return "stranded_returned" }

// ActiveExpedition devuelve la expedición activa de la colonia col (si hay).
func ActiveExpedition(w *world.World, col ids.ID) (ids.ID, *Expedition) {
	r := w.ECS()
	var outID ids.ID
	var out *Expedition
	ecs.Get[Expedition](r).Each(func(id ids.ID, e *Expedition) {
		if out == nil && e.Colony == col && !r.Despawning(id) {
			outID, out = id, e
		}
	})
	return outID, out
}

// expeditionOf devuelve la expedición activa donde participa entity (sin
// haber vuelto ni quedado varado).
func expeditionOf(w *world.World, entity ids.ID) (ids.ID, *Expedition, int) {
	r := w.ECS()
	var outID ids.ID
	var out *Expedition
	idx := -1
	ecs.Get[Expedition](r).Each(func(id ids.ID, e *Expedition) {
		if out != nil || r.Despawning(id) {
			return
		}
		for i, m := range e.Members {
			if m.Entity == entity && !m.Returned && !m.Stranded {
				outID, out, idx = id, e, i
				return
			}
		}
	})
	return outID, out, idx
}

// inExpeditionField indica si entity está en una salida ya partida (viajando,
// reconociendo, regresando) o varada: sus hallazgos no se comparten todavía
// con la colonia, se guardan para reportar al volver al nido.
func inExpeditionField(w *world.World, entity ids.ID) bool {
	if ecs.Get[Stranded](w.ECS()).Has(entity) {
		return true
	}
	_, exp, _ := expeditionOf(w, entity)
	return exp != nil && exp.Phase != ExpGathering
}

// memberReady verifica que entity tenga saciedad y energía >= ExpMinReadyNeed.
func memberReady(w *world.World, entity ids.ID) bool {
	n, ok := ecs.Get[needs.Needs](w.ECS()).Get(entity)
	if !ok {
		return true
	}
	for _, name := range []string{"hunger", "energy"} {
		if v, ok := n.Get(name); ok && v < ExpMinReadyNeed {
			return false
		}
	}
	return true
}

// CanLaunchReason explica por qué una colonia no puede lanzar una expedición
// hacia target con participants ("" = puede partir).
func CanLaunchReason(w *world.World, col, target ids.ID, participants []ids.ID) string {
	r := w.ECS()
	t, ok := ecs.Get[ExpeditionTarget](r).Get(target)
	if !ok {
		return "destino no disponible"
	}
	if len(participants) < t.minMembers() {
		return "faltan integrantes"
	}
	for _, p := range participants {
		if !memberReady(w, p) {
			return "integrantes con poca energía o hambre"
		}
	}
	alive := AliveMembers(w, col)
	p, _ := ecs.Get[ai.Portions](r).Get(col)
	if p.Get(ai.ResourceFood)-int32(len(participants)) < alive {
		return "faltan provisiones en la despensa"
	}
	// Cobertura en casa: al menos 2 miembros disponibles fuera de la salida
	// (o 1 si la colonia tiene solo 2 no varados y la salida admite 1).
	stranded := ecs.Get[Stranded](r)
	agents := ecs.Get[ai.Agent](r)
	nonStranded := 0
	stayHome := 0
	ecs.Get[Member](r).Each(func(e ids.ID, m *Member) {
		if m.Colony != col || r.Despawning(e) || stranded.Has(e) {
			return
		}
		nonStranded++
		if slices.Contains(participants, e) {
			return
		}
		if a := agents.Ptr(e); a != nil && a.Trapped {
			return
		}
		stayHome++
	})
	minHome := 2
	if nonStranded == 2 && t.minMembers() == 1 {
		minHome = 1
	}
	if stayHome < minHome {
		return "falta cobertura en el nido"
	}
	return ""
}

// NoExpeditionReason resume qué impide iniciar una nueva expedición en col
// (para la ficha de colonia en la interfaz).
func NoExpeditionReason(w *world.World, col ids.ID) string {
	if _, exp := ActiveExpedition(w, col); exp != nil {
		return ""
	}
	r := w.ECS()
	if ecs.Get[ExpeditionTarget](r).Len() == 0 {
		return ""
	}
	alive := AliveMembers(w, col)
	if alive == 0 {
		return "sin miembros"
	}
	p, _ := ecs.Get[ai.Portions](r).Get(col)
	if p.Get(ai.ResourceFood) <= alive {
		return "faltan provisiones en la despensa"
	}
	stranded := ecs.Get[Stranded](r)
	available := 0
	ecs.Get[Member](r).Each(func(e ids.ID, m *Member) {
		if m.Colony == col && !r.Despawning(e) && !stranded.Has(e) {
			available++
		}
	})
	if available < 2 {
		return "falta cobertura en el nido"
	}
	return ""
}

func appendUniqueID(list []ids.ID, id ids.ID) []ids.ID {
	if id.IsNil() || slices.Contains(list, id) {
		return list
	}
	return append(list, id)
}

// pickGuide elige al guía entre los índices activos: prioriza a un miembro con
// oficio RoleScout y desempata por orden de ID (D5).
func pickGuide(w *world.World, members []ExpMember, active []int) ids.ID {
	ms := ecs.Get[Member](w.ECS())
	for _, i := range active {
		if m, ok := ms.Get(members[i].Entity); ok && m.Role == RoleScout {
			return members[i].Entity
		}
	}
	return members[active[0]].Entity
}

func portalSignature(w *world.World) uint64 {
	var sig uint64
	ecs.Get[nav.Portal](w.ECS()).Each(func(id ids.ID, p *nav.Portal) {
		v := p.Version
		if p.Disabled {
			v ^= 1 << 63
		}
		sig = sig*1315423911 ^ uint64(id) ^ v
	})
	return sig
}

// nearEntity indica si a está a menos de dist de b en el mismo Space.
func nearEntity(w *world.World, a, b ids.ID, dist units.Length) bool {
	locs := comp.Of(w.ECS()).Location
	la, ok1 := locs.Get(a)
	lb, ok2 := locs.Get(b)
	return ok1 && ok2 && la.Space == lb.Space && la.Pos.XY().DistSq(lb.Pos.XY()) <= int64(dist)*int64(dist)
}

func appendUniqueNote(list []SupportNote, note SupportNote) []SupportNote {
	for _, n := range list {
		if n.BodySpace == note.BodySpace && n.HouseSpace == note.HouseSpace && n.Pose == note.Pose {
			return list
		}
	}
	return append(list, note)
}

func bodyPoseOf(r *ecs.Registry, sp ids.ID) string {
	if z, ok := ecs.Get[attach.BodyZone](r).Get(sp); ok && !z.Body.IsNil() {
		if st, ok := ecs.Get[attach.BodyState](r).Get(z.Body); ok {
			return st.Pose
		}
	}
	parent := space.Of(r).Parent(sp)
	if !parent.IsNil() {
		if st, ok := ecs.Get[attach.BodyState](r).Get(parent); ok {
			return st.Pose
		}
	}
	return ""
}

// reportFindings comunica los hallazgos de la expedición a toda la colonia,
// incorporando cada dato una sola vez y publicando acarreos sobre comida viva.
func reportFindings(w *world.World, col, reporter ids.ID, findings []ids.ID, notes []SupportNote, expID ids.ID) int {
	r := w.ECS()
	agents := ecs.Get[ai.Agent](r)
	members := ecs.Get[Member](r)
	food := ecs.Get[Food](r)
	targets := ecs.Get[ExpeditionTarget](r)
	count := 0
	for _, f := range findings {
		if !r.Alive(f) || r.Despawning(f) {
			continue
		}
		count++
		if t := targets.Ptr(f); t != nil {
			t.Visited = true
		}
		members.Each(func(e ids.ID, m *Member) {
			if m.Colony != col || r.Despawning(e) || inExpeditionField(w, e) {
				return
			}
			if a := agents.Ptr(e); a != nil {
				ai.Learn(w, a, e, f)
			}
		})
		if food.Has(f) && !atNest(w, col, f) && !hauled(w, f) {
			if nest, ok := comp.Of(r).Location.Get(col); ok {
				_ = jobs.PostJob{Type: "haul", Target: f, DestSpace: nest.Space, Dest: nest.Pos.XY()}.Apply(w)
			}
		}
	}
	if colComp := ecs.Get[Colony](r).Ptr(col); colComp != nil {
		for _, n := range notes {
			if !r.Alive(n.BodySpace) || !r.Alive(n.HouseSpace) {
				continue
			}
			before := len(colComp.KnownSupports)
			colComp.KnownSupports = appendUniqueNote(colComp.KnownSupports, n)
			if len(colComp.KnownSupports) > before {
				count++
				w.Emit(SupportDiscovered{
					Colony: col, Entity: reporter,
					BodySpace: n.BodySpace, HouseSpace: n.HouseSpace, Pose: n.Pose,
				})
			}
		}
	}
	if !expID.IsNil() {
		w.Emit(ExpeditionReported{Expedition: expID, Colony: col, Entity: reporter, Findings: count})
	}
	return count
}

// markStranded pasa a un expedicionario a regreso individual (Stranded).
func markStranded(w *world.World, expID ids.ID, exp *Expedition, idx int) {
	e := exp.Members[idx].Entity
	exp.Members[idx].Stranded = true
	ecs.Get[Stranded](w.ECS()).Set(e, Stranded{
		Colony:       exp.Colony,
		Findings:     slices.Clone(exp.Findings),
		SupportNotes: slices.Clone(exp.SupportNotes),
		RetryAt:      w.Tick() + StrandedRetryTicks,
		PortalSig:    portalSignature(w),
	})
	w.Emit(ExpeditionStranded{Expedition: expID, Entity: e})
}

// expeditionSystem gestiona el ciclo de vida de las expediciones activas.
type expeditionSystem struct{}

func (expeditionSystem) Name() string { return "colony.expeditions" }

func (expeditionSystem) Run(w *world.World) {
	r := w.ECS()
	exps := ecs.Get[Expedition](r)
	strandedStore := ecs.Get[Stranded](r)

	// 1. Registrar hallazgos hechos por expedicionarios y varados (del tick anterior).
	for _, ev := range w.Events() {
		switch x := ev.Event.(type) {
		case ai.Noticed:
			if !r.Alive(x.Target) || r.Despawning(x.Target) {
				continue
			}
			if _, exp, _ := expeditionOf(w, x.Entity); exp != nil && exp.Phase != ExpGathering {
				exp.Findings = appendUniqueID(exp.Findings, x.Target)
			} else if st := strandedStore.Ptr(x.Entity); st != nil {
				st.Findings = appendUniqueID(st.Findings, x.Target)
			}
		case space.EntityReparented:
			if x.Cause != "portal" {
				continue
			}
			var note *SupportNote
			spFrom, ok1 := ecs.Get[space.Space](r).Get(x.From)
			spTo, ok2 := ecs.Get[space.Space](r).Get(x.To)
			if ok1 && ok2 && (spFrom.Kind == space.Articulated) != (spTo.Kind == space.Articulated) {
				bodySp, houseSp := x.From, x.To
				if spTo.Kind == space.Articulated {
					bodySp, houseSp = x.To, x.From
				}
				note = &SupportNote{BodySpace: bodySp, HouseSpace: houseSp, Pose: bodyPoseOf(r, bodySp), Tick: w.Tick()}
			}
			if _, exp, _ := expeditionOf(w, x.Entity); exp != nil && exp.Phase != ExpGathering {
				exp.Findings = appendUniqueID(exp.Findings, x.To)
				if note != nil {
					exp.SupportNotes = appendUniqueNote(exp.SupportNotes, *note)
				}
			} else if st := strandedStore.Ptr(x.Entity); st != nil {
				st.Findings = appendUniqueID(st.Findings, x.To)
				if note != nil {
					st.SupportNotes = appendUniqueNote(st.SupportNotes, *note)
				}
			}
		}
	}

	// 2. Actualizar cada expedición activa.
	for _, id := range exps.IDs() {
		exp := exps.Ptr(id)
		if exp == nil || r.Despawning(id) {
			continue
		}
		target := ecs.Get[ExpeditionTarget](r).Ptr(exp.Target)
		if target == nil || r.Despawning(exp.Target) {
			r.Despawn(id)
			continue
		}

		// Quitar miembros muertos.
		aliveMembers := exp.Members[:0]
		for _, m := range exp.Members {
			if r.Alive(m.Entity) && !r.Despawning(m.Entity) {
				aliveMembers = append(aliveMembers, m)
			}
		}
		exp.Members = aliveMembers

		// Verificar guía vivo y activo; si falta, priorizar un explorador (RoleScout) y luego menor ID (D5).
		guideOK := false
		var active []int
		for i, m := range exp.Members {
			if !m.Returned && !m.Stranded {
				active = append(active, i)
				if m.Entity == exp.Guide {
					guideOK = true
				}
			}
		}
		if !guideOK && len(active) > 0 {
			exp.Guide = pickGuide(w, exp.Members, active)
			w.Emit(ExpeditionGuideChanged{Expedition: id, Guide: exp.Guide})
		}

		switch exp.Phase {
		case ExpGathering:
			runGathering(w, id, exp, target)
		case ExpTraveling:
			runTraveling(w, id, exp, target, active)
		case ExpSurveying:
			runSurveying(w, id, exp, target, active)
		case ExpReturning:
			runReturning(w, id, exp, target, active)
		}
	}
}

func runGathering(w *world.World, id ids.ID, exp *Expedition, target *ExpeditionTarget) {
	r := w.ECS()
	arrived := 0
	for i := range exp.Members {
		if nearEntity(w, exp.Members[i].Entity, exp.Colony, GatherRadius) {
			exp.Members[i].Arrived = true
		}
		if exp.Members[i].Arrived {
			arrived++
		}
	}
	maxM := target.maxMembers()
	if arrived < maxM && w.Tick() < exp.PhaseUntil {
		return
	}
	// Evaluar salida con quienes llegaron y están en condiciones.
	var ready []ExpMember
	var readyIDs []ids.ID
	for _, m := range exp.Members {
		if m.Arrived && memberReady(w, m.Entity) {
			ready = append(ready, ExpMember{Entity: m.Entity})
			readyIDs = append(readyIDs, m.Entity)
		}
	}
	reason := CanLaunchReason(w, exp.Colony, exp.Target, readyIDs)
	if reason != "" {
		target.Next = w.Tick() + target.cooldown()
		w.Emit(ExpeditionCanceled{Expedition: id, Colony: exp.Colony, Target: exp.Target, Reason: reason})
		r.Despawn(id)
		return
	}
	// Equipar 1 ración por participante en el mismo tick.
	for _, e := range readyIDs {
		if !equipRation(w, exp.Colony, e, 1) {
			// Reversión defensiva (no debería pasar tras CanLaunchReason).
			for _, prev := range readyIDs {
				returnRations(w, exp.Colony, prev)
			}
			target.Next = w.Tick() + target.cooldown()
			w.Emit(ExpeditionCanceled{Expedition: id, Colony: exp.Colony, Target: exp.Target, Reason: "faltan provisiones"})
			r.Despawn(id)
			return
		}
	}
	exp.Members = ready
	allIdx := make([]int, len(ready))
	for i := range ready {
		allIdx[i] = i
	}
	exp.Guide = pickGuide(w, ready, allIdx)
	exp.Phase = ExpTraveling
	exp.DepartTick = w.Tick()
	exp.MaxUntil = w.Tick() + target.maxDuration()
	exp.WaypointIdx = 0
	if len(target.Waypoints) > 0 {
		exp.Rendezvous = target.Waypoints[0]
	} else {
		exp.Rendezvous = exp.Target
	}
	exp.PhaseUntil = w.Tick() + DefaultRegroupTicks
	w.Emit(ExpeditionDeparted{Expedition: id, Colony: exp.Colony, Target: exp.Target, Members: readyIDs})
}

func startReturn(w *world.World, exp *Expedition, reason string) {
	exp.Phase = ExpReturning
	exp.Rendezvous = exp.Colony
	exp.PhaseUntil = w.Tick() + DefaultReturnTimeout
	if exp.Reason == "" {
		exp.Reason = reason
	}
}

func runTraveling(w *world.World, id ids.ID, exp *Expedition, target *ExpeditionTarget, active []int) {
	if len(active) < target.minMembers() {
		startReturn(w, exp, "under_min")
		return
	}
	if w.Tick() >= exp.MaxUntil {
		startReturn(w, exp, "timeout")
		return
	}
	arrivedCount := 0
	for _, i := range active {
		if nearEntity(w, exp.Members[i].Entity, exp.Rendezvous, ArriveRadius) {
			exp.Members[i].Arrived = true
		}
		if exp.Members[i].Arrived {
			arrivedCount++
		}
	}
	if arrivedCount == len(active) || (w.Tick() >= exp.PhaseUntil && arrivedCount >= target.minMembers()) {
		// Llegaron al punto de reunión actual.
		if exp.WaypointIdx < len(target.Waypoints) {
			exp.Findings = appendUniqueID(exp.Findings, target.Waypoints[exp.WaypointIdx])
			// Comprobar margen de energía y tiempo para seguir.
			elapsed := w.Tick() - exp.DepartTick
			lowMargin := w.Tick()+elapsed > exp.MaxUntil
			for _, i := range active {
				if n, ok := ecs.Get[needs.Needs](w.ECS()).Get(exp.Members[i].Entity); ok {
					if v, ok := n.Get("energy"); ok && v < ExpLowEnergy {
						lowMargin = true
					}
				}
			}
			if lowMargin {
				startReturn(w, exp, "low_margin")
				return
			}
			exp.WaypointIdx++
			if exp.WaypointIdx < len(target.Waypoints) {
				exp.Rendezvous = target.Waypoints[exp.WaypointIdx]
			} else {
				exp.Rendezvous = exp.Target
			}
			for i := range exp.Members {
				exp.Members[i].Arrived = false
			}
			exp.PhaseUntil = w.Tick() + DefaultRegroupTicks
			return
		}
		// Llegaron al destino final: reconocer.
		exp.Phase = ExpSurveying
		exp.PhaseUntil = w.Tick() + target.survey()
		return
	}
	if w.Tick() >= exp.PhaseUntil {
		startReturn(w, exp, "regroup_timeout")
	}
}

func runSurveying(w *world.World, id ids.ID, exp *Expedition, target *ExpeditionTarget, active []int) {
	if len(active) < target.minMembers() {
		startReturn(w, exp, "under_min")
		return
	}
	if w.Tick() < exp.PhaseUntil {
		return
	}
	r := w.ECS()
	exp.Findings = appendUniqueID(exp.Findings, exp.Target)
	for _, cl := range target.ExtraClues {
		if r.Alive(cl) {
			exp.Findings = appendUniqueID(exp.Findings, cl)
		}
	}
	// Compartir notas entre los participantes y sumar diversión.
	for _, i := range active {
		e := exp.Members[i].Entity
		if n := ecs.Get[needs.Needs](r).Ptr(e); n != nil {
			n.Add(ChatNeed, ExpFunGain)
		}
	}
	w.Emit(ExpeditionSurveyed{Expedition: id, Target: exp.Target, Findings: len(exp.Findings)})
	exp.Phase = ExpReturning
	exp.Rendezvous = exp.Colony
	exp.PhaseUntil = w.Tick() + DefaultReturnTimeout
}

func runReturning(w *world.World, id ids.ID, exp *Expedition, target *ExpeditionTarget, active []int) {
	r := w.ECS()
	remaining := 0
	for _, i := range active {
		e := exp.Members[i].Entity
		if nearEntity(w, e, exp.Colony, GatherRadius) {
			reportFindings(w, exp.Colony, e, exp.Findings, exp.SupportNotes, id)
			returnRations(w, exp.Colony, e)
			exp.Members[i].Returned = true
			continue
		}
		remaining++
	}
	if remaining == 0 {
		if exp.Reason != "" {
			target.Next = w.Tick() + target.cooldown()
		} else {
			target.Next = w.Tick() + target.cooldown()/2
		}
		w.Emit(ExpeditionCompleted{
			Expedition: id, Colony: exp.Colony, Target: exp.Target,
			Early: exp.Reason != "", Reason: exp.Reason,
		})
		r.Despawn(id)
	}
}

// --- Acción "expedition" ---

// ExpeditionAction maneja tres roles del agente en una sola acción:
// 1) regresar si está varado (Stranded),
// 2) cumplir la fase de su expedición activa,
// 3) unirse a una convocatoria abierta o proponer una nueva salida.
//
// Doing: Target es la expedición (o el ExpeditionTarget al proponer, o la
// colonia al regresar varado); Offer = 0 (en curso/varado), 1 (unirse), 2 (proponer).
type ExpeditionAction struct{}

func (ExpeditionAction) Name() string { return ExpeditionActionName }

func (ExpeditionAction) Options(c *ai.Ctx) []ai.Option {
	r := c.W.ECS()
	m, ok := ecs.Get[Member](r).Get(c.Actor)
	if !ok {
		return nil
	}
	// 1. Varado: intentar volver al nido cuando corresponda.
	if st, ok := ecs.Get[Stranded](r).Get(c.Actor); ok {
		return []ai.Option{{Target: st.Colony, Offer: 0, Utility: 0.32}}
	}
	// 2. Ya participa en una expedición activa.
	if expID, exp, _ := expeditionOf(c.W, c.Actor); exp != nil {
		return []ai.Option{{Target: expID, Offer: 0, Utility: 0.36}}
	}
	// 3. Proponer o unirse: requiere inclinación, despensa en meta, sin trabajo y en condiciones.
	if m.Adventure <= 0 || Deficit(c.W, m.Colony, ai.ResourceFood) > 0 || !memberReady(c.W, c.Actor) || !jobs.AssignedTo(c.W, c.Actor).IsNil() {
		return nil
	}
	if expID, exp := ActiveExpedition(c.W, m.Colony); exp != nil {
		if exp.Phase != ExpGathering {
			return nil
		}
		t, ok := ecs.Get[ExpeditionTarget](r).Get(exp.Target)
		if !ok || len(exp.Members) >= t.maxMembers() {
			return nil
		}
		u := 0.24 * m.Adventure
		if !t.Visited {
			u += 0.06
		}
		return []ai.Option{{Target: expID, Offer: 1, Utility: u}}
	}
	// Sin expedición activa: evaluar destinos conocidos.
	var out []ai.Option
	ecs.Get[ExpeditionTarget](r).Each(func(tid ids.ID, t *ExpeditionTarget) {
		if t.Colony != m.Colony || r.Despawning(tid) || c.W.Tick() < t.Next {
			return
		}
		if !t.Clue.IsNil() && !c.Agent.Knows(t.Clue) {
			return
		}
		// Comprobación preliminar: despensa y cobertura para al menos 1.
		alive := AliveMembers(c.W, m.Colony)
		p, _ := ecs.Get[ai.Portions](r).Get(m.Colony)
		if p.Get(ai.ResourceFood)-int32(t.minMembers()) < alive {
			return
		}
		u := 0.22 * m.Adventure
		if !t.Visited {
			u += 0.06
			if m.Role == RoleScout {
				u += RoleBonus
			}
		}
		out = append(out, ai.Option{Target: tid, Offer: 2, Utility: u})
	})
	return out
}

func (ExpeditionAction) Utility(c *ai.Ctx, d *ai.Doing) float64 {
	if ecs.Get[Stranded](c.W.ECS()).Has(c.Actor) {
		return 0.32
	}
	if _, exp, _ := expeditionOf(c.W, c.Actor); exp != nil {
		return 0.36
	}
	return 0
}

func (ExpeditionAction) Start(c *ai.Ctx, d *ai.Doing) bool {
	r := c.W.ECS()
	switch d.Offer {
	case 0:
		if ecs.Get[Stranded](r).Has(c.Actor) {
			return true
		}
		_, exp, _ := expeditionOf(c.W, c.Actor)
		return exp != nil
	case 1: // unirse a convocatoria
		exp := ecs.Get[Expedition](r).Ptr(d.Target)
		if exp == nil || exp.Phase != ExpGathering {
			return false
		}
		t, ok := ecs.Get[ExpeditionTarget](r).Get(exp.Target)
		if !ok || len(exp.Members) >= t.maxMembers() {
			return false
		}
		if _, other, _ := expeditionOf(c.W, c.Actor); other != nil {
			return false
		}
		exp.Members = append(exp.Members, ExpMember{Entity: c.Actor})
		slices.SortFunc(exp.Members, func(a, b ExpMember) int { return cmp.Compare(a.Entity, b.Entity) })
		c.W.Emit(ExpeditionJoined{Expedition: d.Target, Entity: c.Actor})
		return true
	case 2: // proponer nueva expedición
		m, ok := ecs.Get[Member](r).Get(c.Actor)
		if !ok {
			return false
		}
		if _, existing := ActiveExpedition(c.W, m.Colony); existing != nil {
			return false
		}
		t, ok := ecs.Get[ExpeditionTarget](r).Get(d.Target)
		if !ok || t.Colony != m.Colony || c.W.Tick() < t.Next {
			return false
		}
		targetID := d.Target
		expID := r.Spawn()
		ecs.Get[Expedition](r).Set(expID, Expedition{
			Colony:     m.Colony,
			Target:     targetID,
			Phase:      ExpGathering,
			Guide:      c.Actor,
			Members:    []ExpMember{{Entity: c.Actor}},
			Rendezvous: m.Colony,
			PhaseUntil: c.W.Tick() + DefaultGatherTicks,
		})
		d.Target = expID
		d.Offer = 0
		c.W.Emit(ExpeditionProposed{Expedition: expID, Colony: m.Colony, Target: targetID, By: c.Actor})
		return true
	}
	return false
}

func (ExpeditionAction) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	w := c.W
	r := w.ECS()
	// Caso 1: varado.
	if st := ecs.Get[Stranded](r).Ptr(c.Actor); st != nil {
		return stepStranded(w, c.Actor, d, st)
	}
	// Caso 2: en expedición activa.
	expID, exp, idx := expeditionOf(w, c.Actor)
	if exp == nil {
		return ai.Done
	}
	switch exp.Phase {
	case ExpGathering:
		l, ok := comp.Of(r).Location.Get(exp.Colony)
		if !ok {
			return ai.Failed
		}
		if !nearEntity(w, c.Actor, exp.Colony, GatherRadius) {
			if nav.Approach(w, c.Actor, l.Space, l.Pos.XY(), &d.Moving) == nav.Failed {
				return ai.Failed
			}
		} else if d.Moving {
			nav.Halt(w, c.Actor)
			d.Moving = false
		}
		return ai.Running
	case ExpTraveling:
		l, ok := comp.Of(r).Location.Get(exp.Rendezvous)
		if !ok {
			return ai.Failed
		}
		if nearEntity(w, c.Actor, exp.Rendezvous, ArriveRadius) {
			if d.Moving {
				nav.Halt(w, c.Actor)
				d.Moving = false
			}
			exp.Members[idx].Arrived = true
			return ai.Running
		}
		switch nav.Approach(w, c.Actor, l.Space, l.Pos.XY(), &d.Moving) {
		case nav.Failed:
			// Si está sobre una zona corporal y un paso/articulación se cerró
			// temporalmente, esperar en un punto de agarre mientras quede margen.
			if onArticulatedSpace(r, c.Actor) && w.Tick() < exp.PhaseUntil {
				if gp, ok := nearestGripLocation(r, c.Actor); ok {
					nav.Approach(w, c.Actor, gp.Space, gp.Pos.XY(), &d.Moving)
				}
				return ai.Running
			}
			// Paso cerrado o inalcanzable: intentar regresar; si tampoco hay
			// camino al nido, queda varado.
			if canReachNest(w, c.Actor, exp.Colony) {
				startReturn(w, exp, "path_blocked")
				d.Moving = false
				return ai.Running
			}
			markStranded(w, expID, exp, idx)
			return ai.Done
		}
		return ai.Running
	case ExpSurveying:
		if d.Moving {
			nav.Halt(w, c.Actor)
			d.Moving = false
		}
		return ai.Running
	case ExpReturning:
		l, ok := comp.Of(r).Location.Get(exp.Colony)
		if !ok {
			return ai.Failed
		}
		if nearEntity(w, c.Actor, exp.Colony, GatherRadius) {
			return ai.Running // runReturning lo completa en el próximo tick
		}
		if nav.Approach(w, c.Actor, l.Space, l.Pos.XY(), &d.Moving) == nav.Failed {
			markStranded(w, expID, exp, idx)
			return ai.Done
		}
		return ai.Running
	}
	return ai.Done
}

func stepStranded(w *world.World, actor ids.ID, d *ai.Doing, st *Stranded) ai.Result {
	r := w.ECS()
	nestLoc, ok := comp.Of(r).Location.Get(st.Colony)
	if !ok {
		return ai.Failed
	}
	if nearEntity(w, actor, st.Colony, GatherRadius) {
		count := reportFindings(w, st.Colony, actor, st.Findings, st.SupportNotes, ids.Nil)
		returnRations(w, st.Colony, actor)
		ecs.Get[Stranded](r).Remove(actor)
		w.Emit(StrandedReturned{Colony: st.Colony, Entity: actor, Findings: count})
		return ai.Done
	}
	sig := portalSignature(w)
	if sig != st.PortalSig || w.Tick() >= st.RetryAt || d.Stage == 1 {
		d.Stage = 1
		switch nav.Approach(w, actor, nestLoc.Space, nestLoc.Pos.XY(), &d.Moving) {
		case nav.Arrived:
			return ai.Running
		case nav.Underway:
			return ai.Running
		case nav.Failed:
			st.PortalSig = sig
			st.RetryAt = w.Tick() + StrandedRetryTicks
			d.Stage, d.Moving = 0, false
		}
	}
	// Mientras espera, acercarse a un punto de agarre (en el cuerpo) o a un refugio en su Space.
	if gp, ok := nearestGripLocation(r, actor); ok {
		nav.Approach(w, actor, gp.Space, gp.Pos.XY(), &d.Moving)
	} else if shLoc, ok := nearestShelter(w, actor); ok {
		nav.Approach(w, actor, shLoc.Space, shLoc.Pos.XY(), &d.Moving)
	}
	return ai.Running
}

func onArticulatedSpace(r *ecs.Registry, actor ids.ID) bool {
	l, ok := comp.Of(r).Location.Get(actor)
	if !ok {
		return false
	}
	sp, ok := ecs.Get[space.Space](r).Get(l.Space)
	return ok && sp.Kind == space.Articulated
}

func nearestGripLocation(r *ecs.Registry, actor ids.ID) (comp.Location, bool) {
	l, ok := comp.Of(r).Location.Get(actor)
	if !ok {
		return comp.Location{}, false
	}
	z, ok := ecs.Get[attach.BodyZone](r).Get(l.Space)
	if !ok || len(z.Grips) == 0 {
		return comp.Location{}, false
	}
	best := z.Grips[0]
	bestD := l.Pos.XY().DistSq(best.Pos)
	for _, g := range z.Grips[1:] {
		if d := l.Pos.XY().DistSq(g.Pos); d < bestD {
			best, bestD = g, d
		}
	}
	return comp.Location{Space: l.Space, Pos: units.V3(best.Pos.X, best.Pos.Y, 0)}, true
}

func nearestShelter(w *world.World, actor ids.ID) (comp.Location, bool) {
	r := w.ECS()
	l, ok := comp.Of(r).Location.Get(actor)
	if !ok {
		return comp.Location{}, false
	}
	var best comp.Location
	var bestD int64
	found := false
	ecs.Get[crush.Shelter](r).Each(func(e ids.ID, _ *crush.Shelter) {
		sl, ok := comp.Of(r).Location.Get(e)
		if !ok || sl.Space != l.Space || r.Despawning(e) {
			return
		}
		d := l.Pos.XY().DistSq(sl.Pos.XY())
		if !found || d < bestD {
			best, bestD, found = sl, d, true
		}
	})
	return best, found
}

func canReachNest(w *world.World, actor, col ids.ID) bool {
	r := w.ECS()
	al, ok1 := comp.Of(r).Location.Get(actor)
	nl, ok2 := comp.Of(r).Location.Get(col)
	if !ok1 || !ok2 {
		return false
	}
	scale := ""
	if mv, ok := ecs.Get[nav.Mover](r).Get(actor); ok {
		scale = mv.Scale
	}
	var router nav.Router
	_, err := router.RouteFor(w, actor, scale, nil, al.Space, al.Pos.XY(), nl.Space, nl.Pos.XY())
	return err == nil
}

func (ExpeditionAction) Stop(c *ai.Ctx, d *ai.Doing, _ bool) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
	}
}

func (ExpeditionAction) Valid(c *ai.Ctx, _ *ai.Doing) bool {
	if ecs.Get[Stranded](c.W.ECS()).Has(c.Actor) {
		return true
	}
	_, exp, _ := expeditionOf(c.W, c.Actor)
	return exp != nil
}

func (ExpeditionAction) Interruptible(*ai.Ctx, *ai.Doing) bool { return true }

// Invariantes de expediciones:
// - como máximo una expedición viva por colonia;
// - cada tiny participa en como máximo una expedición a la vez;
// - como máximo ExpMaxCompany integrantes, ordenados por ID (D5).
func checkExpeditions(w *world.World) error {
	r := w.ECS()
	var err error
	byColony := map[ids.ID]ids.ID{}
	byMember := map[ids.ID]ids.ID{}
	ecs.Get[Expedition](r).Each(func(id ids.ID, e *Expedition) {
		if r.Despawning(id) {
			return
		}
		if prev, dup := byColony[e.Colony]; dup {
			err = errors.Join(err, fmt.Errorf("colony: colonia %v con dos expediciones (%v y %v)", e.Colony, prev, id))
		}
		byColony[e.Colony] = id
		if len(e.Members) > ExpMaxCompany {
			err = errors.Join(err, fmt.Errorf("colony: expedición %v con %d miembros (máx %d)", id, len(e.Members), ExpMaxCompany))
		}
		for i, m := range e.Members {
			if i > 0 && e.Members[i-1].Entity >= m.Entity {
				err = errors.Join(err, fmt.Errorf("colony: miembros de expedición %v desordenados o duplicados", id))
			}
			if !m.Returned && !m.Stranded {
				if prev, dup := byMember[m.Entity]; dup {
					err = errors.Join(err, fmt.Errorf("colony: %v en dos expediciones (%v y %v)", m.Entity, prev, id))
				}
				byMember[m.Entity] = id
			}
		}
	})
	return err
}

func registerExpeditions(w *world.World) error {
	r := w.ECS()
	for _, err := range []error{
		register[ExpeditionTarget](r, "expedition_target"),
		register[Expedition](r, "expedition"),
		register[Stranded](r, "stranded"),
	} {
		if err != nil {
			return err
		}
	}
	w.AddInvariant("colony.expedition", checkExpeditions)
	return errors.Join(
		world.RegisterEvent[ExpeditionProposed](w),
		world.RegisterEvent[ExpeditionJoined](w),
		world.RegisterEvent[ExpeditionDeparted](w),
		world.RegisterEvent[ExpeditionGuideChanged](w),
		world.RegisterEvent[ExpeditionSurveyed](w),
		world.RegisterEvent[ExpeditionReported](w),
		world.RegisterEvent[ExpeditionCompleted](w),
		world.RegisterEvent[ExpeditionCanceled](w),
		world.RegisterEvent[ExpeditionStranded](w),
		world.RegisterEvent[StrandedReturned](w),
		world.RegisterEvent[SupportDiscovered](w),
		w.Register(world.SystemSpec{Phase: clock.PhaseDecide, Order: -35}, expeditionSystem{}),
	)
}
