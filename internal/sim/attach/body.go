package attach

import (
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

const (
	// DefaultGripRadius es el alcance por defecto de un punto de agarre.
	DefaultGripRadius = 4 * units.Centimeter
	// GripHoldTicks es cuánto tiempo se sujeta un tiny ante un cambio de pose.
	GripHoldTicks clock.Tick = 4
)

// BodyState guarda la pose activa de un cuerpo (store "body_state").
type BodyState struct {
	Pose           string     `json:"pose"`
	PoseTick       clock.Tick `json:"pose_tick,omitempty"`
	Tickle         int32      `json:"tickle,omitempty"`
	TickleZone     ids.ID     `json:"tickle_zone,omitempty"`
	TickleCooldown clock.Tick `json:"tickle_cooldown,omitempty"`
	TickleUntil    clock.Tick `json:"tickle_until,omitempty"`
}

// Joint declara un portal interno entre dos zonas del cuerpo cuyo estado
// abierto/cerrado depende de la pose activa del cuerpo (sin recrear el portal).
type Joint struct {
	Portal      ids.ID   `json:"portal"`
	ClosedPoses []string `json:"closed_poses,omitempty"`
	OpenPoses   []string `json:"open_poses,omitempty"`
}

// ClosedIn indica si la articulación está cerrada en pose.
func (j Joint) ClosedIn(pose string) bool {
	if len(j.OpenPoses) > 0 {
		return !slices.Contains(j.OpenPoses, pose)
	}
	return slices.Contains(j.ClosedPoses, pose)
}

// BodyJoints es el componente "body_joints" de un cuerpo: sus articulaciones
// internas dependientes de la pose.
type BodyJoints []Joint

// GripPoint es un lugar marcado donde un tiny puede sujetarse o esperar
// (puño de manga, costura, borde de bolsillo, mechón).
type GripPoint struct {
	Pos    units.Vec2   `json:"pos"`
	Radius units.Length `json:"radius,omitempty"`
	Label  string       `json:"label,omitempty"`
}

func (g GripPoint) radius() units.Length {
	if g.Radius > 0 {
		return g.Radius
	}
	return DefaultGripRadius
}

// BodyZone marca una zona corporal explorable (store "body_zone").
type BodyZone struct {
	Body        ids.ID      `json:"body,omitempty"`
	Region      string      `json:"region,omitempty"`
	Sensitivity int32       `json:"sensitivity,omitempty"`
	RequireGrip bool        `json:"require_grip,omitempty"`
	Grips       []GripPoint `json:"grips,omitempty"`
	// Occluded indica que la ropa o el pelo tapan al tiny y la vista micro
	// debe recortar la capa frontal a su alrededor (propuesta 19.5).
	Occluded bool `json:"occluded,omitempty"`
}

// NearestGrip devuelve el punto de agarre más cercano a pos dentro de su radio.
func (z BodyZone) NearestGrip(pos units.Vec2) (GripPoint, bool) {
	var best GripPoint
	var bestD int64
	found := false
	for _, g := range z.Grips {
		r := g.radius()
		d := pos.DistSq(g.Pos)
		if d <= int64(r)*int64(r) && (!found || d < bestD) {
			best, bestD, found = g, d, true
		}
	}
	return best, found
}

// Gripping es el componente "gripping" de un tiny sujetado a un agarre tras un
// cambio de pose o sacudida.
type Gripping struct {
	Space ids.ID     `json:"space"`
	Pos   units.Vec2 `json:"pos"`
	Label string     `json:"label,omitempty"`
	Until clock.Tick `json:"until,omitempty"`
}

// EntityGripped se emite cuando un tiny se sujeta a un punto de agarre ante un
// cambio de pose.
type EntityGripped struct {
	Entity ids.ID `json:"entity"`
	Space  ids.ID `json:"space"`
	Label  string `json:"label,omitempty"`
}

func (EntityGripped) Kind() string { return "entity_gripped" }

// EntityLostGrip se emite cuando un tiny pierde apoyo en una zona inclinada sin
// punto de agarre al alcance y comienza a caer.
type EntityLostGrip struct {
	Entity ids.ID `json:"entity"`
	Space  ids.ID `json:"space"`
}

func (EntityLostGrip) Kind() string { return "entity_lost_grip" }

// applyJoints actualiza el estado abierto/cerrado de los portales internos del
// cuerpo para la pose dada sin recrearlos.
func applyJoints(w *world.World, body ids.ID, pose string) {
	r := w.ECS()
	joints, ok := ecs.Get[BodyJoints](r).Get(body)
	if !ok {
		return
	}
	portals := ecs.Get[nav.Portal](r)
	for _, j := range joints {
		p := portals.Ptr(j.Portal)
		if p == nil || r.Despawning(j.Portal) {
			continue
		}
		closed := j.ClosedIn(pose)
		if p.Disabled != closed {
			p.Disabled = closed
			p.Version++
		}
	}
}

// checkGripOnPose evalúa agarre y pérdida de apoyo de los tinies ubicados en
// las partes afectadas por un cambio de pose.
func checkGripOnPose(w *world.World, affected []ids.ID) {
	if len(affected) == 0 {
		return
	}
	r := w.ECS()
	g := space.Of(r)
	spaces := ecs.Get[space.Space](r)
	zones := ecs.Get[BodyZone](r)
	grips := ecs.Get[Gripping](r)
	movers := ecs.Get[nav.Mover](r)
	locs := comp.Of(r).Location
	bodies := comp.Of(r).Body

	inAffected := map[ids.ID]bool{}
	for _, id := range affected {
		inAffected[id] = true
	}

	var toDrop []struct {
		e  ids.ID
		sp ids.ID
	}
	bodies.Each(func(e ids.ID, _ *comp.Body) {
		if spaces.Has(e) || r.Despawning(e) {
			return
		}
		l, ok := locs.Get(e)
		if !ok || !inAffected[l.Space] {
			return
		}
		sp, ok := spaces.Get(l.Space)
		if !ok {
			return
		}
		zone, hasZone := zones.Get(l.Space)
		if !hasZone && !sp.IsClimb() {
			return
		}
		if !zone.RequireGrip && !sp.IsClimb() {
			return
		}
		if gp, ok := zone.NearestGrip(l.Pos.XY()); ok {
			grips.Set(e, Gripping{Space: l.Space, Pos: l.Pos.XY(), Label: gp.Label, Until: w.Tick() + GripHoldTicks})
			if mv := movers.Ptr(e); mv != nil && mv.WaitUntil < w.Tick()+GripHoldTicks {
				mv.WaitUntil = w.Tick() + GripHoldTicks
			}
			w.Emit(EntityGripped{Entity: e, Space: l.Space, Label: gp.Label})
			return
		}
		if sp.IsClimb() {
			// Un tramo de trepada vertical (costura/borde) sostiene a quien ya va agarrado.
			label := zone.Region
			grips.Set(e, Gripping{Space: l.Space, Pos: l.Pos.XY(), Label: label, Until: w.Tick() + GripHoldTicks})
			if mv := movers.Ptr(e); mv != nil && mv.WaitUntil < w.Tick()+GripHoldTicks {
				mv.WaitUntil = w.Tick() + GripHoldTicks
			}
			w.Emit(EntityGripped{Entity: e, Space: l.Space, Label: label})
			return
		}
		// Zona inclinada que exige agarre y no hay punto al alcance: pierde apoyo.
		toDrop = append(toDrop, struct {
			e  ids.ID
			sp ids.ID
		}{e, l.Space})
	})

	for _, d := range toDrop {
		nav.Halt(w, d.e)
		grips.Remove(d.e)
		w.Emit(EntityLostGrip{Entity: d.e, Space: d.sp})
		if !space.LandExcluding(w, d.e, d.sp) {
			if root, err := g.Root(d.e); err == nil {
				_, _ = g.Reparent(d.e, root)
			}
		}
	}
}

// gripUpkeep limpia estados Gripping vencidos o cuando la entidad cambió de Space.
type gripUpkeep struct{}

func (gripUpkeep) Name() string { return "attach.grips" }

func (gripUpkeep) Run(w *world.World) {
	r := w.ECS()
	grips := ecs.Get[Gripping](r)
	locs := comp.Of(r).Location
	for _, e := range grips.IDs() {
		gr, _ := grips.Get(e)
		l, ok := locs.Get(e)
		if !ok || l.Space != gr.Space || (gr.Until > 0 && w.Tick() >= gr.Until) {
			grips.Remove(e)
		}
	}
}
