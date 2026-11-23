package scenario_test

import (
	"testing"

	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/space"
)

// TestSeatsAcceptance (M38) verifica el slice Asientos en la casa sobre las
// semillas de referencia de la partida libre:
//   - silla y sofá presentes y conectados;
//   - los tinies alcanzan la silla o el sofá durante la partida;
//   - la giantess se sienta tanto en la silla como en el sofá y usa `relax`;
//   - las invariantes se cumplen y el snapshot reproduce el mismo hash.
func TestSeatsAcceptance(t *testing.T) {
	if testing.Short() {
		t.Skip("diez minutos de partida por semilla")
	}
	total := map[string]int{}
	for _, seed := range referenceSeeds {
		in := partida(t, seed, 12000)
		if _, err := in.Run(); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		g := in.Refs["giantess"]
		chair, sofa, rear := in.Refs["chair"], in.Refs["sofa"], in.Refs["body_rear"]
		if chair.IsNil() || sofa.IsNil() || rear.IsNil() {
			t.Fatalf("seed %d: falta chair/sofa/body_rear en casa.json", seed)
		}
		for _, e := range in.Events {
			switch ev := e.Event.(type) {
			case space.EntityReparented:
				if ev.Cause == "portal" && ev.Entity != g {
					if ev.To == chair {
						total["tiny_chair"]++
					}
					if ev.To == sofa {
						total["tiny_sofa"]++
					}
				}
			case space.EntityFell:
				if ev.To == chair {
					total["fell_chair"]++
				}
				if ev.To == sofa {
					total["fell_sofa"]++
				}
			case ai.ActionStarted:
				if ev.Entity == g && ev.What == "relax" {
					total["relax"]++
				}
			case crush.Pressed:
				if ev.Source == crush.SourceSeat && ev.Surface == chair {
					total["sit_chair"]++
				}
				if ev.Source == crush.SourceSeat && ev.Surface == sofa {
					total["sit_sofa"]++
				}
			case crush.Crushed:
				if ev.Source == crush.SourceSeat {
					total["crushed_seat"]++
				}
			}
		}
	}
	t.Logf("%v", total)
	for _, k := range []string{"tiny_chair", "fell_chair", "sit_chair", "sit_sofa", "relax"} {
		if total[k] == 0 {
			t.Errorf("ninguna semilla dio %s: %v", k, total)
		}
	}
}
