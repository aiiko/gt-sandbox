package main

import (
	"image"
	"testing"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/render"
)

// La rueda cambia el factor de la capa de escena de a un paso entero, dentro
// de su rango, con el punto bajo el cursor fijo. Alejada a 1× en 1080p, la
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
		t.Fatalf("factor de arranque %d, quería el del lienzo (2)", got)
	}
	cursor := [2]float64{700, 150}
	before := g.proj.ToWorld(cursor[0], cursor[1])
	g.wheelZoom(-1, cursor)
	if after := project().ToWorld(cursor[0], cursor[1]); before.Sub(after).Len() > 20000 {
		t.Fatalf("el punto bajo el cursor se movió de %v a %v", before, after)
	}
	if got := g.art.sceneFactor(render.Macro); got != 1 {
		t.Fatalf("alejado: factor %d", got)
	}
	g.wheelZoom(-1, cursor) // ya en el mínimo
	g.wheelZoom(-0.5, cursor)
	if got := g.art.sceneFactor(render.Macro); got != 1 {
		t.Fatalf("pasó el mínimo: factor %d", got)
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
		t.Fatalf("solo %d habitaciones a la vista", rooms)
	}

	g.wheelZoom(5, cursor)
	if got := g.art.sceneFactor(render.Macro); got != 4 {
		t.Fatalf("acercado: factor %d, quería el máximo (4)", got)
	}
	if got := g.art.sceneFactor(render.Micro); got != 2 {
		t.Fatalf("micro sin tocar: factor %d", got)
	}
	delete(g.art.factor, render.Macro) // lo que hace F
	if got := g.art.sceneFactor(render.Macro); got != 2 {
		t.Fatalf("después de F: factor %d", got)
	}
}
