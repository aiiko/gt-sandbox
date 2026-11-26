// Package crush implementa el aplastamiento: lo que baja sobre una superficie
// (una palma, un pie, un objeto que cae) mata a los tinies frágiles que quedan debajo,
// salvo que estén a cubierto en un refugio que resista la fuerza.
//
// El aplastamiento ocurre en un instante: cuando un aplastador se apoya. No
// hay daño parcial ni muerte por caída. La víctima se destruye y en su lugar
// aparecen restos, una entidad nueva (así ningún ID guardado apunta a algo
// distinto de lo que era, D3).
package crush

import (
	"errors"
	"fmt"
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/world"
)

// Crusher es el componente "crusher": algo que aplasta al apoyarse.
//
//   - Si es un Space (la palma), su huella es su extensión, recortada por su
//     grilla de navegación si tiene (los dedos). Su cara de abajo está
//     Thickness por debajo de su superficie, y se apoya sobre otra
//     superficie cuando esa cara queda a ras de ella.
//   - Si no es un Space (la taza), aplasta al aterrizar (entity_fell) en un
//     disco de radio Radius (o el de su Body) alrededor de su posición.
type Crusher struct {
	Source    string       `json:"source"` // "hand", "foot", "object": para eventos y presentación
	Force     int32        `json:"force"`  // se compara con Shelter.Strength
	Thickness units.Length `json:"thickness,omitempty"`
	Radius    units.Length `json:"radius,omitempty"`
	// Pressing son las superficies sobre las que está apoyado (ordenadas por
	// ID). Es estado: solo un apoyo nuevo aplasta.
	Pressing []ids.ID `json:"pressing,omitempty"`
}

// Fragile marca a quien muere aplastado (componente "fragile").
type Fragile struct{}

// Shelter es el componente "shelter" de un refugio (la caja de fósforos del
// nido): protege a quien está a menos de Radius, en su mismo Space, de un
// aplastador con Force ≤ Strength.
type Shelter struct {
	Strength int32        `json:"strength"`
	Radius   units.Length `json:"radius"`
}

// Remains es el componente "remains" de lo que queda de un aplastado.
type Remains struct {
	Name   comp.Name  `json:"name"`
	Of     ids.ID     `json:"of"` // la víctima (ya destruida)
	By     ids.ID     `json:"by"` // el aplastador
	Source string     `json:"source"`
	Tick   clock.Tick `json:"tick"`
	// Colony la completa la colonia de la víctima (sim/colony), si tenía.
	Colony ids.ID `json:"colony,omitempty"`
}

// Crushed se emite por cada víctima. Remains es la entidad de los restos.
type Crushed struct {
	Entity  ids.ID     `json:"entity"`
	By      ids.ID     `json:"by"`
	Source  string     `json:"source"`
	Space   ids.ID     `json:"space"`
	Pos     units.Vec2 `json:"pos"`
	Remains ids.ID     `json:"remains"`
}

func (Crushed) Kind() string { return "crushed" }

// Sheltered se emite cuando un refugio salva a alguien que estaba debajo.
type Sheltered struct {
	Entity  ids.ID `json:"entity"`
	By      ids.ID `json:"by"`
	Shelter ids.ID `json:"shelter"`
}

func (Sheltered) Kind() string { return "sheltered" }

// Pressed se emite cuando un aplastador se apoya sobre una superficie, haya
// víctimas o no (para la presentación y el sonido).
type Pressed struct {
	By      ids.ID `json:"by"`
	Source  string `json:"source"`
	Surface ids.ID `json:"surface"`
}

func (Pressed) Kind() string { return "pressed" }

// SourceSeat es el source de los aplastamientos al sentarse.
const SourceSeat = "seat"

// SourceGarment es el source de los aplastamientos por compresión al calzarse.
const SourceGarment = "garment"

// Alturas del apoyo de un Space aplastador: su cara de abajo está apoyada
// sobre una superficie si queda entre PressBelow por debajo y PressAbove por
// encima de ella. PressAbove es menor que la altura de un tiny (8 mm).
const (
	PressBelow = units.Centimeter
	PressAbove = 5 * units.Millimeter
)

// Install registra componentes, eventos, el sistema y la invariante. Va
// después de space, nav y attach.
func Install(w *world.World) error {
	r := w.ECS()
	_, err1 := ecs.Register[Crusher](r, "crusher")
	_, err2 := ecs.Register[Fragile](r, "fragile")
	_, err3 := ecs.Register[Shelter](r, "shelter")
	_, err4 := ecs.Register[Remains](r, "remains")
	_, err5 := ecs.Register[Stride](r, "stride")
	if err := errors.Join(err1, err2, err3, err4, err5); err != nil {
		return err
	}
	w.AddInvariant("crush", check)
	return errors.Join(
		world.RegisterEvent[Crushed](w),
		world.RegisterEvent[Sheltered](w),
		world.RegisterEvent[Pressed](w),
		world.RegisterEvent[Footstep](w),
		world.RegisterEvent[NearMiss](w),
		// Después de nav.move: la pisada usa la posición ya avanzada.
		w.Register(world.SystemSpec{Phase: clock.PhaseMove, Order: 1 << 19}, steps{}),
		// Antes de space.falls y jobs.upkeep: lo que muere acá libera sus
		// reservas y trabajos en el mismo tick.
		w.Register(world.SystemSpec{Phase: clock.PhaseConsequences, Order: -1}, resolve{}),
	)
}

func check(w *world.World) error {
	r := w.ECS()
	var err error
	ecs.Get[Crusher](r).Each(func(e ids.ID, c *Crusher) {
		if err == nil && !slices.IsSorted(c.Pressing) {
			err = fmt.Errorf("aplastador %v: pressing desordenado", e)
		}
	})
	ecs.Get[Remains](r).Each(func(e ids.ID, m *Remains) {
		if err == nil && (m.Name == "" || m.Source == "") {
			err = fmt.Errorf("restos %v sin nombre o causa", e)
		}
	})
	return err
}
