package main

import (
	"testing"
)

// En macro, cada tiny vivo cuenta en el globo de su superficie, los globos no
// se pisan y un clic en el del escritorio lleva a su vista de cerca.
func TestTinyBadges(t *testing.T) {
	g := artGame(t, morningScenario)
	desk := refOf(t, morningScenario, "desk1")
	for _, tick := range []int{30, 300, 530} {
		f := stepTo(t, g, tick)
		groups := tinyGroups(f, g.cam.Project(g.anchor(f), g.w, g.h))
		want, got := 0, 0
		for _, th := range f.Things {
			if isTiny(th) {
				want++
			}
		}
		for i, gr := range groups {
			got += len(gr.tinies)
			if i > 0 && groups[i-1].surface >= gr.surface {
				t.Fatalf("tick %d: grupos fuera de orden", tick)
			}
			for _, other := range groups[:i] {
				if gr.badge.Overlaps(other.badge) {
					t.Fatalf("tick %d: globos superpuestos %v %v", tick, gr.badge, other.badge)
				}
			}
			if gr.badge.Max.Y > gr.at.Y {
				t.Fatalf("tick %d: el globo no queda encima de los tinies", tick)
			}
		}
		if got != want || want == 0 {
			t.Fatalf("tick %d: %d tinies en globos, %d en el mundo", tick, got, want)
		}
	}

	g = artGame(t, morningScenario)
	f := stepTo(t, g, 30)
	g.art.badges = tinyGroups(f, g.cam.Project(g.anchor(f), g.w, g.h))
	var deskBadge *tinyGroup
	for i := range g.art.badges {
		if g.art.badges[i].surface == desk {
			deskBadge = &g.art.badges[i]
		}
	}
	if deskBadge == nil {
		t.Fatal("sin globo en el escritorio")
	}
	if !g.clickTinyBadge(deskBadge.badge.Min.Add(deskBadge.badge.Size().Div(2))) {
		t.Fatal("el clic no tocó el globo")
	}
	if g.art.focus != desk {
		t.Fatalf("el clic lleva a %v, no al escritorio", g.art.focus)
	}
	if g.clickTinyBadge(deskBadge.badge.Max.Add(deskBadge.badge.Size())) {
		t.Fatal("un clic fuera del globo no lo toca")
	}
}
