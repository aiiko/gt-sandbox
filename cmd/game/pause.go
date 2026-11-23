package main

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/world"
	"gtsandbox/internal/sound"
	"gtsandbox/internal/ui"
	"gtsandbox/internal/view"
)

type pausePage int

const (
	pauseMain pausePage = iota
	pauseOptions
	pauseLoad
	pauseControls
)

// Opciones de la página principal del menú de pausa.
const (
	pItemContinue = iota
	pItemSave
	pItemLoad
	pItemOptions
	pItemControls
	pItemExit
)

// Opciones de la subpágina Opciones.
const (
	optVolume = iota
	optUIScale
	optFullscreen
	optReducedMotion
	optPauseOnThreat
	optBack
)

type saveSlot struct {
	Path string
	Name string
}

// pauseState guarda el estado del menú de pausa de una partida (propuesta, 17.6).
type pauseState struct {
	open   bool
	page   pausePage
	sel    [4]int
	msg    string
	saves  []saveSlot
	layout ui.MenuLayout
	cursor image.Point
}

// saveMeta guarda los metadatos de presentación junto a un snapshot guardado
// desde el menú de pausa. El archivo .json principal sigue siendo un snapshot
// puro compatible con -load y simrun -load.
type saveMeta struct {
	Title   string            `json:"title,omitempty"`
	Atlas   string            `json:"atlas,omitempty"`
	Desk    ids.ID            `json:"desk,omitempty"`
	Visuals map[ids.ID]string `json:"visuals,omitempty"`
}

// saveDirOverride permite redirigir la carpeta de partidas en pruebas.
var saveDirOverride string

func savesDir() string {
	if saveDirOverride != "" {
		return saveDirOverride
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "gt-sandbox", "partidas")
}

func listSaves(dir string) []saveSlot {
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []saveSlot
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".meta.json") {
			continue
		}
		out = append(out, saveSlot{
			Path: filepath.Join(dir, name),
			Name: strings.TrimSuffix(name, ".json"),
		})
	}
	slices.SortFunc(out, func(a, b saveSlot) int { return strings.Compare(b.Name, a.Name) })
	return out
}

func (g *game) openPause() {
	g.pause.open = true
	g.pause.page = pauseMain
	g.pause.sel[pauseMain] = pItemContinue
	g.pause.msg = ""
	g.wasPaused = g.run.Paused
	g.run.Paused = true
}

func (g *game) closePause() {
	g.pause.open = false
	g.pause.msg = ""
	g.run.Paused = g.wasPaused
}

func (g *game) saveSnapshot(dir string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("no hay carpeta de usuario disponible")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	snap, err := g.run.World.Save()
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("partida-%s.json", time.Now().Format("20060102-150405"))
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	if err := world.WriteSnapshot(f, snap); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	meta := saveMeta{Title: g.title, Atlas: g.artAtlas, Desk: g.artDesk, Visuals: g.visuals}
	if data, err := json.MarshalIndent(meta, "", "  "); err == nil {
		_ = os.WriteFile(strings.TrimSuffix(path, ".json")+".meta.json", append(data, '\n'), 0o644)
	}
	return name, nil
}

func (g *game) loadSnapshot(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	w, err := engine.Load(f)
	f.Close()
	if err != nil {
		return err
	}
	g.run = view.NewRunner(w, func() error { w.Step(); return w.CheckInvariants() })
	g.status = func() string { return "" }
	g.sel = ids.Nil
	g.follow = ids.Nil
	g.notices = nil
	g.noticeSeq = 0
	g.threatCooldown = nil
	g.panelScroll = 0

	if data, err := os.ReadFile(strings.TrimSuffix(path, ".json") + ".meta.json"); err == nil {
		var meta saveMeta
		if json.Unmarshal(data, &meta) == nil {
			if meta.Title != "" {
				g.title = meta.Title
			}
			if len(meta.Visuals) > 0 {
				g.visuals = meta.Visuals
			}
			if !meta.Desk.IsNil() {
				g.artDesk = meta.Desk
			}
			if meta.Atlas != "" && meta.Atlas != g.artAtlas {
				g.artAtlas = meta.Atlas
				_ = g.enableArt(meta.Atlas)
			}
		}
	}
	if g.art != nil {
		g.art.focus = ids.Nil
		g.fitArt()
	}
	g.pause.open = false
	g.pause.msg = ""
	g.run.Paused = false
	return nil
}

func (g *game) pauseCount() int {
	switch g.pause.page {
	case pauseOptions:
		return optBack + 1
	case pauseLoad:
		return len(g.pause.saves) + 1 // + Volver
	case pauseControls:
		return 1
	}
	return pItemExit + 1
}

// handlePause aplica la entrada al menú de pausa (teclado y mouse).
func (g *game) handlePause(in menuInput) {
	if g.pause.page == pauseControls {
		if in.back || in.enter || in.pick >= 0 {
			g.pause.page = pauseMain
		}
		return
	}
	n := g.pauseCount()
	sel := &g.pause.sel[g.pause.page]
	if *sel >= n {
		*sel = max(0, n-1)
	}
	if in.hover >= 0 && in.hover < n {
		*sel = in.hover
	}
	switch {
	case in.up:
		*sel = (*sel + n - 1) % n
	case in.down:
		*sel = (*sel + 1) % n
	case in.back:
		if g.pause.page == pauseMain {
			g.closePause()
		} else {
			g.pause.page, g.pause.msg = pauseMain, ""
		}
	case (in.left || in.right) && g.pause.page == pauseOptions:
		dir := 1
		if in.left {
			dir = -1
		}
		g.adjustOption(*sel, dir)
	case in.enter || (in.pick >= 0 && in.pick < n):
		if !in.enter {
			*sel = in.pick
		}
		g.choosePause(*sel)
	}
}

func (g *game) choosePause(sel int) {
	switch g.pause.page {
	case pauseMain:
		switch sel {
		case pItemContinue:
			g.closePause()
		case pItemSave:
			name, err := g.saveSnapshot(savesDir())
			if err != nil {
				g.pause.msg = "Error al guardar: " + err.Error()
			} else {
				g.pause.msg = "Guardado: " + name
			}
		case pItemLoad:
			g.pause.saves = listSaves(savesDir())
			g.pause.page = pauseLoad
			g.pause.sel[pauseLoad] = 0
			g.pause.msg = ""
			if len(g.pause.saves) == 0 {
				g.pause.msg = "No hay partidas guardadas."
			}
		case pItemOptions:
			g.pause.page = pauseOptions
			g.pause.sel[pauseOptions] = 0
			g.pause.msg = ""
		case pItemControls:
			g.pause.page = pauseControls
			g.pause.msg = ""
		case pItemExit:
			if g.fromMenu {
				g.toMenu = true
			} else {
				g.quitGame = true
			}
		}
	case pauseOptions:
		if sel == optBack {
			g.pause.page, g.pause.msg = pauseMain, ""
			return
		}
		g.adjustOption(sel, 1)
	case pauseLoad:
		if sel >= len(g.pause.saves) {
			g.pause.page, g.pause.msg = pauseMain, ""
			return
		}
		if err := g.loadSnapshot(g.pause.saves[sel].Path); err != nil {
			g.pause.msg = "Error al cargar: " + err.Error()
		}
	}
}

var volumeSteps = []float64{0, 0.25, 0.5, 0.75, 1}

func (g *game) adjustOption(opt, dir int) {
	switch opt {
	case optVolume:
		if g.sound != nil {
			cur := g.sound.Master
			idx := 0
			for i, v := range volumeSteps {
				if cur >= v-0.05 {
					idx = i
				}
			}
			idx = (idx + dir + len(volumeSteps)) % len(volumeSteps)
			g.sound.Master = volumeSteps[idx]
			if g.sound.Master > 0 && g.sound.Muted() {
				g.sound.SetMuted(false)
			}
			g.sound.Play(sound.UISelect, 1, 0)
			g.syncPrefs()
		}
	case optUIScale:
		uiScalePref = (uiScalePref + dir + 5) % 5
		g.syncPrefs()
	case optFullscreen:
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
		g.syncPrefs()
	case optReducedMotion:
		if g.art != nil {
			g.art.cut = !g.art.cut
		}
		g.syncPrefs()
	case optPauseOnThreat:
		g.pauseOnThreat = !g.pauseOnThreat
		g.syncPrefs()
	}
}

// syncPrefs vuelca las opciones de presentación actuales a la configuración
// de ventana para que se guarden al salir.
func (g *game) syncPrefs() {
	if window == nil {
		return
	}
	window.cfg.UIScale = uiScalePref
	window.cfg.PauseOnThreat = g.pauseOnThreat
	window.cfg.PanelWidth = g.panelW
	window.cfg.PanelHeight = g.panelH
	if g.art != nil {
		window.cfg.ReducedMotion = g.art.cut
	}
	if g.sound != nil {
		v := g.sound.Master
		window.cfg.Volume = &v
	}
}

func (g *game) applySavedPrefs() {
	if window == nil {
		return
	}
	c := window.cfg
	uiScalePref = max(0, min(4, c.UIScale))
	g.pauseOnThreat = c.PauseOnThreat
	g.panelW, g.panelH = c.PanelWidth, c.PanelHeight
	if g.art != nil && c.ReducedMotion {
		g.art.cut = true
	}
	if g.sound != nil && c.Volume != nil {
		g.sound.Master = max(0, min(1, *c.Volume))
	}
}

func boolLabel(b bool) string {
	if b {
		return "Sí"
	}
	return "No"
}

func uiScaleLabel(pref int) string {
	if pref <= 0 {
		return "Automático"
	}
	return fmt.Sprintf("%d×", pref)
}

func (g *game) pauseView() ui.PauseMenu {
	m := ui.PauseMenu{
		Title:    "Pausa",
		Selected: g.pause.sel[g.pause.page],
		Message:  g.pause.msg,
	}
	switch g.pause.page {
	case pauseMain:
		exitLabel := "Volver al menú principal"
		if !g.fromMenu {
			exitLabel = "Salir"
		}
		m.Items = []ui.MenuItem{
			{Label: "Continuar"},
			{Label: "Guardar partida"},
			{Label: "Cargar partida"},
			{Label: "Opciones"},
			{Label: "Controles"},
			{Label: exitLabel},
		}
		m.Controls = ui.Controls("↑↓", "Elegir", "Enter", "Aceptar", "Esc", "Continuar")
	case pauseOptions:
		m.Subtitle = "Opciones"
		vol := "Sin sonido"
		if g.sound != nil {
			if g.sound.Muted() {
				vol = "Silencio"
			} else {
				vol = fmt.Sprintf("%d%%", int(g.sound.Master*100+0.5))
			}
		}
		cut := false
		if g.art != nil {
			cut = g.art.cut
		}
		m.Items = []ui.MenuItem{
			{Label: "Volumen", Value: vol},
			{Label: "Escala de interfaz", Value: uiScaleLabel(uiScalePref)},
			{Label: "Pantalla completa", Value: boolLabel(ebiten.IsFullscreen())},
			{Label: "Movimiento reducido", Value: boolLabel(cut)},
			{Label: "Pausa ante amenazas", Value: boolLabel(g.pauseOnThreat)},
			{Label: "Volver"},
		}
		m.Controls = ui.Controls("↑↓", "Elegir", "←→", "Cambiar", "Esc", "Volver")
	case pauseLoad:
		m.Subtitle = "Cargar partida"
		for _, s := range g.pause.saves {
			m.Items = append(m.Items, ui.MenuItem{Label: s.Name})
		}
		m.Items = append(m.Items, ui.MenuItem{Label: "Volver"})
		m.Controls = ui.Controls("↑↓", "Elegir", "Enter", "Cargar", "Esc", "Volver")
	}
	return m
}

func (g *game) pauseInput() menuInput {
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
	if g.art != nil {
		x, y := g.art.defaultTransform.Apply(ebiten.CursorPositionF())
		if p, ok := g.art.viewport.Point(x, y); ok {
			if p != g.pause.cursor {
				in.hover = g.pause.layout.Hit(p)
			}
			g.pause.cursor = p
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				in.pick = g.pause.layout.Hit(p)
				if g.pause.page == pauseControls {
					in.pick = 0
				}
			}
		}
	}
	return in
}
