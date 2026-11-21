// simrun ejecuta la simulación headless y reporta el hash del estado final.
//
//	go run ./cmd/simrun -seed 42 -ticks 1000 -spawn 100 -save out.json
//	go run ./cmd/simrun -load out.json -ticks 500
//
// A partir de M3 cargará escenarios.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/world"
)

func main() {
	seed := flag.Uint64("seed", 1, "semilla del mundo (ignorada con -load)")
	ticks := flag.Int("ticks", 1000, "cantidad de ticks a simular")
	spawn := flag.Int("spawn", 0, "entidades a crear en el primer tick (en grilla)")
	load := flag.String("load", "", "snapshot JSON desde el cual continuar")
	save := flag.String("save", "", "guardar snapshot JSON al terminar")
	flag.Parse()

	if err := run(*seed, *ticks, *spawn, *load, *save); err != nil {
		fmt.Fprintln(os.Stderr, "simrun:", err)
		os.Exit(1)
	}
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

	for i := range spawn {
		x := units.Length(i%100) * units.Centimeter
		y := units.Length(i/100) * units.Centimeter
		if _, err := w.Enqueue(cmds.Spawn{
			Name:     comp.Name(fmt.Sprintf("e%d", i)),
			Location: &comp.Location{Pos: units.V3(x, y, 0)},
		}); err != nil {
			return err
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
