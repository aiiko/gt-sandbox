// simbench mide el costo por tick de una multitud o un escenario completo.
//
//	go run ./cmd/simbench -n 200 -ticks 2000
//	go run ./cmd/simbench -scenario tests/scenarios/una_manana_en_la_casa.json
package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"time"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/world"
)

func main() {
	n := flag.Int("n", 200, "cantidad de tinies")
	ticks := flag.Int("ticks", 0, "ticks a medir (por defecto: 2000 o la duración del escenario)")
	seed := flag.Uint64("seed", 1, "semilla")
	path := flag.String("scenario", "", "escenario JSON a medir")
	flag.Parse()

	var seedOverride *uint64
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			seedOverride = seed
		}
	})
	if err := run(*path, *n, *ticks, seedOverride); err != nil {
		fmt.Fprintln(os.Stderr, "simbench:", err)
		os.Exit(1)
	}
}

func run(path string, n, ticks int, seed *uint64) error {
	if ticks < 0 {
		return fmt.Errorf("ticks debe ser >= 0")
	}
	var w *world.World
	var step func() error
	var label string
	if path != "" {
		spec, err := scenario.ReadFile(path)
		if err != nil {
			return err
		}
		if seed != nil {
			spec.Seed = *seed
		}
		in, err := scenario.Build(spec)
		if err != nil {
			return err
		}
		p, err := in.Player()
		if err != nil {
			return err
		}
		w, step = in.World, p.Step
		label = fmt.Sprintf("escenario=%q seed=%d", spec.Name, spec.Seed)
		if ticks == 0 {
			ticks = spec.Ticks
		}
	} else {
		crowdSeed := uint64(1)
		if seed != nil {
			crowdSeed = *seed
		}
		c, err := scenario.NewCrowd(n, crowdSeed)
		if err != nil {
			return err
		}
		w = c.World
		step = func() error { c.Step(); return nil }
		label = fmt.Sprintf("n=%d", n)
		if ticks == 0 {
			ticks = 2000
		}
	}
	if ticks <= 0 {
		return fmt.Errorf("se necesita al menos un tick para medir")
	}
	// El reloj de Windows tiene ~1 ms de resolución: se mide por ventanas de
	// un segundo simulado (clock.TPS ticks) y se reporta ms/tick de cada una.
	var windows []float64
	arrived := 0
	begin := time.Now()
	for t := 0; t < ticks; t += clock.TPS {
		start := time.Now()
		count := min(clock.TPS, ticks-t)
		for range count {
			if err := step(); err != nil {
				return err
			}
			for _, e := range w.Events() {
				if e.Event.Kind() == "move_arrived" {
					arrived++
				}
			}
		}
		windows = append(windows, float64(time.Since(start))/float64(time.Millisecond)/float64(count))
	}
	total := time.Since(begin)
	if err := w.CheckInvariants(); err != nil {
		return err
	}
	slices.Sort(windows)
	fmt.Printf("%s ticks=%d llegadas=%d  ms/tick: avg=%.3f  por ventana: p50=%.3f p99=%.3f peor=%.3f  hash=%#016x\n",
		label, ticks, arrived, float64(total)/float64(time.Millisecond)/float64(ticks),
		windows[len(windows)/2], windows[len(windows)*99/100], windows[len(windows)-1], w.Hash())
	return nil
}
