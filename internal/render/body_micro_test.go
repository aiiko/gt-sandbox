package render

import (
	"testing"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/view"
)

func TestMicroBodyClimbAndOccludedCutaway(t *testing.T) {
	atlas := &Atlas{
		sprites: map[string]map[Scale]variant{},
	}
	bodyID := ids.ID(1)
	sleeveID := ids.ID(2)
	tinyID := ids.ID(3)

	f := &view.Frame{
		Tick: 1,
		Spaces: []view.Space{
			{ID: bodyID, Name: "giantess", Kind: space.Body},
			{
				ID:       sleeveID,
				Name:     "antebrazo izquierdo",
				Kind:     space.Articulated,
				Parent:   bodyID,
				Extent:   units.RectWH(0, 0, 6*units.Centimeter, 12*units.Centimeter),
				Climb:    true,
				Occluded: true,
			},
		},
		Things: []view.Thing{
			{
				ID:    tinyID,
				Name:  "Tiny 01",
				Space: sleeveID,
				Pose:  view.Pose{Pos: units.V3(3*units.Centimeter, 6*units.Centimeter, 0)},
				Body:  true,
			},
		},
	}

	cam := camera.Camera{X: float64(3 * units.Centimeter), Y: float64(6 * units.Centimeter), Zoom: 0.001}
	proj := cam.Project(fixedmath.Identity, 320, 180)
	scene := atlas.Arrange(f, proj, map[ids.ID]string{}, Micro, sleeveID, tinyID)

	if !scene.HasBackdrop() {
		t.Fatal("una zona corporal sin arte propio debía usar vista micro básica (HasBackdrop)")
	}
	if scene.Clip(sleeveID) != "basic" {
		t.Fatalf("clip esperado 'basic', obtenido %q", scene.Clip(sleeveID))
	}
	if !scene.backdrop.climb {
		t.Fatal("zona corporal con Climb=true debía activar vista de costado de trepada")
	}
	if scene.cutaway == nil {
		t.Fatal("zona corporal con Occluded=true debía activar recorte alrededor del tiny seguido")
	}
}
