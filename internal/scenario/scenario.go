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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/world"
)

type Spec struct {
	Name     string        `json:"name"`
	Seed     uint64        `json:"seed"`
	Ticks    int           `json:"ticks"`
	Entities []EntitySpec  `json:"entities"`
	Commands []CommandSpec `json:"commands"`
	Expect   []Expectation `json:"expect"`
	Notes    string        `json:"notes,omitempty"` // libre, para humanos
}

type EntitySpec struct {
	Ref   string          `json:"ref"`
	Spawn json.RawMessage `json:"spawn"`
}

type CommandSpec struct {
	Tick int             `json:"tick"`
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

// Expectation es exactamente una de: Event (ocurrió en [After, By), By = 0
// sin límite),
// NoEvent (nunca ocurrió) o At (posición final, opcionalmente en Space).
// Para Event y NoEvent, Match filtra por campos del evento (con "$ref");
// Entity es un atajo para {"entity": ...}.
type Expectation struct {
	Event   string          `json:"event,omitempty"`
	Entity  string          `json:"entity,omitempty"`
	Match   json.RawMessage `json:"match,omitempty"`
	By      int             `json:"by,omitempty"`
	After   int             `json:"after,omitempty"`
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
	Spec  *Spec
	World *world.World
	Refs  map[string]ids.ID
	Start clock.Tick // tick del mundo que corresponde al tick 0 del escenario
	// Events acumula los eventos publicados desde Start.
	Events []world.EventRecord
}

// Build crea el mundo y sus entidades.
func Build(s *Spec, opts ...engine.Option) (*Instance, error) {
	w, err := engine.New(world.Config{Seed: s.Seed}, opts...)
	if err != nil {
		return nil, err
	}
	in := &Instance{Spec: s, World: w, Refs: map[string]ids.ID{}}
	pending := slices.Clone(s.Entities)
	for len(pending) > 0 {
		var wave []EntitySpec
		var rest []EntitySpec
		for _, e := range pending {
			raw, err := in.resolve(e.Spawn)
			if err != nil {
				rest = append(rest, e)
				continue
			}
			var sp cmds.Spawn
			if err := strictUnmarshal(raw, &sp); err != nil {
				return nil, fmt.Errorf("scenario: entidad %q: %w", e.Ref, err)
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
	w := in.World
	byTick, err := in.commands()
	if err != nil {
		return nil, err
	}
	for t := range in.Spec.Ticks {
		for _, c := range byTick[t] {
			if _, err := w.Enqueue(c); err != nil {
				return nil, err
			}
		}
		w.Step()
		in.Events = append(in.Events, w.Events()...)
		if err := w.CheckInvariants(); err != nil {
			return nil, fmt.Errorf("scenario: tick %d: %w", t, err)
		}
	}
	return in.check()
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

func (in *Instance) check() ([]string, error) {
	var fails []string
	for _, x := range in.Spec.Expect {
		switch {
		case x.Event != "" || x.NoEvent != "":
			match, err := in.matcher(x)
			if err != nil {
				return nil, err
			}
			switch {
			case x.Event != "" && !in.happened(x.Event, match, x.After, x.By):
				fails = append(fails, fmt.Sprintf("no ocurrió %s %s%s (en [%d, %d))", x.Event, x.Entity, x.Match, x.After, x.By))
			case x.NoEvent != "" && in.happened(x.NoEvent, match, x.After, x.By):
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

// happened busca un evento del kind dado cuyos campos coincidan con match;
// debe ocurrir en el intervalo de ticks del escenario [after, by) (by = 0:
// sin límite).
func (in *Instance) happened(kind string, match map[string]any, after, by int) bool {
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
			return true
		}
	}
	return false
}

func strictUnmarshal(raw []byte, v any) error {
	d := json.NewDecoder(strings.NewReader(string(raw)))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
