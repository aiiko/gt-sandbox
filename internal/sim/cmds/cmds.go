// Package cmds define los comandos genéricos del motor: crear, destruir,
// ubicar y reubicar entidades. Son la base para escenarios, herramientas de debug y
// replays. Los comandos de gameplay (órdenes a tinies, etc.) llegarán con sus
// propios sistemas.
//
// Los kinds ("spawn", ...) son parte del formato de los logs de replay.
package cmds

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

var ErrNotAlive = errors.New("entidad no viva")

// Spawn crea una entidad con los componentes básicos indicados y emite
// EntitySpawned con el ID asignado. Con Space, la entidad es una superficie;
// un Space sin Location es una raíz.
type Spawn struct {
	Name     comp.Name      `json:"name,omitempty"`
	Location *comp.Location `json:"location,omitempty"`
	Body     *comp.Body     `json:"body,omitempty"`
	Space    *space.Space   `json:"space,omitempty"`
	// Components asigna componentes de otros subsistemas por nombre de store
	// (p. ej. "mover", "nav_grid"). Los que tienen campo propio arriba no
	// pueden ir acá: sus campos propios pasan por validación.
	Components map[string]json.RawMessage `json:"components,omitempty"`
}

var typedStores = map[string]bool{"name": true, "location": true, "body": true, "space": true}

func (Spawn) Kind() string { return "spawn" }

func (c Spawn) Apply(w *world.World) error {
	if c.Body != nil && (c.Body.Radius < 0 || c.Body.Height < 0 || c.Body.Mass < 0) {
		return errors.New("spawn: body con valores negativos")
	}
	if c.Location != nil {
		if err := space.Of(w.ECS()).CanPlace(ids.Nil, c.Location.Space); err != nil {
			return fmt.Errorf("spawn: %w", err)
		}
	}
	setters, err := decodeComponents(w, c.Components)
	if err != nil {
		return fmt.Errorf("spawn: %w", err)
	}
	s := comp.Of(w.ECS())
	id := w.ECS().Spawn()
	for _, set := range setters {
		set(id)
	}
	if c.Name != "" {
		s.Name.Set(id, c.Name)
	}
	if c.Location != nil {
		s.Location.Set(id, *c.Location)
	}
	if c.Body != nil {
		s.Body.Set(id, *c.Body)
	}
	if c.Space != nil {
		ecs.Get[space.Space](w.ECS()).Set(id, *c.Space)
	}
	w.Emit(EntitySpawned{Entity: id, Name: c.Name})
	return nil
}

// decodeComponents valida y decodifica componentes por nombre de store, en
// orden de nombre. No modifica nada: devuelve los setters.
func decodeComponents(w *world.World, cs map[string]json.RawMessage) ([]func(ids.ID), error) {
	names := make([]string, 0, len(cs))
	for n := range cs {
		if typedStores[n] {
			return nil, fmt.Errorf("%q debe ir en su campo propio", n)
		}
		names = append(names, n)
	}
	slices.Sort(names)
	setters := make([]func(ids.ID), len(names))
	for i, n := range names {
		set, err := w.ECS().DecodeComponent(n, cs[n])
		if err != nil {
			return nil, err
		}
		setters[i] = set
	}
	return setters, nil
}

// SetComponents agrega o reemplaza componentes de una entidad existente, por
// nombre de store (mismas reglas que Spawn.Components). Sirve para
// escenarios y herramientas, p. ej. para referencias circulares que no se
// pueden expresar al crear.
type SetComponents struct {
	Entity     ids.ID                     `json:"entity"`
	Components map[string]json.RawMessage `json:"components"`
}

func (SetComponents) Kind() string { return "set_components" }

func (c SetComponents) Apply(w *world.World) error {
	if !w.ECS().Alive(c.Entity) {
		return fmt.Errorf("set_components %v: %w", c.Entity, ErrNotAlive)
	}
	setters, err := decodeComponents(w, c.Components)
	if err != nil {
		return fmt.Errorf("set_components %v: %w", c.Entity, err)
	}
	for _, set := range setters {
		set(c.Entity)
	}
	return nil
}

// Despawn destruye una entidad al final del tick.
type Despawn struct {
	Entity ids.ID `json:"entity"`
}

func (Despawn) Kind() string { return "despawn" }

func (c Despawn) Apply(w *world.World) error {
	if !w.ECS().Despawn(c.Entity) {
		if w.ECS().Despawning(c.Entity) {
			return fmt.Errorf("despawn %v: ya marcada", c.Entity)
		}
		return fmt.Errorf("despawn %v: %w", c.Entity, ErrNotAlive)
	}
	return nil
}

// SetLocation reubica una entidad en coordenadas locales de un Space
// (teletransporte de debug/escenario; el movimiento normal lo hacen los
// sistemas).
type SetLocation struct {
	Entity   ids.ID        `json:"entity"`
	Location comp.Location `json:"location"`
}

func (SetLocation) Kind() string { return "set_location" }

func (c SetLocation) Apply(w *world.World) error {
	if !w.ECS().Alive(c.Entity) {
		return fmt.Errorf("set_location %v: %w", c.Entity, ErrNotAlive)
	}
	if err := space.Of(w.ECS()).CanPlace(c.Entity, c.Location.Space); err != nil {
		return fmt.Errorf("set_location %v: %w", c.Entity, err)
	}
	comp.Of(w.ECS()).Location.Set(c.Entity, c.Location)
	return nil
}

// Reparent mueve una entidad a otro Space conservando su posición en el mundo.
type Reparent struct {
	Entity ids.ID `json:"entity"`
	Space  ids.ID `json:"space"`
}

func (Reparent) Kind() string { return "reparent" }

func (c Reparent) Apply(w *world.World) error {
	if !w.ECS().Alive(c.Entity) {
		return fmt.Errorf("reparent %v: %w", c.Entity, ErrNotAlive)
	}
	from, err := space.Of(w.ECS()).Reparent(c.Entity, c.Space)
	if err != nil {
		return fmt.Errorf("reparent %v: %w", c.Entity, err)
	}
	w.Emit(space.EntityReparented{Entity: c.Entity, From: from, To: c.Space, Cause: "command"})
	return nil
}

// EntitySpawned se emite al crear una entidad con Spawn.
type EntitySpawned struct {
	Entity ids.ID    `json:"entity"`
	Name   comp.Name `json:"name,omitempty"`
}

func (EntitySpawned) Kind() string { return "entity_spawned" }

// Register habilita los comandos y eventos de este paquete en w.
func Register(w *world.World) error {
	return errors.Join(
		world.RegisterCommand[Spawn](w),
		world.RegisterCommand[Despawn](w),
		world.RegisterCommand[SetLocation](w),
		world.RegisterCommand[Reparent](w),
		world.RegisterCommand[SetComponents](w),
		world.RegisterEvent[EntitySpawned](w),
	)
}
