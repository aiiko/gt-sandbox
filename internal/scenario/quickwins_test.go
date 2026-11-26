package scenario_test

import (
	"bytes"
	"path/filepath"
	"slices"
	"testing"

	"gtsandbox/internal/scenario"
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
