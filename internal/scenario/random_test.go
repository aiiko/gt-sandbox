package scenario

import (
	"encoding/json"
	"reflect"
	"testing"

	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
)

const morning = "../../tests/scenarios/una_manana_en_la_casa.json"

func randomMorning(t *testing.T, seed uint64) *Spec {
	t.Helper()
	base, err := ReadFile(morning)
	if err != nil {
		t.Fatal(err)
	}
	s, err := Randomize(base, seed)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// La misma semilla da la misma partida y otra semilla, otra.
func TestRandomizeDeterministic(t *testing.T) {
	a, b, c := randomMorning(t, 7), randomMorning(t, 7), randomMorning(t, 8)
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	jc, _ := json.Marshal(c)
	if string(ja) != string(jb) {
		t.Fatal("la misma semilla dio partidas distintas")
	}
	if string(ja) == string(jc) {
		t.Fatal("dos semillas dieron la misma partida")
	}
	if a.Seed != 7 || len(a.Expect) != 0 {
		t.Fatalf("seed %d, %d expectativas", a.Seed, len(a.Expect))
	}
}

// Una partida al azar arma tinies en celdas libres del escritorio, cerca del
// nido, y corre sin romper invariantes.
func TestRandomizeBuildsAndRuns(t *testing.T) {
	for _, seed := range []uint64{1, 2, 3, 99} {
		s := randomMorning(t, seed)
		in, err := Build(s)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		desk := in.Refs["desk1"]
		grid, _ := ecs.Get[nav.Grid](in.World.ECS()).Get(desk)
		locs := ecs.Get[comp.Location](in.World.ECS())
		tinies := 0
		for ref, id := range in.Refs {
			if len(ref) != 6 || ref[:4] != "tiny" {
				continue
			}
			tinies++
			l, ok := locs.Get(id)
			if !ok || l.Space != desk {
				t.Fatalf("seed %d: %s no está en el escritorio", seed, ref)
			}
			if !grid.Walkable(l.Pos.XY()) {
				t.Fatalf("seed %d: %s en una celda bloqueada", seed, ref)
			}
		}
		if tinies < RandomMinTinies || tinies > RandomMaxTinies {
			t.Fatalf("seed %d: %d tinies", seed, tinies)
		}
		p, err := in.Player()
		if err != nil {
			t.Fatal(err)
		}
		for range 300 {
			if err := p.Step(); err != nil {
				t.Fatalf("seed %d: %v", seed, err)
			}
		}
	}
}

// Randomize no toca el escenario base.
func TestRandomizeKeepsBase(t *testing.T) {
	base, err := ReadFile(morning)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := ReadFile(morning)
	if _, err := Randomize(base, 5); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(base, before) {
		t.Fatal("Randomize modificó el escenario base")
	}
}

// La partida al azar del juego sale de la mañana peligrosa: los tinies
// sorteados heredan fragile y reflex, y la mañana entera corre sin romper
// invariantes, con aplastados o no según la semilla.
func TestRandomizeDangerousMorning(t *testing.T) {
	base, err := ReadFile("../../tests/scenarios/un_descuido_en_el_escritorio.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, seed := range []uint64{1, 2, 3} {
		s, err := Randomize(base, seed)
		if err != nil {
			t.Fatal(err)
		}
		in, err := Build(s)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		fragile := ecs.Get[crush.Fragile](in.World.ECS())
		for ref, id := range in.Refs {
			if len(ref) == 6 && ref[:4] == "tiny" && !fragile.Has(id) {
				t.Fatalf("seed %d: %s no es frágil", seed, ref)
			}
		}
		p, err := in.Player()
		if err != nil {
			t.Fatal(err)
		}
		crushed := 0
		for !p.Done() {
			if err := p.Step(); err != nil {
				t.Fatalf("seed %d: %v", seed, err)
			}
		}
		for _, e := range in.Events {
			if e.Event.Kind() == "crushed" {
				crushed++
			}
		}
		t.Logf("seed %d: %d aplastados", seed, crushed)
	}
}

// El cuidado de la giantess se sortea entre 0 y RandomMaxCare cuando el base
// lo fija, y cambia entre semillas; el base sin care no gana uno.
func TestRandomizeCare(t *testing.T) {
	base, err := ReadFile("../../tests/scenarios/un_descuido_en_el_escritorio.json")
	if err != nil {
		t.Fatal(err)
	}
	care := func(s *Spec) (float64, bool) {
		var data struct {
			Components struct {
				Agent map[string]any `json:"agent"`
			} `json:"components"`
		}
		if err := json.Unmarshal(s.Commands[0].Data, &data); err != nil {
			t.Fatal(err)
		}
		c, ok := data.Components.Agent["care"].(float64)
		return c, ok
	}
	seen := map[float64]bool{}
	for seed := range uint64(12) {
		s, err := Randomize(base, seed+1)
		if err != nil {
			t.Fatal(err)
		}
		c, ok := care(s)
		if !ok || c < 0 || c > RandomMaxCare {
			t.Fatalf("seed %d: care %v", seed+1, c)
		}
		seen[c] = true
	}
	if len(seen) < 6 {
		t.Fatalf("el cuidado casi no varía: %v", seen)
	}
	if c, _ := care(base); c != 0 {
		t.Fatalf("Randomize cambió el base: care %v", c)
	}
	if _, ok := care(randomMorning(t, 3)); ok {
		t.Fatal("la mañana sin care no debería ganar uno")
	}
}
