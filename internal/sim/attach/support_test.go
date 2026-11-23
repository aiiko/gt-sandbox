package attach_test

import (
	"encoding/json"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// TestSupportAccessRules verifica el criterio 19.6.2:
//   - Apoyos simultáneos: mano en escritorio, pie en piso, antebrazo en sofá.
//   - Roce breve (< StableTicks) no abre portal.
//   - Extremo bloqueado cierra el portal.
//   - Pie plantado (RequireStill) cierra al empezar a caminar.
func TestSupportAccessRules(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 20})

	floor := spawn(t, w, cmds.Spawn{
		Name: "piso",
		Space: &space.Space{
			Kind:   space.Static,
			Extent: units.RectWH(0, 0, 3*units.Meter, 3*units.Meter),
		},
		Components: map[string]json.RawMessage{
			"contact_surface": mustJSON(t, attach.Surface{Tolerance: 4 * units.Centimeter}),
		},
	})
	desk := spawn(t, w, cmds.Spawn{
		Name:     "escritorio",
		Space:    &space.Space{Kind: space.Elevated, Extent: units.RectWH(0, 0, 80*units.Centimeter, 60*units.Centimeter)},
		Location: &comp.Location{Space: floor, Pos: units.V3(0, 0, 75*units.Centimeter)},
		Components: map[string]json.RawMessage{
			"contact_surface": mustJSON(t, attach.Surface{Tolerance: 4 * units.Centimeter}),
			"nav_grid": mustJSON(t, map[string]any{
				"origin": units.Vec2{}, "cell": 2 * units.Centimeter, "version": 0,
				"rows": makeRows(40, 30),
			}),
		},
	})
	sofa := spawn(t, w, cmds.Spawn{
		Name:     "sofá",
		Space:    &space.Space{Kind: space.Elevated, Extent: units.RectWH(0, 0, 80*units.Centimeter, 60*units.Centimeter)},
		Location: &comp.Location{Space: floor, Pos: units.V3(units.Meter, 0, 45*units.Centimeter)},
		Components: map[string]json.RawMessage{
			"contact_surface": mustJSON(t, attach.Surface{Tolerance: 4 * units.Centimeter}),
		},
	})

	giantess := spawn(t, w, cmds.Spawn{
		Name:     "giantess",
		Space:    &space.Space{Kind: space.Body},
		Location: &comp.Location{Space: floor, Pos: units.V3(80*units.Centimeter, 20*units.Centimeter, 0)},
		Components: map[string]json.RawMessage{
			"mover": mustJSON(t, nav.Mover{Speed: units.Meter, Scale: "giantess"}),
		},
	})

	// Mano apoyada en el escritorio (con cara de apoyo + borde de acceso lateral).
	hand := spawn(t, w, cmds.Spawn{
		Name:     "mano",
		Space:    &space.Space{Kind: space.Articulated, Extent: units.RectWH(0, 0, 12*units.Centimeter, 12*units.Centimeter)},
		Location: &comp.Location{Space: giantess, Pos: units.V3(-30*units.Centimeter, 0, 77*units.Centimeter)},
		Components: map[string]json.RawMessage{
			"contact_surface": mustJSON(t, attach.Surface{
				Tolerance:     4 * units.Centimeter,
				Scale:         "tiny",
				StableTicks:   2,
				Support:       &units.Rect{Min: units.V2(2*units.Centimeter, 2*units.Centimeter), Max: units.V2(10*units.Centimeter, 10*units.Centimeter)},
				SupportOffset: -2 * units.Centimeter,
				Access:        []units.Vec2{{X: units.Centimeter, Y: 6 * units.Centimeter}},
				AccessReach:   6 * units.Centimeter,
				Body:          giantess,
			}),
		},
	})
	// Pie plantado en el piso (RequireStill: true, StableTicks: 2).
	foot := spawn(t, w, cmds.Spawn{
		Name:     "pie",
		Space:    &space.Space{Kind: space.Articulated, Extent: units.RectWH(0, 0, 10*units.Centimeter, 24*units.Centimeter)},
		Location: &comp.Location{Space: giantess, Pos: units.V3(0, 0, 2*units.Centimeter)},
		Components: map[string]json.RawMessage{
			"contact_surface": mustJSON(t, attach.Surface{
				Tolerance:     4 * units.Centimeter,
				Scale:         "tiny",
				StableTicks:   2,
				RequireStill:  true,
				SupportOffset: -2 * units.Centimeter,
				Access:        []units.Vec2{{X: units.Centimeter, Y: 12 * units.Centimeter}},
				AccessReach:   6 * units.Centimeter,
				Body:          giantess,
			}),
		},
	})
	// Antebrazo apoyado en el sofá.
	forearm := spawn(t, w, cmds.Spawn{
		Name:     "antebrazo",
		Space:    &space.Space{Kind: space.Articulated, Extent: units.RectWH(0, 0, 10*units.Centimeter, 20*units.Centimeter)},
		Location: &comp.Location{Space: giantess, Pos: units.V3(30*units.Centimeter, 0, 47*units.Centimeter)},
		Components: map[string]json.RawMessage{
			"contact_surface": mustJSON(t, attach.Surface{
				Tolerance:     4 * units.Centimeter,
				Scale:         "tiny",
				StableTicks:   2,
				SupportOffset: -2 * units.Centimeter,
				Access:        []units.Vec2{{X: 5 * units.Centimeter, Y: units.Centimeter}},
				AccessReach:   6 * units.Centimeter,
				Body:          giantess,
			}),
		},
	})
	// Mano derecha para probar roce breve (< StableTicks = 3).
	brushHand := spawn(t, w, cmds.Spawn{
		Name:     "mano roce",
		Space:    &space.Space{Kind: space.Articulated, Extent: units.RectWH(0, 0, 10*units.Centimeter, 10*units.Centimeter)},
		Location: &comp.Location{Space: giantess, Pos: units.V3(-50*units.Centimeter, 0, 120*units.Centimeter)},
		Components: map[string]json.RawMessage{
			"contact_surface": mustJSON(t, attach.Surface{
				Tolerance:     4 * units.Centimeter,
				Scale:         "tiny",
				StableTicks:   3,
				SupportOffset: -2 * units.Centimeter,
				Access:        []units.Vec2{{X: 5 * units.Centimeter, Y: 5 * units.Centimeter}},
				AccessReach:   6 * units.Centimeter,
				Body:          giantess,
			}),
		},
	})
	w.Enqueue(cmds.SetComponents{
		Entity: giantess,
		Components: map[string]json.RawMessage{
			"poses": mustJSON(t, attach.Poses{
				"brush_down": {{Part: brushHand, Pos: units.V3(-50*units.Centimeter, 0, 77*units.Centimeter)}},
				"brush_up":   {{Part: brushHand, Pos: units.V3(-50*units.Centimeter, 0, 120*units.Centimeter)}},
			}),
		},
	})
	w.StepN(3)

	// 1. Verificar los 3 apoyos simultáneos abiertos y con scale "tiny".
	portals := ecs.Get[nav.Portal](w.ECS())
	if portals.Len() != 3 {
		t.Fatalf("se esperaban 3 portales simultáneos (mano-escritorio, pie-piso, antebrazo-sofá), hay %d", portals.Len())
	}
	var handPortalID ids.ID
	var deskEnd units.Vec2
	portals.Each(func(id ids.ID, p *nav.Portal) {
		if p.Scale != "tiny" {
			t.Fatalf("portal %v debía tener scale=tiny, tiene %q", id, p.Scale)
		}
		if (p.A.Space == hand && p.B.Space == desk) || (p.A.Space == desk && p.B.Space == hand) {
			handPortalID = id
			if p.A.Space == desk {
				deskEnd = p.A.Pos
			} else {
				deskEnd = p.B.Pos
			}
		}
	})
	if handPortalID.IsNil() {
		t.Fatal("falta el portal mano-escritorio")
	}

	// 2. Roce breve (1 tick tocando < StableTicks=3) no debe abrir portal.
	w.Enqueue(attach.SetPose{Entity: giantess, Pose: "brush_down"})
	w.Step()
	openedBrush := false
	for _, ev := range w.Events() {
		if po, ok := ev.Event.(attach.PortalOpened); ok && (po.A == brushHand || po.B == brushHand) {
			openedBrush = true
		}
	}
	w.Enqueue(attach.SetPose{Entity: giantess, Pose: "brush_up"})
	w.Step()
	if openedBrush {
		t.Fatal("un roce breve abrió portal antes de StableTicks")
	}

	// 3. Bloquear el extremo en el escritorio debe cerrar o reubicar el portal; si bloqueamos toda la zona, se cierra.
	grid := ecs.Get[nav.Grid](w.ECS()).Ptr(desk)
	cx, cy := grid.CellOf(deskEnd)
	var cells []nav.CellCost
	for dy := -4; dy <= 4; dy++ {
		for dx := -4; dx <= 4; dx++ {
			if grid.In(cx+dx, cy+dy) {
				cells = append(cells, nav.CellCost{X: cx + dx, Y: cy + dy, Cost: nav.Blocked})
			}
		}
	}
	w.Enqueue(nav.SetCells{Space: desk, Cells: cells})
	w.Step()
	if portals.Has(handPortalID) {
		t.Fatal("el portal mano-escritorio debía cerrarse al bloquearse el extremo")
	}

	// 4. Al empezar a caminar la giantess, el portal del pie (RequireStill) se cierra.
	w.Enqueue(nav.MoveTo{Entity: giantess, Space: floor, Goal: units.V2(2*units.Meter, 20*units.Centimeter)})
	w.Step()
	footOpen := false
	portals.Each(func(_ ids.ID, p *nav.Portal) {
		if p.A.Space == foot || p.B.Space == foot {
			footOpen = true
		}
	})
	if footOpen {
		t.Fatal("el portal del pie debía cerrarse al caminar la giantess")
	}
	_ = forearm
	_ = sofa
}

// TestSupportWithdrawnAtCrossingTick verifica 19.3 y 19.6.2:
// Si la giantess retira la mano en el mismo tick en que un tiny llega al portal,
// quien todavía estaba en el escritorio queda en el escritorio y replanea;
// quien ya había cruzado permanece sobre la mano.
// También verifica que apoyar la mano protege al pasajero y aplasta al de abajo.
func TestSupportWithdrawnAtCrossingTickAndCrushProtection(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 21})

	floor := spawn(t, w, cmds.Spawn{Space: &space.Space{Kind: space.Static, Extent: units.RectWH(0, 0, 2*units.Meter, 2*units.Meter)}})
	desk := spawn(t, w, cmds.Spawn{
		Name:     "escritorio",
		Space:    &space.Space{Kind: space.Elevated, Extent: units.RectWH(0, 0, 60*units.Centimeter, 60*units.Centimeter)},
		Location: &comp.Location{Space: floor, Pos: units.V3(0, 0, 75*units.Centimeter)},
		Components: map[string]json.RawMessage{
			"contact_surface": mustJSON(t, attach.Surface{Tolerance: 3 * units.Centimeter}),
		},
	})
	giantess := spawn(t, w, cmds.Spawn{
		Name:     "giantess",
		Space:    &space.Space{Kind: space.Body},
		Location: &comp.Location{Space: floor, Pos: units.V3(30*units.Centimeter, 30*units.Centimeter, 0)},
	})
	palm := spawn(t, w, cmds.Spawn{
		Name:     "palma",
		Space:    &space.Space{Kind: space.Articulated, Extent: units.RectWH(0, 0, 12*units.Centimeter, 12*units.Centimeter)},
		Location: &comp.Location{Space: giantess, Pos: units.V3(0, 0, 120*units.Centimeter)},
		Components: map[string]json.RawMessage{
			"contact_surface": mustJSON(t, attach.Surface{
				Tolerance:     3 * units.Centimeter,
				Scale:         "tiny",
				SupportOffset: -2 * units.Centimeter,
				Access:        []units.Vec2{{X: units.Centimeter, Y: 6 * units.Centimeter}},
				AccessReach:   5 * units.Centimeter,
				Body:          giantess,
			}),
			"crusher": mustJSON(t, crush.Crusher{Source: "hand", Force: 100, Thickness: 2 * units.Centimeter}),
		},
	})
	w.Enqueue(cmds.SetComponents{
		Entity: giantess,
		Components: map[string]json.RawMessage{
			"poses": mustJSON(t, attach.Poses{
				"rest": {{Part: palm, Pos: units.V3(0, 0, 77*units.Centimeter)}},
				"lift": {{Part: palm, Pos: units.V3(0, 0, 120*units.Centimeter)}},
			}),
		},
	})
	w.Step()

	// Pasajero ya sobre la palma + víctima debajo en el escritorio.
	rider := spawn(t, w, cmds.Spawn{
		Name:       "Pasajero",
		Location:   &comp.Location{Space: palm, Pos: units.V3(6*units.Centimeter, 6*units.Centimeter, 0)},
		Body:       &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter},
		Components: map[string]json.RawMessage{"fragile": mustJSON(t, crush.Fragile{})},
	})
	under := spawn(t, w, cmds.Spawn{
		Name:       "Debajo",
		Location:   &comp.Location{Space: desk, Pos: units.V3(36*units.Centimeter, 36*units.Centimeter, 0)},
		Body:       &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter},
		Components: map[string]json.RawMessage{"fragile": mustJSON(t, crush.Fragile{})},
	})

	// Apoyar la mano: el pasajero sobre la palma sobrevive, el de abajo muere.
	w.Enqueue(attach.SetPose{Entity: giantess, Pose: "rest"})
	w.Step()
	if !w.ECS().Alive(rider) {
		t.Fatal("el pasajero sobre la palma no debía morir al apoyarla")
	}
	if w.ECS().Alive(under) {
		t.Fatal("el tiny debajo de la palma debía morir aplastado")
	}

	// El portal está en desk (31cm, 36cm). Ponemos al tiny a 1.5cm (1 cm/tick a 20cm/s):
	// en el tick 1 planifica y avanza 1cm (queda a 0.5cm, todavía en Moving);
	// en el tick 2 llegaría a cruzar el portal, pero retiramos la mano en ese mismo tick.
	late := spawn(t, w, cmds.Spawn{
		Name:     "Tardío",
		Location: &comp.Location{Space: desk, Pos: units.V3(295*units.Millimeter, 36*units.Centimeter, 0)},
		Body:     &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter},
		Components: map[string]json.RawMessage{
			"mover": mustJSON(t, nav.Mover{Speed: 20 * units.Centimeter, Scale: "tiny"}),
		},
	})
	w.Enqueue(nav.MoveTo{Entity: late, Space: palm, Goal: units.V2(6*units.Centimeter, 6*units.Centimeter)})
	w.Step() // planifica y se acerca a 0.5cm del portal (aún en Moving)
	if m, _ := ecs.Get[nav.Mover](w.ECS()).Get(late); m.Status != nav.Moving {
		t.Fatalf("late debía estar Moving antes del retiro, está en %v", m.Status)
	}
	// Retiramos la mano en el mismo tick en que nav.move llegaría a cruzar el portal.
	w.Enqueue(attach.SetPose{Entity: giantess, Pose: "lift"})
	w.Step()

	lLate, _ := comp.Of(w.ECS()).Location.Get(late)
	if lLate.Space != desk {
		t.Fatalf("el tiny que aún no había cruzado debía quedarse en el escritorio, terminó en %v", lLate.Space)
	}
	lRider, _ := comp.Of(w.ECS()).Location.Get(rider)
	if lRider.Space != palm {
		t.Fatalf("el pasajero que ya estaba en la palma debía viajar con ella, terminó en %v", lRider.Space)
	}
	if err := w.CheckInvariants(); err != nil {
		t.Fatal(err)
	}
}

func makeRows(w, h int) []string {
	row := make([]byte, w)
	for i := range row {
		row[i] = '.'
	}
	s := string(row)
	rows := make([]string, h)
	for i := range rows {
		rows[i] = s
	}
	return rows
}
