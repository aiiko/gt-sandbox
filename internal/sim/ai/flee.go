package ai

import (
	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Huida. Un agente con Reflex > 0 que queda bajo una amenaza (crush.Threats)
// en su propio Space se alarma; Reflex ticks después deja lo que hace y
// huye al punto libre más cercano, o a un refugio si queda más cerca. Si no
// hay salida, queda atrapado. Esto corre fuera del período de decisión y
// por encima de cualquier compromiso.
const (
	FleeAction = "flee"
	// FleeRange es hasta dónde busca una salida.
	FleeRange = 30 * units.Centimeter
	// fleeStep es el paso de la búsqueda, en cada una de fleeDirs direcciones.
	fleeStep = 2 * units.Millimeter
	fleeDirs = 16
	// FleeMargin es cuánto se aleja de la huella, además de su radio.
	FleeMargin = units.Millimeter
)

// Eventos de la huida.
type (
	// Alarmed: el agente percibió una amenaza (sombra o vibración).
	Alarmed struct {
		Entity  ids.ID `json:"entity"`
		By      ids.ID `json:"by"`
		Source  string `json:"source"`
		Channel string `json:"channel"`
	}
	// Trapped: no hay salida a su alcance.
	Trapped struct {
		Entity ids.ID `json:"entity"`
		By     ids.ID `json:"by"`
	}
)

func (Alarmed) Kind() string { return "alarmed" }
func (Trapped) Kind() string { return "trapped" }

// threatOn devuelve la primera amenaza que cubre al actor en su Space.
func threatOn(c *Ctx, threats []crush.Threat) (crush.Threat, bool) {
	r := c.W.ECS()
	l, ok := comp.Of(r).Location.Get(c.Actor)
	if !ok {
		return crush.Threat{}, false
	}
	p, _, err := space.Of(r).WorldPose(c.Actor)
	if err != nil {
		return crush.Threat{}, false
	}
	for _, t := range threats {
		if t.Surface == l.Space && t.Covers(p, 0) {
			return t, true
		}
	}
	return crush.Threat{}, false
}

// react maneja a un agente amenazado. Devuelve true si el selector no debe
// seguir con él este tick.
func (s decide) react(c *Ctx, t crush.Threat, threats []crush.Threat) bool {
	a, w := c.Agent, c.W
	if a.Doing != nil && a.Doing.Action == FleeAction {
		return true
	}
	if a.Alarm == 0 {
		a.Alarm = w.Tick() + a.Reflex
		w.Emit(Alarmed{Entity: c.Actor, By: t.By, Source: t.Source, Channel: t.Channel})
		return true
	}
	if w.Tick() < a.Alarm || a.Trapped || w.Tick() < a.RecoverUntil {
		return true
	}
	sp, goal, ok := escape(c, t, threats)
	if !ok {
		a.Trapped = true
		s.drop(c, "trapped")
		nav.Halt(w, c.Actor)
		w.Emit(Trapped{Entity: c.Actor, By: t.By})
		return true
	}
	s.drop(c, FleeAction)
	a.Doing = &Doing{Action: FleeAction, Place: sp, Goal: goal, Since: w.Tick()}
	w.Emit(ActionStarted{Entity: c.Actor, Action: FleeAction, Target: ids.Nil})
	return true
}

// drop interrumpe la acción en curso (si hay) a favor de by.
func (s decide) drop(c *Ctx, by string) {
	a := c.Agent
	if a.Doing == nil {
		return
	}
	d := *a.Doing
	cur := s.acts.get(d.Action)
	what := s.acts.what(c, &d)
	cur.Stop(c, &d, false)
	a.Doing = nil
	c.W.Emit(ActionInterrupted{Entity: c.Actor, Action: d.Action, What: what, Target: d.Target, By: by})
}

// escape busca la salida más cercana: un punto del Space del actor,
// transitable y fuera de todas las amenazas de ese Space (con margen), o un
// refugio que resista la amenaza t si está más cerca. Prueba fleeDirs
// direcciones en pasos de fleeStep hasta FleeRange (orden determinista).
func escape(c *Ctx, t crush.Threat, threats []crush.Threat) (ids.ID, units.Vec2, bool) {
	r := c.W.ECS()
	g := space.Of(r)
	cs := comp.Of(r)
	l, ok := cs.Location.Get(c.Actor)
	if !ok {
		return ids.Nil, units.Vec2{}, false
	}
	sp, ok := ecs.Get[space.Space](r).Get(l.Space)
	if !ok {
		return ids.Nil, units.Vec2{}, false
	}
	margin := FleeMargin
	if b, ok := cs.Body.Get(c.Actor); ok {
		margin += b.Radius
	}
	grid := ecs.Get[nav.Grid](r).Ptr(l.Space)
	from := l.Pos.XY()
	reach := func(units.Vec2) bool { return true }
	if grid != nil {
		reach = grid.Region(from, FleeRange)
	}
	free := func(p units.Vec2) bool {
		if !sp.Extent.Contains(p) || !reach(p) {
			return false
		}
		wp, err := g.ToWorld(l.Space, units.V3(p.X, p.Y, 0))
		if err != nil {
			return false
		}
		for _, x := range threats {
			if x.Surface == l.Space && x.Covers(wp, margin) {
				return false
			}
		}
		return true
	}

	best, found := units.Vec2{}, false
	var bestD int64
	for k := units.Length(1); k*fleeStep <= FleeRange && !found; k++ {
		for i := range fleeDirs {
			off := fixedmath.Rotate(units.V2(k*fleeStep, 0), fixedmath.Turn/fleeDirs*fixedmath.Angle(i))
			if p := from.Add(off); free(p) {
				best, found, bestD = p, true, from.DistSq(p)
				break
			}
		}
	}
	ecs.Get[crush.Shelter](r).Each(func(e ids.ID, sh *crush.Shelter) {
		sl, ok := cs.Location.Get(e)
		if !ok || sl.Space != l.Space || sh.Strength < t.Force || r.Despawning(e) {
			return
		}
		if d := from.DistSq(sl.Pos.XY()); d <= int64(FleeRange)*int64(FleeRange) && reach(sl.Pos.XY()) && (!found || d < bestD) {
			best, found, bestD = sl.Pos.XY(), true, d
		}
	})
	return l.Space, best, found
}

// flee camina a Doing.Goal en el Space Doing.Place. Solo la empieza react.
type flee struct{}

func (flee) Name() string                    { return FleeAction }
func (flee) Options(*Ctx) []Option           { return nil }
func (flee) Utility(*Ctx, *Doing) float64    { return UrgentUtility * 10 }
func (flee) Start(*Ctx, *Doing) bool         { return true }
func (flee) Valid(*Ctx, *Doing) bool         { return false }
func (flee) Interruptible(*Ctx, *Doing) bool { return false }
func (flee) Committed(*Ctx, *Doing) bool     { return true }
func (flee) Step(c *Ctx, d *Doing) Result    { return approachResult(c, d) }

// Stop: si la huida no terminó bien, no vuelve a intentarla mientras siga
// bajo la misma amenaza (react la reinicia al salir de ella).
func (flee) Stop(c *Ctx, d *Doing, done bool) {
	haltIfMoving(c, d)
	if !done {
		c.Agent.Trapped = true
	}
}

func approachResult(c *Ctx, d *Doing) Result {
	switch nav.Approach(c.W, c.Actor, d.Place, d.Goal, &d.Moving) {
	case nav.Arrived:
		return Done
	case nav.Failed:
		return Failed
	}
	return Running
}

func haltIfMoving(c *Ctx, d *Doing) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
	}
}

// underThreat indica si el objetivo de una opción (o el de su trabajo) está
// bajo una amenaza: un agente que huye no vuelve a meterse ahí.
func underThreat(c *Ctx, target ids.ID) bool {
	if len(c.Threats) == 0 {
		return false
	}
	r := c.W.ECS()
	if j, ok := ecs.Get[jobs.Job](r).Get(target); ok {
		target = j.Target
	}
	l, ok := comp.Of(r).Location.Get(target)
	if !ok {
		return false
	}
	p, _, err := space.Of(r).WorldPose(target)
	if err != nil {
		return false
	}
	for _, t := range c.Threats {
		if t.Surface == l.Space && t.Covers(p, crush.NearMissRange) {
			return true
		}
	}
	return false
}

// registerFlee registra los eventos de la huida.
func registerFlee(w *world.World) error {
	if err := world.RegisterEvent[Alarmed](w); err != nil {
		return err
	}
	return world.RegisterEvent[Trapped](w)
}
