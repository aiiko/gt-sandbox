package colony

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/spatial"
	"gtsandbox/internal/sim/world"
)

// Tipos de tesoros perdidos (M63).
const (
	TrinketButton   = "button"
	TrinketClip     = "clip"
	TrinketThread   = "thread"
	TrinketGraphite = "graphite"
	TrinketCoin     = "coin"
)

// Parámetros de regalos y colección (M63).
const (
	GiftAction           = "gift"
	GiftFunMin     int32 = 700
	GiftUtility          = 0.24
	DefaultGiftRad       = 3 * units.Centimeter
	// HoardFunPerKind es cuánto sube el piso de diversión por cada tipo
	// distinto de tesoro en la colección del nido (tope HoardFunCap).
	HoardFunPerKind int32 = 150
	HoardFunCap     int32 = 600
)

// Trinket es el componente "trinket" de un tesoro perdido (botón, clip, hilo,
// mina de grafito). Giver es el tiny que lo dejó como regalo (Nil si no es
// un regalo colocado).
type Trinket struct {
	Kind  string `json:"kind"`
	Uses  int32  `json:"uses,omitempty"`
	Giver ids.ID `json:"giver,omitempty"`
}

// ValidTrinketKind indica si kind es un tipo conocido de tesoro.
func ValidTrinketKind(kind string) bool {
	switch kind {
	case TrinketButton, TrinketClip, TrinketThread, TrinketGraphite, TrinketCoin:
		return true
	}
	return false
}

// Hoard es el componente "hoard" del nido: la colección de tesoros de la
// colonia (lista ordenada de IDs, D5). Los tesoros siguen vivos para poder
// retirarlos y usarlos (M64, M65) o regalar repetidos.
type Hoard struct {
	Items []ids.ID `json:"items,omitempty"`
}

// Has indica si id está en la colección.
func (h Hoard) Has(id ids.ID) bool {
	_, ok := slices.BinarySearch(h.Items, id)
	return ok
}

// Add agrega id a la colección manteniendo el orden y sin duplicados.
func (h *Hoard) Add(id ids.ID) bool {
	i, ok := slices.BinarySearch(h.Items, id)
	if ok {
		return false
	}
	h.Items = slices.Insert(h.Items, i, id)
	return true
}

// Remove quita id de la colección (p. ej. al retirarlo para usar o regalar).
func (h *Hoard) Remove(id ids.ID) bool {
	i, ok := slices.BinarySearch(h.Items, id)
	if !ok {
		return false
	}
	h.Items = slices.Delete(h.Items, i, i+1)
	return true
}

// DistinctKinds cuenta cuántos tipos distintos de tesoro vivo tiene la colección.
func (h Hoard) DistinctKinds(r *ecs.Registry) int {
	trinkets := ecs.Get[Trinket](r)
	seen := map[string]bool{}
	for _, id := range h.Items {
		if !r.Alive(id) || r.Despawning(id) {
			continue
		}
		if tr, ok := trinkets.Get(id); ok && tr.Kind != "" {
			seen[tr.Kind] = true
		}
	}
	return len(seen)
}

// HasKind indica si la colección ya tiene al menos un tesoro del tipo kind.
func (h Hoard) HasKind(r *ecs.Registry, kind string) bool {
	trinkets := ecs.Get[Trinket](r)
	for _, id := range h.Items {
		if !r.Alive(id) || r.Despawning(id) {
			continue
		}
		if tr, ok := trinkets.Get(id); ok && tr.Kind == kind {
			return true
		}
	}
	return false
}

// Duplicate devuelve el primer tesoro repetido de la colección (un tipo que
// aparece al menos dos veces) que no esté reservado ni cargado, o ids.Nil.
func (h Hoard) Duplicate(w *world.World) ids.ID {
	r := w.ECS()
	trinkets := ecs.Get[Trinket](r)
	carried := ecs.Get[jobs.Carried](r)
	counts := map[string]int{}
	for _, id := range h.Items {
		if !r.Alive(id) || r.Despawning(id) {
			continue
		}
		if tr, ok := trinkets.Get(id); ok && tr.Kind != "" {
			counts[tr.Kind]++
		}
	}
	for _, id := range h.Items {
		if !r.Alive(id) || r.Despawning(id) || carried.Has(id) || !jobs.Holder(w, id).IsNil() {
			continue
		}
		if tr, ok := trinkets.Get(id); ok && counts[tr.Kind] >= 2 {
			return id
		}
	}
	return ids.Nil
}

// GiftSpot es el componente "gift_spot": un lugar visible (borde del
// escritorio, mesa del living) donde un tiny puede dejar un tesoro repetido
// como regalo para la giantess.
type GiftSpot struct {
	Radius units.Length `json:"radius,omitempty"`
}

func (g GiftSpot) radius() units.Length {
	if g.Radius > 0 {
		return g.Radius
	}
	return DefaultGiftRad
}

// Eventos de tesoros y regalos (M63).
type (
	TrinketFound struct {
		Entity      ids.ID `json:"entity"`
		Spot        ids.ID `json:"spot"`
		Trinket     ids.ID `json:"trinket"`
		TrinketKind string `json:"kind"`
	}
	TrinketStored struct {
		Colony      ids.ID `json:"colony"`
		Trinket     ids.ID `json:"trinket"`
		TrinketKind string `json:"kind"`
		FirstOfKind bool   `json:"first_of_kind,omitempty"`
		Total       int    `json:"total"`
	}
	GiftPlaced struct {
		Entity      ids.ID `json:"entity"`
		Trinket     ids.ID `json:"trinket"`
		TrinketKind string `json:"kind"`
		Spot        ids.ID `json:"spot"`
		Surface     ids.ID `json:"surface"`
	}
)

func (TrinketFound) Kind() string  { return "trinket_found" }
func (TrinketStored) Kind() string { return "trinket_stored" }
func (GiftPlaced) Kind() string    { return "gift_placed" }

// spawnTrinket crea una entidad tesoro del tipo kind en loc.
func spawnTrinket(w *world.World, kind string, loc comp.Location) ids.ID {
	r := w.ECS()
	id := r.Spawn()
	cs := comp.Of(r)
	cs.Name.Set(id, comp.Name("trinket_"+kind))
	cs.Location.Set(id, loc)
	cs.Body.Set(id, comp.Body{Radius: 2 * units.Millimeter, Height: 1500, Mass: 0})
	uses := int32(0)
	switch kind {
	case TrinketThread:
		uses = 1
	case TrinketGraphite:
		uses = 3
	}
	ecs.Get[Trinket](r).Set(id, Trinket{Kind: kind, Uses: uses})
	ecs.Get[ai.Landmark](r).Set(id, ai.Landmark{})
	return id
}

// inAnyHoard indica si e pertenece a algún Hoard vivo.
func inAnyHoard(r *ecs.Registry, e ids.ID) bool {
	found := false
	ecs.Get[Hoard](r).Each(func(id ids.ID, h *Hoard) {
		if !found && !r.Despawning(id) && h.Has(e) {
			found = true
		}
	})
	return found
}

// onGiftSpot indica si hay algún tesoro regalo ya colocado cerca de spotID.
func giftSpotOccupied(w *world.World, spotID ids.ID) bool {
	r := w.ECS()
	gs, ok := ecs.Get[GiftSpot](r).Get(spotID)
	loc, ok2 := comp.Of(r).Location.Get(spotID)
	if !ok || !ok2 {
		return true
	}
	trinkets := ecs.Get[Trinket](r)
	carried := ecs.Get[jobs.Carried](r)
	found := false
	spatial.Of(w).Near(loc.Space, loc.Pos.XY(), gs.radius(), func(e ids.ID, _ units.Vec2) {
		if !found && !r.Despawning(e) && !carried.Has(e) {
			if tr, ok := trinkets.Get(e); ok && !tr.Giver.IsNil() {
				found = true
			}
		}
	})
	return found
}

// hoardStore guarda tesoros sueltos en el nido (Hoard), limpia IDs muertos y
// publica acarreos de tesoros conocidos cuando no hay déficit de comida ni agua.
type hoardStore struct{}

func (hoardStore) Name() string { return "colony.hoard" }

func (hoardStore) Run(w *world.World) {
	r := w.ECS()
	locs := comp.Of(r).Location
	trinkets := ecs.Get[Trinket](r)
	hoards := ecs.Get[Hoard](r)
	carried := ecs.Get[jobs.Carried](r)

	// 1. Limpiar IDs muertos y guardar tesoros sueltos que llegaron al nido.
	ecs.Get[Colony](r).Each(func(col ids.ID, c *Colony) {
		if r.Despawning(col) {
			return
		}
		h := hoards.Ptr(col)
		if h == nil {
			return
		}
		h.Items = slices.DeleteFunc(h.Items, func(id ids.ID) bool {
			return !r.Alive(id) || r.Despawning(id)
		})
		nest, ok := locs.Get(col)
		if !ok {
			return
		}
		var loose []ids.ID
		spatial.Of(w).Near(nest.Space, nest.Pos.XY(), c.Radius, func(e ids.ID, _ units.Vec2) {
			if r.Despawning(e) || carried.Has(e) || !jobs.Holder(w, e).IsNil() || h.Has(e) {
				return
			}
			if tr, ok := trinkets.Get(e); ok && tr.Giver.IsNil() {
				loose = append(loose, e)
			}
		})
		slices.Sort(loose)
		for _, e := range loose {
			tr, _ := trinkets.Get(e)
			first := !h.HasKind(r, tr.Kind)
			if h.Add(e) {
				// Fijar el tesoro sobre el nido.
				if l := locs.Ptr(e); l != nil {
					l.Space, l.Pos = nest.Space, nest.Pos
				}
				w.Emit(TrinketStored{
					Colony: col, Trinket: e, TrinketKind: tr.Kind,
					FirstOfKind: first, Total: len(h.Items),
				})
			}
		}

		// 2. Publicar acarreo de tesoros conocidos si no hay déficit de comida ni agua.
		if Deficit(w, col, ai.ResourceFood) > 0 || Deficit(w, col, ai.ResourceWater) > 0 {
			return
		}
		var toHaul []ids.ID
		trinkets.Each(func(e ids.ID, tr *Trinket) {
			if r.Despawning(e) || !tr.Giver.IsNil() || h.Has(e) || inAnyHoard(r, e) ||
				atNest(w, col, e) || hauled(w, e) || !colonyKnows(w, col, e) {
				return
			}
			toHaul = append(toHaul, e)
		})
		slices.Sort(toHaul)
		for _, e := range toHaul {
			_ = jobs.PostJob{Type: "haul", Target: e, DestSpace: nest.Space, Dest: nest.Pos.XY()}.Apply(w)
		}
	})
}

// hoardFun eleva el piso de diversión (fun) y frena su decaimiento según la
// cantidad de tipos distintos de tesoros en la colección del nido.
type hoardFun struct{}

func (hoardFun) Name() string { return "colony.hoard_fun" }

func (hoardFun) Run(w *world.World) {
	r := w.ECS()
	hoards := ecs.Get[Hoard](r)
	needsStore := ecs.Get[needs.Needs](r)
	step := w.Tick() / needs.Interval

	ecs.Get[Member](r).Each(func(e ids.ID, m *Member) {
		if r.Despawning(e) {
			return
		}
		h, ok := hoards.Get(m.Colony)
		if !ok {
			return
		}
		kinds := h.DistinctKinds(r)
		if kinds <= 0 {
			return
		}
		n := needsStore.Ptr(e)
		if n == nil {
			return
		}
		floor := min(int32(kinds)*HoardFunPerKind, HoardFunCap)
		// Frena el decaimiento de fun: restaura 1 punto cada max(1, 5-kinds) intervalos.
		every := clock.Tick(max(1, 5-kinds))
		for i := range *n {
			if (*n)[i].Name != "fun" {
				continue
			}
			if (*n)[i].Decay > 0 && step%every == 0 && (*n)[i].Value < needs.Max {
				(*n)[i].Value = min((*n)[i].Value+1, needs.Max)
			}
			if (*n)[i].Value < floor {
				(*n)[i].Value = floor
			}
			break
		}
	})
}

// --- Acción Gift: llevar un tesoro repetido a un gift_spot ---

// Gift es la acción "gift": un tiny con mucha diversión (fun >= GiftFunMin)
// retira un tesoro repetido del Hoard del nido y lo deja en un gift_spot.
//
// Doing: Target = gift_spot, Item = tesoro repetido; Stage 0 busca el tesoro
// en el nido, Stage 1 lo lleva al gift_spot y lo deposita.
type Gift struct{}

func (Gift) Name() string { return GiftAction }

func (Gift) Options(c *ai.Ctx) []ai.Option {
	if c.Needs == nil {
		return nil
	}
	fun, ok := c.Needs.Get("fun")
	if !ok || fun < GiftFunMin {
		return nil
	}
	r := c.W.ECS()
	m, ok := ecs.Get[Member](r).Get(c.Actor)
	if !ok {
		return nil
	}
	h, ok := ecs.Get[Hoard](r).Get(m.Colony)
	if !ok || h.Duplicate(c.W).IsNil() {
		return nil
	}
	locs := comp.Of(r).Location
	var out []ai.Option
	ecs.Get[GiftSpot](r).Each(func(id ids.ID, _ *GiftSpot) {
		loc, ok := locs.Get(id)
		if !ok || r.Despawning(id) || !c.Agent.Knows(id) {
			return
		}
		if h := jobs.Holder(c.W, id); !h.IsNil() && h != c.Actor {
			return
		}
		if giftSpotOccupied(c.W, id) {
			return
		}
		u := GiftUtility - 0.05*ai.Distance(c, loc.Space, loc.Pos.XY())
		out = append(out, ai.Option{Target: id, Utility: u})
	})
	return out
}

func (Gift) Utility(c *ai.Ctx, _ *ai.Doing) float64 { return GiftUtility }

func (Gift) Start(c *ai.Ctx, d *ai.Doing) bool {
	r := c.W.ECS()
	m, ok := ecs.Get[Member](r).Get(c.Actor)
	if !ok {
		return false
	}
	h, ok := ecs.Get[Hoard](r).Get(m.Colony)
	if !ok {
		return false
	}
	dup := h.Duplicate(c.W)
	if dup.IsNil() {
		return false
	}
	if !jobs.Reserve(c.W, d.Target, c.Actor) {
		return false
	}
	if !jobs.Reserve(c.W, dup, c.Actor) {
		jobs.Release(c.W, d.Target, c.Actor)
		return false
	}
	d.Item = dup
	return true
}

func (Gift) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	w := c.W
	r := w.ECS()
	if !r.Alive(d.Target) || r.Despawning(d.Target) || !r.Alive(d.Item) || r.Despawning(d.Item) {
		return ai.Failed
	}
	if !jobs.Reserve(w, d.Target, c.Actor) || !jobs.Reserve(w, d.Item, c.Actor) {
		return ai.Failed
	}
	locs := comp.Of(r).Location
	switch d.Stage {
	case 0:
		itemLoc, ok := locs.Get(d.Item)
		if !ok {
			return ai.Failed
		}
		switch nav.Approach(w, c.Actor, itemLoc.Space, itemLoc.Pos.XY(), &d.Moving) {
		case nav.Failed:
			return ai.Failed
		case nav.Underway:
			return ai.Running
		}
		// Retirar del Hoard y cargar el tesoro.
		if m, ok := ecs.Get[Member](r).Get(c.Actor); ok {
			if h := ecs.Get[Hoard](r).Ptr(m.Colony); h != nil {
				h.Remove(d.Item)
			}
		}
		ecs.Get[jobs.Carried](r).Set(d.Item, jobs.Carried{By: c.Actor})
		d.Stage, d.Moving = 1, false
		return ai.Running
	default:
		spotLoc, ok := locs.Get(d.Target)
		if !ok {
			return ai.Failed
		}
		switch nav.Approach(w, c.Actor, spotLoc.Space, spotLoc.Pos.XY(), &d.Moving) {
		case nav.Failed:
			return ai.Failed
		case nav.Underway:
			return ai.Running
		}
		ecs.Get[jobs.Carried](r).Remove(d.Item)
		if l := locs.Ptr(d.Item); l != nil {
			l.Space, l.Pos = spotLoc.Space, spotLoc.Pos
		}
		kind := ""
		if tr := ecs.Get[Trinket](r).Ptr(d.Item); tr != nil {
			tr.Giver = c.Actor
			kind = tr.Kind
		}
		item := d.Item
		d.Item = ids.Nil
		jobs.Release(w, item, c.Actor)
		jobs.Release(w, d.Target, c.Actor)
		w.Emit(GiftPlaced{
			Entity: c.Actor, Trinket: item, TrinketKind: kind,
			Spot: d.Target, Surface: spotLoc.Space,
		})
		return ai.Done
	}
}

func (Gift) Stop(c *ai.Ctx, d *ai.Doing, _ bool) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
	}
	if !d.Item.IsNil() {
		ecs.Get[jobs.Carried](c.W.ECS()).Remove(d.Item)
		jobs.Release(c.W, d.Item, c.Actor)
	}
	jobs.Release(c.W, d.Target, c.Actor)
}

func (Gift) Valid(c *ai.Ctx, d *ai.Doing) bool {
	r := c.W.ECS()
	return r.Alive(d.Target) && !r.Despawning(d.Target) &&
		r.Alive(d.Item) && !r.Despawning(d.Item) &&
		jobs.Reserve(c.W, d.Target, c.Actor) && jobs.Reserve(c.W, d.Item, c.Actor)
}

func (Gift) Interruptible(*ai.Ctx, *ai.Doing) bool { return true }

func checkTrinkets(w *world.World) error {
	r := w.ECS()
	var err error
	ecs.Get[Trinket](r).Each(func(id ids.ID, tr *Trinket) {
		if !ValidTrinketKind(tr.Kind) {
			err = errors.Join(err, fmt.Errorf("colony: trinket %v tipo inválido %q", id, tr.Kind))
		}
		if tr.Uses < 0 {
			err = errors.Join(err, fmt.Errorf("colony: trinket %v usos negativos (%d)", id, tr.Uses))
		}
	})
	ecs.Get[Hoard](r).Each(func(id ids.ID, h *Hoard) {
		for i, item := range h.Items {
			if i > 0 && cmp.Compare(item, h.Items[i-1]) <= 0 {
				err = errors.Join(err, fmt.Errorf("colony: hoard %v desordenado o duplicado", id))
			}
		}
	})
	return err
}

func registerTrinkets(w *world.World) error {
	r := w.ECS()
	for _, err := range []error{
		register[Trinket](r, "trinket"),
		register[Hoard](r, "hoard"),
		register[GiftSpot](r, "gift_spot"),
	} {
		if err != nil {
			return err
		}
	}
	w.AddInvariant("colony.trinket", checkTrinkets)
	return errors.Join(
		world.RegisterEvent[TrinketFound](w),
		world.RegisterEvent[TrinketStored](w),
		world.RegisterEvent[GiftPlaced](w),
		w.Register(world.SystemSpec{Phase: clock.PhaseConsequences, Order: (1 << 21) + 1}, hoardStore{}),
		w.Register(world.SystemSpec{Phase: clock.PhaseSlow, Period: needs.Interval, Order: 10}, hoardFun{}),
	)
}
