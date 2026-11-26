// Package encounter implementa los encuentros de la giantess con los tinies
// (ver docs/specs/encuentros.md): un registro de dónde vio tinies hace poco
// y qué hacían ("sightings"), enfriamientos por superficie y acción, y las
// acciones que ese registro abre en el selector de ai (observar, compartir
// migajas, ofrecer la palma).
//
// Se activa por contenido: un agente sin el componente "encounter" no
// registra nada ni tiene estas opciones.
package encounter

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

const (
	// SightingTTL es cuánto recuerda un avistamiento sin volver a ver tinies
	// en esa superficie.
	SightingTTL clock.Tick = 2 * 60 * clock.TPS
	// Recent es cuán reciente tiene que ser un avistamiento para abrir una
	// acción de encuentro.
	Recent clock.Tick = 20 * clock.TPS
)

// Encounter es el componente "encounter" de quien tiene encuentros (la
// giantess). Observe, Share y Offer son las utilidades de cada acción (0 =
// no la hace).
type Encounter struct {
	Observe    float64 `json:"observe,omitempty"`
	Share      float64 `json:"share,omitempty"`
	Offer      float64 `json:"offer,omitempty"`
	Carry      float64 `json:"carry,omitempty"`
	NoticeGift float64 `json:"notice_gift,omitempty"`
	// Portion es lo que deja al compartir (una migaja); DropPose, la pose
	// desde la que la deja (vacía = DefaultDropPose).
	Portion  *cmds.Spawn `json:"portion,omitempty"`
	DropPose string      `json:"drop_pose,omitempty"`
	// HoverPose y RestPose son las poses de la palma ofrecida: suspendida y
	// apoyada (vacías = DefaultHoverPose y DefaultRestPose).
	HoverPose string `json:"hover_pose,omitempty"`
	RestPose  string `json:"rest_pose,omitempty"`
	// Sightings y Cooldowns están ordenados por superficie (y acción).
	Sightings []Sighting `json:"sightings,omitempty"`
	Cooldowns []Cooldown `json:"cooldowns,omitempty"`
}

// Sighting es lo último que vio en una superficie: cuándo, cuántos tinies y
// cuántos de ellos buscaban comida (salir a buscar o acarrear).
type Sighting struct {
	Surface  ids.ID     `json:"surface"`
	Last     clock.Tick `json:"last"`
	Seen     int        `json:"seen"`
	Foraging int        `json:"foraging,omitempty"`
}

// Cooldown: no repite Action en Surface hasta Until.
type Cooldown struct {
	Surface ids.ID     `json:"surface"`
	Action  string     `json:"action"`
	Until   clock.Tick `json:"until"`
}

// Sighting devuelve el avistamiento de la superficie s.
func (e *Encounter) Sighting(s ids.ID) (Sighting, bool) {
	i, ok := slices.BinarySearchFunc(e.Sightings, s, func(x Sighting, s ids.ID) int { return cmp.Compare(x.Surface, s) })
	if !ok {
		return Sighting{}, false
	}
	return e.Sightings[i], true
}

// Cooling indica si action está enfriándose en la superficie s.
func (e *Encounter) Cooling(s ids.ID, action string, now clock.Tick) bool {
	for _, c := range e.Cooldowns {
		if c.Surface == s && c.Action == action && c.Until > now {
			return true
		}
	}
	return false
}

// Cool enfría action en la superficie s hasta until.
func (e *Encounter) Cool(s ids.ID, action string, until clock.Tick) {
	for i, c := range e.Cooldowns {
		if c.Surface == s && c.Action == action {
			e.Cooldowns[i].Until = until
			return
		}
	}
	e.Cooldowns = append(e.Cooldowns, Cooldown{Surface: s, Action: action, Until: until})
	slices.SortFunc(e.Cooldowns, compareCooldown)
}

func compareCooldown(a, b Cooldown) int {
	return cmp.Or(cmp.Compare(a.Surface, b.Surface), cmp.Compare(a.Action, b.Action))
}

// note registra que en s vio seen tinies, foraging de ellos buscando comida.
func (e *Encounter) note(s ids.ID, now clock.Tick, seen, foraging int) {
	i, ok := slices.BinarySearchFunc(e.Sightings, s, func(x Sighting, s ids.ID) int { return cmp.Compare(x.Surface, s) })
	if !ok {
		e.Sightings = slices.Insert(e.Sightings, i, Sighting{Surface: s})
	}
	e.Sightings[i].Last, e.Sightings[i].Seen, e.Sightings[i].Foraging = now, seen, foraging
}

// Skipped se emite cuando una acción de encuentro llega y su condición ya no
// se cumple (los tinies se fueron, ya hay migajas, la huella está ocupada).
type Skipped struct {
	Entity  ids.ID `json:"entity"`
	Action  string `json:"action"`
	Surface ids.ID `json:"surface"`
	Reason  string `json:"reason"`
}

func (Skipped) Kind() string { return "encounter_skipped" }

// foraging: buscar comida es salir a buscar, revisar un punto de búsqueda o
// acarrear (un trabajo).
func foraging(d *ai.Doing) bool {
	return d != nil && (d.Action == "forage" || d.Action == "search" || d.Action == "work")
}

// look: cada ai.LookPeriod ticks, quien tiene "encounter" y Watch cuenta los
// agentes de otra escala a menos de Watch por superficie (salvo los que van
// en su mano), olvida los avistamientos viejos y los enfriamientos vencidos.
type look struct{}

func (look) Name() string { return "encounter.look" }

func (look) Run(w *world.World) {
	r := w.ECS()
	agents := ecs.Get[ai.Agent](r)
	locs := comp.Of(r).Location
	g := space.Of(r)
	now := w.Tick()
	ecs.Get[Encounter](r).Each(func(e ids.ID, enc *Encounter) {
		a, ok := agents.Get(e)
		if !ok || a.Watch <= 0 || r.Despawning(e) || !clock.Staggered(now, ai.LookPeriod, uint64(e.Index())) {
			return
		}
		me, _, err := g.WorldPose(e)
		if err != nil {
			return
		}
		type count struct {
			surface        ids.ID
			seen, foraging int
		}
		var counts []count
		r2 := int64(a.Watch) * int64(a.Watch)
		agents.Each(func(o ids.ID, oa *ai.Agent) {
			if o == e || oa.Scale == a.Scale || r.Despawning(o) {
				return
			}
			l, ok := locs.Get(o)
			if !ok || l.Space == a.Hand {
				return
			}
			if p, _, err := g.WorldPose(o); err != nil || p.XY().DistSq(me.XY()) > r2 {
				return
			}
			i, found := slices.BinarySearchFunc(counts, l.Space, func(c count, s ids.ID) int { return cmp.Compare(c.surface, s) })
			if !found {
				counts = slices.Insert(counts, i, count{surface: l.Space})
			}
			counts[i].seen++
			if foraging(oa.Doing) {
				counts[i].foraging++
			}
		})
		for _, c := range counts {
			enc.note(c.surface, now, c.seen, c.foraging)
		}
		enc.Sightings = slices.DeleteFunc(enc.Sightings, func(s Sighting) bool {
			return now-s.Last > SightingTTL || !r.Alive(s.Surface) || r.Despawning(s.Surface)
		})
		enc.Cooldowns = slices.DeleteFunc(enc.Cooldowns, func(c Cooldown) bool { return c.Until <= now })
	})
}

func check(w *world.World) error {
	var err error
	ecs.Get[Encounter](w.ECS()).Each(func(e ids.ID, enc *Encounter) {
		if !slices.IsSortedFunc(enc.Sightings, func(a, b Sighting) int { return cmp.Compare(a.Surface, b.Surface) }) ||
			len(slices.CompactFunc(slices.Clone(enc.Sightings), func(a, b Sighting) bool { return a.Surface == b.Surface })) != len(enc.Sightings) {
			err = errors.Join(err, fmt.Errorf("encounter %s: avistamientos desordenados o repetidos", e))
		}
		if !slices.IsSortedFunc(enc.Cooldowns, compareCooldown) {
			err = errors.Join(err, fmt.Errorf("encounter %s: enfriamientos desordenados", e))
		}
		if !ecs.Get[ai.Agent](w.ECS()).Has(e) {
			err = errors.Join(err, fmt.Errorf("encounter %s: no es agente", e))
		}
	})
	return err
}

// IsAction indica si name es una acción de encuentro.
func IsAction(name string) bool {
	return name == "observe" || name == "share" || name == "offer_palm" || name == CarryAction || name == NoticeGiftAction
}

// Actions son las acciones de encuentro, para ai.Install.
func Actions() []ai.Action {
	return []ai.Action{Observe{}, Share{}, OfferPalm{}, CarryPassenger{}, NoticeGift{}}
}

// Install registra el componente, los eventos y el sistema de avistamientos.
// Las acciones se instalan con ai.Install(w, encounter.Actions()...).
func Install(w *world.World) error {
	if _, err := ecs.Register[Encounter](w.ECS(), "encounter"); err != nil {
		return err
	}
	w.AddInvariant("encounter", check)
	return errors.Join(
		world.RegisterEvent[Skipped](w),
		world.RegisterEvent[Observed](w),
		world.RegisterEvent[Shared](w),
		world.RegisterEvent[PalmOffered](w),
		world.RegisterEvent[PalmWithdrawn](w),
		world.RegisterEvent[CarryStarted](w),
		world.RegisterEvent[PassengerDelivered](w),
		world.RegisterEvent[GiftNoticed](w),
		// Después de ai.look (Order -50) y antes de ai.decide.
		w.Register(world.SystemSpec{Phase: clock.PhaseDecide, Order: -45}, look{}),
	)
}
