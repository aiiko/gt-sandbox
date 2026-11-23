package view

import (
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/encounter"
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
	case crush.Crushed:
		return ev.Remains, ev.Space
	case crush.NearMiss:
		return ev.Entity, ids.Nil
	case ai.Trapped:
		return ev.Entity, ids.Nil
	case ai.Hesitated:
		return ev.Entity, ev.Surface
	case ai.ItemsLeft:
		return ev.Entity, ids.Nil
	case encounter.Observed:
		return ev.Entity, ev.Surface
	case encounter.Shared:
		return ev.Entity, ev.Surface
	case encounter.PalmOffered:
		return ev.Entity, ev.Surface
	case encounter.PalmWithdrawn:
		return ev.Entity, ev.Surface
	case encounter.Skipped:
		return ev.Entity, ev.Surface
	case colony.Searched:
		return ev.Entity, ids.Nil
	case colony.FiberDelivered:
		return ids.Nil, ev.Site
	case colony.RefugeBuilt:
		return ev.By, ev.Site
	case colony.ChatDone:
		return ev.A, ev.Gathering
	}
	return ids.Nil, ids.Nil
}

func storyText(r *ecs.Registry, event world.Event) string {
	names := comp.Of(r).Name
	name := func(id ids.ID) string {
		if n, ok := names.Get(id); ok && n != "" {
			switch n {
			case "cup":
				return "la taza"
			case "plate":
				return "el plato"
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
		case "snack_living":
			return fmt.Sprintf("%s decide merendar en el living.", name(ev.Entity))
		case "relax":
			return fmt.Sprintf("%s decide descansar en el sofá.", name(ev.Entity))
		case "observe":
			return fmt.Sprintf("%s se acerca a mirar a los tinies de %s.", name(ev.Entity), name(ev.Target))
		}
	case ai.ActionDone:
		switch ev.What {
		case "breakfast":
			return fmt.Sprintf("%s termina de desayunar.", name(ev.Entity))
		case "snack_living":
			return fmt.Sprintf("%s termina de merendar.", name(ev.Entity))
		case "relax":
			return fmt.Sprintf("%s termina de descansar.", name(ev.Entity))
		}
	case ai.ItemDropped:
		return fmt.Sprintf("%s deja %s en %s.", name(ev.Entity), name(ev.Item), name(ev.On))
	case ai.ItemsLeft:
		// las migajas de una comida; compartir lo cuenta encounter.Shared
		switch ev.Motive {
		case "breakfast":
			return "Caen migajas del desayuno."
		case "snack_living", "snack":
			return "Caen migajas de la merienda."
		}
	case encounter.Observed:
		if ev.Seen == 1 {
			return fmt.Sprintf("%s se queda mirando a un tiny en %s.", name(ev.Entity), name(ev.Surface))
		}
		return fmt.Sprintf("%s se queda mirando a %d tinies en %s.", name(ev.Entity), ev.Seen, name(ev.Surface))
	case encounter.Shared:
		return fmt.Sprintf("%s deja unas migajas para los tinies de %s.", name(ev.Entity), name(ev.Surface))
	case encounter.PalmOffered:
		return fmt.Sprintf("%s ofrece la palma abierta a los tinies de %s.", name(ev.Entity), name(ev.Surface))
	case encounter.PalmWithdrawn:
		switch len(ev.Boarded) {
		case 0:
			return fmt.Sprintf("Nadie sube a la palma; %s la retira.", name(ev.Entity))
		case 1:
			return fmt.Sprintf("%s subió a la palma; %s la retira.", name(ev.Boarded[0]), name(ev.Entity))
		}
		return fmt.Sprintf("Subieron %d tinies a la palma; %s la retira.", len(ev.Boarded), name(ev.Entity))
	case encounter.Skipped:
		switch ev.Reason {
		case "occupied":
			return fmt.Sprintf("%s ve un tiny donde iba a apoyar la mano y no la apoya.", name(ev.Entity))
		case "enough":
			return fmt.Sprintf("%s ve que ya hay migajas en %s.", name(ev.Entity), name(ev.Surface))
		}
	case jobs.JobAssigned:
		if j := ecs.Get[jobs.Job](r).Ptr(ev.Job); j != nil && j.Type == "haul" {
			return fmt.Sprintf("La colonia asigna un acarreo a %s.", name(ev.By))
		}
	case colony.Stored:
		return fmt.Sprintf("El depósito de %s guarda %d porciones.", name(ev.Colony), ev.Total)
	case colony.Searched:
		// una búsqueda vacía no se cuenta
		if ev.Found {
			return fmt.Sprintf("%s encuentra comida en %s.", name(ev.Entity), name(ev.Spot))
		}
	case colony.FiberDelivered:
		if b, ok := ecs.Get[colony.BuildSite](r).Get(ev.Site); ok {
			return fmt.Sprintf("Una fibra más para %s (%d de %d).", name(ev.Site), ev.Total, b.Required)
		}
	case colony.RefugeBuilt:
		return fmt.Sprintf("%s termina %s.", name(ev.By), name(ev.Site))
	case colony.ChatDone:
		return fmt.Sprintf("%s y %s conversan un rato.", name(ev.A), name(ev.B))
	case attach.PoseChanged:
		if a := ecs.Get[ai.Agent](r).Ptr(ev.Entity); a != nil && !a.Hand.IsNil() {
			switch ev.Pose {
			case "hand_rest", "sit_hand_rest":
				return fmt.Sprintf("%s apoya la mano.", name(ev.Entity))
			case "sit":
				return fmt.Sprintf("%s se sienta.", name(ev.Entity))
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
	case crush.Crushed:
		by := map[string]string{"hand": "la palma", "foot": "un pie", "seat": "la giantess al sentarse"}[ev.Source]
		if by == "" {
			by = name(ev.By)
		}
		return fmt.Sprintf("%s quedó bajo %s en %s.", name(ev.Remains), by, name(ev.Space))
	case crush.NearMiss:
		if ecs.Get[ai.Agent](r).Has(ev.Entity) {
			return fmt.Sprintf("%s se salvó por poco.", name(ev.Entity))
		}
	case ai.Trapped:
		return fmt.Sprintf("%s quedó atrapado sin salida.", name(ev.Entity))
	case ai.Hesitated:
		if ev.Source == crush.SourceSeat {
			return fmt.Sprintf("%s nota tinies en %s y espera antes de sentarse.", name(ev.Entity), name(ev.Surface))
		}
		return fmt.Sprintf("%s ve tinies bajo su mano y espera.", name(ev.Entity))
	case colony.MemberLost:
		return ""
	}
	return ""
}
