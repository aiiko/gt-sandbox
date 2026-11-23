package ui

import (
	"image"
	"slices"
	"testing"
)

// El parlante muestra el volumen por la cantidad de ondas, y el silencio
// sin ondas (tachado).
func TestSoundWaves(t *testing.T) {
	for _, c := range []struct {
		s    Sound
		want int
	}{
		{Sound{Volume: 1}, 3}, {Sound{Volume: 0.5}, 2}, {Sound{Volume: 0.2}, 1},
		{Sound{Volume: 0}, 0}, {Sound{Muted: true, Volume: 1}, 0},
	} {
		if got := c.s.Waves(); got != c.want {
			t.Errorf("%+v: %d ondas, quería %d", c.s, got, c.want)
		}
	}
}

func TestWrap(t *testing.T) {
	got := Wrap([]string{"abcdefghijklmn", "", "corto"}, 8)
	want := []string{"abcdefgh", "  ijklmn", "", "corto"}
	if !slices.Equal(got, want) {
		t.Fatalf("%q", got)
	}
}

// La lista se desplaza para que la opción elegida siempre se vea, y cada
// opción visible se puede tocar.
func TestMenuLayoutScrollsToSelection(t *testing.T) {
	items := make([]MenuItem, 12)
	for sel := range items {
		m := Menu{Items: items, Selected: sel}
		lay := m.Layout(480, 270)
		r := lay.Items[sel]
		if r.Empty() {
			t.Fatalf("elegida %d: no se ve", sel)
		}
		if !r.In(lay.Panel) {
			t.Fatalf("elegida %d: fuera del panel", sel)
		}
		if got := lay.Hit(r.Min.Add(image.Pt(2, 2))); got != sel {
			t.Fatalf("elegida %d: Hit dio %d", sel, got)
		}
		shown := 0
		for _, it := range lay.Items {
			if !it.Empty() {
				shown++
			}
		}
		if shown != menuRows {
			t.Fatalf("elegida %d: %d opciones visibles", sel, shown)
		}
	}
	if lay := (Menu{Items: items[:3]}).Layout(480, 270); lay.Panel.Max.Y > 270 || lay.Hit(image.Pt(0, 0)) != -1 {
		t.Fatalf("panel %v", lay.Panel)
	}
}

// La interfaz de partida (panel lateral con pestañas, ficha derecha con
// botones, avisos, menú de pausa y ayuda) entra sin superponerse en 1×, 2× y
// 3× (propuesta, 17.6 y 17.8).
func TestHUDFitsAt1x2x3x(t *testing.T) {
	hud := HUD{
		Clock:      "07:15",
		Location:   []string{"Dormitorio", "Escritorio"},
		Population: 14,
		ShowPop:    true,
		Pantry:     &Pantry{Portions: 8, Goal: 28},
		Following:  "Tiny 01",
		Panel: &SidePanel{
			Tab: TabColony,
			Colony: &ColonyPanel{
				Portions: 8, Goal: 28,
				Site:       &SiteProgress{Delivered: 2, Required: 4},
				Activities: []ActivityCount{{Label: "Acarrea", Count: 3}, {Label: "Busca", Count: 2}},
				Tinies:     []TinyRow{{Name: "Tiny 01", Activity: "Acarrea", Hungry: true}},
			},
		},
		Card: &Card{
			Name:      "Tiny 01",
			Place:     "Dormitorio › Escritorio",
			Activity:  "Lleva una migaja al nido",
			Needs:     []Need{{Label: "Hambre", Value: 600}, {Label: "Energía", Value: 250}},
			Order:     "Clic der.: ir allí",
			Diary:     []StoryLine{{Time: "07:10", Text: "Explora cocina."}},
			CanFollow: true,
			CanGo:     true,
		},
		Notices: []Notice{
			{Kind: NoticeThreat, Title: "Amenaza a la colonia", Cause: "2 tinies en peligro en escritorio", Focusable: true},
		},
	}
	pause := PauseMenu{
		Title:    "Pausa",
		Subtitle: "Opciones",
		Items: []MenuItem{
			{Label: "Volumen", Value: "100%"},
			{Label: "Escala de interfaz", Value: "Automático"},
			{Label: "Pantalla completa", Value: "No"},
			{Label: "Movimiento reducido", Value: "No"},
			{Label: "Pausa ante amenazas", Value: "Sí"},
			{Label: "Volver"},
		},
		Selected: 2,
		Message:  "Partida guardada.",
	}
	for _, tc := range []struct {
		name string
		w, h int
	}{
		{"1x-1080p", 1920, 1080},
		{"2x-1080p", 960, 540},
		{"3x-1080p", 640, 360},
		{"min-768", 768, 432},
		{"min-480", 480, 270},
	} {
		screen := image.Rect(0, 0, tc.w, tc.h)
		lay := hud.LayoutFor(tc.w, tc.h)
		if !lay.PanelRect.In(screen) || lay.PanelRect.Min.Y < lay.Top || lay.PanelRect.Max.Y > lay.Foot {
			t.Errorf("%s: panel fuera de cuadro: %v (top %d, foot %d)", tc.name, lay.PanelRect, lay.Top, lay.Foot)
		}
		for i, tab := range lay.Tabs {
			if tab.Empty() || !tab.In(lay.PanelRect) {
				t.Errorf("%s: pestaña %d inválida: %v", tc.name, i, tab)
			}
			if i > 0 && tab.Overlaps(lay.Tabs[i-1]) {
				t.Errorf("%s: pestañas %d y %d se superponen", tc.name, i-1, i)
			}
		}
		if !lay.CardRect.In(screen) || lay.CardRect.Min.Y < lay.Top || lay.CardRect.Max.Y > lay.Foot {
			t.Errorf("%s: ficha fuera de cuadro: %v", tc.name, lay.CardRect)
		}
		if lay.PanelRect.Overlaps(lay.CardRect) {
			t.Errorf("%s: panel %v y ficha %v se superponen", tc.name, lay.PanelRect, lay.CardRect)
		}
		if lay.CardFollow.Empty() || lay.CardGo.Empty() || !lay.CardFollow.In(lay.CardRect) || !lay.CardGo.In(lay.CardRect) || lay.CardFollow.Overlaps(lay.CardGo) {
			t.Errorf("%s: botones de ficha inválidos: follow=%v go=%v card=%v", tc.name, lay.CardFollow, lay.CardGo, lay.CardRect)
		}
		for i, nr := range lay.Notices {
			if !nr.In(screen) || !lay.NoticeClose[i].In(nr) {
				t.Errorf("%s: aviso %d fuera de cuadro: %v", tc.name, i, nr)
			}
		}
		pl := pause.Layout(tc.w, tc.h)
		if !pl.Panel.In(screen) || pl.Items[pause.Selected].Empty() || pl.Hit(pl.Items[pause.Selected].Min.Add(image.Pt(2, 2))) != pause.Selected {
			t.Errorf("%s: menú de pausa inválido: %+v", tc.name, pl)
		}
		hr := HelpRect(tc.w, tc.h)
		if hr.Empty() || !hr.In(screen) {
			t.Errorf("%s: ayuda fuera de cuadro: %v", tc.name, hr)
		}
	}
}

// El panel lateral admite ancho y alto personalizados (al arrastrar bordes),
// acotados al lienzo, y expone sus zonas de arrastre.
func TestSidePanelResizeZonesAndClamp(t *testing.T) {
	h := HUD{Panel: &SidePanel{Tab: TabChronicle, Width: 280, Height: 220}}
	lay := h.LayoutFor(960, 540)
	if lay.PanelRect.Dx() != 280 || lay.PanelRect.Dy() != 220 {
		t.Fatalf("tamaño %v, quería 280×220", lay.PanelRect.Size())
	}
	if lay.PanelResizeR.Empty() || lay.PanelResizeB.Empty() || lay.PanelResizeBR.Empty() {
		t.Fatalf("zonas de redimensionado vacías: R=%v B=%v BR=%v", lay.PanelResizeR, lay.PanelResizeB, lay.PanelResizeBR)
	}
	// Valores extremos quedan acotados dentro del lienzo.
	hBig := HUD{Panel: &SidePanel{Tab: TabChronicle, Width: 9999, Height: 9999}}
	lBig := hBig.LayoutFor(960, 540)
	if lBig.PanelRect.Max.X > 960/2 || lBig.PanelRect.Max.Y > lBig.Foot {
		t.Fatalf("tamaño máximo desbordado: %v", lBig.PanelRect)
	}
	hTiny := HUD{Panel: &SidePanel{Tab: TabChronicle, Width: 10, Height: 10}}
	lTiny := hTiny.LayoutFor(960, 540)
	if lTiny.PanelRect.Dx() < 152 || lTiny.PanelRect.Dy() < 88 {
		t.Fatalf("tamaño mínimo demasiado chico: %v", lTiny.PanelRect)
	}
}
