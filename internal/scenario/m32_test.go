package scenario_test

import (
	"slices"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/encounter"
)

// offerRun corre un escenario de la palma ofrecida guardando y cargando a
// cada rato, y devuelve quiénes subieron según palm_withdrawn y según los
// explored hacia la palma.
func offerRun(t *testing.T, file string) (boarded, explored []ids.ID, in *scenario.Instance) {
	t.Helper()
	s, err := scenario.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	in, err = scenario.Build(s)
	if err != nil {
		t.Fatal(err)
	}
	playChecked(t, in, 113, nil)
	if msgs, err := in.Check(); err != nil || len(msgs) > 0 {
		t.Fatalf("%v %v", msgs, err)
	}
	palm := in.Refs["palm"]
	withdrawn := 0
	for _, e := range in.Events {
		switch ev := e.Event.(type) {
		case encounter.PalmWithdrawn:
			withdrawn++
			boarded = append(boarded, ev.Boarded...)
		case ai.Explored:
			if ev.Space == palm && !slices.Contains(explored, ev.Entity) {
				explored = append(explored, ev.Entity)
			}
		}
	}
	if withdrawn > 1 {
		t.Fatalf("retiró la mano %d veces", withdrawn)
	}
	return boarded, explored, in
}

// Criterio 15.3: un tiny cruza a la palma ofrecida por decisión propia, y
// palm_withdrawn nombra a los que subieron.
func TestTinyBoardsOfferedPalm(t *testing.T) {
	boarded, explored, _ := offerRun(t, "../../tests/scenarios/tiny_boards_offered_palm.json")
	if len(boarded) == 0 {
		t.Fatal("nadie subió")
	}
	slices.Sort(boarded)
	slices.Sort(explored)
	if !slices.Equal(boarded, explored) {
		t.Fatalf("subieron %v, exploraron la palma %v", boarded, explored)
	}
}

// Criterio 15.3: sin curiosidad nadie sube, y la mano se retira al acabar
// la espera.
func TestNobodyBoardsOfferedPalm(t *testing.T) {
	boarded, explored, _ := offerRun(t, "../../tests/scenarios/nobody_boards_offered_palm.json")
	if len(boarded) > 0 || len(explored) > 0 {
		t.Fatalf("subieron %v (explored %v)", boarded, explored)
	}
}

// Criterio 15.3: con un tiny inmóvil bajo la huella, la giantess cancela el
// apoyo y nadie muere.
func TestOfferCanceledWhenOccupied(t *testing.T) {
	_, _, in := offerRun(t, "../../tests/scenarios/offer_canceled_when_occupied.json")
	if !in.World.ECS().Alive(in.Refs["tiny17"]) {
		t.Fatal("el tiny de la huella murió")
	}
}
