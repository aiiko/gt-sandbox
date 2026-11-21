package world

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/clock"
)

// Command es la única forma de que algo externo (jugador, UI, tests,
// escenarios) modifique el mundo (D7). Se encola con Enqueue y se aplica al
// inicio del próximo tick, en orden de llegada.
//
// Apply puede emitir eventos. Si devuelve error, el comando no debe haber
// modificado nada y se emite CommandRejected.
type Command interface {
	Kinded
	Apply(w *World) error
}

// Event describe algo que ocurrió durante un tick. Se publica al final del
// tick y lo ven los sistemas del tick siguiente y los observadores externos.
type Event interface {
	Kinded
}

// EventRecord es un evento publicado con su posición en el tiempo.
type EventRecord struct {
	Tick  clock.Tick
	Seq   uint32 // orden dentro del tick
	Event Event
}

// CommandRecord es un comando aplicado, tal como quedó en el log de replay.
type CommandRecord struct {
	Tick clock.Tick `json:"tick"`
	Encoded
}

type queuedCommand struct {
	seq uint64
	cmd Command
}

// EntityDespawned se emite cuando una entidad se destruye en el Flush.
type EntityDespawned struct {
	Entity ids.ID `json:"entity"`
}

func (EntityDespawned) Kind() string { return "entity_despawned" }

// CommandRejected se emite cuando un comando falla al aplicarse.
type CommandRejected struct {
	Seq     uint64 `json:"seq"`
	Command string `json:"command"`
	Error   string `json:"error"`
}

func (CommandRejected) Kind() string { return "command_rejected" }

// Enqueue agrega un comando para el próximo tick. Es seguro llamarlo desde
// otra goroutine (por ejemplo, el loop de render). Devuelve el número de
// secuencia asignado.
func (w *World) Enqueue(c Command) (uint64, error) {
	if !w.commandKinds.has(c.Kind()) {
		return 0, fmt.Errorf("world: comando %q no registrado", c.Kind())
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	seq := w.nextCmdSeq
	w.nextCmdSeq++
	w.queue = append(w.queue, queuedCommand{seq: seq, cmd: c})
	return seq, nil
}

// PendingCommands devuelve cuántos comandos esperan el próximo tick.
func (w *World) PendingCommands() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.queue)
}

func (w *World) applyCommands() {
	w.mu.Lock()
	q := w.queue
	w.queue = nil
	w.mu.Unlock()

	for _, qc := range q {
		if w.recording {
			enc, err := w.commandKinds.encode(qc.cmd)
			if err != nil {
				panic(err) // registrado en Enqueue: solo falla si no es serializable
			}
			w.cmdLog = append(w.cmdLog, CommandRecord{Tick: w.tick, Encoded: enc})
		}
		if err := qc.cmd.Apply(w); err != nil {
			w.Emit(CommandRejected{Seq: qc.seq, Command: qc.cmd.Kind(), Error: err.Error()})
		}
	}
}

// Emit publica un evento. Solo puede llamarse durante Step; el tipo de evento
// debe estar registrado (para poder guardarlo).
func (w *World) Emit(e Event) {
	if !w.stepping {
		panic("world: Emit fuera de Step")
	}
	if !w.eventKinds.has(e.Kind()) {
		panic(fmt.Sprintf("world: evento %q no registrado", e.Kind()))
	}
	w.pendingEvents = append(w.pendingEvents, EventRecord{
		Tick: w.tick, Seq: uint32(len(w.pendingEvents)), Event: e,
	})
}

// Events devuelve los eventos publicados en el último tick completado.
// Durante Step, son los del tick anterior. El slice es de solo lectura.
func (w *World) Events() []EventRecord { return w.published }

// SetRecording activa o desactiva el log de comandos aplicados.
func (w *World) SetRecording(on bool) { w.recording = on }

// CommandLog devuelve una copia del log de comandos aplicados.
func (w *World) CommandLog() []CommandRecord { return slices.Clone(w.cmdLog) }

// DecodeCommand reconstruye un comando de un registro serializado.
func (w *World) DecodeCommand(e Encoded) (Command, error) { return w.commandKinds.decode(e) }

// Replay reproduce un log de comandos hasta el tick `until`. Los registros
// anteriores al tick actual se saltean (ya están reflejados en el estado, p.
// ej. al continuar desde un snapshot). El mundo no debe tener comandos
// encolados: se mezclarían con los del log.
func Replay(w *World, log []CommandRecord, until clock.Tick) error {
	if w.PendingCommands() > 0 {
		return errors.New("world: Replay con comandos encolados")
	}
	if !slices.IsSortedFunc(log, func(a, b CommandRecord) int { return cmp.Compare(a.Tick, b.Tick) }) {
		return errors.New("world: log de replay desordenado")
	}
	cmds := make([]Command, len(log))
	for i, r := range log {
		c, err := w.DecodeCommand(r.Encoded)
		if err != nil {
			return fmt.Errorf("world: registro %d: %w", i, err)
		}
		cmds[i] = c
	}
	i := 0
	for i < len(log) && log[i].Tick < w.tick {
		i++
	}
	for w.tick < until {
		for i < len(log) && log[i].Tick == w.tick {
			if _, err := w.Enqueue(cmds[i]); err != nil {
				return err
			}
			i++
		}
		w.Step()
	}
	return nil
}

// Stateful lo implementan los sistemas que guardan estado propio que afecta
// la simulación. Ese estado entra en Hash() y en los snapshots.
type Stateful interface {
	SaveState() (json.RawMessage, error)
	LoadState(json.RawMessage) error
}
