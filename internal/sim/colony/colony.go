// Package colony implementa colonias de tinies: un nido con depósito de
// comida, miembros que comparten lo que ven, trabajos de acarreo que la
// colonia publica sola y la acción "forage" (salir a buscar).
//
// La colonia es una entidad ubicada en el nido, con "colony" y además un
// smart object: sus ofertas (comer del depósito, descansar) y su depósito
// son componentes de ai ("offers", "portions").
package colony

import (
	"errors"
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/spatial"
	"gtsandbox/internal/sim/world"
)

// Colony es el componente "colony" de la entidad nido. Lo que llega a menos
// de Radius del nido se guarda en el depósito.
type Colony struct {
	Radius units.Length `json:"radius"`
}

// Member es el componente "member" de un tiny: a qué colonia pertenece y
// cuánto le gusta salir a buscar cuando no tiene nada mejor (utilidad).
type Member struct {
	Colony ids.ID  `json:"colony"`
	Forage float64 `json:"forage,omitempty"`
}

// Food marca algo que la colonia junta (componente "food"). Al guardarlo
// suma sus porciones (o 1) al depósito.
type Food struct{}

// Stored se emite al guardar comida en el depósito.
type Stored struct {
	Colony   ids.ID `json:"colony"`
	Item     ids.ID `json:"item"`
	Portions int32  `json:"portions"`
	Total    int32  `json:"total"`
}

func (Stored) Kind() string { return "stored" }

// share: los miembros conocen su nido y se cuentan lo que vieron; la comida
// nueva genera un trabajo de acarreo al nido.
type share struct{}

func (share) Name() string { return "colony.share" }

func (share) Run(w *world.World) {
	r := w.ECS()
	members := ecs.Get[Member](r)
	agents := ecs.Get[ai.Agent](r)
	members.Each(func(e ids.ID, m *Member) {
		if a := agents.Ptr(e); a != nil && !r.Despawning(e) {
			ai.Learn(w, a, e, m.Colony)
		}
	})
	for _, ev := range w.Events() { // del tick anterior
		n, ok := ev.Event.(ai.Noticed)
		if !ok {
			continue
		}
		m, ok := members.Get(n.Entity)
		if !ok || !r.Alive(n.Target) || r.Despawning(n.Target) {
			continue
		}
		members.Each(func(e ids.ID, o *Member) {
			if a := agents.Ptr(e); a != nil && o.Colony == m.Colony && e != n.Entity && !r.Despawning(e) {
				ai.Learn(w, a, e, n.Target)
			}
		})
		if ecs.Get[Food](r).Has(n.Target) && !atNest(w, m.Colony, n.Target) && !hauled(w, n.Target) {
			nest, ok := comp.Of(r).Location.Get(m.Colony)
			if !ok {
				continue
			}
			_ = jobs.PostJob{Type: "haul", Target: n.Target, DestSpace: nest.Space, Dest: nest.Pos.XY()}.Apply(w)
		}
	}
}

// hauled: ya hay un trabajo sobre e, o alguien lo carga.
func hauled(w *world.World, e ids.ID) bool {
	if ecs.Get[jobs.Carried](w.ECS()).Has(e) {
		return true
	}
	found := false
	ecs.Get[jobs.Job](w.ECS()).Each(func(id ids.ID, j *jobs.Job) {
		found = found || j.Target == e && !w.ECS().Despawning(id)
	})
	return found
}

// atNest: e está a menos de Radius del nido col.
func atNest(w *world.World, col, e ids.ID) bool {
	c, ok := ecs.Get[Colony](w.ECS()).Get(col)
	locs := comp.Of(w.ECS()).Location
	nest, ok2 := locs.Get(col)
	l, ok3 := locs.Get(e)
	return ok && ok2 && ok3 && l.Space == nest.Space &&
		l.Pos.XY().DistSq(nest.Pos.XY()) <= int64(c.Radius)*int64(c.Radius)
}

// store guarda en el depósito la comida suelta que está en el nido.
type store struct{}

func (store) Name() string { return "colony.store" }

func (store) Run(w *world.World) {
	r := w.ECS()
	locs := comp.Of(r).Location
	food := ecs.Get[Food](r)
	portions := ecs.Get[ai.Portions](r)
	ecs.Get[Colony](r).Each(func(col ids.ID, c *Colony) {
		nest, ok := locs.Get(col)
		if !ok || r.Despawning(col) {
			return
		}
		var items []ids.ID
		spatial.Of(w).Near(nest.Space, nest.Pos.XY(), c.Radius, func(e ids.ID, _ units.Vec2) {
			if food.Has(e) && !r.Despawning(e) && !ecs.Get[jobs.Carried](r).Has(e) && jobs.Holder(w, e).IsNil() {
				items = append(items, e)
			}
		})
		for _, e := range items {
			n := int32(1)
			if p, ok := portions.Get(e); ok {
				n = p.Left
			}
			dep, _ := portions.Get(col)
			dep.Left += n
			portions.Set(col, dep)
			r.Despawn(e)
			w.Emit(Stored{Colony: col, Item: e, Portions: n, Total: dep.Left})
		}
	})
}

func check(w *world.World) error {
	r := w.ECS()
	var err error
	ecs.Get[Member](r).Each(func(e ids.ID, m *Member) {
		if r.Alive(m.Colony) && !ecs.Get[Colony](r).Has(m.Colony) {
			err = errors.Join(err, fmt.Errorf("colony: %v es miembro de %v, que no es una colonia", e, m.Colony))
		}
	})
	return err
}

// Install registra stores, eventos, sistemas e invariantes. Requiere ai y
// jobs; la acción Forage se pasa a ai.Install.
func Install(w *world.World) error {
	r := w.ECS()
	for _, err := range []error{
		register[Colony](r, "colony"), register[Member](r, "member"), register[Food](r, "food"),
	} {
		if err != nil {
			return err
		}
	}
	w.AddInvariant("colony", check)
	return errors.Join(
		world.RegisterEvent[Stored](w),
		// Después de ai.look (lo visto este tick se comparte el siguiente) y
		// antes de ai.decide.
		w.Register(world.SystemSpec{Phase: clock.PhaseDecide, Order: -40}, share{}),
		w.Register(world.SystemSpec{Phase: clock.PhaseConsequences, Order: 1 << 21}, store{}),
	)
}

func register[T any](r *ecs.Registry, name string) error {
	_, err := ecs.Register[T](r, name)
	return err
}
