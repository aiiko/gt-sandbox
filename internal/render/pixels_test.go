package render

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"testing/fstest"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/view"
)

func TestPixelViewport(t *testing.T) {
	for _, tc := range []struct {
		w, h   int
		factor int
		origin image.Point
	}{
		{480, 270, 1, image.Pt(0, 0)},
		{960, 540, 2, image.Pt(0, 0)},
		{1920, 1080, 4, image.Pt(0, 0)},
		{1031, 601, 2, image.Pt(35, 30)},
	} {
		v := PixelViewport(image.Rect(0, 0, tc.w, tc.h))
		if v.Factor != tc.factor || v.Rect.Min != tc.origin {
			t.Fatalf("%dx%d: viewport %+v", tc.w, tc.h, v)
		}
		point := image.Pt(123, 95)
		at := v.Rect.Min.Add(point.Mul(tc.factor))
		for ox := range tc.factor {
			for oy := range tc.factor {
				if got, ok := v.Point(float64(at.X+ox), float64(at.Y+oy)); !ok || got != point {
					t.Fatalf("%dx%d: puntero %v, quería %v", tc.w, tc.h, got, point)
				}
			}
		}
		for _, outside := range []image.Point{v.Rect.Min.Sub(image.Pt(1, 0)), v.Rect.Max.Sub(image.Pt(0, 1))} {
			if _, ok := v.Point(float64(outside.X), float64(outside.Y)); ok {
				t.Fatalf("%dx%d: se aceptó la banda en %v", tc.w, tc.h, outside)
			}
		}
	}
}

const sampleVariant = `{"pivot":[1,1],"clips":{"idle":{"frame_ticks":1,"frames":[[0,0,3,3]]}}}`

func sampleAtlas(t *testing.T, manifest string) fstest.MapFS {
	t.Helper()
	im := image.NewNRGBA(image.Rect(0, 0, 8, 4))
	for y := range 3 {
		for x := range 3 {
			im.SetNRGBA(x, y, color.NRGBA{R: 200, A: 255})
		}
	}
	im.SetNRGBA(4, 1, color.NRGBA{G: 200, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, im); err != nil {
		t.Fatal(err)
	}
	return fstest.MapFS{
		"props/atlas.json": {Data: []byte(manifest)},
		"props/atlas.png":  {Data: buf.Bytes()},
	}
}

func TestAtlasRejectsInvalidExports(t *testing.T) {
	for _, tc := range []struct{ micro, want string }{
		{`null`, "pivote inválido"},
		{strings.Replace(sampleVariant, "[1,1]", "[3,1]", 1), "fuera del cuadro"},
		{strings.Replace(sampleVariant, "[0,0,3,3]", "[7,0,3,3]", 1), "rectángulo fuera"},
		{strings.Replace(sampleVariant, "[0,0,3,3]", "[0,0,3]", 1), "rectángulo fuera"},
		{strings.Replace(sampleVariant, `"frame_ticks":1`, `"frame_ticks":0`, 1), "clip \"idle\""},
		{strings.Replace(sampleVariant, `"idle"`, `"walk"`, 1), "falta el clip idle"},
	} {
		manifest := `{"image":"atlas.png","sprites":{"cup":{"macro":` + sampleVariant + `,"micro":` + tc.micro + `}}}`
		_, err := LoadAtlas(sampleAtlas(t, manifest), "props/atlas.json")
		if err == nil || !strings.Contains(err.Error(), `atlas.json, sprite "cup", variante micro`) || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("micro %s: error %v, quería %q e identificación del recurso", tc.micro, err, tc.want)
		}
	}
	manifest := `{"image":"atlas.png","sprites":{"cup":{"macro":` + sampleVariant + `}}}`
	if _, err := LoadAtlas(sampleAtlas(t, manifest), "props/atlas.json"); err == nil || !strings.Contains(err.Error(), "variante micro: falta") {
		t.Fatalf("variante ausente: %v", err)
	}
}

func TestSpritePickingUsesSnappedInterpolatedFrame(t *testing.T) {
	walk := `{"pivot":[1,1],"clips":{"idle":{"frame_ticks":1,"frames":[[0,0,3,3]]},"walk":{"frame_ticks":2,"frames":[[0,0,3,3],[3,0,3,3]]}}}`
	manifest := `{"image":"atlas.png","sprites":{
		"table":{"macro":` + sampleVariant + `,"micro":` + sampleVariant + `},
		"person":{"macro":` + sampleVariant + `,"micro":` + walk + `}}}`
	a, err := LoadAtlas(sampleAtlas(t, manifest), "props/atlas.json")
	if err != nil {
		t.Fatal(err)
	}
	previous := &view.Frame{Things: []view.Thing{{ID: 2, Pose: view.Pose{Pos: units.V3(400, 0, 0)}}}}
	current := &view.Frame{
		Tick:   2,
		Spaces: []view.Space{{ID: 1, Pose: view.Pose{Pos: units.V3(200, 0, 0)}}},
		Things: []view.Thing{{ID: 2, Space: 1, Name: "Otro nombre", Mover: &view.Mover{Status: "moving"}}},
	}
	f := view.Interpolate(previous, current, 0.5)
	proj := (camera.Camera{Zoom: 0.003}).Project(fixedmath.Identity, CanvasWidth, CanvasHeight)
	s := a.Arrange(f, proj, map[ids.ID]string{1: "table", 2: "person"}, Micro, 1)
	v := PixelViewport(image.Rect(0, 0, 1031, 601))
	at := v.Rect.Min.Add(image.Pt(241, 135).Mul(v.Factor))
	point, _ := v.Point(float64(at.X), float64(at.Y))
	if got := s.Pick(point); got != 2 {
		t.Fatalf("pivote del sprite interpolado: %v, quería tiny", got)
	}
	if got := s.Pick(point.Sub(image.Pt(1, 0))); got != 1 {
		t.Fatalf("píxel transparente: %v, quería mesa", got)
	}
	if got := s.Pick(image.Pt(239, 135)); !got.IsNil() {
		t.Fatalf("fuera de los sprites: %v", got)
	}
}
