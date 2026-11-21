// Package ai elige y ejecuta acciones de los agentes con utility AI.
//
// Cada agente (componente "agent") hace a lo sumo una acción a la vez. Cada
// DecidePeriod ticks (repartido entre agentes) el selector junta las opciones
// de todas las acciones registradas, se queda con la de mayor utilidad y, si
// supera a la actual por más que la inercia, la cambia (histéresis: así no
// oscila entre dos opciones parecidas). La ejecución avanza un paso por tick.
//
// Los objetos declaran qué ofrecen (componente "offers", smart objects); la
// acción "use" los aprovecha. La acción "work" toma trabajos de la cola de
// jobs. Ciertos eventos (una caída) suspenden la acción en curso; tras
// recuperarse, el agente la retoma.
package ai

import (
	"cmp"
	"errors"
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Parámetros del selector.
const (
	DecidePeriod   clock.Tick = 10 // cada cuánto decide un agente
	RecoverTicks   clock.Tick = 2 * clock.TPS
	DefaultInertia            = 0.15 // ventaja de la acción en curso
	// MinUtility: por debajo de esto el agente prefiere no hacer nada.
	MinUtility = 0.01
)

// Weight es la importancia de una necesidad para un agente (personalidad).
type Weight struct {
	Need string  `json:"need"`
	W    float64 `json:"w"`
}

// Doing es una acción en curso. Stage y Progress los usa la acción; Moving
// es el estado de nav.Approach.
type Doing struct {
	Action   string     `json:"action"`
	Target   ids.ID     `json:"target"`
	Offer    int        `json:"offer,omitempty"`
	Stage    int        `json:"stage,omitempty"`
	Progress clock.Tick `json:"progress,omitempty"`
	Moving   bool       `json:"moving,omitempty"`
	Since    clock.Tick `json:"since"`
}

// Agent es el componente "agent".
type Agent struct {
	Scale        string   `json:"scale"`             // qué ofertas puede usar
	Weights      []Weight `json:"weights,omitempty"` // falta = 1
	Work         float64  `json:"work,omitempty"`    // utilidad de tomar un trabajo
	DistanceCost float64  `json:"distance_cost,omitempty"`
	// Inertia es la ventaja de la acción en curso; nil = DefaultInertia.
	Inertia *float64 `json:"inertia,omitempty"`

	Doing        *Doing     `json:"doing,omitempty"`
	Suspended    *Doing     `json:"suspended,omitempty"`
	RecoverUntil clock.Tick `json:"recover_until,omitempty"`
}

// Weight devuelve el peso de una necesidad.
func (a *Agent) Weight(need string) float64 {
	for _, w := range a.Weights {
		if w.Need == need {
			return w.W
		}
	}
	return 1
}

func (a *Agent) inertia() float64 {
	if a.Inertia == nil {
		return DefaultInertia
	}
	return *a.Inertia
}

// Ctx es lo que recibe una acción.
type Ctx struct {
	W     *world.World
	Actor ids.ID
	Agent *Agent
	Needs needs.Needs // nil si no tiene
}

// Option es una acción posible con su utilidad.
type Option struct {
	Action  string
	Target  ids.ID
	Offer   int
	Utility float64
}

// Result de un paso.
type Result uint8

const (
	Running Result = iota
	Done
	Failed
)

// Action es un tipo de acción. Las implementaciones no guardan estado: todo
// lo que necesitan entre ticks va en Doing (y así se guarda y se hashea).
type Action interface {
	Name() string
	// Options devuelve las opciones posibles (ya filtradas por precondiciones).
	Options(c *Ctx) []Option
	// Utility es la utilidad de seguir con d.
	Utility(c *Ctx, d *Doing) float64
	// Start intenta empezar d (p. ej. reservar). Si falla no modifica nada.
	Start(c *Ctx, d *Doing) bool
	// Step avanza d un tick.
	Step(c *Ctx, d *Doing) Result
	// Stop termina d (liberar reservas, etc.); done indica si terminó bien.
	Stop(c *Ctx, d *Doing, done bool)
	// Valid indica si d puede retomarse tras una suspensión.
	Valid(c *Ctx, d *Doing) bool
	// Interruptible indica si el selector puede cambiar d por otra.
	Interruptible(c *Ctx, d *Doing) bool
}

// actions es el registro de acciones del mundo (en orden de registro).
type actions struct{ list []Action }

func (s *actions) get(name string) Action {
	for _, a := range s.list {
		if a.Name() == name {
			return a
		}
	}
	return nil
}

// Eventos.
type (
	ActionStarted struct {
		Entity ids.ID `json:"entity"`
		Action string `json:"action"`
		Target ids.ID `json:"target"`
	}
	ActionDone struct {
		Entity ids.ID `json:"entity"`
		Action string `json:"action"`
		Target ids.ID `json:"target"`
	}
	ActionFailed struct {
		Entity ids.ID `json:"entity"`
		Action string `json:"action"`
		Target ids.ID `json:"target"`
	}
	ActionInterrupted struct {
		Entity ids.ID `json:"entity"`
		Action string `json:"action"`
		Target ids.ID `json:"target"`
		By     string `json:"by"` // acción que la reemplazó
	}
	ActionSuspended struct {
		Entity ids.ID `json:"entity"`
		Action string `json:"action"`
		Target ids.ID `json:"target"`
		Cause  string `json:"cause"` // kind del evento
	}
	ActionResumed struct {
		Entity ids.ID `json:"entity"`
		Action string `json:"action"`
		Target ids.ID `json:"target"`
	}
)

func (ActionStarted) Kind() string     { return "action_started" }
func (ActionDone) Kind() string        { return "action_done" }
func (ActionFailed) Kind() string      { return "action_failed" }
func (ActionInterrupted) Kind() string { return "action_interrupted" }
func (ActionSuspended) Kind() string   { return "action_suspended" }
func (ActionResumed) Kind() string     { return "action_resumed" }

// interrupting devuelve, para el evento e, la entidad que debe suspender lo
// que hace (o Nil). Por ahora: caer.
func interrupting(e world.Event) ids.ID {
	if f, ok := e.(space.EntityFell); ok {
		return f.Entity
	}
	return ids.Nil
}

type agentIter struct {
	agents *ecs.Store[Agent]
	needs  *ecs.Store[needs.Needs]
}

func (it agentIter) each(w *world.World, fn func(c *Ctx)) {
	it.agents.Each(func(e ids.ID, a *Agent) {
		if w.ECS().Despawning(e) {
			return
		}
		c := &Ctx{W: w, Actor: e, Agent: a}
		if n := it.needs.Ptr(e); n != nil {
			c.Needs = *n
		}
		fn(c)
	})
}

// decide: suspensiones por eventos, reanudaciones y selección.
type decide struct {
	acts *actions
	it   agentIter
}

func (decide) Name() string { return "ai.decide" }

func (s decide) Run(w *world.World) {
	hit := map[ids.ID]string{}
	for _, e := range w.Events() { // eventos del tick anterior
		if id := interrupting(e.Event); !id.IsNil() {
			if _, seen := hit[id]; !seen {
				hit[id] = e.Event.Kind()
			}
		}
	}
	s.it.each(w, func(c *Ctx) {
		a := c.Agent
		if cause, ok := hit[c.Actor]; ok {
			a.RecoverUntil = w.Tick() + RecoverTicks
			if a.Doing != nil && a.Suspended == nil {
				d := a.Doing
				nav.Halt(w, c.Actor)
				d.Moving = false
				a.Suspended, a.Doing = d, nil
				w.Emit(ActionSuspended{Entity: c.Actor, Action: d.Action, Target: d.Target, Cause: cause})
			}
		}
		if w.Tick() < a.RecoverUntil {
			return
		}
		if a.Doing == nil && a.Suspended != nil {
			d := a.Suspended
			a.Suspended = nil
			if act := s.acts.get(d.Action); act != nil && act.Valid(c, d) {
				a.Doing = d
				w.Emit(ActionResumed{Entity: c.Actor, Action: d.Action, Target: d.Target})
				return
			}
			if act := s.acts.get(d.Action); act != nil {
				act.Stop(c, d, false)
			}
		}
		if !clock.Staggered(w.Tick(), DecidePeriod, uint64(c.Actor.Index())) {
			return
		}
		best, ok := s.best(c)
		if !ok || best.Utility < MinUtility {
			return
		}
		if a.Doing != nil {
			cur := s.acts.get(a.Doing.Action)
			same := best.Action == a.Doing.Action && best.Target == a.Doing.Target && best.Offer == a.Doing.Offer
			if same || !cur.Interruptible(c, a.Doing) || best.Utility <= cur.Utility(c, a.Doing)+a.inertia() {
				return
			}
			d := *a.Doing
			if !s.start(c, best) {
				return
			}
			cur.Stop(c, &d, false)
			w.Emit(ActionInterrupted{Entity: c.Actor, Action: d.Action, Target: d.Target, By: best.Action})
			return
		}
		s.start(c, best)
	})
}

// best devuelve la mejor opción. Desempate: acción, objetivo, oferta.
func (s decide) best(c *Ctx) (Option, bool) {
	var best Option
	found := false
	for _, act := range s.acts.list {
		for _, o := range act.Options(c) {
			o.Action = act.Name()
			if !found || o.Utility > best.Utility || o.Utility == best.Utility && cmp.Or(
				cmp.Compare(o.Action, best.Action), cmp.Compare(o.Target, best.Target), cmp.Compare(o.Offer, best.Offer)) < 0 {
				best, found = o, true
			}
		}
	}
	return best, found
}

func (s decide) start(c *Ctx, o Option) bool {
	d := &Doing{Action: o.Action, Target: o.Target, Offer: o.Offer, Since: c.W.Tick()}
	if !s.acts.get(o.Action).Start(c, d) {
		return false
	}
	c.Agent.Doing = d
	c.W.Emit(ActionStarted{Entity: c.Actor, Action: d.Action, Target: d.Target})
	return true
}

// act avanza las acciones en curso.
type act struct {
	acts *actions
	it   agentIter
}

func (act) Name() string { return "ai.act" }

func (s act) Run(w *world.World) {
	s.it.each(w, func(c *Ctx) {
		a := c.Agent
		if a.Doing == nil || w.Tick() < a.RecoverUntil {
			return
		}
		d := a.Doing
		x := s.acts.get(d.Action)
		switch x.Step(c, d) {
		case Done:
			x.Stop(c, d, true)
			a.Doing = nil
			w.Emit(ActionDone{Entity: c.Actor, Action: d.Action, Target: d.Target})
		case Failed:
			x.Stop(c, d, false)
			a.Doing = nil
			w.Emit(ActionFailed{Entity: c.Actor, Action: d.Action, Target: d.Target})
		}
	})
}

func check(w *world.World) error {
	var err error
	ecs.Get[Agent](w.ECS()).Each(func(e ids.ID, a *Agent) {
		for _, d := range []*Doing{a.Doing, a.Suspended} {
			if d != nil && d.Action == "" {
				err = errors.Join(err, fmt.Errorf("ai %v: acción sin nombre", e))
			}
		}
	})
	return err
}

// Install registra stores, eventos, sistemas e invariantes, con las
// acciones del núcleo ("use", "work") y las extra que se pasen.
func Install(w *world.World, extra ...Action) error {
	r := w.ECS()
	agents, err := ecs.Register[Agent](r, "agent")
	if err != nil {
		return err
	}
	if _, err := ecs.Register[Offers](r, "offers"); err != nil {
		return err
	}
	acts := &actions{list: append([]Action{use{}, work{}}, extra...)}
	it := agentIter{agents: agents, needs: ecs.Get[needs.Needs](r)}
	w.AddInvariant("ai", check)
	return errors.Join(
		world.RegisterEvent[ActionStarted](w),
		world.RegisterEvent[ActionDone](w),
		world.RegisterEvent[ActionFailed](w),
		world.RegisterEvent[ActionInterrupted](w),
		world.RegisterEvent[ActionSuspended](w),
		world.RegisterEvent[ActionResumed](w),
		w.Register(world.SystemSpec{Phase: clock.PhaseDecide}, decide{acts, it}),
		// Antes de nav.plan, para que los pedidos de movimiento se planeen
		// en el mismo tick.
		w.Register(world.SystemSpec{Phase: clock.PhaseAct, Order: -100}, act{acts, it}),
	)
}
