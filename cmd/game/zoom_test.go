package main

import (
	"image"
	"testing"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/render"
)

// La rueda recorre los niveles de la capa de escena de a uno, sin pasarse de
// los extremos, con el punto bajo el cursor fijo. Alejada a 1× en 1080p, la
// casa entera entra en pantalla; F vuelve al factor del lienzo.
func TestWheelZoomSteps(t *testing.T) {
	g := artGame(t, morningScenario)
	g.art.viewport = render.PixelViewport(image.Rect(0, 0, 1920, 1080), image.Pt(g.w, g.h))
	g.fit()
	f := g.run.Cur
	project := func() camera.Projector {
		g.syncArtZoom(f)
		g.proj = g.cam.Project(fixedmath.Identity, g.w, g.h)
		return g.proj
	}
	project()
	if got := g.art.sceneFactor(render.Macro); got != 2 {
		t.Fatalf("factor de arranque %v, quería el del lienzo (2)", got)
	}
	cursor := [2]float64{700, 150}
	before := g.proj.ToWorld(cursor[0], cursor[1])
	g.wheelZoom(-1, cursor)
	if after := project().ToWorld(cursor[0], cursor[1]); before.Sub(after).Len() > 20000 {
		t.Fatalf("el punto bajo el cursor se movió de %v a %v", before, after)
	}
	if got := g.art.sceneFactor(render.Macro); got != 1 {
		t.Fatalf("alejado: factor %v", got)
	}
	for _, want := range []float64{0.5, 0.5} {
		g.wheelZoom(-1, cursor)
		if got := g.art.sceneFactor(render.Macro); got != want {
			t.Fatalf("alejando: factor %v, quería %v", got, want)
		}
	}
	// Reducida a ½, la capa de escena tiene 2 × 2 píxeles del arte por
	// píxel físico.
	if size := g.sceneLayer(render.Macro).Size(); size != image.Pt(3840, 2160) {
		t.Fatalf("capa reducida %v", size)
	}
	g.wheelZoom(-0.5, cursor) // media muesca no es un paso
	g.wheelZoom(1.5, cursor)
	if got := g.art.sceneFactor(render.Macro); got != 1 {
		t.Fatalf("acercando un nivel: factor %v", got)
	}

	g.fit() // re-centra con el factor alejado
	layer := g.sceneLayer(render.Macro)
	sp := project().Resized(g.artZoom(render.Macro), layer.Size().X, layer.Size().Y)
	if layer.Size() != image.Pt(1920, 1080) {
		t.Fatalf("capa de escena %v", layer.Size())
	}
	scene := g.art.atlas.Arrange(f, sp, g.visuals, render.Macro, g.artFocus(f))
	screen, rooms := image.Rectangle{Max: layer.Size()}, 0
	for _, s := range f.Spaces {
		if s.Name == "" || g.visuals[s.ID] == "" {
			continue
		}
		if r, ok := scene.Bounds(s.ID); ok {
			rooms++
			if !r.In(screen) {
				t.Errorf("alejada, %s queda fuera: %v", s.Name, r)
			}
		}
	}
	if rooms < 3 {
		t.Fatalf("solo %v habitaciones a la vista", rooms)
	}

	g.wheelZoom(10, cursor)
	if got := g.art.sceneFactor(render.Macro); got != 8 {
		t.Fatalf("acercado: factor %v, quería el máximo (8)", got)
	}
	if got := g.art.sceneFactor(render.Micro); got != 2 {
		t.Fatalf("micro sin tocar: factor %v", got)
	}
	delete(g.art.factor, render.Macro) // lo que hace F
	if got := g.art.sceneFactor(render.Macro); got != 2 {
		t.Fatalf("después de F: factor %v", got)
	}
}
