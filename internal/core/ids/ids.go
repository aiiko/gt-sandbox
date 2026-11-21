// Package ids implementa identificadores generacionales (decisión D3).
//
// Un ID combina un índice de slot (32 bits bajos) y una generación (32 bits
// altos). Cuando un slot se libera, su generación se incrementa: cualquier ID
// viejo que apunte a ese slot deja de estar vivo y se detecta, en lugar de
// referirse silenciosamente a otra entidad.
//
// La generación nunca vale 0, así que el ID cero (Nil) nunca es válido.
package ids

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gtsandbox/internal/core/statehash"
)

// ID es un identificador generacional.
type ID uint64

// Nil es el ID inválido.
const Nil ID = 0

// Make construye un ID a partir de índice y generación.
func Make(index, gen uint32) ID { return ID(uint64(gen)<<32 | uint64(index)) }

func (id ID) Index() uint32 { return uint32(id) }
func (id ID) Gen() uint32   { return uint32(id >> 32) }
func (id ID) IsNil() bool   { return id == Nil }

func (id ID) String() string {
	if id.IsNil() {
		return "nil"
	}
	return fmt.Sprintf("%d#%d", id.Index(), id.Gen())
}

// MarshalText serializa como "índice#generación" (o "nil"), legible en
// snapshots y logs, y sin problemas de precisión en JSON.
func (id ID) MarshalText() ([]byte, error) { return []byte(id.String()), nil }

// UnmarshalText es la inversa de MarshalText.
func (id *ID) UnmarshalText(b []byte) error {
	s := string(b)
	if s == "nil" {
		*id = Nil
		return nil
	}
	is, gs, ok := strings.Cut(s, "#")
	if !ok {
		return fmt.Errorf("ids: formato inválido %q (esperado índice#gen)", s)
	}
	idx, err := strconv.ParseUint(is, 10, 32)
	if err != nil {
		return fmt.Errorf("ids: índice inválido en %q: %w", s, err)
	}
	gen, err := strconv.ParseUint(gs, 10, 32)
	if err != nil {
		return fmt.Errorf("ids: generación inválida en %q: %w", s, err)
	}
	if gen == 0 {
		return fmt.Errorf("ids: generación 0 inválida en %q", s)
	}
	*id = Make(uint32(idx), uint32(gen))
	return nil
}

// Allocator reparte IDs generacionales de forma determinista.
//
// Los slots liberados se reutilizan en orden FIFO: así un slot tarda lo máximo
// posible en reaparecer, lo que hace más visibles los bugs de referencias
// viejas en debug.
type Allocator struct {
	gens  []uint32 // generación actual de cada slot
	alive []bool
	free  []uint32 // cola FIFO de slots libres
	head  int      // inicio de la cola dentro de free
	live  int
}

// Alloc devuelve un ID nuevo y vivo.
func (a *Allocator) Alloc() ID {
	if a.head < len(a.free) {
		idx := a.free[a.head]
		a.head++
		a.compact()
		a.alive[idx] = true
		a.live++
		return Make(idx, a.gens[idx])
	}
	idx := uint32(len(a.gens))
	if uint64(idx) == 1<<32-1 {
		panic("ids: se agotaron los índices")
	}
	a.gens = append(a.gens, 1)
	a.alive = append(a.alive, true)
	a.live++
	return Make(idx, 1)
}

// Free libera id. Devuelve false si ya no estaba vivo.
func (a *Allocator) Free(id ID) bool {
	if !a.Alive(id) {
		return false
	}
	idx := id.Index()
	a.alive[idx] = false
	g := a.gens[idx] + 1
	if g == 0 { // wrap: la generación 0 está reservada
		g = 1
	}
	a.gens[idx] = g
	a.free = append(a.free, idx)
	a.live--
	return true
}

// Alive indica si id fue emitido por este allocator y sigue vivo.
func (a *Allocator) Alive(id ID) bool {
	idx := id.Index()
	return !id.IsNil() && int(idx) < len(a.gens) && a.alive[idx] && a.gens[idx] == id.Gen()
}

// Len devuelve la cantidad de IDs vivos.
func (a *Allocator) Len() int { return a.live }

// Each recorre los IDs vivos en orden de índice ascendente (orden estable, D5).
func (a *Allocator) Each(fn func(ID)) {
	for i, ok := range a.alive {
		if ok {
			fn(Make(uint32(i), a.gens[i]))
		}
	}
}

// HashState escribe el estado completo del allocator en h.
func (a *Allocator) HashState(h *statehash.Hasher) {
	h.Int(len(a.gens))
	for i, g := range a.gens {
		h.U32(g)
		h.Bool(a.alive[i])
	}
	h.Int(len(a.free) - a.head)
	for _, idx := range a.free[a.head:] {
		h.U32(idx)
	}
}

// AllocatorState es el estado serializable de un Allocator.
type AllocatorState struct {
	Gens  []uint32 `json:"gens"`
	Alive []bool   `json:"alive"`
	Free  []uint32 `json:"free"` // en orden FIFO
}

// State devuelve una copia del estado interno.
func (a *Allocator) State() AllocatorState {
	return AllocatorState{
		Gens:  append([]uint32{}, a.gens...),
		Alive: append([]bool{}, a.alive...),
		Free:  append([]uint32{}, a.free[a.head:]...),
	}
}

// Restore reemplaza el estado del allocator, validando su consistencia:
// generaciones no nulas y cada slot muerto presente exactamente una vez en la
// cola de libres.
func (a *Allocator) Restore(st AllocatorState) error {
	n := len(st.Gens)
	if len(st.Alive) != n {
		return errors.New("ids: gens y alive tienen distinto largo")
	}
	inFree := make([]bool, n)
	for _, idx := range st.Free {
		if int(idx) >= n {
			return fmt.Errorf("ids: slot libre %d fuera de rango", idx)
		}
		if st.Alive[idx] {
			return fmt.Errorf("ids: slot %d vivo y libre a la vez", idx)
		}
		if inFree[idx] {
			return fmt.Errorf("ids: slot %d duplicado en libres", idx)
		}
		inFree[idx] = true
	}
	live := 0
	for i := range n {
		if st.Gens[i] == 0 {
			return fmt.Errorf("ids: slot %d con generación 0", i)
		}
		if st.Alive[i] {
			live++
		} else if !inFree[i] {
			return fmt.Errorf("ids: slot muerto %d no está en libres", i)
		}
	}
	*a = Allocator{
		gens:  append([]uint32{}, st.Gens...),
		alive: append([]bool{}, st.Alive...),
		free:  append([]uint32{}, st.Free...),
		live:  live,
	}
	return nil
}

// compact evita que la cola crezca sin límite cuando se consume.
func (a *Allocator) compact() {
	if a.head > 64 && a.head*2 > len(a.free) {
		n := copy(a.free, a.free[a.head:])
		a.free = a.free[:n]
		a.head = 0
	}
}
