package main

import (
	"fmt"
	"image"
	"maps"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gtsandbox/internal/camera"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/render"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/encounter"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/ui"
	"gtsandbox/internal/view"
)

func clockAt(t clock.Tick) string { return view.ClockAt(t) }

// displayNames traduce los nombres internos que no están en castellano.
var displayNames = map[string]string{
	"house":     "Casa",
	"giantess":  "Giantess",
	"cup":       "la taza",
	"plate":     "el plato",
	"crumb":     "una migaja",
	"chair":     "silla",
	"sofa":      "sofá",
	"body_rear": "trasero",
	"fiber":     "una fibra",
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
// Escritorio», «Giantess › Palma»), o la cadena completa desde el cuerpo si
// está anidado en una zona corporal («Giantess › Antebrazo izquierdo › Puño»).
// En macro, la casa y la habitación de lo seguido o elegido.
func (g *game) location(f *view.Frame, focus ids.ID) []string {
	if s, ok := f.SpaceByID(focus); ok {
		if s.Parent.IsNil() {
			return []string{capital(g.display(s.ID))}
		}
		c := chain(f, s.ID)
		for i, id := range c {
			if sp, ok := f.SpaceByID(id); ok && sp.Kind == space.Body {
				out := make([]string, 0, len(c)-i+1)
				for _, cid := range c[i:] {
					out = append(out, capital(g.display(cid)))
				}
				if t, ok := f.ThingByID(g.follow); ok && t.Space == focus && t.GripLabel != "" {
					out = append(out, capital(t.GripLabel))
				}
				return out
			}
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
	"use:breakfast":    {"Desayuna", "desayunar"},
	"use:snack":        {"Merienda", "merendar"},
	"use:snack_living": {"Merienda en el living", "merendar en el living"},
	"use:watch":        {"Mira televisión", "mirar televisión"},
	"use:eat":          {"Come en el nido", "comer en el nido"},
	"use:drink":        {"Toma agua", "tomar agua"},
	"use:rest":         {"Descansa en el nido", "descansar en el nido"},
	"use:wash_hands":   {"Se lava las manos", "lavarse las manos"},
	"use:wash_dishes":  {"Lava los platos", "lavar los platos"},
	"use:shower":       {"Se ducha", "ducharse"},
	"use:nibble":       {"Mordisquea una migaja", "mordisquear una migaja"},
	"use:shelter":      {"Descansa en la taza", "descansar en la taza"},
	"use:sleep":        {"Duerme", "dormir"},
	"forage":           {"Busca comida", "buscar comida"},
	"work":             {"Trabaja para la colonia", "trabajar"},
	"explore":          {"Explora", "explorar"},
	"flee":             {"Huye", "huir"},
	"use:write":        {"Escribe", "escribir"},
	"use:relax":        {"Descansa en el sofá", "descansar en el sofá"},
	"observe":          {"Mira a los tinies", "mirar a los tinies"},
	"share":            {"Comparte migajas", "compartir migajas"},
	"offer_palm":       {"Ofrece la palma", "ofrecer la palma"},
	"notice_gift":      {"Mira un regalo", "mirar un regalo"},
	"gift":             {"Lleva un regalo", "llevar un regalo"},
	"watch":            {"Vigila el paso", "vigilar el paso"},
	"scout":            {"Explora el mapa", "explorar el mapa"},
	"eat_ration":       {"Come una ración", "comer una ración"},
	"expedition":       {"En expedición", "salir de expedición"},
}

var roleLabels = map[string]string{
	colony.RoleGatherer: "Recolector",
	colony.RolePorter:   "Porteador",
	colony.RoleBuilder:  "Constructor",
	colony.RoleLookout:  "Vigía",
	colony.RoleScout:    "Explorador",
}

// remainsCause dice qué aplastó a quien dejó unos restos.
var remainsCause = map[string]string{"hand": "Aplastado por la palma", "foot": "Aplastado por un pie",
	"object": "Aplastado por un objeto", "seat": "Aplastado al sentarse", "garment": "Aplastado dentro de una prenda"}

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

// lifeText dice el motivo y el destino de lo que hace un tiny de la
// colonia: qué punto revisa, adónde lleva lo que carga, si vuelve al nido a
// comer o a descansar, con quién conversa. "" si no aplica.
func (g *game) lifeText(t view.Thing, carrying string) string {
	if strings.HasSuffix(t.Activity, " (suspendida)") {
		return ""
	}
	goal := g.display(t.Goal)
	switch t.Activity {
	case "search":
		if t.Step == "goto" {
			return "Va a revisar " + goal
		}
		return "Revisa " + goal
	case "gift":
		if carrying != "" {
			return "Lleva un regalo a " + goal
		}
		return "Busca un regalo para dejar en " + goal
	case "chat":
		switch t.Step {
		case "goto":
			return "Va al punto de reunión"
		case "wait":
			return "Espera a alguien para conversar"
		}
		if t.With.IsNil() {
			return "Conversa"
		}
		return "Conversa con " + g.display(t.With)
	case "use:eat":
		if t.Step == "goto" {
			return "Vuelve al nido a comer"
		}
	case "use:drink":
		if g.visuals[t.Goal] == "nest" {
			if t.Step == "goto" {
				return "Vuelve al nido a tomar agua"
			}
			return "Toma agua en el nido"
		}
		place := g.display(t.Space)
		if place != "" {
			if t.Step == "goto" {
				return "Va a tomar agua en " + place
			}
			return "Toma agua en " + place
		}
	case "use:rest":
		if t.Step == "goto" || t.Step == "" && t.Mover != nil && t.Mover.Status == "moving" {
			return "Vuelve al nido a descansar"
		}
	case "work":
		switch t.Job {
		case "pluck":
			return "Desprende una fibra de " + goal
		case "build":
			return "Arma " + goal
		case "haul":
			if carrying == "" {
				return "Va a buscar " + goal
			}
			if g.visuals[t.Goal] == "fiber" {
				return "Lleva " + carrying + " al refugio"
			}
			if g.visuals[t.Goal] == "water_drop" || g.visuals[t.Goal] == "puddle" {
				return "Lleva agua al nido"
			}
			return "Lleva " + carrying + " al nido"
		}
	}
	return ""
}

var needLabels = map[string]string{"hunger": "Hambre", "thirst": "Sed", "energy": "Energía", "fun": "Diversión", "hygiene": "Higiene"}

// card arma el inspector resumido de la entidad elegida.
func (g *game) card(f *view.Frame, id ids.ID) *ui.Card {
	var activity, step, cause string
	var goal ids.ID
	var needs []view.Need
	hasMover := false
	if t, ok := f.ThingByID(id); ok {
		activity, step, goal, needs, hasMover = t.Activity, t.Step, t.Goal, t.Needs, t.Mover != nil
		cause = t.Remains
	} else if s, ok := f.SpaceByID(id); ok {
		activity, step, goal, needs = s.Activity, s.Step, s.Goal, s.Needs
	} else {
		return nil
	}
	c := &ui.Card{
		Name:      capital(g.display(id)),
		Emote:     f.Emote(id),
		Following: g.follow == id,
		CanFollow: true,
		CanGo:     true,
	}
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
	if t, ok := f.ThingByID(id); ok {
		if text := g.lifeText(*t, carrying); text != "" {
			c.Activity = text
		}
		if t.Stranded {
			c.Activity = "Varado · busca volver al nido"
		}
		if lbl := roleLabels[t.Role]; lbl != "" && cause == "" {
			c.Name = c.Name + " · " + lbl
		}
		if t.Why != "" && cause == "" {
			c.Activity += " · " + t.Why
		}
		if t.Rations > 0 && cause == "" && carrying == "" {
			c.Carrying = fmt.Sprintf("Lleva %d ración", t.Rations)
			if t.Rations > 1 {
				c.Carrying = fmt.Sprintf("Lleva %d raciones", t.Rations)
			}
		}
		if t.GripLabel != "" && cause == "" {
			if c.Activity != "" {
				c.Activity += " · Agarrado a " + t.GripLabel
			} else {
				c.Activity = "Agarrado a " + t.GripLabel
			}
		}
		if t.NearestRefuge != "" && cause == "" && c.Dest == "" {
			c.Dest = "Refugio cercano: " + capital(t.NearestRefuge)
		}
		// Si integra la expedición activa, mostrar objetivo, fase y compañeros.
		if cause == "" && len(f.Colonies) > 0 && f.Colonies[0].Expedition != nil {
			exp := f.Colonies[0].Expedition
			if slices.Contains(exp.Members, id) {
				var mates []string
				for _, m := range exp.Members {
					if m != id {
						mates = append(mates, capital(g.display(m)))
					}
				}
				info := fmt.Sprintf("Exp: %s (%s)", capital(exp.Name), exp.Phase)
				if exp.Reason != "" {
					info += " · " + exp.Reason
				}
				if len(mates) > 0 {
					info += " · con " + strings.Join(mates, ", ")
				}
				c.Dest = info
			}
		}
	}
	if cause != "" {
		c.Activity = remainsCause[cause]
	}
	if !goal.IsNil() && cause == "" {
		if dest := g.destName(f, goal); dest != "" && !strings.Contains(strings.ToLower(c.Activity), strings.ToLower(dest)) {
			c.Dest = "Destino: " + dest
		}
	}
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
	c.Diary = g.diary(id)
	return c
}

// destName nombra el destino de una acción (el lugar o el objeto objetivo).
func (g *game) destName(f *view.Frame, goal ids.ID) string {
	if goal.IsNil() {
		return ""
	}
	if s, ok := f.SpaceByID(goal); ok && s.Name != "" {
		return capital(g.display(goal))
	}
	if t, ok := f.ThingByID(goal); ok {
		if t.Name != "" {
			return capital(g.display(goal))
		}
		if !t.Space.IsNil() {
			return capital(g.display(t.Space))
		}
	}
	return ""
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
	s := g.run.Chronicle.Lines()
	return s[max(0, len(s)-storyShown):]
}

// diaryShown es cuántas entradas del diario muestra la ficha.
const diaryShown = 3

// diary son las últimas entradas del diario de una entidad o lugar.
func (g *game) diary(id ids.ID) []ui.StoryLine {
	d := g.run.Chronicle.Diary(id)
	// en su propio diario, «Tiny 03 explora cocina» es «Explora cocina»
	self := g.display(id) + " "
	var out []ui.StoryLine
	for _, e := range d[max(0, len(d)-diaryShown):] {
		text := e.Text
		if t, ok := strings.CutPrefix(text, self); ok && self != " " {
			text = t
		}
		out = append(out, ui.StoryLine{Time: clockAt(e.Tick), Text: capital(text)})
	}
	return out
}

// deathNoticeDuration es cuánto tiempo real permanece en pantalla un aviso
// (independiente de x1/x4/x16), y deathNoticeMax cuántos se apilan a la vez.
const (
	deathNoticeDuration  = 10 * time.Second
	threatNoticeCooldown = 8 * time.Second
	deathNoticeMax       = 3
)

type hudNotice struct {
	kind         ui.NoticeKind
	title, cause string
	entry        view.StoryEntry
	left         time.Duration
}

// tinyPopulation cuenta los tinies vivos en el frame.
func tinyPopulation(f *view.Frame) int {
	n := 0
	for _, t := range f.Things {
		if isTiny(t) {
			n++
		}
	}
	return n
}

// deathCause describe la causa del deceso de un evento crush.Crushed.
func (g *game) deathCause(ev crush.Crushed) string {
	base := remainsCause[ev.Source]
	if ev.Source == "object" {
		if by := g.display(ev.By); by != "" {
			base = "Aplastado por " + by
		}
	}
	if base == "" {
		base = "Aplastado"
	}
	if place := g.display(ev.Space); place != "" {
		return base + " en " + place
	}
	return base
}

func (g *game) pushNotice(n hudNotice) {
	g.notices = append([]hudNotice{n}, g.notices...)
	if len(g.notices) > deathNoticeMax {
		g.notices = g.notices[:deathNoticeMax]
	}
}

// updateNotices descuenta dt en tiempo real de las notificaciones activas y
// suma avisos nuevos (decesos, amenazas, palma ofrecida y refugio terminado).
func (g *game) updateNotices(dt time.Duration) {
	if dt > 0 {
		for i := range g.notices {
			g.notices[i].left -= dt
		}
		g.notices = slices.DeleteFunc(g.notices, func(n hudNotice) bool { return n.left <= 0 })
		for sp, left := range g.threatCooldown {
			if left <= dt {
				delete(g.threatCooldown, sp)
			} else {
				g.threatCooldown[sp] = left - dt
			}
		}
	}
	threatTriggered := false
	for _, pub := range g.run.Events {
		if pub.Seq <= g.noticeSeq {
			continue
		}
		g.noticeSeq = pub.Seq
		switch ev := pub.Event.(type) {
		case crush.Crushed:
			who := capital(g.display(ev.Remains))
			if who == "" {
				who = "Un tiny"
			}
			g.pushNotice(hudNotice{
				kind:  ui.NoticeDeath,
				title: who + " falleció",
				cause: g.deathCause(ev),
				entry: view.StoryEntry{Tick: pub.Tick, Actor: ev.Remains, Place: ev.Space},
				left:  deathNoticeDuration,
			})
		case encounter.PalmOffered:
			place := g.display(ev.Surface)
			cause := "La giantess ofrece la palma"
			if place != "" {
				cause = "En " + place
			}
			g.pushNotice(hudNotice{
				kind:  ui.NoticePalm,
				title: "Palma ofrecida",
				cause: cause,
				entry: view.StoryEntry{Tick: pub.Tick, Place: ev.Surface},
				left:  deathNoticeDuration,
			})
		case colony.RefugeBuilt:
			sp := spaceOf(g.run.Cur, ev.Site)
			place := g.display(sp)
			cause := "Nuevo refugio disponible"
			if place != "" {
				cause = "Terminado en " + place
			}
			g.pushNotice(hudNotice{
				kind:  ui.NoticeRefuge,
				title: "Refugio terminado",
				cause: cause,
				entry: view.StoryEntry{Tick: pub.Tick, Actor: ev.By, Place: sp},
				left:  deathNoticeDuration,
			})
		case ai.Trapped:
			sp := spaceOf(g.run.Cur, ev.Entity)
			if g.threatCooldown[sp] <= 0 {
				if g.threatCooldown == nil {
					g.threatCooldown = map[ids.ID]time.Duration{}
				}
				g.threatCooldown[sp] = threatNoticeCooldown
				who := capital(g.display(ev.Entity))
				g.pushNotice(hudNotice{
					kind:  ui.NoticeThreat,
					title: "Amenaza a la colonia",
					cause: who + " quedó atrapado",
					entry: view.StoryEntry{Tick: pub.Tick, Actor: ev.Entity, Place: sp},
					left:  deathNoticeDuration,
				})
				threatTriggered = true
			}
		}
	}
	// Amenazas activas: tinies alarmados o superficie con tinies bajo una mano/pisada (17.7).
	alarmedBySpace := map[ids.ID]int{}
	tiniesBySpace := map[ids.ID]int{}
	for _, t := range g.run.Cur.Things {
		if !isTiny(t) {
			continue
		}
		tiniesBySpace[t.Space]++
		if t.Alarmed {
			alarmedBySpace[t.Space]++
		}
	}
	for sp := range threatenedSurfaces(g.run.Cur) {
		if tiniesBySpace[sp] > 0 && alarmedBySpace[sp] == 0 {
			alarmedBySpace[sp] = tiniesBySpace[sp]
		}
	}
	for _, sp := range slices.Sorted(maps.Keys(alarmedBySpace)) {
		if g.threatCooldown[sp] > 0 {
			continue
		}
		if g.threatCooldown == nil {
			g.threatCooldown = map[ids.ID]time.Duration{}
		}
		g.threatCooldown[sp] = threatNoticeCooldown
		n := alarmedBySpace[sp]
		cause := fmt.Sprintf("%d tinies en peligro", n)
		if n == 1 {
			cause = "1 tiny en peligro"
		}
		if place := g.display(sp); place != "" {
			cause += " en " + place
		}
		g.pushNotice(hudNotice{
			kind:  ui.NoticeThreat,
			title: "Amenaza a la colonia",
			cause: cause,
			entry: view.StoryEntry{Tick: g.run.Cur.Tick, Place: sp},
			left:  deathNoticeDuration,
		})
		threatTriggered = true
	}
	if threatTriggered && g.pauseOnThreat && !g.run.Paused && !g.pause.open && g.shot == "" {
		g.run.Paused = true
	}
}

// dismissNotice cierra manualmente la notificación i.
func (g *game) dismissNotice(i int) {
	if i >= 0 && i < len(g.notices) {
		g.notices = slices.Delete(g.notices, i, i+1)
	}
}

// partida arma la interfaz de partida de un frame.
func (g *game) partida(f *view.Frame, focus ids.ID) ui.HUD {
	g.updateNotices(0)
	pauseLbl := "Pausa"
	if g.run.Paused {
		pauseLbl = "Reanudar"
	}
	ctxKey, ctxLbl := "H", "Panel"
	if g.story {
		ctxLbl = "Ocultar"
	}
	if !g.sel.IsNil() {
		ctxKey, ctxLbl = "T", "Seguir"
		if g.follow == g.sel {
			ctxLbl = "Soltar"
		}
	}
	h := ui.HUD{
		Clock:      clockAt(f.Tick),
		Paused:     g.run.Paused,
		Speed:      max(0, slices.Index(view.Speeds, g.run.Speed)),
		Location:   g.location(f, focus),
		Population: tinyPopulation(f),
		ShowPop:    true,
		Controls:   ui.Controls("Espacio", pauseLbl, "Tab", g.scaleLabel(f), ctxKey, ctxLbl, "F1", "Ayuda"),
	}
	if len(f.Colonies) > 0 {
		col := f.Colonies[0]
		h.Pantry = &ui.Pantry{Portions: col.Portions, Goal: col.Goal}
	}
	for _, dn := range g.notices {
		h.Notices = append(h.Notices, ui.Notice{
			Kind:      dn.kind,
			Title:     dn.title,
			Cause:     dn.cause,
			Focusable: g.storyTarget(f, dn.entry) != ids.Nil,
		})
	}
	if !g.follow.IsNil() {
		h.Following = capital(g.display(g.follow))
	}
	if g.sound != nil {
		h.Sound = &ui.Sound{Muted: g.sound.Muted(), Volume: g.sound.Master}
	}
	if !g.sel.IsNil() && g.sel != focus {
		h.Card = g.card(f, g.sel)
	}
	// El retrato de la giantess está siempre visible abajo a la izquierda.
	face := ""
	if gi := g.giantessID(); !gi.IsNil() {
		face = f.Expression(gi)
		h.Portrait = face
	}
	if !focus.IsNil() {
		h.Minimap = g.minimap(f, focus)
	}
	if g.story {
		h.Face = view.FaceNeutral
		if face != "" {
			h.Face = face
		}
		for _, e := range g.storyEntries() {
			h.Story = append(h.Story, storyLine(f, g, e))
		}
		h.Panel = g.sidePanel(f, h.Face)
	}
	h.Tooltip = g.hoverTooltip(f, focus)
	return h
}

func storyLine(f *view.Frame, g *game, e view.StoryEntry) ui.StoryLine {
	return ui.StoryLine{
		Time:      clockAt(e.Tick),
		Text:      capital(e.Text),
		Focusable: g.storyTarget(f, e) != ids.Nil,
		Tone:      storyTone(e),
		Icon:      storyIcon(e),
	}
}

// sidePanel arma el panel lateral izquierdo con sus tres pestañas.
func (g *game) sidePanel(f *view.Frame, face string) *ui.SidePanel {
	p := &ui.SidePanel{Tab: g.panelTab, Scroll: g.panelScroll, Width: g.panelW, Height: g.panelH}
	for _, e := range g.run.Chronicle.Lines() {
		p.Chronicle = append(p.Chronicle, storyLine(f, g, e))
	}
	p.Colony = g.colonyPanel(f)
	p.Giantess = g.giantessPanel(f, face)
	return p
}

// shortActivity agrupa las actividades de los tinies en categorías legibles
// para el resumen de la pestaña Colonia.
func shortActivity(t view.Thing) string {
	if t.Alarmed || t.Activity == "flee" {
		return "Huye"
	}
	base, _ := strings.CutSuffix(t.Activity, " (suspendida)")
	switch base {
	case "work":
		switch t.Job {
		case "pluck":
			return "Fibras"
		case "build":
			return "Refugio"
		default:
			return "Acarrea"
		}
	case "search", "forage":
		return "Busca"
	case "use:eat", "use:nibble":
		return "Come"
	case "use:rest", "use:shelter":
		return "Descansa"
	case "chat":
		return "Conversa"
	case "explore", "scout":
		return "Explora"
	case "watch":
		return "Vigila"
	case "expedition":
		return "Expedición"
	case "eat_ration":
		return "Ración"
	}
	return "Libre"
}

func (g *game) colonyPanel(f *view.Frame) *ui.ColonyPanel {
	cp := &ui.ColonyPanel{}
	if len(f.Colonies) > 0 {
		col := f.Colonies[0]
		cp.Portions, cp.Goal = col.Portions, col.Goal
		cp.NoExpReason, cp.Stranded = col.NoExpReason, col.Stranded
		if col.Site != nil {
			cp.Site = &ui.SiteProgress{
				Delivered: col.Site.Delivered,
				Required:  col.Site.Required,
				Done:      col.Site.Done,
			}
		}
		g.panelExpMembers = g.panelExpMembers[:0]
		if col.Expedition != nil {
			phase := map[string]string{
				"gathering": "convoca",
				"traveling": "en camino",
				"surveying": "reconoce",
				"returning": "vuelve",
			}[col.Expedition.Phase]
			if phase == "" {
				phase = col.Expedition.Phase
			}
			cp.Expedition = fmt.Sprintf("Exp: %s · %s (%d)", capital(col.Expedition.Name), phase, len(col.Expedition.Members))
			if col.Expedition.Reason != "" {
				cp.Expedition += " · " + col.Expedition.Reason
			}
			for _, m := range col.Expedition.Members {
				place := ""
				if sp := spaceOf(f, m); !sp.IsNil() {
					place = g.display(sp)
				}
				cp.ExpMembers = append(cp.ExpMembers, ui.ExpMemberRow{
					Name:  capital(g.display(m)),
					Place: place,
				})
				g.panelExpMembers = append(g.panelExpMembers, m)
			}
		}
	}
	counts := map[string]int{}
	rep := map[string]ids.ID{}
	g.panelTinies = g.panelTinies[:0]
	for _, t := range f.Things {
		if !isTiny(t) {
			continue
		}
		cat := shortActivity(t)
		counts[cat]++
		if rep[cat].IsNil() {
			rep[cat] = t.ID
		}
		hunger, _ := view.NeedOf(t.Needs, "hunger")
		energy, _ := view.NeedOf(t.Needs, "energy")
		cp.Tinies = append(cp.Tinies, ui.TinyRow{
			Name:     capital(g.display(t.ID)),
			Role:     roleLabels[t.Role],
			Activity: cat,
			Hungry:   hunger < 300,
			Tired:    energy < 300,
			Alarmed:  t.Alarmed,
		})
		g.panelTinies = append(g.panelTinies, t.ID)
	}
	g.panelActivities = g.panelActivities[:0]
	for _, label := range slices.Sorted(maps.Keys(counts)) {
		cp.Activities = append(cp.Activities, ui.ActivityCount{Label: label, Count: counts[label]})
		g.panelActivities = append(g.panelActivities, rep[label])
	}
	return cp
}

func (g *game) giantessPanel(f *view.Frame, face string) *ui.GiantessPanel {
	gi := g.giantessID()
	if gi.IsNil() {
		return nil
	}
	s, ok := f.SpaceByID(gi)
	if !ok {
		return nil
	}
	carrying := ""
	for _, t := range f.Things {
		if t.CarriedBy == gi {
			carrying = g.display(t.ID)
		}
	}
	gp := &ui.GiantessPanel{
		Face:     face,
		Activity: activityText(s.Activity, s.Step, carrying),
		Dest:     g.destName(f, s.Goal),
	}
	for _, n := range s.Needs {
		label, ok := needLabels[n.Name]
		if !ok {
			label = capital(n.Name)
		}
		gp.Needs = append(gp.Needs, ui.Need{Label: label, Value: n.Value})
	}
	return gp
}

// storyTone colorea la hora y el icono de una línea de la crónica: ámbar para
// peligro, cian para encuentros y turquesa para otros hitos.
func storyTone(e view.StoryEntry) ui.Tone {
	switch {
	case e.Grave || e.Kind == view.KindDanger:
		return ui.ToneDanger
	case e.Kind == view.KindEncounter:
		return ui.ToneEncounter
	case e.Level == view.LevelMilestone:
		return ui.ToneMilestone
	}
	return ui.ToneNormal
}

func storyIcon(e view.StoryEntry) ui.StoryIcon {
	switch e.Kind {
	case view.KindDanger:
		return ui.IconDanger
	case view.KindEncounter:
		return ui.IconEncounter
	case view.KindColony:
		return ui.IconColony
	case view.KindRoutine:
		return ui.IconRoutine
	case view.KindSummary:
		return ui.IconSummary
	}
	if e.Grave {
		return ui.IconDanger
	}
	if e.Level == view.LevelMilestone {
		return ui.IconColony
	}
	return ui.IconNone
}

// goToEntity lleva la cámara a la entidad elegida (botón «Ir allí» de la
// ficha): abre la vista micro de su superficie si tiene una, o centra macro
// sobre ella.
func (g *game) goToEntity(id ids.ID) {
	if id.IsNil() || g.art == nil {
		return
	}
	f := g.run.Cur
	g.stopFollowing()
	if ms := g.microSurface(f, id); !ms.IsNil() {
		g.art.focus = ms
		g.fitArt()
		if t, ok := f.ThingByID(id); ok {
			if s, ok := f.SpaceByID(ms); ok {
				pos := s.MicroPos(t.Pose.Pos, t.LocalPos)
				g.cam.X = float64(pos.X - s.Pose.Pos.X)
				g.cam.Y = float64(pos.Y - s.Pose.Pos.Y)
			}
		}
		g.startArtFade()
		return
	}
	if p, ok := entityPose(f, id); ok {
		if !g.art.focus.IsNil() {
			g.art.focus = ids.Nil
			g.fitArt()
			g.startArtFade()
		}
		g.cam.X, g.cam.Y = float64(p.Pos.X), float64(p.Pos.Y)
	}
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
