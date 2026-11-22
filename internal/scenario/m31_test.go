package scenario_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/encounter"
	"gtsandbox/internal/sim/jobs"
)

// Criterio 15.3: cada porción que la giantess deja al compartir se puede
// seguir hasta que la colonia la guarda (shared → noticed → haul → stored),
// y la historia la distingue de las migajas de una comida (motivo).
func TestSharedCrumbsAreTraceable(t *testing.T) {
	s, err := scenario.ReadFile("../../tests/scenarios/giantess_shares_crumbs.json")
	if err != nil {
		t.Fatal(err)
	}
	in, err := scenario.Build(s)
	if err != nil {
		t.Fatal(err)
	}
	playChecked(t, in, 113, nil)
	var shared []ids.ID
	noticed, hauled, stored := map[ids.ID]bool{}, map[ids.ID]bool{}, map[ids.ID]bool{}
	for _, e := range in.Events {
		switch ev := e.Event.(type) {
		case encounter.Shared:
			shared = append(shared, ev.Items...)
		case ai.ItemsLeft:
			if ev.Motive != "share" {
				t.Errorf("migajas con motivo %q en un escenario sin comidas", ev.Motive)
			}
		case ai.Noticed:
			noticed[ev.Target] = true
		case jobs.JobPosted:
			if ev.Type == "haul" {
				hauled[ev.Target] = true
			}
		case colony.Stored:
			stored[ev.Item] = true
		}
	}
	if len(shared) != encounter.ShareCount {
		t.Fatalf("compartió %d porciones", len(shared))
	}
	for _, it := range shared {
		if !noticed[it] || !hauled[it] || !stored[it] {
			t.Errorf("porción %s: noticed %v, haul %v, stored %v", it, noticed[it], hauled[it], stored[it])
		}
	}
}

// Con migajas suficientes a la vista, compartir no es opción.
func TestNoSharingWithCrumbsInSight(t *testing.T) {
	s, err := scenario.ReadFile("../../tests/scenarios/giantess_skips_sharing.json")
	if err != nil {
		t.Fatal(err)
	}
	for i := range s.Commands {
		s.Commands[i].Tick = 0 // las migajas están desde el principio
	}
	s.Ticks, s.Expect = 100, nil
	in, err := scenario.Build(s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := in.Run(); err != nil {
		t.Fatal(err)
	}
	for _, e := range in.Events {
		if a, ok := e.Event.(ai.ActionStarted); ok && a.Action == "share" {
			t.Fatalf("tick %d: compartió con migajas a la vista", e.Tick)
		}
	}
}
