// simrun ejecuta la simulación headless y reporta el hash del estado final.
//
// En M0 solo corre un mundo vacío; a partir de M3 cargará escenarios.
//
//	go run ./cmd/simrun -seed 42 -ticks 10000
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/world"
)

func main() {
	seed := flag.Uint64("seed", 1, "semilla del mundo")
	ticks := flag.Int("ticks", 1000, "cantidad de ticks a simular")
	flag.Parse()

	if *ticks < 0 {
		fmt.Fprintln(os.Stderr, "ticks debe ser >= 0")
		os.Exit(2)
	}

	w := world.New(world.Config{Seed: *seed})

	start := time.Now()
	w.StepN(*ticks)
	elapsed := time.Since(start)

	fmt.Printf("seed=%d ticks=%d simulado=%v real=%v hash=%#016x\n",
		w.Seed(), w.Tick(), clock.Tick(*ticks).Duration(), elapsed, w.Hash())
}
