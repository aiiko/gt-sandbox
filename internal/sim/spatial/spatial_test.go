package spatial

import (
	"math/rand/v2"
	"slices"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
)

// Near devuelve exactamente lo mismo que la fuerza bruta.
func TestNearMatchesBruteForce(t *testing.T) {
	r := ecs.NewRegistry()
	s, _ := comp.Register(r)
	spaces := []ids.ID{r.Spawn(), r.Spawn()}
	rnd := rand.New(rand.NewPCG(3, 4))
	coord := func() units.Length { return units.Length(rnd.IntN(int(20*units.Centimeter))) - 10*units.Centimeter }
	for range 500 {
		s.Location.Set(r.Spawn(), comp.Location{Space: spaces[rnd.IntN(2)], Pos: units.V3(coord(), coord(), 0)})
	}
	var ix Index
	ix.Rebuild(s.Location)
	for range 200 {
		sp, c, rad := spaces[rnd.IntN(2)], units.V2(coord(), coord()), units.Length(rnd.IntN(int(5*units.Centimeter)))
		var got, want []ids.ID
		ix.Near(sp, c, rad, func(e ids.ID, _ units.Vec2) { got = append(got, e) })
		s.Location.Each(func(e ids.ID, l *comp.Location) {
			if l.Space == sp && l.Pos.XY().DistSq(c) <= int64(rad)*int64(rad) {
				want = append(want, e)
			}
		})
		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Fatalf("Near(%v, %v): got %v, want %v", c, rad, got, want)
		}
	}
}
