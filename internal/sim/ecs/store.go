package ecs

import (
	"encoding/json"
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/statehash"
)

// Store guarda un tipo de componente para muchas entidades (sparse set).
//
// Los componentes son datos puros serializables a JSON: el hash del estado y
// los snapshots se derivan de esa serialización, así que ningún campo puede
// quedar fuera por olvido. Por eso los campos deben ser exportados.
//
// El orden de iteración es el orden denso interno. Depende solo de la
// secuencia de operaciones (determinista) y se preserva en los snapshots.
type Store[T any] struct {
	name      string
	reg       *Registry
	sparse    []int32 // índice de entidad -> posición densa + 1 (0 = ausente)
	dense     []ids.ID
	vals      []T
	iterating int
}

// Name devuelve el nombre con el que se registró el store.
func (s *Store[T]) Name() string { return s.name }

// Len devuelve cuántas entidades tienen este componente.
func (s *Store[T]) Len() int { return len(s.dense) }

func (s *Store[T]) pos(id ids.ID) (int, bool) {
	idx := int(id.Index())
	if idx >= len(s.sparse) {
		return 0, false
	}
	p := int(s.sparse[idx]) - 1
	if p < 0 || s.dense[p] != id {
		return 0, false
	}
	return p, true
}

// Has indica si id tiene el componente. Un ID viejo (otra generación) nunca
// lo tiene.
func (s *Store[T]) Has(id ids.ID) bool {
	_, ok := s.pos(id)
	return ok
}

// Get devuelve una copia del componente.
func (s *Store[T]) Get(id ids.ID) (T, bool) {
	if p, ok := s.pos(id); ok {
		return s.vals[p], true
	}
	var zero T
	return zero, false
}

// Ptr devuelve un puntero al componente para modificarlo en el lugar, o nil.
// El puntero se invalida con cualquier Set de una entidad nueva o Remove.
func (s *Store[T]) Ptr(id ids.ID) *T {
	if p, ok := s.pos(id); ok {
		return &s.vals[p]
	}
	return nil
}

// Set agrega o reemplaza el componente de id. Panic si la entidad no está viva
// (es un error de programación: validar antes con Registry.Alive) o si se
// agrega una entidad nueva durante Each.
func (s *Store[T]) Set(id ids.ID, v T) {
	if !s.reg.Alive(id) {
		panic(fmt.Sprintf("ecs: %s.Set sobre entidad no viva %v", s.name, id))
	}
	if p, ok := s.pos(id); ok {
		s.vals[p] = v
		return
	}
	if s.iterating > 0 {
		panic(fmt.Sprintf("ecs: %s.Set de entidad nueva durante Each", s.name))
	}
	idx := int(id.Index())
	if idx >= len(s.sparse) {
		s.sparse = append(s.sparse, make([]int32, idx+1-len(s.sparse))...)
	}
	s.dense = append(s.dense, id)
	s.vals = append(s.vals, v)
	s.sparse[idx] = int32(len(s.dense))
}

// Remove quita el componente. Devuelve false si no estaba.
func (s *Store[T]) Remove(id ids.ID) bool {
	if s.iterating > 0 {
		panic(fmt.Sprintf("ecs: %s.Remove durante Each", s.name))
	}
	return s.remove(id)
}

func (s *Store[T]) remove(id ids.ID) bool {
	p, ok := s.pos(id)
	if !ok {
		return false
	}
	last := len(s.dense) - 1
	if p != last {
		s.dense[p] = s.dense[last]
		s.vals[p] = s.vals[last]
		s.sparse[s.dense[p].Index()] = int32(p + 1)
	}
	var zero T
	s.vals[last] = zero
	s.dense = s.dense[:last]
	s.vals = s.vals[:last]
	s.sparse[id.Index()] = 0
	return true
}

// Each recorre todas las entidades con el componente, en orden estable.
// Durante la iteración se puede modificar el valor vía el puntero, hacer Set
// sobre entidades que ya tienen el componente y marcar despawns (son
// diferidos); agregar o quitar componentes de este store hace panic.
func (s *Store[T]) Each(fn func(id ids.ID, v *T)) {
	s.iterating++
	defer func() { s.iterating-- }()
	for i := range s.dense {
		fn(s.dense[i], &s.vals[i])
	}
}

// IDs devuelve una copia de las entidades con el componente, en orden estable.
func (s *Store[T]) IDs() []ids.ID { return append([]ids.ID(nil), s.dense...) }

func (s *Store[T]) hashState(h *statehash.Hasher) {
	h.String(s.name)
	h.Int(len(s.dense))
	for i, id := range s.dense {
		h.U64(uint64(id))
		b, err := json.Marshal(s.vals[i])
		if err != nil {
			panic(fmt.Sprintf("ecs: %s no serializable: %v", s.name, err))
		}
		h.Bytes(b)
	}
}

type entry[T any] struct {
	ID ids.ID `json:"id"`
	V  T      `json:"v"`
}

func (s *Store[T]) encode() (json.RawMessage, error) {
	out := make([]entry[T], len(s.dense))
	for i, id := range s.dense {
		out[i] = entry[T]{ID: id, V: s.vals[i]}
	}
	return json.Marshal(out)
}

func (s *Store[T]) encodeOne(id ids.ID) (json.RawMessage, bool) {
	v, ok := s.Get(id)
	if !ok {
		return nil, false
	}
	raw, err := json.Marshal(v)
	if err != nil {
		raw, _ = json.Marshal(err.Error())
	}
	return raw, true
}

func (s *Store[T]) decode(raw json.RawMessage) error {
	var in []entry[T]
	if err := strictUnmarshal(raw, &in); err != nil {
		return fmt.Errorf("ecs: store %s: %w", s.name, err)
	}
	s.sparse, s.dense, s.vals = nil, nil, nil
	for _, e := range in {
		if !s.reg.Alive(e.ID) {
			return fmt.Errorf("ecs: store %s: entidad %v no viva", s.name, e.ID)
		}
		if s.Has(e.ID) {
			return fmt.Errorf("ecs: store %s: entidad %v duplicada", s.name, e.ID)
		}
		s.Set(e.ID, e.V)
	}
	return nil
}

func (s *Store[T]) decodeValue(raw json.RawMessage) (func(ids.ID), error) {
	var v T
	if err := strictUnmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("ecs: store %s: %w", s.name, err)
	}
	return func(id ids.ID) { s.Set(id, v) }, nil
}

func (s *Store[T]) check() error {
	present := 0
	for _, p := range s.sparse {
		if p != 0 {
			present++
		}
	}
	if present != len(s.dense) || len(s.vals) != len(s.dense) {
		return fmt.Errorf("ecs: store %s: sparse/dense desincronizados", s.name)
	}
	for i, id := range s.dense {
		if !s.reg.Alive(id) {
			return fmt.Errorf("ecs: store %s: componente de entidad muerta %v", s.name, id)
		}
		if int(id.Index()) >= len(s.sparse) || s.sparse[id.Index()] != int32(i+1) {
			return fmt.Errorf("ecs: store %s: índice disperso roto para %v", s.name, id)
		}
	}
	return nil
}
