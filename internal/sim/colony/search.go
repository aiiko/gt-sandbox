package colony

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/spatial"
	"gtsandbox/internal/sim/world"
)

// Valores por defecto de un punto de búsqueda.
const (
	DefaultLookTicks clock.Tick = 3 * clock.TPS
	DefaultCooldown  clock.Tick = 120 * clock.TPS
)

// SearchSpot es el componente "search_spot": un lugar que vale la pena
// revisar cuando falta comida (un rincón del escritorio, el borde de la
// silla). Lo revisa un tiny a la vez (reserva). LookTicks es cuánto se
// queda mirando; si no encontró nada, el punto no se ofrece hasta
// EmptyUntil (ahora + Cooldown).
type SearchSpot struct {
	LookTicks  clock.Tick `json:"look_ticks,omitempty"`
	Cooldown   clock.Tick `json:"cooldown,omitempty"`
	EmptyUntil clock.Tick `json:"empty_until,omitempty"`
}

func (s SearchSpot) look() clock.Tick { return cmpOr(s.LookTicks, DefaultLookTicks) }

func (s SearchSpot) cooldown() clock.Tick { return cmpOr(s.Cooldown, DefaultCooldown) }

func cmpOr[T comparable](v, def T) T {
	var zero T
	if v == zero {
		return def
	}
	return v
}

// Searched se emite al terminar de revisar un punto: Found dice si vio
// comida.
type Searched struct {
	Entity ids.ID `json:"entity"`
	Spot   ids.ID `json:"spot"`
	Found  bool   `json:"found"`
}

func (Searched) Kind() string { return "searched" }

// SearchAction es el nombre de la acción.
const SearchAction = "search"

// Search es la acción "search": revisar un punto de búsqueda conocido
// cuando la despensa está por debajo de su meta. Vale Member.Search por el
// déficit (colony.Deficit), menos la distancia; con Member.Search por
// debajo de Agent.Work, un acarreo conocido siempre gana. Termina al ver
// comida o al acabar de mirar.
//
// Doing: Target es el punto; Stage 0 camina, 1 mira (Progress).
type Search struct{}

func (Search) Name() string { return SearchAction }

func (Search) Options(c *ai.Ctx) []ai.Option {
	r := c.W.ECS()
	m, ok := ecs.Get[Member](r).Get(c.Actor)
	if !ok || m.Search <= 0 {
		return nil
	}
	def := Deficit(c.W, m.Colony, ai.ResourceFood)
	if def <= 0 {
		return nil
	}
	hasRoles := HasRoles(c.W, m.Colony)
	if hasRoles && PendingFoodPortions(c.W, m.Colony) >= MissingPortions(c.W, m.Colony) {
		return nil // ya hay comida suficiente en acarreos para cubrir el faltante
	}
	locs := comp.Of(r).Location
	var spots []struct {
		id  ids.ID
		loc comp.Location
	}
	ecs.Get[SearchSpot](r).Each(func(e ids.ID, s *SearchSpot) {
		l, ok := locs.Get(e)
		if !ok || r.Despawning(e) || !c.Agent.Knows(e) || c.W.Tick() < s.EmptyUntil {
			return
		}
		if h := jobs.Holder(c.W, e); !h.IsNil() && h != c.Actor {
			return
		}
		if knownFood(c, l) {
			return // ya se sabe que ahí hay comida: la cola de trabajos se encarga
		}
		spots = append(spots, struct {
			id  ids.ID
			loc comp.Location
		}{e, l})
	})
	if len(spots) == 0 {
		return nil
	}
	bonus := 0.0
	if hasRoles {
		switch {
		case m.Role == RoleGatherer:
			bonus = RoleBonus
		case len(spots) == 1 && specialistAvailable(c.W, m.Colony, RoleGatherer, c.Actor):
			return nil // un solo punto libre y hay recolector disponible: dejárselo
		}
	}
	out := make([]ai.Option, 0, len(spots))
	for _, s := range spots {
		u := m.Search*def + bonus - c.Agent.DistanceCost*ai.Distance(c, s.loc.Space, s.loc.Pos.XY())
		out = append(out, ai.Option{Target: s.id, Utility: u})
	}
	return out
}

func (Search) Utility(c *ai.Ctx, _ *ai.Doing) float64 {
	m, _ := ecs.Get[Member](c.W.ECS()).Get(c.Actor)
	u := m.Search * Deficit(c.W, m.Colony, ai.ResourceFood)
	if HasRoles(c.W, m.Colony) && m.Role == RoleGatherer {
		u += RoleBonus
	}
	return u
}

func (Search) Start(c *ai.Ctx, d *ai.Doing) bool {
	return ecs.Get[SearchSpot](c.W.ECS()).Has(d.Target) && jobs.Reserve(c.W, d.Target, c.Actor)
}

func (Search) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	w := c.W
	s := ecs.Get[SearchSpot](w.ECS()).Ptr(d.Target)
	l, ok := comp.Of(w.ECS()).Location.Get(d.Target)
	if s == nil || !ok || !jobs.Reserve(w, d.Target, c.Actor) {
		return ai.Failed
	}
	if d.Stage == 0 {
		switch nav.Approach(w, c.Actor, l.Space, l.Pos.XY(), &d.Moving) {
		case nav.Failed:
			return ai.Failed
		case nav.Underway:
			return ai.Running
		}
		d.Stage, d.Moving = 1, false
	}
	found := seesFood(c)
	if d.Progress++; !found && d.Progress < s.look() {
		return ai.Running
	}
	if !found {
		s.EmptyUntil = w.Tick() + s.cooldown()
	}
	w.Emit(Searched{Entity: c.Actor, Spot: d.Target, Found: found})
	return ai.Done
}

// seesFood: el actor ve comida suelta (en su Space, a menos de su vista).
func seesFood(c *ai.Ctx) bool {
	l, ok := comp.Of(c.W.ECS()).Location.Get(c.Actor)
	return ok && foodNear(c, l, false)
}

// knownFood: el actor sabe que hay comida suelta a menos de su vista del
// lugar l.
func knownFood(c *ai.Ctx, l comp.Location) bool { return foodNear(c, l, true) }

func foodNear(c *ai.Ctx, l comp.Location, known bool) bool {
	r := c.W.ECS()
	if c.Agent.Sight <= 0 {
		return false
	}
	food, carried := ecs.Get[Food](r), ecs.Get[jobs.Carried](r)
	found := false
	spatial.Of(c.W).Near(l.Space, l.Pos.XY(), c.Agent.Sight, func(e ids.ID, _ units.Vec2) {
		found = found || food.Has(e) && !carried.Has(e) && !r.Despawning(e) && (!known || c.Agent.Knows(e))
	})
	return found
}

func (Search) Stop(c *ai.Ctx, d *ai.Doing, _ bool) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
	}
	jobs.Release(c.W, d.Target, c.Actor)
}

func (Search) Valid(c *ai.Ctx, d *ai.Doing) bool {
	return ecs.Get[SearchSpot](c.W.ECS()).Has(d.Target) && jobs.Reserve(c.W, d.Target, c.Actor)
}

func (Search) Interruptible(*ai.Ctx, *ai.Doing) bool { return true }

// registerSearch registra el store y el evento de los puntos de búsqueda.
func registerSearch(w *world.World) error {
	if err := register[SearchSpot](w.ECS(), "search_spot"); err != nil {
		return err
	}
	return world.RegisterEvent[Searched](w)
}
