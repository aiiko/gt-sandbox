package space

import (
	"errors"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/world"
)

// EntityReparented se emite cuando una entidad cambia de Space.
type EntityReparented struct {
	Entity ids.ID `json:"entity"`
	From   ids.ID `json:"from"`
	To     ids.ID `json:"to"`
	Cause  string `json:"cause"` // "command", "space_destroyed", ...
}

func (EntityReparented) Kind() string { return "entity_reparented" }

// Install registra el store "space", sus eventos, los sistemas de huérfanos
// y caídas, y las invariantes. Requiere que comp ya esté registrado.
func Install(w *world.World) error {
	if _, err := ecs.Register[Space](w.ECS(), "space"); err != nil {
		return err
	}
	err := errors.Join(
		world.RegisterEvent[EntityReparented](w),
		world.RegisterEvent[EntityFell](w),
		w.Register(world.SystemSpec{Phase: clock.PhaseConsequences}, falls{}),
		// Último de la fase Flush: después de él solo queda ecs.Flush.
		w.Register(world.SystemSpec{Phase: clock.PhaseFlush, Order: 1 << 30}, orphans{}),
	)
	w.AddInvariant("space", func(w *world.World) error { return Of(w.ECS()).CheckInvariants() })
	return err
}

// orphans reubica a los ocupantes de un Space que va a destruirse en el
// ancestro vivo más cercano, conservando su posición en el mundo (después
// otros sistemas decidirán si caen). Si no queda ningún ancestro (se destruye
// la raíz), los ocupantes se destruyen en cascada.
type orphans struct{}

func (orphans) Name() string { return "space.orphans" }

func (orphans) Run(w *world.World) {
	r := w.ECS()
	g := Of(r)
	for i := 0; i < r.PendingDespawns(); i++ { // la lista crece con las cascadas
		dying := r.PendingAt(i)
		if !g.spaces.Has(dying) {
			continue
		}
		target := g.Parent(dying)
		for !target.IsNil() && r.Despawning(target) {
			target = g.Parent(target)
		}
		var occupants []ids.ID
		g.loc.Each(func(e ids.ID, l *comp.Location) {
			if l.Space == dying && !r.Despawning(e) {
				occupants = append(occupants, e)
			}
		})
		for _, e := range occupants {
			if target.IsNil() {
				r.Despawn(e)
				continue
			}
			if _, err := g.Reparent(e, target); err != nil {
				panic(err) // target es ancestro vivo: no puede fallar
			}
			w.Emit(EntityReparented{Entity: e, From: dying, To: target, Cause: "space_destroyed"})
		}
	}
}
