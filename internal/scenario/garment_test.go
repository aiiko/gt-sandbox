package scenario_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/world"
)

func assertSnapshotContinuity(t *testing.T, scenFile string, saveTick, afterTicks int, checkAtSave func(*testing.T, *scenario.Instance)) {
	t.Helper()
	spec, err := scenario.ReadFile(filepath.Join("..", "..", "tests", "scenarios", scenFile))
	if err != nil {
		t.Fatal(err)
	}
	inA, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	pA, err := inA.Player()
	if err != nil {
		t.Fatal(err)
	}
	for range saveTick {
		if err := pA.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if checkAtSave != nil {
		checkAtSave(t, inA)
	}
	snap, err := inA.World.Save()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := world.WriteSnapshot(&buf, snap); err != nil {
		t.Fatal(err)
	}
	wB, err := engine.Load(&buf)
	if err != nil {
		t.Fatal(err)
	}
	for range afterTicks {
		if err := pA.Step(); err != nil {
			t.Fatal(err)
		}
		wB.Step()
	}
	if inA.World.Hash() != wB.Hash() {
		t.Fatalf("%s: continuidad rota tras snapshot en tick %d: %x vs %x", scenFile, saveTick, inA.World.Hash(), wB.Hash())
	}
	if err := wB.CheckInvariants(); err != nil {
		t.Fatalf("%s: invariante rota tras load: %v", scenFile, err)
	}
}

func TestGarmentSnapshotContinuity(t *testing.T) {
	t.Run("mid_donning", func(t *testing.T) {
		assertSnapshotContinuity(t, "garment_donning_flee_and_crush.json", 18, 60, func(t *testing.T, in *scenario.Instance) {
			g, ok := ecs.Get[attach.Garment](in.World.ECS()).Get(in.Refs["slipper_left"])
			if !ok || g.State != attach.GarmentStateDonning {
				t.Fatalf("esperaba prenda en donning al guardar, obtuve %+v", g)
			}
		})
	})
	t.Run("mid_worn_ride", func(t *testing.T) {
		assertSnapshotContinuity(t, "garment_ride_and_doff.json", 180, 120, func(t *testing.T, in *scenario.Instance) {
			g, ok := ecs.Get[attach.Garment](in.World.ECS()).Get(in.Refs["slipper_left"])
			if !ok || g.State != attach.GarmentStateWorn {
				t.Fatalf("esperaba prenda en worn al guardar, obtuve %+v", g)
			}
		})
	})
	t.Run("cardigan_pocket_ride", func(t *testing.T) {
		assertSnapshotContinuity(t, "cardigan_pocket_ride.json", 160, 140, func(t *testing.T, in *scenario.Instance) {
			g, ok := ecs.Get[attach.Garment](in.World.ECS()).Get(in.Refs["cardigan_pocket"])
			if !ok || g.State != attach.GarmentStateWorn {
				t.Fatalf("esperaba cárdigan en worn al guardar, obtuve %+v", g)
			}
		})
	})
}

func TestCasaGarmentsRandomizeAndStep(t *testing.T) {
	spec, err := scenario.ReadFile(filepath.Join("..", "..", "content", "partidas", "casa.json"))
	if err != nil {
		t.Fatal(err)
	}
	randSpec, err := scenario.Randomize(spec, 42)
	if err != nil {
		t.Fatal(err)
	}
	in, err := scenario.Build(randSpec)
	if err != nil {
		t.Fatal(err)
	}
	p, err := in.Player()
	if err != nil {
		t.Fatal(err)
	}
	for range 300 {
		if err := p.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if gStore := ecs.Get[attach.Garment](in.World.ECS()); gStore.Len() < 3 {
		t.Fatalf("esperaba >= 3 prendas en casa.json, hubo %d", gStore.Len())
	}
}
