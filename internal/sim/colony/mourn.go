package colony

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/world"
)

// Valores por defecto de la zona de peligro que deja una muerte.
const (
	DefaultDangerRadius            = 10 * units.Centimeter
	DefaultDangerTicks  clock.Tick = 2 * 60 * clock.TPS
)

// MemberLost se emite cuando un miembro de la colonia muere aplastado.
type MemberLost struct {
	Colony  ids.ID `json:"colony"`
	Entity  ids.ID `json:"entity"`
	Cause   string `json:"cause"`
	Remains ids.ID `json:"remains"`
	Danger  ids.ID `json:"danger"`
}

func (MemberLost) Kind() string { return "member_lost" }

// DangerExpired se emite cuando vence una zona de peligro.
type DangerExpired struct {
	Danger ids.ID `json:"danger"`
	Space  ids.ID `json:"space"`
}

func (DangerExpired) Kind() string { return "danger_expired" }

// mourn (fase Consequences, después de crush.resolve): por cada restos nuevos
// de un miembro, completa su colonia, emite member_lost y marca una zona de
// peligro que los de su escala evitan hasta que vence. La víctima todavía no
// se destruyó (eso pasa en el Flush), así que su "member" se puede leer. La
// colonia se entera en el acto: lo que les pasa a los suyos se sabe enseguida,
// como lo que ven.
//
// También borra las zonas vencidas.
type mourn struct{}

func (mourn) Name() string { return "colony.mourn" }

func (mourn) Run(w *world.World) {
	r := w.ECS()
	members := ecs.Get[Member](r)
	colonies := ecs.Get[Colony](r)
	dangers := ecs.Get[ai.Danger](r)
	locs := comp.Of(r).Location

	for _, e := range dangers.IDs() {
		d, _ := dangers.Get(e)
		if d.Until <= w.Tick() && !r.Despawning(e) {
			r.Despawn(e)
			w.Emit(DangerExpired{Danger: e, Space: d.Space})
		}
	}

	type loss struct {
		remains ids.ID
		m       crush.Remains
	}
	var lost []loss
	ecs.Get[crush.Remains](r).Each(func(e ids.ID, m *crush.Remains) {
		if m.Tick == w.Tick() && m.Colony.IsNil() {
			if mem, ok := members.Get(m.Of); ok {
				m.Colony = mem.Colony
				lost = append(lost, loss{e, *m})
			}
		}
	})
	for _, l := range lost {
		c, _ := colonies.Get(l.m.Colony)
		loc, ok := locs.Get(l.remains)
		if !ok {
			continue
		}
		scale := "tiny"
		if a, ok := ecs.Get[ai.Agent](r).Get(l.m.Of); ok && a.Scale != "" {
			scale = a.Scale
		}
		zone := r.Spawn()
		dangers.Set(zone, ai.Danger{
			Scale: scale, Space: loc.Space, Center: loc.Pos.XY(),
			Radius: orDefault(c.DangerRadius, DefaultDangerRadius),
			Until:  w.Tick() + orDefault(c.DangerTicks, DefaultDangerTicks),
		})
		w.Emit(MemberLost{Colony: l.m.Colony, Entity: l.m.Of, Cause: l.m.Source, Remains: l.remains, Danger: zone})
	}
}

func orDefault[T units.Length | clock.Tick](v, def T) T {
	if v > 0 {
		return v
	}
	return def
}
