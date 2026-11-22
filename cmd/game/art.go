package main

import (
	"fmt"
	"image"
	"image/color"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/render"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/ui"
	"gtsandbox/internal/view"
)

const artScenario = "tests/scenarios/muestra_pixel_art.json"

const (
	artHeader       = 23
	artFooter       = 22
	artFadeDuration = 180 * time.Millisecond
)

// pixelGame reemplaza la presentación final del visor para el modo de arte.
type pixelGame struct{ *game }

type artMode struct {
	atlas            *render.Atlas
	focus            ids.ID
	last             render.ArtScene
	proj             camera.Projector
	viewport         render.Viewport
	defaultTransform ebiten.GeoM // Ebitengine usa esta transformación para el cursor
	macro            camera.Camera
	cut              bool
	fade             time.Duration
	pendingFade      bool
	presented        *ebiten.Image
	fadeFrom         *ebiten.Image
	layout           ui.Layout   // zonas de la interfaz de partida del último frame
	debug            bool        // F3: overlays técnicos sobre el arte
	badges           []tinyGroup // globos de tinies del último frame macro
	clickOn          ids.ID      // lo que eligió el último clic sobre la escena
	clickAt          units.Vec2  // y dónde, en el mundo
}

func (g *game) enableArt(path string) error {
	if len(g.visuals) == 0 {
		return fmt.Errorf("el escenario no contiene entidades con visual")
	}
	if path == "" {
		path = g.artAtlas
		if path == "" {
			path = "content/art/muestra.json"
		}
	}
	atlas, err := render.LoadAtlas(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if g.micro != "" && !atlas.Supports(render.Micro) {
		return fmt.Errorf("atlas %s: el escenario no ofrece vista micro", path)
	}
	for _, id := range slices.Sorted(maps.Keys(g.visuals)) {
		if err := atlas.Require(g.visuals[id]); err != nil {
			return fmt.Errorf("entidad %s: %w", id, err)
		}
	}
	g.art = &artMode{atlas: atlas}
	return nil
}

func (pixelGame) Layout(_, _ int) (int, int) {
	return render.CanvasWidth, render.CanvasHeight
}

func (g pixelGame) DrawFinalScreen(screen ebiten.FinalScreen, canvas *ebiten.Image, defaultTransform ebiten.GeoM) {
	a := g.art
	a.defaultTransform = defaultTransform
	a.viewport = render.PixelViewport(screen.Bounds())
	a.viewport.Blit(screen, canvas)
	if g.wantsShot() {
		bounds := screen.Bounds()
		shot := ebiten.NewImageWithOptions(bounds, nil)
		defer shot.Deallocate()
		a.viewport.Blit(shot, canvas)
		g.shotDone, g.shotErr = true, savePNG(shot, g.shot)
	}
}

func (g *game) fitArt() {
	f := g.run.Cur
	focus := g.artFocus(f)
	g.cam.Zoom = 16.0 / float64(40*units.Centimeter)
	if !focus.IsNil() {
		g.cam.Zoom = 16.0 / float64(units.Centimeter)
	}
	if g.followsInView(f) {
		g.cam.Anchor = ids.Nil
		g.cam.X, g.cam.Y = 0, 0
		return
	}
	if s, ok := f.SpaceByID(focus); ok {
		g.cam.Anchor = s.ID
		// El ancla sigue la traslación del Space con el norte arriba.
		center := s.Pose.Transform().Apply(units.V3((s.Extent.Min.X+s.Extent.Max.X)/2, (s.Extent.Min.Y+s.Extent.Max.Y)/2, 0))
		g.cam.X, g.cam.Y = float64(center.X-s.Pose.Pos.X), float64(center.Y-s.Pose.Pos.Y)
		if offset, ok := g.art.atlas.MicroFocus(g.visuals[s.ID]); ok {
			g.cam.X = float64(offset.X) / g.cam.Zoom
			g.cam.Y = -float64(offset.Y) / g.cam.Zoom
		}
		return
	}
	g.cam.Anchor = ids.Nil
	g.cam.Fit(f.Bounds(), g.w, g.h)
	g.cam.Zoom = 16.0 / float64(40*units.Centimeter)
}

func (g *game) microSurface(f *view.Frame, id ids.ID) ids.ID {
	if t, ok := f.ThingByID(id); ok {
		id = t.Space
	}
	if s, ok := f.SpaceByID(id); ok && !s.Extent.Empty() && g.art.atlas.Visible(g.visuals[id], render.Micro) {
		return id
	}
	return ids.Nil
}

func (g *game) artFocus(f *view.Frame) ids.ID {
	if g.art.focus.IsNil() {
		return ids.Nil
	}
	if id := g.microSurface(f, g.follow); !id.IsNil() {
		return id
	}
	return g.art.focus
}

func (g *game) followsInView(f *view.Frame) bool {
	return !g.follow.IsNil() && (g.art == nil || g.art.focus.IsNil() || !g.microSurface(f, g.follow).IsNil())
}

func (g *game) toggleArtMicro() {
	a := g.art
	if !a.atlas.Supports(render.Micro) {
		return
	}
	if !a.focus.IsNil() {
		a.focus = ids.Nil
		g.cam = a.macro
		if !g.follow.IsNil() {
			g.fitArt()
		}
		g.startArtFade()
		return
	}
	f := g.run.Cur
	if s, ok := f.SpaceByID(g.microTarget(f)); ok {
		a.macro, a.focus = g.cam, s.ID
		if !g.follow.IsNil() {
			a.macro.Follow(ids.Nil, fixedmath.Identity, g.proj)
		}
		g.fitArt()
		if t, ok := f.ThingByID(g.sel); ok && t.Space == s.ID && !g.followsInView(f) {
			g.cam.X = float64(t.Pose.Pos.X - s.Pose.Pos.X)
			g.cam.Y = float64(t.Pose.Pos.Y - s.Pose.Pos.Y)
		}
		g.startArtFade()
	}
}

// microTarget es la superficie a la que lleva Tab desde la vista general:
// la de lo seguido; si no, la de lo elegido (una superficie, como el piso de
// una habitación, es la suya); si no, la que tiene más tinies y, sin tinies,
// el escritorio. Nil si ninguna tiene vista de cerca.
func (g *game) microTarget(f *view.Frame) ids.ID {
	for _, id := range []ids.ID{g.follow, g.sel, g.busiestSurface(f), g.artDesk} {
		if s := g.microSurface(f, id); !s.IsNil() {
			return s
		}
	}
	return ids.Nil
}

// scaleLabel nombra a dónde lleva Tab: la superficie de cerca o la casa.
func (g *game) scaleLabel(f *view.Frame) string {
	if g.art == nil || !g.art.atlas.Supports(render.Micro) {
		return "Escala"
	}
	if !g.art.focus.IsNil() {
		return "Casa"
	}
	if t := g.microTarget(f); !t.IsNil() {
		return capital(g.display(t))
	}
	return "Escala"
}

func (g *game) startArtFade() {
	if g.fitted && !g.art.cut && g.shot == "" && g.art.presented != nil {
		g.art.fade, g.art.pendingFade = artFadeDuration, true
	}
}

func (a *artMode) blend(dst *ebiten.Image) {
	if a.presented == nil {
		a.presented = ebiten.NewImage(render.CanvasWidth, render.CanvasHeight)
	}
	if a.pendingFade {
		if a.fadeFrom == nil {
			a.fadeFrom = ebiten.NewImage(render.CanvasWidth, render.CanvasHeight)
		}
		a.fadeFrom.Clear()
		a.fadeFrom.DrawImage(a.presented, nil)
		a.pendingFade = false
	}
	if a.fade > 0 {
		op := &ebiten.DrawImageOptions{}
		op.ColorScale.ScaleAlpha(float32(a.fade) / float32(artFadeDuration))
		dst.DrawImage(a.fadeFrom, op)
	}
	a.presented.Clear()
	a.presented.DrawImage(dst, nil)
}

func (g *game) artMouse() {
	a := g.art
	if a.fade > 0 {
		return
	}
	x, y := a.defaultTransform.Apply(ebiten.CursorPositionF())
	pixel, ok := a.viewport.Point(x, y)
	top, foot := artHeader, render.CanvasHeight-artFooter
	if g.artAtlas != "" {
		top, foot = a.layout.Top, a.layout.Foot
	}
	if !ok || pixel.Y < top || pixel.Y >= foot {
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for i, r := range a.layout.Story {
			if pixel.In(r) {
				if entries := g.storyEntries(); i < len(entries) {
					g.focusStory(entries[i])
				}
				return
			}
		}
		if g.clickTinyBadge(pixel) {
			return
		}
		g.sel = a.last.Pick(pixel)
		a.clickOn, a.clickAt = g.sel, a.proj.ToWorld(float64(pixel.X)+0.5, float64(pixel.Y)+0.5)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.moveSelected(a.proj.ToWorld(float64(pixel.X)+0.5, float64(pixel.Y)+0.5))
	}
}

func (g *game) drawArt(dst *ebiten.Image, f *view.Frame, proj camera.Projector) {
	a := g.art
	if !a.cut && g.shot == "" {
		defer a.blend(dst)
	}
	focus := g.artFocus(f)
	scale := render.Macro
	if !focus.IsNil() {
		scale = render.Micro
	}
	a.last, a.proj = a.atlas.Arrange(f, proj, g.visuals, scale, focus, g.sel, g.follow), proj
	a.last.Draw(dst, g.sel)

	if house := g.artAtlas != ""; house {
		if focus.IsNil() {
			for _, room := range f.Spaces {
				if room.Kind != space.Static || g.visuals[room.ID] == "" || room.Name == "" {
					continue
				}
				c := view.Corners(room)
				south := c[0].Add(c[1].Sub(c[0]).MulDiv(1, 2))
				x, y := proj.ToScreen(south)
				label := capital(room.Name)
				w := ui.TextWidth(label)
				r := image.Rect(int(x)-w/2-4, int(y)+8, int(x)+w/2+4, int(y)+23)
				ui.Box(dst, r, ui.Ivory)
				ui.DrawText(dst, label, r.Min.X+4, r.Min.Y+1, ui.Plum)
			}
			a.badges = tinyGroups(f, proj)
			g.drawTinyBadges(dst, a.badges)
		} else {
			a.badges = nil
		}
		g.drawEmotes(dst, f, scale)
		a.layout = g.partida(f, focus).Draw(dst, a.atlas)
		if a.debug {
			g.drawDebug(dst, f, proj, a.layout.Top+1)
		}
		if g.confirmExit {
			drawExitPrompt(dst)
		}
		return
	}
	ink := color.NRGBA{R: 25, G: 30, B: 38, A: 255}
	vector.FillRect(dst, 0, 0, render.CanvasWidth, artHeader, ink, false)
	vector.FillRect(dst, 0, render.CanvasHeight-artFooter, render.CanvasWidth, artFooter, ink, false)
	ebitenutil.DebugPrintAt(dst, "PRUEBA DE ARTE / "+string(scale), 10, 4)
	if name := g.name(g.sel); !g.sel.IsNil() {
		ebitenutil.DebugPrintAt(dst, name, 250, 4)
	}
	ebitenutil.DebugPrintAt(dst, "Tab  Escala     Clic  Elegir     T  Seguir     F  Centrar", 10, render.CanvasHeight-19)
}

// drawExitPrompt pregunta, sobre la partida en pausa, si volver al menú.
func drawExitPrompt(dst *ebiten.Image) {
	const title, keys = "¿Volver al menú principal?", "Enter  Volver      Esc  Seguir jugando"
	w := max(ui.TextWidth(title), ui.TextWidth(keys)) + 20
	x, y := (render.CanvasWidth-w)/2, render.CanvasHeight/2-20
	ui.Box(dst, image.Rect(x, y, x+w, y+40), ui.Ivory)
	ui.DrawText(dst, title, (render.CanvasWidth-ui.TextWidth(title))/2, y+5, ui.Plum)
	ui.DrawText(dst, keys, (render.CanvasWidth-ui.TextWidth(keys))/2, y+21, ui.PlumSoft)
}
