package view

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
)

// SurfaceAt devuelve el Space más alto cuya extensión contiene el punto de
// mundo p (en el plano), y p en coordenadas locales de ese Space.
func (f *Frame) SurfaceAt(p units.Vec2) (ids.ID, units.Vec2, bool) {
	for i := len(f.Spaces) - 1; i >= 0; i-- {
		s := &f.Spaces[i]
		if s.Extent.Empty() {
			continue
		}
		l := s.Pose.Transform().ApplyInv(units.V3(p.X, p.Y, s.Pose.Pos.Z)).XY()
		if s.Extent.Contains(l) {
			return s.ID, l, true
		}
	}
	return ids.Nil, units.Vec2{}, false
}

// Pick elige qué hay en el punto de mundo p: la Thing más cercana a menos de
// max(radio, tol); si no hay, un Space sin extensión (un marco, como el cuerpo
// de la giantess) cuyo origen esté a menos de tol; si no, la superficie más
// alta que contiene p.
func (f *Frame) Pick(p units.Vec2, tol units.Length) ids.ID {
	best, bestD := ids.Nil, int64(-1)
	try := func(id ids.ID, at units.Vec3, r units.Length) {
		d := at.XY().DistSq(p)
		if r = max(r, tol); d <= int64(r)*int64(r) && (bestD < 0 || d < bestD) {
			best, bestD = id, d
		}
	}
	for _, t := range f.Things {
		try(t.ID, t.Pose.Pos, t.Radius)
	}
	if best.IsNil() {
		for _, s := range f.Spaces {
			if s.Extent.Empty() {
				try(s.ID, s.Pose.Pos, 0)
			}
		}
	}
	if best.IsNil() {
		best, _, _ = f.SurfaceAt(p)
	}
	return best
}
