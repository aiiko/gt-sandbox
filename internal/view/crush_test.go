package view

import (
	"strings"
	"testing"

	"gtsandbox/internal/scenario"
)

// En el escenario de cierre, al bajar la mano: el frame muestra los restos
// con su causa y el espíritu, los que huyen llevan la gota, la historia cuenta
// la muerte con causa y lugar, y la giantess deja ver dónde pisará al caminar.
func TestCrushPresentation(t *testing.T) {
	spec, err := scenario.ReadFile("../../tests/scenarios/un_descuido_en_el_escritorio.json")
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
	sawFootprint, sawFlee := false, false
	for in.World.Tick()-in.Start < 200 {
		r.StepOnce()
		f := r.Cur
		for _, fp := range f.Footprints {
			if fp.Walker == in.Refs["giantess"] && fp.Length > 0 && fp.Dir.LengthSq() > 0 {
				sawFootprint = true
			}
		}
		for _, th := range f.Things {
			if th.Activity == "flee" && f.Emote(th.ID) == EmoteSweat && th.Alarmed {
				sawFlee = true
			}
		}
	}
	if !sawFootprint {
		t.Error("el frame nunca mostró la próxima pisada de la giantess")
	}
	if !sawFlee {
		t.Error("ningún tiny huyó con la gota")
	}
	f := r.Cur
	remains := 0
	for _, th := range f.Things {
		if th.Remains != "" {
			remains++
			if th.Remains != "hand" || !f.Fresh(th.ID) || f.Emote(th.ID) != EmoteSpirit {
				t.Errorf("restos: %+v, frescos %v, símbolo %q", th, f.Fresh(th.ID), f.Emote(th.ID))
			}
			if in.Visuals[th.ID] != scenario.RemainsVisual {
				t.Errorf("los restos tienen visual %q", in.Visuals[th.ID])
			}
		}
	}
	if remains == 0 {
		t.Fatal("no hay restos en el frame")
	}
	told := false
	for _, s := range r.Story {
		told = told || strings.Contains(s.Text, "quedó bajo la palma en escritorio")
	}
	if !told {
		t.Errorf("la historia no cuenta la muerte: %+v", r.Story)
	}
}
