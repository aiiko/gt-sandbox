package view

import (
	"errors"
	"testing"

	"gtsandbox/internal/core/units"
	"gtsandbox/internal/scenario"
)

var errEnd = errors.New("fin")

func load(t *testing.T) (*scenario.Instance, *scenario.Player) {
	t.Helper()
	spec, err := scenario.ReadFile("../../tests/scenarios/giantess_carries_tiny.json")
	if err != nil {
		t.Fatal(err)
	}
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	p, err := in.Player()
	if err != nil {
		t.Fatal(err)
	}
	return in, p
}

// newRunner corre el escenario hasta su último tick y se detiene ahí.
func newRunner(t *testing.T) (*scenario.Instance, *Runner) {
	in, p := load(t)
	return in, NewRunner(in.World, func() error {
		if p.Done() {
			return errEnd
		}
		return p.Step()
	})
}

// El frame refleja el escenario: la palma lleva la grilla, el contacto abre un
// portal dinámico, el tiny en camino tiene ruta, los Spaces van de abajo
// hacia arriba, y Pick encuentra al tiny y a las superficies.
func TestCaptureAndPick(t *testing.T) {
	in, r := newRunner(t)
	for in.World.Tick()-in.Start < 30 {
		r.StepOnce()
	}
	f := r.Cur
	for i := 1; i < len(f.Spaces); i++ {
		if f.Spaces[i].Pose.Pos.Z < f.Spaces[i-1].Pose.Pos.Z {
			t.Fatal("Spaces fuera de orden")
		}
	}
	palm, _ := f.SpaceByID(in.Refs["palm"])
	if palm == nil || palm.Grid == nil || palm.Depth != 3 || !palm.Surface {
		t.Fatalf("palma: %+v", palm)
	}
	if len(f.Contacts) == 0 || len(f.Portals) != 2 || !f.Portals[1].Dynamic && !f.Portals[0].Dynamic {
		t.Fatalf("contactos %v, portales %v", f.Contacts, f.Portals)
	}
	tiny, _ := f.ThingByID(in.Refs["tiny"])
	if tiny == nil || tiny.Mover == nil || tiny.Mover.Status != "moving" || len(tiny.Mover.Route) < 2 {
		t.Fatalf("tiny: %+v", tiny)
	}
	if got := f.Pick(tiny.Pose.Pos.XY(), units.Millimeter); got != tiny.ID {
		t.Fatalf("Pick en el tiny devolvió %v", got)
	}
	d1, _ := f.SpaceByID(in.Refs["desk1"])
	corner := d1.Pose.Pos.XY().Add(units.V2(units.Centimeter, units.Centimeter))
	if got, l, ok := f.SurfaceAt(corner); !ok || got != d1.ID || l != units.V2(units.Centimeter, units.Centimeter) {
		t.Fatalf("SurfaceAt: %v %v %v", got, l, ok)
	}
}

func TestInterpolate(t *testing.T) {
	a := &Frame{Things: []Thing{{ID: 1, Pose: Pose{Pos: units.V3(0, 0, 0)}}, {ID: 2}}}
	b := &Frame{Things: []Thing{{ID: 1, Pose: Pose{Pos: units.V3(10, 20, 0)}}, {ID: 2, Pose: Pose{Pos: units.V3(2*units.Meter, 0, 0)}}}}
	m := Interpolate(a, b, 0.5)
	if m.Things[0].Pose.Pos != units.V3(5, 10, 0) || m.Things[1].Pose != b.Things[1].Pose {
		t.Fatalf("%+v", m.Things)
	}
	if b.Things[0].Pose.Pos != units.V3(10, 20, 0) {
		t.Fatal("Interpolate modificó cur")
	}
}
