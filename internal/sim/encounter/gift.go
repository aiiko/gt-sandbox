package encounter

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

const (
	// NoticeGiftAction es el nombre de la acción.
	NoticeGiftAction = "notice_gift"
	// NoticeGiftTicks es cuánto se detiene la giantess a mirar un regalo.
	NoticeGiftTicks clock.Tick = 4 * clock.TPS
	// NoticeGiftCooldown evita repetir notice_gift de inmediato en la misma superficie.
	NoticeGiftCooldown clock.Tick = 2 * 60 * clock.TPS
)

// GiftNoticed se emite cuando la giantess nota un regalo dejado por un tiny:
// se detiene, mira y baja su enfriamiento de share en esa superficie.
type GiftNoticed struct {
	Entity      ids.ID `json:"entity"`
	Surface     ids.ID `json:"surface"`
	Trinket     ids.ID `json:"trinket"`
	TrinketKind string `json:"kind"`
	Giver       ids.ID `json:"giver,omitempty"`
}

func (GiftNoticed) Kind() string { return "gift_noticed" }

// NoticeGift es la reacción de la giantess al ver un tesoro dejado como regalo
// en una superficie (M63): se detiene, mira y baja su enfriamiento de share.
//
// Doing: Target = superficie, Item = tesoro regalo; Stage 0 camina al lugar,
// Stage 1 mira el regalo.
type NoticeGift struct{}

func (NoticeGift) Name() string { return NoticeGiftAction }

func (NoticeGift) What(*ai.Ctx, *ai.Doing) string { return NoticeGiftAction }

func (NoticeGift) utility(enc *Encounter) float64 {
	if enc == nil {
		return 0
	}
	if enc.NoticeGift > 0 {
		return enc.NoticeGift
	}
	if enc.Observe > 0 {
		return enc.Observe + 0.08
	}
	return 0
}

func (n NoticeGift) Options(c *ai.Ctx) []ai.Option {
	enc := Of(c)
	u0 := n.utility(enc)
	if enc == nil || u0 <= 0 {
		return nil
	}
	r := c.W.ECS()
	locs := comp.Of(r).Location
	carried := ecs.Get[jobs.Carried](r)
	g := space.Of(r)
	me, _, err := g.WorldPose(c.Actor)
	if err != nil {
		return nil
	}
	maxD2 := int64(c.Agent.Watch) * int64(c.Agent.Watch) * 4 // rango amplio para notar regalos en muebles cercanos

	seenSurf := map[ids.ID]bool{}
	var out []ai.Option
	ecs.Get[colony.Trinket](r).Each(func(id ids.ID, tr *colony.Trinket) {
		if r.Despawning(id) || tr.Giver.IsNil() || carried.Has(id) {
			return
		}
		l, ok := locs.Get(id)
		if !ok || seenSurf[l.Space] || enc.Cooling(l.Space, NoticeGiftAction, c.W.Tick()) {
			return
		}
		sp, ok := spotOf(c.W, l.Space)
		if !ok {
			return
		}
		if maxD2 > 0 {
			if p, _, err := g.WorldPose(id); err != nil || p.XY().DistSq(me.XY()) > maxD2 {
				return
			}
		}
		seenSurf[l.Space] = true
		out = append(out, ai.Option{Target: l.Space, Utility: u0 - c.Agent.DistanceCost*distance(c, sp)})
	})
	return out
}

func (n NoticeGift) Utility(c *ai.Ctx, _ *ai.Doing) float64 { return n.utility(Of(c)) }

func giftOnSurface(w *world.World, surface ids.ID) ids.ID {
	r := w.ECS()
	locs := comp.Of(r).Location
	carried := ecs.Get[jobs.Carried](r)
	var best ids.ID
	ecs.Get[colony.Trinket](r).Each(func(id ids.ID, tr *colony.Trinket) {
		if !best.IsNil() || r.Despawning(id) || tr.Giver.IsNil() || carried.Has(id) {
			return
		}
		if l, ok := locs.Get(id); ok && l.Space == surface {
			best = id
		}
	})
	return best
}

func (NoticeGift) Start(c *ai.Ctx, d *ai.Doing) bool {
	enc := Of(c)
	if _, ok := spotOf(c.W, d.Target); !ok || enc == nil {
		return false
	}
	gift := giftOnSurface(c.W, d.Target)
	if gift.IsNil() {
		return false
	}
	d.Item = gift
	enc.Cool(d.Target, NoticeGiftAction, c.W.Tick()+NoticeGiftCooldown)
	return true
}

func (NoticeGift) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	sp, ok := spotOf(c.W, d.Target)
	if !ok {
		return ai.Failed
	}
	r := c.W.ECS()
	switch d.Stage {
	case 0:
		if res, arrived := goTo(c, d, sp); !arrived {
			return res
		}
		if !r.Alive(d.Item) || r.Despawning(d.Item) {
			return skip(c, d, "gone")
		}
		d.Stage = 1
		fallthrough
	default:
		if d.Progress++; d.Progress < NoticeGiftTicks {
			return ai.Running
		}
		tr, ok := ecs.Get[colony.Trinket](r).Get(d.Item)
		if !ok || tr.Giver.IsNil() {
			return skip(c, d, "gone")
		}
		// Baja el enfriamiento de share en esta superficie.
		if enc := Of(c); enc != nil {
			now := c.W.Tick()
			for i, cd := range enc.Cooldowns {
				if cd.Surface == d.Target && cd.Action == "share" && cd.Until > now {
					rem := cd.Until - now
					enc.Cooldowns[i].Until = now + rem/2
					break
				}
			}
		}
		item := d.Item
		d.Item = ids.Nil
		r.Despawn(item)
		c.W.Emit(GiftNoticed{
			Entity: c.Actor, Surface: d.Target,
			Trinket: item, TrinketKind: tr.Kind, Giver: tr.Giver,
		})
		return ai.Done
	}
}

func (NoticeGift) Stop(c *ai.Ctx, d *ai.Doing, _ bool) { halt(c, d) }

func (NoticeGift) Valid(c *ai.Ctx, d *ai.Doing) bool {
	_, ok := spotOf(c.W, d.Target)
	return ok
}

func (NoticeGift) Interruptible(*ai.Ctx, *ai.Doing) bool { return true }
