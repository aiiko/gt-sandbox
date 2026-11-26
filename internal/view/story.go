package view

import (
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/encounter"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

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
	case encounter.CarryStarted:
		return ev.Entity, ev.Surface
	case encounter.PassengerDelivered:
		if len(ev.Passengers) > 0 {
			return ev.Passengers[0], ev.Surface
		}
		return ev.Entity, ev.Surface
	case encounter.Skipped:
		return ev.Entity, ev.Surface
	case encounter.GiftNoticed:
		return ev.Entity, ev.Surface
	case colony.Searched:
		return ev.Entity, ids.Nil
	case colony.TrinketFound:
		return ev.Entity, ev.Spot
	case colony.TrinketStored:
		return ids.Nil, ev.Colony
	case colony.GiftPlaced:
		return ev.Entity, ev.Surface
	case colony.FiberDelivered:
		return ids.Nil, ev.Site
	case colony.RefugeBuilt:
		return ev.By, ev.Site
	case colony.ChatDone:
		return ev.A, ev.Gathering
	case colony.LookoutWarned:
		return ev.By, ev.Space
	case colony.ExpeditionDeparted:
		if len(ev.Members) > 0 {
			return ev.Members[0], ev.Target
		}
		return ids.Nil, ev.Target
	case colony.ExpeditionReported:
		return ev.Entity, ev.Colony
	case colony.ExpeditionCompleted:
		return ids.Nil, ev.Target
	case colony.ExpeditionStranded:
		return ev.Entity, ids.Nil
	case colony.StrandedReturned:
		return ev.Entity, ev.Colony
	case colony.RationEaten:
		return ev.Entity, ids.Nil
	case colony.SupportDiscovered:
		return ev.Entity, ev.BodySpace
	case attach.EntityGripped:
		return ev.Entity, ev.Space
	case attach.BodyTickled:
		return ev.Body, ev.Zone
	}
	return ids.Nil, ids.Nil
}

// storyText es el texto de un evento para la historia ("" si no se cuenta).
// Las actividades de la giantess las cuenta la rutina de la crónica.
func storyText(k chronicler, event world.Event) string {
	r, name, isHand := k.r, k.name, k.isHand
	switch ev := event.(type) {
	case ai.ItemDropped:
		return fmt.Sprintf("%s deja %s en %s.", name(ev.Entity), name(ev.Item), name(ev.On))
	case ai.ItemsLeft:
		// las migajas de una comida; compartir lo cuenta encounter.Shared
		switch ev.Motive {
		case "breakfast":
			return "Caen migajas del desayuno."
		case "snack_living", "snack":
			return "Caen migajas de la merienda."
		case "wash_hands":
			return "Quedan gotas de agua en el baño."
		case "wash_dishes":
			return "Quedan gotas de agua en la mesada."
		case "shower":
			return "Quedan gotas de agua y un charco junto a la ducha."
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
	case encounter.CarryStarted:
		switch len(ev.Passengers) {
		case 0:
			return ""
		case 1:
			return fmt.Sprintf("%s lleva a %s en la palma hasta %s.", name(ev.Entity), name(ev.Passengers[0]), name(ev.Surface))
		}
		return fmt.Sprintf("%s lleva a %d tinies en la palma hasta %s.", name(ev.Entity), len(ev.Passengers), name(ev.Surface))
	case encounter.PassengerDelivered:
		switch len(ev.Passengers) {
		case 0:
			return ""
		case 1:
			return fmt.Sprintf("%s bajó a %s en %s.", name(ev.Entity), name(ev.Passengers[0]), name(ev.Surface))
		}
		return fmt.Sprintf("%s bajó a %d tinies en %s.", name(ev.Entity), len(ev.Passengers), name(ev.Surface))
	case encounter.Skipped:
		switch ev.Reason {
		case "occupied":
			return fmt.Sprintf("%s ve un tiny donde iba a apoyar la mano y no la apoya.", name(ev.Entity))
		case "enough":
			return fmt.Sprintf("%s ve que ya hay migajas en %s.", name(ev.Entity), name(ev.Surface))
		}
	case encounter.GiftNoticed:
		if !ev.Giver.IsNil() {
			return fmt.Sprintf("Ella encontró el regalo de %s.", name(ev.Giver))
		}
		return fmt.Sprintf("Ella encontró un regalo en %s.", name(ev.Surface))
	case colony.TrinketFound:
		return fmt.Sprintf("%s encontró %s.", name(ev.Entity), trinketArticle(ev.TrinketKind))
	case colony.TrinketStored:
		if ev.FirstOfKind {
			return fmt.Sprintf("La colonia encontró %s.", trinketFirst(ev.TrinketKind))
		}
		return fmt.Sprintf("La colección de %s suma %d tesoros.", name(ev.Colony), ev.Total)
	case colony.GiftPlaced:
		return fmt.Sprintf("%s dejó %s como regalo en %s.", name(ev.Entity), trinketArticle(ev.TrinketKind), name(ev.Surface))
	case jobs.JobAssigned:
		if j := ecs.Get[jobs.Job](r).Ptr(ev.Job); j != nil && j.Type == "haul" {
			return fmt.Sprintf("La colonia asigna un acarreo a %s.", name(ev.By))
		}
	case colony.Stored:
		if ev.Resource == ai.ResourceWater {
			return fmt.Sprintf("El depósito de %s guarda %d porciones de agua.", name(ev.Colony), ev.Total)
		}
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
	case colony.LookoutWarned:
		return fmt.Sprintf("%s avisa de peligro en %s.", name(ev.By), name(ev.Space))
	case colony.ExpeditionDeparted:
		if len(ev.Members) == 1 {
			return fmt.Sprintf("%s parte en expedición hacia %s.", name(ev.Members[0]), name(ev.Target))
		}
		return fmt.Sprintf("Sale una expedición de %d tinies hacia %s.", len(ev.Members), name(ev.Target))
	case colony.ExpeditionReported:
		return fmt.Sprintf("%s regresa al nido con %d hallazgos.", name(ev.Entity), ev.Findings)
	case colony.ExpeditionCompleted:
		if ev.Early {
			return fmt.Sprintf("La expedición hacia %s vuelve antes (%s).", name(ev.Target), ev.Reason)
		}
		return fmt.Sprintf("Termina la expedición hacia %s.", name(ev.Target))
	case colony.ExpeditionStranded:
		return fmt.Sprintf("%s quedó varado y espera para volver.", name(ev.Entity))
	case colony.SupportDiscovered:
		when := posePhrase(ev.Pose)
		if when != "" {
			return fmt.Sprintf("Desde %s se puede bajar a %s %s.", name(ev.BodySpace), name(ev.HouseSpace), when)
		}
		return fmt.Sprintf("Desde %s se puede bajar a %s.", name(ev.BodySpace), name(ev.HouseSpace))
	case attach.EntityGripped:
		if ev.Label != "" {
			return fmt.Sprintf("%s se sujeta en %s.", name(ev.Entity), ev.Label)
		}
		return fmt.Sprintf("%s se sujeta en %s.", name(ev.Entity), name(ev.Space))
	case colony.StrandedReturned:
		return fmt.Sprintf("%s logró volver al nido.", name(ev.Entity))
	case colony.RationEaten:
		return fmt.Sprintf("%s come una ración en el camino.", name(ev.Entity))
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
		if ev.Source == crush.SourceGarment {
			return fmt.Sprintf("%s quedó dentro de %s.", name(ev.Remains), withArticle(name(ev.By)))
		}
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
		if ev.Source == attach.DefaultGarmentSource {
			return fmt.Sprintf("%s nota tinies dentro de %s y espera.", name(ev.Entity), withArticle(name(ev.Surface)))
		}
		return fmt.Sprintf("%s ve tinies bajo su mano y espera.", name(ev.Entity))
	case attach.BodyTickled:
		region := ev.Region
		if region == "" {
			region = name(ev.Zone)
		}
		if region == "" {
			return fmt.Sprintf("%s siente un cosquilleo repentino.", name(ev.Body))
		}
		return fmt.Sprintf("%s siente cosquillas en %s.", name(ev.Body), withArticle(region))
	case colony.MemberLost:
		return ""
	}
	return ""
}

func trinketArticle(kind string) string {
	switch kind {
	case colony.TrinketButton:
		return "un botón"
	case colony.TrinketClip:
		return "un clip"
	case colony.TrinketThread:
		return "un hilo"
	case colony.TrinketGraphite:
		return "una mina de grafito"
	case colony.TrinketCoin:
		return "una moneda"
	default:
		return "un tesoro"
	}
}

func trinketFirst(kind string) string {
	switch kind {
	case colony.TrinketButton:
		return "su primer botón"
	case colony.TrinketClip:
		return "su primer clip"
	case colony.TrinketThread:
		return "su primer hilo"
	case colony.TrinketGraphite:
		return "su primera mina de grafito"
	case colony.TrinketCoin:
		return "su primera moneda"
	default:
		return "su primer tesoro"
	}
}

func withArticle(s string) string {
	switch s {
	case "baño":
		return "el baño"
	case "dormitorio":
		return "el dormitorio"
	case "living":
		return "el living"
	case "pasillo":
		return "el pasillo"
	case "genkan", "recibidor":
		return "el " + s
	case "cocina":
		return "la cocina"
	}
	if len(s) == 0 {
		return s
	}
	for _, pref := range []string{"la ", "el ", "las ", "los ", "un ", "una "} {
		if len(s) >= len(pref) && s[:len(pref)] == pref {
			return s
		}
	}
	if s[len(s)-1] == 'a' {
		return "la " + s
	}
	return "el " + s
}

func posePhrase(pose string) string {
	switch pose {
	case "lie_down":
		return "cuando se recuesta"
	case "sit", "sit_hand_rest", "sit_hand_hover", "sit_set_table", "sit_hand_sofa":
		return "cuando se sienta"
	case "hand_rest":
		return "cuando apoya la mano"
	case "idle":
		return "cuando está parada"
	case "":
		return ""
	default:
		return "en pose " + pose
	}
}
