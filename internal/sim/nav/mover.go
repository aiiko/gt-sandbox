package nav

import (
	"errors"
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Status es el estado de un Mover.
type Status uint8

const (
	Idle     Status = iota
	Planning        // espera ruta (nav.plan)
	Moving          // recorriendo Legs (nav.move)
)

var statusNames = [...]string{"idle", "planning", "moving"}

func (s Status) String() string {
	if int(s) < len(statusNames) {
		return statusNames[s]
	}
	return fmt.Sprintf("Status(%d)", s)
}

func (s Status) MarshalText() ([]byte, error) {
	if int(s) >= len(statusNames) {
		return nil, fmt.Errorf("nav: status inválido %d", s)
	}
	return []byte(statusNames[s]), nil
}

func (s *Status) UnmarshalText(b []byte) error {
	for i, n := range statusNames {
		if n == string(b) {
			*s = Status(i)
			return nil
		}
	}
	return fmt.Errorf("nav: status desconocido %q", b)
}

// Mover es el componente de las entidades que caminan.
//
// GoalSpace/Goal es el destino (coordenadas locales de GoalSpace). Legs es
// la ruta planificada: el primer tramo está en el Space actual de la entidad.
type Mover struct {
	Speed     units.Length `json:"speed"` // µm por segundo
	Status    Status       `json:"status"`
	GoalSpace ids.ID       `json:"goal_space"`
	Goal      units.Vec2   `json:"goal"`
	Legs      []Leg        `json:"legs,omitempty"`
}

// idle devuelve el mover detenido, conservando velocidad y último destino.
func (m Mover) idle() Mover { return Mover{Speed: m.Speed, GoalSpace: m.GoalSpace, Goal: m.Goal} }

// Razones de MoveFailed.
const (
	ReasonNoPath      = "no_path"
	ReasonNotWalkable = "not_walkable"
	ReasonBudget      = "search_budget"
)

// MoveArrived se emite cuando un Mover llega a su destino.
type MoveArrived struct {
	Entity ids.ID `json:"entity"`
}

func (MoveArrived) Kind() string { return "move_arrived" }

// MoveFailed se emite cuando un movimiento se abandona.
type MoveFailed struct {
	Entity ids.ID `json:"entity"`
	Reason string `json:"reason"`
}

func (MoveFailed) Kind() string { return "move_failed" }

// MoveTo pide que una entidad camine hasta Goal, en coordenadas locales de
// Space (Nil = el Space actual). Reemplaza cualquier movimiento en curso.
type MoveTo struct {
	Entity ids.ID     `json:"entity"`
	Space  ids.ID     `json:"space,omitempty"`
	Goal   units.Vec2 `json:"goal"`
}

func (MoveTo) Kind() string { return "move_to" }

func (c MoveTo) Apply(w *world.World) error {
	m := ecs.Get[Mover](w.ECS()).Ptr(c.Entity)
	loc, ok := comp.Of(w.ECS()).Location.Get(c.Entity)
	if m == nil || !ok {
		return fmt.Errorf("move_to %v: no es un mover ubicado", c.Entity)
	}
	target := c.Space
	if target.IsNil() {
		target = loc.Space
	} else if !space.Of(w.ECS()).IsSpace(target) {
		return fmt.Errorf("move_to %v: %w: %v", c.Entity, space.ErrNotSpace, target)
	}
	*m = Mover{Speed: m.Speed, Status: Planning, GoalSpace: target, Goal: c.Goal}
	return nil
}

// Stop detiene el movimiento de una entidad, sin evento.
type Stop struct {
	Entity ids.ID `json:"entity"`
}

func (Stop) Kind() string { return "stop" }

func (c Stop) Apply(w *world.World) error {
	m := ecs.Get[Mover](w.ECS()).Ptr(c.Entity)
	if m == nil {
		return fmt.Errorf("stop %v: no es un mover", c.Entity)
	}
	*m = m.idle()
	return nil
}

// CellCost asigna un costo a una celda.
type CellCost struct {
	X    int  `json:"x"`
	Y    int  `json:"y"`
	Cost Cost `json:"cost"`
}

// SetCells cambia celdas de la grilla de un Space. Se rechaza entero si
// alguna celda está fuera de rango o si bloquearía una celda ocupada por un
// mover (aplastar tinies es asunto de sistemas posteriores, no de nav).
type SetCells struct {
	Space ids.ID     `json:"space"`
	Cells []CellCost `json:"cells"`
}

func (SetCells) Kind() string { return "set_cells" }

func (c SetCells) Apply(w *world.World) error {
	g := ecs.Get[Grid](w.ECS()).Ptr(c.Space)
	if g == nil {
		return fmt.Errorf("set_cells %v: el Space no tiene grilla", c.Space)
	}
	occupied := map[int]bool{}
	locs := comp.Of(w.ECS()).Location
	ecs.Get[Mover](w.ECS()).Each(func(e ids.ID, _ *Mover) {
		if l, ok := locs.Get(e); ok && l.Space == c.Space {
			x, y := g.CellOf(l.Pos.XY())
			occupied[y*g.W+x] = true
		}
	})
	for _, cc := range c.Cells {
		if !g.In(cc.X, cc.Y) || cc.Cost > MaxCost {
			return fmt.Errorf("set_cells: celda inválida %+v", cc)
		}
		if cc.Cost == Blocked && occupied[cc.Y*g.W+cc.X] {
			return fmt.Errorf("set_cells: celda (%d,%d) ocupada por un mover", cc.X, cc.Y)
		}
	}
	for _, cc := range c.Cells {
		g.Set(cc.X, cc.Y, cc.Cost)
	}
	return nil
}

// PlanBudget es la cantidad máxima de planificaciones por tick (cada una
// corre hasta MaxSearches búsquedas A*). Los movers que no entran esperan al
// tick siguiente, en orden del store.
const PlanBudget = 64

// planner calcula rutas para los movers en Planning.
type planner struct{ router Router }

func (*planner) Name() string { return "nav.plan" }

func (p *planner) Run(w *world.World) {
	locs := comp.Of(w.ECS()).Location
	budget := PlanBudget
	ecs.Get[Mover](w.ECS()).Each(func(e ids.ID, m *Mover) {
		if m.Status != Planning || budget == 0 {
			return
		}
		budget--
		loc, _ := locs.Get(e)
		legs, err := p.router.Route(w.ECS(), loc.Space, loc.Pos.XY(), m.GoalSpace, m.Goal)
		if err != nil {
			fail(w, e, m, reason(err))
			return
		}
		m.Status, m.Legs = Moving, legs
	})
}

func reason(err error) string {
	switch {
	case errors.Is(err, ErrNotWalkable):
		return ReasonNotWalkable
	case errors.Is(err, ErrSearchBudget):
		return ReasonBudget
	default:
		return ReasonNoPath
	}
}

func fail(w *world.World, e ids.ID, m *Mover, why string) {
	*m = m.idle()
	w.Emit(MoveFailed{Entity: e, Reason: why})
}

// mover avanza a los movers en Moving por sus tramos y cruza portales. Si la
// ruta quedó obsoleta (cambió una grilla o un portal, o la entidad ya no está
// en el Space del primer tramo), el mover se queda quieto este tick y vuelve
// a Planning: la ruta se recalcula desde donde está.
type mover struct{}

func (mover) Name() string { return "nav.move" }

func (mover) Run(w *world.World) {
	reg := w.ECS()
	grids := ecs.Get[Grid](reg)
	portals := ecs.Get[Portal](reg)
	locs := comp.Of(reg).Location
	g := space.Of(reg)
	ecs.Get[Mover](reg).Each(func(e ids.ID, m *Mover) {
		if m.Status != Moving {
			return
		}
		loc := locs.Ptr(e)
		if loc == nil || !routeValid(m.Legs, loc.Space, g, grids, portals) {
			m.Status, m.Legs = Planning, nil
			return
		}
		pos := loc.Pos.XY()
		left := m.Speed / clock.TPS
		for len(m.Legs) > 0 {
			leg := &m.Legs[0]
			if len(leg.Path) > 0 {
				if left == 0 {
					break
				}
				d := leg.Path[0].Sub(pos)
				n := d.Len()
				if n <= left {
					pos, left = leg.Path[0], left-n
					leg.Path = leg.Path[1:]
					continue
				}
				pos = pos.Add(units.Vec2{X: d.X.MulDiv(int64(left), int64(n)), Y: d.Y.MulDiv(int64(left), int64(n))})
				left = 0
				break
			}
			if leg.Portal.IsNil() { // fin de la ruta
				m.Legs = nil
				break
			}
			_, out := portals.Ptr(leg.Portal).ends(leg.Reverse)
			if g.CanPlace(e, out.Space) != nil { // p. ej. cruzaría hacia sí mismo
				loc.Pos.X, loc.Pos.Y = pos.X, pos.Y
				fail(w, e, m, ReasonNoPath)
				return
			}
			from := loc.Space
			loc.Space, pos = out.Space, out.Pos
			w.Emit(space.EntityReparented{Entity: e, From: from, To: out.Space, Cause: "portal"})
			m.Legs = m.Legs[1:]
		}
		loc.Pos.X, loc.Pos.Y = pos.X, pos.Y
		if len(m.Legs) == 0 {
			*m = m.idle()
			w.Emit(MoveArrived{Entity: e})
		}
	})
}

// routeValid indica si la ruta sigue vigente: cada tramo en un Space vivo
// con la misma versión de grilla, y cada portal vivo, habilitado, con la
// misma versión y conectando los Spaces de los tramos que une.
func routeValid(legs []Leg, current ids.ID, g space.Graph, grids *ecs.Store[Grid], portals *ecs.Store[Portal]) bool {
	if len(legs) == 0 || legs[0].Space != current {
		return false
	}
	for i, l := range legs {
		if !g.IsSpace(l.Space) {
			return false
		}
		if grid := grids.Ptr(l.Space); grid != nil && grid.Version != l.GridVersion {
			return false
		}
		if l.Portal.IsNil() {
			return i == len(legs)-1
		}
		p := portals.Ptr(l.Portal)
		if p == nil || p.Disabled || p.Version != l.PortalVersion || i+1 >= len(legs) {
			return false
		}
		if in, out := p.ends(l.Reverse); in.Space != l.Space || out.Space != legs[i+1].Space {
			return false
		}
	}
	return true
}

// Install registra los stores "nav_grid", "mover" y "portal", comandos,
// eventos, sistemas e invariantes de nav. Requiere comp y space.
func Install(w *world.World) error {
	r := w.ECS()
	if _, err := ecs.Register[Grid](r, "nav_grid"); err != nil {
		return err
	}
	if _, err := ecs.Register[Mover](r, "mover"); err != nil {
		return err
	}
	if _, err := ecs.Register[Portal](r, "portal"); err != nil {
		return err
	}
	w.AddInvariant("nav.mover", checkMovers)
	w.AddInvariant("nav.portal", checkPortals)
	return errors.Join(
		world.RegisterCommand[MoveTo](w),
		world.RegisterCommand[Stop](w),
		world.RegisterCommand[SetCells](w),
		world.RegisterCommand[SetPortal](w),
		world.RegisterEvent[MoveArrived](w),
		world.RegisterEvent[MoveFailed](w),
		w.Register(world.SystemSpec{Phase: clock.PhaseAct}, &planner{}),
		w.Register(world.SystemSpec{Phase: clock.PhaseMove}, mover{}),
		// Mismo Order que space.orphans y registrado después: corre a
		// continuación, cuando ya se conocen todas las cascadas.
		w.Register(world.SystemSpec{Phase: clock.PhaseFlush, Order: 1 << 30}, portalCleanup{}),
	)
}

// checkMovers: un mover en Moving tiene ruta, y si su primer tramo está en el
// Space donde se encuentra (si no, nav.move replaneará), está en una celda
// transitable.
func checkMovers(w *world.World) error {
	grids := ecs.Get[Grid](w.ECS())
	locs := comp.Of(w.ECS()).Location
	var err error
	ecs.Get[Mover](w.ECS()).Each(func(e ids.ID, m *Mover) {
		if err != nil || m.Status != Moving {
			return
		}
		if len(m.Legs) == 0 {
			err = fmt.Errorf("mover %v en moving sin ruta", e)
			return
		}
		l, ok := locs.Get(e)
		if g := grids.Ptr(l.Space); ok && l.Space == m.Legs[0].Space && g != nil && !g.Walkable(l.Pos.XY()) {
			err = fmt.Errorf("mover %v en celda bloqueada %v", e, l.Pos)
		}
	})
	return err
}
