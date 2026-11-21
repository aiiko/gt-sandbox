package world

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Kinded es cualquier valor polimórfico serializable (comandos, eventos).
// Kind debe ser un nombre estable: se usa en snapshots y logs de replay.
type Kinded interface {
	Kind() string
}

// Encoded es un valor polimórfico serializado: su kind y sus datos JSON.
type Encoded struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

// kindRegistry mapea kinds a fábricas para poder decodificar.
type kindRegistry[T Kinded] struct {
	what      string
	factories map[string]func() T
}

func newKindRegistry[T Kinded](what string) *kindRegistry[T] {
	return &kindRegistry[T]{what: what, factories: map[string]func() T{}}
}

func (r *kindRegistry[T]) add(f func() T) error {
	k := f().Kind()
	if k == "" {
		return fmt.Errorf("world: %s con kind vacío", r.what)
	}
	if _, dup := r.factories[k]; dup {
		return fmt.Errorf("world: %s %q ya registrado", r.what, k)
	}
	r.factories[k] = f
	return nil
}

func (r *kindRegistry[T]) has(kind string) bool {
	_, ok := r.factories[kind]
	return ok
}

func (r *kindRegistry[T]) encode(v T) (Encoded, error) {
	k := v.Kind()
	if !r.has(k) {
		return Encoded{}, fmt.Errorf("world: %s %q no registrado", r.what, k)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return Encoded{}, fmt.Errorf("world: %s %q: %w", r.what, k, err)
	}
	return Encoded{Kind: k, Data: b}, nil
}

func (r *kindRegistry[T]) decode(e Encoded) (T, error) {
	var zero T
	f, ok := r.factories[e.Kind]
	if !ok {
		return zero, fmt.Errorf("world: %s %q desconocido", r.what, e.Kind)
	}
	v := f()
	dec := json.NewDecoder(bytes.NewReader(e.Data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return zero, fmt.Errorf("world: %s %q: %w", r.what, e.Kind, err)
	}
	return v, nil
}

// RegisterCommand habilita el tipo de comando C (se registra *C, para poder
// decodificarlo). Uso: world.RegisterCommand[cmds.Spawn](w).
func RegisterCommand[C any, PC interface {
	*C
	Command
}](w *World) error {
	return w.commandKinds.add(func() Command { return PC(new(C)) })
}

// RegisterEvent habilita el tipo de evento E. Uso: world.RegisterEvent[MyEvent](w).
func RegisterEvent[E any, PE interface {
	*E
	Event
}](w *World) error {
	return w.eventKinds.add(func() Event { return PE(new(E)) })
}
