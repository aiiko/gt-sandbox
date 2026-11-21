package fixedmath

import (
	"fmt"

	"gtsandbox/internal/core/units"
)

// Angle es un ángulo en pasos de 1/Turn de vuelta, antihorario, alrededor del
// eje Z. Se normaliza a [0, Turn).
type Angle int32

const (
	Turn        Angle = 4096
	QuarterTurn Angle = Turn / 4
	HalfTurn    Angle = Turn / 2

	fracBits = 30
	one      = int64(1) << fracBits
)

// Degrees convierte grados enteros a Angle (redondeando al paso más cercano).
func Degrees(d int) Angle {
	n := int64(d) * int64(Turn)
	if n >= 0 {
		n += 180
	} else {
		n -= 180
	}
	return Angle(n / 360).Norm()
}

// Norm devuelve el ángulo equivalente en [0, Turn).
func (a Angle) Norm() Angle {
	a %= Turn
	if a < 0 {
		a += Turn
	}
	return a
}

// Add suma ángulos normalizando.
func (a Angle) Add(b Angle) Angle { return (a + b).Norm() }

// Sub resta ángulos normalizando.
func (a Angle) Sub(b Angle) Angle { return (a - b).Norm() }

// DegreesF devuelve el ángulo en grados. Solo para mostrar.
func (a Angle) DegreesF() float64 { return float64(a.Norm()) * 360 / float64(Turn) }

func (a Angle) String() string { return fmt.Sprintf("%.1f°", a.DegreesF()) }

// sincos devuelve seno y coseno en punto fijo (escala 2^30).
func sincos(a Angle) (sin, cos int64) {
	return sinFixed(a), sinFixed(a + QuarterTurn)
}

func sinFixed(a Angle) int64 {
	a = a.Norm()
	q, r := a/QuarterTurn, a%QuarterTurn
	switch q {
	case 0:
		return int64(sinTable[r])
	case 1:
		return int64(sinTable[QuarterTurn-r])
	case 2:
		return -int64(sinTable[r])
	default:
		return -int64(sinTable[QuarterTurn-r])
	}
}

// roundShift divide por 2^30 redondeando al entero más cercano (mitades hacia +∞).
func roundShift(v int64) int64 { return (v + one/2) >> fracBits }

// Rotate rota v por a alrededor del origen.
func Rotate(v units.Vec2, a Angle) units.Vec2 {
	s, c := sincos(a)
	x, y := int64(v.X), int64(v.Y)
	return units.Vec2{
		X: units.Length(roundShift(x*c - y*s)),
		Y: units.Length(roundShift(x*s + y*c)),
	}
}

// Transform ubica un sistema de coordenadas hijo dentro de su padre:
// primero rota (alrededor de Z) y después traslada.
type Transform struct {
	Pos units.Vec3 `json:"pos"`
	Rot Angle      `json:"rot"`
}

// Identity es la transformación nula.
var Identity = Transform{}

// Apply lleva un punto de coordenadas hijas a coordenadas del padre.
func (t Transform) Apply(p units.Vec3) units.Vec3 {
	r := Rotate(p.XY(), t.Rot)
	return units.Vec3{X: r.X + t.Pos.X, Y: r.Y + t.Pos.Y, Z: p.Z + t.Pos.Z}
}

// ApplyInv lleva un punto de coordenadas del padre a coordenadas hijas.
func (t Transform) ApplyInv(p units.Vec3) units.Vec3 {
	d := p.Sub(t.Pos)
	r := Rotate(d.XY(), -t.Rot)
	return units.Vec3{X: r.X, Y: r.Y, Z: d.Z}
}

// Mul compone: el resultado aplica primero child y después t.
// Si t lleva de A a B y child de C a A, t.Mul(child) lleva de C a B.
func (t Transform) Mul(child Transform) Transform {
	return Transform{Pos: t.Apply(child.Pos), Rot: t.Rot.Add(child.Rot)}
}

// Inverse devuelve la transformación inversa.
func (t Transform) Inverse() Transform {
	// x_hijo = R⁻¹·(x - p) = R⁻¹·x - R⁻¹·p
	return Transform{Pos: Transform{Rot: t.Rot}.ApplyInv(t.Pos).Neg(), Rot: (-t.Rot).Norm()}
}

func (t Transform) String() string { return fmt.Sprintf("{%v %v}", t.Pos, t.Rot) }
