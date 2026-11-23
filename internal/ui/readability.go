package ui

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// BadgeActivity es la actividad dominante que muestra un globo de tinies en
// macro (propuesta, 17.7).
type BadgeActivity uint8

const (
	BadgeIdle BadgeActivity = iota
	BadgeEat
	BadgeCarry
	BadgeRest
	BadgeFlee
	BadgeChat
	BadgeSearch
	BadgeWork
)

func drawBadgeActivity(dst *ebiten.Image, x, y int, a BadgeActivity) {
	fx, fy := float32(x), float32(y)
	switch a {
	case BadgeFlee:
		// Exclamación «!»
		vector.FillRect(dst, fx+2, fy, 1, 3, Plum, false)
		vector.FillRect(dst, fx+2, fy+4, 1, 1, Plum, false)
	case BadgeCarry:
		// Paquetito / migaja
		vector.FillRect(dst, fx+1, fy+1, 3, 3, Amber, false)
		vector.FillRect(dst, fx+1, fy, 3, 1, Plum, false)
		vector.FillRect(dst, fx+1, fy+4, 3, 1, Plum, false)
	case BadgeEat:
		// Bocado turquesa
		vector.FillRect(dst, fx+1, fy+1, 3, 3, Turquoise, false)
		vector.FillRect(dst, fx+2, fy+2, 1, 1, Plum, false)
	case BadgeRest:
		// «z» chica
		vector.FillRect(dst, fx+1, fy, 3, 1, PlumSoft, false)
		vector.FillRect(dst, fx+2, fy+1, 1, 1, PlumSoft, false)
		vector.FillRect(dst, fx+1, fy+2, 1, 1, PlumSoft, false)
		vector.FillRect(dst, fx+1, fy+3, 3, 1, PlumSoft, false)
	case BadgeChat:
		// Globito de diálogo
		vector.FillRect(dst, fx, fy, 5, 3, Lilac, false)
		vector.FillRect(dst, fx+1, fy+3, 1, 1, Lilac, false)
	case BadgeSearch:
		// Lupita / signo
		vector.FillRect(dst, fx+1, fy, 2, 2, Plum, false)
		vector.FillRect(dst, fx+3, fy+2, 1, 2, Plum, false)
	case BadgeWork:
		// Bloque de obra
		vector.FillRect(dst, fx, fy+1, 5, 3, Green, false)
		vector.FillRect(dst, fx+2, fy, 1, 1, Plum, false)
	}
}

// Tooltip es la información al pasar el cursor por encima de una entidad o
// un globo de tinies (propuesta, 17.7).
type Tooltip struct {
	Pos   image.Point
	Title string
	Lines []string
}

func (t Tooltip) draw(dst *ebiten.Image, W, H int) {
	if t.Title == "" && len(t.Lines) == 0 {
		return
	}
	w := TextWidth(t.Title)
	for _, l := range t.Lines {
		w = max(w, TextWidth(l))
	}
	w = min(210, w+12)
	rows := len(t.Lines)
	if t.Title != "" {
		rows++
	}
	h := rows*LineHeight + 6
	x := t.Pos.X + 10
	y := t.Pos.Y + 12
	if x+w > W-4 {
		x = max(4, t.Pos.X-w-8)
	}
	if y+h > H-footH-4 {
		y = max(topH+4, t.Pos.Y-h-8)
	}
	r := image.Rect(x, y, x+w, y+h)
	Box(dst, r, Ivory)
	ty := r.Min.Y + 3
	if t.Title != "" {
		DrawText(dst, Fit(t.Title, w-10), r.Min.X+5, ty, Plum)
		ty += LineHeight
	}
	for _, l := range t.Lines {
		DrawText(dst, Fit(l, w-10), r.Min.X+5, ty, PlumSoft)
		ty += LineHeight
	}
}

// MinimapRoom es una habitación proyectada dentro del miniplano (0..1 en X e Y).
type MinimapRoom struct {
	X0, Y0, X1, Y1 float32
	Active         bool
}

// Minimap es el recuadro chico de la casa que se muestra en vista micro
// (propuesta, 17.7): habitaciones, superficie activa y posición de la giantess.
type Minimap struct {
	Rooms       []MinimapRoom
	HasGiantess bool
	GX, GY      float32 // posición normalizada (0..1) de la giantess
	HasFocus    bool
	FX, FY      float32 // posición normalizada (0..1) de la superficie enfocada
}

func (m Minimap) draw(dst *ebiten.Image, r image.Rectangle) {
	if len(m.Rooms) == 0 {
		return
	}
	Box(dst, r, IvoryPanel)
	in := image.Rect(r.Min.X+5, r.Min.Y+5, r.Max.X-5, r.Max.Y-5)
	iw, ih := float32(in.Dx()), float32(in.Dy())
	for _, rm := range m.Rooms {
		x0 := float32(in.Min.X) + rm.X0*iw
		x1 := float32(in.Min.X) + rm.X1*iw
		// +Y de mundo va hacia arriba en pantalla.
		y0 := float32(in.Max.Y) - rm.Y1*ih
		y1 := float32(in.Max.Y) - rm.Y0*ih
		fill := IvoryDim
		if rm.Active {
			fill = Turquoise
		}
		vector.FillRect(dst, x0, y0, max(1, x1-x0), max(1, y1-y0), Plum, false)
		if x1-x0 > 2 && y1-y0 > 2 {
			vector.FillRect(dst, x0+1, y0+1, x1-x0-2, y1-y0-2, fill, false)
		}
	}
	if m.HasFocus {
		fx := float32(in.Min.X) + m.FX*iw
		fy := float32(in.Max.Y) - m.FY*ih
		vector.FillRect(dst, fx-3, fy-3, 6, 6, Plum, false)
		vector.FillRect(dst, fx-2, fy-2, 4, 4, Turquoise, false)
	}
	if m.HasGiantess {
		gx := float32(in.Min.X) + m.GX*iw
		gy := float32(in.Max.Y) - m.GY*ih
		vector.FillRect(dst, gx-2, gy-2, 5, 5, Plum, false)
		vector.FillRect(dst, gx-1, gy-1, 3, 3, Coral, false)
	}
}

// RoomLabel es un rótulo de habitación en macro que evita los paneles de la
// interfaz y se pega al borde visible si la habitación queda tapada (17.7).
type RoomLabel struct {
	Text string
	At   image.Point // punto preferido (borde sur de la habitación)
	Area image.Rectangle
}

// PlaceRoomLabels ubica cada rótulo dentro de safe evitando obstacles.
func PlaceRoomLabels(labels []RoomLabel, safe image.Rectangle, obstacles []image.Rectangle) []image.Rectangle {
	out := make([]image.Rectangle, len(labels))
	for i, l := range labels {
		w := TextWidth(l.Text) + 8
		h := 15
		r := image.Rect(l.At.X-w/2, l.At.Y, l.At.X-w/2+w, l.At.Y+h)
		r = clampRect(r, safe)
		for range 4 {
			hit := false
			for _, ob := range obstacles {
				if ob.Empty() || !r.Overlaps(ob) {
					continue
				}
				hit = true
				// Empujar hacia el lado visible más cercano.
				cands := []image.Rectangle{
					clampRect(image.Rect(ob.Max.X+3, r.Min.Y, ob.Max.X+3+w, r.Min.Y+h), safe),
					clampRect(image.Rect(ob.Min.X-3-w, r.Min.Y, ob.Min.X-3, r.Min.Y+h), safe),
					clampRect(image.Rect(r.Min.X, ob.Min.Y-3-h, r.Min.X+w, ob.Min.Y-3), safe),
					clampRect(image.Rect(r.Min.X, ob.Max.Y+3, r.Min.X+w, ob.Max.Y+3+h), safe),
				}
				best := r
				bestDist := math.MaxInt
				for _, c := range cands {
					if c.Overlaps(ob) {
						continue
					}
					d := absInt(c.Min.X-r.Min.X) + absInt(c.Min.Y-r.Min.Y)
					if d < bestDist {
						bestDist, best = d, c
					}
				}
				r = best
			}
			if !hit {
				break
			}
		}
		out[i] = r
	}
	return out
}

func clampRect(r, bounds image.Rectangle) image.Rectangle {
	if r.Min.X < bounds.Min.X {
		r = r.Add(image.Pt(bounds.Min.X-r.Min.X, 0))
	}
	if r.Max.X > bounds.Max.X {
		r = r.Add(image.Pt(bounds.Max.X-r.Max.X, 0))
	}
	if r.Min.Y < bounds.Min.Y {
		r = r.Add(image.Pt(0, bounds.Min.Y-r.Min.Y))
	}
	if r.Max.Y > bounds.Max.Y {
		r = r.Add(image.Pt(0, bounds.Max.Y-r.Max.Y))
	}
	return r
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// DrawRoomLabels dibuja los rótulos de habitaciones evitando los paneles.
func DrawRoomLabels(dst *ebiten.Image, labels []RoomLabel, safe image.Rectangle, obstacles []image.Rectangle) {
	rects := PlaceRoomLabels(labels, safe, obstacles)
	for i, r := range rects {
		Box(dst, r, Ivory)
		DrawText(dst, labels[i].Text, r.Min.X+4, r.Min.Y+1, Plum)
	}
}

// DrawOutline dibuja un contorno resaltado alrededor de un polígono cerrado
// (resalte de superficie al pasar el cursor por la etiqueta de Tab, 17.7).
func DrawOutline(dst *ebiten.Image, pts []image.Point, c color.Color) {
	if len(pts) < 2 {
		return
	}
	for i := range pts {
		a, b := pts[i], pts[(i+1)%len(pts)]
		vector.StrokeLine(dst, float32(a.X), float32(a.Y), float32(b.X), float32(b.Y), 2, c, false)
	}
}
