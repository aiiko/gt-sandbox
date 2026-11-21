// game es el visor de debug (M6): corre la simulación a 20 TPS, la dibuja a
// 60 FPS interpolando, y permite inspeccionar y mandar comandos.
//
//	go run ./cmd/game                                   # escenario de M5
//	go run ./cmd/game -scenario tests/scenarios/floor_to_desk_by_leg.json
//	go run ./cmd/game -crowd 200 -seed 7
//	go run ./cmd/game -load s.json
//
// Controles: Espacio pausa | N paso | 1/2/3 velocidad x1/x4/x16 | Tab vista
// micro del Space elegido / macro | F encuadrar | rueda zoom | arrastrar o
// flechas para mover | click elegir | click derecho mover lo elegido ·
// G/P/C/R/L overlays (grillas, portales, contactos, rutas, nombres) | Esc
// soltar.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/render"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/ui"
	"gtsandbox/internal/view"
)

func main() {
	scen := flag.String("scenario", "tests/scenarios/giantess_carries_tiny.json", "escenario JSON a reproducir")
	crowd := flag.Int("crowd", 0, "en vez de un escenario, n tinies caminando por un escritorio")
	seed := flag.Uint64("seed", 1, "semilla para -crowd")
	load := flag.String("load", "", "en vez de un escenario, continuar un snapshot JSON")
	shot := flag.String("shot", "", "guardar una captura PNG al llegar a -shot-tick y salir")
	micro := flag.String("micro", "", "empezar en la vista micro del Space con este nombre")
	shotTick := flag.Int("shot-tick", 0, "tick del escenario (o del mundo) para -shot; corre a x16")
	flag.Parse()

	g, err := newGame(*scen, *crowd, *seed, *load)
	if err != nil {
		fmt.Fprintln(os.Stderr, "game:", err)
		os.Exit(1)
	}
	g.micro = *micro
	if *shot != "" {
		g.shot, g.shotTick = *shot, clock.Tick(*shotTick)+g.run.World.Tick()
		g.run.Speed = view.Speeds[len(view.Speeds)-1]
	}
	ebiten.SetWindowTitle("gt-sandbox | " + g.title)
	ebiten.SetWindowSize(1280, 800)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintln(os.Stderr, "game:", err)
		os.Exit(1)
	}
}

type game struct {
	title  string
	run    *view.Runner
	status func() string // estado del escenario, para el HUD

	cam     camera.Camera
	proj    camera.Projector
	layers  render.Layers
	sel     ids.ID
	w, h    int
	fitted  bool
	drag    bool
	pressAt [2]int
	last    [2]int

	micro    string // -micro
	shot     string // captura pedida por -shot
	shotTick clock.Tick
	shotDone bool
}

func newGame(scen string, crowd int, seed uint64, load string) (*game, error) {
	g := &game{layers: render.AllLayers, cam: camera.Camera{Zoom: 1e-3}}
	switch {
	case load != "":
		f, err := os.Open(load)
		if err != nil {
			return nil, err
		}
		w, err := engine.Load(f)
		f.Close()
		if err != nil {
			return nil, err
		}
		g.title = load
		g.run = view.NewRunner(w, func() error { w.Step(); return w.CheckInvariants() })
		g.status = func() string { return "" }
	case crowd > 0:
		c, err := scenario.NewCrowd(crowd, seed)
		if err != nil {
			return nil, err
		}
		g.title = fmt.Sprintf("crowd %d", crowd)
		g.run = view.NewRunner(c.World, func() error { c.Step(); return nil })
		g.status = func() string { return "" }
	default:
		spec, err := scenario.ReadFile(scen)
		if err != nil {
			return nil, err
		}
		in, err := scenario.Build(spec)
		if err != nil {
			return nil, err
		}
		p, err := in.Player()
		if err != nil {
			return nil, err
		}
		g.title = spec.Name
		g.run = view.NewRunner(in.World, p.Step)
		var verdict string
		g.status = func() string {
			if !p.Done() {
				return fmt.Sprintf("escenario %d/%d", p.Tick(), spec.Ticks)
			}
			if verdict == "" {
				fails, err := in.Check()
				switch {
				case err != nil:
					verdict = "error: " + err.Error()
				case len(fails) > 0:
					verdict = fmt.Sprintf("%d expectativas FALLAN: %s", len(fails), fails[0])
				default:
					verdict = fmt.Sprintf("%d expectativas OK", len(spec.Expect))
				}
			}
			return fmt.Sprintf("escenario %d/%d terminado | %s", p.Tick(), spec.Ticks, verdict)
		}
	}
	return g, nil
}

func (g *game) Layout(w, h int) (int, int) { return w, h }

// anchor devuelve la pose del ancla de la cámara; si el ancla ya no existe,
// vuelve a la vista macro.
func (g *game) anchor(f *view.Frame) fixedmath.Transform {
	if !g.cam.Anchor.IsNil() {
		if s, ok := f.SpaceByID(g.cam.Anchor); ok {
			return s.Pose.Transform()
		}
		g.cam.Follow(ids.Nil, fixedmath.Identity, g.proj)
	}
	return fixedmath.Identity
}

func (g *game) frame() *view.Frame {
	return view.Interpolate(g.run.Prev, g.run.Cur, g.run.Alpha())
}

func (g *game) fit() {
	f := g.run.Cur
	if s, ok := f.SpaceByID(g.cam.Anchor); ok && !s.Extent.Empty() {
		g.cam.Fit(s.Extent, g.w, g.h)
		return
	}
	g.cam.Anchor = ids.Nil
	b := f.Bounds()
	if b.Empty() {
		b = units.RectWH(-units.Meter/2, -units.Meter/2, units.Meter, units.Meter)
	}
	g.cam.Fit(b, g.w, g.h)
}

func (g *game) Update() error {
	if g.w == 0 {
		return nil // todavía sin Layout
	}
	if !g.fitted {
		g.fit()
		g.fitted = true
		for _, sp := range g.run.Cur.Spaces {
			if g.micro != "" && sp.Name == g.micro {
				g.sel = sp.ID
				g.proj = g.cam.Project(fixedmath.Identity, g.w, g.h)
				g.toggleMicro()
			}
		}
	}
	g.proj = g.cam.Project(g.anchor(g.frame()), g.w, g.h)
	g.keys()
	g.mouse()
	if g.shotDone {
		return ebiten.Termination
	}
	if g.shot != "" && g.run.World.Tick() >= g.shotTick {
		g.run.Paused = true
	} else {
		g.run.Advance(time.Second / time.Duration(ebiten.TPS()))
	}
	return nil
}

func (g *game) keys() {
	r := g.run
	k := inpututil.IsKeyJustPressed
	switch {
	case k(ebiten.KeySpace):
		r.Paused = !r.Paused
	case k(ebiten.KeyN), k(ebiten.KeyPeriod):
		r.Paused = true
		r.StepOnce()
	case k(ebiten.KeyDigit1), k(ebiten.KeyDigit2), k(ebiten.KeyDigit3):
		for i, key := range []ebiten.Key{ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3} {
			if k(key) {
				r.Speed = view.Speeds[i]
			}
		}
	case k(ebiten.KeyTab):
		g.toggleMicro()
	case k(ebiten.KeyF):
		g.fit()
	case k(ebiten.KeyEscape):
		g.sel = ids.Nil
	case k(ebiten.KeyG):
		g.layers.Grids = !g.layers.Grids
	case k(ebiten.KeyP):
		g.layers.Portals = !g.layers.Portals
	case k(ebiten.KeyC):
		g.layers.Contacts = !g.layers.Contacts
	case k(ebiten.KeyR):
		g.layers.Routes = !g.layers.Routes
	case k(ebiten.KeyL):
		g.layers.Labels = !g.layers.Labels
	}
	const step = 8
	for _, a := range []struct {
		key    ebiten.Key
		dx, dy float64
	}{{ebiten.KeyArrowLeft, step, 0}, {ebiten.KeyArrowRight, -step, 0}, {ebiten.KeyArrowUp, 0, step}, {ebiten.KeyArrowDown, 0, -step}} {
		if ebiten.IsKeyPressed(a.key) {
			g.cam.Pan(a.dx, a.dy)
		}
	}
}

// toggleMicro alterna entre la vista macro y la vista micro del Space
// elegido (o del Space donde está la entidad elegida).
func (g *game) toggleMicro() {
	f := g.run.Cur
	if !g.cam.Anchor.IsNil() {
		g.cam.Follow(ids.Nil, fixedmath.Identity, g.proj)
		g.fit()
		return
	}
	target := g.sel
	if t, ok := f.ThingByID(target); ok {
		target = t.Space
	}
	if s, ok := f.SpaceByID(target); ok {
		g.cam.Follow(s.ID, s.Pose.Transform(), g.proj)
		if !s.Extent.Empty() {
			g.cam.Fit(s.Extent, g.w, g.h)
		}
	}
}

func (g *game) mouse() {
	x, y := ebiten.CursorPosition()
	if _, wy := ebiten.Wheel(); wy != 0 {
		f := 1.25
		if wy < 0 {
			f = 1 / f
		}
		g.cam.ZoomAt(g.proj, float64(x), float64(y), f)
	}
	switch {
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		g.pressAt, g.last, g.drag = [2]int{x, y}, [2]int{x, y}, false
	case ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft):
		if abs(x-g.pressAt[0])+abs(y-g.pressAt[1]) > 4 {
			g.drag = true
		}
		if g.drag {
			g.cam.Pan(float64(x-g.last[0]), float64(y-g.last[1]))
		}
		g.last = [2]int{x, y}
	case inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && !g.drag:
		tol := units.Length(6 / g.cam.Zoom) // 6 px
		g.sel = g.run.Cur.Pick(g.proj.ToWorld(float64(x), float64(y)), tol)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.moveSelected(g.proj.ToWorld(float64(x), float64(y)))
	}
}

// moveSelected manda move_to a la entidad elegida, hacia la superficie más
// alta bajo el cursor. Si se rechaza, se ve en el log de eventos.
func (g *game) moveSelected(p units.Vec2) {
	t, ok := g.run.Cur.ThingByID(g.sel)
	if !ok || t.Mover == nil {
		return
	}
	if s, l, ok := g.run.Cur.SurfaceAt(p); ok {
		_ = g.run.Enqueue(nav.MoveTo{Entity: t.ID, Space: s, Goal: l})
	}
}

func abs(v int) int { return max(v, -v) }

func (g *game) Draw(screen *ebiten.Image) {
	g.w, g.h = screen.Bounds().Dx(), screen.Bounds().Dy()
	f := g.frame()
	proj := g.cam.Project(g.anchor(f), g.w, g.h)
	render.Draw(screen, render.Scene{Frame: f, Proj: proj, Layers: g.layers, Selected: g.sel})
	g.hud(screen)
	if g.shot != "" && g.run.Paused && g.run.World.Tick() >= g.shotTick && !g.shotDone {
		g.shotDone = true
		if err := savePNG(screen, g.shot); err != nil {
			fmt.Fprintln(os.Stderr, "game:", err)
		}
	}
}

func savePNG(img *ebiten.Image, path string) error {
	b := img.Bounds()
	rgba := image.NewRGBA(b)
	img.ReadPixels(rgba.Pix)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, rgba); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func (g *game) hud(screen *ebiten.Image) {
	r := g.run
	state := fmt.Sprintf("x%d", r.Speed)
	if r.Paused {
		state = "PAUSA"
	}
	mode := "macro"
	if !g.cam.Anchor.IsNil() {
		mode = "micro " + g.name(g.cam.Anchor)
	}
	lines := []string{
		fmt.Sprintf("tick %d | %s | hash %#016x | %.0f fps", r.World.Tick(), state, r.Hash(), ebiten.ActualFPS()),
		fmt.Sprintf("vista %s | %s/px", mode, units.Length(1/g.cam.Zoom)),
	}
	if s := g.status(); s != "" {
		lines = append(lines, s)
	}
	if r.Err != nil {
		lines = append(lines, "ERROR: "+r.Err.Error())
	}
	lines = append(lines, "espacio pausa | N paso | 1/2/3 vel | tab micro | F encuadrar | click elegir | click der. mover | G P C R L capas")
	ui.Panel(screen, 8, 8, ui.Wrap(lines, (g.w-40)/ui.CharW))

	if !g.sel.IsNil() {
		cols := 60
		ui.Panel(screen, g.w-cols*ui.CharW-20, 8, ui.Wrap(view.Inspect(r.World, g.sel), cols))
	}

	const n = 12
	log := r.Log[max(0, len(r.Log)-n):]
	names := g.names()
	lines = lines[:0]
	for _, e := range log {
		lines = append(lines, fmt.Sprintf("%5d %s %s", e.Tick, e.Kind, names.Replace(e.Data)))
	}
	cols := (g.w - 40) / ui.CharW
	for i := range lines {
		if r := []rune(lines[i]); len(r) > cols {
			lines[i] = string(r[:cols-1]) + "~"
		}
	}
	ui.Panel(screen, 8, g.h-len(lines)*ui.LineH-16, lines)
}

// names reemplaza los IDs entre comillas por nombres, para el log.
func (g *game) names() *strings.Replacer {
	var pairs []string
	add := func(id ids.ID, name string) {
		if name != "" {
			pairs = append(pairs, `"`+id.String()+`"`, name)
		}
	}
	for _, s := range g.run.Cur.Spaces {
		add(s.ID, s.Name)
	}
	for _, t := range g.run.Cur.Things {
		add(t.ID, t.Name)
	}
	return strings.NewReplacer(pairs...)
}

func (g *game) name(id ids.ID) string {
	if s, ok := g.run.Cur.SpaceByID(id); ok && s.Name != "" {
		return s.Name
	}
	return id.String()
}
