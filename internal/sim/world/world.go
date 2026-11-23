// Package world contiene el estado autoritativo de la simulación y su loop.
//
// Reglas (ver CLAUDE.md):
//   - El estado solo cambia dentro de Step(), a través de comandos y sistemas (D7).
//   - Los sistemas corren secuencialmente en un orden total y estable (D5, D8):
//     fase, luego Order, luego orden de registro.
//   - Toda aleatoriedad sale de w.RNG(nombre) (D4).
//
// Un tick (Step) hace, en orden:
//  1. aplica los comandos encolados (en orden de secuencia);
//  2. corre los sistemas que tocan este tick, fase por fase;
//  3. destruye las entidades marcadas para despawn (emite EntityDespawned);
//  4. publica los eventos del tick (visibles en Events() hasta el próximo Step).
package world

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"

	"gtsandbox/internal/core/rng"
	"gtsandbox/internal/core/statehash"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/ecs"
)

// HashVersion identifica el formato de Hash(). Incrementarlo si cambia qué se
// hashea o cómo, para que los golden files viejos fallen de forma explícita.
const HashVersion = 10

// System es una unidad de lógica que corre en una fase del tick.
//
// Si el sistema guarda estado propio que afecte la simulación, debe
// implementar Stateful para que entre en Hash() y en los snapshots.
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
	ecs      *ecs.Registry
	systems  []registered
	names    map[string]bool
	stepping bool

	commandKinds *kindRegistry[Command]
	eventKinds   *kindRegistry[Event]

	mu         sync.Mutex // protege queue y nextCmdSeq (Enqueue es concurrente)
	queue      []queuedCommand
	nextCmdSeq uint64

	pendingEvents []EventRecord
	published     []EventRecord

	recording bool
	cmdLog    []CommandRecord

	invariants []invariant
	resources  map[reflect.Type]any
}

type invariant struct {
	name  string
	check func(*World) error
}

// New crea un mundo vacío con los eventos del núcleo registrados.
func New(cfg Config) *World {
	w := &World{
		seed:         cfg.Seed,
		rng:          rng.NewSet(cfg.Seed),
		ecs:          ecs.NewRegistry(),
		names:        map[string]bool{},
		commandKinds: newKindRegistry[Command]("comando"),
		eventKinds:   newKindRegistry[Event]("evento"),
	}
	must(RegisterEvent[EntityDespawned](w))
	must(RegisterEvent[CommandRejected](w))
	return w
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	ErrDuplicateSystem = errors.New("world: sistema duplicado")
	ErrInvalidPhase    = errors.New("world: fase inválida")
	ErrDuringStep      = errors.New("world: no se puede registrar durante Step")
)

// Register agrega un sistema. Debe hacerse antes de empezar a simular; los
// nombres son únicos y estables (se usan en snapshots).
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
	must(w.Register(spec, sys))
}

// Step avanza la simulación exactamente un tick.
func (w *World) Step() {
	if w.stepping {
		panic("world: Step reentrante")
	}
	w.stepping = true
	defer func() { w.stepping = false }()

	w.applyCommands()
	for _, r := range w.systems {
		if clock.Due(w.tick, r.spec.Period, r.spec.Offset) {
			r.sys.Run(w)
		}
	}
	for _, id := range w.ecs.Flush() {
		w.Emit(EntityDespawned{Entity: id})
	}
	w.published = w.pendingEvents
	w.pendingEvents = nil
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

// ECS devuelve el registro de entidades y componentes.
func (w *World) ECS() *ecs.Registry { return w.ecs }

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

// AddInvariant registra un chequeo de consistencia que CheckInvariants
// ejecutará (en orden de registro). Pensado para que cada subsistema declare
// sus propias invariantes.
func (w *World) AddInvariant(name string, check func(*World) error) {
	w.invariants = append(w.invariants, invariant{name, check})
}

// CheckInvariants verifica la consistencia interna del mundo. Es costoso:
// para tests y builds de debug.
func (w *World) CheckInvariants() error {
	if err := w.ecs.CheckInvariants(); err != nil {
		return err
	}
	for _, inv := range w.invariants {
		if err := inv.check(w); err != nil {
			return fmt.Errorf("invariante %s: %w", inv.name, err)
		}
	}
	return nil
}

// Hash devuelve un hash del estado lógico completo. Dos mundos con la misma
// semilla, los mismos sistemas y los mismos inputs deben tener el mismo hash
// en el mismo tick. Los comandos encolados todavía no aplicados no cuentan.
func (w *World) Hash() uint64 {
	h := statehash.New()
	h.Int(HashVersion)
	h.U64(w.seed)
	h.U64(uint64(w.tick))
	w.rng.HashState(h)
	w.ecs.HashState(h)
	w.mu.Lock()
	h.U64(w.nextCmdSeq)
	w.mu.Unlock()
	h.Int(len(w.published))
	for _, e := range w.published {
		enc, err := w.eventKinds.encode(e.Event)
		if err != nil {
			panic(err)
		}
		h.U64(uint64(e.Tick))
		h.U32(e.Seq)
		h.String(enc.Kind)
		h.Bytes(enc.Data)
	}
	for _, r := range w.systems {
		if st, ok := r.sys.(Stateful); ok {
			raw, err := st.SaveState()
			if err != nil {
				panic(fmt.Sprintf("world: sistema %q: %v", r.sys.Name(), err))
			}
			h.String(r.sys.Name())
			h.Bytes(raw)
		}
	}
	return h.Sum()
}
