package scenario_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/ecs"
)

// playSearches corre colony_searches_spots con el depósito inicial dado y
// devuelve cuántas búsquedas empezaron. En cada tick verifica que dos
// tinies no revisen el mismo punto.
func playSearches(t *testing.T, portions int32) int {
	t.Helper()
	spec, err := scenario.ReadFile("../../tests/scenarios/colony_searches_spots.json")
	if err != nil {
		t.Fatal(err)
	}
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	ecs.Get[ai.Portions](in.World.ECS()).Ptr(in.Refs["nest"]).Set(ai.ResourceFood, portions)
	p, err := in.Player()
	if err != nil {
		t.Fatal(err)
	}
	for !p.Done() {
		if err := p.Step(); err != nil {
			t.Fatal(err)
		}
		by := map[ids.ID]ids.ID{}
		ecs.Get[ai.Agent](in.World.ECS()).Each(func(e ids.ID, a *ai.Agent) {
			if a.Doing == nil || a.Doing.Action != colony.SearchAction {
				return
			}
			if o, dup := by[a.Doing.Target]; dup {
				t.Fatalf("tick %d: %v y %v revisan el mismo punto %v", p.Tick(), o, e, a.Doing.Target)
			}
			by[a.Doing.Target] = e
		})
	}
	n := 0
	for _, e := range in.Events {
		if s, ok := e.Event.(ai.ActionStarted); ok && s.Action == colony.SearchAction {
			n++
		}
	}
	return n
}

// M40: nunca dos tinies revisan el mismo punto.
func TestSearchSpotReservedOnce(t *testing.T) {
	if n := playSearches(t, 0); n < 2 {
		t.Fatalf("solo %d búsquedas", n)
	}
}

// M40: con la despensa en la meta se inician menos salidas de búsqueda que
// con el nido vacío.
func TestStockedColonySearchesLess(t *testing.T) {
	low, full := playSearches(t, 0), playSearches(t, 6)
	if full >= low {
		t.Fatalf("búsquedas: vacía %d, llena %d", low, full)
	}
	t.Logf("búsquedas: vacía %d, llena %d", low, full)
}
