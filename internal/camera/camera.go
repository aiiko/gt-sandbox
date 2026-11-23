// Package camera proyecta coordenadas de mundo (µm, enteros) a la pantalla
// (píxeles, float) y vuelve. Es matemática pura, sin Ebitengine.
//
// La cámara sigue a un Space ancla: su centro está en coordenadas locales de
// ese Space y la vista gira con él. Con el ancla en una raíz es la vista
// macro; con el ancla en una palma, la vista micro que acompaña a la mano.
// Las coordenadas se pasan a float relativas al ancla, así que la precisión
// no depende de lo lejos que esté del origen del mundo.
package camera

import (
	"image"
	"math"

	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
)

// Límites de zoom, en píxeles por µm (10 px/m a 1 px/µm).
const (
	MinZoom = 1e-5
	MaxZoom = 1.0
)

// Camera es el estado de la cámara. En pantalla, +Y local va hacia arriba.
type Camera struct {
	Anchor ids.ID  // Space que sigue la cámara (Nil: mundo)
	X, Y   float64 // centro, en µm locales del ancla
	Zoom   float64 // píxeles por µm
}

// Projector convierte coordenadas para un ancla y un tamaño de pantalla dados.
type Projector struct {
	ox, oy, cos, sin float64
	rot              fixedmath.Angle
	cam              Camera
	hw, hh           float64
}

// Project arma el Projector: anchor es la pose de mundo del ancla (local →
// mundo) y w×h el tamaño de la pantalla.
func (c Camera) Project(anchor fixedmath.Transform, w, h int) Projector {
	a := Radians(anchor.Rot)
	return Projector{
		ox: float64(anchor.Pos.X), oy: float64(anchor.Pos.Y),
		cos: math.Cos(a), sin: math.Sin(a), rot: anchor.Rot,
		cam: c, hw: float64(w) / 2, hh: float64(h) / 2,
	}
}

// Radians convierte un Angle a radianes (solo para dibujar).
func Radians(a fixedmath.Angle) float64 {
	return float64(a.Norm()) * 2 * math.Pi / float64(fixedmath.Turn)
}

// Bounds es el rectángulo de pantalla para el que se armó el Projector.
func (p Projector) Bounds() image.Rectangle {
	return image.Rect(0, 0, int(math.Round(2*p.hw)), int(math.Round(2*p.hh)))
}

// Local lleva un punto de mundo a coordenadas del ancla (µm, float).
func (p Projector) Local(v units.Vec3) (float64, float64) {
	dx, dy := float64(v.X)-p.ox, float64(v.Y)-p.oy
	return dx*p.cos + dy*p.sin, -dx*p.sin + dy*p.cos
}

// ToScreen lleva un punto de mundo a píxeles.
func (p Projector) ToScreen(v units.Vec3) (float32, float32) {
	lx, ly := p.Local(v)
	return p.localToScreen(lx, ly)
}

func (p Projector) localToScreen(lx, ly float64) (float32, float32) {
	return float32(p.hw + (lx-p.cam.X)*p.cam.Zoom), float32(p.hh - (ly-p.cam.Y)*p.cam.Zoom)
}

func (p Projector) screenToLocal(sx, sy float64) (float64, float64) {
	return (sx-p.hw)/p.cam.Zoom + p.cam.X, -(sy-p.hh)/p.cam.Zoom + p.cam.Y
}

// ToWorld lleva un píxel al plano de mundo.
func (p Projector) ToWorld(sx, sy float64) units.Vec2 {
	lx, ly := p.screenToLocal(sx, sy)
	return units.V2(
		units.Length(math.Round(p.ox+lx*p.cos-ly*p.sin)),
		units.Length(math.Round(p.oy+lx*p.sin+ly*p.cos)),
	)
}

// Pixels convierte una longitud a píxeles.
func (p Projector) Pixels(l units.Length) float32 { return float32(float64(l) * p.cam.Zoom) }

// ScreenAngle devuelve un ángulo de mundo en radianes de pantalla (sentido
// horario, porque la Y de pantalla va hacia abajo).
func (p Projector) ScreenAngle(a fixedmath.Angle) float64 { return -Radians(a.Sub(p.rot)) }

// ZoomAt multiplica el zoom por f manteniendo fijo el punto bajo (sx, sy).
func (c *Camera) ZoomAt(p Projector, sx, sy, f float64) {
	lx, ly := p.screenToLocal(sx, sy)
	c.Zoom = min(max(c.Zoom*f, MinZoom), MaxZoom)
	c.X = lx - (sx-p.hw)/c.Zoom
	c.Y = ly + (sy-p.hh)/c.Zoom
}

// Pan desplaza la vista en píxeles (arrastrar el mundo con el mouse).
func (c *Camera) Pan(dx, dy float64) {
	c.X -= dx / c.Zoom
	c.Y += dy / c.Zoom
}

// Fit centra el rectángulo r (en coordenadas del ancla) y elige el zoom para
// que ocupe el 90 % de la pantalla.
func (c *Camera) Fit(r units.Rect, w, h int) {
	c.X = (float64(r.Min.X) + float64(r.Max.X)) / 2
	c.Y = (float64(r.Min.Y) + float64(r.Max.Y)) / 2
	rw, rh := max(float64(r.Width()), 1), max(float64(r.Height()), 1)
	c.Zoom = min(max(0.9*min(float64(w)/rw, float64(h)/rh), MinZoom), MaxZoom)
}

// Follow cambia el ancla sin mover la vista: el punto del centro de la
// pantalla sigue en el centro. old es el Projector actual y anchor la pose
// del ancla nueva.
func (c *Camera) Follow(id ids.ID, anchor fixedmath.Transform, old Projector) {
	center := old.ToWorld(old.hw, old.hh)
	l := anchor.ApplyInv(units.V3(center.X, center.Y, anchor.Pos.Z))
	c.Anchor, c.X, c.Y = id, float64(l.X), float64(l.Y)
}
