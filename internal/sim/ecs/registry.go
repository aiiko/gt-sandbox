// Package ecs administra entidades y sus componentes (decisión D9).
//
// Es una implementación propia y simple: un allocator de IDs generacionales y
// un Store por tipo de componente. Los sistemas solo usan esta API, así que
// podría reemplazarse por un ECS por arquetipos sin tocarlos.
//
// Despawn es diferido: la entidad sigue viva (y con sus componentes) hasta
// Flush, que el World llama al final de cada tick. Así las iteraciones en
// curso nunca se invalidan.
package ecs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/statehash"
)

type anyStore interface {
	Name() string
	Len() int
	Has(ids.ID) bool
	remove(ids.ID) bool
	hashState(h *statehash.Hasher)
	encode() (json.RawMessage, error)
	decode(json.RawMessage) error
	decodeValue(json.RawMessage) (func(ids.ID), error)
	check() error
}

// Registry contiene todas las entidades y stores de un mundo.
type Registry struct {
	alloc      ids.Allocator
	stores     []anyStore // en orden de registro
	byName     map[string]anyStore
	byType     map[reflect.Type]anyStore
	pending    []ids.ID // despawns en orden de pedido
	pendingSet map[ids.ID]struct{}
}

func NewRegistry() *Registry {
	return &Registry{
		byName:     map[string]anyStore{},
		byType:     map[reflect.Type]anyStore{},
		pendingSet: map[ids.ID]struct{}{},
	}
}

// Register crea el store del componente T con un nombre único y estable (se
// usa en snapshots: cambiarlo rompe partidas guardadas).
func Register[T any](r *Registry, name string) (*Store[T], error) {
	t := reflect.TypeFor[T]()
	if name == "" {
		return nil, errors.New("ecs: nombre de store vacío")
	}
	if _, dup := r.byName[name]; dup {
		return nil, fmt.Errorf("ecs: store %q ya registrado", name)
	}
	if _, dup := r.byType[t]; dup {
		return nil, fmt.Errorf("ecs: tipo %v ya registrado", t)
	}
	var zero T
	if _, err := json.Marshal(zero); err != nil {
		return nil, fmt.Errorf("ecs: componente %v no serializable: %w", t, err)
	}
	s := &Store[T]{name: name, reg: r}
	r.stores = append(r.stores, s)
	r.byName[name] = s
	r.byType[t] = s
	return s, nil
}

// MustRegister es Register con panic. Para setup.
func MustRegister[T any](r *Registry, name string) *Store[T] {
	s, err := Register[T](r, name)
	if err != nil {
		panic(err)
	}
	return s
}

// Lookup devuelve el store del componente T, si está registrado.
func Lookup[T any](r *Registry) (*Store[T], bool) {
	s, ok := r.byType[reflect.TypeFor[T]()]
	if !ok {
		return nil, false
	}
	return s.(*Store[T]), true
}

// Get devuelve el store del componente T. Panic si no está registrado.
// Los sistemas deberían llamarlo una vez en su setup y guardar el puntero.
func Get[T any](r *Registry) *Store[T] {
	s, ok := Lookup[T](r)
	if !ok {
		panic(fmt.Sprintf("ecs: componente %v no registrado", reflect.TypeFor[T]()))
	}
	return s
}

// StoreNames devuelve los nombres de los stores en orden de registro.
func (r *Registry) StoreNames() []string {
	out := make([]string, len(r.stores))
	for i, s := range r.stores {
		out[i] = s.Name()
	}
	return out
}

// DecodeComponent decodifica un valor JSON para el store `name` y devuelve
// una función que lo asigna a una entidad. Permite crear entidades con
// componentes arbitrarios desde datos (comandos, escenarios) validando todo
// antes de modificar nada.
func (r *Registry) DecodeComponent(name string, raw json.RawMessage) (func(ids.ID), error) {
	s, ok := r.byName[name]
	if !ok {
		return nil, fmt.Errorf("ecs: store %q desconocido", name)
	}
	return s.decodeValue(raw)
}

// Spawn crea una entidad sin componentes.
func (r *Registry) Spawn() ids.ID { return r.alloc.Alloc() }

// Alive indica si la entidad existe (incluye las marcadas para despawn).
func (r *Registry) Alive(id ids.ID) bool { return r.alloc.Alive(id) }

// Despawn marca la entidad para destruirse en el próximo Flush. Devuelve
// false si no está viva o ya estaba marcada.
func (r *Registry) Despawn(id ids.ID) bool {
	if !r.Alive(id) {
		return false
	}
	if _, dup := r.pendingSet[id]; dup {
		return false
	}
	r.pendingSet[id] = struct{}{}
	r.pending = append(r.pending, id)
	return true
}

// Despawning indica si la entidad está marcada para despawn.
func (r *Registry) Despawning(id ids.ID) bool {
	_, ok := r.pendingSet[id]
	return ok
}

// PendingDespawns devuelve cuántas entidades esperan Flush.
func (r *Registry) PendingDespawns() int { return len(r.pending) }

// PendingAt devuelve la i-ésima entidad marcada, en orden de pedido. Permite
// recorrer la lista mientras crece (despawns en cascada).
func (r *Registry) PendingAt(i int) ids.ID { return r.pending[i] }

// Flush destruye las entidades marcadas, en orden de pedido, quitando sus
// componentes de todos los stores. Devuelve las entidades destruidas.
func (r *Registry) Flush() []ids.ID {
	if len(r.pending) == 0 {
		return nil
	}
	done := r.pending
	for _, id := range done {
		for _, s := range r.stores {
			s.remove(id)
		}
		r.alloc.Free(id)
	}
	r.pending = nil
	clear(r.pendingSet)
	return done
}

// Len devuelve la cantidad de entidades vivas.
func (r *Registry) Len() int { return r.alloc.Len() }

// Each recorre las entidades vivas en orden de índice.
func (r *Registry) Each(fn func(ids.ID)) { r.alloc.Each(fn) }

// HashState escribe entidades, stores y despawns pendientes.
func (r *Registry) HashState(h *statehash.Hasher) {
	r.alloc.HashState(h)
	h.Int(len(r.stores))
	for _, s := range r.stores {
		s.hashState(h)
	}
	h.Int(len(r.pending))
	for _, id := range r.pending {
		h.U64(uint64(id))
	}
}

// CheckInvariants verifica la consistencia interna. Pensado para tests y
// builds de debug; es O(entidades).
func (r *Registry) CheckInvariants() error {
	for _, s := range r.stores {
		if err := s.check(); err != nil {
			return err
		}
	}
	for _, id := range r.pending {
		if !r.Alive(id) {
			return fmt.Errorf("ecs: despawn pendiente de entidad muerta %v", id)
		}
	}
	if len(r.pending) != len(r.pendingSet) {
		return errors.New("ecs: pending y pendingSet desincronizados")
	}
	return nil
}

// State es el estado serializable de un Registry.
type State struct {
	Entities   ids.AllocatorState         `json:"entities"`
	Components map[string]json.RawMessage `json:"components"`
}

// Save serializa entidades y componentes. Falla si hay despawns pendientes:
// solo se guarda entre ticks, cuando ya se hizo Flush.
func (r *Registry) Save() (State, error) {
	if len(r.pending) > 0 {
		return State{}, errors.New("ecs: no se puede guardar con despawns pendientes")
	}
	st := State{Entities: r.alloc.State(), Components: map[string]json.RawMessage{}}
	for _, s := range r.stores {
		raw, err := s.encode()
		if err != nil {
			return State{}, fmt.Errorf("ecs: store %s: %w", s.Name(), err)
		}
		st.Components[s.Name()] = raw
	}
	return st, nil
}

// Restore carga un estado guardado. El registry debe estar vacío y tener
// registrados exactamente los mismos stores que el estado.
func (r *Registry) Restore(st State) error {
	if r.Len() != 0 || len(r.pending) != 0 {
		return errors.New("ecs: Restore requiere un registry vacío")
	}
	if err := sameKeys(r.StoreNames(), st.Components); err != nil {
		return err
	}
	if err := r.alloc.Restore(st.Entities); err != nil {
		return err
	}
	for _, s := range r.stores {
		if err := s.decode(st.Components[s.Name()]); err != nil {
			return err
		}
	}
	return r.CheckInvariants()
}

func sameKeys(names []string, m map[string]json.RawMessage) error {
	for _, n := range names {
		if _, ok := m[n]; !ok {
			return fmt.Errorf("ecs: el estado no tiene el store %q", n)
		}
	}
	if len(m) != len(names) {
		var extra []string
		for k := range m {
			if !slices.Contains(names, k) {
				extra = append(extra, k)
			}
		}
		slices.Sort(extra)
		return fmt.Errorf("ecs: el estado tiene stores desconocidos: %v", extra)
	}
	return nil
}

func strictUnmarshal(raw []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
