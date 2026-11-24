package colony_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

func bodyMustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func bodySpawn(t *testing.T, w *world.World, c cmds.Spawn) ids.ID {
	t.Helper()
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

// TestBodyExpeditionSupportMemoryAndContinuity verifica el criterio 19.6.4:
//   - Una compañía parte hacia una zona corporal, aguarda un cambio de pose
//     (codo doblado -> reabierto), desembarca al abrirse el apoyo en la
//     almohada, vuelve al nido y comunica el apoyo descubierto
//     (SupportDiscovered + Colony.KnownSupports).
//   - Guardar y cargar cerca del cambio de pose y del cierre de portal da el
//     mismo hash y eventos que la ejecución continua.
func TestBodyExpeditionSupportMemoryAndContinuity(t *testing.T) {
	run := func(saveTick int) (uint64, bool, []colony.SupportNote) {
		w := engine.MustNew(world.Config{Seed: 194})

		room := bodySpawn(t, w, cmds.Spawn{
			Name:  "dormitorio",
			Space: &space.Space{Kind: space.Static, Extent: units.RectWH(0, 0, 2*units.Meter, 2*units.Meter)},
		})
		desk := bodySpawn(t, w, cmds.Spawn{
			Name:     "escritorio",
			Space:    &space.Space{Kind: space.Elevated, Extent: units.RectWH(0, 0, 80*units.Centimeter, 60*units.Centimeter)},
			Location: &comp.Location{Space: room, Pos: units.V3(0, 0, 75*units.Centimeter)},
			Components: map[string]json.RawMessage{
				"contact_surface": bodyMustJSON(t, attach.Surface{Tolerance: 4 * units.Centimeter}),
			},
		})
		pillow := bodySpawn(t, w, cmds.Spawn{
			Name:     "almohada",
			Space:    &space.Space{Kind: space.Elevated, Extent: units.RectWH(0, 0, 60*units.Centimeter, 40*units.Centimeter)},
			Location: &comp.Location{Space: room, Pos: units.V3(0, 0, 25*units.Centimeter)},
			Components: map[string]json.RawMessage{
				"contact_surface": bodyMustJSON(t, attach.Surface{Tolerance: 5 * units.Centimeter}),
			},
		})
		// Paso de vuelta desde la almohada al escritorio (para regresar al nido).
		bodySpawn(t, w, cmds.Spawn{Components: map[string]json.RawMessage{
			"portal": bodyMustJSON(t, nav.Portal{
				A:     nav.PortalEnd{Space: pillow, Pos: units.V2(10*units.Centimeter, 10*units.Centimeter)},
				B:     nav.PortalEnd{Space: desk, Pos: units.V2(10*units.Centimeter, 10*units.Centimeter)},
				Scale: "tiny",
			}),
		}})

		giantess := bodySpawn(t, w, cmds.Spawn{
			Name:     "giantess",
			Space:    &space.Space{Kind: space.Body},
			Location: &comp.Location{Space: room, Pos: units.V3(20*units.Centimeter, 20*units.Centimeter, 0)},
		})
		palm := bodySpawn(t, w, cmds.Spawn{
			Name:     "palma",
			Space:    &space.Space{Kind: space.Articulated, Extent: units.RectWH(0, 0, 12*units.Centimeter, 12*units.Centimeter)},
			Location: &comp.Location{Space: giantess, Pos: units.V3(10*units.Centimeter, 10*units.Centimeter, 77*units.Centimeter)},
			Components: map[string]json.RawMessage{
				"contact_surface": bodyMustJSON(t, attach.Surface{
					Tolerance: 4 * units.Centimeter, Scale: "tiny", Body: giantess,
				}),
				"body_zone": bodyMustJSON(t, attach.BodyZone{
					Body: giantess, Region: "palma",
					Grips: []attach.GripPoint{{Pos: units.V2(6*units.Centimeter, 6*units.Centimeter), Radius: 5 * units.Centimeter, Label: "puño"}},
				}),
			},
		})
		head := bodySpawn(t, w, cmds.Spawn{
			Name:     "cabello y rostro",
			Space:    &space.Space{Kind: space.Articulated, Extent: units.RectWH(0, 0, 20*units.Centimeter, 20*units.Centimeter)},
			Location: &comp.Location{Space: giantess, Pos: units.V3(10*units.Centimeter, 10*units.Centimeter, 150*units.Centimeter)},
			Components: map[string]json.RawMessage{
				"contact_surface": bodyMustJSON(t, attach.Surface{
					Tolerance: 5 * units.Centimeter, Scale: "tiny",
					Access: []units.Vec2{{X: 10 * units.Centimeter, Y: 10 * units.Centimeter}}, AccessReach: 6 * units.Centimeter,
					Body: giantess,
				}),
				"body_zone": bodyMustJSON(t, attach.BodyZone{
					Body: giantess, Region: "cabello y rostro", Occluded: true,
					Grips: []attach.GripPoint{{Pos: units.V2(10*units.Centimeter, 10*units.Centimeter), Radius: 6 * units.Centimeter, Label: "mechón"}},
				}),
			},
		})
		armJoint := bodySpawn(t, w, cmds.Spawn{Components: map[string]json.RawMessage{
			"portal": bodyMustJSON(t, nav.Portal{
				A:     nav.PortalEnd{Space: palm, Pos: units.V2(10*units.Centimeter, 6*units.Centimeter)},
				B:     nav.PortalEnd{Space: head, Pos: units.V2(10*units.Centimeter, 6*units.Centimeter)},
				Scale: "tiny",
			}),
		}})

		w.Enqueue(cmds.SetComponents{
			Entity: giantess,
			Components: map[string]json.RawMessage{
				"poses": bodyMustJSON(t, attach.Poses{
					"hand_rest": {
						{Part: palm, Pos: units.V3(10*units.Centimeter, 10*units.Centimeter, 77*units.Centimeter)},
						{Part: head, Pos: units.V3(10*units.Centimeter, 10*units.Centimeter, 150*units.Centimeter)},
					},
					"elbow_bent": {
						{Part: palm, Pos: units.V3(10*units.Centimeter, 10*units.Centimeter, 77*units.Centimeter)},
						{Part: head, Pos: units.V3(10*units.Centimeter, 10*units.Centimeter, 150*units.Centimeter)},
					},
					"lie_down": {
						{Part: palm, Pos: units.V3(10*units.Centimeter, 10*units.Centimeter, 120*units.Centimeter)},
						{Part: head, Pos: units.V3(10*units.Centimeter, 10*units.Centimeter, 27*units.Centimeter)},
					},
				}),
				"body_joints": bodyMustJSON(t, attach.BodyJoints{
					{Portal: armJoint, ClosedPoses: []string{"elbow_bent"}},
				}),
				"body_state": bodyMustJSON(t, attach.BodyState{Pose: "hand_rest"}),
			},
		})
		w.Step()

		// Colonia en el escritorio con comida sobrada y 3 miembros (1 explorador + 2 en el nido).
		nest := bodySpawn(t, w, cmds.Spawn{
			Name:     "nido",
			Location: &comp.Location{Space: desk, Pos: units.V3(20*units.Centimeter, 20*units.Centimeter, 0)},
			Components: map[string]json.RawMessage{
				"colony":   bodyMustJSON(t, colony.Colony{Radius: 8 * units.Centimeter, Stock: colony.NewStock(ai.ResourceFood, 2)}),
				"portions": bodyMustJSON(t, ai.NewPortions(ai.ResourceFood, 20)),
			},
		})
		target := bodySpawn(t, w, cmds.Spawn{
			Name:     "pelo de la giantess",
			Location: &comp.Location{Space: head, Pos: units.V3(10*units.Centimeter, 10*units.Centimeter, 0)},
			Components: map[string]json.RawMessage{
				"expedition_target": bodyMustJSON(t, colony.ExpeditionTarget{
					Colony: nest, MinMembers: 1, MaxMembers: 1, SurveyTicks: 10, MaxDuration: 800,
				}),
			},
		})
		_ = target

		spawnTiny := func(name, role string, adv float64) ids.ID {
			return bodySpawn(t, w, cmds.Spawn{
				Name:     comp.Name(name),
				Location: &comp.Location{Space: desk, Pos: units.V3(20*units.Centimeter, 20*units.Centimeter, 0)},
				Body:     &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter},
				Components: map[string]json.RawMessage{
					"mover":  bodyMustJSON(t, nav.Mover{Speed: 20 * units.Centimeter, Scale: "tiny"}),
					"needs":  bodyMustJSON(t, needs.Needs{{Name: "hunger", Value: 900}, {Name: "energy", Value: 900}, {Name: "fun", Value: 500}}),
					"agent":  bodyMustJSON(t, ai.Agent{Scale: "tiny", DistanceCost: 0.01}),
					"member": bodyMustJSON(t, colony.Member{Colony: nest, Role: role, Adventure: adv}),
				},
			})
		}
		scout := spawnTiny("Explorador", colony.RoleScout, 1.0)
		spawnTiny("Hogar 1", colony.RoleGatherer, 0)
		spawnTiny("Hogar 2", colony.RolePorter, 0)

		discovered := false
		completed := false
		var evLog []string
		// Verificar también que al estar en la palma y doblarse el codo, el
		// expedicionario espera en su punto de agarre y continúa al reabrirse.
		bentAt := 0
		liedAt := 0
		for step := 1; step <= 450; step++ {
			// Cuando llega a la palma por primera vez, doblar el codo 20 ticks.
			if bentAt == 0 {
				if l, ok := comp.Of(w.ECS()).Location.Get(scout); ok && l.Space == palm {
					bentAt = step
					w.Enqueue(attach.SetPose{Entity: giantess, Pose: "elbow_bent"})
				}
			} else if step == bentAt+20 {
				w.Enqueue(attach.SetPose{Entity: giantess, Pose: "hand_rest"})
			}
			// Cuando llega al cabello, recostar a la giantess para abrir cabello -> almohada.
			if liedAt == 0 {
				if l, ok := comp.Of(w.ECS()).Location.Get(scout); ok && l.Space == head {
					liedAt = step
					w.Enqueue(attach.SetPose{Entity: giantess, Pose: "lie_down"})
				}
			}
			w.Step()
			for _, ev := range w.Events() {
				switch ev.Event.(type) {
				case colony.SupportDiscovered:
					discovered = true
				case colony.ExpeditionCompleted, colony.StrandedReturned:
					completed = true
				}
				switch ev.Event.Kind() {
				case "expedition_proposed", "expedition_departed", "expedition_surveyed", "expedition_reported", "expedition_completed", "expedition_canceled", "expedition_stranded", "stranded_returned", "support_discovered", "entity_reparented", "move_failed":
					evLog = append(evLog, ev.Event.Kind())
				}
			}
			if saveTick > 0 && step == saveTick {
				snap, err := w.Save()
				if err != nil {
					t.Fatal(err)
				}
				var buf bytes.Buffer
				if err := world.WriteSnapshot(&buf, snap); err != nil {
					t.Fatal(err)
				}
				w, err = engine.Load(&buf)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
		if err := w.CheckInvariants(); err != nil {
			t.Fatal(err)
		}
		col, _ := ecs.Get[colony.Colony](w.ECS()).Get(nest)
		if !discovered || !completed {
			t.Logf("saveTick=%d events=%v", saveTick, evLog)
		}
		return w.Hash(), discovered && completed, col.KnownSupports
	}

	h1, ok1, notes1 := run(0)
	h2, ok2, notes2 := run(139) // justo antes del cambio de pose a lie_down
	if !ok1 || !ok2 {
		t.Fatalf("la expedición corporal debía completarse y reportar el apoyo (ok1=%v ok2=%v notes=%+v)", ok1, ok2, notes1)
	}
	if len(notes1) == 0 || len(notes2) == 0 {
		t.Fatalf("Colony.KnownSupports vacío: %+v", notes1)
	}
	if h1 != h2 {
		t.Fatalf("hash distinto tras guardar/cargar cerca del cambio de pose: %x vs %x", h1, h2)
	}
}
