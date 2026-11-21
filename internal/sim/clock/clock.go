// Package clock define el tiempo de la simulación (decisión D2) y las fases
// fijas de cada tick.
//
// La simulación nunca lee el reloj real: el tiempo es solo el contador de
// ticks. Cada tick representa TickDuration de tiempo simulado.
package clock

import (
	"fmt"
	"time"
)

// TPS es la frecuencia lógica de la simulación.
const TPS = 20

// TickDuration es el tiempo simulado que representa un tick.
const TickDuration = time.Second / TPS

// Tick es un instante de la simulación: el número de ticks transcurridos.
type Tick uint64

// TicksFor convierte una duración simulada a ticks (redondeo hacia arriba,
// para que un intervalo nunca sea 0 ticks si la duración es positiva).
func TicksFor(d time.Duration) Tick {
	if d <= 0 {
		return 0
	}
	return Tick((d + TickDuration - 1) / TickDuration)
}

// Duration devuelve el tiempo simulado equivalente.
func (t Tick) Duration() time.Duration { return time.Duration(t) * TickDuration }

func (t Tick) String() string { return fmt.Sprintf("t%d", uint64(t)) }

// Phase ordena los sistemas dentro de un tick. El orden es parte del contrato
// del motor: agregar fases nuevas es un cambio de arquitectura (ADR).
type Phase uint8

const (
	PhaseCommands     Phase = iota // aplica comandos externos encolados
	PhaseDecide                    // percepción y toma de decisiones de IA
	PhaseAct                       // ejecución de acciones/jobs, intenciones de movimiento
	PhaseMove                      // movimiento local, transforms, attachments
	PhaseTopology                  // contactos, portales dinámicos
	PhaseConsequences              // reacciones físicas: caídas, impactos...
	PhaseSlow                      // necesidades, procesos lentos (por intervalo)
	PhaseFlush                     // publicación de eventos, despawns diferidos
	phaseCount
)

// PhaseCount es la cantidad de fases.
const PhaseCount = int(phaseCount)

var phaseNames = [...]string{
	"Commands", "Decide", "Act", "Move", "Topology", "Consequences", "Slow", "Flush",
}

func (p Phase) Valid() bool { return p < phaseCount }

func (p Phase) String() string {
	if !p.Valid() {
		return fmt.Sprintf("Phase(%d)", p)
	}
	return phaseNames[p]
}

// Due indica si algo con período `period` y desfase `offset` debe ejecutarse
// en el tick t. period 0 o 1 significa "todos los ticks".
func Due(t Tick, period, offset Tick) bool {
	if period <= 1 {
		return true
	}
	return (t+offset)%period == 0
}

// Staggered reparte trabajo periódico entre ticks usando una clave estable
// (típicamente el índice de entidad), para que no todas las entidades
// recalculen en el mismo tick.
func Staggered(t Tick, period Tick, key uint64) bool {
	if period <= 1 {
		return true
	}
	return Due(t, period, Tick(key%uint64(period)))
}
