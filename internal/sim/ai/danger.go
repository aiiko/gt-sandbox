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

// Cuidado de la giantess al apoyar la mano o al sentarse.
const (
	// HoverExtra es cuánto más sostiene el aplastador en el aire, por unidad
	// de Care, antes de apoyarlo.
	HoverExtra clock.Tick = 40
	// CareWaitMax es lo máximo que espera a que se aparten los tinies que vio
	// debajo; después apoya igual.
	CareWaitMax clock.Tick = 100
	// SeatCareScale reduce la atención al sentarse (fuera de la línea de vista
	// directa): hace falta un Care más alto para notar tinies en el asiento.
	SeatCareScale = 0.45
)

// Hesitated se emite cuando la giantess ve tinies debajo de su mano o en el
// asiento y espera a que se aparten.
type Hesitated struct {
	Entity  ids.ID `json:"entity"`
	Surface ids.ID `json:"surface"`
	Source  string `json:"source,omitempty"`
	Seen    int    `json:"seen"`
}

func (Hesitated) Kind() string { return "hesitated" }

// careful decide si un paso pose de un agente con Care debe esperar. Aplica
// cuando su mano o su asiento (SourceSeat montado sobre el actor) están
// suspendidos sobre una superficie (crush.Threats): ese pose es el que apoya.
// d.Progress cuenta los ticks de espera y d.Wary indica que vio tinies debajo.
func careful(c *Ctx, d *Doing) bool {
	a := c.Agent
	if a.Care <= 0 {
		return false
	}
	g := space.Of(c.W.ECS())
	var hover *crush.Threat
	for _, t := range crush.Threats(c.W) {
		if !a.Hand.IsNil() && t.By == a.Hand {
			tt := t
			hover = &tt
			break
		}
		if hover == nil && t.Source == crush.SourceSeat && g.Parent(t.By) == c.Actor {
			tt := t
			hover = &tt
		}
	}
	if hover == nil {
		return false
	}
	care := min(a.Care, 1)
	if hover.Source == crush.SourceSeat {
		care = care * care * SeatCareScale
	}
	d.Progress++
	seen := visibleUnder(c, *hover)
	if d.Progress == 1 && len(seen) > 0 {
		// tirada de atención: cada tiny visible es una chance de notarlos
		p := 1 - math.Pow(1-care, float64(len(seen)))
		if c.W.RNG("ai.attention").Float64() < p {
			d.Wary = true
			c.W.Emit(Hesitated{Entity: c.Actor, Surface: hover.Surface, Source: hover.Source, Seen: len(seen)})
		}
	}
	if d.Progress <= clock.Tick(care*float64(HoverExtra)) {
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
