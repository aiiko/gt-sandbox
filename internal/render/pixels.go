package render

import (
	"cmp"
	"image"
	"image/color"
	"math"
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/view"
)

// El lienzo de la escena llena la pantalla: mide pantalla / factor, con el
// factor entero más grande que deja al menos MinCanvasWidth × MinCanvasHeight
// píxeles lógicos (propuesta, 17.2). El mínimo deja ver la casa entera a
// 80 px/m (720 × 240) con margen para la interfaz: N = 2 en 1080p y 3 en 1440p.
const (
	MinCanvasWidth  = 768
	MinCanvasHeight = 432
)

// MaxCanvasFactor limita el factor de escena: en pantallas grandes el
// lienzo crece en vez de agrandar la interfaz (a 1440p, N = 3 y no 5).
const MaxCanvasFactor = 3

// CanvasFactor es el factor de escena para una pantalla física de w × h
// píxeles: el mayor entero, hasta MaxCanvasFactor, que deja un lienzo de al
// menos el mínimo (1 si la pantalla es más chica).
func CanvasFactor(w, h int) int {
	return max(1, min(w/MinCanvasWidth, h/MinCanvasHeight, MaxCanvasFactor))
}

// CanvasFactorFor devuelve el factor del lienzo de interfaz para w × h: con
// pref <= 0 usa CanvasFactor (automático); con pref en 1..4 lo aplica sin
// bajar el lienzo de 480 × 270 píxeles lógicos.
func CanvasFactorFor(w, h, pref int) int {
	if pref <= 0 {
		return CanvasFactor(w, h)
	}
	limit := max(1, min(w/480, h/270))
	return min(pref, limit)
}

// CanvasSize es el tamaño del lienzo que llena una pantalla física de w × h,
// redondeado hacia arriba para que no queden bandas.
func CanvasSize(w, h int) (int, int) {
	return CanvasSizeFor(w, h, 0)
}

// CanvasSizeFor es CanvasSize con un factor de interfaz preferido (0 = auto).
func CanvasSizeFor(w, h, pref int) (int, int) {
	n := CanvasFactorFor(w, h, pref)
	return (max(w, 1) + n - 1) / n, (max(h, 1) + n - 1) / n
}

// Viewport es el rectángulo del lienzo en píxeles físicos de la pantalla final.
// Puede exceder la pantalla en menos de un píxel del lienzo por lado. Un
// píxel del lienzo mide Factor píxeles físicos o, con Down > 1, 1/Down: la
// capa de escena alejada, que se reduce promediando (mipmaps).
type Viewport struct {
	Rect   image.Rectangle
	Factor int
	Down   int // 0 o 1: sin reducir
}

// PixelViewport centra un lienzo de tamaño canvas sobre la pantalla, escalado
// por el factor de escena de esa pantalla.
func PixelViewport(screen image.Rectangle, canvas image.Point) Viewport {
	return PixelViewportFor(screen, canvas, 0)
}

// PixelViewportFor es PixelViewport con un factor de interfaz preferido (0 = auto).
func PixelViewportFor(screen image.Rectangle, canvas image.Point, pref int) Viewport {
	factor := CanvasFactorFor(screen.Dx(), screen.Dy(), pref)
	size := canvas.Mul(factor)
	origin := screen.Min.Add(screen.Size().Sub(size).Div(2))
	return Viewport{Rect: image.Rectangle{Min: origin, Max: origin.Add(size)}, Factor: factor}
}

func (v Viewport) down() int { return max(1, v.Down) }

// Point convierte un punto físico al píxel del lienzo que aparece debajo.
func (v Viewport) Point(x, y float64) (image.Point, bool) {
	p := image.Pt(int(math.Floor(x)), int(math.Floor(y)))
	if v.Factor == 0 || !p.In(v.Rect) {
		return image.Point{}, false
	}
	return p.Sub(v.Rect.Min).Mul(v.down()).Div(v.Factor), true
}

// Blit presenta el lienzo sin mezclar los píxeles vecinos, sobre el fondo.
func (v Viewport) Blit(screen interface {
	Fill(color.Color)
	DrawImage(*ebiten.Image, *ebiten.DrawImageOptions)
}, canvas *ebiten.Image) {
	screen.Fill(Backdrop)
	v.Draw(screen, canvas)
}

// Draw presenta el lienzo sin fondo, así la interfaz se compone sobre la
// capa de escena. Ampliado, sin mezclar los píxeles vecinos; reducido, con
// filtro lineal, que Ebitengine resuelve con mipmaps (cada píxel físico es
// el promedio de los del arte que cubre).
func (v Viewport) Draw(screen interface {
	DrawImage(*ebiten.Image, *ebiten.DrawImageOptions)
}, canvas *ebiten.Image) {
	if v.Factor == 0 {
		return
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	if v.down() > 1 {
		op.Filter = ebiten.FilterLinear
	}
	k := float64(v.Factor) / float64(v.down())
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(float64(v.Rect.Min.X), float64(v.Rect.Min.Y))
	screen.DrawImage(canvas, op)
}

// Backdrop es el fondo de la pantalla fuera del lienzo.
var Backdrop = color.NRGBA{R: 16, G: 19, B: 24, A: 255}

// Scene es la capa de escena que cubre el mismo rectángulo que el lienzo de
// la interfaz con otra escala (el zoom de la rueda, propuesta 17.4): factor
// píxeles físicos por píxel del arte si es 1 o más (entero) y, por debajo de
// 1, 1/factor píxeles del arte por píxel físico (½, ¼). Mide lo justo para
// cubrir el lienzo y queda centrada, a menos de un píxel físico del centro
// del lienzo. Con el factor del lienzo es el lienzo mismo.
func (v Viewport) Scene(factor float64) Viewport {
	if v.Factor == 0 || factor <= 0 {
		return Viewport{}
	}
	up, down := int(math.Round(factor)), 0
	if factor < 1 {
		up, down = 1, int(math.Round(1/factor))
	}
	size := v.Rect.Size()
	blocks := image.Pt((size.X+up-1)/up, (size.Y+up-1)/up)
	extra := blocks.Mul(up).Sub(size)
	origin := v.Rect.Min.Sub(extra.Div(2))
	return Viewport{Rect: image.Rectangle{Min: origin, Max: origin.Add(blocks.Mul(up))}, Factor: up, Down: down}
}

// Size es el tamaño del lienzo en sus propios píxeles.
func (v Viewport) Size() image.Point {
	if v.Factor == 0 {
		return image.Point{}
	}
	return v.Rect.Size().Mul(v.down()).Div(v.Factor)
}

type sprite struct {
	id    ids.ID
	src   image.Rectangle
	dest  image.Rectangle
	layer int
	y     float32
	group ids.ID
	z     units.Length
	alpha float32 // 0 = opaco
	ghost bool    // sombras y señales: no se eligen ni llevan el marco de selección
	space bool
}

// climbFront es cuánto más alto que la cima de la tira se ordena quien la
// trepa: lo justo para quedar delante del mueble al que lleva.
const climbFront = 10 * units.Centimeter

// FlattenClimbs devuelve el frame como se dibuja: lo que está en un Space
// que se trepa (ClimbVisual) pasa a la base de la tira, porque su Y local es
// la altura y ninguna vista dibuja la altura. Sin tiras que
// trepar devuelve el mismo frame.
func FlattenClimbs(f *view.Frame, visuals map[ids.ID]string) *view.Frame {
	var out *view.Frame
	for i, t := range f.Things {
		if visuals[t.Space] != ClimbVisual {
			continue
		}
		s, ok := f.SpaceByID(t.Space)
		if !ok {
			continue
		}
		if out == nil {
			c := *f
			c.Things = slices.Clone(f.Things)
			out = &c
		}
		tr := s.Pose.Transform()
		local := tr.ApplyInv(t.Pose.Pos)
		local.Y = 0
		out.Things[i].Pose.Pos = tr.Apply(local)
	}
	if out == nil {
		return f
	}
	return out
}

// look ajusta cómo se dibuja una colocación.
type look struct {
	visual string  // "" = el visual de la entidad
	alpha  float32 // 0 = opaco
	ghost  bool
	layer  int    // 0 = la capa de la variante
	top    bool   // encima de todo (el reflejo de lo seguido bajo una superficie más alta)
	over   ids.ID // macro: objeto cuyo lugar ocupa (la giantess en la cama)
}

const (
	// PassVisual es el visual de la señal de un paso entre superficies en contacto.
	PassVisual = "pass"
	// FootprintVisual es el visual de la sombra de la próxima pisada (macro).
	FootprintVisual = "footprint"
	// ClimbVisual marca un Space que se trepa (una pata, un respaldo): la
	// simulación lo modela como una tira cuyo eje Y local es la altura. No
	// se dibuja, y lo que está en él se ve en la base de la tira
	// (FlattenClimbs), de espaldas y caminando, no corrido hacia el norte.
	ClimbVisual = "climb"
	// FootVisual es el pie de la giantess visto de cerca: su sombra antes de
	// bajar (shadow_<dir>) y el pie apoyado (foot_<dir>).
	FootVisual = "foot"
	// GripVisual marca un punto de agarre (costura, puño, mechón) en una zona corporal.
	GripVisual = "grip_point"
	// ShadowReach es la altura a la que la sombra anticipada de una mano deja de verse.
	ShadowReach = units.Meter
	// fallLift limita cuántos píxeles cae un sprite en la animación de caída.
	fallLift = 48
)

const (
	layerBackground = iota
	layerGround
	layerDepth
	layerForeground
)

// ArtScene comparte las colocaciones de sprites entre el dibujo y el picking.
type ArtScene struct {
	atlas    *Atlas
	sprites  []sprite
	clips    map[ids.ID]string
	marks    map[ids.ID][]string
	lifts    map[ids.ID]int
	backdrop *basicSurface
	cutaway  *cutawayMask
}

// cutawayMask recorta la capa frontal de una cavidad cerrada (pasacables del
// zócalo o bolsillo del sofá) alrededor del tiny seguido (propuesta, 18.7).
type cutawayMask struct {
	outer image.Rectangle
	hole  image.Rectangle
}

// basicSurface dibuja una superficie micro sin arte específico en el atlas
// (patas, respaldos, pasacables, bolsillo del sofá o cualquier Space nuevo):
// límites, tono de material y entradas/salidas. Ver propuesta 18.7.
type basicSurface struct {
	id      ids.ID
	rect    image.Rectangle
	climb   bool
	portals []portalMark
}

type portalMark struct {
	at       image.Point
	disabled bool
}

// HasBackdrop indica si la escena usó la vista micro básica (sin arte propio).
func (s ArtScene) HasBackdrop() bool { return s.backdrop != nil }

// Clip devuelve el clip con el que se dibujó una entidad, o "" si no se dibujó.
func (s ArtScene) Clip(id ids.ID) string { return s.clips[id] }

// Marks devuelve los clips de las sombras y señales colocadas para una
// entidad (un portal, una mano, un tiny que cae), estén o no en pantalla.
func (s ArtScene) Marks(id ids.ID) []string { return s.marks[id] }

// Lift devuelve cuántos píxeles por encima de su pivote se dibujó una
// entidad (una carga en la mano, un tiny que cae).
func (s ArtScene) Lift(id ids.ID) int { return s.lifts[id] }

// Arrange proyecta un frame ya interpolado y alinea los pivotes al píxel.
// El foco micro incluye su superficie y los descendientes de esa superficie.
func (a *Atlas) Arrange(f *view.Frame, p camera.Projector, visuals map[ids.ID]string, scale Scale, focus ids.ID, featured ...ids.ID) ArtScene {
	// En macro (o en micro de otra superficie), quien trepa se ve en la base;
	// al enfocar en micro la propia pata/respaldo, su eje Y es la altura real.
	if scale == Macro || visuals[focus] != ClimbVisual {
		f = FlattenClimbs(f, visuals)
	}
	canvas := p.Bounds()
	scope := map[ids.ID]bool{focus: true}
	for _, c := range f.Contacts {
		if c.A == focus {
			scope[c.B] = true
		}
		if c.B == focus {
			scope[c.A] = true
		}
	}
	inFocus := func(spaceID ids.ID) bool {
		if focus.IsNil() {
			return true
		}
		for !spaceID.IsNil() {
			if scope[spaceID] {
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
	visibleAs := func(id ids.ID, visual string) (variant, bool) {
		v, ok := a.sprites[visual][scale]
		return v, ok && !v.Hidden && (!v.SelectedOnly || slices.Contains(featured, id))
	}
	visible := func(id ids.ID) (variant, bool) { return visibleAs(id, visuals[id]) }
	out := ArtScene{atlas: a, clips: map[ids.ID]string{}, marks: map[ids.ID][]string{}, lifts: map[ids.ID]int{}}
	carrying, inContact, resting := map[ids.ID]bool{}, map[ids.ID]bool{}, map[ids.ID]bool{}
	for _, t := range f.Things {
		if !t.CarriedBy.IsNil() {
			carrying[t.CarriedBy] = true
		}
	}
	for _, c := range f.Contacts {
		inContact[c.A], inContact[c.B] = true, true
	}
	for _, s := range f.Spaces {
		if inContact[s.ID] {
			resting[s.Parent] = true
		}
	}
	add := func(id, support, carrier ids.ID, pos units.Vec3, names []string, isSpace bool, offset image.Point, lk look) {
		visual := lk.visual
		if visual == "" {
			visual = visuals[id]
		}
		v, ok := visibleAs(id, visual)
		if !ok {
			return
		}
		name := v.resolve(names)
		if lk.top {
			// no es una marca: es la misma entidad vista a través de lo que la tapa
		} else if lk.ghost {
			out.marks[id] = append(out.marks[id], name)
		} else {
			out.clips[id], out.lifts[id] = name, -offset.Y
		}
		x, y := p.ToScreen(pos)
		at := image.Pt(int(math.Round(float64(x)))-v.Pivot[0], int(math.Round(float64(y)))-v.Pivot[1]).Add(offset)
		group, depth := id, y+float32(v.DepthOffset)
		height := pos.Z
		layer := layerDepth
		switch v.Layer {
		case "background":
			layer = layerBackground
		case "ground":
			layer = layerGround
		default:
			if !carrier.IsNil() {
				group = carrier
				depth = y + float32(a.sprites[visuals[carrier]][scale].DepthOffset)
			}
			if scale == Macro {
				group, depth = a.depthGroup(f, p, visuals, support, group, depth)
				if t, ok := f.ThingByID(lk.over); ok {
					_, ty := p.ToScreen(t.Pose.Pos)
					group, depth = t.ID, ty+float32(a.sprites[visuals[t.ID]][Macro].DepthOffset)+occupantDepth
				}
			}
		}
		if scale == Micro {
			if isSpace {
				layer = layerBackground
			} else if sp, ok := f.SpaceByID(support); ok {
				group, height = sp.ID, sp.Pose.Pos.Z
				if visuals[sp.ID] == ClimbVisual {
					// quien trepa va por fuera de la pata: delante del mueble
					// al que lleva, aunque la base de la tira esté en el piso
					height += sp.Extent.Max.Y + climbFront
				}
			}
		}
		if lk.layer != 0 {
			layer = lk.layer
		}
		if lk.top {
			group, height, layer = id, units.Length(math.MaxInt64), layerForeground
		}
		addFrame := func(src image.Rectangle, layer int) {
			dest := image.Rectangle{Min: at, Max: at.Add(src.Size())}
			if dest.Overlaps(canvas) {
				out.sprites = append(out.sprites, sprite{id: id, src: src, dest: dest, layer: layer, y: depth, group: group, z: height,
					alpha: lk.alpha, ghost: lk.ghost, space: isSpace})
			}
		}
		addFrame(v.frame(name, uint64(f.Tick)), layer)
		if v.Foreground != nil {
			addFrame(frameRect(v.Foreground), layerForeground)
		}
	}
	microPos := func(spaceID ids.ID, worldPos, localPos units.Vec3) units.Vec3 {
		if scale != Micro {
			return worldPos
		}
		if sp, ok := f.SpaceByID(spaceID); ok && sp.UsesMicroLocal() {
			return sp.MicroPos(worldPos, localPos)
		}
		return worldPos
	}
	for _, s := range f.Spaces {
		if !inFocus(s.ID) {
			continue
		}
		// Patas, respaldos, cavidades (bolsillo del sofá) y zonas corporales
		// proyectadas en 3D solo dibujan su micro cuando son el foco. Las demás
		// partes articuladas del cuerpo (salvo la palma) solo se dibujan cuando
		// son el foco o están en contacto directo con él, para que la cabeza o
		// las caderas no tapen el pie o el suelo en vista micro.
		if scale == Micro && s.ID != focus {
			if visuals[s.ID] == ClimbVisual || visuals[s.ID] == "sofa_pocket" || s.Projected {
				continue
			}
			if s.Kind == space.Articulated && visuals[s.ID] != "palm" && !scope[s.ID] {
				continue
			}
		}
		if v, ok := visible(s.ID); ok && v.Tile {
			out.tiles(f, p, s, v)
			continue
		}
		names := actor{mover: s.Mover, facing: s.Pose.Rot, activity: s.Activity, step: s.Step,
			carrying: carrying[s.ID], resting: resting[s.ID]}.clips()
		if s.Mover == nil && !s.Extent.Empty() {
			names = surfaceClips(s.Pose.Rot, inContact[s.ID])
		}
		support, lk := s.Parent, look{}
		if scale == Macro {
			support = a.seatOf(f, visuals, s, support)
			if support == s.Parent {
				lk.over = a.placeOf(f, visuals, s)
			}
		}
		add(s.ID, support, ids.Nil, s.Pose.Pos, names, true, image.Point{}, lk)
	}
	if scale == Macro {
		// Sombra de la próxima pisada de quien camina: más oscura cuanto más
		// cerca está de bajar.
		for _, fp := range f.Footprints {
			w, ok := f.SpaceByID(fp.Walker)
			if !ok || w.Mover == nil || w.Mover.Status != "moving" || !inFocus(fp.Space) {
				continue
			}
			add(fp.Walker, fp.Space, ids.Nil, fp.Pos, []string{"idle"}, false, image.Point{},
				look{visual: FootprintVisual, ghost: true, layer: layerGround, alpha: 0.35})
		}
		for _, pt := range f.Portals {
			if !pt.Dynamic && !pt.Disabled {
				mid := units.V3(pt.A.X+(pt.B.X-pt.A.X)/2, pt.A.Y+(pt.B.Y-pt.A.Y)/2, pt.A.Z)
				add(pt.ID, ids.Nil, ids.Nil, mid, []string{"idle"}, false, image.Point{}, look{})
			}
		}
	}
	if scale == Micro {
		// Recorte de capa frontal alrededor del tiny seguido en cavidades o zonas
		// ocultas por ropa o pelo (propuesta 18.7 y 19.5).
		if fs, ok := f.SpaceByID(focus); ok && !fs.Extent.Empty() {
			if v := visuals[focus]; v == "cable_duct" || v == "sofa_pocket" || fs.Occluded {
				b := buildBackdrop(f, p, *fs, visuals[fs.ID] == ClimbVisual || fs.Climb)
				for _, fid := range featured {
					if th, ok := f.ThingByID(fid); ok && th.Space == focus {
						sx, sy := p.ToScreen(microPos(th.Space, th.Pose.Pos, th.LocalPos))
						cx, cy := int(math.Round(float64(sx))), int(math.Round(float64(sy)))
						out.cutaway = &cutawayMask{
							outer: b.rect,
							hole:  image.Rect(cx-26, cy-26, cx+26, cy+26),
						}
						break
					}
				}
			}
		}
		// Superficie básica si el Space enfocado no tiene arte micro propio
		// (patas, respaldos, pasacables, bolsillo del sofá o zonas corporales).
		if fs, ok := f.SpaceByID(focus); ok && !fs.Extent.Empty() {
			if _, vis := visible(fs.ID); !vis {
				out.backdrop = buildBackdrop(f, p, *fs, visuals[fs.ID] == ClimbVisual || fs.Climb)
				out.clips[fs.ID] = "basic"
			}
		}
		// Señal del paso en cada extremo visible de un portal de contacto.
		// El pie: su sombra antes de bajar, más oscura cuanto más cerca, y el
		// pie apoyado durante StepTicks, por encima de lo que pisa.
		for _, fp := range f.Footprints {
			w, ok := f.SpaceByID(fp.Walker)
			if !ok || w.Mover == nil || w.Mover.Status != "moving" || !inFocus(fp.Space) || fp.Stride <= 0 {
				continue
			}
			near := 1 - float32(fp.Remaining)/float32(fp.Stride)
			add(fp.Walker, fp.Space, ids.Nil, fp.Pos, []string{"shadow_" + heading(fp.Dir)}, false, image.Point{},
				look{visual: FootVisual, ghost: true, layer: layerGround, alpha: 0.15 + 0.45*max(near, 0)})
		}
		for _, st := range f.Steps {
			if inFocus(st.Space) {
				add(st.Walker, st.Space, ids.Nil, st.Pos, []string{"foot_" + heading(st.Dir)}, false, image.Point{},
					look{visual: FootVisual, ghost: true, layer: layerForeground})
			}
		}
		focusArticulated := false
		if fs, ok := f.SpaceByID(focus); ok {
			focusArticulated = fs.Kind == space.Articulated
			for _, gr := range fs.Grips {
				pos := fs.MicroPos(fs.Pose.Transform().Apply(units.V3(gr.Pos.X, gr.Pos.Y, 0)), units.V3(gr.Pos.X, gr.Pos.Y, 0))
				add(fs.ID, fs.ID, ids.Nil, pos, []string{"idle"}, false, image.Point{},
					look{visual: GripVisual, ghost: true, layer: layerGround, alpha: 0.78})
			}
		}
		for _, pt := range f.Portals {
			if !pt.Dynamic && !focusArticulated {
				continue
			}
			state := []string{"open"}
			if pt.Disabled {
				state = []string{"closed"}
			}
			for _, end := range []struct {
				pos   units.Vec3
				local units.Vec2
				space ids.ID
			}{{pt.A, pt.ALocal, pt.ASpace}, {pt.B, pt.BLocal, pt.BSpace}} {
				if !pt.Dynamic && end.space != focus {
					continue
				}
				if inFocus(end.space) {
					pos := microPos(end.space, end.pos, units.V3(end.local.X, end.local.Y, 0))
					add(pt.ID, end.space, ids.Nil, pos, state, false, image.Point{},
						look{visual: PassVisual, ghost: true, layer: layerGround})
				}
			}
		}
		// Sombra anticipada de una superficie móvil que todavía no se apoya
		// sobre la superficie enfocada (solo visuales con clip shadow_<dir>, y
		// nunca sobre otra zona del mismo cuerpo).
		if fs, ok := f.SpaceByID(focus); ok && !fs.BodyZone {
			for _, s := range f.Spaces {
				if s.Mover != nil || s.Extent.Empty() || inFocus(s.ID) {
					continue
				}
				shadowClip := "shadow_" + compass(s.Pose.Rot)
				v, ok := visible(s.ID)
				if !ok || v.Tile {
					continue
				}
				if _, has := v.Clips[shadowClip]; !has {
					continue
				}
				if lk, offset, ok := handShadow(fs, s); ok {
					add(s.ID, fs.ID, ids.Nil, s.Pose.Pos, []string{shadowClip}, false, offset, lk)
				}
			}
		}
	}
	for _, t := range f.Things {
		if t.CarriedBy.IsNil() && inFocus(t.Space) {
			names := actor{mover: t.Mover, facing: t.Pose.Rot, activity: t.Activity, step: t.Step,
				carrying: carrying[t.ID], built: t.Built}.clips()
			pos := microPos(t.Space, t.Pose.Pos, t.LocalPos)
			var offset image.Point
			if fl, ok := f.FallOf(t.ID); ok {
				// Cae desde arriba con aceleración; su sombra lo espera en el apoyo.
				lift := fallOffset(p, fl, f.Tick)
				names, offset = append([]string{"fall"}, names...), image.Pt(0, -lift)
				add(t.ID, t.Space, ids.Nil, pos, []string{"shadow"}, false, image.Point{},
					look{ghost: true, layer: layerGround, alpha: 0.3 + 0.5*float32(fallLift-lift)/fallLift})
			}
			add(t.ID, t.Space, ids.Nil, pos, names, false, offset, look{})
			if scale == Micro && slices.Contains(featured, t.ID) {
				// Lo seguido se ve a través de una superficie que lo tapa (la
				// mano apoyada encima del extremo del paso).
				add(t.ID, t.Space, ids.Nil, pos, names, false, offset, look{ghost: true, top: true, alpha: 0.45})
			}
		}
	}
	for _, t := range f.Things {
		if t.CarriedBy.IsNil() {
			continue
		}
		v, ok := visible(t.CarriedBy)
		if !ok {
			continue
		}
		var support ids.ID
		var pos units.Vec3
		if carrier, ok := f.ThingByID(t.CarriedBy); ok {
			support, pos = carrier.Space, microPos(carrier.Space, carrier.Pose.Pos, carrier.LocalPos)
		} else if carrier, ok := f.SpaceByID(t.CarriedBy); ok {
			support, pos = carrier.Parent, carrier.Pose.Pos
		} else {
			continue
		}
		if inFocus(support) {
			var offset image.Point
			if v.Carry != nil {
				offset = image.Pt(v.Carry[0], v.Carry[1])
				// La carga va del lado hacia el que mira el porteador.
				if strings.HasSuffix(out.clips[t.CarriedBy], "_west") {
					offset.X = -offset.X
				}
			}
			add(t.ID, support, t.CarriedBy, pos, []string{"held"}, false, offset, look{})
		}
	}
	slices.SortStableFunc(out.sprites, func(a, b sprite) int {
		if scale == Micro {
			return cmp.Or(cmp.Compare(a.z, b.z), cmp.Compare(a.group, b.group), cmp.Compare(a.layer, b.layer), cmp.Compare(a.y, b.y))
		}
		return cmp.Or(cmp.Compare(a.layer, b.layer), cmp.Compare(a.y, b.y), cmp.Compare(a.group, b.group))
	})
	return out
}

// tiles cubre la extensión de un Space con su baldosa, desde la esquina
// mínima, solo donde se ve.
func (s *ArtScene) tiles(f *view.Frame, p camera.Projector, sp view.Space, v variant) {
	src := v.frame("idle", uint64(f.Tick))
	w, h := src.Dx(), src.Dy()
	perMeter := float64(p.Pixels(units.Meter))
	if perMeter <= 0 {
		return
	}
	stepX := units.Length(math.Round(float64(w) / perMeter * float64(units.Meter)))
	stepY := units.Length(math.Round(float64(h) / perMeter * float64(units.Meter)))
	t := sp.Pose.Transform()
	canvas := p.Bounds()
	px := func(l units.Length) int { return int(math.Round(float64(l) * perMeter / float64(units.Meter))) }
	for y := sp.Extent.Min.Y; y < sp.Extent.Max.Y; y += stepY {
		for x := sp.Extent.Min.X; x < sp.Extent.Max.X; x += stepX {
			// La esquina mínima de la baldosa queda abajo a la izquierda. La
			// última fila y la última columna se recortan en el borde.
			cut := src
			cut.Max.X = cut.Min.X + min(w, px(sp.Extent.Max.X-x))
			cut.Min.Y = cut.Max.Y - min(h, px(sp.Extent.Max.Y-y))
			if cut.Empty() {
				continue
			}
			sx, sy := p.ToScreen(t.Apply(units.V3(x, y, 0)))
			at := image.Pt(int(math.Round(float64(sx))), int(math.Round(float64(sy)))-cut.Dy())
			dest := image.Rectangle{Min: at, Max: at.Add(cut.Size())}
			if dest.Overlaps(canvas) {
				s.sprites = append(s.sprites, sprite{id: sp.ID, src: cut, dest: dest, layer: layerBackground,
					y: float32(dest.Max.Y), group: sp.ID, z: sp.Pose.Pos.Z, space: true})
			}
		}
	}
	s.clips[sp.ID] = "idle"
}

// heading da el punto cardinal más cercano a una dirección de mundo.
func heading(d units.Vec2) string {
	switch {
	case d.X.Abs() >= d.Y.Abs() && d.X >= 0:
		return "east"
	case d.X.Abs() >= d.Y.Abs():
		return "west"
	case d.Y > 0:
		return "north"
	}
	return "south"
}

// handShadow decide si una superficie móvil s (una mano) proyecta su sombra
// sobre la superficie enfocada fs: está encima, a menos de ShadowReach, y su
// centro cae sobre fs o cerca del borde. Cuanto más baja, más oscura y más
// pegada al apoyo.
func handShadow(fs *view.Space, s view.Space) (look, image.Point, bool) {
	dz := s.Pose.Pos.Z - fs.Pose.Pos.Z
	if fs.Extent.Empty() || dz <= 0 || dz >= ShadowReach {
		return look{}, image.Point{}, false
	}
	local := fs.Pose.Transform().ApplyInv(s.Pose.Pos).XY()
	margin := max(s.Extent.Max.X-s.Extent.Min.X, s.Extent.Max.Y-s.Extent.Min.Y) / 2
	r := fs.Extent
	if local.X < r.Min.X-margin || local.X > r.Max.X+margin || local.Y < r.Min.Y-margin || local.Y > r.Max.Y+margin {
		return look{}, image.Point{}, false
	}
	h := float64(dz) / float64(ShadowReach)
	offset := image.Pt(2+int(math.Round(12*h)), 4+int(math.Round(18*h)))
	return look{ghost: true, layer: layerGround, alpha: float32(0.42 - 0.27*h)}, offset, true
}

// fallOffset es cuántos píxeles por encima de su apoyo se dibuja una entidad
// que cae: parte de la altura de la caída (limitada a fallLift) y baja con
// aceleración hasta tocar el apoyo al final de view.FallTicks.
func fallOffset(p camera.Projector, fl view.Fall, tick clock.Tick) int {
	x0, _ := p.ToScreen(units.V3(0, 0, 0))
	x1, _ := p.ToScreen(units.V3(fl.Height, 0, 0))
	start := min(math.Abs(float64(x1-x0)), fallLift)
	t := min(float64(tick-fl.Tick)/float64(view.FallTicks), 1)
	return int(math.Round(start * (1 - t*t)))
}

// depthGroup mantiene al ocupante junto al mueble que lo sostiene.
func (a *Atlas) depthGroup(f *view.Frame, p camera.Projector, visuals map[ids.ID]string, support, group ids.ID, depth float32) (ids.ID, float32) {
	for !support.IsNil() {
		s, ok := f.SpaceByID(support)
		if !ok {
			break
		}
		if v, ok := a.sprites[visuals[s.ID]][Macro]; ok && v.Layer == "" && !s.Extent.Empty() {
			_, y := p.ToScreen(s.Pose.Pos)
			// el ocupante va justo encima de su mueble
			group, depth = s.ID, y+float32(v.DepthOffset)+occupantDepth
		}
		support = s.Parent
	}
	return group, depth
}

// occupantDepth deja al ocupante después del mueble con la misma profundidad.
const occupantDepth = 0.01

// placeOf devuelve el objeto con sprite macro cuyo lugar ocupa un agente
// quieto (la giantess acostada en la cama, que no tiene extensión propia).
func (a *Atlas) placeOf(f *view.Frame, visuals map[ids.ID]string, s view.Space) ids.ID {
	if s.Mover == nil || s.Mover.Status == "moving" {
		return ids.Nil
	}
	for _, t := range f.Things {
		if t.Space != s.Parent || t.Mover != nil || !t.CarriedBy.IsNil() {
			continue // un tiny o algo en la mano no es un lugar
		}
		if v, ok := a.sprites[visuals[t.ID]][Macro]; !ok || v.Layer != "" {
			continue
		}
		d := t.Pose.Pos.Sub(s.Pose.Pos)
		if d.X.Abs() <= placeRadius && d.Y.Abs() <= placeRadius {
			return t.ID
		}
	}
	return ids.Nil
}

// placeRadius es cuánto puede apartarse del punto del objeto quien ocupa su lugar.
const placeRadius = 10 * units.Centimeter

// seatOf devuelve el mueble sobre el que está quieto un agente que es un
// Space (la giantess sentada en el sofá o la silla, o en la cama): el
// hermano con sprite macro cuya extensión contiene su posición. Sin él,
// devuelve parent.
func (a *Atlas) seatOf(f *view.Frame, visuals map[ids.ID]string, s view.Space, parent ids.ID) ids.ID {
	if s.Mover == nil || s.Mover.Status == "moving" {
		return parent
	}
	for _, o := range f.Spaces {
		if o.ID == s.ID || o.Parent != parent || o.Mover != nil || o.Extent.Empty() {
			continue
		}
		if v, ok := a.sprites[visuals[o.ID]][Macro]; !ok || v.Layer != "" {
			continue
		}
		local := o.Pose.Transform().ApplyInv(s.Pose.Pos)
		e := o.Extent
		if local.X >= e.Min.X && local.X <= e.Max.X && local.Y >= e.Min.Y && local.Y <= e.Max.Y {
			return o.ID
		}
	}
	return parent
}

// poseClips relaciona la acción de una oferta en uso con la pose que la dibuja.
var poseClips = map[string]string{
	"breakfast":    "eat",
	"snack":        "eat",
	"snack_living": "eat",
	"eat":          "eat",
	"nibble":       "eat",
	"watch":        "sit",
	"relax":        "sit",
	"sleep":        "sleep",
}

// actor es el estado de una entidad del que sale su clip.
type actor struct {
	mover          *view.Mover
	facing         fixedmath.Angle
	activity, step string
	carrying       bool // lleva algo
	resting        bool // tiene una parte apoyada en otra superficie
	built          bool // una obra terminada (refugio auxiliar)
}

func compass(facing fixedmath.Angle) string {
	directions := [...]string{"east", "north", "west", "south"}
	quarter := fixedmath.QuarterTurn
	return directions[((facing.Norm()+quarter/2)/quarter)%4]
}

// clips devuelve los clips candidatos, del más específico al más general.
func (a actor) clips() []string {
	if a.built {
		return []string{"built", "idle"}
	}
	if a.mover == nil {
		return []string{"idle"}
	}
	direction := compass(a.facing)
	if a.mover.Status == "moving" {
		if len(a.mover.Route) > 1 {
			d := a.mover.Route[1].Sub(a.mover.Route[0])
			switch {
			case d.X.Abs() >= d.Y.Abs() && d.X > 0:
				direction = "east"
			case d.X.Abs() >= d.Y.Abs() && d.X < 0:
				direction = "west"
			case d.Y > 0:
				direction = "north"
			case d.Y < 0:
				direction = "south"
			case d.Z > 0:
				direction = "north"
			case d.Z < 0:
				direction = "south"
			}
		}
		walk := []string{"walk_" + direction, "walk", "idle_" + direction}
		if a.carrying {
			walk = append([]string{"carry_" + direction}, walk...)
		}
		return walk
	}
	var out []string
	if action, ok := strings.CutPrefix(a.activity, "use:"); ok && (a.step == "" || a.step == "use") {
		if pose, ok := poseClips[action]; ok {
			out = append(out, pose+"_"+direction, pose)
		}
	}
	if a.resting {
		out = append(out, "reach_"+direction, "reach")
	}
	return append(out, "idle_"+direction)
}

// surfaceClips elige la variante apoyada o elevada de una superficie móvil
// según su orientación de mundo.
func surfaceClips(facing fixedmath.Angle, contact bool) []string {
	state := "raised"
	if contact {
		state = "rested"
	}
	return []string{state + "_" + compass(facing), state, "idle"}
}

// Pick devuelve el sprite visible en el píxel, respetando su transparencia.
func (s ArtScene) Pick(p image.Point) ids.ID {
	for i := len(s.sprites) - 1; i >= 0; i-- {
		sp := s.sprites[i]
		if !sp.ghost && p.In(sp.dest) {
			at := p.Sub(sp.dest.Min).Add(sp.src.Min)
			if _, _, _, alpha := s.atlas.pixels.At(at.X, at.Y).RGBA(); alpha > 0 {
				return sp.id
			}
		}
	}
	if s.backdrop != nil && p.In(s.backdrop.rect) {
		return s.backdrop.id
	}
	return ids.Nil
}

// Placement es un sprite colocado, en orden de dibujo.
type Placement struct {
	ID    ids.ID
	Src   image.Rectangle // cuadro del atlas
	Dest  image.Rectangle // destino en el lienzo, alineado al píxel
	Layer int             // 0 fondo, 1 suelo, 2 profundidad, 3 primer plano
	Alpha float32         // 0 = opaco
	Ghost bool            // sombra o señal
	Space bool            // una superficie (la base sobre la que se apoyan otros)
}

// Placements devuelve los sprites en el orden en que se dibujan.
func (s ArtScene) Placements() []Placement {
	out := make([]Placement, len(s.sprites))
	for i, sp := range s.sprites {
		out[i] = Placement{ID: sp.id, Src: sp.src, Dest: sp.dest, Layer: sp.layer, Alpha: sp.alpha, Ghost: sp.ghost, Space: sp.space}
	}
	return out
}

// Bounds devuelve el rectángulo de pantalla del sprite principal de una
// entidad (sin sombras, señales ni primer plano).
func (s ArtScene) Bounds(id ids.ID) (image.Rectangle, bool) {
	var r image.Rectangle
	for _, sp := range s.sprites {
		if sp.id == id && !sp.ghost && sp.layer != layerForeground {
			r = r.Union(sp.dest)
		}
	}
	return r, !r.Empty()
}

func (a *Atlas) tex() *ebiten.Image {
	if a.texture == nil {
		a.texture = ebiten.NewImageFromImage(a.pixels)
	}
	return a.texture
}

// DrawClip dibuja el primer cuadro de un clip de un visual con su pivote en
// at. Sirve a la interfaz (retrato, símbolos). Devuelve false si no existe.
func (a *Atlas) DrawClip(dst *ebiten.Image, visual string, scale Scale, clip string, at image.Point) bool {
	v, ok := a.sprites[visual][scale]
	if !ok || v.Hidden {
		return false
	}
	c, ok := v.Clips[clip]
	if !ok {
		return false
	}
	src := frameRect(c.Frames[0])
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Translate(float64(at.X-v.Pivot[0]), float64(at.Y-v.Pivot[1]))
	dst.DrawImage(a.tex().SubImage(src).(*ebiten.Image), op)
	return true
}

func (s ArtScene) Draw(dst *ebiten.Image, selected ids.ID) {
	dst.Fill(color.NRGBA{R: 35, G: 41, B: 50, A: 255})
	if s.backdrop != nil {
		drawBackdrop(dst, s.backdrop)
	}
	s.atlas.tex()
	var selection image.Rectangle
	for _, sp := range s.sprites {
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		op.GeoM.Translate(float64(sp.dest.Min.X), float64(sp.dest.Min.Y))
		if sp.alpha > 0 {
			op.ColorScale.ScaleAlpha(sp.alpha)
		}
		dst.DrawImage(s.atlas.texture.SubImage(sp.src).(*ebiten.Image), op)
		if sp.id == selected && !sp.ghost {
			selection = sp.dest.Inset(-2)
		}
	}
	if s.cutaway != nil {
		drawCutaway(dst, s.cutaway)
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

func buildBackdrop(f *view.Frame, p camera.Projector, sp view.Space, climb bool) *basicSurface {
	tr := sp.Pose.Transform()
	corners := [4]units.Vec2{
		sp.Extent.Min,
		{X: sp.Extent.Max.X, Y: sp.Extent.Min.Y},
		sp.Extent.Max,
		{X: sp.Extent.Min.X, Y: sp.Extent.Max.Y},
	}
	var rect image.Rectangle
	for i, c := range corners {
		sx, sy := p.ToScreen(sp.MicroPos(tr.Apply(units.V3(c.X, c.Y, 0)), units.V3(c.X, c.Y, 0)))
		pt := image.Pt(int(math.Round(float64(sx))), int(math.Round(float64(sy))))
		r := image.Rectangle{Min: pt, Max: pt.Add(image.Pt(1, 1))}
		if i == 0 {
			rect = r
		} else {
			rect = rect.Union(r)
		}
	}
	var marks []portalMark
	for _, pt := range f.Portals {
		for _, end := range [2]struct {
			sp    ids.ID
			pos   units.Vec3
			local units.Vec2
		}{{pt.ASpace, pt.A, pt.ALocal}, {pt.BSpace, pt.B, pt.BLocal}} {
			if end.sp != sp.ID {
				continue
			}
			sx, sy := p.ToScreen(sp.MicroPos(end.pos, units.V3(end.local.X, end.local.Y, 0)))
			marks = append(marks, portalMark{
				at:       image.Pt(int(math.Round(float64(sx))), int(math.Round(float64(sy)))),
				disabled: pt.Disabled,
			})
		}
	}
	return &basicSurface{id: sp.ID, rect: rect, climb: climb, portals: marks}
}

func drawBackdrop(dst *ebiten.Image, b *basicSurface) {
	r := b.rect.Intersect(dst.Bounds())
	if r.Empty() {
		return
	}
	fill := color.NRGBA{R: 74, G: 82, B: 94, A: 255}
	border := color.NRGBA{R: 132, G: 146, B: 164, A: 255}
	if b.climb {
		// Tira vertical (pata/respaldo): tono madera cálida con borde definido.
		fill = color.NRGBA{R: 184, G: 134, B: 84, A: 255}
		border = color.NRGBA{R: 112, G: 74, B: 42, A: 255}
	}
	x, y, w, h := float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy())
	vector.FillRect(dst, x, y, w, h, fill, false)
	vector.StrokeRect(dst, x, y, w, h, 1, border, false)
	for _, m := range b.portals {
		c := color.NRGBA{R: 118, G: 216, B: 203, A: 255}
		if m.disabled {
			c = color.NRGBA{R: 196, G: 96, B: 96, A: 255}
		}
		vector.FillRect(dst, float32(m.at.X-4), float32(m.at.Y-2), 8, 4, c, false)
	}
}

// drawCutaway dibuja el techo/capa frontal semitransparente de una cavidad
// dejando un recorte limpio alrededor del tiny seguido.
func drawCutaway(dst *ebiten.Image, m *cutawayMask) {
	outer := m.outer.Intersect(dst.Bounds())
	if outer.Empty() {
		return
	}
	hole := m.hole.Intersect(outer)
	shade := color.NRGBA{R: 28, G: 24, B: 38, A: 115}
	fill := func(r image.Rectangle) {
		if !r.Empty() {
			vector.FillRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), shade, false)
		}
	}
	if hole.Empty() {
		fill(outer)
		return
	}
	fill(image.Rect(outer.Min.X, outer.Min.Y, outer.Max.X, hole.Min.Y))
	fill(image.Rect(outer.Min.X, hole.Max.Y, outer.Max.X, outer.Max.Y))
	fill(image.Rect(outer.Min.X, hole.Min.Y, hole.Min.X, hole.Max.Y))
	fill(image.Rect(hole.Max.X, hole.Min.Y, outer.Max.X, hole.Max.Y))
}
