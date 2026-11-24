package sound

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/view"
)

// Bucles de ambiente. La casa no tiene música: suena por su ambiente.
const (
	AmbBedroom = "amb_bedroom" // gorriones por la ventana y un fūrin
	AmbKitchen = "amb_kitchen" // el zumbido de la heladera
	AmbLiving  = "amb_living"  // el tic del reloj
	AmbTV      = "amb_tv"      // el murmullo de la TV encendida
	AmbMicro   = "amb_micro"   // tono de sala grave: todo amplificado
)

// Ambients son todos los bucles, en orden estable.
var Ambients = []string{AmbBedroom, AmbKitchen, AmbLiving, AmbTV, AmbMicro}

// rooms da el bucle de cada habitación según su visual.
var rooms = map[string]string{
	"bedroom":      AmbBedroom,
	"kitchen":      AmbKitchen,
	"living":       AmbLiving,
	"apt_bedroom":  AmbBedroom,
	"apt_kitchen":  AmbKitchen,
	"apt_living":   AmbLiving,
	"apt_hallway":  AmbLiving,
	"apt_bathroom": AmbKitchen,
	"apt_genkan":   AmbLiving,
}

const (
	// FadeFrames es cuánto tarda el ambiente en pasar de silencio a pleno.
	FadeFrames = 48
	// microRoom es el nivel de la habitación detrás del tono de sala en micro.
	microRoom = 0.45
	// pausedLevel baja el ambiente durante la pausa.
	pausedLevel = 0.4
)

// Ambience lleva el nivel de cada bucle de ambiente y lo acerca, frame a
// frame, al que corresponde a lo que se ve: se funde al cambiar de habitación
// o de escala, sin cortes.
type Ambience struct {
	level map[string]float64
}

// AmbientTargets es el nivel (0..1) que corresponde a cada bucle. Suena la
// habitación de lo que se mira: en micro, la de la superficie enfocada, con
// el tono de sala por delante; en macro, la de lo seguido o, si no, la de la
// giantess. La TV encendida suma su murmullo al living.
func AmbientTargets(f *view.Frame, v View, follow ids.ID, paused bool) map[string]float64 {
	out := map[string]float64{}
	var giantess ids.ID
	for _, s := range f.Spaces {
		if v.Visual(s.ID) == "giantess" {
			giantess = s.ID
		}
	}
	subject := giantess
	switch {
	case v.Micro:
		subject = v.Focus
	case !follow.IsNil():
		subject = follow
	}
	level := 1.0
	if v.Micro {
		level = microRoom
		out[AmbMicro] = 1
	}
	room := roomOf(f, subject)
	if id, ok := rooms[v.Visual(room)]; ok {
		out[id] = level
	}
	if s, ok := f.SpaceByID(giantess); ok && (v.Visual(room) == "living" || v.Visual(room) == "apt_living") && watching(s) {
		out[AmbTV] = 0.8 * level
	}
	if paused {
		for id := range out {
			out[id] *= pausedLevel
		}
	}
	return out
}

// watching dice si la giantess está mirando la TV (encendida).
func watching(s *view.Space) bool {
	return s.Activity == "use:watch" && (s.Step == "" || s.Step == "use")
}

// roomOf sube por la jerarquía hasta el Space hijo de la raíz.
func roomOf(f *view.Frame, id ids.ID) ids.ID {
	if t, ok := f.ThingByID(id); ok {
		id = t.Space
	}
	for range 64 {
		s, ok := f.SpaceByID(id)
		if !ok {
			return ids.Nil
		}
		parent, ok := f.SpaceByID(s.Parent)
		if !ok {
			return ids.Nil
		}
		if parent.Parent.IsNil() {
			return s.ID
		}
		id = s.Parent
	}
	return ids.Nil
}

// Observe acerca cada bucle a su nivel y se lo pasa al mezclador.
func (a *Ambience) Observe(m *Mixer, f *view.Frame, v View, follow ids.ID, paused bool) {
	if a.level == nil {
		a.level = map[string]float64{}
	}
	target := AmbientTargets(f, v, follow, paused)
	const step = 1.0 / FadeFrames
	for _, id := range Ambients {
		cur, want := a.level[id], target[id]
		switch {
		case cur < want:
			cur = min(want, cur+step)
		case cur > want:
			cur = max(want, cur-step)
		}
		a.level[id] = cur
		m.Loop(id, cur)
	}
}
