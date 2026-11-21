package world

import (
	"errors"
	"testing"

	"gtsandbox/internal/core/statehash"
	"gtsandbox/internal/sim/clock"
)

// noise es un sistema de prueba que consume aleatoriedad y acumula estado,
// para que el hash refleje algo no trivial.
type noise struct {
	name string
	acc  uint64
	runs int
}

func (n *noise) Name() string { return n.name }

func (n *noise) Run(w *World) {
	r := w.RNG(n.name)
	for range 1 + r.IntN(4) {
		n.acc = n.acc*31 + r.Uint64()
	}
	n.runs++
}

func (n *noise) HashState(h *statehash.Hasher) {
	h.U64(n.acc)
	h.Int(n.runs)
}

func newNoisyWorld(seed uint64) *World {
	w := New(Config{Seed: seed})
	w.MustRegister(SystemSpec{Phase: clock.PhaseDecide}, &noise{name: "decide"})
	w.MustRegister(SystemSpec{Phase: clock.PhaseSlow, Period: 7, Offset: 3}, &noise{name: "slow"})
	return w
}

// Criterio de aceptación M0: 10.000 ticks con la misma semilla dan el mismo hash.
func TestDeterminism10kTicks(t *testing.T) {
	a, b := newNoisyWorld(42), newNoisyWorld(42)
	a.StepN(10_000)
	b.StepN(10_000)
	if a.Hash() != b.Hash() {
		t.Fatalf("hash distinto con la misma semilla: %#x vs %#x", a.Hash(), b.Hash())
	}
	if a.Tick() != 10_000 {
		t.Fatalf("Tick = %d", a.Tick())
	}
}

func TestDeterminismEmptyWorld(t *testing.T) {
	a, b := New(Config{Seed: 1}), New(Config{Seed: 1})
	a.StepN(10_000)
	b.StepN(10_000)
	if a.Hash() != b.Hash() {
		t.Fatal("mundos vacíos divergieron")
	}
}

func TestHashSensitivity(t *testing.T) {
	a, b := newNoisyWorld(1), newNoisyWorld(2)
	a.StepN(100)
	b.StepN(100)
	if a.Hash() == b.Hash() {
		t.Fatal("semillas distintas deberían divergir")
	}
	c := newNoisyWorld(1)
	c.StepN(101)
	if a.Hash() == c.Hash() {
		t.Fatal("ticks distintos deberían divergir")
	}
}

// Agregar un sistema que usa su propio stream no cambia lo que ven los demás.
func TestNewSystemDoesNotPerturbOthers(t *testing.T) {
	a := newNoisyWorld(5)
	b := newNoisyWorld(5)
	extra := &noise{name: "extra"}
	b.MustRegister(SystemSpec{Phase: clock.PhaseAct}, extra)
	a.StepN(500)
	b.StepN(500)
	decideA := findNoise(t, a, "decide")
	decideB := findNoise(t, b, "decide")
	if decideA.acc != decideB.acc {
		t.Fatal("el sistema nuevo alteró la aleatoriedad de 'decide'")
	}
}

func findNoise(t *testing.T, w *World, name string) *noise {
	t.Helper()
	for _, r := range w.systems {
		if r.sys.Name() == name {
			return r.sys.(*noise)
		}
	}
	t.Fatalf("sistema %q no encontrado", name)
	return nil
}

type recorder struct {
	name string
	log  *[]string
}

func (r recorder) Name() string { return r.name }
func (r recorder) Run(w *World) { *r.log = append(*r.log, r.name) }

func TestExecutionOrder(t *testing.T) {
	var log []string
	w := New(Config{})
	w.MustRegister(SystemSpec{Phase: clock.PhaseSlow}, recorder{"slow", &log})
	w.MustRegister(SystemSpec{Phase: clock.PhaseMove, Order: 10}, recorder{"move10", &log})
	w.MustRegister(SystemSpec{Phase: clock.PhaseMove, Order: -1}, recorder{"move-1", &log})
	w.MustRegister(SystemSpec{Phase: clock.PhaseMove, Order: 10}, recorder{"move10b", &log})
	w.MustRegister(SystemSpec{Phase: clock.PhaseCommands}, recorder{"cmd", &log})
	w.Step()
	want := []string{"cmd", "move-1", "move10", "move10b", "slow"}
	if len(log) != len(want) {
		t.Fatalf("log = %v", log)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Fatalf("orden = %v, esperado %v", log, want)
		}
	}
}

func TestPeriodicSystem(t *testing.T) {
	n := &noise{name: "p"}
	w := New(Config{})
	w.MustRegister(SystemSpec{Phase: clock.PhaseSlow, Period: 20}, n)
	w.StepN(100)
	if n.runs != 5 {
		t.Fatalf("período 20 en 100 ticks: %d ejecuciones", n.runs)
	}
}

func TestRegisterErrors(t *testing.T) {
	w := New(Config{})
	w.MustRegister(SystemSpec{}, &noise{name: "a"})
	if err := w.Register(SystemSpec{}, &noise{name: "a"}); !errors.Is(err, ErrDuplicateSystem) {
		t.Fatalf("duplicado: %v", err)
	}
	if err := w.Register(SystemSpec{Phase: clock.Phase(200)}, &noise{name: "b"}); !errors.Is(err, ErrInvalidPhase) {
		t.Fatalf("fase inválida: %v", err)
	}
}

type registerer struct{ err error }

func (r *registerer) Name() string { return "registerer" }
func (r *registerer) Run(w *World) {
	r.err = w.Register(SystemSpec{}, &noise{name: "late"})
}

func TestRegisterDuringStepRejected(t *testing.T) {
	w := New(Config{})
	r := &registerer{}
	w.MustRegister(SystemSpec{}, r)
	w.Step()
	if !errors.Is(r.err, ErrDuringStep) {
		t.Fatalf("se esperaba ErrDuringStep, vino %v", r.err)
	}
}

func BenchmarkStepEmpty(b *testing.B) {
	w := New(Config{})
	for b.Loop() {
		w.Step()
	}
}
