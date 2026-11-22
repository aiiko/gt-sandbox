// Package nav implementa la navegación local: una grilla de costos por Space,
// A* sobre ella, y el movimiento de entidades con Mover.
//
// La navegación entre Spaces (portales, rutas jerárquicas) llega en M4.
package nav

import (
	"encoding/json"
	"fmt"

	"gtsandbox/internal/core/units"
)

// Cost es el costo de atravesar una celda: 0 = bloqueada, 1..9 = transitable
// (1 normal, mayor = más lento o menos deseable).
type Cost uint8

const (
	Blocked Cost = 0
	Free    Cost = 1
	MaxCost Cost = 9
)

// MaxCells limita el tamaño de una grilla (≈ 2 m × 2 m a 1 mm).
const MaxCells = 1 << 22

// Grid es el componente de navegación de un Space: una grilla regular de
// celdas cuadradas en coordenadas locales del Space. La celda (0,0) empieza
// en Origin; x crece con X e y con Y.
//
// En JSON se representa como filas de texto ('#' bloqueada, '.' costo 1,
// '2'..'9' costo), legible en snapshots y escribible a mano en escenarios.
// La fila 0 es la de menor Y.
type Grid struct {
	Origin units.Vec2
	Cell   units.Length
	W, H   int
	// Version se incrementa con cada cambio; las rutas la recuerdan para
	// saber si quedaron obsoletas.
	Version uint64
	cost    []Cost
}

// NewGrid crea una grilla con todas las celdas libres.
func NewGrid(origin units.Vec2, cell units.Length, w, h int) (*Grid, error) {
	if cell <= 0 || w <= 0 || h <= 0 || w*h > MaxCells {
		return nil, fmt.Errorf("nav: grilla inválida (cell=%v, %dx%d)", cell, w, h)
	}
	g := &Grid{Origin: origin, Cell: cell, W: w, H: h, cost: make([]Cost, w*h)}
	for i := range g.cost {
		g.cost[i] = Free
	}
	return g, nil
}

func (g *Grid) In(x, y int) bool { return x >= 0 && y >= 0 && x < g.W && y < g.H }

// At devuelve el costo de una celda; fuera de la grilla es Blocked.
func (g *Grid) At(x, y int) Cost {
	if !g.In(x, y) {
		return Blocked
	}
	return g.cost[y*g.W+x]
}

// Set cambia el costo de una celda e incrementa Version si cambió.
func (g *Grid) Set(x, y int, c Cost) {
	if !g.In(x, y) {
		panic(fmt.Sprintf("nav: celda (%d,%d) fuera de la grilla %dx%d", x, y, g.W, g.H))
	}
	if c > MaxCost {
		c = MaxCost
	}
	if g.cost[y*g.W+x] != c {
		g.cost[y*g.W+x] = c
		g.Version++
	}
}

// CellOf devuelve la celda que contiene p (coordenadas locales del Space).
func (g *Grid) CellOf(p units.Vec2) (x, y int) {
	return floorDiv(p.X-g.Origin.X, g.Cell), floorDiv(p.Y-g.Origin.Y, g.Cell)
}

// Center devuelve el centro de una celda.
func (g *Grid) Center(x, y int) units.Vec2 {
	return units.Vec2{
		X: g.Origin.X + units.Length(x)*g.Cell + g.Cell/2,
		Y: g.Origin.Y + units.Length(y)*g.Cell + g.Cell/2,
	}
}

// Walkable indica si el punto p cae en una celda transitable.
func (g *Grid) Walkable(p units.Vec2) bool { return g.At(g.CellOf(p)) != Blocked }

func floorDiv(a, b units.Length) int {
	q := a / b
	if (a%b != 0) && (a < 0) {
		q--
	}
	return int(q)
}

type gridJSON struct {
	Origin  units.Vec2   `json:"origin"`
	Cell    units.Length `json:"cell"`
	Version uint64       `json:"version"`
	Rows    []string     `json:"rows"`
}

func (g Grid) MarshalJSON() ([]byte, error) {
	rows := make([]string, g.H)
	buf := make([]byte, g.W)
	for y := range g.H {
		for x := range g.W {
			switch c := g.cost[y*g.W+x]; c {
			case Blocked:
				buf[x] = '#'
			case Free:
				buf[x] = '.'
			default:
				buf[x] = '0' + byte(c)
			}
		}
		rows[y] = string(buf)
	}
	return json.Marshal(gridJSON{g.Origin, g.Cell, g.Version, rows})
}

func (g *Grid) UnmarshalJSON(b []byte) error {
	var j gridJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	if len(j.Rows) == 0 { // valor cero (p. ej. validación de serialización)
		*g = Grid{}
		return nil
	}
	n, err := NewGrid(j.Origin, j.Cell, len(j.Rows[0]), len(j.Rows))
	if err != nil {
		return err
	}
	for y, row := range j.Rows {
		if len(row) != n.W {
			return fmt.Errorf("nav: fila %d mide %d, se esperaba %d", y, len(row), n.W)
		}
		for x := range len(row) {
			switch ch := row[x]; {
			case ch == '#':
				n.cost[y*n.W+x] = Blocked
			case ch == '.':
				n.cost[y*n.W+x] = Free
			case ch >= '2' && ch <= '9':
				n.cost[y*n.W+x] = Cost(ch - '0')
			default:
				return fmt.Errorf("nav: carácter %q inválido en fila %d", ch, y)
			}
		}
	}
	n.Version = j.Version
	*g = *n
	return nil
}

// Region devuelve una prueba de alcanzabilidad: si un punto cae en una celda
// a la que se llega desde from por celdas transitables vecinas (4 vecinos),
// sin alejarse más de radius en cada eje. Fuera de la grilla, nada.
func (g *Grid) Region(from units.Vec2, radius units.Length) func(units.Vec2) bool {
	fx, fy := g.CellOf(from)
	r := int(radius/g.Cell) + 1
	seen := map[[2]int]bool{}
	if g.At(fx, fy) != Blocked {
		seen[[2]int{fx, fy}] = true
		queue := [][2]int{{fx, fy}}
		for len(queue) > 0 {
			c := queue[0]
			queue = queue[1:]
			for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				n := [2]int{c[0] + d[0], c[1] + d[1]}
				if seen[n] || n[0]-fx > r || fx-n[0] > r || n[1]-fy > r || fy-n[1] > r || g.At(n[0], n[1]) == Blocked {
					continue
				}
				seen[n] = true
				queue = append(queue, n)
			}
		}
	}
	return func(p units.Vec2) bool {
		x, y := g.CellOf(p)
		return seen[[2]int{x, y}]
	}
}
