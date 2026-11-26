package scenario_test

import (
	"bytes"
	"path/filepath"
	"slices"
	"testing"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/encounter"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/world"
)

func verifySnapshotContinuity(t *testing.T, scenFile string, saveTick, afterTicks int) {
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
	specB := *spec
	specB.Commands = nil
	for _, c := range spec.Commands {
		if c.Tick >= saveTick {
			c.Tick -= saveTick
			specB.Commands = append(specB.Commands, c)
		}
	}
	inB := &scenario.Instance{Spec: &specB, World: wB, Refs: inA.Refs}
	pB, err := inB.Player()
	if err != nil {
		t.Fatal(err)
	}
	for range afterTicks {
		if err := pA.Step(); err != nil {
			t.Fatal(err)
		}
		if err := pB.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if inA.World.Hash() != wB.Hash() {
		t.Fatalf("%s: continuidad rota tras snapshot en tick %d: %x vs %x", scenFile, saveTick, inA.World.Hash(), wB.Hash())
	}
	if err := wB.CheckInvariants(); err != nil {
		t.Fatalf("%s: invariante rota tras load: %v", scenFile, err)
	}
}

func TestQW1CarryPassengerSnapshotAndOutcome(t *testing.T) {
	for _, tc := range []struct {
		file       string
		saveTick   int
		afterTicks int
	}{
		{"giantess_carries_passenger_to_desk.json", 200, 400},
		{"giantess_cancels_carry_when_occupied.json", 180, 220},
	} {
		t.Run(tc.file, func(t *testing.T) {
			verifySnapshotContinuity(t, tc.file, tc.saveTick, tc.afterTicks)
		})
	}

	spec, err := scenario.ReadFile("../../tests/scenarios/giantess_carries_passenger_to_desk.json")
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
	var delivered []encounter.PassengerDelivered
	for range spec.Ticks {
		if err := p.Step(); err != nil {
			t.Fatal(err)
		}
		for _, ev := range in.World.Events() {
			if d, ok := ev.Event.(encounter.PassengerDelivered); ok {
				delivered = append(delivered, d)
			}
		}
	}
	if len(delivered) != 1 || !slices.Contains(delivered[0].Passengers, in.Refs["passenger"]) {
		t.Fatalf("esperaba entrega de passenger %v, vino %+v", in.Refs["passenger"], delivered)
	}
}

func TestQW2GarmentsAndGenkanSnapshot(t *testing.T) {
	for _, tc := range []struct {
		file       string
		saveTick   int
		afterTicks int
	}{
		{"genkan_expedition_and_loot.json", 160, 400},
		{"sock_donning_and_hair_tie_shelter.json", 18, 120},
	} {
		t.Run(tc.file, func(t *testing.T) {
			verifySnapshotContinuity(t, tc.file, tc.saveTick, tc.afterTicks)
		})
	}
}

func TestQW3TickleSnapshot(t *testing.T) {
	verifySnapshotContinuity(t, "tiny_tickles_giantess_neck.json", 15, 90)
}

func TestQW4BondSnapshotAndTrustEvolution(t *testing.T) {
	for _, tc := range []struct {
		file       string
		saveTick   int
		afterTicks int
	}{
		{"tinies_react_to_observe_by_trust.json", 60, 120},
		{"bond_evolves_from_share_and_near_miss.json", 100, 180},
	} {
		t.Run(tc.file, func(t *testing.T) {
			verifySnapshotContinuity(t, tc.file, tc.saveTick, tc.afterTicks)
		})
	}

	in := run(t, "bond_evolves_from_share_and_near_miss.json", nil)
	r := in.World.ECS()
	bonds := ecs.Get[encounter.Bond](r)
	if b, ok := bonds.Get(in.Refs["desk_tiny"]); !ok || b.Trust <= 10 {
		t.Fatalf("desk_tiny debía subir su confianza tras shared (>10), tiene %+v (ok=%v)", b, ok)
	}
	if b, ok := bonds.Get(in.Refs["floor_tiny"]); !ok || b.Trust >= 10 {
		t.Fatalf("floor_tiny debía bajar su confianza tras near_miss (<10), tiene %+v (ok=%v)", b, ok)
	}
}
