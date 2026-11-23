package main

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/view"
)

// lifeGame abre la partida libre de una semilla con el arte de la casa.
func lifeGame(t *testing.T, seed uint64, ticks int) *game {
	t.Helper()
	const base = "../../" + randomBase
	spec, err := scenario.ReadFile(base)
	if err != nil {
		t.Fatal(err)
	}
	if spec, err = scenario.Randomize(spec, seed); err != nil {
		t.Fatal(err)
	}
	spec.Ticks = ticks
	g, err := newScenarioGame(spec, base)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.enableArt(""); err != nil {
		t.Fatal(err)
	}
	g.w, g.h = (pixelGame{g}).Layout(960, 540)
	g.fit()
	return g
}

// M44: siguiendo a los tinies de la partida libre se distinguen la
// búsqueda, el acarreo con su destino, el regreso al nido por hambre o
// cansancio y la reunión (texto de la ficha y símbolo), y la historia cuenta
// hallazgos, entregas y reuniones.
func TestPartidaLifeReadable(t *testing.T) {
	if testing.Short() {
		t.Skip("cinco minutos de partida")
	}
	const ticks = 6000
	want := map[string]bool{
		"Va a revisar": false, "Revisa": false, "Va a buscar": false, "Lleva": false,
		"Vuelve al nido a comer": false, "Vuelve al nido a descansar": false,
		"Espera a alguien para conversar": false, "Conversa con": false,
	}
	emotes := map[string]bool{}
	story := map[string]bool{"encuentra comida": false, "Una fibra más": false, "conversan": false}
	for _, seed := range []uint64{1, 99} { // la 1 arma el refugio; en la 99 encuentran comida
		g := lifeGame(t, seed, ticks)
		for range ticks {
			f := stepTo(t, g, int(g.run.Cur.Tick)+1)
			for _, th := range f.Things {
				if th.Needs == nil || th.Name == "giantess" {
					continue
				}
				c := g.card(f, th.ID)
				if c == nil {
					continue
				}
				for prefix := range want {
					if strings.HasPrefix(c.Activity, prefix) {
						want[prefix] = true
					}
				}
				switch th.Activity {
				case "search", "chat", "use:rest":
					if e := f.Emote(th.ID); e != "" {
						emotes[th.Activity+":"+e] = true
					}
				}
			}
		}
		for _, e := range g.run.Chronicle.Recent() {
			for k := range story {
				if strings.Contains(e.Text, k) {
					story[k] = true
				}
			}
		}
	}
	for _, prefix := range slices.Sorted(maps.Keys(want)) {
		if !want[prefix] {
			t.Errorf("ninguna ficha dijo %q", prefix)
		}
	}
	for _, e := range []string{"search:" + view.EmoteExplore, "chat:" + view.EmoteChat, "use:rest:" + view.EmoteSleep} {
		if !emotes[e] {
			t.Errorf("no se vio el símbolo %s (vistos %v)", e, slices.Sorted(maps.Keys(emotes)))
		}
	}
	for k, ok := range story {
		if !ok {
			t.Errorf("la historia no cuenta %q", k)
		}
	}
}
