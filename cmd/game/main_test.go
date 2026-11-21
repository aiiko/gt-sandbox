package main

import (
	"image"
	"slices"
	"strings"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/render"
	"gtsandbox/internal/view"
)

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
	hash := g.run.World.Hash()
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
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		scene := g.art.atlas.Arrange(f, proj, g.visuals, layout.scale, g.art.focus)
		if got := scene.Pick(layout.cup); cup.IsNil() || got != cup {
			t.Errorf("%s: seleccionar la taza devolvió %s, quería %s", layout.scale, got, cup)
		}
		if got := scene.Pick(layout.tiny); tiny.IsNil() || got != tiny {
			t.Errorf("%s: seleccionar al tiny devolvió %s, quería %s", layout.scale, got, tiny)
		}
		g.sel = tiny
		g.toggleMicro()
	}
	if got := g.run.World.Hash(); got != hash {
		t.Fatalf("la presentación cambió el hash: %x, quería %x", got, hash)
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
		"preparar el desayuno", "deja la taza en escritorio", "Aparecen migajas",
		"asigna un acarreo", "sube a la mano", "levanta la mano", "entra en living",
		"baja de la mano", "explora mesa del living",
	}
	for _, e := range g.run.Story {
		if len(pending) > 0 && strings.Contains(e.Text, pending[0]) {
			pending = pending[1:]
		}
	}
	if len(pending) != 0 {
		t.Fatalf("faltan eventos en la historia: %v", pending)
	}
}
