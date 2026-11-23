// Package view arma, a partir del mundo, una foto plana y serializable de lo
// que hay que dibujar (Frame), y controla el avance del tiempo para visores
// interactivos (Runner).
//
// No importa Ebitengine: el renderer dibuja solo desde un Frame, así que otro
// visor (un inspector web, un test) puede usar los mismos datos.
package view

import (
	"cmp"
	"slices"

	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/encounter"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Pose es una posición y orientación en coordenadas de mundo.
type Pose struct {
	Pos units.Vec3      `json:"pos"`
	Rot fixedmath.Angle `json:"rot"`
}

// Transform devuelve la pose como transformación local→mundo.
func (p Pose) Transform() fixedmath.Transform { return fixedmath.Transform{Pos: p.Pos, Rot: p.Rot} }

// Grid es una copia de una grilla de navegación, en coordenadas del Space.
type Grid struct {
	Origin  units.Vec2   `json:"origin"`
	Cell    units.Length `json:"cell"`
	W       int          `json:"w"`
	H       int          `json:"h"`
	Version uint64       `json:"version"`
	Cost    []nav.Cost   `json:"cost"` // fila por fila, desde y = 0
}

// Grip es un punto de agarre (costura, puño, mechón) en coordenadas locales del Space.
type Grip struct {
	Pos   units.Vec2 `json:"pos"`
	Label string     `json:"label,omitempty"`
}

// Space es un Space con su pose de mundo.
type Space struct {
	ID      ids.ID     `json:"id"`
	Name    string     `json:"name,omitempty"`
	Kind    space.Kind `json:"kind"`
	Parent  ids.ID     `json:"parent"`
	Pose    Pose       `json:"pose"`
	Extent  units.Rect `json:"extent"`
	Grid    *Grid      `json:"grid,omitempty"`
	Surface bool       `json:"surface,omitempty"` // tiene contact_surface
	Depth   int        `json:"depth"`             // 0 = raíz
	Mover   *Mover     `json:"mover,omitempty"`
	// Activity y Step describen lo que hace un agente que es un Space (la giantess).
	Activity string `json:"activity,omitempty"`
	Step     string `json:"step,omitempty"`
	Goal     ids.ID `json:"goal,omitempty"`
	Needs    []Need `json:"needs,omitempty"`
	// Climb, Occluded, Projected y Grips informan a la vista micro sobre zonas
	// corporales (trepadas de costado, ropa/pelo que tapan, proyección 3D y agarres).
	Climb     bool   `json:"climb,omitempty"`
	Occluded  bool   `json:"occluded,omitempty"`
	Projected bool   `json:"projected,omitempty"`
	Grips     []Grip `json:"grips,omitempty"`
}

// MicroPos devuelve la posición sobre el plano micro del Space: en superficies
// con proyección 3D (trepadas o planos inclinados del cuerpo) proyecta las
// coordenadas locales 2D sobre la baldosa dibujada; en superficies planas
// devuelve worldPos sin cambios.
func (s Space) MicroPos(worldPos, localPos units.Vec3) units.Vec3 {
	if !s.Projected {
		return worldPos
	}
	return s.Pose.Transform().Apply(units.V3(localPos.X, localPos.Y, 0))
}

// Expedition resume una expedición activa de la colonia para la interfaz.
type Expedition struct {
	ID       ids.ID   `json:"id"`
	Target   ids.ID   `json:"target"`
	Name     string   `json:"name,omitempty"`
	Phase    string   `json:"phase"`
	Guide    ids.ID   `json:"guide,omitempty"`
	Members  []ids.ID `json:"members,omitempty"`
	Reason   string   `json:"reason,omitempty"`
	Findings int      `json:"findings,omitempty"`
}

// Colony resume el estado de una colonia (despensa, obra y expedición) para
// la interfaz de partida.
type Colony struct {
	ID          ids.ID      `json:"id"`
	Name        string      `json:"name,omitempty"`
	Space       ids.ID      `json:"space"`
	Portions    int32       `json:"portions"`
	Goal        int32       `json:"goal"` // meta de porciones (Stock * miembros vivos)
	Members     int         `json:"members"`
	Site        *Site       `json:"site,omitempty"`
	Expedition  *Expedition `json:"expedition,omitempty"`
	Stranded    int         `json:"stranded,omitempty"`
	NoExpReason string      `json:"no_exp_reason,omitempty"`
}

// Site es el avance de la obra del refugio auxiliar de una colonia.
type Site struct {
	ID        ids.ID `json:"id"`
	Name      string `json:"name,omitempty"`
	Space     ids.ID `json:"space"`
	Delivered int32  `json:"delivered"`
	Required  int32  `json:"required"`
	Done      bool   `json:"done,omitempty"`
}

// Need es el valor de una necesidad de un agente, en [0, needs.Max].
type Need struct {
	Name  string `json:"name"`
	Value int32  `json:"value"`
}

// Thing es una entidad ubicada que no es un Space (un tiny, un objeto).
type Thing struct {
	ID        ids.ID       `json:"id"`
	Name      string       `json:"name,omitempty"`
	Space     ids.ID       `json:"space"`
	Pose      Pose         `json:"pose"`
	LocalPos  units.Vec3   `json:"local_pos,omitempty"`
	Radius    units.Length `json:"radius"`
	Height    units.Length `json:"height"`
	Body      bool         `json:"body"`
	Mover     *Mover       `json:"mover,omitempty"`
	CarriedBy ids.ID       `json:"carried_by,omitempty"`
	// Activity resume lo que hace un agente ("use:eat", "work", "work (suspendida)").
	Activity string `json:"activity,omitempty"`
	// Step es el paso en curso de una oferta de varios pasos ("goto", "pose", "use").
	Step  string `json:"step,omitempty"`
	Needs []Need `json:"needs,omitempty"`
	// Goal es sobre qué actúa un agente: el objeto de la oferta, el punto
	// que revisa, el punto de reunión o el objetivo de su trabajo (Job es el
	// tipo de trabajo). With es con quién conversa.
	Goal ids.ID `json:"goal,omitempty"`
	Job  string `json:"job,omitempty"`
	With ids.ID `json:"with,omitempty"`
	// Built: la obra de un refugio auxiliar ya está terminada.
	Built bool `json:"built,omitempty"`
	// Remains es la causa ("hand", "foot", "object") si son los restos de un
	// aplastado. Alarmed: un agente percibió una amenaza y todavía no huye o
	// huye; Trapped: no encontró salida.
	Remains       string `json:"remains,omitempty"`
	Alarmed       bool   `json:"alarmed,omitempty"`
	Trapped       bool   `json:"trapped,omitempty"`
	Role          string `json:"role,omitempty"`
	Why           string `json:"why,omitempty"`
	Rations       int32  `json:"rations,omitempty"`
	Stranded      bool   `json:"stranded,omitempty"`
	NearestRefuge string `json:"nearest_refuge,omitempty"`
	GripLabel     string `json:"grip_label,omitempty"`
}

// ExpeditionMark es un hito o pista de expedición para el miniplano en micro.
type ExpeditionMark struct {
	ID        ids.ID     `json:"id"`
	Pos       units.Vec3 `json:"pos"`
	Confirmed bool       `json:"confirmed,omitempty"`
}

// Footprint es la próxima pisada prevista de quien camina: centro en el
// mundo, dirección de la marcha (en el mundo, sin normalizar) y tamaño del
// pie. La presentación dibuja su sombra antes de que baje.
type Footprint struct {
	Walker ids.ID       `json:"walker"`
	Space  ids.ID       `json:"space"`
	Pos    units.Vec3   `json:"pos"`
	Dir    units.Vec2   `json:"dir"`
	Length units.Length `json:"length"`
	Width  units.Length `json:"width"`
	// Stride y Remaining: el largo del paso y cuánto falta para que el pie
	// baje (la sombra se oscurece al acercarse).
	Stride    units.Length `json:"stride"`
	Remaining units.Length `json:"remaining"`
}

// Step es una pisada reciente: el pie apoyado en Pos (mundo), mirando hacia
// Dir, desde Tick. Se dibuja durante StepTicks.
type Step struct {
	Walker ids.ID     `json:"walker"`
	Space  ids.ID     `json:"space"`
	Pos    units.Vec3 `json:"pos"`
	Dir    units.Vec2 `json:"dir"`
	Tick   clock.Tick `json:"tick"`
}

// StepTicks es cuánto se ve un pie apoyado.
const StepTicks clock.Tick = 8

// Mover resume un nav.Mover. Route son los puntos de mundo que faltan
// recorrer, empezando por la posición actual.
type Mover struct {
	Status string       `json:"status"`
	Route  []units.Vec3 `json:"route,omitempty"`
}

// Portal es un portal con sus extremos en coordenadas de mundo y el Space de
// cada extremo.
type Portal struct {
	ID       ids.ID     `json:"id"`
	A        units.Vec3 `json:"a"`
	B        units.Vec3 `json:"b"`
	ALocal   units.Vec2 `json:"a_local,omitempty"`
	BLocal   units.Vec2 `json:"b_local,omitempty"`
	ASpace   ids.ID     `json:"a_space"`
	BSpace   ids.ID     `json:"b_space"`
	Disabled bool       `json:"disabled,omitempty"`
	Dynamic  bool       `json:"dynamic,omitempty"` // lo creó un contacto
}

// Contact es un contacto activo entre dos superficies.
type Contact struct {
	ID    ids.ID     `json:"id"`
	A     ids.ID     `json:"a"`
	B     ids.ID     `json:"b"`
	Since clock.Tick `json:"since"`
}

// Fall es una caída reciente (evento entity_fell). La simulación la resuelve
// en un tick; la presentación la anima durante FallTicks.
type Fall struct {
	Entity ids.ID       `json:"entity"`
	From   ids.ID       `json:"from"`
	To     ids.ID       `json:"to"`
	Height units.Length `json:"height"`
	Tick   clock.Tick   `json:"tick"` // primer frame en que el tiny ya está en To
}

// FallTicks es cuánto dura la animación de una caída.
const FallTicks = 10

// Frame es lo que hay que dibujar de un tick. Spaces va de abajo hacia
// arriba (por altura de mundo y luego ID); el resto, por ID. Falls la
// completa el Runner con las caídas de los últimos FallTicks.
type Frame struct {
	Tick     clock.Tick `json:"tick"`
	Spaces   []Space    `json:"spaces"`
	Things   []Thing    `json:"things"`
	Portals  []Portal   `json:"portals"`
	Contacts []Contact  `json:"contacts"`
	Falls    []Fall     `json:"falls,omitempty"`
	Cues     []Cue      `json:"cues,omitempty"`
	// Footprints son las pisadas previstas (sombra del pie).
	Footprints []Footprint `json:"footprints,omitempty"`
	// Steps son las pisadas recientes (el pie apoyado).
	Steps []Step `json:"steps,omitempty"`
	// Colonies resume el estado de cada colonia (despensa y refugio).
	Colonies []Colony `json:"colonies,omitempty"`
	// Marks son pistas vistas y destinos confirmados para el miniplano en micro.
	Marks []ExpeditionMark `json:"marks,omitempty"`
}

// FallOf devuelve la caída en curso de una entidad.
func (f *Frame) FallOf(id ids.ID) (Fall, bool) {
	for _, fl := range f.Falls {
		if fl.Entity == id && f.Tick-fl.Tick < FallTicks {
			return fl, true
		}
	}
	return Fall{}, false
}

// SpaceByID busca un Space del frame.
func (f *Frame) SpaceByID(id ids.ID) (*Space, bool) {
	for i := range f.Spaces {
		if f.Spaces[i].ID == id {
			return &f.Spaces[i], true
		}
	}
	return nil, false
}

// ThingByID busca una Thing del frame.
func (f *Frame) ThingByID(id ids.ID) (*Thing, bool) {
	for i := range f.Things {
		if f.Things[i].ID == id {
			return &f.Things[i], true
		}
	}
	return nil, false
}

// Capturer arma Frames. Guarda copias de las grillas por versión para no
// copiarlas en cada frame.
type Capturer struct {
	grids map[ids.ID]*Grid
}

// Capture arma el Frame del estado actual de w. Solo lee.
func (c *Capturer) Capture(w *world.World) *Frame {
	if c.grids == nil {
		c.grids = map[ids.ID]*Grid{}
	}
	r := w.ECS()
	g := space.Of(r)
	cs := comp.Of(r)
	names := func(id ids.ID) string { n, _ := cs.Name.Get(id); return string(n) }
	pose := func(id ids.ID) (Pose, bool) {
		p, rot, err := g.WorldPose(id)
		return Pose{p, rot}, err == nil
	}
	f := &Frame{Tick: w.Tick()}

	movers := ecs.Get[nav.Mover](r)
	mover := func(id ids.ID, pos units.Vec3) *Mover {
		m, ok := movers.Get(id)
		if !ok {
			return nil
		}
		out := &Mover{Status: m.Status.String()}
		if m.Status == nav.Moving {
			out.Route = route(g, pos, m.Legs)
		}
		return out
	}
	agents := ecs.Get[ai.Agent](r)
	grids := ecs.Get[nav.Grid](r)
	surfaces := ecs.Get[attach.Surface](r)
	zones := ecs.Get[attach.BodyZone](r)
	seen := map[ids.ID]bool{}
	ecs.Get[space.Space](r).Each(func(id ids.ID, sp *space.Space) {
		p, ok := pose(id)
		if !ok {
			return
		}
		s := Space{ID: id, Name: names(id), Kind: sp.Kind, Parent: g.Parent(id), Pose: p,
			Extent: sp.Extent, Surface: surfaces.Has(id), Mover: mover(id, p.Pos),
			Climb: sp.IsClimb(), Projected: sp.Projection != nil}
		if z, ok := zones.Get(id); ok {
			s.Occluded = z.Occluded
			if len(z.Grips) > 0 {
				s.Grips = make([]Grip, len(z.Grips))
				for i, gr := range z.Grips {
					s.Grips[i] = Grip{Pos: gr.Pos, Label: gr.Label}
				}
			}
		}
		if a, ok := agents.Get(id); ok {
			s.Activity, s.Step = activity(r, a)
			s.Goal, _, _ = purpose(r, id, a)
		}
		s.Needs = needsOf(r, id)
		for e := s.Parent; !e.IsNil() && s.Depth < space.MaxDepth; e = g.Parent(e) {
			s.Depth++
		}
		if gr, ok := grids.Get(id); ok {
			s.Grid = c.grid(id, &gr)
			seen[id] = true
		}
		f.Spaces = append(f.Spaces, s)
	})
	for id := range c.grids { // borrar copias de grillas que ya no existen
		if !seen[id] {
			delete(c.grids, id)
		}
	}
	slices.SortFunc(f.Spaces, func(a, b Space) int {
		return cmp.Or(cmp.Compare(a.Pose.Pos.Z, b.Pose.Pos.Z), cmp.Compare(a.ID, b.ID))
	})

	carried := ecs.Get[jobs.Carried](r)
	spaces := ecs.Get[space.Space](r)
	remains := ecs.Get[crush.Remains](r)
	cs.Location.Each(func(id ids.ID, l *comp.Location) {
		if spaces.Has(id) {
			return
		}
		p, ok := pose(id)
		if !ok {
			return
		}
		t := Thing{ID: id, Name: names(id), Space: l.Space, Pose: p, LocalPos: l.Pos}
		if b, ok := cs.Body.Get(id); ok {
			t.Radius, t.Height, t.Body = b.Radius, b.Height, true
		}
		t.Mover = mover(id, p.Pos)
		if c, ok := carried.Get(id); ok {
			t.CarriedBy = c.By
		}
		if a, ok := agents.Get(id); ok {
			t.Activity, t.Step = activity(r, a)
			t.Goal, t.Job, t.With = purpose(r, id, a)
			t.Alarmed = a.Reflex > 0 && a.Alarm > 0
			t.Trapped = a.Trapped
		}
		if b, ok := ecs.Get[colony.BuildSite](r).Get(id); ok {
			t.Built = b.Done
		}
		if m, ok := ecs.Get[colony.Member](r).Get(id); ok {
			t.Role = m.Role
			t.Why = colony.JobReason(w, id)
		}
		if pr, ok := ecs.Get[colony.Provisions](r).Get(id); ok {
			t.Rations = pr.Rations
		}
		t.Stranded = ecs.Get[colony.Stranded](r).Has(id)
		if t.Stranded || t.Activity == colony.ExpeditionActionName {
			t.NearestRefuge = nearestShelterName(r, g, names, l.Space, p.Pos)
		}
		if m, ok := remains.Get(id); ok {
			t.Remains = m.Source
		}
		if gr, ok := ecs.Get[attach.Gripping](r).Get(id); ok {
			t.GripLabel = gr.Label
		}
		t.Needs = needsOf(r, id)
		f.Things = append(f.Things, t)
	})
	strides := ecs.Get[crush.Stride](r)
	for _, st := range crush.NextSteps(w) {
		at, err1 := g.ToWorld(st.Space, units.V3(st.Center.X, st.Center.Y, 0))
		ahead, err2 := g.ToWorld(st.Space, units.V3(st.Center.X+st.Dir.X, st.Center.Y+st.Dir.Y, 0))
		if err1 == nil && err2 == nil {
			f.Footprints = append(f.Footprints, Footprint{Walker: st.Entity, Space: st.Space, Pos: at,
				Dir: ahead.Sub(at).XY(), Length: st.Length, Width: st.Width, Remaining: st.Remaining})
			if s, ok := strides.Get(st.Entity); ok {
				f.Footprints[len(f.Footprints)-1].Stride = s.Length
			}
		}
	}

	contacts := ecs.Get[attach.Contact](r)
	ecs.Get[nav.Portal](r).Each(func(id ids.ID, p *nav.Portal) {
		a, errA := g.ToWorld(p.A.Space, units.V3(p.A.Pos.X, p.A.Pos.Y, 0))
		b, errB := g.ToWorld(p.B.Space, units.V3(p.B.Pos.X, p.B.Pos.Y, 0))
		if errA != nil || errB != nil {
			return
		}
		f.Portals = append(f.Portals, Portal{ID: id, A: a, B: b, ALocal: p.A.Pos, BLocal: p.B.Pos,
			ASpace: p.A.Space, BSpace: p.B.Space, Disabled: p.Disabled, Dynamic: contacts.Has(id)})
	})
	contacts.Each(func(id ids.ID, k *attach.Contact) {
		f.Contacts = append(f.Contacts, Contact{ID: id, A: k.A, B: k.B, Since: k.Since})
	})

	// Colonias: despensa, meta y avance de la obra del refugio.
	colMembers := map[ids.ID]int{}
	ecs.Get[colony.Member](r).Each(func(e ids.ID, m *colony.Member) {
		if !r.Despawning(e) {
			colMembers[m.Colony]++
		}
	})
	colSites := map[ids.ID]*Site{}
	ecs.Get[colony.BuildSite](r).Each(func(id ids.ID, b *colony.BuildSite) {
		if _, exists := colSites[b.Colony]; exists {
			return
		}
		var sp ids.ID
		if l, ok := cs.Location.Get(id); ok {
			sp = l.Space
		}
		colSites[b.Colony] = &Site{ID: id, Name: names(id), Space: sp, Delivered: b.Delivered, Required: b.Required, Done: b.Done}
	})
	portions := ecs.Get[ai.Portions](r)
	strandedStore := ecs.Get[colony.Stranded](r)
	ecs.Get[colony.Colony](r).Each(func(id ids.ID, col *colony.Colony) {
		var sp ids.ID
		if l, ok := cs.Location.Get(id); ok {
			sp = l.Space
		}
		p, _ := portions.Get(id)
		m := colMembers[id]
		strandedCount := 0
		strandedStore.Each(func(e ids.ID, st *colony.Stranded) {
			if st.Colony == id && !r.Despawning(e) {
				strandedCount++
			}
		})
		var expView *Expedition
		if expID, exp := colony.ActiveExpedition(w, id); exp != nil {
			var mems []ids.ID
			for _, em := range exp.Members {
				if !em.Returned && !em.Stranded {
					mems = append(mems, em.Entity)
				}
			}
			expView = &Expedition{
				ID: expID, Target: exp.Target, Name: names(exp.Target),
				Phase: string(exp.Phase), Guide: exp.Guide, Members: mems,
				Reason: exp.Reason, Findings: len(exp.Findings),
			}
		}
		f.Colonies = append(f.Colonies, Colony{
			ID: id, Name: names(id), Space: sp,
			Portions: p.Left, Goal: col.Stock * int32(m), Members: m,
			Site: colSites[id], Expedition: expView, Stranded: strandedCount,
			NoExpReason: colony.NoExpeditionReason(w, id),
		})
	})
	ecs.Get[colony.ExpeditionTarget](r).Each(func(id ids.ID, t *colony.ExpeditionTarget) {
		if r.Despawning(id) {
			return
		}
		if p, ok := pose(id); ok {
			f.Marks = append(f.Marks, ExpeditionMark{ID: id, Pos: p.Pos, Confirmed: t.Visited})
		}
	})
	byID := func(a, b ids.ID) int { return cmp.Compare(a, b) }
	slices.SortFunc(f.Things, func(a, b Thing) int { return byID(a.ID, b.ID) })
	slices.SortFunc(f.Portals, func(a, b Portal) int { return byID(a.ID, b.ID) })
	slices.SortFunc(f.Contacts, func(a, b Contact) int { return byID(a.ID, b.ID) })
	slices.SortFunc(f.Colonies, func(a, b Colony) int { return byID(a.ID, b.ID) })
	slices.SortFunc(f.Marks, func(a, b ExpeditionMark) int { return byID(a.ID, b.ID) })
	return f
}

func nearestShelterName(r *ecs.Registry, g space.Graph, names func(ids.ID) string, sp ids.ID, pos units.Vec3) string {
	var bestName string
	var bestD int64
	found := false
	ecs.Get[crush.Shelter](r).Each(func(id ids.ID, _ *crush.Shelter) {
		if r.Despawning(id) {
			return
		}
		wp, _, err := g.WorldPose(id)
		if err != nil {
			return
		}
		d := pos.XY().DistSq(wp.XY())
		if !found || d < bestD {
			bestName, bestD, found = names(id), d, true
		}
	})
	return bestName
}

func (c *Capturer) grid(id ids.ID, g *nav.Grid) *Grid {
	if v, ok := c.grids[id]; ok && v.Version == g.Version && v.W == g.W && v.H == g.H &&
		v.Origin == g.Origin && v.Cell == g.Cell {
		return v
	}
	v := &Grid{Origin: g.Origin, Cell: g.Cell, W: g.W, H: g.H, Version: g.Version, Cost: make([]nav.Cost, g.W*g.H)}
	for y := range g.H {
		for x := range g.W {
			v.Cost[y*g.W+x] = g.At(x, y)
		}
	}
	c.grids[id] = v
	return v
}

// activity resume la acción de un agente y, si la oferta tiene pasos, el
// paso en curso. Una acción suspendida no tiene paso.
func needsOf(r *ecs.Registry, id ids.ID) []Need {
	n, ok := ecs.Get[needs.Needs](r).Get(id)
	if !ok {
		return nil
	}
	out := make([]Need, len(n))
	for i, x := range n {
		out[i] = Need{Name: x.Name, Value: x.Value}
	}
	return out
}

// NeedOf devuelve el valor de una necesidad de la lista.
func NeedOf(list []Need, name string) (int32, bool) {
	for _, n := range list {
		if n.Name == name {
			return n.Value, true
		}
	}
	return 0, false
}

func activity(r *ecs.Registry, a ai.Agent) (string, string) {
	d, suffix := a.Doing, ""
	if d == nil && a.Suspended != nil {
		d, suffix = a.Suspended, " (suspendida)"
	}
	if d == nil {
		return "", ""
	}
	name, step := d.Action, ""
	if encounter.IsAction(d.Action) && d.Stage == 0 && suffix == "" {
		step = "goto" // va hacia la superficie
	}
	if suffix == "" {
		switch d.Action {
		case colony.SearchAction:
			step = [...]string{"goto", "look"}[min(d.Stage, 1)]
		case colony.ChatAction:
			step = [...]string{"goto", "wait", "talk"}[min(d.Stage, 2)]
		}
	}
	if os, ok := ecs.Get[ai.Offers](r).Get(d.Target); ok && d.Action == "use" && d.Offer < len(os) {
		o := os[d.Offer]
		name += ":" + o.Action
		if suffix == "" && d.Stage < len(o.Steps) {
			step = o.Steps[d.Stage].Do
		}
	}
	return name + suffix, step
}

// purpose dice sobre qué actúa el agente id: el objetivo de su acción (o
// de su trabajo, con el tipo de trabajo) y con quién conversa.
func purpose(r *ecs.Registry, id ids.ID, a ai.Agent) (goal ids.ID, job string, with ids.ID) {
	d := a.Doing
	if d == nil {
		d = a.Suspended
	}
	if d == nil {
		return ids.Nil, "", ids.Nil
	}
	goal = d.Target
	if j, ok := ecs.Get[jobs.Job](r).Get(d.Target); ok && d.Action == "work" {
		goal, job = j.Target, j.Type
	}
	if g, ok := ecs.Get[colony.Gathering](r).Get(d.Target); ok && d.Action == colony.ChatAction {
		for _, s := range g.Seats {
			if s.By != id {
				with = s.By
			}
		}
	}
	return goal, job, with
}

// route pasa los tramos pendientes a puntos de mundo.
func route(g space.Graph, from units.Vec3, legs []nav.Leg) []units.Vec3 {
	out := []units.Vec3{from}
	for _, l := range legs {
		for _, p := range l.Path {
			if w, err := g.ToWorld(l.Space, units.V3(p.X, p.Y, 0)); err == nil {
				out = append(out, w)
			}
		}
	}
	return out
}
