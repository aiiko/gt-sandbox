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
	u := m.Forage
	if HasRoles(c.W, m.Colony) && m.Role == RoleGatherer && Deficit(c.W, m.Colony, ai.ResourceFood) > 0 &&
		PendingFoodPortions(c.W, m.Colony) < MissingPortions(c.W, m.Colony) {
		u += 0.04
	}
	return []ai.Option{{Utility: u}}
}

func (Forage) Utility(c *ai.Ctx, _ *ai.Doing) float64 {
	m, _ := ecs.Get[Member](c.W.ECS()).Get(c.Actor)
	u := m.Forage
	if HasRoles(c.W, m.Colony) && m.Role == RoleGatherer && Deficit(c.W, m.Colony, ai.ResourceFood) > 0 &&
		PendingFoodPortions(c.W, m.Colony) < MissingPortions(c.W, m.Colony) {
		u += 0.04
	}
	return u
}

// Start elige el destino: un punto transitable al azar cercano.
func (Forage) Start(c *ai.Ctx, d *ai.Doing) bool {
	l, ok := comp.Of(c.W.ECS()).Location.Get(c.Actor)
	p, ok2 := ai.WanderPoint(c.W, c.Actor, ForageRange, "colony.forage")
	if !ok || !ok2 {
		return false
	}
	d.Place, d.Goal = l.Space, p
	return true
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
