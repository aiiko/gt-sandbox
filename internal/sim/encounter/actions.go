package encounter

import (
	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Of devuelve el componente encounter del actor (nil si no tiene).
func Of(c *ai.Ctx) *Encounter { return ecs.Get[Encounter](c.W.ECS()).Ptr(c.Actor) }

// spot es dónde se para la giantess para atender la superficie s: el lugar
// ("place") del mueble, que es la superficie misma.
type spot struct {
	space  ids.ID
	goal   units.Vec2
	facing *fixedmath.Angle
}

func spotOf(w *world.World, s ids.ID) (spot, bool) {
	r := w.ECS()
	pl, ok := ecs.Get[ai.Place](r).Get(s)
	loc, ok2 := comp.Of(r).Location.Get(s)
	if !ok || !ok2 || r.Despawning(s) || !space.Of(r).IsSpace(s) {
		return spot{}, false
	}
	return spot{space: loc.Space, goal: loc.Pos.XY().Add(pl.Spot), facing: pl.Facing}, true
}

// distance es la distancia de mundo, en metros, entre el actor y el punto.
func distance(c *ai.Ctx, sp spot) float64 {
	g := space.Of(c.W.ECS())
	a, _, err1 := g.WorldPose(c.Actor)
	b, err2 := g.ToWorld(sp.space, units.V3(sp.goal.X, sp.goal.Y, 0))
	if err1 != nil || err2 != nil {
		return 0
	}
	return float64(a.Sub(b).XY().Len()) / float64(units.Meter)
}

// recent indica si el avistamiento es lo bastante nuevo para un encuentro.
func recent(w *world.World, s Sighting) bool { return w.Tick()-s.Last <= Recent && s.Seen > 0 }

// options arma una opción de action por superficie con avistamiento
// reciente, lugar donde pararse y sin enfriamiento, si ok lo acepta. La
// utilidad es fija (utility) menos el costo de la distancia.
func options(c *ai.Ctx, action string, utility float64, ok func(enc *Encounter, s Sighting) bool) []ai.Option {
	enc := Of(c)
	if enc == nil || utility <= 0 {
		return nil
	}
	var out []ai.Option
	for _, s := range enc.Sightings {
		if !recent(c.W, s) || enc.Cooling(s.Surface, action, c.W.Tick()) || ok != nil && !ok(enc, s) {
			continue
		}
		sp, found := spotOf(c.W, s.Surface)
		if !found {
			continue
		}
		out = append(out, ai.Option{Target: s.Surface, Utility: utility - c.Agent.DistanceCost*distance(c, sp)})
	}
	return out
}

// skip termina un encuentro cuya condición ya no se cumple.
func skip(c *ai.Ctx, d *ai.Doing, reason string) ai.Result {
	c.W.Emit(Skipped{Entity: c.Actor, Action: d.Action, Surface: d.Target, Reason: reason})
	return ai.Done
}

// sighting es el avistamiento vigente de la superficie de d.
func sighting(c *ai.Ctx, d *ai.Doing) (Sighting, bool) {
	enc := Of(c)
	if enc == nil {
		return Sighting{}, false
	}
	s, ok := enc.Sighting(d.Target)
	return s, ok && recent(c.W, s)
}

// --- observe: ir a mirar una superficie con tinies ---

const (
	// ObserveTicks es cuánto se queda mirando.
	ObserveTicks = 10 * 20
	// ObserveCooldown es cuánto tarda en volver a mirar la misma superficie.
	ObserveCooldown = 3 * 60 * 20
)

// Observed se emite al terminar de mirar: cuántos tinies vio al final.
type Observed struct {
	Entity  ids.ID `json:"entity"`
	Surface ids.ID `json:"surface"`
	Seen    int    `json:"seen"`
}

func (Observed) Kind() string { return "observed" }

// Observe es la acción "observe": ir al lugar de una superficie donde vio
// tinies hace poco y quedarse mirando con la mano en reposo (fuera de su
// alcance, así que no los asusta). Mientras mira, encounter.look sigue
// registrando. Al empezar, la superficie se enfría para esta acción.
type Observe struct{}

func (Observe) Name() string { return "observe" }

func (Observe) What(*ai.Ctx, *ai.Doing) string { return "observe" }

func (Observe) Options(c *ai.Ctx) []ai.Option {
	enc := Of(c)
	if enc == nil {
		return nil
	}
	return options(c, "observe", enc.Observe, nil)
}

func (Observe) Utility(c *ai.Ctx, _ *ai.Doing) float64 {
	if enc := Of(c); enc != nil {
		return enc.Observe
	}
	return 0
}

func (Observe) Start(c *ai.Ctx, d *ai.Doing) bool {
	enc := Of(c)
	if _, ok := spotOf(c.W, d.Target); !ok || enc == nil {
		return false
	}
	enc.Cool(d.Target, "observe", c.W.Tick()+ObserveCooldown)
	return true
}

func (Observe) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	sp, ok := spotOf(c.W, d.Target)
	if !ok {
		return ai.Failed
	}
	switch d.Stage {
	case 0:
		if r, arrived := goTo(c, d, sp); !arrived {
			return r
		}
		if _, ok := sighting(c, d); !ok {
			return skip(c, d, "gone")
		}
		d.Stage = 1
		fallthrough
	default:
		if d.Progress++; d.Progress < ObserveTicks {
			return ai.Running
		}
		s, _ := Of(c).Sighting(d.Target)
		c.W.Emit(Observed{Entity: c.Actor, Surface: d.Target, Seen: s.Seen})
		return ai.Done
	}
}

// goTo acerca al actor al lugar sp; al llegar lo orienta. arrived indica si
// ya está.
func goTo(c *ai.Ctx, d *ai.Doing, sp spot) (ai.Result, bool) {
	switch nav.Approach(c.W, c.Actor, sp.space, sp.goal, &d.Moving) {
	case nav.Failed:
		return ai.Failed, false
	case nav.Underway:
		return ai.Running, false
	}
	if l := comp.Of(c.W.ECS()).Location.Ptr(c.Actor); sp.facing != nil && l != nil {
		l.Facing = *sp.facing
	}
	return ai.Running, true
}

func (Observe) Stop(c *ai.Ctx, d *ai.Doing, _ bool) { halt(c, d) }

// halt detiene al actor si se estaba moviendo.
func halt(c *ai.Ctx, d *ai.Doing) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
		d.Moving = false
	}
}

func (Observe) Valid(c *ai.Ctx, d *ai.Doing) bool {
	_, ok := spotOf(c.W, d.Target)
	return ok
}

func (Observe) Interruptible(*ai.Ctx, *ai.Doing) bool { return true }
