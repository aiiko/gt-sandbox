package space_test

import (
	"errors"
	"testing"

	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

const (
	cm = units.Centimeter
	mm = units.Millimeter
	m  = units.Meter
)

// house arma: casa (raíz) > habitación A y B > escritorio en A;
// giantess (Body) en A > palma (Articulated) > tiny.
type house struct {
	w                                         *world.World
	g                                         space.Graph
	root, roomA, roomB, desk, gts, palm, tiny ids.ID
}

func newHouse(t *testing.T) *house {
	t.Helper()
	h := &house{w: engine.MustNew(world.Config{Seed: 1})}
	spawn := func(name string, kind space.Kind, in ids.ID, pos units.Vec3, facing fixedmath.Angle) ids.ID {
		c := cmds.Spawn{Name: comp.Name(name)}
		if kind != 255 {
			c.Space = &space.Space{Kind: kind, Extent: units.RectWH(0, 0, 4*m, 4*m)}
		}
		if !in.IsNil() {
			c.Location = &comp.Location{Space: in, Pos: pos, Facing: facing}
		}
		if _, err := h.w.Enqueue(c); err != nil {
			t.Fatal(err)
		}
		h.w.Step()
		for _, e := range h.w.Events() {
			if s, ok := e.Event.(cmds.EntitySpawned); ok {
				return s.Entity
			}
		}
		t.Fatalf("spawn %s rechazado: %+v", name, h.w.Events())
		return ids.Nil
	}
	h.root = spawn("casa", space.Static, ids.Nil, units.Vec3{}, 0)
	h.roomA = spawn("habitación A", space.Static, h.root, units.Vec3{}, 0)
	h.roomB = spawn("habitación B", space.Static, h.root, units.V3(5*m, 0, 0), fixedmath.QuarterTurn)
	h.desk = spawn("escritorio", space.Elevated, h.roomA, units.V3(1*m, 2*m, 75*cm), 0)
	h.gts = spawn("giantess", space.Body, h.roomA, units.V3(2*m, 1*m, 0), 0)
	h.palm = spawn("palma", space.Articulated, h.gts, units.V3(30*cm, 0, 1*m), 0)
	h.tiny = spawn("tiny", 255, h.palm, units.V3(2*cm, 0, 0), 0)
	h.g = space.Of(h.w.ECS())
	return h
}

func (h *house) loc(e ids.ID) comp.Location {
	l, _ := comp.Of(h.w.ECS()).Location.Get(e)
	return l
}

func (h *house) worldPos(t *testing.T, e ids.ID) units.Vec3 {
	t.Helper()
	p, _, err := h.g.WorldPose(e)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func (h *house) apply(t *testing.T, c world.Command) []world.EventRecord {
	t.Helper()
	if _, err := h.w.Enqueue(c); err != nil {
		t.Fatal(err)
	}
	h.w.Step()
	if err := h.w.CheckInvariants(); err != nil {
		t.Fatal(err)
	}
	return h.w.Events()
}

// Aceptación M2: un tiny sobre una mano en movimiento conserva su posición
// local y su posición en el mundo sigue a la giantess. Valores esperados
// calculados a mano (rotaciones de 90°: exactas).
func TestTinyRidesMovingHand(t *testing.T) {
	h := newHouse(t)
	before := h.loc(h.tiny)
	// Giantess en (2m,1m): palma en +30cm en x, tiny 2cm más.
	if got := h.worldPos(t, h.tiny); got != units.V3(2*m+32*cm, 1*m, 1*m) {
		t.Fatalf("posición inicial %v", got)
	}
	// La giantess camina a (3m,2m) y gira 90°: lo que estaba en +x queda en +y.
	h.apply(t, cmds.SetLocation{Entity: h.gts, Location: comp.Location{
		Space: h.roomA, Pos: units.V3(3*m, 2*m, 0), Facing: fixedmath.QuarterTurn,
	}})
	if h.loc(h.tiny) != before {
		t.Fatal("la posición local del tiny cambió")
	}
	if got := h.worldPos(t, h.tiny); got != units.V3(3*m, 2*m+32*cm, 1*m) {
		t.Fatalf("posición tras mover a la giantess %v", got)
	}
}

// Aceptación M2: la giantess cambia de habitación; sus partes siguen montadas
// en ella y nadie se mueve en el mundo por el cambio de contenedor.
func TestGiantessChangesRoom(t *testing.T) {
	h := newHouse(t)
	tinyBefore := h.worldPos(t, h.tiny)
	evs := h.apply(t, cmds.Reparent{Entity: h.gts, Space: h.roomB})
	if len(evs) != 1 || evs[0].Event.(space.EntityReparented).To != h.roomB {
		t.Fatalf("eventos %+v", evs)
	}
	if h.loc(h.gts).Space != h.roomB || h.loc(h.palm).Space != h.gts || h.loc(h.tiny).Space != h.palm {
		t.Fatal("la jerarquía del cuerpo se rompió")
	}
	if got := h.worldPos(t, h.tiny); got != tinyBefore {
		t.Fatalf("el reparent movió al tiny: %v -> %v", tinyBefore, got)
	}
	// Habitación B está rotada 90°: la posición local de la giantess cambia.
	if h.loc(h.gts).Pos == units.V3(2*m, 1*m, 0) {
		t.Fatal("la posición local debería expresarse en coordenadas de B")
	}
}

// Aceptación M2: Reparent conserva la posición en el mundo. Exacto con
// ángulos de 90°; con ángulos arbitrarios, error de pocos µm.
func TestReparentPreservesWorldPosition(t *testing.T) {
	h := newHouse(t)
	before := h.worldPos(t, h.tiny)
	h.apply(t, cmds.Reparent{Entity: h.tiny, Space: h.desk})
	if got := h.worldPos(t, h.tiny); got != before {
		t.Fatalf("90°: %v -> %v", before, got)
	}

	h.apply(t, cmds.SetLocation{Entity: h.gts, Location: comp.Location{
		Space: h.roomA, Pos: units.V3(2*m, 1*m, 0), Facing: fixedmath.Degrees(37),
	}})
	h.apply(t, cmds.Reparent{Entity: h.tiny, Space: h.palm})
	before = h.worldPos(t, h.tiny)
	h.apply(t, cmds.Reparent{Entity: h.tiny, Space: h.roomB})
	if d := h.worldPos(t, h.tiny).Sub(before); d.X.Abs() > 3 || d.Y.Abs() > 3 || d.Z != 0 {
		t.Fatalf("37°: error %v", d)
	}
}

func TestCyclesAndInvalidTargetsRejected(t *testing.T) {
	h := newHouse(t)
	for name, c := range map[string]world.Command{
		"giantess dentro de su palma":    cmds.Reparent{Entity: h.gts, Space: h.palm},
		"space dentro de sí mismo":       cmds.SetLocation{Entity: h.palm, Location: comp.Location{Space: h.palm}},
		"dentro de algo que no es space": cmds.SetLocation{Entity: h.desk, Location: comp.Location{Space: h.tiny}},
		"spawn en space inexistente":     cmds.Spawn{Location: &comp.Location{Space: ids.Make(99, 1)}},
	} {
		evs := h.apply(t, c)
		if len(evs) != 1 {
			t.Errorf("%s: eventos %+v", name, evs)
			continue
		}
		if _, ok := evs[0].Event.(world.CommandRejected); !ok {
			t.Errorf("%s: debería rechazarse", name)
		}
	}
}

// Si se destruye un Space, sus ocupantes pasan al ancestro vivo más cercano
// sin moverse en el mundo; si se destruye la raíz, todo cae en cascada.
func TestDestroyedSpaceOrphans(t *testing.T) {
	h := newHouse(t)
	before := h.worldPos(t, h.tiny)
	h.w.Enqueue(cmds.Despawn{Entity: h.palm})
	h.apply(t, cmds.Despawn{Entity: h.gts}) // palma y giantess a la vez
	if h.loc(h.tiny).Space != h.roomA || h.worldPos(t, h.tiny) != before {
		t.Fatalf("tiny en %v, mundo %v", h.loc(h.tiny).Space, h.worldPos(t, h.tiny))
	}

	h.apply(t, cmds.Despawn{Entity: h.root})
	if n := h.w.ECS().Len(); n != 0 {
		t.Fatalf("quedaron %d entidades tras destruir la raíz", n)
	}
}

func TestDifferentRoots(t *testing.T) {
	h := newHouse(t)
	h.apply(t, cmds.Spawn{Space: &space.Space{}})
	other := h.w.Events()[0].Event.(cmds.EntitySpawned).Entity
	if _, err := h.g.Reparent(h.tiny, other); !errors.Is(err, space.ErrDifferentRoots) {
		t.Fatalf("err = %v", err)
	}
}
