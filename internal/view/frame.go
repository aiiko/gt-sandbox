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
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
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
}

// Thing es una entidad ubicada que no es un Space (un tiny, un objeto).
type Thing struct {
	ID     ids.ID       `json:"id"`
	Name   string       `json:"name,omitempty"`
	Space  ids.ID       `json:"space"`
	Pose   Pose         `json:"pose"`
	Radius units.Length `json:"radius"`
	Height units.Length `json:"height"`
	Body   bool         `json:"body"`
	Mover  *Mover       `json:"mover,omitempty"`
}

// Mover resume un nav.Mover. Route son los puntos de mundo que faltan
// recorrer, empezando por la posición actual.
type Mover struct {
	Status string       `json:"status"`
	Route  []units.Vec3 `json:"route,omitempty"`
}

// Portal es un portal con sus extremos en coordenadas de mundo.
type Portal struct {
	ID       ids.ID     `json:"id"`
	A        units.Vec3 `json:"a"`
	B        units.Vec3 `json:"b"`
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

// Frame es lo que hay que dibujar de un tick. Spaces va de abajo hacia
// arriba (por altura de mundo y luego ID); el resto, por ID.
type Frame struct {
	Tick     clock.Tick `json:"tick"`
	Spaces   []Space    `json:"spaces"`
	Things   []Thing    `json:"things"`
	Portals  []Portal   `json:"portals"`
	Contacts []Contact  `json:"contacts"`
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

	grids := ecs.Get[nav.Grid](r)
	surfaces := ecs.Get[attach.Surface](r)
	seen := map[ids.ID]bool{}
	ecs.Get[space.Space](r).Each(func(id ids.ID, sp *space.Space) {
		p, ok := pose(id)
		if !ok {
			return
		}
		s := Space{ID: id, Name: names(id), Kind: sp.Kind, Parent: g.Parent(id), Pose: p,
			Extent: sp.Extent, Surface: surfaces.Has(id)}
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

	movers := ecs.Get[nav.Mover](r)
	spaces := ecs.Get[space.Space](r)
	cs.Location.Each(func(id ids.ID, l *comp.Location) {
		if spaces.Has(id) {
			return
		}
		p, ok := pose(id)
		if !ok {
			return
		}
		t := Thing{ID: id, Name: names(id), Space: l.Space, Pose: p}
		if b, ok := cs.Body.Get(id); ok {
			t.Radius, t.Height, t.Body = b.Radius, b.Height, true
		}
		if m, ok := movers.Get(id); ok {
			t.Mover = &Mover{Status: m.Status.String()}
			if m.Status == nav.Moving {
				t.Mover.Route = route(g, p.Pos, m.Legs)
			}
		}
		f.Things = append(f.Things, t)
	})

	contacts := ecs.Get[attach.Contact](r)
	ecs.Get[nav.Portal](r).Each(func(id ids.ID, p *nav.Portal) {
		a, errA := g.ToWorld(p.A.Space, units.V3(p.A.Pos.X, p.A.Pos.Y, 0))
		b, errB := g.ToWorld(p.B.Space, units.V3(p.B.Pos.X, p.B.Pos.Y, 0))
		if errA != nil || errB != nil {
			return
		}
		f.Portals = append(f.Portals, Portal{ID: id, A: a, B: b, Disabled: p.Disabled, Dynamic: contacts.Has(id)})
	})
	contacts.Each(func(id ids.ID, k *attach.Contact) {
		f.Contacts = append(f.Contacts, Contact{ID: id, A: k.A, B: k.B, Since: k.Since})
	})
	byID := func(a, b ids.ID) int { return cmp.Compare(a, b) }
	slices.SortFunc(f.Things, func(a, b Thing) int { return byID(a.ID, b.ID) })
	slices.SortFunc(f.Portals, func(a, b Portal) int { return byID(a.ID, b.ID) })
	slices.SortFunc(f.Contacts, func(a, b Contact) int { return byID(a.ID, b.ID) })
	return f
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
