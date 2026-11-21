package main

import (
	"image"
	"testing"
	"time"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/render"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/view"
)

// headlessHashes corre un escenario sin presentación y guarda el hash de cada
// tick (unos ticks más allá del final, porque a x16 el visor puede pasarse).
func headlessHashes(t *testing.T, path string, extra int) (map[clock.Tick]uint64, clock.Tick) {
	t.Helper()
	spec, err := scenario.ReadFile("../../" + path)
	if err != nil {
		t.Fatal(err)
	}
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	p, err := in.Player()
	if err != nil {
		t.Fatal(err)
	}
	hashes := map[clock.Tick]uint64{in.World.Tick(): in.World.Hash()}
	for range spec.Ticks + extra {
		if err := p.Step(); err != nil {
			t.Fatal(err)
		}
		hashes[in.World.Tick()] = in.World.Hash()
	}
	return hashes, in.World.Tick() - clock.Tick(extra)
}

// La presentación no cambia la simulación. Es la única prueba que compara
// con la ejecución headless: recorre la mañana completa con cada forma de
// mirarla y exige, en cada tick que alcanza, el mismo hash que sin pantalla.
//
//   - velocidad: x1, x16, paso a paso, frames irregulares mezclando pausas y
//     velocidades, y frames lentos que pierden tiempo (no ticks);
//   - vistas: el visor de debug con su cámara, y el arte en macro y micro con
//     interpolación, Arrange y picking en cada frame;
//   - seguimiento: seguir a un tiny por todo el viaje, la interfaz de partida,
//     enfocar desde la historia, el modo de debug y volver al encuadre general;
//   - sonido: en el arte, el mezclador observa la interfaz en cada frame.
func TestPresentationMatchesHeadless(t *testing.T) {
	const extra = 2 * view.MaxStepsPerFrame
	want, end := headlessHashes(t, morningScenario, extra)
	frame := time.Second / 60

	type drive func(g *game, i int)
	cases := map[string]struct {
		art   bool
		drive drive
	}{
		"x1":   {false, func(g *game, _ int) { g.run.Advance(frame) }},
		"x16":  {false, func(g *game, _ int) { g.run.Speed = 16; g.run.Advance(frame) }},
		"paso": {false, func(g *game, _ int) { g.run.Paused = true; g.run.StepOnce() }},
		"mezcla": {false, func(g *game, i int) {
			g.run.Speed = view.Speeds[i%3]
			g.run.Paused = i%7 == 0
			g.run.Advance(time.Duration(i%40) * time.Millisecond)
		}},
		"lento": {false, func(g *game, _ int) { g.run.Speed = 16; g.run.Advance(time.Second) }},
		"visor": {false, func(g *game, i int) {
			g.run.Speed = 4
			g.run.Advance(frame)
			switch i % 300 {
			case 50:
				g.sel = g.giantessID()
				g.toggleFollow()
			case 150:
				g.toggleMicro()
			case 250:
				g.toggleMicro()
				g.stopFollowing()
				g.fit()
			}
			f := g.frame()
			g.proj = g.cam.Project(g.anchor(f), g.w, g.h)
			f.Pick(units.V2(units.Meter, units.Meter), units.Centimeter)
		}},
		"vistas": {true, func(g *game, i int) {
			g.run.Speed = 4
			g.run.Advance(frame)
			if i%150 == 0 {
				g.sel = g.artDesk
				g.toggleMicro()
			}
			arrange(g)
		}},
		"seguir": {true, func(g *game, i int) {
			g.run.Advance(frame * 3)
			switch i {
			case 0:
				g.story = true
				g.startFollowing("Tiny 01")
			case 900:
				g.art.debug = true
			case 1000:
				if e := g.storyEntries(); len(e) > 0 {
					g.focusStory(e[len(e)-1])
				}
			case 1200:
				g.stopFollowing()
				if !g.art.focus.IsNil() {
					g.toggleArtMicro()
				}
				g.fit()
			}
			scene, f, focus := arrange(g)
			scene.Pick(image.Pt(240, 135))
			g.partida(f, focus)
		}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			g, err := newGame("../../"+morningScenario, 0, 0, "")
			if err != nil {
				t.Fatal(err)
			}
			g.w, g.h = 1280, 800
			if c.art {
				if err := g.enableArt(""); err != nil {
					t.Fatal(err)
				}
				g.w, g.h = (pixelGame{g}).Layout(960, 540)
			}
			g.fit()
			var sounds *soundRecorder
			if c.art {
				sounds = withSound(t, g)
			}
			for i := 0; g.run.World.Tick() < end; i++ {
				c.drive(g, i)
				g.observeSound()
				if g.run.Err != nil {
					t.Fatal(g.run.Err)
				}
				tick := g.run.World.Tick()
				h, ok := want[tick]
				if !ok {
					t.Fatalf("tick %d fuera de la ejecución headless", tick)
				}
				if got := g.run.Hash(); got != h {
					t.Fatalf("tick %d: hash %#x, headless %#x", tick, got, h)
				}
				if i > 1_000_000 {
					t.Fatal("no avanza")
				}
			}
			if sounds != nil && len(sounds.played) == 0 {
				t.Error("la interfaz no sonó")
			}
		})
	}
}

// arrange coloca la escena de arte del frame actual, como el Draw del juego.
func arrange(g *game) (render.ArtScene, *view.Frame, ids.ID) {
	f := g.frame()
	focus := g.artFocus(f)
	scale := render.Macro
	if !focus.IsNil() {
		scale = render.Micro
	}
	g.proj = g.cam.Project(g.anchor(f), g.w, g.h)
	return g.art.atlas.Arrange(f, g.proj, g.visuals, scale, focus, g.sel, g.follow), f, focus
}
