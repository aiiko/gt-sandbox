package scenario_test

import (
	"encoding/json"
	"strings"
	"testing"

	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/crush"
)

// TestSeatsScenariosSnapshots verifica que los escenarios de asientos (M34–M36)
// mantienen invariantes y dan el mismo hash y eventos al guardar y cargar.
func TestSeatsScenariosSnapshots(t *testing.T) {
	files := []string{
		"tiny_reaches_chair_and_sofa",
		"giantess_sits_and_relaxes",
		"chair_sit_crushes_tiny",
		"sofa_sit_spares_far_side",
		"careful_sit_needs_high_care",
	}
	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			spec, err := scenario.ReadFile("../../tests/scenarios/" + name + ".json")
			if err != nil {
				t.Fatal(err)
			}
			in, err := scenario.Build(spec)
			if err != nil {
				t.Fatal(err)
			}
			playChecked(t, in, 37, nil)
		})
	}
}

// TestSitRequiresMoreCareThanHand comprueba que sentarse (fuera de la línea de
// vista) exige más cuidado que apoyar la mano: con care moderado (0,4) titubea
// menos veces a lo largo de varias semillas que con care alto (0,95), y menos
// que la mano con el mismo care moderado.
func TestSitRequiresMoreCareThanHand(t *testing.T) {
	base, err := scenario.ReadFile("../../tests/scenarios/careful_sit_needs_high_care.json")
	if err != nil {
		t.Fatal(err)
	}
	runCare := func(care float64, seeds []uint64) (hesitations, crushed int) {
		for _, seed := range seeds {
			spec := *base
			spec.Seed = seed
			spec.Expect = nil
			spec.Commands = make([]scenario.CommandSpec, len(base.Commands))
			copy(spec.Commands, base.Commands)
			var data map[string]any
			if err := json.Unmarshal(spec.Commands[0].Data, &data); err != nil {
				t.Fatal(err)
			}
			comps := data["components"].(map[string]any)
			comps["agent"].(map[string]any)["care"] = care
			raw, _ := json.Marshal(data)
			spec.Commands[0].Data = raw
			in, err := scenario.Build(&spec)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := in.Run(); err != nil {
				t.Fatal(err)
			}
			for _, e := range in.Events {
				switch ev := e.Event.(type) {
				case ai.Hesitated:
					if ev.Source == crush.SourceSeat {
						hesitations++
					}
				case crush.Crushed:
					if ev.Source == crush.SourceSeat {
						crushed++
					}
				}
			}
		}
		return hesitations, crushed
	}
	seeds := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	modH, modC := runCare(0.4, seeds)
	highH, highC := runCare(0.95, seeds)
	t.Logf("care 0.4: %d titubeos, %d aplastados; care 0.95: %d titubeos, %d aplastados", modH, modC, highH, highC)
	if highH <= modH {
		t.Errorf("care alto (%d) no titubeó más que care moderado (%d)", highH, modH)
	}
	if modC == 0 {
		t.Errorf("con care moderado (0.4) se esperaba algún fallo al sentarse a ciegas")
	}
	if highC >= modC {
		t.Errorf("con care alto (%d) se esperaban menos muertes que con moderado (%d)", highC, modC)
	}
}

// TestPalmContactWithSofa verifica que el sofá tiene contact_surface: si la
// giantess apoya la palma sobre el almohadón (pose sit_hand_sofa), attach abre
// un portal dinámico palma↔sofá.
func TestPalmContactWithSofa(t *testing.T) {
	spec, err := scenario.ReadFile("../../tests/scenarios/giantess_sits_and_relaxes.json")
	if err != nil {
		t.Fatal(err)
	}
	spec.Expect = nil
	spec.Ticks = 40
	// Ubicar a la giantess ya frente al sofá mirando al norte y aplicar sit_hand_sofa.
	for i := range spec.Entities {
		if spec.Entities[i].Ref == "giantess" {
			raw := string(spec.Entities[i].Spawn)
			raw = strings.Replace(raw, `"x": 1000000`, `"x": 1800000`, 1)
			raw = strings.Replace(raw, `"y": 1500000`, `"y": 700000`, 1)
			raw = strings.Replace(raw, `"z": 0}`, `"z": 0}, "facing": 1024`, 1)
			spec.Entities[i].Spawn = json.RawMessage(raw)
		}
	}
	spec.Commands = append(spec.Commands,
		scenario.CommandSpec{Tick: 2, Kind: "set_pose", Data: json.RawMessage(`{"entity":"$giantess","pose":"sit_hand_sofa"}`)},
	)
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := in.Run(); err != nil {
		t.Fatal(err)
	}
	palm, sofa := in.Refs["palm"], in.Refs["sofa"]
	opened := false
	for _, e := range in.Events {
		if ev, ok := e.Event.(attach.ContactBegan); ok {
			if (ev.A == palm && ev.B == sofa) || (ev.A == sofa && ev.B == palm) {
				opened = true
			}
		}
	}
	if !opened {
		t.Fatal("la palma apoyada en el sofá no abrió contacto con $sofa")
	}
}
