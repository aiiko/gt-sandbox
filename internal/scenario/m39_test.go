package scenario_test

import (
	"testing"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/needs"
)

// M39: descansar en el nido sube la energía (la del final del descanso es
// mayor que la del principio, aunque siga bajando mientras camina).
func TestRestRaisesEnergy(t *testing.T) {
	spec, err := scenario.ReadFile("../../tests/scenarios/tiny_rests_at_nest.json")
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
	tiny := in.Refs["t"]
	energy := func() int32 {
		n, _ := ecs.Get[needs.Needs](in.World.ECS()).Get(tiny)
		v, _ := n.Get("energy")
		return v
	}
	var before int32 = -1
	rested := false
	for !p.Done() && !rested {
		if err := p.Step(); err != nil {
			t.Fatal(err)
		}
		a, _ := ecs.Get[ai.Agent](in.World.ECS()).Get(tiny)
		if before < 0 && a.Doing != nil && a.Doing.Action == "use" && a.Doing.Stage == 1 {
			before = energy() // empezó el paso use
		}
		for _, e := range in.World.Events() {
			if d, ok := e.Event.(ai.ActionDone); ok && d.Entity == tiny && d.What == "rest" {
				rested = true
			}
		}
	}
	if !rested || before < 0 {
		t.Fatalf("no descansó (antes %d)", before)
	}
	if after := energy(); after <= before {
		t.Fatalf("energía %d → %d", before, after)
	}
}

// M39: el déficit de la despensa va de 1 (vacía) a 0 (en la meta), sin meta
// es 0 y cuenta solo miembros vivos.
func TestPantryDeficit(t *testing.T) {
	in := run(t, "tiny_rests_at_nest.json", func(s *scenario.Spec) { s.Ticks = 1 })
	w := in.World
	nest := in.Refs["nest"]
	set := func(stock, left int32) {
		c := ecs.Get[colony.Colony](w.ECS()).Ptr(nest)
		c.Stock = stock
		ecs.Get[ai.Portions](w.ECS()).Ptr(nest).Left = left
	}
	for _, tc := range []struct {
		stock, left int32
		want        float64
	}{{0, 0, 0}, {4, 0, 1}, {4, 1, 0.75}, {4, 4, 0}, {4, 9, 0}} {
		set(tc.stock, tc.left)
		if got := colony.Deficit(w, nest); got != tc.want {
			t.Errorf("stock %d, depósito %d: déficit %v, quería %v", tc.stock, tc.left, got, tc.want)
		}
	}
}
