package ui

import (
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gtsandbox/internal/render"
)

// Estilo de la interfaz de partida (propuesta, «Interfaz y alcance de arte»):
// paneles marfil con esquinas redondeadas y borde ciruela, texto ciruela en
// una fuente pixel de 12 px con tildes y eñes.
var (
	Ivory     = color.NRGBA{0xFF, 0xF1, 0xD8, 0xFF}
	IvoryDim  = color.NRGBA{0xF2, 0xE0, 0xC4, 0xFF}
	Plum      = color.NRGBA{0x3F, 0x36, 0x5C, 0xFF}
	PlumSoft  = color.NRGBA{0x79, 0x69, 0xA4, 0xFF}
	Turquoise = color.NRGBA{0x76, 0xD8, 0xCB, 0xFF}
	Coral     = color.NRGBA{0xFF, 0x9C, 0xAD, 0xFF}
	Lilac     = color.NRGBA{0x9A, 0x8B, 0xE3, 0xFF}
)

// LineHeight es el alto de una línea de texto de la interfaz de partida.
const LineHeight = 13

var face = text.NewGoXFace(bitmapfont.Face)

// TextWidth mide un texto en píxeles.
func TextWidth(s string) int { return int(text.Advance(s, face) + 0.5) }

// DrawText escribe s con la esquina superior izquierda en (x, y).
func DrawText(dst *ebiten.Image, s string, x, y int, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, face, op)
}

// Fit recorta s con «…» para que mida a lo sumo w píxeles.
func Fit(s string, w int) string {
	if TextWidth(s) <= w {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && TextWidth(string(r)+"…") > w {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

// WrapWidth corta s en líneas de a lo sumo w píxeles, por palabras.
func WrapWidth(s string, w int) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(s) {
		next := word
		if line != "" {
			next = line + " " + word
		}
		if line != "" && TextWidth(next) > w {
			lines = append(lines, line)
			next = word
		}
		line = next
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// Box dibuja un panel marfil con esquinas redondeadas de 2 px y borde ciruela.
func Box(dst *ebiten.Image, r image.Rectangle, fill color.Color) {
	x, y, w, h := float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy())
	vector.FillRect(dst, x+2, y, w-4, h, Plum, false)
	vector.FillRect(dst, x, y+2, w, h-4, Plum, false)
	vector.FillRect(dst, x+1, y+1, w-2, h-2, Plum, false)
	vector.FillRect(dst, x+2, y+1, w-4, h-2, fill, false)
	vector.FillRect(dst, x+1, y+2, w-2, h-4, fill, false)
	// sombra de un píxel abajo para despegar el panel del escenario
	vector.FillRect(dst, x+2, y+h, w-4, 1, color.NRGBA{0x3F, 0x36, 0x5C, 0x60}, false)
}

// Control es un atajo de teclado de la barra inferior.
type Control struct{ Key, Label string }

// Controls arma la barra de controles a partir de pares tecla, rótulo.
func Controls(pairs ...string) []Control {
	out := make([]Control, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, Control{Key: pairs[i], Label: pairs[i+1]})
	}
	return out
}

// Need es una barra de necesidad del inspector resumido (Value en [0, 1000]).
type Need struct {
	Label string
	Value int32
}

// Card es el inspector resumido de la entidad elegida.
type Card struct {
	Name     string
	Place    string
	Activity string
	Carrying string
	Needs    []Need
	Order    string // orden disponible ("Clic der. mover aquí")
	Emote    string // símbolo actual, para repetirlo junto al nombre
}

// StoryLine es una entrada del panel de historia.
type StoryLine struct {
	Time      string
	Text      string
	Focusable bool // se puede enfocar al actor o al lugar
}

// HUD es todo lo que muestra la interfaz de partida en un frame. Lo arma el
// juego a partir del Frame; Draw solo lo dibuja.
type HUD struct {
	Clock     string
	Paused    bool
	Speed     int // índice en view.Speeds, 0..2
	Location  []string
	Following string
	Card      *Card
	Story     []StoryLine
	Face      string // expresión del retrato; "" oculta el panel de historia
	Controls  []Control
	Sound     *Sound // nil: sin sonido (capturas, sin dispositivo)
}

// Sound es el estado del sonido que muestra la barra superior: un parlante
// con ondas según el volumen, o tachado en silencio, y la tecla que lo
// alterna. Se lee por forma, sin color.
type Sound struct {
	Muted  bool
	Volume float64 // 0..1
}

// SoundWidth es el ancho del indicador de sonido en la barra superior.
const SoundWidth = 30

// Layout son las zonas interactivas que dejó el último Draw.
type Layout struct {
	Story []image.Rectangle // una por entrada de Story, en el mismo orden
	Top   int               // alto de la barra superior
	Foot  int               // y de la barra inferior
}

// Separator es el separador del rótulo jerárquico de ubicación.
const Separator = " › "

const (
	topH  = 19
	footH = 18
)

// Draw dibuja la interfaz de partida sobre el lienzo de 480 × 270.
func (h HUD) Draw(dst *ebiten.Image, atlas *render.Atlas) Layout {
	W, H := dst.Bounds().Dx(), dst.Bounds().Dy()
	lay := Layout{Top: topH + 2, Foot: H - footH}

	// Barra superior: reloj y velocidad, ubicación, seguimiento.
	Box(dst, image.Rect(2, 2, W-2, topH), Ivory)
	drawClock(dst, 7, 6)
	DrawText(dst, h.Clock, 18, 4, Plum)
	sx := 18 + TextWidth(h.Clock) + 6
	drawSpeed(dst, sx, 6, h.Paused, h.Speed)
	state := []string{"x1", "x4", "x16"}[min(max(h.Speed, 0), 2)]
	if h.Paused {
		state = "Pausa"
	}
	DrawText(dst, state, sx+14, 4, Plum)
	left := sx + 14 + TextWidth("Pausa") + 10
	if h.Sound != nil {
		h.Sound.draw(dst, left, 5)
		left += SoundWidth + 6
	}
	loc := strings.Join(h.Location, Separator)
	right := W - 8
	if h.Following != "" {
		f := "Siguiendo a " + h.Following
		fw := TextWidth(f)
		drawEye(dst, right-fw-14, 7)
		DrawText(dst, f, right-fw, 4, Plum)
		right -= fw + 20
	}
	loc = Fit(loc, right-left)
	lx := max(left, (W-TextWidth(loc))/2)
	DrawText(dst, loc, lx, 4, Plum)

	drawControls(dst, h.Controls)

	// Panel de historia con el retrato de la giantess.
	if h.Face != "" {
		lay.Story = h.drawStory(dst, atlas, H-footH-4)
	}
	if h.Card != nil {
		h.Card.draw(dst, atlas, W-4, H-footH-4)
	}
	return lay
}

// drawControls dibuja la barra inferior: controles como teclas con su rótulo.
func drawControls(dst *ebiten.Image, controls []Control) {
	W, H := dst.Bounds().Dx(), dst.Bounds().Dy()
	vector.FillRect(dst, 0, float32(H-footH), float32(W), float32(footH), Plum, false)
	x := 4
	for _, c := range controls {
		kw := TextWidth(c.Key) + 6
		lw := TextWidth(c.Label)
		if x+kw+3+lw > W-4 {
			break
		}
		Box(dst, image.Rect(x, H-footH+2, x+kw, H-2), Ivory)
		DrawText(dst, c.Key, x+3, H-footH+2, Plum)
		DrawText(dst, c.Label, x+kw+3, H-footH+3, Ivory)
		x += kw + 3 + lw + 8
	}
}

func (h HUD) drawStory(dst *ebiten.Image, atlas *render.Atlas, bottom int) []image.Rectangle {
	const pw = 52
	pr := image.Rect(4, bottom-pw, 4+pw, bottom)
	Box(dst, pr, IvoryDim)
	atlas.DrawClip(dst, "portrait", render.Macro, h.Face, pr.Min.Add(image.Pt(2, 2)))

	const bw = 236
	type entry struct {
		lines []string
		time  string
		focus bool
		index int // en h.Story; -1 para el texto de espera
	}
	var entries []entry
	n := 0
	for i := len(h.Story) - 1; i >= 0 && n < 4; i-- {
		lines := WrapWidth(h.Story[i].Text, bw-48)
		if n+len(lines) > 4 {
			break
		}
		entries = append([]entry{{lines, h.Story[i].Time, h.Story[i].Focusable, i}}, entries...)
		n += len(lines)
	}
	if n == 0 {
		entries = []entry{{lines: []string{"La casa todavía duerme."}, index: -1}}
		n = 1
	}
	bh := n*LineHeight + 8
	br := image.Rect(pr.Max.X+7, bottom-bh, pr.Max.X+7+bw, bottom)
	Box(dst, br, Ivory)
	// cola del globo hacia el retrato
	for i := range 5 {
		vector.FillRect(dst, float32(br.Min.X-5+i), float32(bottom-16+i), 1, float32(9-2*i), Plum, false)
		if i > 0 {
			vector.FillRect(dst, float32(br.Min.X-5+i), float32(bottom-15+i), 1, float32(7-2*i+1), Ivory, false)
		}
	}
	// las entradas que no entraron quedan con un rectángulo vacío
	rects := make([]image.Rectangle, len(h.Story))
	y := br.Min.Y + 4
	for _, e := range entries {
		top := y
		if e.time != "" {
			DrawText(dst, e.time, br.Min.X+5, y, PlumSoft)
		}
		for _, l := range e.lines {
			DrawText(dst, l, br.Min.X+43, y, Plum)
			y += LineHeight
		}
		r := image.Rect(br.Min.X+2, top, br.Max.X-2, y)
		if e.focus {
			// marca de «se puede enfocar»: una flecha turquesa al final
			ax, ay := float32(br.Max.X-8), float32(top+4)
			vector.FillRect(dst, ax, ay, 2, 5, Turquoise, false)
			vector.FillRect(dst, ax+2, ay+1, 1, 3, Turquoise, false)
			vector.FillRect(dst, ax+3, ay+2, 1, 1, Turquoise, false)
		}
		if e.index >= 0 {
			rects[e.index] = r
		}
	}
	return rects
}

func (c Card) draw(dst *ebiten.Image, atlas *render.Atlas, right, bottom int) {
	const w = 150
	lines := 1
	if c.Place != "" {
		lines++
	}
	if c.Activity != "" {
		lines++
	}
	if c.Carrying != "" {
		lines++
	}
	if c.Order != "" {
		lines++
	}
	h := lines*LineHeight + len(c.Needs)*11 + 8
	r := image.Rect(right-w, bottom-h, right, bottom)
	Box(dst, r, Ivory)
	x, y := r.Min.X+5, r.Min.Y+4
	name := Fit(c.Name, w-24)
	DrawText(dst, name, x, y, Plum)
	if c.Emote != "" {
		atlas.DrawClip(dst, "emote", render.Macro, c.Emote, image.Pt(x+TextWidth(name)+8, y+11))
	}
	y += LineHeight
	for _, l := range []string{c.Place, c.Activity, c.Carrying} {
		if l != "" {
			DrawText(dst, Fit(l, w-10), x, y, PlumSoft)
			y += LineHeight
		}
	}
	for _, n := range c.Needs {
		drawNeed(dst, x, y+1, w-10, n)
		y += 11
	}
	if c.Order != "" {
		DrawText(dst, Fit(c.Order, w-10), x, y, PlumSoft)
	}
}

// drawNeed dibuja una necesidad: rótulo, barra y valor. Una necesidad baja
// lleva además un «!» y la barra rayada, para leerse sin color.
func drawNeed(dst *ebiten.Image, x, y, w int, n Need) {
	const labelW = 58
	low := n.Value < 300
	label := n.Label
	if low {
		label += " !"
	}
	DrawText(dst, label, x, y-3, Plum)
	bx, bw := x+labelW, w-labelW
	vector.FillRect(dst, float32(bx), float32(y+1), float32(bw), 7, Plum, false)
	vector.FillRect(dst, float32(bx+1), float32(y+2), float32(bw-2), 5, IvoryDim, false)
	fill := int(float64(bw-2) * float64(n.Value) / 1000)
	c := Turquoise
	if low {
		c = Coral
	}
	vector.FillRect(dst, float32(bx+1), float32(y+2), float32(fill), 5, c, false)
	if low {
		for i := 0; i < fill; i += 3 {
			vector.FillRect(dst, float32(bx+1+i), float32(y+2), 1, 5, Plum, false)
		}
	}
}

func drawClock(dst *ebiten.Image, x, y int) {
	fx, fy := float32(x), float32(y)
	vector.FillRect(dst, fx+1, fy, 6, 8, Plum, false)
	vector.FillRect(dst, fx, fy+1, 8, 6, Plum, false)
	vector.FillRect(dst, fx+1, fy+1, 6, 6, Ivory, false)
	vector.FillRect(dst, fx+3, fy+2, 1, 3, Plum, false)
	vector.FillRect(dst, fx+4, fy+4, 2, 1, Plum, false)
}

// drawSpeed dibuja la velocidad por forma: dos barras en pausa, uno a tres
// triángulos según la velocidad.
func drawSpeed(dst *ebiten.Image, x, y int, paused bool, speed int) {
	fx, fy := float32(x), float32(y)
	if paused {
		vector.FillRect(dst, fx+2, fy, 3, 8, Plum, false)
		vector.FillRect(dst, fx+7, fy, 3, 8, Plum, false)
		return
	}
	for i := range min(max(speed, 0), 2) + 1 {
		ox := fx + float32(i*4)
		for k := range 4 {
			vector.FillRect(dst, ox+float32(k), fy+float32(k), 1, float32(8-2*k), Plum, false)
		}
	}
}

// Waves es cuántas ondas muestra el parlante: ninguna en silencio y de una a
// tres según el volumen.
func (s Sound) Waves() int {
	if s.Muted || s.Volume <= 0 {
		return 0
	}
	return min(3, 1+int(s.Volume*3-0.001))
}

// draw dibuja el parlante y la tecla M.
func (s Sound) draw(dst *ebiten.Image, x, y int) {
	fx, fy := float32(x), float32(y)
	vector.FillRect(dst, fx, fy+3, 3, 4, Plum, false)
	for k := range 4 {
		vector.FillRect(dst, fx+3+float32(k), fy+2-float32(k), 1, 6+2*float32(k), Plum, false)
	}
	if s.Muted || s.Volume <= 0 {
		for k := range 5 { // tachado: una equis junto al parlante
			vector.FillRect(dst, fx+9+float32(k), fy+2+float32(k), 1, 1, Coral, false)
			vector.FillRect(dst, fx+13-float32(k), fy+2+float32(k), 1, 1, Coral, false)
		}
	} else {
		for i := range s.Waves() {
			ox := fx + 9 + float32(2*i)
			h := float32(2 + 2*i)
			vector.FillRect(dst, ox, fy+5-h/2, 1, h, Plum, false)
		}
	}
	kx := x + 17
	Box(dst, image.Rect(kx, y-2, kx+TextWidth("M")+6, y+11), IvoryDim)
	DrawText(dst, "M", kx+3, y-2, Plum)
}

func drawEye(dst *ebiten.Image, x, y int) {
	fx, fy := float32(x), float32(y)
	vector.FillRect(dst, fx+2, fy, 7, 1, Plum, false)
	vector.FillRect(dst, fx, fy+1, 11, 4, Plum, false)
	vector.FillRect(dst, fx+2, fy+5, 7, 1, Plum, false)
	vector.FillRect(dst, fx+2, fy+1, 7, 4, Ivory, false)
	vector.FillRect(dst, fx+4, fy+1, 3, 4, Lilac, false)
	vector.FillRect(dst, fx+5, fy+2, 1, 2, Plum, false)
}
