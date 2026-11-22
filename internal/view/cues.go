package view

import (
	"strings"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Símbolos de expresión (propuesta, «Expresiones y señales»). Cada uno tiene
// una forma distinta, así que se leen también sin color.
const (
	EmoteDiscover = "discover" // «!»: descubre algo
	EmoteExplore  = "explore"  // «?»: explora
	EmoteSweat    = "sweat"    // gota: lo levantan o cae
	EmoteSleep    = "sleep"    // «zzz»: descansa
	EmoteNote     = "note"     // nota musical: se divierte
	EmoteVein     = "vein"     // vena: le interrumpen una acción
	EmoteSpirit   = "spirit"   // espíritu que sube: un tiny murió aplastado (sobre sus restos)
	EmoteLook     = "look"     // ojo: la giantess se queda mirando a los tinies
	cueSurprise   = "surprise" // la giantess nota algo en su mano (solo el retrato)
)

// Expresiones del retrato de la giantess.
const (
	FaceNeutral   = "neutral"
	FaceHappy     = "happy"
	FaceHungry    = "hungry"
	FaceSurprised = "surprised"
	FaceSleepy    = "sleepy"
	FaceCurious   = "curious"
)

// CueTicks es cuánto dura en pantalla una señal que sale de un evento.
const CueTicks clock.Tick = 40

// LowNeed es el valor por debajo del cual una necesidad se nota en la cara.
const LowNeed = 300

// Cue es una señal de presentación que dejó un evento: el símbolo Kind sobre
// Entity desde Tick (el primer frame que ya refleja el evento).
type Cue struct {
	Entity ids.ID     `json:"entity"`
	Kind   string     `json:"kind"`
	Tick   clock.Tick `json:"tick"`
}

// cuesOf traduce un evento publicado a señales. Solo lee el registro.
func cuesOf(r *ecs.Registry, event world.Event) []Cue {
	agents := ecs.Get[ai.Agent](r)
	handOwner := func(id ids.ID) ids.ID {
		var owner ids.ID
		agents.Each(func(e ids.ID, a *ai.Agent) {
			if a.Hand == id && owner.IsNil() {
				owner = e
			}
		})
		return owner
	}
	switch ev := event.(type) {
	case ai.Noticed:
		if a, ok := agents.Get(ev.Entity); ok && a.Scale != "giantess" && ecs.Get[colony.Food](r).Has(ev.Target) {
			return []Cue{{Entity: ev.Entity, Kind: EmoteDiscover}}
		}
	case space.EntityReparented:
		if ev.Cause == "portal" {
			if owner := handOwner(ev.To); !owner.IsNil() {
				return []Cue{{Entity: ev.Entity, Kind: EmoteSweat}, {Entity: owner, Kind: cueSurprise}}
			}
		}
	case ai.ActionInterrupted:
		// huir no enoja: la gota de la alarma queda
		if a, ok := agents.Get(ev.Entity); ev.Action == "use" && !(ok && a.Doing != nil && a.Doing.Action == ai.FleeAction) {
			return []Cue{{Entity: ev.Entity, Kind: EmoteVein}}
		}
	case ai.Alarmed:
		return []Cue{{Entity: ev.Entity, Kind: EmoteSweat}}
	case ai.Trapped:
		return []Cue{{Entity: ev.Entity, Kind: EmoteSweat}}
	case crush.NearMiss:
		return []Cue{{Entity: ev.Entity, Kind: EmoteDiscover}}
	case ai.Hesitated:
		return []Cue{{Entity: ev.Entity, Kind: EmoteExplore}}
	case crush.Crushed:
		return []Cue{{Entity: ev.Remains, Kind: EmoteSpirit}}
	}
	if n, ok := event.(ai.Noticed); ok && ecs.Get[crush.Remains](r).Has(n.Target) {
		if a, ok := agents.Get(n.Entity); ok && a.Scale == "giantess" {
			return []Cue{{Entity: n.Entity, Kind: cueSurprise}}
		}
	}
	return nil
}

// Fresh indica si los restos de id son recientes: el espíritu todavía sube
// y el tiny se dibuja aplanado, con sus colores.
func (f *Frame) Fresh(id ids.ID) bool {
	_, ok := f.cue(id, func(k string) bool { return k == EmoteSpirit })
	return ok
}

// cue devuelve la señal más reciente de una entidad, si sigue vigente.
func (f *Frame) cue(id ids.ID, keep func(string) bool) (Cue, bool) {
	var best Cue
	found := false
	for _, c := range f.Cues {
		if c.Entity == id && keep(c.Kind) && f.Tick-c.Tick < CueTicks && (!found || c.Tick >= best.Tick) {
			best, found = c, true
		}
	}
	return best, found
}

// Emote devuelve el símbolo que se dibuja sobre una entidad, o "". Huir o
// estar atrapado da la gota antes que nada; si no, una señal reciente de un
// evento tiene prioridad; si no hay, sale del estado: caer o viajar en una
// mano levantada (gota), explorar («?»), descansar («zzz»), divertirse (nota).
func (f *Frame) Emote(id ids.ID) string {
	// el susto manda: mientras huye o está atrapado, la gota
	if t, ok := f.ThingByID(id); ok && (t.Trapped || t.Activity == ai.FleeAction) {
		return EmoteSweat
	}
	if c, ok := f.cue(id, func(k string) bool { return k != cueSurprise }); ok {
		return c.Kind
	}
	var activity, step string
	if t, ok := f.ThingByID(id); ok {
		activity, step = t.Activity, t.Step
		if _, ok := f.FallOf(id); ok && t.Needs != nil {
			return EmoteSweat
		}
		// un agente (tiene necesidades) sobre una mano que no está apoyada
		if s, ok := f.SpaceByID(t.Space); ok && t.Needs != nil && s.Kind == space.Articulated && !f.touching(s.ID) {
			return EmoteSweat
		}
	} else if s, ok := f.SpaceByID(id); ok {
		activity, step = s.Activity, s.Step
	} else {
		return ""
	}
	using := step == "" || step == "use"
	switch {
	case activity == "explore" || strings.HasPrefix(activity, "explore "):
		return EmoteExplore
	case (activity == "use:shelter" || activity == "use:sleep") && using:
		return EmoteSleep
	case activity == "use:watch" && using:
		return EmoteNote
	case activity == "observe" && step != "goto":
		return EmoteLook
	}
	return ""
}

// Expression devuelve la cara del retrato de un agente: sorprendida si algo
// acaba de subir a su mano, contenta mientras come o se divierte, curiosa en
// un encuentro con los tinies, somnolienta o hambrienta según sus
// necesidades, y neutral si no.
func (f *Frame) Expression(id ids.ID) string {
	s, ok := f.SpaceByID(id)
	if !ok {
		return FaceNeutral
	}
	if _, ok := f.cue(id, func(k string) bool { return k == cueSurprise }); ok {
		return FaceSurprised
	}
	using := s.Step == "" || s.Step == "use"
	switch s.Activity {
	case "use:breakfast", "use:snack", "use:snack_living", "use:eat", "use:watch":
		if using {
			return FaceHappy
		}
	case "observe", "share", "offer_palm":
		if s.Step != "goto" {
			return FaceCurious
		}
	}
	if v, ok := NeedOf(s.Needs, "energy"); ok && v < LowNeed {
		return FaceSleepy
	}
	if v, ok := NeedOf(s.Needs, "hunger"); ok && v < LowNeed {
		return FaceHungry
	}
	return FaceNeutral
}

// touching indica si un Space está en algún contacto.
func (f *Frame) touching(id ids.ID) bool {
	for _, c := range f.Contacts {
		if c.A == id || c.B == id {
			return true
		}
	}
	return false
}
