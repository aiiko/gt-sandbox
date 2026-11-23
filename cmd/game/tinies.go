package main

import (
	"cmp"
	"image"
	"maps"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/render"
	"gtsandbox/internal/ui"
	"gtsandbox/internal/view"
)

// tinyGroup son los tinies de una misma superficie, como se indican en la
// vista macro (donde los tinies no se dibujan).
type tinyGroup struct {
	surface    ids.ID
	tinies     []ids.ID
	dots       []image.Point // cada tiny en el lienzo
	at         image.Point   // centro del grupo en el lienzo
	alarmed    bool
	threatened bool
	activity   ui.BadgeActivity
	badge      image.Rectangle // dónde quedó el globo
}

// isTiny dice si una cosa del Frame es un tiny vivo: un agente con
// necesidades que no es un Space (la giantess sí lo es) ni son restos.
func isTiny(t view.Thing) bool { return len(t.Needs) > 0 && t.Remains == "" }

// tinyGroups agrupa los tinies por superficie, en orden de ID, y ubica cada
// globo para que no pise a los anteriores. Quien trepa una pata se marca en
// su base, como lo dibuja la vista macro (render.FlattenClimbs).
func tinyGroups(f *view.Frame, proj camera.Projector, visuals map[ids.ID]string) []tinyGroup {
	f = render.FlattenClimbs(f, visuals)
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
	threats := threatenedSurfaces(f)
	var placed []image.Rectangle
	for i := range out {
		g := &out[i]
		g.threatened = threats[g.surface]
		g.activity = dominantBadgeActivity(f, g.tinies)
		var sum image.Point
		top := math.MaxInt
		for _, d := range g.dots {
			sum = sum.Add(d)
			top = min(top, d.Y)
		}
		g.at = image.Pt(sum.X/len(g.dots), top)
		b := ui.TinyBadge{Count: len(g.tinies), Alarmed: g.alarmed, Threatened: g.threatened, Activity: g.activity}
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

func dominantBadgeActivity(f *view.Frame, tinies []ids.ID) ui.BadgeActivity {
	counts := map[string]int{}
	for _, id := range tinies {
		if t, ok := f.ThingByID(id); ok {
			cat := shortActivity(*t)
			if cat == "Huye" {
				return ui.BadgeFlee
			}
			if cat != "Libre" {
				counts[cat]++
			}
		}
	}
	best, bestN := "", 0
	for _, cat := range slices.Sorted(maps.Keys(counts)) {
		if counts[cat] > bestN {
			best, bestN = cat, counts[cat]
		}
	}
	return badgeActivityOf(best)
}

// drawTinyBadges marca en macro dónde están los tinies: un punto por tiny y
// un globo por superficie con cuántos hay, su actividad dominante y si hay amenaza.
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
		ui.TinyBadge{Count: len(gr.tinies), Alarmed: gr.alarmed, Threatened: gr.threatened, Activity: gr.activity}.Draw(dst, g.art.atlas, gr.badge)
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
			g.focusTinies()
		} else {
			g.sel = gr.tinies[0]
		}
		return true
	}
	return false
}

// Media ventana de la vista micro (16 px por cm en un lienzo de 480 × 270,
// sin las barras): el grupo de tinies que entra en ella es el que se encuadra.
const (
	clusterHalfW = 12 * units.Centimeter
	clusterHalfH = 6 * units.Centimeter
)

// tinyCluster busca, entre los tinies de la superficie s, el grupo más
// numeroso que entra en una vista micro y devuelve el centro de su extensión y
// cuántos tiene (0 si no hay tinies). A igual cantidad gana el primero en el
// orden del Frame.
func tinyCluster(f *view.Frame, s ids.ID) (units.Vec2, int) {
	sp, hasSp := f.SpaceByID(s)
	var pts []units.Vec2
	for _, t := range f.Things {
		if isTiny(t) && t.Space == s {
			pos := t.Pose.Pos
			if hasSp && sp.UsesMicroLocal() {
				pos = sp.MicroPos(t.Pose.Pos, t.LocalPos)
			}
			pts = append(pts, pos.XY())
		}
	}
	var best units.Vec2
	most := 0
	for _, p := range pts {
		lo, hi := p, p
		n := 0
		for _, q := range pts {
			d := q.Sub(p)
			if max(d.X, -d.X) <= clusterHalfW && max(d.Y, -d.Y) <= clusterHalfH {
				lo = units.Vec2{X: min(lo.X, q.X), Y: min(lo.Y, q.Y)}
				hi = units.Vec2{X: max(hi.X, q.X), Y: max(hi.Y, q.Y)}
				n++
			}
		}
		if n > most {
			// El centro de la extensión: el grupo mide a lo sumo el doble de
			// la media ventana, así que entra entero.
			most, best = n, units.Vec2{X: (lo.X + hi.X) / 2, Y: (lo.Y + hi.Y) / 2}
		}
	}
	return best, most
}

// busiestSurface es la superficie con vista de cerca donde hay más tinies
// (a igual cantidad, la de menor ID); Nil si no hay tinies.
func (g *game) busiestSurface(f *view.Frame) ids.ID {
	count := map[ids.ID]int{}
	for _, t := range f.Things {
		if isTiny(t) {
			if s := g.microSurface(f, t.ID); !s.IsNil() {
				count[s]++
			}
		}
	}
	best := ids.Nil
	for _, s := range slices.Sorted(maps.Keys(count)) {
		if best.IsNil() || count[s] > count[best] {
			best = s
		}
	}
	return best
}

// focusTinies, recién entrados a la vista de cerca con Tab, encuadra lo que
// importa en vez del encuadre fijo de la superficie: el grupo más numeroso de
// tinies que hay en ella o, si no hay, el punto donde se hizo clic para
// elegirla (un clic en una habitación en macro elige su piso). No cambia nada
// si se sigue a alguien o si lo elegido es un tiny de esa superficie (ya quedó
// encuadrado).
func (g *game) focusTinies() {
	a, f := g.art, g.run.Cur
	if a.focus.IsNil() || !g.follow.IsNil() {
		return
	}
	if t, ok := f.ThingByID(g.sel); ok && t.Space == a.focus {
		return
	}
	s, ok := f.SpaceByID(a.focus)
	if !ok {
		return
	}
	c, n := tinyCluster(f, s.ID)
	if n == 0 {
		if a.clickOn != s.ID {
			return
		}
		c = a.clickAt
	}
	g.cam.X = float64(c.X - s.Pose.Pos.X)
	g.cam.Y = float64(c.Y - s.Pose.Pos.Y)
}
