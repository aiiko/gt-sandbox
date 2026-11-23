package colony

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Explorador de la colonia (oficio "scout").
//
// A diferencia de ai.explore (curiosidad individual solo sobre portales del
// Space actual), la acción "scout" recorre el mapa al servicio de la colonia:
// consulta todos los portales conocidos por el actor, prioriza superficies que
// ningún miembro de la colonia visitó todavía y, al llegar, hace un barrido
// amplio (ScoutRange) para descubrir nuevos portales, pistas de expedición,
// refugios y comida. Cede ante despensa crítica (CriticalFood).

const (
	ScoutAction             = "scout"
	ScoutRange              = 50 * units.Centimeter
	ScoutRevisit clock.Tick = 180 * clock.TPS
)

// Scout es la acción "scout". Doing: Target es el portal elegido (o Nil si
// hace barrido en el Space actual); Place/Goal es el destino; Stage 0 llega
// al destino, Stage 1 barre los alrededores.
type Scout struct{}

func (Scout) Name() string { return ScoutAction }

// colonyVisitedAt devuelve el tick más reciente en que algún miembro vivo de
// col visitó el Space sp (visited=false si nadie lo visitó).
func colonyVisitedAt(w *world.World, col, sp ids.ID) (clock.Tick, bool) {
	r := w.ECS()
	agents := ecs.Get[ai.Agent](r)
	var latest clock.Tick
	visited := false
	ecs.Get[Member](r).Each(func(e ids.ID, m *Member) {
		if m.Colony != col || r.Despawning(e) {
			return
		}
		a := agents.Ptr(e)
		if a == nil {
			return
		}
		for _, v := range a.Visited {
			if v.Space == sp && (!visited || v.At > latest) {
				latest, visited = v.At, true
			}
		}
	})
	return latest, visited
}

func (Scout) Options(c *ai.Ctx) []ai.Option {
	r := c.W.ECS()
	m, ok := ecs.Get[Member](r).Get(c.Actor)
	if !ok || !HasRoles(c.W, m.Colony) || CriticalFood(c.W, m.Colony) {
		return nil
	}
	if m.Role != RoleScout {
		if Deficit(c.W, m.Colony) > 0 || specialistAvailable(c.W, m.Colony, RoleScout, c.Actor) {
			return nil
		}
	}
	l, ok := comp.Of(r).Location.Get(c.Actor)
	if !ok {
		return nil
	}
	mvScale := ""
	if mv, ok := ecs.Get[nav.Mover](r).Get(c.Actor); ok {
		mvScale = mv.Scale
	}
	portals := ecs.Get[nav.Portal](r)
	contacts := ecs.Get[attach.Contact](r)
	g := space.Of(r)
	bonus := 0.0
	if m.Role == RoleScout {
		bonus = RoleBonus
	}

	var out []ai.Option
	seenSpaces := map[ids.ID]bool{l.Space: true}
	for _, k := range c.Agent.Known {
		p := portals.Ptr(k.Entity)
		if p == nil || p.Disabled || r.Despawning(k.Entity) || contacts.Has(k.Entity) {
			continue
		}
		if p.Scale != "" && p.Scale != mvScale {
			continue
		}
		for _, end := range [2]nav.PortalEnd{p.A, p.B} {
			if seenSpaces[end.Space] || !g.IsSpace(end.Space) || ai.InDanger(c.W, c.Agent.Scale, end.Space, end.Pos) {
				continue
			}
			at, vis := colonyVisitedAt(c.W, m.Colony, end.Space)
			base := 0.0
			switch {
			case !vis:
				base = 0.20
			case c.W.Tick()-at >= ScoutRevisit:
				base = 0.10
			default:
				continue
			}
			seenSpaces[end.Space] = true
			u := base + bonus - c.Agent.DistanceCost*ai.Distance(c, end.Space, end.Pos)
			out = append(out, ai.Option{Target: k.Entity, Offer: endOffer(p, end.Space), Utility: u})
		}
	}
	// Si es explorador y no hay portales hacia zonas sin visitar, hace barrido
	// amplio en su superficie actual para encontrar pistas o portales lejanos.
	if len(out) == 0 && m.Role == RoleScout {
		out = append(out, ai.Option{Utility: 0.12})
	}
	return out
}

func endOffer(p *nav.Portal, sp ids.ID) int {
	if p.B.Space == sp {
		return 1
	}
	return 0
}

func (Scout) Utility(c *ai.Ctx, d *ai.Doing) float64 {
	m, ok := ecs.Get[Member](c.W.ECS()).Get(c.Actor)
	if !ok || CriticalFood(c.W, m.Colony) {
		return 0
	}
	u := 0.12
	if !d.Target.IsNil() {
		u = 0.20
	}
	if m.Role == RoleScout {
		u += RoleBonus
	}
	return u
}

func (Scout) Start(c *ai.Ctx, d *ai.Doing) bool {
	r := c.W.ECS()
	if !d.Target.IsNil() {
		p := ecs.Get[nav.Portal](r).Ptr(d.Target)
		if p == nil || p.Disabled {
			return false
		}
		end := p.A
		if d.Offer == 1 {
			end = p.B
		}
		d.Place, d.Goal = end.Space, end.Pos
		return space.Of(r).IsSpace(d.Place)
	}
	l, ok := comp.Of(r).Location.Get(c.Actor)
	p, ok2 := ai.WanderPoint(c.W, c.Actor, ScoutRange, "colony.scout")
	if !ok || !ok2 {
		return false
	}
	d.Place, d.Goal, d.Stage = l.Space, p, 1
	return true
}

func (Scout) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	switch nav.Approach(c.W, c.Actor, d.Place, d.Goal, &d.Moving) {
	case nav.Failed:
		if d.Stage == 0 {
			return ai.Failed
		}
		return ai.Done
	case nav.Underway:
		return ai.Running
	}
	if d.Stage == 1 {
		return ai.Done
	}
	// Llegó a la superficie nueva: hacer un barrido amplio para descubrir pistas.
	d.Stage, d.Moving = 1, false
	if p, ok := ai.WanderPoint(c.W, c.Actor, ScoutRange, "colony.scout"); ok {
		d.Goal = p
		return ai.Running
	}
	return ai.Done
}

func (Scout) Stop(c *ai.Ctx, d *ai.Doing, _ bool) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
	}
}

func (Scout) Valid(c *ai.Ctx, d *ai.Doing) bool { return space.Of(c.W.ECS()).IsSpace(d.Place) }

func (Scout) Interruptible(*ai.Ctx, *ai.Doing) bool { return true }
