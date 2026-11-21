// Package spatial mantiene un índice de vecindad por Space: buckets regulares
// sobre las posiciones locales (comp.Location). Es estado derivado (un
// recurso del mundo): no se hashea ni se guarda, se reconstruye.
//
// Se reconstruye dos veces por tick: al empezar los sistemas (refleja los
// comandos aplicados) y al terminar la fase Move (refleja el movimiento).
// Entre medio puede estar desactualizado respecto de lo que se mueva en la
// misma fase.
package spatial

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/world"
)

// Bucket es el lado de cada bucket del índice.
const Bucket = 2 * units.Centimeter

type key struct {
	space  ids.ID
	bx, by int64
}

type item struct {
	id  ids.ID
	pos units.Vec2
}

// Index responde consultas de vecindad dentro de un Space.
type Index struct {
	buckets map[key][]item
}

func bucketOf(v units.Length) int64 {
	q := int64(v / Bucket)
	if v%Bucket != 0 && v < 0 {
		q--
	}
	return q
}

// Rebuild recalcula el índice desde los Locations.
func (ix *Index) Rebuild(locs interface {
	Each(func(ids.ID, *comp.Location))
}) {
	if ix.buckets == nil {
		ix.buckets = map[key][]item{}
	}
	clear(ix.buckets)
	locs.Each(func(e ids.ID, l *comp.Location) {
		k := key{l.Space, bucketOf(l.Pos.X), bucketOf(l.Pos.Y)}
		ix.buckets[k] = append(ix.buckets[k], item{e, l.Pos.XY()})
	})
}

// Near llama a fn con cada entidad de space a distancia <= radius de center
// (en el plano). El orden es determinista: por bucket (y, luego x) y dentro
// del bucket en orden del store de Location.
func (ix *Index) Near(space ids.ID, center units.Vec2, radius units.Length, fn func(ids.ID, units.Vec2)) {
	r2 := int64(radius) * int64(radius)
	for by := bucketOf(center.Y - radius); by <= bucketOf(center.Y+radius); by++ {
		for bx := bucketOf(center.X - radius); bx <= bucketOf(center.X+radius); bx++ {
			for _, it := range ix.buckets[key{space, bx, by}] {
				if it.pos.DistSq(center) <= r2 {
					fn(it.id, it.pos)
				}
			}
		}
	}
}

// Of devuelve el índice del mundo (instalado por Install).
func Of(w *world.World) *Index { return world.Resource[Index](w) }

type rebuild struct{ name string }

func (r rebuild) Name() string { return r.name }
func (rebuild) Run(w *world.World) {
	Of(w).Rebuild(comp.Of(w.ECS()).Location)
}

// Install registra el recurso Index y los sistemas que lo reconstruyen.
func Install(w *world.World) error {
	world.SetResource(w, &Index{})
	if err := w.Register(world.SystemSpec{Phase: clock.PhaseCommands, Order: -1 << 30}, rebuild{"spatial.pre"}); err != nil {
		return err
	}
	return w.Register(world.SystemSpec{Phase: clock.PhaseMove, Order: 1 << 30}, rebuild{"spatial.post"})
}
