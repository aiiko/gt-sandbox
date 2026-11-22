package ai

import (
	"math"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Danger es el componente "danger" de una zona que los agentes de Scale
// evitan hasta Until: ahí murió alguien. La crea quien sabe de muertes (la
// colonia) y la borra al vencer.
type Danger struct {
	Scale  string       `json:"scale"`
	Space  ids.ID       `json:"space"`
	Center units.Vec2   `json:"center"`
	Radius units.Length `json:"radius"`
	Until  clock.Tick   `json:"until"`
}

// InDanger indica si el punto p del Space sp está en una zona de peligro
// vigente para agentes de la escala scale.
func InDanger(w *world.World, scale string, sp ids.ID, p units.Vec2) bool {
	found := false
	ecs.Get[Danger](w.ECS()).Each(func(e ids.ID, d *Danger) {
		found = found || d.Scale == scale && d.Space == sp && d.Until > w.Tick() && !w.ECS().Despawning(e) &&
			p.DistSq(d.Center) <= int64(d.Radius)*int64(d.Radius)
	})
	return found
}

// endangered indica si el objetivo de una opción (o el de su trabajo) está
// en una zona de peligro para el actor.
func endangered(c *Ctx, target ids.ID) bool {
	r := c.W.ECS()
	if ecs.Get[Danger](r).Len() == 0 {
		return false
	}
	if j, ok := ecs.Get[jobs.Job](r).Get(target); ok {
		target = j.Target
	}
	l, ok := comp.Of(r).Location.Get(target)
	return ok && InDanger(c.W, c.Agent.Scale, l.Space, l.Pos.XY())
}

// Cuidado de la giantess al apoyar la mano.
const (
	// HoverExtra es cuánto más sostiene la mano en el aire, por unidad de
	// Care, antes de apoyarla.
	HoverExtra clock.Tick = 40
	// CareWaitMax es lo máximo que espera a que se aparten los tinies que vio
	// debajo; después apoya igual.
	CareWaitMax clock.Tick = 100
)

// Hesitated se emite cuando la giantess ve tinies debajo de su mano y espera
// a que se aparten.
type Hesitated struct {
	Entity  ids.ID `json:"entity"`
	Surface ids.ID `json:"surface"`
	Seen    int    `json:"seen"`
}

func (Hesitated) Kind() string { return "hesitated" }

// careful decide si un paso pose de un agente con Care debe esperar. Solo
// aplica si su mano está suspendida sobre una superficie (crush.Threats):
// ese pose es el que la apoya. d.Progress cuenta los ticks de espera y
// d.Wary indica que vio tinies debajo.
func careful(c *Ctx, d *Doing) bool {
	a := c.Agent
	if a.Care <= 0 || a.Hand.IsNil() {
		return false
	}
	var hover *crush.Threat
	for _, t := range crush.Threats(c.W) {
		if t.By == a.Hand {
			hover = &t
			break
		}
	}
	if hover == nil {
		return false
	}
	d.Progress++
	seen := visibleUnder(c, *hover)
	if d.Progress == 1 && len(seen) > 0 {
		// tirada de atención: cada tiny visible es una chance de notarlos
		p := 1 - math.Pow(1-min(a.Care, 1), float64(len(seen)))
		if c.W.RNG("ai.attention").Float64() < p {
			d.Wary = true
			c.W.Emit(Hesitated{Entity: c.Actor, Surface: hover.Surface, Seen: len(seen)})
		}
	}
	if d.Progress <= clock.Tick(a.Care*float64(HoverExtra)) {
		return true
	}
	return d.Wary && d.Progress <= CareWaitMax && len(seen) > 0
}

// visibleUnder son los frágiles bajo la huella (con margen de susto) que el
// agente alcanza a ver (a menos de Watch en el mundo).
func visibleUnder(c *Ctx, t crush.Threat) []ids.ID {
	if c.Agent.Watch <= 0 {
		return nil
	}
	r := c.W.ECS()
	g := space.Of(r)
	me, _, err := g.WorldPose(c.Actor)
	if err != nil {
		return nil
	}
	r2 := int64(c.Agent.Watch) * int64(c.Agent.Watch)
	var out []ids.ID
	for _, e := range t.Under(c.W, crush.NearMissRange) {
		if p, _, err := g.WorldPose(e); err == nil && p.XY().DistSq(me.XY()) <= r2 {
			out = append(out, e)
		}
	}
	return out
}
