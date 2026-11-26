package crush_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

func newGarmentCrushWorld(t *testing.T) *world.World {
	t.Helper()
	w := world.New(world.Config{Seed: 1})
	if _, err := comp.Register(w.ECS()); err != nil {
		t.Fatal(err)
	}
	for _, fn := range []func(*world.World) error{space.Install, nav.Install, attach.Install, crush.Install} {
		if err := fn(w); err != nil {
			t.Fatal(err)
		}
	}
	return w
}

func TestGarmentInnerPressureThreatCrushAndSafeZones(t *testing.T) {
	w := newGarmentCrushWorld(t)
	r := w.ECS()
	cs := comp.Of(r)
	cm := units.Centimeter

	room := r.Spawn()
	ecs.Get[space.Space](r).Set(room, space.Space{Kind: space.Static, Extent: units.Rect{Max: units.V2(300*cm, 300*cm)}})

	giantess := r.Spawn()
	ecs.Get[space.Space](r).Set(giantess, space.Space{Kind: space.Body})
	cs.Location.Set(giantess, comp.Location{Space: room, Pos: units.V3(100*cm, 100*cm, 0)})

	foot := r.Spawn()
	ecs.Get[space.Space](r).Set(foot, space.Space{Kind: space.Articulated, Extent: units.Rect{Max: units.V2(10*cm, 24*cm)}})
	cs.Location.Set(foot, comp.Location{Space: giantess, Pos: units.V3(0, 0, 2*cm)})

	slipper := r.Spawn()
	ecs.Get[space.Space](r).Set(slipper, space.Space{
		Kind:   space.Rigid,
		Extent: units.Rect{Min: units.V2(-5*cm, -12*cm), Max: units.V2(5*cm, 12*cm)},
	})
	cs.Location.Set(slipper, comp.Location{Space: room, Pos: units.V3(120*cm, 100*cm, 0)})
	ecs.Get[attach.Garment](r).Set(slipper, attach.Garment{
		Owner:     giantess,
		Slot:      "slipper_left",
		State:     attach.GarmentStatePlaced,
		MountPart: foot,
		InnerPressure: &attach.InnerPressure{
			Zone:  units.Rect{Min: units.V2(-4*cm, -8*cm), Max: units.V2(4*cm, 7*cm)},
			Force: 50,
		},
	})

	spawnTiny := func(name string, x, y units.Length) ids.ID {
		e := r.Spawn()
		cs.Name.Set(e, comp.Name(name))
		cs.Location.Set(e, comp.Location{Space: slipper, Pos: units.V3(x, y, 0)})
		cs.Body.Set(e, comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter})
		ecs.Get[crush.Fragile](r).Set(e, crush.Fragile{})
		return e
	}

	victim := spawnTiny("victim", 0, 0)        // en la plantilla
	toeSafe := spawnTiny("toe_safe", 0, 10*cm) // en la puntera segura
	sheltered := spawnTiny("sheltered", 0, -5*cm)

	sh := r.Spawn()
	cs.Location.Set(sh, comp.Location{Space: slipper, Pos: units.V3(0, -5*cm, 0)})
	ecs.Get[crush.Shelter](r).Set(sh, crush.Shelter{Strength: 60, Radius: 2 * cm})

	if _, err := w.Enqueue(attach.DonGarment{Actor: giantess, Garment: slipper, WarnTicks: 3}); err != nil {
		t.Fatal(err)
	}
	w.Step()
	threats := crush.Threats(w)
	if len(threats) != 1 || threats[0].Surface != slipper || threats[0].Source != crush.SourceGarment {
		t.Fatalf("esperaba 1 amenaza en la prenda durante donning, obtuve %+v", threats)
	}

	var crushed, saved, worn bool
	for range 3 {
		w.Step()
		for _, ev := range w.Events() {
			switch e := ev.Event.(type) {
			case crush.Crushed:
				if e.Entity == victim && e.Source == crush.SourceGarment {
					crushed = true
				}
			case crush.Sheltered:
				if e.Entity == sheltered {
					saved = true
				}
			case attach.GarmentWorn:
				if e.Garment == slipper {
					worn = true
				}
			}
		}
	}
	if !worn || !crushed || !saved {
		t.Fatalf("resultados inesperados: worn=%v crushed=%v saved=%v", worn, crushed, saved)
	}
	if r.Alive(victim) {
		t.Fatal("victim en InnerPressure debía morir aplastado")
	}
	if !r.Alive(toeSafe) || !r.Alive(sheltered) {
		t.Fatal("toeSafe y sheltered debían sobrevivir al calce")
	}
}
