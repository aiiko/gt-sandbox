package ecs

import (
	"encoding/json"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/statehash"
)

type pos struct{ X, Y int64 }
type tag struct{ Label string }

func setup() (*Registry, *Store[pos], *Store[tag]) {
	r := NewRegistry()
	return r, MustRegister[pos](r, "pos"), MustRegister[tag](r, "tag")
}

func mustPanic(t *testing.T, what string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s debe hacer panic", what)
		}
	}()
	fn()
}

func TestStore(t *testing.T) {
	r, ps, _ := setup()
	a, b := r.Spawn(), r.Spawn()
	ps.Set(a, pos{1, 2})
	ps.Set(b, pos{3, 4})
	ps.Ptr(b).X = 30
	if !ps.Remove(a) || ps.Has(a) {
		t.Fatal("Remove")
	}
	if v, _ := ps.Get(b); v != (pos{30, 4}) {
		t.Fatalf("Ptr o swap-remove corrompieron el valor: %v", v)
	}
}

func TestDeferredDespawn(t *testing.T) {
	r, ps, ts := setup()
	a := r.Spawn()
	ps.Set(a, pos{})
	ts.Set(a, tag{"x"})
	if !r.Despawn(a) || r.Despawn(a) || !ps.Has(a) {
		t.Fatal("Despawn acepta una vez y la entidad sigue viva hasta Flush")
	}
	r.Flush()
	if r.Alive(a) || ps.Has(a) || ts.Has(a) {
		t.Fatal("Flush no limpió todo")
	}
	b := r.Spawn() // reutiliza el slot con otra generación
	ps.Set(b, pos{9, 9})
	if ps.Has(a) {
		t.Fatal("ID viejo accede al componente de la entidad nueva")
	}
	mustPanic(t, "Set sobre entidad muerta", func() { ps.Set(a, pos{}) })
}

func TestEachGuards(t *testing.T) {
	r, ps, _ := setup()
	for range 5 {
		ps.Set(r.Spawn(), pos{})
	}
	ps.Each(func(id ids.ID, _ *pos) { r.Despawn(id) }) // diferido: seguro
	r.Flush()
	if ps.Len() != 0 {
		t.Fatal("despawn durante Each")
	}
	ps.Set(r.Spawn(), pos{})
	other := r.Spawn()
	mustPanic(t, "agregar durante Each", func() {
		ps.Each(func(ids.ID, *pos) { ps.Set(other, pos{}) })
	})
}

func TestRegisterErrors(t *testing.T) {
	r, _, _ := setup()
	for name, err := range map[string]error{
		"nombre duplicado": func() error { _, e := Register[struct{ A int }](r, "pos"); return e }(),
		"tipo duplicado":   func() error { _, e := Register[pos](r, "pos2"); return e }(),
		"no serializable":  func() error { _, e := Register[chan int](r, "ch"); return e }(),
	} {
		if err == nil {
			t.Errorf("%s: debería fallar", name)
		}
	}
}

func hashOf(r *Registry) uint64 {
	h := statehash.New()
	r.HashState(h)
	return h.Sum()
}

func TestSaveRestore(t *testing.T) {
	r, ps, ts := setup()
	var all []ids.ID
	for i := range 10 {
		id := r.Spawn()
		all = append(all, id)
		ps.Set(id, pos{int64(i), 0})
		if i%3 == 0 {
			ts.Set(id, tag{"t"})
		}
	}
	r.Despawn(all[2])
	r.Flush()
	ps.Remove(all[5]) // altera el orden denso

	st, err := r.Save()
	if err != nil {
		t.Fatal(err)
	}
	r2, _, _ := setup()
	if err := r2.Restore(st); err != nil || hashOf(r) != hashOf(r2) {
		t.Fatalf("restore: %v", err)
	}

	bad := func(raw string) State {
		c := State{Entities: st.Entities, Components: map[string]json.RawMessage{"tag": st.Components["tag"]}}
		c.Components["pos"] = json.RawMessage(raw)
		return c
	}
	for name, s := range map[string]State{
		"entidad inexistente": bad(`[{"id":"50#1","v":{"X":0,"Y":0}}]`),
		"campo desconocido":   bad(`[{"id":"0#1","v":{"Q":1}}]`),
		"store faltante":      {Entities: st.Entities, Components: map[string]json.RawMessage{}},
	} {
		r3, _, _ := setup()
		if r3.Restore(s) == nil {
			t.Errorf("%s: debería fallar", name)
		}
	}

	r.Despawn(all[0])
	if _, err := r.Save(); err == nil {
		t.Fatal("Save con despawns pendientes debe fallar")
	}
}

// Spawn/despawn masivo sin referencias colgantes. `go test -fuzz FuzzRegistry`
// explora más secuencias.
func FuzzRegistry(f *testing.F) {
	f.Add([]byte{0, 0, 1, 2, 3, 0, 4, 2, 5, 0, 1, 4, 5})
	f.Fuzz(func(t *testing.T, ops []byte) {
		r, ps, ts := setup()
		var live, dead []ids.ID
		for i, op := range ops {
			k := int(op >> 3)
			if op%6 != 0 && op%6 != 5 && len(live) == 0 {
				continue
			}
			switch op % 6 {
			case 0:
				live = append(live, r.Spawn())
			case 1:
				ps.Set(live[k%len(live)], pos{int64(i), 0})
			case 2:
				ts.Set(live[k%len(live)], tag{"f"})
			case 3:
				ps.Remove(live[k%len(live)])
			case 4:
				j := k % len(live)
				r.Despawn(live[j])
				dead = append(dead, live[j])
				live = append(live[:j], live[j+1:]...)
			case 5:
				r.Flush()
				for _, id := range dead {
					if r.Alive(id) || ps.Has(id) || ts.Has(id) {
						t.Fatalf("referencia colgante a %v", id)
					}
				}
				dead = dead[:0]
			}
			if err := r.CheckInvariants(); err != nil {
				t.Fatal(err)
			}
		}
	})
}

func TestComponents(t *testing.T) {
	r, ps, ts := setup()
	a := r.Spawn()
	ts.Set(a, tag{"x"})
	ps.Set(a, pos{1, 2})
	got := r.Components(a)
	if len(got) != 2 || got[0].Store != "pos" || string(got[1].JSON) != `{"Label":"x"}` {
		t.Fatalf("%+v", got)
	}
}
