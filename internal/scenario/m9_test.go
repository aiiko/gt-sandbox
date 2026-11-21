package scenario_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/needs"
)

// Aceptación de M9: no hay trabajos antes de que alguien vea comida, el
// depósito crece y en 10 minutos ningún tiny llega a hambre 0.
func TestColonyFeeds(t *testing.T) {
	spec, err := scenario.ReadFile("../../tests/scenarios/colony_gathers.json")
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
	knows := map[[2]ids.ID]bool{} // lo que ya conocía cada uno al empezar
	ecs.Get[ai.Agent](in.World.ECS()).Each(func(e ids.ID, a *ai.Agent) {
		for _, k := range a.Known {
			knows[[2]ids.ID{e, k.Entity}] = true
		}
	})
	minHunger := int32(needs.Max)
	for !p.Done() {
		if err := p.Step(); err != nil {
			t.Fatal(err)
		}
		ecs.Get[colony.Member](in.World.ECS()).Each(func(e ids.ID, _ *colony.Member) {
			n, _ := ecs.Get[needs.Needs](in.World.ECS()).Get(e)
			v, _ := n.Get("hunger")
			minHunger = min(minHunger, v)
		})
	}
	food := map[ids.ID]bool{}
	stored := 0
	for _, e := range in.Events {
		switch ev := e.Event.(type) {
		case ai.Noticed:
			food[ev.Target] = true
			knows[[2]ids.ID{ev.Entity, ev.Target}] = true
		case ai.ActionStarted:
			if ev.Action == "use" && !knows[[2]ids.ID{ev.Entity, ev.Target}] {
				t.Fatalf("tick %d: %v usa %v sin conocerlo", e.Tick, ev.Entity, ev.Target)
			}
		case jobs.JobPosted:
			if !food[ev.Target] {
				t.Fatalf("tick %d: trabajo sobre %v antes de verlo", e.Tick, ev.Target)
			}
		case colony.Stored:
			stored++
		}
	}
	if minHunger <= 0 || stored < 2 {
		t.Fatalf("hambre mínima %d, guardados %d", minHunger, stored)
	}
	t.Logf("hambre mínima %d, guardados %d", minHunger, stored)
}
