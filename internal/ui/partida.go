package ui

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gtsandbox/internal/render"
)

// Estilo de la interfaz de partida (propuesta, 17.6):
// paneles marfil con esquinas redondeadas y borde ciruela, texto ciruela en
// una fuente pixel de 12 px con tildes y eñes.
var (
	Ivory      = color.NRGBA{0xFF, 0xF1, 0xD8, 0xFF}
	IvoryPanel = color.NRGBA{0xFF, 0xF1, 0xD8, 0xE8}
	IvoryDim   = color.NRGBA{0xF2, 0xE0, 0xC4, 0xFF}
	Plum       = color.NRGBA{0x3F, 0x36, 0x5C, 0xFF}
	PlumSoft   = color.NRGBA{0x79, 0x69, 0xA4, 0xFF}
	Turquoise  = color.NRGBA{0x76, 0xD8, 0xCB, 0xFF}
	Coral      = color.NRGBA{0xFF, 0x9C, 0xAD, 0xFF}
	Lilac      = color.NRGBA{0x9A, 0x8B, 0xE3, 0xFF}
	Amber      = color.NRGBA{0xE8, 0x9B, 0x2D, 0xFF}
	Green      = color.NRGBA{0x6E, 0xC4, 0x8F, 0xFF}
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

// Card es la ficha derecha de la entidad elegida (propuesta, 17.6).
type Card struct {
	Name      string
	Place     string
	Activity  string
	Dest      string // destino cuando aporta información distinta a Activity
	Carrying  string
	Needs     []Need
	Order     string // orden disponible ("Clic der.: ir allí")
	Emote     string // símbolo actual, para repetirlo junto al nombre
	Diary     []StoryLine
	Following bool // ya se está siguiendo a esta entidad
	CanFollow bool // muestra el botón Seguir
	CanGo     bool // muestra el botón Ir allí
}

// StoryIcon identifica el icono pequeño que acompaña un hito de la crónica.
type StoryIcon uint8

const (
	IconNone StoryIcon = iota
	IconDanger
	IconEncounter
	IconColony
	IconRoutine
	IconSummary
)

// StoryLine es una entrada del panel de historia.
type StoryLine struct {
	Time      string
	Text      string
	Focusable bool // se puede enfocar al actor o al lugar
	Tone      Tone
	Icon      StoryIcon
}

// Tone es el color de la hora y del icono de una línea de la crónica.
type Tone uint8

const (
	ToneNormal    Tone = iota
	ToneMilestone      // hito de colonia: turquesa
	ToneDanger         // peligro / deceso: ámbar o coral
	ToneEncounter      // encuentro: cian (turquesa)
)

func (t Tone) color() color.Color {
	switch t {
	case ToneMilestone, ToneEncounter:
		return Turquoise
	case ToneDanger:
		return Amber
	}
	return PlumSoft
}

// NoticeKind distingue los tipos de aviso arriba a la derecha (propuesta, 17.6).
type NoticeKind uint8

const (
	NoticeDeath NoticeKind = iota
	NoticeThreat
	NoticePalm
	NoticeRefuge
)

func (k NoticeKind) stripe() color.Color {
	switch k {
	case NoticeThreat:
		return Amber
	case NoticePalm:
		return Turquoise
	case NoticeRefuge:
		return Green
	}
	return Coral
}

// Notice es una notificación temporal en pantalla.
type Notice struct {
	Kind      NoticeKind
	Title     string
	Cause     string
	Focusable bool
}

// Pantry resume el depósito de la colonia para la barra superior.
type Pantry struct {
	Portions int32
	Goal     int32
}

// PanelTab es la pestaña activa del panel lateral izquierdo.
type PanelTab int

const (
	TabChronicle PanelTab = iota
	TabColony
	TabGiantess
)

var tabLabels = [3]string{"Crónica", "Colonia", "Giantess"}

// ActivityCount resume cuántos tinies realizan una actividad en la pestaña Colonia.
type ActivityCount struct {
	Label string
	Count int
}

// TinyRow es una fila de la lista de tinies en la pestaña Colonia.
type TinyRow struct {
	Name     string
	Activity string
	Hungry   bool
	Tired    bool
	Alarmed  bool
}

// SiteProgress resume el avance de la obra del refugio auxiliar.
type SiteProgress struct {
	Delivered int32
	Required  int32
	Done      bool
}

// ColonyPanel es el contenido de la pestaña Colonia.
type ColonyPanel struct {
	Portions   int32
	Goal       int32
	Site       *SiteProgress
	Activities []ActivityCount
	Tinies     []TinyRow
}

// GiantessPanel es el contenido de la pestaña Giantess.
type GiantessPanel struct {
	Face     string
	Activity string
	Dest     string
	Needs    []Need
}

// SidePanel es el panel lateral izquierdo con tres pestañas (propuesta, 17.6).
type SidePanel struct {
	Tab       PanelTab
	Scroll    int // desplazamiento (en filas desde el final en Crónica, desde el principio en Colonia)
	Chronicle []StoryLine
	Colony    *ColonyPanel
	Giantess  *GiantessPanel
}

// HUD es todo lo que muestra la interfaz de partida en un frame. Lo arma el
// juego a partir del Frame; Draw solo lo dibuja.
type HUD struct {
	Clock      string
	Paused     bool
	Speed      int // índice en view.Speeds, 0..2
	Location   []string
	Following  string
	Population int      // cantidad de tinies vivos
	ShowPop    bool     // muestra el indicador de población en la barra superior
	Pantry     *Pantry  // despensa de la colonia en la barra superior
	Notices    []Notice // notificaciones activas (decesos, amenazas, palma, refugio)
	Card       *Card
	Panel      *SidePanel  // panel lateral izquierdo (nil si está plegado con H)
	Story      []StoryLine // últimas líneas de la crónica (compatibilidad y resúmenes)
	Face       string      // expresión actual de la giantess
	Controls   []Control
	Sound      *Sound // nil: sin sonido (capturas, sin dispositivo)
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
	PanelRect       image.Rectangle    // panel lateral izquierdo (vacío si está plegado)
	PanelClose      image.Rectangle    // botón «×» para plegar el panel
	Tabs            [3]image.Rectangle // pestañas Crónica, Colonia, Giantess
	Story           []image.Rectangle  // una por entrada de Panel.Chronicle
	Activities      []image.Rectangle  // una por entrada de Panel.Colony.Activities
	Tinies          []image.Rectangle  // una por entrada de Panel.Colony.Tinies
	PanelScrollUp   image.Rectangle
	PanelScrollDown image.Rectangle
	GiantessFocus   image.Rectangle
	CardRect        image.Rectangle   // ficha derecha (vacía si no hay)
	CardFollow      image.Rectangle   // botón Seguir de la ficha
	CardGo          image.Rectangle   // botón Ir allí de la ficha
	Notices         []image.Rectangle // una por entrada de Notices
	NoticeClose     []image.Rectangle // botón «×» de cada entrada de Notices
	Top             int               // alto de la barra superior
	Foot            int               // y de la barra inferior
}

// Separator es el separador del rótulo jerárquico de ubicación.
const Separator = " › "

const (
	topH  = 19
	footH = 18
)

// sidePanelRect ubica el panel lateral izquierdo según el tamaño del lienzo y
// la pestaña activa, manteniendo un tamaño compacto que no tape media casa.
func (p *SidePanel) rect(W, H int) image.Rectangle {
	w := min(204, max(172, W/4))
	avail := H - topH - footH - 10
	h := min(162, max(108, avail))
	if p != nil && p.Tab == TabGiantess {
		needs := 0
		if p.Giantess != nil {
			needs = len(p.Giantess.Needs)
		}
		h = min(h, 82+needs*12)
	}
	return image.Rect(4, topH+5, 4+w, topH+5+h)
}

func sidePanelTabs(pr image.Rectangle) (tabs [3]image.Rectangle, closeBtn image.Rectangle) {
	const closeW = 14
	closeBtn = image.Rect(pr.Max.X-closeW-2, pr.Min.Y+2, pr.Max.X-2, pr.Min.Y+17)
	usable := pr.Dx() - 4 - closeW
	tw := usable / 3
	for i := range 3 {
		x0 := pr.Min.X + 2 + i*tw
		x1 := x0 + tw
		if i == 2 {
			x1 = closeBtn.Min.X
		}
		tabs[i] = image.Rect(x0, pr.Min.Y+2, x1, pr.Min.Y+17)
	}
	return tabs, closeBtn
}

// cardRect calcula el rectángulo de la ficha derecha y sus botones sin dibujar.
func (c Card) layout(W, H int) (rect, followBtn, goBtn image.Rectangle) {
	w := 164
	if len(c.Diary) > 0 {
		w = min(204, W/2-8)
	}
	lines := 1
	for _, s := range []string{c.Place, c.Activity, c.Dest, c.Carrying, c.Order} {
		if s != "" {
			lines++
		}
	}
	if len(c.Diary) > 0 {
		lines += len(c.Diary)
	}
	h := lines*LineHeight + len(c.Needs)*11 + 8
	if len(c.Diary) > 0 {
		h += 3
	}
	hasBtns := c.CanFollow || c.CanGo
	if hasBtns {
		h += 18
	}
	bottom := H - footH - 4
	top := max(topH+5, bottom-h)
	rect = image.Rect(W-4-w, top, W-4, bottom)
	if hasBtns {
		by := rect.Max.Y - 17
		bx := rect.Min.X + 5
		bw := (w - 14) / 2
		if c.CanFollow {
			followBtn = image.Rect(bx, by, bx+bw, by+14)
		}
		if c.CanGo {
			gx := bx + bw + 4
			if !c.CanFollow {
				gx = bx
			}
			goBtn = image.Rect(gx, by, gx+bw, by+14)
		}
	}
	return rect, followBtn, goBtn
}

// LayoutFor calcula las zonas principales del HUD para un lienzo de W × H sin
// dibujar (útil para pruebas de ajuste a 1×, 2× y 3×).
func (h HUD) LayoutFor(W, H int) Layout {
	lay := Layout{Top: topH + 2, Foot: H - footH}
	if h.Panel != nil {
		pr := h.Panel.rect(W, H)
		lay.PanelRect = pr
		lay.Tabs, lay.PanelClose = sidePanelTabs(pr)
	}
	if h.Card != nil {
		lay.CardRect, lay.CardFollow, lay.CardGo = h.Card.layout(W, H)
	}
	if len(h.Notices) > 0 {
		lay.Notices, lay.NoticeClose = noticesLayout(h.Notices, W-4, topH+5)
	}
	return lay
}

// Draw dibuja la interfaz de partida sobre el lienzo.
func (h HUD) Draw(dst *ebiten.Image, atlas *render.Atlas) Layout {
	W, H := dst.Bounds().Dx(), dst.Bounds().Dy()
	lay := Layout{Top: topH + 2, Foot: H - footH}

	// Barra superior: reloj y velocidad, sonido, ubicación, despensa, población, seguimiento.
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
	left := sx + 14 + TextWidth("Pausa") + 8
	if h.Sound != nil {
		h.Sound.draw(dst, left, 5)
		left += SoundWidth + 6
	}
	right := W - 8
	if h.ShowPop {
		pop := strconv.Itoa(h.Population)
		pw := TextWidth(pop)
		DrawText(dst, pop, right-pw, 4, Plum)
		if atlas != nil {
			atlas.DrawClip(dst, "tiny", render.Macro, "idle", image.Pt(right-pw-6, 10))
		}
		right -= pw + 16
	}
	if h.Pantry != nil {
		txt := strconv.Itoa(int(h.Pantry.Portions))
		if h.Pantry.Goal > 0 {
			txt = fmt.Sprintf("%d/%d", h.Pantry.Portions, h.Pantry.Goal)
		}
		tw := TextWidth(txt)
		ink := Plum
		if h.Pantry.Goal > 0 && h.Pantry.Portions < h.Pantry.Goal/2 {
			ink = Amber
		}
		DrawText(dst, txt, right-tw, 4, ink)
		drawPantryIcon(dst, right-tw-10, 6)
		right -= tw + 18
	}
	if h.Following != "" && right-left > 110 {
		f := Fit("Siguiendo a "+h.Following, max(60, (right-left)/2))
		fw := TextWidth(f)
		drawEye(dst, right-fw-13, 7)
		DrawText(dst, f, right-fw, 4, Plum)
		right -= fw + 18
	}
	loc := strings.Join(h.Location, Separator)
	loc = Fit(loc, max(right-left, 0))
	lx := max(left, (W-TextWidth(loc))/2)
	if lx+TextWidth(loc) > right {
		lx = left
	}
	DrawText(dst, loc, lx, 4, Plum)

	drawControls(dst, h.Controls)

	if len(h.Notices) > 0 {
		lay.Notices, lay.NoticeClose = drawNotices(dst, atlas, h.Notices, W-4, topH+5)
	}
	if h.Panel != nil {
		h.drawSidePanel(dst, atlas, &lay, W, H)
	}
	if h.Card != nil {
		lay.CardRect, lay.CardFollow, lay.CardGo = h.Card.draw(dst, atlas, W, H)
	}
	return lay
}

func (h HUD) drawSidePanel(dst *ebiten.Image, atlas *render.Atlas, lay *Layout, W, H int) {
	p := h.Panel
	pr := p.rect(W, H)
	lay.PanelRect = pr
	Box(dst, pr, IvoryPanel)

	// Pestañas arriba: Crónica | Colonia | Giantess + botón «×» para plegar.
	lay.Tabs, lay.PanelClose = sidePanelTabs(pr)
	for i, tr := range lay.Tabs {
		active := PanelTab(i) == p.Tab
		fill := IvoryDim
		ink := PlumSoft
		if active {
			fill = Ivory
			ink = Plum
		}
		vector.FillRect(dst, float32(tr.Min.X), float32(tr.Min.Y), float32(tr.Dx()), float32(tr.Dy()), fill, false)
		if active {
			vector.FillRect(dst, float32(tr.Min.X+2), float32(tr.Max.Y-2), float32(tr.Dx()-4), 2, Turquoise, false)
		}
		if i > 0 {
			vector.FillRect(dst, float32(tr.Min.X), float32(tr.Min.Y+2), 1, float32(tr.Dy()-4), PlumSoft, false)
		}
		lbl := Fit(tabLabels[i], tr.Dx()-4)
		DrawText(dst, lbl, tr.Min.X+(tr.Dx()-TextWidth(lbl))/2, tr.Min.Y+1, ink)
	}
	// Botón «×» para plegar el panel con un clic (equivalente a H).
	cr := lay.PanelClose
	vector.FillRect(dst, float32(cr.Min.X), float32(cr.Min.Y), float32(cr.Dx()), float32(cr.Dy()), IvoryDim, false)
	vector.FillRect(dst, float32(cr.Min.X), float32(cr.Min.Y+2), 1, float32(cr.Dy()-4), PlumSoft, false)
	cx, cy := float32(cr.Min.X+5), float32(cr.Min.Y+5)
	for k := range 5 {
		vector.FillRect(dst, cx+float32(k), cy+float32(k), 1, 1, PlumSoft, false)
		vector.FillRect(dst, cx+4-float32(k), cy+float32(k), 1, 1, PlumSoft, false)
	}
	vector.FillRect(dst, float32(pr.Min.X+2), float32(pr.Min.Y+17), float32(pr.Dx()-4), 1, Plum, false)

	body := image.Rect(pr.Min.X+4, pr.Min.Y+20, pr.Max.X-4, pr.Max.Y-4)
	switch p.Tab {
	case TabChronicle:
		h.drawChronicleTab(dst, lay, body)
	case TabColony:
		h.drawColonyTab(dst, lay, body)
	case TabGiantess:
		h.drawGiantessTab(dst, atlas, lay, body)
	}
}

func (h HUD) drawChronicleTab(dst *ebiten.Image, lay *Layout, body image.Rectangle) {
	p := h.Panel
	lay.Story = make([]image.Rectangle, len(p.Chronicle))
	if len(p.Chronicle) == 0 {
		DrawText(dst, "La casa todavía duerme.", body.Min.X+2, body.Min.Y+4, PlumSoft)
		return
	}
	textW := body.Dx() - 50
	type wrapped struct {
		idx   int
		lines []string
	}
	all := make([]wrapped, len(p.Chronicle))
	for i, s := range p.Chronicle {
		ls := WrapWidth(s.Text, max(40, textW))
		if len(ls) == 0 {
			ls = []string{""}
		}
		all[i] = wrapped{idx: i, lines: ls}
	}
	maxLines := max(1, (body.Dy()-12)/LineHeight)
	end := max(1, len(all)-max(0, p.Scroll))
	end = min(end, len(all))
	start := end
	used := 0
	for start > 0 {
		n := len(all[start-1].lines)
		if used+n > maxLines && used > 0 {
			break
		}
		used += n
		start--
	}
	y := body.Min.Y + 2
	for _, item := range all[start:end] {
		s := p.Chronicle[item.idx]
		top := y
		drawStoryIcon(dst, body.Min.X+1, y+3, s.Icon, s.Tone)
		if s.Time != "" {
			DrawText(dst, s.Time, body.Min.X+9, y, s.Tone.color())
		}
		for _, l := range item.lines {
			if y+LineHeight > body.Max.Y-10 {
				break
			}
			DrawText(dst, l, body.Min.X+43, y, Plum)
			y += LineHeight
		}
		r := image.Rect(body.Min.X, top, body.Max.X-2, y)
		if s.Focusable {
			ax, ay := float32(body.Max.X-7), float32(top+4)
			vector.FillRect(dst, ax, ay, 2, 5, Turquoise, false)
			vector.FillRect(dst, ax+2, ay+1, 1, 3, Turquoise, false)
			vector.FillRect(dst, ax+3, ay+2, 1, 1, Turquoise, false)
		}
		lay.Story[item.idx] = r
	}
	drawScrollBar(dst, body, start > 0, end < len(all), &lay.PanelScrollUp, &lay.PanelScrollDown)
}

func (h HUD) drawColonyTab(dst *ebiten.Image, lay *Layout, body image.Rectangle) {
	col := h.Panel.Colony
	if col == nil {
		DrawText(dst, "Sin colonia activa.", body.Min.X+2, body.Min.Y+4, PlumSoft)
		return
	}
	x, y, w := body.Min.X+2, body.Min.Y+2, body.Dx()-4

	// Barra de la despensa.
	pantryVal := int32(1000)
	pantryTxt := strconv.Itoa(int(col.Portions))
	if col.Goal > 0 {
		pantryVal = min(1000, col.Portions*1000/col.Goal)
		pantryTxt = fmt.Sprintf("%d/%d", col.Portions, col.Goal)
	}
	drawProgressBar(dst, x, y, w, "Despensa", pantryTxt, pantryVal, col.Goal > 0 && col.Portions < col.Goal/2)
	y += 14

	// Avance del refugio auxiliar.
	if col.Site != nil {
		s := col.Site
		val := int32(0)
		if s.Required > 0 {
			val = min(1000, s.Delivered*1000/s.Required)
		}
		txt := fmt.Sprintf("%d/%d", s.Delivered, s.Required)
		if s.Done {
			val = 1000
			txt = "Listo"
		}
		drawProgressBar(dst, x, y, w, "Refugio", txt, val, false)
		y += 14
	}

	// Resumen de actividades (clic elige un tiny de esa actividad).
	if len(col.Activities) > 0 {
		vector.FillRect(dst, float32(x), float32(y), float32(w), 1, IvoryDim, false)
		y += 3
		lay.Activities = make([]image.Rectangle, len(col.Activities))
		cx := x
		for i, a := range col.Activities {
			chip := fmt.Sprintf("%s %d", a.Label, a.Count)
			cw := TextWidth(chip) + 8
			if cx+cw > x+w && cx > x {
				cx = x
				y += 15
			}
			if y+14 > body.Max.Y-28 {
				break
			}
			cr := image.Rect(cx, y, cx+cw, y+13)
			vector.FillRect(dst, float32(cr.Min.X), float32(cr.Min.Y), float32(cr.Dx()), float32(cr.Dy()), IvoryDim, false)
			DrawText(dst, chip, cr.Min.X+4, cr.Min.Y, Plum)
			lay.Activities[i] = cr
			cx += cw + 3
		}
		y += 16
	}

	// Lista desplazable de tinies con actividad y alertas.
	vector.FillRect(dst, float32(x), float32(y), float32(w), 1, IvoryDim, false)
	y += 3
	lay.Tinies = make([]image.Rectangle, len(col.Tinies))
	rowsAvail := max(1, (body.Max.Y-12-y)/LineHeight)
	start := min(max(0, h.Panel.Scroll), max(0, len(col.Tinies)-rowsAvail))
	end := min(len(col.Tinies), start+rowsAvail)
	for i := start; i < end; i++ {
		t := col.Tinies[i]
		rr := image.Rect(body.Min.X, y, body.Max.X-2, y+LineHeight)
		lay.Tinies[i] = rr
		alertW := 0
		if t.Alarmed {
			DrawText(dst, "!", rr.Max.X-8, y, Coral)
			alertW += 10
		}
		if t.Tired {
			DrawText(dst, "E", rr.Max.X-alertW-9, y, Amber)
			alertW += 10
		}
		if t.Hungry {
			DrawText(dst, "H", rr.Max.X-alertW-9, y, Amber)
			alertW += 10
		}
		nameW := 52
		DrawText(dst, Fit(t.Name, nameW), x, y, Plum)
		DrawText(dst, Fit(t.Activity, max(20, w-nameW-alertW-4)), x+nameW+3, y, PlumSoft)
		y += LineHeight
	}
	drawScrollBar(dst, body, start > 0, end < len(col.Tinies), &lay.PanelScrollUp, &lay.PanelScrollDown)
}

func (h HUD) drawGiantessTab(dst *ebiten.Image, atlas *render.Atlas, lay *Layout, body image.Rectangle) {
	gp := h.Panel.Giantess
	if gp == nil {
		DrawText(dst, "Sin giantess en escena.", body.Min.X+2, body.Min.Y+4, PlumSoft)
		return
	}
	x, y, w := body.Min.X+2, body.Min.Y+2, body.Dx()-4
	const pw = 52
	pr := image.Rect(x, y, x+pw, y+pw)
	Box(dst, pr, IvoryDim)
	if atlas != nil && gp.Face != "" {
		atlas.DrawClip(dst, "portrait", render.Macro, gp.Face, pr.Min.Add(image.Pt(2, 2)))
	}
	lay.GiantessFocus = pr

	tx := pr.Max.X + 6
	tw := max(40, w-pw-6)
	ty := y + 2
	DrawText(dst, "Giantess", tx, ty, Plum)
	ty += LineHeight + 1
	act := gp.Activity
	if act == "" {
		act = "En reposo"
	}
	for _, l := range WrapWidth(act, tw) {
		if ty+LineHeight > pr.Max.Y+4 {
			break
		}
		DrawText(dst, l, tx, ty, PlumSoft)
		ty += LineHeight
	}
	if gp.Dest != "" && ty+LineHeight <= pr.Max.Y+14 {
		DrawText(dst, Fit("Destino: "+gp.Dest, tw), tx, ty, PlumSoft)
	}

	ny := pr.Max.Y + 8
	for _, n := range gp.Needs {
		if ny+11 > body.Max.Y {
			break
		}
		drawNeed(dst, x, ny+1, w, n)
		ny += 12
	}
}

func drawScrollBar(dst *ebiten.Image, body image.Rectangle, canUp, canDown bool, upRect, downRect *image.Rectangle) {
	if !canUp && !canDown {
		return
	}
	by := body.Max.Y - 10
	if canUp {
		r := image.Rect(body.Max.X-24, by, body.Max.X-13, body.Max.Y)
		*upRect = r
		drawArrow(dst, r.Min.X+5, r.Min.Y+4, -1)
	}
	if canDown {
		r := image.Rect(body.Max.X-12, by, body.Max.X-1, body.Max.Y)
		*downRect = r
		drawArrow(dst, r.Min.X+5, r.Min.Y+4, 1)
	}
}

func drawProgressBar(dst *ebiten.Image, x, y, w int, label, valueText string, permille int32, low bool) {
	const labelW = 54
	DrawText(dst, label, x, y-2, Plum)
	vw := TextWidth(valueText)
	bx := x + labelW
	bw := max(20, w-labelW-vw-5)
	vector.FillRect(dst, float32(bx), float32(y+2), float32(bw), 7, Plum, false)
	vector.FillRect(dst, float32(bx+1), float32(y+3), float32(bw-2), 5, IvoryDim, false)
	fill := int(float64(bw-2) * float64(max(0, min(1000, permille))) / 1000)
	c := Turquoise
	if low {
		c = Amber
	}
	vector.FillRect(dst, float32(bx+1), float32(y+3), float32(fill), 5, c, false)
	DrawText(dst, valueText, bx+bw+4, y-2, PlumSoft)
}

func drawStoryIcon(dst *ebiten.Image, x, y int, icon StoryIcon, tone Tone) {
	fx, fy := float32(x), float32(y)
	c := tone.color()
	switch icon {
	case IconDanger:
		// Triángulo de advertencia / exclamación ámbar.
		vector.FillRect(dst, fx+3, fy, 1, 4, Amber, false)
		vector.FillRect(dst, fx+3, fy+5, 1, 1, Amber, false)
		vector.FillRect(dst, fx+1, fy+2, 1, 3, Amber, false)
		vector.FillRect(dst, fx+5, fy+2, 1, 3, Amber, false)
	case IconEncounter:
		// Rombo cian.
		vector.FillRect(dst, fx+3, fy, 1, 6, Turquoise, false)
		vector.FillRect(dst, fx+2, fy+1, 3, 4, Turquoise, false)
		vector.FillRect(dst, fx+1, fy+2, 5, 2, Turquoise, false)
	case IconColony:
		// Banderita / hito de colonia turquesa.
		vector.FillRect(dst, fx+2, fy, 1, 6, Turquoise, false)
		vector.FillRect(dst, fx+3, fy, 3, 3, Turquoise, false)
	case IconRoutine:
		vector.FillRect(dst, fx+2, fy+2, 3, 3, PlumSoft, false)
	case IconSummary:
		vector.FillRect(dst, fx+1, fy+2, 5, 1, PlumSoft, false)
		vector.FillRect(dst, fx+1, fy+4, 4, 1, PlumSoft, false)
	default:
		vector.FillRect(dst, fx+3, fy+2, 2, 2, c, false)
	}
}

func drawPantryIcon(dst *ebiten.Image, x, y int) {
	fx, fy := float32(x), float32(y)
	vector.FillRect(dst, fx+1, fy+1, 6, 5, Plum, false)
	vector.FillRect(dst, fx, fy+2, 8, 3, Plum, false)
	vector.FillRect(dst, fx+1, fy+2, 6, 3, Amber, false)
	vector.FillRect(dst, fx+2, fy+2, 2, 1, Ivory, false)
}

func noticesLayout(notices []Notice, right, top int) ([]image.Rectangle, []image.Rectangle) {
	const w = 188
	rects := make([]image.Rectangle, len(notices))
	closes := make([]image.Rectangle, len(notices))
	y := top
	for i, n := range notices {
		causeLines := WrapWidth(n.Cause, w-18)
		if len(causeLines) > 2 {
			causeLines = causeLines[:2]
		}
		h := (1+len(causeLines))*LineHeight + 6
		r := image.Rect(right-w, y, right, y+h)
		rects[i] = r
		closes[i] = image.Rect(r.Max.X-13, r.Min.Y+2, r.Max.X-2, r.Min.Y+13)
		y = r.Max.Y + 3
	}
	return rects, closes
}

// drawNotices dibuja las notificaciones temporales (decesos, amenazas, palma
// ofrecida, refugio terminado) apiladas arriba a la derecha.
func drawNotices(dst *ebiten.Image, atlas *render.Atlas, notices []Notice, right, top int) ([]image.Rectangle, []image.Rectangle) {
	const w = 188
	rects, closes := noticesLayout(notices, right, top)
	for i, n := range notices {
		r := rects[i]
		cr := closes[i]
		causeLines := WrapWidth(n.Cause, w-18)
		if len(causeLines) > 2 {
			causeLines = causeLines[:2]
		}
		Box(dst, r, Ivory)
		vector.FillRect(dst, float32(r.Min.X+2), float32(r.Min.Y+2), 3, float32(r.Dy()-4), n.Kind.stripe(), false)
		tx := r.Min.X + 9
		ty := r.Min.Y + 3
		title := Fit(n.Title, w-42)
		DrawText(dst, title, tx, ty, Plum)
		if n.Kind == NoticeDeath && atlas != nil {
			atlas.DrawClip(dst, "emote", render.Macro, "spirit", image.Pt(tx+TextWidth(title)+7, ty+11))
		}
		if n.Focusable {
			ax, ay := float32(r.Max.X-20), float32(ty+4)
			vector.FillRect(dst, ax, ay, 2, 5, Turquoise, false)
			vector.FillRect(dst, ax+2, ay+1, 1, 3, Turquoise, false)
			vector.FillRect(dst, ax+3, ay+2, 1, 1, Turquoise, false)
		}
		cx, cy := float32(cr.Min.X+3), float32(cr.Min.Y+3)
		for k := range 5 {
			vector.FillRect(dst, cx+float32(k), cy+float32(k), 1, 1, PlumSoft, false)
			vector.FillRect(dst, cx+4-float32(k), cy+float32(k), 1, 1, PlumSoft, false)
		}
		ty += LineHeight
		for _, line := range causeLines {
			DrawText(dst, line, tx, ty, PlumSoft)
			ty += LineHeight
		}
	}
	return rects, closes
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

func (c Card) draw(dst *ebiten.Image, atlas *render.Atlas, W, H int) (rect, followBtn, goBtn image.Rectangle) {
	r, followBtn, goBtn := c.layout(W, H)
	w := r.Dx()
	Box(dst, r, Ivory)
	x, y := r.Min.X+5, r.Min.Y+4
	name := Fit(c.Name, w-24)
	DrawText(dst, name, x, y, Plum)
	if c.Emote != "" && atlas != nil {
		atlas.DrawClip(dst, "emote", render.Macro, c.Emote, image.Pt(x+TextWidth(name)+8, y+11))
	}
	y += LineHeight
	for _, l := range []string{c.Place, c.Activity, c.Dest, c.Carrying} {
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
		y += LineHeight
	}
	if len(c.Diary) > 0 {
		vector.FillRect(dst, float32(x), float32(y+1), float32(w-10), 1, IvoryDim, false)
		y += 3
		for _, d := range c.Diary {
			if y+LineHeight > r.Max.Y-18 {
				break
			}
			DrawText(dst, d.Time, x, y, PlumSoft)
			DrawText(dst, Fit(d.Text, w-10-36), x+36, y, Plum)
			y += LineHeight
		}
	}
	if !followBtn.Empty() {
		lbl := "Seguir"
		fill := IvoryDim
		if c.Following {
			lbl = "Soltar"
			fill = Turquoise
		}
		drawButton(dst, followBtn, lbl, fill)
	}
	if !goBtn.Empty() {
		drawButton(dst, goBtn, "Ir allí", IvoryDim)
	}
	return r, followBtn, goBtn
}

func drawButton(dst *ebiten.Image, r image.Rectangle, label string, fill color.Color) {
	Box(dst, r, fill)
	lbl := Fit(label, r.Dx()-4)
	DrawText(dst, lbl, r.Min.X+(r.Dx()-TextWidth(lbl))/2, r.Min.Y, Plum)
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
	bx, bw := x+labelW, max(16, w-labelW)
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

// TinyBadge es el indicador de los tinies de una superficie en la vista
// macro, donde no se dibujan: el marcador del tiny y cuántos hay, con un
// palito (TinyStem) que baja hasta ellos. En coral si alguno está alarmado.
type TinyBadge struct {
	Count   int
	Alarmed bool
}

const (
	tinyBadgeH   = 13
	tinyBadgeGap = 6 // del borde inferior del globo hasta los tinies
)

// Rect es el globo para tinies centrados en at; lift lo sube más (para no
// pisar otro globo).
func (b TinyBadge) Rect(at image.Point, lift int) image.Rectangle {
	w := 4 + 5 + 3 + TextWidth(strconv.Itoa(b.Count)) + 4
	bottom := at.Y - tinyBadgeGap - lift
	return image.Rect(at.X-w/2, bottom-tinyBadgeH, at.X-w/2+w, bottom)
}

// TinyStem dibuja el palito que baja del globo r hasta los tinies en at. Se
// dibujan todos antes que los globos, para que un palito no cruce otro globo.
func TinyStem(dst *ebiten.Image, r image.Rectangle, at image.Point) {
	vector.FillRect(dst, float32(at.X), float32(r.Max.Y), 1, float32(at.Y-2-r.Max.Y), Plum, false)
}

// Draw dibuja el globo r.
func (b TinyBadge) Draw(dst *ebiten.Image, atlas *render.Atlas, r image.Rectangle) {
	fill := Ivory
	if b.Alarmed {
		fill = Coral
	}
	Box(dst, r, fill)
	atlas.DrawClip(dst, "tiny", render.Macro, "idle", image.Pt(r.Min.X+6, r.Min.Y+6))
	DrawText(dst, strconv.Itoa(b.Count), r.Min.X+12, r.Min.Y, Plum)
}
