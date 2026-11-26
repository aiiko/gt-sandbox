package attach

import (
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/world"
)

const (
	// TickleThreshold es cuánta estimulación acumulada dispara la reacción de cosquillas.
	TickleThreshold int32 = 250
	// TickleDecay es cuánto decrece el medidor por tick sin estímulo activo.
	TickleDecay int32 = 3
	// TicklePauseTicks es cuántos ticks se frena brevemente el cuerpo al sentir cosquillas.
	TicklePauseTicks clock.Tick = 25
	// TickleCooldownTicks evita encadenar cosquillas seguidas tras una reacción.
	TickleCooldownTicks clock.Tick = 45 * clock.TPS
)

// BodyTickled se emite cuando el movimiento de una entidad de otra escala
// dispara una reacción de cosquillas sobre una zona corporal sensible.
type BodyTickled struct {
	Body   ids.ID `json:"body"`
	Zone   ids.ID `json:"zone"`
	By     ids.ID `json:"by"`
	Region string `json:"region,omitempty"`
}

func (BodyTickled) Kind() string { return "body_tickled" }

// tickle acumula cosquillas sobre zonas corporales con Sensitivity > 0 cuando
// otra entidad de distinta escala camina por ellas sin estar sujetada.
type tickle struct{}

func (tickle) Name() string { return "attach.tickle" }

func (tickle) Run(w *world.World) {
	r := w.ECS()
	zones := ecs.Get[BodyZone](r)
	var bodyIDs []ids.ID
	seenBody := map[ids.ID]bool{}
	zones.Each(func(zoneID ids.ID, z *BodyZone) {
		if z.Sensitivity > 0 && !z.Body.IsNil() && !r.Despawning(zoneID) && !r.Despawning(z.Body) && !seenBody[z.Body] {
			seenBody[z.Body] = true
			bodyIDs = append(bodyIDs, z.Body)
		}
	})
	if len(bodyIDs) == 0 {
		return
	}
	slices.Sort(bodyIDs)

	zoneIDs := zones.IDs()
	slices.Sort(zoneIDs)

	movers := ecs.Get[nav.Mover](r)
	moverIDs := movers.IDs()
	slices.Sort(moverIDs)

	locs := comp.Of(r).Location
	grips := ecs.Get[Gripping](r)
	states := ecs.Get[BodyState](r)
	now := w.Tick()

	for _, bodyID := range bodyIDs {
		bodyScale := ""
		if bmv, ok := movers.Get(bodyID); ok {
			bodyScale = bmv.Scale
		}
		var bestZone, triggerBy ids.ID
		var bestSens int32
		var bestRegion string

		for _, zid := range zoneIDs {
			z, ok := zones.Get(zid)
			if !ok || z.Body != bodyID || z.Sensitivity <= 0 || r.Despawning(zid) {
				continue
			}
			for _, eid := range moverIDs {
				if eid == bodyID || r.Despawning(eid) {
					continue
				}
				mv, ok := movers.Get(eid)
				if !ok || mv.Status != nav.Moving || mv.Scale == bodyScale {
					continue
				}
				loc, ok := locs.Get(eid)
				if !ok || loc.Space != zid {
					continue
				}
				if gr, ok := grips.Get(eid); ok && gr.Space == zid {
					continue
				}
				if z.Sensitivity > bestSens || (z.Sensitivity == bestSens && (bestZone.IsNil() || zid < bestZone || (zid == bestZone && eid < triggerBy))) {
					bestSens = z.Sensitivity
					bestZone = zid
					bestRegion = z.Region
					triggerBy = eid
				}
				break // moverIDs está ordenado: la primera entidad activa en esta zona alcanza para determinismo
			}
		}

		hadState := states.Has(bodyID)
		st, _ := states.Get(bodyID)
		changed := false

		if bestSens > 0 && now >= st.TickleCooldown {
			st.Tickle += bestSens
			st.TickleZone = bestZone
			changed = true
			if st.Tickle >= TickleThreshold {
				st.Tickle = 0
				st.TickleCooldown = now + TickleCooldownTicks
				st.TickleUntil = now + TicklePauseTicks
				if bmv := movers.Ptr(bodyID); bmv != nil && bmv.WaitUntil < now+TicklePauseTicks {
					bmv.WaitUntil = now + TicklePauseTicks
				}
				checkGripOnPose(w, []ids.ID{bestZone})
				w.Emit(BodyTickled{
					Body:   bodyID,
					Zone:   bestZone,
					By:     triggerBy,
					Region: bestRegion,
				})
			}
		} else if st.Tickle > 0 {
			st.Tickle = max(0, st.Tickle-TickleDecay)
			if st.Tickle == 0 {
				st.TickleZone = ids.Nil
			}
			changed = true
		}

		if changed || (hadState && (st.TickleUntil > 0 && now >= st.TickleUntil && now >= st.TickleCooldown)) {
			if hadState || st.Tickle > 0 || st.TickleCooldown > now || st.TickleUntil > now {
				states.Set(bodyID, st)
			}
		}
	}
}
