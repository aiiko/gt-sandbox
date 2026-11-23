package scenario_test

import (
	"maps"
	"slices"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
)

// lifeTicks son los primeros cinco minutos de una partida.
const lifeTicks = 5 * 60 * 20

// tinyActivities cuenta las actividades que completaron los tinies, con los
// nombres de la propuesta: buscar, acarrear, comer, descansar, preparar el
// refugio y reunirse.
func tinyActivities(in *scenario.Instance) map[string]int {
	jobType := map[ids.ID]string{}
	out := map[string]int{}
	for _, e := range in.Events {
		switch ev := e.Event.(type) {
		case jobs.JobPosted:
			jobType[ev.Job] = ev.Type
		case ai.ActionDone:
			if ev.Entity == in.Refs["giantess"] {
				continue
			}
			switch {
			case ev.Action == colony.SearchAction:
				out["buscar"]++
			case ev.Action == colony.ChatAction:
				out["reunirse"]++
			case ev.Action == "use" && (ev.What == "eat" || ev.What == "nibble"):
				out["comer"]++
			case ev.Action == "use" && ev.What == "rest":
				out["descansar"]++
			case ev.Action == "work" && jobType[ev.Target] == "haul":
				out["acarrear"]++
			case ev.Action == "work" && (jobType[ev.Target] == colony.JobPluck || jobType[ev.Target] == colony.JobBuild):
				out["refugio"]++
			}
		}
	}
	return out
}

// M43: en los primeros cinco minutos de cada semilla de referencia, los
// tinies completan al menos tres actividades distintas.
func TestPartidaTinyRoutine(t *testing.T) {
	if testing.Short() {
		t.Skip("cinco minutos de partida por semilla")
	}
	total := map[string]int{}
	for _, seed := range referenceSeeds {
		in := partida(t, seed, lifeTicks)
		if _, err := in.Run(); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		acts := tinyActivities(in)
		t.Logf("seed %d: %v", seed, acts)
		if len(acts) < 3 {
			t.Errorf("seed %d: solo %v", seed, slices.Sorted(maps.Keys(acts)))
		}
		for k, v := range acts {
			total[k] += v
		}
	}
	t.Logf("total: %v", total)
}

// M43: al continuar desde snapshots tomados a mitad de una búsqueda y de la
// preparación del refugio se obtienen los mismos eventos y el mismo hash
// que en la ejecución continua.
func TestPartidaLifeSnapshotContinuation(t *testing.T) {
	if testing.Short() {
		t.Skip("cinco minutos de partida con snapshots")
	}
	const every = 41 // más corto que una búsqueda: algún snapshot cae a mitad
	in := partida(t, 1, lifeTicks)
	site := in.Refs["site"]
	searching, building := false, false
	playChecked(t, in, every, func() {
		w := in.World
		if (w.Tick()-1)%every != 0 {
			return
		}
		ecs.Get[ai.Agent](w.ECS()).Each(func(_ ids.ID, a *ai.Agent) {
			searching = searching || a.Doing != nil && a.Doing.Action == colony.SearchAction
		})
		if s, ok := ecs.Get[colony.BuildSite](w.ECS()).Get(site); ok && s.Delivered > 0 && s.Delivered < s.Required {
			building = true
		}
	})
	if !searching || !building {
		t.Fatalf("snapshots a mitad de una búsqueda %v, de la obra %v", searching, building)
	}
}
