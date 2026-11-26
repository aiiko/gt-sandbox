package encounter_test

import (
	"encoding/json"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/encounter"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func spawnEntity(t *testing.T, w *world.World, c cmds.Spawn) ids.ID {
	t.Helper()
	w.Enqueue(c)
	w.Step()
	for _, ev := range w.Events() {
		if s, ok := ev.Event.(cmds.EntitySpawned); ok {
			return s.Entity
		}
	}
	t.Fatal("spawn no emitió entity_spawned")
	return ids.Nil
}

type emitterSys struct {
	fn func(w *world.World)
}

func (emitterSys) Name() string         { return "test.emitter" }
func (s emitterSys) Run(w *world.World) { s.fn(w) }

func TestBondEvolvesAndTiniesReactToObserve(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 77})
	gRoom, _ := nav.NewGrid(units.Vec2{}, 10*units.Centimeter, 20, 20)
	room := spawnEntity(t, w, cmds.Spawn{
		Space:      &space.Space{Kind: space.Static, Extent: units.RectWH(0, 0, 2*units.Meter, 2*units.Meter)},
		Components: map[string]json.RawMessage{"nav_grid": mustJSON(t, gRoom)},
	})
	gDesk, _ := nav.NewGrid(units.Vec2{}, 5*units.Centimeter, 8, 8)
	desk := spawnEntity(t, w, cmds.Spawn{
		Space:    &space.Space{Kind: space.Elevated, Extent: units.RectWH(0, 0, 40*units.Centimeter, 40*units.Centimeter)},
		Location: &comp.Location{Space: room, Pos: units.V3(units.Meter, units.Meter, 75*units.Centimeter)},
		Components: map[string]json.RawMessage{
			"nav_grid": mustJSON(t, gDesk),
			"place":    mustJSON(t, ai.Place{Spot: units.V2(20*units.Centimeter, -20*units.Centimeter)}),
		},
	})
	giantess := spawnEntity(t, w, cmds.Spawn{
		Name:     "Giantess",
		Space:    &space.Space{Kind: space.Body},
		Location: &comp.Location{Space: room, Pos: units.V3(120*units.Centimeter, 80*units.Centimeter, 0)},
		Components: map[string]json.RawMessage{
			"mover": mustJSON(t, nav.Mover{Speed: units.Meter, Scale: "giantess"}),
			"agent": mustJSON(t, ai.Agent{Scale: "giantess", Sight: 2 * units.Meter, Watch: 2 * units.Meter}),
			"encounter": mustJSON(t, encounter.Encounter{
				Personality: "curious",
				Observe:     0.6,
				Sightings:   []encounter.Sighting{{Surface: desk, Last: 0, Seen: 2}},
			}),
			"needs": mustJSON(t, []needs.Need{{Name: "fun", Value: 900, Decay: 0}}),
		},
	})
	spawnTiny := func(name string, pos units.Vec2, trust int32) ids.ID {
		return spawnEntity(t, w, cmds.Spawn{
			Name:     comp.Name(name),
			Location: &comp.Location{Space: desk, Pos: units.V3(pos.X, pos.Y, 0)},
			Body:     &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter, Mass: 1},
			Components: map[string]json.RawMessage{
				"mover": mustJSON(t, nav.Mover{Speed: 4 * units.Centimeter, Scale: "tiny"}),
				"agent": mustJSON(t, ai.Agent{Scale: "tiny", Sight: 30 * units.Centimeter, Reflex: 4}),
				"bond":  mustJSON(t, encounter.Bond{Trust: trust}),
				"needs": mustJSON(t, []needs.Need{{Name: "hunger", Value: 950, Decay: 0}}),
			},
		})
	}
	waver := spawnTiny("Tiny Confianza", units.V2(10*units.Centimeter, 10*units.Centimeter), 120)
	shyer := spawnTiny("Tiny Cautela", units.V2(25*units.Centimeter, 25*units.Centimeter), -100)

	var reacted []encounter.TinyReacted
	for range 120 {
		w.Step()
		for _, ev := range w.Events() {
			if tr, ok := ev.Event.(encounter.TinyReacted); ok {
				reacted = append(reacted, tr)
			}
		}
	}
	gotWave, gotShy := false, false
	for _, r := range reacted {
		if r.Entity == waver && r.Reaction == encounter.ReactionWave && r.Giantess == giantess {
			gotWave = true
		}
		if r.Entity == shyer && r.Reaction == encounter.ReactionShy && r.Giantess == giantess {
			gotShy = true
		}
	}
	if !gotWave || !gotShy {
		t.Fatalf("faltaron reacciones tras observe: wave=%v shy=%v eventos=%+v", gotWave, gotShy, reacted)
	}

	var queue []world.Event
	if err := w.Register(world.SystemSpec{Phase: clock.PhaseConsequences, Order: 1}, emitterSys{fn: func(w *world.World) {
		for _, ev := range queue {
			w.Emit(ev)
		}
		queue = nil
	}}); err != nil {
		t.Fatal(err)
	}

	// Shared sube trust; NearMiss baja trust; Crushed reduce Affinity y fija CautionUntil.
	beforeWave := ecs.Get[encounter.Bond](w.ECS()).Ptr(waver).Trust
	queue = append(queue, encounter.Shared{Entity: giantess, Surface: desk})
	w.Step()
	w.Step()
	afterShare := ecs.Get[encounter.Bond](w.ECS()).Ptr(waver).Trust
	if afterShare <= beforeWave {
		t.Fatalf("esperaba subida de confianza tras share (%d -> %d)", beforeWave, afterShare)
	}

	queue = append(queue, crush.NearMiss{Entity: waver, By: giantess})
	w.Step()
	w.Step()
	afterMiss := ecs.Get[encounter.Bond](w.ECS()).Ptr(waver).Trust
	if afterMiss >= afterShare {
		t.Fatalf("esperaba caída de confianza tras near_miss (%d -> %d)", afterShare, afterMiss)
	}

	affBefore := ecs.Get[encounter.Encounter](w.ECS()).Ptr(giantess).Affinity
	queue = append(queue, crush.Crushed{Entity: shyer, By: giantess, Space: desk, Source: "hand"})
	w.Step()
	w.Step()
	enc := ecs.Get[encounter.Encounter](w.ECS()).Ptr(giantess)
	if enc.Affinity >= affBefore || enc.CautionUntil <= w.Tick() {
		t.Fatalf("crushed debía bajar afinidad (%d->%d) y fijar CautionUntil (%d vs %d)",
			affBefore, enc.Affinity, enc.CautionUntil, w.Tick())
	}
}
