package attach

import (
	"cmp"
	"fmt"
	"slices"

	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// GarmentState describe el estado de una prenda explorable.
type GarmentState string

const (
	GarmentStatePlaced  GarmentState = "placed"  // apoyada en un mueble o piso
	GarmentStateDonning GarmentState = "donning" // poniéndosela: ventana de aviso previa
	GarmentStateWorn    GarmentState = "worn"    // montada sobre MountPart del cuerpo
	GarmentStateDoffing GarmentState = "doffing" // sacándosela (opcional, breve)
)

const (
	// DefaultDonWarnTicks es el aviso base (en ticks) al ponerse una prenda.
	DefaultDonWarnTicks clock.Tick = 20
	// DefaultGarmentSource es la causa de aplastamiento por defecto al calzarse.
	DefaultGarmentSource = "garment"
)

// InnerPressure define la región interior de la prenda que se comprime
// al completarse el calce (donning -> worn). Todo lo que queda fuera de
// Zone (borde, puntera, bolsillo) es seguro durante el calce.
type InnerPressure struct {
	Zone   units.Rect `json:"zone"`
	Force  int32      `json:"force"`
	Source string     `json:"source,omitempty"` // por defecto "garment"
}

// EffectiveSource devuelve el Source de la presión o DefaultGarmentSource.
func (p InnerPressure) EffectiveSource() string {
	if p.Source != "" {
		return p.Source
	}
	return DefaultGarmentSource
}

// Garment es el componente "garment" de una prenda explorable (un Space rígido
// que alterna entre apoyado en la casa y montado sobre el cuerpo).
type Garment struct {
	Owner ids.ID       `json:"owner"`
	Slot  string       `json:"slot"` // único por Owner en estado worn (ej. "slipper_left", "cardigan")
	State GarmentState `json:"state"`

	// Ventana de transición (donning / doffing).
	StateUntil clock.Tick `json:"state_until,omitempty"`
	// Destino pendiente cuando State == "doffing".
	DoffSpace  ids.ID          `json:"doff_space,omitempty"`
	DoffPos    units.Vec3      `json:"doff_pos,omitempty"`
	DoffFacing fixedmath.Angle `json:"doff_facing,omitempty"`

	// Configuración al estar puesta (mounted sobre MountPart).
	MountPart      ids.ID            `json:"mount_part"`
	WornPos        units.Vec3        `json:"worn_pos"`
	WornFacing     fixedmath.Angle   `json:"worn_facing,omitempty"`
	WornProjection *space.Projection `json:"worn_projection,omitempty"`

	// Proyección al estar apoyada (nil = plano XY).
	PlacedProjection *space.Projection `json:"placed_projection,omitempty"`

	// Portales internos prenda <-> zonas del cuerpo (p. ej. pantufla <-> tobillo).
	// Habilitados solo cuando State == "worn"; se alterna Disabled sin recrearlos.
	BodyPortals []ids.ID `json:"body_portals,omitempty"`

	// Presión interior al completar el calce (nil para prendas/bolsillos sin compresión).
	InnerPressure *InnerPressure `json:"inner_pressure,omitempty"`

	// Aviso base en ticks antes de completar el calce (0 = DefaultDonWarnTicks).
	DonWarnTicks clock.Tick `json:"don_warn_ticks,omitempty"`
}

func validGarmentState(st GarmentState) bool {
	switch st {
	case GarmentStatePlaced, GarmentStateDonning, GarmentStateWorn, GarmentStateDoffing:
		return true
	}
	return false
}

// DonGarment inicia el calce de Garment por Actor abriendo la ventana donning.
type DonGarment struct {
	Actor     ids.ID `json:"actor"`
	Garment   ids.ID `json:"garment"`
	WarnTicks int64  `json:"warn_ticks,omitempty"`
}

func (DonGarment) Kind() string { return "don_garment" }

func (c DonGarment) Apply(w *world.World) error {
	r := w.ECS()
	if !r.Alive(c.Actor) || r.Despawning(c.Actor) {
		return fmt.Errorf("don_garment: actor %v no vivo", c.Actor)
	}
	if !r.Alive(c.Garment) || r.Despawning(c.Garment) {
		return fmt.Errorf("don_garment: prenda %v no viva", c.Garment)
	}
	g := space.Of(r)
	if !g.IsSpace(c.Garment) {
		return fmt.Errorf("don_garment: %v no es un Space", c.Garment)
	}
	garments := ecs.Get[Garment](r)
	gar, ok := garments.Get(c.Garment)
	if !ok {
		return fmt.Errorf("don_garment: %v no tiene componente garment", c.Garment)
	}
	if gar.Owner != c.Actor {
		return fmt.Errorf("don_garment: %v no pertenece a %v", c.Garment, c.Actor)
	}
	if gar.Slot == "" {
		return fmt.Errorf("don_garment: %v sin slot", c.Garment)
	}
	if gar.State != GarmentStatePlaced {
		return fmt.Errorf("don_garment: %v en estado %q (se esperaba placed)", c.Garment, gar.State)
	}
	if err := validateMountTarget(r, g, c.Garment, gar); err != nil {
		return fmt.Errorf("don_garment: %w", err)
	}
	if hasOtherActiveSlot(r, garments, c.Garment, gar.Owner, gar.Slot) {
		return fmt.Errorf("don_garment: %v ya usa el slot %q", gar.Owner, gar.Slot)
	}
	warn := effectiveWarnTicks(c.WarnTicks, gar.DonWarnTicks)
	ptr := garments.Ptr(c.Garment)
	ptr.State = GarmentStateDonning
	ptr.StateUntil = w.Tick() + warn
	w.Emit(GarmentDonning{Actor: c.Actor, Garment: c.Garment, Slot: ptr.Slot, Until: ptr.StateUntil})
	return nil
}

func effectiveWarnTicks(override int64, base clock.Tick) clock.Tick {
	if override < 0 {
		return 1
	}
	if override > 0 {
		return clock.Tick(override)
	}
	if base > 0 {
		return base
	}
	return DefaultDonWarnTicks
}

func validateMountTarget(r *ecs.Registry, g space.Graph, garID ids.ID, gar Garment) error {
	if !r.Alive(gar.MountPart) || r.Despawning(gar.MountPart) || !g.IsSpace(gar.MountPart) {
		return fmt.Errorf("mount_part %v inválido", gar.MountPart)
	}
	if gar.MountPart != gar.Owner && !mountedOn(g, gar.MountPart, gar.Owner) {
		return fmt.Errorf("mount_part %v no pertenece a %v", gar.MountPart, gar.Owner)
	}
	if err := g.CanPlace(garID, gar.MountPart); err != nil {
		return err
	}
	garRoot, err1 := g.Root(garID)
	mntRoot, err2 := g.Root(gar.MountPart)
	if err1 != nil || err2 != nil || garRoot != mntRoot {
		return fmt.Errorf("mount_part %v en otra raíz", gar.MountPart)
	}
	if gar.WornProjection != nil {
		sp, _ := ecs.Get[space.Space](r).Get(garID)
		if err := gar.WornProjection.Valid(sp.Extent); err != nil {
			return err
		}
	}
	return nil
}

func hasOtherActiveSlot(r *ecs.Registry, garments *ecs.Store[Garment], self, owner ids.ID, slot string) bool {
	dup := false
	garments.Each(func(id ids.ID, g *Garment) {
		if dup || id == self || r.Despawning(id) {
			return
		}
		if g.Owner == owner && g.Slot == slot && g.State != GarmentStatePlaced {
			dup = true
		}
	})
	return dup
}

// CancelDon cancela un calce en curso y devuelve la prenda a placed.
type CancelDon struct {
	Actor   ids.ID `json:"actor"`
	Garment ids.ID `json:"garment"`
}

func (CancelDon) Kind() string { return "cancel_don" }

func (c CancelDon) Apply(w *world.World) error {
	r := w.ECS()
	if !r.Alive(c.Garment) || r.Despawning(c.Garment) {
		return fmt.Errorf("cancel_don: prenda %v no viva", c.Garment)
	}
	garments := ecs.Get[Garment](r)
	gar := garments.Ptr(c.Garment)
	if gar == nil {
		return fmt.Errorf("cancel_don: %v no tiene componente garment", c.Garment)
	}
	if gar.Owner != c.Actor {
		return fmt.Errorf("cancel_don: %v no pertenece a %v", c.Garment, c.Actor)
	}
	if gar.State != GarmentStateDonning {
		return fmt.Errorf("cancel_don: %v en estado %q (se esperaba donning)", c.Garment, gar.State)
	}
	gar.State = GarmentStatePlaced
	gar.StateUntil = 0
	w.Emit(GarmentDonCancelled{Actor: c.Actor, Garment: c.Garment, Slot: gar.Slot})
	return nil
}

// DoffGarment quita una prenda puesta y la apoya en TargetSpace en (Pos, Facing).
type DoffGarment struct {
	Actor       ids.ID          `json:"actor"`
	Garment     ids.ID          `json:"garment"`
	TargetSpace ids.ID          `json:"target_space"`
	Pos         units.Vec3      `json:"pos"`
	Facing      fixedmath.Angle `json:"facing,omitempty"`
}

func (DoffGarment) Kind() string { return "doff_garment" }

func (c DoffGarment) Apply(w *world.World) error {
	r := w.ECS()
	if !r.Alive(c.Actor) || r.Despawning(c.Actor) {
		return fmt.Errorf("doff_garment: actor %v no vivo", c.Actor)
	}
	if !r.Alive(c.Garment) || r.Despawning(c.Garment) {
		return fmt.Errorf("doff_garment: prenda %v no viva", c.Garment)
	}
	if !r.Alive(c.TargetSpace) || r.Despawning(c.TargetSpace) {
		return fmt.Errorf("doff_garment: destino %v no vivo", c.TargetSpace)
	}
	g := space.Of(r)
	if !g.IsSpace(c.Garment) || !g.IsSpace(c.TargetSpace) {
		return fmt.Errorf("doff_garment: prenda o destino no es un Space")
	}
	garments := ecs.Get[Garment](r)
	gar, ok := garments.Get(c.Garment)
	if !ok {
		return fmt.Errorf("doff_garment: %v no tiene componente garment", c.Garment)
	}
	if gar.Owner != c.Actor {
		return fmt.Errorf("doff_garment: %v no pertenece a %v", c.Garment, c.Actor)
	}
	if gar.State != GarmentStateWorn {
		return fmt.Errorf("doff_garment: %v en estado %q (se esperaba worn)", c.Garment, gar.State)
	}
	if err := g.CanPlace(c.Garment, c.TargetSpace); err != nil {
		return fmt.Errorf("doff_garment: %w", err)
	}
	garRoot, err1 := g.Root(c.Garment)
	dstRoot, err2 := g.Root(c.TargetSpace)
	if err1 != nil || err2 != nil || garRoot != dstRoot {
		return fmt.Errorf("doff_garment: destino %v en otra raíz", c.TargetSpace)
	}
	if gar.PlacedProjection != nil {
		sp, _ := ecs.Get[space.Space](r).Get(c.Garment)
		if err := gar.PlacedProjection.Valid(sp.Extent); err != nil {
			return fmt.Errorf("doff_garment: %w", err)
		}
	}
	executeDoff(w, g, c.Garment, garments.Ptr(c.Garment), c.TargetSpace, c.Pos, c.Facing)
	return nil
}

func executeDoff(w *world.World, g space.Graph, garID ids.ID, gar *Garment, targetSpace ids.ID, pos units.Vec3, facing fixedmath.Angle) {
	r := w.ECS()
	setBodyPortalsDisabled(r, gar.BodyPortals, true)
	endGarmentContacts(w, garID)
	_, _ = g.Reparent(garID, targetSpace)
	if loc := comp.Of(r).Location.Ptr(garID); loc != nil {
		loc.Space = targetSpace
		loc.Pos = pos
		loc.Facing = facing
	}
	if sp := ecs.Get[space.Space](r).Ptr(garID); sp != nil {
		sp.Projection = gar.PlacedProjection.Clone()
	}
	gar.State = GarmentStatePlaced
	gar.StateUntil = 0
	gar.DoffSpace = ids.Nil
	gar.DoffPos = units.Vec3{}
	gar.DoffFacing = 0
	w.Emit(GarmentDoffed{Actor: gar.Owner, Garment: garID, Slot: gar.Slot, Surface: targetSpace})
}

// Eventos de prendas.
type (
	GarmentDonning struct {
		Actor   ids.ID     `json:"actor"`
		Garment ids.ID     `json:"garment"`
		Slot    string     `json:"slot"`
		Until   clock.Tick `json:"until"`
	}
	GarmentDonCancelled struct {
		Actor   ids.ID `json:"actor"`
		Garment ids.ID `json:"garment"`
		Slot    string `json:"slot"`
	}
	GarmentWorn struct {
		Actor     ids.ID `json:"actor"`
		Garment   ids.ID `json:"garment"`
		Slot      string `json:"slot"`
		MountPart ids.ID `json:"mount_part"`
	}
	GarmentDoffed struct {
		Actor   ids.ID `json:"actor"`
		Garment ids.ID `json:"garment"`
		Slot    string `json:"slot"`
		Surface ids.ID `json:"surface"`
	}
)

func (GarmentDonning) Kind() string      { return "garment_donning" }
func (GarmentDonCancelled) Kind() string { return "garment_don_cancelled" }
func (GarmentWorn) Kind() string         { return "garment_worn" }
func (GarmentDoffed) Kind() string       { return "garment_doffed" }

// PendingGarmentPressure marca una prenda que acaba de completar donning -> worn
// en el tick actual para que crush.resolve evalúe su compresión interior.
type PendingGarmentPressure struct {
	Garment  ids.ID
	Owner    ids.ID
	Pressure InnerPressure
	Tick     clock.Tick
}

type garmentRuntime struct {
	pressures []PendingGarmentPressure
	justWorn  map[ids.ID]clock.Tick
}

func getGarmentRuntime(w *world.World) *garmentRuntime {
	rt := world.Resource[garmentRuntime](w)
	if rt == nil {
		rt = &garmentRuntime{justWorn: map[ids.ID]clock.Tick{}}
		world.SetResource(w, rt)
	}
	return rt
}

// ConsumeGarmentPressures devuelve y limpia las compresiones interiores
// pendientes del tick actual, en orden ascendente de Garment ID (D5).
func ConsumeGarmentPressures(w *world.World) []PendingGarmentPressure {
	rt := world.Resource[garmentRuntime](w)
	if rt == nil || len(rt.pressures) == 0 {
		return nil
	}
	tick := w.Tick()
	var out []PendingGarmentPressure
	for _, p := range rt.pressures {
		if p.Tick == tick {
			out = append(out, p)
		}
	}
	rt.pressures = nil
	slices.SortFunc(out, func(a, b PendingGarmentPressure) int { return cmp.Compare(a.Garment, b.Garment) })
	return out
}

func isJustWorn(w *world.World, id ids.ID) bool {
	rt := world.Resource[garmentRuntime](w)
	if rt == nil {
		return false
	}
	t, ok := rt.justWorn[id]
	return ok && t == w.Tick()
}

func setBodyPortalsDisabled(r *ecs.Registry, portalIDs []ids.ID, disabled bool) {
	portals := ecs.Get[nav.Portal](r)
	for _, pid := range portalIDs {
		if !r.Alive(pid) || r.Despawning(pid) {
			continue
		}
		p := portals.Ptr(pid)
		if p == nil || p.Disabled == disabled {
			continue
		}
		p.Disabled = disabled
		p.Version++
	}
}

func endGarmentContacts(w *world.World, garID ids.ID) {
	r := w.ECS()
	cs := ecs.Get[Contact](r)
	var toEnd []ids.ID
	for _, cid := range cs.IDs() {
		if c, ok := cs.Get(cid); ok && (c.A == garID || c.B == garID) {
			toEnd = append(toEnd, cid)
		}
	}
	slices.Sort(toEnd)
	for _, cid := range toEnd {
		end(w, cid)
	}
}

// garments completa las transiciones donning/doffing de prendas y limpia las
// prendas cuyo dueño o MountPart dejaron de existir.
type garments struct{}

func (garments) Name() string { return "attach.garments" }

func (garments) Run(w *world.World) {
	r := w.ECS()
	g := space.Of(r)
	store := ecs.Get[Garment](r)
	spaces := ecs.Get[space.Space](r)
	zones := ecs.Get[BodyZone](r)
	locs := comp.Of(r).Location
	rt := getGarmentRuntime(w)
	// Limpiar marcas de ticks anteriores.
	for id, tk := range rt.justWorn {
		if tk < w.Tick() {
			delete(rt.justWorn, id)
		}
	}

	all := slices.Clone(store.IDs())
	slices.Sort(all)
	for _, id := range all {
		if r.Despawning(id) {
			continue
		}
		gar := store.Ptr(id)
		if gar == nil {
			continue
		}
		ownerAlive := r.Alive(gar.Owner) && !r.Despawning(gar.Owner)
		mountAlive := r.Alive(gar.MountPart) && !r.Despawning(gar.MountPart) && g.IsSpace(gar.MountPart)
		if !ownerAlive || !mountAlive {
			switch gar.State {
			case GarmentStateDonning:
				gar.State = GarmentStatePlaced
				gar.StateUntil = 0
				setBodyPortalsDisabled(r, gar.BodyPortals, true)
			case GarmentStateDoffing:
				gar.State = GarmentStatePlaced
				gar.StateUntil = 0
				gar.DoffSpace = ids.Nil
				setBodyPortalsDisabled(r, gar.BodyPortals, true)
			case GarmentStateWorn:
				setBodyPortalsDisabled(r, gar.BodyPortals, true)
				gar.State = GarmentStatePlaced
				gar.StateUntil = 0
			}
			continue
		}

		switch gar.State {
		case GarmentStateDonning:
			if w.Tick() < gar.StateUntil {
				continue
			}
			if err := validateMountTarget(r, g, id, *gar); err != nil || hasOtherActiveSlot(r, store, id, gar.Owner, gar.Slot) {
				gar.State = GarmentStatePlaced
				gar.StateUntil = 0
				setBodyPortalsDisabled(r, gar.BodyPortals, true)
				w.Emit(GarmentDonCancelled{Actor: gar.Owner, Garment: id, Slot: gar.Slot})
				continue
			}
			if gar.InnerPressure != nil {
				rt.pressures = append(rt.pressures, PendingGarmentPressure{
					Garment:  id,
					Owner:    gar.Owner,
					Pressure: *gar.InnerPressure,
					Tick:     w.Tick(),
				})
			}
			endGarmentContacts(w, id)
			rt.justWorn[id] = w.Tick()
			_, _ = g.Reparent(id, gar.MountPart)
			if loc := locs.Ptr(id); loc != nil {
				loc.Space = gar.MountPart
				loc.Pos = gar.WornPos
				loc.Facing = gar.WornFacing
			}
			if sp := spaces.Ptr(id); sp != nil {
				sp.Projection = gar.WornProjection.Clone()
			}
			gar.State = GarmentStateWorn
			gar.StateUntil = 0
			setBodyPortalsDisabled(r, gar.BodyPortals, false)
			if zone, hasZone := zones.Get(id); (hasZone && zone.RequireGrip) || gar.WornProjection != nil {
				checkGripOnPose(w, []ids.ID{id})
			}
			w.Emit(GarmentWorn{Actor: gar.Owner, Garment: id, Slot: gar.Slot, MountPart: gar.MountPart})

		case GarmentStateDoffing:
			if w.Tick() < gar.StateUntil {
				continue
			}
			if r.Alive(gar.DoffSpace) && !r.Despawning(gar.DoffSpace) && g.IsSpace(gar.DoffSpace) && g.CanPlace(id, gar.DoffSpace) == nil {
				executeDoff(w, g, id, gar, gar.DoffSpace, gar.DoffPos, gar.DoffFacing)
			} else {
				gar.State = GarmentStateWorn
				gar.StateUntil = 0
				gar.DoffSpace = ids.Nil
			}
		}
	}
}

func checkGarments(w *world.World) error {
	r := w.ECS()
	store := ecs.Get[Garment](r)
	locs := comp.Of(r).Location
	portals := ecs.Get[nav.Portal](r)
	g := space.Of(r)

	all := slices.Clone(store.IDs())
	slices.Sort(all)
	type slotKey struct {
		owner ids.ID
		slot  string
	}
	activeSlots := map[slotKey]ids.ID{}

	for _, id := range all {
		gar, _ := store.Get(id)
		if gar.Slot == "" {
			return fmt.Errorf("prenda %v sin slot", id)
		}
		if !validGarmentState(gar.State) {
			return fmt.Errorf("prenda %v con estado inválido %q", id, gar.State)
		}
		if gar.State == GarmentStateDonning && gar.StateUntil == 0 {
			return fmt.Errorf("prenda %v en donning con state_until == 0", id)
		}
		if gar.State == GarmentStateWorn {
			if !r.Alive(gar.Owner) || !r.Alive(gar.MountPart) || !g.IsSpace(gar.MountPart) {
				return fmt.Errorf("prenda %v en worn con owner o mount_part no vivo", id)
			}
			l, ok := locs.Get(id)
			if !ok || l.Space != gar.MountPart {
				return fmt.Errorf("prenda %v en worn con space %v != mount_part %v", id, l.Space, gar.MountPart)
			}
		}
		if gar.State == GarmentStateWorn || gar.State == GarmentStateDonning {
			k := slotKey{gar.Owner, gar.Slot}
			if other, dup := activeSlots[k]; dup {
				return fmt.Errorf("slot duplicado (%v, %q) en prendas %v y %v", gar.Owner, gar.Slot, other, id)
			}
			activeSlots[k] = id
		}
		if gar.State == GarmentStatePlaced || gar.State == GarmentStateDonning {
			for _, pid := range gar.BodyPortals {
				if r.Alive(pid) && !r.Despawning(pid) {
					if p, ok := portals.Get(pid); ok && !p.Disabled {
						return fmt.Errorf("prenda %v en estado %q tiene portal corporal %v habilitado", id, gar.State, pid)
					}
				}
			}
		}
	}
	return nil
}
