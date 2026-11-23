package main

import (
	"slices"
	"strings"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/render"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/encounter"
	"gtsandbox/internal/ui"
	"gtsandbox/internal/view"
)

// El panel lateral ofrece las tres pestañas (Crónica, Colonia y Giantess),
// la barra superior muestra la despensa, la ficha trae los botones Seguir e
// Ir allí y la barra inferior muestra 4 controles contextuales con F1 Ayuda.
func TestSidePanelColonyGiantessAndCard(t *testing.T) {
	g := artGame(t, randomBase)
	g.story = true
	f := stepTo(t, g, 120)

	h := g.partida(f, ids.Nil)
	if h.Pantry == nil || h.Pantry.Goal <= 0 {
		t.Fatalf("barra superior sin despensa: %+v", h.Pantry)
	}
	if len(h.Controls) != 4 || h.Controls[3].Key != "F1" {
		t.Fatalf("barra inferior inesperada: %+v", h.Controls)
	}
	if h.Panel == nil || h.Panel.Colony == nil || h.Panel.Giantess == nil {
		t.Fatalf("panel lateral incompleto: %+v", h.Panel)
	}
	if len(h.Panel.Colony.Tinies) == 0 || len(h.Panel.Colony.Activities) == 0 {
		t.Fatalf("pestaña Colonia sin tinies o actividades: %+v", h.Panel.Colony)
	}
	if h.Panel.Colony.Site == nil || h.Panel.Colony.Site.Required <= 0 {
		t.Fatalf("pestaña Colonia sin avance del refugio: %+v", h.Panel.Colony.Site)
	}
	if len(g.panelActivities) != len(h.Panel.Colony.Activities) || len(g.panelTinies) != len(h.Panel.Colony.Tinies) {
		t.Fatalf("índices de clic desalineados: act %d/%d, tinies %d/%d",
			len(g.panelActivities), len(h.Panel.Colony.Activities), len(g.panelTinies), len(h.Panel.Colony.Tinies))
	}
	if h.Panel.Giantess.Face == "" || len(h.Panel.Giantess.Needs) == 0 {
		t.Fatalf("pestaña Giantess incompleta: %+v", h.Panel.Giantess)
	}

	// Elegir un tiny desde la actividad y probar los botones de la ficha.
	tinyID := g.panelActivities[0]
	g.sel = tinyID
	hSel := g.partida(f, ids.Nil)
	if hSel.Card == nil || !hSel.Card.CanFollow || !hSel.Card.CanGo || hSel.Card.Following {
		t.Fatalf("ficha sin botones: %+v", hSel.Card)
	}
	if !slices.ContainsFunc(hSel.Controls, func(c ui.Control) bool { return c.Key == "T" && c.Label == "Seguir" }) {
		t.Errorf("con entidad elegida, la barra debería ofrecer T Seguir: %+v", hSel.Controls)
	}
	g.goToEntity(tinyID)
	if g.art.focus.IsNil() {
		t.Error("Ir allí sobre un tiny en el escritorio debería abrir su vista micro")
	}
}

// El menú de pausa pausa la partida, navega por Opciones, guarda y carga el
// snapshot conservando el hash exacto, y al cerrar devuelve la pausa previa.
func TestPauseMenuOptionsAndSaveLoad(t *testing.T) {
	t.Chdir("../..")
	defer func(old int) { uiScalePref = old }(uiScalePref)
	defer func(old string) { saveDirOverride = old }(saveDirOverride)
	uiScalePref = 0
	dir := t.TempDir()
	saveDirOverride = dir

	g, err := newGame(morningScenario, 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.enableArt(""); err != nil {
		t.Fatal(err)
	}
	g.fromMenu = true
	g.w, g.h = 960, 540
	g.startView()
	for range 60 {
		g.run.StepOnce()
	}
	savedTick := g.run.World.Tick()
	savedHash := g.run.Hash()

	// Abrir pausa detiene la simulación; cerrar restaura wasPaused.
	g.run.Paused = false
	g.openPause()
	if !g.pause.open || !g.run.Paused {
		t.Fatal("openPause debería abrir el menú y pausar")
	}

	// Opciones: cambiar escala de interfaz, movimiento reducido y pausa ante amenazas.
	g.handlePause(menuInput{pick: pItemOptions, hover: -1})
	if g.pause.page != pauseOptions {
		t.Fatalf("página %d, quería pauseOptions", g.pause.page)
	}
	g.handlePause(menuInput{pick: optUIScale, hover: -1})
	if uiScalePref != 1 {
		t.Fatalf("uiScalePref %d, quería 1", uiScalePref)
	}
	if f := render.CanvasFactorFor(1920, 1080, uiScalePref); f != 1 {
		t.Fatalf("CanvasFactorFor(1920, 1080, 1) = %d", f)
	}
	g.handlePause(menuInput{pick: optReducedMotion, hover: -1})
	if !g.art.cut {
		t.Fatal("Movimiento reducido no se activó")
	}
	g.handlePause(menuInput{pick: optPauseOnThreat, hover: -1})
	if !g.pauseOnThreat {
		t.Fatal("Pausa ante amenazas no se activó")
	}
	g.handlePause(menuInput{back: true, pick: -1, hover: -1})
	if g.pause.page != pauseMain {
		t.Fatalf("Esc en Opciones debería volver a pauseMain: %d", g.pause.page)
	}

	// Guardar partida, avanzar 30 ticks y cargar desde la lista.
	g.handlePause(menuInput{pick: pItemSave, hover: -1})
	if !strings.HasPrefix(g.pause.msg, "Guardado:") {
		t.Fatalf("guardar falló: %q", g.pause.msg)
	}
	for range 30 {
		g.run.StepOnce()
	}
	if g.run.World.Tick() == savedTick {
		t.Fatal("la simulación no avanzó tras guardar")
	}
	g.handlePause(menuInput{pick: pItemLoad, hover: -1})
	if g.pause.page != pauseLoad || len(g.pause.saves) != 1 {
		t.Fatalf("lista de partidas: página %d, %d partidas", g.pause.page, len(g.pause.saves))
	}
	g.handlePause(menuInput{pick: 0, hover: -1})
	if g.pause.open {
		t.Fatalf("cargar debería cerrar el menú (msg=%q)", g.pause.msg)
	}
	if g.run.World.Tick() != savedTick || g.run.Hash() != savedHash {
		t.Fatalf("cargado tick %d hash %x, quería tick %d hash %x", g.run.World.Tick(), g.run.Hash(), savedTick, savedHash)
	}
	if len(g.visuals) == 0 {
		t.Fatal("al cargar se perdieron los visuales")
	}
}

// Con «Pausa ante amenazas» activa, una amenaza a la colonia pausa la partida.
func TestPauseOnThreatPausesGame(t *testing.T) {
	const sc = "tests/scenarios/un_descuido_en_el_escritorio.json"
	g := artGame(t, sc)
	g.pauseOnThreat = true
	g.run.Paused = false

	paused := false
	for tick := 1; tick <= 400; tick++ {
		stepTo(t, g, tick)
		g.updateNotices(0)
		if g.run.Paused {
			paused = true
			break
		}
	}
	if !paused {
		t.Fatal("la partida no se pausó al aparecer una amenaza con pauseOnThreat activo")
	}
	if len(g.notices) == 0 || g.notices[0].kind != ui.NoticeThreat {
		t.Fatalf("se esperaba un aviso de amenaza: %+v", g.notices)
	}
}

// Palma ofrecida y refugio terminado generan sus avisos correspondientes.
func TestPalmAndRefugeNotices(t *testing.T) {
	g := artGame(t, randomBase)
	f := g.run.Cur
	g.run.Events = append(g.run.Events,
		view.Published{Seq: g.noticeSeq + 1, Tick: f.Tick, Event: encounter.PalmOffered{Surface: g.artDesk}},
		view.Published{Seq: g.noticeSeq + 2, Tick: f.Tick, Event: colony.RefugeBuilt{Site: g.artDesk}},
	)
	g.updateNotices(0)
	var kinds []ui.NoticeKind
	for _, n := range g.notices {
		kinds = append(kinds, n.kind)
	}
	if !slices.Contains(kinds, ui.NoticePalm) || !slices.Contains(kinds, ui.NoticeRefuge) {
		t.Fatalf("faltan avisos de palma/refugio: %+v", g.notices)
	}
}
