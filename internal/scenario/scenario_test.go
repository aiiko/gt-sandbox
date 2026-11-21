package scenario

import (
	"path/filepath"
	"testing"

	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
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

// Guardar y cargar a mitad de un movimiento (grilla, ruta y estado del mover
// en el snapshot) no cambia el futuro.
func TestSaveLoadMidMove(t *testing.T) {
	spec, err := ReadFile("../../tests/scenarios/replan_when_gap_closes.json")
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
	w := in.World
	for tick := range 60 {
		for _, c := range byTick[tick] {
			w.Enqueue(c)
		}
		w.Step()
	}
	m, _ := ecs.Get[nav.Mover](w.ECS()).Get(in.Refs["tiny"])
	if m.Status != nav.Moving || len(m.Path) < 2 {
		t.Fatalf("el tiny debería estar a mitad de camino: %+v", m)
	}
	snap, err := w.Save()
	if err != nil {
		t.Fatal(err)
	}
	w2, err := engine.Restore(snap)
	if err != nil {
		t.Fatal(err)
	}
	w.StepN(600)
	w2.StepN(600)
	if w.Hash() != w2.Hash() {
		t.Fatal("el mundo restaurado divergió")
	}
}
