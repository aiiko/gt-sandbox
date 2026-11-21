package main

import (
	"maps"
	"testing"

	"gtsandbox/internal/sound"
)

// El ambiente sigue a lo que se mira: la habitación de la giantess en macro,
// el tono de sala con la habitación detrás en micro, la TV cuando la giantess
// la mira desde el living, y la pausa lo baja.
func TestAmbienceFollowsView(t *testing.T) {
	g := artGame(t, morningScenario)
	withSound(t, g)
	targets := func() map[string]float64 {
		return sound.AmbientTargets(g.run.Cur, g.soundView(), g.follow, g.run.Paused)
	}
	check := func(what string, want map[string]float64) {
		t.Helper()
		if got := targets(); !maps.Equal(got, want) {
			t.Errorf("%s: %v, quería %v", what, got, want)
		}
	}
	check("macro, la giantess en el dormitorio", map[string]float64{sound.AmbBedroom: 1})
	g.run.Paused = true
	check("en pausa", map[string]float64{sound.AmbBedroom: 0.4})
	g.run.Paused = false
	stepTo(t, g, 70)
	check("macro, la giantess en la cocina", map[string]float64{sound.AmbKitchen: 1})
	g.sel = g.artDesk
	g.toggleArtMicro()
	check("micro del escritorio", map[string]float64{sound.AmbMicro: 1, sound.AmbBedroom: 0.45})
	g.toggleArtMicro()
	stepTo(t, g, 620)
	check("macro, la giantess mira la TV", map[string]float64{sound.AmbLiving: 1, sound.AmbTV: 0.8})

	// los fundidos: tras FadeFrames frames la cocina quedó en silencio y el
	// living a pleno, sin saltos
	for range sound.FadeFrames {
		g.observeSound()
	}
	vol := sound.DefaultVolumes[sound.Ambient]
	if g.sound.LoopVolume(sound.AmbLiving) != vol || g.sound.LoopVolume(sound.AmbKitchen) != 0 {
		t.Errorf("living %.3f, cocina %.3f", g.sound.LoopVolume(sound.AmbLiving), g.sound.LoopVolume(sound.AmbKitchen))
	}
}
