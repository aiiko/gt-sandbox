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
		target := g.busiestSurface(f)
		center, n := tinyCluster(f, target)
		if n == 0 {
			t.Fatalf("seed %d: sin tinies", seed)
		}
		tab(g)
		if g.art.focus != target {
			t.Fatalf("seed %d: Tab lleva a %v, no a la superficie con tinies %v", seed, g.art.focus, target)
		}
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		// El grupo cabe en la media ventana alrededor de su centro: todo lo
		// que cae ahí (al menos los n del grupo) tiene que verse.
		inBox, inView := 0, 0
		for _, th := range f.Things {
			if !isTiny(th) || th.Space != target {
				continue
			}
			d := th.Pose.Pos.XY().Sub(center)
			if max(d.X, -d.X) > clusterHalfW || max(d.Y, -d.Y) > clusterHalfH {
				continue
			}
			inBox++
			x, y := proj.ToScreen(th.Pose.Pos)
			if x >= 0 && x < float32(g.w) && y >= artHeader && y < float32(g.h-artFooter) {
				inView++
			}
		}
		if inBox < n || inView != inBox {
			t.Fatalf("seed %d: grupo de %d, %d alrededor del centro, %d en pantalla", seed, n, inBox, inView)
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

// Con una habitación elegida (un clic en su piso en macro), Tab abre ese
// piso: encuadra a los tinies que haya en él o, si no hay, el punto donde se
// hizo clic, no el rincón fijo de la vista de cerca.
func TestTabFromRoomFramesFloor(t *testing.T) {
	g := artGame(t, morningScenario)
	f := stepTo(t, g, 1)
	kitchen := refOf(t, morningScenario, "kitchen")
	k, _ := f.SpaceByID(kitchen)
	click := k.Pose.Pos.XY().Add(units.Vec2{X: 2 * units.Meter, Y: units.Meter / 2})
	g.sel, g.art.clickOn, g.art.clickAt = kitchen, kitchen, click
	tab(g)
	if g.art.focus != kitchen || g.sel != kitchen {
		t.Fatalf("Tab abre %s con %s elegido, no la cocina", g.name(g.art.focus), g.name(g.sel))
	}
	if g.cam.X != float64(click.X-k.Pose.Pos.X) || g.cam.Y != float64(click.Y-k.Pose.Pos.Y) {
		t.Fatalf("encuadre %v,%v; se esperaba el clic", g.cam.X, g.cam.Y)
	}

	// En la mañana peligrosa, Tiny 21 empieza en el tatami del dormitorio.
	const descuido = "tests/scenarios/un_descuido_en_el_escritorio.json"
	g = artGame(t, descuido)
	f = stepTo(t, g, 1)
	bedroom := refOf(t, descuido, "bedroom")
	b, _ := f.SpaceByID(bedroom)
	g.sel, g.art.clickOn, g.art.clickAt = bedroom, bedroom, b.Pose.Pos.XY()
	tab(g)
	c, n := tinyCluster(f, bedroom)
	if g.art.focus != bedroom || n == 0 {
		t.Fatalf("Tab abre %s; %d tinies en el piso", g.name(g.art.focus), n)
	}
	if g.cam.X != float64(c.X-b.Pose.Pos.X) || g.cam.Y != float64(c.Y-b.Pose.Pos.Y) {
		t.Fatal("con tinies en el piso, Tab los encuadra a ellos y no al clic")
	}
}
