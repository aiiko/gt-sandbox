package ai_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

func setupGarmentAIWorld(t *testing.T, care float64) (*world.World, ids.ID, ids.ID, ids.ID, ids.ID) {
	t.Helper()
	w := engine.MustNew(world.Config{Seed: 7})
	r := w.ECS()
	cs := comp.Of(r)
	cm := units.Centimeter

	room := r.Spawn()
	cs.Name.Set(room, "bedroom")
	ecs.Get[space.Space](r).Set(room, space.Space{Kind: space.Static, Extent: units.Rect{Max: units.V2(300*cm, 300*cm)}})
	gRoom, _ := nav.NewGrid(units.Vec2{}, 10*cm, 30, 30)
	ecs.Get[nav.Grid](r).Set(room, *gRoom)

	giantess := r.Spawn()
	cs.Name.Set(giantess, "giantess")
	ecs.Get[space.Space](r).Set(giantess, space.Space{Kind: space.Body})
	cs.Location.Set(giantess, comp.Location{Space: room, Pos: units.V3(100*cm, 100*cm, 0)})
	ecs.Get[nav.Mover](r).Set(giantess, nav.Mover{Speed: 100 * cm, Scale: "giantess"})
	ecs.Get[ai.Agent](r).Set(giantess, ai.Agent{
		Scale: "giantess",
		Care:  care,
		Watch: 200 * cm,
	})
	ecs.Get[needs.Needs](r).Set(giantess, needs.Needs{{Name: "energy", Value: 200}, {Name: "hunger", Value: 900}})

	foot := r.Spawn()
	cs.Name.Set(foot, "foot_l")
	ecs.Get[space.Space](r).Set(foot, space.Space{Kind: space.Articulated, Extent: units.Rect{Max: units.V2(10*cm, 24*cm)}})
	cs.Location.Set(foot, comp.Location{Space: giantess, Pos: units.V3(-10*cm, 0, 2*cm)})

	slipper := r.Spawn()
	cs.Name.Set(slipper, "pantufla izquierda")
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
		WornPos:   units.V3(5*cm, 12*cm, -cm),
		InnerPressure: &attach.InnerPressure{
			Zone:  units.Rect{Min: units.V2(-4*cm, -8*cm), Max: units.V2(4*cm, 7*cm)},
			Force: 50,
		},
		DonWarnTicks: 10,
	})

	bed := r.Spawn()
	cs.Name.Set(bed, "bed")
	cs.Location.Set(bed, comp.Location{Space: room, Pos: units.V3(220*cm, 100*cm, 0)})
	ecs.Get[ai.Offers](r).Set(bed, ai.Offers{{
		Action:  "sleep",
		Scale:   "giantess",
		Effects: []ai.Effect{{Need: "energy", Amount: 500}},
		Steps: []ai.Step{
			{Do: ai.StepGoto},
			{Do: ai.StepDoff, Slot: "slipper_left", Offset: units.V2(-20*cm, -20*cm)},
			{Do: ai.StepUse, Ticks: 20},
			{Do: ai.StepDon, Slot: "slipper_left"},
		},
	}})

	return w, room, giantess, slipper, bed
}

func TestStepDonDoffAndNoOp(t *testing.T) {
	w, room, giantess, slipper, bed := setupGarmentAIWorld(t, 0)
	r := w.ECS()

	// Agregar un segundo slot inexistente en otra oferta no debe bloquear.
	os := ecs.Get[ai.Offers](r).Ptr(bed)
	(*os)[0].Steps = append((*os)[0].Steps, ai.Step{Do: ai.StepDon, Slot: "missing_slot"})

	var sawWorn, sawDoffed, sawDone bool
	for range 200 {
		w.Step()
		for _, ev := range w.Events() {
			switch e := ev.Event.(type) {
			case attach.GarmentWorn:
				if e.Garment == slipper {
					sawWorn = true
				}
			case attach.GarmentDoffed:
				if e.Garment == slipper {
					sawDoffed = true
				}
			case ai.ActionDone:
				if e.Entity == giantess && e.What == "sleep" {
					sawDone = true
				}
			}
		}
		if sawDone {
			break
		}
	}
	if !sawDone || !sawWorn {
		t.Fatalf("esperaba completar sleep con don/no-op: done=%v worn=%v doffed=%v", sawDone, sawWorn, sawDoffed)
	}
	// Ahora la pantufla está puesta: volver a dormir debe sacársela y volvérsela a poner.
	_ = (needs.SetNeed{Entity: giantess, Need: "energy", Value: 100}).Apply(w)
	sawDoffed, sawDone = false, false
	for range 200 {
		w.Step()
		for _, ev := range w.Events() {
			switch e := ev.Event.(type) {
			case attach.GarmentDoffed:
				if e.Garment == slipper && e.Surface == room {
					sawDoffed = true
				}
			case ai.ActionDone:
				if e.Entity == giantess && e.What == "sleep" {
					sawDone = true
				}
			}
		}
		if sawDone {
			break
		}
	}
	if !sawDoffed || !sawDone {
		t.Fatalf("esperaba doff + re-don al dormir: doffed=%v done=%v", sawDoffed, sawDone)
	}
}

func TestStepDonCareHesitatesAndUrgentInterruptCancels(t *testing.T) {
	w, _, giantess, slipper, bed := setupGarmentAIWorld(t, 1.0)
	r := w.ECS()
	cs := comp.Of(r)
	cm := units.Centimeter

	// Colocar un tiny dentro de la pantufla.
	tiny := r.Spawn()
	cs.Name.Set(tiny, "tiny")
	cs.Location.Set(tiny, comp.Location{Space: slipper, Pos: units.V3(0, 0, 0)})
	cs.Body.Set(tiny, comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter})
	ecs.Get[crush.Fragile](r).Set(tiny, crush.Fragile{})

	// Oferta que calza la pantufla antes de use, para probar cuidado + cancelación.
	ecs.Get[ai.Offers](r).Set(bed, ai.Offers{{
		Action:  "prep",
		Scale:   "giantess",
		Effects: []ai.Effect{{Need: "energy", Amount: 200}},
		Steps: []ai.Step{
			{Do: ai.StepDon, Slot: "slipper_left"},
			{Do: ai.StepUse, Ticks: 5},
		},
	}})

	// Agregar un objeto de comida urgente para interrumpir durante donning.
	food := r.Spawn()
	cs.Name.Set(food, "snack")
	cs.Location.Set(food, comp.Location{Space: cs.Location.Ptr(giantess).Space, Pos: units.V3(100*cm, 100*cm, 0)})
	ecs.Get[ai.Offers](r).Set(food, ai.Offers{{
		Action:   "eat",
		Scale:    "giantess",
		Effects:  []ai.Effect{{Need: "hunger", Amount: 900}},
		Duration: 10,
	}})

	var sawHesitated, sawDonning, sawCancelled bool
	for i := range 220 {
		// Una vez iniciada la ventana donning, bajar hunger para disparar interrupción urgente.
		if sawDonning && !sawCancelled {
			_ = (needs.SetNeed{Entity: giantess, Need: "hunger", Value: 10}).Apply(w)
		}
		w.Step()
		for _, ev := range w.Events() {
			switch e := ev.Event.(type) {
			case ai.Hesitated:
				if e.Entity == giantess && e.Surface == slipper {
					sawHesitated = true
				}
			case attach.GarmentDonning:
				if e.Garment == slipper {
					sawDonning = true
					if e.Until-w.Tick() <= 10 {
						t.Fatalf("esperaba warnTicks extendido por care, obtuvo %v", e.Until-w.Tick())
					}
				}
			case attach.GarmentDonCancelled:
				if e.Garment == slipper {
					sawCancelled = true
				}
			}
		}
		_ = clock.Tick(i)
		if sawCancelled {
			break
		}
	}
	if !sawHesitated || !sawDonning || !sawCancelled {
		t.Fatalf("esperaba hesitated=%v donning=%v cancelled=%v", sawHesitated, sawDonning, sawCancelled)
	}
	if g, _ := ecs.Get[attach.Garment](r).Get(slipper); g.State != attach.GarmentStatePlaced {
		t.Fatalf("tras interrupción la prenda quedó en %q", g.State)
	}
}
