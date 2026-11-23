package view

import (
	"strings"
	"testing"

	"gtsandbox/internal/scenario"
)

func TestSeatsStoryAndExpression(t *testing.T) {
	for _, tc := range []struct {
		file string
		want []string
	}{
		{"giantess_sits_and_relaxes.json", []string{"Descansa en el sofá, hasta las", "se sienta"}},
		{"chair_sit_crushes_tiny.json", []string{"quedó bajo la giantess al sentarse en silla"}},
		{"careful_sit_needs_high_care.json", []string{"nota tinies en silla y espera antes de sentarse"}},
	} {
		t.Run(tc.file, func(t *testing.T) {
			sawHappy := false
			run, _ := runScenario(t, tc.file, func(f *Frame, in *scenario.Instance) {
				if tc.file == "giantess_sits_and_relaxes.json" && f.Expression(in.Refs["giantess"]) == FaceHappy {
					sawHappy = true
				}
			})
			var lines []string
			for _, e := range run.Chronicle.Recent() {
				lines = append(lines, e.Text)
			}
			all := strings.Join(lines, "\n")
			for _, w := range tc.want {
				if !strings.Contains(all, w) {
					t.Errorf("falta %q en la historia:\n%s", w, all)
				}
			}
			if tc.file == "giantess_sits_and_relaxes.json" && !sawHappy {
				t.Error("la giantess no se mostró contenta (FaceHappy) al descansar en el sofá")
			}
		})
	}
}
