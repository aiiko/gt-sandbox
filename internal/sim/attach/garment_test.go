package attach_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

func newGarmentTestWorld(t *testing.T) *world.World {
	t.Helper()
	w := world.New(world.Config{Seed: 1})
	if _, err := comp.Register(w.ECS()); err != nil {
		t.Fatal(err)
	}
	for _, fn := range []func(*world.World) error{space.Install, nav.Install, attach.Install} {
		if err := fn(w); err != nil {
			t.Fatal(err)
		}
	}
	return w
}

func spawnSpace(w *world.World, name string, kind space.Kind, parent ids.ID, pos units.Vec3, ext units.Rect) ids.ID {
	r := w.ECS()
	cs := comp.Of(r)
	id := r.Spawn()
	cs.Name.Set(id, comp.Name(name))
	ecs.Get[space.Space](r).Set(id, space.Space{Kind: kind, Extent: ext})
	if !parent.IsNil() {
		cs.Location.Set(id, comp.Location{Space: parent, Pos: pos})
	}
	return id
}

func TestGarmentLifecycleAndPortals(t *testing.T) {
	w := newGarmentTestWorld(t)
	r := w.ECS()
	cs := comp.Of(r)
	cm := units.Centimeter

	room := spawnSpace(w, "bedroom", space.Static, ids.Nil, units.Vec3{}, units.Rect{Max: units.V2(300*cm, 300*cm)})
	giantess := spawnSpace(w, "giantess", space.Body, room, units.V3(100*cm, 100*cm, 0), units.Rect{})
	foot := spawnSpace(w, "foot_l", space.Articulated, giantess, units.V3(-15*cm, 0, 2*cm), units.Rect{Max: units.V2(10*cm, 24*cm)})
	ankle := spawnSpace(w, "calf_l", space.Articulated, giantess, units.V3(-15*cm, 0, 10*cm), units.Rect{Max: units.V2(8*cm, 16*cm)})

	slipper := spawnSpace(w, "slipper", space.Rigid, room, units.V3(120*cm, 100*cm, 0), units.Rect{
		Min: units.V2(-5*cm, -12*cm),
		Max: units.V2(5*cm, 12*cm),
	})
	portalID := r.Spawn()
	ecs.Get[nav.Portal](r).Set(portalID, nav.Portal{
		A:        nav.PortalEnd{Space: slipper, Pos: units.V2(0, -10*cm)},
		B:        nav.PortalEnd{Space: ankle, Pos: units.V2(4*cm, 2*cm)},
		Scale:    "tiny",
		Disabled: true,
	})
	ecs.Get[attach.Garment](r).Set(slipper, attach.Garment{
		Owner:       giantess,
		Slot:        "slipper_left",
		State:       attach.GarmentStatePlaced,
		MountPart:   foot,
		WornPos:     units.V3(5*cm, 12*cm, -cm),
		BodyPortals: []ids.ID{portalID},
	})

	// Tiny adentro en la puntera segura: debe conservar sus coordenadas locales.
	tiny := r.Spawn()
	localPos := units.V3(1*cm, 9*cm, 0)
	cs.Location.Set(tiny, comp.Location{Space: slipper, Pos: localPos})
	cs.Body.Set(tiny, comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter})

	// Validación atómica de DoffGarment cuando aún está placed.
	if err := (attach.DoffGarment{Actor: giantess, Garment: slipper, TargetSpace: room}).Apply(w); err == nil {
		t.Fatal("esperaba error al hacer doff de prenda placed")
	}

	// DonGarment con actor ajeno falla sin mutar nada.
	if err := (attach.DonGarment{Actor: foot, Garment: slipper}).Apply(w); err == nil {
		t.Fatal("esperaba error con actor distinto de owner")
	}
	if g, _ := ecs.Get[attach.Garment](r).Get(slipper); g.State != attach.GarmentStatePlaced {
		t.Fatalf("don fallido mutó el estado a %q", g.State)
	}

	// CancelDon funciona en donning y vuelve a placed.
	if _, err := w.Enqueue(attach.DonGarment{Actor: giantess, Garment: slipper, WarnTicks: 10}); err != nil {
		t.Fatal(err)
	}
	w.Step()
	if g, _ := ecs.Get[attach.Garment](r).Get(slipper); g.State != attach.GarmentStateDonning {
		t.Fatalf("esperaba estado donning, obtuve %q", g.State)
	}
	if _, err := w.Enqueue(attach.CancelDon{Actor: giantess, Garment: slipper}); err != nil {
		t.Fatal(err)
	}
	w.Step()
	if g, _ := ecs.Get[attach.Garment](r).Get(slipper); g.State != attach.GarmentStatePlaced || g.StateUntil != 0 {
		t.Fatalf("cancel_don no devolvió a placed: %+v", g)
	}

	// DonGarment hasta completar -> worn, reparented, portal habilitado, tiny intacto.
	if _, err := w.Enqueue(attach.DonGarment{Actor: giantess, Garment: slipper, WarnTicks: 5}); err != nil {
		t.Fatal(err)
	}
	w.StepN(6)
	if err := w.CheckInvariants(); err != nil {
		t.Fatal(err)
	}
	g, _ := ecs.Get[attach.Garment](r).Get(slipper)
	if g.State != attach.GarmentStateWorn {
		t.Fatalf("esperaba estado worn, obtuve %q", g.State)
	}
	if l, _ := cs.Location.Get(slipper); l.Space != foot {
		t.Fatalf("prenda worn en %v, esperaba %v", l.Space, foot)
	}
	if p, _ := ecs.Get[nav.Portal](r).Get(portalID); p.Disabled {
		t.Fatal("el portal corporal no se habilitó al pasar a worn")
	}
	if tl, _ := cs.Location.Get(tiny); tl.Space != slipper || tl.Pos != localPos {
		t.Fatalf("el tiny perdió sus coordenadas locales: %+v", tl)
	}

	// DoffGarment -> placed en room, portal corporal deshabilitado.
	if _, err := w.Enqueue(attach.DoffGarment{
		Actor:       giantess,
		Garment:     slipper,
		TargetSpace: room,
		Pos:         units.V3(200*cm, 150*cm, 0),
	}); err != nil {
		t.Fatal(err)
	}
	w.Step()
	if err := w.CheckInvariants(); err != nil {
		t.Fatal(err)
	}
	g, _ = ecs.Get[attach.Garment](r).Get(slipper)
	if g.State != attach.GarmentStatePlaced {
		t.Fatalf("esperaba estado placed tras doff, obtuve %q", g.State)
	}
	if p, _ := ecs.Get[nav.Portal](r).Get(portalID); !p.Disabled {
		t.Fatal("el portal corporal no se deshabilitó tras doff")
	}
}

func TestGarmentDuplicateSlotAndGripAndCleanup(t *testing.T) {
	w := newGarmentTestWorld(t)
	r := w.ECS()
	cs := comp.Of(r)
	cm := units.Centimeter

	room := spawnSpace(w, "bedroom", space.Static, ids.Nil, units.Vec3{}, units.Rect{Max: units.V2(300*cm, 300*cm)})
	giantess := spawnSpace(w, "giantess", space.Body, room, units.V3(100*cm, 100*cm, 0), units.Rect{})
	foot := spawnSpace(w, "foot_l", space.Articulated, giantess, units.V3(-15*cm, 0, 2*cm), units.Rect{Max: units.V2(10*cm, 24*cm)})

	ext := units.Rect{Min: units.V2(-5*cm, -12*cm), Max: units.V2(5*cm, 12*cm)}
	s1 := spawnSpace(w, "s1", space.Rigid, room, units.V3(110*cm, 100*cm, 0), ext)
	s2 := spawnSpace(w, "s2", space.Rigid, room, units.V3(130*cm, 100*cm, 0), ext)
	ecs.Get[attach.BodyZone](r).Set(s1, attach.BodyZone{
		Body:        giantess,
		RequireGrip: true,
		Grips:       []attach.GripPoint{{Pos: units.V2(0, 9*cm), Radius: 3 * cm, Label: "puntera"}},
	})
	ecs.Get[attach.Garment](r).Set(s1, attach.Garment{
		Owner: giantess, Slot: "slipper_left", State: attach.GarmentStatePlaced, MountPart: foot,
		WornProjection: &space.Projection{
			AnchorA: units.V3(0, -12*cm, 0),
			AnchorB: units.V3(0, 12*cm, 12*cm),
		},
	})
	ecs.Get[attach.Garment](r).Set(s2, attach.Garment{
		Owner: giantess, Slot: "slipper_left", State: attach.GarmentStatePlaced, MountPart: foot,
	})

	holder := r.Spawn()
	cs.Location.Set(holder, comp.Location{Space: s1, Pos: units.V3(0, 9*cm, 0)})
	cs.Body.Set(holder, comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter})
	dropper := r.Spawn()
	cs.Location.Set(dropper, comp.Location{Space: s1, Pos: units.V3(0, 0, 0)})
	cs.Body.Set(dropper, comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter})

	if _, err := w.Enqueue(attach.DonGarment{Actor: giantess, Garment: s1, WarnTicks: 2}); err != nil {
		t.Fatal(err)
	}
	w.Step()
	// Mientras s1 está donning, s2 en el mismo slot debe rechazarse.
	if err := (attach.DonGarment{Actor: giantess, Garment: s2}).Apply(w); err == nil {
		t.Fatal("esperaba rechazo por slot duplicado")
	}
	w.StepN(2)
	if !ecs.Get[attach.Gripping](r).Has(holder) {
		t.Fatal("holder junto al GripPoint debía quedar agarrado al inclinarse la prenda")
	}
	if l, _ := cs.Location.Get(dropper); l.Space == s1 {
		t.Fatal("dropper sin agarre debía caer de la prenda inclinada")
	}

	// Si MountPart se destruye, la prenda pasa a placed y pasa las invariantes.
	r.Despawn(foot)
	w.Step()
	if err := w.CheckInvariants(); err != nil {
		t.Fatal(err)
	}
	if g, _ := ecs.Get[attach.Garment](r).Get(s1); g.State != attach.GarmentStatePlaced {
		t.Fatalf("esperaba que la prenda vuelva a placed si muere MountPart, obtuve %q", g.State)
	}
}
