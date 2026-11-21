// simbench mide el costo por tick de una multitud de tinies caminando
// (scenario.Crowd): A*, movimiento e índice espacial.
//
//	go run ./cmd/simbench -n 200 -ticks 2000
package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"time"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/clock"
)

func main() {
	n := flag.Int("n", 200, "cantidad de tinies")
	ticks := flag.Int("ticks", 2000, "ticks a medir")
	seed := flag.Uint64("seed", 1, "semilla")
	flag.Parse()

	c, err := scenario.NewCrowd(*n, *seed)
	if err != nil {
		fmt.Fprintln(os.Stderr, "simbench:", err)
		os.Exit(1)
	}
	// El reloj de Windows tiene ~1 ms de resolución: se mide por ventanas de
	// un segundo simulado (clock.TPS ticks) y se reporta ms/tick de cada una.
	var windows []float64
	arrived := 0
	begin := time.Now()
	for t := 0; t < *ticks; t += clock.TPS {
		start := time.Now()
		for range clock.TPS {
			c.Step()
			for _, e := range c.World.Events() {
				if e.Event.Kind() == "move_arrived" {
					arrived++
				}
			}
		}
		windows = append(windows, float64(time.Since(start).Microseconds())/1000/clock.TPS)
	}
	total := time.Since(begin)
	if err := c.World.CheckInvariants(); err != nil {
		fmt.Fprintln(os.Stderr, "simbench:", err)
		os.Exit(1)
	}
	slices.Sort(windows)
	fmt.Printf("n=%d ticks=%d llegadas=%d  ms/tick: avg=%.3f  por segundo simulado: p50=%.3f p99=%.3f peor=%.3f  hash=%#016x\n",
		*n, len(windows)*clock.TPS, arrived, float64(total.Microseconds())/1000/float64(len(windows)*clock.TPS),
		windows[len(windows)/2], windows[len(windows)*99/100], windows[len(windows)-1], c.World.Hash())
}
