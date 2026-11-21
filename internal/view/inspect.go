package view

import (
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/world"
)

// Inspect describe una entidad para el panel del visor: una línea por
// componente, con el JSON del snapshot. Solo lee.
func Inspect(w *world.World, id ids.ID) []string {
	if !w.ECS().Alive(id) {
		return []string{fmt.Sprintf("%v (muerta)", id)}
	}
	out := []string{fmt.Sprintf("entidad %v", id)}
	for _, c := range w.ECS().Components(id) {
		out = append(out, c.Store+": "+string(c.JSON))
	}
	return out
}

// Bounds es la caja de mundo (en el plano) que cubre las extensiones de
// todos los Spaces. Vacía si no hay ninguna.
func (f *Frame) Bounds() units.Rect {
	var b units.Rect
	first := true
	for _, s := range f.Spaces {
		if s.Extent.Empty() {
			continue
		}
		for _, c := range Corners(s) {
			if first {
				b, first = units.Rect{Min: c.XY(), Max: c.XY()}, false
			}
			b.Min.X, b.Min.Y = min(b.Min.X, c.X), min(b.Min.Y, c.Y)
			b.Max.X, b.Max.Y = max(b.Max.X, c.X), max(b.Max.Y, c.Y)
		}
	}
	return b
}

// Corners devuelve las esquinas de la extensión de s en el mundo, en orden
// antihorario.
func Corners(s Space) [4]units.Vec3 {
	t, e := s.Pose.Transform(), s.Extent
	return [4]units.Vec3{
		t.Apply(units.V3(e.Min.X, e.Min.Y, 0)), t.Apply(units.V3(e.Max.X, e.Min.Y, 0)),
		t.Apply(units.V3(e.Max.X, e.Max.Y, 0)), t.Apply(units.V3(e.Min.X, e.Max.Y, 0)),
	}
}
