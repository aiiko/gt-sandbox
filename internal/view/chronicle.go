package view

import (
	"fmt"
	"strings"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/encounter"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// La mañana empieza a las 7:00 y cada segundo simulado es un minuto del reloj
// de la casa.
const (
	DayStart         = 7 * 60
	MinutesPerSecond = 1
)

// ClockAt es la hora de la casa en un tick («07:24»).
func ClockAt(t clock.Tick) string {
	m := DayStart + int(t)/clock.TPS*MinutesPerSecond
	return fmt.Sprintf("%02d:%02d", m/60%24, m%60)
}

// StoryLevel es el nivel de una entrada de la historia (propuesta, 17.5).
type StoryLevel uint8

const (
	// LevelDetail queda solo en los diarios de su actor y su lugar.
	LevelDetail StoryLevel = iota
	// LevelMilestone es un hito de la crónica: muertes, sustos, encuentros,
	// primeras veces y cambios de la colonia.
	LevelMilestone
	// LevelRoutine es una actividad de la giantess, en una sola línea que se
	// completa al terminar.
	LevelRoutine
	// LevelSummary resume lo que hizo la colonia en un período.
	LevelSummary
)

// StoryKind clasifica el motivo de una entrada para el icono y el color en
// la crónica (propuesta, 17.6).
type StoryKind uint8

const (
	KindGeneral   StoryKind = iota
	KindDanger              // muerte, tiny atrapado, susto
	KindEncounter           // observa, comparte, ofrece palma
	KindColony              // primera vez, despensa vacía, refugio terminado
	KindRoutine             // actividad de la giantess
	KindSummary             // resumen de la colonia
)

// StoryEntry es una entrada de la historia. Actor es quien la protagoniza y
// Place la superficie o habitación donde ocurre (pueden faltar); la interfaz
// los usa para enfocar desde la entrada. Seq la identifica: una línea de
// rutina que se completa conserva su Seq.
type StoryEntry struct {
	Seq   int
	Tick  clock.Tick
	End   clock.Tick // rutina y resumen: hasta cuándo (0 = en curso)
	Text  string
	Actor ids.ID
	Place ids.ID
	Level StoryLevel
	Kind  StoryKind
	Grave bool // hito grave (una muerte, un tiny atrapado): va con aviso
	Count int  // detalles agrupados en el diario del lugar (0 = uno solo)

	plural string // formato con %d para agrupar («%d tinies exploran cocina.»)
	extra  ids.ID // otro protagonista (el compañero de una conversación)
}

// Tamaños de lo que guarda la crónica.
const (
	DiarySize = 40 // entradas por diario
	// SummaryTicks es el período del resumen de la colonia: 30 minutos del
	// reloj de la casa.
	SummaryTicks = 30 * clock.TPS / MinutesPerSecond
	// GroupTicks es cuánto pueden separarse dos detalles para agruparse.
	GroupTicks = clock.TPS
)

// Chronicle arma la historia de la partida a partir de los eventos
// publicados (propuesta, 17.5): la crónica, con hitos, la rutina de la
// giantess y resúmenes de la colonia, y un diario por entidad y por lugar con
// el detalle. Es presentación: sale de los eventos, no se hashea ni se
// guarda, y al cargar una partida empieza vacía.
type Chronicle struct {
	lines   []StoryEntry // la crónica, últimas LogSize
	recent  []StoryEntry // todo, también el detalle, últimas LogSize
	diaries map[ids.ID][]StoryEntry
	seq     int

	seeded  bool
	firsts  map[ids.ID]bool   // lugares a los que ya llegó un tiny
	routine map[ids.ID]int    // Seq de la actividad en curso de cada giantess
	empty   map[ids.ID]bool   // despensas que ya se contaron vacías
	period  clock.Tick        // inicio del período del resumen
	tally   [tallyKinds]int32 // lo que hizo la colonia en el período
}

type tallyKind int

const (
	tallyStored tallyKind = iota
	tallyFiber
	tallyChat
	tallyFound
	tallyKinds
)

var tallyWords = [tallyKinds][2]string{
	tallyStored: {"migaja al nido", "migajas al nido"},
	tallyFiber:  {"fibra", "fibras"},
	tallyChat:   {"conversación", "conversaciones"},
	tallyFound:  {"hallazgo de comida", "hallazgos de comida"},
}

// Lines es la crónica, en orden.
func (c *Chronicle) Lines() []StoryEntry { return c.lines }

// Recent son las últimas entradas de todos los niveles, en orden.
func (c *Chronicle) Recent() []StoryEntry { return c.recent }

// Diary es el diario de una entidad o de un lugar, en orden.
func (c *Chronicle) Diary(id ids.ID) []StoryEntry { return c.diaries[id] }

// Last es la última entrada de la crónica (cero si no hay).
func (c *Chronicle) Last() StoryEntry {
	if len(c.lines) == 0 {
		return StoryEntry{}
	}
	return c.lines[len(c.lines)-1]
}

// Record clasifica un evento publicado en el tick t.
func (c *Chronicle) Record(r *ecs.Registry, t clock.Tick, event world.Event) {
	c.seed(r)
	k := chronicler{r: r}
	switch ev := event.(type) {
	case ai.ActionStarted:
		if k.isGiantess(ev.Entity) {
			c.startRoutine(k, t, ev)
		}
		return
	case ai.ActionDone:
		if seq, ok := c.routine[ev.Entity]; ok {
			c.closeRoutine(ev.Entity, seq, t)
		}
		return
	case crush.Crushed:
		// los restos heredan el diario de quien murió
		if d, ok := c.diaries[ev.Entity]; ok && !ev.Remains.IsNil() {
			c.diaries[ev.Remains] = append(c.diaries[ev.Remains], d...)
			delete(c.diaries, ev.Entity)
		}
	case ai.Consumed:
		if ecs.Get[colony.Colony](r).Has(ev.Target) && ev.Left == 0 && !c.empty[ev.Target] {
			c.empty[ev.Target] = true
			c.add(StoryEntry{Tick: t, Text: fmt.Sprintf("La despensa de %s se vació.", k.name(ev.Target)), Place: ev.Target, Level: LevelMilestone, Kind: KindColony})
		}
		return
	case colony.Stored:
		c.tally[tallyStored]++
		if ev.Total > 0 {
			delete(c.empty, ev.Colony)
		}
	case colony.FiberDelivered:
		c.tally[tallyFiber]++
	case colony.ChatDone:
		c.tally[tallyChat]++
	case colony.Searched:
		if ev.Found {
			c.tally[tallyFound]++
		}
	case space.EntityReparented:
		// la primera vez que un tiny llega a un lugar es un hito
		if ev.Cause == "portal" && !k.isGiantess(ev.Entity) && ecs.Get[ai.Agent](r).Has(ev.Entity) && !k.isHand(ev.To) && !c.firsts[ev.To] {
			c.firsts[ev.To] = true
			c.add(StoryEntry{Tick: t, Text: fmt.Sprintf("Por primera vez un tiny llega a %s.", k.name(ev.To)), Actor: ev.Entity, Place: ev.To, Level: LevelMilestone, Kind: KindColony})
		}
	}
	e, ok := k.entry(event)
	if !ok {
		return
	}
	e.Tick = t
	c.add(e)
}

// Flush cierra el período del resumen de la colonia cuando el reloj pasa a
// la media hora siguiente. Se llama una vez por tick, después de Record.
func (c *Chronicle) Flush(now clock.Tick) {
	if now >= (c.period/SummaryTicks+1)*SummaryTicks {
		c.summarize(now)
	}
}

// summarize escribe el resumen del período que termina en now (si pasó algo)
// y empieza otro.
func (c *Chronicle) summarize(now clock.Tick) {
	var parts []string
	for k, n := range c.tally {
		if n > 0 {
			w := tallyWords[k][1]
			if n == 1 {
				w = tallyWords[k][0]
			}
			parts = append(parts, fmt.Sprintf("%d %s", n, w))
		}
	}
	if len(parts) > 0 {
		text := parts[0]
		if n := len(parts); n > 1 {
			text = strings.Join(parts[:n-1], ", ") + " y " + parts[n-1]
		}
		// lleva la hora en que se escribe, así la crónica queda en orden
		c.add(StoryEntry{Tick: now, End: now, Text: "Desde las " + ClockAt(c.period) + ", la colonia: " + text + ".", Level: LevelSummary, Kind: KindSummary})
	}
	c.period, c.tally = now, [tallyKinds]int32{}
}

// routineText es cada actividad de la giantess en la crónica; true si va con
// el lugar donde ocurre. Las de los encuentros son hitos aparte.
var routineText = map[string]struct {
	text  string
	place bool
}{
	"breakfast":    {"Desayuna", true},
	"watch":        {"Mira televisión", true},
	"sleep":        {"Duerme", true},
	"snack_living": {"Merienda", true},
	"snack":        {"Merienda", true},
	"relax":        {"Descansa en el sofá", false},
}

func (c *Chronicle) startRoutine(k chronicler, t clock.Tick, ev ai.ActionStarted) {
	if seq, ok := c.routine[ev.Entity]; ok {
		c.closeRoutine(ev.Entity, seq, t)
	}
	// un cambio de actividad también cierra el resumen de la colonia
	c.summarize(t)
	what, ok := routineText[ev.What]
	if !ok {
		return
	}
	text := what.text
	if place := k.room(ev.Target); what.place && !place.IsNil() {
		text += " en " + k.name(place)
	}
	e := c.add(StoryEntry{Tick: t, Text: text + ".", Actor: ev.Entity, Level: LevelRoutine, Kind: KindRoutine})
	c.routine[ev.Entity] = e.Seq
}

// closeRoutine completa la línea de la actividad: hasta cuándo duró.
func (c *Chronicle) closeRoutine(actor ids.ID, seq int, t clock.Tick) {
	delete(c.routine, actor)
	c.update(seq, func(e *StoryEntry) {
		e.End = t
		if ClockAt(t) != ClockAt(e.Tick) {
			e.Text = strings.TrimSuffix(e.Text, ".") + ", hasta las " + ClockAt(t) + "."
		}
	})
}

// update cambia la entrada seq donde esté: crónica, recientes y diarios.
func (c *Chronicle) update(seq int, fn func(*StoryEntry)) {
	for _, list := range [][]StoryEntry{c.lines, c.recent} {
		for i := range list {
			if list[i].Seq == seq {
				fn(&list[i])
			}
		}
	}
	for _, d := range c.diaries {
		for i := range d {
			if d[i].Seq == seq {
				fn(&d[i])
			}
		}
	}
}

// add guarda una entrada nueva: en la crónica si no es detalle, en los
// recientes y en los diarios de su actor, su lugar y su otro protagonista.
// Un mismo texto en el mismo tick (varios eventos de lo mismo) entra una vez.
func (c *Chronicle) add(e StoryEntry) StoryEntry {
	for i := len(c.recent) - 1; i >= 0 && c.recent[i].Tick == e.Tick; i-- {
		if r := c.recent[i]; r.Text == e.Text && r.Actor == e.Actor && r.Place == e.Place {
			return r
		}
	}
	// un hito que llega junto con otro igual se suma a su línea, pero cada
	// protagonista lo guarda en su diario
	merged := e.Level != LevelDetail && c.merge(c.lines, e)
	c.seq++
	e.Seq = c.seq
	if e.Level != LevelDetail && !merged {
		c.lines = keepLast(append(c.lines, e), LogSize)
	}
	c.recent = keepLast(append(c.recent, e), LogSize)
	for _, id := range []ids.ID{e.Actor, e.extra} {
		if !id.IsNil() {
			c.diaries[id] = keepLast(append(c.diaries[id], e), DiarySize)
		}
	}
	if p := e.Place; !p.IsNil() && p != e.Actor && p != e.extra {
		// los detalles de un mismo tipo y lugar que llegan juntos se agrupan
		if !c.merge(c.diaries[p], e) {
			c.diaries[p] = keepLast(append(c.diaries[p], e), DiarySize)
		}
	}
	return e
}

// merge suma e a la última entrada de list si es del mismo tipo y llegó
// junto con ella («3 tinies exploran cocina»). La entrada agrupada ya no
// tiene un solo actor.
func (c *Chronicle) merge(list []StoryEntry, e StoryEntry) bool {
	n := len(list)
	if n == 0 || e.plural == "" || list[n-1].plural != e.plural || e.Tick-list[n-1].Tick > GroupTicks {
		return false
	}
	last := &list[n-1]
	last.Count = max(last.Count, 1) + 1
	last.Text = fmt.Sprintf(last.plural, last.Count)
	last.Actor = ids.Nil
	return true
}

func keepLast(s []StoryEntry, n int) []StoryEntry {
	if over := len(s) - n; over > 0 {
		s = append(s[:0], s[over:]...)
	}
	return s
}

// seed marca, la primera vez, los lugares donde ya están los tinies: llegar
// de vuelta no es una primera vez.
func (c *Chronicle) seed(r *ecs.Registry) {
	if c.seeded {
		return
	}
	c.seeded = true
	c.diaries, c.firsts, c.routine, c.empty = map[ids.ID][]StoryEntry{}, map[ids.ID]bool{}, map[ids.ID]int{}, map[ids.ID]bool{}
	loc := comp.Of(r).Location
	ecs.Get[ai.Agent](r).Each(func(id ids.ID, a *ai.Agent) {
		if a.Scale == "giantess" {
			return
		}
		if l, ok := loc.Get(id); ok {
			c.firsts[l.Space] = true
		}
	})
}

// chronicler arma el texto y el nivel de cada evento con el registro del
// tick en que se publicó.
type chronicler struct{ r *ecs.Registry }

func (k chronicler) name(id ids.ID) string {
	if n, ok := comp.Of(k.r).Name.Get(id); ok && n != "" {
		switch n {
		case "cup":
			return "la taza"
		case "plate":
			return "el plato"
		}
		return string(n)
	}
	if rem, ok := ecs.Get[crush.Remains](k.r).Get(id); ok && rem.Name != "" {
		return string(rem.Name)
	}
	return id.String()
}

func (k chronicler) isGiantess(id ids.ID) bool {
	a := ecs.Get[ai.Agent](k.r).Ptr(id)
	return a != nil && a.Scale == "giantess"
}

func (k chronicler) isHand(id ids.ID) bool {
	s, ok := ecs.Get[space.Space](k.r).Get(id)
	if !ok || s.Kind != space.Articulated {
		return false
	}
	l, _ := comp.Of(k.r).Location.Get(id)
	a := ecs.Get[ai.Agent](k.r).Ptr(l.Space)
	return a != nil && a.Hand == id
}

// room es la habitación donde está una entidad: el Space con nombre más
// alto bajo la raíz de su cadena.
func (k chronicler) room(id ids.ID) ids.ID {
	loc := comp.Of(k.r).Location
	spaces := ecs.Get[space.Space](k.r)
	room := ids.Nil
	for range 16 {
		l, ok := loc.Get(id)
		if !ok || l.Space.IsNil() {
			break
		}
		if spaces.Has(l.Space) {
			if _, parent := loc.Get(l.Space); parent {
				room = l.Space
			}
		}
		id = l.Space
	}
	return room
}

// entry es la entrada de un evento, con su nivel; false si no se cuenta.
func (k chronicler) entry(event world.Event) (StoryEntry, bool) {
	text := storyText(k, event)
	if text == "" {
		return StoryEntry{}, false
	}
	e := StoryEntry{Text: text}
	e.Actor, e.Place = storySubjects(event)
	switch ev := event.(type) {
	case crush.Crushed:
		e.Level, e.Kind, e.Grave = LevelMilestone, KindDanger, true
	case ai.Trapped:
		e.Level, e.Kind, e.Grave = LevelMilestone, KindDanger, true
	case crush.NearMiss:
		e.Level, e.Kind, e.plural = LevelMilestone, KindDanger, "%d tinies se salvaron por poco."
	case ai.Hesitated, encounter.Observed, encounter.Shared,
		encounter.PalmWithdrawn, encounter.Skipped:
		e.Level, e.Kind = LevelMilestone, KindEncounter
	case colony.RefugeBuilt:
		e.Level, e.Kind = LevelMilestone, KindColony
	case ai.Explored:
		e.plural = "%d tinies exploran " + k.name(ev.Space) + "."
	case space.EntityReparented:
		if k.isHand(ev.To) {
			e.plural = "%d tinies suben a la mano."
		}
	case colony.ChatDone:
		e.extra = ev.B
	}
	return e, true
}
