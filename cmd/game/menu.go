package main

import (
	"fmt"
	"image"
	"path/filepath"
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"gtsandbox/internal/render"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sound"
	"gtsandbox/internal/ui"
)

// scenarioDir es la carpeta de escenarios que lista el menú.
const scenarioDir = "tests/scenarios"

type menuPage int

const (
	pageMain menuPage = iota
	pageScenarios
)

// Opciones del menú principal, en orden.
const (
	itemNewGame = iota
	itemScenarios
	itemDebug
	itemQuit
)

// scenarioChoice es un escenario jugable de la lista.
type scenarioChoice struct {
	Path, Name, Notes string
}

// listScenarios devuelve, ordenados por archivo, los escenarios de dir que
// se juegan con la interfaz de partida (los que declaran atlas). Los demás
// son pruebas del motor sin arte.
func listScenarios(dir string) ([]scenarioChoice, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	slices.Sort(files)
	var out []scenarioChoice
	for _, f := range files {
		spec, err := scenario.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		if spec.Atlas == "" {
			continue
		}
		out = append(out, scenarioChoice{Path: filepath.ToSlash(f), Name: spec.Name, Notes: spec.Notes})
	}
	return out, nil
}

type menuActionKind int

const (
	actNone menuActionKind = iota
	actNewGame
	actScenario
	actQuit
)

type menuAction struct {
	kind menuActionKind
	path string // escenario elegido, con actScenario
}

// menuInput es lo que el jugador hizo en un frame. Pick y Hover son índices
// de opción (-1: ninguna).
type menuInput struct {
	up, down, left, right, enter, back bool
	pick, hover                        int
}

// menu es el estado del menú principal: la página, la opción elegida en cada
// una y el modo debug para las partidas que se abran.
type menu struct {
	page      menuPage
	sel       [2]int
	debug     bool
	scenarios []scenarioChoice
	err       string // por qué no se pudo abrir la última partida
}

func (m *menu) count() int {
	if m.page == pageScenarios {
		return len(m.scenarios) + 1 // + Volver
	}
	return itemQuit + 1
}

// handle aplica la entrada y devuelve qué partida abrir, si alguna.
func (m *menu) handle(in menuInput) menuAction {
	n, sel := m.count(), &m.sel[m.page]
	if in.hover >= 0 && in.hover < n {
		*sel = in.hover
	}
	switch {
	case in.up:
		*sel = (*sel + n - 1) % n
	case in.down:
		*sel = (*sel + 1) % n
	case in.back && m.page == pageScenarios:
		m.page, m.err = pageMain, ""
	case (in.left || in.right) && m.page == pageMain && *sel == itemDebug:
		m.debug = !m.debug
	case in.enter || (in.pick >= 0 && in.pick < n):
		if !in.enter {
			*sel = in.pick
		}
		return m.choose()
	}
	return menuAction{}
}

func (m *menu) choose() menuAction {
	sel := m.sel[m.page]
	if m.page == pageScenarios {
		if sel == len(m.scenarios) {
			m.page, m.err = pageMain, ""
			return menuAction{}
		}
		return menuAction{kind: actScenario, path: m.scenarios[sel].Path}
	}
	switch sel {
	case itemNewGame:
		return menuAction{kind: actNewGame}
	case itemScenarios:
		m.page, m.sel[pageScenarios], m.err = pageScenarios, 0, ""
	case itemDebug:
		m.debug = !m.debug
	case itemQuit:
		return menuAction{kind: actQuit}
	}
	return menuAction{}
}

var mainNotes = [...]string{
	itemNewGame:   "Una colonia sorteada y frágil que la giantess puede aplastar. Ella cambia de rutina cada partida y, si ve tinies, se detiene a mirarlos, les deja migajas o les ofrece la palma.",
	itemScenarios: "Historias armadas para ver momentos puntuales de la casa.",
	itemDebug:     "Muestra los datos técnicos sobre el arte: grillas, rutas y estado de la simulación. En la partida, F3 los alterna.",
	itemQuit:      "Cerrar el juego.",
}

// view arma la pantalla del menú.
func (m *menu) view() ui.Menu {
	v := ui.Menu{Title: "gt-sandbox", Selected: m.sel[m.page]}
	if m.page == pageScenarios {
		v.Subtitle = "Escenarios"
		for _, s := range m.scenarios {
			v.Items = append(v.Items, ui.MenuItem{Label: s.Name})
		}
		v.Items = append(v.Items, ui.MenuItem{Label: "Volver"})
		v.Note = "Volver al menú principal."
		if v.Selected < len(m.scenarios) {
			v.Note = m.scenarios[v.Selected].Notes
		}
		v.Controls = ui.Controls("↑↓", "Elegir", "Enter", "Jugar", "Esc", "Volver")
	} else {
		v.Subtitle = "Una casa, una giantess y su colonia de tinies"
		debug := "No"
		if m.debug {
			debug = "Sí"
		}
		v.Items = []ui.MenuItem{{Label: "Nueva partida"}, {Label: "Escenarios"}, {Label: "Modo debug", Value: debug}, {Label: "Salir"}}
		v.Note = mainNotes[v.Selected]
		v.Controls = ui.Controls("↑↓", "Elegir", "Enter", "Aceptar", "←→", "Cambiar")
	}
	if m.err != "" {
		v.Note = "No se pudo abrir: " + m.err
	}
	return v
}

// app es el juego completo: el menú principal y la partida que se abre desde
// él, en el lienzo de pixel art. Esc (sin nada elegido) vuelve al menú.
type app struct {
	menu          menu
	g             *game
	sound         *sound.Mixer
	reducedMotion bool

	layout           ui.MenuLayout
	viewport         render.Viewport
	defaultTransform ebiten.GeoM
	cursor           image.Point
	seed             func() uint64 // semilla de cada partida nueva
}

func newApp(debug, reducedMotion bool) (*app, error) {
	list, err := listScenarios(scenarioDir)
	if err != nil {
		return nil, err
	}
	return &app{
		menu:          menu{debug: debug, scenarios: list},
		reducedMotion: reducedMotion,
		seed:          randomSeed,
	}, nil
}

const appTitle = "gt-sandbox"

// randomSeed elige la semilla de una partida nueva. Es presentación: la
// simulación solo recibe el número.
func randomSeed() uint64 { return uint64(time.Now().UnixNano())%999_999 + 1 }

func (*app) Layout(_, _ int) (int, int) { return render.CanvasWidth, render.CanvasHeight }

func (a *app) Update() error {
	if a.g != nil {
		if a.g.toMenu {
			a.leave()
			return nil
		}
		return a.g.Update()
	}
	before := a.menu.sel
	act := a.menu.handle(a.input())
	if a.sound != nil {
		if a.menu.sel != before || act.kind != actNone {
			a.sound.Play(sound.UISelect, 1, 0)
		}
		a.sound.Advance()
	}
	switch act.kind {
	case actQuit:
		return ebiten.Termination
	case actNewGame:
		seed := a.seed()
		a.start(func() (*game, error) { return newRandomGame(seed) })
	case actScenario:
		a.start(func() (*game, error) { return newGame(act.path, 0, 0, "") })
	}
	return nil
}

// input lee teclado y mouse en coordenadas del lienzo.
func (a *app) input() menuInput {
	k := inpututil.IsKeyJustPressed
	in := menuInput{
		up:    k(ebiten.KeyArrowUp) || k(ebiten.KeyW),
		down:  k(ebiten.KeyArrowDown) || k(ebiten.KeyS),
		left:  k(ebiten.KeyArrowLeft) || k(ebiten.KeyA),
		right: k(ebiten.KeyArrowRight) || k(ebiten.KeyD),
		enter: k(ebiten.KeyEnter) || k(ebiten.KeyNumpadEnter) || k(ebiten.KeySpace),
		back:  k(ebiten.KeyEscape) || k(ebiten.KeyBackspace),
		pick:  -1, hover: -1,
	}
	if _, wy := ebiten.Wheel(); wy > 0 {
		in.up = true
	} else if wy < 0 {
		in.down = true
	}
	x, y := a.defaultTransform.Apply(ebiten.CursorPositionF())
	if p, ok := a.viewport.Point(x, y); ok {
		if p != a.cursor {
			in.hover = a.layout.Hit(p)
		}
		a.cursor = p
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			in.pick = a.layout.Hit(p)
		}
	}
	return in
}

// start abre una partida; si falla, el menú muestra por qué.
func (a *app) start(build func() (*game, error)) {
	g, err := build()
	if err == nil {
		err = g.enableArt("")
	}
	if err != nil {
		a.menu.err = err.Error()
		return
	}
	g.art.cut, g.art.debug = a.reducedMotion, a.menu.debug
	g.story, g.fromMenu, g.sound = true, true, a.sound
	a.g, a.menu.err = g, ""
	ebiten.SetWindowTitle(appTitle + " | " + g.title)
}

// leave cierra la partida y vuelve al menú.
func (a *app) leave() {
	if a.sound != nil {
		a.sound.Hush()
	}
	a.g = nil
	ebiten.SetWindowTitle(appTitle)
}

func (a *app) Draw(screen *ebiten.Image) {
	if a.g != nil {
		a.g.Draw(screen)
		return
	}
	a.layout = a.menu.view().Draw(screen)
}

func (a *app) DrawFinalScreen(screen ebiten.FinalScreen, canvas *ebiten.Image, defaultTransform ebiten.GeoM) {
	if a.g != nil {
		pixelGame{a.g}.DrawFinalScreen(screen, canvas, defaultTransform)
		return
	}
	a.defaultTransform = defaultTransform
	a.viewport = render.PixelViewport(screen.Bounds())
	a.viewport.Blit(screen, canvas)
}

// randomBase es el escenario del que sale cada partida al azar
// (tools/scenarios/partida.py): la mañana peligrosa, con tinies frágiles que
// se asustan, la palma que aplasta y las pisadas de la giantess, más los
// perfiles de arranque de la giantess y la merienda en el living.
const randomBase = "content/partidas/casa.json"

// newRandomGame arma una partida al azar a partir de randomBase.
func newRandomGame(seed uint64) (*game, error) {
	base, err := scenario.ReadFile(randomBase)
	if err != nil {
		return nil, err
	}
	spec, err := scenario.Randomize(base, seed)
	if err != nil {
		return nil, err
	}
	spec.Name = fmt.Sprintf("Partida al azar (semilla %d, giantess %s)", seed, spec.Profile)
	spec.Notes = ""
	return newScenarioGame(spec, randomBase)
}
