package sound

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/view"
)

// Sonidos de la escena.
const (
	StepTatami  = "giantess_step_tatami"
	StepWood    = "giantess_step_wood"
	StepTile    = "giantess_step_tile"
	DoorFusuma  = "door_fusuma"
	DoorNoren   = "door_noren"
	FridgeOpen  = "fridge_open"
	FridgeClose = "fridge_close"
	CupClink    = "cup_clink"
	Bite        = "giantess_bite"
	Zabuton     = "zabuton"
	TVOn        = "tv_on"
	HandNear    = "hand_near"  // micro: la mano se acerca
	HandLand    = "hand_land"  // micro: la mano se apoya
	HandLift    = "hand_lift"  // micro: la mano se levanta
	HandTap     = "hand_tap"   // macro: la mano se apoya
	TinyStep    = "tiny_step_" // + 1, 2 o 3 según la paleta
	CrumbPop    = "crumb_pop"  // un tiny levanta una migaja
	CrumbTick   = "crumb_tick" // una migaja cae sobre una superficie
	NestStore   = "nest_store"
	TinyHop     = "tiny_hop"
	TinyFall    = "tiny_fall"
	TinyLand    = "tiny_land"
	CrushPlaf   = "crush_plaf" // micro: la palma o un objeto aplasta a un tiny
	CrushTum    = "crush_tum"  // micro: una pisada aplasta a un tiny
	CrushFar    = "crush_far"  // macro: un golpe sordo donde murió un tiny
	PassOpen    = "pass_open"
	PassClose   = "pass_close"
	Emote       = "emote_" // + el símbolo (view.Emote*)
)

// Cadencias de los sonidos que salen del estado (en ticks).
const (
	giantessStepTicks clock.Tick = 10
	tinyStepTicks     clock.Tick = 8
	biteTicks         clock.Tick = 30
)

// minor son los sonidos que se omiten a x4 y x16: los que se repiten. Quedan
// los que cuentan algo (la mano, las caídas, los descubrimientos).
var minor = map[string]bool{
	StepTatami: true, StepWood: true, StepTile: true, Bite: true,
	TinyStep + "1": true, TinyStep + "2": true, TinyStep + "3": true,
	CrumbPop: true, CrumbTick: true,
	Emote + view.EmoteExplore: true, Emote + view.EmoteSleep: true, Emote + view.EmoteNote: true,
}

// floors da el sonido de los pasos de la giantess según la habitación.
var floors = map[string]string{"bedroom": StepTatami, "living": StepWood, "kitchen": StepTile}

// tinySteps da la variante de pasitos según la paleta del tiny.
var tinySteps = map[string]string{"tiny": TinyStep + "1", "tiny_lavender": TinyStep + "2", "tiny_ochre": TinyStep + "3"}

// View es lo que el sonido necesita saber de la presentación.
type View struct {
	Micro bool // vista micro: suenan los tinies y la mano de cerca
	// Focus es la superficie de la vista micro. La mano es enorme: se oye
	// cuando se apoya en ella (o cuando ella es la palma), aunque su centro
	// quede fuera de cuadro.
	Focus ids.ID
	Fast  bool // x4 o x16: se omiten los sonidos menores
	// Locate dice dónde queda en pantalla un punto del mundo: el panorama
	// (-1..1) y si se ve. Lo que no se ve no suena.
	Locate func(p units.Vec3) (pan float64, ok bool)
	// Visual es el identificador visual de una entidad ("giantess", "palm").
	Visual func(id ids.ID) string
	// ShowsEmote dice si el símbolo de una entidad se dibuja (en micro, los
	// de todos los tinies; en macro, la giantess y lo elegido o seguido).
	ShowsEmote func(id ids.ID) bool
}

// Request es un sonido pedido por la escena.
type Request struct {
	ID     string
	Volume float64
	Pan    float64
}

// Scene traduce los eventos publicados y los cambios entre frames en
// sonidos, con perspectiva por escala: en macro suenan la giantess y la casa;
// en micro, los tinies, la mano de cerca y la señal de paso. Solo lee.
type Scene struct {
	last *view.Frame
	seq  uint64
}

// Observe pide al mezclador los sonidos del frame actual.
func (s *Scene) Observe(m *Mixer, cur *view.Frame, events []view.Published, v View) {
	for _, r := range s.Sounds(cur, events, v) {
		m.Play(r.ID, r.Volume, r.Pan)
	}
}

// Sounds devuelve los sonidos nuevos desde la última observación. La
// primera observación solo registra el estado.
func (s *Scene) Sounds(cur *view.Frame, events []view.Published, v View) []Request {
	prev := s.last
	var fresh []view.Published
	for _, e := range events {
		if e.Seq > s.seq {
			fresh = append(fresh, e)
		}
	}
	if n := len(events); n > 0 {
		s.seq = max(s.seq, events[n-1].Seq)
	}
	if prev != nil && cur.Tick == prev.Tick {
		return nil
	}
	s.last = cur
	if prev == nil {
		return nil
	}
	c := &collector{v: v}
	for _, e := range fresh {
		c.event(prev, cur, e.Event)
	}
	c.state(prev, cur)
	return c.out
}

type collector struct {
	v   View
	out []Request
}

func (c *collector) add(id string, at units.Vec3, volume float64) {
	if c.v.Fast && minor[id] {
		return
	}
	pan, ok := c.v.Locate(at)
	if !ok {
		return
	}
	c.out = append(c.out, Request{ID: id, Volume: volume, Pan: pan})
}

// addNear suena aunque el punto quede fuera de cuadro (centrado entonces).
func (c *collector) addNear(id string, at units.Vec3, volume float64) {
	if c.v.Fast && minor[id] {
		return
	}
	pan, _ := c.v.Locate(at)
	c.out = append(c.out, Request{ID: id, Volume: volume, Pan: pan})
}

// contactPoint es dónde se cruza un contacto: el extremo del portal dinámico
// entre las dos superficies, o la mano si no hay portal.
func contactPoint(a, b, h ids.ID, frames ...*view.Frame) units.Vec3 {
	for _, f := range frames {
		for _, p := range f.Portals {
			if p.Dynamic && (p.ASpace == a && p.BSpace == b || p.ASpace == b && p.BSpace == a) {
				return p.A
			}
		}
	}
	at, _ := pos(h, frames...)
	return at
}

// over dice si la mano está sobre la superficie enfocada.
func over(f *view.Frame, h, focus ids.ID) bool {
	if h == focus {
		return true
	}
	hs, ok1 := f.SpaceByID(h)
	fs, ok2 := f.SpaceByID(focus)
	if !ok1 || !ok2 || fs.Extent.Empty() {
		return false
	}
	local := units.V2(hs.Pose.Pos.X-fs.Pose.Pos.X, hs.Pose.Pos.Y-fs.Pose.Pos.Y)
	return fs.Extent.Contains(local)
}

// pos es la posición de mundo de una entidad en el frame más reciente que la
// tenga.
func pos(id ids.ID, frames ...*view.Frame) (units.Vec3, bool) {
	for _, f := range frames {
		if t, ok := f.ThingByID(id); ok {
			return t.Pose.Pos, true
		}
		if s, ok := f.SpaceByID(id); ok {
			return s.Pose.Pos, true
		}
	}
	return units.Vec3{}, false
}

// hand devuelve la mano de un par de superficies en contacto.
func hand(f *view.Frame, a, b ids.ID) (ids.ID, bool) {
	for _, id := range []ids.ID{a, b} {
		if s, ok := f.SpaceByID(id); ok && s.Kind == space.Articulated {
			return id, true
		}
	}
	return ids.Nil, false
}

// palmOf devuelve la mano de un agente que es un Space (la giantess).
func palmOf(f *view.Frame, owner ids.ID) (ids.ID, bool) {
	for _, s := range f.Spaces {
		if s.Parent == owner && s.Kind == space.Articulated {
			return s.ID, true
		}
	}
	return ids.Nil, false
}

// creature dice si una Thing es un personaje que camina (un tiny), no un
// objeto.
func creature(f *view.Frame, id ids.ID) bool {
	t, ok := f.ThingByID(id)
	return ok && (t.Needs != nil || t.Mover != nil)
}

func (c *collector) event(prev, cur *view.Frame, e any) {
	v := c.v
	switch ev := e.(type) {
	case attach.ContactBegan:
		if h, ok := hand(cur, ev.A, ev.B); ok {
			switch {
			case !v.Micro:
				p, _ := pos(h, cur)
				c.add(HandTap, p, 0.8)
			case v.Focus == ev.A || v.Focus == ev.B:
				c.addNear(HandLand, contactPoint(ev.A, ev.B, h, cur), 1)
			}
		}
	case attach.ContactEnded:
		if h, ok := hand(cur, ev.A, ev.B); ok && v.Micro && (v.Focus == ev.A || v.Focus == ev.B) {
			c.addNear(HandLift, contactPoint(ev.A, ev.B, h, prev, cur), 1)
		}
	case attach.PoseChanged:
		if (ev.Pose == "hand_hover" || ev.Pose == "set_table") && v.Micro {
			if h, ok := palmOf(cur, ev.Entity); ok && over(cur, h, v.Focus) {
				p, _ := pos(h, cur)
				c.addNear(HandNear, p, 0.9)
			}
		}
	case attach.PortalOpened:
		c.portal(PassOpen, ev.Portal, cur, prev)
	case attach.PortalClosed:
		c.portal(PassClose, ev.Portal, prev, cur)
	case space.EntityReparented:
		if ev.Cause != "portal" {
			return
		}
		p, ok := pos(ev.Entity, cur)
		if !ok {
			return
		}
		if _, isSpace := cur.SpaceByID(ev.Entity); isSpace && !v.Micro {
			door := DoorFusuma
			if v.Visual(ev.From) == "kitchen" || v.Visual(ev.To) == "kitchen" {
				door = DoorNoren
			}
			c.add(door, p, 1)
			if v.Visual(ev.From) == "kitchen" {
				c.add(FridgeClose, p, 0.6)
			}
		}
		if creature(cur, ev.Entity) && v.Micro && (v.Focus == ev.From || v.Focus == ev.To) {
			c.addNear(TinyHop, p, 0.9)
		}
	case nav.MoveArrived:
		s, ok := cur.SpaceByID(ev.Entity)
		if !ok || v.Micro {
			return
		}
		switch v.Visual(s.Parent) {
		case "kitchen":
			c.add(FridgeOpen, s.Pose.Pos, 1)
		case "living":
			c.add(Zabuton, s.Pose.Pos, 0.8)
			for _, t := range cur.Things {
				if v.Visual(t.ID) == "tv" {
					c.add(TVOn, t.Pose.Pos, 0.7)
				}
			}
		}
	case ai.ItemDropped:
		if p, ok := pos(ev.Item, cur); ok {
			if v.Micro && ev.On == v.Focus {
				c.addNear(CupClink, p, 1)
			} else {
				c.add(CupClink, p, 1)
			}
		}
	case space.EntityFell:
		p, ok := pos(ev.Entity, cur)
		if !ok || !v.Micro {
			return
		}
		switch {
		case creature(cur, ev.Entity):
			c.add(TinyFall, p, 1)
		case v.Visual(ev.Entity) == "crumb":
			c.add(CrumbTick, p, 0.7)
		}
	case crush.Crushed:
		at, ok := pos(ev.Remains, cur)
		if !ok {
			return
		}
		switch {
		case !v.Micro:
			c.add(CrushFar, at, 0.8)
		case ev.Source == crush.SourceFoot:
			c.addNear(CrushTum, at, 1)
		default:
			c.add(CrushPlaf, at, 1)
		}
	case colony.Stored:
		if p, ok := pos(ev.Colony, cur); ok && v.Micro {
			c.add(NestStore, p, 0.9)
		}
	}
}

// portal suena la señal de paso, que solo se dibuja en micro y en los
// extremos que tocan la superficie enfocada.
func (c *collector) portal(id string, portal ids.ID, frames ...*view.Frame) {
	if !c.v.Micro {
		return
	}
	for _, f := range frames {
		for _, p := range f.Portals {
			if p.ID != portal {
				continue
			}
			switch c.v.Focus {
			case p.ASpace:
				c.addNear(id, p.A, 0.8)
			case p.BSpace:
				c.addNear(id, p.B, 0.8)
			}
			return
		}
	}
}

// crossed dice si entre prev y cur pasó un múltiplo de every (desplazado).
func crossed(prev, cur, every, offset clock.Tick) bool {
	return (prev+offset)/every != (cur+offset)/every
}

func (c *collector) state(prev, cur *view.Frame) {
	v := c.v
	if !v.Micro {
		for _, s := range cur.Spaces {
			if v.Visual(s.ID) != "giantess" {
				continue
			}
			before, ok := prev.SpaceByID(s.ID)
			if ok && before.Pose.Pos != s.Pose.Pos && crossed(prev.Tick, cur.Tick, giantessStepTicks, 0) {
				if step, ok := floors[v.Visual(s.Parent)]; ok {
					c.add(step, s.Pose.Pos, 0.8)
				}
			}
			if s.Activity == "use:breakfast" && s.Step == "use" && crossed(prev.Tick, cur.Tick, biteTicks, 0) {
				c.add(Bite, s.Pose.Pos, 0.7)
			}
		}
	}
	if v.Micro {
		for i, t := range cur.Things {
			before, ok := prev.ThingByID(t.ID)
			if !ok {
				continue
			}
			if (t.Needs != nil || t.Mover != nil) && before.Pose.Pos != t.Pose.Pos && t.CarriedBy.IsNil() && v.ShowsEmote(t.ID) {
				if _, falling := cur.FallOf(t.ID); !falling && crossed(prev.Tick, cur.Tick, tinyStepTicks, clock.Tick(i)) {
					step, ok := tinySteps[v.Visual(t.ID)]
					if !ok {
						step = TinyStep + "1"
					}
					c.add(step, t.Pose.Pos, 0.4)
				}
			}
			if !t.CarriedBy.IsNil() && before.CarriedBy.IsNil() {
				c.add(CrumbPop, t.Pose.Pos, 0.8)
			}
		}
		for _, fl := range cur.Falls {
			land := fl.Tick + view.FallTicks - 1
			if prev.Tick < land && land <= cur.Tick && creature(cur, fl.Entity) {
				if p, ok := pos(fl.Entity, cur); ok {
					c.add(TinyLand, p, 1)
				}
			}
		}
	}
	emote := func(id ids.ID, at units.Vec3) {
		if !v.ShowsEmote(id) {
			return
		}
		if e := cur.Emote(id); e != "" && e != prev.Emote(id) {
			c.add(Emote+e, at, 0.8)
		}
	}
	for _, t := range cur.Things {
		emote(t.ID, t.Pose.Pos)
	}
	for _, s := range cur.Spaces {
		if s.Needs != nil {
			emote(s.ID, s.Pose.Pos)
		}
	}
}
