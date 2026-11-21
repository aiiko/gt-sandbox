package view

import (
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// StoryEntry resume un evento publicado en el tick indicado. Actor es quien
// lo protagoniza y Place la superficie o habitación donde ocurre (pueden
// faltar); la interfaz los usa para enfocar desde la entrada.
type StoryEntry struct {
	Tick  clock.Tick
	Text  string
	Actor ids.ID
	Place ids.ID
}

// storySubjects devuelve el actor y el lugar de un evento de la historia.
func storySubjects(event world.Event) (actor, place ids.ID) {
	switch ev := event.(type) {
	case ai.ActionStarted:
		return ev.Entity, ids.Nil
	case ai.ActionDone:
		return ev.Entity, ids.Nil
	case ai.ItemDropped:
		return ev.Entity, ev.On
	case jobs.JobAssigned:
		return ev.By, ids.Nil
	case colony.Stored:
		return ids.Nil, ev.Colony
	case attach.PoseChanged:
		return ev.Entity, ids.Nil
	case space.EntityReparented:
		return ev.Entity, ev.To
	case ai.Explored:
		return ev.Entity, ev.Space
	}
	return ids.Nil, ids.Nil
}

func storyText(r *ecs.Registry, event world.Event) string {
	names := comp.Of(r).Name
	name := func(id ids.ID) string {
		if n, ok := names.Get(id); ok && n != "" {
			if n == "cup" {
				return "la taza"
			}
			return string(n)
		}
		return id.String()
	}
	isHand := func(id ids.ID) bool {
		s, ok := ecs.Get[space.Space](r).Get(id)
		if !ok || s.Kind != space.Articulated {
			return false
		}
		l, _ := comp.Of(r).Location.Get(id)
		a := ecs.Get[ai.Agent](r).Ptr(l.Space)
		return a != nil && a.Hand == id
	}
	switch ev := event.(type) {
	case ai.ActionStarted:
		switch ev.What {
		case "breakfast":
			return fmt.Sprintf("%s decide preparar el desayuno.", name(ev.Entity))
		case "watch":
			return fmt.Sprintf("%s decide mirar televisión.", name(ev.Entity))
		case "sleep":
			return fmt.Sprintf("%s se dispone a dormir.", name(ev.Entity))
		}
	case ai.ActionDone:
		if ev.What == "breakfast" {
			return fmt.Sprintf("%s termina de desayunar.", name(ev.Entity))
		}
	case ai.ItemDropped:
		return fmt.Sprintf("%s deja %s en %s.", name(ev.Entity), name(ev.Item), name(ev.On))
	case cmds.EntitySpawned:
		if ev.Name == "crumb" {
			return "Aparecen migajas."
		}
	case jobs.JobAssigned:
		if j := ecs.Get[jobs.Job](r).Ptr(ev.Job); j != nil && j.Type == "haul" {
			return fmt.Sprintf("La colonia asigna un acarreo a %s.", name(ev.By))
		}
	case colony.Stored:
		return fmt.Sprintf("El depósito de %s guarda %d porciones.", name(ev.Colony), ev.Total)
	case attach.PoseChanged:
		if a := ecs.Get[ai.Agent](r).Ptr(ev.Entity); a != nil && !a.Hand.IsNil() {
			switch ev.Pose {
			case "hand_rest":
				return fmt.Sprintf("%s apoya la mano.", name(ev.Entity))
			case a.IdlePose:
				return fmt.Sprintf("%s levanta la mano.", name(ev.Entity))
			}
		}
	case space.EntityReparented:
		if ev.Cause == "portal" {
			switch {
			case isHand(ev.To):
				return fmt.Sprintf("%s sube a la mano.", name(ev.Entity))
			case isHand(ev.From):
				return fmt.Sprintf("%s baja de la mano a %s.", name(ev.Entity), name(ev.To))
			}
			if a := ecs.Get[ai.Agent](r).Ptr(ev.Entity); a != nil && a.Scale == "giantess" {
				return fmt.Sprintf("%s entra en %s.", name(ev.Entity), name(ev.To))
			}
		}
	case ai.Explored:
		if !isHand(ev.Space) {
			return fmt.Sprintf("%s explora %s.", name(ev.Entity), name(ev.Space))
		}
	}
	return ""
}
