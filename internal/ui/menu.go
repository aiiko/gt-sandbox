package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// MenuItem es una opción del menú. Value se muestra a la derecha (el estado
// de un interruptor, por ejemplo «No»).
type MenuItem struct {
	Label string
	Value string
}

// Menu es una pantalla de menú: título, opciones con una elegida, una nota
// sobre la elegida y la barra de controles. Lo arma el juego; Draw solo lo
// dibuja.
type Menu struct {
	Title    string
	Subtitle string
	Items    []MenuItem
	Selected int
	Note     string // descripción de la opción elegida
	Controls []Control
}

// MenuLayout son las zonas interactivas del menú.
type MenuLayout struct {
	Panel image.Rectangle
	Items []image.Rectangle // una por opción; vacía si la opción no se ve
}

// Layout ubica el panel y las opciones en un lienzo de W × H, desplazando la
// lista para que la elegida siempre se vea.
func (m Menu) Layout(W, H int) MenuLayout {
	lay := MenuLayout{Items: make([]image.Rectangle, len(m.Items))}
	first := max(0, min(m.Selected-menuRows+1, len(m.Items)-menuRows))
	visible := min(menuRows, len(m.Items))
	x0 := (W - menuWidth) / 2
	lay.Panel = image.Rect(x0, menuTop, x0+menuWidth, menuTop+visible*menuRow+8)
	for i := first; i < first+visible; i++ {
		y := menuTop + 4 + (i-first)*menuRow
		lay.Items[i] = image.Rect(x0+4, y, x0+menuWidth-4, y+menuRow)
	}
	return lay
}

// Hit devuelve la opción bajo el punto p, o -1.
func (l MenuLayout) Hit(p image.Point) int {
	for i, r := range l.Items {
		if p.In(r) {
			return i
		}
	}
	return -1
}

// Medidas del menú en el lienzo de 480 × 270.
const (
	menuWidth = 300
	menuRow   = 16
	menuTop   = 64
	menuRows  = 7 // opciones visibles a la vez
	noteLines = 3
)

var (
	menuBackground = color.NRGBA{0x2B, 0x25, 0x40, 0xFF}
	menuStripe     = color.NRGBA{0x34, 0x2D, 0x4D, 0xFF}
)

// Draw dibuja el menú ocupando todo dst.
func (m Menu) Draw(dst *ebiten.Image) MenuLayout {
	W, H := dst.Bounds().Dx(), dst.Bounds().Dy()
	dst.Fill(menuBackground)
	for y := 0; y < H; y += 12 {
		vector.FillRect(dst, 0, float32(y), float32(W), 6, menuStripe, false)
	}

	// Título al doble de tamaño, con sombra.
	tw := 2 * TextWidth(m.Title)
	drawTextScaled(dst, m.Title, (W-tw)/2+1, 15, 2, Plum)
	drawTextScaled(dst, m.Title, (W-tw)/2, 14, 2, Ivory)
	if m.Subtitle != "" {
		DrawText(dst, m.Subtitle, (W-TextWidth(m.Subtitle))/2, 42, Turquoise)
	}

	lay := m.Layout(W, H)
	Box(dst, lay.Panel, Ivory)
	for i, r := range lay.Items {
		if r.Empty() {
			continue
		}
		ink := Plum
		if i == m.Selected {
			vector.FillRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), Lilac, false)
			ink = Ivory
			drawCaret(dst, r.Min.X+4, r.Min.Y+5, ink)
		}
		value := m.Items[i].Value
		vw := TextWidth(value)
		DrawText(dst, Fit(m.Items[i].Label, r.Dx()-24-vw), r.Min.X+14, r.Min.Y+1, ink)
		if value != "" {
			DrawText(dst, value, r.Max.X-6-vw, r.Min.Y+1, ink)
		}
	}
	panel := lay.Panel
	// Flechas si hay opciones fuera de la vista.
	if len(m.Items) > 0 && lay.Items[0].Empty() {
		drawArrow(dst, panel.Max.X-10, panel.Min.Y+3, -1)
	}
	if n := len(m.Items); n > 0 && lay.Items[n-1].Empty() {
		drawArrow(dst, panel.Max.X-10, panel.Max.Y-4, 1)
	}

	if m.Note != "" {
		lines := WrapWidth(m.Note, menuWidth+60-12)
		if len(lines) > noteLines {
			lines = lines[:noteLines]
			lines[noteLines-1] = Fit(lines[noteLines-1]+" …", menuWidth+60-12)
		}
		nx := (W - menuWidth - 60) / 2
		ny := panel.Max.Y + 8
		Box(dst, image.Rect(nx, ny, nx+menuWidth+60, ny+len(lines)*LineHeight+6), IvoryDim)
		for i, l := range lines {
			DrawText(dst, l, nx+6, ny+2+i*LineHeight, PlumSoft)
		}
	}
	drawControls(dst, m.Controls)
	return lay
}

// drawTextScaled escribe s agrandado scale veces.
func drawTextScaled(dst *ebiten.Image, s string, x, y int, scale float64, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, face, op)
}

// drawCaret dibuja el triángulo que marca la opción elegida.
func drawCaret(dst *ebiten.Image, x, y int, c color.Color) {
	for i := range 3 {
		vector.FillRect(dst, float32(x+i), float32(y+i), 1, float32(7-2*i), c, false)
	}
}

// drawArrow dibuja una flecha chica hacia arriba (dir -1) o abajo (dir 1).
func drawArrow(dst *ebiten.Image, x, y, dir int) {
	for i := range 3 {
		vector.FillRect(dst, float32(x-2+i), float32(y+dir*i), float32(5-2*i), 1, PlumSoft, false)
	}
}
