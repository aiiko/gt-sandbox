package ai

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
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
func (a *Agent) Knows(e ids.ID) bool { return a.Sight <= 0 || a.known(e) }

func (a *Agent) known(e ids.ID) bool {
	for _, k := range a.Known {
		if k.Entity == e {
			return true
		}
	}
	return false
}

// Learn hace que el agente actor conozca e (si no lo conocía) y emite
// Noticed. También lo registra un agente sin Sight (que igual conoce todo),
// para que Noticed diga qué notó.
func Learn(w *world.World, a *Agent, actor, e ids.ID) {
	if a.known(e) || !w.ECS().Alive(e) {
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

// look: cada LookPeriod ticks, cada agente con vista registra el Space donde
// está y aprende los smart objects y los extremos de portal que tiene a la
// vista; cada agente con Watch nota a los agentes de otra escala cercanos.
type look struct{ it agentIter }

func (look) Name() string { return "ai.look" }

func (s look) Run(w *world.World) {
	r := w.ECS()
	offers := ecs.Get[Offers](r)
	locs := comp.Of(r).Location
	ix := spatial.Of(w)
	s.it.each(w, func(c *Ctx) {
		a := c.Agent
		if (a.Sight <= 0 && a.Watch <= 0) || !clock.Staggered(w.Tick(), LookPeriod, uint64(c.Actor.Index())) {
			return
		}
		a.Known = keep(a.Known, func(k Known) bool { return r.Alive(k.Entity) && !r.Despawning(k.Entity) })
		l, ok := locs.Get(c.Actor)
		if !ok {
			return
		}
		if a.Watch > 0 {
			s.watch(w, c)
		}
		if a.Sight <= 0 {
			return
		}
		a.visit(l.Space)
		ix.Near(l.Space, l.Pos.XY(), a.Sight, func(e ids.ID, _ units.Vec2) {
			if e != c.Actor && offers.Has(e) && !r.Despawning(e) {
				Learn(w, a, c.Actor, e)
			}
		})
		r2 := int64(a.Sight) * int64(a.Sight)
		ecs.Get[nav.Portal](r).Each(func(e ids.ID, p *nav.Portal) {
			if mine, _, ok := side(p, l.Space); ok && !r.Despawning(e) && mine.Pos.DistSq(l.Pos.XY()) <= r2 {
				Learn(w, a, c.Actor, e)
			}
		})
	})
}

// watch: el agente nota a los agentes de otra escala y a los restos de
// aplastados a menos de Watch (en el mundo).
func (s look) watch(w *world.World, c *Ctx) {
	g := space.Of(w.ECS())
	me, _, err := g.WorldPose(c.Actor)
	if err != nil {
		return
	}
	r2 := int64(c.Agent.Watch) * int64(c.Agent.Watch)
	var seen []ids.ID
	s.it.agents.Each(func(e ids.ID, o *Agent) {
		if e == c.Actor || o.Scale == c.Agent.Scale || c.Agent.known(e) || w.ECS().Despawning(e) {
			return
		}
		if p, _, err := g.WorldPose(e); err == nil && p.XY().DistSq(me.XY()) <= r2 {
			seen = append(seen, e)
		}
	})
	// también los restos de un aplastado
	ecs.Get[crush.Remains](w.ECS()).Each(func(e ids.ID, _ *crush.Remains) {
		if c.Agent.known(e) || w.ECS().Despawning(e) {
			return
		}
		if p, _, err := g.WorldPose(e); err == nil && p.XY().DistSq(me.XY()) <= r2 {
			seen = append(seen, e)
		}
	})
	for _, e := range seen { // fuera del Each: Learn no toca el store, pero por las dudas
		Learn(w, c.Agent, c.Actor, e)
	}
}
