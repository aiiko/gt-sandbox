package scenario_test

import (
	"bytes"
	"testing"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/world"
)

// El recorrido desde la última pisada y el pie que sigue son estado: guardar
// entre las dos pisadas y continuar da las mismas pisadas y el mismo hash.
func TestStrideSurvivesSave(t *testing.T) {
	spec, err := scenario.ReadFile("../../tests/scenarios/step_crushes_tiny_on_floor.json")
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
	for range 10 { // entre la pisada del tick 7 y la del 14
		if err := p.Step(); err != nil {
			t.Fatal(err)
		}
	}
	snap, err := in.World.Save()
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
	steps := func(w *world.World) []crush.Footstep {
		var out []crush.Footstep
		for range 30 {
			w.Step()
			for _, e := range w.Events() {
				if f, ok := e.Event.(crush.Footstep); ok {
					out = append(out, f)
				}
			}
		}
		return out
	}
	a, b := steps(in.World), steps(w2)
	if len(a) != 1 || len(b) != 1 || a[0] != b[0] || !a[0].Right {
		t.Fatalf("pisadas: original %+v, restaurado %+v", a, b)
	}
	if in.World.Hash() != w2.Hash() {
		t.Fatal("el mundo restaurado diverge")
	}
}
