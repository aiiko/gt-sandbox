package fixedmath

import (
	"math/rand/v2"
	"testing"

	"gtsandbox/internal/core/units"
)

func TestAngles(t *testing.T) {
	if Degrees(90) != QuarterTurn || Degrees(-90) != 3*QuarterTurn || Degrees(360) != 0 {
		t.Fatal("Degrees")
	}
	if sinFixed(0) != 0 || sinFixed(QuarterTurn) != one || sinFixed(HalfTurn) != 0 || sinFixed(-QuarterTurn) != -one {
		t.Fatal("valores exactos del seno")
	}
	for a := Angle(0); a < Turn; a += 37 {
		s, c := sincos(a)
		// sin² + cos² ≈ 1 (error relativo de la tabla < 1e-8)
		if d := (s*s+c*c)>>30 - one; d < -8 || d > 8 {
			t.Fatalf("sin²+cos² en %v: desvío %d", a, d)
		}
	}
}

func TestQuarterTurnsExact(t *testing.T) {
	v := units.V2(3*units.Centimeter, -7)
	want := []units.Vec2{v, {X: 7, Y: 3 * units.Centimeter}, {X: -3 * units.Centimeter, Y: 7}, {X: -7, Y: -3 * units.Centimeter}}
	for i, w := range want {
		if got := Rotate(v, Angle(i)*QuarterTurn); got != w {
			t.Fatalf("%d×90°: %v, esperado %v", i, got, w)
		}
	}
}

func TestTransformRoundTrip(t *testing.T) {
	r := rand.New(rand.NewPCG(5, 6))
	rnd := func() units.Length { return units.Length(r.IntN(20*int(units.Meter))) - 10*units.Meter }
	for range 1000 {
		tr := Transform{Pos: units.V3(rnd(), rnd(), rnd()), Rot: Angle(r.IntN(int(Turn)))}
		p := units.V3(rnd(), rnd(), rnd())
		back := tr.ApplyInv(tr.Apply(p))
		if d := back.Sub(p); d.X.Abs() > 1 || d.Y.Abs() > 1 || d.Z != 0 {
			t.Fatalf("Apply/ApplyInv: error %v con %v", d, tr)
		}
		inv := tr.Inverse().Apply(tr.Apply(p))
		if d := inv.Sub(p); d.X.Abs() > 2 || d.Y.Abs() > 2 || d.Z != 0 {
			t.Fatalf("Inverse: error %v", d)
		}
	}
}

func TestMulComposes(t *testing.T) {
	parent := Transform{Pos: units.V3(units.Meter, 0, 0), Rot: QuarterTurn}
	child := Transform{Pos: units.V3(10*units.Centimeter, 0, units.Meter), Rot: QuarterTurn}
	p := units.V3(units.Centimeter, 0, 0)
	if got, want := parent.Mul(child).Apply(p), parent.Apply(child.Apply(p)); got != want {
		t.Fatalf("Mul: %v vs %v", got, want)
	}
	// Con cuartos de vuelta, todo exacto: (1cm,0) rota 180° → (-1cm,0); + (0,10cm,1m) + (1m,0,0).
	if got := parent.Mul(child).Apply(p); got != units.V3(units.Meter-units.Centimeter, 10*units.Centimeter, units.Meter) {
		t.Fatalf("composición = %v", got)
	}
}
