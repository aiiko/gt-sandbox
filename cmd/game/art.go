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
	// Capa de escena (propuesta, 17.4): el arte se dibuja en su propia
	// imagen con un factor que cambia la rueda, y la interfaz se compone
	// encima con el factor del lienzo.
	factor    map[render.Scale]float64 // factor elegido con la rueda (sin elegir: el del lienzo)
	wheel     float64                  // giro de la rueda que todavía no llega a un paso
	scene     *ebiten.Image
	sceneView render.Viewport // dónde cae la capa de escena en la pantalla final
}

// zoomLevels son los factores de escena (píxeles físicos por píxel del arte)
// que recorre la rueda: enteros, para que el arte ampliado nunca se
// re-muestree, y ½, una reducción exacta que promedia. Más abajo la casa
// queda diminuta en 1080p y la capa reducida pesa demasiado (¼ en 4K serían
// 15360 × 8640 píxeles del arte).
var zoomLevels = []float64{0.5, 1, 2, 3, 4, 6, 8}

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

// Layout da el lienzo que llena la ventana (propuesta, 17.2).
func (pixelGame) Layout(w, h int) (int, int) { return canvasLayout(w, h) }

func (g pixelGame) DrawFinalScreen(screen ebiten.FinalScreen, canvas *ebiten.Image, defaultTransform ebiten.GeoM) {
	a := g.art
	a.defaultTransform = defaultTransform
	a.viewport = render.PixelViewport(screen.Bounds(), canvas.Bounds().Size())
	if g.wantsShot() {
		shot := ebiten.NewImageWithOptions(screen.Bounds(), nil)
		defer shot.Deallocate()
		a.compose(shot, canvas)
		g.shotDone, g.shotErr = true, savePNG(shot, g.shot)
	}
	if a.cut || g.shot != "" {
		a.compose(screen, canvas)
		return
	}
	a.blend(screen, canvas)
}

// compose arma la pantalla final: fondo, capa de escena con su factor y la
// interfaz encima con el del lienzo.
func (a *artMode) compose(dst interface {
	Fill(color.Color)
	DrawImage(*ebiten.Image, *ebiten.DrawImageOptions)
}, canvas *ebiten.Image) {
	dst.Fill(render.Backdrop)
	if a.scene != nil {
		a.sceneView.Draw(dst, a.scene)
	}
	a.viewport.Draw(dst, canvas)
}

// uiFactor es el factor del lienzo de la interfaz (1 antes del primer frame
// y en las pruebas, que no pasan por la pantalla final).
func (a *artMode) uiFactor() int { return max(1, a.viewport.Factor) }

// sceneFactor es el factor de la capa de escena de una escala: el elegido con
// la rueda o, si no se eligió, el del lienzo (un píxel del arte por píxel
// del lienzo).
func (a *artMode) sceneFactor(scale render.Scale) float64 {
	if n := a.factor[scale]; n > 0 {
		return n
	}
	return float64(a.uiFactor())
}

// artScale es la escala que se ve: micro con una superficie en foco.
func (g *game) artScale(f *view.Frame) render.Scale {
	if g.artFocus(f).IsNil() {
		return render.Macro
	}
	return render.Micro
}

// sceneZoom es el zoom de la cámara (píxeles del lienzo por µm) con el que la
// capa de escena muestra el arte de la escala a su factor.
func (g *game) sceneZoom(scale render.Scale) float64 {
	return g.artZoom(scale) * g.art.sceneFactor(scale) / float64(g.art.uiFactor())
}

// sceneLayer es dónde cae la capa de escena de la escala en la pantalla. Sin
// pantalla final todavía (y en las pruebas) es el lienzo a 1×.
func (g *game) sceneLayer(scale render.Scale) render.Viewport {
	v := g.art.viewport
	if v.Factor == 0 {
		v = render.Viewport{Rect: image.Rect(0, 0, g.w, g.h), Factor: 1}
	}
	return v.Scene(g.art.sceneFactor(scale))
}

// wheelZoom cambia el factor de escena de a un nivel por muesca de la rueda,
// con el punto del lienzo bajo el cursor fijo (el centro si se sigue a
// alguien).
func (g *game) wheelZoom(wy float64, cursor [2]float64) {
	a := g.art
	a.wheel += wy
	steps := int(a.wheel)
	if steps == 0 {
		return
	}
	a.wheel -= float64(steps)
	f := g.run.Cur
	scale := g.artScale(f)
	old := a.sceneFactor(scale)
	i, found := slices.BinarySearch(zoomLevels, old)
	if !found && steps > 0 {
		i-- // un factor fuera de la lista: el primer paso cae en su vecino
	}
	n := zoomLevels[min(max(i+steps, 0), len(zoomLevels)-1)]
	if n == old {
		a.wheel = 0
		return
	}
	if a.factor == nil {
		a.factor = map[render.Scale]float64{}
	}
	a.factor[scale] = n
	if g.followsInView(f) {
		cursor = [2]float64{float64(g.w) / 2, float64(g.h) / 2}
	}
	g.cam.ZoomAt(g.proj, cursor[0], cursor[1], n/old)
	g.cam.Zoom = g.sceneZoom(scale)
}

func (g *game) fitArt() {
	f := g.run.Cur
	focus := g.artFocus(f)
	g.cam.Zoom = g.sceneZoom(g.artScale(f))
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
			g.cam.X = float64(offset.X) / g.artZoom(render.Micro)
			g.cam.Y = -float64(offset.Y) / g.artZoom(render.Micro)
		}
		return
	}
	g.cam.Anchor = ids.Nil
	g.cam.Fit(f.Bounds(), g.w, g.h)
	g.cam.Zoom = g.sceneZoom(render.Macro)
}

// syncArtZoom ajusta el zoom de la cámara al factor de escena actual: cambia
// si la ventana cambió de factor de lienzo o si Tab volvió a una escala con
// otro factor. El centro de la vista no se mueve.
func (g *game) syncArtZoom(f *view.Frame) {
	if g.art != nil {
		g.cam.Zoom = g.sceneZoom(g.artScale(f))
	}
}

// artZoom es el zoom (px/µm) con el que un píxel del lienzo es un píxel del
// arte de la escala, según la densidad que declara el atlas.
func (g *game) artZoom(scale render.Scale) float64 {
	return float64(g.art.atlas.Density(scale)) / float64(units.Meter)
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

// artFocus es la superficie que se ve de cerca (Nil: la casa). De cerca, lo
// seguido manda: se ve su superficie y, mientras está en una sin vista de
// cerca (trepando una pata), la casa; al llegar a otra con vista de cerca se
// vuelve a ella.
func (g *game) artFocus(f *view.Frame) ids.ID {
	if g.art.focus.IsNil() {
		return ids.Nil
	}
	if _, ok := entityPose(f, g.follow); ok && !g.follow.IsNil() {
		return g.microSurface(f, g.follow)
	}
	return g.art.focus
}

// followsInView dice si la cámara acompaña a lo seguido: siempre que haya
// algo seguido, porque artFocus muestra la vista donde está.
func (g *game) followsInView(*view.Frame) bool { return !g.follow.IsNil() }

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

// blend presenta la pantalla compuesta con el fundido en curso encima.
// presented guarda lo último que se vio, fundido incluido, para que un cambio
// de escala en medio de otro funda desde lo que había en pantalla.
func (a *artMode) blend(screen ebiten.FinalScreen, canvas *ebiten.Image) {
	size := screen.Bounds().Size()
	if a.presented == nil || a.presented.Bounds().Size() != size {
		// la ventana cambió de tamaño: el fundido en curso ya no calza
		if a.presented != nil {
			a.presented.Deallocate()
		}
		a.presented, a.fadeFrom, a.fade = ebiten.NewImage(size.X, size.Y), nil, 0
	}
	if a.pendingFade {
		if a.fadeFrom == nil {
			a.fadeFrom = ebiten.NewImage(size.X, size.Y)
		}
		a.fadeFrom.Clear()
		a.fadeFrom.DrawImage(a.presented, nil)
		a.pendingFade = false
	}
	a.compose(a.presented, canvas)
	if a.fade > 0 {
		op := &ebiten.DrawImageOptions{}
		op.ColorScale.ScaleAlpha(float32(a.fade) / float32(artFadeDuration))
		a.presented.DrawImage(a.fadeFrom, op)
	}
	screen.DrawImage(a.presented, nil)
}

func (g *game) artMouse() {
	a := g.art
	if a.fade > 0 {
		return
	}
	x, y := a.defaultTransform.Apply(ebiten.CursorPositionF())
	pixel, ok := a.viewport.Point(x, y)
	top, foot := artHeader, g.h-artFooter
	if g.artAtlas != "" {
		top, foot = a.layout.Top, a.layout.Foot
	}
	if !ok || pixel.Y < top || pixel.Y >= foot {
		return
	}
	if _, wy := ebiten.Wheel(); wy != 0 {
		v := a.viewport
		g.wheelZoom(wy, [2]float64{(x - float64(v.Rect.Min.X)) / float64(v.Factor), (y - float64(v.Rect.Min.Y)) / float64(v.Factor)})
	}
	// La escena se elige en sus propios píxeles; la interfaz, en los del lienzo.
	at, onScene := a.sceneView.Point(x, y)
	world := a.proj.ToWorld(float64(at.X)+0.5, float64(at.Y)+0.5)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for i, r := range a.layout.NoticeClose {
			if pixel.In(r) {
				g.dismissNotice(i)
				return
			}
		}
		for i, r := range a.layout.Notices {
			if pixel.In(r) {
				if i < len(g.notices) {
					g.focusStory(g.notices[i].entry)
				}
				return
			}
		}
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
		if !onScene {
			return
		}
		g.sel = a.last.Pick(at)
		a.clickOn, a.clickAt = g.sel, world
	}
	if onScene && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.moveSelected(world)
	}
}

func (g *game) drawArt(dst *ebiten.Image, f *view.Frame, proj camera.Projector) {
	a := g.art
	focus := g.artFocus(f)
	scale := g.artScale(f)
	a.sceneView = g.sceneLayer(scale)
	size := a.sceneView.Size()
	if a.scene == nil || a.scene.Bounds().Size() != size {
		if a.scene != nil {
			a.scene.Deallocate()
		}
		a.scene = ebiten.NewImage(size.X, size.Y)
	}
	sp := proj.Resized(g.artZoom(scale), size.X, size.Y)
	a.last, a.proj = a.atlas.Arrange(f, sp, g.visuals, scale, focus, g.sel, g.follow), sp
	a.last.Draw(a.scene, g.sel)

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
			a.badges = tinyGroups(f, proj, g.visuals)
			g.drawTinyBadges(dst, a.badges)
		} else {
			a.badges = nil
		}
		g.drawEmotes(a.scene, f, scale)
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
	W, H := dst.Bounds().Dx(), dst.Bounds().Dy()
	vector.FillRect(dst, 0, 0, float32(W), artHeader, ink, false)
	vector.FillRect(dst, 0, float32(H-artFooter), float32(W), artFooter, ink, false)
	ebitenutil.DebugPrintAt(dst, "PRUEBA DE ARTE / "+string(scale), 10, 4)
	if name := g.name(g.sel); !g.sel.IsNil() {
		ebitenutil.DebugPrintAt(dst, name, 250, 4)
	}
	ebitenutil.DebugPrintAt(dst, "Tab  Escala     Clic  Elegir     T  Seguir     F  Centrar", 10, H-19)
}

// drawExitPrompt pregunta, sobre la partida en pausa, si volver al menú.
func drawExitPrompt(dst *ebiten.Image) {
	const title, keys = "¿Volver al menú principal?", "Enter  Volver      Esc  Seguir jugando"
	W, H := dst.Bounds().Dx(), dst.Bounds().Dy()
	w := max(ui.TextWidth(title), ui.TextWidth(keys)) + 20
	x, y := (W-w)/2, H/2-20
	ui.Box(dst, image.Rect(x, y, x+w, y+40), ui.Ivory)
	ui.DrawText(dst, title, (W-ui.TextWidth(title))/2, y+5, ui.Plum)
	ui.DrawText(dst, keys, (W-ui.TextWidth(keys))/2, y+21, ui.PlumSoft)
}
