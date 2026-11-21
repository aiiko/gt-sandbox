package units

import "testing"

func TestConstants(t *testing.T) {
	if Millimeter != 1_000 || Centimeter != 10_000 || Meter != 1_000_000 {
		t.Fatalf("constantes incorrectas: mm=%d cm=%d m=%d", Millimeter, Centimeter, Meter)
	}
}

func TestProposalScale(t *testing.T) {
	// Giantess de 170 cm y tiny de 8 mm: relación 212,5×.
	giantess := 170 * Centimeter
	tiny := 8 * Millimeter
	if got := giantess * 10 / tiny; got != 2125 {
		t.Fatalf("relación×10 = %d, esperado 2125", got)
	}
}

func TestVec3Ops(t *testing.T) {
	a := V3(1*Millimeter, 2*Millimeter, 3*Millimeter)
	b := V3(500, 500, 500)
	if got := a.Add(b).Sub(b); got != a {
		t.Fatalf("Add/Sub no es inversa: %v", got)
	}
	if got := a.Scale(2); got != V3(2*Millimeter, 4*Millimeter, 6*Millimeter) {
		t.Fatalf("Scale: %v", got)
	}
	if got := a.MulDiv(1, 2); got != V3(500, 1000, 1500) {
		t.Fatalf("MulDiv: %v", got)
	}
	if got := V3(3, 4, 0).LengthSq(); got != 25 {
		t.Fatalf("LengthSq: %d", got)
	}
}

func TestLengthSqNoOverflowInHouseRange(t *testing.T) {
	// 1 km por eje debe ser seguro.
	v := V3(1000*Meter, 1000*Meter, 1000*Meter)
	if v.LengthSq() <= 0 {
		t.Fatal("overflow en LengthSq dentro del rango documentado")
	}
}

func TestRect(t *testing.T) {
	r := RectWH(0, 0, 10*Centimeter, 5*Centimeter)
	if !r.Contains(V2(0, 0)) || r.Contains(V2(10*Centimeter, 0)) {
		t.Fatal("Contains debe ser semiabierto")
	}
	o := RectWH(5*Centimeter, 2*Centimeter, 10*Centimeter, 10*Centimeter)
	in := r.Intersect(o)
	if in != RectWH(5*Centimeter, 2*Centimeter, 5*Centimeter, 3*Centimeter) {
		t.Fatalf("Intersect: %v", in)
	}
	if r.Overlaps(RectWH(10*Centimeter, 0, 1, 1)) {
		t.Fatal("rectángulos adyacentes no se solapan")
	}
}

func TestString(t *testing.T) {
	cases := map[Length]string{
		5:                "5µm",
		8 * Millimeter:   "8.000mm",
		170 * Centimeter: "1.700m",
		-25 * Centimeter: "-25.00cm",
	}
	for l, want := range cases {
		if got := l.String(); got != want {
			t.Errorf("%d.String() = %q, esperado %q", int64(l), got, want)
		}
	}
}
