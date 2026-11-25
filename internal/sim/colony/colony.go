// Package colony implementa colonias de tinies: un nido con depósito de
// comida, miembros que comparten lo que ven, trabajos de acarreo que la
// colonia publica sola y la acción "forage" (salir a buscar).
//
// La colonia es una entidad ubicada en el nido, con "colony" y además un
// smart object: sus ofertas (comer del depósito, descansar) y su depósito
// son componentes de ai ("offers", "portions").
package colony

import (
	"cmp"
	"encoding/json"
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
	"gtsandbox/internal/sim/spatial"
	"gtsandbox/internal/sim/world"
)

// SupportNote recuerda un apoyo observado entre una zona corporal y una
// superficie de la casa junto con la pose en que se abrió (propuesta 19.4).
type SupportNote struct {
	BodySpace  ids.ID     `json:"body_space"`
	HouseSpace ids.ID     `json:"house_space"`
	Pose       string     `json:"pose,omitempty"`
	Tick       clock.Tick `json:"tick"`
}

// SupportDiscovered se emite cuando un expedicionario o explorador comunica a
// la colonia un apoyo útil descubierto en el cuerpo de la giantess.
type SupportDiscovered struct {
	Colony     ids.ID `json:"colony"`
	Entity     ids.ID `json:"entity"`
	BodySpace  ids.ID `json:"body_space"`
	HouseSpace ids.ID `json:"house_space"`
	Pose       string `json:"pose,omitempty"`
}

func (SupportDiscovered) Kind() string { return "support_discovered" }

// StockGoal es la meta de un recurso en el depósito de la colonia, en
// porciones por miembro vivo.
type StockGoal struct {
	Resource string `json:"resource"`
	Amount   int32  `json:"amount"`
}

// Stock guarda las metas del depósito por recurso, serializado como lista
// ordenada por nombre (D5). En JSON también acepta un entero (meta de "food")
// o un mapa {"food": 4, "water": 2}.
type Stock []StockGoal

// NewStock crea un Stock con una sola meta para el recurso dado.
func NewStock(res string, amount int32) Stock {
	var s Stock
	s.Set(res, amount)
	return s
}

// Get devuelve la meta por miembro vivo del recurso res ("" = "food").
func (s Stock) Get(res string) int32 {
	res = ai.NormalizeResource(res)
	for _, g := range s {
		if g.Resource == res {
			return g.Amount
		}
	}
	return 0
}

// Set fija la meta por miembro vivo del recurso res manteniendo el orden (D5).
func (s *Stock) Set(res string, amount int32) {
	res = ai.NormalizeResource(res)
	for i, g := range *s {
		if g.Resource == res {
			if amount <= 0 {
				*s = slices.Delete(*s, i, i+1)
			} else {
				(*s)[i].Amount = amount
			}
			return
		}
	}
	if amount <= 0 {
		return
	}
	*s = append(*s, StockGoal{Resource: res, Amount: amount})
	slices.SortFunc(*s, func(a, b StockGoal) int { return cmp.Compare(a.Resource, b.Resource) })
}

// MarshalJSON serializa siempre como lista ordenada por nombre (D5).
func (s Stock) MarshalJSON() ([]byte, error) {
	out := make([]StockGoal, 0, len(s))
	for _, g := range s {
		if g.Amount > 0 {
			out = append(out, StockGoal{Resource: ai.NormalizeResource(g.Resource), Amount: g.Amount})
		}
	}
	slices.SortFunc(out, func(a, b StockGoal) int { return cmp.Compare(a.Resource, b.Resource) })
	return json.Marshal(out)
}

// UnmarshalJSON acepta entero (meta de food), objeto {"food":4,"water":2} o lista ordenada.
func (s *Stock) UnmarshalJSON(data []byte) error {
	*s = nil
	trimmed := data
	for len(trimmed) > 0 && (trimmed[0] == ' ' || trimmed[0] == '\t' || trimmed[0] == '\n' || trimmed[0] == '\r') {
		trimmed = trimmed[1:]
	}
	if len(trimmed) == 0 {
		return nil
	}
	switch trimmed[0] {
	case '[':
		var list []StockGoal
		if err := json.Unmarshal(data, &list); err != nil {
			return err
		}
		for _, g := range list {
			s.Set(g.Resource, g.Amount)
		}
	case '{':
		var m map[string]int32
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		for k, v := range m {
			s.Set(k, v)
		}
	default:
		var n int32
		if err := json.Unmarshal(data, &n); err != nil {
			return err
		}
		s.Set(ai.ResourceFood, n)
	}
	return nil
}

// Colony es el componente "colony" de la entidad nido. Lo que llega a menos
// de Radius del nido se guarda en el depósito.
type Colony struct {
	Radius units.Length `json:"radius"`
	// DangerRadius y DangerTicks son el tamaño y la duración de la zona de
	// peligro que deja la muerte de un miembro (0 = por defecto).
	DangerRadius units.Length `json:"danger_radius,omitempty"`
	DangerTicks  clock.Tick   `json:"danger_ticks,omitempty"`
	// Stock es la meta del depósito por recurso, en porciones por miembro vivo.
	// Ver Deficit.
	Stock Stock `json:"stock,omitempty"`
	// KnownSupports guarda los apoyos cuerpo↔casa reportados a la colonia.
	KnownSupports []SupportNote `json:"known_supports,omitempty"`
}

// Member es el componente "member" de un tiny: a qué colonia pertenece,
// cuánto le gusta salir a buscar cuando no tiene nada mejor (utilidad),
// su oficio preferido (Role) y su inclinación por la aventura (Adventure).
type Member struct {
	Colony ids.ID  `json:"colony"`
	Forage float64 `json:"forage,omitempty"`
	// Search es la utilidad de revisar un punto de búsqueda con la despensa
	// vacía (se escala por Deficit). Ver la acción Search.
	Search float64 `json:"search,omitempty"`
	// Role es el oficio preferido: "", "gatherer", "porter", "builder", "lookout".
	Role string `json:"role,omitempty"`
	// Adventure es la inclinación por participar en expediciones [0, 1].
	Adventure float64 `json:"adventure,omitempty"`
}

// Food marca algo que la colonia junta (componente "food"). Al guardarlo
// suma sus porciones (o 1) al depósito bajo "food".
type Food struct{}

// Water marca agua que la colonia puede acarrear al nido (componente "water").
type Water struct{}

// WaterDrop describe una gota o charco de agua (componente "water_drop").
// Volume refleja cuántas tomas/porciones de agua le quedan; se sincroniza con
// ai.Portions ("water") cuando un tiny toma de ella.
type WaterDrop struct {
	Volume int32 `json:"volume"`
}

// Stored se emite al guardar comida o agua en el depósito.
type Stored struct {
	Colony   ids.ID `json:"colony"`
	Item     ids.ID `json:"item"`
	Resource string `json:"resource,omitempty"`
	Portions int32  `json:"portions"`
	Total    int32  `json:"total"`
}

func (Stored) Kind() string { return "stored" }

// share: los miembros conocen su nido y se cuentan lo que vieron; la comida
// nueva genera un trabajo de acarreo al nido.
type share struct{}

func (share) Name() string { return "colony.share" }

func (share) Run(w *world.World) {
	r := w.ECS()
	members := ecs.Get[Member](r)
	agents := ecs.Get[ai.Agent](r)
	members.Each(func(e ids.ID, m *Member) {
		if a := agents.Ptr(e); a != nil && !r.Despawning(e) {
			ai.Learn(w, a, e, m.Colony)
		}
	})
	for _, ev := range w.Events() { // del tick anterior
		n, ok := ev.Event.(ai.Noticed)
		if !ok {
			continue
		}
		m, ok := members.Get(n.Entity)
		if !ok || !r.Alive(n.Target) || r.Despawning(n.Target) || inExpeditionField(w, n.Entity) {
			continue
		}
		members.Each(func(e ids.ID, o *Member) {
			if a := agents.Ptr(e); a != nil && o.Colony == m.Colony && e != n.Entity && !r.Despawning(e) && !inExpeditionField(w, e) {
				ai.Learn(w, a, e, n.Target)
			}
		})
		if (ecs.Get[Food](r).Has(n.Target) || ecs.Get[Water](r).Has(n.Target)) && !atNest(w, m.Colony, n.Target) && !hauled(w, n.Target) {
			nest, ok := comp.Of(r).Location.Get(m.Colony)
			if !ok {
				continue
			}
			_ = jobs.PostJob{Type: "haul", Target: n.Target, DestSpace: nest.Space, Dest: nest.Pos.XY()}.Apply(w)
		}
	}
}

// Deficit es cuánto le falta al depósito de la colonia col para llegar a
// su meta del recurso dado (Colony.Stock.Get(resource) por miembro vivo), en
// [0, 1]: 0 en la meta o sin meta, 1 vacío. Si resource == "" usa "food".
// Se calcula al vuelo: no es estado.
func Deficit(w *world.World, col ids.ID, resource string) float64 {
	resource = ai.NormalizeResource(resource)
	r := w.ECS()
	c, ok := ecs.Get[Colony](r).Get(col)
	if !ok {
		return 0
	}
	stock := c.Stock.Get(resource)
	if stock <= 0 {
		return 0
	}
	n := int32(0)
	ecs.Get[Member](r).Each(func(e ids.ID, m *Member) {
		if m.Colony == col && !r.Despawning(e) {
			n++
		}
	})
	goal := stock * n
	if goal <= 0 {
		return 0
	}
	p, _ := ecs.Get[ai.Portions](r).Get(col)
	have := p.Get(resource)
	if have >= goal {
		return 0
	}
	return float64(goal-max(have, 0)) / float64(goal)
}

// hauled: ya hay un trabajo sobre e, o alguien lo carga.
func hauled(w *world.World, e ids.ID) bool {
	if ecs.Get[jobs.Carried](w.ECS()).Has(e) {
		return true
	}
	found := false
	ecs.Get[jobs.Job](w.ECS()).Each(func(id ids.ID, j *jobs.Job) {
		found = found || j.Target == e && !w.ECS().Despawning(id)
	})
	return found
}

// atNest: e está a menos de Radius del nido col.
func atNest(w *world.World, col, e ids.ID) bool {
	c, ok := ecs.Get[Colony](w.ECS()).Get(col)
	locs := comp.Of(w.ECS()).Location
	nest, ok2 := locs.Get(col)
	l, ok3 := locs.Get(e)
	return ok && ok2 && ok3 && l.Space == nest.Space &&
		l.Pos.XY().DistSq(nest.Pos.XY()) <= int64(c.Radius)*int64(c.Radius)
}

// store guarda en el depósito la comida y el agua sueltas que están en el nido,
// y sincroniza WaterDrop.Volume con Portions("water").
type store struct{}

func (store) Name() string { return "colony.store" }

func (store) Run(w *world.World) {
	r := w.ECS()
	locs := comp.Of(r).Location
	food := ecs.Get[Food](r)
	water := ecs.Get[Water](r)
	drops := ecs.Get[WaterDrop](r)
	portions := ecs.Get[ai.Portions](r)

	// Sincronizar WaterDrop.Volume con Portions("water") tras consumos o al nacer.
	drops.Each(func(id ids.ID, d *WaterDrop) {
		if r.Despawning(id) {
			return
		}
		if p, ok := portions.Get(id); ok {
			d.Volume = p.Get(ai.ResourceWater)
		} else if d.Volume > 0 {
			portions.Set(id, ai.NewPortions(ai.ResourceWater, d.Volume))
		}
	})

	ecs.Get[Colony](r).Each(func(col ids.ID, c *Colony) {
		nest, ok := locs.Get(col)
		if !ok || r.Despawning(col) {
			return
		}
		var items []ids.ID
		spatial.Of(w).Near(nest.Space, nest.Pos.XY(), c.Radius, func(e ids.ID, _ units.Vec2) {
			if (food.Has(e) || water.Has(e)) && !r.Despawning(e) && !ecs.Get[jobs.Carried](r).Has(e) && jobs.Holder(w, e).IsNil() {
				items = append(items, e)
			}
		})
		for _, e := range items {
			res := ai.ResourceFood
			if water.Has(e) && !food.Has(e) {
				res = ai.ResourceWater
			}
			n := int32(1)
			if p, ok := portions.Get(e); ok && p.Get(res) > 0 {
				n = p.Get(res)
			} else if res == ai.ResourceWater {
				if d, ok := drops.Get(e); ok && d.Volume > 0 {
					n = d.Volume
				}
			}
			dep, _ := portions.Get(col)
			total := dep.Add(res, n)
			portions.Set(col, dep)
			r.Despawn(e)
			w.Emit(Stored{Colony: col, Item: e, Resource: res, Portions: n, Total: total})
		}
	})
}

func check(w *world.World) error {
	r := w.ECS()
	var err error
	ecs.Get[Member](r).Each(func(e ids.ID, m *Member) {
		if r.Alive(m.Colony) && !ecs.Get[Colony](r).Has(m.Colony) {
			err = errors.Join(err, fmt.Errorf("colony: %v es miembro de %v, que no es una colonia", e, m.Colony))
		}
	})
	ecs.Get[WaterDrop](r).Each(func(e ids.ID, d *WaterDrop) {
		if d.Volume < 0 {
			err = errors.Join(err, fmt.Errorf("colony: %v tiene volumen de agua negativo (%d)", e, d.Volume))
		}
	})
	ecs.Get[Colony](r).Each(func(e ids.ID, c *Colony) {
		for i, g := range c.Stock {
			if g.Resource == "" || g.Amount <= 0 {
				err = errors.Join(err, fmt.Errorf("colony %v: stock inválido %+v", e, g))
			}
			if i > 0 && g.Resource <= c.Stock[i-1].Resource {
				err = errors.Join(err, fmt.Errorf("colony %v: stock sin ordenar o duplicado", e))
			}
		}
	})
	return err
}

// Install registra stores, eventos, sistemas e invariantes. Requiere ai y
// jobs; las acciones de Actions se pasan a ai.Install.
func Install(w *world.World) error {
	r := w.ECS()
	for _, err := range []error{
		register[Colony](r, "colony"), register[Member](r, "member"),
		register[Food](r, "food"), register[Water](r, "water"), register[WaterDrop](r, "water_drop"),
	} {
		if err != nil {
			return err
		}
	}
	w.AddInvariant("colony", check)
	registerRoles(w)
	return errors.Join(
		registerSearch(w),
		registerBuild(w),
		registerGather(w),
		registerLookout(w),
		registerProvisions(w),
		registerExpeditions(w),
		registerTrinkets(w),
		world.RegisterEvent[Stored](w),
		world.RegisterEvent[MemberLost](w),
		world.RegisterEvent[DangerExpired](w),
		// Después de crush.resolve (Order -1) y antes de jobs.upkeep.
		w.Register(world.SystemSpec{Phase: clock.PhaseConsequences, Order: 0}, mourn{}),
		// Después de ai.look (lo visto este tick se comparte el siguiente) y
		// antes de ai.decide.
		w.Register(world.SystemSpec{Phase: clock.PhaseDecide, Order: -40}, share{}),
		w.Register(world.SystemSpec{Phase: clock.PhaseConsequences, Order: 1 << 21}, store{}),
	)
}

// Actions son las acciones de la colonia, para ai.Install.
func Actions() []ai.Action {
	return []ai.Action{Forage{}, Search{}, Chat{}, Watch{}, Scout{}, EatRation{}, ExpeditionAction{}, Gift{}}
}

func register[T any](r *ecs.Registry, name string) error {
	_, err := ecs.Register[T](r, name)
	return err
}
