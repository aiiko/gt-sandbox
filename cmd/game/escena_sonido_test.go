package main

import (
	"slices"
	"strings"
	"testing"
	"time"

	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sound"
	"gtsandbox/internal/view"
)

// heard es un sonido con el tick del frame en que sonó.
type heard struct {
	tick clock.Tick
	id   string
}

// listening recorre un escenario de a un tick, observando el sonido en cada
// frame como el juego, y verifica que los eventos que tienen sonido suenen en
// el mismo frame en que se publican (sameFrame da el sonido, o su prefijo,
// que corresponde a un evento). before corre antes de cada paso.
type listening struct {
	g      *game
	r      *soundRecorder
	heard  []heard
	counts map[string]int
}

func listen(t *testing.T, g *game, until clock.Tick, before func(clock.Tick), sameFrame func(e view.Published, cur *view.Frame) string) *listening {
	t.Helper()
	l := &listening{g: g, r: withSound(t, g), counts: map[string]int{}}
	var seq uint64
	for g.run.Cur.Tick < until {
		if before != nil {
			before(g.run.Cur.Tick)
		}
		n := len(l.r.played)
		g.run.StepOnce()
		if g.run.Err != nil {
			t.Fatal(g.run.Err)
		}
		g.observeSound()
		now := l.r.played[n:]
		for _, id := range now {
			l.heard = append(l.heard, heard{g.run.Cur.Tick, id})
			l.counts[id]++
		}
		for _, e := range g.run.Events {
			if e.Seq <= seq {
				continue
			}
			seq = e.Seq
			if want := sameFrame(e, g.run.Cur); want != "" && !slices.ContainsFunc(now, func(id string) bool { return strings.HasPrefix(id, want) }) {
				t.Errorf("tick %d: %s no sonó con %s (sonó %v)", g.run.Cur.Tick, want, e.Event.Kind(), now)
			}
		}
	}
	return l
}

func (l *listening) expect(t *testing.T, view string, ids ...string) {
	t.Helper()
	for _, id := range ids {
		if l.counts[id] == 0 {
			t.Errorf("%s: nunca sonó %s", view, id)
		}
	}
}

func (l *listening) forbid(t *testing.T, view string, prefixes ...string) {
	t.Helper()
	for id := range l.counts {
		for _, p := range prefixes {
			if strings.HasPrefix(id, p) {
				t.Errorf("%s: no debería sonar %s", view, id)
			}
		}
	}
}

// first es el tick en que sonó id por primera vez después de from (o el
// máximo si nunca).
func (l *listening) first(id string, from ...clock.Tick) clock.Tick {
	after := clock.Tick(0)
	if len(from) > 0 {
		after = from[0] + 1
	}
	for _, h := range l.heard {
		if h.id == id && h.tick >= after {
			return h.tick
		}
	}
	return never
}

const never = clock.Tick(1 << 62)

// Aceptación de M21: la mañana suena con perspectiva por escala, y cada
// momento clave suena en el frame de su evento.
func TestMorningSounds(t *testing.T) {
	end := clock.Tick(900)

	// Macro: la giantess y la casa. Los tinies no se oyen.
	g := artGame(t, morningScenario)
	l := listen(t, g, end, nil, func(e view.Published, cur *view.Frame) string {
		switch ev := e.Event.(type) {
		case attach.ContactBegan:
			return sound.HandTap
		case space.EntityReparented:
			if _, ok := cur.SpaceByID(ev.Entity); ok && ev.Cause == "portal" {
				return "door_"
			}
		}
		if e.Event.Kind() == "item_dropped" {
			return sound.CupClink
		}
		return ""
	})
	l.expect(t, "macro", sound.DoorNoren, sound.FridgeOpen, sound.FridgeClose, sound.CupClink, sound.Bite,
		sound.HandTap, sound.DoorFusuma, sound.Zabuton, sound.TVOn, sound.Emote+view.EmoteNote,
		sound.StepTatami, sound.StepTile, sound.StepWood)
	l.forbid(t, "macro", "tiny_", "hand_land", "pass_", "crumb_")

	// Micro del escritorio: el desayuno visto desde la colonia.
	g = artGame(t, morningScenario)
	g.sel = g.artDesk
	g.toggleArtMicro()
	desk := g.artDesk
	l = listen(t, g, end, nil, func(e view.Published, _ *view.Frame) string {
		switch ev := e.Event.(type) {
		case attach.ContactBegan:
			if ev.A == desk || ev.B == desk {
				return sound.HandLand
			}
		case attach.PortalOpened:
			if ev.A == desk || ev.B == desk {
				return sound.PassOpen
			}
		}
		return ""
	})
	l.expect(t, "escritorio", sound.CrumbTick, sound.HandNear, sound.HandLand, sound.PassOpen,
		sound.Emote+view.EmoteDiscover, sound.CrumbPop, sound.NestStore, sound.TinyHop, sound.CupClink)
	l.forbid(t, "escritorio", "giantess_step", "door_", "fridge_", "tv_on", "zabuton")
	if !(l.first(sound.HandNear) < l.first(sound.HandLand) && l.first(sound.Emote+view.EmoteDiscover) < l.first(sound.TinyHop)) {
		t.Errorf("orden: la mano se acerca (%d) antes de apoyarse (%d); el descubrimiento (%d) antes del ascenso (%d)",
			l.first(sound.HandNear), l.first(sound.HandLand), l.first(sound.Emote+view.EmoteDiscover), l.first(sound.TinyHop))
	}
	// Veinte tinies caminando no superan el límite de cada variante de pasitos.
	lib := g.sound.Lib
	last := map[string]int{}
	for i, h := range l.heard {
		if !strings.HasPrefix(h.id, sound.TinyStep) {
			continue
		}
		d, _ := lib.Def(h.id)
		if j, ok := last[h.id]; ok && int64(h.tick-l.heard[j].tick) < int64(d.CooldownMS)*sound.FPS/1000 {
			t.Fatalf("%s repetido a los %d frames", h.id, h.tick-l.heard[j].tick)
		}
		last[h.id] = i
	}

	// Siguiendo a Tiny 01: el ascenso, el viaje en la palma y la llegada.
	g = artGame(t, morningScenario)
	l = listen(t, g, end, func(tick clock.Tick) {
		if tick == 180 {
			g.startFollowing("Tiny 01")
		}
	}, func(e view.Published, _ *view.Frame) string {
		if _, ok := e.Event.(attach.ContactEnded); ok {
			return sound.HandLift // Tiny 01 viaja en la palma: la vista es la palma
		}
		return ""
	})
	l.expect(t, "seguir", sound.UIFollow, sound.TinyHop, sound.Emote+view.EmoteSweat, sound.HandLift,
		sound.PassClose, sound.HandLand, sound.PassOpen)
	lift := l.first(sound.HandLift)
	land := l.first(sound.HandLand, lift)
	if lift == never || land == never {
		t.Fatalf("el viaje no sonó: levantar %d, apoyar %d", lift, land)
	}
	if l.first(sound.TinyHop, land) == never {
		t.Error("Tiny 01 no sonó al bajar en el living")
	}

	// La caída: el silbido al caer y el «puf» al terminar la animación.
	g = artGame(t, fallScenario)
	l = listen(t, g, 40, func(clock.Tick) {
		if g.follow.IsNil() {
			g.startFollowing("Tiny 01")
		}
	}, func(e view.Published, cur *view.Frame) string {
		if ev, ok := e.Event.(space.EntityFell); ok {
			if th, ok := cur.ThingByID(ev.Entity); ok && th.Mover != nil {
				return sound.TinyFall
			}
		}
		return ""
	})
	fall, landed := l.first(sound.TinyFall), l.first(sound.TinyLand)
	if fall == never || landed != fall+view.FallTicks-1 {
		t.Errorf("caída en %d, aterrizaje en %d", fall, landed)
	}
}

// A x4 y x16 se omiten los sonidos que se repiten y quedan los que cuentan
// algo; en pausa no avanza nada y no suena la escena.
func TestFastAndPausedSounds(t *testing.T) {
	g := artGame(t, morningScenario)
	g.sel = g.artDesk
	g.toggleArtMicro()
	r := withSound(t, g)
	g.run.Speed = 16
	for g.run.Cur.Tick < 400 {
		g.run.StepOnce()
		g.observeSound()
	}
	for _, id := range r.played {
		if strings.HasPrefix(id, sound.TinyStep) || id == sound.CrumbPop {
			t.Errorf("a x16 sonó %s", id)
		}
	}
	if !slices.Contains(r.played, sound.HandLand) {
		t.Error("a x16 la mano apoyada debe sonar")
	}
	g.run.Paused = true
	g.observeSound()
	n := len(r.played)
	for range 60 {
		g.run.Advance(time.Second / 60)
		g.observeSound()
	}
	if len(r.played) != n {
		t.Errorf("en pausa sonó %v", r.played[n:])
	}
}
