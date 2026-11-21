package artpack

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

// El atlas versionado tiene que salir de sus fuentes sin cambios.
func TestCasaAtlasMatchesSources(t *testing.T) {
	r, err := Pack(os.DirFS("../../content/art/src/casa"))
	if err != nil {
		t.Fatal(err)
	}
	for file, want := range map[string][]byte{"casa.png": r.PNG, "casa.json": r.Manifest} {
		got, err := os.ReadFile("../../content/art/" + file)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("content/art/%s no coincide con las fuentes: correr go run ./cmd/artpack", file)
		}
	}
}

func pngBytes(t *testing.T, w, h int, fill func(x, y int) color.NRGBA) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetNRGBA(x, y, fill(x, y))
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestPackLayersAndSharedFrames(t *testing.T) {
	red := color.NRGBA{255, 0, 0, 255}
	clear := color.NRGBA{}
	fsys := fstest.MapFS{
		"atlas.json": {Data: []byte(`{"image":"a.png","width":6,"sprites":{
			"s":{"macro":{"pivot":[1,1],"carry":[0,-2],"clips":{
				"idle":{"frame_ticks":1,"frames":[["base.png","top.png"]]},
				"walk":{"frame_ticks":2,"frames":["copy.png","other.png"]}}},
			"micro":{"hidden":true}}}}`)},
		"base.png":  {Data: pngBytes(t, 3, 2, func(x, y int) color.NRGBA { return red })},
		"top.png":   {Data: pngBytes(t, 3, 2, func(x, y int) color.NRGBA { return map[bool]color.NRGBA{true: {0, 0, 255, 255}, false: clear}[x == 0] })},
		"copy.png":  {Data: pngBytes(t, 3, 2, func(x, y int) color.NRGBA { return map[bool]color.NRGBA{true: {0, 0, 255, 255}, false: red}[x == 0] })},
		"other.png": {Data: pngBytes(t, 4, 1, func(x, y int) color.NRGBA { return red })},
	}
	r, err := Pack(fsys)
	if err != nil {
		t.Fatal(err)
	}
	var m outManifest
	if err := json.Unmarshal(r.Manifest, &m); err != nil {
		t.Fatal(err)
	}
	v := m.Sprites["s"]["macro"]
	idle, walk := v.Clips["idle"].Frames[0], v.Clips["walk"].Frames
	if !slicesEqual(idle, walk[0]) {
		t.Fatalf("la composición de capas igual a copy.png debería compartir lugar: %v y %v", idle, walk[0])
	}
	if walk[1][2] != 4 || walk[1][3] != 1 || walk[1][1] != 3 {
		t.Fatalf("el cuadro de 4 px no entra al lado y va al estante siguiente: %v", walk[1])
	}
	if !m.Sprites["s"]["micro"].Hidden || v.Carry[1] != -2 {
		t.Fatalf("se perdieron campos: %+v", m.Sprites["s"])
	}
	img, err := png.Decode(bytes.NewReader(r.PNG))
	if err != nil {
		t.Fatal(err)
	}
	if got := color.NRGBAModel.Convert(img.At(idle[0], idle[1])).(color.NRGBA); got.B != 255 || got.R != 0 {
		t.Fatalf("la capa superior no tapa a la base: %v", got)
	}
	if !strings.Contains(string(r.Manifest), `"pivot": [1, 1]`) {
		t.Fatalf("las listas de números deberían quedar en una línea:\n%s", r.Manifest)
	}
	again, _ := Pack(fsys)
	if !bytes.Equal(again.PNG, r.PNG) || !bytes.Equal(again.Manifest, r.Manifest) {
		t.Fatal("dos armados de las mismas fuentes difieren")
	}
}

func TestPackRejectsInvalidSources(t *testing.T) {
	one := pngBytes(t, 2, 2, func(x, y int) color.NRGBA { return color.NRGBA{A: 255} })
	two := pngBytes(t, 3, 2, func(x, y int) color.NRGBA { return color.NRGBA{A: 255} })
	for name, c := range map[string]struct{ manifest, want string }{
		"falta":       {`{"image":"a.png","width":8,"sprites":{"s":{"macro":{"clips":{"idle":{"frame_ticks":1,"frames":["no.png"]}}}}}}`, `clip "idle", cuadro 1`},
		"tamaños":     {`{"image":"a.png","width":8,"sprites":{"s":{"macro":{"clips":{"idle":{"frame_ticks":1,"frames":[["one.png","two.png"]]}}}}}}`, "mide"},
		"ancho":       {`{"image":"a.png","width":1,"sprites":{"s":{"macro":{"clips":{"idle":{"frame_ticks":1,"frames":["one.png"]}}}}}}`, "más ancho"},
		"ruta":        {`{"image":"a.png","width":8,"sprites":{"s":{"macro":{"clips":{"idle":{"frame_ticks":1,"frames":["../one.png"]}}}}}}`, "inválida"},
		"desconocido": {`{"image":"a.png","width":8,"rect":1,"sprites":{}}`, "unknown field"},
	} {
		_, err := Pack(fstest.MapFS{"atlas.json": {Data: []byte(c.manifest)}, "one.png": {Data: one}, "two.png": {Data: two}})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %v, quería %q", name, err, c.want)
		}
	}
}

func slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
