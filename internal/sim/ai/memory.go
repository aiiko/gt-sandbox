package ai

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/spatial"
	"gtsandbox/internal/sim/world"
)

// Percepción y memoria.
//
// Un agente con Sight > 0 solo usa los smart objects que conoce: los que vio
// (a distancia <= Sight, en su mismo Space) o los que le contaron (Learn).
// Sin Sight, conoce todo (la giantess, por ahora).
const (
	LookPeriod clock.Tick = 10 // cada cuánto mira un agente
	MaxKnown              = 64 // al pasarse, olvida lo más viejo
	// RetryTicks es cuánto evita un agente un objetivo con el que falló.
	RetryTicks clock.Tick = 30 * clock.TPS
)

// Known es algo que el agente conoce, y desde cuándo.
type Known struct {
	Entity ids.ID     `json:"entity"`
	Since  clock.Tick `json:"since"`
}

// Avoid es un objetivo que el agente no reintenta hasta Until.
type Avoid struct {
	Target ids.ID     `json:"target"`
	Until  clock.Tick `json:"until"`
}

// Noticed se emite cuando un agente se entera de algo nuevo (lo ve o se lo
// cuentan).
type Noticed struct {
	Entity ids.ID `json:"entity"`
	Target ids.ID `json:"target"`
}

func (Noticed) Kind() string { return "noticed" }

// Knows indica si el agente conoce e.
func (a *Agent) Knows(e ids.ID) bool {
	if a.Sight <= 0 {
		return true
	}
	for _, k := range a.Known {
		if k.Entity == e {
			return true
		}
	}
	return false
}

// Learn hace que el agente actor conozca e (si no lo conocía) y emite
// Noticed.
func Learn(w *world.World, a *Agent, actor, e ids.ID) {
	if a.Sight <= 0 || a.Knows(e) || !w.ECS().Alive(e) {
		return
	}
	if len(a.Known) >= MaxKnown {
		a.Known = a.Known[1:]
	}
	a.Known = append(a.Known, Known{Entity: e, Since: w.Tick()})
	w.Emit(Noticed{Entity: actor, Target: e})
}

func (a *Agent) avoids(target ids.ID, now clock.Tick) bool {
	for _, x := range a.Avoid {
		if x.Target == target && now < x.Until {
			return true
		}
	}
	return false
}

func (a *Agent) avoid(target ids.ID, now clock.Tick) {
	a.prune(now)
	a.Avoid = append(a.Avoid, Avoid{Target: target, Until: now + RetryTicks})
}

// prune olvida lo que ya no existe y las evitaciones vencidas.
func (a *Agent) prune(now clock.Tick) {
	a.Avoid = keep(a.Avoid, func(x Avoid) bool { return now < x.Until })
}

func keep[T any](s []T, ok func(T) bool) []T {
	out := s[:0]
	for _, x := range s {
		if ok(x) {
			out = append(out, x)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// look: cada agente con vista mira a su alrededor cada LookPeriod ticks.
type look struct{ it agentIter }

func (look) Name() string { return "ai.look" }

func (s look) Run(w *world.World) {
	r := w.ECS()
	offers := ecs.Get[Offers](r)
	locs := comp.Of(r).Location
	ix := spatial.Of(w)
	s.it.each(w, func(c *Ctx) {
		a := c.Agent
		if a.Sight <= 0 || !clock.Staggered(w.Tick(), LookPeriod, uint64(c.Actor.Index())) {
			return
		}
		a.Known = keep(a.Known, func(k Known) bool { return r.Alive(k.Entity) && !r.Despawning(k.Entity) })
		l, ok := locs.Get(c.Actor)
		if !ok {
			return
		}
		ix.Near(l.Space, l.Pos.XY(), a.Sight, func(e ids.ID, _ units.Vec2) {
			if e != c.Actor && offers.Has(e) && !r.Despawning(e) {
				Learn(w, a, c.Actor, e)
			}
		})
	})
}
