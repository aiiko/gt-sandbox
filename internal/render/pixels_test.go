package render

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"slices"
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
		canvas image.Point
	}{
		{480, 270, 1, image.Pt(480, 270)},
		{960, 540, 2, image.Pt(480, 270)},
		{1920, 1080, 4, image.Pt(480, 270)},
		{1920, 1009, 3, image.Pt(640, 337)}, // ventana maximizada en 1080p
		{1031, 601, 2, image.Pt(516, 301)},
		{2560, 1440, 5, image.Pt(512, 288)},
		{400, 200, 1, image.Pt(400, 200)},
	} {
		cw, ch := CanvasSize(tc.w, tc.h)
		v := PixelViewport(image.Rect(0, 0, tc.w, tc.h), image.Pt(cw, ch))
		if v.Factor != tc.factor || image.Pt(cw, ch) != tc.canvas {
			t.Fatalf("%dx%d: lienzo %dx%d, viewport %+v", tc.w, tc.h, cw, ch, v)
		}
		// Sin bandas: el lienzo cubre la pantalla y se pasa menos de un
		// píxel del lienzo, repartido entre los dos lados.
		screen := image.Rect(0, 0, tc.w, tc.h)
		if !screen.In(v.Rect) || v.Rect.Dx()-tc.w >= tc.factor || v.Rect.Dy()-tc.h >= tc.factor || v.Rect.Min.X > 0 || v.Rect.Min.Y > 0 {
			t.Fatalf("%dx%d: el lienzo no llena la pantalla: %+v", tc.w, tc.h, v)
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
	proj := (camera.Camera{Zoom: 0.003}).Project(fixedmath.Identity, MinCanvasWidth, MinCanvasHeight)
	s := a.Arrange(f, proj, map[ids.ID]string{1: "table", 2: "person"}, Micro, 1)
	v := PixelViewport(image.Rect(0, 0, 960, 540), image.Pt(MinCanvasWidth, MinCanvasHeight))
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

func TestMacroOcclusionUsesSupportAndForeground(t *testing.T) {
	room := `{"pivot":[1,1],"layer":"background","foreground":[3,0,3,3],"clips":{"idle":{"frame_ticks":1,"frames":[[0,0,3,3]]}}}`
	manifest := `{"image":"atlas.png","scales":["macro"],"sprites":{"room":{"macro":` + room +
		`},"thing":{"macro":` + sampleVariant + `}}}`
	a, err := LoadAtlas(sampleAtlas(t, manifest), "props/atlas.json")
	if err != nil {
		t.Fatal(err)
	}
	frame := &view.Frame{
		Spaces: []view.Space{
			{ID: 1},
			{ID: 2, Parent: 1, Extent: units.RectWH(0, 0, 1000, 1000), Pose: view.Pose{Pos: units.V3(-1000, 0, 2000)}},
		},
		Things: []view.Thing{
			{ID: 3, Space: 2, Pose: view.Pose{Pos: units.V3(-1000, 1000, 2000)}},
			{ID: 4, Space: 1, Pose: view.Pose{Pos: units.V3(-1000, 1000, 0)}},
		},
	}
	visuals := map[ids.ID]string{1: "room", 2: "thing", 3: "thing", 4: "thing"}
	proj := (camera.Camera{Zoom: 0.001}).Project(fixedmath.Identity, MinCanvasWidth, MinCanvasHeight)
	behind := a.Arrange(frame, proj, visuals, Macro, ids.Nil)
	if id := behind.Pick(image.Pt(239, 134)); id != 3 {
		t.Fatalf("detrás de la mesa y su ocupante: se eligió %s", id)
	}
	frame.Things[1].Pose.Pos.Y = -1000
	front := a.Arrange(frame, proj, visuals, Macro, ids.Nil)
	if id := front.Pick(image.Pt(239, 135)); id != 4 {
		t.Fatalf("delante de la mesa: se eligió %s", id)
	}
	if id := front.Pick(image.Pt(240, 135)); id != 1 {
		t.Fatalf("pared de primer plano: se eligió %s", id)
	}
}

func TestArtDirection(t *testing.T) {
	m := &view.Mover{Status: "moving"}
	for _, c := range []struct {
		goal units.Vec3
		name string
	}{
		{units.V3(20, 1, 0), "walk_east"},
		{units.V3(1, 20, 0), "walk_north"},
		{units.V3(-20, 1, 0), "walk_west"},
		{units.V3(1, -20, 0), "walk_south"},
	} {
		m.Route = []units.Vec3{{}, c.goal}
		if got := (actor{mover: m}).clips()[0]; got != c.name {
			t.Errorf("ruta a %v: %s, quería %s", c.goal, got, c.name)
		}
	}
	m.Route = []units.Vec3{{}, units.V3(-20, 1, 0)}
	if got := (actor{mover: m, carrying: true}).clips(); !slices.Equal(got[:2], []string{"carry_west", "walk_west"}) {
		t.Errorf("acarreo hacia el oeste: %v", got)
	}
	still := &view.Mover{Status: "idle"}
	for _, c := range []struct {
		a    actor
		want []string
	}{
		{actor{mover: still, facing: fixedmath.QuarterTurn}, []string{"idle_north"}},
		{actor{mover: still, facing: fixedmath.QuarterTurn, activity: "use:breakfast", step: "use", resting: true},
			[]string{"eat_north", "eat", "reach_north", "reach", "idle_north"}},
		{actor{mover: still, activity: "use:breakfast", step: "pose", resting: true}, []string{"reach_east", "reach", "idle_east"}},
		{actor{mover: still, activity: "use:watch"}, []string{"sit_east", "sit", "idle_east"}},
		{actor{mover: still, activity: "use:watch (suspendida)"}, []string{"idle_east"}},
		{actor{}, []string{"idle"}},
	} {
		if got := c.a.clips(); !slices.Equal(got, c.want) {
			t.Errorf("%+v: %v, quería %v", c.a, got, c.want)
		}
	}
	if got := surfaceClips(fixedmath.HalfTurn, true); !slices.Equal(got, []string{"rested_west", "rested", "idle"}) {
		t.Errorf("superficie apoyada mirando al oeste: %v", got)
	}
	if got := surfaceClips(0, false)[0]; got != "raised_east" {
		t.Errorf("superficie elevada: %s", got)
	}
	v := variant{Clips: map[string]clip{"idle": {}, "walk": {}, "carry_west": {}}}
	for names, want := range map[string]string{
		"carry_east,walk_east,walk": "walk",
		"carry_west,walk_west":      "carry_west",
		"walk_north":                "walk",
		"sit_north,idle_north":      "idle",
	} {
		if got := v.resolve(strings.Split(names, ",")); got != want {
			t.Errorf("resolve(%s) = %s, quería %s", names, got, want)
		}
	}
}

func TestMicroCargoAndContactLayers(t *testing.T) {
	carrier := `{"pivot":[1,1],"carry":[4,-2],"clips":{"idle":{"frame_ticks":1,"frames":[[0,0,3,3]]}}}`
	nest := `{"pivot":[1,1],"layer":"ground","foreground":[3,0,3,3],"clips":{"idle":{"frame_ticks":1,"frames":[[0,0,3,3]]}}}`
	manifest := `{"image":"atlas.png","scales":["micro"],"sprites":{
		"surface":{"micro":` + sampleVariant + `},
		"nest":{"micro":` + nest + `},
		"carrier":{"micro":` + carrier + `}}}`
	a, err := LoadAtlas(sampleAtlas(t, manifest), "props/atlas.json")
	if err != nil {
		t.Fatal(err)
	}
	f := &view.Frame{
		Spaces: []view.Space{
			{ID: 1, Pose: view.Pose{Pos: units.V3(-5000, 0, 0)}},
			{ID: 2, Pose: view.Pose{Pos: units.V3(-5000, 0, 2000)}},
		},
		Things: []view.Thing{
			{ID: 3, Space: 1},
			{ID: 4, Space: 1, CarriedBy: 3, Pose: view.Pose{Pos: units.V3(-9000, 0, 0)}},
			{ID: 5, Space: 1},
		},
		Contacts: []view.Contact{{A: 1, B: 2}},
	}
	proj := (camera.Camera{Zoom: 0.001}).Project(fixedmath.Identity, MinCanvasWidth, MinCanvasHeight)
	visuals := map[ids.ID]string{1: "surface", 2: "surface", 3: "carrier", 4: "surface", 5: "nest"}
	s := a.Arrange(f, proj, visuals, Micro, 1)
	for point, want := range map[image.Point]ids.ID{
		image.Pt(244, 133): 4, // el anclaje del porteador, no el punto propio del objeto
		image.Pt(239, 135): 3, // tiny delante de la base del nido
		image.Pt(240, 135): 5, // borde del nido tapa parte del tiny
		image.Pt(235, 135): 2, // palma más alta en contacto con el escritorio
	} {
		if got := s.Pick(point); got != want {
			t.Errorf("píxel %v: %s, quería %s", point, got, want)
		}
	}
	f.Contacts = nil
	s = a.Arrange(f, proj, visuals, Micro, 1)
	if got := s.Pick(image.Pt(235, 135)); got != 1 {
		t.Fatalf("palma fuera del foco: %s, quería escritorio", got)
	}
}

// Un piso en baldosas cubre su extensión desde la esquina mínima y recorta
// la última fila y columna en el borde, sin invadir al vecino.
func TestTiledFloorCoversExtent(t *testing.T) {
	tile := `{"pivot":[0,0],"layer":"background","tile":true,"clips":{"idle":{"frame_ticks":1,"frames":[[0,0,3,3]]}}}`
	manifest := `{"image":"atlas.png","sprites":{"floor":{"macro":` + sampleVariant + `,"micro":` + tile + `}}}`
	a, err := LoadAtlas(sampleAtlas(t, manifest), "props/atlas.json")
	if err != nil {
		t.Fatal(err)
	}
	f := &view.Frame{Spaces: []view.Space{{ID: 1, Extent: units.Rect{Max: units.V2(2500, 2500)}}}}
	// 0,003 px/µm: cada baldosa de 3 px cubre 1 mm; la extensión, 2,5 mm (7,5 px).
	proj := (camera.Camera{Zoom: 0.003}).Project(fixedmath.Identity, MinCanvasWidth, MinCanvasHeight)
	s := a.Arrange(f, proj, map[ids.ID]string{1: "floor"}, Micro, 1)
	var union image.Rectangle
	n := 0
	for _, p := range s.Placements() {
		if p.ID == 1 && p.Space {
			union, n = union.Union(p.Dest), n+1
			if p.Src.Dx() > 3 || p.Src.Dy() > 3 || p.Dest.Size() != p.Src.Size() {
				t.Errorf("baldosa %v → %v", p.Src, p.Dest)
			}
		}
	}
	if n != 9 || union.Dx() != 8 || union.Dy() != 8 {
		t.Fatalf("%d baldosas cubren %v, se esperaban 9 en 8×8 px", n, union)
	}
	x0, y0 := proj.ToScreen(units.V3(0, 0, 0))
	if union.Min.X != int(x0+0.5) || union.Max.Y != int(y0+0.5) {
		t.Errorf("la esquina mínima %v,%v no queda abajo a la izquierda de %v", x0, y0, union)
	}
	bad := strings.Replace(tile, `"pivot":[0,0]`, `"pivot":[1,1]`, 1)
	manifest = `{"image":"atlas.png","sprites":{"floor":{"macro":` + sampleVariant + `,"micro":` + bad + `}}}`
	if _, err := LoadAtlas(sampleAtlas(t, manifest), "props/atlas.json"); err == nil || !strings.Contains(err.Error(), "baldosa") {
		t.Fatalf("pivote de baldosa: %v", err)
	}
}
