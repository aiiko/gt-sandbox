package main

import (
	"encoding/json"
	"maps"
	"os"
	"testing"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sound"
)

// Aceptación de M22: la mezcla de referencia recorre la mañana completa con
// el recorrido de cámara registrado y su resumen (cada pedido y los niveles
// del ambiente por segundo) coincide con el guardado. -update lo reescribe.
// También verifica que el ambiente suene sin cortes: nunca queda en silencio
// (salvo el primer fundido) y ningún bucle salta más de un paso de fundido.
func TestReferenceMix(t *testing.T) {
	path := "../../" + mixManifest
	spec, err := readMix(path)
	if err != nil {
		t.Fatal(err)
	}
	sc, err := scenario.ReadFile("../../" + spec.Scenario)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Seed != spec.Seed {
		t.Fatalf("la mezcla dice semilla %d y el escenario %d", spec.Seed, sc.Seed)
	}
	lib, err := sound.Load("../../" + audioManifest)
	if err != nil {
		t.Fatal(err)
	}
	g := artGame(t, spec.Scenario)
	out, err := runMix(g, spec, lib, false)
	if err != nil {
		t.Fatal(err)
	}
	if *updateCaptures {
		spec.Frames, spec.Digest, spec.Counts = out.frame, out.digest(), out.counts
		data, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if out.frame != spec.Frames || out.digest() != spec.Digest || !maps.Equal(out.counts, spec.Counts) {
		t.Errorf("la mezcla cambió: %d frames, resumen %s (registrado %d, %s); si fue a propósito, -update", out.frame, out.digest(), spec.Frames, spec.Digest)
	}
	for _, id := range []string{sound.HandNear, sound.HandLand, sound.HandLift, sound.TinyHop, sound.Emote + "discover", sound.DoorFusuma, sound.TVOn} {
		if out.counts[id] == 0 {
			t.Errorf("la mezcla no tiene %s", id)
		}
	}
	// el ambiente: continuo, con fundidos al cambiar de habitación o de escala
	if out.silences > 0 {
		t.Errorf("el ambiente quedó en silencio %d frames", out.silences)
	}
	if step := sound.DefaultVolumes[sound.Ambient] / sound.FadeFrames; out.maxJump > step+1e-9 {
		t.Errorf("un bucle saltó %.4f en un frame (fundido: %.4f)", out.maxJump, step)
	}
}
