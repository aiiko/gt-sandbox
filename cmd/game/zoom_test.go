package main

import (
	"image"

	"gtsandbox/internal/core/ids"
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

// Siguiendo de cerca a un tiny que pasa a una superficie sin vista de cerca
// (trepa una pata), la vista pasa a la casa centrada en él, y vuelve a la
// de cerca cuando llega a otra que la tiene.
func TestFollowLeavesMicroSurface(t *testing.T) {
	g := artGame(t, randomBase)
	f := g.run.Cur
	climber, desk := ids.Nil, ids.Nil
	for i := 0; i < 6000 && climber.IsNil(); i++ {
		g.run.StepOnce()
		f = g.run.Cur
		for _, th := range f.Things {
			if isTiny(th) && g.visuals[th.Space] == render.ClimbVisual {
				climber = th.ID
			}
		}
	}
	for _, s := range f.Spaces {
		if s.Name == "escritorio" {
			desk = s.ID
		}
	}
	if climber.IsNil() || desk.IsNil() {
		t.Fatal("ningún tiny trepó una pata")
	}
	g.art.focus, g.sel = desk, climber
	g.toggleFollow()
	if got := g.artFocus(f); !got.IsNil() {
		t.Fatalf("trepando: vista de cerca de %s", g.name(got))
	}
	if !g.followsInView(f) || g.art.focus.IsNil() {
		t.Fatal("la cámara dejó de acompañar al tiny o se olvidó de la vista de cerca")
	}
	th, _ := f.ThingByID(climber)
	flat, _ := render.FlattenClimbs(f, g.visuals).ThingByID(climber)
	proj := g.cam.Project(g.anchor(f), g.w, g.h)
	if x, y := proj.ToScreen(flat.Pose.Pos); x != float32(g.w)/2 || y != float32(g.h)/2 {
		t.Fatalf("el tiny no está en el centro: %v,%v", x, y)
	}
	if flat.Pose.Pos.Y == th.Pose.Pos.Y {
		t.Fatal("en macro el tiny sigue corrido en la pata")
	}
	for i := 0; i < 2000 && g.microSurface(g.run.Cur, climber).IsNil(); i++ {
		g.run.StepOnce()
	}
	if s := g.microSurface(g.run.Cur, climber); s.IsNil() || g.artFocus(g.run.Cur) != s {
		t.Fatalf("al llegar arriba no volvió a la vista de cerca (%s)", g.name(g.artFocus(g.run.Cur)))
	}
}
