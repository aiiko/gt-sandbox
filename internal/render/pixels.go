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

// look ajusta cómo se dibuja una colocación.
type look struct {
	visual string  // "" = el visual de la entidad
	alpha  float32 // 0 = opaco
	ghost  bool
	layer  int  // 0 = la capa de la variante
	top    bool // encima de todo (el reflejo de lo seguido bajo una superficie más alta)
}

const (
	// PassVisual es el visual de la señal de un paso entre superficies en contacto.
	PassVisual = "pass"
	// FootprintVisual es el visual de la sombra de la próxima pisada (macro).
	FootprintVisual = "footprint"
	// FootVisual es el pie de la giantess visto de cerca: su sombra antes de
	// bajar (shadow_<dir>) y el pie apoyado (foot_<dir>).
	FootVisual = "foot"
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
	atlas   *Atlas
	sprites []sprite
	clips   map[ids.ID]string
	marks   map[ids.ID][]string
	lifts   map[ids.ID]int
}

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
			}
		}
		if scale == Micro {
			if isSpace {
				layer = layerBackground
			} else if sp, ok := f.SpaceByID(support); ok {
				group, height = sp.ID, sp.Pose.Pos.Z
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
			if dest.Overlaps(image.Rect(0, 0, CanvasWidth, CanvasHeight)) {
				out.sprites = append(out.sprites, sprite{id: id, src: src, dest: dest, layer: layer, y: depth, group: group, z: height,
					alpha: lk.alpha, ghost: lk.ghost, space: isSpace})
			}
		}
		addFrame(v.frame(name, uint64(f.Tick)), layer)
		if v.Foreground != nil {
			addFrame(frameRect(v.Foreground), layerForeground)
		}
	}
	for _, s := range f.Spaces {
		if !inFocus(s.ID) {
			continue
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
		add(s.ID, s.Parent, ids.Nil, s.Pose.Pos, names, true, image.Point{}, look{})
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
		for _, pt := range f.Portals {
			if !pt.Dynamic {
				continue
			}
			state := []string{"open"}
			if pt.Disabled {
				state = []string{"closed"}
			}
			for _, end := range []struct {
				pos   units.Vec3
				space ids.ID
			}{{pt.A, pt.ASpace}, {pt.B, pt.BSpace}} {
				if inFocus(end.space) {
					add(pt.ID, end.space, ids.Nil, end.pos, state, false, image.Point{},
						look{visual: PassVisual, ghost: true, layer: layerGround})
				}
			}
		}
		// Sombra anticipada de una superficie móvil que todavía no se apoya
		// sobre la superficie enfocada.
		if fs, ok := f.SpaceByID(focus); ok {
			for _, s := range f.Spaces {
				if s.Mover != nil || s.Extent.Empty() || inFocus(s.ID) {
					continue
				}
				if lk, offset, ok := handShadow(fs, s); ok {
					add(s.ID, fs.ID, ids.Nil, s.Pose.Pos, []string{"shadow_" + compass(s.Pose.Rot)}, false, offset, lk)
				}
			}
		}
	}
	for _, t := range f.Things {
		if t.CarriedBy.IsNil() && inFocus(t.Space) {
			names := actor{mover: t.Mover, facing: t.Pose.Rot, activity: t.Activity, step: t.Step,
				carrying: carrying[t.ID]}.clips()
			var offset image.Point
			if fl, ok := f.FallOf(t.ID); ok {
				// Cae desde arriba con aceleración; su sombra lo espera en el apoyo.
				lift := fallOffset(p, fl, f.Tick)
				names, offset = append([]string{"fall"}, names...), image.Pt(0, -lift)
				add(t.ID, t.Space, ids.Nil, t.Pose.Pos, []string{"shadow"}, false, image.Point{},
					look{ghost: true, layer: layerGround, alpha: 0.3 + 0.5*float32(fallLift-lift)/fallLift})
			}
			add(t.ID, t.Space, ids.Nil, t.Pose.Pos, names, false, offset, look{})
			if scale == Micro && slices.Contains(featured, t.ID) {
				// Lo seguido se ve a través de una superficie que lo tapa (la
				// mano apoyada encima del extremo del paso).
				add(t.ID, t.Space, ids.Nil, t.Pose.Pos, names, false, offset, look{ghost: true, top: true, alpha: 0.45})
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
			support, pos = carrier.Space, carrier.Pose.Pos
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
	canvas := image.Rect(0, 0, CanvasWidth, CanvasHeight)
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
			group, depth = s.ID, y+float32(v.DepthOffset)
		}
		support = s.Parent
	}
	return group, depth
}

// poseClips relaciona la acción de una oferta en uso con la pose que la dibuja.
var poseClips = map[string]string{
	"breakfast": "eat",
	"snack":     "eat",
	"eat":       "eat",
	"nibble":    "eat",
	"watch":     "sit",
}

// actor es el estado de una entidad del que sale su clip.
type actor struct {
	mover          *view.Mover
	facing         fixedmath.Angle
	activity, step string
	carrying       bool // lleva algo
	resting        bool // tiene una parte apoyada en otra superficie
}

func compass(facing fixedmath.Angle) string {
	directions := [...]string{"east", "north", "west", "south"}
	quarter := fixedmath.QuarterTurn
	return directions[((facing.Norm()+quarter/2)/quarter)%4]
}

// clips devuelve los clips candidatos, del más específico al más general.
func (a actor) clips() []string {
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
