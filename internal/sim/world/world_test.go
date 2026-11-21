package world

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"gtsandbox/internal/sim/clock"
)

// noise consume aleatoriedad y acumula estado, para que el hash refleje algo.
type noise struct {
	name string
	Acc  uint64 `json:"acc"`
	Runs int    `json:"runs"`
}

func (n *noise) Name() string { return n.name }
func (n *noise) Run(w *World) {
	n.Acc = n.Acc*31 + w.RNG(n.name).Uint64()
	n.Runs++
}
func (n *noise) SaveState() (json.RawMessage, error) { return json.Marshal(n) }
func (n *noise) LoadState(raw json.RawMessage) error { return json.Unmarshal(raw, n) }

func newNoisyWorld(seed uint64) *World {
	w := New(Config{Seed: seed})
	w.MustRegister(SystemSpec{Phase: clock.PhaseDecide}, &noise{name: "decide"})
	w.MustRegister(SystemSpec{Phase: clock.PhaseSlow, Period: 7, Offset: 3}, &noise{name: "slow"})
	return w
}

// Criterio de aceptación M0: 10.000 ticks con la misma semilla dan el mismo hash.
func TestDeterminism(t *testing.T) {
	a, b, c := newNoisyWorld(42), newNoisyWorld(42), newNoisyWorld(43)
	a.StepN(10_000)
	b.StepN(10_000)
	c.StepN(10_000)
	if a.Hash() != b.Hash() {
		t.Fatal("hash distinto con la misma semilla")
	}
	if a.Hash() == c.Hash() {
		t.Fatal("semillas distintas deberían divergir")
	}
}

type recorder struct {
	name string
	log  *[]string
}

func (r recorder) Name() string { return r.name }
func (r recorder) Run(*World)   { *r.log = append(*r.log, r.name) }

func TestExecutionOrderAndPeriod(t *testing.T) {
	var log []string
	w := New(Config{})
	w.MustRegister(SystemSpec{Phase: clock.PhaseSlow, Period: 2}, recorder{"slow", &log})
	w.MustRegister(SystemSpec{Phase: clock.PhaseMove, Order: 10}, recorder{"move10", &log})
	w.MustRegister(SystemSpec{Phase: clock.PhaseMove, Order: -1}, recorder{"move-1", &log})
	w.MustRegister(SystemSpec{Phase: clock.PhaseCommands}, recorder{"cmd", &log})
	w.StepN(2)
	want := []string{"cmd", "move-1", "move10", "slow", "cmd", "move-1", "move10"}
	if !slices.Equal(log, want) {
		t.Fatalf("orden = %v, esperado %v", log, want)
	}
}

type registerer struct{ err error }

func (r *registerer) Name() string { return "registerer" }
func (r *registerer) Run(w *World) { r.err = w.Register(SystemSpec{}, &noise{name: "late"}) }

func TestRegisterErrors(t *testing.T) {
	w := New(Config{})
	r := &registerer{}
	w.MustRegister(SystemSpec{}, r)
	if err := w.Register(SystemSpec{}, &registerer{}); !errors.Is(err, ErrDuplicateSystem) {
		t.Errorf("duplicado: %v", err)
	}
	if err := w.Register(SystemSpec{Phase: 200}, &noise{name: "b"}); !errors.Is(err, ErrInvalidPhase) {
		t.Errorf("fase inválida: %v", err)
	}
	w.Step()
	if !errors.Is(r.err, ErrDuringStep) {
		t.Errorf("durante Step: %v", r.err)
	}
}

func BenchmarkStepEmpty(b *testing.B) {
	w := New(Config{})
	for b.Loop() {
		w.Step()
	}
}
