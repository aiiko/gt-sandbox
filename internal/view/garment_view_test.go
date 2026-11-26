package view_test

import (
	"slices"
	"strings"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
	"gtsandbox/internal/view"
)

func TestGarmentFrameAndChronicleMilestones(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 1})
	r := w.ECS()
	cs := comp.Of(r)
	cm := units.Centimeter

	house := r.Spawn()
	cs.Name.Set(house, "house")
	ecs.Get[space.Space](r).Set(house, space.Space{Kind: space.Static})

	bedroom := r.Spawn()
	cs.Name.Set(bedroom, "dormitorio")
	ecs.Get[space.Space](r).Set(bedroom, space.Space{Kind: space.Static, Extent: units.Rect{Max: units.V2(300*cm, 300*cm)}})
	cs.Location.Set(bedroom, comp.Location{Space: house})

	bathroom := r.Spawn()
	cs.Name.Set(bathroom, "baño")
	ecs.Get[space.Space](r).Set(bathroom, space.Space{Kind: space.Static, Extent: units.Rect{Max: units.V2(300*cm, 300*cm)}})
	cs.Location.Set(bathroom, comp.Location{Space: house, Pos: units.V3(300*cm, 0, 0)})

	giantess := r.Spawn()
	cs.Name.Set(giantess, "Giantess")
	ecs.Get[space.Space](r).Set(giantess, space.Space{Kind: space.Body})
	cs.Location.Set(giantess, comp.Location{Space: bedroom, Pos: units.V3(100*cm, 100*cm, 0)})
	ecs.Get[ai.Agent](r).Set(giantess, ai.Agent{Scale: "giantess"})

	foot := r.Spawn()
	cs.Name.Set(foot, "pie izquierdo")
	ecs.Get[space.Space](r).Set(foot, space.Space{Kind: space.Articulated, Extent: units.Rect{Max: units.V2(10*cm, 24*cm)}})
	cs.Location.Set(foot, comp.Location{Space: giantess, Pos: units.V3(-10*cm, 0, 2*cm)})

	slipper := r.Spawn()
	cs.Name.Set(slipper, "pantufla izquierda")
	ecs.Get[space.Space](r).Set(slipper, space.Space{
		Kind:   space.Rigid,
		Extent: units.Rect{Min: units.V2(-5*cm, -12*cm), Max: units.V2(5*cm, 12*cm)},
	})
	cs.Location.Set(slipper, comp.Location{Space: bedroom, Pos: units.V3(120*cm, 100*cm, 0)})
	ecs.Get[attach.Garment](r).Set(slipper, attach.Garment{
		Owner:        giantess,
		Slot:         "slipper_left",
		State:        attach.GarmentStatePlaced,
		MountPart:    foot,
		DonWarnTicks: 10,
		InnerPressure: &attach.InnerPressure{
			Zone:  units.Rect{Min: units.V2(-4*cm, -8*cm), Max: units.V2(4*cm, 7*cm)},
			Force: 50,
		},
	})

	portalOut := r.Spawn()
	ecs.Get[nav.Portal](r).Set(portalOut, nav.Portal{
		A: nav.PortalEnd{Space: slipper, Pos: units.V2(0, -10*cm)},
		B: nav.PortalEnd{Space: bedroom, Pos: units.V2(120*cm, 90*cm)},
	})

	rider := r.Spawn()
	cs.Name.Set(rider, "Tiny 04")
	cs.Location.Set(rider, comp.Location{Space: slipper, Pos: units.V3(0, 9*cm, 0)})
	cs.Body.Set(rider, comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter})
	ecs.Get[ai.Agent](r).Set(rider, ai.Agent{Scale: "tiny"})
	ecs.Get[crush.Fragile](r).Set(rider, crush.Fragile{})

	escaper := r.Spawn()
	cs.Name.Set(escaper, "Tiny 02")
	cs.Location.Set(escaper, comp.Location{Space: slipper, Pos: units.V3(0, -9*cm, 0)})
	cs.Body.Set(escaper, comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter})
	ecs.Get[ai.Agent](r).Set(escaper, ai.Agent{Scale: "tiny"})
	ecs.Get[nav.Mover](r).Set(escaper, nav.Mover{Speed: 40 * cm, Scale: "tiny"})

	victim := r.Spawn()
	cs.Name.Set(victim, "Tiny 07")
	cs.Location.Set(victim, comp.Location{Space: slipper, Pos: units.V3(0, 0, 0)})
	cs.Body.Set(victim, comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter})
	ecs.Get[ai.Agent](r).Set(victim, ai.Agent{Scale: "tiny"})
	ecs.Get[crush.Fragile](r).Set(victim, crush.Fragile{})

	run := view.NewRunner(w, func() error { w.Step(); return nil })
	if sp, ok := run.Cur.SpaceByID(slipper); !ok || sp.GarmentState != "placed" {
		t.Fatalf("esperaba garment_state placed, obtuve %+v", sp)
	}

	_ = run.Enqueue(attach.DonGarment{Actor: giantess, Garment: slipper, WarnTicks: 10})
	_ = run.Enqueue(nav.MoveTo{Entity: escaper, Space: bedroom, Goal: units.V2(120*cm, 90*cm)})
	run.StepOnce()
	run.StepOnce()
	if sp, ok := run.Cur.SpaceByID(slipper); !ok || sp.GarmentState != "donning" || len(run.Cur.Threats) != 1 {
		t.Fatalf("esperaba garment_state donning y 1 threat, obtuve state=%q threats=%d", sp.GarmentState, len(run.Cur.Threats))
	}

	for range 12 {
		run.StepOnce()
	}
	_, _ = space.Of(r).Reparent(giantess, bathroom)
	_ = run.Enqueue(attach.DoffGarment{Actor: giantess, Garment: slipper, TargetSpace: bathroom, Pos: units.V3(100*cm, 100*cm, 0)})
	run.StepOnce()

	var texts []string
	for _, line := range run.Chronicle.Lines() {
		texts = append(texts, line.Text)
	}
	for _, want := range []string{
		"Tiny 02 salió de la pantufla izquierda justo antes de que se la pusiera.",
		"Tiny 07 quedó dentro de la pantufla izquierda.",
		"Tiny 04 viajó en la pantufla izquierda hasta el baño.",
	} {
		if !slices.ContainsFunc(texts, func(s string) bool { return strings.Contains(s, want) }) {
			t.Errorf("falta hito %q en crónica: %v", want, texts)
		}
	}
	_ = ids.Nil
}
