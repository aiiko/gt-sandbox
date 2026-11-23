package camera

import (
	"image"
	"testing"

	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/units"
)

func near(a, b units.Vec2, tol units.Length) bool {
	d := a.Sub(b)
	return d.X.Abs() <= tol && d.Y.Abs() <= tol
}

// Ida y vuelta mundo→pantalla→mundo con un ancla rotada y lejos del origen;
// +Y local va hacia arriba en pantalla.
func TestProjectRoundTrip(t *testing.T) {
	anchor := fixedmath.Transform{Pos: units.V3(5000*units.Meter, -300*units.Meter, units.Meter), Rot: fixedmath.Degrees(30)}
	c := Camera{X: 1000, Y: 2000, Zoom: 0.01}
	p := c.Project(anchor, 800, 600)
	for _, l := range []units.Vec3{{}, units.V3(1000, 2000, 0), units.V3(-7*units.Centimeter, 3*units.Centimeter, 0)} {
		w := anchor.Apply(l)
		sx, sy := p.ToScreen(w)
		if got := p.ToWorld(float64(sx), float64(sy)); !near(got, w.XY(), 200) {
			t.Fatalf("%v → (%v,%v) → %v", w, sx, sy, got)
		}
	}
	_, y0 := p.ToScreen(anchor.Apply(units.V3(0, 0, 0)))
	_, y1 := p.ToScreen(anchor.Apply(units.V3(0, units.Centimeter, 0)))
	if y1 >= y0 {
		t.Fatal("+Y local debe ir hacia arriba")
	}
}

// ZoomAt deja fijo el punto bajo el cursor; Follow no mueve la vista.
func TestZoomAtAndFollow(t *testing.T) {
	c := Camera{Zoom: 0.001}
	p := c.Project(fixedmath.Identity, 800, 600)
	before := p.ToWorld(100, 50)
	c.ZoomAt(p, 100, 50, 8)
	p = c.Project(fixedmath.Identity, 800, 600)
	if after := p.ToWorld(100, 50); !near(before, after, 2) {
		t.Fatalf("zoom movió el punto: %v → %v", before, after)
	}
	center := p.ToWorld(400, 300)
	anchor := fixedmath.Transform{Pos: units.V3(units.Meter, 0, 0), Rot: fixedmath.QuarterTurn}
	c.Follow(7, anchor, p)
	if got := c.Project(anchor, 800, 600).ToWorld(400, 300); !near(got, center, 2) {
		t.Fatalf("follow movió el centro: %v → %v", center, got)
	}
}

// Resized conserva el encuadre: el centro de una pantalla es el de la otra y
// las distancias escalan con el zoom.
func TestResizedKeepsCenter(t *testing.T) {
	anchor := fixedmath.Transform{Pos: units.V3(3*units.Meter, -2*units.Meter, 0)}
	p := Camera{X: 1500, Y: -800, Zoom: 0.0004}.Project(anchor, 960, 540)
	q := p.Resized(0.0002, 1920, 1080)
	if got, want := q.ToWorld(960, 540), p.ToWorld(480, 270); !near(got, want, 5) {
		t.Fatalf("centro %v, quería %v", got, want)
	}
	if q.Bounds() != image.Rect(0, 0, 1920, 1080) || q.Pixels(units.Meter) != p.Pixels(units.Meter)/2 {
		t.Fatalf("pantalla %v, %v px/m", q.Bounds(), q.Pixels(units.Meter))
	}
}
