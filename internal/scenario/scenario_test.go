package scenario

import (
	"path/filepath"
	"testing"

	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/world"
)

// Cada archivo de tests/scenarios debe cumplir todas sus expectativas, y
// correrlo dos veces debe dar el mismo hash.
func TestScenarios(t *testing.T) {
	files, _ := filepath.Glob("../../tests/scenarios/*.json")
	if len(files) == 0 {
		t.Fatal("no hay escenarios")
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			var hashes [2]uint64
			for i := range hashes {
				spec, err := ReadFile(f)
				if err != nil {
					t.Fatal(err)
				}
				in, err := Build(spec)
				if err != nil {
					t.Fatal(err)
				}
				fails, err := in.Run()
				if err != nil {
					t.Fatal(err)
				}
				for _, msg := range fails {
					t.Error(msg)
				}
				hashes[i] = in.World.Hash()
			}
			if hashes[0] != hashes[1] {
				t.Fatal("escenario no determinista")
			}
		})
	}
}

// Una multitud corre sin romper invariantes y nadie queda trabado en
// planning más allá de lo que explica el presupuesto de A*.
func TestCrowd(t *testing.T) {
	c, err := NewCrowd(200, 5)
	if err != nil {
		t.Fatal(err)
	}
	arrived := 0
	for range 400 {
		c.Step()
		for _, e := range c.World.Events() {
			if e.Event.Kind() == "move_arrived" {
				arrived++
			}
		}
	}
	if err := c.World.CheckInvariants(); err != nil {
		t.Fatal(err)
	}
	if arrived == 0 {
		t.Fatal("nadie llegó a destino en 20 s")
	}
}

// Guardar y cargar a mitad de un movimiento (grillas, portales, ruta y
// estado del mover en el snapshot) no cambia el futuro, incluidos los
// comandos posteriores al guardado.
func TestSaveLoadMidMove(t *testing.T) {
	cases := []struct {
		file, who  string
		save, legs int // tick del guardado y tramos que debe tener la ruta de who
	}{
		{"replan_when_gap_closes", "tiny", 60, 1},
		{"portal_disabled_midway", "tiny", 150, 3},    // suelo, pata A, escritorio
		{"giantess_carries_tiny", "giantess", 120, 2}, // habitación A, puerta, B
	}
	for _, c := range cases {
		spec, err := ReadFile("../../tests/scenarios/" + c.file + ".json")
		if err != nil {
			t.Fatal(err)
		}
		in, err := Build(spec)
		if err != nil {
			t.Fatal(err)
		}
		byTick, err := in.commands()
		if err != nil {
			t.Fatal(err)
		}
		worlds := []*world.World{in.World}
		for tick := range spec.Ticks {
			if tick == c.save {
				m, _ := ecs.Get[nav.Mover](in.World.ECS()).Get(in.Refs[c.who])
				if m.Status != nav.Moving || len(m.Legs) != c.legs {
					t.Fatalf("%s: %s debería estar a mitad de camino: %+v", c.file, c.who, m)
				}
				snap, err := in.World.Save()
				if err != nil {
					t.Fatal(err)
				}
				w2, err := engine.Restore(snap)
				if err != nil {
					t.Fatal(err)
				}
				worlds = append(worlds, w2)
			}
			for _, w := range worlds {
				for _, cmd := range byTick[tick] {
					w.Enqueue(cmd)
				}
				w.Step()
			}
		}
		if worlds[0].Hash() != worlds[1].Hash() {
			t.Fatalf("%s: el mundo restaurado divergió", c.file)
		}
	}
}
