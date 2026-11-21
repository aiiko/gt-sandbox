package main

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gtsandbox/internal/render"
	"gtsandbox/internal/scenario"
)

func directions(clips ...string) []string {
	var out []string
	for _, c := range clips {
		for _, d := range []string{"north", "east", "south", "west"} {
			out = append(out, c+"_"+d)
		}
	}
	return out
}

// requiredClips son los clips que la presentación de la mañana en la casa
// pide a cada visual: caminata, acarreo y caída del tiny, poses de la
// giantess, palma elevada y apoyada, taza en mano.
var requiredClips = map[string]map[render.Scale][]string{
	"giantess":      {render.Macro: directions("idle", "walk", "reach", "sit", "eat")},
	"tiny":          {render.Micro: append(directions("idle", "walk", "carry"), "fall")},
	"tiny_lavender": {render.Micro: append(directions("idle", "walk", "carry"), "fall")},
	"tiny_ochre":    {render.Micro: append(directions("idle", "walk", "carry"), "fall")},
	"palm":          {render.Micro: {"raised", "rested"}},
	"cup":           {render.Macro: {"held"}, render.Micro: {"held"}},
	"crumb":         {render.Micro: {"idle"}},
}

// Cada visual del escenario y del contenido existe en el atlas de la casa y
// tiene los clips que va a pedir la presentación.
func TestMorningAtlasCoversPresentation(t *testing.T) {
	spec, err := scenario.ReadFile("../../" + morningScenario)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Atlas == "" {
		t.Fatal("el escenario no declara atlas")
	}
	path := filepath.Join("../../"+filepath.Dir(morningScenario), spec.Atlas)
	atlas, err := render.LoadAtlas(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		t.Fatal(err)
	}
	used := map[string]bool{}
	for _, e := range spec.Entities {
		if e.Visual != "" {
			used[e.Visual] = true
		}
	}
	objects, err := filepath.Glob("../../content/objects/*.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range objects {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var obj struct {
			Visual string `json:"visual"`
		}
		if err := json.Unmarshal(b, &obj); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		if obj.Visual != "" {
			used[obj.Visual] = true
		}
	}
	for _, id := range slices.Sorted(maps.Keys(used)) {
		if err := atlas.Require(id); err != nil {
			t.Error(err)
		}
	}
	for _, id := range slices.Sorted(maps.Keys(requiredClips)) {
		if !used[id] {
			t.Errorf("visual %q requerido pero ni el escenario ni el contenido lo usan", id)
		}
		for scale, want := range requiredClips[id] {
			have := atlas.Clips(id, scale)
			for _, clip := range want {
				if !slices.Contains(have, clip) {
					t.Errorf("visual %q, %s: falta el clip %q", id, scale, clip)
				}
			}
		}
	}
}
