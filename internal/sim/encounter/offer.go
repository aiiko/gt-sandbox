package encounter

import (
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
)

const (
	// OfferAnnounce es cuánto sostiene la mano en el aire antes de apoyarla:
	// la sombra avisa y los tinies con reflejos se apartan.
	OfferAnnounce = 3 * 20
	// OfferWait es cuánto deja la mano apoyada, con el portal abierto.
	OfferWait = 20 * 20
	// OfferCooldown es cuánto tarda en volver a ofrecerla en la misma
	// superficie.
	OfferCooldown = 5 * 60 * 20
	// Poses por defecto de la oferta (ver Encounter.HoverPose y RestPose).
	DefaultHoverPose = "hand_hover"
	DefaultRestPose  = "hand_rest"
)

// Pasos de la oferta: ir (lo hace la acción), suspender la mano, esperar,
// apoyarla y esperar con la mano apoyada.
const (
	offerHover = 1
	offerRest  = 3
)

// Eventos de la palma ofrecida.
type (
	PalmOffered struct {
		Entity  ids.ID `json:"entity"`
		Surface ids.ID `json:"surface"`
	}
	// PalmWithdrawn se emite al retirar la mano: Boarded son los tinies que
	// subieron mientras estuvo apoyada.
	PalmWithdrawn struct {
		Entity  ids.ID   `json:"entity"`
		Surface ids.ID   `json:"surface"`
		Boarded []ids.ID `json:"boarded"`
	}
)

func (PalmOffered) Kind() string   { return "palm_offered" }
func (PalmWithdrawn) Kind() string { return "palm_withdrawn" }

// OfferPalm es la acción "offer_palm": apoyar la mano abierta junto a los
// tinies de una superficie y esperar con el portal de contacto abierto, por
// si alguno quiere subir. Primero la sostiene en el aire (la sombra anuncia
// el apoyo y los tinies con reflejos se apartan); si al ir a apoyarla hay
// frágiles bajo la huella, cancela. Los que suben quedan en doing.items.
type OfferPalm struct{}

func (OfferPalm) Name() string { return "offer_palm" }

func (OfferPalm) What(*ai.Ctx, *ai.Doing) string { return "offer_palm" }

func (OfferPalm) Options(c *ai.Ctx) []ai.Option {
	enc := Of(c)
	r := c.W.ECS()
	if enc == nil || c.Agent.Hand.IsNil() || !ecs.Get[attach.Surface](r).Has(c.Agent.Hand) {
		return nil
	}
	return options(c, "offer_palm", enc.Offer, func(_ *Encounter, s Sighting) bool {
		return ecs.Get[attach.Surface](r).Has(s.Surface)
	})
}

func (OfferPalm) Utility(c *ai.Ctx, _ *ai.Doing) float64 {
	if enc := Of(c); enc != nil {
		return enc.Offer
	}
	return 0
}

func palmOffer(c *ai.Ctx, enc *Encounter, s ids.ID) (ai.Offer, comp.Location, bool) {
	r := c.W.ECS()
	pl, ok := ecs.Get[ai.Place](r).Get(s)
	loc, ok2 := comp.Of(r).Location.Get(s)
	if !ok || !ok2 {
		return ai.Offer{}, loc, false
	}
	hover, rest := enc.HoverPose, enc.RestPose
	if hover == "" {
		hover = DefaultHoverPose
	}
	if rest == "" {
		rest = DefaultRestPose
	}
	return ai.Offer{
		Action: "offer_palm", Scale: c.Agent.Scale, Approach: pl.Spot, Facing: pl.Facing,
		Steps: []ai.Step{
			{Do: ai.StepGoto},
			{Do: ai.StepPose, Pose: hover},
			{Do: ai.StepWait, Ticks: OfferAnnounce},
			{Do: ai.StepPose, Pose: rest},
			{Do: ai.StepWait, Ticks: OfferWait},
		},
	}, loc, true
}

func (OfferPalm) Start(c *ai.Ctx, d *ai.Doing) bool {
	enc := Of(c)
	if enc == nil {
		return false
	}
	if _, _, ok := palmOffer(c, enc, d.Target); !ok {
		return false
	}
	enc.Cool(d.Target, "offer_palm", c.W.Tick()+OfferCooldown)
	return true
}

// occupied indica si hay frágiles bajo la huella de la mano suspendida sobre
// la superficie s.
func occupied(c *ai.Ctx, s ids.ID) bool {
	for _, t := range crush.Threats(c.W) {
		if t.By == c.Agent.Hand && t.Surface == s && len(t.Under(c.W, 0)) > 0 {
			return true
		}
	}
	return false
}

func (OfferPalm) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	enc := Of(c)
	if enc == nil {
		return ai.Failed
	}
	o, loc, ok := palmOffer(c, enc, d.Target)
	if !ok {
		return ai.Failed
	}
	if d.Stage < offerHover {
		sp, _ := spotOf(c.W, d.Target)
		if r, arrived := goTo(c, d, sp); !arrived {
			return r
		}
		if _, ok := sighting(c, d); !ok {
			return skip(c, d, "gone")
		}
		d.Stage, d.Moving = offerHover, false
	}
	if d.Stage < offerRest {
		// hasta el anuncio inclusive
		head := o
		head.Steps = o.Steps[:offerRest]
		if res := ai.RunSteps(c, d, head, loc); res != ai.Done {
			return res
		}
	}
	if d.Stage == offerRest {
		if occupied(c, d.Target) {
			return skip(c, d, "occupied")
		}
		if res := ai.RunSteps(c, d, o, loc); d.Stage == offerRest {
			return res // todavía no la apoyó (cuidado)
		}
		c.W.Emit(PalmOffered{Entity: c.Actor, Surface: d.Target})
		return ai.Running
	}
	d.Items = boarded(c, d.Items)
	return ai.RunSteps(c, d, o, loc)
}

// boarded suma a list los agentes de otra escala que están en la mano.
func boarded(c *ai.Ctx, list []ids.ID) []ids.ID {
	r := c.W.ECS()
	locs := comp.Of(r).Location
	ecs.Get[ai.Agent](r).Each(func(e ids.ID, a *ai.Agent) {
		if a.Scale == c.Agent.Scale || r.Despawning(e) || slices.Contains(list, e) {
			return
		}
		if l, ok := locs.Get(e); ok && l.Space == c.Agent.Hand {
			list = append(list, e)
		}
	})
	return list
}

// Stop retira la mano; si llegó a apoyarla, emite palm_withdrawn.
func (OfferPalm) Stop(c *ai.Ctx, d *ai.Doing, _ bool) {
	if d.Stage > offerRest {
		c.W.Emit(PalmWithdrawn{Entity: c.Actor, Surface: d.Target, Boarded: append([]ids.ID{}, boarded(c, d.Items)...)})
	}
	ai.UndoSteps(c, d)
}

func (OfferPalm) Valid(c *ai.Ctx, d *ai.Doing) bool {
	_, ok := spotOf(c.W, d.Target)
	return ok
}

func (OfferPalm) Interruptible(*ai.Ctx, *ai.Doing) bool { return true }

// Committed: con la mano suspendida o apoyada, solo algo urgente la
// interrumpe.
func (OfferPalm) Committed(_ *ai.Ctx, d *ai.Doing) bool { return d.Posed }
