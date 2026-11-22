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
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/crush"
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
	// UrgentUtility es lo que tiene que valer una opción para interrumpir una
	// acción comprometida (que dejaría cosas a medias).
	UrgentUtility = 0.5
)

// Weight es la importancia de una necesidad para un agente (personalidad).
type Weight struct {
	Need string  `json:"need"`
	W    float64 `json:"w"`
}

// Doing es una acción en curso. Stage y Progress los usa la acción; Moving
// es el estado de nav.Approach. Place, Item y Posed son de "use": el lugar
// elegido para el paso en curso, lo que carga y si cambió de pose.
type Doing struct {
	Action   string     `json:"action"`
	Target   ids.ID     `json:"target"`
	Offer    int        `json:"offer,omitempty"`
	Stage    int        `json:"stage,omitempty"`
	Progress clock.Tick `json:"progress,omitempty"`
	Moving   bool       `json:"moving,omitempty"`
	Place    ids.ID     `json:"place,omitempty"`
	Item     ids.ID     `json:"item,omitempty"`
	Posed    bool       `json:"posed,omitempty"`
	Wary     bool       `json:"wary,omitempty"`  // vio tinies debajo de la mano (pose con Care)
	Goal     units.Vec2 `json:"goal,omitempty"`  // destino de acciones sin objetivo (forage)
	Items    []ids.ID   `json:"items,omitempty"` // lo que dejaron los pasos spawn
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
	// Hand es la parte desde la que deja cosas (un Space montado sobre el
	// actor); IdlePose, la pose a la que vuelve al terminar una oferta.
	Hand     ids.ID `json:"hand,omitempty"`
	IdlePose string `json:"idle_pose,omitempty"`
	// Sight es el radio de visión (0 = conoce todo); Known, lo que conoce.
	Sight units.Length `json:"sight,omitempty"`
	Known []Known      `json:"known,omitempty"`
	// Avoid son objetivos con los que falló hace poco (ver RetryTicks).
	Avoid []Avoid `json:"avoid,omitempty"`
	// Curiosity es la utilidad de explorar un Space nuevo; Visited, los
	// Spaces donde estuvo y cuándo por última vez (solo se registran con
	// Sight). Revisit es cuánto tarda un Space visitado en volver a
	// interesarle (0 = nunca); al volver vale Curiosity*RevisitFactor.
	Curiosity float64    `json:"curiosity,omitempty"`
	Visited   []Visit    `json:"visited,omitempty"`
	Revisit   clock.Tick `json:"revisit,omitempty"`
	// Watch es el radio (en el mundo) en el que nota a otros agentes.
	Watch units.Length `json:"watch,omitempty"`

	// Reflex son los ticks que tarda en reaccionar a una amenaza
	// (crush.Threats); 0 = no huye. Alarm es el tick en que reacciona (0 =
	// sin alarma) y Trapped, que ya no encontró salida.
	Reflex clock.Tick `json:"reflex,omitempty"`
	// Care (0 a 1) es cuánto cuida dónde apoya la mano: la sostiene más en
	// el aire y puede notar a los tinies de debajo y esperar.
	Care    float64    `json:"care,omitempty"`
	Alarm   clock.Tick `json:"alarm,omitempty"`
	Trapped bool       `json:"trapped,omitempty"`

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
	// Threats son las amenazas del tick (solo para agentes con Reflex).
	Threats []crush.Threat
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
		What   string `json:"what,omitempty"` // qué oferta, para "use"
		Target ids.ID `json:"target"`
	}
	ActionDone struct {
		Entity ids.ID `json:"entity"`
		Action string `json:"action"`
		What   string `json:"what,omitempty"` // qué oferta, para "use"
		Target ids.ID `json:"target"`
	}
	ActionFailed struct {
		Entity ids.ID `json:"entity"`
		Action string `json:"action"`
		What   string `json:"what,omitempty"` // qué oferta, para "use"
		Target ids.ID `json:"target"`
	}
	ActionInterrupted struct {
		Entity ids.ID `json:"entity"`
		Action string `json:"action"`
		What   string `json:"what,omitempty"` // qué oferta, para "use"
		Target ids.ID `json:"target"`
		By     string `json:"by"` // qué la reemplazó (oferta o acción)
	}
	ActionSuspended struct {
		Entity ids.ID `json:"entity"`
		Action string `json:"action"`
		What   string `json:"what,omitempty"` // qué oferta, para "use"
		Target ids.ID `json:"target"`
		Cause  string `json:"cause"` // kind del evento
	}
	ActionResumed struct {
		Entity ids.ID `json:"entity"`
		Action string `json:"action"`
		What   string `json:"what,omitempty"` // qué oferta, para "use"
		Target ids.ID `json:"target"`
	}
)

func (ActionStarted) Kind() string     { return "action_started" }
func (ActionDone) Kind() string        { return "action_done" }
func (ActionFailed) Kind() string      { return "action_failed" }
func (ActionInterrupted) Kind() string { return "action_interrupted" }
func (ActionSuspended) Kind() string   { return "action_suspended" }
func (ActionResumed) Kind() string     { return "action_resumed" }

// describer es una acción que sabe nombrar lo que hace (p. ej. "use" dice
// qué oferta).
type describer interface {
	What(c *Ctx, d *Doing) string
}

func (s *actions) what(c *Ctx, d *Doing) string {
	if x, ok := s.get(d.Action).(describer); ok {
		return x.What(c, d)
	}
	return ""
}

// interrupting devuelve, para el evento e, la entidad que debe suspender lo
// que hace (o Nil): caer o salvarse por poco de un aplastamiento.
func interrupting(e world.Event) ids.ID {
	switch f := e.(type) {
	case space.EntityFell:
		return f.Entity
	case crush.NearMiss:
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
	var threats []crush.Threat
	threatened := false
	s.it.each(w, func(c *Ctx) {
		a := c.Agent
		if cause, ok := hit[c.Actor]; ok {
			a.RecoverUntil = w.Tick() + RecoverTicks
			if a.Doing != nil && a.Suspended == nil {
				d := a.Doing
				nav.Halt(w, c.Actor)
				d.Moving = false
				a.Suspended, a.Doing = d, nil
				w.Emit(ActionSuspended{Entity: c.Actor, Action: d.Action, What: s.acts.what(c, d), Target: d.Target, Cause: cause})
			}
		}
		if a.Reflex > 0 {
			if !threatened {
				threats, threatened = crush.Threats(w), true
			}
			c.Threats = threats
			if t, ok := threatOn(c, threats); ok {
				if s.react(c, t, threats) {
					return
				}
			} else {
				a.Alarm, a.Trapped = 0, false
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
				w.Emit(ActionResumed{Entity: c.Actor, Action: d.Action, What: s.acts.what(c, d), Target: d.Target})
				return
			}
			if act := s.acts.get(d.Action); act != nil {
				act.Stop(c, d, false)
			}
		}
		if !clock.Staggered(w.Tick(), DecidePeriod, uint64(c.Actor.Index())) {
			return
		}
		a.prune(w.Tick())
		best, ok := s.best(c)
		if !ok || best.Utility < MinUtility {
			return
		}
		if a.Doing != nil {
			cur := s.acts.get(a.Doing.Action)
			same := best.Action == a.Doing.Action && best.Target == a.Doing.Target && best.Offer == a.Doing.Offer
			if same || !cur.Interruptible(c, a.Doing) || best.Utility <= cur.Utility(c, a.Doing)+a.inertia() ||
				committed(cur, c, a.Doing) && best.Utility < UrgentUtility {
				return
			}
			d := *a.Doing
			what := s.acts.what(c, &d)
			if !s.start(c, best) {
				return
			}
			cur.Stop(c, &d, false)
			by := cmp.Or(s.acts.what(c, a.Doing), best.Action)
			w.Emit(ActionInterrupted{Entity: c.Actor, Action: d.Action, What: what, Target: d.Target, By: by})
			return
		}
		s.start(c, best)
	})
}

// committer es opcional: una acción comprometida solo la interrumpe algo
// urgente.
type committer interface {
	Committed(c *Ctx, d *Doing) bool
}

func committed(act Action, c *Ctx, d *Doing) bool {
	x, ok := act.(committer)
	return ok && x.Committed(c, d)
}

// best devuelve la mejor opción. Desempate: acción, objetivo, oferta.
func (s decide) best(c *Ctx) (Option, bool) {
	var best Option
	found := false
	for _, act := range s.acts.list {
		for _, o := range act.Options(c) {
			o.Action = act.Name()
			if !o.Target.IsNil() && (c.Agent.avoids(o.Target, c.W.Tick()) || underThreat(c, o.Target) || endangered(c, o.Target)) {
				continue
			}
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
	c.W.Emit(ActionStarted{Entity: c.Actor, Action: d.Action, What: s.acts.what(c, d), Target: d.Target})
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
		what := s.acts.what(c, d) // antes del paso: puede consumir el objeto
		switch x.Step(c, d) {
		case Done:
			x.Stop(c, d, true)
			a.Doing = nil
			w.Emit(ActionDone{Entity: c.Actor, Action: d.Action, What: what, Target: d.Target})
		case Failed:
			x.Stop(c, d, false)
			a.Doing = nil
			if !d.Target.IsNil() {
				a.avoid(d.Target, w.Tick())
			}
			w.Emit(ActionFailed{Entity: c.Actor, Action: d.Action, What: what, Target: d.Target})
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
	ecs.Get[Offers](w.ECS()).Each(func(e ids.ID, os *Offers) {
		if e2 := os.Validate(); e2 != nil {
			err = errors.Join(err, fmt.Errorf("ai %v: %w", e, e2))
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
	if err := registerUse(w); err != nil {
		return err
	}
	if err := registerFlee(w); err != nil {
		return err
	}
	if _, err := ecs.Register[Danger](r, "danger"); err != nil {
		return err
	}
	if err := world.RegisterEvent[Hesitated](w); err != nil {
		return err
	}
	acts := &actions{list: append([]Action{use{}, work{}, explore{}, flee{}}, extra...)}
	it := agentIter{agents: agents, needs: ecs.Get[needs.Needs](r)}
	w.AddInvariant("ai", check)
	return errors.Join(
		world.RegisterEvent[ActionStarted](w),
		world.RegisterEvent[ActionDone](w),
		world.RegisterEvent[ActionFailed](w),
		world.RegisterEvent[ActionInterrupted](w),
		world.RegisterEvent[ActionSuspended](w),
		world.RegisterEvent[ActionResumed](w),
		world.RegisterEvent[Noticed](w),
		world.RegisterEvent[Explored](w),
		w.Register(world.SystemSpec{Phase: clock.PhaseDecide, Order: -50}, look{it}),
		w.Register(world.SystemSpec{Phase: clock.PhaseDecide}, decide{acts, it}),
		// Antes de nav.plan, para que los pedidos de movimiento se planeen
		// en el mismo tick.
		w.Register(world.SystemSpec{Phase: clock.PhaseAct, Order: -100}, act{acts, it}),
	)
}
