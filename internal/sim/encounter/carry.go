package encounter

import (
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
)

const (
	// CarryAction es el nombre de la acción "llevar a un pasajero".
	CarryAction = "carry_passenger"
	// CarryWait es el tiempo máximo que sostiene la mano apoyada para que
	// bajen los pasajeros.
	CarryWait clock.Tick = 15 * clock.TPS
	// CarryMinWait es la espera mínima antes de dar por terminado el descenso.
	CarryMinWait clock.Tick = 2 * clock.TPS
	// CarryCooldown evita repetir el traslado hacia la misma superficie enseguida.
	CarryCooldown clock.Tick = 3 * 60 * clock.TPS
)

// Eventos de transporte deliberado de pasajeros en la palma.
type (
	CarryStarted struct {
		Entity     ids.ID   `json:"entity"`
		Surface    ids.ID   `json:"surface"`
		Passengers []ids.ID `json:"passengers"`
	}
	PassengerDelivered struct {
		Entity     ids.ID   `json:"entity"`
		Surface    ids.ID   `json:"surface"`
		Passengers []ids.ID `json:"passengers"`
	}
)

func (CarryStarted) Kind() string       { return "carry_started" }
func (PassengerDelivered) Kind() string { return "passenger_delivered" }

// CarryPassenger lleva intencionalmente a los tinies que viajan en la palma
// hasta una superficie segura con lugar donde apoyarla (priorizando la del
// nido si está lejos, o una mesa distinta si ya está junto a ella).
type CarryPassenger struct{}

func (CarryPassenger) Name() string { return CarryAction }

func (CarryPassenger) What(*ai.Ctx, *ai.Doing) string { return CarryAction }

// handPassengers devuelve los agentes de otra escala vivos en la mano de c,
// ordenados por ID (D5).
func handPassengers(c *ai.Ctx) []ids.ID {
	if c.Agent.Hand.IsNil() {
		return nil
	}
	r := c.W.ECS()
	locs := comp.Of(r).Location
	var out []ids.ID
	ecs.Get[ai.Agent](r).Each(func(id ids.ID, a *ai.Agent) {
		if id == c.Actor || a.Scale == c.Agent.Scale || r.Despawning(id) {
			return
		}
		if l, ok := locs.Get(id); ok && l.Space == c.Agent.Hand {
			out = append(out, id)
		}
	})
	slices.Sort(out)
	return out
}

// homeSurfaces reúne las superficies donde viven los nidos de passengers.
func homeSurfaces(c *ai.Ctx, passengers []ids.ID) map[ids.ID]bool {
	r := c.W.ECS()
	members := ecs.Get[colony.Member](r)
	locs := comp.Of(r).Location
	out := map[ids.ID]bool{}
	for _, pid := range passengers {
		m, ok := members.Get(pid)
		if !ok || m.Colony.IsNil() || !r.Alive(m.Colony) || r.Despawning(m.Colony) {
			continue
		}
		if l, ok := locs.Get(m.Colony); ok && !l.Space.IsNil() {
			out[l.Space] = true
		}
	}
	return out
}

// posesForSurface elige las poses de suspensión y apoyo que alcanzan la altura
// de la superficie s (p. ej. hand_rest para el escritorio a 75 cm o
// sit_table_rest para la mesa del living a 65 cm).
func posesForSurface(c *ai.Ctx, enc *Encounter, s ids.ID) (string, string) {
	hover, rest := enc.HoverPose, enc.RestPose
	if hover == "" {
		hover = DefaultHoverPose
	}
	if rest == "" {
		rest = DefaultRestPose
	}
	r := c.W.ECS()
	poses, ok := ecs.Get[attach.Poses](r).Get(c.Actor)
	if !ok {
		return hover, rest
	}
	sloc, ok := comp.Of(r).Location.Get(s)
	if !ok {
		return hover, rest
	}
	var thickness units.Length
	if cr, ok := ecs.Get[crush.Crusher](r).Get(c.Agent.Hand); ok {
		thickness = cr.Thickness
	}
	matches := func(poseName string) bool {
		parts, ok := poses[poseName]
		if !ok {
			return false
		}
		for _, p := range parts {
			if p.Part == c.Agent.Hand {
				diff := (p.Pos.Z - thickness) - sloc.Pos.Z
				return diff >= -crush.PressBelow && diff <= crush.PressAbove
			}
		}
		return false
	}
	if matches(rest) {
		return hover, rest
	}
	if _, ok := poses["sit_table_hover"]; ok && matches("sit_table_rest") {
		return "sit_table_hover", "sit_table_rest"
	}
	return hover, rest
}

func carryTargets(c *ai.Ctx, enc *Encounter, passengers []ids.ID) []ai.Option {
	r := c.W.ECS()
	surfaces := ecs.Get[attach.Surface](r)
	places := ecs.Get[ai.Place](r)
	homes := homeSurfaces(c, passengers)

	idsList := slices.Clone(places.IDs())
	slices.Sort(idsList)
	var out []ai.Option
	for _, s := range idsList {
		if !surfaces.Has(s) || enc.Cooling(s, CarryAction, c.W.Tick()) {
			continue
		}
		sp, ok := spotOf(c.W, s)
		if !ok {
			continue
		}
		dist := distance(c, sp)
		u := enc.Carry - c.Agent.DistanceCost*dist
		if homes[s] {
			u += 0.08
		}
		if dist < 0.6 {
			u -= 0.12
		}
		if u > 0 {
			out = append(out, ai.Option{Target: s, Utility: u})
		}
	}
	return out
}

func (CarryPassenger) Options(c *ai.Ctx) []ai.Option {
	enc := Of(c)
	r := c.W.ECS()
	if enc == nil || enc.Carry <= 0 || c.Agent.Hand.IsNil() || !ecs.Get[attach.Surface](r).Has(c.Agent.Hand) {
		return nil
	}
	passengers := handPassengers(c)
	if len(passengers) == 0 {
		return nil
	}
	return carryTargets(c, enc, passengers)
}

func (CarryPassenger) Utility(c *ai.Ctx, d *ai.Doing) float64 {
	enc := Of(c)
	if enc == nil || enc.Carry <= 0 {
		return 0
	}
	if len(handPassengers(c)) == 0 && !d.Posed {
		return 0
	}
	return enc.Carry
}

func carryOffer(c *ai.Ctx, enc *Encounter, s ids.ID) (ai.Offer, comp.Location, bool) {
	r := c.W.ECS()
	pl, ok := ecs.Get[ai.Place](r).Get(s)
	loc, ok2 := comp.Of(r).Location.Get(s)
	if !ok || !ok2 {
		return ai.Offer{}, loc, false
	}
	hover, rest := posesForSurface(c, enc, s)
	return ai.Offer{
		Action: CarryAction, Scale: c.Agent.Scale, Approach: pl.Spot, Facing: pl.Facing,
		Steps: []ai.Step{
			{Do: ai.StepGoto},
			{Do: ai.StepPose, Pose: hover},
			{Do: ai.StepWait, Ticks: OfferAnnounce},
			{Do: ai.StepPose, Pose: rest},
			{Do: ai.StepWait, Ticks: CarryWait},
		},
	}, loc, true
}

func (CarryPassenger) Start(c *ai.Ctx, d *ai.Doing) bool {
	enc := Of(c)
	if enc == nil || enc.Carry <= 0 {
		return false
	}
	passengers := handPassengers(c)
	if len(passengers) == 0 {
		return false
	}
	if _, _, ok := carryOffer(c, enc, d.Target); !ok {
		return false
	}
	d.Items = passengers
	enc.Cool(d.Target, CarryAction, c.W.Tick()+CarryCooldown)
	c.W.Emit(CarryStarted{Entity: c.Actor, Surface: d.Target, Passengers: slices.Clone(passengers)})
	return true
}

func anyOnHand(c *ai.Ctx, list []ids.ID) bool {
	r := c.W.ECS()
	locs := comp.Of(r).Location
	for _, id := range list {
		if !r.Alive(id) || r.Despawning(id) {
			continue
		}
		if l, ok := locs.Get(id); ok && l.Space == c.Agent.Hand {
			return true
		}
	}
	return false
}

func (CarryPassenger) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	enc := Of(c)
	if enc == nil {
		return ai.Failed
	}
	o, loc, ok := carryOffer(c, enc, d.Target)
	if !ok {
		return ai.Failed
	}
	if d.Stage < offerHover {
		sp, _ := spotOf(c.W, d.Target)
		if r, arrived := goTo(c, d, sp); !arrived {
			return r
		}
		d.Items = boarded(c, d.Items)
		if !anyOnHand(c, d.Items) {
			return skip(c, d, "gone")
		}
		d.Stage, d.Moving = offerHover, false
	}
	if d.Stage < offerRest {
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
		return ai.Running
	}
	d.Items = boarded(c, d.Items)
	res := ai.RunSteps(c, d, o, loc)
	if res == ai.Running && d.Progress >= CarryMinWait && !anyOnHand(c, d.Items) {
		return ai.Done
	}
	return res
}

func (CarryPassenger) Stop(c *ai.Ctx, d *ai.Doing, _ bool) {
	if d.Stage > offerRest {
		r := c.W.ECS()
		locs := comp.Of(r).Location
		var delivered []ids.ID
		for _, id := range d.Items {
			if !r.Alive(id) || r.Despawning(id) {
				continue
			}
			if l, ok := locs.Get(id); ok && l.Space != c.Agent.Hand {
				delivered = append(delivered, id)
			}
		}
		slices.Sort(delivered)
		if len(delivered) > 0 {
			c.W.Emit(PassengerDelivered{Entity: c.Actor, Surface: d.Target, Passengers: delivered})
		}
	}
	ai.UndoSteps(c, d)
}

func (CarryPassenger) Valid(c *ai.Ctx, d *ai.Doing) bool {
	_, ok := spotOf(c.W, d.Target)
	return ok
}

func (CarryPassenger) Interruptible(*ai.Ctx, *ai.Doing) bool { return true }

func (CarryPassenger) Committed(_ *ai.Ctx, d *ai.Doing) bool { return d.Posed }
