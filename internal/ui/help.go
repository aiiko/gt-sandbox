package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gtsandbox/internal/render"
)

// helpControl es una fila de la columna de controles de la ayuda completa.
type helpControl struct{ Key, Desc string }

var helpControls = []helpControl{
	{"Espacio", "Pausar o reanudar"},
	{"1 / 2 / 3", "Velocidad x1, x4, x16"},
	{"N", "Avanzar un tick"},
	{"Tab", "Cambiar vista macro / micro"},
	{"Rueda", "Acercar o alejar la escena"},
	{"WASD / ↑↓←→", "Mover la cámara"},
	{"T", "Seguir a lo elegido"},
	{"F", "Volver al encuadre general"},
	{"H", "Mostrar u ocultar el panel"},
	{"Clic der.", "Ordenar ir a un punto"},
	{"M", "Silenciar sonido"},
	{"F11", "Pantalla completa"},
	{"F3", "Overlays técnicos"},
	{"Esc", "Soltar o abrir menú de pausa"},
}

// helpSymbol es una fila de la leyenda de símbolos de la ayuda completa.
type helpSymbol struct{ Clip, Label string }

var helpSymbols = []helpSymbol{
	{"discover", "Descubrimiento o susto"},
	{"explore", "Exploración o atención"},
	{"sweat", "En peligro o en la mano"},
	{"sleep", "Descansando o durmiendo"},
	{"note", "Mirando TV o conversando"},
	{"eye", "Giantess observando tinies"},
	{"anger", "Actividad interrumpida"},
	{"spirit", "Deceso de un tiny"},
}

// HelpRect devuelve el rectángulo del panel de ayuda completa en W × H.
func HelpRect(W, H int) image.Rectangle {
	w := min(444, W-16)
	h := min(232, H-topH-footH-12)
	x0 := (W - w) / 2
	y0 := max(topH+4, (H-footH-h)/2)
	return image.Rect(x0, y0, x0+w, y0+h)
}

// DrawHelp dibuja la ventana de ayuda completa con todos los controles y la
// leyenda de símbolos (F1 o ?, propuesta 17.6).
func DrawHelp(dst *ebiten.Image, atlas *render.Atlas) image.Rectangle {
	W, H := dst.Bounds().Dx(), dst.Bounds().Dy()
	vector.FillRect(dst, 0, 0, float32(W), float32(H), color.NRGBA{0x1B, 0x16, 0x2B, 0xB0}, false)

	r := HelpRect(W, H)
	Box(dst, r, Ivory)

	title := "Controles y leyenda de símbolos"
	DrawText(dst, title, r.Min.X+(r.Dx()-TextWidth(title))/2, r.Min.Y+4, Plum)
	vector.FillRect(dst, float32(r.Min.X+6), float32(r.Min.Y+18), float32(r.Dx()-12), 1, IvoryDim, false)

	mid := r.Min.X + r.Dx()*54/100
	vector.FillRect(dst, float32(mid), float32(r.Min.Y+22), 1, float32(r.Dy()-38), IvoryDim, false)

	// Columna izquierda: controles.
	lx := r.Min.X + 8
	ly := r.Min.Y + 23
	DrawText(dst, "Controles", lx, ly, PlumSoft)
	ly += LineHeight + 1
	maxY := r.Max.Y - 18
	for _, c := range helpControls {
		if ly+LineHeight > maxY {
			break
		}
		DrawText(dst, c.Key, lx, ly, Plum)
		DrawText(dst, Fit(c.Desc, mid-lx-74), lx+68, ly, PlumSoft)
		ly += LineHeight
	}

	// Columna derecha: leyenda de símbolos.
	rx := mid + 8
	ry := r.Min.Y + 23
	DrawText(dst, "Símbolos", rx, ry, PlumSoft)
	ry += LineHeight + 3
	for _, s := range helpSymbols {
		if ry+LineHeight > maxY {
			break
		}
		if atlas != nil {
			atlas.DrawClip(dst, "emote", render.Macro, s.Clip, image.Pt(rx+6, ry+11))
		}
		DrawText(dst, Fit(s.Label, r.Max.X-rx-24), rx+16, ry, Plum)
		ry += LineHeight + 2
	}

	hint := "F1 / ? / Esc para cerrar"
	DrawText(dst, hint, r.Min.X+(r.Dx()-TextWidth(hint))/2, r.Max.Y-LineHeight-3, PlumSoft)
	return r
}
