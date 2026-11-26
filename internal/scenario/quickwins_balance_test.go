package scenario_test

import (
	"bytes"
	"maps"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/encounter"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/world"
)

func TestQuickWinsMultiSeedBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("multi-seed balance skipped in -short")
	}
	base, err := scenario.ReadFile("../../content/partidas/casa.json")
	if err != nil {
		t.Fatal(err)
	}

	seeds := []uint64{1, 7, 21, 42, 99}
	var (
		totalInitial     int
		totalAlive       int
		totalShowers     int
		totalWashes      int
		totalCookRice    int
		totalWaterPlant  int
		totalDrinks      int
		totalEncounters  int
		totalFoodStored  int
		totalWaterStored int
		totalTrinkets    int
		totalCoins       int
		totalBodyTickled int
		totalTinyReacted int
	)

	for _, seed := range seeds {
		spec, err := scenario.Randomize(base, seed)
		if err != nil {
			t.Fatalf("seed %d randomize: %v", seed, err)
		}
		spec.Ticks = 12000
		in, err := scenario.Build(spec)
		if err != nil {
			t.Fatalf("seed %d build: %v", seed, err)
		}

		initial := 0
		ecs.Get[colony.Member](in.World.ECS()).Each(func(_ ids.ID, _ *colony.Member) {
			initial++
		})

		if _, err := in.Run(); err != nil {
			t.Fatalf("seed %d run: %v", seed, err)
		}

		alive := 0
		ecs.Get[colony.Member](in.World.ECS()).Each(func(id ids.ID, _ *colony.Member) {
			if !in.World.ECS().Despawning(id) {
				alive++
			}
		})

		var (
			showers, washes, cookRice, waterPlant, drinks int
			encounters, foodStored, waterStored           int
			trinkets, coins, bodyTickled, tinyReacted     int
		)
		for _, ev := range in.Events {
			switch e := ev.Event.(type) {
			case ai.ActionDone:
				switch e.What {
				case "shower":
					showers++
				case "wash_hands", "wash_dishes":
					washes++
				case "cook_rice":
					cookRice++
				case "water_plant":
					waterPlant++
				case "drink":
					drinks++
				}
			case encounter.Observed:
				encounters++
			case encounter.Shared:
				encounters++
			case encounter.PalmWithdrawn:
				encounters++
			case encounter.PassengerDelivered:
				encounters++
			case encounter.TinyReacted:
				tinyReacted++
			case colony.Stored:
				switch e.Resource {
				case ai.ResourceFood:
					foodStored++
				case ai.ResourceWater:
					waterStored++
				}
			case colony.TrinketFound:
				trinkets++
				if e.TrinketKind == "coin" {
					coins++
				}
			case attach.BodyTickled:
				bodyTickled++
			}
		}

		t.Logf("seed %2d: alive=%d/%d showers=%d washes=%d cook_rice=%d water_plant=%d enc=%d food=%d water=%d drinks=%d trinkets=%d(coins=%d) tickle=%d react=%d",
			seed, alive, initial, showers, washes, cookRice, waterPlant, encounters, foodStored, waterStored, drinks, trinkets, coins, bodyTickled, tinyReacted)

		totalInitial += initial
		totalAlive += alive
		totalShowers += showers
		totalWashes += washes
		totalCookRice += cookRice
		totalWaterPlant += waterPlant
		totalDrinks += drinks
		totalEncounters += encounters
		totalFoodStored += foodStored
		totalWaterStored += waterStored
		totalTrinkets += trinkets
		totalCoins += coins
		totalBodyTickled += bodyTickled
		totalTinyReacted += tinyReacted
	}

	avgSurvival := 0.0
	if totalInitial > 0 {
		avgSurvival = float64(totalAlive) / float64(totalInitial)
	}
	t.Logf("aggregate (%d seeds): survival=%.1f%% (%d/%d) showers=%d washes=%d cook_rice=%d water_plant=%d enc=%d food=%d water=%d drinks=%d trinkets=%d(coins=%d) tickle=%d react=%d",
		len(seeds), avgSurvival*100, totalAlive, totalInitial, totalShowers, totalWashes, totalCookRice, totalWaterPlant,
		totalEncounters, totalFoodStored, totalWaterStored, totalDrinks, totalTrinkets, totalCoins, totalBodyTickled, totalTinyReacted)

	if avgSurvival < 0.60 {
		t.Errorf("supervivencia media baja: %.1f%% (< 60%%)", avgSurvival*100)
	}
	if totalShowers < 5 {
		t.Errorf("pocas duchas agregadas: %d (< 5)", totalShowers)
	}
	if totalFoodStored == 0 {
		t.Errorf("no se almacenó comida en las 5 semillas")
	}
	if totalWaterStored == 0 {
		t.Errorf("no se almacenó agua en las 5 semillas")
	}
	if totalDrinks == 0 {
		t.Errorf("ningún tiny bebió en las 5 semillas")
	}
	if totalEncounters == 0 {
		t.Errorf("sin encuentros registrados en las 5 semillas")
	}
	if totalCookRice+totalWaterPlant == 0 {
		t.Errorf("la giantess nunca cocinó arroz ni regó la maceta en las 5 semillas")
	}

	// Continuidad tras snapshot a mitad de partida aleatorizada (seed 7).
	spec7, err := scenario.Randomize(base, 7)
	if err != nil {
		t.Fatalf("seed 7 snapshot randomize: %v", err)
	}
	const saveTick = 1130
	const afterTicks = 400
	inA, err := scenario.Build(spec7)
	if err != nil {
		t.Fatalf("seed 7 snapshot build: %v", err)
	}
	pA, err := inA.Player()
	if err != nil {
		t.Fatal(err)
	}
	for range saveTick {
		if err := pA.Step(); err != nil {
			t.Fatal(err)
		}
	}
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
	specB := *spec7
	specB.Commands = nil
	for _, c := range spec7.Commands {
		if c.Tick >= saveTick {
			c.Tick -= saveTick
			specB.Commands = append(specB.Commands, c)
		}
	}
	inB := &scenario.Instance{Spec: &specB, World: wB, Refs: inA.Refs, Visuals: maps.Clone(inA.Visuals)}
	pB, err := inB.Player()
	if err != nil {
		t.Fatal(err)
	}
	for range afterTicks {
		if err := pA.Step(); err != nil {
			t.Fatal(err)
		}
		if err := pB.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if inA.World.Hash() != wB.Hash() {
		t.Fatalf("seed 7: continuidad rota tras snapshot en tick %d: %x vs %x", saveTick, inA.World.Hash(), wB.Hash())
	}
	if err := wB.CheckInvariants(); err != nil {
		t.Fatalf("seed 7: invariante rota tras load: %v", err)
	}
}
