package scenario_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
)

// M41: de la servilleta al refugio. Cada fibra cuenta una vez, la obra
// conserva su ID y recibe shelter, y al continuar desde snapshots (uno cae
// a mitad de la obra) se obtienen el mismo log y el mismo hash.
func TestColonyBuildsRefuge(t *testing.T) {
	spec, err := scenario.ReadFile("../../tests/scenarios/colony_builds_refuge.json")
	if err != nil {
		t.Fatal(err)
	}
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	site := in.Refs["site"]
	midway := false
	playChecked(t, in, 97, func() {
		if s, ok := ecs.Get[colony.BuildSite](in.World.ECS()).Get(site); ok && s.Delivered > 0 && s.Delivered < s.Required {
			midway = true
		}
	})
	if !midway {
		t.Fatal("ningún tick a mitad de la obra")
	}
	counted := map[ids.ID]bool{}
	for _, e := range in.Events {
		if d, ok := e.Event.(colony.FiberDelivered); ok {
			if counted[d.Item] {
				t.Fatalf("la fibra %v contó dos veces", d.Item)
			}
			counted[d.Item] = true
		}
	}
	s, ok := ecs.Get[colony.BuildSite](in.World.ECS()).Get(site)
	if !ok || !s.Done || s.Delivered != s.Required || len(counted) != int(s.Required) {
		t.Fatalf("obra %+v, fibras contadas %d", s, len(counted))
	}
	if !ecs.Get[crush.Shelter](in.World.ECS()).Has(site) {
		t.Fatal("el refugio terminado no protege")
	}
}
