package world

import (
	"bytes"
	"errors"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/ecs"
)

// El replay y save/load con comandos reales se prueban en engine; acá solo la
// semántica de comandos, eventos y snapshot.

type counter struct {
	V int `json:"v"`
}

type spawnCounter struct {
	V int `json:"v"`
}

func (spawnCounter) Kind() string { return "spawn_counter" }
func (c spawnCounter) Apply(w *World) error {
	if c.V < 0 {
		return errors.New("negativo")
	}
	id := w.ECS().Spawn()
	ecs.Get[counter](w.ECS()).Set(id, counter{c.V})
	w.Emit(counterSpawned{Entity: id})
	return nil
}

type counterSpawned struct {
	Entity ids.ID `json:"entity"`
}

func (counterSpawned) Kind() string { return "counter_spawned" }

// incr suma 1 a cada contador y cuenta los eventos que ve.
type incr struct{ seen int }

func (*incr) Name() string { return "incr" }
func (s *incr) Run(w *World) {
	s.seen += len(w.Events())
	ecs.Get[counter](w.ECS()).Each(func(_ ids.ID, c *counter) { c.V += 1 + w.RNG("incr").IntN(3) })
}

func newTestWorld(t *testing.T, seed uint64) (*World, *incr) {
	t.Helper()
	w := New(Config{Seed: seed})
	ecs.MustRegister[counter](w.ECS(), "counter")
	if err := errors.Join(RegisterCommand[spawnCounter](w), RegisterEvent[counterSpawned](w)); err != nil {
		t.Fatal(err)
	}
	s := &incr{}
	w.MustRegister(SystemSpec{Phase: clock.PhaseAct}, s)
	return w, s
}

func enqueue(t *testing.T, w *World, c Command) {
	t.Helper()
	if _, err := w.Enqueue(c); err != nil {
		t.Fatal(err)
	}
}

func TestCommandsAndEvents(t *testing.T) {
	w, sys := newTestWorld(t, 1)
	enqueue(t, w, spawnCounter{V: 10})
	enqueue(t, w, spawnCounter{V: -1})
	if w.ECS().Len() != 0 {
		t.Fatal("el comando no debe aplicarse antes de Step")
	}
	w.Step()
	evs := w.Events()
	if w.ECS().Len() != 1 || len(evs) != 2 || sys.seen != 0 {
		t.Fatalf("t0: vivas=%d eventos=%+v seen=%d", w.ECS().Len(), evs, sys.seen)
	}
	if rej, ok := evs[1].Event.(CommandRejected); !ok || rej.Command != "spawn_counter" {
		t.Fatalf("se esperaba CommandRejected, vino %+v", evs[1].Event)
	}

	// Los sistemas ven los eventos del tick anterior; los eventos duran un tick.
	id := evs[0].Event.(counterSpawned).Entity
	w.ECS().Despawn(id)
	w.Step()
	if sys.seen != 2 || len(w.Events()) != 1 || w.Events()[0].Event != (EntityDespawned{Entity: id}) {
		t.Fatalf("t1: seen=%d eventos=%+v", sys.seen, w.Events())
	}

	if _, err := New(Config{}).Enqueue(spawnCounter{}); err == nil {
		t.Error("comando no registrado debe fallar en Enqueue")
	}
	defer func() {
		if recover() == nil {
			t.Error("Emit fuera de Step debe hacer panic")
		}
	}()
	w.Emit(counterSpawned{})
}

func TestSnapshotRoundTrip(t *testing.T) {
	a, _ := newTestWorld(t, 77)
	enqueue(t, a, spawnCounter{V: 1})
	a.StepN(5)
	enqueue(t, a, spawnCounter{V: 100}) // queda pendiente en el snapshot

	snap, err := a.Save()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := WriteSnapshot(&buf, snap); err != nil {
		t.Fatal(err)
	}
	back, err := ReadSnapshot(&buf)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := newTestWorld(t, 0)
	if err := b.Restore(back); err != nil {
		t.Fatal(err)
	}
	if a.Hash() != b.Hash() || b.PendingCommands() != 1 {
		t.Fatal("estado restaurado distinto")
	}
	a.StepN(10)
	b.StepN(10)
	if a.Hash() != b.Hash() {
		t.Fatal("divergen tras continuar desde el snapshot")
	}

	used, _ := newTestWorld(t, 1)
	used.Step()
	if used.Restore(snap) == nil {
		t.Error("Restore sobre mundo usado debe fallar")
	}
	if New(Config{}).Restore(snap) == nil {
		t.Error("Restore sobre mundo con otra forma debe fallar")
	}
	if _, err := ReadSnapshot(bytes.NewReader([]byte(`{"schema":1,"bogus":true}`))); err == nil {
		t.Error("campo desconocido debe fallar")
	}
}

// Un evento publicado que viaja en el snapshot vuelve con su tipo de valor,
// igual que el emitido: los sistemas del tick siguiente lo reconocen.
func TestRestoredEventsKeepValueType(t *testing.T) {
	a, _ := newTestWorld(t, 3)
	enqueue(t, a, spawnCounter{V: 1})
	a.Step()
	snap, err := a.Save()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := newTestWorld(t, 0)
	if err := b.Restore(snap); err != nil {
		t.Fatal(err)
	}
	evs := b.Events()
	if len(evs) != 1 {
		t.Fatalf("eventos restaurados: %v", evs)
	}
	if _, ok := evs[0].Event.(counterSpawned); !ok {
		t.Fatalf("evento restaurado de tipo %T", evs[0].Event)
	}
}
