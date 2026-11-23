package scenario_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
)

// holds dice si la acción d (en curso o suspendida) usa e: su objetivo, el
// objetivo de su trabajo o lo que carga.
func holds(r *ecs.Registry, d *ai.Doing, e ids.ID) bool {
	if d == nil {
		return false
	}
	if d.Target == e || d.Item == e {
		return true
	}
	j, ok := ecs.Get[jobs.Job](r).Get(d.Target)
	return ok && d.Action == "work" && j.Target == e
}

// Aceptación de la propuesta (16.4) sobre la partida libre: en cada tick de
// los primeros cinco minutos de cada semilla de referencia, nadie revisa un
// punto ni toma un trabajo que ya tiene otro; toda reserva es de alguien
// que la está usando (las tareas abandonadas la liberaron), y cada fibra
// entregada cuenta una vez. El resto de 16.4 lo cubren TestPartidaTinyRoutine,
// TestPartidaLifeSnapshotContinuation, TestStockedColonySearchesLess,
// TestRestRaisesEnergy, TestChatRaisesFun, TestColonyBuildsRefuge, los
// escenarios del slice (tools/scenarios/vida.py) y, en la presentación,
// TestPartidaLifeReadable.
func TestLifeAcceptance(t *testing.T) {
	if testing.Short() {
		t.Skip("cinco minutos de partida por semilla")
	}
	for _, seed := range referenceSeeds {
		in := partida(t, seed, lifeTicks)
		p, err := in.Player()
		if err != nil {
			t.Fatal(err)
		}
		r := in.World.ECS()
		agents := ecs.Get[ai.Agent](r)
		for !p.Done() {
			if err := p.Step(); err != nil {
				t.Fatal(err)
			}
			spots, assigned := map[ids.ID]ids.ID{}, map[ids.ID]ids.ID{}
			agents.Each(func(e ids.ID, a *ai.Agent) {
				d := a.Doing
				if d == nil {
					return
				}
				if d.Action == colony.SearchAction {
					if o, dup := spots[d.Target]; dup {
						t.Fatalf("seed %d, tick %d: %v y %v revisan el mismo punto", seed, p.Tick(), o, e)
					}
					spots[d.Target] = e
				}
				if j, ok := ecs.Get[jobs.Job](r).Get(d.Target); ok && d.Action == "work" {
					if o, dup := assigned[j.Target]; dup {
						t.Fatalf("seed %d, tick %d: %v y %v trabajan sobre %v", seed, p.Tick(), o, e, j.Target)
					}
					assigned[j.Target] = e
				}
			})
			ecs.Get[jobs.Reservation](r).Each(func(e ids.ID, res *jobs.Reservation) {
				a := agents.Ptr(res.By)
				if a == nil || r.Despawning(res.By) {
					return // la libera jobs.upkeep
				}
				if !holds(r, a.Doing, e) && !holds(r, a.Suspended, e) {
					t.Fatalf("seed %d, tick %d: %v reservó %v y ya no lo usa", seed, p.Tick(), res.By, e)
				}
			})
		}
		counted := map[ids.ID]bool{}
		for _, e := range in.Events {
			if d, ok := e.Event.(colony.FiberDelivered); ok {
				if counted[d.Item] {
					t.Fatalf("seed %d: la fibra %v contó dos veces", seed, d.Item)
				}
				counted[d.Item] = true
			}
		}
	}
}
