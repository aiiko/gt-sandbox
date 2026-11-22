package view

import (
	"strings"
	"testing"

	"gtsandbox/internal/scenario"
)

// runScenario corre un escenario con un Runner y devuelve el Runner al final,
// con lo que vio cada frame en seen.
func runScenario(t *testing.T, file string, seen func(f *Frame, in *scenario.Instance)) (*Runner, *scenario.Instance) {
	t.Helper()
	spec, err := scenario.ReadFile("../../tests/scenarios/" + file)
	if err != nil {
		t.Fatal(err)
	}
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	p, err := in.Player()
	if err != nil {
		t.Fatal(err)
	}
	r := NewRunner(in.World, func() error {
		if p.Done() {
			return errEnd
		}
		return p.Step()
	})
	for !p.Done() {
		r.StepOnce()
		if seen != nil {
			seen(r.Cur, in)
		}
	}
	return r, in
}

func storyHas(r *Runner, part string) bool {
	for _, s := range r.Story {
		if strings.Contains(s.Text, part) {
			return true
		}
	}
	return false
}

// Criterio 15.3: la historia distingue las migajas de una merienda de las que
// la giantess deja a propósito para los tinies, y cuenta los desenlaces de la
// palma ofrecida solo cuando ocurren.
func TestEncounterStory(t *testing.T) {
	r, _ := runScenario(t, "giantess_snacks_in_living.json", nil)
	if !storyHas(r, "Caen migajas de la merienda.") || storyHas(r, "para los tinies") {
		t.Errorf("merienda: %+v", r.Story)
	}
	r, _ = runScenario(t, "giantess_shares_crumbs.json", nil)
	if !storyHas(r, "deja unas migajas para los tinies de escritorio") || storyHas(r, "Caen migajas") {
		t.Errorf("compartir: %+v", r.Story)
	}
	r, _ = runScenario(t, "tiny_boards_offered_palm.json", nil)
	if !storyHas(r, "ofrece la palma abierta") || !storyHas(r, "Subieron 3 tinies a la palma") {
		t.Errorf("alguien sube: %+v", r.Story)
	}
	r, _ = runScenario(t, "nobody_boards_offered_palm.json", nil)
	if !storyHas(r, "Nadie sube a la palma") || storyHas(r, "Subieron") {
		t.Errorf("nadie sube: %+v", r.Story)
	}
	r, _ = runScenario(t, "offer_canceled_when_occupied.json", nil)
	if !storyHas(r, "no la apoya") || storyHas(r, "ofrece la palma abierta") {
		t.Errorf("cancelada: %+v", r.Story)
	}
}

// Mientras observa, la giantess muestra el ojo y el retrato curioso; al ir
// hacia el escritorio, ninguno de los dos.
func TestObservePresentation(t *testing.T) {
	look, curious, goingPlain := false, false, false
	runScenario(t, "giantess_observes_desk.json", func(f *Frame, in *scenario.Instance) {
		g := in.Refs["giantess"]
		s, _ := f.SpaceByID(g)
		switch {
		case s.Activity == "observe" && s.Step == "goto":
			goingPlain = goingPlain || f.Emote(g) == "" && f.Expression(g) != FaceCurious
		case s.Activity == "observe":
			look = look || f.Emote(g) == EmoteLook
			curious = curious || f.Expression(g) == FaceCurious
		}
	})
	if !look || !curious || !goingPlain {
		t.Fatalf("ojo %v, curiosa %v, yendo sin símbolo %v", look, curious, goingPlain)
	}
}
