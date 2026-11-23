package scenario_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/needs"
)

// M42: una conversación completada sube la diversión de los dos
// participantes en Gain, en el mismo tick.
func TestChatRaisesFun(t *testing.T) {
	spec, err := scenario.ReadFile("../../tests/scenarios/tinies_chat.json")
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
	fun := func(e ids.ID) int32 {
		n, _ := ecs.Get[needs.Needs](in.World.ECS()).Get(e)
		v, _ := n.Get(colony.ChatNeed)
		return v
	}
	a, b := in.Refs["a"], in.Refs["b"]
	before := map[ids.ID]int32{a: fun(a), b: fun(b)}
	for !p.Done() {
		if err := p.Step(); err != nil {
			t.Fatal(err)
		}
		for _, e := range in.World.Events() {
			if _, ok := e.Event.(colony.ChatDone); ok {
				for _, x := range []ids.ID{a, b} {
					if got := fun(x); got != before[x]+colony.DefaultChatGain {
						t.Fatalf("diversión de %v: %d → %d", x, before[x], got)
					}
				}
				return
			}
		}
	}
	t.Fatal("no conversaron")
}
