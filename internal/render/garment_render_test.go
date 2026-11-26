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

func TestGarmentMicroThreatShadowCutawayAndMacroHidden(t *testing.T) {
	atlas := &Atlas{
		sprites: map[string]map[Scale]variant{
			"slipper": {
				Macro: variant{
					Pivot: []int{7, 12},
					Layer: "ground",
					Clips: map[string]clip{
						"idle": {FrameTicks: 1, Frames: [][]int{{0, 0, 14, 24}}},
					},
				},
				Micro: variant{
					Pivot: []int{104, 216},
					Layer: "background",
					Clips: map[string]clip{
						"idle":    {FrameTicks: 1, Frames: [][]int{{0, 0, 208, 432}}},
						"donning": {FrameTicks: 1, Frames: [][]int{{0, 0, 208, 432}}},
					},
				},
			},
			"tiny": {
				Micro: variant{
					Pivot: []int{8, 14},
					Clips: map[string]clip{
						"idle":       {FrameTicks: 1, Frames: [][]int{{0, 0, 16, 16}}},
						"idle_south": {FrameTicks: 1, Frames: [][]int{{0, 0, 16, 16}}},
					},
				},
			},
		},
	}

	roomID := ids.ID(1)
	slipperID := ids.ID(2)
	tinyID := ids.ID(3)
	cm := units.Centimeter

	f := &view.Frame{
		Tick: 1,
		Spaces: []view.Space{
			{ID: roomID, Name: "dormitorio", Kind: space.Static, Extent: units.RectWH(0, 0, 300*cm, 300*cm)},
			{
				ID:           slipperID,
				Name:         "pantufla izquierda",
				Kind:         space.Rigid,
				Parent:       roomID,
				Pose:         view.Pose{Pos: units.V3(100*cm, 100*cm, 0)},
				Extent:       units.Rect{Min: units.V2(-5*cm, -12*cm), Max: units.V2(5*cm, 12*cm)},
				BodyZone:     true,
				Occluded:     true,
				GarmentState: "donning",
			},
		},
		Things: []view.Thing{
			{
				ID:       tinyID,
				Name:     "Tiny 01",
				Space:    slipperID,
				Pose:     view.Pose{Pos: units.V3(100*cm, 109*cm, 0)},
				LocalPos: units.V3(0, 9*cm, 0),
				Body:     true,
			},
		},
		Threats: []view.Threat{
			{
				By:       slipperID,
				Surface:  slipperID,
				Source:   "garment",
				Zone:     units.Rect{Min: units.V2(-4*cm, -8*cm), Max: units.V2(4*cm, 7*cm)},
				Progress: 0.5,
			},
		},
	}

	visuals := map[ids.ID]string{slipperID: "slipper", tinyID: "tiny"}
	cam := camera.Camera{X: float64(100 * cm), Y: float64(100 * cm), Zoom: float64(DefaultMicroDensity) / float64(units.Meter)}
	proj := cam.Project(fixedmath.Identity, 480, 270)

	microScene := atlas.Arrange(f, proj, visuals, Micro, slipperID, tinyID)
	if microScene.Clip(slipperID) != "donning" {
		t.Fatalf("esperaba clip donning en micro, obtuve %q", microScene.Clip(slipperID))
	}
	if microScene.ThreatShadows() != 1 {
		t.Fatalf("esperaba 1 sombra de amenaza en micro, obtuve %d", microScene.ThreatShadows())
	}
	if microScene.cutaway == nil || microScene.cutaway.hole.Empty() {
		t.Fatalf("esperaba cutaway activo en prenda Occluded")
	}

	// En macro cuando está worn no debe dibujarse su sprite.
	f.Spaces[1].GarmentState = "worn"
	macroCam := camera.Camera{X: float64(100 * cm), Y: float64(100 * cm), Zoom: 80.0 / float64(units.Meter)}
	macroScene := atlas.Arrange(f, macroCam.Project(fixedmath.Identity, 480, 270), visuals, Macro, ids.Nil)
	if macroScene.Clip(slipperID) != "" {
		t.Fatalf("prenda worn no debía dibujarse en macro, obtuvo %q", macroScene.Clip(slipperID))
	}
}
