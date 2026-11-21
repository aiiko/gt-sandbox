// Package render dibuja un view.Frame con Ebitengine: formas y colores, sin
// arte. Solo lee el Frame; nunca toca la simulación.
package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/view"
)

// Layers indica qué overlays se dibujan.
type Layers struct {
	Grids, Portals, Contacts, Routes, Labels bool
}

// AllLayers tiene todos los overlays prendidos.
var AllLayers = Layers{true, true, true, true, true}

// MaxThingLabels: con más entidades que esto, solo se rotula la elegida.
const MaxThingLabels = 40

// Scene es todo lo que hace falta para dibujar un frame.
type Scene struct {
	Frame    *view.Frame
	Proj     camera.Projector
	Layers   Layers
	Selected ids.ID
}

// Colores.
var (
	Background = color.NRGBA{0x1b, 0x1d, 0x22, 0xff}
	kindFill   = map[space.Kind]color.NRGBA{
		space.Static:      {0x3a, 0x40, 0x4c, 0xff},
		space.Elevated:    {0x5b, 0x4a, 0x36, 0xff},
		space.Body:        {0x6a, 0x3f, 0x5c, 0xff},
		space.Articulated: {0x9a, 0x5a, 0x7a, 0xff},
	}
	outline     = color.NRGBA{0xd0, 0xd4, 0xdc, 0x80}
	blocked     = color.NRGBA{0x10, 0x10, 0x12, 0xe0}
	slow        = color.NRGBA{0xc8, 0x8a, 0x30, 0x50}
	gridLine    = color.NRGBA{0xff, 0xff, 0xff, 0x10}
	portalStat  = color.NRGBA{0x40, 0xd0, 0xe0, 0xff}
	portalDyn   = color.NRGBA{0xf0, 0x50, 0xd0, 0xff}
	portalOff   = color.NRGBA{0x80, 0x80, 0x80, 0xff}
	contactLine = color.NRGBA{0xf0, 0x50, 0xd0, 0x60}
	routeLine   = color.NRGBA{0xf0, 0xd0, 0x40, 0xc0}
	thingFill   = color.NRGBA{0x70, 0xe0, 0x80, 0xff}
	thingMove   = color.NRGBA{0xf0, 0xd0, 0x40, 0xff}
	thingPlan   = color.NRGBA{0x60, 0xa0, 0xff, 0xff}
	selected    = color.NRGBA{0xff, 0xff, 0xff, 0xff}
	frameMark   = color.NRGBA{0xe0, 0x90, 0xc0, 0xff}
)

// Draw dibuja la escena en dst.
func Draw(dst *ebiten.Image, s Scene) {
	dst.Fill(Background)
	f, p := s.Frame, s.Proj
	for _, sp := range f.Spaces {
		drawSpace(dst, p, sp, s.Layers.Grids, sp.ID == s.Selected)
	}
	if s.Layers.Contacts {
		for _, c := range f.Contacts {
			a, okA := f.SpaceByID(c.A)
			b, okB := f.SpaceByID(c.B)
			if okA && okB {
				ax, ay := p.ToScreen(center(*a))
				bx, by := p.ToScreen(center(*b))
				vector.StrokeLine(dst, ax, ay, bx, by, 3, contactLine, true)
			}
		}
	}
	if s.Layers.Portals {
		for _, pt := range f.Portals {
			c := portalStat
			if pt.Disabled {
				c = portalOff
			} else if pt.Dynamic {
				c = portalDyn
			}
			ax, ay := p.ToScreen(pt.A)
			bx, by := p.ToScreen(pt.B)
			vector.StrokeLine(dst, ax, ay, bx, by, 1.5, c, true)
			vector.FillCircle(dst, ax, ay, 3, c, true)
			vector.FillCircle(dst, bx, by, 3, c, true)
		}
	}
	for _, t := range f.Things {
		if s.Layers.Routes && t.Mover != nil && len(t.Mover.Route) > 1 {
			polyline(dst, p, t.Mover.Route, 1.5, routeLine)
		}
	}
	for _, t := range f.Things {
		drawThing(dst, p, t, t.ID == s.Selected)
	}
	if s.Layers.Labels {
		for _, sp := range f.Spaces {
			if sp.Name != "" {
				x, y := p.ToScreen(labelPoint(sp))
				ebitenutil.DebugPrintAt(dst, sp.Name, int(x)+3, int(y)-16)
			}
		}
		crowded := len(f.Things) > MaxThingLabels
		for _, t := range f.Things {
			if t.Name != "" && (!crowded || t.ID == s.Selected) {
				x, y := p.ToScreen(t.Pose.Pos)
				ebitenutil.DebugPrintAt(dst, t.Name, int(x)+6, int(y)+2)
			}
		}
	}
}

func center(s view.Space) units.Vec3 {
	e := s.Extent
	return s.Pose.Transform().Apply(units.V3((e.Min.X+e.Max.X)/2, (e.Min.Y+e.Max.Y)/2, 0))
}

func labelPoint(s view.Space) units.Vec3 {
	if s.Extent.Empty() {
		return s.Pose.Pos
	}
	return view.Corners(s)[3]
}

func drawSpace(dst *ebiten.Image, p camera.Projector, s view.Space, grids, sel bool) {
	if s.Extent.Empty() { // marco sin superficie (el cuerpo de la giantess)
		x, y := p.ToScreen(s.Pose.Pos)
		c := frameMark
		if sel {
			c = selected
		}
		vector.StrokeLine(dst, x-6, y, x+6, y, 1.5, c, true)
		vector.StrokeLine(dst, x, y-6, x, y+6, 1.5, c, true)
		a := p.ScreenAngle(s.Pose.Rot)
		vector.StrokeLine(dst, x, y, x+float32(12*math.Cos(a)), y+float32(12*math.Sin(a)), 1.5, c, true)
		return
	}
	var path vector.Path
	quad(&path, p, view.Corners(s))
	fill(dst, &path, kindFill[s.Kind])
	if grids && s.Grid != nil {
		drawGrid(dst, p, s)
	}
	c, w := outline, float32(1)
	if sel {
		c, w = selected, 2
	} else if s.Surface {
		c = portalDyn
	}
	stroke(dst, &path, w, c)
}

func drawGrid(dst *ebiten.Image, p camera.Projector, s view.Space) {
	g, t := s.Grid, s.Pose.Transform()
	px := p.Pixels(g.Cell)
	if px < 1 {
		return
	}
	cell := func(x, y int) [4]units.Vec3 {
		x0, y0 := g.Origin.X+units.Length(x)*g.Cell, g.Origin.Y+units.Length(y)*g.Cell
		x1, y1 := x0+g.Cell, y0+g.Cell
		return [4]units.Vec3{t.Apply(units.V3(x0, y0, 0)), t.Apply(units.V3(x1, y0, 0)),
			t.Apply(units.V3(x1, y1, 0)), t.Apply(units.V3(x0, y1, 0))}
	}
	var hard, soft vector.Path
	for y := range g.H {
		for x := range g.W {
			switch c := g.Cost[y*g.W+x]; {
			case c == nav.Blocked:
				quad(&hard, p, cell(x, y))
			case c > nav.Free:
				quad(&soft, p, cell(x, y))
			}
		}
	}
	fill(dst, &hard, blocked)
	fill(dst, &soft, slow)
	if px < 10 {
		return
	}
	var lines vector.Path
	at := func(x, y int) (float32, float32) {
		return p.ToScreen(t.Apply(units.V3(g.Origin.X+units.Length(x)*g.Cell, g.Origin.Y+units.Length(y)*g.Cell, 0)))
	}
	for x := 0; x <= g.W; x++ {
		lines.MoveTo(at(x, 0))
		lines.LineTo(at(x, g.H))
	}
	for y := 0; y <= g.H; y++ {
		lines.MoveTo(at(0, y))
		lines.LineTo(at(g.W, y))
	}
	stroke(dst, &lines, 1, gridLine)
}

func drawThing(dst *ebiten.Image, p camera.Projector, t view.Thing, sel bool) {
	x, y := p.ToScreen(t.Pose.Pos)
	r := max(p.Pixels(t.Radius), 3)
	c := thingFill
	if t.Mover != nil {
		switch t.Mover.Status {
		case "moving":
			c = thingMove
		case "planning":
			c = thingPlan
		}
	}
	vector.FillCircle(dst, x, y, r, c, true)
	a := p.ScreenAngle(t.Pose.Rot)
	vector.StrokeLine(dst, x, y, x+r*float32(math.Cos(a)), y+r*float32(math.Sin(a)), 1, Background, true)
	if sel {
		vector.StrokeCircle(dst, x, y, r+3, 1.5, selected, true)
	}
}

func quad(path *vector.Path, p camera.Projector, c [4]units.Vec3) {
	path.MoveTo(p.ToScreen(c[0]))
	for _, v := range c[1:] {
		path.LineTo(p.ToScreen(v))
	}
	path.Close()
}

func polyline(dst *ebiten.Image, p camera.Projector, pts []units.Vec3, w float32, c color.Color) {
	var path vector.Path
	path.MoveTo(p.ToScreen(pts[0]))
	for _, v := range pts[1:] {
		path.LineTo(p.ToScreen(v))
	}
	stroke(dst, &path, w, c)
}

func opts(c color.Color) *vector.DrawPathOptions {
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	return op
}

func fill(dst *ebiten.Image, path *vector.Path, c color.Color) {
	vector.FillPath(dst, path, nil, opts(c))
}

func stroke(dst *ebiten.Image, path *vector.Path, w float32, c color.Color) {
	vector.StrokePath(dst, path, &vector.StrokeOptions{Width: w, LineJoin: vector.LineJoinRound}, opts(c))
}
