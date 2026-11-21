// Package world contiene el estado autoritativo de la simulación y su loop.
//
// Reglas (ver CLAUDE.md):
//   - El estado solo cambia dentro de Step(), a través de sistemas (D7).
//   - Los sistemas corren secuencialmente en un orden total y estable (D5, D8):
//     fase, luego Order, luego orden de registro.
//   - Toda aleatoriedad sale de w.RNG(nombre) (D4).
package world

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"gtsandbox/internal/core/rng"
	"gtsandbox/internal/core/statehash"
	"gtsandbox/internal/sim/clock"
)

// HashVersion identifica el formato de Hash(). Incrementarlo si cambia qué se
// hashea o cómo, para que los golden files viejos fallen de forma explícita.
const HashVersion = 1

// System es una unidad de lógica que corre en una fase del tick.
//
// Si el sistema guarda estado propio que afecte la simulación, debe implementar
// también statehash.Hashable para que entre en World.Hash().
type System interface {
	Name() string
	Run(w *World)
}

// SystemSpec indica cuándo corre un sistema.
type SystemSpec struct {
	Phase clock.Phase
	// Order desempata sistemas de la misma fase (menor primero).
	Order int
	// Period en ticks; 0 o 1 = todos los ticks.
	Period clock.Tick
	// Offset desplaza la ejecución periódica para repartir carga.
	Offset clock.Tick
}

type registered struct {
	spec SystemSpec
	sys  System
	seq  int
}

// Config es la configuración inicial de un mundo.
type Config struct {
	Seed uint64
}

// World es el estado completo de la simulación.
type World struct {
	seed     uint64
	tick     clock.Tick
	rng      *rng.Set
	systems  []registered
	names    map[string]bool
	stepping bool
}

// New crea un mundo vacío.
func New(cfg Config) *World {
	return &World{
		seed:  cfg.Seed,
		rng:   rng.NewSet(cfg.Seed),
		names: map[string]bool{},
	}
}

var (
	ErrDuplicateSystem = errors.New("world: sistema duplicado")
	ErrInvalidPhase    = errors.New("world: fase inválida")
	ErrDuringStep      = errors.New("world: no se puede registrar durante Step")
)

// Register agrega un sistema. Debe hacerse antes de empezar a simular; los
// nombres son únicos.
func (w *World) Register(spec SystemSpec, sys System) error {
	if w.stepping {
		return ErrDuringStep
	}
	if !spec.Phase.Valid() {
		return fmt.Errorf("%w: %v", ErrInvalidPhase, spec.Phase)
	}
	name := sys.Name()
	if w.names[name] {
		return fmt.Errorf("%w: %q", ErrDuplicateSystem, name)
	}
	w.names[name] = true
	w.systems = append(w.systems, registered{spec: spec, sys: sys, seq: len(w.systems)})
	slices.SortStableFunc(w.systems, func(a, b registered) int {
		return cmp.Or(
			cmp.Compare(a.spec.Phase, b.spec.Phase),
			cmp.Compare(a.spec.Order, b.spec.Order),
			cmp.Compare(a.seq, b.seq),
		)
	})
	return nil
}

// MustRegister es Register que hace panic ante error. Para setup.
func (w *World) MustRegister(spec SystemSpec, sys System) {
	if err := w.Register(spec, sys); err != nil {
		panic(err)
	}
}

// Step avanza la simulación exactamente un tick.
func (w *World) Step() {
	if w.stepping {
		panic("world: Step reentrante")
	}
	w.stepping = true
	defer func() { w.stepping = false }()

	for _, r := range w.systems {
		if clock.Due(w.tick, r.spec.Period, r.spec.Offset) {
			r.sys.Run(w)
		}
	}
	w.tick++
}

// StepN avanza n ticks.
func (w *World) StepN(n int) {
	for range n {
		w.Step()
	}
}

// Tick devuelve el tick en curso. Durante Step es el tick que se está
// procesando; afuera, la cantidad de ticks completados.
func (w *World) Tick() clock.Tick { return w.tick }

// Seed devuelve la semilla del mundo.
func (w *World) Seed() uint64 { return w.seed }

// RNG devuelve el stream determinista `name`. Por convención, cada sistema usa
// un stream con su propio nombre.
func (w *World) RNG(name string) *rng.Stream { return w.rng.Get(name) }

// Systems devuelve los nombres de los sistemas en orden de ejecución.
func (w *World) Systems() []string {
	out := make([]string, len(w.systems))
	for i, r := range w.systems {
		out[i] = r.sys.Name()
	}
	return out
}

// Hash devuelve un hash del estado lógico completo. Dos mundos con la misma
// semilla, los mismos sistemas y los mismos inputs deben tener el mismo hash
// en el mismo tick.
func (w *World) Hash() uint64 {
	h := statehash.New()
	h.Int(HashVersion)
	h.U64(w.seed)
	h.U64(uint64(w.tick))
	w.rng.HashState(h)
	for _, r := range w.systems {
		if hs, ok := r.sys.(statehash.Hashable); ok {
			h.String(r.sys.Name())
			hs.HashState(h)
		}
	}
	return h.Sum()
}
