package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"gtsandbox/internal/render"
)

// deviceScale es la escala de la pantalla (píxeles físicos por píxel lógico
// de la ventana). run la conecta al monitor; en los tests vale 1.
var deviceScale = func() float64 { return 1 }

// uiScalePref es la escala de interfaz preferida (0 = Automático, 1..4).
var uiScalePref int

// canvasLayout da el lienzo que llena una ventana de w × h píxeles lógicos:
// el factor de escena se elige con los píxeles físicos (propuesta, 17.2 y 17.6).
func canvasLayout(w, h int) (int, int) {
	s := deviceScale()
	return render.CanvasSizeFor(int(math.Ceil(float64(w)*s)), int(math.Ceil(float64(h)*s)), uiScalePref)
}

// windowConfig es la configuración local de presentación que recuerda el juego
// entre sesiones (ventana y opciones de presentación). No entra en partidas ni
// en la simulación.
type windowConfig struct {
	Width         int      `json:"width"`
	Height        int      `json:"height"`
	Maximized     bool     `json:"maximized"`
	Fullscreen    bool     `json:"fullscreen"`
	Volume        *float64 `json:"volume,omitempty"`
	UIScale       int      `json:"ui_scale,omitempty"`
	ReducedMotion bool     `json:"reduced_motion,omitempty"`
	PauseOnThreat bool     `json:"pause_on_threat,omitempty"`
}

// windowFile es dónde se guarda la ventana ("" si no hay carpeta de usuario).
func windowFile() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "gt-sandbox", "ventana.json")
}

// loadWindow lee la ventana guardada; ok es false si no hay una válida.
func loadWindow(path string) (windowConfig, bool) {
	var c windowConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return c, false
	}
	if json.Unmarshal(data, &c) != nil || c.Width < render.MinCanvasWidth || c.Height < render.MinCanvasHeight {
		return windowConfig{}, false
	}
	return c, true
}

// saveWindow guarda la ventana, creando la carpeta si hace falta.
func saveWindow(path string, c windowConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// defaultWindow es el tamaño de la ventana restaurada la primera vez: el
// mayor múltiplo del lienzo mínimo que entra en el 90 % del monitor (en
// píxeles lógicos). La ventana además arranca maximizada.
func defaultWindow(monitorW, monitorH int) (int, int) {
	n := max(1, min(monitorW*9/10/render.MinCanvasWidth, monitorH*9/10/render.MinCanvasHeight))
	return render.MinCanvasWidth * n, render.MinCanvasHeight * n
}

// windowTracker recuerda la ventana mientras corre el juego para guardarla
// al salir. Con -size, -shot o -bench no se usa.
type windowTracker struct {
	path string
	cfg  windowConfig
}

// open aplica la ventana guardada (o la de la primera vez) antes de RunGame.
func (t *windowTracker) open() {
	c, ok := loadWindow(t.path)
	if !ok {
		c = windowConfig{Width: 960, Height: 540, Maximized: true}
		if m := ebiten.Monitor(); m != nil {
			if w, h := m.Size(); w > 0 && h > 0 {
				c.Width, c.Height = defaultWindow(w, h)
			}
		}
	}
	t.cfg = c
	uiScalePref = max(0, min(4, c.UIScale))
	ebiten.SetWindowSize(c.Width, c.Height)
	if c.Maximized {
		ebiten.MaximizeWindow()
	}
	ebiten.SetFullscreen(c.Fullscreen)
}

// observe anota el estado actual de la ventana. El tamaño solo se toma con
// la ventana restaurada, para volver a él al salir de maximizado.
func (t *windowTracker) observe() {
	t.cfg.Fullscreen = ebiten.IsFullscreen()
	if t.cfg.Fullscreen {
		return
	}
	t.cfg.Maximized = ebiten.IsWindowMaximized()
	if !t.cfg.Maximized {
		if w, h := ebiten.WindowSize(); w >= render.MinCanvasWidth && h >= render.MinCanvasHeight {
			t.cfg.Width, t.cfg.Height = w, h
		}
	}
}

func (t *windowTracker) save() error { return saveWindow(t.path, t.cfg) }

// window es la ventana que se recuerda en esta sesión (nil: no se recuerda).
var window *windowTracker

// windowKeys atiende las teclas de la ventana: F11 alterna pantalla completa.
// Se llama una sola vez por tick.
func windowKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if window != nil {
		window.observe()
	}
}
