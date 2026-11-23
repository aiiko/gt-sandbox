package main

import (
	"fmt"
	"image"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/render"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/ui"
	"gtsandbox/internal/view"
)

const doubleClickWindow = 350 * time.Millisecond

// selectOrFollow elige id y, si se hizo doble clic sobre el mismo id en
// doubleClickWindow, activa el seguimiento (propuesta, 17.7).
func (g *game) selectOrFollow(id ids.ID) {
	now := time.Now()
	a := g.art
	isDouble := !id.IsNil() && id == a.lastClickID && now.Sub(a.lastClickTime) <= doubleClickWindow
	a.lastClickID, a.lastClickTime = id, now
	g.sel = id
	if isDouble && g.follow != id {
		g.stopFollowing()
		g.toggleFollow()
	}
}

// threatenedSurfaces devuelve las superficies que tienen una amenaza
// inminente: una pisada prevista de quien camina o una superficie móvil
// suspendida encima a menos de render.ShadowReach (propuesta, 17.7).
func threatenedSurfaces(f *view.Frame) map[ids.ID]bool {
	out := map[ids.ID]bool{}
	for _, fp := range f.Footprints {
		if w, ok := f.SpaceByID(fp.Walker); ok && w.Mover != nil && w.Mover.Status == "moving" {
			out[fp.Space] = true
		}
	}
	for _, s := range f.Spaces {
		if s.Mover != nil || s.Extent.Empty() || s.Kind != space.Articulated {
			continue
		}
		for _, fs := range f.Spaces {
			if fs.ID == s.ID || fs.Extent.Empty() {
				continue
			}
			dz := s.Pose.Pos.Z - fs.Pose.Pos.Z
			if dz <= 0 || dz >= render.ShadowReach {
				continue
			}
			local := fs.Pose.Transform().ApplyInv(s.Pose.Pos).XY()
			margin := max(s.Extent.Max.X-s.Extent.Min.X, s.Extent.Max.Y-s.Extent.Min.Y) / 2
			r := fs.Extent
			if local.X >= r.Min.X-margin && local.X <= r.Max.X+margin && local.Y >= r.Min.Y-margin && local.Y <= r.Max.Y+margin {
				out[fs.ID] = true
			}
		}
	}
	return out
}

func badgeActivityOf(cat string) ui.BadgeActivity {
	switch cat {
	case "Huye":
		return ui.BadgeFlee
	case "Acarrea":
		return ui.BadgeCarry
	case "Come":
		return ui.BadgeEat
	case "Descansa":
		return ui.BadgeRest
	case "Conversa":
		return ui.BadgeChat
	case "Busca", "Explora":
		return ui.BadgeSearch
	case "Fibras", "Refugio":
		return ui.BadgeWork
	}
	return ui.BadgeIdle
}

// hoverTooltip arma la información al pasar el cursor (propuesta, 17.7).
func (g *game) hoverTooltip(f *view.Frame, focus ids.ID) *ui.Tooltip {
	a := g.art
	if a == nil || !a.hoverOK || g.pause.open || g.help || g.panelResize != 0 || g.shot != "" {
		return nil
	}
	p := a.hoverPixel
	lay := a.layout
	if p.In(lay.PortraitRect) {
		if gi := g.giantessID(); !gi.IsNil() {
			if s, ok := f.SpaceByID(gi); ok {
				lines := []string{}
				if act := activityText(s.Activity, s.Step, ""); act != "" {
					lines = append(lines, act)
				}
				if dest := g.destName(f, s.Goal); dest != "" {
					lines = append(lines, "Destino: "+dest)
				}
				return &ui.Tooltip{Pos: p, Title: "Giantess", Lines: lines}
			}
		}
		return nil
	}
	if p.Y < lay.Top || p.Y >= lay.Foot || p.In(lay.PanelRect) || p.In(lay.CardRect) {
		return nil
	}
	for _, r := range lay.Notices {
		if p.In(r) {
			return nil
		}
	}
	// En macro, sobre un globo de tinies: lista de quiénes están ahí y qué hacen.
	if focus.IsNil() {
		for _, gr := range a.badges {
			if !p.In(gr.badge) {
				continue
			}
			title := fmt.Sprintf("Tinies en %s (%d)", g.display(gr.surface), len(gr.tinies))
			var lines []string
			for i, id := range gr.tinies {
				if i >= 5 {
					lines = append(lines, fmt.Sprintf("+%d más", len(gr.tinies)-i))
					break
				}
				line := capital(g.display(id))
				if t, ok := f.ThingByID(id); ok {
					line += " — " + shortActivity(*t)
				}
				lines = append(lines, line)
			}
			return &ui.Tooltip{Pos: p, Title: title, Lines: lines}
		}
	}
	if !a.hoverOnScene {
		return nil
	}
	id := a.last.Pick(a.hoverAt)
	if id.IsNil() {
		return nil
	}
	title := capital(g.display(id))
	var lines []string
	if t, ok := f.ThingByID(id); ok {
		carrying := ""
		for _, o := range f.Things {
			if o.CarriedBy == id {
				carrying = g.display(o.ID)
			}
		}
		act := activityText(t.Activity, t.Step, carrying)
		if lt := g.lifeText(*t, carrying); lt != "" {
			act = lt
		}
		if t.Remains != "" {
			act = remainsCause[t.Remains]
		}
		if act != "" {
			lines = append(lines, act)
		}
		for _, col := range f.Colonies {
			if col.ID == id {
				lines = append(lines, fmt.Sprintf("Despensa: %d/%d", col.Portions, col.Goal))
			}
		}
	} else if s, ok := f.SpaceByID(id); ok {
		if act := activityText(s.Activity, s.Step, ""); act != "" {
			lines = append(lines, act)
		}
	}
	return &ui.Tooltip{Pos: p, Title: title, Lines: lines}
}

// minimap arma el miniplano de la casa para la vista micro (propuesta, 17.7).
func (g *game) minimap(f *view.Frame, focus ids.ID) *ui.Minimap {
	if focus.IsNil() {
		return nil
	}
	b := f.Bounds()
	bw, bh := float64(b.Max.X-b.Min.X), float64(b.Max.Y-b.Min.Y)
	if bw <= 0 || bh <= 0 {
		return nil
	}
	norm := func(x, y units.Length) (float32, float32) {
		nx := float32((float64(x - b.Min.X)) / bw)
		ny := float32((float64(y - b.Min.Y)) / bh)
		return max(0, min(1, nx)), max(0, min(1, ny))
	}
	activeRoom := ids.Nil
	if c := chain(f, focus); len(c) > 0 {
		activeRoom = c[0]
	}
	mm := &ui.Minimap{}
	for _, s := range f.Spaces {
		if s.Kind != space.Static || s.Extent.Empty() || g.visuals[s.ID] == "" {
			continue
		}
		corners := view.Corners(s)
		minX, maxX := corners[0].X, corners[0].X
		minY, maxY := corners[0].Y, corners[0].Y
		for _, c := range corners[1:] {
			minX, maxX = min(minX, c.X), max(maxX, c.X)
			minY, maxY = min(minY, c.Y), max(maxY, c.Y)
		}
		x0, y0 := norm(minX, minY)
		x1, y1 := norm(maxX, maxY)
		mm.Rooms = append(mm.Rooms, ui.MinimapRoom{X0: x0, Y0: y0, X1: x1, Y1: y1, Active: s.ID == activeRoom})
	}
	if len(mm.Rooms) == 0 {
		return nil
	}
	for _, mk := range f.Marks {
		mx, my := norm(mk.Pos.X, mk.Pos.Y)
		mm.Pins = append(mm.Pins, ui.MinimapPin{X: mx, Y: my, Confirmed: mk.Confirmed})
	}
	if sp, ok := f.SpaceByID(focus); ok {
		mm.HasFocus = true
		mm.FX, mm.FY = norm(sp.Pose.Pos.X, sp.Pose.Pos.Y)
	}
	if gi := g.giantessID(); !gi.IsNil() {
		if gs, ok := f.SpaceByID(gi); ok {
			mm.HasGiantess = true
			mm.GX, mm.GY = norm(gs.Pose.Pos.X, gs.Pose.Pos.Y)
		}
	}
	return mm
}

// roomLabels arma los rótulos de las habitaciones en macro para pasarlos a HUD.
func (g *game) roomLabels(f *view.Frame, proj camera.Projector) []ui.RoomLabel {
	var out []ui.RoomLabel
	for _, room := range f.Spaces {
		if room.Kind != space.Static || g.visuals[room.ID] == "" || room.Name == "" {
			continue
		}
		c := view.Corners(room)
		south := c[0].Add(c[1].Sub(c[0]).MulDiv(1, 2))
		x, y := proj.ToScreen(south)
		out = append(out, ui.RoomLabel{
			Text: capital(room.Name),
			At:   image.Pt(int(math.Round(float64(x))), int(math.Round(float64(y)))+8),
		})
	}
	return out
}

// drawTabHighlight resalta en macro la superficie a la que lleva Tab al
// pasar el cursor por la etiqueta de Tab en la barra inferior (17.7).
func (g *game) drawTabHighlight(dst *ebiten.Image, f *view.Frame, proj camera.Projector) {
	if g.art == nil || !g.art.tabHover {
		return
	}
	target := g.microTarget(f)
	s, ok := f.SpaceByID(target)
	if !ok || s.Extent.Empty() {
		return
	}
	corners := view.Corners(*s)
	pts := make([]image.Point, len(corners))
	for i, c := range corners {
		x, y := proj.ToScreen(c)
		pts[i] = image.Pt(int(math.Round(float64(x))), int(math.Round(float64(y))))
	}
	ui.DrawOutline(dst, pts, ui.Turquoise)
}
