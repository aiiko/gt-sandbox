package ids

import (
	"math/rand/v2"
	"testing"

	"gtsandbox/internal/core/statehash"
)

func TestStaleIDDetected(t *testing.T) {
	var a Allocator
	x := a.Alloc()
	if !a.Free(x) || a.Free(x) || a.Alive(x) {
		t.Fatal("Free debe aceptar una sola vez y matar el ID")
	}
	y := a.Alloc()
	if y.Index() != x.Index() || a.Alive(x) || !a.Alive(y) {
		t.Fatalf("slot reutilizado pero ID viejo no detectado: x=%v y=%v", x, y)
	}
	if a.Alive(Nil) || a.Alive(Make(99, 1)) {
		t.Fatal("IDs inválidos reportados como vivos")
	}
}

func TestGenerationWrapSkipsZero(t *testing.T) {
	var a Allocator
	id := a.Alloc()
	a.gens[id.Index()] = ^uint32(0)
	a.Free(Make(id.Index(), ^uint32(0)))
	if g := a.Alloc().Gen(); g != 1 {
		t.Fatalf("tras wrap la generación debe ser 1, vino %d", g)
	}
}

// Contra un modelo: nunca dos IDs vivos iguales, nada liberado sigue vivo.
func TestAgainstModel(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	var a Allocator
	var live, dead []ID
	for range 5000 {
		if len(live) == 0 || r.IntN(3) != 0 {
			live = append(live, a.Alloc())
			continue
		}
		i := r.IntN(len(live))
		a.Free(live[i])
		dead = append(dead, live[i])
		live = append(live[:i], live[i+1:]...)
	}
	seen := map[ID]bool{}
	for _, id := range live {
		if seen[id] || !a.Alive(id) {
			t.Fatalf("%v duplicado o no vivo", id)
		}
		seen[id] = true
	}
	for _, id := range dead {
		if a.Alive(id) {
			t.Fatalf("%v liberado pero vivo", id)
		}
	}
	if a.Len() != len(live) {
		t.Fatalf("Len=%d, esperado %d", a.Len(), len(live))
	}
}

func TestText(t *testing.T) {
	for _, id := range []ID{Nil, Make(0, 1), Make(123, 456)} {
		b, _ := id.MarshalText()
		var got ID
		if err := got.UnmarshalText(b); err != nil || got != id {
			t.Fatalf("round trip %v -> %s -> %v (%v)", id, b, got, err)
		}
	}
	var x ID
	for _, bad := range []string{"", "3", "a#1", "3#0"} {
		if x.UnmarshalText([]byte(bad)) == nil {
			t.Errorf("%q debería ser inválido", bad)
		}
	}
}

func TestStateRestore(t *testing.T) {
	var a Allocator
	xs := []ID{a.Alloc(), a.Alloc(), a.Alloc()}
	a.Free(xs[1])
	var b Allocator
	if err := b.Restore(a.State()); err != nil {
		t.Fatal(err)
	}
	ha, hb := statehash.New(), statehash.New()
	a.HashState(ha)
	b.HashState(hb)
	if ha.Sum() != hb.Sum() || a.Alloc() != b.Alloc() {
		t.Fatal("el estado restaurado difiere")
	}

	bad := []AllocatorState{
		{Gens: []uint32{0}, Alive: []bool{true}},                    // generación 0
		{Gens: []uint32{1}, Alive: []bool{false}},                   // muerto fuera de libres
		{Gens: []uint32{1}, Alive: []bool{true}, Free: []uint32{0}}, // vivo y libre
	}
	for i, st := range bad {
		if (&Allocator{}).Restore(st) == nil {
			t.Errorf("caso %d debería fallar", i)
		}
	}
}
