package world

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"

	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/ecs"
)

// SnapshotSchema es la versión del formato de Snapshot. Incrementarlo ante
// cambios incompatibles y agregar una migración en ReadSnapshot.
const SnapshotSchema = 15

// Snapshot es el estado completo y serializable de un mundo entre ticks.
//
// No incluye la "forma" del mundo (qué sistemas, stores, comandos y eventos
// existen): eso es código. Para restaurar se construye un mundo con la misma
// forma y se llama a Restore.
type Snapshot struct {
	Schema          int                        `json:"schema"`
	Seed            uint64                     `json:"seed"`
	Tick            clock.Tick                 `json:"tick"`
	RNG             map[string][]byte          `json:"rng"`
	ECS             ecs.State                  `json:"ecs"`
	Systems         map[string]json.RawMessage `json:"systems"`
	Events          []SavedEvent               `json:"events"`
	NextCommandSeq  uint64                     `json:"next_command_seq"`
	PendingCommands []SavedCommand             `json:"pending_commands"`
}

// SavedEvent es un evento publicado en un snapshot.
type SavedEvent struct {
	Tick clock.Tick `json:"tick"`
	Seq  uint32     `json:"seq"`
	Encoded
}

// SavedCommand es un comando encolado no aplicado todavía.
type SavedCommand struct {
	Seq uint64 `json:"seq"`
	Encoded
}

// Save captura el estado del mundo. Solo puede llamarse entre ticks.
func (w *World) Save() (*Snapshot, error) {
	if w.stepping {
		return nil, errors.New("world: Save durante Step")
	}
	st, err := w.ecs.Save()
	if err != nil {
		return nil, err
	}
	s := &Snapshot{
		Schema:  SnapshotSchema,
		Seed:    w.seed,
		Tick:    w.tick,
		RNG:     w.rng.State(),
		ECS:     st,
		Systems: map[string]json.RawMessage{},
		Events:  []SavedEvent{},
	}
	for _, r := range w.systems {
		if sf, ok := r.sys.(Stateful); ok {
			raw, err := sf.SaveState()
			if err != nil {
				return nil, fmt.Errorf("world: sistema %q: %w", r.sys.Name(), err)
			}
			s.Systems[r.sys.Name()] = raw
		}
	}
	for _, e := range w.published {
		enc, err := w.eventKinds.encode(e.Event)
		if err != nil {
			return nil, err
		}
		s.Events = append(s.Events, SavedEvent{Tick: e.Tick, Seq: e.Seq, Encoded: enc})
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	s.NextCommandSeq = w.nextCmdSeq
	s.PendingCommands = []SavedCommand{}
	for _, qc := range w.queue {
		enc, err := w.commandKinds.encode(qc.cmd)
		if err != nil {
			return nil, err
		}
		s.PendingCommands = append(s.PendingCommands, SavedCommand{Seq: qc.seq, Encoded: enc})
	}
	return s, nil
}

// Restore carga un snapshot en un mundo recién creado (tick 0, vacío) que
// tenga la misma forma: mismos stores, sistemas con estado, comandos y eventos.
// La semilla de la configuración se reemplaza por la del snapshot.
func (w *World) Restore(s *Snapshot) error {
	if w.stepping {
		return errors.New("world: Restore durante Step")
	}
	if s.Schema != SnapshotSchema {
		return fmt.Errorf("world: schema de snapshot %d, se esperaba %d", s.Schema, SnapshotSchema)
	}
	if w.tick != 0 || w.ecs.Len() != 0 || len(w.published) != 0 || w.PendingCommands() != 0 {
		return errors.New("world: Restore requiere un mundo recién creado")
	}

	stateful := map[string]Stateful{}
	for _, r := range w.systems {
		if sf, ok := r.sys.(Stateful); ok {
			stateful[r.sys.Name()] = sf
		}
	}
	for name := range stateful {
		if _, ok := s.Systems[name]; !ok {
			return fmt.Errorf("world: el snapshot no tiene estado del sistema %q", name)
		}
	}
	for name := range s.Systems {
		if _, ok := stateful[name]; !ok {
			return fmt.Errorf("world: el snapshot tiene estado de un sistema desconocido %q", name)
		}
	}

	events := make([]EventRecord, len(s.Events))
	for i, se := range s.Events {
		e, err := w.eventKinds.decode(se.Encoded)
		if err != nil {
			return err
		}
		events[i] = EventRecord{Tick: se.Tick, Seq: se.Seq, Event: e}
	}
	queue := make([]queuedCommand, len(s.PendingCommands))
	for i, sc := range s.PendingCommands {
		c, err := w.commandKinds.decode(sc.Encoded)
		if err != nil {
			return err
		}
		queue[i] = queuedCommand{seq: sc.Seq, cmd: c}
	}
	if !slices.IsSortedFunc(queue, func(a, b queuedCommand) int { return cmp.Compare(a.seq, b.seq) }) {
		return errors.New("world: comandos pendientes desordenados")
	}

	if err := w.rng.Restore(s.Seed, s.RNG); err != nil {
		return err
	}
	if err := w.ecs.Restore(s.ECS); err != nil {
		return err
	}
	// Orden estable para que un error parcial sea reproducible.
	names := make([]string, 0, len(stateful))
	for n := range stateful {
		names = append(names, n)
	}
	slices.Sort(names)
	for _, n := range names {
		if err := stateful[n].LoadState(s.Systems[n]); err != nil {
			return fmt.Errorf("world: sistema %q: %w", n, err)
		}
	}

	w.seed = s.Seed
	w.tick = s.Tick
	w.published = events
	w.mu.Lock()
	w.queue = queue
	w.nextCmdSeq = s.NextCommandSeq
	w.mu.Unlock()
	return nil
}

// WriteSnapshot serializa un snapshot como JSON indentado (legible y
// diffeable; el formato binario llegará cuando el tamaño lo justifique).
func WriteSnapshot(out io.Writer, s *Snapshot) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

// ReadSnapshot lee un snapshot JSON, rechazando campos desconocidos.
func ReadSnapshot(in io.Reader) (*Snapshot, error) {
	dec := json.NewDecoder(in)
	dec.DisallowUnknownFields()
	var s Snapshot
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("world: snapshot inválido: %w", err)
	}
	return &s, nil
}
