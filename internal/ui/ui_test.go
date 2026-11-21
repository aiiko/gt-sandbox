package ui

import (
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
