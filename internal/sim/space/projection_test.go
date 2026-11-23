package space_test

import (
	"testing"

	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

func TestProjectionAnchorAndComposition(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 1})
	spawn := func(c cmds.Spawn) {
		w.Enqueue(c)
		w.Step()
	}

	// Habitación raíz.
	w.Enqueue(cmds.Spawn{Space: &space.Space{Kind: space.Static, Extent: units.RectWH(0, 0, 2*units.Meter, 2*units.Meter)}})
	w.Step()
	room := w.Events()[0].Event.(cmds.EntitySpawned).Entity

	// Giantess en (1m, 1m, 0) girada 90°.
	w.Enqueue(cmds.Spawn{
		Space:    &space.Space{Kind: space.Body},
		Location: &comp.Location{Space: room, Pos: units.V3(units.Meter, units.Meter, 0), Facing: fixedmath.QuarterTurn},
	})
	w.Step()
	giantess := w.Events()[0].Event.(cmds.EntitySpawned).Entity

	// Antebrazo: extensión 2D 4×10 cm proyectada en diagonal 30 cm hacia arriba y 40 cm en +X (longitud real 50 cm).
	proj := &space.Projection{
		AnchorA: units.V3(0, 0, 75*units.Centimeter),
		AnchorB: units.V3(40*units.Centimeter, 0, 105*units.Centimeter),
		Climb:   true,
	}
	w.Enqueue(cmds.Spawn{
		Space: &space.Space{
			Kind:       space.Articulated,
			Extent:     units.RectWH(0, 0, 4*units.Centimeter, 10*units.Centimeter),
			Projection: proj,
		},
		Location: &comp.Location{Space: giantess, Pos: units.Vec3{}},
	})
	w.Step()
	forearm := w.Events()[0].Event.(cmds.EntitySpawned).Entity

	// Tiny a mitad de la manga: (2cm, 5cm, 0) local.
	spawn(cmds.Spawn{
		Location: &comp.Location{Space: forearm, Pos: units.V3(2*units.Centimeter, 5*units.Centimeter, 0)},
		Body:     &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter},
	})
	tiny := w.Events()[0].Event.(cmds.EntitySpawned).Entity

	g := space.Of(w.ECS())
	wp, _, err := g.WorldPose(tiny)
	if err != nil {
		t.Fatal(err)
	}
	// En el marco de la giantess, la mitad de AnchorA..AnchorB es (20cm, 0, 90cm).
	// Girado 90° antihorario y trasladado a (1m, 1m, 0): (1m, 1.20m, 90cm).
	want := units.V3(units.Meter, 120*units.Centimeter, 90*units.Centimeter)
	if wp.Sub(want).Len() > 2*units.Micrometer {
		t.Fatalf("WorldPose = %v, quería %v", wp, want)
	}

	// FromWorld debe invertir ToWorld con precisión de pocos µm.
	back, err := g.FromWorld(forearm, wp)
	if err != nil {
		t.Fatal(err)
	}
	wantLocal := units.V3(2*units.Centimeter, 5*units.Centimeter, 0)
	if back.Sub(wantLocal).Len() > 10*units.Micrometer {
		t.Fatalf("FromWorld = %v, quería %v", back, wantLocal)
	}

	// SegmentLen3D de toda la extensión Y (10 cm locales) debe medir 50 cm reales.
	sp, _ := w.ECS().Components(forearm)[0], true
	_ = sp
	fsp, _ := g.ToWorld(forearm, units.V3(2*units.Centimeter, 0, 0))
	fend, _ := g.ToWorld(forearm, units.V3(2*units.Centimeter, 10*units.Centimeter, 0))
	if got := fend.Sub(fsp).Len(); (got - 50*units.Centimeter).Abs() > 10*units.Micrometer {
		t.Fatalf("longitud 3D = %v, quería 50cm", got)
	}
}

func TestClimbExcludedFromFallLandingAndSlopedReceivesFall(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 1})

	// Piso a Z=0.
	w.Enqueue(cmds.Spawn{Space: &space.Space{Kind: space.Static, Extent: units.RectWH(0, 0, 2*units.Meter, 2*units.Meter)}})
	w.Step()
	floor := w.Events()[0].Event.(cmds.EntitySpawned).Entity

	// Tramo vertical de trepada en X=50cm, Y=50cm, de Z=20cm a Z=80cm.
	w.Enqueue(cmds.Spawn{
		Space: &space.Space{
			Kind:   space.Articulated,
			Extent: units.RectWH(0, 0, 10*units.Centimeter, 60*units.Centimeter),
			Projection: &space.Projection{
				AnchorA: units.V3(50*units.Centimeter, 50*units.Centimeter, 20*units.Centimeter),
				AnchorB: units.V3(50*units.Centimeter, 50*units.Centimeter, 80*units.Centimeter),
				Climb:   true,
			},
		},
		Location: &comp.Location{Space: floor},
	})
	w.Step()

	// Superficie inclinada suave (30°) a Z=30cm..50cm en X=[30cm..70cm], Y=[30cm..70cm].
	w.Enqueue(cmds.Spawn{
		Space: &space.Space{
			Kind:   space.Articulated,
			Extent: units.RectWH(0, 0, 40*units.Centimeter, 40*units.Centimeter),
			Projection: &space.Projection{
				Origin: units.V3(30*units.Centimeter, 30*units.Centimeter, 30*units.Centimeter),
				AxisU:  units.V3(space.BasisScale, 0, 0),
				AxisV:  units.V3(0, space.BasisScale, 500*units.Millimeter), // pendiente 0.5
			},
		},
		Location: &comp.Location{Space: floor},
	})
	w.Step()
	slope := w.Events()[0].Event.(cmds.EntitySpawned).Entity

	// Tiny soltado en el aire en (50cm, 50cm, 100cm): ignora el tramo vertical
	// y aterriza en la superficie inclinada a Z = 30cm + 0.5*20cm = 40cm.
	w.Enqueue(cmds.Spawn{
		Location: &comp.Location{Space: floor, Pos: units.V3(50*units.Centimeter, 50*units.Centimeter, units.Meter)},
		Body:     &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter},
	})
	w.Step()
	var fell space.EntityFell
	found := false
	for _, ev := range w.Events() {
		if f, ok := ev.Event.(space.EntityFell); ok {
			fell, found = f, true
		}
	}
	if !found {
		t.Fatal("se esperaba entity_fell")
	}
	if fell.To != slope {
		t.Fatalf("aterrizó en %v, quería la superficie inclinada %v (no el tramo vertical)", fell.To, slope)
	}
	if (fell.Height - 60*units.Centimeter).Abs() > units.Millimeter {
		t.Fatalf("altura de caída = %v, quería 60cm", fell.Height)
	}
}
