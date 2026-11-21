package space

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/world"
)

// SupportTolerance es cuánto puede despegarse una entidad de su superficie
// (en Z local) y seguir apoyada.
const SupportTolerance = units.Millimeter

// EntityFell se emite cuando una entidad sin apoyo aterriza en otra
// superficie. Height es la altura de la caída.
type EntityFell struct {
	Entity ids.ID       `json:"entity"`
	From   ids.ID       `json:"from"`
	To     ids.ID       `json:"to"`
	Height units.Length `json:"height"`
}

func (EntityFell) Kind() string { return "entity_fell" }

// Supported indica si e está apoyada en su Space: dentro de la extensión y
// sobre la superficie (|Z| ≤ SupportTolerance).
func (g Graph) Supported(e ids.ID) bool {
	l, ok := g.loc.Get(e)
	if !ok {
		return true
	}
	sp, ok := g.spaces.Get(l.Space)
	return ok && l.Pos.Z.Abs() <= SupportTolerance && sp.Extent.Contains(l.Pos.XY())
}

// Below busca la superficie más alta debajo de e (sin contar e ni lo montado
// sobre e): un Space de la misma raíz cuya extensión contiene la vertical de
// e y cuya superficie está a la altura de e o más abajo. Devuelve también la
// posición local de e en esa superficie y la altura de la caída.
func (g Graph) Below(e ids.ID) (target ids.ID, local units.Vec3, height units.Length, ok bool) {
	pos, _, err := g.WorldPose(e)
	if err != nil {
		return ids.Nil, units.Vec3{}, 0, false
	}
	root, err := g.Root(e)
	if err != nil {
		return ids.Nil, units.Vec3{}, 0, false
	}
	g.spaces.Each(func(s ids.ID, sp *Space) {
		if s == e || sp.Extent.Empty() || g.reg.Despawning(s) || g.CanPlace(e, s) != nil {
			return
		}
		if r, err := g.Root(s); err != nil || r != root {
			return
		}
		l, err := g.FromWorld(s, pos)
		if err != nil || l.Z < -SupportTolerance || !sp.Extent.Contains(l.XY()) {
			return
		}
		if !ok || l.Z < height || (l.Z == height && s < target) {
			target, local, height, ok = s, l, l.Z, true
		}
	})
	return target, local, max(height, 0), ok
}

// Land hace aterrizar ya mismo a e si no está apoyada, como lo haría
// space.falls al final del tick (emite EntityFell). Sirve para apoyar
// objetos que un sistema deja en el aire, p. ej. a la altura de una mano.
func Land(w *world.World, e ids.ID) bool {
	g := Of(w.ECS())
	if g.spaces.Has(e) || g.Supported(e) {
		return false
	}
	to, local, h, ok := g.Below(e)
	if !ok {
		return false
	}
	return land(w, g, e, g.Parent(e), to, local, h)
}

func land(w *world.World, g Graph, e, from, to ids.ID, local units.Vec3, h units.Length) bool {
	if _, err := g.Reparent(e, to); err != nil {
		return false // misma raíz y CanPlace ya verificados
	}
	g.loc.Ptr(e).Pos = units.V3(local.X, local.Y, 0)
	w.Emit(EntityFell{Entity: e, From: from, To: to, Height: h})
	return true
}

// falls hace aterrizar a las entidades físicas (con comp.Body, que no sean
// Spaces) que quedaron sin apoyo: porque su Space desapareció y las reubicó
// space.orphans, porque una pose las dejó en el aire, etc. La caída es
// instantánea; el daño y la animación son de otros sistemas.
//
// Si no hay nada debajo, la entidad queda donde está (sin evento).
type falls struct{}

func (falls) Name() string { return "space.falls" }

func (falls) Run(w *world.World) {
	r := w.ECS()
	g := Of(r)
	type landing struct {
		e, from, to ids.ID
		local       units.Vec3
		h           units.Length
	}
	var todo []landing
	ecs.Get[comp.Body](r).Each(func(e ids.ID, _ *comp.Body) {
		if g.spaces.Has(e) || r.Despawning(e) || g.Supported(e) {
			return
		}
		if to, local, h, ok := g.Below(e); ok {
			todo = append(todo, landing{e, g.Parent(e), to, local, h})
		}
	})
	for _, l := range todo {
		land(w, g, l.e, l.from, l.to, l.local, l.h)
	}
}
