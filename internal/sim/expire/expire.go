// Package expire implementa vencimiento genérico de entidades: todo lo que
// tiene el componente "expires" hace despawn al llegar a su tick de vencimiento
// (en orden de ID) y emite Expired.
package expire

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/world"
)

// Expires marca una entidad con fecha de vencimiento (componente "expires").
// Si At == 0 y TTL > 0, el sistema fija At = w.Tick() + TTL la primera vez que
// corre (útil para objetos creados por plantillas con StepSpawn).
type Expires struct {
	At   clock.Tick `json:"at,omitempty"`
	TTL  clock.Tick `json:"ttl,omitempty"`
	Kind string     `json:"kind,omitempty"`
}

// Expired se emite cuando una entidad vence y es marcada para despawn.
type Expired struct {
	Entity   ids.ID `json:"entity"`
	ItemKind string `json:"kind,omitempty"`
}

func (Expired) Kind() string { return "expired" }

type system struct{}

func (system) Name() string { return "expire" }

func (system) Run(w *world.World) {
	r := w.ECS()
	store := ecs.Get[Expires](r)
	names := comp.Of(r).Name
	now := w.Tick()

	type dueItem struct {
		id   ids.ID
		kind string
	}
	var due []dueItem

	store.Each(func(id ids.ID, ex *Expires) {
		if r.Despawning(id) {
			return
		}
		if ex.At == 0 && ex.TTL > 0 {
			ex.At = now + ex.TTL
			ex.TTL = 0
		}
		if now >= ex.At {
			kind := ex.Kind
			if kind == "" {
				if n, ok := names.Get(id); ok {
					kind = string(n)
				}
			}
			due = append(due, dueItem{id: id, kind: kind})
		}
	})

	slices.SortFunc(due, func(a, b dueItem) int { return cmp.Compare(a.id, b.id) })
	for _, d := range due {
		r.Despawn(d.id)
		w.Emit(Expired{Entity: d.id, ItemKind: d.kind})
	}
}

func check(w *world.World) error {
	r := w.ECS()
	now := w.Tick()
	var err error
	ecs.Get[Expires](r).Each(func(id ids.ID, ex *Expires) {
		if r.Despawning(id) {
			return
		}
		if ex.At == 0 && ex.TTL > 0 {
			return // recién creada antes de su primer Step
		}
		if now > 0 && ex.At < now {
			err = errors.Join(err, fmt.Errorf("expire: %v venció en %v (tick actual %v)", id, ex.At, now))
		}
	})
	return err
}

// Install registra el store "expires", el evento "expired", la invariante y el
// sistema en la última fase (PhaseFlush).
func Install(w *world.World) error {
	if _, err := ecs.Register[Expires](w.ECS(), "expires"); err != nil {
		return err
	}
	w.AddInvariant("expire", check)
	return errors.Join(
		world.RegisterEvent[Expired](w),
		w.Register(world.SystemSpec{Phase: clock.PhaseFlush, Order: 0}, system{}),
	)
}
