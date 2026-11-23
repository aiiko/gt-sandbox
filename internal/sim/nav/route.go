package nav

import (
	"math"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Leg es un tramo de una ruta dentro de un Space. Al terminar Path, si
// Portal no es Nil, el mover cruza ese portal y sigue con el tramo siguiente.
// Las versiones permiten detectar que la ruta quedó obsoleta.
type Leg struct {
	Space         ids.ID       `json:"space"`
	Path          []units.Vec2 `json:"path,omitempty"`
	GridVersion   uint64       `json:"grid_version"`
	Portal        ids.ID       `json:"portal,omitempty"`
	Reverse       bool         `json:"reverse,omitempty"` // cruza de B a A
	PortalVersion uint64       `json:"portal_version,omitempty"`
}

// MaxSearches limita las búsquedas A* de una planificación.
const MaxSearches = 128

// Router planifica rutas entre Spaces: un grafo cuyos nodos son el origen,
// el destino y los extremos de los portales habilitados. Las aristas dentro
// de un Space empiezan estimadas (línea recta, cota inferior); se corre
// Dijkstra, se calcula con A* cada tramo estimado de la mejor ruta y se
// repite hasta que la mejor ruta sea toda exacta. Así A* corre solo en los
// tramos candidatos.
//
// Es un espacio de trabajo reutilizable, no estado de la simulación.
type Router struct {
	search Searcher

	nodes   []rnode
	bySpace map[ids.ID][]int
	exact   map[[2]int]edge
	dist    []int64
	prev    []int
	done    []bool
}

type rnode struct {
	space   ids.ID
	pos     units.Vec2
	portal  ids.ID // Nil para origen y destino
	partner int    // índice del otro extremo del portal, -1 si no hay
	reverse bool   // cruzar desde este extremo es ir de B a A
	version uint64
	cost    int64
}

type edge struct {
	cost int64
	path []units.Vec2
	ok   bool
}

const inf = math.MaxInt64

// Route busca una ruta desde (fromSpace, from) hasta (toSpace, to).
func (r *Router) Route(reg *ecs.Registry, fromSpace ids.ID, from units.Vec2, toSpace ids.ID, to units.Vec2) ([]Leg, error) {
	return r.route(reg, "", nil, nil, ids.Nil, fromSpace, from, toSpace, to)
}

// RouteFor busca una ruta aplicando el filtro de escala y penalizaciones de
// portales (si se proveyeron).
func (r *Router) RouteFor(w *world.World, actor ids.ID, scale string, penalty PortalPenalty, fromSpace ids.ID, from units.Vec2, toSpace ids.ID, to units.Vec2) ([]Leg, error) {
	return r.route(w.ECS(), scale, w, penalty, actor, fromSpace, from, toSpace, to)
}

func (r *Router) route(reg *ecs.Registry, scale string, w *world.World, penalty PortalPenalty, actor ids.ID, fromSpace ids.ID, from units.Vec2, toSpace ids.ID, to units.Vec2) ([]Leg, error) {
	grids := ecs.Get[Grid](reg)
	spaces := ecs.Get[space.Space](reg)
	g := space.Of(reg)
	if !g.IsSpace(fromSpace) || !g.IsSpace(toSpace) {
		return nil, ErrNotWalkable
	}
	for _, end := range [2]PortalEnd{{fromSpace, from}, {toSpace, to}} {
		if grid := grids.Ptr(end.Space); grid != nil && !grid.Walkable(end.Pos) {
			return nil, ErrNotWalkable
		}
	}
	r.nodes = append(r.nodes[:0], rnode{space: fromSpace, pos: from, partner: -1}, rnode{space: toSpace, pos: to, partner: -1})
	ecs.Get[Portal](reg).Each(func(e ids.ID, p *Portal) {
		if p.Disabled || (p.Scale != "" && p.Scale != scale) || !g.IsSpace(p.A.Space) || !g.IsSpace(p.B.Space) {
			return
		}
		c := int64(p.Cost)
		if penalty != nil && w != nil {
			c += int64(penalty(w, actor, e))
		}
		i := len(r.nodes)
		r.nodes = append(r.nodes,
			rnode{space: p.A.Space, pos: p.A.Pos, portal: e, partner: i + 1, version: p.Version, cost: c},
			rnode{space: p.B.Space, pos: p.B.Pos, portal: e, partner: i, reverse: true, version: p.Version, cost: c})
	})
	if r.bySpace == nil {
		r.bySpace, r.exact = map[ids.ID][]int{}, map[[2]int]edge{}
	}
	clear(r.bySpace)
	clear(r.exact)
	for i, n := range r.nodes {
		r.bySpace[n.space] = append(r.bySpace[n.space], i)
	}

	searches := 0
	for {
		path := r.dijkstra(spaces)
		if path == nil {
			return nil, ErrNoPath
		}
		refined := false
		for k := 0; k+1 < len(path); k++ {
			u, v := path[k], path[k+1]
			if r.isCrossing(u, v) {
				continue
			}
			if _, ok := r.exact[[2]int{u, v}]; ok {
				continue
			}
			e := edge{}
			sp, _ := spaces.Get(r.nodes[u].space)
			if grid := grids.Ptr(r.nodes[u].space); grid == nil {
				e = edge{cost: int64(sp.SegmentLen3D(r.nodes[u].pos, r.nodes[v].pos)), path: []units.Vec2{r.nodes[v].pos}, ok: true}
			} else {
				if searches++; searches > MaxSearches {
					return nil, ErrSearchBudget
				}
				p, c, err := r.search.FindPath(grid, r.nodes[u].pos, r.nodes[v].pos, grid.W*grid.H)
				if err == ErrSearchBudget {
					return nil, err
				}
				cost := int64(c)
				if err == nil && sp.Projection != nil {
					cost = projectedPathCost(sp, r.nodes[u].pos, p, cost)
				}
				e = edge{cost: cost, path: p, ok: err == nil}
			}
			r.exact[[2]int{u, v}] = e
			refined = true
		}
		if !refined {
			return r.legs(path, grids), nil
		}
	}
}

func projectedPathCost(sp space.Space, start units.Vec2, path []units.Vec2, cost2D int64) int64 {
	var d2, d3 int64
	cur := start
	for _, next := range path {
		d2 += int64(next.Sub(cur).Len())
		d3 += int64(sp.SegmentLen3D(cur, next))
		cur = next
	}
	if d2 <= 0 {
		return cost2D
	}
	return cost2D * d3 / d2
}

func (r *Router) isCrossing(u, v int) bool { return r.nodes[u].partner == v }

// cost de la arista u→v: cruce de portal, tramo exacto o estimación.
func (r *Router) cost(spaces *ecs.Store[space.Space], u, v int) int64 {
	if r.isCrossing(u, v) {
		return r.nodes[u].cost
	}
	if e, ok := r.exact[[2]int{u, v}]; ok {
		if !e.ok {
			return inf
		}
		return e.cost
	}
	if sp, ok := spaces.Get(r.nodes[u].space); ok && sp.Projection != nil {
		return int64(sp.SegmentLen3D(r.nodes[u].pos, r.nodes[v].pos))
	}
	return int64(r.nodes[v].pos.Sub(r.nodes[u].pos).Len())
}

// dijkstra devuelve la secuencia de nodos de 0 (origen) a 1 (destino), o
// nil. Desempata por índice de nodo: determinista.
func (r *Router) dijkstra(spaces *ecs.Store[space.Space]) []int {
	n := len(r.nodes)
	r.dist = r.dist[:0]
	r.prev = r.prev[:0]
	r.done = r.done[:0]
	for range n {
		r.dist = append(r.dist, inf)
		r.prev = append(r.prev, -1)
		r.done = append(r.done, false)
	}
	r.dist[0] = 0
	relax := func(u, v int) {
		c := r.cost(spaces, u, v)
		if c == inf || r.done[v] {
			return
		}
		if d := r.dist[u] + c; d < r.dist[v] {
			r.dist[v], r.prev[v] = d, u
		}
	}
	for {
		u := -1
		for i := range n {
			if !r.done[i] && r.dist[i] != inf && (u < 0 || r.dist[i] < r.dist[u]) {
				u = i
			}
		}
		if u < 0 {
			return nil
		}
		if u == 1 {
			break
		}
		r.done[u] = true
		for _, v := range r.bySpace[r.nodes[u].space] {
			if v != u && v != 0 {
				relax(u, v)
			}
		}
		if p := r.nodes[u].partner; p >= 0 {
			relax(u, p)
		}
	}
	var path []int
	for v := 1; v >= 0; v = r.prev[v] {
		path = append(path, v)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

func (r *Router) legs(path []int, grids *ecs.Store[Grid]) []Leg {
	newLeg := func(s ids.ID) Leg {
		l := Leg{Space: s}
		if g := grids.Ptr(s); g != nil {
			l.GridVersion = g.Version
		}
		return l
	}
	out := []Leg{newLeg(r.nodes[0].space)}
	for k := 0; k+1 < len(path); k++ {
		u, v := path[k], path[k+1]
		cur := &out[len(out)-1]
		if r.isCrossing(u, v) {
			n := r.nodes[u]
			cur.Portal, cur.Reverse, cur.PortalVersion = n.portal, n.reverse, n.version
			out = append(out, newLeg(r.nodes[v].space))
			continue
		}
		cur.Path = append(cur.Path, r.exact[[2]int{u, v}].path...)
	}
	return out
}
