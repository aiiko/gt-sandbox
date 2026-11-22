package scenario_test

import (
	"slices"
	"strings"
	"testing"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/needs"
)

// partidaBase es la base de «Nueva partida» (tools/scenarios/partida.py).
const partidaBase = "../../content/partidas/casa.json"

// referenceSeeds son las semillas de referencia de la partida libre.
var referenceSeeds = []uint64{1, 7, 21, 42, 99}

// partida arma la partida libre de una semilla con la duración dada.
func partida(t *testing.T, seed uint64, ticks int) *scenario.Instance {
	t.Helper()
	base, err := scenario.ReadFile(partidaBase)
	if err != nil {
		t.Fatal(err)
	}
	s, err := scenario.Randomize(base, seed)
	if err != nil {
		t.Fatal(err)
	}
	s.Ticks = ticks
	in, err := scenario.Build(s)
	if err != nil {
		t.Fatalf("seed %d: %v", seed, err)
	}
	return in
}

// giantessRoutine son las actividades que empezó la giantess, en orden.
func giantessRoutine(in *scenario.Instance) []string {
	g := in.Refs["giantess"]
	var out []string
	for _, e := range in.Events {
		if a, ok := e.Event.(ai.ActionStarted); ok && a.Entity == g && a.What != "" {
			out = append(out, a.What)
		}
	}
	return out
}

// Cada partida elige un perfil de arranque de la giantess y lo aplica a sus
// necesidades, que decaen; la base no cambia.
func TestPartidaProfiles(t *testing.T) {
	base, err := scenario.ReadFile(partidaBase)
	if err != nil {
		t.Fatal(err)
	}
	if base.Random == nil || len(base.Random.Starts) < 3 {
		t.Fatal("la base no trae perfiles de arranque")
	}
	profiles := map[string]bool{}
	for seed := range uint64(20) {
		s, err := scenario.Randomize(base, seed+1)
		if err != nil {
			t.Fatal(err)
		}
		if s.Random != nil || s.Profile == "" {
			t.Fatalf("seed %d: perfil %q, random %v", seed+1, s.Profile, s.Random)
		}
		profiles[s.Profile] = true
		in, err := scenario.Build(s)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := ecs.Get[needs.Needs](in.World.ECS()).Get(in.Refs["giantess"])
		start := base.Random.Starts[slices.IndexFunc(base.Random.Starts, func(st scenario.RandomStart) bool { return st.Name == s.Profile })]
		for _, n := range start.Needs {
			i := slices.IndexFunc(got, func(m needs.Need) bool { return m.Name == n.Name })
			if i < 0 || got[i] != n {
				t.Fatalf("seed %d (%s): %s es %+v, se esperaba %+v", seed+1, s.Profile, n.Name, got, n)
			}
		}
	}
	if len(profiles) != len(base.Random.Starts) {
		t.Fatalf("en 20 semillas salieron %d perfiles de %d", len(profiles), len(base.Random.Starts))
	}
	if again, _ := scenario.ReadFile(partidaBase); again.Random == nil || base.Profile != "" {
		t.Fatal("Randomize tocó la base")
	}
}

// Criterio 15.3: en la misma casa, semillas de referencia distintas producen
// al menos dos órdenes de actividades de la giantess, y la rutina no se agota:
// sigue decidiendo pasados los primeros minutos.
func TestPartidaRoutineVaries(t *testing.T) {
	orders := map[string][]uint64{}
	for _, seed := range referenceSeeds {
		in := partida(t, seed, 6000) // cinco minutos
		if _, err := in.Run(); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		r := giantessRoutine(in)
		if len(r) < 3 {
			t.Errorf("seed %d: la giantess solo hizo %v", seed, r)
		}
		key := strings.Join(r, " → ")
		orders[key] = append(orders[key], seed)
		t.Logf("seed %d (%s): %s", seed, in.Spec.Profile, key)
	}
	if len(orders) < 2 {
		t.Fatalf("todas las semillas dan la misma rutina: %v", orders)
	}
}

// Criterio 15.3: con la misma semilla, la partida y su continuación desde un
// snapshot reproducen las mismas decisiones y eventos.
func TestPartidaSnapshotContinuation(t *testing.T) {
	for _, seed := range []uint64{7, 42} {
		in := partida(t, seed, 3000)
		playChecked(t, in, 113, nil)
	}
}
