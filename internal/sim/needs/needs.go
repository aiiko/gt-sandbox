// Package needs mantiene las necesidades de los agentes: valores enteros en
// [0, Max] (Max = satisfecha, 0 = urgente) que bajan por intervalo.
package needs

import (
	"errors"
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/world"
)

// Max es el valor de una necesidad completamente satisfecha.
const Max = 1000

// Interval es cada cuántos ticks decaen las necesidades (1 s).
const Interval clock.Tick = clock.TPS

// Need es una necesidad. Decay es cuánto baja por Interval.
type Need struct {
	Name  string `json:"name"`
	Value int32  `json:"value"`
	Decay int32  `json:"decay,omitempty"`
}

// Needs es el componente "needs": la lista de necesidades de un agente.
type Needs []Need

// Get devuelve el valor de una necesidad.
func (n Needs) Get(name string) (int32, bool) {
	for _, x := range n {
		if x.Name == name {
			return x.Value, true
		}
	}
	return 0, false
}

// Add suma delta a una necesidad, recortando a [0, Max]. Devuelve false si
// no existe.
func (n Needs) Add(name string, delta int32) bool {
	for i := range n {
		if n[i].Name == name {
			n[i].Value = clamp(n[i].Value + delta)
			return true
		}
	}
	return false
}

func clamp(v int32) int32 { return min(max(v, 0), Max) }

// SetNeed fija el valor de una necesidad (debug y escenarios).
type SetNeed struct {
	Entity ids.ID `json:"entity"`
	Need   string `json:"need"`
	Value  int32  `json:"value"`
}

func (SetNeed) Kind() string { return "set_need" }

func (c SetNeed) Apply(w *world.World) error {
	n := ecs.Get[Needs](w.ECS()).Ptr(c.Entity)
	if n == nil {
		return fmt.Errorf("set_need %v: no tiene necesidades", c.Entity)
	}
	if c.Value < 0 || c.Value > Max {
		return fmt.Errorf("set_need: valor fuera de rango: %d", c.Value)
	}
	for i := range *n {
		if (*n)[i].Name == c.Need {
			(*n)[i].Value = c.Value
			return nil
		}
	}
	return fmt.Errorf("set_need %v: no tiene la necesidad %q", c.Entity, c.Need)
}

// decay baja las necesidades una vez por Interval.
type decay struct{}

func (decay) Name() string { return "needs.decay" }

func (decay) Run(w *world.World) {
	ecs.Get[Needs](w.ECS()).Each(func(_ ids.ID, n *Needs) {
		for i := range *n {
			(*n)[i].Value = clamp((*n)[i].Value - (*n)[i].Decay)
		}
	})
}

func check(w *world.World) error {
	var err error
	ecs.Get[Needs](w.ECS()).Each(func(e ids.ID, n *Needs) {
		seen := map[string]bool{}
		for _, x := range *n {
			if x.Name == "" || seen[x.Name] || x.Value < 0 || x.Value > Max || x.Decay < 0 {
				err = errors.Join(err, fmt.Errorf("needs %v: necesidad inválida %+v", e, x))
			}
			seen[x.Name] = true
		}
	})
	return err
}

// Install registra el store "needs", set_need y el decaimiento (fase Slow).
func Install(w *world.World) error {
	if _, err := ecs.Register[Needs](w.ECS(), "needs"); err != nil {
		return err
	}
	w.AddInvariant("needs", check)
	return errors.Join(
		world.RegisterCommand[SetNeed](w),
		w.Register(world.SystemSpec{Phase: clock.PhaseSlow, Period: Interval}, decay{}),
	)
}
