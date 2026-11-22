package main

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/view"
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

// tab hace lo que la tecla Tab en el arte.
func tab(g *game) {
	g.toggleMicro()
	g.focusTinies()
}

// Tab sin nada elegido lleva a la superficie con más tinies y encuadra su
// grupo más numeroso: todos sus tinies quedan en pantalla, entre las barras.
func TestTabFramesTinies(t *testing.T) {
	t.Chdir("../..")
	for _, seed := range []uint64{1, 5, 9} {
		g, err := newRandomGame(seed)
		if err != nil {
			t.Fatal(err)
		}
		if err := g.enableArt(""); err != nil {
			t.Fatal(err)
		}
		g.w, g.h = (pixelGame{g}).Layout(960, 540)
		g.fit()
		f := stepTo(t, g, 20)
		target := g.busiestSurface(f, ids.Nil)
		center, n := tinyCluster(f, target)
		if n == 0 {
			t.Fatalf("seed %d: sin tinies", seed)
		}
		tab(g)
		if g.art.focus != target {
			t.Fatalf("seed %d: Tab lleva a %v, no a la superficie con tinies %v", seed, g.art.focus, target)
		}
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		inView := 0
		for _, th := range f.Things {
			if !isTiny(th) || th.Space != target {
				continue
			}
			d := th.Pose.Pos.XY().Sub(center)
			if max(d.X, -d.X) > clusterHalfW || max(d.Y, -d.Y) > clusterHalfH {
				continue
			}
			x, y := proj.ToScreen(th.Pose.Pos)
			if x >= 0 && x < float32(g.w) && y >= artHeader && y < float32(g.h-artFooter) {
				inView++
			}
		}
		if inView != n {
			t.Fatalf("seed %d: %d de %d tinies del grupo en pantalla", seed, inView, n)
		}
	}
}

// El grupo más numeroso gana aunque haya tinies sueltos lejos.
func TestTinyClusterPicksDensest(t *testing.T) {
	desk := ids.ID(7)
	tiny := func(x, y units.Length) view.Thing {
		return view.Thing{Space: desk, Needs: []view.Need{{}}, Pose: view.Pose{Pos: units.V3(x, y, 0)}}
	}
	f := &view.Frame{Things: []view.Thing{
		tiny(0, 0),
		tiny(50*units.Centimeter, 0), tiny(52*units.Centimeter, units.Centimeter), tiny(54*units.Centimeter, 0),
		{Space: desk, Pose: view.Pose{Pos: units.V3(0, 0, 0)}}, // una migaja no cuenta
	}}
	c, n := tinyCluster(f, desk)
	if n != 3 || c.X != 52*units.Centimeter {
		t.Fatalf("centro %v con %d", c, n)
	}
	if _, n := tinyCluster(f, ids.ID(8)); n != 0 {
		t.Fatal("otra superficie no tiene tinies")
	}
}

// Una habitación elegida (un clic en su piso en macro) no lleva a un piso
// vacío: Tab abre donde hay tinies dentro de ella o, si no hay, en la casa, y
// suelta la habitación para que no quede su ficha abierta.
func TestTabFromRoomGoesToTinies(t *testing.T) {
	for _, room := range []string{"bedroom", "kitchen"} {
		g := artGame(t, morningScenario)
		f := stepTo(t, g, 1)
		g.sel = refOf(t, morningScenario, room)
		tab(g)
		if g.art.focus != g.artDesk {
			t.Fatalf("%s: Tab abre %s, no el escritorio", room, g.name(g.art.focus))
		}
		if !g.sel.IsNil() {
			t.Fatalf("%s: la habitación sigue elegida", room)
		}
		if _, n := tinyCluster(f, g.art.focus); n == 0 {
			t.Fatalf("%s: sin tinies en la superficie abierta", room)
		}
	}
}
