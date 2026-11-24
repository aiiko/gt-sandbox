package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/render"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/view"
)

const capturesManifest = "docs/capturas/capturas.json"

var updateCaptures = flag.Bool("update", false, "reescribir los resúmenes de docs/capturas/capturas.json")

// captureSpec es una captura de referencia: escenario, semilla, tick y vista
// para reproducirla con cmd/capturas (o cmd/game -shot), y el resumen de la
// escena que verifica TestReferenceCaptures.
type captureSpec struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Scenario    string    `json:"scenario"`
	Seed        uint64    `json:"seed"`
	Tick        int       `json:"tick"`
	View        string    `json:"view"`
	Micro       string    `json:"micro,omitempty"`
	Follow      string    `json:"follow,omitempty"`
	Pan         []float64 `json:"pan,omitempty"`
	Digest      string    `json:"digest,omitempty"`
}

type captureManifest struct {
	Size     string        `json:"size"`
	Captures []captureSpec `json:"captures"`
}

func readCaptures(t *testing.T) captureManifest {
	t.Helper()
	data, err := os.ReadFile("../../" + capturesManifest)
	if err != nil {
		t.Fatal(err)
	}
	var m captureManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// captureScene reproduce una captura como lo hace -shot: encuadre inicial,
// pasos exactos hasta el tick y el frame actual (en pausa, sin interpolar).
func captureScene(t *testing.T, c captureSpec) (*game, *view.Frame, render.ArtScene, ids.ID) {
	t.Helper()
	g := artGame(t, c.Scenario)
	g.story = c.Scenario == morningScenario
	g.micro, g.followName = c.Micro, c.Follow
	if len(c.Pan) == 2 {
		g.pan = [2]float64{c.Pan[0], c.Pan[1]}
	}
	g.startView()
	stepTo(t, g, int(g.run.Cur.Tick)+c.Tick)
	g.run.Paused = true
	f := g.frame()
	focus := g.artFocus(f)
	scale := render.Macro
	if !focus.IsNil() {
		scale = render.Micro
	}
	proj := g.cam.Project(g.anchor(f), g.w, g.h)
	return g, f, g.art.atlas.Arrange(f, proj, g.visuals, scale, focus, g.sel, g.follow), focus
}

// Cada captura de referencia se reproduce sin GPU y se verifica:
//   - nitidez: cada sprite se copia sin escalar a un destino entero;
//   - orden de dibujo: lo que se apoya en una superficie va después de ella;
//   - animación y contacto: la palma usa rested solo si está en un contacto;
//   - el resumen de la escena (sprites, cuadros, símbolos e interfaz) coincide
//     con el registrado, así un cambio de dibujo se nota y se revisa.
//
// go test ./cmd/game -run TestReferenceCaptures -update reescribe los
// resúmenes; go run ./cmd/capturas regenera los PNG.
func TestReferenceCaptures(t *testing.T) {
	m := readCaptures(t)
	changed := false
	for i, c := range m.Captures {
		spec, err := scenario.ReadFile("../../" + c.Scenario)
		if err != nil {
			t.Fatal(err)
		}
		if spec.Seed != c.Seed {
			t.Errorf("%s: semilla %d, el escenario usa %d", c.Name, c.Seed, spec.Seed)
		}
		g, f, scene, focus := captureScene(t, c)
		if (c.View == "macro") != focus.IsNil() {
			t.Errorf("%s: vista %q, foco %s", c.Name, c.View, g.name(focus))
		}
		things := map[ids.ID]view.Thing{}
		for _, th := range f.Things {
			things[th.ID] = th
		}
		base := map[ids.ID]int{}
		var lines []string
		for k, p := range scene.Placements() {
			sc := max(1, p.Scale)
			if p.Dest.Size() != p.Src.Size().Mul(sc) {
				t.Errorf("%s: %s escalado %v → %v (escala %d)", c.Name, p.ID, p.Src, p.Dest, sc)
			}
			if p.Space && p.Layer != 3 {
				if _, ok := base[p.ID]; !ok {
					base[p.ID] = k
				}
			}
			if th, ok := things[p.ID]; ok && !p.Ghost && p.Layer != 3 {
				if b, ok := base[th.Space]; !ok && slices.ContainsFunc(scene.Placements(), func(q render.Placement) bool { return q.ID == th.Space && q.Space }) {
					t.Errorf("%s: %s se dibuja antes que su superficie", c.Name, th.Name)
				} else if ok && b > k {
					t.Errorf("%s: %s se dibuja antes que su superficie", c.Name, th.Name)
				}
			}
			lines = append(lines, fmt.Sprintf("%s %v %v %d %.2f %v", p.ID, p.Src, p.Dest, p.Layer, p.Alpha, p.Ghost))
		}
		for _, s := range f.Spaces {
			if g.visuals[s.ID] != "palm" {
				continue
			}
			clip := scene.Clip(s.ID)
			touching := slices.ContainsFunc(f.Contacts, func(ct view.Contact) bool { return ct.A == s.ID || ct.B == s.ID })
			if clip != "" && strings.HasPrefix(clip, "rested_") != touching {
				t.Errorf("%s: palma %q con contacto %v", c.Name, clip, touching)
			}
		}
		for _, th := range f.Things {
			if e := f.Emote(th.ID); e != "" {
				lines = append(lines, "emote "+th.Name+" "+e)
			}
		}
		h := g.partida(f, focus)
		lines = append(lines, fmt.Sprintf("hud %s %v %q %s %d", h.Clock, h.Location, h.Following, h.Face, len(h.Story)))
		digest := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(lines, "\n"))))[:16]
		if digest != c.Digest {
			if *updateCaptures {
				m.Captures[i].Digest, changed = digest, true
			} else {
				t.Errorf("%s: la escena cambió (resumen %s, registrado %s); revisar y regenerar con -update y go run ./cmd/capturas", c.Name, digest, c.Digest)
			}
		}
	}
	if changed {
		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("../../"+capturesManifest, append(data, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
