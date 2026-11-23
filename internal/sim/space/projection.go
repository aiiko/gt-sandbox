package space

import (
	"errors"
	"math"
	"math/bits"

	"gtsandbox/internal/core/units"
)

// BasisScale es la escala de los vectores eje en una Projection (1 m = 1e6 µm).
// Un eje de norma BasisScale conserva las distancias 1:1 entre la grilla 2D y el mundo.
const BasisScale = units.Meter

var ErrInvalidProjection = errors.New("space: proyección inválida o degenerada")

// Projection define cómo se proyecta la grilla 2D local de un Space al espacio
// 3D de su marco (Location). Permite superficies inclinadas y tramos de trepada
// verticales o diagonales conservando grillas 2D en µm (propuesta 19.2).
//
// Hay dos modos:
//   - Tramo angosto (AnchorA != AnchorB): la coordenada Y local en
//     [Extent.Min.Y, Extent.Max.Y] interpola linealmente de AnchorA a AnchorB;
//     X local se desplaza perpendicularmente en el plano XY (o en X si el tramo
//     es puramente vertical).
//   - Superficie (AxisU != 0 && AxisV != 0): Origin es el origen en el marco
//     del Space y AxisU / AxisV son los desplazamientos por BasisScale a lo
//     largo de X e Y locales.
type Projection struct {
	Origin  units.Vec3   `json:"origin,omitempty"`
	AxisU   units.Vec3   `json:"axis_u,omitempty"`
	AxisV   units.Vec3   `json:"axis_v,omitempty"`
	AnchorA units.Vec3   `json:"anchor_a,omitempty"`
	AnchorB units.Vec3   `json:"anchor_b,omitempty"`
	Width   units.Length `json:"width,omitempty"`
	// Climb marca un tramo de trepada vertical/empinado: sostiene a quien
	// ya está agarrado pero no recibe aterrizajes de caídas.
	Climb bool `json:"climb,omitempty"`
}

// Clone devuelve una copia independiente de p (nil si p == nil).
func (p *Projection) Clone() *Projection {
	if p == nil {
		return nil
	}
	c := *p
	return &c
}

func (p *Projection) isAnchorMode() bool {
	return p != nil && p.AnchorA != p.AnchorB
}

// Valid verifica que la proyección esté bien definida para ext.
func (p *Projection) Valid(ext units.Rect) error {
	if p == nil {
		return nil
	}
	if p.isAnchorMode() {
		if ext.Height() <= 0 {
			return ErrInvalidProjection
		}
		return nil
	}
	if p.AxisU.LengthSq() == 0 || p.AxisV.LengthSq() == 0 {
		return ErrInvalidProjection
	}
	return nil
}

// basis devuelve el marco afín (origin, u, v, n) con u, v, n en escala BasisScale.
func (p *Projection) basis(ext units.Rect) (origin, u, v, n units.Vec3, ok bool) {
	if p == nil {
		return units.Vec3{}, units.V3(BasisScale, 0, 0), units.V3(0, BasisScale, 0), units.V3(0, 0, BasisScale), true
	}
	if p.isAnchorMode() {
		h := ext.Height()
		if h <= 0 {
			return units.Vec3{}, units.Vec3{}, units.Vec3{}, units.Vec3{}, false
		}
		d := p.AnchorB.Sub(p.AnchorA)
		v = units.V3(
			units.Length(mulDiv64(int64(d.X), int64(BasisScale), int64(h))),
			units.Length(mulDiv64(int64(d.Y), int64(BasisScale), int64(h))),
			units.Length(mulDiv64(int64(d.Z), int64(BasisScale), int64(h))),
		)
		if dxy := d.XY().Len(); dxy > 0 {
			u = units.V3(
				units.Length(mulDiv64(-int64(d.Y), int64(BasisScale), int64(dxy))),
				units.Length(mulDiv64(int64(d.X), int64(BasisScale), int64(dxy))),
				0,
			)
		} else {
			u = units.V3(BasisScale, 0, 0)
		}
		midX := (ext.Min.X + ext.Max.X) / 2
		origin = p.AnchorA.Sub(scaleVec3(u, midX)).Sub(scaleVec3(v, ext.Min.Y))
		n = unitCross(u, v)
		return origin, u, v, n, true
	}
	if p.AxisU.LengthSq() == 0 || p.AxisV.LengthSq() == 0 {
		return units.Vec3{}, units.Vec3{}, units.Vec3{}, units.Vec3{}, false
	}
	n = unitCross(p.AxisU, p.AxisV)
	return p.Origin, p.AxisU, p.AxisV, n, true
}

// Project convierte un punto local (x, y, z) de la grilla de s a coordenadas
// 3D en el marco de Location de s. Si s no tiene Projection, devuelve p.
func (s Space) Project(p units.Vec3) units.Vec3 {
	if s.Projection == nil {
		return p
	}
	origin, u, v, n, ok := s.Projection.basis(s.Extent)
	if !ok {
		return p
	}
	return origin.Add(scaleVec3(u, p.X)).Add(scaleVec3(v, p.Y)).Add(scaleVec3(n, p.Z))
}

// Unproject convierte un punto 3D en el marco de Location de s a coordenadas
// locales (x, y, z) de la grilla de s. Si s no tiene Projection, devuelve q.
func (s Space) Unproject(q units.Vec3) units.Vec3 {
	if s.Projection == nil {
		return q
	}
	origin, u, v, n, ok := s.Projection.basis(s.Extent)
	if !ok {
		return q
	}
	d := q.Sub(origin)
	uu := dot64(u, u)
	vv := dot64(v, v)
	uv := dot64(u, v)
	du := dot64(d, u)
	dv := dot64(d, v)
	var px, py int64
	if uv == 0 {
		px = mulDiv64(du, int64(BasisScale), uu)
		py = mulDiv64(dv, int64(BasisScale), vv)
	} else {
		// Sistema 2x2 en mm para evitar desbordamiento del determinante.
		const sc = 1000
		a, b, c := uu/(sc*sc), uv/(sc*sc), vv/(sc*sc)
		det := a*c - b*b
		if det != 0 {
			ru, rv := du/(sc*sc), dv/(sc*sc)
			px = mulDiv64(ru*c-rv*b, int64(BasisScale), det)
			py = mulDiv64(a*rv-b*ru, int64(BasisScale), det)
		} else {
			px = mulDiv64(du, int64(BasisScale), uu)
			py = mulDiv64(dv, int64(BasisScale), vv)
		}
	}
	var pz int64
	if nn := dot64(n, n); nn > 0 {
		pz = mulDiv64(dot64(d, n), int64(BasisScale), nn)
	}
	return units.V3(units.Length(px), units.Length(py), units.Length(pz))
}

// IsClimb indica si s es un tramo vertical o demasiado empinado (>60°) para
// recibir aterrizajes de caídas.
func (s Space) IsClimb() bool {
	if s.Projection == nil {
		return false
	}
	if s.Projection.Climb {
		return true
	}
	_, _, _, n, ok := s.Projection.basis(s.Extent)
	return ok && n.Z.Abs() < BasisScale/2
}

// LandingAt intersecta la vertical en localXY (en el marco de Location de s)
// con la superficie proyectada de s. Devuelve las coordenadas locales (u, v)
// dentro de s.Extent y la altura Z de la superficie en ese punto.
func (s Space) LandingAt(localXY units.Vec2) (units.Vec2, units.Length, bool) {
	if s.Projection == nil {
		if !s.Extent.Contains(localXY) {
			return units.Vec2{}, 0, false
		}
		return localXY, 0, true
	}
	origin, u, v, _, ok := s.Projection.basis(s.Extent)
	if !ok {
		return units.Vec2{}, 0, false
	}
	det := int64(u.X)*int64(v.Y) - int64(v.X)*int64(u.Y)
	if det == 0 {
		return units.Vec2{}, 0, false
	}
	dx := int64(localXY.X - origin.X)
	dy := int64(localXY.Y - origin.Y)
	px := units.Length(mulDiv64(dx*int64(v.Y)-dy*int64(v.X), int64(BasisScale), det))
	py := units.Length(mulDiv64(int64(u.X)*dy-int64(u.Y)*dx, int64(BasisScale), det))
	uv := units.V2(px, py)
	if !s.Extent.Contains(uv) {
		return units.Vec2{}, 0, false
	}
	surfZ := origin.Z +
		units.Length(mulDiv64(int64(u.Z), int64(px), int64(BasisScale))) +
		units.Length(mulDiv64(int64(v.Z), int64(py), int64(BasisScale)))
	return uv, surfZ, true
}

// SegmentLen3D devuelve la longitud real en 3D del segmento local a→b sobre s.
// Para superficies planas (Projection == nil) coincide con la distancia 2D.
func (s Space) SegmentLen3D(a, b units.Vec2) units.Length {
	if s.Projection == nil {
		return b.Sub(a).Len()
	}
	pa := s.Project(units.V3(a.X, a.Y, 0))
	pb := s.Project(units.V3(b.X, b.Y, 0))
	return pb.Sub(pa).Len()
}

func scaleVec3(v units.Vec3, s units.Length) units.Vec3 {
	return units.V3(
		units.Length(mulDiv64(int64(v.X), int64(s), int64(BasisScale))),
		units.Length(mulDiv64(int64(v.Y), int64(s), int64(BasisScale))),
		units.Length(mulDiv64(int64(v.Z), int64(s), int64(BasisScale))),
	)
}

func dot64(a, b units.Vec3) int64 {
	return int64(a.X)*int64(b.X) + int64(a.Y)*int64(b.Y) + int64(a.Z)*int64(b.Z)
}

// unitCross devuelve (u × v) normalizado a BasisScale (con +Z hacia arriba si
// es posible). Si u y v son colineales, devuelve (0, 0, BasisScale).
func unitCross(u, v units.Vec3) units.Vec3 {
	// Reducir por 1000 para que el producto cruz quepa holgado en int64.
	ux, uy, uz := int64(u.X)/1000, int64(u.Y)/1000, int64(u.Z)/1000
	vx, vy, vz := int64(v.X)/1000, int64(v.Y)/1000, int64(v.Z)/1000
	cx := uy*vz - uz*vy
	cy := uz*vx - ux*vz
	cz := ux*vy - uy*vx
	lenSq := uint64(cx*cx + cy*cy + cz*cz)
	if lenSq == 0 {
		return units.V3(0, 0, BasisScale)
	}
	l := int64(units.Isqrt(lenSq))
	if l == 0 {
		return units.V3(0, 0, BasisScale)
	}
	return units.V3(
		units.Length(mulDiv64(cx, int64(BasisScale), l)),
		units.Length(mulDiv64(cy, int64(BasisScale), l)),
		units.Length(mulDiv64(cz, int64(BasisScale), l)),
	)
}

func absU64(v int64) uint64 {
	if v < 0 {
		return uint64(-v)
	}
	return uint64(v)
}

// mulDiv64 calcula a*b/den con producto intermedio de 128 bits y redondeo
// hacia cero (determinista y sin floats).
func mulDiv64(a, b, den int64) int64 {
	if den == 0 || a == 0 || b == 0 {
		return 0
	}
	neg := (a < 0) != (b < 0) != (den < 0)
	hi, lo := bits.Mul64(absU64(a), absU64(b))
	ud := absU64(den)
	if hi >= ud {
		if neg {
			return math.MinInt64
		}
		return math.MaxInt64
	}
	q, _ := bits.Div64(hi, lo, ud)
	if neg {
		return -int64(q)
	}
	return int64(q)
}
