package ai

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/space"
)

// Effect es cuánto sube una necesidad al usar una oferta completa.
type Effect struct {
	Need   string `json:"need"`
	Amount int32  `json:"amount"`
}

// Offer es una acción que ofrece un objeto (smart object): a qué escala de
// agente, qué efectos tiene, cuánto dura y desde dónde se usa (Approach,
// relativo a la posición del objeto en su Space).
type Offer struct {
	Action   string     `json:"action"`
	Scale    string     `json:"scale"`
	Effects  []Effect   `json:"effects"`
	Duration clock.Tick `json:"duration"`
	Approach units.Vec2 `json:"approach,omitempty"`
}

// Offers es el componente "offers".
type Offers []Offer

// discomfort es la curva de urgencia de una necesidad: 0 satisfecha, 1 vacía,
// cuadrática (las necesidades bajas pesan mucho más).
func discomfort(v int32) float64 {
	x := float64(needs.Max-v) / needs.Max
	return x * x
}

// gain es la utilidad de subir las necesidades según effects, escalados por
// frac (la parte que falta hacer).
func gain(c *Ctx, effects []Effect, frac float64) float64 {
	u := 0.0
	for _, e := range effects {
		v, ok := c.Needs.Get(e.Need)
		if !ok {
			continue
		}
		after := min(needs.Max, v+int32(float64(e.Amount)*frac))
		u += c.Agent.Weight(e.Need) * (discomfort(v) - discomfort(after))
	}
	return u
}

// distance es la distancia de mundo, en metros, entre el actor y un punto.
func distance(c *Ctx, sp ids.ID, p units.Vec2) float64 {
	g := space.Of(c.W.ECS())
	a, _, err1 := g.WorldPose(c.Actor)
	b, err2 := g.ToWorld(sp, units.V3(p.X, p.Y, 0))
	if err1 != nil || err2 != nil {
		return 0
	}
	return float64(a.Sub(b).XY().Len()) / float64(units.Meter)
}

// --- use: usar una oferta de un smart object ---

type use struct{}

func (use) Name() string { return "use" }

func offerOf(c *Ctx, d *Doing) (Offer, comp.Location, bool) {
	r := c.W.ECS()
	os, ok := ecs.Get[Offers](r).Get(d.Target)
	loc, ok2 := comp.Of(r).Location.Get(d.Target)
	if !ok || !ok2 || d.Offer < 0 || d.Offer >= len(os) || r.Despawning(d.Target) {
		return Offer{}, loc, false
	}
	return os[d.Offer], loc, true
}

func spot(loc comp.Location, o Offer) units.Vec2 { return loc.Pos.XY().Add(o.Approach) }

func (use) Options(c *Ctx) []Option {
	if c.Needs == nil {
		return nil
	}
	var out []Option
	locs := comp.Of(c.W.ECS()).Location
	ecs.Get[Offers](c.W.ECS()).Each(func(e ids.ID, os *Offers) {
		loc, ok := locs.Get(e)
		if !ok || c.W.ECS().Despawning(e) || e == c.Actor {
			return
		}
		if h := jobs.Holder(c.W, e); !h.IsNil() && h != c.Actor {
			return
		}
		for i, o := range *os {
			if o.Scale != c.Agent.Scale || o.Duration <= 0 {
				continue
			}
			u := gain(c, o.Effects, 1) - c.Agent.DistanceCost*distance(c, loc.Space, spot(loc, o))
			out = append(out, Option{Target: e, Offer: i, Utility: u})
		}
	})
	return out
}

func (use) Utility(c *Ctx, d *Doing) float64 {
	o, _, ok := offerOf(c, d)
	if !ok {
		return 0
	}
	return gain(c, o.Effects, float64(o.Duration-d.Progress)/float64(o.Duration))
}

func (use) Start(c *Ctx, d *Doing) bool {
	_, _, ok := offerOf(c, d)
	return ok && jobs.Reserve(c.W, d.Target, c.Actor)
}

func (use) Step(c *Ctx, d *Doing) Result {
	o, loc, ok := offerOf(c, d)
	if !ok || !jobs.Reserve(c.W, d.Target, c.Actor) {
		return Failed
	}
	if d.Stage == 0 {
		switch nav.Approach(c.W, c.Actor, loc.Space, spot(loc, o), &d.Moving) {
		case nav.Arrived:
			d.Stage = 1
		case nav.Failed:
			return Failed
		}
		return Running
	}
	// Los efectos se aplican de a poco: al terminar suman exactamente Amount.
	d.Progress++
	if n := ecs.Get[needs.Needs](c.W.ECS()).Ptr(c.Actor); n != nil {
		for _, e := range o.Effects {
			p, dur := int64(d.Progress), int64(o.Duration)
			n.Add(e.Need, int32(int64(e.Amount)*p/dur-int64(e.Amount)*(p-1)/dur))
		}
	}
	if d.Progress >= o.Duration {
		return Done
	}
	return Running
}

func (use) Stop(c *Ctx, d *Doing, _ bool) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
	}
	jobs.Release(c.W, d.Target, c.Actor)
}

func (use) Valid(c *Ctx, d *Doing) bool {
	_, _, ok := offerOf(c, d)
	return ok && jobs.Reserve(c.W, d.Target, c.Actor)
}

func (use) Interruptible(*Ctx, *Doing) bool { return true }

// --- work: tomar y hacer un trabajo de la cola ---

type work struct{}

func (work) Name() string { return "work" }

func (work) Options(c *Ctx) []Option {
	if c.Agent.Work <= 0 || !jobs.AssignedTo(c.W, c.Actor).IsNil() {
		return nil
	}
	var out []Option
	locs := comp.Of(c.W.ECS()).Location
	ecs.Get[jobs.Job](c.W.ECS()).Each(func(id ids.ID, _ *jobs.Job) {
		j, ok := jobs.Open(c.W, id)
		if !ok {
			return
		}
		loc, _ := locs.Get(j.Target)
		u := c.Agent.Work + 0.01*float64(j.Priority) - c.Agent.DistanceCost*distance(c, loc.Space, loc.Pos.XY())
		out = append(out, Option{Target: id, Utility: u})
	})
	return out
}

func (work) Utility(c *Ctx, _ *Doing) float64 { return c.Agent.Work }

func (work) Start(c *Ctx, d *Doing) bool { return jobs.Assign(c.W, d.Target, c.Actor) }

func (work) Step(c *Ctx, d *Doing) Result {
	switch jobs.Step(c.W, d.Target, c.Actor, &d.Moving) {
	case jobs.Done:
		return Done
	case jobs.Failed:
		return Failed
	}
	return Running
}

func (work) Stop(c *Ctx, d *Doing, done bool) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
	}
	if !done {
		jobs.Unassign(c.W, d.Target, "stopped")
	}
}

func (work) Valid(c *Ctx, d *Doing) bool {
	j, ok := ecs.Get[jobs.Job](c.W.ECS()).Get(d.Target)
	return ok && j.Assignee == c.Actor && !c.W.ECS().Despawning(d.Target)
}

func (work) Interruptible(*Ctx, *Doing) bool { return true }
