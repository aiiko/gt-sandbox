package scenario_test

import (
	"encoding/json"
	"testing"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
)

// Si algo más urgente interrumpe el desayuno mientras come, la giantess
// levanta la mano (cierra el portal) y deja la taza a sus pies, sin reserva.
func TestBreakfastInterrupted(t *testing.T) {
	in := run(t, "giantess_breakfast.json", func(s *scenario.Spec) {
		s.Commands = append(s.Commands, scenario.CommandSpec{Tick: 250, Kind: "set_need",
			Data: json.RawMessage(`{"entity": "$giantess", "need": "energy", "value": 0}`)})
	})
	var cup, posed = in.Refs["giantess"], ""
	interrupted, dropped, closed := false, false, false
	for _, e := range in.Events {
		switch ev := e.Event.(type) {
		case ai.ActionInterrupted:
			interrupted = interrupted || ev.What == "breakfast"
		case ai.ItemDropped:
			cup, dropped = ev.Item, ev.On == in.Refs["bedroom"]
		case attach.PoseChanged:
			posed = ev.Pose
		case attach.PortalClosed:
			closed = interrupted
		}
	}
	if !interrupted || !dropped || !closed || posed != "idle" {
		t.Fatalf("interrumpido=%v taza en el piso=%v portal cerrado=%v pose=%q", interrupted, dropped, closed, posed)
	}
	r := in.World.ECS()
	if ecs.Get[jobs.Carried](r).Has(cup) || ecs.Get[jobs.Reservation](r).Has(cup) {
		t.Fatal("la taza sigue cargada o reservada")
	}
	if l, _ := comp.Of(r).Location.Get(cup); l.Pos.Z != 0 {
		t.Fatalf("la taza quedó en el aire: %v", l)
	}
}
