package view_test

import (
	"encoding/json"
	"strings"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
	"gtsandbox/internal/view"
)

func TestBodyFrameFlagsAndSupportChronicle(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 1})
	spawn := func(c cmds.Spawn) ids.ID {
		w.Enqueue(c)
		w.Step()
		for _, ev := range w.Events() {
			if sp, ok := ev.Event.(cmds.EntitySpawned); ok {
				return sp.Entity
			}
		}
		t.Fatal("spawn falló")
		return ids.Nil
	}
	mustJSON := func(v any) json.RawMessage {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	room := spawn(cmds.Spawn{Name: "dormitorio", Space: &space.Space{Kind: space.Static, Extent: units.RectWH(0, 0, 2*units.Meter, 2*units.Meter)}})
	pillow := spawn(cmds.Spawn{Name: "almohada", Space: &space.Space{Kind: space.Elevated, Extent: units.RectWH(0, 0, 50*units.Centimeter, 30*units.Centimeter)}, Location: &comp.Location{Space: room}})
	giantess := spawn(cmds.Spawn{Name: "giantess", Space: &space.Space{Kind: space.Body}, Location: &comp.Location{Space: room}})
	hair := spawn(cmds.Spawn{
		Name: "cabello y rostro",
		Space: &space.Space{
			Kind:   space.Articulated,
			Extent: units.RectWH(0, 0, 20*units.Centimeter, 20*units.Centimeter),
			Projection: &space.Projection{
				AnchorA: units.V3(0, 0, 140*units.Centimeter),
				AnchorB: units.V3(0, 0, 160*units.Centimeter),
				Climb:   true,
			},
		},
		Location: &comp.Location{Space: giantess},
		Components: map[string]json.RawMessage{
			"body_zone": mustJSON(attach.BodyZone{Body: giantess, Region: "cabello y rostro", Occluded: true}),
		},
	})
	tiny := spawn(cmds.Spawn{
		Name:     "Tiny 01",
		Location: &comp.Location{Space: hair, Pos: units.V3(10*units.Centimeter, 10*units.Centimeter, 0)},
		Body:     &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter},
		Components: map[string]json.RawMessage{
			"gripping": mustJSON(attach.Gripping{Space: hair, Label: "mechón izquierdo", Until: 100}),
		},
	})

	var cap view.Capturer
	f := cap.Capture(w)
	sp, ok := f.SpaceByID(hair)
	if !ok || !sp.Climb || !sp.Occluded {
		t.Fatalf("Space de cabello debía marcar Climb=true y Occluded=true: %+v", sp)
	}
	th, ok := f.ThingByID(tiny)
	if !ok || th.GripLabel != "mechón izquierdo" {
		t.Fatalf("Thing debía exponer GripLabel=%q, tiene %q", "mechón izquierdo", th.GripLabel)
	}

	var ch view.Chronicle
	ch.Record(w.ECS(), 10, colony.SupportDiscovered{
		Entity: tiny, BodySpace: hair, HouseSpace: pillow, Pose: "lie_down",
	})
	lines := ch.Lines()
	if len(lines) != 1 || !strings.Contains(lines[0].Text, "cabello y rostro") || !strings.Contains(lines[0].Text, "almohada") || !strings.Contains(lines[0].Text, "cuando se recuesta") {
		t.Fatalf("crónica inesperada para SupportDiscovered: %+v", lines)
	}
}
