package attach_test

import (
	"bytes"
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

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestHandToShoulderClimbAndElbowJoint verifica el criterio 19.6.1:
//   - Ruta mano (palma) → antebrazo → brazo → hombro con longitud real 3D.
//   - Mover el brazo conserva la posición local del tiny mientras trepa.
//   - Doblar el codo cierra su paso sin recrear el portal; al volver a
//     extenderlo el mismo tiny continúa y llega al hombro.
//   - Guardar y cargar a mitad de la trepada da el mismo hash y eventos.
func TestHandToShoulderClimbAndElbowJoint(t *testing.T) {
	run := func(saveAt int) (uint64, bool) {
		w := engine.MustNew(world.Config{Seed: 19})
		room := spawn(t, w, cmds.Spawn{Space: &space.Space{Kind: space.Static, Extent: units.RectWH(0, 0, 2*units.Meter, 2*units.Meter)}})
		giantess := spawn(t, w, cmds.Spawn{
			Space:    &space.Space{Kind: space.Body},
			Location: &comp.Location{Space: room, Pos: units.V3(units.Meter, units.Meter, 0)},
		})
		palm := spawn(t, w, cmds.Spawn{
			Name:     "palma",
			Space:    &space.Space{Kind: space.Articulated, Extent: units.RectWH(0, 0, 10*units.Centimeter, 10*units.Centimeter)},
			Location: &comp.Location{Space: giantess, Pos: units.V3(0, 0, 75*units.Centimeter)},
		})
		// Antebrazo: extensión 4×10cm proyectada a 30cm verticales (longitud 3D real = 30cm).
		forearmProj := &space.Projection{
			AnchorA: units.V3(5*units.Centimeter, 10*units.Centimeter, 75*units.Centimeter),
			AnchorB: units.V3(5*units.Centimeter, 10*units.Centimeter, 105*units.Centimeter),
			Climb:   true,
		}
		forearm := spawn(t, w, cmds.Spawn{
			Name: "antebrazo izquierdo",
			Space: &space.Space{
				Kind:       space.Articulated,
				Extent:     units.RectWH(0, 0, 4*units.Centimeter, 10*units.Centimeter),
				Projection: forearmProj,
			},
			Location: &comp.Location{Space: giantess},
			Components: map[string]json.RawMessage{
				"body_zone": mustJSON(t, attach.BodyZone{
					Body:        giantess,
					Region:      "antebrazo izquierdo",
					RequireGrip: true,
					Grips:       []attach.GripPoint{{Pos: units.V2(2*units.Centimeter, 2*units.Centimeter), Radius: 5 * units.Centimeter, Label: "puño de la manga"}},
				}),
			},
		})
		// Brazo: extensión 4×10cm proyectada a 30cm verticales.
		upperProj := &space.Projection{
			AnchorA: units.V3(5*units.Centimeter, 10*units.Centimeter, 105*units.Centimeter),
			AnchorB: units.V3(5*units.Centimeter, 10*units.Centimeter, 135*units.Centimeter),
			Climb:   true,
		}
		upperArm := spawn(t, w, cmds.Spawn{
			Name: "brazo izquierdo",
			Space: &space.Space{
				Kind:       space.Articulated,
				Extent:     units.RectWH(0, 0, 4*units.Centimeter, 10*units.Centimeter),
				Projection: upperProj,
			},
			Location: &comp.Location{Space: giantess},
		})
		shoulder := spawn(t, w, cmds.Spawn{
			Name:     "hombro izquierdo",
			Space:    &space.Space{Kind: space.Articulated, Extent: units.RectWH(0, 0, 12*units.Centimeter, 12*units.Centimeter)},
			Location: &comp.Location{Space: giantess, Pos: units.V3(0, 5*units.Centimeter, 135*units.Centimeter)},
		})

		// Portales internos: muñeca, codo, hombro.
		spawn(t, w, cmds.Spawn{Components: map[string]json.RawMessage{
			"portal": mustJSON(t, nav.Portal{
				A: nav.PortalEnd{Space: palm, Pos: units.V2(5*units.Centimeter, 9*units.Centimeter)},
				B: nav.PortalEnd{Space: forearm, Pos: units.V2(2*units.Centimeter, 1*units.Centimeter)},
			}),
		}})
		elbow := spawn(t, w, cmds.Spawn{Components: map[string]json.RawMessage{
			"portal": mustJSON(t, nav.Portal{
				A: nav.PortalEnd{Space: forearm, Pos: units.V2(2*units.Centimeter, 9*units.Centimeter)},
				B: nav.PortalEnd{Space: upperArm, Pos: units.V2(2*units.Centimeter, 1*units.Centimeter)},
			}),
		}})
		spawn(t, w, cmds.Spawn{Components: map[string]json.RawMessage{
			"portal": mustJSON(t, nav.Portal{
				A: nav.PortalEnd{Space: upperArm, Pos: units.V2(2*units.Centimeter, 9*units.Centimeter)},
				B: nav.PortalEnd{Space: shoulder, Pos: units.V2(6*units.Centimeter, 6*units.Centimeter)},
			}),
		}})

		// Poses y articulaciones en la giantess.
		w.Enqueue(cmds.SetComponents{
			Entity: giantess,
			Components: map[string]json.RawMessage{
				"poses": mustJSON(t, attach.Poses{
					"extended": {
						{Part: forearm, Pos: units.Vec3{}, Projection: forearmProj},
						{Part: upperArm, Pos: units.Vec3{}, Projection: upperProj},
					},
					"elbow_bent": {
						{Part: forearm, Pos: units.V3(5*units.Centimeter, 0, 10*units.Centimeter), Projection: forearmProj},
						{Part: upperArm, Pos: units.Vec3{}, Projection: upperProj},
					},
				}),
				"body_joints": mustJSON(t, attach.BodyJoints{
					{Portal: elbow, ClosedPoses: []string{"elbow_bent"}},
				}),
			},
		})
		w.Step()

		// Tiny en la palma, velocidad 20 cm/s (1 cm/tick).
		tiny := spawn(t, w, cmds.Spawn{
			Name:     "Tiny 01",
			Location: &comp.Location{Space: palm, Pos: units.V3(5*units.Centimeter, 5*units.Centimeter, 0)},
			Body:     &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter},
			Components: map[string]json.RawMessage{
				"mover": mustJSON(t, nav.Mover{Speed: 20 * units.Centimeter}),
			},
		})

		// Pedir ir al hombro.
		w.Enqueue(nav.MoveTo{Entity: tiny, Space: shoulder, Goal: units.V2(6*units.Centimeter, 6*units.Centimeter)})
		w.StepN(15) // ya está trepando el antebrazo (4cm palma + ~10cm reales de 30cm)

		locBefore, _ := comp.Of(w.ECS()).Location.Get(tiny)
		if locBefore.Space != forearm {
			t.Fatalf("en tick 15 debía estar en forearm, está en %v", locBefore.Space)
		}

		// Doblar el codo: mueve el antebrazo y cierra el portal del codo (sin recrearlo).
		w.Enqueue(attach.SetPose{Entity: giantess, Pose: "elbow_bent"})
		w.Step()
		locAfter, _ := comp.Of(w.ECS()).Location.Get(tiny)
		if locAfter.Space != forearm || locAfter.Pos != locBefore.Pos {
			t.Fatalf("al mover el brazo cambió la posición local: %+v → %+v", locBefore, locAfter)
		}
		if p := ecs.Get[nav.Portal](w.ECS()).Ptr(elbow); p == nil || !p.Disabled {
			t.Fatalf("el portal del codo debía cerrarse sin borrarse: %+v", p)
		}

		// Guardar y cargar a mitad de la trepada si se pidió.
		if saveAt > 0 {
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

		w.StepN(10)
		// Extender el codo de nuevo: se reabre el mismo portal y el tiny llega al hombro.
		w.Enqueue(attach.SetPose{Entity: giantess, Pose: "extended"})
		w.Enqueue(nav.MoveTo{Entity: tiny, Space: shoulder, Goal: units.V2(6*units.Centimeter, 6*units.Centimeter)})
		arrived := false
		for range 120 {
			w.Step()
			for _, ev := range w.Events() {
				if ma, ok := ev.Event.(nav.MoveArrived); ok && ma.Entity == tiny {
					arrived = true
				}
			}
		}
		if err := w.CheckInvariants(); err != nil {
			t.Fatal(err)
		}
		if l, _ := comp.Of(w.ECS()).Location.Get(tiny); l.Space != shoulder {
			t.Fatalf("no llegó al hombro: está en %v", l.Space)
		}
		return w.Hash(), arrived
	}

	h1, ok1 := run(0)
	h2, ok2 := run(1)
	if !ok1 || !ok2 {
		t.Fatal("el tiny no completó la llegada al hombro")
	}
	if h1 != h2 {
		t.Fatalf("hash distinto tras guardar/cargar a mitad de trepada: %x vs %x", h1, h2)
	}
}

func TestGripHoldsOrFallsOnPoseChange(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 2})
	floor := spawn(t, w, cmds.Spawn{Space: &space.Space{Kind: space.Static, Extent: units.RectWH(0, 0, 2*units.Meter, 2*units.Meter)}})
	giantess := spawn(t, w, cmds.Spawn{
		Space:    &space.Space{Kind: space.Body},
		Location: &comp.Location{Space: floor, Pos: units.V3(units.Meter, units.Meter, 0)},
	})
	sleeve := spawn(t, w, cmds.Spawn{
		Space:    &space.Space{Kind: space.Articulated, Extent: units.RectWH(0, 0, 20*units.Centimeter, 20*units.Centimeter)},
		Location: &comp.Location{Space: giantess, Pos: units.V3(0, 0, 80*units.Centimeter)},
		Components: map[string]json.RawMessage{
			"body_zone": mustJSON(t, attach.BodyZone{
				Body:        giantess,
				Region:      "manga",
				RequireGrip: true,
				Grips:       []attach.GripPoint{{Pos: units.V2(2*units.Centimeter, 2*units.Centimeter), Radius: 4 * units.Centimeter, Label: "puño"}},
			}),
		},
	})
	w.Enqueue(cmds.SetComponents{
		Entity: giantess,
		Components: map[string]json.RawMessage{
			"poses": mustJSON(t, attach.Poses{
				"jolt": {{Part: sleeve, Pos: units.V3(5*units.Centimeter, 0, 90*units.Centimeter)}},
			}),
		},
	})
	w.Step()

	// Tiny A junto al agarre (se sujeta); Tiny B lejos del agarre (cae al piso).
	tinyA := spawn(t, w, cmds.Spawn{
		Location: &comp.Location{Space: sleeve, Pos: units.V3(3*units.Centimeter, 2*units.Centimeter, 0)},
		Body:     &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter},
	})
	tinyB := spawn(t, w, cmds.Spawn{
		Location: &comp.Location{Space: sleeve, Pos: units.V3(18*units.Centimeter, 18*units.Centimeter, 0)},
		Body:     &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter},
	})

	w.Enqueue(attach.SetPose{Entity: giantess, Pose: "jolt"})
	w.Step()

	var gripped, lost, fell ids.ID
	for _, ev := range w.Events() {
		switch x := ev.Event.(type) {
		case attach.EntityGripped:
			gripped = x.Entity
		case attach.EntityLostGrip:
			lost = x.Entity
		case space.EntityFell:
			fell = x.Entity
		}
	}
	if gripped != tinyA {
		t.Fatalf("tinyA debía agarrarse, gripped=%v", gripped)
	}
	if lost != tinyB || fell != tinyB {
		t.Fatalf("tinyB debía perder agarre y caer, lost=%v fell=%v", lost, fell)
	}
	if l, _ := comp.Of(w.ECS()).Location.Get(tinyA); l.Space != sleeve {
		t.Fatalf("tinyA debía seguir en la manga, está en %v", l.Space)
	}
	if l, _ := comp.Of(w.ECS()).Location.Get(tinyB); l.Space != floor {
		t.Fatalf("tinyB debía aterrizar en el piso, está en %v", l.Space)
	}
}
