// Package units define las unidades físicas canónicas de la simulación.
//
// Toda magnitud espacial autoritativa se expresa como Length: micrómetros en
// int64 (decisión D1). Los floats solo aparecen en conversiones para mostrar.
//
// Rango seguro: las operaciones al cuadrado (LengthSq, DistSq) no desbordan
// mientras las coordenadas se mantengan por debajo de ~1,7·10⁹ µm (~1,7 km)
// por eje, muy por encima de cualquier casa.
package units

import "fmt"

// Length es una longitud en micrómetros.
type Length int64

const (
	Micrometer Length = 1
	Millimeter        = 1000 * Micrometer
	Centimeter        = 10 * Millimeter
	Meter             = 100 * Centimeter
)

// Millimeters devuelve la longitud en mm como float. Solo para mostrar/debug.
func (l Length) Millimeters() float64 { return float64(l) / float64(Millimeter) }

// Meters devuelve la longitud en metros como float. Solo para mostrar/debug.
func (l Length) Meters() float64 { return float64(l) / float64(Meter) }

// Abs devuelve el valor absoluto.
func (l Length) Abs() Length {
	if l < 0 {
		return -l
	}
	return l
}

// MulDiv calcula l*num/den con redondeo hacia cero. Útil para escalar sin floats.
// Panic si den == 0.
func (l Length) MulDiv(num, den int64) Length {
	return Length(int64(l) * num / den)
}

func (l Length) String() string {
	a := l.Abs()
	switch {
	case a >= Meter:
		return fmt.Sprintf("%.3fm", l.Meters())
	case a >= Centimeter:
		return fmt.Sprintf("%.2fcm", float64(l)/float64(Centimeter))
	case a >= Millimeter:
		return fmt.Sprintf("%.3fmm", l.Millimeters())
	default:
		return fmt.Sprintf("%dµm", int64(l))
	}
}

// Vec3 es una posición o desplazamiento en coordenadas locales de un Space.
// X/Y forman el plano de la superficie; Z es la elevación.
type Vec3 struct {
	X, Y, Z Length
}

func V3(x, y, z Length) Vec3 { return Vec3{x, y, z} }

func (a Vec3) Add(b Vec3) Vec3 { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Sub(b Vec3) Vec3 { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Neg() Vec3       { return Vec3{-a.X, -a.Y, -a.Z} }

// Scale multiplica cada componente por k.
func (a Vec3) Scale(k int64) Vec3 {
	return Vec3{a.X * Length(k), a.Y * Length(k), a.Z * Length(k)}
}

// MulDiv escala cada componente por num/den (redondeo hacia cero).
func (a Vec3) MulDiv(num, den int64) Vec3 {
	return Vec3{a.X.MulDiv(num, den), a.Y.MulDiv(num, den), a.Z.MulDiv(num, den)}
}

// LengthSq devuelve el cuadrado de la norma, en µm².
func (a Vec3) LengthSq() int64 {
	x, y, z := int64(a.X), int64(a.Y), int64(a.Z)
	return x*x + y*y + z*z
}

// DistSq devuelve la distancia al cuadrado entre a y b, en µm².
func (a Vec3) DistSq(b Vec3) int64 { return a.Sub(b).LengthSq() }

// XY devuelve la proyección sobre el plano de la superficie.
func (a Vec3) XY() Vec2 { return Vec2{a.X, a.Y} }

func (a Vec3) String() string { return fmt.Sprintf("(%v, %v, %v)", a.X, a.Y, a.Z) }

// Vec2 es una posición sobre el plano de una superficie.
type Vec2 struct {
	X, Y Length
}

func V2(x, y Length) Vec2 { return Vec2{x, y} }

func (a Vec2) Add(b Vec2) Vec2 { return Vec2{a.X + b.X, a.Y + b.Y} }
func (a Vec2) Sub(b Vec2) Vec2 { return Vec2{a.X - b.X, a.Y - b.Y} }

func (a Vec2) LengthSq() int64 {
	x, y := int64(a.X), int64(a.Y)
	return x*x + y*y
}

func (a Vec2) DistSq(b Vec2) int64 { return a.Sub(b).LengthSq() }

func (a Vec2) String() string { return fmt.Sprintf("(%v, %v)", a.X, a.Y) }

// Rect es un rectángulo alineado a ejes, semiabierto: [Min, Max).
type Rect struct {
	Min, Max Vec2
}

// RectWH construye un rectángulo desde una esquina y un tamaño.
func RectWH(x, y, w, h Length) Rect {
	return Rect{Min: Vec2{x, y}, Max: Vec2{x + w, y + h}}
}

func (r Rect) Width() Length  { return r.Max.X - r.Min.X }
func (r Rect) Height() Length { return r.Max.Y - r.Min.Y }

// Empty indica si el rectángulo no tiene área.
func (r Rect) Empty() bool { return r.Max.X <= r.Min.X || r.Max.Y <= r.Min.Y }

// Contains indica si p está dentro de r (borde máximo excluido).
func (r Rect) Contains(p Vec2) bool {
	return p.X >= r.Min.X && p.X < r.Max.X && p.Y >= r.Min.Y && p.Y < r.Max.Y
}

// Intersect devuelve la intersección; puede ser vacía.
func (r Rect) Intersect(o Rect) Rect {
	out := Rect{
		Min: Vec2{max(r.Min.X, o.Min.X), max(r.Min.Y, o.Min.Y)},
		Max: Vec2{min(r.Max.X, o.Max.X), min(r.Max.Y, o.Max.Y)},
	}
	if out.Empty() {
		return Rect{}
	}
	return out
}

// Overlaps indica si r y o comparten área.
func (r Rect) Overlaps(o Rect) bool { return !r.Intersect(o).Empty() }

func (r Rect) String() string { return fmt.Sprintf("[%v - %v]", r.Min, r.Max) }
