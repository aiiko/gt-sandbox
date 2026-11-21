package ids

import (
	"math/rand/v2"
	"testing"

	"gtsandbox/internal/core/statehash"
)

func TestMakeRoundTrip(t *testing.T) {
	id := Make(123, 456)
	if id.Index() != 123 || id.Gen() != 456 {
		t.Fatalf("round trip: %v", id)
	}
	if !Nil.IsNil() || id.IsNil() {
		t.Fatal("IsNil")
	}
}

func TestStaleIDDetected(t *testing.T) {
	var a Allocator
	x := a.Alloc()
	if !a.Free(x) {
		t.Fatal("Free de ID vivo debe devolver true")
	}
	if a.Alive(x) {
		t.Fatal("ID liberado sigue vivo")
	}
	if a.Free(x) {
		t.Fatal("doble Free debe devolver false")
	}
	y := a.Alloc()
	if y.Index() != x.Index() {
		t.Fatalf("se esperaba reutilizar el slot %d, vino %v", x.Index(), y)
	}
	if y == x || a.Alive(x) || !a.Alive(y) {
		t.Fatalf("referencia vieja no detectada: x=%v y=%v", x, y)
	}
}

func TestNilAndForeignNeverAlive(t *testing.T) {
	var a Allocator
	a.Alloc()
	if a.Alive(Nil) || a.Alive(Make(99, 1)) || a.Alive(Make(0, 7)) {
		t.Fatal("IDs inválidos reportados como vivos")
	}
}

func TestFIFOReuse(t *testing.T) {
	var a Allocator
	ids := []ID{a.Alloc(), a.Alloc(), a.Alloc()}
	a.Free(ids[2])
	a.Free(ids[0])
	if got := a.Alloc().Index(); got != 2 {
		t.Fatalf("FIFO: esperado slot 2, vino %d", got)
	}
	if got := a.Alloc().Index(); got != 0 {
		t.Fatalf("FIFO: esperado slot 0, vino %d", got)
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

// Modelo de referencia: el allocator nunca debe tener dos IDs vivos iguales,
// ni reportar vivo algo liberado, ni perder la cuenta.
func TestRandomizedAgainstModel(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	var a Allocator
	live := map[ID]bool{}
	var dead []ID
	var order []ID
	for step := 0; step < 50_000; step++ {
		if len(order) == 0 || r.IntN(3) != 0 {
			id := a.Alloc()
			if live[id] {
				t.Fatalf("step %d: ID duplicado %v", step, id)
			}
			live[id] = true
			order = append(order, id)
		} else {
			i := r.IntN(len(order))
			id := order[i]
			order[i] = order[len(order)-1]
			order = order[:len(order)-1]
			delete(live, id)
			dead = append(dead, id)
			if !a.Free(id) {
				t.Fatalf("step %d: Free falló para %v", step, id)
			}
		}
		if a.Len() != len(live) {
			t.Fatalf("step %d: Len=%d, modelo=%d", step, a.Len(), len(live))
		}
	}
	for _, id := range dead {
		if a.Alive(id) {
			t.Fatalf("ID muerto %v reportado vivo", id)
		}
	}
	count := 0
	prev := -1
	a.Each(func(id ID) {
		if !live[id] {
			t.Fatalf("Each devolvió ID no vivo %v", id)
		}
		if int(id.Index()) <= prev {
			t.Fatal("Each no está en orden ascendente")
		}
		prev = int(id.Index())
		count++
	})
	if count != len(live) {
		t.Fatalf("Each recorrió %d, esperado %d", count, len(live))
	}
}

func TestHashDeterministic(t *testing.T) {
	build := func() uint64 {
		var a Allocator
		x := a.Alloc()
		a.Alloc()
		a.Free(x)
		h := statehash.New()
		a.HashState(h)
		return h.Sum()
	}
	if build() != build() {
		t.Fatal("hash no determinista")
	}
}

func FuzzAllocator(f *testing.F) {
	f.Add([]byte{0, 0, 1, 0, 1, 1})
	f.Fuzz(func(t *testing.T, ops []byte) {
		var a Allocator
		var live []ID
		for _, op := range ops {
			if op%2 == 0 || len(live) == 0 {
				live = append(live, a.Alloc())
				continue
			}
			i := int(op/2) % len(live)
			if !a.Free(live[i]) {
				t.Fatalf("Free falló para %v", live[i])
			}
			if a.Alive(live[i]) {
				t.Fatalf("%v vivo tras Free", live[i])
			}
			live = append(live[:i], live[i+1:]...)
		}
		for _, id := range live {
			if !a.Alive(id) {
				t.Fatalf("%v debería estar vivo", id)
			}
		}
	})
}
