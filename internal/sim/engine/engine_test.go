package engine

import (
	"bytes"
	"encoding/json"
	"math/rand/v2"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// wander mueve cada Location un paso aleatorio por tick y cuenta los pasos
// (estado propio del sistema).
type wander struct{ steps uint64 }

func (*wander) Name() string { return "wander" }
func (s *wander) Run(w *world.World) {
	r := w.RNG("wander")
	comp.Of(w.ECS()).Location.Each(func(_ ids.ID, l *comp.Location) {
		l.Pos.X += units.Length(r.IntN(2001) - 1000)
		s.steps++
	})
}
func (s *wander) SaveState() (json.RawMessage, error) { return json.Marshal(s.steps) }
func (s *wander) LoadState(raw json.RawMessage) error { return json.Unmarshal(raw, &s.steps) }

func withWander(w *world.World) error {
	return w.Register(world.SystemSpec{Phase: clock.PhaseMove}, &wander{})
}

// randomCommand simula un jugador que arma y desarma jerarquías de Spaces;
// incluye comandos inválidos (ciclos, entidades muertas) a propósito.
func randomCommand(r *rand.Rand, w *world.World) world.Command {
	var live, spaces []ids.ID
	g := space.Of(w.ECS())
	w.ECS().Each(func(id ids.ID) {
		live = append(live, id)
		if g.IsSpace(id) {
			spaces = append(spaces, id)
		}
	})
	if len(spaces) == 0 {
		return cmds.Spawn{Space: &space.Space{}} // raíz
	}
	pick := func(xs []ids.ID) ids.ID { return xs[r.IntN(len(xs))] }
	loc := &comp.Location{Space: pick(spaces), Pos: units.V3(units.Length(r.IntN(1e6)), 0, 0)}
	switch r.IntN(6) {
	case 0, 1:
		return cmds.Spawn{Location: loc}
	case 2:
		return cmds.Spawn{Location: loc, Space: &space.Space{Kind: space.Rigid}}
	case 3:
		return cmds.Reparent{Entity: pick(live), Space: pick(spaces)}
	case 4:
		if r.IntN(4) == 0 {
			return cmds.Despawn{Entity: pick(live)}
		}
		return cmds.Despawn{Entity: ids.Make(999, 1)}
	default:
		return cmds.SetLocation{Entity: pick(live), Location: *loc}
	}
}

// Criterios de aceptación de M1: sin referencias colgantes, replay reproduce
// el estado, y save -> load -> continuar da el mismo hash.
func TestM1Acceptance(t *testing.T) {
	const total, saveAt = 300, 120
	a := MustNew(world.Config{Seed: 2026}, withWander)
	a.SetRecording(true)
	r := rand.New(rand.NewPCG(1, 1))
	var snap bytes.Buffer
	var saveHash uint64
	for a.Tick() < total {
		if a.Tick() == saveAt {
			s, err := a.Save()
			if err != nil {
				t.Fatal(err)
			}
			world.WriteSnapshot(&snap, s)
			saveHash = a.Hash()
		}
		for range r.IntN(4) {
			if _, err := a.Enqueue(randomCommand(r, a)); err != nil {
				t.Fatal(err)
			}
		}
		a.Step()
		if err := a.CheckInvariants(); err != nil {
			t.Fatalf("t%d: %v", a.Tick(), err)
		}
	}
	log := a.CommandLog()

	b := MustNew(world.Config{Seed: 2026}, withWander)
	if err := world.Replay(b, log, total); err != nil || b.Hash() != a.Hash() {
		t.Fatalf("replay no reproduce el estado final (%v)", err)
	}

	c, err := Load(&snap, withWander)
	if err != nil || c.Hash() != saveHash {
		t.Fatalf("estado cargado distinto al guardado (%v)", err)
	}
	if err := world.Replay(c, log, total); err != nil || c.Hash() != a.Hash() {
		t.Fatalf("load + replay no reproduce el estado final (%v)", err)
	}

	if _, err := Restore(mustSave(t, a)); err == nil {
		t.Fatal("cargar sin el sistema con estado debe fallar")
	}
}

func mustSave(t *testing.T, w *world.World) *world.Snapshot {
	t.Helper()
	s, err := w.Save()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestCommandValidation(t *testing.T) {
	w := MustNew(world.Config{})
	w.Enqueue(cmds.Spawn{Name: "a"}) // sin Location: válido
	w.Step()
	id := w.Events()[0].Event.(cmds.EntitySpawned).Entity

	for _, c := range []world.Command{
		cmds.Despawn{Entity: id},
		cmds.Despawn{Entity: id},                  // ya marcada
		cmds.SetLocation{Entity: ids.Make(42, 1)}, // no existe
		cmds.Spawn{Body: &comp.Body{Radius: -1}},  // inválido
	} {
		w.Enqueue(c)
	}
	w.Step()
	rejected := 0
	for _, e := range w.Events() {
		if _, ok := e.Event.(world.CommandRejected); ok {
			rejected++
		}
	}
	if rejected != 3 || w.ECS().Len() != 0 {
		t.Fatalf("rechazados=%d vivas=%d", rejected, w.ECS().Len())
	}
}

func TestAcceptanceIsNotTrivial(t *testing.T) {
	// Guardia: si los comandos aleatorios se rechazaran todos, el test de
	// aceptación pasaría sin probar nada.
	w := MustNew(world.Config{Seed: 3})
	r := rand.New(rand.NewPCG(1, 1))
	reparented, spaces := 0, 0
	for range 200 {
		w.Enqueue(randomCommand(r, w))
		w.Step()
		for _, e := range w.Events() {
			if _, ok := e.Event.(space.EntityReparented); ok {
				reparented++
			}
		}
	}
	w.ECS().Each(func(id ids.ID) {
		if space.Of(w.ECS()).IsSpace(id) {
			spaces++
		}
	})
	if reparented < 10 || spaces < 3 || w.ECS().Len() < 20 {
		t.Fatalf("escenario trivial: reparents=%d spaces=%d entidades=%d", reparented, spaces, w.ECS().Len())
	}
}
