package colony

import (
	"errors"
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/world"
)

// Provisiones de expedición (propuesta, 18.3).
//
// Al partir en expedición, cada participante recibe una porción de la despensa
// del nido en su reserva personal (Provisions.Rations) en el mismo tick. Puede
// comer una ración durante una pausa segura (acción "eat_ration"), y al volver
// vivo al nido las sobrantes regresan al depósito en el mismo tick. Si muere,
// su reserva muere con la entidad.

const (
	EatRationAction               = "eat_ration"
	EatRationTicks     clock.Tick = 2 * clock.TPS
	EatRationThreshold int32      = 550
	EatRationGain      int32      = 350
)

// Provisions es el componente "provisions" con las raciones personales que
// lleva un expedicionario.
type Provisions struct {
	Rations int32 `json:"rations"`
}

// Eventos de provisiones.
type (
	ProvisionsTaken struct {
		Colony ids.ID `json:"colony"`
		Entity ids.ID `json:"entity"`
		Amount int32  `json:"amount"`
	}
	RationEaten struct {
		Entity ids.ID `json:"entity"`
		Left   int32  `json:"left"`
	}
	ProvisionsReturned struct {
		Colony ids.ID `json:"colony"`
		Entity ids.ID `json:"entity"`
		Amount int32  `json:"amount"`
	}
)

func (ProvisionsTaken) Kind() string    { return "provisions_taken" }
func (RationEaten) Kind() string        { return "ration_eaten" }
func (ProvisionsReturned) Kind() string { return "provisions_returned" }

// equipRation traslada amount porciones del depósito de col a la reserva
// personal de entity en el mismo tick. Devuelve false si no alcanzan.
func equipRation(w *world.World, col, entity ids.ID, amount int32) bool {
	if amount <= 0 {
		return true
	}
	r := w.ECS()
	portions := ecs.Get[ai.Portions](r)
	dep, ok := portions.Get(col)
	if !ok || dep.Get(ai.ResourceFood) < amount {
		return false
	}
	dep.Add(ai.ResourceFood, -amount)
	portions.Set(col, dep)
	provs := ecs.Get[Provisions](r)
	p, _ := provs.Get(entity)
	p.Rations += amount
	provs.Set(entity, p)
	w.Emit(ProvisionsTaken{Colony: col, Entity: entity, Amount: amount})
	return true
}

// returnRations devuelve al depósito de col todas las raciones sobrantes de
// entity en el mismo tick.
func returnRations(w *world.World, col, entity ids.ID) int32 {
	r := w.ECS()
	provs := ecs.Get[Provisions](r)
	p, ok := provs.Get(entity)
	if !ok || p.Rations <= 0 {
		return 0
	}
	amount := p.Rations
	p.Rations = 0
	provs.Remove(entity)
	portions := ecs.Get[ai.Portions](r)
	dep, _ := portions.Get(col)
	dep.Add(ai.ResourceFood, amount)
	portions.Set(col, dep)
	w.Emit(ProvisionsReturned{Colony: col, Entity: entity, Amount: amount})
	return amount
}

// EatRation es la acción "eat_ration": comer una ración personal durante una
// pausa segura para recuperar saciedad en camino.
type EatRation struct{}

func (EatRation) Name() string { return EatRationAction }

func (EatRation) Options(c *ai.Ctx) []ai.Option {
	r := c.W.ECS()
	p, ok := ecs.Get[Provisions](r).Get(c.Actor)
	if !ok || p.Rations <= 0 || c.Needs == nil {
		return nil
	}
	h, ok := c.Needs.Get("hunger")
	if !ok || h >= EatRationThreshold {
		return nil
	}
	x := float64(needs.Max-h) / needs.Max
	u := 0.45 + 0.35*x*x
	return []ai.Option{{Utility: u}}
}

func (EatRation) Utility(c *ai.Ctx, _ *ai.Doing) float64 {
	if c.Needs == nil {
		return 0
	}
	h, ok := c.Needs.Get("hunger")
	if !ok {
		return 0
	}
	x := float64(needs.Max-h) / needs.Max
	return 0.45 + 0.35*x*x
}

func (EatRation) Start(c *ai.Ctx, _ *ai.Doing) bool {
	p, ok := ecs.Get[Provisions](c.W.ECS()).Get(c.Actor)
	if !ok || p.Rations <= 0 {
		return false
	}
	nav.Halt(c.W, c.Actor)
	return true
}

func (EatRation) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	r := c.W.ECS()
	provs := ecs.Get[Provisions](r)
	p, ok := provs.Get(c.Actor)
	if !ok || p.Rations <= 0 {
		return ai.Failed
	}
	if d.Progress++; d.Progress < EatRationTicks {
		return ai.Running
	}
	p.Rations--
	if p.Rations == 0 {
		provs.Remove(c.Actor)
	} else {
		provs.Set(c.Actor, p)
	}
	if n := ecs.Get[needs.Needs](r).Ptr(c.Actor); n != nil {
		n.Add("hunger", EatRationGain)
	}
	c.W.Emit(RationEaten{Entity: c.Actor, Left: p.Rations})
	return ai.Done
}

func (EatRation) Stop(*ai.Ctx, *ai.Doing, bool) {}

func (EatRation) Valid(c *ai.Ctx, _ *ai.Doing) bool {
	p, ok := ecs.Get[Provisions](c.W.ECS()).Get(c.Actor)
	return ok && p.Rations > 0
}

func (EatRation) Interruptible(*ai.Ctx, *ai.Doing) bool { return true }

func checkProvisions(w *world.World) error {
	var err error
	ecs.Get[Provisions](w.ECS()).Each(func(e ids.ID, p *Provisions) {
		if p.Rations < 0 {
			err = errors.Join(err, fmt.Errorf("colony: %v tiene raciones negativas (%d)", e, p.Rations))
		}
	})
	return err
}

func registerProvisions(w *world.World) error {
	if err := register[Provisions](w.ECS(), "provisions"); err != nil {
		return err
	}
	w.AddInvariant("colony.provisions", checkProvisions)
	return errors.Join(
		world.RegisterEvent[ProvisionsTaken](w),
		world.RegisterEvent[RationEaten](w),
		world.RegisterEvent[ProvisionsReturned](w),
	)
}
