package main

import (
	"slices"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sound"
)

// soundRecorder es el backend de las pruebas: guarda lo que se pidió.
type soundRecorder struct {
	played  []string
	stopped int
}

func (r *soundRecorder) Play(id string, _, _ float64) { r.played = append(r.played, id) }
func (r *soundRecorder) StopAll()                     { r.stopped++ }

// withSound conecta al juego un mezclador con los sonidos reales y un
// grabador en lugar del dispositivo.
func withSound(t *testing.T, g *game) *soundRecorder {
	t.Helper()
	lib, err := sound.Load("../../" + audioManifest)
	if err != nil {
		t.Fatal(err)
	}
	r := &soundRecorder{}
	g.sound = sound.NewMixer(lib, r)
	g.observeSound()
	return r
}

// Aceptación de M20: cada acción de la interfaz suena (elegir, seguir con su
// cambio a micro, volver a macro, pausar y reanudar, una entrada nueva en la
// historia) y M silencia hasta volver a pulsarla.
func TestInterfaceSounds(t *testing.T) {
	g := artGame(t, morningScenario)
	r := withSound(t, g)
	frame := func() []string {
		before := len(r.played)
		for range 20 { // pasa el tiempo mínimo entre repeticiones
			g.observeSound()
		}
		return r.played[before:]
	}
	expect := func(what string, want ...string) {
		t.Helper()
		if got := frame(); !slices.Equal(got, want) {
			t.Errorf("%s: sonó %v, quería %v", what, got, want)
		}
	}
	expect("sin cambios")
	g.sel = refOf(t, morningScenario, "fridge")
	expect("elegir la heladera", sound.UISelect)
	g.sel = ids.Nil
	g.startFollowing("Tiny 01")
	expect("seguir a un tiny", sound.UIFollow, sound.UIZoomIn)
	g.toggleArtMicro()
	expect("volver a macro", sound.UIZoomOut)
	g.run.Paused = true
	expect("pausar", sound.UIPause)
	g.run.Paused = false
	expect("reanudar", sound.UIResume)

	g.story = true
	stories := 0
	for range 400 {
		g.run.StepOnce()
		for _, id := range frame() {
			if id == sound.UIStory {
				stories++
			}
		}
	}
	if stories == 0 {
		t.Error("las entradas nuevas de la historia no sonaron")
	}

	g.toggleMute()
	if !g.sound.Muted() || r.stopped != 1 {
		t.Fatal("M no silenció")
	}
	g.run.Paused = true
	expect("en silencio")
	g.toggleMute()
	if got := r.played[len(r.played)-1]; got != sound.UISelect {
		t.Errorf("al volver el sonido sonó %s", got)
	}
	g.run.Paused = false
	expect("sin silencio", sound.UIResume)
}
