// simrun ejecuta la simulación headless y reporta el hash del estado final.
//
//	go run ./cmd/simrun -seed 42 -ticks 1000 -spawn 100 -save out.json
//	go run ./cmd/simrun -load out.json -ticks 500
//	go run ./cmd/simrun -scenario tests/scenarios/tiny_crosses_desk.json [-events]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

func main() {
	seed := flag.Uint64("seed", 1, "semilla del mundo (ignorada con -load)")
	ticks := flag.Int("ticks", 1000, "cantidad de ticks a simular")
	spawn := flag.Int("spawn", 0, "entidades a crear en el primer tick (en grilla)")
	load := flag.String("load", "", "snapshot JSON desde el cual continuar")
	save := flag.String("save", "", "guardar snapshot JSON al terminar")
	scen := flag.String("scenario", "", "correr un escenario JSON y verificar sus expectativas")
	events := flag.Bool("events", false, "con -scenario: imprimir todos los eventos")
	flag.Parse()

	if *scen != "" {
		if err := runScenario(*scen, *events); err != nil {
			fmt.Fprintln(os.Stderr, "simrun:", err)
			os.Exit(1)
		}
		return
	}
	if err := run(*seed, *ticks, *spawn, *load, *save); err != nil {
		fmt.Fprintln(os.Stderr, "simrun:", err)
		os.Exit(1)
	}
}

func runScenario(path string, events bool) error {
	spec, err := scenario.ReadFile(path)
	if err != nil {
		return err
	}
	in, err := scenario.Build(spec)
	if err != nil {
		return err
	}
	start := time.Now()
	fails, err := in.Run()
	if err != nil {
		return err
	}
	if events {
		names := map[ids.ID]string{}
		for ref, id := range in.Refs {
			names[id] = ref
		}
		for _, e := range in.Events {
			b, _ := json.Marshal(e.Event)
			s := string(b)
			for id, ref := range names {
				s = strings.ReplaceAll(s, `"`+id.String()+`"`, ref)
			}
			fmt.Printf("%5d %s %s\n", e.Tick-in.Start, e.Event.Kind(), s)
		}
	}
	fmt.Printf("%s: ticks=%d eventos=%d real=%v hash=%#016x\n",
		spec.Name, spec.Ticks, len(in.Events), time.Since(start), in.World.Hash())
	for _, f := range fails {
		fmt.Println("  FALLA:", f)
	}
	if len(fails) > 0 {
		return fmt.Errorf("%d expectativas fallaron", len(fails))
	}
	fmt.Printf("  %d expectativas OK\n", len(spec.Expect))
	return nil
}

// spawnFloor crea un Space raíz de 10×10 m (consume un tick) y devuelve su ID.
func spawnFloor(w *world.World) (ids.ID, error) {
	if _, err := w.Enqueue(cmds.Spawn{
		Name:  "suelo",
		Space: &space.Space{Kind: space.Static, Extent: units.RectWH(0, 0, 10*units.Meter, 10*units.Meter)},
	}); err != nil {
		return ids.Nil, err
	}
	w.Step()
	for _, e := range w.Events() {
		if s, ok := e.Event.(cmds.EntitySpawned); ok {
			return s.Entity, nil
		}
	}
	return ids.Nil, fmt.Errorf("no se pudo crear el suelo")
}

func run(seed uint64, ticks, spawn int, load, save string) error {
	if ticks < 0 || spawn < 0 {
		return fmt.Errorf("ticks y spawn deben ser >= 0")
	}

	var w *world.World
	var err error
	if load != "" {
		f, err := os.Open(load)
		if err != nil {
			return err
		}
		w, err = engine.Load(f)
		f.Close()
		if err != nil {
			return err
		}
	} else if w, err = engine.New(world.Config{Seed: seed}); err != nil {
		return err
	}

	if spawn > 0 {
		floor, err := spawnFloor(w)
		if err != nil {
			return err
		}
		for i := range spawn {
			x := units.Length(i%100) * units.Centimeter
			y := units.Length(i/100) * units.Centimeter
			if _, err := w.Enqueue(cmds.Spawn{
				Name:     comp.Name(fmt.Sprintf("e%d", i)),
				Location: &comp.Location{Space: floor, Pos: units.V3(x, y, 0)},
			}); err != nil {
				return err
			}
		}
	}

	start := time.Now()
	from := w.Tick()
	w.StepN(ticks)
	elapsed := time.Since(start)

	if err := w.CheckInvariants(); err != nil {
		return err
	}
	fmt.Printf("seed=%d ticks=%d..%d entidades=%d real=%v hash=%#016x\n",
		w.Seed(), from, w.Tick(), w.ECS().Len(), elapsed, w.Hash())

	if save != "" {
		s, err := w.Save()
		if err != nil {
			return err
		}
		f, err := os.Create(save)
		if err != nil {
			return err
		}
		if err := world.WriteSnapshot(f, s); err != nil {
			f.Close()
			return err
		}
		return f.Close()
	}
	return nil
}
