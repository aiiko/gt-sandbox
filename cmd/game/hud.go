package main

import (
	"fmt"
	"image"
	"maps"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/render"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/ui"
	"gtsandbox/internal/view"
)

// La mañana empieza a las 7:00 y cada segundo simulado es un minuto del reloj
// de la casa.
const (
	dayStart         = 7 * 60
	minutesPerSecond = 1
)

func clockAt(t clock.Tick) string {
	m := dayStart + int(t)/clock.TPS*minutesPerSecond
	return fmt.Sprintf("%02d:%02d", m/60%24, m%60)
}

// displayNames traduce los nombres internos que no están en castellano.
var displayNames = map[string]string{
	"house":    "Casa",
	"giantess": "Giantess",
	"cup":      "la taza",
	"crumb":    "una migaja",
}

// display devuelve el nombre de una entidad para la interfaz de partida.
func (g *game) display(id ids.ID) string {
	n := g.name(id)
	if d, ok := displayNames[n]; ok {
		return d
	}
	return n
}

// capital pone en mayúscula la primera letra.
func capital(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

// spaceOf devuelve el Space que contiene a una entidad.
func spaceOf(f *view.Frame, id ids.ID) ids.ID {
	if t, ok := f.ThingByID(id); ok {
		return t.Space
	}
	if s, ok := f.SpaceByID(id); ok {
		return s.Parent
	}
	return ids.Nil
}

// chain devuelve los Spaces desde la habitación (hija de la raíz) hasta s.
func chain(f *view.Frame, s ids.ID) []ids.ID {
	var out []ids.ID
	for !s.IsNil() {
		sp, ok := f.SpaceByID(s)
		if !ok || sp.Parent.IsNil() {
			break
		}
		out = append([]ids.ID{s}, out...)
		s = sp.Parent
	}
	return out
}

// location es el rótulo jerárquico de ubicación. En micro, el Space que
// contiene a la superficie enfocada y la superficie («Dormitorio ›
// Escritorio», «Giantess › Palma»). En macro, la casa y la habitación de lo
// seguido o elegido.
func (g *game) location(f *view.Frame, focus ids.ID) []string {
	if s, ok := f.SpaceByID(focus); ok {
		if s.Parent.IsNil() {
			return []string{capital(g.display(s.ID))}
		}
		return []string{capital(g.display(s.Parent)), capital(g.display(s.ID))}
	}
	root := "Casa"
	for _, s := range f.Spaces {
		if s.Parent.IsNil() {
			root = capital(g.display(s.ID))
			break
		}
	}
	out := []string{root}
	for _, id := range []ids.ID{g.follow, g.sel} {
		if c := chain(f, spaceOf(f, id)); len(c) > 0 {
			return append(out, capital(g.display(c[0])))
		}
	}
	return out
}

// activities da el presente y el infinitivo de cada actividad publicada.
var activities = map[string][2]string{
	"use:breakfast": {"Desayuna", "desayunar"},
	"use:snack":     {"Merienda", "merendar"},
	"use:watch":     {"Mira televisión", "mirar televisión"},
	"use:eat":       {"Come en el nido", "comer en el nido"},
	"use:nibble":    {"Mordisquea una migaja", "mordisquear una migaja"},
	"use:shelter":   {"Descansa en la taza", "descansar en la taza"},
	"use:sleep":     {"Duerme", "dormir"},
	"forage":        {"Busca comida", "buscar comida"},
	"work":          {"Trabaja para la colonia", "trabajar"},
	"explore":       {"Explora", "explorar"},
}

// activityText resume en castellano lo que hace un agente.
func activityText(activity, step string, carrying string) string {
	if activity == "" {
		return ""
	}
	base, suspended := strings.CutSuffix(activity, " (suspendida)")
	words, ok := activities[base]
	text := base
	switch {
	case ok && step == "goto":
		text = "Va a " + words[1]
	case ok:
		text = words[0]
	}
	if base == "work" && carrying != "" {
		text = "Acarrea " + carrying
	}
	if suspended {
		text += " (en pausa)"
	}
	return text
}

var needLabels = map[string]string{"hunger": "Hambre", "energy": "Energía", "fun": "Diversión"}

// card arma el inspector resumido de la entidad elegida.
func (g *game) card(f *view.Frame, id ids.ID) *ui.Card {
	var activity, step string
	var needs []view.Need
	hasMover := false
	if t, ok := f.ThingByID(id); ok {
		activity, step, needs, hasMover = t.Activity, t.Step, t.Needs, t.Mover != nil
	} else if s, ok := f.SpaceByID(id); ok {
		activity, step, needs = s.Activity, s.Step, s.Needs
	} else {
		return nil
	}
	c := &ui.Card{Name: capital(g.display(id)), Emote: f.Emote(id)}
	var place []string
	for _, s := range chain(f, spaceOf(f, id)) {
		place = append(place, capital(g.display(s)))
	}
	c.Place = strings.Join(place, ui.Separator)
	carrying := ""
	for _, t := range f.Things {
		if t.CarriedBy == id {
			carrying = g.display(t.ID)
		}
	}
	c.Activity = activityText(activity, step, carrying)
	if carrying != "" && !strings.HasPrefix(c.Activity, "Acarrea") {
		c.Carrying = "Lleva " + carrying
	}
	for _, n := range needs {
		label, ok := needLabels[n.Name]
		if !ok {
			label = capital(n.Name)
		}
		c.Needs = append(c.Needs, ui.Need{Label: label, Value: n.Value})
	}
	if hasMover {
		c.Order = "Clic der.: ir allí"
	}
	return c
}

// giantessID es la entidad que dibuja el retrato.
func (g *game) giantessID() ids.ID {
	for _, id := range slices.Sorted(maps.Keys(g.visuals)) {
		if g.visuals[id] == "giantess" {
			return id
		}
	}
	return ids.Nil
}

// storyShown es cuántas entradas recientes de la historia pasan al panel.
const storyShown = 6

func (g *game) storyEntries() []view.StoryEntry {
	s := g.run.Story
	return s[max(0, len(s)-storyShown):]
}

// partida arma la interfaz de partida de un frame.
func (g *game) partida(f *view.Frame, focus ids.ID) ui.HUD {
	h := ui.HUD{
		Clock:    clockAt(f.Tick),
		Paused:   g.run.Paused,
		Speed:    max(0, slices.Index(view.Speeds, g.run.Speed)),
		Location: g.location(f, focus),
		Controls: ui.Controls("Espacio", "Pausa", "1-3", "Vel.", "Tab", "Escala", "T", "Seguir",
			"F", "General", "H", "Historia", "F3", "Debug"),
	}
	if !g.follow.IsNil() {
		h.Following = capital(g.display(g.follow))
	}
	if !g.sel.IsNil() && g.sel != focus {
		h.Card = g.card(f, g.sel)
	}
	if g.story {
		h.Face = view.FaceNeutral
		if gi := g.giantessID(); !gi.IsNil() {
			h.Face = f.Expression(gi)
		}
		for _, e := range g.storyEntries() {
			h.Story = append(h.Story, ui.StoryLine{Time: clockAt(e.Tick), Text: capital(e.Text), Focusable: g.storyTarget(f, e) != ids.Nil})
		}
	}
	return h
}

// storyTarget es lo que se enfoca desde una entrada: el actor si sigue en el
// mundo, si no el lugar.
func (g *game) storyTarget(f *view.Frame, e view.StoryEntry) ids.ID {
	for _, id := range []ids.ID{e.Actor, e.Place} {
		if _, ok := entityPose(f, id); ok && !id.IsNil() {
			return id
		}
	}
	return ids.Nil
}

// focusStory enfoca desde una entrada de la historia: sigue al actor (en la
// vista de su superficie si tiene una) o abre la vista micro del lugar.
func (g *game) focusStory(e view.StoryEntry) {
	f := g.run.Cur
	target := g.storyTarget(f, e)
	if target.IsNil() {
		return
	}
	a := g.art
	if target == e.Actor {
		g.sel = target
		if g.follow != target {
			g.stopFollowing()
			g.toggleFollow()
		}
		// en la vista micro de su superficie; la giantess no tiene: en macro
		if a.focus.IsNil() != g.microSurface(f, target).IsNil() {
			g.toggleArtMicro()
		}
		return
	}
	g.stopFollowing()
	g.sel = target
	if g.microSurface(f, target).IsNil() {
		return
	}
	if a.focus.IsNil() {
		g.toggleArtMicro()
	} else {
		a.focus = target
		g.fitArt()
		g.startArtFade()
	}
}

// emoteBob hace subir y bajar el símbolo un píxel.
const emoteBob = 10

// drawEmotes dibuja los símbolos de expresión sobre las cabezas: en micro
// sobre cada tiny, en macro solo sobre la giantess y lo elegido o seguido.
func (g *game) drawEmotes(dst *ebiten.Image, f *view.Frame, scale render.Scale) {
	a := g.art
	var who []ids.ID
	if scale == render.Micro {
		for _, t := range f.Things {
			who = append(who, t.ID)
		}
	} else {
		who = []ids.ID{g.giantessID(), g.sel, g.follow}
	}
	seen := map[ids.ID]bool{}
	for _, id := range who {
		if id.IsNil() || seen[id] {
			continue
		}
		seen[id] = true
		e := f.Emote(id)
		if e == "" {
			continue
		}
		r, ok := a.last.Bounds(id)
		if !ok {
			continue
		}
		bob := int(f.Tick/emoteBob) % 2
		a.atlas.DrawClip(dst, "emote", scale, e, image.Pt((r.Min.X+r.Max.X)/2, r.Min.Y-1-bob))
	}
}

// drawDebug superpone al arte los overlays técnicos y una línea de estado.
func (g *game) drawDebug(dst *ebiten.Image, f *view.Frame, proj camera.Projector, top int) {
	render.DrawOverlays(dst, render.Scene{Frame: f, Proj: proj, Layers: g.layers, Selected: g.sel})
	line := fmt.Sprintf("tick %d  hash %#016x  %.0f fps", g.run.World.Tick(), g.run.Hash(), ebiten.ActualFPS())
	vector.FillRect(dst, 2, float32(top), float32(len(line)*6+8), 18, ui.Plum, false)
	ebitenutil.DebugPrintAt(dst, line, 6, top+1)
}
