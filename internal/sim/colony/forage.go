package colony

import (
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
)

// ForageRange es hasta dónde sale a buscar un tiny en cada salida.
const ForageRange = 30 * units.Centimeter

// Forage es la acción "forage": caminar a un punto al azar cercano (en el
// mismo Space) para ver qué hay. Vale Member.Forage, así que solo gana
// cuando no hay nada mejor.
type Forage struct{}

func (Forage) Name() string { return "forage" }

func (Forage) Options(c *ai.Ctx) []ai.Option {
	m, ok := ecs.Get[Member](c.W.ECS()).Get(c.Actor)
	if !ok || m.Forage <= 0 {
		return nil
	}
	return []ai.Option{{Utility: m.Forage}}
}

func (Forage) Utility(c *ai.Ctx, _ *ai.Doing) float64 {
	m, _ := ecs.Get[Member](c.W.ECS()).Get(c.Actor)
	return m.Forage
}

// Start elige el destino: hasta 8 intentos de un punto transitable al azar.
func (Forage) Start(c *ai.Ctx, d *ai.Doing) bool {
	r := c.W.ECS()
	l, ok := comp.Of(r).Location.Get(c.Actor)
	sp, ok2 := ecs.Get[space.Space](r).Get(l.Space)
	if !ok || !ok2 {
		return false
	}
	grid := ecs.Get[nav.Grid](r).Ptr(l.Space)
	rng := c.W.RNG("colony.forage")
	for range 8 {
		p := l.Pos.XY().Add(units.V2(
			units.Length(rng.Int64N(int64(2*ForageRange+1)))-ForageRange,
			units.Length(rng.Int64N(int64(2*ForageRange+1)))-ForageRange))
		if sp.Extent.Contains(p) && (grid == nil || grid.Walkable(p)) {
			d.Place, d.Goal = l.Space, p
			return true
		}
	}
	return false
}

func (Forage) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	switch nav.Approach(c.W, c.Actor, d.Place, d.Goal, &d.Moving) {
	case nav.Arrived:
		return ai.Done
	case nav.Failed:
		return ai.Failed
	}
	return ai.Running
}

func (Forage) Stop(c *ai.Ctx, d *ai.Doing, _ bool) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
	}
}

func (Forage) Valid(c *ai.Ctx, d *ai.Doing) bool { return space.Of(c.W.ECS()).IsSpace(d.Place) }

func (Forage) Interruptible(*ai.Ctx, *ai.Doing) bool { return true }
