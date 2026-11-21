package nav

import (
	"errors"
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/world"
)

// Status es el estado de un Mover.
type Status uint8

const (
	Idle     Status = iota
	Planning        // espera ruta (nav.plan)
	Moving          // recorriendo Path (nav.move)
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

// Mover es el componente de las entidades que caminan por su Space.
//
// Goal y Path están en coordenadas locales de Space (el Space en el que se
// pidió el movimiento). Si la entidad cambia de Space, el movimiento falla.
type Mover struct {
	Speed       units.Length `json:"speed"` // µm por segundo
	Status      Status       `json:"status"`
	Space       ids.ID       `json:"space"`
	Goal        units.Vec2   `json:"goal"`
	Path        []units.Vec2 `json:"path,omitempty"`
	PathVersion uint64       `json:"path_version"`
}

// Razones de MoveFailed.
const (
	ReasonNoPath       = "no_path"
	ReasonNotWalkable  = "not_walkable"
	ReasonBudget       = "search_budget"
	ReasonSpaceChanged = "space_changed"
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

// MoveTo pide que una entidad camine hasta Goal (coordenadas locales de su
// Space actual). Reemplaza cualquier movimiento en curso.
type MoveTo struct {
	Entity ids.ID     `json:"entity"`
	Goal   units.Vec2 `json:"goal"`
}

func (MoveTo) Kind() string { return "move_to" }

func (c MoveTo) Apply(w *world.World) error {
	m := ecs.Get[Mover](w.ECS()).Ptr(c.Entity)
	loc, ok := comp.Of(w.ECS()).Location.Get(c.Entity)
	if m == nil || !ok {
		return fmt.Errorf("move_to %v: no es un mover ubicado", c.Entity)
	}
	*m = Mover{Speed: m.Speed, Status: Planning, Space: loc.Space, Goal: c.Goal}
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
	*m = Mover{Speed: m.Speed, Space: m.Space, Goal: m.Goal}
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

// PlanBudget es la cantidad máxima de búsquedas A* por tick. Los movers que
// no entran esperan al tick siguiente (en orden del store).
const PlanBudget = 64

// planner corre A* para los movers en Planning.
type planner struct{ search Searcher }

func (*planner) Name() string { return "nav.plan" }

func (p *planner) Run(w *world.World) {
	grids := ecs.Get[Grid](w.ECS())
	locs := comp.Of(w.ECS()).Location
	budget := PlanBudget
	ecs.Get[Mover](w.ECS()).Each(func(e ids.ID, m *Mover) {
		if m.Status != Planning || budget == 0 {
			return
		}
		loc, _ := locs.Get(e)
		if loc.Space != m.Space {
			fail(w, e, m, ReasonSpaceChanged)
			return
		}
		g := grids.Ptr(m.Space)
		if g == nil { // Space sin grilla: línea recta
			m.Status, m.Path = Moving, []units.Vec2{m.Goal}
			return
		}
		budget--
		path, err := p.search.FindPath(g, loc.Pos.XY(), m.Goal, g.W*g.H)
		if err != nil {
			fail(w, e, m, reason(err))
			return
		}
		m.Status, m.Path, m.PathVersion = Moving, path, g.Version
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
	*m = Mover{Speed: m.Speed, Space: m.Space, Goal: m.Goal}
	w.Emit(MoveFailed{Entity: e, Reason: why})
}

// mover avanza a los movers en Moving por su Path. Si la grilla cambió desde
// que se planeó la ruta, el mover se queda quieto y vuelve a Planning.
type mover struct{}

func (mover) Name() string { return "nav.move" }

func (mover) Run(w *world.World) {
	grids := ecs.Get[Grid](w.ECS())
	locs := comp.Of(w.ECS()).Location
	ecs.Get[Mover](w.ECS()).Each(func(e ids.ID, m *Mover) {
		if m.Status != Moving {
			return
		}
		loc := locs.Ptr(e)
		if loc == nil || loc.Space != m.Space {
			fail(w, e, m, ReasonSpaceChanged)
			return
		}
		if g := grids.Ptr(m.Space); g != nil && g.Version != m.PathVersion {
			m.Status, m.Path = Planning, nil
			return
		}
		pos := loc.Pos.XY()
		left := m.Speed / clock.TPS
		for left > 0 && len(m.Path) > 0 {
			d := m.Path[0].Sub(pos)
			n := d.Len()
			if n <= left {
				pos, left = m.Path[0], left-n
				m.Path = m.Path[1:]
				continue
			}
			pos = pos.Add(units.Vec2{X: d.X.MulDiv(int64(left), int64(n)), Y: d.Y.MulDiv(int64(left), int64(n))})
			left = 0
		}
		loc.Pos.X, loc.Pos.Y = pos.X, pos.Y
		if len(m.Path) == 0 {
			m.Status, m.Path = Idle, nil
			w.Emit(MoveArrived{Entity: e})
		}
	})
}

// Install registra los stores "nav_grid" y "mover", comandos, eventos,
// sistemas e invariantes de nav. Requiere comp y space.
func Install(w *world.World) error {
	r := w.ECS()
	if _, err := ecs.Register[Grid](r, "nav_grid"); err != nil {
		return err
	}
	if _, err := ecs.Register[Mover](r, "mover"); err != nil {
		return err
	}
	w.AddInvariant("nav", checkInvariants)
	return errors.Join(
		world.RegisterCommand[MoveTo](w),
		world.RegisterCommand[Stop](w),
		world.RegisterCommand[SetCells](w),
		world.RegisterEvent[MoveArrived](w),
		world.RegisterEvent[MoveFailed](w),
		w.Register(world.SystemSpec{Phase: clock.PhaseAct}, &planner{}),
		w.Register(world.SystemSpec{Phase: clock.PhaseMove}, mover{}),
	)
}

// checkInvariants: un mover en Moving está en una celda transitable de su
// grilla, y su ruta no es vacía.
func checkInvariants(w *world.World) error {
	grids := ecs.Get[Grid](w.ECS())
	locs := comp.Of(w.ECS()).Location
	var err error
	ecs.Get[Mover](w.ECS()).Each(func(e ids.ID, m *Mover) {
		if err != nil || m.Status != Moving {
			return
		}
		if len(m.Path) == 0 {
			err = fmt.Errorf("mover %v en moving sin ruta", e)
			return
		}
		l, ok := locs.Get(e)
		if g := grids.Ptr(m.Space); ok && g != nil && !g.Walkable(l.Pos.XY()) {
			err = fmt.Errorf("mover %v en celda bloqueada %v", e, l.Pos)
		}
	})
	return err
}
