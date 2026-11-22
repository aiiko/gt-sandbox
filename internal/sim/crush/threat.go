package crush

import (
	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// ThreatHeight es la altura bajo la cual un Space aplastador suspendido
// sobre una superficie la amenaza (su sombra se ve desde abajo).
const ThreatHeight = 20 * units.Centimeter

// NearMissRange es la distancia a la huella dentro de la cual quien
// sobrevive a un apoyo se lleva un susto (near_miss).
const NearMissRange = 2 * units.Centimeter

// Threat es un aplastador por bajar: una mano suspendida sobre una
// superficie o la próxima pisada de quien camina.
type Threat struct {
	By      ids.ID
	Source  string
	Force   int32
	Surface ids.ID
	// Channel es cómo se percibe: "shadow" (mano) o "vibration" (pisada).
	Channel string
	inside  func(units.Vec3) bool
}

// Covers indica si el punto del mundo p está en la huella o a menos de
// margin de ella (prueba p y ocho puntos a margin alrededor).
func (t Threat) Covers(p units.Vec3, margin units.Length) bool {
	if t.inside(p) {
		return true
	}
	if margin <= 0 {
		return false
	}
	for i := range 8 {
		off := fixedmath.Rotate(units.V2(margin, 0), fixedmath.Turn/8*fixedmath.Angle(i))
		if t.inside(p.Add(units.V3(off.X, off.Y, 0))) {
			return true
		}
	}
	return false
}

// Under devuelve, en orden de store, los frágiles apoyados en la superficie
// de la amenaza que quedan bajo su huella o a menos de margin.
func (t Threat) Under(w *world.World, margin units.Length) []ids.ID {
	r := w.ECS()
	g := space.Of(r)
	locs := comp.Of(r).Location
	var out []ids.ID
	ecs.Get[Fragile](r).Each(func(e ids.ID, _ *Fragile) {
		l, ok := locs.Get(e)
		if !ok || l.Space != t.Surface || r.Despawning(e) || g.IsSpace(e) {
			return
		}
		if p, _, err := g.WorldPose(e); err == nil && t.Covers(p, margin) {
			out = append(out, e)
		}
	})
	return out
}

// Threats devuelve las amenazas del estado actual, en orden estable. Es una
// función pura del estado (no hay caché que guardar ni reconstruir).
func Threats(w *world.World) []Threat {
	r := w.ECS()
	g := space.Of(r)
	var out []Threat
	ecs.Get[Crusher](r).Each(func(e ids.ID, c *Crusher) {
		if !g.IsSpace(e) || r.Despawning(e) {
			return
		}
		for _, im := range over(r, g, e, *c, PressAbove+1, ThreatHeight) {
			out = append(out, Threat{By: e, Source: c.Source, Force: c.Force, Surface: im.surface, Channel: "shadow", inside: im.inside})
		}
	})
	locs := comp.Of(r).Location
	ecs.Get[Stride](r).Each(func(e ids.ID, st *Stride) {
		l, ok := locs.Get(e)
		if !ok || r.Despawning(e) || st.Length <= 0 || st.Space != l.Space || st.Dir == (units.Vec2{}) {
			return
		}
		f := next(e, *st, l.Space, l.Pos.XY())
		out = append(out, Threat{By: e, Source: SourceFoot, Force: st.Force, Surface: l.Space, Channel: "vibration", inside: footprint(g, f)})
	})
	return out
}

// next predice la próxima pisada de quien camina en la dirección del último
// tick.
func next(e ids.ID, st Stride, sp ids.ID, pos units.Vec2) footfall {
	d := st.Dir
	n := int64(d.Len())
	ahead := int64(st.Length - st.Travel)
	at := pos.Add(units.V2(d.X.MulDiv(ahead, n), d.Y.MulDiv(ahead, n)))
	return footfall{e: e, s: st, space: sp, center: at.Add(sideOffset(d, n, st)), dir: d}
}

// sideOffset es el desplazamiento lateral del pie que sigue.
func sideOffset(d units.Vec2, n int64, st Stride) units.Vec2 {
	side := int64(st.Side)
	if st.Right {
		side = -side
	}
	return units.V2((-d.Y).MulDiv(side, n), d.X.MulDiv(side, n)) // izquierda = antihorario
}

// nearMisses emite near_miss para los frágiles de la superficie del impacto
// que siguen vivos y quedaron a menos de NearMissRange de la huella.
func nearMisses(w *world.World, g space.Graph, im impact) {
	r := w.ECS()
	locs := comp.Of(r).Location
	t := Threat{inside: im.inside}
	ecs.Get[Fragile](r).Each(func(e ids.ID, _ *Fragile) {
		l, ok := locs.Get(e)
		if !ok || l.Space != im.surface || e == im.by || r.Despawning(e) || g.IsSpace(e) {
			return
		}
		p, _, err := g.WorldPose(e)
		if err != nil || !t.Covers(p, NearMissRange) {
			return
		}
		w.Emit(NearMiss{Entity: e, By: im.by, Source: im.c.Source})
	})
}

// NearMiss se emite cuando alguien frágil sobrevive a un apoyo muy cerca de
// la huella (o dentro, a cubierto).
type NearMiss struct {
	Entity ids.ID `json:"entity"`
	By     ids.ID `json:"by"`
	Source string `json:"source"`
}

func (NearMiss) Kind() string { return "near_miss" }
