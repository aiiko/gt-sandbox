package encounter

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/world"
)

const (
	// ShareBelow: comparte si en la superficie hay menos que esto de comida
	// suelta a la vista.
	ShareBelow = 2
	// ShareCount es cuántas porciones deja.
	ShareCount = 2
	// ShareCooldown es cuánto tarda en volver a compartir en la misma
	// superficie.
	ShareCooldown = 4 * 60 * 20
	// DefaultDropPose es la pose desde la que deja la comida (la mano
	// suspendida sobre la mesa), si Encounter.DropPose está vacío.
	DefaultDropPose = "set_table"
)

// Shared se emite al dejar comida para los tinies: qué dejó y dónde.
type Shared struct {
	Entity  ids.ID   `json:"entity"`
	Surface ids.ID   `json:"surface"`
	Items   []ids.ID `json:"items"`
}

func (Shared) Kind() string { return "shared" }

// LooseFood cuenta la comida a la vista en la superficie s: con "food" y sin
// cargar (una migaja que un tiny va a buscar sigue sobre la mesa).
func LooseFood(w *world.World, s ids.ID) int {
	r := w.ECS()
	locs := comp.Of(r).Location
	carried := ecs.Get[jobs.Carried](r)
	n := 0
	ecs.Get[colony.Food](r).Each(func(e ids.ID, _ *colony.Food) {
		if l, ok := locs.Get(e); ok && l.Space == s && !carried.Has(e) && !r.Despawning(e) {
			n++
		}
	})
	return n
}

// Share es la acción "share": dejar unas porciones donde vio tinies
// buscando comida y hay poca a la vista. Usa los pasos de oferta de ai (ir
// al lugar, suspender la mano, crear las porciones, esperar) y al terminar
// emite shared con los IDs de lo que dejó.
type Share struct{}

// Pasos de share: los índices importan para la comprobación al llegar.
const sharePose = 1

func (Share) Name() string { return "share" }

func (Share) What(*ai.Ctx, *ai.Doing) string { return "share" }

// wants: hay tinies buscando comida y poca comida suelta.
func wants(w *world.World, s Sighting) bool {
	return s.Foraging > 0 && LooseFood(w, s.Surface) < ShareBelow
}

func (Share) Options(c *ai.Ctx) []ai.Option {
	enc := Of(c)
	if enc == nil || enc.Portion == nil || c.Agent.Hand.IsNil() {
		return nil
	}
	return options(c, "share", enc.Share, func(_ *Encounter, s Sighting) bool { return wants(c.W, s) })
}

func (Share) Utility(c *ai.Ctx, _ *ai.Doing) float64 {
	if enc := Of(c); enc != nil {
		return enc.Share
	}
	return 0
}

// offer arma la oferta sintética de share sobre la superficie s.
func shareOffer(c *ai.Ctx, enc *Encounter, s ids.ID) (ai.Offer, comp.Location, bool) {
	r := c.W.ECS()
	pl, ok := ecs.Get[ai.Place](r).Get(s)
	loc, ok2 := comp.Of(r).Location.Get(s)
	if !ok || !ok2 || enc.Portion == nil {
		return ai.Offer{}, loc, false
	}
	pose := enc.DropPose
	if pose == "" {
		pose = DefaultDropPose
	}
	return ai.Offer{
		Action: "share", Scale: c.Agent.Scale, Approach: pl.Spot, Facing: pl.Facing,
		Steps: []ai.Step{
			{Do: ai.StepGoto},
			{Do: ai.StepPose, Pose: pose},
			{Do: ai.StepSpawn, Item: enc.Portion, Count: ShareCount, Offset: units.V2(100000, 0), Spread: 50000},
			{Do: ai.StepWait, Ticks: 12},
		},
	}, loc, true
}

func (Share) Start(c *ai.Ctx, d *ai.Doing) bool {
	enc := Of(c)
	if enc == nil {
		return false
	}
	if _, _, ok := shareOffer(c, enc, d.Target); !ok {
		return false
	}
	enc.Cool(d.Target, "share", c.W.Tick()+ShareCooldown)
	return true
}

func (Share) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	enc := Of(c)
	if enc == nil {
		return ai.Failed
	}
	o, loc, ok := shareOffer(c, enc, d.Target)
	if !ok {
		return ai.Failed
	}
	if d.Stage < sharePose {
		// el paso goto lo hace la acción, para comprobar al llegar
		sp, _ := spotOf(c.W, d.Target)
		if r, arrived := goTo(c, d, sp); !arrived {
			return r
		}
		s, ok := sighting(c, d)
		switch {
		case !ok || s.Foraging == 0:
			return skip(c, d, "gone")
		case LooseFood(c.W, d.Target) >= ShareBelow:
			return skip(c, d, "enough")
		}
		d.Stage, d.Moving = sharePose, false
	}
	res := ai.RunSteps(c, d, o, loc)
	if res == ai.Done {
		c.W.Emit(Shared{Entity: c.Actor, Surface: d.Target, Items: d.Items})
	}
	return res
}

func (Share) Stop(c *ai.Ctx, d *ai.Doing, _ bool) { ai.UndoSteps(c, d) }

func (Share) Valid(c *ai.Ctx, d *ai.Doing) bool {
	_, ok := spotOf(c.W, d.Target)
	return ok
}

// Interruptible hasta dejar la comida; Committed con la mano suspendida.
func (Share) Interruptible(_ *ai.Ctx, d *ai.Doing) bool { return d.Stage <= sharePose }

func (Share) Committed(_ *ai.Ctx, d *ai.Doing) bool { return d.Posed }
