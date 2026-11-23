package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// PauseMenu es el menú modal de pausa (Esc en la partida, propuesta 17.6):
// se dibuja sobre un velo oscuro encima del mundo en pausa.
type PauseMenu struct {
	Title    string
	Subtitle string
	Items    []MenuItem
	Selected int
	Message  string // aviso breve (p. ej. "Partida guardada")
	Controls []Control
}

const (
	pauseRow = 16
	pausePad = 8
)

// Layout ubica el panel modal y las opciones en un lienzo de W × H.
func (p PauseMenu) Layout(W, H int) MenuLayout {
	lay := MenuLayout{Items: make([]image.Rectangle, len(p.Items))}
	w := min(280, W-16)
	headerH := 20
	if p.Subtitle != "" {
		headerH += LineHeight
	}
	msgH := 0
	if p.Message != "" {
		msgH = LineHeight + 6
	}
	maxRows := max(1, (H-footH-headerH-msgH-36)/pauseRow)
	visible := min(maxRows, len(p.Items))
	first := max(0, min(p.Selected-visible+1, len(p.Items)-visible))
	totalH := headerH + visible*pauseRow + pausePad*2 + msgH
	x0 := (W - w) / 2
	y0 := max(topH+4, (H-footH-totalH)/2)
	lay.Panel = image.Rect(x0, y0, x0+w, y0+totalH)
	listTop := y0 + headerH + 4
	for i := first; i < first+visible; i++ {
		y := listTop + (i-first)*pauseRow
		lay.Items[i] = image.Rect(x0+6, y, x0+w-6, y+pauseRow)
	}
	return lay
}

// Draw dibuja el menú de pausa sobre la escena actual.
func (p PauseMenu) Draw(dst *ebiten.Image) MenuLayout {
	W, H := dst.Bounds().Dx(), dst.Bounds().Dy()
	// Velo semitransparente para destacar el menú sobre la escena en pausa.
	vector.FillRect(dst, 0, 0, float32(W), float32(H), color.NRGBA{0x1B, 0x16, 0x2B, 0xB0}, false)

	lay := p.Layout(W, H)
	Box(dst, lay.Panel, Ivory)

	y := lay.Panel.Min.Y + 5
	DrawText(dst, p.Title, (W-TextWidth(p.Title))/2, y, Plum)
	y += LineHeight + 2
	if p.Subtitle != "" {
		DrawText(dst, Fit(p.Subtitle, lay.Panel.Dx()-16), (W-min(TextWidth(p.Subtitle), lay.Panel.Dx()-16))/2, y, PlumSoft)
		y += LineHeight
	}
	vector.FillRect(dst, float32(lay.Panel.Min.X+6), float32(y), float32(lay.Panel.Dx()-12), 1, IvoryDim, false)

	for i, r := range lay.Items {
		if r.Empty() {
			continue
		}
		ink := Plum
		if i == p.Selected {
			vector.FillRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), Lilac, false)
			ink = Ivory
			drawCaret(dst, r.Min.X+4, r.Min.Y+5, ink)
		}
		val := p.Items[i].Value
		vw := TextWidth(val)
		DrawText(dst, Fit(p.Items[i].Label, r.Dx()-22-vw), r.Min.X+13, r.Min.Y+1, ink)
		if val != "" {
			DrawText(dst, val, r.Max.X-5-vw, r.Min.Y+1, ink)
		}
	}
	if len(p.Items) > 0 && lay.Items[0].Empty() {
		drawArrow(dst, lay.Panel.Max.X-10, lay.Panel.Min.Y+22, -1)
	}
	if n := len(p.Items); n > 0 && lay.Items[n-1].Empty() {
		drawArrow(dst, lay.Panel.Max.X-10, lay.Panel.Max.Y-10, 1)
	}
	if p.Message != "" {
		my := lay.Panel.Max.Y - LineHeight - 5
		DrawText(dst, Fit(p.Message, lay.Panel.Dx()-16), lay.Panel.Min.X+8, my, Turquoise)
	}
	if len(p.Controls) > 0 {
		drawControls(dst, p.Controls)
	}
	return lay
}
