package main

import (
	"slices"
	"strings"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/ui"
	"gtsandbox/internal/view"
)

// Recorre la mañana completa en la interfaz de partida y verifica que los
// momentos clave se vean con su símbolo: el descubrimiento de las migajas
// («!»), el ascenso a la palma (gota), la exploración («?»), el descanso en la
// taza («zzz») y la TV (nota); y que el retrato cambie con la giantess.
func TestPartidaShowsMorningCues(t *testing.T) {
	g := artGame(t, morningScenario)
	g.story = true // el panel de historia con el retrato
	giantess := refOf(t, morningScenario, "giantess")
	palm := refOf(t, morningScenario, "palm")
	spec, err := scenario.ReadFile("../../" + morningScenario)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]sighting{}
	faces := map[string]int{}
	ascents := 0
	var prev *view.Frame
	for range spec.Ticks {
		f := stepTo(t, g, int(g.run.Cur.Tick)+1)
		for _, th := range f.Things {
			if e := f.Emote(th.ID); e != "" {
				if _, ok := seen[e]; !ok {
					seen[e] = sighting{f.Tick, th.Name}
				}
			}
			// subir a la palma muestra la gota en ese mismo frame
			if prev != nil && th.Space == palm {
				if p, ok := prev.ThingByID(th.ID); ok && p.Space != palm && th.Needs != nil {
					ascents++
					if f.Emote(th.ID) != view.EmoteSweat {
						t.Errorf("tick %d: %s sube a la palma con %q", f.Tick, th.Name, f.Emote(th.ID))
					}
				}
			}
		}
		if e := f.Emote(giantess); e != "" {
			if _, ok := seen["giantess:"+e]; !ok {
				seen["giantess:"+e] = sighting{f.Tick, "giantess"}
			}
		}
		h := g.partida(f, ids.Nil)
		faces[h.Face]++
		if h.Clock == "" || len(h.Controls) == 0 || len(h.Location) == 0 {
			t.Fatalf("tick %d: interfaz incompleta %+v", f.Tick, h)
		}
		prev = f
	}
	for _, k := range []string{view.EmoteDiscover, view.EmoteSweat, view.EmoteExplore, view.EmoteSleep, "giantess:" + view.EmoteNote} {
		if _, ok := seen[k]; !ok {
			t.Errorf("la mañana nunca mostró %s: %v", k, seen)
		}
	}
	if seen[view.EmoteDiscover].tick > seen[view.EmoteSweat].tick {
		t.Errorf("el descubrimiento de las migajas (%v) debería ir antes del ascenso (%v)", seen[view.EmoteDiscover], seen[view.EmoteSweat])
	}
	if ascents < 3 {
		t.Errorf("solo %d ascensos a la palma", ascents)
	}
	for _, face := range []string{view.FaceHungry, view.FaceHappy, view.FaceSurprised, view.FaceNeutral} {
		if faces[face] == 0 {
			t.Errorf("el retrato nunca mostró %s: %v", face, faces)
		}
	}
}

type sighting struct {
	tick clock.Tick
	who  string
}

// La ubicación es jerárquica en micro («Dormitorio › Escritorio», «Giantess ›
// Palma») y en macro nombra la habitación de lo seguido; el inspector
// resumido trae lugar, actividad en castellano, necesidades y la orden.
func TestPartidaLocationAndCard(t *testing.T) {
	g := artGame(t, morningScenario)
	palm := refOf(t, morningScenario, "palm")
	f := stepTo(t, g, 300)
	if got := g.location(f, g.artDesk); !slices.Equal(got, []string{"Dormitorio", "Escritorio"}) {
		t.Errorf("micro del escritorio: %v", got)
	}
	if got := g.location(f, palm); !slices.Equal(got, []string{"Giantess", "Palma"}) {
		t.Errorf("micro de la palma: %v", got)
	}
	g.startFollowing("Tiny 05")
	if g.follow.IsNil() {
		t.Fatal("no sigue a Tiny 05")
	}
	g.toggleArtMicro() // volver a macro siguiendo
	if got := g.location(f, ids.Nil); !slices.Equal(got, []string{"Casa", "Dormitorio"}) {
		t.Errorf("macro siguiendo: %v", got)
	}
	if got := strings.Join(g.location(f, g.artDesk), ui.Separator); got != "Dormitorio › Escritorio" {
		t.Errorf("rótulo %q", got)
	}
	c := g.card(f, g.follow)
	if c == nil || c.Name != "Tiny 05" || c.Place != "Dormitorio › Escritorio" || c.Activity == "" || c.Order == "" || len(c.Needs) == 0 {
		t.Fatalf("inspector %+v", c)
	}
	for _, n := range c.Needs {
		if n.Label != "Hambre" && n.Label != "Energía" {
			t.Errorf("necesidad sin rótulo en castellano: %+v", n)
		}
	}
	if g.card(f, g.giantessID()) == nil {
		t.Error("la giantess no tiene inspector")
	}
}

func TestActivityText(t *testing.T) {
	for _, c := range []struct{ activity, step, carrying, want string }{
		{"use:breakfast", "goto", "", "Va a desayunar"},
		{"use:breakfast", "use", "", "Desayuna"},
		{"use:watch", "", "", "Mira televisión"},
		{"work", "", "una migaja", "Acarrea una migaja"},
		{"work (suspendida)", "", "", "Trabaja para la colonia (en pausa)"},
		{"explore", "", "", "Explora"},
		{"flee", "", "", "Huye"},
		{"dance", "", "", "dance"},
	} {
		if got := activityText(c.activity, c.step, c.carrying); got != c.want {
			t.Errorf("%s/%s: %q, quería %q", c.activity, c.step, got, c.want)
		}
	}
	if clockAt(0) != "07:00" || clockAt(20*75) != "08:15" {
		t.Errorf("reloj %s %s", clockAt(0), clockAt(20*75))
	}
}

// Desde una entrada de la historia se enfoca al actor (siguiéndolo en la vista
// de su superficie) o, si ya no está, al lugar.
func TestPartidaStoryFocus(t *testing.T) {
	g := artGame(t, morningScenario)
	palm := refOf(t, morningScenario, "palm")
	stepTo(t, g, 300)
	var climb, start view.StoryEntry
	for _, e := range g.run.Story {
		if strings.HasSuffix(e.Text, "sube a la mano.") && climb.Actor.IsNil() {
			climb = e
		}
		if strings.Contains(e.Text, "desayuno") {
			start = e
		}
	}
	if climb.Actor.IsNil() || climb.Place != palm || start.Actor.IsNil() {
		t.Fatalf("faltan entradas: %+v %+v", climb, start)
	}
	g.focusStory(climb)
	f := g.run.Cur
	if g.follow != climb.Actor || g.artFocus(f) != palm {
		t.Fatalf("siguiendo %s en %s", g.name(g.follow), g.name(g.artFocus(f)))
	}
	h := g.partida(f, g.artFocus(f))
	if !slices.Equal(h.Location, []string{"Giantess", "Palma"}) || h.Following == "" {
		t.Errorf("interfaz %+v", h)
	}
	// la giantess no tiene vista micro: se la sigue en macro
	g.focusStory(start)
	if g.follow != start.Actor || !g.art.focus.IsNil() {
		t.Errorf("siguiendo %s, foco %s", g.name(g.follow), g.name(g.art.focus))
	}
	g.story = true
	if !slices.ContainsFunc(g.partida(g.run.Cur, ids.Nil).Story, func(l ui.StoryLine) bool { return l.Focusable }) {
		t.Error("ninguna entrada se puede enfocar")
	}
}

// Tab nombra su destino, que es el mismo al que lleva: donde hay más tinies
// (al empezar, el escritorio), la superficie elegida (una habitación es su
// piso), la de lo seguido, y de cerca, la casa.
func TestTabNamesItsDestination(t *testing.T) {
	g := artGame(t, morningScenario)
	f := stepTo(t, g, 1)
	tab := func() string {
		for _, c := range g.partida(f, g.art.focus).Controls {
			if c.Key == "Tab" {
				return c.Label
			}
		}
		return ""
	}
	for _, c := range []struct {
		sel, follow, want string
		target            ids.ID
	}{
		{"", "", "Escritorio", g.artDesk},
		{"bedroom", "", "Dormitorio", refOf(t, morningScenario, "bedroom")},
		{"kitchen", "", "Cocina", refOf(t, morningScenario, "kitchen")},
		{"giantess", "", "Escritorio", g.artDesk},
		{"bedroom", "tiny01", "Escritorio", g.artDesk},
		{"desk2", "", "Mesa del living", refOf(t, morningScenario, "desk2")},
	} {
		g.sel, g.follow = ids.Nil, ids.Nil
		if c.sel != "" {
			g.sel = refOf(t, morningScenario, c.sel)
		}
		if c.follow != "" {
			g.follow = refOf(t, morningScenario, c.follow)
		}
		if got := tab(); got != c.want || g.microTarget(f) != c.target {
			t.Errorf("elegido %q, seguido %q: Tab dice %q y lleva a %v; se esperaba %q", c.sel, c.follow, got, g.microTarget(f), c.want)
		}
	}
	g.sel, g.follow = refOf(t, morningScenario, "bedroom"), ids.Nil
	g.toggleArtMicro()
	if g.art.focus != refOf(t, morningScenario, "bedroom") || tab() != "Casa" {
		t.Fatalf("de cerca: foco %v, Tab dice %q", g.art.focus, tab())
	}
}

// La barra superior muestra la población viva de tinies y, al morir uno,
// aparece una notificación temporal con la causa del deceso.
func TestPopulationAndDeathNotice(t *testing.T) {
	const sc = "tests/scenarios/un_descuido_en_el_escritorio.json"
	g := artGame(t, sc)
	f0 := g.run.Cur
	h0 := g.partida(f0, ids.Nil)
	if !h0.ShowPop || h0.Population == 0 {
		t.Fatalf("población inicial inválida: %+v", h0)
	}
	if len(h0.Notices) != 0 {
		t.Fatalf("no debería haber notificaciones al inicio: %+v", h0.Notices)
	}
	var deathTick clock.Tick
	var sawNotice ui.Notice
	for tick := 1; tick <= 600; tick++ {
		f := stepTo(t, g, tick)
		h := g.partida(f, ids.Nil)
		if len(h.Notices) > 0 {
			deathTick = f.Tick
			sawNotice = h.Notices[0]
			if h.Population >= h0.Population {
				t.Fatalf("tick %d: murió un tiny pero la población siguió en %d", tick, h.Population)
			}
			break
		}
	}
	if deathTick == 0 {
		t.Fatal("no apareció ninguna notificación de deceso en un_descuido_en_el_escritorio")
	}
	if !strings.Contains(sawNotice.Title, "falleció") || !strings.Contains(sawNotice.Cause, "Aplastado") {
		t.Errorf("notificación inesperada: %+v", sawNotice)
	}
	if !sawNotice.Focusable {
		t.Errorf("la notificación debería permitir enfocar los restos: %+v", sawNotice)
	}
	// Sigue visible tras 5 s reales, se puede cerrar manualmente y expira tras deathNoticeDuration.
	g.updateNotices(deathNoticeDuration / 2)
	before := len(g.partida(g.run.Cur, ids.Nil).Notices)
	if before == 0 {
		t.Error("la notificación desapareció antes de la mitad de su duración real")
	}
	g.dismissNotice(0)
	if after := len(g.partida(g.run.Cur, ids.Nil).Notices); after != before-1 {
		t.Errorf("dismissNotice dejó %d notificaciones, se esperaban %d", after, before-1)
	}
	g.updateNotices(deathNoticeDuration)
	for _, n := range g.partida(g.run.Cur, ids.Nil).Notices {
		if n.Title == sawNotice.Title {
			t.Errorf("la notificación %q no expiró tras %v reales", n.Title, deathNoticeDuration)
		}
	}
}
