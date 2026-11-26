package crush

import (
	"cmp"
	"fmt"
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// resolve (fase Consequences) detecta los apoyos nuevos y aplasta lo que
// quedó debajo:
//
//   - un Space aplastador se apoya sobre una superficie cuando su cara de
//     abajo queda a ras de ella (entre PressBelow por debajo y PressAbove
//     por encima) y sus extensiones se solapan en el plano. Solo aplasta el
//     apoyo nuevo: un tiny que camina por debajo de una mano ya apoyada no
//     muere;
//   - un objeto aplastador aplasta al aterrizar (entity_fell del tick
//     anterior), en el Space donde quedó.
type resolve struct{}

func (resolve) Name() string { return "crush.resolve" }

// impact es un apoyo nuevo de By sobre Surface. Inside dice si un punto del
// mundo cae en la huella; gap es la altura de la cara de abajo sobre la
// superficie.
type impact struct {
	by      ids.ID
	c       Crusher
	surface ids.ID
	gap     units.Length
	inside  func(units.Vec3) bool
}

func (resolve) Run(w *world.World) {
	r := w.ECS()
	g := space.Of(r)
	crushers := ecs.Get[Crusher](r)
	var impacts []impact

	crushers.Each(func(e ids.ID, c *Crusher) {
		if !g.IsSpace(e) || r.Despawning(e) {
			c.Pressing = nil
			return
		}
		now := pressing(r, g, e, *c)
		for _, p := range now {
			if !slices.Contains(c.Pressing, p.surface) {
				impacts = append(impacts, p)
			}
		}
		c.Pressing = c.Pressing[:0]
		for _, p := range now {
			c.Pressing = append(c.Pressing, p.surface)
		}
		if len(c.Pressing) == 0 {
			c.Pressing = nil
		}
	})

	locs, bodies := comp.Of(r).Location, comp.Of(r).Body
	for _, ev := range w.Events() {
		f, ok := ev.Event.(space.EntityFell)
		if !ok || g.IsSpace(f.Entity) || !r.Alive(f.Entity) || r.Despawning(f.Entity) {
			continue
		}
		c, ok := crushers.Get(f.Entity)
		l, ok2 := locs.Get(f.Entity)
		if !ok || !ok2 {
			continue
		}
		center, _, err := g.WorldPose(f.Entity)
		if err != nil {
			continue
		}
		radius := c.Radius
		if b, ok := bodies.Get(f.Entity); ok && radius == 0 {
			radius = b.Radius
		}
		rr := int64(radius) * int64(radius)
		impacts = append(impacts, impact{by: f.Entity, c: c, surface: l.Space, inside: func(p units.Vec3) bool {
			return p.XY().DistSq(center.XY()) <= rr
		}})
	}

	for _, gp := range attach.ConsumeGarmentPressures(w) {
		garID := gp.Garment
		zone := gp.Pressure.Zone
		impacts = append(impacts, impact{
			by:      garID,
			c:       Crusher{Source: gp.Pressure.EffectiveSource(), Force: gp.Pressure.Force},
			surface: garID,
			gap:     0,
			inside: func(p units.Vec3) bool {
				l, err := g.FromWorld(garID, p)
				return err == nil && zone.Contains(l.XY())
			},
		})
	}

	for _, im := range impacts {
		w.Emit(Pressed{By: im.by, Source: im.c.Source, Surface: im.surface})
		for _, v := range victims(w, g, im) {
			kill(w, g, v, im)
		}
		nearMisses(w, g, im)
	}
}

// pressing devuelve los apoyos actuales del Space aplastador e, ordenados
// por superficie.
func pressing(r *ecs.Registry, g space.Graph, e ids.ID, c Crusher) []impact {
	return over(r, g, e, c, -PressBelow, PressAbove)
}

// over devuelve las superficies sobre las que está el Space aplastador e,
// con su cara de abajo entre lo y hi por encima de ellas, ordenadas.
func over(r *ecs.Registry, g space.Graph, e ids.ID, c Crusher, lo, hi units.Length) []impact {
	spaces := ecs.Get[space.Space](r)
	sp, ok := spaces.Get(e)
	if !ok || sp.Extent.Empty() {
		return nil
	}
	root, err := g.Root(e)
	if err != nil {
		return nil
	}
	faceRect := sp.Extent
	faceZOff := -c.Thickness
	if sf, ok := ecs.Get[attach.Surface](r).Get(e); ok {
		if sf.Support != nil && !sf.Support.Empty() {
			faceRect = *sf.Support
		}
		if sf.SupportOffset != 0 {
			faceZOff = sf.SupportOffset
		}
	}
	bounds, bottom, ok := faceWorldBounds(g, e, faceRect, faceZOff)
	if !ok {
		return nil
	}
	grid := ecs.Get[nav.Grid](r).Ptr(e)
	inside := func(p units.Vec3) bool {
		l, err := g.FromWorld(e, p)
		return err == nil && faceRect.Contains(l.XY()) && (grid == nil || grid.Walkable(l.XY()))
	}
	var out []impact
	spaces.Each(func(s ids.ID, ssp *space.Space) {
		if ssp.Extent.Empty() || r.Despawning(s) || g.CanPlace(e, s) != nil {
			return // s es e o está montado sobre e
		}
		if rs, err := g.Root(s); err != nil || rs != root {
			return
		}
		sBounds, sTop, ok := faceWorldBounds(g, s, ssp.Extent, 0)
		if !ok {
			return
		}
		gap := bottom - sTop
		if gap < lo || gap > hi || !bounds.Overlaps(sBounds) {
			return
		}
		out = append(out, impact{by: e, c: c, surface: s, gap: max(gap, 0), inside: inside})
	})
	slices.SortFunc(out, func(a, b impact) int { return cmp.Compare(a.surface, b.surface) })
	return out
}

func faceWorldBounds(g space.Graph, id ids.ID, rc units.Rect, zOff units.Length) (units.Rect, units.Length, bool) {
	corners := [4]units.Vec2{rc.Min, {X: rc.Max.X, Y: rc.Min.Y}, rc.Max, {X: rc.Min.X, Y: rc.Max.Y}}
	out := units.Rect{Min: units.V2(1<<62, 1<<62), Max: units.V2(-1<<62, -1<<62)}
	zMin := units.Length(1 << 62)
	for _, c := range corners {
		p, err := g.ToWorld(id, units.V3(c.X, c.Y, zOff))
		if err != nil {
			return units.Rect{}, 0, false
		}
		out.Min.X, out.Min.Y = min(out.Min.X, p.X), min(out.Min.Y, p.Y)
		out.Max.X, out.Max.Y = max(out.Max.X, p.X), max(out.Max.Y, p.Y)
		zMin = min(zMin, p.Z)
	}
	return out, zMin, true
}

// victims devuelve, en orden de store, las entidades frágiles apoyadas en la
// superficie del impacto que quedan dentro de la huella y son más altas que
// el hueco bajo la cara de abajo. Las que salva un refugio emiten Sheltered.
func victims(w *world.World, g space.Graph, im impact) []ids.ID {
	r := w.ECS()
	locs, bodies := comp.Of(r).Location, comp.Of(r).Body
	shelters := ecs.Get[Shelter](r)
	var out []ids.ID
	ecs.Get[Fragile](r).Each(func(e ids.ID, _ *Fragile) {
		l, ok := locs.Get(e)
		if !ok || l.Space != im.surface || e == im.by || g.IsSpace(e) || r.Despawning(e) {
			return
		}
		if b, ok := bodies.Get(e); ok && b.Height <= im.gap {
			return
		}
		p, _, err := g.WorldPose(e)
		if err != nil || !im.inside(p) {
			return
		}
		var by ids.ID
		shelters.Each(func(s ids.ID, sh *Shelter) {
			sl, ok := locs.Get(s)
			if !by.IsNil() || !ok || sl.Space != l.Space || r.Despawning(s) || sh.Strength < im.c.Force {
				return
			}
			if l.Pos.XY().DistSq(sl.Pos.XY()) <= int64(sh.Radius)*int64(sh.Radius) {
				by = s
			}
		})
		if !by.IsNil() {
			w.Emit(Sheltered{Entity: e, By: im.by, Shelter: by})
			return
		}
		out = append(out, e)
	})
	return out
}

// kill destruye a la víctima, crea sus restos y emite Crushed.
func kill(w *world.World, g space.Graph, v ids.ID, im impact) {
	r := w.ECS()
	cs := comp.Of(r)
	l, _ := cs.Location.Get(v)
	name, _ := cs.Name.Get(v)
	if name == "" {
		name = comp.Name(fmt.Sprint(v))
	}
	radius := units.Millimeter
	if b, ok := cs.Body.Get(v); ok {
		radius = b.Radius
	}
	rem := r.Spawn()
	cs.Location.Set(rem, comp.Location{Space: l.Space, Pos: units.V3(l.Pos.X, l.Pos.Y, 0), Facing: l.Facing})
	cs.Body.Set(rem, comp.Body{Radius: radius, Height: units.Millimeter})
	cs.Name.Set(rem, name)
	ecs.Get[Remains](r).Set(rem, Remains{Name: name, Of: v, By: im.by, Source: im.c.Source, Tick: w.Tick()})
	r.Despawn(v)
	w.Emit(Crushed{Entity: v, By: im.by, Source: im.c.Source, Space: l.Space, Pos: l.Pos.XY(), Remains: rem})
}
