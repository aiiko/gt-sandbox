package statehash

import "testing"

func TestLengthPrefixAvoidsAmbiguity(t *testing.T) {
	a, b := New(), New()
	a.String("ab")
	a.String("c")
	b.String("a")
	b.String("bc")
	if a.Sum() == b.Sum() {
		t.Fatal(`"ab","c" y "a","bc" no deberían colisionar`)
	}
}

func TestStable(t *testing.T) {
	h := New()
	h.U64(1)
	h.I64(-1)
	h.Bool(true)
	h.F64(0.5)
	h.String("x")
	// Valor fijo: si cambia, cambió el formato del hash. Eso invalida golden
	// files y replays, así que hay que incrementar world.HashVersion.
	const want uint64 = 0xb0ecd6421ba83cad
	if got := h.Sum(); got != want {
		t.Fatalf("formato de hash cambió: %#x, esperado %#x", got, want)
	}
}
