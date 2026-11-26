package encounter

import (
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

const (
	// BondMin y BondMax acotan Trust (por tiny) y Affinity (de la giantess).
	BondMin int32 = -500
	BondMax int32 = 500

	// ReactObserveAction es la acción de un tiny que responde al ser observado.
	ReactObserveAction = "react_observe"
	// ReactOfferWave y ReactOfferShy codifican la variante en ai.Doing.Offer.
	ReactOfferWave = 0
	ReactOfferShy  = 1
	// ReactionWave y ReactionShy son las respuestas según confianza.
	ReactionWave = "wave"
	ReactionShy  = "shy"

	// ReactTicks y ReactCooldown regulan la duración y frecuencia de reacciones.
	ReactTicks         = 20
	ReactCooldown      = 120 * clock.TPS
	WaveTrustThreshold = int32(80)
	ShyTrustThreshold  = int32(-60)
	CautionTicks       = 30 * clock.TPS
)

// Bond guarda el vínculo afectivo de un tiny hacia la giantess (store "bond").
type Bond struct {
	Trust        int32      `json:"trust"`
	LastReaction clock.Tick `json:"last_reaction,omitempty"`
}

// TinyReacted se emite cuando un tiny responde activamente a la mirada de la
// giantess (saludando si confía o encogiéndose si teme).
type TinyReacted struct {
	Entity   ids.ID `json:"entity"`
	Giantess ids.ID `json:"giantess"`
	Reaction string `json:"reaction"`
	Trust    int32  `json:"trust"`
}

func (TinyReacted) Kind() string { return "tiny_reacted" }

func clampBond(v int32) int32 {
	return min(BondMax, max(BondMin, v))
}

func adjustTrust(bonds *ecs.Store[Bond], id ids.ID, delta int32) {
	if b := bonds.Ptr(id); b != nil {
		b.Trust = clampBond(b.Trust + delta)
	}
}

func personalityMod(enc *Encounter, action string, now clock.Tick) float64 {
	if enc == nil || (enc.Personality == "" && enc.Affinity == 0 && now >= enc.CautionUntil) {
		return 0
	}
	mod := float64(enc.Affinity) / float64(BondMax) * 0.04
	switch enc.Personality {
	case "gentle":
		if action == "share" || action == "offer_palm" || action == CarryAction {
			mod += 0.03
		}
	case "curious":
		if action == "observe" {
			mod += 0.04
		}
	case "distracted":
		if action == "observe" || action == "offer_palm" {
			mod -= 0.03
		}
	}
	if now < enc.CautionUntil && (action == "share" || action == "offer_palm" || action == CarryAction) {
		mod -= 0.03
	}
	return mod
}

// bonds actualiza el vínculo (Bond.Trust) y la afinidad/cautela de la
// giantess a partir de los eventos del tick anterior (D5).
type bonds struct{}

func (bonds) Name() string { return "encounter.bonds" }

func (bonds) Run(w *world.World) {
	events := w.Events()
	if len(events) == 0 {
		return
	}
	r := w.ECS()
	bondStore := ecs.Get[Bond](r)
	encStore := ecs.Get[Encounter](r)
	hasAnyBond := bondStore.Len() > 0
	hasPersonality := false
	encStore.Each(func(_ ids.ID, enc *Encounter) {
		if enc.Personality != "" || enc.Affinity != 0 {
			hasPersonality = true
		}
	})
	if !hasAnyBond && !hasPersonality {
		return
	}

	locs := comp.Of(r).Location
	remains := ecs.Get[crush.Remains](r)
	now := w.Tick()

	// Helper determinista para obtener los tinies con Bond ubicados en surf.
	bondedIn := func(surf ids.ID) []ids.ID {
		var list []ids.ID
		bondStore.Each(func(id ids.ID, _ *Bond) {
			if r.Despawning(id) {
				return
			}
			if l, ok := locs.Get(id); ok && l.Space == surf {
				list = append(list, id)
			}
		})
		slices.Sort(list)
		return list
	}

	// Encuentra el dueño con Encounter de una entidad causante (ej. la palma -> giantess).
	findEncounter := func(id ids.ID) *Encounter {
		g := space.Of(r)
		cur := id
		for range space.MaxDepth + 1 {
			if cur.IsNil() {
				break
			}
			if enc := encStore.Ptr(cur); enc != nil && !r.Despawning(cur) {
				return enc
			}
			cur = g.Parent(cur)
		}
		return nil
	}

	for _, published := range events {
		switch ev := published.Event.(type) {
		case Shared:
			for _, id := range bondedIn(ev.Surface) {
				adjustTrust(bondStore, id, 35)
			}
			if enc := encStore.Ptr(ev.Entity); enc != nil {
				enc.Affinity = clampBond(enc.Affinity + 10)
			}
		case PassengerDelivered:
			for _, id := range ev.Passengers {
				if !r.Despawning(id) {
					adjustTrust(bondStore, id, 50)
				}
			}
			if enc := encStore.Ptr(ev.Entity); enc != nil {
				enc.Affinity = clampBond(enc.Affinity + 15)
			}
		case PalmWithdrawn:
			for _, id := range ev.Boarded {
				if !r.Despawning(id) {
					adjustTrust(bondStore, id, 10)
				}
			}
			if len(ev.Boarded) > 0 {
				if enc := encStore.Ptr(ev.Entity); enc != nil {
					enc.Affinity = clampBond(enc.Affinity + 5)
				}
			}
		case GiftNoticed:
			for _, id := range bondedIn(ev.Surface) {
				adjustTrust(bondStore, id, 15)
			}
			if enc := encStore.Ptr(ev.Entity); enc != nil {
				enc.Affinity = clampBond(enc.Affinity + 25)
			}
		case crush.NearMiss:
			if !r.Despawning(ev.Entity) {
				adjustTrust(bondStore, ev.Entity, -45)
			}
		case crush.Crushed:
			for _, id := range bondedIn(ev.Space) {
				adjustTrust(bondStore, id, -80)
			}
			if enc := findEncounter(ev.By); enc != nil {
				enc.Affinity = clampBond(enc.Affinity - 20)
				if enc.CautionUntil < now+CautionTicks {
					enc.CautionUntil = now + CautionTicks
				}
			}
		case ai.Noticed:
			if remains.Has(ev.Target) && !r.Despawning(ev.Entity) {
				adjustTrust(bondStore, ev.Entity, -25)
			}
		}
	}
}

// ReactObserve es la acción "react_observe": un tiny con vínculo (Bond)
// saluda ("wave") si confía mucho o se encoge ("shy") si teme, cuando una
// giantess está observando activamente su superficie.
type ReactObserve struct{}

func (ReactObserve) Name() string { return ReactObserveAction }

// ReactionOfOffer traduce ai.Doing.Offer al nombre de reacción correspondiente.
func ReactionOfOffer(offer int) string {
	if offer == ReactOfferShy {
		return ReactionShy
	}
	return ReactionWave
}

func (ReactObserve) What(_ *ai.Ctx, d *ai.Doing) string {
	if d != nil {
		return ReactionOfOffer(d.Offer)
	}
	return ReactionWave
}

func reactionChoice(trust int32) (kind string, offer int, util float64, ok bool) {
	switch {
	case trust >= WaveTrustThreshold:
		return ReactionWave, ReactOfferWave, 0.38, true
	case trust <= ShyTrustThreshold:
		return ReactionShy, ReactOfferShy, 0.42, true
	default:
		return "", 0, 0, false
	}
}

func observingGiantess(c *ai.Ctx, surface ids.ID) ids.ID {
	r := c.W.ECS()
	agents := ecs.Get[ai.Agent](r)
	encStore := ecs.Get[Encounter](r)
	idsList := encStore.IDs()
	slices.Sort(idsList)
	for _, gid := range idsList {
		if gid == c.Actor || r.Despawning(gid) {
			continue
		}
		a, ok := agents.Get(gid)
		if !ok || a.Doing == nil {
			continue
		}
		if a.Doing.Action == "observe" && a.Doing.Stage >= 1 && a.Doing.Target == surface {
			return gid
		}
	}
	return ids.Nil
}

func (ReactObserve) Options(c *ai.Ctx) []ai.Option {
	r := c.W.ECS()
	b, ok := ecs.Get[Bond](r).Get(c.Actor)
	if !ok {
		return nil
	}
	now := c.W.Tick()
	if b.LastReaction > 0 && now < b.LastReaction+ReactCooldown {
		return nil
	}
	_, offer, util, ok := reactionChoice(b.Trust)
	if !ok {
		return nil
	}
	loc, ok := comp.Of(r).Location.Get(c.Actor)
	if !ok || loc.Space.IsNil() {
		return nil
	}
	gid := observingGiantess(c, loc.Space)
	if gid.IsNil() {
		return nil
	}
	return []ai.Option{{Target: gid, Offer: offer, Utility: util}}
}

func (ReactObserve) Utility(c *ai.Ctx, d *ai.Doing) float64 {
	b, ok := ecs.Get[Bond](c.W.ECS()).Get(c.Actor)
	if !ok {
		return 0
	}
	if d != nil && d.Offer == ReactOfferShy {
		return 0.42
	}
	if _, _, util, ok := reactionChoice(b.Trust); ok {
		return util
	}
	return 0
}

func (ReactObserve) Start(c *ai.Ctx, d *ai.Doing) bool {
	r := c.W.ECS()
	b, ok := ecs.Get[Bond](r).Get(c.Actor)
	if !ok {
		return false
	}
	loc, ok := comp.Of(r).Location.Get(c.Actor)
	if !ok || loc.Space.IsNil() {
		return false
	}
	_, offer, _, ok := reactionChoice(b.Trust)
	if !ok {
		return false
	}
	if gid := observingGiantess(c, loc.Space); gid != d.Target {
		return false
	}
	d.Offer = offer
	d.Place = loc.Space
	return true
}

func (ReactObserve) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	r := c.W.ECS()
	b := ecs.Get[Bond](r).Ptr(c.Actor)
	if b == nil || !r.Alive(d.Target) || r.Despawning(d.Target) {
		return ai.Failed
	}
	if d.Progress++; d.Progress < ReactTicks {
		return ai.Running
	}
	reaction := ReactionOfOffer(d.Offer)
	b.LastReaction = c.W.Tick()
	if reaction == ReactionWave {
		if enc := ecs.Get[Encounter](r).Ptr(d.Target); enc != nil {
			enc.Affinity = clampBond(enc.Affinity + 10)
		}
	}
	c.W.Emit(TinyReacted{
		Entity:   c.Actor,
		Giantess: d.Target,
		Reaction: reaction,
		Trust:    b.Trust,
	})
	return ai.Done
}

func (ReactObserve) Stop(*ai.Ctx, *ai.Doing, bool)     {}
func (ReactObserve) Valid(c *ai.Ctx, _ *ai.Doing) bool { return ecs.Get[Bond](c.W.ECS()).Has(c.Actor) }
func (ReactObserve) Interruptible(*ai.Ctx, *ai.Doing) bool {
	return true
}
