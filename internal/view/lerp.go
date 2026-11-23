package view

import (
	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/units"
)

// MaxLerp es la distancia a partir de la cual un cambio de posición entre
// dos ticks se toma como salto (se dibuja directo, sin interpolar).
const MaxLerp = units.Meter

// Interpolate devuelve una copia de cur con las poses de Spaces y Things
// interpoladas desde prev (alpha 0 = prev, 1 = cur). Lo demás (rutas,
// portales, grillas) es el de cur. Solo para dibujar.
func Interpolate(prev, cur *Frame, alpha float64) *Frame {
	if prev == nil || prev == cur || alpha >= 1 {
		return cur
	}
	alpha = max(alpha, 0)
	out := *cur
	out.Spaces = append([]Space(nil), cur.Spaces...)
	out.Things = append([]Thing(nil), cur.Things...)
	for i := range out.Spaces {
		if p, ok := prev.SpaceByID(out.Spaces[i].ID); ok {
			out.Spaces[i].Pose = lerpPose(p.Pose, out.Spaces[i].Pose, alpha)
		}
	}
	for i := range out.Things {
		if p, ok := prev.ThingByID(out.Things[i].ID); ok {
			t := &out.Things[i]
			d := t.Pose.Pos.Sub(p.Pose.Pos)
			t.Pose = lerpPose(p.Pose, t.Pose, alpha)
			if p.Space == t.Space {
				dl := t.LocalPos.Sub(p.LocalPos)
				if dl.LengthSq() <= int64(MaxLerp)*int64(MaxLerp) {
					l := func(x units.Length) units.Length { return units.Length(float64(x) * alpha) }
					t.LocalPos = p.LocalPos.Add(units.V3(l(dl.X), l(dl.Y), l(dl.Z)))
				}
			}
			if t.Mover != nil && len(t.Mover.Route) > 0 && d.LengthSq() <= int64(MaxLerp)*int64(MaxLerp) {
				m := *t.Mover
				m.Route = append([]units.Vec3{t.Pose.Pos}, m.Route[1:]...)
				t.Mover = &m
			}
		}
	}
	return &out
}

func lerpPose(a, b Pose, t float64) Pose {
	d := b.Pos.Sub(a.Pos)
	if d.LengthSq() > int64(MaxLerp)*int64(MaxLerp) {
		return b
	}
	l := func(x units.Length) units.Length { return units.Length(float64(x) * t) }
	da := b.Rot.Sub(a.Rot)
	if da > fixedmath.HalfTurn { // girar por el camino corto
		da -= fixedmath.Turn
	}
	return Pose{
		Pos: a.Pos.Add(units.V3(l(d.X), l(d.Y), l(d.Z))),
		Rot: a.Rot.Add(fixedmath.Angle(float64(da) * t)),
	}
}
