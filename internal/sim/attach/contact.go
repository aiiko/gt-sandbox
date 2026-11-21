// Package attach implementa lo que vuelve a un cuerpo un "mapa móvil":
// contactos entre superficies, portales dinámicos derivados de ellos y poses
// (conjuntos de transforms de las partes de un cuerpo).
package attach

import (
	"cmp"
	"fmt"
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Surface marca un Space que participa de contactos (store
// "contact_surface"): una palma, un escritorio, una mesa.
//
// Dos superficies se tocan si están en la misma raíz, sus alturas en el
// mundo difieren a lo sumo en la mayor de sus Tolerance y sus extensiones se
// solapan en el plano. El contacto abre un portal si el solapamiento mide al
// menos MinOverlap (el mayor de ambos) en cada lado.
type Surface struct {
	Tolerance  units.Length `json:"tolerance"`
	MinOverlap units.Length `json:"min_overlap"`
}

// Valores por defecto cuando Surface los deja en cero.
const (
	DefaultTolerance  = 3 * units.Centimeter
	DefaultMinOverlap = 1 * units.Centimeter
)

func (s Surface) tolerance() units.Length {
	if s.Tolerance > 0 {
		return s.Tolerance
	}
	return DefaultTolerance
}

func (s Surface) minOverlap() units.Length {
	if s.MinOverlap > 0 {
		return s.MinOverlap
	}
	return DefaultMinOverlap
}

// Contact es el componente de la entidad que representa un contacto activo
// entre dos superficies (A < B por ID). Mientras el contacto permita pasar,
// la misma entidad tiene un nav.Portal.
type Contact struct {
	A     ids.ID     `json:"a"`
	B     ids.ID     `json:"b"`
	Since clock.Tick `json:"since"` // tick en que empezó
}

type ContactBegan struct {
	Contact ids.ID `json:"contact"`
	A       ids.ID `json:"a"`
	B       ids.ID `json:"b"`
}

func (ContactBegan) Kind() string { return "contact_began" }

type ContactEnded struct {
	Contact ids.ID `json:"contact"`
	A       ids.ID `json:"a"`
	B       ids.ID `json:"b"`
}

func (ContactEnded) Kind() string { return "contact_ended" }

type PortalOpened struct {
	Portal ids.ID `json:"portal"`
	A      ids.ID `json:"a"`
	B      ids.ID `json:"b"`
}

func (PortalOpened) Kind() string { return "portal_opened" }

type PortalClosed struct {
	Portal ids.ID `json:"portal"`
	A      ids.ID `json:"a"`
	B      ids.ID `json:"b"`
}

func (PortalClosed) Kind() string { return "portal_closed" }

// PortalDrift es cuánto pueden separarse en el mundo los dos extremos de un
// portal de contacto antes de recalcularlo (p. ej. si la mano se desliza).
const PortalDrift = 5 * units.Millimeter

// contacts recalcula todos los contactos en la fase Topology.
type contacts struct{}

func (contacts) Name() string { return "attach.contacts" }

type surf struct {
	id     ids.ID
	s      Surface
	root   ids.ID
	z      units.Length
	bounds units.Rect // AABB de la extensión en el mundo
	extent units.Rect
}

func (contacts) Run(w *world.World) {
	r := w.ECS()
	g := space.Of(r)
	spaces := ecs.Get[space.Space](r)
	cs := ecs.Get[Contact](r)
	portals := ecs.Get[nav.Portal](r)

	var surfs []surf
	ecs.Get[Surface](r).Each(func(id ids.ID, s *Surface) {
		sp, ok := spaces.Get(id)
		if !ok || sp.Extent.Empty() || r.Despawning(id) {
			return
		}
		t, err := g.Frame(id)
		root, err2 := g.Root(id)
		if err != nil || err2 != nil {
			return
		}
		surfs = append(surfs, surf{id, *s, root, t.Pos.Z, worldBounds(t.Apply, sp.Extent), sp.Extent})
	})
	slices.SortFunc(surfs, func(a, b surf) int { return cmp.Compare(a.id, b.id) })

	existing := map[[2]ids.ID]ids.ID{}
	cs.Each(func(e ids.ID, c *Contact) { existing[[2]ids.ID{c.A, c.B}] = e })
	seen := map[ids.ID]bool{}

	for i := range surfs {
		for j := i + 1; j < len(surfs); j++ {
			a, b := surfs[i], surfs[j]
			tol := max(a.s.tolerance(), b.s.tolerance())
			over := a.bounds.Intersect(b.bounds)
			if a.root != b.root || (a.z-b.z).Abs() > tol || over.Empty() {
				continue
			}
			key := [2]ids.ID{a.id, b.id}
			e, ok := existing[key]
			if !ok {
				e = r.Spawn()
				cs.Set(e, Contact{A: a.id, B: b.id, Since: w.Tick()})
				w.Emit(ContactBegan{Contact: e, A: a.id, B: b.id})
			}
			seen[e] = true
			minSide := max(a.s.minOverlap(), b.s.minOverlap())
			var ends *[2]nav.PortalEnd
			if over.Width() >= minSide && over.Height() >= minSide {
				if p := portals.Ptr(e); p != nil && stillValid(g, p) {
					continue
				}
				ends = findEnds(r, g, a, b, over)
			}
			switch p := portals.Ptr(e); {
			case ends == nil && p != nil:
				portals.Remove(e)
				w.Emit(PortalClosed{Portal: e, A: a.id, B: b.id})
			case ends != nil && p == nil:
				portals.Set(e, nav.Portal{A: ends[0], B: ends[1]})
				w.Emit(PortalOpened{Portal: e, A: a.id, B: b.id})
			case ends != nil:
				p.A, p.B = ends[0], ends[1]
				p.Version++
			}
		}
	}

	for _, e := range cs.IDs() {
		if !seen[e] {
			end(w, e)
		}
	}
}

// end termina un contacto: emite los eventos, le quita los componentes (así
// nadie lo termina dos veces) y destruye la entidad.
func end(w *world.World, e ids.ID) {
	r := w.ECS()
	cs, portals := ecs.Get[Contact](r), ecs.Get[nav.Portal](r)
	c, _ := cs.Get(e)
	if portals.Remove(e) {
		w.Emit(PortalClosed{Portal: e, A: c.A, B: c.B})
	}
	w.Emit(ContactEnded{Contact: e, A: c.A, B: c.B})
	cs.Remove(e)
	r.Despawn(e) // si nav.portals ya la marcó, no hace nada
}

func worldBounds(apply func(units.Vec3) units.Vec3, r units.Rect) units.Rect {
	corners := [4]units.Vec2{r.Min, {X: r.Max.X, Y: r.Min.Y}, r.Max, {X: r.Min.X, Y: r.Max.Y}}
	out := units.Rect{Min: units.V2(1<<62, 1<<62), Max: units.V2(-1<<62, -1<<62)}
	for _, c := range corners {
		p := apply(units.V3(c.X, c.Y, 0))
		out.Min.X, out.Min.Y = min(out.Min.X, p.X), min(out.Min.Y, p.Y)
		out.Max.X, out.Max.Y = max(out.Max.X, p.X), max(out.Max.Y, p.Y)
	}
	return out
}

// stillValid: los extremos siguen coincidiendo en el mundo (a menos de
// PortalDrift) y siguen siendo transitables en ambos lados.
func stillValid(g space.Graph, p *nav.Portal) bool {
	wa, err1 := g.ToWorld(p.A.Space, units.V3(p.A.Pos.X, p.A.Pos.Y, 0))
	wb, err2 := g.ToWorld(p.B.Space, units.V3(p.B.Pos.X, p.B.Pos.Y, 0))
	return err1 == nil && err2 == nil && wa.XY().DistSq(wb.XY()) <= int64(PortalDrift)*int64(PortalDrift)
}

// findEnds busca un punto del solapamiento (en el mundo) que caiga dentro de
// ambas extensiones y sea transitable en ambas grillas. Prueba el centro y
// después una grilla de 5×5 muestras, de la más cercana al centro a la más
// lejana (orden determinista).
func findEnds(r *ecs.Registry, g space.Graph, a, b surf, over units.Rect) *[2]nav.PortalEnd {
	center := units.V2((over.Min.X+over.Max.X)/2, (over.Min.Y+over.Max.Y)/2)
	cands := []units.Vec2{center}
	for iy := range 5 {
		for ix := range 5 {
			cands = append(cands, units.V2(
				over.Min.X+over.Width()*units.Length(2*ix+1)/10,
				over.Min.Y+over.Height()*units.Length(2*iy+1)/10))
		}
	}
	slices.SortStableFunc(cands[1:], func(p, q units.Vec2) int {
		return cmp.Compare(p.DistSq(center), q.DistSq(center))
	})
	grids := ecs.Get[nav.Grid](r)
	local := func(s surf, p units.Vec2) (units.Vec2, bool) {
		l, err := g.FromWorld(s.id, units.V3(p.X, p.Y, s.z))
		if err != nil || !s.extent.Contains(l.XY()) {
			return units.Vec2{}, false
		}
		if grid := grids.Ptr(s.id); grid != nil && !grid.Walkable(l.XY()) {
			return units.Vec2{}, false
		}
		return l.XY(), true
	}
	for _, p := range cands {
		la, okA := local(a, p)
		lb, okB := local(b, p)
		if okA && okB {
			return &[2]nav.PortalEnd{{Space: a.id, Pos: la}, {Space: b.id, Pos: lb}}
		}
	}
	return nil
}

// contactCleanup termina los contactos cuyas superficies se destruyen,
// emitiendo sus eventos. Corre al final del Flush.
type contactCleanup struct{}

func (contactCleanup) Name() string { return "attach.cleanup" }

func (contactCleanup) Run(w *world.World) {
	r := w.ECS()
	g := space.Of(r)
	gone := func(s ids.ID) bool { return !g.IsSpace(s) || r.Despawning(s) }
	for _, e := range ecs.Get[Contact](r).IDs() {
		c, _ := ecs.Get[Contact](r).Get(e)
		if gone(c.A) || gone(c.B) {
			end(w, e)
		}
	}
}

func checkContacts(w *world.World) error {
	g := space.Of(w.ECS())
	var err error
	ecs.Get[Contact](w.ECS()).Each(func(e ids.ID, c *Contact) {
		if err == nil && (!g.IsSpace(c.A) || !g.IsSpace(c.B) || c.A >= c.B) {
			err = fmt.Errorf("contacto %v inválido: %+v", e, *c)
		}
	})
	return err
}
