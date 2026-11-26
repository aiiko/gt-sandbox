package ai

import (
	"math"
	"slices"

	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Danger es el componente "danger" de una zona que los agentes de Scale
// evitan hasta Until: ahí murió alguien. La crea quien sabe de muertes (la
// colonia) y la borra al vencer.
type Danger struct {
	Scale  string       `json:"scale"`
	Space  ids.ID       `json:"space"`
	Center units.Vec2   `json:"center"`
	Radius units.Length `json:"radius"`
	Until  clock.Tick   `json:"until"`
}

// InDanger indica si el punto p del Space sp está en una zona de peligro
// vigente para agentes de la escala scale.
func InDanger(w *world.World, scale string, sp ids.ID, p units.Vec2) bool {
	found := false
	ecs.Get[Danger](w.ECS()).Each(func(e ids.ID, d *Danger) {
		found = found || d.Scale == scale && d.Space == sp && d.Until > w.Tick() && !w.ECS().Despawning(e) &&
			p.DistSq(d.Center) <= int64(d.Radius)*int64(d.Radius)
	})
	return found
}

// endangered indica si el objetivo de una opción (o el de su trabajo) está
// en una zona de peligro para el actor.
func endangered(c *Ctx, target ids.ID) bool {
	r := c.W.ECS()
	if ecs.Get[Danger](r).Len() == 0 {
		return false
	}
	if j, ok := ecs.Get[jobs.Job](r).Get(target); ok {
		target = j.Target
	}
	l, ok := comp.Of(r).Location.Get(target)
	return ok && InDanger(c.W, c.Agent.Scale, l.Space, l.Pos.XY())
}

// Cuidado de la giantess al apoyar la mano o al sentarse.
const (
	// HoverExtra es cuánto más sostiene el aplastador en el aire, por unidad
	// de Care, antes de apoyarlo.
	HoverExtra clock.Tick = 40
	// CareWaitMax es lo máximo que espera a que se aparten los tinies que vio
	// debajo; después apoya igual.
	CareWaitMax clock.Tick = 100
	// SeatCareScale reduce la atención al sentarse (fuera de la línea de vista
	// directa): hace falta un Care más alto para notar tinies en el asiento.
	SeatCareScale = 0.45
)

// Hesitated se emite cuando la giantess ve tinies debajo de su mano o en el
// asiento y espera a que se aparten.
type Hesitated struct {
	Entity  ids.ID `json:"entity"`
	Surface ids.ID `json:"surface"`
	Source  string `json:"source,omitempty"`
	Seen    int    `json:"seen"`
}

func (Hesitated) Kind() string { return "hesitated" }

// careful decide si un paso pose de un agente con Care debe esperar. Aplica
// cuando su mano o su asiento (SourceSeat montado sobre el actor) están
// suspendidos sobre una superficie (crush.Threats): ese pose es el que apoya.
// d.Progress cuenta los ticks de espera y d.Wary indica que vio tinies debajo.
func careful(c *Ctx, d *Doing) bool {
	a := c.Agent
	if a.Care <= 0 {
		return false
	}
	g := space.Of(c.W.ECS())
	var hover *crush.Threat
	for _, t := range crush.Threats(c.W) {
		if !a.Hand.IsNil() && t.By == a.Hand {
			tt := t
			hover = &tt
			break
		}
		if hover == nil && t.Source == crush.SourceSeat && g.Parent(t.By) == c.Actor {
			tt := t
			hover = &tt
		}
	}
	if hover == nil {
		return false
	}
	care := min(a.Care, 1)
	if hover.Source == crush.SourceSeat {
		care = care * care * SeatCareScale
	}
	d.Progress++
	seen := visibleUnder(c, *hover)
	if d.Progress == 1 && len(seen) > 0 {
		// tirada de atención: cada tiny visible es una chance de notarlos
		p := 1 - math.Pow(1-care, float64(len(seen)))
		if c.W.RNG("ai.attention").Float64() < p {
			d.Wary = true
			c.W.Emit(Hesitated{Entity: c.Actor, Surface: hover.Surface, Source: hover.Source, Seen: len(seen)})
		}
	}
	if d.Progress <= clock.Tick(care*float64(HoverExtra)) {
		return true
	}
	return d.Wary && d.Progress <= CareWaitMax && len(seen) > 0
}

// visibleUnder son los frágiles bajo la huella (con margen de susto) que el
// agente alcanza a ver (a menos de Watch en el mundo).
func visibleUnder(c *Ctx, t crush.Threat) []ids.ID {
	if c.Agent.Watch <= 0 {
		return nil
	}
	r := c.W.ECS()
	g := space.Of(r)
	me, _, err := g.WorldPose(c.Actor)
	if err != nil {
		return nil
	}
	r2 := int64(c.Agent.Watch) * int64(c.Agent.Watch)
	var out []ids.ID
	for _, e := range t.Under(c.W, crush.NearMissRange) {
		if p, _, err := g.WorldPose(e); err == nil && p.XY().DistSq(me.XY()) <= r2 {
			out = append(out, e)
		}
	}
	return out
}

const garmentReach = 60 * units.Centimeter

func findGarment(w *world.World, actor ids.ID, slot string, wantState attach.GarmentState, anyState bool) (ids.ID, attach.Garment, bool) {
	r := w.ECS()
	store := ecs.Get[attach.Garment](r)
	all := slices.Clone(store.IDs())
	slices.Sort(all)
	for _, id := range all {
		if r.Despawning(id) {
			continue
		}
		g, _ := store.Get(id)
		if g.Owner == actor && g.Slot == slot && (anyState || g.State == wantState) {
			return id, g, true
		}
	}
	return ids.Nil, attach.Garment{}, false
}

func hasDonningGarment(w *world.World, actor ids.ID) bool {
	r := w.ECS()
	store := ecs.Get[attach.Garment](r)
	for _, id := range store.IDs() {
		if r.Despawning(id) {
			continue
		}
		if g, ok := store.Get(id); ok && g.Owner == actor && g.State == attach.GarmentStateDonning {
			return true
		}
	}
	return false
}

func cancelDonningGarments(w *world.World, actor ids.ID) {
	r := w.ECS()
	store := ecs.Get[attach.Garment](r)
	all := slices.Clone(store.IDs())
	slices.Sort(all)
	for _, id := range all {
		if r.Despawning(id) {
			continue
		}
		if g, ok := store.Get(id); ok && g.Owner == actor && g.State == attach.GarmentStateDonning {
			_ = (attach.CancelDon{Actor: actor, Garment: id}).Apply(w)
		}
	}
}

func visibleInsideGarment(c *Ctx, garID ids.ID) []ids.ID {
	if c.Agent.Watch <= 0 {
		return nil
	}
	r := c.W.ECS()
	g := space.Of(r)
	me, _, err := g.WorldPose(c.Actor)
	if err != nil {
		return nil
	}
	r2 := int64(c.Agent.Watch) * int64(c.Agent.Watch)
	locs := comp.Of(r).Location
	frag := ecs.Get[crush.Fragile](r)
	all := slices.Clone(frag.IDs())
	slices.Sort(all)
	var out []ids.ID
	for _, e := range all {
		if r.Despawning(e) || g.IsSpace(e) {
			continue
		}
		l, ok := locs.Get(e)
		if !ok || l.Space != garID {
			continue
		}
		if p, _, err := g.WorldPose(e); err == nil && p.XY().DistSq(me.XY()) <= r2 {
			out = append(out, e)
		}
	}
	return out
}

func staticRoomOf(r *ecs.Registry, g space.Graph, id ids.ID) ids.ID {
	spaces := ecs.Get[space.Space](r)
	cur := id
	for range space.MaxDepth {
		parent := g.Parent(cur)
		if parent.IsNil() {
			break
		}
		if g.Parent(parent).IsNil() {
			if sp, ok := spaces.Get(cur); ok && sp.Kind == space.Static {
				return cur
			}
		}
		cur = parent
	}
	return ids.Nil
}

func stepDon(c *Ctx, d *Doing, s Step) stepResult {
	w := c.W
	r := w.ECS()
	garID, gar, ok := findGarment(w, c.Actor, s.Slot, "", true)
	if !ok || gar.State == attach.GarmentStateWorn {
		return stepNext
	}
	if gar.State == attach.GarmentStateDonning || gar.State == attach.GarmentStateDoffing {
		return stepWait
	}
	g := space.Of(r)
	if actorRoom, garRoom := staticRoomOf(r, g, c.Actor), staticRoomOf(r, g, garID); !actorRoom.IsNil() && !garRoom.IsNil() && actorRoom != garRoom {
		return stepNext
	}
	actorPos, _, err1 := g.WorldPose(c.Actor)
	garPos, _, err2 := g.WorldPose(garID)
	if err1 != nil || err2 != nil {
		return stepFail
	}
	if actorPos.XY().DistSq(garPos.XY()) > int64(garmentReach)*int64(garmentReach) {
		sp, goal, ok := garmentApproachTarget(w, c.Actor, garPos)
		if !ok {
			return stepFail
		}
		switch nav.Approach(w, c.Actor, sp, goal, &d.Moving) {
		case nav.Failed:
			return stepFail
		case nav.Underway:
			return stepWait
		}
	}
	if d.Moving {
		nav.Halt(w, c.Actor)
		d.Moving = false
	}
	if c.Agent.Care > 0 {
		care := min(c.Agent.Care, 1)
		seen := visibleInsideGarment(c, garID)
		d.Progress++
		if d.Progress == 1 && len(seen) > 0 {
			p := 1 - math.Pow(1-care, float64(len(seen)))
			if w.RNG("ai.attention").Float64() < p {
				d.Wary = true
				w.Emit(Hesitated{Entity: c.Actor, Surface: garID, Source: attach.DefaultGarmentSource, Seen: len(seen)})
			}
		}
		if d.Wary && d.Progress <= CareWaitMax && len(seen) > 0 {
			return stepWait
		}
	}
	baseWarn := gar.DonWarnTicks
	if baseWarn == 0 {
		baseWarn = attach.DefaultDonWarnTicks
	}
	care := min(max(c.Agent.Care, 0), 1)
	warn := baseWarn + clock.Tick(care*float64(HoverExtra))
	if err := (attach.DonGarment{Actor: c.Actor, Garment: garID, WarnTicks: int64(warn)}).Apply(w); err != nil {
		return stepFail
	}
	return stepWait
}

func garmentApproachTarget(w *world.World, actor ids.ID, garWorld units.Vec3) (ids.ID, units.Vec2, bool) {
	r := w.ECS()
	g := space.Of(r)
	spaces := ecs.Get[space.Space](r)
	grids := ecs.Get[nav.Grid](r)
	actorLoc, ok := comp.Of(r).Location.Get(actor)
	if !ok {
		return ids.Nil, units.Vec2{}, false
	}
	root, err := g.Root(actor)
	if err != nil {
		return ids.Nil, units.Vec2{}, false
	}
	// Buscar el Space Static (habitación) que contiene garWorld o el Space del actor.
	target := actorLoc.Space
	allSpaces := slices.Clone(spaces.IDs())
	slices.Sort(allSpaces)
	for _, sid := range allSpaces {
		if r.Despawning(sid) {
			continue
		}
		sp, _ := spaces.Get(sid)
		if sp.Kind != space.Static || sp.Extent.Empty() {
			continue
		}
		if sr, err := g.Root(sid); err != nil || sr != root {
			continue
		}
		if loc, err := g.FromWorld(sid, garWorld); err == nil && sp.Extent.Contains(loc.XY()) {
			target = sid
			break
		}
	}
	sp, ok := spaces.Get(target)
	if !ok {
		return ids.Nil, units.Vec2{}, false
	}
	loc, err := g.FromWorld(target, garWorld)
	if err != nil {
		return ids.Nil, units.Vec2{}, false
	}
	pt := nearestWalkableInSpace(sp.Extent, grids.Ptr(target), loc.XY())
	return target, pt, true
}

func stepDoff(c *Ctx, s Step) stepResult {
	w := c.W
	r := w.ECS()
	garID, _, ok := findGarment(w, c.Actor, s.Slot, attach.GarmentStateWorn, false)
	if !ok {
		return stepNext
	}
	actorLoc, ok := comp.Of(r).Location.Get(c.Actor)
	if !ok {
		return stepFail
	}
	targetSpace := actorLoc.Space
	rawPt := actorLoc.Pos.XY().Add(fixedmath.Rotate(s.Offset, actorLoc.Facing))
	if s.Place != "" {
		if placeSp, ok := nearestPlaceSpace(w, c.Actor, s.Place); ok {
			targetSpace = placeSp
			sp, _ := ecs.Get[space.Space](r).Get(placeSp)
			center := units.V2((sp.Extent.Min.X+sp.Extent.Max.X)/2, (sp.Extent.Min.Y+sp.Extent.Max.Y)/2)
			rawPt = center.Add(s.Offset)
		}
	}
	sp, ok := ecs.Get[space.Space](r).Get(targetSpace)
	if !ok {
		return stepFail
	}
	pt := nearestWalkableInSpace(sp.Extent, ecs.Get[nav.Grid](r).Ptr(targetSpace), rawPt)
	if err := (attach.DoffGarment{
		Actor:       c.Actor,
		Garment:     garID,
		TargetSpace: targetSpace,
		Pos:         units.V3(pt.X, pt.Y, 0),
		Facing:      actorLoc.Facing,
	}).Apply(w); err != nil {
		return stepFail
	}
	return stepNext
}

func nearestPlaceSpace(w *world.World, actor ids.ID, tag string) (ids.ID, bool) {
	r := w.ECS()
	g := space.Of(r)
	actorPos, _, err := g.WorldPose(actor)
	if err != nil {
		return ids.Nil, false
	}
	spaces := ecs.Get[space.Space](r)
	places := ecs.Get[Place](r)
	names := comp.Of(r).Name
	const maxDist = 150 * units.Centimeter
	maxD2 := int64(maxDist) * int64(maxDist)
	var best ids.ID
	var bestD int64
	found := false
	all := slices.Clone(spaces.IDs())
	slices.Sort(all)
	for _, sid := range all {
		if sid == actor || r.Despawning(sid) {
			continue
		}
		sp, _ := spaces.Get(sid)
		if sp.Extent.Empty() {
			continue
		}
		matched := false
		if p, ok := places.Get(sid); ok && hasTag(p.Tags, tag) {
			matched = true
		} else if n, ok := names.Get(sid); ok && string(n) == tag {
			matched = true
		}
		if !matched {
			continue
		}
		pos, _, err := g.WorldPose(sid)
		if err != nil {
			continue
		}
		d := actorPos.XY().DistSq(pos.XY())
		if d <= maxD2 && (!found || d < bestD) {
			best, bestD, found = sid, d, true
		}
	}
	return best, found
}

func nearestWalkableInSpace(ext units.Rect, grid *nav.Grid, pt units.Vec2) units.Vec2 {
	if !ext.Empty() {
		pt = units.V2(
			min(max(pt.X, ext.Min.X), ext.Max.X-1),
			min(max(pt.Y, ext.Min.Y), ext.Max.Y-1),
		)
	}
	if grid == nil || grid.Walkable(pt) {
		return pt
	}
	var best units.Vec2
	var bestD int64
	found := false
	for y := range grid.H {
		for x := range grid.W {
			if grid.At(x, y) == nav.Blocked {
				continue
			}
			c := grid.Center(x, y)
			if !ext.Empty() && !ext.Contains(c) {
				continue
			}
			if d := pt.DistSq(c); !found || d < bestD {
				best, bestD, found = c, d, true
			}
		}
	}
	if found {
		return best
	}
	return pt
}
