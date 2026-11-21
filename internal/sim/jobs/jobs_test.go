package jobs_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/world"
)

// script corre una función dentro de Step (las reservas cambian estado).
type script struct{ fn *func(w *world.World) }

func (script) Name() string { return "test.script" }
func (s script) Run(w *world.World) {
	if *s.fn != nil {
		(*s.fn)(w)
		*s.fn = nil
	}
}

func released(w *world.World) string {
	for _, e := range w.Events() {
		if r, ok := e.Event.(jobs.ReservationReleased); ok {
			return r.Reason
		}
	}
	return ""
}

// Una reserva es exclusiva, se renueva, vence sin renovar y se libera si
// muere quien la tiene.
func TestReservations(t *testing.T) {
	var fn func(w *world.World)
	w := engine.MustNew(world.Config{Seed: 1}, func(w *world.World) error {
		return w.Register(world.SystemSpec{Phase: clock.PhaseAct}, script{&fn})
	})
	do := func(f func(w *world.World)) { fn = f; w.Step() }
	r := w.ECS()
	crumb, a, b := r.Spawn(), r.Spawn(), r.Spawn()
	do(func(w *world.World) {
		if !jobs.Reserve(w, crumb, a) || jobs.Reserve(w, crumb, b) || !jobs.Reserve(w, crumb, a) {
			t.Fatal("la reserva no es exclusiva o no se renueva")
		}
	})
	for range jobs.TTL {
		if w.Step(); released(w) != "" {
			t.Fatalf("venció antes de tiempo, en %v", w.Tick())
		}
	}
	if w.Step(); released(w) != "expired" {
		t.Fatal("no venció")
	}
	do(func(w *world.World) { jobs.Reserve(w, crumb, b); r.Despawn(b) })
	if released(w) != "holder_gone" || jobs.Holder(w, crumb) != ids.Nil {
		t.Fatal("no se liberó al morir quien la tenía")
	}
}
