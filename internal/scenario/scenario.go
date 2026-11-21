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
//	               {"no_event": "move_failed"},
//	               {"at": "$tiny", "pos": {"x": 0, "y": 0}, "tol": 1000}]
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

// Expectation es exactamente una de: Event (ocurrió, opcionalmente para
// Entity y antes de By), NoEvent (nunca ocurrió) o At (posición final).
type Expectation struct {
	Event   string      `json:"event,omitempty"`
	Entity  string      `json:"entity,omitempty"`
	By      int         `json:"by,omitempty"`
	NoEvent string      `json:"no_event,omitempty"`
	At      string      `json:"at,omitempty"`
	Pos     *units.Vec2 `json:"pos,omitempty"`
	Tol     int64       `json:"tol,omitempty"`
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
		case x.Event != "":
			var want ids.ID
			if x.Entity != "" {
				id, ok := in.Refs[strings.TrimPrefix(x.Entity, "$")]
				if !ok {
					return nil, fmt.Errorf("scenario: ref desconocido %q", x.Entity)
				}
				want = id
			}
			if !in.happened(x.Event, want, x.By) {
				fails = append(fails, fmt.Sprintf("no ocurrió %s %s (by %d)", x.Event, x.Entity, x.By))
			}
		case x.NoEvent != "":
			if in.happened(x.NoEvent, ids.Nil, 0) {
				fails = append(fails, "ocurrió "+x.NoEvent)
			}
		case x.At != "" && x.Pos != nil:
			id := in.Refs[strings.TrimPrefix(x.At, "$")]
			l, ok := comp.Of(in.World.ECS()).Location.Get(id)
			if !ok {
				fails = append(fails, x.At+" sin Location")
			} else if d := l.Pos.XY().DistSq(*x.Pos); d > x.Tol*x.Tol {
				fails = append(fails, fmt.Sprintf("%s en %v, se esperaba %v ±%d", x.At, l.Pos.XY(), *x.Pos, x.Tol))
			}
		default:
			return nil, fmt.Errorf("scenario: expectativa vacía o incompleta %+v", x)
		}
	}
	return fails, nil
}

// happened busca un evento del kind dado; con entity != Nil, debe tener ese
// campo "entity"; con by > 0, debe ocurrir antes del tick by del escenario.
func (in *Instance) happened(kind string, entity ids.ID, by int) bool {
	for _, e := range in.Events {
		if e.Event.Kind() != kind || (by > 0 && int(e.Tick-in.Start) >= by) {
			continue
		}
		if entity.IsNil() {
			return true
		}
		b, _ := json.Marshal(e.Event)
		var f struct {
			Entity ids.ID `json:"entity"`
		}
		if json.Unmarshal(b, &f) == nil && f.Entity == entity {
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
