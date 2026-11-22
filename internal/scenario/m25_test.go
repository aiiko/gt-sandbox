package scenario_test

import (
	"testing"

	"gtsandbox/internal/scenario"
)

// Los escenarios de aplastamiento dan el mismo log y el mismo hash si se
// guarda y se carga a cada rato: el apoyo en curso, las pisadas, la alarma,
// la huida y los restos son estado.
func TestCrushScenariosSnapshots(t *testing.T) {
	for _, name := range []string{
		"palm_crushes_tiny", "tiny_on_palm_is_safe", "nest_shelters_from_palm", "cup_lands_on_tiny",
		"haul_requeued_after_death", "step_crushes_tiny_on_floor", "tiny_flees_shadow",
		"tiny_drops_crumb_to_flee", "tiny_flees_footsteps",
	} {
		t.Run(name, func(t *testing.T) {
			spec, err := scenario.ReadFile("../../tests/scenarios/" + name + ".json")
			if err != nil {
				t.Fatal(err)
			}
			in, err := scenario.Build(spec)
			if err != nil {
				t.Fatal(err)
			}
			playWithSnapshots(t, in, 3)
			fails, err := in.Check()
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range fails {
				t.Error(f)
			}
		})
	}
}
