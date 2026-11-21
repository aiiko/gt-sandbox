package scenario

import (
	"encoding/json"
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/rng"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Crowd es un mundo de carga: n tinies caminando por un escritorio de
// 1,2 × 0,8 m (celdas de 1 cm) con obstáculos aleatorios. Cada vez que uno
// llega o falla, Step le da otro destino. Para benchmarks y pruebas de
// estabilidad.
type Crowd struct {
	World  *world.World
	Desk   ids.ID
	Tinies []ids.ID
	grid   *nav.Grid
	rnd    *rng.Stream
}

// Tamaño del escritorio de Crowd.
const (
	CrowdCell = units.Centimeter
	CrowdW    = 120
	CrowdH    = 80
)

// NewCrowd arma el mundo (consume dos ticks) y manda a todos a caminar.
func NewCrowd(n int, seed uint64, opts ...engine.Option) (*Crowd, error) {
	w, err := engine.New(world.Config{Seed: seed}, opts...)
	if err != nil {
		return nil, err
	}
	// Generador propio (fuera de la simulación): decide inputs, como un jugador.
	c := &Crowd{World: w, rnd: rng.NewStream(seed, "crowd")}
	c.grid, _ = nav.NewGrid(units.Vec2{}, CrowdCell, CrowdW, CrowdH)
	for range 25 { // cajas de 2..10 cm
		bw, bh := 2+c.rnd.IntN(9), 2+c.rnd.IntN(9)
		x0, y0 := c.rnd.IntN(CrowdW-bw), c.rnd.IntN(CrowdH-bh)
		for y := y0; y < y0+bh; y++ {
			for x := x0; x < x0+bw; x++ {
				c.grid.Set(x, y, nav.Blocked)
			}
		}
	}
	gridJSON, err := c.grid.MarshalJSON()
	if err != nil {
		return nil, err
	}
	ext := units.RectWH(0, 0, CrowdW*CrowdCell, CrowdH*CrowdCell)
	if c.Desk, err = spawnOne(w, cmds.Spawn{Name: "desk", Space: &space.Space{Kind: space.Static, Extent: ext},
		Components: map[string]json.RawMessage{"nav_grid": gridJSON}}); err != nil {
		return nil, err
	}
	for i := range n {
		p := c.freePoint()
		if _, err := w.Enqueue(cmds.Spawn{
			Name:       comp.Name(fmt.Sprintf("tiny%d", i)),
			Location:   &comp.Location{Space: c.Desk, Pos: units.V3(p.X, p.Y, 0)},
			Body:       &comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter},
			Components: map[string]json.RawMessage{"mover": json.RawMessage(`{"speed":20000}`)},
		}); err != nil {
			return nil, err
		}
	}
	w.Step()
	for _, e := range w.Events() {
		if s, ok := e.Event.(cmds.EntitySpawned); ok {
			c.Tinies = append(c.Tinies, s.Entity)
			c.send(s.Entity)
		}
	}
	if len(c.Tinies) != n {
		return nil, fmt.Errorf("crowd: se crearon %d de %d", len(c.Tinies), n)
	}
	return c, nil
}

func (c *Crowd) freePoint() units.Vec2 {
	for {
		p := units.V2(units.Length(c.rnd.IntN(CrowdW*int(CrowdCell))), units.Length(c.rnd.IntN(CrowdH*int(CrowdCell))))
		if c.grid.Walkable(p) {
			return p
		}
	}
}

func (c *Crowd) send(e ids.ID) {
	if _, err := c.World.Enqueue(nav.MoveTo{Entity: e, Goal: c.freePoint()}); err != nil {
		panic(err)
	}
}

// Step avanza un tick y reasigna destino a quien llegó o falló.
func (c *Crowd) Step() {
	c.World.Step()
	for _, e := range c.World.Events() {
		switch ev := e.Event.(type) {
		case nav.MoveArrived:
			c.send(ev.Entity)
		case nav.MoveFailed:
			c.send(ev.Entity)
		}
	}
}

func spawnOne(w *world.World, s cmds.Spawn) (ids.ID, error) {
	if _, err := w.Enqueue(s); err != nil {
		return ids.Nil, err
	}
	w.Step()
	for _, e := range w.Events() {
		switch ev := e.Event.(type) {
		case cmds.EntitySpawned:
			return ev.Entity, nil
		case world.CommandRejected:
			return ids.Nil, fmt.Errorf("spawn rechazado: %s", ev.Error)
		}
	}
	return ids.Nil, fmt.Errorf("spawn sin evento")
}
