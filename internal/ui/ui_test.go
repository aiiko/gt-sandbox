package ui

import (
	"image"
	"slices"
	"testing"
)

// El parlante muestra el volumen por la cantidad de ondas, y el silencio
// sin ondas (tachado).
func TestSoundWaves(t *testing.T) {
	for _, c := range []struct {
		s    Sound
		want int
	}{
		{Sound{Volume: 1}, 3}, {Sound{Volume: 0.5}, 2}, {Sound{Volume: 0.2}, 1},
		{Sound{Volume: 0}, 0}, {Sound{Muted: true, Volume: 1}, 0},
	} {
		if got := c.s.Waves(); got != c.want {
			t.Errorf("%+v: %d ondas, quería %d", c.s, got, c.want)
		}
	}
}

func TestWrap(t *testing.T) {
	got := Wrap([]string{"abcdefghijklmn", "", "corto"}, 8)
	want := []string{"abcdefgh", "  ijklmn", "", "corto"}
	if !slices.Equal(got, want) {
		t.Fatalf("%q", got)
	}
}

// La lista se desplaza para que la opción elegida siempre se vea, y cada
// opción visible se puede tocar.
func TestMenuLayoutScrollsToSelection(t *testing.T) {
	items := make([]MenuItem, 12)
	for sel := range items {
		m := Menu{Items: items, Selected: sel}
		lay := m.Layout(480, 270)
		r := lay.Items[sel]
		if r.Empty() {
			t.Fatalf("elegida %d: no se ve", sel)
		}
		if !r.In(lay.Panel) {
			t.Fatalf("elegida %d: fuera del panel", sel)
		}
		if got := lay.Hit(r.Min.Add(image.Pt(2, 2))); got != sel {
			t.Fatalf("elegida %d: Hit dio %d", sel, got)
		}
		shown := 0
		for _, it := range lay.Items {
			if !it.Empty() {
				shown++
			}
		}
		if shown != menuRows {
			t.Fatalf("elegida %d: %d opciones visibles", sel, shown)
		}
	}
	if lay := (Menu{Items: items[:3]}).Layout(480, 270); lay.Panel.Max.Y > 270 || lay.Hit(image.Pt(0, 0)) != -1 {
		t.Fatalf("panel %v", lay.Panel)
	}
}
