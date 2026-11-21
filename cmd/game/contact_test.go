package main

import (
	"slices"
	"strings"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/render"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/view"
)

const fallScenario = "tests/scenarios/tiny_cae_de_la_palma.json"

func artGame(t *testing.T, path string) *game {
	t.Helper()
	g, err := newGame("../../"+path, 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.enableArt(""); err != nil {
		t.Fatal(err)
	}
	g.w, g.h = (pixelGame{g}).Layout(960, 540)
	g.fit()
	return g
}

func refOf(t *testing.T, path, ref string) ids.ID {
	t.Helper()
	spec, err := scenario.ReadFile("../../" + path)
	if err != nil {
		t.Fatal(err)
	}
	// los IDs salen igual en cada construcción del mismo escenario
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	id := in.Refs[ref]
	if id.IsNil() {
		t.Fatalf("%s: falta %s", path, ref)
	}
	return id
}

// stepTo avanza el juego hasta el tick dado y devuelve el frame actual.
func stepTo(t *testing.T, g *game, tick int) *view.Frame {
	t.Helper()
	for int(g.run.Cur.Tick) < tick {
		g.run.StepOnce()
		if g.run.Err != nil {
			t.Fatal(g.run.Err)
		}
	}
	return g.run.Cur
}

// En cada tick de la mañana, la señal del paso aparece en los dos extremos de
// cada portal de contacto de la superficie enfocada (abierta o cerrada según
// el portal), y en ningún otro momento: no hay señales durante el viaje.
func TestMorningPassSignalMatchesPortals(t *testing.T) {
	g := artGame(t, morningScenario)
	palm := refOf(t, morningScenario, "palm")
	desk2 := refOf(t, morningScenario, "desk2")
	spec, err := scenario.ReadFile("../../" + morningScenario)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for range spec.Ticks {
		f := stepTo(t, g, int(g.run.Cur.Tick)+1)
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		for _, focus := range []ids.ID{palm, g.artDesk, desk2} {
			scene := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, focus)
			for _, pt := range f.Portals {
				want := 0
				if pt.Dynamic && (pt.ASpace == focus || pt.BSpace == focus) {
					want = 2
				}
				marks := scene.Marks(pt.ID)
				if len(marks) != want {
					t.Fatalf("tick %d, foco %s: portal %s (%s–%s) con %d señales, quería %d", f.Tick, g.name(focus), pt.ID, g.name(pt.ASpace), g.name(pt.BSpace), len(marks), want)
				}
				state := "open"
				if pt.Disabled {
					state = "closed"
				}
				for _, m := range marks {
					if m != state {
						t.Fatalf("tick %d: señal %q en un portal %s", f.Tick, m, state)
					}
				}
				if want > 0 {
					seen[g.name(focus)+":"+state]++
				}
			}
		}
		if len(f.Contacts) == 0 && f.Tick > 300 && f.Tick < 560 {
			seen["viaje"]++
		}
	}
	for _, k := range []string{"palma:open", "escritorio:open", "mesa del living:open", "viaje"} {
		if seen[k] == 0 {
			t.Errorf("el recorrido nunca mostró %s: %v", k, seen)
		}
	}
}

// Mientras la mano está suspendida sobre un mueble, antes de apoyarse, su
// sombra se ve en la vista micro de ese mueble. Al apoyarse, la sombra se va
// y la palma se dibuja en el foco.
func TestHandShadowAnticipatesContact(t *testing.T) {
	g := artGame(t, morningScenario)
	palm := refOf(t, morningScenario, "palm")
	desk2 := refOf(t, morningScenario, "desk2")
	spec, err := scenario.ReadFile("../../" + morningScenario)
	if err != nil {
		t.Fatal(err)
	}
	touching := func(f *view.Frame, s ids.ID) bool {
		for _, c := range f.Contacts {
			if c.A == palm && c.B == s || c.B == palm && c.A == s {
				return true
			}
		}
		return false
	}
	shadows := map[ids.ID]int{}
	arrivals := 0
	var prev *view.Frame
	for range spec.Ticks {
		f := stepTo(t, g, int(g.run.Cur.Tick)+1)
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		for _, s := range []ids.ID{g.artDesk, desk2} {
			marks := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, s).Marks(palm)
			shadow := slices.ContainsFunc(marks, func(m string) bool { return strings.HasPrefix(m, "shadow_") })
			if shadow && touching(f, s) {
				t.Fatalf("tick %d: sombra de la palma apoyada en %s", f.Tick, g.name(s))
			}
			if shadow {
				shadows[s]++
			}
			if touching(f, s) && prev != nil && !touching(prev, s) {
				arrivals++
				// el tick anterior al contacto ya mostraba la sombra
				pproj := g.cam.Project(g.anchor(prev), g.w, g.h)
				pm := g.art.atlas.Arrange(prev, pproj, g.visuals, render.Micro, s).Marks(palm)
				if !slices.ContainsFunc(pm, func(m string) bool { return strings.HasPrefix(m, "shadow_") }) {
					t.Errorf("tick %d: la palma llegó a %s sin sombra previa", f.Tick, g.name(s))
				}
			}
		}
		prev = f
	}
	if arrivals != 2 || shadows[g.artDesk] < 10 || shadows[desk2] < 10 {
		t.Errorf("llegadas %d, ticks con sombra %v", arrivals, shadows)
	}
}

// Siguiendo a Tiny 01, la vista micro lo acompaña del escritorio a la palma,
// durante el viaje y al bajar en la mesa del living: el foco es siempre la
// superficie donde está y la cámara queda centrada en él.
func TestFollowTinyThroughTrip(t *testing.T) {
	g := artGame(t, morningScenario)
	g.startFollowing("Tiny 01")
	tiny := g.follow
	if tiny.IsNil() || g.art.focus.IsNil() {
		t.Fatal("no sigue a Tiny 01 en la vista micro")
	}
	spec, err := scenario.ReadFile("../../" + morningScenario)
	if err != nil {
		t.Fatal(err)
	}
	var surfaces []string
	for range spec.Ticks {
		f := stepTo(t, g, int(g.run.Cur.Tick)+1)
		th, ok := f.ThingByID(tiny)
		if !ok {
			t.Fatalf("tick %d: Tiny 01 desapareció", f.Tick)
		}
		if g.follow != tiny || g.artFocus(f) != th.Space {
			t.Fatalf("tick %d: foco %s, Tiny 01 en %s", f.Tick, g.name(g.artFocus(f)), g.name(th.Space))
		}
		if a := g.anchor(f); a.Pos != th.Pose.Pos {
			t.Fatalf("tick %d: cámara en %v, Tiny 01 en %v", f.Tick, a.Pos, th.Pose.Pos)
		}
		if n := g.name(th.Space); len(surfaces) == 0 || surfaces[len(surfaces)-1] != n {
			surfaces = append(surfaces, n)
		}
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		if clip := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, g.artFocus(f), g.sel, g.follow).Clip(tiny); clip == "" {
			t.Fatalf("tick %d: Tiny 01 no se dibuja en %s", f.Tick, g.name(th.Space))
		}
	}
	if want := []string{"escritorio", "palma", "mesa del living"}; !slices.Equal(surfaces, want) {
		t.Errorf("recorrido %v, quería %v", surfaces, want)
	}
}

// En el escenario de la caída, el tiny que se pasa del borde de la palma se
// dibuja cayendo con su sombra en el escritorio durante view.FallTicks, y el
// seguimiento pasa de la palma al escritorio sin perderlo.
func TestFallIsAnimatedAndFollowed(t *testing.T) {
	g := artGame(t, fallScenario)
	desk := refOf(t, fallScenario, "desk1")
	g.startFollowing("Tiny 01")
	tiny := g.follow
	if tiny.IsNil() {
		t.Fatal("no sigue a Tiny 01")
	}
	falling, landed := 0, false
	var lifts []int
	for range 30 {
		f := stepTo(t, g, int(g.run.Cur.Tick)+1)
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		focus := g.artFocus(f)
		scene := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, focus, g.sel, g.follow)
		fl, isFalling := f.FallOf(tiny)
		if isFalling {
			falling++
			if focus != desk || fl.To != desk || scene.Clip(tiny) != "fall" || !slices.Contains(scene.Marks(tiny), "shadow") {
				t.Fatalf("tick %d: foco %s, clip %q, marcas %v", f.Tick, g.name(focus), scene.Clip(tiny), scene.Marks(tiny))
			}
			lifts = append(lifts, scene.Lift(tiny))
		} else if falling > 0 {
			landed = true
			if strings.HasPrefix(scene.Clip(tiny), "fall") || len(scene.Marks(tiny)) > 0 {
				t.Fatalf("tick %d: sigue cayendo (%q, %v)", f.Tick, scene.Clip(tiny), scene.Marks(tiny))
			}
		}
		if g.follow != tiny {
			t.Fatalf("tick %d: perdió el seguimiento", f.Tick)
		}
	}
	if falling != view.FallTicks || !landed {
		t.Fatalf("cayó %d ticks (quería %d), aterrizó %v", falling, view.FallTicks, landed)
	}
	// baja desde arriba, cada vez más rápido, hasta tocar el escritorio
	if lifts[0] <= 0 || !slices.IsSortedFunc(lifts, func(a, b int) int { return b - a }) || lifts[len(lifts)-1] > lifts[0]/4 {
		t.Errorf("alturas de la caída %v", lifts)
	}
}
