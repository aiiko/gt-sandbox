package attach

import (
	"errors"
	"fmt"

	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// PartPose es la ubicación de una parte del cuerpo respecto de su Space
// contenedor (que no cambia: una pose mueve partes, no las reengancha).
type PartPose struct {
	Part   ids.ID          `json:"part"`
	Pos    units.Vec3      `json:"pos"`
	Facing fixedmath.Angle `json:"facing"`
}

// Poses es el componente de un cuerpo (store "poses"): poses con nombre,
// cada una un conjunto de transforms de sus partes. No hay esqueleto: una
// pose es solo cambiar Locations.
type Poses map[string][]PartPose

// SetPose aplica la pose Pose del cuerpo Entity. Es instantáneo (la
// interpolación vendrá con la animación). Todas las partes deben estar
// montadas, directa o indirectamente, sobre Entity.
type SetPose struct {
	Entity ids.ID `json:"entity"`
	Pose   string `json:"pose"`
}

func (SetPose) Kind() string { return "set_pose" }

func (c SetPose) Apply(w *world.World) error {
	r := w.ECS()
	poses, ok := ecs.Get[Poses](r).Get(c.Entity)
	if !ok {
		return fmt.Errorf("set_pose %v: no tiene poses", c.Entity)
	}
	parts, ok := poses[c.Pose]
	if !ok {
		return fmt.Errorf("set_pose %v: pose desconocida %q", c.Entity, c.Pose)
	}
	g := space.Of(r)
	locs := comp.Of(r).Location
	for _, p := range parts {
		if !locs.Has(p.Part) || !mountedOn(g, p.Part, c.Entity) {
			return fmt.Errorf("set_pose %v: %v no es parte del cuerpo", c.Entity, p.Part)
		}
	}
	for _, p := range parts {
		l := locs.Ptr(p.Part)
		l.Pos, l.Facing = p.Pos, p.Facing
	}
	w.Emit(PoseChanged{Entity: c.Entity, Pose: c.Pose})
	return nil
}

// mountedOn indica si e está montado (directa o indirectamente) sobre body.
func mountedOn(g space.Graph, e, body ids.ID) bool {
	cur := g.Parent(e)
	for depth := 0; !cur.IsNil() && depth <= space.MaxDepth; depth++ {
		if cur == body {
			return true
		}
		cur = g.Parent(cur)
	}
	return false
}

type PoseChanged struct {
	Entity ids.ID `json:"entity"`
	Pose   string `json:"pose"`
}

func (PoseChanged) Kind() string { return "pose_changed" }

// Install registra stores, comandos, eventos y sistemas de attach. Requiere
// comp, space y nav.
func Install(w *world.World) error {
	r := w.ECS()
	if _, err := ecs.Register[Surface](r, "contact_surface"); err != nil {
		return err
	}
	if _, err := ecs.Register[Contact](r, "contact"); err != nil {
		return err
	}
	if _, err := ecs.Register[Poses](r, "poses"); err != nil {
		return err
	}
	w.AddInvariant("attach.contact", checkContacts)
	return errors.Join(
		world.RegisterCommand[SetPose](w),
		world.RegisterEvent[PoseChanged](w),
		world.RegisterEvent[ContactBegan](w),
		world.RegisterEvent[ContactEnded](w),
		world.RegisterEvent[PortalOpened](w),
		world.RegisterEvent[PortalClosed](w),
		w.Register(world.SystemSpec{Phase: clock.PhaseTopology}, contacts{}),
		// Mismo Order que space.orphans y nav.portals, registrado después.
		w.Register(world.SystemSpec{Phase: clock.PhaseFlush, Order: 1 << 30}, contactCleanup{}),
	)
}
