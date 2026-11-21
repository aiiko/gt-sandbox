package attach_test

import (
	"encoding/json"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// killer destruye una entidad en la fase Slow (después de Topology).
type killer struct {
	target *ids.ID
	at     clock.Tick
}

func (killer) Name() string { return "test.killer" }
func (k killer) Run(w *world.World) {
	if w.Tick() == k.at {
		w.ECS().Despawn(*k.target)
	}
}

func spawn(t *testing.T, w *world.World, c cmds.Spawn) ids.ID {
	t.Helper()
	w.Enqueue(c)
	w.Step()
	for _, e := range w.Events() {
		if s, ok := e.Event.(cmds.EntitySpawned); ok {
			return s.Entity
		}
	}
	t.Fatal("spawn falló")
	return ids.Nil
}

func kinds(w *world.World) map[string]int {
	out := map[string]int{}
	for _, e := range w.Events() {
		out[e.Event.Kind()]++
	}
	return out
}

// Dos superficies que se tocan abren un portal estable (su versión no cambia
// si nada se mueve). Si un sistema destruye una superficie después de
// Topology, el contacto termina igual en ese tick, con sus eventos.
func TestContactLifecycle(t *testing.T) {
	var victim ids.ID
	w := engine.MustNew(world.Config{Seed: 1}, func(w *world.World) error {
		return w.Register(world.SystemSpec{Phase: clock.PhaseSlow}, killer{&victim, 10})
	})
	surface := map[string]json.RawMessage{"contact_surface": json.RawMessage(`{}`)}
	room := spawn(t, w, cmds.Spawn{Space: &space.Space{Extent: units.RectWH(0, 0, units.Meter, units.Meter)}})
	spawn(t, w, cmds.Spawn{
		Space:      &space.Space{Kind: space.Elevated, Extent: units.RectWH(0, 0, 50*units.Centimeter, 50*units.Centimeter)},
		Location:   &comp.Location{Space: room, Pos: units.V3(0, 0, 70*units.Centimeter)},
		Components: surface,
	})
	victim = spawn(t, w, cmds.Spawn{
		Space:      &space.Space{Kind: space.Articulated, Extent: units.RectWH(0, 0, 10*units.Centimeter, 10*units.Centimeter)},
		Location:   &comp.Location{Space: room, Pos: units.V3(45*units.Centimeter, 0, 72*units.Centimeter)},
		Components: surface,
	})
	if k := kinds(w); k["contact_began"] != 1 || k["portal_opened"] != 1 {
		t.Fatalf("eventos al tocar: %v", k)
	}
	portals := ecs.Get[nav.Portal](w.ECS())
	var p nav.Portal
	portals.Each(func(_ ids.ID, v *nav.Portal) { p = *v })
	w.StepN(3)
	portals.Each(func(_ ids.ID, v *nav.Portal) {
		if v.Version != p.Version || v.A != p.A {
			t.Fatalf("el portal cambió sin que nada se moviera: %+v → %+v", p, *v)
		}
	})
	for w.Tick() <= 10 {
		w.Step()
	}
	if k := kinds(w); k["contact_ended"] != 1 || k["portal_closed"] != 1 {
		t.Fatalf("eventos al destruir: %v", k)
	}
	if err := w.CheckInvariants(); err != nil {
		t.Fatal(err)
	}
	if ecs.Get[attach.Contact](w.ECS()).Len() != 0 || portals.Len() != 0 {
		t.Fatal("quedaron contactos o portales")
	}
}
