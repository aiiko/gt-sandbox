// Package ui dibuja los paneles de texto del visor de debug (HUD, inspector,
// log de eventos) con la fuente de debug de Ebitengine.
package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Medidas de la fuente de debug.
const (
	CharW = 6
	LineH = 16
)

var panelBG = color.NRGBA{0x00, 0x00, 0x00, 0xb0}

// Panel dibuja lines sobre un fondo semitransparente con la esquina superior
// izquierda en (x, y). Devuelve el alto usado.
func Panel(dst *ebiten.Image, x, y int, lines []string) int {
	if len(lines) == 0 {
		return 0
	}
	w := 0
	for _, l := range lines {
		w = max(w, len([]rune(l)))
	}
	h := len(lines)*LineH + 8
	vector.FillRect(dst, float32(x), float32(y), float32(w*CharW+12), float32(h), panelBG, false)
	for i, l := range lines {
		ebitenutil.DebugPrintAt(dst, l, x+6, y+4+i*LineH)
	}
	return h
}

// Wrap corta cada línea en trozos de a lo sumo cols caracteres (las
// continuaciones van con sangría).
func Wrap(lines []string, cols int) []string {
	cols = max(cols, 8)
	var out []string
	for _, l := range lines {
		r := []rune(l)
		out = append(out, string(r[:min(len(r), cols)]))
		for r = r[min(len(r), cols):]; len(r) > 0; r = r[min(len(r), cols-2):] {
			out = append(out, "  "+string(r[:min(len(r), cols-2)]))
		}
	}
	return out
}
