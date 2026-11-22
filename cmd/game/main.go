// game es el juego: sin flags abre el menú principal (nueva partida al azar
// o un escenario de la casa) y todo se dibuja en pixel art. Con -scenario,
// -load, -crowd o -new-game abre eso directamente: en pixel art si el mundo
// tiene visuales; si no (o con -debug-view), en el visor de debug, que corre
// la simulación a 20 TPS, la dibuja a 60 FPS interpolando y permite
// inspeccionar y mandar comandos.
//
//	go run ./cmd/game                                   # menú principal
//	go run ./cmd/game -new-game -seed 7                 # partida al azar
//	go run ./cmd/game -scenario tests/scenarios/una_manana_en_la_casa.json
//	go run ./cmd/game -debug-view                       # la casa en el visor de debug
//	go run ./cmd/game -scenario tests/scenarios/floor_to_desk_by_leg.json
//	go run ./cmd/game -crowd 200 -seed 7
//	go run ./cmd/game -load s.json
//
// Controles del visor de debug: Espacio pausa | N paso | 1/2/3 velocidad
// x1/x4/x16 | Tab vista micro del Space elegido / macro | F encuadrar | rueda
// zoom | arrastrar, flechas o WASD para mover | click elegir | click derecho mover
// lo elegido · T seguir | H historia/eventos | RePág/AvPág recorrer la
// historia | G/P/C/R/L overlays (grillas, portales, contactos, rutas,
// nombres) | Esc soltar. En pixel art, F3 superpone los overlays técnicos y,
// en una partida del menú, Esc sin nada elegido pregunta si volver al menú.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
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
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sound"
	"gtsandbox/internal/ui"
	"gtsandbox/internal/view"
)

const morningScenario = "tests/scenarios/una_manana_en_la_casa.json"

func main() {
	scen := flag.String("scenario", "", "escenario JSON a reproducir")
	newRandom := flag.Bool("new-game", false, "empezar una partida al azar (con -seed, repetible)")
	crowd := flag.Int("crowd", 0, "en vez de un escenario, n tinies caminando por un escritorio")
	seed := flag.Uint64("seed", 0, "semilla para -crowd (por defecto 1) y -new-game (por defecto, al azar)")
	load := flag.String("load", "", "en vez de un escenario, continuar un snapshot JSON")
	shot := flag.String("shot", "", "guardar una captura PNG al llegar a -shot-tick y salir")
	micro := flag.String("micro", "", "empezar en la vista micro del Space con este nombre")
	follow := flag.String("follow", "", "empezar siguiendo a la entidad con este nombre (en pixel art, en su vista micro)")
	shotTick := flag.Int("shot-tick", 0, "tick del escenario (o del mundo) para -shot; corre a x16")
	story := flag.Bool("story", false, "iniciar en modo historia")
	flag.Bool("art", true, "obsoleto: el pixel art es el modo por defecto (ver -debug-view)")
	debugView := flag.Bool("debug-view", false, "usar el visor de debug en vez del pixel art")
	atlas := flag.String("atlas", "", "reemplazar el manifiesto del escenario para el pixel art")
	reducedMotion := flag.Bool("reduced-motion", false, "cambiar de escala con corte directo en pixel art")
	debugArt := flag.Bool("debug", false, "arrancar con los overlays técnicos (F3) sobre el pixel art")
	pan := flag.String("pan", "", "desplazar el encuadre inicial dx,dy píxeles del lienzo (capturas)")
	bench := flag.Duration("bench", 0, "medir simulación y dibujo durante esta duración (a x4, sin vsync) y salir")
	size := flag.String("size", "", "tamaño de ventana, por ejemplo 960x540")
	mute := flag.Bool("mute", false, "arrancar en silencio (M alterna)")
	volume := flag.Float64("volume", 1, "volumen general del sonido, de 0 a 1")
	mix := flag.String("mix", "", "renderizar la mezcla de referencia ("+mixManifest+") a este WAV y salir")
	flag.Parse()

	fail := func(err error) {
		fmt.Fprintln(os.Stderr, "game:", err)
		os.Exit(1)
	}
	if *mix != "" {
		if err := renderMix(*mix); err != nil {
			fail(err)
		}
		return
	}
	width, height := render.CanvasWidth*2, render.CanvasHeight*2
	if *size != "" {
		widthText, heightText, _ := strings.Cut(*size, "x")
		w, widthErr := strconv.Atoi(widthText)
		h, heightErr := strconv.Atoi(heightText)
		if widthErr != nil || heightErr != nil || w < 1 || h < 1 {
			fail(fmt.Errorf("-size debe ser un tamaño positivo como 960x540"))
		}
		width, height = w, h
	}
	direct := *scen != "" || *load != "" || *crowd > 0 || *newRandom || *debugView ||
		*shot != "" || *bench > 0 || *micro != "" || *follow != "" || *pan != ""

	if !direct {
		a, err := newApp(*debugArt, *reducedMotion)
		if err != nil {
			fail(err)
		}
		mixer, err := loadSound(audioManifest, *volume, *mute)
		if err != nil {
			fmt.Fprintln(os.Stderr, "game: sin sonido:", err)
		}
		a.sound = mixer
		ebiten.SetWindowSizeLimits(render.CanvasWidth, render.CanvasHeight, -1, -1)
		run(a, appTitle, width, height, fail)
		return
	}

	if *scen == "" {
		*scen = morningScenario
	}
	var g *game
	var err error
	if *newRandom {
		if *seed == 0 {
			*seed = randomSeed()
		}
		g, err = newRandomGame(*seed)
	} else {
		g, err = newGame(*scen, *crowd, max(*seed, 1), *load)
	}
	if err != nil {
		fail(err)
	}
	g.micro, g.followName = *micro, *follow
	if *pan != "" {
		dx, dy, _ := strings.Cut(*pan, ",")
		x, errX := strconv.ParseFloat(dx, 64)
		y, errY := strconv.ParseFloat(dy, 64)
		if errX != nil || errY != nil {
			fail(fmt.Errorf("-pan debe ser dx,dy, por ejemplo 200,-90"))
		}
		g.pan = [2]float64{x, y}
	}
	g.story = (*scen == morningScenario || *newRandom) && *load == "" && *crowd <= 0
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "story" {
			g.story = *story
		}
	})
	if *bench > 0 {
		g.bench = &benchmark{duration: *bench}
		g.run.Speed = 4
		ebiten.SetVsyncEnabled(false)
	}
	if *shot != "" {
		g.shot, g.shotTick = *shot, clock.Tick(*shotTick)+g.run.World.Tick()
		g.run.Speed = view.Speeds[len(view.Speeds)-1]
	}
	var app ebiten.Game = g
	if !*debugView && len(g.visuals) > 0 {
		if err := g.enableArt(*atlas); err != nil {
			fail(err)
		}
		g.art.cut, g.art.debug = *reducedMotion, *debugArt
		app = pixelGame{g}
		ebiten.SetWindowSizeLimits(render.CanvasWidth, render.CanvasHeight, -1, -1)
	} else if *size == "" {
		width, height = 1280, 800
	}
	if *shot == "" && *bench == 0 {
		mixer, err := loadSound(audioManifest, *volume, *mute)
		if err != nil {
			fmt.Fprintln(os.Stderr, "game: sin sonido:", err)
		}
		g.sound = mixer
	}
	run(app, appTitle+" | "+g.title, width, height, fail)
}

// run abre la ventana y corre el juego hasta que termina.
func run(app ebiten.Game, title string, width, height int, fail func(error)) {
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(app); err != nil {
		fail(err)
	}
}

type game struct {
	title    string
	run      *view.Runner
	status   func() string // estado del escenario, para el HUD
	visuals  map[ids.ID]string
	artDesk  ids.ID
	artAtlas string
	art      *artMode

	cam     camera.Camera
	proj    camera.Projector
	layers  render.Layers
	sel     ids.ID
	follow  ids.ID
	w, h    int
	fitted  bool
	drag    bool
	pressAt [2]int
	last    [2]int
	story   bool
	logBack int // líneas desde el final de la historia

	micro      string     // -micro
	followName string     // -follow
	pan        [2]float64 // -pan
	microExact bool       // -micro abre exactamente el Space pedido
	bench      *benchmark // -bench
	sound      *sound.Mixer
	uiSounds   sound.UIWatcher
	scene      sound.Scene
	ambience   sound.Ambience
	shot       string // captura pedida por -shot
	shotTick   clock.Tick
	shotDone   bool
	shotErr    error

	fromMenu    bool // abierta desde el menú: Esc ofrece volver
	confirmExit bool // preguntando si volver al menú
	wasPaused   bool // pausa de antes de preguntar
	toMenu      bool // el jugador eligió volver al menú
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
		return newScenarioGame(spec, scen)
	}
	return g, nil
}

// newScenarioGame arma el visor de un escenario ya leído; path es su archivo,
// contra el que se resuelve el atlas.
func newScenarioGame(spec *scenario.Spec, path string) (*game, error) {
	g := &game{layers: render.AllLayers, cam: camera.Camera{Zoom: 1e-3}}
	in, err := scenario.Build(spec)
	if err != nil {
		return nil, err
	}
	p, err := in.Player()
	if err != nil {
		return nil, err
	}
	g.title = spec.Name
	g.visuals = in.Visuals
	g.artDesk = in.Refs["desk"]
	if g.artDesk.IsNil() {
		g.artDesk = in.Refs["desk1"]
	}
	if spec.Atlas != "" {
		g.artAtlas = filepath.Join(filepath.Dir(path), spec.Atlas)
	}
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
	return g, nil
}

func (g *game) Layout(w, h int) (int, int) { return w, h }

// remainsOf devuelve los restos de una entidad que murió aplastada, si el
// evento sigue en el registro reciente.
func remainsOf(events []view.Published, id ids.ID) ids.ID {
	for i := len(events) - 1; i >= 0; i-- {
		if c, ok := events[i].Event.(crush.Crushed); ok && c.Entity == id {
			return c.Remains
		}
	}
	return ids.Nil
}

// anchor usa la pose interpolada de la entidad seguida o del Space micro.
// El seguimiento acompaña la traslación sin rotar con cada giro del actor.
func (g *game) anchor(f *view.Frame) fixedmath.Transform {
	if !g.follow.IsNil() {
		if _, ok := entityPose(f, g.follow); !ok {
			// Si murió aplastado, el seguimiento pasa a sus restos.
			if r := remainsOf(g.run.Events, g.follow); !r.IsNil() {
				if g.sel == g.follow {
					g.sel = r
				}
				g.follow = r
			}
		}
		if p, ok := entityPose(f, g.follow); !ok {
			g.stopFollowing()
		} else if g.followsInView(f) {
			return fixedmath.Transform{Pos: p.Pos}
		}
	}
	if !g.cam.Anchor.IsNil() {
		if s, ok := f.SpaceByID(g.cam.Anchor); ok {
			if g.art != nil {
				return fixedmath.Transform{Pos: s.Pose.Pos}
			}
			return s.Pose.Transform()
		}
		g.cam.Follow(ids.Nil, fixedmath.Identity, g.proj)
	}
	return fixedmath.Identity
}

func entityPose(f *view.Frame, id ids.ID) (view.Pose, bool) {
	if t, ok := f.ThingByID(id); ok {
		return t.Pose, true
	}
	if s, ok := f.SpaceByID(id); ok {
		return s.Pose, true
	}
	return view.Pose{}, false
}

func (g *game) toggleFollow() {
	if g.sel.IsNil() || g.follow == g.sel {
		g.stopFollowing()
		return
	}
	if _, ok := entityPose(g.run.Cur, g.sel); ok {
		g.follow = g.sel
		if g.followsInView(g.run.Cur) {
			g.cam.Anchor, g.cam.X, g.cam.Y = ids.Nil, 0, 0
		}
	}
}

// startFollowing elige y sigue a la entidad con ese nombre; en -art pasa a
// la vista micro de la superficie donde está.
func (g *game) startFollowing(name string) {
	if name == "" {
		return
	}
	f := g.run.Cur
	for _, t := range f.Things {
		if t.Name == name {
			g.sel = t.ID
		}
	}
	for _, s := range f.Spaces {
		if s.Name == name {
			g.sel = s.ID
		}
	}
	if g.sel.IsNil() {
		return
	}
	g.toggleFollow()
	if g.art != nil && g.art.focus.IsNil() {
		g.toggleMicro()
	}
}

func (g *game) stopFollowing() {
	if !g.follow.IsNil() {
		if g.art != nil {
			g.art.focus = g.artFocus(g.run.Cur)
		}
		g.follow = ids.Nil
		g.cam.Follow(ids.Nil, fixedmath.Identity, g.proj)
	}
}

func (g *game) frame() *view.Frame {
	return view.Interpolate(g.run.Prev, g.run.Cur, g.run.Alpha())
}

func (g *game) fit() {
	if g.art != nil {
		g.fitArt()
		return
	}
	if !g.follow.IsNil() {
		g.cam.X, g.cam.Y = 0, 0
		return
	}
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
	if g.shotDone {
		if g.shotErr != nil {
			return g.shotErr
		}
		return ebiten.Termination
	}
	if g.w == 0 {
		return nil // todavía sin Layout
	}
	if !g.fitted {
		g.startView()
	}
	dt := time.Second / time.Duration(ebiten.TPS())
	if a := g.art; a != nil && !a.pendingFade {
		a.fade = max(0, a.fade-dt)
	}
	if g.confirmExit {
		g.exitPrompt()
		g.observeSound()
		return nil
	}
	f := g.frame()
	g.proj = g.cam.Project(g.anchor(f), g.w, g.h)
	g.keys()
	g.proj = g.cam.Project(g.anchor(f), g.w, g.h)
	g.mouse(f)
	switch {
	case g.shot != "" && g.run.World.Tick() >= g.shotTick:
		g.run.Paused = true
	case g.bench != nil:
		if g.bench.update(g) {
			return ebiten.Termination
		}
	case g.shot != "":
		// La captura cae exactamente en su tick: a x16 se avanza de a 16
		// ticks sin pasarse.
		for range min(clock.Tick(view.Speeds[len(view.Speeds)-1]), g.shotTick-g.run.World.Tick()) {
			g.run.StepOnce()
		}
	default:
		g.run.Advance(dt)
	}
	g.observeSound()
	return nil
}

// startView arma el encuadre inicial: el general, la vista micro de -micro y
// el seguimiento de -follow. Lo usan el juego y las pruebas de capturas.
func (g *game) startView() {
	g.fit()
	for _, sp := range g.run.Cur.Spaces {
		if g.micro != "" && sp.Name == g.micro {
			g.sel = sp.ID
			g.proj = g.cam.Project(fixedmath.Identity, g.w, g.h)
			g.microExact = true
			g.toggleMicro()
			g.microExact = false
		}
	}
	g.startFollowing(g.followName)
	if g.pan != [2]float64{} {
		g.cam.Pan(g.pan[0], g.pan[1])
	}
	g.fitted = true
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
		if g.art != nil {
			g.focusTinies()
		}
	case k(ebiten.KeyT):
		g.toggleFollow()
	case k(ebiten.KeyM):
		g.toggleMute()
	case k(ebiten.KeyF3) && g.art != nil:
		g.art.debug = !g.art.debug
	case k(ebiten.KeyH):
		g.story, g.logBack = !g.story, 0
	case g.story && k(ebiten.KeyPageUp):
		g.logBack = min(g.logBack+storyRows, max(0, len(r.Story)-storyRows))
	case g.story && k(ebiten.KeyPageDown):
		g.logBack = max(0, g.logBack-storyRows)
	case k(ebiten.KeyF):
		if g.art != nil && g.artAtlas != "" && g.art.focus.IsNil() {
			g.stopFollowing()
		}
		g.fit()
	case k(ebiten.KeyEscape):
		if g.fromMenu && g.sel.IsNil() && g.follow.IsNil() {
			g.confirmExit, g.wasPaused, r.Paused = true, r.Paused, true
			return
		}
		g.sel = ids.Nil
		g.stopFollowing()
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
	if dx, dy := panKeys(ebiten.IsKeyPressed); dx != 0 || dy != 0 {
		g.cam.Pan(dx, dy)
	}
}

// panStep es cuánto se mueve la cámara por frame con una tecla apretada.
const panStep = 8

// panKeys devuelve el desplazamiento de la cámara según las flechas o WASD
// apretadas; la flecha y su letra juntas no suman.
func panKeys(pressed func(ebiten.Key) bool) (dx, dy float64) {
	held := func(a, b ebiten.Key) bool { return pressed(a) || pressed(b) }
	if held(ebiten.KeyArrowLeft, ebiten.KeyA) {
		dx += panStep
	}
	if held(ebiten.KeyArrowRight, ebiten.KeyD) {
		dx -= panStep
	}
	if held(ebiten.KeyArrowUp, ebiten.KeyW) {
		dy += panStep
	}
	if held(ebiten.KeyArrowDown, ebiten.KeyS) {
		dy -= panStep
	}
	return dx, dy
}

// exitPrompt atiende la pregunta de volver al menú: Enter vuelve, Esc sigue
// jugando como estaba.
func (g *game) exitPrompt() {
	k := inpututil.IsKeyJustPressed
	switch {
	case k(ebiten.KeyEnter), k(ebiten.KeyNumpadEnter):
		g.toMenu = true
	case k(ebiten.KeyEscape):
		g.confirmExit, g.run.Paused = false, g.wasPaused
	}
}

// toggleMicro alterna entre la vista macro y la vista micro del Space
// elegido (o del Space donde está la entidad elegida).
func (g *game) toggleMicro() {
	if g.art != nil {
		g.toggleArtMicro()
		return
	}
	f := g.run.Cur
	g.stopFollowing()
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

func (g *game) mouse(f *view.Frame) {
	if g.art != nil {
		g.artMouse()
		return
	}
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
		g.sel = f.Pick(g.proj.ToWorld(float64(x), float64(y)), tol)
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
	if g.bench != nil {
		g.bench.draw()
	}
	g.w, g.h = screen.Bounds().Dx(), screen.Bounds().Dy()
	f := g.frame()
	proj := g.cam.Project(g.anchor(f), g.w, g.h)
	if g.art != nil {
		g.drawArt(screen, f, proj)
		return
	}
	render.Draw(screen, render.Scene{
		Frame: f, Proj: proj, Layers: g.layers, Selected: g.sel, HideMoverLabels: g.story,
	})
	g.hud(screen)
	if g.wantsShot() {
		g.shotDone = true
		g.shotErr = savePNG(screen, g.shot)
	}
}

func (g *game) wantsShot() bool {
	return g.shot != "" && g.run.Paused && g.run.World.Tick() >= g.shotTick && !g.shotDone
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
	if !g.follow.IsNil() {
		mode = "siguiendo a " + g.name(g.follow)
	} else if !g.cam.Anchor.IsNil() {
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
	lines = append(lines, "espacio pausa | N paso | 1/2/3 vel | tab micro | F encuadrar | click elegir | T seguir | H historia",
		"click der. mover | G P C R L capas")
	ui.Panel(screen, 8, 8, ui.Wrap(lines, (g.w-40)/ui.CharW))

	if !g.sel.IsNil() {
		cols := 60
		ui.Panel(screen, g.w-cols*ui.CharW-20, 8, ui.Wrap(view.Inspect(r.World, g.sel), cols))
	}

	if g.story {
		g.storyPanel(screen)
		return
	}
	const n = 12
	log := r.Log[max(0, len(r.Log)-n):]
	names := g.names()
	lines = lines[:0]
	for _, e := range log {
		lines = append(lines, fmt.Sprintf("%5d %s %s", e.Tick, e.Kind, names.Replace(e.Data)))
	}
	g.logPanel(screen, lines)
}

const storyRows = 8

func (g *game) storyPanel(screen *ebiten.Image) {
	story := g.run.Story
	end := len(story) - g.logBack
	start := max(0, end-storyRows)
	lines := []string{"Historia | H eventos | RePág/AvPág recorrer"}
	if end == 0 {
		lines = append(lines, "Todavía no ocurrieron eventos de la historia.")
	}
	for _, e := range story[start:end] {
		sec := e.Tick / clock.TPS
		lines = append(lines, fmt.Sprintf("%02d:%02d  %s", sec/60, sec%60, e.Text))
	}
	g.logPanel(screen, lines)
}

func (g *game) logPanel(screen *ebiten.Image, lines []string) {
	cols := max(8, (g.w-40)/ui.CharW)
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
	if t, ok := g.run.Cur.ThingByID(id); ok && t.Name != "" {
		return t.Name
	}
	return id.String()
}
