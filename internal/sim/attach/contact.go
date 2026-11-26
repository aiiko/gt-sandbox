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
// "contact_surface"): una palma, un escritorio, un pie, un sofá.
//
// Dos superficies se tocan si están en la misma raíz, sus alturas en el
// mundo difieren a lo sumo en la mayor de sus Tolerance y sus zonas de apoyo
// se solapan (o sus bordes de acceso quedan al alcance). Si los campos nuevos
// quedan en cero, se aplica la regla clásica de solapamiento.
type Surface struct {
	Tolerance  units.Length `json:"tolerance"`
	MinOverlap units.Length `json:"min_overlap"`
	// Scale restringe quién puede cruzar los portales abiertos por esta
	// superficie (p. ej. "tiny"). Vacío = cualquier escala.
	Scale string `json:"scale,omitempty"`
	// StableTicks exige que el contacto se sostenga tantos ticks antes de
	// abrir el portal, evitando aperturas por un roce breve (propuesta 19.3).
	StableTicks clock.Tick `json:"stable_ticks,omitempty"`
	// RequireStill exige que el cuerpo dueño no esté caminando (ej. pie plantado).
	RequireStill bool `json:"require_still,omitempty"`
	// Support es la cara de apoyo en coordenadas locales (nil = Space.Extent).
	Support *units.Rect `json:"support,omitempty"`
	// SupportOffset desplaza en Z local la cara de apoyo respecto de la
	// superficie transitable (p. ej. -2 cm para el reverso de la mano).
	SupportOffset units.Length `json:"support_offset,omitempty"`
	// Access declara puntos locales de borde de acceso por donde entra o sale
	// un tiny (pueden quedar al costado de la huella de apoyo).
	Access []units.Vec2 `json:"access,omitempty"`
	// AccessReach es la distancia máxima entre un punto de acceso y el punto
	// transitable de la otra superficie (0 = DefaultAccessReach).
	AccessReach units.Length `json:"access_reach,omitempty"`
	// Body es la entidad cuerpo dueña de esta parte (para comprobar marcha y
	// evitar autocontactos accidentales entre partes del mismo cuerpo).
	Body ids.ID `json:"body,omitempty"`
	// SelfContact habilita contactos entre dos partes del mismo Body (p. ej.
	// mano descansando sobre la rodilla).
	SelfContact bool `json:"self_contact,omitempty"`
}

// Valores por defecto cuando Surface los deja en cero.
const (
	DefaultTolerance   = 3 * units.Centimeter
	DefaultMinOverlap  = 1 * units.Centimeter
	DefaultAccessReach = 6 * units.Centimeter
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

func (s Surface) accessReach() units.Length {
	if s.AccessReach > 0 {
		return s.AccessReach
	}
	return DefaultAccessReach
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
// portal de contacto clásico antes de recalcularlo (p. ej. si la mano se desliza).
const PortalDrift = 5 * units.Millimeter

// contacts recalcula todos los contactos en la fase Topology.
type contacts struct{}

func (contacts) Name() string { return "attach.contacts" }

type surf struct {
	id     ids.ID
	s      Surface
	root   ids.ID
	zMin   units.Length
	zMax   units.Length
	bounds units.Rect // AABB de la cara de apoyo en el mundo
	extent units.Rect
	moving bool
}

func buildSurf(r *ecs.Registry, g space.Graph, id ids.ID, s Surface) (surf, bool) {
	sp, ok := ecs.Get[space.Space](r).Get(id)
	if !ok || sp.Extent.Empty() || r.Despawning(id) {
		return surf{}, false
	}
	root, err := g.Root(id)
	if err != nil {
		return surf{}, false
	}
	rect := sp.Extent
	if s.Support != nil && !s.Support.Empty() {
		rect = *s.Support
	}
	bounds, zMin, zMax, ok := projectedBounds(g, id, rect, s.SupportOffset)
	if !ok {
		return surf{}, false
	}
	moving := false
	if s.RequireStill {
		movers := ecs.Get[nav.Mover](r)
		check := id
		if !s.Body.IsNil() && mountedOn(g, id, s.Body) {
			check = s.Body
		}
		if mv, ok := movers.Get(check); ok && mv.Status == nav.Moving {
			moving = true
		}
	}
	return surf{id: id, s: s, root: root, zMin: zMin, zMax: zMax, bounds: bounds, extent: sp.Extent, moving: moving}, true
}

func projectedBounds(g space.Graph, id ids.ID, r units.Rect, zOff units.Length) (units.Rect, units.Length, units.Length, bool) {
	corners := [4]units.Vec2{r.Min, {X: r.Max.X, Y: r.Min.Y}, r.Max, {X: r.Min.X, Y: r.Max.Y}}
	out := units.Rect{Min: units.V2(1<<62, 1<<62), Max: units.V2(-1<<62, -1<<62)}
	zMin, zMax := units.Length(1<<62), units.Length(-1<<62)
	for _, c := range corners {
		p, err := g.ToWorld(id, units.V3(c.X, c.Y, zOff))
		if err != nil {
			return units.Rect{}, 0, 0, false
		}
		out.Min.X, out.Min.Y = min(out.Min.X, p.X), min(out.Min.Y, p.Y)
		out.Max.X, out.Max.Y = max(out.Max.X, p.X), max(out.Max.Y, p.Y)
		zMin, zMax = min(zMin, p.Z), max(zMax, p.Z)
	}
	return out, zMin, zMax, true
}

func expandRect(r units.Rect, m units.Length) units.Rect {
	return units.Rect{
		Min: units.V2(r.Min.X-m, r.Min.Y-m),
		Max: units.V2(r.Max.X+m, r.Max.Y+m),
	}
}

func hasAccess(a, b surf) bool {
	return len(a.s.Access) > 0 || len(b.s.Access) > 0
}

func portalScale(a, b Surface) string {
	if a.Scale != "" {
		return a.Scale
	}
	return b.Scale
}

func surfacesTouch(a, b surf) (units.Rect, bool) {
	if a.root != b.root {
		return units.Rect{}, false
	}
	if !a.s.Body.IsNil() && a.s.Body == b.s.Body && !(a.s.SelfContact && b.s.SelfContact) {
		return units.Rect{}, false
	}
	tol := max(a.s.tolerance(), b.s.tolerance())
	if a.zMin > b.zMax+tol || b.zMin > a.zMax+tol {
		return units.Rect{}, false
	}
	if hasAccess(a, b) {
		reach := max(a.s.accessReach(), b.s.accessReach())
		over := expandRect(a.bounds, reach).Intersect(b.bounds)
		return over, !over.Empty()
	}
	over := a.bounds.Intersect(b.bounds)
	return over, !over.Empty()
}

func (contacts) Run(w *world.World) {
	r := w.ECS()
	g := space.Of(r)
	cs := ecs.Get[Contact](r)
	portals := ecs.Get[nav.Portal](r)

	var surfs []surf
	ecs.Get[Surface](r).Each(func(id ids.ID, s *Surface) {
		if sf, ok := buildSurf(r, g, id, *s); ok {
			surfs = append(surfs, sf)
		}
	})
	slices.SortFunc(surfs, func(a, b surf) int { return cmp.Compare(a.id, b.id) })

	existing := map[[2]ids.ID]ids.ID{}
	cs.Each(func(e ids.ID, c *Contact) { existing[[2]ids.ID{c.A, c.B}] = e })
	seen := map[ids.ID]bool{}

	for i := range surfs {
		for j := i + 1; j < len(surfs); j++ {
			a, b := surfs[i], surfs[j]
			over, touches := surfacesTouch(a, b)
			if !touches {
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

			c, _ := cs.Get(e)
			reqStable := max(a.s.StableTicks, b.s.StableTicks)
			stillOK := !a.moving && !b.moving
			ready := stillOK && (w.Tick()-c.Since >= reqStable)

			var ends *[2]nav.PortalEnd
			if ready {
				if hasAccess(a, b) {
					if p := portals.Ptr(e); p != nil && stillValid(r, g, a, b, p) {
						continue
					}
					ends = findAccessEnds(r, g, a, b)
				} else {
					minSide := max(a.s.minOverlap(), b.s.minOverlap())
					if over.Width() >= minSide && over.Height() >= minSide {
						if p := portals.Ptr(e); p != nil && stillValid(r, g, a, b, p) {
							continue
						}
						ends = findEnds(r, g, a, b, over)
					}
				}
			}

			scale := portalScale(a.s, b.s)
			switch p := portals.Ptr(e); {
			case ends == nil && p != nil:
				portals.Remove(e)
				w.Emit(PortalClosed{Portal: e, A: a.id, B: b.id})
			case ends != nil && p == nil:
				portals.Set(e, nav.Portal{A: ends[0], B: ends[1], Scale: scale})
				w.Emit(PortalOpened{Portal: e, A: a.id, B: b.id})
			case ends != nil:
				p.A, p.B, p.Scale = ends[0], ends[1], scale
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

// PortalCrossingValid revalida un portal de contacto dinámico en el instante
// exacto en que un mover va a cruzarlo (con las poses y movimientos del tick
// actual ya aplicados). Ver propuesta 19.3.
func PortalCrossingValid(w *world.World, _, portalID ids.ID) bool {
	r := w.ECS()
	c, ok := ecs.Get[Contact](r).Get(portalID)
	if !ok {
		return true // portal estático: ya validado por nav.routeValid
	}
	if isJustWorn(w, c.A) || isJustWorn(w, c.B) {
		return false
	}
	p := ecs.Get[nav.Portal](r).Ptr(portalID)
	if p == nil || p.Disabled {
		return false
	}
	surfs := ecs.Get[Surface](r)
	saData, okA := surfs.Get(c.A)
	sbData, okB := surfs.Get(c.B)
	if !okA || !okB {
		return false
	}
	g := space.Of(r)
	a, okA := buildSurf(r, g, c.A, saData)
	b, okB := buildSurf(r, g, c.B, sbData)
	if !okA || !okB {
		return false
	}
	if _, touches := surfacesTouch(a, b); !touches {
		return false
	}
	return stillValid(r, g, a, b, p)
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

// stillValid verifica que los extremos sigan dentro de sus extensiones,
// transitables en sus grillas, estables y dentro de la tolerancia de deriva.
func stillValid(r *ecs.Registry, g space.Graph, a, b surf, p *nav.Portal) bool {
	if a.moving || b.moving {
		return false
	}
	grids := ecs.Get[nav.Grid](r)
	if !a.extent.Contains(p.A.Pos) || !b.extent.Contains(p.B.Pos) {
		return false
	}
	if ga := grids.Ptr(a.id); ga != nil && !ga.Walkable(p.A.Pos) {
		return false
	}
	if gb := grids.Ptr(b.id); gb != nil && !gb.Walkable(p.B.Pos) {
		return false
	}
	wa, err1 := g.ToWorld(p.A.Space, units.V3(p.A.Pos.X, p.A.Pos.Y, 0))
	wb, err2 := g.ToWorld(p.B.Space, units.V3(p.B.Pos.X, p.B.Pos.Y, 0))
	if err1 != nil || err2 != nil {
		return false
	}
	if hasAccess(a, b) {
		reach := max(a.s.accessReach(), b.s.accessReach())
		return wa.DistSq(wb) <= int64(reach)*int64(reach)
	}
	return wa.XY().DistSq(wb.XY()) <= int64(PortalDrift)*int64(PortalDrift)
}

// findAccessEnds busca un par de extremos transitables usando los bordes de
// acceso declarados en a o b (en orden determinista).
func findAccessEnds(r *ecs.Registry, g space.Graph, a, b surf) *[2]nav.PortalEnd {
	grids := ecs.Get[nav.Grid](r)
	reach := max(a.s.accessReach(), b.s.accessReach())
	reachSq := int64(reach) * int64(reach)

	walkable := func(s surf, pt units.Vec2) bool {
		if !s.extent.Contains(pt) {
			return false
		}
		if gr := grids.Ptr(s.id); gr != nil && !gr.Walkable(pt) {
			return false
		}
		return true
	}

	match := func(src, dst surf) (units.Vec2, units.Vec2, bool) {
		for _, pt := range src.s.Access {
			if !walkable(src, pt) {
				continue
			}
			wp, err := g.ToWorld(src.id, units.V3(pt.X, pt.Y, 0))
			if err != nil {
				continue
			}
			ld, err := g.FromWorld(dst.id, wp)
			if err != nil {
				continue
			}
			// Probar primero el punto proyectado y luego anillos deterministas hasta reach.
			step := units.Centimeter
			if gr := grids.Ptr(dst.id); gr != nil && gr.Cell > 0 && gr.Cell < step {
				step = gr.Cell
			}
			base := clampToRect(ld.XY(), dst.extent)
			var best units.Vec2
			var bestD int64
			found := false
			checkCand := func(cand units.Vec2) {
				if !walkable(dst, cand) {
					return
				}
				wd, err := g.ToWorld(dst.id, units.V3(cand.X, cand.Y, 0))
				if err != nil {
					return
				}
				d := wp.DistSq(wd)
				if d <= reachSq && (!found || d < bestD) {
					best, bestD, found = cand, d, true
				}
			}
			checkCand(base)
			if !found {
				steps := int(reach/step) + 1
				for ring := 1; ring <= steps && !found; ring++ {
					off := units.Length(ring) * step
					for dy := -off; dy <= off; dy += step {
						for dx := -off; dx <= off; dx += step {
							if dx.Abs() != off && dy.Abs() != off {
								continue
							}
							checkCand(units.V2(base.X+dx, base.Y+dy))
						}
					}
				}
			}
			if found {
				return pt, best, true
			}
		}
		return units.Vec2{}, units.Vec2{}, false
	}

	if len(a.s.Access) > 0 {
		if pa, pb, ok := match(a, b); ok {
			return &[2]nav.PortalEnd{{Space: a.id, Pos: pa}, {Space: b.id, Pos: pb}}
		}
	}
	if len(b.s.Access) > 0 {
		if pb, pa, ok := match(b, a); ok {
			return &[2]nav.PortalEnd{{Space: a.id, Pos: pa}, {Space: b.id, Pos: pb}}
		}
	}
	return nil
}

func clampToRect(p units.Vec2, r units.Rect) units.Vec2 {
	x := min(max(p.X, r.Min.X), r.Max.X-1)
	y := min(max(p.Y, r.Min.Y), r.Max.Y-1)
	return units.V2(x, y)
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
		l, err := g.FromWorld(s.id, units.V3(p.X, p.Y, s.zMin))
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
