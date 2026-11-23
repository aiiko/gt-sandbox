package colony

import (
	"errors"
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Refugio auxiliar. Una obra (build_site) junto a una ruta de salida pide
// fibras de papel a una fuente (fiber_source, la servilleta) cuando la
// despensa de su colonia está en la meta. Los encargos son trabajos: pluck
// (desprender una fibra de la fuente), haul (llevarla a la obra) y build
// (armar el refugio con todas las fibras). Al terminar, la obra recibe
// "shelter" y conserva su ID.

// Tipos de trabajo.
const (
	JobPluck = "pluck"
	JobBuild = "build"
)

// Tiempos de los encargos.
const (
	PluckTicks clock.Tick = 2 * clock.TPS
	BuildTicks clock.Tick = 5 * clock.TPS
)

// FiberSource es el componente "fiber_source" de la servilleta: cuántas
// fibras quedan y qué se crea al desprender una (Item).
type FiberSource struct {
	Left int32       `json:"left"`
	Item *cmds.Spawn `json:"item,omitempty"`
}

// Material es el componente "material" de una fibra: para qué obra es.
// Lo completa el trabajo pluck.
type Material struct {
	Site ids.ID `json:"site,omitempty"`
}

// BuildSite es el componente "build_site" de una obra. Colony es la colonia
// que la arma y Source, de dónde salen las fibras. Lo que llega a menos de
// Radius cuenta como entregado. Al terminar (Done) recibe Shelter.
type BuildSite struct {
	Colony    ids.ID        `json:"colony"`
	Source    ids.ID        `json:"source"`
	Required  int32         `json:"required"`
	Delivered int32         `json:"delivered,omitempty"`
	Radius    units.Length  `json:"radius"`
	Shelter   crush.Shelter `json:"shelter"`
	Done      bool          `json:"done,omitempty"`
}

// Eventos de la obra.
type (
	FiberPlucked struct {
		Entity ids.ID `json:"entity"`
		Source ids.ID `json:"source"`
		Fiber  ids.ID `json:"fiber"`
		Left   int32  `json:"left"`
	}
	FiberDelivered struct {
		Site  ids.ID `json:"site"`
		Item  ids.ID `json:"item"`
		Total int32  `json:"total"`
	}
	RefugeBuilt struct {
		Site ids.ID `json:"site"`
		By   ids.ID `json:"by"`
	}
)

func (FiberPlucked) Kind() string   { return "fiber_plucked" }
func (FiberDelivered) Kind() string { return "fiber_delivered" }
func (RefugeBuilt) Kind() string    { return "refuge_built" }

func init() {
	jobs.RegisterKind(JobPluck, pluck)
	jobs.RegisterKind(JobBuild, build)
}

// approachThen camina al objetivo del trabajo (etapa 0) y espera ticks
// (etapa 1). Devuelve Done al terminar la espera.
func approachThen(w *world.World, j *jobs.Job, actor ids.ID, moving *bool, ticks clock.Tick) jobs.Result {
	switch j.Stage {
	case 0:
		t, ok := comp.Of(w.ECS()).Location.Get(j.Target)
		if !ok {
			return jobs.Failed
		}
		switch nav.Approach(w, actor, t.Space, t.Pos.XY(), moving) {
		case nav.Arrived:
			j.Stage, j.Wait = 1, ticks
		case nav.Failed:
			return jobs.Failed
		}
	case 1:
		if j.Wait > 0 {
			j.Wait--
			return jobs.Running
		}
		return jobs.Done
	}
	return jobs.Running
}

// siteOf es la primera obra sin terminar que saca fibras de src.
func siteOf(w *world.World, src ids.ID) ids.ID {
	var out ids.ID
	ecs.Get[BuildSite](w.ECS()).Each(func(e ids.ID, s *BuildSite) {
		if out.IsNil() && s.Source == src && !s.Done && !w.ECS().Despawning(e) {
			out = e
		}
	})
	return out
}

// pluck: ir a la fuente, desprender una fibra y dejarla ahí (para su obra).
func pluck(w *world.World, j *jobs.Job, actor ids.ID, moving *bool) jobs.Result {
	src := ecs.Get[FiberSource](w.ECS()).Ptr(j.Target)
	if src == nil || src.Left <= 0 || src.Item == nil {
		return jobs.Failed
	}
	res := approachThen(w, j, actor, moving, PluckTicks)
	if res != jobs.Done {
		return res
	}
	l, ok := comp.Of(w.ECS()).Location.Get(actor)
	if !ok {
		return jobs.Failed
	}
	item := *src.Item
	item.Location = &comp.Location{Space: l.Space, Pos: units.V3(l.Pos.X, l.Pos.Y, 0)}
	id, err := item.Do(w)
	if err != nil {
		return jobs.Failed
	}
	space.Land(w, id)
	ecs.Get[Material](w.ECS()).Set(id, Material{Site: siteOf(w, j.Target)})
	src.Left--
	if a := ecs.Get[ai.Agent](w.ECS()).Ptr(actor); a != nil {
		ai.Learn(w, a, actor, id)
	}
	w.Emit(FiberPlucked{Entity: actor, Source: j.Target, Fiber: id, Left: src.Left})
	return jobs.Done
}

// build: ir a la obra con todas las fibras y armar el refugio.
func build(w *world.World, j *jobs.Job, actor ids.ID, moving *bool) jobs.Result {
	r := w.ECS()
	s := ecs.Get[BuildSite](r).Ptr(j.Target)
	if s == nil || s.Done || s.Delivered < s.Required {
		return jobs.Failed
	}
	res := approachThen(w, j, actor, moving, BuildTicks)
	if res != jobs.Done {
		return res
	}
	s.Done = true
	ecs.Get[crush.Shelter](r).Set(j.Target, s.Shelter)
	w.Emit(RefugeBuilt{Site: j.Target, By: actor})
	return jobs.Done
}

// site: cada obra sin terminar pide lo que le falta. Corre en Decide,
// después de colony.share (los trabajos nuevos se eligen en el mismo tick).
type site struct{}

func (site) Name() string { return "colony.site" }

func (site) Run(w *world.World) {
	r := w.ECS()
	sites := ecs.Get[BuildSite](r)
	for _, e := range sites.IDs() {
		s, _ := sites.Get(e)
		if s.Done || r.Despawning(e) {
			continue
		}
		loc, ok := comp.Of(r).Location.Get(e)
		if !ok {
			continue
		}
		post := func(typ string, target ids.ID) {
			_ = jobs.PostJob{Type: typ, Target: target, DestSpace: loc.Space, Dest: loc.Pos.XY()}.Apply(w)
		}
		// fibras de la obra: sueltas sin trabajo → acarreo
		fibers := int32(0)
		ecs.Get[Material](r).Each(func(f ids.ID, m *Material) {
			if m.Site != e || r.Despawning(f) {
				return
			}
			fibers++
			if !hauled(w, f) && !nearSite(w, s, loc, f) {
				post("haul", f)
			}
		})
		if s.Delivered >= s.Required {
			if !jobOn(w, JobBuild, e) {
				post(JobBuild, e)
			}
			continue
		}
		src, ok := ecs.Get[FiberSource](r).Get(s.Source)
		plucks := jobOn(w, JobPluck, s.Source)
		if !ok || src.Left <= 0 {
			if plucks {
				cancelJobsOn(w, JobPluck, s.Source)
			}
			continue
		}
		need := s.Required - s.Delivered - fibers
		if need > 0 && !plucks && Deficit(w, s.Colony) == 0 && colonyKnows(w, s.Colony, s.Source) {
			post(JobPluck, s.Source)
		}
	}
}

// nearSite: e está a menos de Radius de la obra.
func nearSite(w *world.World, s BuildSite, loc comp.Location, e ids.ID) bool {
	l, ok := comp.Of(w.ECS()).Location.Get(e)
	return ok && l.Space == loc.Space && l.Pos.XY().DistSq(loc.Pos.XY()) <= int64(s.Radius)*int64(s.Radius)
}

// jobOn: hay un trabajo de tipo typ sobre target.
func jobOn(w *world.World, typ string, target ids.ID) bool {
	found := false
	ecs.Get[jobs.Job](w.ECS()).Each(func(id ids.ID, j *jobs.Job) {
		found = found || j.Type == typ && j.Target == target && !w.ECS().Despawning(id)
	})
	return found
}

func cancelJobsOn(w *world.World, typ string, target ids.ID) {
	store := ecs.Get[jobs.Job](w.ECS())
	for _, id := range store.IDs() {
		if j, _ := store.Get(id); j.Type == typ && j.Target == target && !w.ECS().Despawning(id) {
			_ = jobs.CancelJob{Job: id}.Apply(w)
		}
	}
}

// colonyKnows: algún miembro de col conoce e.
func colonyKnows(w *world.World, col, e ids.ID) bool {
	agents := ecs.Get[ai.Agent](w.ECS())
	found := false
	ecs.Get[Member](w.ECS()).Each(func(m ids.ID, mm *Member) {
		if a := agents.Ptr(m); !found && a != nil && mm.Colony == col && !w.ECS().Despawning(m) {
			found = a.Knows(e)
		}
	})
	return found
}

// deliver: la fibra suelta, sin reservar y a menos de Radius de su obra
// cuenta una vez y desaparece. Corre en Consequences, como colony.store.
type deliver struct{}

func (deliver) Name() string { return "colony.deliver" }

func (deliver) Run(w *world.World) {
	r := w.ECS()
	sites := ecs.Get[BuildSite](r)
	carried := ecs.Get[jobs.Carried](r)
	mats := ecs.Get[Material](r)
	for _, f := range mats.IDs() {
		m, _ := mats.Get(f)
		s := sites.Ptr(m.Site)
		if s == nil || s.Done || s.Delivered >= s.Required || r.Despawning(f) || r.Despawning(m.Site) ||
			carried.Has(f) || !jobs.Holder(w, f).IsNil() {
			continue
		}
		loc, ok := comp.Of(r).Location.Get(m.Site)
		if !ok || !nearSite(w, *s, loc, f) {
			continue
		}
		s.Delivered++
		r.Despawn(f)
		w.Emit(FiberDelivered{Site: m.Site, Item: f, Total: s.Delivered})
	}
}

func checkSites(w *world.World) error {
	r := w.ECS()
	var err error
	shelters := ecs.Get[crush.Shelter](r)
	ecs.Get[BuildSite](r).Each(func(e ids.ID, s *BuildSite) {
		if s.Delivered < 0 || s.Delivered > s.Required {
			err = errors.Join(err, fmt.Errorf("colony: obra %v con %d de %d fibras", e, s.Delivered, s.Required))
		}
		if s.Done && !shelters.Has(e) {
			err = errors.Join(err, fmt.Errorf("colony: obra %v terminada sin shelter", e))
		}
	})
	return err
}

// registerBuild registra stores, eventos, sistemas e invariantes de la obra.
func registerBuild(w *world.World) error {
	r := w.ECS()
	for _, err := range []error{
		register[FiberSource](r, "fiber_source"), register[Material](r, "material"), register[BuildSite](r, "build_site"),
	} {
		if err != nil {
			return err
		}
	}
	w.AddInvariant("colony.build", checkSites)
	return errors.Join(
		world.RegisterEvent[FiberPlucked](w),
		world.RegisterEvent[FiberDelivered](w),
		world.RegisterEvent[RefugeBuilt](w),
		w.Register(world.SystemSpec{Phase: clock.PhaseDecide, Order: -39}, site{}),
		// Junto a colony.store.
		w.Register(world.SystemSpec{Phase: clock.PhaseConsequences, Order: 1<<21 + 1}, deliver{}),
	)
}
