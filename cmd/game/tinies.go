package main

import (
	"cmp"
	"image"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/ui"
	"gtsandbox/internal/view"
)

// tinyGroup son los tinies de una misma superficie, como se indican en la
// vista macro (donde los tinies no se dibujan).
type tinyGroup struct {
	surface ids.ID
	tinies  []ids.ID
	dots    []image.Point // cada tiny en el lienzo
	at      image.Point   // centro del grupo en el lienzo
	alarmed bool
	badge   image.Rectangle // dónde quedó el globo
}

// isTiny dice si una cosa del Frame es un tiny vivo: un agente con
// necesidades que no es un Space (la giantess sí lo es) ni son restos.
func isTiny(t view.Thing) bool { return len(t.Needs) > 0 && t.Remains == "" }

// tinyGroups agrupa los tinies por superficie, en orden de ID, y ubica cada
// globo para que no pise a los anteriores.
func tinyGroups(f *view.Frame, proj camera.Projector) []tinyGroup {
	var out []tinyGroup
	index := map[ids.ID]int{}
	for _, t := range f.Things {
		if !isTiny(t) {
			continue
		}
		i, ok := index[t.Space]
		if !ok {
			i = len(out)
			index[t.Space] = i
			out = append(out, tinyGroup{surface: t.Space})
		}
		x, y := proj.ToScreen(t.Pose.Pos)
		g := &out[i]
		g.tinies = append(g.tinies, t.ID)
		g.dots = append(g.dots, image.Pt(int(math.Round(float64(x))), int(math.Round(float64(y)))))
		g.alarmed = g.alarmed || t.Alarmed
	}
	slices.SortFunc(out, func(a, b tinyGroup) int { return cmp.Compare(a.surface, b.surface) })
	var placed []image.Rectangle
	for i := range out {
		g := &out[i]
		var sum image.Point
		top := math.MaxInt
		for _, d := range g.dots {
			sum = sum.Add(d)
			top = min(top, d.Y)
		}
		g.at = image.Pt(sum.X/len(g.dots), top)
		b := ui.TinyBadge{Count: len(g.tinies), Alarmed: g.alarmed}
		for lift := 0; ; lift += 4 {
			g.badge = b.Rect(g.at, lift)
			if !slices.ContainsFunc(placed, g.badge.Overlaps) || lift > 80 {
				break
			}
		}
		placed = append(placed, g.badge)
	}
	return out
}

// drawTinyBadges marca en macro dónde están los tinies: un punto por tiny y
// un globo por superficie con cuántos hay.
func (g *game) drawTinyBadges(dst *ebiten.Image, groups []tinyGroup) {
	for _, gr := range groups {
		for _, d := range gr.dots {
			vector.FillRect(dst, float32(d.X-1), float32(d.Y-1), 3, 3, ui.Plum, false)
			vector.FillRect(dst, float32(d.X), float32(d.Y), 1, 1, ui.Turquoise, false)
		}
	}
	for _, gr := range groups {
		ui.TinyStem(dst, gr.badge, gr.at)
	}
	for _, gr := range groups {
		ui.TinyBadge{Count: len(gr.tinies), Alarmed: gr.alarmed}.Draw(dst, g.art.atlas, gr.badge)
	}
}

// clickTinyBadge atiende un clic sobre un globo: entra a la vista de cerca de
// esa superficie o, si no tiene, elige al primer tiny del grupo.
func (g *game) clickTinyBadge(p image.Point) bool {
	for _, gr := range g.art.badges {
		if !p.In(gr.badge) {
			continue
		}
		if s := g.microSurface(g.run.Cur, gr.surface); !s.IsNil() {
			g.stopFollowing()
			g.sel = s
			g.toggleArtMicro()
		} else {
			g.sel = gr.tinies[0]
		}
		return true
	}
	return false
}
