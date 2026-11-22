package scenario_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/encounter"
)

// Aceptación del slice (propuesta 15.3) sobre las semillas de referencia de la
// partida libre, diez minutos cada una. Lo que vale en cada partida:
//
//   - un encuentro no se repite en la misma superficie antes de su
//     enfriamiento;
//   - la palma ofrecida nunca aplasta: ningún aplastamiento mientras la
//     giantess ofrece la palma, y solo se retira la que se apoyó.
//
// Y entre todas las semillas aparecen todos los desenlaces: mirar, compartir
// (con alguna porción guardada en el nido), un tiny que sube a la palma
// ofrecida y una oferta sin pasajeros o cancelada. Las rutinas distintas y la
// continuación desde snapshot están en m28_test.go; los desenlaces aislados,
// en sus escenarios.
func TestEncountersAcceptance(t *testing.T) {
	if testing.Short() {
		t.Skip("diez minutos de partida por semilla")
	}
	cooldown := map[string]clock.Tick{
		"observe": encounter.ObserveCooldown, "share": encounter.ShareCooldown, "offer_palm": encounter.OfferCooldown,
	}
	total := map[string]int{}
	for _, seed := range referenceSeeds {
		in := partida(t, seed, 12000)
		if _, err := in.Run(); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		g := in.Refs["giantess"]
		type key struct {
			action  string
			surface ids.ID
		}
		last := map[key]clock.Tick{}
		doing, offered := "", false
		var shared []ids.ID
		stored := map[ids.ID]bool{}
		for _, e := range in.Events {
			switch ev := e.Event.(type) {
			case ai.ActionStarted:
				if ev.Entity != g {
					continue
				}
				doing = ev.Action
				if c, ok := cooldown[ev.Action]; ok {
					k := key{ev.Action, ev.Target}
					if prev, seen := last[k]; seen && e.Tick-prev < c {
						t.Errorf("seed %d: %s en %s a los %d ticks del anterior", seed, ev.Action, ev.Target, e.Tick-prev)
					}
					last[k] = e.Tick
				}
			case ai.ActionDone, ai.ActionFailed:
				// (una interrupción no: la acción nueva ya empezó antes)
				if who, what := ended(ev); who == g && what == doing {
					doing = ""
				}
			case crush.Crushed:
				if doing == "offer_palm" {
					t.Errorf("seed %d, tick %d: aplastamiento (%s) durante la palma ofrecida", seed, e.Tick, ev.Source)
				}
			case encounter.Observed:
				total["observed"]++
			case encounter.Shared:
				shared = append(shared, ev.Items...)
			case colony.Stored:
				stored[ev.Item] = true
			case encounter.PalmOffered:
				offered = true
				total["offered"]++
			case encounter.PalmWithdrawn:
				if !offered {
					t.Errorf("seed %d: retiró una palma que no ofreció", seed)
				}
				offered = false
				if len(ev.Boarded) > 0 {
					total["boarded"]++
				} else {
					total["empty"]++
				}
			case encounter.Skipped:
				if ev.Reason == "occupied" {
					total["empty"]++
				}
			}
		}
		total["shared"] += len(shared)
		for _, it := range shared {
			if stored[it] {
				total["stored"]++
			}
		}
	}
	for _, k := range []string{"observed", "shared", "stored", "offered", "boarded", "empty"} {
		if total[k] == 0 {
			t.Errorf("ninguna semilla dio %s: %v", k, total)
		}
	}
	t.Logf("%v", total)
}

// ended es quién terminó qué acción.
func ended(ev any) (ids.ID, string) {
	switch ev := ev.(type) {
	case ai.ActionDone:
		return ev.Entity, ev.Action
	case ai.ActionFailed:
		return ev.Entity, ev.Action
	}
	return ids.Nil, ""
}
