// Package scenario arma y corre mundos descriptos en JSON: entidades,
// comandos programados y expectativas. Sirve para tests de integración
// (tests/scenarios/*.json), para simrun y para benchmarks.
//
// Formato:
//
//	{
//	  "name": "...", "seed": 1, "ticks": 600,
//	  "entities": [{"ref": "desk", "spawn": {<cmds.Spawn>}}],
//	  "commands": [{"tick": 0, "kind": "move_to", "data": {...}}],
//	  "expect":   [{"event": "move_arrived", "entity": "$tiny", "by": 500},
//	               {"event": "entity_reparented", "match": {"to": "$desk"}},
//	               {"no_event": "move_failed"},
//	               {"at": "$tiny", "space": "$desk", "pos": {"x": 0, "y": 0}, "tol": 1000}]
//	}
//
// Cualquier string "$ref" dentro de spawn o data se reemplaza por el ID de la
// entidad con ese ref. Las entidades se crean antes del tick 0 del escenario,
// en tantos ticks de armado como haga falta para resolver las referencias;
// los ticks de "commands", "by" y "ticks" cuentan desde el fin del armado.
package scenario

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"

	"gtsandbox/internal/content"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/world"
)

// RemainsVisual es el identificador visual de los restos de un aplastado
// (no salen de un spawn, así que no traen visual propio).
const RemainsVisual = "remains"

type Spec struct {
	Name     string        `json:"name"`
	Seed     uint64        `json:"seed"`
	Ticks    int           `json:"ticks"`
	Entities []EntitySpec  `json:"entities"`
	Commands []CommandSpec `json:"commands"`
	Expect   []Expectation `json:"expect"`
	Notes    string        `json:"notes,omitempty"` // libre, para humanos
	Atlas    string        `json:"atlas,omitempty"` // presentación, relativo al archivo
}

// EntitySpec es una entidad a crear. Con Object, Spawn se combina con esa
// definición de content/objects (Spawn reemplaza campos y componentes).
type EntitySpec struct {
	Ref    string          `json:"ref"`
	Object string          `json:"object,omitempty"`
	Visual string          `json:"visual,omitempty"` // identificador de presentación
	Spawn  json.RawMessage `json:"spawn"`
}

type CommandSpec struct {
	Tick int             `json:"tick"`
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

// Expectation es exactamente una de: Event (ocurrió en [After, By), By = 0
// sin límite; con Count, exactamente esa cantidad de veces),
// NoEvent (nunca ocurrió) o At (posición final, opcionalmente en Space).
// Para Event y NoEvent, Match filtra por campos del evento (con "$ref");
// Entity es un atajo para {"entity": ...}.
type Expectation struct {
	Event   string          `json:"event,omitempty"`
	Entity  string          `json:"entity,omitempty"`
	Match   json.RawMessage `json:"match,omitempty"`
	By      int             `json:"by,omitempty"`
	After   int             `json:"after,omitempty"`
	Count   int             `json:"count,omitempty"`
	NoEvent string          `json:"no_event,omitempty"`
	At      string          `json:"at,omitempty"`
	Space   string          `json:"space,omitempty"` // con At: el Space esperado
	Pos     *units.Vec2     `json:"pos,omitempty"`
	Tol     int64           `json:"tol,omitempty"`
}

// Read parsea un Spec rechazando campos desconocidos.
func Read(r io.Reader) (*Spec, error) {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	var s Spec
	if err := d.Decode(&s); err != nil {
		return nil, fmt.Errorf("scenario: %w", err)
	}
	return &s, nil
}

// ReadFile lee un Spec de un archivo.
func ReadFile(path string) (*Spec, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Read(f)
}

// Instance es un escenario armado, listo para correr.
type Instance struct {
	Spec    *Spec
	World   *world.World
	Refs    map[string]ids.ID
	Visuals map[ids.ID]string
	Start   clock.Tick // tick del mundo que corresponde al tick 0 del escenario
	// Events acumula los eventos publicados desde Start.
	Events []world.EventRecord
}

// Build crea el mundo y sus entidades.
func Build(s *Spec, opts ...engine.Option) (*Instance, error) {
	w, err := engine.New(world.Config{Seed: s.Seed}, opts...)
	if err != nil {
		return nil, err
	}
	in := &Instance{Spec: s, World: w, Refs: map[string]ids.ID{}, Visuals: map[ids.ID]string{}}
	pending := slices.Clone(s.Entities)
	for len(pending) > 0 {
		var wave []EntitySpec
		var rest []EntitySpec
		for _, e := range pending {
			spawn := e.Spawn
			if e.Object != "" {
				lib, err := content.Default()
				if err != nil {
					return nil, err
				}
				if spawn, err = lib.Spawn(e.Object, spawn); err != nil {
					return nil, fmt.Errorf("scenario: entidad %q: %w", e.Ref, err)
				}
			}
			raw, err := in.resolve(spawn)
			if err != nil {
				rest = append(rest, e)
				continue
			}
			var sp cmds.Spawn
			if err := strictUnmarshal(raw, &sp); err != nil {
				return nil, fmt.Errorf("scenario: entidad %q: %w", e.Ref, err)
			}
			if e.Visual != "" {
				sp.Visual = e.Visual
			}
			if _, err := w.Enqueue(sp); err != nil {
				return nil, err
			}
			wave = append(wave, e)
		}
		if len(wave) == 0 {
			return nil, fmt.Errorf("scenario: referencias sin resolver en %q", rest[0].Ref)
		}
		w.Step()
		var spawned []ids.ID
		for _, ev := range w.Events() {
			in.recordVisual(ev.Event)
			switch e := ev.Event.(type) {
			case cmds.EntitySpawned:
				spawned = append(spawned, e.Entity)
			case world.CommandRejected:
				return nil, fmt.Errorf("scenario: spawn rechazado: %s", e.Error)
			}
		}
		for i, e := range wave {
			if e.Ref != "" {
				if _, dup := in.Refs[e.Ref]; dup {
					return nil, fmt.Errorf("scenario: ref duplicado %q", e.Ref)
				}
				in.Refs[e.Ref] = spawned[i]
			}
		}
		pending = rest
	}
	in.Start = w.Tick()
	return in, nil
}

// resolve reemplaza los "$ref" de raw. Falla si alguno no existe todavía.
func (in *Instance) resolve(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return raw, nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	var missing error
	var walk func(any) any
	walk = func(v any) any {
		switch x := v.(type) {
		case string:
			if ref, ok := strings.CutPrefix(x, "$"); ok {
				id, found := in.Refs[ref]
				if !found {
					missing = fmt.Errorf("scenario: ref desconocido %q", x)
					return x
				}
				return id.String()
			}
		case map[string]any:
			for k, e := range x {
				x[k] = walk(e)
			}
		case []any:
			for i, e := range x {
				x[i] = walk(e)
			}
		}
		return v
	}
	v = walk(v)
	if missing != nil {
		return nil, missing
	}
	return json.Marshal(v)
}

// Run corre el escenario completo y devuelve las expectativas que fallaron.
func (in *Instance) Run() ([]string, error) {
	p, err := in.Player()
	if err != nil {
		return nil, err
	}
	for !p.Done() {
		if err := p.Step(); err != nil {
			return nil, err
		}
	}
	return in.Check()
}

// Player corre un escenario tick a tick, para visores interactivos. Pasado
// Spec.Ticks sigue avanzando el mundo, sin comandos.
type Player struct {
	in     *Instance
	byTick map[int][]world.Command
	t      int
}

// Player prepara la reproducción tick a tick.
func (in *Instance) Player() (*Player, error) {
	byTick, err := in.commands()
	if err != nil {
		return nil, err
	}
	return &Player{in: in, byTick: byTick}, nil
}

// Tick es el próximo tick del escenario (relativo a Start).
func (p *Player) Tick() int { return p.t }

// Done indica si ya se corrieron los Spec.Ticks del escenario.
func (p *Player) Done() bool { return p.t >= p.in.Spec.Ticks }

// Step encola los comandos del tick, avanza el mundo y chequea invariantes.
func (p *Player) Step() error {
	w := p.in.World
	for _, c := range p.byTick[p.t] {
		if _, err := w.Enqueue(c); err != nil {
			return err
		}
	}
	w.Step()
	p.in.Events = append(p.in.Events, w.Events()...)
	for _, e := range w.Events() {
		p.in.recordVisual(e.Event)
	}
	if err := w.CheckInvariants(); err != nil {
		return fmt.Errorf("scenario: tick %d: %w", p.t, err)
	}
	p.t++
	return nil
}

func (in *Instance) recordVisual(event world.Event) {
	switch e := event.(type) {
	case cmds.EntitySpawned:
		if e.Visual != "" {
			in.Visuals[e.Entity] = e.Visual
		}
	case world.EntityDespawned:
		delete(in.Visuals, e.Entity)
	case crush.Crushed:
		in.Visuals[e.Remains] = RemainsVisual
	}
}

// commands decodifica los comandos del Spec, agrupados por tick.
func (in *Instance) commands() (map[int][]world.Command, error) {
	byTick := map[int][]world.Command{}
	for i, c := range in.Spec.Commands {
		data, err := in.resolve(c.Data)
		if err != nil {
			return nil, err
		}
		cmd, err := in.World.DecodeCommand(world.Encoded{Kind: c.Kind, Data: data})
		if err != nil {
			return nil, fmt.Errorf("scenario: comando %d: %w", i, err)
		}
		byTick[c.Tick] = append(byTick[c.Tick], cmd)
	}
	return byTick, nil
}

// Check evalúa las expectativas contra los eventos acumulados.
func (in *Instance) Check() ([]string, error) {
	var fails []string
	for _, x := range in.Spec.Expect {
		switch {
		case x.Event != "" || x.NoEvent != "":
			match, err := in.matcher(x)
			if err != nil {
				return nil, err
			}
			switch n := in.count(cmp.Or(x.Event, x.NoEvent), match, x.After, x.By); {
			case x.Event != "" && x.Count > 0 && n != x.Count:
				fails = append(fails, fmt.Sprintf("%s %s%s ocurrió %d veces, se esperaban %d", x.Event, x.Entity, x.Match, n, x.Count))
			case x.Event != "" && n == 0:
				fails = append(fails, fmt.Sprintf("no ocurrió %s %s%s (en [%d, %d))", x.Event, x.Entity, x.Match, x.After, x.By))
			case x.NoEvent != "" && n > 0:
				fails = append(fails, fmt.Sprintf("ocurrió %s %s%s", x.NoEvent, x.Entity, x.Match))
			}
		case x.At != "" && x.Pos != nil:
			id := in.Refs[strings.TrimPrefix(x.At, "$")]
			l, ok := comp.Of(in.World.ECS()).Location.Get(id)
			if !ok {
				fails = append(fails, x.At+" sin Location")
			} else if sp, want := in.Refs[strings.TrimPrefix(x.Space, "$")]; x.Space != "" && (!want || l.Space != sp) {
				fails = append(fails, fmt.Sprintf("%s en el Space %v, se esperaba %s", x.At, l.Space, x.Space))
			} else if d := l.Pos.XY().DistSq(*x.Pos); d > x.Tol*x.Tol {
				fails = append(fails, fmt.Sprintf("%s en %v, se esperaba %v ±%d", x.At, l.Pos.XY(), *x.Pos, x.Tol))
			}
		default:
			return nil, fmt.Errorf("scenario: expectativa vacía o incompleta %+v", x)
		}
	}
	return fails, nil
}

// matcher arma el filtro de campos de una expectativa de evento.
func (in *Instance) matcher(x Expectation) (map[string]any, error) {
	match := map[string]any{}
	if len(x.Match) > 0 {
		raw, err := in.resolve(x.Match)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &match); err != nil {
			return nil, fmt.Errorf("scenario: match: %w", err)
		}
	}
	if x.Entity != "" {
		id, ok := in.Refs[strings.TrimPrefix(x.Entity, "$")]
		if !ok {
			return nil, fmt.Errorf("scenario: ref desconocido %q", x.Entity)
		}
		match["entity"] = id.String()
	}
	return match, nil
}

// count cuenta los eventos del kind dado cuyos campos coinciden con match,
// en el intervalo de ticks del escenario [after, by) (by = 0: sin límite).
func (in *Instance) count(kind string, match map[string]any, after, by int) int {
	n := 0
	for _, e := range in.Events {
		t := int(e.Tick - in.Start)
		if e.Event.Kind() != kind || t < after || (by > 0 && t >= by) {
			continue
		}
		b, _ := json.Marshal(e.Event)
		var fields map[string]any
		if json.Unmarshal(b, &fields) != nil {
			continue
		}
		ok := true
		for k, v := range match {
			if !reflect.DeepEqual(fields[k], v) {
				ok = false
			}
		}
		if ok {
			n++
		}
	}
	return n
}

func strictUnmarshal(raw []byte, v any) error {
	d := json.NewDecoder(strings.NewReader(string(raw)))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
