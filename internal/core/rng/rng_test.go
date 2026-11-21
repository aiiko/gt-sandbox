package rng

import (
	"testing"

	"gtsandbox/internal/core/statehash"
)

func draw(s *Stream, n int) []uint64 {
	out := make([]uint64, n)
	for i := range out {
		out[i] = s.Uint64()
	}
	return out
}

func equal(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSameSeedSameSequence(t *testing.T) {
	if !equal(draw(NewStream(42, "ai"), 100), draw(NewStream(42, "ai"), 100)) {
		t.Fatal("misma semilla y nombre deben dar la misma secuencia")
	}
}

func TestStreamsIndependent(t *testing.T) {
	if equal(draw(NewStream(42, "ai"), 10), draw(NewStream(42, "physics"), 10)) {
		t.Fatal("streams distintos no deberían coincidir")
	}
	if equal(draw(NewStream(1, "ai"), 10), draw(NewStream(2, "ai"), 10)) {
		t.Fatal("semillas distintas no deberían coincidir")
	}
}

// Agregar un stream nuevo o crearlos en otro orden no altera a los demás.
func TestCreationOrderIrrelevant(t *testing.T) {
	a := NewSet(7)
	aAI := a.Get("ai")
	a.Get("weather")
	b := NewSet(7)
	b.Get("weather")
	b.Get("new_system")
	bAI := b.Get("ai")
	if !equal(draw(aAI, 50), draw(bAI, 50)) {
		t.Fatal("el orden de creación cambió la secuencia")
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	s := NewStream(9, "x")
	draw(s, 13)
	state, err := s.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	want := draw(s, 20)
	r := NewStream(0, "x")
	if err := r.UnmarshalBinary(state); err != nil {
		t.Fatal(err)
	}
	if !equal(draw(r, 20), want) {
		t.Fatal("restaurar el estado no reproduce la secuencia")
	}
}

func TestSetHashTracksState(t *testing.T) {
	hash := func(s *Set) uint64 {
		h := statehash.New()
		s.HashState(h)
		return h.Sum()
	}
	a, b := NewSet(3), NewSet(3)
	a.Get("ai")
	b.Get("ai")
	if hash(a) != hash(b) {
		t.Fatal("estados iguales, hashes distintos")
	}
	a.Get("ai").Uint64()
	if hash(a) == hash(b) {
		t.Fatal("consumir un número debe cambiar el hash")
	}
}

func TestSetStateRestore(t *testing.T) {
	a := NewSet(11)
	a.Get("ai").Uint64()
	a.Get("phys").IntN(10)
	b := NewSet(0)
	if err := b.Restore(a.Seed(), a.State()); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"ai", "phys", "new"} {
		if !equal(draw(a.Get(n), 10), draw(b.Get(n), 10)) {
			t.Fatalf("stream %q divergió tras restaurar", n)
		}
	}
	if err := b.Restore(1, map[string][]byte{"x": {1, 2}}); err == nil {
		t.Fatal("estado corrupto debería fallar")
	}
}
