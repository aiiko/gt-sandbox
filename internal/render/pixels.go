package render

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/view"
)

const (
	CanvasWidth  = 480
	CanvasHeight = 270
)

// Viewport es el rectángulo del lienzo en píxeles físicos de la pantalla final.
type Viewport struct {
	Rect   image.Rectangle
	Factor int
}

func PixelViewport(screen image.Rectangle) Viewport {
	factor := min(screen.Dx()/CanvasWidth, screen.Dy()/CanvasHeight)
	if factor < 1 {
		return Viewport{}
	}
	w, h := CanvasWidth*factor, CanvasHeight*factor
	origin := screen.Min.Add(image.Pt((screen.Dx()-w)/2, (screen.Dy()-h)/2))
	return Viewport{Rect: image.Rectangle{Min: origin, Max: origin.Add(image.Pt(w, h))}, Factor: factor}
}

// Point convierte un punto físico al píxel del lienzo que aparece debajo.
func (v Viewport) Point(x, y float64) (image.Point, bool) {
	p := image.Pt(int(math.Floor(x)), int(math.Floor(y)))
	if !p.In(v.Rect) {
		return image.Point{}, false
	}
	return p.Sub(v.Rect.Min).Div(v.Factor), true
}

// Blit presenta el lienzo sin mezclar los píxeles vecinos.
func (v Viewport) Blit(screen interface {
	Fill(color.Color)
	DrawImage(*ebiten.Image, *ebiten.DrawImageOptions)
}, canvas *ebiten.Image) {
	screen.Fill(color.NRGBA{R: 16, G: 19, B: 24, A: 255})
	if v.Factor == 0 {
		return
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Scale(float64(v.Factor), float64(v.Factor))
	op.GeoM.Translate(float64(v.Rect.Min.X), float64(v.Rect.Min.Y))
	screen.DrawImage(canvas, op)
}

type sprite struct {
	id   ids.ID
	src  image.Rectangle
	dest image.Rectangle
}

// ArtScene comparte las colocaciones de sprites entre el dibujo y el picking.
type ArtScene struct {
	atlas   *Atlas
	sprites []sprite
}

// Arrange proyecta un frame ya interpolado y alinea los pivotes al píxel.
// El foco micro incluye su superficie y los descendientes de esa superficie.
func (a *Atlas) Arrange(f *view.Frame, p camera.Projector, visuals map[ids.ID]string, scale Scale, focus ids.ID) ArtScene {
	inFocus := func(spaceID ids.ID) bool {
		if focus.IsNil() {
			return true
		}
		for !spaceID.IsNil() {
			if spaceID == focus {
				return true
			}
			s, ok := f.SpaceByID(spaceID)
			if !ok {
				return false
			}
			spaceID = s.Parent
		}
		return false
	}
	out := ArtScene{atlas: a}
	add := func(id ids.ID, pos units.Vec3, name string) {
		src, pivot, ok := a.frame(visuals[id], scale, name, uint64(f.Tick))
		if !ok {
			return
		}
		x, y := p.ToScreen(pos)
		at := image.Pt(int(math.Round(float64(x))), int(math.Round(float64(y)))).Sub(pivot)
		dest := image.Rectangle{Min: at, Max: at.Add(src.Size())}
		if dest.Overlaps(image.Rect(0, 0, CanvasWidth, CanvasHeight)) {
			out.sprites = append(out.sprites, sprite{id: id, src: src, dest: dest})
		}
	}
	for _, s := range f.Spaces {
		if inFocus(s.ID) {
			add(s.ID, s.Pose.Pos, "idle")
		}
	}
	for _, t := range f.Things {
		if !inFocus(t.Space) {
			continue
		}
		name := "idle"
		if t.Mover != nil && t.Mover.Status == "moving" {
			name = "walk"
		}
		add(t.ID, t.Pose.Pos, name)
	}
	return out
}

// Pick devuelve el sprite visible en el píxel, respetando su transparencia.
func (s ArtScene) Pick(p image.Point) ids.ID {
	for i := len(s.sprites) - 1; i >= 0; i-- {
		sp := s.sprites[i]
		if p.In(sp.dest) {
			at := p.Sub(sp.dest.Min).Add(sp.src.Min)
			if _, _, _, alpha := s.atlas.pixels.At(at.X, at.Y).RGBA(); alpha > 0 {
				return sp.id
			}
		}
	}
	return ids.Nil
}

func (s ArtScene) Draw(dst *ebiten.Image, selected ids.ID) {
	dst.Fill(color.NRGBA{R: 35, G: 41, B: 50, A: 255})
	if s.atlas.texture == nil {
		s.atlas.texture = ebiten.NewImageFromImage(s.atlas.pixels)
	}
	var selection image.Rectangle
	for _, sp := range s.sprites {
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		op.GeoM.Translate(float64(sp.dest.Min.X), float64(sp.dest.Min.Y))
		dst.DrawImage(s.atlas.texture.SubImage(sp.src).(*ebiten.Image), op)
		if sp.id == selected {
			selection = sp.dest.Inset(-2)
		}
	}
	if !selection.Empty() {
		c := color.NRGBA{R: 118, G: 216, B: 203, A: 255}
		for _, x := range []float32{float32(selection.Min.X), float32(selection.Max.X - 1)} {
			for _, y := range []float32{float32(selection.Min.Y), float32(selection.Max.Y - 1)} {
				vector.FillRect(dst, x-1, y, 3, 1, c, false)
				vector.FillRect(dst, x, y-1, 1, 3, c, false)
			}
		}
	}
}
