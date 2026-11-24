package scenario_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/expire"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/world"
)

// M61: un tiny sediento toma de una gota de agua, la achica y recupera sed;
// si sigue tomando hasta vaciarla, la gota desaparece.
func TestM61TinyDrinksDrop(t *testing.T) {
	in := run(t, "tiny_drinks_drop.json", nil)
	r := in.World.ECS()
	n, ok := ecs.Get[needs.Needs](r).Get(in.Refs["t"])
	if !ok {
		t.Fatal("el tiny no tiene necesidades")
	}
	thirst, ok := n.Get("thirst")
	if !ok || thirst <= 250 {
		t.Fatalf("la sed debía subir tras beber, quedó en %d", thirst)
	}
	drop := in.Refs["drop"]
	if r.Alive(drop) {
		wd, ok := ecs.Get[colony.WaterDrop](r).Get(drop)
		if !ok || wd.Volume >= 2 {
			t.Fatalf("la gota debía achicarse (vol < 2), tiene %+v", wd)
		}
	}
}

// M61: una gota que nadie usa vence, desaparece y emite expire.Expired.
func TestM61DropEvaporates(t *testing.T) {
	in := run(t, "drop_evaporates.json", nil)
	if in.World.ECS().Alive(in.Refs["drop"]) {
		t.Fatal("la gota debía desaparecer al vencer")
	}
	var saw bool
	for _, ev := range in.Events {
		if ex, ok := ev.Event.(expire.Expired); ok && ex.Entity == in.Refs["drop"] && ex.ItemKind == "water_drop" {
			saw = true
		}
	}
	if !saw {
		t.Fatal("no se emitió expire.Expired para la gota")
	}
}

// M61: un tiny acarrea una gota al nido (sube water del depósito) y otro tiny
// sediento bebe del nido.
func TestM61WaterHauledToNest(t *testing.T) {
	in := run(t, "water_hauled_to_nest.json", nil)
	r := in.World.ECS()
	nest := in.Refs["nest"]
	p, ok := ecs.Get[ai.Portions](r).Get(nest)
	if !ok {
		t.Fatal("el nido no tiene portions")
	}
	// La gota tenía 3 porciones y drinker tomó al menos 1 (pueden quedar 1 o 2).
	if p.Get(ai.ResourceWater) >= 3 {
		t.Fatalf("esperaba consumo de agua del nido (<3), hay %d", p.Get(ai.ResourceWater))
	}
	// Verificar continuidad desde snapshot a mitad del escenario.
	spec, err := scenario.ReadFile(filepath.Join("..", "..", "tests", "scenarios", "water_hauled_to_nest.json"))
	if err != nil {
		t.Fatal(err)
	}
	inA, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	inA.World.StepN(400)
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
		t.Fatalf("continuidad desde snapshot rota: %x vs %x", inA.World.Hash(), wB.Hash())
	}
}

// M61: la giantess se lava las manos en el lavamanos y deja gotas con vencimiento.
func TestM61GiantessWashesHands(t *testing.T) {
	in := run(t, "giantess_washes_hands.json", nil)
	r := in.World.ECS()
	count := 0
	ecs.Get[colony.WaterDrop](r).Each(func(_ ids.ID, _ *colony.WaterDrop) { count++ })
	if count < 2 {
		t.Fatalf("esperaba al menos 2 gotas tras wash_hands, encontré %d", count)
	}
}

// M61: la giantess se ducha y deja muchas gotas de agua y un charco.
func TestM61GiantessTakesShower(t *testing.T) {
	in := run(t, "giantess_takes_shower.json", nil)
	r := in.World.ECS()
	count := 0
	ecs.Get[colony.WaterDrop](r).Each(func(_ ids.ID, _ *colony.WaterDrop) { count++ })
	if count < 6 {
		t.Fatalf("esperaba al menos 6 gotas/charco tras shower, encontré %d", count)
	}
}

// M61: en casa.json la giantess tiene necesidad hygiene y puede usar la ducha
// y el lavamanos cuando su higiene está baja.
func TestM61CasaGiantessUsesShowerAndWash(t *testing.T) {
	spec, err := scenario.ReadFile(filepath.Join("..", "..", "content", "partidas", "casa.json"))
	if err != nil {
		t.Fatal(err)
	}
	spec.Ticks = 1200
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	g := in.Refs["giantess"]
	n := ecs.Get[needs.Needs](in.World.ECS()).Ptr(g)
	if n == nil {
		t.Fatal("giantess sin needs en casa.json")
	}
	if _, ok := n.Get("hygiene"); !ok {
		t.Fatal("giantess no tiene necesidad hygiene en casa.json")
	}
	// Bajar hygiene al mínimo y dejar otras necesidades altas para que vaya al baño.
	for i := range *n {
		switch (*n)[i].Name {
		case "hygiene":
			(*n)[i].Value = 50
		default:
			(*n)[i].Value = 950
		}
	}
	if _, err := in.Run(); err != nil {
		t.Fatal(err)
	}
	washed := false
	var gEvents []string
	for _, ev := range in.Events {
		switch e := ev.Event.(type) {
		case ai.ActionStarted:
			if e.Entity == g {
				gEvents = append(gEvents, "start:"+e.Action+":"+e.What)
			}
		case ai.ActionDone:
			if e.Entity == g {
				gEvents = append(gEvents, "done:"+e.Action+":"+e.What)
				if e.What == "shower" || e.What == "wash_hands" || e.What == "wash_dishes" {
					washed = true
				}
			}
		case ai.ActionFailed:
			if e.Entity == g {
				gEvents = append(gEvents, "fail:"+e.Action+":"+e.What)
			}
		}
	}
	if !washed {
		t.Fatalf("en casa.json la giantess con higiene baja no usó ducha/lavamanos/bacha; eventos=%v", gEvents)
	}
	drops := 0
	ecs.Get[colony.WaterDrop](in.World.ECS()).Each(func(_ ids.ID, _ *colony.WaterDrop) { drops++ })
	if drops < 6 {
		t.Fatalf("esperaba al menos 6 gotas/charco tras ducharse en casa.json, encontré %d (eventos=%v)", drops, gEvents)
	}
}

// M61: en una partida libre (casa.json) la giantess llega a usar la ducha
// (no solo lavarse las manos) y los tinies acarrean agua y beben sin quedar
// todos secos.
func TestM61CasaShowerAndWaterLoop(t *testing.T) {
	base, err := scenario.ReadFile(filepath.Join("..", "..", "content", "partidas", "casa.json"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := scenario.Randomize(base, 1)
	if err != nil {
		t.Fatal(err)
	}
	spec.Ticks = 12000
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := in.Run(); err != nil {
		t.Fatal(err)
	}
	showers, washes, waterStored, drinks := 0, 0, 0, 0
	for _, ev := range in.Events {
		switch e := ev.Event.(type) {
		case ai.ActionDone:
			switch e.What {
			case "shower":
				showers++
			case "wash_hands", "wash_dishes":
				washes++
			case "drink":
				drinks++
			}
		case colony.Stored:
			if e.Resource == ai.ResourceWater {
				waterStored++
			}
		}
	}
	nestWater := int32(0)
	if p, ok := ecs.Get[ai.Portions](in.World.ECS()).Get(in.Refs["nest"]); ok {
		nestWater = p.Get(ai.ResourceWater)
	}
	alive, thirsty := 0, 0
	ecs.Get[colony.Member](in.World.ECS()).Each(func(id ids.ID, _ *colony.Member) {
		if in.World.ECS().Despawning(id) {
			return
		}
		alive++
		if n, ok := ecs.Get[needs.Needs](in.World.ECS()).Get(id); ok {
			if th, ok := n.Get("thirst"); ok && th < 300 {
				thirsty++
			}
		}
	})
	t.Logf("casa seed 1 (12000 ticks): showers=%d washes=%d waterStored=%d drinks=%d nestWater=%d thirsty=%d/%d",
		showers, washes, waterStored, drinks, nestWater, thirsty, alive)
	if showers == 0 {
		t.Errorf("la giantess nunca usó la ducha en 12000 ticks (washes=%d)", washes)
	}
	if waterStored == 0 {
		t.Errorf("los tinies nunca acarrearon agua al nido en 12000 ticks")
	}
	if drinks == 0 {
		t.Errorf("ningún tiny bebió agua en 12000 ticks")
	}
	if alive > 0 && thirsty*2 > alive {
		t.Errorf("más de la mitad de la colonia quedó sedienta (%d/%d con thirst < 300)", thirsty, alive)
	}
}

// M61: Deficit y Stock funcionan por recurso independiente (food y water).
func TestM61DeficitByResource(t *testing.T) {
	in := run(t, "tiny_drinks_drop.json", func(s *scenario.Spec) { s.Ticks = 1 })
	w := in.World
	nest := in.Refs["nest"]
	p := ecs.Get[ai.Portions](w.ECS()).Ptr(nest)
	p.Set(ai.ResourceFood, 2)
	p.Set(ai.ResourceWater, 1)
	if got := colony.Deficit(w, nest, ai.ResourceFood); got != 0 {
		t.Errorf("déficit food = %v, quería 0", got)
	}
	if got := colony.Deficit(w, nest, ai.ResourceWater); got != 0.5 {
		t.Errorf("déficit water = %v, quería 0.5", got)
	}
}
