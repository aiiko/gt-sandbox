// Package comp define los componentes básicos del motor.
//
// Los componentes son datos puros: sin lógica, sin punteros, sin referencias
// directas a otros componentes. Las referencias a otras entidades se guardan
// como ids.ID y se validan con Registry.Alive antes de usarse (D3).
//
// Los nombres de store ("location", ...) son parte del formato de guardado.
package comp

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ecs"
)

// Location ubica una entidad sobre un Space.
//
// Space es el ID de la entidad que provee la superficie (a partir de M2 los
// Spaces son entidades con su propio componente). ids.Nil significa "sin
// superficie" hasta que exista el sistema de Spaces.
// Pos está en coordenadas locales del Space, en µm.
type Location struct {
	Space ids.ID     `json:"space"`
	Pos   units.Vec3 `json:"pos"`
}

// Body describe el tamaño físico de una entidad.
type Body struct {
	Radius units.Length `json:"radius"` // radio de colisión en el plano
	Height units.Length `json:"height"`
	Mass   units.Mass   `json:"mass"`
}

// Name es un nombre legible, para debug y contenido.
type Name string

// Stores agrupa los stores de los componentes básicos.
type Stores struct {
	Location *ecs.Store[Location]
	Body     *ecs.Store[Body]
	Name     *ecs.Store[Name]
}

// Register registra los componentes básicos en r.
func Register(r *ecs.Registry) (Stores, error) {
	var s Stores
	var err error
	if s.Location, err = ecs.Register[Location](r, "location"); err != nil {
		return s, err
	}
	if s.Body, err = ecs.Register[Body](r, "body"); err != nil {
		return s, err
	}
	if s.Name, err = ecs.Register[Name](r, "name"); err != nil {
		return s, err
	}
	return s, nil
}

// Of devuelve los stores básicos de un registry donde ya se registraron.
func Of(r *ecs.Registry) Stores {
	return Stores{
		Location: ecs.Get[Location](r),
		Body:     ecs.Get[Body](r),
		Name:     ecs.Get[Name](r),
	}
}
