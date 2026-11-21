package ai

import (
	"cmp"
	"errors"
	"fmt"

	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Offer es una acción que ofrece un objeto (smart object): a qué escala de
// agente, qué efectos tiene y cómo se hace (Steps). Sin Steps, la oferta es
// "ir al punto de uso y usar durante Duration". Approach y Facing son el
// punto de uso y la orientación, relativos a la posición del objeto en su
// Space.
type Offer struct {
	Action   string           `json:"action"`
	Scale    string           `json:"scale"`
	Effects  []Effect         `json:"effects"`
	Duration clock.Tick       `json:"duration,omitempty"`
	Approach units.Vec2       `json:"approach,omitempty"`
	Facing   *fixedmath.Angle `json:"facing,omitempty"`
	Steps    []Step           `json:"steps,omitempty"`
}

// Offers es el componente "offers".
type Offers []Offer

// Pasos de una oferta (Step.Do).
const (
	StepGoto    = "goto"    // ir al punto de uso, o al lugar Place más cercano
	StepWait    = "wait"    // esperar Ticks
	StepUse     = "use"     // aplicar los efectos de a poco durante Ticks (uno solo por oferta)
	StepPose    = "pose"    // poner al actor en la pose Pose (se deshace al terminar)
	StepTake    = "take"    // crear Item y cargarlo
	StepDrop    = "drop"    // dejar lo cargado en Offset
	StepSpawn   = "spawn"   // crear Count copias de Item alrededor de Offset
	StepConsume = "consume" // gastar Amount porciones del objeto (sin porciones, todo; ver Keep)
)

// Step es un paso de una oferta. Offset es relativo a la mano del actor
// (Agent.Hand) o, si no tiene, al actor, y gira con él. Lo que se deja o
// se crea cae a la superficie que haya debajo.
type Step struct {
	Do     string       `json:"do"`
	Place  string       `json:"place,omitempty"`
	Ticks  clock.Tick   `json:"ticks,omitempty"`
	Pose   string       `json:"pose,omitempty"`
	Item   *cmds.Spawn  `json:"item,omitempty"`
	Count  int          `json:"count,omitempty"`
	Offset units.Vec2   `json:"offset,omitempty"`
	Spread units.Length `json:"spread,omitempty"`
	Amount int32        `json:"amount,omitempty"`
	Keep   bool         `json:"keep,omitempty"` // consume: en 0 porciones no desaparece (un depósito)
}

// Place marca un lugar donde se hacen cosas (componente "place"): un
// escritorio, una mesa. Spot y Facing son relativos a la posición de la
// entidad en su Space, como Offer.Approach.
type Place struct {
	Tags   []string         `json:"tags"`
	Spot   units.Vec2       `json:"spot,omitempty"`
	Facing *fixedmath.Angle `json:"facing,omitempty"`
}

// Portions es lo que queda de un objeto consumible (componente "portions").
type Portions struct {
	Left int32 `json:"left"`
}

// Eventos de las ofertas.
type (
	ItemDropped struct {
		Entity ids.ID `json:"entity"`
		Item   ids.ID `json:"item"`
		On     ids.ID `json:"on"` // Space donde quedó
	}
	Consumed struct {
		Entity ids.ID `json:"entity"`
		Target ids.ID `json:"target"`
		Left   int32  `json:"left"`
	}
)

func (ItemDropped) Kind() string { return "item_dropped" }
func (Consumed) Kind() string    { return "consumed" }

// plan devuelve los pasos de la oferta.
func (o Offer) plan() []Step {
	if len(o.Steps) > 0 {
		return o.Steps
	}
	return []Step{{Do: StepGoto}, {Do: StepUse, Ticks: o.Duration}}
}

// useStep es el índice del paso "use".
func (o Offer) useStep() int {
	for i, s := range o.plan() {
		if s.Do == StepUse {
			return i
		}
	}
	return -1
}

// Validate verifica que la oferta se pueda ejecutar.
func (o Offer) Validate() error {
	if o.Action == "" || o.Scale == "" {
		return errors.New("oferta sin action o scale")
	}
	uses, carrying := 0, false
	var err error
	bad := func(i int, msg string) { err = errors.Join(err, fmt.Errorf("%s: paso %d: %s", o.Action, i, msg)) }
	for i, s := range o.plan() {
		switch s.Do {
		case StepGoto, StepConsume:
		case StepWait, StepUse:
			if s.Ticks <= 0 {
				bad(i, "ticks debe ser positivo")
			}
			if s.Do == StepUse {
				uses++
			}
		case StepPose:
			if s.Pose == "" {
				bad(i, "falta pose")
			}
		case StepTake, StepSpawn:
			if s.Item == nil {
				bad(i, "falta item")
			} else if _, ok := s.Item.Components["agent"]; ok {
				bad(i, "un item no puede ser agente")
			}
			if s.Do == StepTake {
				if carrying {
					bad(i, "ya carga algo")
				}
				carrying = true
			}
		case StepDrop:
			if !carrying {
				bad(i, "no carga nada")
			}
			carrying = false
		default:
			bad(i, fmt.Sprintf("paso desconocido %q", s.Do))
		}
	}
	if uses != 1 {
		err = errors.Join(err, fmt.Errorf("%s: debe tener exactamente un paso use", o.Action))
	}
	return err
}

// Validate verifica todas las ofertas.
func (os Offers) Validate() error {
	var err error
	for _, o := range os {
		err = errors.Join(err, o.Validate())
	}
	return err
}

// --- use: usar una oferta de un smart object ---

type use struct{}

func (use) Name() string { return "use" }

func offerOf(c *Ctx, d *Doing) (Offer, comp.Location, bool) {
	r := c.W.ECS()
	os, ok := ecs.Get[Offers](r).Get(d.Target)
	loc, ok2 := comp.Of(r).Location.Get(d.Target)
	if !ok || !ok2 || d.Offer < 0 || d.Offer >= len(os) || r.Despawning(d.Target) {
		return Offer{}, loc, false
	}
	return os[d.Offer], loc, true
}

// What es el nombre de la oferta.
func (use) What(c *Ctx, d *Doing) string {
	if os, ok := ecs.Get[Offers](c.W.ECS()).Get(d.Target); ok && d.Offer >= 0 && d.Offer < len(os) {
		return os[d.Offer].Action
	}
	return ""
}

func spot(loc comp.Location, o Offer) units.Vec2 { return loc.Pos.XY().Add(o.Approach) }

// nearestPlace es el lugar con la etiqueta tag más cercano al actor
// (desempate por ID).
func nearestPlace(c *Ctx, tag string) (ids.ID, bool) {
	best, bestD, found := ids.Nil, 0.0, false
	locs := comp.Of(c.W.ECS()).Location
	ecs.Get[Place](c.W.ECS()).Each(func(e ids.ID, p *Place) {
		loc, ok := locs.Get(e)
		if !ok || c.W.ECS().Despawning(e) || !hasTag(p.Tags, tag) {
			return
		}
		d := distance(c, loc.Space, loc.Pos.XY().Add(p.Spot))
		if !found || d < bestD || d == bestD && cmp.Less(e, best) {
			best, bestD, found = e, d, true
		}
	})
	return best, found
}

func hasTag(tags []string, t string) bool {
	for _, x := range tags {
		if x == t {
			return true
		}
	}
	return false
}

func (use) Options(c *Ctx) []Option {
	if c.Needs == nil {
		return nil
	}
	var out []Option
	locs := comp.Of(c.W.ECS()).Location
	ecs.Get[Offers](c.W.ECS()).Each(func(e ids.ID, os *Offers) {
		loc, ok := locs.Get(e)
		if !ok || c.W.ECS().Despawning(e) || e == c.Actor {
			return
		}
		if h := jobs.Holder(c.W, e); !h.IsNil() && h != c.Actor {
			return
		}
	offers:
		for i, o := range *os {
			if o.Scale != c.Agent.Scale || wasteful(c, o.Effects) || !c.Agent.Knows(e) || !stocked(c, e, o) || o.Validate() != nil {
				continue
			}
			for _, s := range o.plan() {
				if s.Place != "" {
					if _, ok := nearestPlace(c, s.Place); !ok {
						continue offers
					}
				}
			}
			u := gain(c, o.Effects, 1) - c.Agent.DistanceCost*distance(c, loc.Space, spot(loc, o))
			out = append(out, Option{Target: e, Offer: i, Utility: u})
		}
	})
	return out
}

// stocked: si la oferta consume porciones, el objeto tiene suficientes.
func stocked(c *Ctx, e ids.ID, o Offer) bool {
	p, ok := ecs.Get[Portions](c.W.ECS()).Get(e)
	if !ok {
		return true
	}
	for _, s := range o.plan() {
		if s.Do == StepConsume && p.Left < max(s.Amount, 1) {
			return false
		}
	}
	return true
}

// wasteful: todos los efectos se desperdiciarían en más de la mitad (no se
// come sin hambre aunque no haya nada mejor que hacer).
func wasteful(c *Ctx, effects []Effect) bool {
	for _, e := range effects {
		if v, ok := c.Needs.Get(e.Need); ok && v+e.Amount/2 <= needs.Max {
			return false
		}
	}
	return len(effects) > 0
}

// Utility: lo que falta del efecto. Después del paso use no queda nada (y
// la acción ya no se puede interrumpir).
func (use) Utility(c *Ctx, d *Doing) float64 {
	o, _, ok := offerOf(c, d)
	if !ok {
		return 0
	}
	switch u := o.useStep(); {
	case d.Stage < u:
		return gain(c, o.Effects, 1)
	case d.Stage == u:
		t := o.plan()[u].Ticks
		return gain(c, o.Effects, float64(t-d.Progress)/float64(t))
	}
	return 0
}

func (use) Start(c *Ctx, d *Doing) bool {
	_, _, ok := offerOf(c, d)
	return ok && jobs.Reserve(c.W, d.Target, c.Actor)
}

type stepResult uint8

const (
	stepNext stepResult = iota
	stepWait
	stepFail
)

// Step avanza los pasos: los instantáneos se encadenan en el mismo tick.
func (use) Step(c *Ctx, d *Doing) Result {
	o, loc, ok := offerOf(c, d)
	if !ok || !jobs.Reserve(c.W, d.Target, c.Actor) {
		return Failed
	}
	if !d.Item.IsNil() && !jobs.Reserve(c.W, d.Item, c.Actor) {
		return Failed
	}
	steps := o.plan()
	for d.Stage < len(steps) {
		switch doStep(c, d, o, loc, steps[d.Stage]) {
		case stepWait:
			return Running
		case stepFail:
			return Failed
		}
		d.Stage++
		d.Progress, d.Moving, d.Place = 0, false, ids.Nil
	}
	return Done
}

func doStep(c *Ctx, d *Doing, o Offer, loc comp.Location, s Step) stepResult {
	w := c.W
	switch s.Do {
	case StepGoto:
		sp, goal, facing := loc.Space, spot(loc, o), o.Facing
		if s.Place != "" {
			if d.Place.IsNil() || !w.ECS().Alive(d.Place) {
				p, ok := nearestPlace(c, s.Place)
				if !ok {
					return stepFail
				}
				d.Place = p
			}
			pl, _ := ecs.Get[Place](w.ECS()).Get(d.Place)
			ploc, _ := comp.Of(w.ECS()).Location.Get(d.Place)
			sp, goal, facing = ploc.Space, ploc.Pos.XY().Add(pl.Spot), pl.Facing
		}
		switch nav.Approach(w, c.Actor, sp, goal, &d.Moving) {
		case nav.Failed:
			return stepFail
		case nav.Underway:
			return stepWait
		}
		if l := comp.Of(w.ECS()).Location.Ptr(c.Actor); facing != nil && l != nil {
			l.Facing = *facing
		}
	case StepWait:
		if d.Progress++; d.Progress < s.Ticks {
			return stepWait
		}
	case StepUse:
		// Los efectos se aplican de a poco: al terminar suman exactamente Amount.
		d.Progress++
		if n := ecs.Get[needs.Needs](w.ECS()).Ptr(c.Actor); n != nil {
			p, dur := int64(d.Progress), int64(s.Ticks)
			for _, e := range o.Effects {
				n.Add(e.Need, int32(int64(e.Amount)*p/dur-int64(e.Amount)*(p-1)/dur))
			}
		}
		if d.Progress < s.Ticks {
			return stepWait
		}
	case StepPose:
		if err := (attach.SetPose{Entity: c.Actor, Pose: s.Pose}).Apply(w); err != nil {
			return stepFail
		}
		d.Posed = s.Pose != c.Agent.IdlePose
	case StepTake:
		l, _ := comp.Of(w.ECS()).Location.Get(c.Actor)
		item, ok := place(c, *s.Item, l)
		if !ok || !jobs.Reserve(w, item, c.Actor) {
			return stepFail
		}
		ecs.Get[jobs.Carried](w.ECS()).Set(item, jobs.Carried{By: c.Actor})
		d.Item = item
	case StepDrop:
		if !putDown(c, d, s.Offset, true) {
			return stepFail
		}
	case StepSpawn:
		n := max(s.Count, 1)
		_, facing, _ := space.Of(w.ECS()).WorldPose(c.Actor)
		for i := range n {
			off := s.Offset
			if s.Spread > 0 {
				ring := fixedmath.Rotate(units.V2(s.Spread, 0), fixedmath.Turn*fixedmath.Angle(i)/fixedmath.Angle(n))
				off = off.Add(ring)
			}
			at, ok := dropPoint(c, off, true, facing)
			if !ok {
				return stepFail
			}
			if _, ok := place(c, *s.Item, at); !ok {
				return stepFail
			}
		}
	case StepConsume:
		amount := max(s.Amount, 1)
		left := int32(0)
		if p := ecs.Get[Portions](w.ECS()).Ptr(d.Target); p != nil {
			p.Left = max(p.Left-amount, 0)
			left = p.Left
		}
		w.Emit(Consumed{Entity: c.Actor, Target: d.Target, Left: left})
		if left == 0 && !s.Keep {
			w.ECS().Despawn(d.Target)
		}
	}
	return stepNext
}

// dropPoint es la ubicación (en el Space del actor) de off, relativo a la
// mano si hand y el actor tiene una, o al actor; off gira con facing.
func dropPoint(c *Ctx, off units.Vec2, hand bool, facing fixedmath.Angle) (comp.Location, bool) {
	g := space.Of(c.W.ECS())
	l, ok := comp.Of(c.W.ECS()).Location.Get(c.Actor)
	if !ok {
		return l, false
	}
	base := c.Actor
	if h := c.Agent.Hand; hand && g.IsSpace(h) && !c.W.ECS().Despawning(h) {
		base = h
	}
	p, _, err := g.WorldPose(base)
	if err != nil {
		return l, false
	}
	r := fixedmath.Rotate(off, facing)
	local, err := g.FromWorld(l.Space, units.V3(p.X+r.X, p.Y+r.Y, p.Z))
	if err != nil {
		return l, false
	}
	return comp.Location{Space: l.Space, Pos: local}, true
}

// place crea item en at y lo hace caer a la superficie de abajo.
func place(c *Ctx, item cmds.Spawn, at comp.Location) (ids.ID, bool) {
	item.Location = &at
	id, err := item.Do(c.W)
	if err != nil {
		return ids.Nil, false
	}
	space.Land(c.W, id)
	return id, true
}

// putDown deja lo cargado en off (ver dropPoint) y libera su reserva.
func putDown(c *Ctx, d *Doing, off units.Vec2, hand bool) bool {
	item := d.Item
	d.Item = ids.Nil
	r := c.W.ECS()
	if item.IsNil() || !r.Alive(item) {
		return false
	}
	_, facing, _ := space.Of(r).WorldPose(c.Actor)
	at, ok := dropPoint(c, off, hand, facing)
	ecs.Get[jobs.Carried](r).Remove(item)
	jobs.Release(c.W, item, c.Actor)
	if l := comp.Of(r).Location.Ptr(item); ok && l != nil {
		*l = at
		space.Land(c.W, item)
	}
	on, _ := comp.Of(r).Location.Get(item)
	c.W.Emit(ItemDropped{Entity: c.Actor, Item: item, On: on.Space})
	return true
}

// Stop deshace lo que quedó a medias: detiene el movimiento, deja lo
// cargado a los pies del actor y vuelve a la pose de reposo.
func (use) Stop(c *Ctx, d *Doing, _ bool) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
	}
	if !d.Item.IsNil() {
		putDown(c, d, units.Vec2{}, false)
	}
	if d.Posed && c.Agent.IdlePose != "" {
		_ = (attach.SetPose{Entity: c.Actor, Pose: c.Agent.IdlePose}).Apply(c.W)
		d.Posed = false
	}
	jobs.Release(c.W, d.Target, c.Actor)
}

func (use) Valid(c *Ctx, d *Doing) bool {
	_, _, ok := offerOf(c, d)
	return ok && jobs.Reserve(c.W, d.Target, c.Actor) && (d.Item.IsNil() || jobs.Reserve(c.W, d.Item, c.Actor))
}

// Interruptible: hasta terminar el paso use. Lo que sigue (dejar cosas,
// limpiar) se completa siempre.
func (use) Interruptible(c *Ctx, d *Doing) bool {
	o, _, ok := offerOf(c, d)
	return !ok || d.Stage <= o.useStep()
}

func registerUse(w *world.World) error {
	r := w.ECS()
	for _, err := range []error{
		storeErr(ecs.Register[Offers](r, "offers")),
		storeErr(ecs.Register[Place](r, "place")),
		storeErr(ecs.Register[Portions](r, "portions")),
	} {
		if err != nil {
			return err
		}
	}
	return errors.Join(world.RegisterEvent[ItemDropped](w), world.RegisterEvent[Consumed](w))
}

func storeErr[T any](_ *ecs.Store[T], err error) error { return err }
