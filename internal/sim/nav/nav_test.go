package nav

import (
	"encoding/json"
	"math/rand/v2"
	"slices"
	"testing"

	"gtsandbox/internal/core/units"
)

func TestGridJSON(t *testing.T) {
	g, _ := NewGrid(units.V2(-5, 7), 10, 4, 2)
	g.Set(1, 0, Blocked)
	g.Set(3, 1, 5)
	b, _ := json.Marshal(g)
	if want := `{"origin":{"x":-5,"y":7},"cell":10,"version":2,"rows":[".#..","...5"]}`; string(b) != want {
		t.Fatalf("got %s", b)
	}
	var back Grid
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	b2, _ := json.Marshal(back)
	if string(b2) != string(b) {
		t.Fatalf("roundtrip: %s", b2)
	}
	if x, y := g.CellOf(units.V2(-6, 6)); x != -1 || y != -1 {
		t.Fatalf("CellOf con negativos: %d,%d", x, y)
	}
}

// reachable: BFS con las mismas reglas de movimiento que A*.
func reachable(g *Grid, sx, sy, tx, ty int) bool {
	seen := make([]bool, g.W*g.H)
	q := [][2]int{{sx, sy}}
	seen[sy*g.W+sx] = true
	for len(q) > 0 {
		c := q[0]
		q = q[1:]
		if c[0] == tx && c[1] == ty {
			return true
		}
		for _, d := range dirs {
			nx, ny := c[0]+d[0], c[1]+d[1]
			if g.At(nx, ny) == Blocked || seen[ny*g.W+nx] {
				continue
			}
			if d[0] != 0 && d[1] != 0 && (g.At(c[0]+d[0], c[1]) == Blocked || g.At(c[0], c[1]+d[1]) == Blocked) {
				continue
			}
			seen[ny*g.W+nx] = true
			q = append(q, [2]int{nx, ny})
		}
	}
	return false
}

// En grillas aleatorias: hay ruta si y solo si el destino es alcanzable, la
// ruta termina exactamente en el destino y ningún tramo recto pisa una celda
// bloqueada. El mismo input da la misma ruta.
func TestFindPathRandomGrids(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	var s Searcher
	const cell = 1000
	for range 300 {
		g, _ := NewGrid(units.V2(0, 0), cell, 5+r.IntN(25), 5+r.IntN(25))
		for range g.W * g.H / 3 {
			g.Set(r.IntN(g.W), r.IntN(g.H), Cost(r.IntN(4))) // ~25% bloqueadas
		}
		rnd := func() units.Vec2 {
			return units.V2(units.Length(r.IntN(g.W*cell)), units.Length(r.IntN(g.H*cell)))
		}
		from, to := rnd(), rnd()
		path, err := s.FindPath(g, from, to, 0)
		sx, sy := g.CellOf(from)
		tx, ty := g.CellOf(to)
		switch {
		case g.At(sx, sy) == Blocked || g.At(tx, ty) == Blocked:
			if err != ErrNotWalkable {
				t.Fatalf("esperaba ErrNotWalkable, got %v", err)
			}
			continue
		case !reachable(g, sx, sy, tx, ty):
			if err != ErrNoPath {
				t.Fatalf("esperaba ErrNoPath, got %v", err)
			}
			continue
		case err != nil:
			t.Fatalf("alcanzable pero %v", err)
		}
		if path[len(path)-1] != to {
			t.Fatalf("termina en %v, no en %v", path[len(path)-1], to)
		}
		prev := from
		for _, p := range path {
			d := p.Sub(prev)
			n := int64(d.Len())/(cell/10) + 1
			for i := int64(0); i <= n; i++ {
				q := prev.Add(units.V2(d.X.MulDiv(i, n), d.Y.MulDiv(i, n)))
				if !g.Walkable(q) {
					t.Fatalf("el tramo %v→%v pisa celda bloqueada en %v", prev, p, q)
				}
			}
			prev = p
		}
		again, _ := s.FindPath(g, from, to, 0)
		if !slices.Equal(path, again) {
			t.Fatal("A* no determinista")
		}
	}
}

func TestFindPathPrefersCheapCells(t *testing.T) {
	// Cruzar la columna cara cuesta más que rodearla por la fila 2.
	g := mustGrid(t, `{"origin":{"x":0,"y":0},"cell":10,"version":0,"rows":["..9..","..9..","....."]}`)
	var s Searcher
	path, err := s.FindPath(g, g.Center(0, 0), g.Center(4, 0), 0)
	if err != nil || !slices.Contains(path, g.Center(2, 2)) {
		t.Fatalf("path %v err %v", path, err)
	}
	if _, err := s.FindPath(g, g.Center(0, 0), g.Center(4, 0), 3); err != ErrSearchBudget {
		t.Fatalf("presupuesto: %v", err)
	}
}

func mustGrid(t *testing.T, js string) *Grid {
	t.Helper()
	var g Grid
	if err := json.Unmarshal([]byte(js), &g); err != nil {
		t.Fatal(err)
	}
	return &g
}
