package main

import (
	"slices"
	"testing"
)

func noInput() menuInput { return menuInput{pick: -1, hover: -1} }

// La lista ofrece solo los escenarios con la interfaz de partida (los que
// declaran atlas), en orden de archivo; las pruebas del motor sin arte no.
func TestMenuListsPlayableScenarios(t *testing.T) {
	list, err := listScenarios("../../" + scenarioDir)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, s := range list {
		paths = append(paths, s.Path[len("../../"):])
		if s.Name == "" {
			t.Errorf("%s sin nombre", s.Path)
		}
	}
	want := []string{
		"tests/scenarios/body_desk_to_pillow.json",
		"tests/scenarios/body_floor_to_sofa.json",
		"tests/scenarios/tiny_cae_de_la_palma.json",
		"tests/scenarios/un_descuido_cuidadosa.json",
		"tests/scenarios/un_descuido_en_el_escritorio.json",
		"tests/scenarios/una_manana_en_la_casa.json",
	}
	if !slices.Equal(paths, want) {
		t.Fatalf("escenarios %v", paths)
	}
}

// El menú principal se recorre con las flechas (dando la vuelta), alterna el
// modo debug y abre la lista; desde la lista, Esc o «Volver» regresan.
func TestMenuNavigation(t *testing.T) {
	m := menu{scenarios: []scenarioChoice{{Path: "a.json", Name: "A"}, {Path: "b.json", Name: "B"}}}
	press := func(in menuInput) menuAction { return m.handle(in) }
	in := noInput()

	in.up = true
	press(in)
	if m.sel[pageMain] != itemQuit {
		t.Fatalf("arriba desde la primera va a la última: %d", m.sel[pageMain])
	}
	in = noInput()
	in.down = true
	press(in)
	if m.sel[pageMain] != itemNewGame {
		t.Fatalf("abajo desde la última vuelve a la primera: %d", m.sel[pageMain])
	}
	in = noInput()
	in.enter = true
	if a := press(in); a.kind != actNewGame {
		t.Fatalf("Nueva partida: %+v", a)
	}

	in = noInput()
	in.hover = itemDebug
	press(in)
	in = noInput()
	in.right = true
	press(in)
	if !m.debug || m.view().Items[itemDebug].Value != "Sí" {
		t.Fatal("→ enciende el modo debug")
	}
	in = noInput()
	in.pick = itemDebug
	press(in)
	if m.debug {
		t.Fatal("el clic lo vuelve a apagar")
	}

	in = noInput()
	in.pick = itemScenarios
	press(in)
	if m.page != pageScenarios || len(m.view().Items) != 3 {
		t.Fatalf("página %d, %d opciones", m.page, len(m.view().Items))
	}
	in = noInput()
	in.down = true
	press(in)
	in = noInput()
	in.enter = true
	if a := press(in); a.kind != actScenario || a.path != "b.json" {
		t.Fatalf("elegir B: %+v", a)
	}
	in = noInput()
	in.back = true
	press(in)
	if m.page != pageMain || m.sel[pageMain] != itemScenarios {
		t.Fatalf("Esc vuelve al menú principal en Escenarios: %d %d", m.page, m.sel[pageMain])
	}
	in = noInput()
	in.enter = true
	press(in)
	in = noInput()
	in.pick = 2 // Volver
	if a := press(in); a.kind != actNone || m.page != pageMain {
		t.Fatalf("Volver: %+v, página %d", a, m.page)
	}
	in = noInput()
	in.pick = itemQuit
	if a := press(in); a.kind != actQuit {
		t.Fatalf("Salir: %+v", a)
	}
}

// Cada opción del menú tiene su nota, y un error al abrir se muestra en su
// lugar.
func TestMenuNotes(t *testing.T) {
	m := menu{scenarios: []scenarioChoice{{Name: "A", Notes: "nota A"}}}
	for i := range m.count() {
		m.sel[pageMain] = i
		if m.view().Note == "" {
			t.Errorf("opción %d sin nota", i)
		}
	}
	m.page = pageScenarios
	if m.view().Note != "nota A" {
		t.Fatalf("nota %q", m.view().Note)
	}
	m.err = "falta el atlas"
	if m.view().Note != "No se pudo abrir: falta el atlas" {
		t.Fatalf("nota %q", m.view().Note)
	}
}

// Una partida al azar se dibuja en pixel art con la interfaz de partida, y
// la misma semilla da el mismo mundo.
func TestRandomGame(t *testing.T) {
	t.Chdir("../..")
	var hashes []uint64
	for _, seed := range []uint64{3, 3, 4} {
		g, err := newRandomGame(seed)
		if err != nil {
			t.Fatal(err)
		}
		if err := g.enableArt(""); err != nil {
			t.Fatal(err)
		}
		if g.artAtlas == "" || g.artDesk.IsNil() {
			t.Fatalf("seed %d: sin la interfaz de partida (atlas %q)", seed, g.artAtlas)
		}
		for range 40 {
			g.run.StepOnce()
		}
		if g.run.Err != nil {
			t.Fatal(g.run.Err)
		}
		hashes = append(hashes, g.run.Hash())
	}
	if hashes[0] != hashes[1] || hashes[0] == hashes[2] {
		t.Fatalf("hashes %x", hashes)
	}
}

// Abrir una partida desde el menú la deja lista para volver; un escenario
// que no se puede abrir deja el error en el menú.
func TestAppStartAndLeave(t *testing.T) {
	t.Chdir("../..")
	a, err := newApp(true, false)
	if err != nil {
		t.Fatal(err)
	}
	a.start(func() (*game, error) { return newGame(morningScenario, 0, 0, "") })
	if a.g == nil || !a.g.fromMenu || !a.g.story || !a.g.art.debug {
		t.Fatalf("partida %+v", a.g)
	}
	a.g.toMenu = true
	if err := a.Update(); err != nil || a.g != nil {
		t.Fatalf("volver al menú: %v, %v", err, a.g)
	}
	a.start(func() (*game, error) { return newGame(scenarioDir+"/palm_crushes_tiny.json", 0, 0, "") })
	if a.g != nil || a.menu.err == "" {
		t.Fatalf("un escenario sin arte no se abre: %v %q", a.g, a.menu.err)
	}
}
