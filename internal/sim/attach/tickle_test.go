package attach_test

import (
	"encoding/json"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

type tickleEntities struct {
	room, giantess, neck, tiny ids.ID
}

func setupTickleWorld(t *testing.T, gripPos *units.Vec2, gripped bool) (*world.World, tickleEntities) {
	t.Helper()
	w := engine.MustNew(world.Config{Seed: 41})
	room := spawn(t, w, cmds.Spawn{
		Space: &space.Space{Kind: space.Static, Extent: units.RectWH(0, 0, 2*units.Meter, 2*units.Meter)},
		Components: map[string]json.RawMessage{
			"nav_grid": mustJSON(t, flatGrid(t, 20, 20, 10*units.Centimeter)),
		},
	})
	giantess := spawn(t, w, cmds.Spawn{
		Space:    &space.Space{Kind: space.Body},
		Location: &comp.Location{Space: room, Pos: units.V3(units.Meter, units.Meter, 0)},
		Components: map[string]json.RawMessage{
			"mover":      mustJSON(t, nav.Mover{Speed: units.Meter, Scale: "giantess"}),
			"body_state": mustJSON(t, attach.BodyState{Pose: "idle"}),
		},
	})
	neck := spawn(t, w, cmds.Spawn{
		Name: "cuello y nuca",
		Space: &space.Space{
			Kind:   space.Articulated,
			Extent: units.RectWH(0, 0, 20*units.Centimeter, 20*units.Centimeter),
		},
		Location: &comp.Location{Space: giantess},
		Components: map[string]json.RawMessage{
			"nav_grid": mustJSON(t, flatGrid(t, 20, 20, units.Centimeter)),
			"body_zone": mustJSON(t, attach.BodyZone{
				Body:        giantess,
				Region:      "cuello y nuca",
				Sensitivity: 50,
				RequireGrip: true,
				Grips: []attach.GripPoint{
					{Pos: units.V2(18*units.Centimeter, 18*units.Centimeter), Radius: 3 * units.Centimeter, Label: "borde del cuello"},
				},
			}),
		},
	})
	startPos := units.V3(2*units.Centimeter, 2*units.Centimeter, 0)
	if gripPos != nil {
		startPos = units.V3(gripPos.X, gripPos.Y, 0)
	}
	tinyComps := map[string]json.RawMessage{
		"mover": mustJSON(t, nav.Mover{Speed: 4 * units.Centimeter, Scale: "tiny"}),
	}
	if gripped {
		tinyComps["gripping"] = mustJSON(t, attach.Gripping{
			Space: neck,
			Pos:   startPos.XY(),
			Label: "borde del cuello",
			Until: 500,
		})
	}
	tiny := spawn(t, w, cmds.Spawn{
		Location:   &comp.Location{Space: neck, Pos: startPos},
		Body:       &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter, Mass: 1},
		Components: tinyComps,
	})
	poses := attach.Poses{
		"idle": {},
		"step": {},
	}
	ecs.Get[attach.Poses](w.ECS()).Set(giantess, poses)
	return w, tickleEntities{room: room, giantess: giantess, neck: neck, tiny: tiny}
}

func flatGrid(t *testing.T, w, h int, cell units.Length) *nav.Grid {
	t.Helper()
	g, err := nav.NewGrid(units.Vec2{}, cell, w, h)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestTickleAccumulatesTriggersAndCoolsDown(t *testing.T) {
	w, ent := setupTickleWorld(t, nil, false)
	w.Enqueue(nav.MoveTo{Entity: ent.giantess, Space: ent.room, Goal: units.V2(180*units.Centimeter, units.Meter)})
	w.Enqueue(nav.MoveTo{Entity: ent.tiny, Space: ent.neck, Goal: units.V2(15*units.Centimeter, 15*units.Centimeter)})

	var tickled []attach.BodyTickled
	var dropped []attach.EntityLostGrip
	for range 30 {
		w.Step()
		for _, ev := range w.Events() {
			if bt, ok := ev.Event.(attach.BodyTickled); ok {
				tickled = append(tickled, bt)
			}
			if lg, ok := ev.Event.(attach.EntityLostGrip); ok {
				dropped = append(dropped, lg)
			}
		}
	}
	if len(tickled) != 1 {
		t.Fatalf("esperaba exactamente 1 evento body_tickled (por cooldown), obtuve %d", len(tickled))
	}
	if tickled[0].Body != ent.giantess || tickled[0].Zone != ent.neck || tickled[0].By != ent.tiny || tickled[0].Region != "cuello y nuca" {
		t.Fatalf("evento body_tickled inesperado: %+v", tickled[0])
	}
	if len(dropped) == 0 {
		t.Fatal("el tiny sin agarre debería perder apoyo al sacudirse la zona")
	}
	st, ok := ecs.Get[attach.BodyState](w.ECS()).Get(ent.giantess)
	if !ok || st.TickleCooldown <= w.Tick() || st.TickleUntil == 0 {
		t.Fatalf("estado de cosquillas inválido tras disparar: %+v", st)
	}
	mv, _ := ecs.Get[nav.Mover](w.ECS()).Get(ent.giantess)
	if mv.WaitUntil < st.TickleUntil {
		t.Fatalf("la giantess debía frenar hasta %d, WaitUntil=%d", st.TickleUntil, mv.WaitUntil)
	}
	// Cambiar de pose debe preservar los campos de cosquillas.
	w.Enqueue(attach.SetPose{Entity: ent.giantess, Pose: "step"})
	w.Step()
	st2, _ := ecs.Get[attach.BodyState](w.ECS()).Get(ent.giantess)
	if st2.Pose != "step" || st2.TickleCooldown != st.TickleCooldown || st2.TickleUntil != st.TickleUntil {
		t.Fatalf("SetPose borró el estado de cosquillas: antes=%+v despues=%+v", st, st2)
	}
}

func TestGrippedTinyDoesNotTriggerTickle(t *testing.T) {
	pos := units.V2(18*units.Centimeter, 18*units.Centimeter)
	w, ent := setupTickleWorld(t, &pos, true)
	w.Enqueue(nav.MoveTo{Entity: ent.tiny, Space: ent.neck, Goal: units.V2(2*units.Centimeter, 2*units.Centimeter)})
	for range 25 {
		w.Step()
		for _, ev := range w.Events() {
			if _, ok := ev.Event.(attach.BodyTickled); ok {
				t.Fatal("un tiny sujeto al borde no debía disparar cosquillas")
			}
		}
	}
	st, _ := ecs.Get[attach.BodyState](w.ECS()).Get(ent.giantess)
	if st.Tickle != 0 {
		t.Fatalf("no debía acumular cosquillas mientras está agarrado, Tickle=%d", st.Tickle)
	}
}
