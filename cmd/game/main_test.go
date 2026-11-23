package main

import (
	"image"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/render"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/view"
)

func TestMorningMicroArtKeepsIdentityAndCamera(t *testing.T) {
	g, err := newGame("../../"+morningScenario, 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.enableArt(""); err != nil {
		t.Fatal(err)
	}
	g.w, g.h = (pixelGame{g}).Layout(960, 540)
	g.fit()
	g.cam.Pan(11, 3)
	macro := g.cam

	var carriedCup ids.ID
	var sawHaul bool
	for range 565 {
		g.run.StepOnce()
		for _, item := range g.run.Cur.Things {
			if !item.CarriedBy.IsNil() {
				switch g.visuals[item.ID] {
				case "cup":
					carriedCup = item.ID
				case "crumb":
					sawHaul = true
				}
			}
		}
	}
	if g.run.Err != nil {
		t.Fatal(g.run.Err)
	}
	f := g.run.Cur
	var cup, tiny view.Thing
	for _, item := range f.Things {
		switch {
		case g.visuals[item.ID] == "cup":
			cup = item
		case item.Name == "Tiny 04":
			tiny = item
		}
	}
	if cup.ID.IsNil() || cup.ID != carriedCup || cup.Space != g.artDesk || !cup.CarriedBy.IsNil() {
		t.Fatalf("la taza debe estar sobre el escritorio: %+v", cup)
	}
	if !sawHaul {
		t.Fatal("el visor no recibió una migaja en acarreo")
	}
	if tiny.ID.IsNil() {
		t.Fatal("falta el tiny")
	}
	pickAround := func(scale render.Scale, target view.Thing, radius int) ids.ID {
		t.Helper()
		g.proj = g.cam.Project(g.anchor(f), g.w, g.h)
		scene := g.art.atlas.Arrange(f, g.proj, g.visuals, scale, g.artFocus(f), g.sel, g.follow)
		x, y := g.proj.ToScreen(target.Pose.Pos)
		p := image.Pt(int(math.Round(float64(x))), int(math.Round(float64(y))))
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if got := scene.Pick(p.Add(image.Pt(dx, dy))); got == target.ID {
					return got
				}
			}
		}
		return ids.Nil
	}
	g.sel = pickAround(render.Macro, cup, 8)
	if g.sel != cup.ID {
		t.Fatal("la taza del desayuno no se puede seleccionar en macro")
	}
	g.toggleArtMicro()
	if g.artFocus(f) != g.artDesk || g.sel != cup.ID || pickAround(render.Micro, cup, 0) != cup.ID {
		t.Fatal("micro debe centrar y mostrar la misma taza")
	}
	g.toggleArtMicro()
	if g.cam != macro || g.sel != cup.ID {
		t.Fatal("al salir de micro se perdió el encuadre anterior o la selección")
	}

	g.sel = tiny.ID
	g.toggleFollow()
	for step, scale := range []render.Scale{render.Macro, render.Micro, render.Macro} {
		if pickAround(scale, tiny, 16) != tiny.ID {
			t.Fatalf("no se seleccionó el tiny en %s", scale)
		}
		x, y := g.proj.ToScreen(tiny.Pose.Pos)
		if x != float32(g.w)/2 || y != float32(g.h)/2 || g.follow != tiny.ID || g.sel != tiny.ID {
			t.Fatalf("%s: se perdió el seguimiento en (%f, %f)", scale, x, y)
		}
		if scale == render.Macro && !g.art.focus.IsNil() {
			t.Fatal("no volvió a macro")
		}
		if scale == render.Micro && g.art.focus != tiny.Space {
			t.Fatal("micro no enfocó el Space del tiny")
		}
		if step < 2 {
			g.toggleArtMicro()
		}
	}
}

func TestMorningMacroArt(t *testing.T) {
	path := "../../" + morningScenario
	g, err := newGame(path, 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.enableArt(""); err != nil {
		t.Fatal(err)
	}
	if !g.art.atlas.Supports(render.Macro) {
		t.Fatal("el atlas de la casa debe ofrecer macro")
	}
	g.w, g.h = (pixelGame{g}).Layout(960, 540)
	g.fit()
	f := g.run.Cur
	g.proj = g.cam.Project(g.anchor(f), g.w, g.h)
	scene := g.art.atlas.Arrange(f, g.proj, g.visuals, render.Macro, ids.Nil)
	screen := func(v units.Vec3) image.Point {
		x, y := g.proj.ToScreen(v)
		return image.Pt(int(math.Round(float64(x))), int(math.Round(float64(y))))
	}
	gi, _ := f.SpaceByID(g.giantessID())
	// al empezar está parada sobre el borde del futón: se elige su brazo este
	g.sel = scene.Pick(screen(gi.Pose.Pos).Add(image.Pt(7, -28)))
	if g.visuals[g.sel] != "giantess" {
		t.Fatalf("clic en la giantess devolvió %s", g.name(g.sel))
	}
	for _, pt := range f.Portals {
		want := g.visuals[pt.ID]
		if want != "door" && want != "door_noren" {
			continue
		}
		at := screen(units.V3(pt.A.X+(pt.B.X-pt.A.X)/2, pt.A.Y+(pt.B.Y-pt.A.Y)/2, pt.A.Z))
		if id := scene.Pick(at); g.visuals[id] != want {
			t.Errorf("umbral en %v: se seleccionó %s, quería %s", at, g.name(id), want)
		}
	}
	g.toggleFollow()
	spec, err := scenario.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var visited []string
	for range spec.Ticks {
		g.run.StepOnce()
		if g.run.Err != nil {
			t.Fatal(g.run.Err)
		}
		f = view.Interpolate(g.run.Prev, g.run.Cur, 0.5)
		actor, ok := f.SpaceByID(g.follow)
		if !ok {
			t.Fatal("se perdió el seguimiento")
		}
		room := g.name(actor.Parent)
		if len(visited) == 0 || visited[len(visited)-1] != room {
			visited = append(visited, room)
		}
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		if x, y := proj.ToScreen(actor.Pose.Pos); x != float32(g.w)/2 || y != float32(g.h)/2 {
			t.Fatalf("seguimiento en %s: punto de apoyo en (%f, %f)", room, x, y)
		}
	}
	want := []string{"dormitorio", "cocina", "dormitorio", "living"}
	if !slices.Equal(visited, want) {
		t.Fatalf("recorrido de cámara %v, quería %v", visited, want)
	}
}

func TestArtSampleUsesSameEntitiesAtBothScales(t *testing.T) {
	g, err := newGame("../../"+artScenario, 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.enableArt("../../content/art/muestra.json"); err != nil {
		t.Fatal(err)
	}
	g.w, g.h = (pixelGame{g}).Layout(1031, 601)
	g.fit()
	f := g.run.Cur
	var cup, tiny ids.ID
	for id, visual := range g.visuals {
		switch visual {
		case "cup":
			cup = id
		case "tiny":
			tiny = id
		}
	}
	for _, layout := range []struct {
		scale     render.Scale
		cup, tiny image.Point
	}{
		{render.Macro, image.Pt(228, 134), image.Pt(235, 137)},
		{render.Micro, image.Pt(135, 85), image.Pt(229, 214)},
	} {
		// Las posiciones son las del lienzo mínimo; el de la ventana de
		// 1031×601 (516×301) lo rodea centrado.
		shift := image.Pt((g.w-480)/2, (g.h-270)/2)
		layout.cup, layout.tiny = layout.cup.Add(shift), layout.tiny.Add(shift)
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		scene := g.art.atlas.Arrange(f, proj, g.visuals, layout.scale, g.art.focus)
		if got := scene.Pick(layout.cup); cup.IsNil() || got != cup {
			t.Errorf("%s: seleccionar la taza devolvió %s, quería %s", layout.scale, got, cup)
		}
		if got := scene.Pick(layout.tiny); tiny.IsNil() || got != tiny {
			t.Errorf("%s: seleccionar al tiny devolvió %s, quería %s", layout.scale, got, tiny)
		}
		g.sel = g.artDesk
		g.toggleMicro()
	}
}

func TestMorningViewer(t *testing.T) {
	g, err := newGame("../../"+morningScenario, 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	g.w, g.h = 1280, 800
	g.fit()
	for _, t := range g.run.Cur.Things {
		if t.Name == "Tiny 03" {
			g.sel = t.ID
			break
		}
	}
	g.toggleFollow()
	if g.follow.IsNil() {
		t.Fatal("no se activó el seguimiento")
	}

	var surfaces []string
	for range 760 {
		g.run.StepOnce()
		if g.run.Err != nil {
			t.Fatal(g.run.Err)
		}
		f := view.Interpolate(g.run.Prev, g.run.Cur, 0.5)
		tiny, ok := f.ThingByID(g.follow)
		if !ok {
			t.Fatal("se perdió la entidad seguida")
		}
		g.proj = g.cam.Project(g.anchor(f), g.w, g.h)
		x, y := g.proj.ToScreen(tiny.Pose.Pos)
		if x != 640 || y != 400 {
			t.Fatalf("la cámara perdió el centro de %s: (%f, %f)", tiny.Name, x, y)
		}
		s, _ := f.SpaceByID(tiny.Space)
		if len(surfaces) == 0 || surfaces[len(surfaces)-1] != s.Name {
			surfaces = append(surfaces, s.Name)
		}
	}
	if want := []string{"escritorio", "palma", "mesa del living"}; !slices.Equal(surfaces, want) {
		t.Fatalf("la cámara acompañó al tiny en %v, esperaba %v", surfaces, want)
	}
	center := g.proj.ToWorld(640, 400)
	g.toggleFollow()
	proj := g.cam.Project(g.anchor(g.run.Cur), g.w, g.h)
	if got := proj.ToWorld(640, 400); !g.follow.IsNil() || got != center {
		t.Fatalf("la cámara saltó al dejar de seguir: %v, quería %v", got, center)
	}

	pending := []string{
		"Desayuna", "deja la taza en escritorio", "Caen migajas del desayuno",
		"asigna un acarreo", "sube a la mano", "levanta la mano", "entra en living",
		"baja de la mano", "explora mesa del living",
	}
	for _, e := range g.run.Chronicle.Recent() {
		if len(pending) > 0 && strings.Contains(e.Text, pending[0]) {
			pending = pending[1:]
		}
	}
	if len(pending) != 0 {
		t.Fatalf("faltan eventos en la historia: %v", pending)
	}
}

// Las flechas y WASD mueven la cámara igual; la flecha y su letra juntas no
// suman, y direcciones opuestas se anulan.
func TestPanKeys(t *testing.T) {
	keys := func(ks ...ebiten.Key) func(ebiten.Key) bool {
		return func(k ebiten.Key) bool { return slices.Contains(ks, k) }
	}
	for _, c := range []struct {
		keys   []ebiten.Key
		dx, dy float64
	}{
		{nil, 0, 0},
		{[]ebiten.Key{ebiten.KeyA}, panStep, 0},
		{[]ebiten.Key{ebiten.KeyArrowLeft}, panStep, 0},
		{[]ebiten.Key{ebiten.KeyA, ebiten.KeyArrowLeft}, panStep, 0},
		{[]ebiten.Key{ebiten.KeyD}, -panStep, 0},
		{[]ebiten.Key{ebiten.KeyW}, 0, panStep},
		{[]ebiten.Key{ebiten.KeyS}, 0, -panStep},
		{[]ebiten.Key{ebiten.KeyW, ebiten.KeyD}, -panStep, panStep},
		{[]ebiten.Key{ebiten.KeyA, ebiten.KeyD}, 0, 0},
	} {
		if dx, dy := panKeys(keys(c.keys...)); dx != c.dx || dy != c.dy {
			t.Errorf("%v: %v,%v; se esperaba %v,%v", c.keys, dx, dy, c.dx, c.dy)
		}
	}
}
