package scenario_test

import (
	"bytes"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/world"
)

// Solo un apoyo nuevo aplasta: un tiny que camina por debajo de la palma ya
// apoyada no muere, tampoco después de guardar y cargar (el apoyo en curso,
// crusher.pressing, es parte del estado). Restos con nombre y causa.
func TestCrushOnlyOnNewPress(t *testing.T) {
	spec, err := scenario.ReadFile("../../tests/scenarios/palm_crushes_tiny.json")
	if err != nil {
		t.Fatal(err)
	}
	spec.Ticks = 60 // hasta antes de levantar la mano
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := in.Run(); err != nil {
		t.Fatal(err)
	}
	w := in.World
	r := w.ECS()

	var remains []crush.Remains
	ecs.Get[crush.Remains](r).Each(func(_ ids.ID, m *crush.Remains) { remains = append(remains, *m) })
	if len(remains) != 1 || remains[0].Name != "under" || remains[0].Source != "hand" || remains[0].By != in.Refs["palm"] {
		t.Fatalf("restos = %+v", remains)
	}
	if p, _ := ecs.Get[crush.Crusher](r).Get(in.Refs["palm"]); len(p.Pressing) != 1 || p.Pressing[0] != in.Refs["desk1"] {
		t.Fatalf("pressing = %v", p.Pressing)
	}

	// El tiny de al lado camina hasta debajo de la palma apoyada.
	beside := in.Refs["beside"]
	if _, err := w.Enqueue(nav.MoveTo{Entity: beside, Space: in.Refs["desk1"], Goal: units.V2(200000, 240000)}); err != nil {
		t.Fatal(err)
	}
	w.StepN(5)

	snap, err := w.Save()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := world.WriteSnapshot(&buf, snap); err != nil {
		t.Fatal(err)
	}
	w2, err := engine.Load(&buf)
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range []*world.World{w, w2} {
		for range 200 {
			x.Step()
			for _, e := range x.Events() {
				if c, ok := e.Event.(crush.Crushed); ok {
					t.Fatalf("tick %d: aplastado sin apoyo nuevo: %+v", e.Tick, c)
				}
			}
			if err := x.CheckInvariants(); err != nil {
				t.Fatal(err)
			}
		}
	}
	if w.Hash() != w2.Hash() {
		t.Fatal("el mundo restaurado diverge")
	}
	l, _ := comp.Of(r).Location.Get(beside)
	if d := l.Pos.XY().DistSq(units.V2(200000, 240000)); d > 1000*1000 {
		t.Fatalf("el tiny no llegó debajo de la palma: %v", l.Pos)
	}
}
