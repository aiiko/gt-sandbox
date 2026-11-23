package ai

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Exploración: un agente curioso cruza los portales que conoce hacia Spaces
// donde nunca estuvo, y al llegar da una vuelta. Los portales dinámicos (una
// mano apoyada) valen el doble: son raros y no duran.
const (
	ExploreRange   = 10 * units.Centimeter // la vuelta al llegar
	DynamicPortalX = 2.0                   // multiplicador de curiosidad
	RevisitFactor  = 0.5                   // curiosidad por un Space que ya conoce
)

// Visit es un Space donde estuvo el agente y la última vez que lo vio ahí.
type Visit struct {
	Space ids.ID     `json:"space"`
	At    clock.Tick `json:"at"`
}

// Explored se emite al llegar por primera vez a un Space explorando.
type Explored struct {
	Entity ids.ID `json:"entity"`
	Space  ids.ID `json:"space"`
}

func (Explored) Kind() string { return "explored" }

// HasVisited indica si el agente estuvo en el Space s.
func (a *Agent) HasVisited(s ids.ID) bool {
	_, ok := a.visitedAt(s)
	return ok
}

func (a *Agent) visitedAt(s ids.ID) (clock.Tick, bool) {
	for _, v := range a.Visited {
		if v.Space == s {
			return v.At, true
		}
	}
	return 0, false
}

func (a *Agent) visit(s ids.ID, now clock.Tick) {
	for i, v := range a.Visited {
		if v.Space == s {
			a.Visited[i].At = now
			return
		}
	}
	a.Visited = append(a.Visited, Visit{Space: s, At: now})
}

// curiosityFor es cuánto le interesa el Space s: todo si es nuevo, una parte
// si lo visitó hace más de Revisit, nada si no.
func (a *Agent) curiosityFor(s ids.ID, now clock.Tick) float64 {
	at, ok := a.visitedAt(s)
	switch {
	case !ok:
		return a.Curiosity
	case a.Revisit > 0 && now-at >= a.Revisit:
		return a.Curiosity * RevisitFactor
	}
	return 0
}

type explore struct{}

func (explore) Name() string { return "explore" }

// side devuelve, para un portal, el extremo propio y el otro si uno de sus
// extremos está en here.
func side(p *nav.Portal, here ids.ID) (mine, other nav.PortalEnd, ok bool) {
	switch here {
	case p.A.Space:
		return p.A, p.B, true
	case p.B.Space:
		return p.B, p.A, true
	}
	return nav.PortalEnd{}, nav.PortalEnd{}, false
}

func (explore) Options(c *Ctx) []Option {
	a := c.Agent
	if a.Curiosity <= 0 {
		return nil
	}
	r := c.W.ECS()
	l, ok := comp.Of(r).Location.Get(c.Actor)
	if !ok {
		return nil
	}
	portals := ecs.Get[nav.Portal](r)
	spaces := ecs.Get[space.Space](r)
	hereSp, _ := spaces.Get(l.Space)
	g := space.Of(r)
	var out []Option
	for _, k := range a.Known {
		p := portals.Ptr(k.Entity)
		if p == nil || p.Disabled || r.Despawning(k.Entity) {
			continue
		}
		mine, other, ok := side(p, l.Space)
		if !ok || !g.IsSpace(other.Space) || InDanger(c.W, a.Scale, other.Space, other.Pos) {
			continue
		}
		// La curiosidad casual sube a una palma apoyada y baja de ella, pero
		// no encadena recorridos internos por el cuerpo entero (eso es asunto
		// de exploradores y expediciones; propuesta 19.4).
		if otherSp, ok := spaces.Get(other.Space); ok && hereSp.Kind == space.Articulated && otherSp.Kind == space.Articulated {
			continue
		}
		u := a.curiosityFor(other.Space, c.W.Tick())
		if u <= 0 {
			continue
		}
		if ecs.Get[attach.Contact](r).Has(k.Entity) {
			u *= DynamicPortalX
		}
		u -= a.DistanceCost * distance(c, l.Space, mine.Pos)
		out = append(out, Option{Target: k.Entity, Utility: u})
	}
	return out
}

func (explore) Utility(c *Ctx, _ *Doing) float64 { return c.Agent.Curiosity }

func (explore) Start(c *Ctx, d *Doing) bool {
	r := c.W.ECS()
	p := ecs.Get[nav.Portal](r).Ptr(d.Target)
	l, ok := comp.Of(r).Location.Get(c.Actor)
	if p == nil || !ok {
		return false
	}
	_, other, ok := side(p, l.Space)
	d.Place, d.Goal = other.Space, other.Pos
	return ok
}

func (explore) Step(c *Ctx, d *Doing) Result {
	switch nav.Approach(c.W, c.Actor, d.Place, d.Goal, &d.Moving) {
	case nav.Failed:
		if d.Stage == 0 {
			return Failed
		}
		return Done // la vuelta es opcional
	case nav.Underway:
		return Running
	}
	if d.Stage == 1 {
		return Done
	}
	// Llegó al Space nuevo: lo registra y elige un punto para dar una vuelta.
	c.Agent.visit(d.Place, c.W.Tick())
	c.W.Emit(Explored{Entity: c.Actor, Space: d.Place})
	d.Stage = 1
	if p, ok := WanderPoint(c.W, c.Actor, ExploreRange, "ai.explore"); ok {
		d.Goal = p
		return Running
	}
	return Done
}

// WanderPoint elige un punto transitable al azar a menos de rng del actor,
// en su Space (hasta 8 intentos).
func WanderPoint(w *world.World, actor ids.ID, rng units.Length, stream string) (units.Vec2, bool) {
	r := w.ECS()
	l, ok := comp.Of(r).Location.Get(actor)
	sp, ok2 := ecs.Get[space.Space](r).Get(l.Space)
	if !ok || !ok2 {
		return units.Vec2{}, false
	}
	grid := ecs.Get[nav.Grid](r).Ptr(l.Space)
	scale := ""
	if a, ok := ecs.Get[Agent](r).Get(actor); ok {
		scale = a.Scale
	}
	s := w.RNG(stream)
	for range 8 {
		p := l.Pos.XY().Add(units.V2(
			units.Length(s.Int64N(int64(2*rng+1)))-rng,
			units.Length(s.Int64N(int64(2*rng+1)))-rng))
		if sp.Extent.Contains(p) && (grid == nil || grid.Walkable(p)) && !InDanger(w, scale, l.Space, p) {
			return p, true
		}
	}
	return units.Vec2{}, false
}

func (explore) Stop(c *Ctx, d *Doing, _ bool) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
	}
}

func (explore) Valid(c *Ctx, d *Doing) bool { return space.Of(c.W.ECS()).IsSpace(d.Place) }

func (explore) Interruptible(*Ctx, *Doing) bool { return true }
