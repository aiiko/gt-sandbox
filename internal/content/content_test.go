package content

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"
)

func TestDefaultAndMerge(t *testing.T) {
	lib, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(lib.Names(), ","); got != "bed,chair,crumb,cup,desk,fiber,fridge,napkin,nest,plate,puddle,refuge_site,shower_tray,sink,sofa,tv,washbasin,water_drop" {
		t.Fatalf("objetos: %s", got)
	}
	raw, err := lib.Spawn("crumb", json.RawMessage(`{"name":"c1","components":{"mover":{"speed":1}}}`))
	if err != nil {
		t.Fatal(err)
	}
	var sp struct {
		Name       string
		Body       map[string]int
		Components map[string]json.RawMessage
	}
	if err := json.Unmarshal(raw, &sp); err != nil || sp.Name != "c1" || sp.Body["radius"] != 1500 ||
		sp.Components["offers"] == nil || sp.Components["mover"] == nil {
		t.Fatalf("%s", raw)
	}
}

func TestLoadRejects(t *testing.T) {
	for name, body := range map[string]string{
		"campo desconocido":      `{"name":"x","colour":1}`,
		"componente desconocido": `{"name":"x","components":{"nope":{}}}`,
		"oferta inválida":        `{"name":"x","components":{"offers":[{"action":"a","duraton":3}]}}`,
		"nombre distinto":        `{"name":"y"}`,
	} {
		fs := fstest.MapFS{"objects/x.json": {Data: []byte(body)}}
		if _, err := Load(fs); err == nil {
			t.Errorf("%s: se aceptó", name)
		}
	}
}

// "item": "<objeto>" se expande a la definición completa; un ciclo es error.
func TestItems(t *testing.T) {
	lib, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := lib.Spawn("fridge", nil)
	if !strings.Contains(string(raw), `"name":"cup"`) || !strings.Contains(string(raw), `"portions"`) {
		t.Fatalf("items sin expandir: %s", raw)
	}
	obj := func(name, item string) string {
		return `{"name":"` + name + `","components":{"offers":[{"action":"a","scale":"s","effects":[],` +
			`"steps":[{"do":"use","ticks":1},{"do":"spawn","item":"` + item + `"}]}]}}`
	}
	fs := fstest.MapFS{"objects/x.json": {Data: []byte(obj("x", "y"))}, "objects/y.json": {Data: []byte(obj("y", "x"))}}
	if _, err := Load(fs); err == nil || !strings.Contains(err.Error(), "ciclo") {
		t.Fatalf("ciclo aceptado: %v", err)
	}
}
