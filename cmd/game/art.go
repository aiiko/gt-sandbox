package main

import (
	"fmt"
	"image/color"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/render"
	"gtsandbox/internal/view"
)

const artScenario = "tests/scenarios/muestra_pixel_art.json"

const (
	artHeader = 23
	artFooter = 22
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
}

func (g *game) enableArt(path string) error {
	if len(g.visuals) == 0 {
		return fmt.Errorf("el escenario no contiene entidades con visual")
	}
	atlas, err := render.LoadAtlas(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
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
	if !g.follow.IsNil() {
		g.cam.X, g.cam.Y = 0, 0
		return
	}
	f := g.run.Cur
	if s, ok := f.SpaceByID(g.art.focus); ok {
		g.cam.Anchor = s.ID
		// El ancla sigue la traslación del Space con el norte arriba.
		center := s.Pose.Transform().Apply(units.V3((s.Extent.Min.X+s.Extent.Max.X)/2, (s.Extent.Min.Y+s.Extent.Max.Y)/2, 0))
		g.cam.X, g.cam.Y = float64(center.X-s.Pose.Pos.X), float64(center.Y-s.Pose.Pos.Y)
		g.cam.Zoom = 16.0 / float64(units.Centimeter)
		return
	}
	g.cam.Anchor = ids.Nil
	g.cam.Fit(f.Bounds(), g.w, g.h)
	g.cam.Zoom = 16.0 / float64(40*units.Centimeter)
}

func (g *game) toggleArtMicro() {
	if !g.art.focus.IsNil() {
		g.art.focus = ids.Nil
		g.stopFollowing()
		g.fitArt()
		return
	}
	target := g.sel
	f := g.run.Cur
	if t, ok := f.ThingByID(target); ok {
		target = t.Space
	}
	if s, ok := f.SpaceByID(target); !ok || s.Parent.IsNil() || s.Extent.Empty() {
		target = g.artDesk
	}
	if s, ok := f.SpaceByID(target); ok {
		g.stopFollowing()
		g.art.focus = s.ID
		g.fitArt()
	}
}

func (g *game) artMouse() {
	a := g.art
	x, y := a.defaultTransform.Apply(ebiten.CursorPositionF())
	pixel, ok := a.viewport.Point(x, y)
	if !ok || pixel.Y < artHeader || pixel.Y >= render.CanvasHeight-artFooter {
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.sel = a.last.Pick(pixel)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.moveSelected(a.proj.ToWorld(float64(pixel.X)+0.5, float64(pixel.Y)+0.5))
	}
}

func (g *game) drawArt(dst *ebiten.Image, f *view.Frame, proj camera.Projector) {
	a := g.art
	scale := render.Macro
	if !a.focus.IsNil() {
		scale = render.Micro
	}
	a.last, a.proj = a.atlas.Arrange(f, proj, g.visuals, scale, a.focus), proj
	a.last.Draw(dst, g.sel)

	ink := color.NRGBA{R: 25, G: 30, B: 38, A: 255}
	vector.FillRect(dst, 0, 0, render.CanvasWidth, artHeader, ink, false)
	vector.FillRect(dst, 0, render.CanvasHeight-artFooter, render.CanvasWidth, artFooter, ink, false)
	ebitenutil.DebugPrintAt(dst, "PRUEBA DE ARTE / "+string(scale), 10, 4)
	if name := g.name(g.sel); !g.sel.IsNil() {
		ebitenutil.DebugPrintAt(dst, name, 250, 4)
	}
	ebitenutil.DebugPrintAt(dst, "Tab  Escala     Clic  Elegir     T  Seguir     F  Centrar", 10, render.CanvasHeight-19)
}
