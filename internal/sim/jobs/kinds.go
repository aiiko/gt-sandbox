package jobs

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/world"
)

// Result es el resultado de avanzar un trabajo un tick.
type Result uint8

const (
	Running Result = iota
	Done
	Failed
)

// Executor avanza un trabajo de tipo dado un tick, para actor. moving es el
// estado de nav.Approach del que ejecuta (vive en su acción, no en el job).
type Executor func(w *world.World, j *Job, actor ids.ID, moving *bool) Result

// Kinds son los tipos de trabajo conocidos.
var Kinds = map[string]Executor{"haul": haul}

// HandleTicks es lo que tarda levantar o dejar algo (1 s).
const HandleTicks clock.Tick = clock.TPS

// Step avanza el trabajo id, que debe estar asignado a actor, renovando la
// reserva. Al terminar lo completa; si falla, lo devuelve a la cola.
func Step(w *world.World, id, actor ids.ID, moving *bool) Result {
	j := ecs.Get[Job](w.ECS()).Ptr(id)
	if j == nil || w.ECS().Despawning(id) || j.Assignee != actor || !Reserve(w, j.Target, actor) {
		return Failed
	}
	switch Kinds[j.Type](w, j, actor, moving) {
	case Done:
		complete(w, id, j)
		return Done
	case Failed:
		Unassign(w, id, "failed")
		return Failed
	}
	return Running
}

// haul: ir al objetivo, levantarlo, llevarlo al destino y dejarlo.
func haul(w *world.World, j *Job, actor ids.ID, moving *bool) Result {
	locs := comp.Of(w.ECS()).Location
	carried := ecs.Get[Carried](w.ECS())
	wait := func(next int) Result {
		if j.Wait > 0 {
			j.Wait--
			return Running
		}
		j.Stage = next
		return Running
	}
	switch j.Stage {
	case 0: // ir al objetivo
		t, ok := locs.Get(j.Target)
		if !ok {
			return Failed
		}
		switch nav.Approach(w, actor, t.Space, t.Pos.XY(), moving) {
		case nav.Arrived:
			j.Stage, j.Wait = 1, HandleTicks
		case nav.Failed:
			return Failed
		}
	case 1: // levantar
		if j.Wait == 0 {
			carried.Set(j.Target, Carried{By: actor})
		}
		return wait(2)
	case 2: // llevar
		switch nav.Approach(w, actor, j.DestSpace, j.Dest, moving) {
		case nav.Arrived:
			j.Stage, j.Wait = 3, HandleTicks
		case nav.Failed:
			return Failed
		}
	case 3: // dejar
		if j.Wait > 0 {
			j.Wait--
			return Running
		}
		carried.Remove(j.Target)
		if p := locs.Ptr(j.Target); p != nil {
			a := locs.Ptr(actor)
			p.Space, p.Pos = a.Space, units.V3(a.Pos.X, a.Pos.Y, 0)
		}
		return Done
	}
	return Running
}
