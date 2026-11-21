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
	if got := strings.Join(lib.Names(), ","); got != "bed,crumb,fridge,tv" {
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
