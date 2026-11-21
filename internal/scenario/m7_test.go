package scenario_test

import (
	"encoding/json"
	"strings"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/world"
)

func run(t *testing.T, file string, edit func(*scenario.Spec)) *scenario.Instance {
	t.Helper()
	spec, err := scenario.ReadFile("../../tests/scenarios/" + file)
	if err != nil {
		t.Fatal(err)
	}
	if edit != nil {
		edit(spec)
	}
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := in.Run(); err != nil {
		t.Fatal(err)
	}
	return in
}

// oscillations cuenta interrupciones de una acción que a su vez había
// interrumpido a otra (A→B→A...).
func oscillations(events []world.EventRecord) int {
	interrupter := false // la acción en curso empezó interrumpiendo a otra
	n := 0
	for _, e := range events {
		switch e.Event.(type) {
		case ai.ActionInterrupted:
			if interrupter {
				n++
			}
			interrupter = true
		case ai.ActionDone, ai.ActionFailed:
			interrupter = false
		}
	}
	return n
}

// Aceptación de M7: la giantess (3 necesidades, 3 acciones) no oscila; sin
// inercia, sí (el test detecta la oscilación).
func TestStableChoice(t *testing.T) {
	if n := oscillations(run(t, "giantess_needs_stable.json", nil).Events); n != 0 {
		t.Fatalf("oscila %d veces con la inercia por defecto", n)
	}
	noInertia := func(s *scenario.Spec) {
		for i, e := range s.Entities {
			if e.Ref == "giantess" {
				s.Entities[i].Spawn = json.RawMessage(strings.Replace(string(e.Spawn),
					`"distance_cost": 0.01`, `"distance_cost": 0.01, "inertia": 0`, 1))
			}
		}
	}
	if n := oscillations(run(t, "giantess_needs_stable.json", noInertia).Events); n == 0 {
		t.Fatal("sin inercia debería oscilar")
	}
}

// Aceptación de M7: cada migaja se reserva una sola vez, cada trabajo lo toma
// un solo tiny y el tercero nunca empieza nada.
func TestCrumbReservedOnce(t *testing.T) {
	in := run(t, "tinies_share_crumbs.json", nil)
	reserved := map[ids.ID]int{}
	workers := map[ids.ID]bool{}
	for _, e := range in.Events {
		switch ev := e.Event.(type) {
		case jobs.Reserved:
			reserved[ev.Target]++
		case ai.ActionStarted:
			workers[ev.Entity] = true
		}
	}
	if reserved[in.Refs["crumb1"]] != 1 || reserved[in.Refs["crumb2"]] != 1 || len(workers) != 2 {
		t.Fatalf("reservas %v, trabajadores %v", reserved, workers)
	}
}

// Lo que se carga viaja con quien lo lleva.
func TestCarriedFollows(t *testing.T) {
	spec, err := scenario.ReadFile("../../tests/scenarios/tinies_share_crumbs.json")
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
	r := in.World.ECS()
	locs, carried := comp.Of(r).Location, ecs.Get[jobs.Carried](r)
	n := 0
	for !p.Done() {
		if err := p.Step(); err != nil {
			t.Fatal(err)
		}
		carried.Each(func(e ids.ID, c *jobs.Carried) {
			if a, _ := locs.Get(e); a != must(locs.Get(c.By)) {
				t.Fatalf("tick %v: %v no sigue a %v", in.World.Tick(), e, c.By)
			}
			n++
		})
	}
	if n == 0 {
		t.Fatal("nadie cargó nada")
	}
}

func must[T any](v T, _ bool) T { return v }
