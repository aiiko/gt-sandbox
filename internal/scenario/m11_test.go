package scenario_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

func TestUnaMananaEnLaCasa(t *testing.T) {
	successes := 0
	for _, seed := range []uint64{1, 7, 21, 42, 99} {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			spec, err := scenario.ReadFile("../../tests/scenarios/una_manana_en_la_casa.json")
			if err != nil {
				t.Fatal(err)
			}
			spec.Seed = seed
			if spec.Ticks >= 20*60*clock.TPS {
				t.Fatal("el escenario debe durar menos de 20 minutos")
			}
			in, err := scenario.Build(spec)
			if err != nil {
				t.Fatal(err)
			}
			members := map[ids.ID]bool{}
			ecs.Get[colony.Member](in.World.ECS()).Each(func(e ids.ID, m *colony.Member) {
				if m.Colony == in.Refs["nest"] {
					members[e] = true
				}
			})
			if len(members) != 20 {
				t.Fatalf("hay %d miembros, se esperaban 20", len(members))
			}
			playMorningWithSnapshots(t, in)
			for _, e := range in.Events {
				switch ev := e.Event.(type) {
				case world.CommandRejected:
					t.Fatalf("tick %d: comando rechazado: %s", e.Tick, ev.Error)
				case space.EntityFell:
					if members[ev.Entity] {
						t.Fatalf("tick %d: se cayó un tiny: %+v", e.Tick, ev)
					}
				}
			}
			tick, err := morningSequence(in, members)
			if err != nil {
				t.Log(err)
				return
			}
			successes++
			t.Logf("secuencia completa a los %.2f s simulados", float64(tick-in.Start)/clock.TPS)
		})
	}
	if successes < 3 {
		t.Fatalf("secuencia completa en %d de 5 semillas, se requieren al menos 3", successes)
	}
}

// Después del comando inicial, el escenario usa solo sistemas del motor.
// Comparamos el log tick a tick con una copia que se guarda y recarga
// periódicamente, además del hash en cada recarga y al terminar.
func playMorningWithSnapshots(t *testing.T, in *scenario.Instance) {
	t.Helper()
	p, err := in.Player()
	if err != nil {
		t.Fatal(err)
	}
	var restored *world.World
	for !p.Done() {
		if err := p.Step(); err != nil {
			t.Fatal(err)
		}
		if restored != nil {
			restored.Step()
			if !reflect.DeepEqual(restored.Events(), in.World.Events()) {
				t.Fatalf("el log restaurado divergió en el tick %d", p.Tick())
			}
		}
		if (p.Tick()-1)%113 != 0 && !p.Done() {
			continue
		}
		if restored != nil && restored.Hash() != in.World.Hash() {
			t.Fatalf("el estado restaurado divergió en el tick %d", p.Tick())
		}
		source := restored
		if source == nil {
			source = in.World
		}
		snap, err := source.Save()
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err := world.WriteSnapshot(&buf, snap); err != nil {
			t.Fatal(err)
		}
		restored, err = engine.Load(&buf)
		if err != nil {
			t.Fatal(err)
		}
		if restored.Hash() != in.World.Hash() {
			t.Fatalf("el snapshot cambió el estado en el tick %d", p.Tick())
		}
	}
}

func morningSequence(in *scenario.Instance, members map[ids.ID]bool) (clock.Tick, error) {
	refs := in.Refs
	giantess, palm, desk, table := refs["giantess"], refs["palm"], refs["desk1"], refs["desk2"]
	crumbs := map[ids.ID]bool{}
	noticed := map[ids.ID]bool{}
	hauling := map[ids.ID]bool{}
	riders := map[ids.ID]bool{}
	disembarked := map[ids.ID]bool{}
	var cup ids.ID
	var breakfast, served, assigned, raised, lifted, moved, rested bool
	for _, e := range in.Events {
		switch ev := e.Event.(type) {
		case ai.ActionStarted:
			breakfast = breakfast || ev.Entity == giantess && ev.What == "breakfast" && ev.Target == refs["fridge"]
		case cmds.EntitySpawned:
			if breakfast && ev.Name == "cup" {
				cup = ev.Entity
			}
			if breakfast && ev.Name == "crumb" {
				crumbs[ev.Entity] = true
			}
		case ai.ItemDropped:
			served = served || ev.Entity == giantess && ev.Item == cup && ev.On == desk
		case ai.Noticed:
			if served && members[ev.Entity] && crumbs[ev.Target] {
				noticed[ev.Target] = true
			}
		case jobs.JobPosted:
			hauling[ev.Job] = ev.Type == "haul" && noticed[ev.Target]
		case jobs.JobAssigned:
			assigned = assigned || hauling[ev.Job] && members[ev.By]
		case attach.PoseChanged:
			raised = raised || len(riders) > 0 && ev.Entity == giantess && ev.Pose == "idle"
		case attach.PortalClosed:
			lifted = lifted || raised && ev.A == desk && ev.B == palm
		case attach.PortalOpened:
			rested = rested || moved && ev.A == table && ev.B == palm
		case space.EntityReparented:
			if ev.Cause != "portal" {
				continue
			}
			if members[ev.Entity] {
				if assigned && !lifted && ev.From == desk && ev.To == palm {
					riders[ev.Entity] = true
				}
				if ev.From == palm {
					disembarked[ev.Entity] = riders[ev.Entity] && rested && ev.To == table
					delete(riders, ev.Entity)
				}
			}
			moved = moved || lifted && len(riders) > 0 && ev.Entity == giantess &&
				ev.From == refs["bedroom"] && ev.To == refs["living"]
		case ai.Explored:
			if ev.Space == table && disembarked[ev.Entity] {
				return e.Tick, nil
			}
		}
	}
	return 0, fmt.Errorf("desayuno=%v taza=%v migajas=%d acarreo=%v mano levantada=%v viaje=%v mano apoyada en el living=%v; faltó explorar el living",
		breakfast, served, len(crumbs), assigned, lifted, moved, rested)
}
