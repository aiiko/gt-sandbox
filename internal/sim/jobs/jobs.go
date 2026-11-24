// Package jobs implementa reservas y trabajos.
//
// Una reserva (componente "reservation" en la entidad reservada) dice quién
// la tiene y hasta cuándo; como es un componente, una entidad no puede tener
// dos. El que la tiene la renueva mientras la usa; si deja de hacerlo, vence.
//
// Un trabajo es una entidad con "job": qué hacer, sobre qué objetivo, quién lo
// tomó y en qué etapa va. Tomar un trabajo exige reservar su objetivo. La
// etapa vive en el trabajo, así que un trabajo interrumpido se retoma donde
// quedó. Quién decide tomar trabajos es la IA (acción "work").
package jobs

import (
	"errors"
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// TTL es cuánto dura una reserva sin renovar (10 s).
const TTL clock.Tick = 10 * clock.TPS

// Reservation es el componente "reservation" de una entidad reservada.
type Reservation struct {
	By    ids.ID     `json:"by"`
	Until clock.Tick `json:"until"` // vence al terminar este tick
}

// Job es el componente "job" de un trabajo.
type Job struct {
	Type      string     `json:"type"`
	Target    ids.ID     `json:"target"`
	DestSpace ids.ID     `json:"dest_space,omitempty"`
	Dest      units.Vec2 `json:"dest"`
	Priority  int        `json:"priority,omitempty"`
	Assignee  ids.ID     `json:"assignee,omitempty"`
	Stage     int        `json:"stage,omitempty"`
	Wait      clock.Tick `json:"wait,omitempty"` // ticks que faltan en la etapa actual
	Since     clock.Tick `json:"since"`
}

// Carried es el componente "carried": la entidad viaja con By.
type Carried struct {
	By ids.ID `json:"by"`
}

// Eventos.
type (
	Reserved struct {
		Target ids.ID `json:"target"`
		By     ids.ID `json:"by"`
	}
	ReservationReleased struct {
		Target ids.ID `json:"target"`
		By     ids.ID `json:"by"`
		Reason string `json:"reason"` // released, expired, holder_gone
	}
	JobPosted struct {
		Job    ids.ID `json:"job"`
		Type   string `json:"type"`
		Target ids.ID `json:"target"`
	}
	JobAssigned struct {
		Job ids.ID `json:"job"`
		By  ids.ID `json:"by"`
	}
	JobUnassigned struct {
		Job    ids.ID `json:"job"`
		By     ids.ID `json:"by"`
		Reason string `json:"reason"`
	}
	JobDone struct {
		Job ids.ID `json:"job"`
		By  ids.ID `json:"by"`
	}
	JobCanceled struct {
		Job    ids.ID `json:"job"`
		Reason string `json:"reason"`
	}
)

func (Reserved) Kind() string            { return "reserved" }
func (ReservationReleased) Kind() string { return "reservation_released" }
func (JobPosted) Kind() string           { return "job_posted" }
func (JobAssigned) Kind() string         { return "job_assigned" }
func (JobUnassigned) Kind() string       { return "job_unassigned" }
func (JobDone) Kind() string             { return "job_done" }
func (JobCanceled) Kind() string         { return "job_canceled" }

// --- Reservas ---

// Holder devuelve quién tiene reservada e (Nil si nadie o si venció).
func Holder(w *world.World, e ids.ID) ids.ID {
	r, ok := ecs.Get[Reservation](w.ECS()).Get(e)
	if !ok || r.Until < w.Tick() || !w.ECS().Alive(r.By) {
		return ids.Nil
	}
	return r.By
}

// Reserve reserva e para by hasta dentro de TTL ticks, o renueva la reserva
// si ya es suya. Falla si e no existe o la tiene otro.
func Reserve(w *world.World, e, by ids.ID) bool {
	r := w.ECS()
	if !r.Alive(e) || r.Despawning(e) {
		return false
	}
	h := Holder(w, e)
	if !h.IsNil() && h != by {
		return false
	}
	ecs.Get[Reservation](r).Set(e, Reservation{By: by, Until: w.Tick() + TTL})
	if h.IsNil() {
		w.Emit(Reserved{Target: e, By: by})
	}
	return true
}

// Release libera la reserva de e si es de by.
func Release(w *world.World, e, by ids.ID) {
	release(w, e, by, "released")
}

func release(w *world.World, e, by ids.ID, reason string) {
	s := ecs.Get[Reservation](w.ECS())
	if r, ok := s.Get(e); ok && r.By == by {
		s.Remove(e)
		w.Emit(ReservationReleased{Target: e, By: by, Reason: reason})
	}
}

// --- Trabajos ---

// PostJob publica un trabajo. Type debe ser un tipo registrado (Kinds).
// DestSpace vacío = el Space donde está el objetivo.
type PostJob struct {
	Type      string     `json:"type"`
	Target    ids.ID     `json:"target"`
	DestSpace ids.ID     `json:"dest_space,omitempty"`
	Dest      units.Vec2 `json:"dest"`
	Priority  int        `json:"priority,omitempty"`
}

func (PostJob) Kind() string { return "post_job" }

func (c PostJob) Apply(w *world.World) error {
	r := w.ECS()
	if _, ok := Kinds[c.Type]; !ok {
		return fmt.Errorf("post_job: tipo desconocido %q", c.Type)
	}
	loc, ok := comp.Of(r).Location.Get(c.Target)
	if !ok {
		return fmt.Errorf("post_job: el objetivo %v no está ubicado", c.Target)
	}
	dest := c.DestSpace
	if dest.IsNil() {
		dest = loc.Space
	}
	if !space.Of(r).IsSpace(dest) {
		return fmt.Errorf("post_job: destino %v: %w", dest, space.ErrNotSpace)
	}
	id := r.Spawn()
	ecs.Get[Job](r).Set(id, Job{Type: c.Type, Target: c.Target, DestSpace: dest, Dest: c.Dest,
		Priority: c.Priority, Since: w.Tick()})
	w.Emit(JobPosted{Job: id, Type: c.Type, Target: c.Target})
	return nil
}

// CancelJob cancela un trabajo: libera la reserva y suelta lo que se lleve.
type CancelJob struct {
	Job ids.ID `json:"job"`
}

func (CancelJob) Kind() string { return "cancel_job" }

func (c CancelJob) Apply(w *world.World) error {
	j, ok := ecs.Get[Job](w.ECS()).Get(c.Job)
	if !ok || w.ECS().Despawning(c.Job) {
		return fmt.Errorf("cancel_job: %v no es un trabajo", c.Job)
	}
	cancel(w, c.Job, j, "canceled")
	return nil
}

func cancel(w *world.World, id ids.ID, j Job, reason string) {
	if !j.Assignee.IsNil() {
		drop(w, j.Target, j.Assignee)
		release(w, j.Target, j.Assignee, "released")
	}
	w.ECS().Despawn(id)
	w.Emit(JobCanceled{Job: id, Reason: reason})
}

// Open indica si el trabajo puede tomarse: vivo, sin asignar, con el
// objetivo vivo y sin reservar por otro.
func Open(w *world.World, id ids.ID) (Job, bool) {
	j, ok := ecs.Get[Job](w.ECS()).Get(id)
	if !ok || w.ECS().Despawning(id) || !j.Assignee.IsNil() || !w.ECS().Alive(j.Target) {
		return j, false
	}
	return j, Holder(w, j.Target).IsNil()
}

// Assign da el trabajo a by, reservando el objetivo. Falla sin modificar nada
// si el trabajo no está abierto.
func Assign(w *world.World, id, by ids.ID) bool {
	if _, ok := Open(w, id); !ok {
		return false
	}
	j := ecs.Get[Job](w.ECS()).Ptr(id)
	if !Reserve(w, j.Target, by) {
		return false
	}
	j.Assignee = by
	w.Emit(JobAssigned{Job: id, By: by})
	return true
}

// Unassign devuelve el trabajo a la cola: suelta lo que se lleve, libera la
// reserva y vuelve a la primera etapa.
func Unassign(w *world.World, id ids.ID, reason string) {
	j := ecs.Get[Job](w.ECS()).Ptr(id)
	if j == nil || j.Assignee.IsNil() {
		return
	}
	by := j.Assignee
	drop(w, j.Target, by)
	release(w, j.Target, by, "released")
	j.Assignee, j.Stage, j.Wait = ids.Nil, 0, 0
	w.Emit(JobUnassigned{Job: id, By: by, Reason: reason})
}

// AssignedTo devuelve el trabajo asignado a by, si tiene.
func AssignedTo(w *world.World, by ids.ID) ids.ID {
	var out ids.ID
	ecs.Get[Job](w.ECS()).Each(func(id ids.ID, j *Job) {
		if out.IsNil() && j.Assignee == by && !w.ECS().Despawning(id) {
			out = id
		}
	})
	return out
}

func complete(w *world.World, id ids.ID, j *Job) {
	release(w, j.Target, j.Assignee, "released")
	w.ECS().Despawn(id)
	w.Emit(JobDone{Job: id, By: j.Assignee})
}

// drop suelta e si lo lleva by: queda donde está.
func drop(w *world.World, e, by ids.ID) {
	s := ecs.Get[Carried](w.ECS())
	if c, ok := s.Get(e); ok && c.By == by {
		s.Remove(e)
	}
}

// --- Sistemas ---

// upkeep vence reservas, cancela trabajos sin objetivo y devuelve a la cola
// los de agentes que murieron o perdieron la reserva.
type upkeep struct{}

func (upkeep) Name() string { return "jobs.upkeep" }

func (upkeep) Run(w *world.World) {
	r := w.ECS()
	res := ecs.Get[Reservation](r)
	for _, e := range res.IDs() {
		v, _ := res.Get(e)
		switch {
		case !r.Alive(v.By) || r.Despawning(v.By):
			release(w, e, v.By, "holder_gone")
		case v.Until < w.Tick():
			release(w, e, v.By, "expired")
		}
	}
	jobs := ecs.Get[Job](r)
	for _, id := range jobs.IDs() {
		j, _ := jobs.Get(id)
		switch {
		case r.Despawning(id):
		case !r.Alive(j.Target) || r.Despawning(j.Target):
			cancel(w, id, j, "target_gone")
		case j.Assignee.IsNil():
		case !r.Alive(j.Assignee) || r.Despawning(j.Assignee):
			Unassign(w, id, "assignee_gone")
		case Holder(w, j.Target) != j.Assignee:
			Unassign(w, id, "reservation_lost")
		}
	}
	carried := ecs.Get[Carried](r)
	for _, e := range carried.IDs() {
		if c, _ := carried.Get(e); !r.Alive(c.By) || r.Despawning(c.By) {
			carried.Remove(e)
		}
	}
}

// carry lleva lo cargado a la posición de quien lo lleva.
type carry struct{}

func (carry) Name() string { return "jobs.carry" }

func (carry) Run(w *world.World) {
	locs := comp.Of(w.ECS()).Location
	ecs.Get[Carried](w.ECS()).Each(func(e ids.ID, c *Carried) {
		l, ok := locs.Get(c.By)
		if p := locs.Ptr(e); ok && p != nil {
			p.Space, p.Pos = l.Space, l.Pos
		}
	})
}

func check(w *world.World) error {
	r := w.ECS()
	var err error
	ecs.Get[Job](r).Each(func(id ids.ID, j *Job) {
		if r.Despawning(id) || j.Assignee.IsNil() {
			return
		}
		if res, ok := ecs.Get[Reservation](r).Get(j.Target); !ok || res.By != j.Assignee {
			err = errors.Join(err, fmt.Errorf("jobs: %v asignado a %v sin reservar %v", id, j.Assignee, j.Target))
		}
	})
	ecs.Get[Carried](r).Each(func(e ids.ID, c *Carried) {
		if res, ok := ecs.Get[Reservation](r).Get(e); r.Alive(c.By) && (!ok || res.By != c.By) {
			err = errors.Join(err, fmt.Errorf("jobs: %v cargado por %v sin reserva", e, c.By))
		}
	})
	return err
}

// Install registra stores, comandos, eventos, sistemas e invariantes.
func Install(w *world.World) error {
	r := w.ECS()
	for _, err := range []error{
		register[Reservation](r, "reservation"), register[Job](r, "job"), register[Carried](r, "carried"),
	} {
		if err != nil {
			return err
		}
	}
	w.AddInvariant("jobs", check)
	return errors.Join(
		world.RegisterCommand[PostJob](w),
		world.RegisterCommand[CancelJob](w),
		world.RegisterEvent[Reserved](w),
		world.RegisterEvent[ReservationReleased](w),
		world.RegisterEvent[JobPosted](w),
		world.RegisterEvent[JobAssigned](w),
		world.RegisterEvent[JobUnassigned](w),
		world.RegisterEvent[JobDone](w),
		world.RegisterEvent[JobCanceled](w),
		// Después del movimiento y antes de reconstruir el índice espacial.
		w.Register(world.SystemSpec{Phase: clock.PhaseMove, Order: 1 << 20}, carry{}),
		w.Register(world.SystemSpec{Phase: clock.PhaseFlush, Order: 1 << 20}, upkeep{}),
	)
}

func register[T any](r *ecs.Registry, name string) error {
	_, err := ecs.Register[T](r, name)
	return err
}
