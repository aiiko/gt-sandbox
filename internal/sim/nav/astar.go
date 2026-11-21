package nav

import (
	"errors"

	"gtsandbox/internal/core/units"
)

var (
	ErrNoPath       = errors.New("nav: no hay camino")
	ErrNotWalkable  = errors.New("nav: origen o destino no transitable")
	ErrSearchBudget = errors.New("nav: búsqueda excedió el presupuesto")
)

// Costos de paso por celda de costo 1 (enteros: 10 y 14 ≈ 10·√2).
const (
	stepStraight = 10
	stepDiagonal = 14
)

var dirs = [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

// Searcher ejecuta A* reutilizando memoria entre búsquedas. No es estado de
// la simulación (no se guarda): es un espacio de trabajo.
type Searcher struct {
	stamp  []uint32 // stamp[i] == gen ⇒ g/parent/closed válidos en esta búsqueda
	gen    uint32
	g      []int32
	parent []int32
	closed []bool
	open   nodeHeap
}

func (s *Searcher) reset(n int) {
	if len(s.stamp) < n {
		s.stamp = make([]uint32, n)
		s.g = make([]int32, n)
		s.parent = make([]int32, n)
		s.closed = make([]bool, n)
		s.gen = 0
	}
	s.gen++
	if s.gen == 0 { // wrap: limpiar
		clear(s.stamp)
		s.gen = 1
	}
	s.open = s.open[:0]
}

func (s *Searcher) touch(i int32) {
	if s.stamp[i] != s.gen {
		s.stamp[i] = s.gen
		s.g[i] = -1
		s.closed[i] = false
	}
}

// FindPath busca un camino de from a to (coordenadas locales) y devuelve los
// puntos a recorrer: centros de celda donde cambia la dirección, terminando
// exactamente en to. No incluye from.
//
// Movimiento en 8 direcciones sin cortar esquinas: un paso diagonal requiere
// que las dos celdas ortogonales adyacentes sean transitables. Así, el
// segmento recto entre puntos consecutivos nunca toca una celda bloqueada.
// maxExpand limita los nodos expandidos (0 = sin límite).
func (s *Searcher) FindPath(g *Grid, from, to units.Vec2, maxExpand int) ([]units.Vec2, error) {
	sx, sy := g.CellOf(from)
	tx, ty := g.CellOf(to)
	if g.At(sx, sy) == Blocked || g.At(tx, ty) == Blocked {
		return nil, ErrNotWalkable
	}
	if sx == tx && sy == ty {
		return []units.Vec2{to}, nil
	}
	s.reset(g.W * g.H)
	start, goal := int32(sy*g.W+sx), int32(ty*g.W+tx)
	s.touch(start)
	s.g[start] = 0
	s.open.push(node{f: heuristic(sx, sy, tx, ty), h: heuristic(sx, sy, tx, ty), idx: start})

	expanded := 0
	for len(s.open) > 0 {
		cur := s.open.pop()
		if s.closed[cur.idx] {
			continue
		}
		if cur.idx == goal {
			return s.build(g, start, goal, to), nil
		}
		s.closed[cur.idx] = true
		if expanded++; maxExpand > 0 && expanded > maxExpand {
			return nil, ErrSearchBudget
		}
		cx, cy := int(cur.idx)%g.W, int(cur.idx)/g.W
		for _, d := range dirs {
			nx, ny := cx+d[0], cy+d[1]
			c := g.At(nx, ny)
			if c == Blocked {
				continue
			}
			step := int32(stepStraight)
			if d[0] != 0 && d[1] != 0 {
				if g.At(cx+d[0], cy) == Blocked || g.At(cx, cy+d[1]) == Blocked {
					continue // no cortar esquinas
				}
				step = stepDiagonal
			}
			ni := int32(ny*g.W + nx)
			s.touch(ni)
			if s.closed[ni] {
				continue
			}
			ng := s.g[cur.idx] + step*int32(c)
			if s.g[ni] >= 0 && ng >= s.g[ni] {
				continue
			}
			s.g[ni] = ng
			s.parent[ni] = cur.idx
			h := heuristic(nx, ny, tx, ty)
			s.open.push(node{f: ng + h, h: h, idx: ni})
		}
	}
	return nil, ErrNoPath
}

// heuristic octil, admisible porque el costo mínimo por celda es 1.
func heuristic(x, y, tx, ty int) int32 {
	dx, dy := abs(x-tx), abs(y-ty)
	return int32(stepStraight*max(dx, dy) + (stepDiagonal-stepStraight)*min(dx, dy))
}

func (s *Searcher) build(g *Grid, start, goal int32, to units.Vec2) []units.Vec2 {
	var cells []int32
	for i := goal; i != start; i = s.parent[i] {
		cells = append(cells, i)
	}
	// cells va de goal hacia start; recorrerlo al revés.
	out := make([]units.Vec2, 0, 8)
	dir := func(a, b int32) (int, int) {
		return int(b)%g.W - int(a)%g.W, int(b)/g.W - int(a)/g.W
	}
	prev := start
	for k := len(cells) - 1; k >= 0; k-- {
		c := cells[k]
		last := k == 0
		// Se saltean las celdas colineales: el segmento recto entre los puntos
		// que quedan atraviesa solo celdas de la ruta o vecinas ortogonales
		// de un paso diagonal (transitables por la regla de esquinas). Vale
		// también para los extremos, que son puntos arbitrarios de su celda.
		if !last {
			dx1, dy1 := dir(prev, c)
			dx2, dy2 := dir(c, cells[k-1])
			if dx1 == dx2 && dy1 == dy2 {
				prev = c
				continue
			}
		}
		if last {
			out = append(out, to)
		} else {
			out = append(out, g.Center(int(c)%g.W, int(c)/g.W))
		}
		prev = c
	}
	return out
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// node ordena por f, luego h (preferir el más cercano al destino), luego
// índice: desempate total y determinista.
type node struct {
	f, h int32
	idx  int32
}

func (a node) less(b node) bool {
	if a.f != b.f {
		return a.f < b.f
	}
	if a.h != b.h {
		return a.h < b.h
	}
	return a.idx < b.idx
}

type nodeHeap []node

func (h *nodeHeap) push(n node) {
	*h = append(*h, n)
	a := *h
	i := len(a) - 1
	for i > 0 {
		p := (i - 1) / 2
		if !a[i].less(a[p]) {
			break
		}
		a[i], a[p] = a[p], a[i]
		i = p
	}
}

func (h *nodeHeap) pop() node {
	a := *h
	top := a[0]
	last := len(a) - 1
	a[0] = a[last]
	a = a[:last]
	i := 0
	for {
		l, r, m := 2*i+1, 2*i+2, i
		if l < len(a) && a[l].less(a[m]) {
			m = l
		}
		if r < len(a) && a[r].less(a[m]) {
			m = r
		}
		if m == i {
			break
		}
		a[i], a[m] = a[m], a[i]
		i = m
	}
	*h = a
	return top
}
