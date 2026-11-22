package scenario_test

import (
	"cmp"
	"slices"
	"testing"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/encounter"
)

// La giantess registra en qué superficie ve tinies, en orden estable, y
// olvida lo que dejó de ver hace más de SightingTTL; el registro sobrevive a
// guardar y cargar.
func TestGiantessLogsSightings(t *testing.T) {
	s, err := scenario.ReadFile("../../tests/scenarios/giantess_logs_sightings.json")
	if err != nil {
		t.Fatal(err)
	}
	in, err := scenario.Build(s)
	if err != nil {
		t.Fatal(err)
	}
	g, desk, kitchen := in.Refs["giantess"], in.Refs["desk1"], in.Refs["kitchen"]
	encs := ecs.Get[encounter.Encounter](in.World.ECS())
	sawDesk, maxSurfaces := false, 0
	playChecked(t, in, 113, func() {
		enc, ok := encs.Get(g)
		if !ok {
			t.Fatal("la giantess perdió su registro")
		}
		if !slices.IsSortedFunc(enc.Sightings, func(a, b encounter.Sighting) int { return cmp.Compare(a.Surface, b.Surface) }) {
			t.Fatalf("avistamientos desordenados: %+v", enc.Sightings)
		}
		maxSurfaces = max(maxSurfaces, len(enc.Sightings))
		if s, ok := enc.Sighting(desk); ok && s.Seen > 0 {
			sawDesk = true
		}
		_, old := enc.Sighting(kitchen)
		if now := in.World.Tick(); old && now > encounter.SightingTTL+20 {
			t.Fatalf("tick %d: el avistamiento viejo de la cocina sigue", now)
		} else if !old && now < encounter.SightingTTL-20 {
			t.Fatalf("tick %d: olvidó la cocina antes de tiempo", now)
		}
	})
	if !sawDesk {
		t.Fatal("nunca registró a los tinies del escritorio")
	}
	if maxSurfaces < 2 {
		t.Fatalf("solo registró %d superficies a la vez", maxSurfaces)
	}
}
