package main

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"gtsandbox/internal/sim/clock"
)

// benchmark mide la presentación: corre el juego a x4 sin vsync durante una
// duración fija y cuenta los frames dibujados. El costo por frame sale del
// total (tiempo / frames), porque el reloj de Windows no resuelve llamadas de
// menos de un milisegundo. La simulación se mide aparte, en lote y sin
// pantalla, con cmd/simbench.
type benchmark struct {
	duration time.Duration
	start    time.Time
	frames   int
	tick0    clock.Tick
}

func (b *benchmark) update(g *game) (done bool) {
	if b.start.IsZero() {
		b.start, b.tick0 = time.Now(), g.run.World.Tick()
	}
	if time.Since(b.start) >= b.duration {
		b.report(g)
		return true
	}
	g.run.Advance(time.Second / time.Duration(ebiten.TPS()))
	return false
}

func (b *benchmark) draw() {
	if !b.start.IsZero() {
		b.frames++
	}
}

func (b *benchmark) report(g *game) {
	elapsed := time.Since(b.start)
	w, h := ebiten.WindowSize()
	fps := float64(b.frames) / elapsed.Seconds()
	fmt.Printf("bench %s: ventana %dx%d, %.1f s, ticks %d→%d\n", g.title, w, h, elapsed.Seconds(), b.tick0, g.run.World.Tick())
	fmt.Printf("  presentación: %.1f FPS sin vsync, %.3f ms/frame (%d frames)\n", fps, 1000/fps, b.frames)
}
