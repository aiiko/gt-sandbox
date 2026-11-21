package needs_test

import (
	"testing"

	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/world"
)

// Las necesidades bajan Decay por intervalo, sin pasar de 0.
func TestDecay(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 1})
	e := w.ECS().Spawn()
	s := ecs.Get[needs.Needs](w.ECS())
	s.Set(e, needs.Needs{{Name: "hunger", Value: 25, Decay: 10}, {Name: "fun", Value: 500}})
	w.StepN(int(3 * needs.Interval))
	n, _ := s.Get(e)
	if h, _ := n.Get("hunger"); h != 0 {
		t.Fatalf("hunger = %d", h)
	}
	if f, _ := n.Get("fun"); f != 500 {
		t.Fatalf("fun = %d", f)
	}
	w.Enqueue(needs.SetNeed{Entity: e, Need: "hunger", Value: 2000})
	w.Step()
	if len(w.Events()) == 0 || w.Events()[0].Event.Kind() != "command_rejected" {
		t.Fatal("set_need fuera de rango debía rechazarse")
	}
}
