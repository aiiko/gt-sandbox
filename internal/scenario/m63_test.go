package scenario_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/encounter"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/world"
)

// M63: un tiny revisa un punto de búsqueda con botín, sortea un tesoro de
// forma reproducible con colony.loot y lo guarda en el hoard del nido.
func TestM63TinyFindsTrinket(t *testing.T) {
	in1 := run(t, "tiny_finds_trinket.json", nil)
	in2 := run(t, "tiny_finds_trinket.json", nil)
	if in1.World.Hash() != in2.World.Hash() {
		t.Fatalf("sorteo de botín no reproducible: %x vs %x", in1.World.Hash(), in2.World.Hash())
	}
	h, ok := ecs.Get[colony.Hoard](in1.World.ECS()).Get(in1.Refs["nest"])
	if !ok || len(h.Items) == 0 {
		t.Fatalf("esperaba al menos un tesoro en el hoard del nido, obtuve %+v", h)
	}
}

// M63: cada tesoro distinto en el hoard del nido sube el piso de diversión
// (fun) y frena su decaimiento.
func TestM63HoardRaisesFun(t *testing.T) {
	in := run(t, "hoard_raises_fun.json", nil)
	r := in.World.ECS()
	h, ok := ecs.Get[colony.Hoard](r).Get(in.Refs["nest"])
	if !ok || h.DistinctKinds(r) != 2 {
		t.Fatalf("esperaba 2 tipos distintos en hoard, obtuve %+v", h)
	}
	n, ok := ecs.Get[needs.Needs](r).Get(in.Refs["t"])
	if !ok {
		t.Fatal("tiny sin necesidades")
	}
	fun, _ := n.Get("fun")
	wantFloor := int32(2) * colony.HoardFunPerKind
	if fun < wantFloor {
		t.Fatalf("fun cayó por debajo del piso del hoard (%d < %d)", fun, wantFloor)
	}
}

// M63: un tiny con mucha diversión deja un tesoro repetido en un gift_spot;
// la giantess lo nota (notice_gift) y baja su enfriamiento de share.
func TestM63GiantessNoticesGift(t *testing.T) {
	in := run(t, "giantess_notices_gift.json", nil)
	r := in.World.ECS()
	g := in.Refs["giantess"]
	desk := in.Refs["desk1"]

	enc, ok := ecs.Get[encounter.Encounter](r).Get(g)
	if !ok {
		t.Fatal("giantess sin encounter")
	}
	for _, cd := range enc.Cooldowns {
		if cd.Surface == desk && cd.Action == "share" && cd.Until >= 6000 {
			t.Fatalf("el enfriamiento de share debía bajar tras notice_gift, sigue en %v", cd.Until)
		}
	}

	// Continuidad desde snapshot a mitad del escenario.
	spec, err := scenario.ReadFile(filepath.Join("..", "..", "tests", "scenarios", "giantess_notices_gift.json"))
	if err != nil {
		t.Fatal(err)
	}
	inA, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	inA.World.StepN(300)
	snap, err := inA.World.Save()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := world.WriteSnapshot(&buf, snap); err != nil {
		t.Fatal(err)
	}
	wB, err := engine.Load(&buf)
	if err != nil {
		t.Fatal(err)
	}
	inA.World.StepN(400)
	wB.StepN(400)
	if inA.World.Hash() != wB.Hash() {
		t.Fatalf("continuidad desde snapshot rota en M63: %x vs %x", inA.World.Hash(), wB.Hash())
	}
}
