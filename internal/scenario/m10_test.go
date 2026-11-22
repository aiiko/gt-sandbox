package scenario_test

import (
	"encoding/json"
	"testing"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/ecs"
)

func at0(kind, data string) func(*scenario.Spec) {
	return func(s *scenario.Spec) {
		s.Commands = append(s.Commands, scenario.CommandSpec{Tick: 0, Kind: kind, Data: json.RawMessage(data)})
	}
}

// El tiny recuerda el escritorio nuevo; sin curiosidad no sube a la mano.
func TestTinyRidesHand(t *testing.T) {
	for _, curious := range []bool{true, false} {
		edit := at0("set_components", `{"entity": "$tiny", "components": {"agent": {"scale": "tiny", "sight": 200000}}}`)
		if curious {
			edit = nil
		}
		in := run(t, "tiny_rides_hand.json", edit)
		a, _ := ecs.Get[ai.Agent](in.World.ECS()).Get(in.Refs["tiny"])
		if got := a.HasVisited(in.Refs["desk2"]); got != curious {
			t.Errorf("curioso=%v: visitó desk2 = %v (%v)", curious, got, a.Visited)
		}
	}
}

// Con hambre, el desayuno empieza cortando la escritura, pero una vez con la
// taza en la mano no lo corta volver a tener ganas de escribir.
func TestBreakfastCommitted(t *testing.T) {
	in := run(t, "tiny_rides_hand.json", at0("set_need", `{"entity": "$giantess", "need": "hunger", "value": 450}`))
	for _, e := range in.Events {
		if ev, ok := e.Event.(ai.ActionInterrupted); ok && ev.What == "breakfast" {
			t.Fatalf("tick %d: %+v", e.Tick, ev)
		}
	}
}
