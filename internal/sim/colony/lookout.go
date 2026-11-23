package colony

import (
	"errors"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Vigía y avisos (propuesta, 18.2 y 18.3).
//
// Un puesto de vigía (lookout_post) es un marcador reservable cerca de un
// portal conocido. Un tiny (con preferencia para el oficio "lookout") vigila
// el tramo durante un turno breve; si percibe una sombra o una próxima pisada
// que alcanza el tramo, emite un aviso (lookout_warned) una sola vez por
// amenaza. Los miembros cercanos que se dirigen al tramo esperan o buscan otro
// camino (penalización en nav.Router); quienes ya están bajo peligro inmediato
// huyen con el sistema existente.

const (
	WatchAction                    = "watch"
	DefaultShiftTicks   clock.Tick = 12 * clock.TPS
	DefaultMinShift     clock.Tick = 2 * clock.TPS
	DefaultWarnTicks    clock.Tick = 4 * clock.TPS
	DefaultLookoutPause clock.Tick = 8 * clock.TPS
	// DangerPortalPenalty es el costo extra que suma un portal bajo aviso o
	// zona de peligro para que nav.Router elija un desvío (p. ej. pasacables).
	DangerPortalPenalty = 20 * units.Meter
)

// LookoutPost es el componente "lookout_post" de un puesto de vigía.
type LookoutPost struct {
	Colony      ids.ID       `json:"colony"`
	Portal      ids.ID       `json:"portal"`
	WatchPoint  units.Vec2   `json:"watch_point"`
	WatchRadius units.Length `json:"watch_radius"`
	AlertRadius units.Length `json:"alert_radius"`
	ShiftTicks  clock.Tick   `json:"shift_ticks,omitempty"`
	WarnTicks   clock.Tick   `json:"warn_ticks,omitempty"`
	Cooldown    clock.Tick   `json:"cooldown,omitempty"`
	Next        clock.Tick   `json:"next,omitempty"`
	LastThreat  ids.ID       `json:"last_threat,omitempty"`
	WarnUntil   clock.Tick   `json:"warn_until,omitempty"`
	WarnSource  string       `json:"warn_source,omitempty"`
}

func (p LookoutPost) shift() clock.Tick    { return cmpOr(p.ShiftTicks, DefaultShiftTicks) }
func (p LookoutPost) warn() clock.Tick     { return cmpOr(p.WarnTicks, DefaultWarnTicks) }
func (p LookoutPost) cooldown() clock.Tick { return cmpOr(p.Cooldown, DefaultLookoutPause) }

// LookoutWarned se emite cuando un vigía percibe una amenaza sobre el tramo
// vigilado y avisa a los miembros cercanos.
type LookoutWarned struct {
	Post   ids.ID       `json:"post"`
	By     ids.ID       `json:"by"`
	Space  ids.ID       `json:"space"`
	Center units.Vec2   `json:"center"`
	Radius units.Length `json:"radius"`
	Until  clock.Tick   `json:"until"`
	Source string       `json:"source"`
}

func (LookoutWarned) Kind() string { return "lookout_warned" }

// Watch es la acción "watch": ir a un puesto de vigía conocido y observar el
// tramo vigilado. Doing: Target es el puesto; Stage 0 camina, 1 vigila.
type Watch struct{}

func (Watch) Name() string { return WatchAction }

// routeActive indica si algún miembro de la colonia (distinto de exclude) está
// usando o acercándose al portal vigilado, o hay una expedición en marcha.
func routeActive(w *world.World, p LookoutPost, postSpace ids.ID, exclude ids.ID) bool {
	r := w.ECS()
	locs := comp.Of(r).Location
	movers := ecs.Get[nav.Mover](r)
	active := false
	ecs.Get[Member](r).Each(func(e ids.ID, m *Member) {
		if active || e == exclude || m.Colony != p.Colony || r.Despawning(e) {
			return
		}
		if mv, ok := movers.Get(e); ok && mv.Status == nav.Moving {
			for _, leg := range mv.Legs {
				if !p.Portal.IsNil() && leg.Portal == p.Portal {
					active = true
					return
				}
			}
		}
		if l, ok := locs.Get(e); ok && l.Space == postSpace &&
			l.Pos.XY().DistSq(p.WatchPoint) <= int64(p.AlertRadius)*int64(p.AlertRadius) {
			active = true
		}
	})
	return active
}

func (Watch) Options(c *ai.Ctx) []ai.Option {
	r := c.W.ECS()
	m, ok := ecs.Get[Member](r).Get(c.Actor)
	if !ok {
		return nil
	}
	locs := comp.Of(r).Location
	critical := CriticalFood(c.W, m.Colony)
	var out []ai.Option
	ecs.Get[LookoutPost](r).Each(func(e ids.ID, p *LookoutPost) {
		if p.Colony != m.Colony || r.Despawning(e) || !c.Agent.Knows(e) || c.W.Tick() < p.Next {
			return
		}
		if h := jobs.Holder(c.W, e); !h.IsNil() && h != c.Actor {
			return
		}
		l, ok := locs.Get(e)
		if !ok {
			return
		}
		busy := routeActive(c.W, *p, l.Space, c.Actor)
		if !busy && m.Role != RoleLookout {
			return
		}
		if critical && m.Role != RoleLookout {
			return
		}
		u := 0.16
		if !busy {
			u = 0.06
		}
		if m.Role == RoleLookout {
			u += RoleBonus
		} else if HasRoles(c.W, m.Colony) && specialistAvailable(c.W, m.Colony, RoleLookout, c.Actor) {
			return
		}
		u -= c.Agent.DistanceCost * ai.Distance(c, l.Space, l.Pos.XY())
		out = append(out, ai.Option{Target: e, Utility: u})
	})
	return out
}

func (Watch) Utility(c *ai.Ctx, d *ai.Doing) float64 {
	r := c.W.ECS()
	m, ok := ecs.Get[Member](r).Get(c.Actor)
	p, ok2 := ecs.Get[LookoutPost](r).Get(d.Target)
	l, ok3 := comp.Of(r).Location.Get(d.Target)
	if !ok || !ok2 || !ok3 {
		return 0
	}
	u := 0.06
	if routeActive(c.W, p, l.Space, c.Actor) {
		u = 0.16
	}
	if m.Role == RoleLookout {
		u += RoleBonus
	}
	return u
}

func (Watch) Start(c *ai.Ctx, d *ai.Doing) bool {
	return ecs.Get[LookoutPost](c.W.ECS()).Has(d.Target) && jobs.Reserve(c.W, d.Target, c.Actor)
}

func (Watch) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	w := c.W
	r := w.ECS()
	p := ecs.Get[LookoutPost](r).Ptr(d.Target)
	l, ok := comp.Of(r).Location.Get(d.Target)
	if p == nil || !ok || !jobs.Reserve(w, d.Target, c.Actor) {
		return ai.Failed
	}
	if d.Stage == 0 {
		switch nav.Approach(w, c.Actor, l.Space, l.Pos.XY(), &d.Moving) {
		case nav.Failed:
			return ai.Failed
		case nav.Underway:
			return ai.Running
		}
		d.Stage, d.Moving, d.Progress = 1, false, 0
	}
	// Observar amenazas sobre el tramo vigilado.
	if t, hit := threatOnZone(w, l.Space, p.WatchPoint, p.WatchRadius); hit {
		if w.Tick() >= p.WarnUntil || p.LastThreat != t.By {
			p.WarnUntil = w.Tick() + p.warn()
			p.LastThreat = t.By
			p.WarnSource = t.Source
			w.Emit(LookoutWarned{
				Post: d.Target, By: c.Actor, Space: l.Space,
				Center: p.WatchPoint, Radius: p.WatchRadius,
				Until: p.WarnUntil, Source: t.Source,
			})
		}
	}
	d.Progress++
	if d.Progress >= p.shift() || (d.Progress >= DefaultMinShift && !routeActive(w, *p, l.Space, c.Actor)) {
		p.Next = w.Tick() + p.cooldown()
		return ai.Done
	}
	return ai.Running
}

func (Watch) Stop(c *ai.Ctx, d *ai.Doing, _ bool) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
	}
	jobs.Release(c.W, d.Target, c.Actor)
}

func (Watch) Valid(c *ai.Ctx, d *ai.Doing) bool {
	return ecs.Get[LookoutPost](c.W.ECS()).Has(d.Target) && jobs.Reserve(c.W, d.Target, c.Actor)
}

func (Watch) Interruptible(*ai.Ctx, *ai.Doing) bool { return true }

// threatOnZone devuelve la primera amenaza sobre el Space sp que toca el
// círculo (center, radius).
func threatOnZone(w *world.World, sp ids.ID, center units.Vec2, radius units.Length) (crush.Threat, bool) {
	wp, err := space.Of(w.ECS()).ToWorld(sp, units.V3(center.X, center.Y, 0))
	if err != nil {
		return crush.Threat{}, false
	}
	for _, t := range crush.Threats(w) {
		if t.Surface == sp && t.Covers(wp, radius) {
			return t, true
		}
	}
	return crush.Threat{}, false
}

// lookoutSystem vence avisos cuya amenaza terminó y detiene o replanea a los
// miembros cercanos que se dirigen a un tramo avisado.
type lookoutSystem struct{}

func (lookoutSystem) Name() string { return "colony.lookout" }

func (lookoutSystem) Run(w *world.World) {
	r := w.ECS()
	posts := ecs.Get[LookoutPost](r)
	if posts.Len() == 0 {
		return
	}
	locs := comp.Of(r).Location
	movers := ecs.Get[nav.Mover](r)
	agents := ecs.Get[ai.Agent](r)
	members := ecs.Get[Member](r)

	posts.Each(func(pid ids.ID, p *LookoutPost) {
		if r.Despawning(pid) || w.Tick() >= p.WarnUntil {
			return
		}
		pl, ok := locs.Get(pid)
		if !ok {
			return
		}
		// Si la amenaza que disparó el aviso ya no cubre el tramo, vence antes.
		if _, hit := threatOnZone(w, pl.Space, p.WatchPoint, p.WatchRadius); !hit {
			p.WarnUntil = w.Tick()
			return
		}
		alertSq := int64(p.AlertRadius) * int64(p.AlertRadius)
		watchSq := int64(p.WatchRadius) * int64(p.WatchRadius)
		members.Each(func(e ids.ID, m *Member) {
			if m.Colony != p.Colony || r.Despawning(e) {
				return
			}
			a := agents.Ptr(e)
			if a != nil && a.Doing != nil && (a.Doing.Action == ai.FleeAction || a.Doing.Action == WatchAction) {
				return
			}
			l, ok := locs.Get(e)
			if !ok || l.Space != pl.Space || l.Pos.XY().DistSq(pl.Pos.XY()) > alertSq {
				return
			}
			mv := movers.Ptr(e)
			if mv == nil || mv.Status != nav.Moving || len(mv.Legs) == 0 || mv.Legs[0].Space != pl.Space {
				return
			}
			// Si ya está dentro del tramo amenazado, la huida individual se encarga.
			if l.Pos.XY().DistSq(p.WatchPoint) <= watchSq {
				return
			}
			if headsInto(mv.Legs[0].Path, p.WatchPoint, watchSq) {
				// Si el tramo cruza el portal vigilado, replanear para permitir desvío.
				if !p.Portal.IsNil() && mv.Legs[0].Portal == p.Portal {
					mv.Status, mv.Legs = nav.Planning, nil
					return
				}
				// Si no hay desvío, esperar a resguardo mientras dure el aviso.
				mv.WaitUntil = max(mv.WaitUntil, w.Tick()+2)
			}
		})
	})
}

// headsInto indica si alguno de los próximos puntos del tramo entra en el
// círculo (center, radiusSq).
func headsInto(path []units.Vec2, center units.Vec2, radiusSq int64) bool {
	for _, pt := range path {
		if pt.DistSq(center) <= radiusSq {
			return true
		}
	}
	return false
}

// portalDangerPenalty penaliza portales cuyos extremos caen en una zona de
// peligro (ai.Danger) o bajo un aviso de vigía activo, para que los tinies
// prefieran desvíos seguros (p. ej. el pasacables en lugar de la puerta).
func portalDangerPenalty(w *world.World, actor, portalID ids.ID) units.Length {
	r := w.ECS()
	a, ok := ecs.Get[ai.Agent](r).Get(actor)
	if !ok || a.Scale == "" {
		return 0
	}
	p, ok := ecs.Get[nav.Portal](r).Get(portalID)
	if !ok {
		return 0
	}
	for _, end := range [2]nav.PortalEnd{p.A, p.B} {
		if ai.InDanger(w, a.Scale, end.Space, end.Pos) {
			return DangerPortalPenalty
		}
	}
	// Revisar avisos de vigía activos.
	if m, ok := ecs.Get[Member](r).Get(actor); ok {
		locs := comp.Of(r).Location
		penalized := false
		ecs.Get[LookoutPost](r).Each(func(pid ids.ID, post *LookoutPost) {
			if penalized || post.Colony != m.Colony || w.Tick() >= post.WarnUntil || r.Despawning(pid) {
				return
			}
			if post.Portal == portalID {
				penalized = true
				return
			}
			if pl, ok := locs.Get(pid); ok {
				r2 := int64(post.WatchRadius) * int64(post.WatchRadius)
				for _, end := range [2]nav.PortalEnd{p.A, p.B} {
					if end.Space == pl.Space && end.Pos.DistSq(post.WatchPoint) <= r2 {
						penalized = true
						return
					}
				}
			}
		})
		if penalized {
			return DangerPortalPenalty
		}
	}
	return 0
}

// registerLookout registra el store, evento, sistema y penalización de portales.
func registerLookout(w *world.World) error {
	if err := register[LookoutPost](w.ECS(), "lookout_post"); err != nil {
		return err
	}
	nav.SetPortalPenalty(w, portalDangerPenalty)
	return errors.Join(
		world.RegisterEvent[LookoutWarned](w),
		// Después de ai.act (Order -100) y antes de nav.plan (Order 0).
		w.Register(world.SystemSpec{Phase: clock.PhaseAct, Order: -50}, lookoutSystem{}),
	)
}
