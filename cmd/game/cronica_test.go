package main

import (
	"strings"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/encounter"
	"gtsandbox/internal/view"
)

// Criterio de aceptación de la crónica (propuesta, 17.5): en los primeros 5
// minutos de la partida base con su semilla, la crónica no cuenta
// exploraciones, subidas ni acarreos, pero sí cada aplastamiento, susto y
// encuentro. El diario de un tiny que exploró guarda sus exploraciones, y
// los restos heredan el diario de quien murió.
func TestChronicleAcceptance(t *testing.T) {
	g, err := newGame("../../"+randomBase, 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	end := g.run.World.Tick() + 5*60*20
	crushed, scared, met := 0, 0, 0
	explorer, remains := ids.Nil, ids.Nil
	for g.run.World.Tick() < end {
		g.run.StepOnce()
		if g.run.Err != nil {
			t.Fatal(g.run.Err)
		}
		for _, e := range g.run.World.Events() {
			switch ev := e.Event.(type) {
			case crush.Crushed:
				crushed++
				if len(g.run.Chronicle.Diary(ev.Remains)) > 1 {
					remains = ev.Remains
				}
			case crush.NearMiss:
				if ev.Entity.Gen() > 0 && len(g.run.Chronicle.Diary(ev.Entity)) > 0 {
					scared++
				}
			case encounter.Observed, encounter.Shared, encounter.PalmWithdrawn:
				met++
			case encounter.Skipped:
				// «gone»: los tinies se fueron antes; no hubo encuentro
				if ev.Reason != "gone" {
					met++
				}
			case ai.Explored:
				explorer = ev.Entity
			}
		}
	}
	lines := g.run.Chronicle.Lines()
	gotCrushed, gotScared, gotMet := 0, 0, 0
	for _, e := range lines {
		for _, w := range []string{"explora", "sube a la mano", "asigna un acarreo"} {
			if strings.Contains(e.Text, w) {
				t.Errorf("la crónica cuenta un detalle: %q", e.Text)
			}
		}
		switch {
		case strings.Contains(e.Text, "quedó bajo"):
			gotCrushed++
		case strings.Contains(e.Text, "por poco"):
			gotScared += max(e.Count, 1)
		case e.Level == view.LevelMilestone && (strings.Contains(e.Text, "mirando") || strings.Contains(e.Text, "migajas para") ||
			strings.Contains(e.Text, "palma") || strings.Contains(e.Text, "no la apoya") || strings.Contains(e.Text, "ya hay migajas")):
			gotMet++
		}
	}
	if crushed == 0 || gotCrushed != crushed {
		t.Errorf("aplastamientos: %d en la crónica, %d eventos", gotCrushed, crushed)
	}
	if scared == 0 || gotScared != scared {
		t.Errorf("sustos: %d en la crónica, %d eventos", gotScared, scared)
	}
	if met == 0 || gotMet != met {
		t.Errorf("encuentros: %d en la crónica, %d eventos", gotMet, met)
	}
	routine := false
	for _, e := range lines {
		routine = routine || e.Level == view.LevelRoutine && strings.Contains(e.Text, "hasta las")
	}
	if !routine {
		t.Error("la crónica no tiene ninguna actividad completa de la giantess")
	}

	explored := false
	for _, e := range g.run.Chronicle.Diary(explorer) {
		explored = explored || strings.Contains(e.Text, "explora")
	}
	if explorer.IsNil() || !explored {
		t.Errorf("el diario de %s no guarda sus exploraciones", g.name(explorer))
	}
	d := g.run.Chronicle.Diary(remains)
	if len(d) < 2 || !strings.Contains(d[len(d)-1].Text, "quedó bajo") {
		t.Errorf("los restos no heredaron el diario: %+v", d)
	}
	// la ficha muestra el diario
	g.sel = explorer
	if c := g.card(g.run.Cur, explorer); c == nil || len(c.Diary) == 0 {
		t.Errorf("la ficha no muestra el diario: %+v", c)
	}
}
