package main

import (
	"testing"

	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/render"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/space"
)

// Verifica que todas las zonas corporales, el futón y la almohada en casa.json
// tienen visual micro propio en el atlas (sin caer al respaldo básico) y que
// las marcas de agarre y articulaciones se dibujan en vista micro.
func TestBodyZonesHaveMicroArtInCasa(t *testing.T) {
	g := artGame(t, randomBase)
	f := g.run.Cur
	g.w, g.h = 480, 270

	var checked int
	for _, s := range f.Spaces {
		if s.Kind != space.Articulated && s.Name != "futón" && s.Name != "almohada" {
			continue
		}
		if s.Extent.Empty() {
			continue
		}
		// seat es solo sombra al sentarse (no superficie caminable).
		if g.visuals[s.ID] == "seat" {
			continue
		}
		if g.visuals[s.ID] == "" {
			t.Errorf("superficie %q (%v) no tiene visual asignado", s.Name, s.ID)
			continue
		}
		if !g.art.atlas.Visible(g.visuals[s.ID], render.Micro) {
			t.Errorf("visual %q de %q no es visible en micro", g.visuals[s.ID], s.Name)
			continue
		}
		g.art.focus = s.ID
		g.fitArt()
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		scene := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, s.ID)
		if scene.HasBackdrop() {
			t.Errorf("superficie %q (%s) cayó a vista micro básica", s.Name, g.visuals[s.ID])
		}
		if scene.Clip(s.ID) == "" {
			t.Errorf("superficie %q (%s) no dibujó ningún clip en micro", s.Name, g.visuals[s.ID])
		}
		if len(s.Grips) > 0 && len(scene.Marks(s.ID)) == 0 {
			t.Errorf("zona corporal %q tiene %d agarres pero no dibujó marcas de agarre", s.Name, len(s.Grips))
		}
		checked++
	}
	if checked < 20 {
		t.Fatalf("se esperaban >= 20 superficies corporales/futón/almohada verificadas, hubo %d", checked)
	}
}

// Verifica visualmente los recorridos completos de exploración corporal
// (escritorio→almohada y piso→sofá) siguiendo a Tiny 01 en vista micro:
// en cada zona corporal proyectada en 3D el tiny se mantiene visible dentro
// del encuadre, sobre la baldosa de arte de esa zona, con seguimiento estable.
func TestBodyExplorationMicroVisualsEndToEnd(t *testing.T) {
	cases := []struct {
		name      string
		scenario  string
		ticks     int
		wantZones []string
	}{
		{
			name:      "desk_to_pillow",
			scenario:  "tests/scenarios/body_desk_to_pillow.json",
			ticks:     1050,
			wantZones: []string{"palma", "antebrazo derecho", "brazo derecho", "hombro derecho", "cuello y nuca", "cabello y rostro", "almohada", "futón"},
		},
		{
			name:      "floor_to_sofa",
			scenario:  "tests/scenarios/body_floor_to_sofa.json",
			ticks:     700,
			wantZones: []string{"pie izquierdo", "pantorrilla izquierda", "muslo izquierdo", "sofá"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := artGame(t, tc.scenario)
			g.w, g.h = 480, 270
			g.startFollowing("Tiny 01")
			if g.follow.IsNil() || g.artFocus(g.run.Cur).IsNil() {
				t.Fatal("no se inició el seguimiento en micro de Tiny 01")
			}

			seen := map[string]bool{}
			for range tc.ticks {
				g.run.StepOnce()
				f := g.run.Cur
				th, ok := f.ThingByID(g.follow)
				if !ok {
					t.Fatalf("tick %d: Tiny 01 desapareció", f.Tick)
				}
				focus := g.artFocus(f)
				if focus != th.Space {
					t.Fatalf("tick %d: foco %v, Tiny 01 en %v", f.Tick, g.name(focus), g.name(th.Space))
				}
				sp, _ := f.SpaceByID(focus)
				seen[sp.Name] = true

				g.syncArtZoom(f)
				proj := g.cam.Project(g.anchor(f), g.w, g.h)
				scene := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, focus, g.sel, g.follow)
				if scene.HasBackdrop() {
					t.Fatalf("tick %d (%s): no debía usar vista básica", f.Tick, sp.Name)
				}
				if scene.Clip(th.ID) == "" {
					t.Fatalf("tick %d (%s): Tiny 01 sin clip en micro", f.Tick, sp.Name)
				}
				bounds, ok := scene.Bounds(th.ID)
				if !ok || !bounds.Overlaps(proj.Bounds()) {
					t.Fatalf("tick %d (%s): Tiny 01 fuera de pantalla en micro (bounds=%v canvas=%v)", f.Tick, sp.Name, bounds, proj.Bounds())
				}
				if scene.Clip(sp.ID) == "" {
					t.Fatalf("tick %d (%s): la superficie no se dibujó en micro", f.Tick, sp.Name)
				}
			}

			for _, z := range tc.wantZones {
				if !seen[z] {
					t.Errorf("Tiny 01 nunca pasó visualmente por %q (vistos: %v)", z, seen)
				}
			}
		})
	}
}

// Verifica que en vista micro de una zona corporal articulada los portales
// internos (articulaciones) dibujan la señal de paso (abierta o cerrada).
func TestBodyArticulatedPortalMarksVisible(t *testing.T) {
	g := artGame(t, "tests/scenarios/body_desk_to_pillow.json")
	g.w, g.h = 480, 270
	g.run.StepOnce()
	f := g.run.Cur

	var forearmID ids.ID
	for _, s := range f.Spaces {
		if s.Name == "antebrazo derecho" {
			forearmID = s.ID
			break
		}
	}
	if forearmID.IsNil() {
		t.Fatal("falta antebrazo derecho")
	}
	g.art.focus = forearmID
	g.fitArt()
	proj := g.cam.Project(g.anchor(f), g.w, g.h)
	scene := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, forearmID)

	portalMarks := 0
	for _, pt := range f.Portals {
		if pt.ASpace == forearmID || pt.BSpace == forearmID {
			if len(scene.Marks(pt.ID)) > 0 {
				portalMarks++
			}
		}
	}
	if portalMarks < 2 {
		t.Fatalf("antebrazo derecho debía mostrar señales de paso de sus articulaciones (muñeca y codo), mostró %d", portalMarks)
	}
}

// Regresión: cuando la giantess gira (90°, 180°, 270°), en vista micro de una
// zona corporal (p. ej. pecho) el tiny, los agarres y los portales deben seguir
// alineados dentro del sprite derecho de esa zona, y sin sombra de asiento encima.
func TestBodyMicroAlignmentWhenGiantessRotated(t *testing.T) {
	g := artGame(t, randomBase)
	g.w, g.h = 960, 540
	g.run.StepOnce()

	var giantessID, chestID, tinyID ids.ID
	for _, s := range g.run.Cur.Spaces {
		switch s.Name {
		case "giantess":
			giantessID = s.ID
		case "pecho":
			chestID = s.ID
		}
	}
	for _, th := range g.run.Cur.Things {
		if isTiny(th) {
			tinyID = th.ID
			break
		}
	}
	if giantessID.IsNil() || chestID.IsNil() || tinyID.IsNil() {
		t.Fatalf("faltan entidades: giantess=%v chest=%v tiny=%v", giantessID, chestID, tinyID)
	}

	locs := ecs.Get[comp.Location](g.run.World.ECS())
	if l, ok := locs.Get(tinyID); ok {
		l.Space = chestID
		l.Pos = units.V3(14*units.Centimeter, 18*units.Centimeter, 0)
		locs.Set(tinyID, l)
	}

	for _, rot := range []fixedmath.Angle{fixedmath.QuarterTurn, fixedmath.HalfTurn, 3 * fixedmath.QuarterTurn} {
		if l, ok := locs.Get(giantessID); ok {
			l.Facing = rot
			locs.Set(giantessID, l)
		}
		g.run.StepOnce()
		f := g.run.Cur

		g.art.focus = chestID
		g.fitArt()
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		scene := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, chestID, tinyID)

		chestBounds, ok := scene.Bounds(chestID)
		if !ok {
			t.Fatalf("rot=%v: pecho sin bounds", rot)
		}
		tinyBounds, ok := scene.Bounds(tinyID)
		if !ok {
			t.Fatalf("rot=%v: tiny sin bounds", rot)
		}
		if !tinyBounds.In(chestBounds) {
			t.Fatalf("rot=%v: tiny %v quedó fuera del sprite de pecho %v", rot, tinyBounds, chestBounds)
		}
		for _, pl := range scene.Placements() {
			if pl.Ghost && !pl.Dest.Overlaps(chestBounds) {
				t.Fatalf("rot=%v: marca/portal %v en %v quedó fuera de pecho %v", rot, pl.ID, pl.Dest, chestBounds)
			}
			if pl.Ghost && g.visuals[pl.ID] == "seat" {
				t.Fatalf("rot=%v: no debe dibujarse sombra de asiento sobre una zona corporal", rot)
			}
		}
	}
}

// Verifica que en vista micro del piso se ven ambos pies (calzados) todo el
// tiempo y que al caminar uno se levanta mientras el otro queda apoyado.
func TestMicroBothFeetVisibleAndAnimatedWhileWalking(t *testing.T) {
	g := artGame(t, randomBase)
	g.w, g.h = 960, 540
	g.run.StepOnce()
	f := g.run.Cur

	var giantessID, roomID ids.ID
	for _, s := range f.Spaces {
		if s.Name == "giantess" {
			giantessID, roomID = s.ID, s.Parent
			break
		}
	}
	if giantessID.IsNil() || roomID.IsNil() {
		t.Fatal("falta giantess o habitación")
	}

	// Quieta o caminando, siempre deben existir 2 pies en f.Feet.
	if len(f.Feet) != 2 {
		t.Fatalf("se esperaban 2 pies en Frame.Feet, hay %d", len(f.Feet))
	}

	// Avanzar hasta que la giantess camine y verificar que un pie se eleva (Lift > 0)
	// mientras el otro permanece apoyado (Lift == 0) y ambos se dibujan en micro.
	sawLift := false
	for range 300 {
		g.run.StepOnce()
		f = g.run.Cur
		if len(f.Feet) != 2 {
			t.Fatalf("tick %d: se esperaban 2 pies, hay %d", f.Tick, len(f.Feet))
		}
		lifted, planted := 0, 0
		for _, ft := range f.Feet {
			if ft.Lift > 0.05 {
				lifted++
			} else {
				planted++
			}
		}
		if lifted == 1 && planted == 1 {
			sawLift = true
			roomID = f.Feet[0].Space
			g.art.focus = roomID
			g.cam.Anchor = ids.Nil
			g.cam.X, g.cam.Y = float64(f.Feet[0].Pos.X), float64(f.Feet[0].Pos.Y)
			g.syncArtZoom(f)
			proj := g.cam.Project(g.anchor(f), g.w, g.h)
			scene := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, roomID)
			marks := scene.Marks(giantessID)
			if len(marks) < 4 {
				t.Fatalf("se esperaban al menos 2 pies + 2 sombras en micro, hubo %v", marks)
			}
			break
		}
	}
	if !sawLift {
		t.Fatal("nunca se observó un pie levantado y otro apoyado mientras la giantess camina")
	}
}
