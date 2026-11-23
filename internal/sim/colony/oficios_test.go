package colony_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

func newDeskWorld(t *testing.T, seed uint64) (*world.World, ids.ID, ids.ID) {
	t.Helper()
	w := engine.MustNew(world.Config{Seed: seed})
	r := w.ECS()
	root := r.Spawn()
	ecs.Get[space.Space](r).Set(root, space.Space{Kind: space.Static})

	desk := r.Spawn()
	ecs.Get[space.Space](r).Set(desk, space.Space{
		Kind:   space.Elevated,
		Extent: units.Rect{Min: units.V2(0, 0), Max: units.V2(60*units.Centimeter, 40*units.Centimeter)},
	})
	comp.Of(r).Location.Set(desk, comp.Location{Space: root, Pos: units.V3(0, 0, 75*units.Centimeter)})
	g, err := nav.NewGrid(units.V2(0, 0), 2*units.Centimeter, 30, 20)
	if err != nil {
		t.Fatal(err)
	}
	ecs.Get[nav.Grid](r).Set(desk, *g)

	nest := r.Spawn()
	comp.Of(r).Name.Set(nest, "nido")
	comp.Of(r).Location.Set(nest, comp.Location{Space: desk, Pos: units.V3(10*units.Centimeter, 10*units.Centimeter, 0)})
	ecs.Get[colony.Colony](r).Set(nest, colony.Colony{Radius: 4 * units.Centimeter, Stock: 2})
	ecs.Get[ai.Portions](r).Set(nest, ai.Portions{Left: 0})
	return w, desk, nest
}

func spawnTiny(w *world.World, sp, nest ids.ID, x, y units.Length, role string, adv float64, hunger, energy int32) ids.ID {
	r := w.ECS()
	e := r.Spawn()
	comp.Of(r).Location.Set(e, comp.Location{Space: sp, Pos: units.V3(x, y, 0)})
	comp.Of(r).Body.Set(e, comp.Body{Radius: 2 * units.Millimeter, Height: 8 * units.Millimeter})
	ecs.Get[nav.Mover](r).Set(e, nav.Mover{Speed: 4 * units.Centimeter, Scale: "tiny"})
	ecs.Get[ai.Agent](r).Set(e, ai.Agent{Scale: "tiny", Work: 0.2, DistanceCost: 0.1, Sight: 15 * units.Centimeter, Reflex: 2})
	ecs.Get[colony.Member](r).Set(e, colony.Member{Colony: nest, Forage: 0.02, Search: 0.16, Role: role, Adventure: adv})
	ecs.Get[needs.Needs](r).Set(e, needs.Needs{
		{Name: "hunger", Value: hunger, Decay: 0},
		{Name: "energy", Value: energy, Decay: 0},
		{Name: "fun", Value: 500, Decay: 0},
	})
	return e
}

func spawnCrumb(w *world.World, sp ids.ID, x, y units.Length) ids.ID {
	r := w.ECS()
	e := r.Spawn()
	comp.Of(r).Location.Set(e, comp.Location{Space: sp, Pos: units.V3(x, y, 0)})
	ecs.Get[colony.Food](r).Set(e, colony.Food{})
	ecs.Get[ai.Landmark](r).Set(e, ai.Landmark{})
	return e
}

// M50: Con despensa baja y comida al alcance, el recolector encuentra y el
// porteador entrega. Si el porteador falta, otro miembro cubre el acarreo.
func TestRolesAndCoverage(t *testing.T) {
	w, desk, nest := newDeskWorld(t, 1)
	r := w.ECS()

	spot := r.Spawn()
	comp.Of(r).Location.Set(spot, comp.Location{Space: desk, Pos: units.V3(45*units.Centimeter, 30*units.Centimeter, 0)})
	ecs.Get[colony.SearchSpot](r).Set(spot, colony.SearchSpot{LookTicks: 20})
	ecs.Get[ai.Landmark](r).Set(spot, ai.Landmark{Range: units.Meter})

	crumb := spawnCrumb(w, desk, 45*units.Centimeter, 30*units.Centimeter)

	gatherer := spawnTiny(w, desk, nest, 12*units.Centimeter, 10*units.Centimeter, colony.RoleGatherer, 0, 800, 800)
	porter := spawnTiny(w, desk, nest, 14*units.Centimeter, 10*units.Centimeter, colony.RolePorter, 0, 800, 800)
	builder := spawnTiny(w, desk, nest, 16*units.Centimeter, 10*units.Centimeter, colony.RoleBuilder, 0, 800, 800)
	_ = builder

	var searchedBy, hauledBy ids.ID
	stored := false
	for range 1200 {
		w.Step()
		for _, ev := range w.Events() {
			switch x := ev.Event.(type) {
			case colony.Searched:
				if x.Found && searchedBy.IsNil() {
					searchedBy = x.Entity
				}
			case jobs.JobAssigned:
				if hauledBy.IsNil() {
					hauledBy = x.By
				}
			case colony.Stored:
				if x.Item == crumb {
					stored = true
				}
			}
		}
		if stored {
			break
		}
	}
	if searchedBy != gatherer {
		t.Fatalf("se esperaba que buscara el recolector (%v), buscó %v", gatherer, searchedBy)
	}
	if !stored {
		t.Fatal("la migaja no se guardó en el nido")
	}
	if hauledBy != porter {
		t.Fatalf("se esperaba que acarreara el porteador (%v), acarreó %v", porter, hauledBy)
	}

	// Suplencia: sin porteador (despawneado), el constructor cubre un acarreo nuevo.
	r.Despawn(porter)
	w.Step()
	crumb2 := spawnCrumb(w, desk, 20*units.Centimeter, 12*units.Centimeter)
	w.Enqueue(jobs.PostJob{Type: "haul", Target: crumb2, DestSpace: desk, Dest: units.V2(10*units.Centimeter, 10*units.Centimeter)})
	stored2 := false
	for range 800 {
		w.Step()
		for _, ev := range w.Events() {
			if x, ok := ev.Event.(colony.Stored); ok && x.Item == crumb2 {
				stored2 = true
			}
		}
		if stored2 {
			break
		}
	}
	if !stored2 {
		t.Fatal("sin porteador, otro miembro debería haber cubierto el acarreo")
	}
}

// M50: Una colonia solo con generalistas también se abastece.
func TestGeneralistColonyFeedsItself(t *testing.T) {
	w, desk, nest := newDeskWorld(t, 2)
	crumb := spawnCrumb(w, desk, 25*units.Centimeter, 10*units.Centimeter)
	spawnTiny(w, desk, nest, 12*units.Centimeter, 10*units.Centimeter, colony.RoleGeneralist, 0, 800, 800)
	w.Enqueue(jobs.PostJob{Type: "haul", Target: crumb, DestSpace: desk, Dest: units.V2(10*units.Centimeter, 10*units.Centimeter)})

	stored := false
	for range 600 {
		w.Step()
		for _, ev := range w.Events() {
			if x, ok := ev.Event.(colony.Stored); ok && x.Item == crumb {
				stored = true
			}
		}
		if stored {
			break
		}
	}
	if !stored {
		t.Fatal("una colonia de generalistas debe poder abastecerse")
	}
}

// M51: El filtro de escala en nav.Portal impide a la giantess usar el
// pasacables ("tiny") mientras los tinies sí lo usan.
func TestNavPortalScaleFilter(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 3})
	r := w.ECS()
	root := r.Spawn()
	ecs.Get[space.Space](r).Set(root, space.Space{Kind: space.Static})

	roomA := r.Spawn()
	roomB := r.Spawn()
	for _, rm := range []ids.ID{roomA, roomB} {
		ecs.Get[space.Space](r).Set(rm, space.Space{Kind: space.Static, Extent: units.Rect{Max: units.V2(units.Meter, units.Meter)}})
		comp.Of(r).Location.Set(rm, comp.Location{Space: root})
	}
	narrow := r.Spawn()
	ecs.Get[nav.Portal](r).Set(narrow, nav.Portal{
		A:     nav.PortalEnd{Space: roomA, Pos: units.V2(90*units.Centimeter, 50*units.Centimeter)},
		B:     nav.PortalEnd{Space: roomB, Pos: units.V2(10*units.Centimeter, 50*units.Centimeter)},
		Scale: "tiny",
	})

	var router nav.Router
	if _, err := router.RouteFor(w, ids.Nil, "tiny", nil, roomA, units.V2(50*units.Centimeter, 50*units.Centimeter), roomB, units.V2(50*units.Centimeter, 50*units.Centimeter)); err != nil {
		t.Fatalf("tiny debería poder cruzar el portal estrecho: %v", err)
	}
	if _, err := router.RouteFor(w, ids.Nil, "giantess", nil, roomA, units.V2(50*units.Centimeter, 50*units.Centimeter), roomB, units.V2(50*units.Centimeter, 50*units.Centimeter)); err == nil {
		t.Fatal("giantess no debería poder cruzar un portal con scale=tiny")
	}
}

// M51: Un vigía en su puesto detecta una amenaza y emite un aviso una sola vez;
// el puesto queda libre al terminar el turno.
func TestLookoutWarnsAndReleasesPost(t *testing.T) {
	w, desk, nest := newDeskWorld(t, 4)
	r := w.ECS()
	ecs.Get[ai.Portions](r).Set(nest, ai.Portions{Left: 10})

	post := r.Spawn()
	comp.Of(r).Location.Set(post, comp.Location{Space: desk, Pos: units.V3(20*units.Centimeter, 20*units.Centimeter, 0)})
	ecs.Get[colony.LookoutPost](r).Set(post, colony.LookoutPost{
		Colony:      nest,
		WatchPoint:  units.V2(40*units.Centimeter, 20*units.Centimeter),
		WatchRadius: 8 * units.Centimeter,
		AlertRadius: 30 * units.Centimeter,
		ShiftTicks:  120,
		WarnTicks:   60,
	})
	ecs.Get[ai.Landmark](r).Set(post, ai.Landmark{Range: units.Meter})

	// Aplastador suspendido sobre el tramo vigilado (40, 20 cm).
	hand := r.Spawn()
	ecs.Get[space.Space](r).Set(hand, space.Space{
		Kind:   space.Articulated,
		Extent: units.Rect{Min: units.V2(-5*units.Centimeter, -5*units.Centimeter), Max: units.V2(5*units.Centimeter, 5*units.Centimeter)},
	})
	comp.Of(r).Location.Set(hand, comp.Location{Space: desk, Pos: units.V3(40*units.Centimeter, 20*units.Centimeter, 10*units.Centimeter)})
	ecs.Get[crush.Crusher](r).Set(hand, crush.Crusher{Source: "hand", Force: 10, Thickness: 2 * units.Centimeter})

	lookout := spawnTiny(w, desk, nest, 18*units.Centimeter, 20*units.Centimeter, colony.RoleLookout, 0, 900, 900)
	_ = lookout

	warnCount := 0
	watchDone := false
	for range 300 {
		w.Step()
		for _, ev := range w.Events() {
			if _, ok := ev.Event.(colony.LookoutWarned); ok {
				warnCount++
			}
			if x, ok := ev.Event.(ai.ActionDone); ok && x.Action == colony.WatchAction {
				watchDone = true
			}
		}
		if watchDone {
			break
		}
	}
	if !watchDone {
		t.Fatal("el vigía no terminó su turno")
	}
	if warnCount != 1 {
		t.Fatalf("se esperaba 1 aviso por la misma amenaza, hubo %d", warnCount)
	}
	if h := jobs.Holder(w, post); !h.IsNil() {
		t.Fatalf("el puesto debería quedar libre tras el turno, reservado por %v", h)
	}
}

// M52: Expedición solitaria y en compañía (2 a 4), rechazo de 5º miembro,
// conservación de raciones, hallazgos solo publicados al regresar y snapshot.
func TestExpeditionLifecycleAndConservation(t *testing.T) {
	w, desk, nest := newDeskWorld(t, 5)
	r := w.ECS()
	ecs.Get[ai.Portions](r).Set(nest, ai.Portions{Left: 20})

	// Destino de compañía (2 a 4) en el otro extremo del escritorio con una migaja oculta.
	clue := r.Spawn()
	comp.Of(r).Location.Set(clue, comp.Location{Space: desk, Pos: units.V3(15*units.Centimeter, 10*units.Centimeter, 0)})
	ecs.Get[ai.Landmark](r).Set(clue, ai.Landmark{Range: units.Meter})

	target := r.Spawn()
	comp.Of(r).Name.Set(target, "mirador")
	comp.Of(r).Location.Set(target, comp.Location{Space: desk, Pos: units.V3(52*units.Centimeter, 32*units.Centimeter, 0)})
	ecs.Get[colony.ExpeditionTarget](r).Set(target, colony.ExpeditionTarget{
		Colony:      nest,
		Clue:        clue,
		MinMembers:  2,
		MaxMembers:  4,
		SurveyTicks: 30,
		MaxDuration: 2000,
	})
	ecs.Get[ai.Landmark](r).Set(target, ai.Landmark{Range: 10 * units.Centimeter})

	hiddenCrumb := spawnCrumb(w, desk, 52*units.Centimeter, 32*units.Centimeter)

	// 7 tinies: 5 aventureros + 2 caseros (cobertura en el nido).
	var home [2]ids.ID
	for i := range 5 {
		spawnTiny(w, desk, nest, 10*units.Centimeter+units.Length(i)*units.Centimeter, 12*units.Centimeter, colony.RoleGeneralist, 0.9, 900, 900)
	}
	home[0] = spawnTiny(w, desk, nest, 8*units.Centimeter, 8*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)
	home[1] = spawnTiny(w, desk, nest, 9*units.Centimeter, 8*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)

	var departed, surveyed, reported, completed bool
	var departedCount int
	var snap *world.Snapshot
	var snapTick clock.Tick

	for tick := range 2500 {
		w.Step()
		// Tomar un snapshot a mitad de la convocatoria o viaje para probar continuidad.
		if snap == nil {
			if _, exp := colony.ActiveExpedition(w, nest); exp != nil && exp.Phase == colony.ExpGathering {
				var err error
				snap, err = w.Save()
				if err != nil {
					t.Fatal(err)
				}
				snapTick = w.Tick()
			}
		}
		for _, ev := range w.Events() {
			switch x := ev.Event.(type) {
			case colony.ExpeditionDeparted:
				departed = true
				departedCount = len(x.Members)
			case colony.ExpeditionSurveyed:
				surveyed = true
				// Antes de reportar, los caseros NO deben conocer la migaja descubierta.
				for _, h := range home {
					if a, _ := ecs.Get[ai.Agent](r).Get(h); a.Knows(hiddenCrumb) {
						t.Fatalf("tick %d: el casero %v conoció la migaja antes del regreso", tick, h)
					}
				}
			case colony.ExpeditionReported:
				reported = true
			case colony.ExpeditionCompleted:
				completed = true
			}
		}
		if completed {
			break
		}
	}

	if !departed || !surveyed || !reported || !completed {
		t.Fatalf("fases incompletas: departed=%v surveyed=%v reported=%v completed=%v", departed, surveyed, reported, completed)
	}
	if departedCount < 2 || departedCount > 4 {
		t.Fatalf("compañía fuera de rango [2,4]: %d", departedCount)
	}
	// Tras el regreso, los caseros sí conocen la migaja y todas las raciones volvieron (20 porciones).
	for _, h := range home {
		if a, _ := ecs.Get[ai.Agent](r).Get(h); !a.Knows(hiddenCrumb) {
			t.Fatalf("tras el regreso, el casero %v debería conocer la migaja", h)
		}
	}
	dep, _ := ecs.Get[ai.Portions](r).Get(nest)
	if dep.Left != 20 {
		t.Fatalf("conservación de porciones rota: quedaron %d, se esperaban 20", dep.Left)
	}

	// Continuidad desde snapshot a mitad de la convocatoria: mismo hash al llegar al mismo tick.
	if snap == nil {
		t.Fatal("no se capturó snapshot en ExpGathering")
	}
	w2, err := engine.Restore(snap)
	if err != nil {
		t.Fatal(err)
	}
	for w2.Tick() < w.Tick() {
		w2.Step()
	}
	if w2.Hash() != w.Hash() {
		t.Fatalf("divergencia de hash al continuar desde tick %d: %#x != %#x", snapTick, w2.Hash(), w.Hash())
	}
}

// M52: Cierre de portal con grupo dividido -> varado, espera y regresa al reabrirse.
func TestExpeditionStrandedAndRecovery(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 6})
	r := w.ECS()
	root := r.Spawn()
	ecs.Get[space.Space](r).Set(root, space.Space{Kind: space.Static})

	roomA := r.Spawn()
	roomB := r.Spawn()
	for _, rm := range []ids.ID{roomA, roomB} {
		ecs.Get[space.Space](r).Set(rm, space.Space{Kind: space.Static, Extent: units.Rect{Max: units.V2(40*units.Centimeter, 40*units.Centimeter)}})
		comp.Of(r).Location.Set(rm, comp.Location{Space: root})
		g, err := nav.NewGrid(units.V2(0, 0), 2*units.Centimeter, 20, 20)
		if err != nil {
			t.Fatal(err)
		}
		ecs.Get[nav.Grid](r).Set(rm, *g)
	}
	portal := r.Spawn()
	ecs.Get[nav.Portal](r).Set(portal, nav.Portal{
		A: nav.PortalEnd{Space: roomA, Pos: units.V2(35*units.Centimeter, 20*units.Centimeter)},
		B: nav.PortalEnd{Space: roomB, Pos: units.V2(5*units.Centimeter, 20*units.Centimeter)},
	})

	nest := r.Spawn()
	comp.Of(r).Location.Set(nest, comp.Location{Space: roomA, Pos: units.V3(10*units.Centimeter, 20*units.Centimeter, 0)})
	ecs.Get[colony.Colony](r).Set(nest, colony.Colony{Radius: 4 * units.Centimeter, Stock: 2})
	ecs.Get[ai.Portions](r).Set(nest, ai.Portions{Left: 10})

	// Refugio en roomB para que el varado se resguarde.
	refuge := r.Spawn()
	comp.Of(r).Location.Set(refuge, comp.Location{Space: roomB, Pos: units.V3(15*units.Centimeter, 20*units.Centimeter, 0)})
	ecs.Get[crush.Shelter](r).Set(refuge, crush.Shelter{Strength: 500, Radius: 4 * units.Centimeter})

	target := r.Spawn()
	comp.Of(r).Location.Set(target, comp.Location{Space: roomB, Pos: units.V3(25*units.Centimeter, 20*units.Centimeter, 0)})
	ecs.Get[colony.ExpeditionTarget](r).Set(target, colony.ExpeditionTarget{
		Colony:      nest,
		Clue:        portal,
		MinMembers:  1,
		MaxMembers:  1,
		SurveyTicks: 20,
		MaxDuration: 2000,
	})

	scout := spawnTiny(w, roomA, nest, 12*units.Centimeter, 20*units.Centimeter, colony.RoleGeneralist, 0.9, 900, 900)
	spawnTiny(w, roomA, nest, 8*units.Centimeter, 18*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)
	// Darle conocimiento del portal al explorador (sin emitir evento fuera de Step).
	a := ecs.Get[ai.Agent](r).Ptr(scout)
	a.Known = append(a.Known, ai.Known{Entity: portal})

	closed := false
	stranded := false
	reopened := false
	returned := false
	for range 2500 {
		w.Step()
		loc, _ := comp.Of(r).Location.Get(scout)
		if !closed && loc.Space == roomB {
			w.Enqueue(nav.SetPortal{Portal: portal, Enabled: false})
			closed = true
		}
		for _, ev := range w.Events() {
			if _, ok := ev.Event.(colony.ExpeditionStranded); ok {
				stranded = true
			}
			if _, ok := ev.Event.(colony.StrandedReturned); ok {
				returned = true
			}
		}
		if stranded && !reopened {
			w.Enqueue(nav.SetPortal{Portal: portal, Enabled: true})
			reopened = true
		}
		if returned {
			break
		}
	}
	if !closed || !stranded || !returned {
		t.Fatalf("flujo de varado incompleto: closed=%v stranded=%v returned=%v", closed, stranded, returned)
	}
	// La ración volvió al depósito.
	dep, _ := ecs.Get[ai.Portions](r).Get(nest)
	if dep.Left != 10 {
		t.Fatalf("ración perdida tras regreso del varado: quedaron %d, se esperaban 10", dep.Left)
	}
}

// M52: Reemplazo de guía al morir el guía actual.
func TestExpeditionGuideReplacement(t *testing.T) {
	w, desk, nest := newDeskWorld(t, 7)
	r := w.ECS()
	ecs.Get[ai.Portions](r).Set(nest, ai.Portions{Left: 20})

	clue := r.Spawn()
	comp.Of(r).Location.Set(clue, comp.Location{Space: desk, Pos: units.V3(12*units.Centimeter, 10*units.Centimeter, 0)})
	ecs.Get[ai.Landmark](r).Set(clue, ai.Landmark{Range: units.Meter})

	target := r.Spawn()
	comp.Of(r).Location.Set(target, comp.Location{Space: desk, Pos: units.V3(55*units.Centimeter, 35*units.Centimeter, 0)})
	ecs.Get[colony.ExpeditionTarget](r).Set(target, colony.ExpeditionTarget{
		Colony: nest, Clue: clue, MinMembers: 2, MaxMembers: 3, SurveyTicks: 20, MaxDuration: 2000,
	})

	for i := range 3 {
		spawnTiny(w, desk, nest, 10*units.Centimeter+units.Length(i)*units.Centimeter, 12*units.Centimeter, colony.RoleGeneralist, 0.9, 900, 900)
	}
	spawnTiny(w, desk, nest, 8*units.Centimeter, 8*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)
	spawnTiny(w, desk, nest, 9*units.Centimeter, 8*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)

	killed := false
	replaced := false
	completed := false
	for range 2500 {
		w.Step()
		if !killed {
			if _, exp := colony.ActiveExpedition(w, nest); exp != nil && exp.Phase == colony.ExpTraveling && !exp.Guide.IsNil() {
				r.Despawn(exp.Guide)
				killed = true
			}
		}
		for _, ev := range w.Events() {
			if _, ok := ev.Event.(colony.ExpeditionGuideChanged); ok {
				replaced = true
			}
			if _, ok := ev.Event.(colony.ExpeditionCompleted); ok {
				completed = true
			}
		}
		if completed {
			break
		}
	}
	if !killed || !replaced || !completed {
		t.Fatalf("reemplazo de guía incompleto: killed=%v replaced=%v completed=%v", killed, replaced, completed)
	}
	// La ración del guía muerto se perdió con él (19), las de los 2 supervivientes volvieron.
	dep, _ := ecs.Get[ai.Portions](r).Get(nest)
	if dep.Left != 19 {
		t.Fatalf("conservación tras baja: quedaron %d, se esperaban 19", dep.Left)
	}
}

// M52: Falta de provisiones cancela la convocatoria sin perder porciones.
func TestExpeditionCanceledWhenLowProvisions(t *testing.T) {
	w, desk, nest := newDeskWorld(t, 8)
	r := w.ECS()
	// 4 tinies y 5 porciones: alcanza preliminarmente para 1, pero al llegar 2 (min=2) no alcanza (5-2=3 < 4).
	ecs.Get[ai.Portions](r).Set(nest, ai.Portions{Left: 5})

	clue := r.Spawn()
	comp.Of(r).Location.Set(clue, comp.Location{Space: desk, Pos: units.V3(12*units.Centimeter, 10*units.Centimeter, 0)})
	ecs.Get[ai.Landmark](r).Set(clue, ai.Landmark{Range: units.Meter})

	target := r.Spawn()
	comp.Of(r).Location.Set(target, comp.Location{Space: desk, Pos: units.V3(50*units.Centimeter, 30*units.Centimeter, 0)})
	ecs.Get[colony.ExpeditionTarget](r).Set(target, colony.ExpeditionTarget{
		Colony: nest, Clue: clue, MinMembers: 2, MaxMembers: 2,
	})

	// Proponer manualmente en gathering para probar la cancelación al vencer.
	a := spawnTiny(w, desk, nest, 10*units.Centimeter, 10*units.Centimeter, colony.RoleGeneralist, 0.9, 900, 900)
	b := spawnTiny(w, desk, nest, 11*units.Centimeter, 10*units.Centimeter, colony.RoleGeneralist, 0.9, 900, 900)
	spawnTiny(w, desk, nest, 8*units.Centimeter, 8*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)
	spawnTiny(w, desk, nest, 9*units.Centimeter, 8*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)

	expID := r.Spawn()
	ecs.Get[colony.Expedition](r).Set(expID, colony.Expedition{
		Colony:     nest,
		Target:     target,
		Phase:      colony.ExpGathering,
		Guide:      a,
		Members:    []colony.ExpMember{{Entity: a}, {Entity: b}},
		Rendezvous: nest,
		PhaseUntil: 10,
	})

	canceled := false
	for range 30 {
		w.Step()
		for _, ev := range w.Events() {
			if _, ok := ev.Event.(colony.ExpeditionCanceled); ok {
				canceled = true
			}
		}
	}
	if !canceled {
		t.Fatal("debería haberse cancelado por falta de provisiones")
	}
	dep, _ := ecs.Get[ai.Portions](r).Get(nest)
	if dep.Left != 5 {
		t.Fatalf("las porciones no deben cambiar al cancelar: quedaron %d", dep.Left)
	}
}

// M51 (18.8.2): En una misma escena con aviso del vigía:
// - el tiny cercano que se dirige al tramo amenazado espera (WaitUntil > tick);
// - el tiny lejano conserva su movimiento habitual sin esperar;
// - el tiny que ya está bajo la amenaza reacciona con la huida existente.
func TestLookoutNearbyWaitsDistantUnaffectedAndThreatenedFlees(t *testing.T) {
	w, desk, nest := newDeskWorld(t, 11)
	r := w.ECS()
	ecs.Get[ai.Portions](r).Set(nest, ai.Portions{Left: 10})

	post := r.Spawn()
	comp.Of(r).Location.Set(post, comp.Location{Space: desk, Pos: units.V3(20*units.Centimeter, 20*units.Centimeter, 0)})
	ecs.Get[colony.LookoutPost](r).Set(post, colony.LookoutPost{
		Colony:      nest,
		WatchPoint:  units.V2(35*units.Centimeter, 20*units.Centimeter),
		WatchRadius: 8 * units.Centimeter,
		AlertRadius: 15 * units.Centimeter,
		ShiftTicks:  200,
		WarnTicks:   120,
	})
	ecs.Get[ai.Landmark](r).Set(post, ai.Landmark{Range: units.Meter})

	// Mano suspendida sobre (35, 20 cm).
	hand := r.Spawn()
	ecs.Get[space.Space](r).Set(hand, space.Space{
		Kind:   space.Articulated,
		Extent: units.Rect{Min: units.V2(-5*units.Centimeter, -5*units.Centimeter), Max: units.V2(5*units.Centimeter, 5*units.Centimeter)},
	})
	comp.Of(r).Location.Set(hand, comp.Location{Space: desk, Pos: units.V3(35*units.Centimeter, 20*units.Centimeter, 10*units.Centimeter)})
	ecs.Get[crush.Crusher](r).Set(hand, crush.Crusher{Source: "hand", Force: 10, Thickness: 2 * units.Centimeter})

	spawnTiny(w, desk, nest, 20*units.Centimeter, 20*units.Centimeter, colony.RoleLookout, 0, 900, 900)
	nearby := spawnTiny(w, desk, nest, 22*units.Centimeter, 20*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)
	distant := spawnTiny(w, desk, nest, 55*units.Centimeter, 35*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)
	under := spawnTiny(w, desk, nest, 35*units.Centimeter, 20*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)
	// Quitar IA a nearby/distant para controlar sus rutas exactas; under mantiene IA para huir.
	ecs.Get[ai.Agent](r).Remove(nearby)
	ecs.Get[ai.Agent](r).Remove(distant)

	w.Enqueue(nav.MoveTo{Entity: nearby, Space: desk, Goal: units.V2(35*units.Centimeter, 20*units.Centimeter)})
	w.Enqueue(nav.MoveTo{Entity: distant, Space: desk, Goal: units.V2(50*units.Centimeter, 35*units.Centimeter)})

	warned, nearbyWaited, distantWaited, underFled := false, false, false, false
	for range 120 {
		w.Step()
		for _, ev := range w.Events() {
			if _, ok := ev.Event.(colony.LookoutWarned); ok {
				warned = true
			}
			if x, ok := ev.Event.(ai.ActionStarted); ok && x.Entity == under && x.Action == ai.FleeAction {
				underFled = true
			}
		}
		if mv, ok := ecs.Get[nav.Mover](r).Get(nearby); ok && mv.WaitUntil > w.Tick() {
			nearbyWaited = true
		}
		if mv, ok := ecs.Get[nav.Mover](r).Get(distant); ok && mv.WaitUntil > w.Tick() {
			distantWaited = true
		}
	}
	if !warned || !nearbyWaited || distantWaited || !underFled {
		t.Fatalf("reacciones ante vigía incorrectas: warned=%v nearbyWaited=%v distantWaited=%v underFled=%v",
			warned, nearbyWaited, distantWaited, underFled)
	}
}

// M52 (18.8.6): Continuidad determinista desde un snapshot tomado a mitad de
// una retirada (ExpReturning).
func TestExpeditionSnapshotMidReturn(t *testing.T) {
	w, desk, nest := newDeskWorld(t, 12)
	r := w.ECS()
	ecs.Get[ai.Portions](r).Set(nest, ai.Portions{Left: 20})

	clue := r.Spawn()
	comp.Of(r).Location.Set(clue, comp.Location{Space: desk, Pos: units.V3(12*units.Centimeter, 10*units.Centimeter, 0)})
	ecs.Get[ai.Landmark](r).Set(clue, ai.Landmark{Range: units.Meter})

	target := r.Spawn()
	comp.Of(r).Location.Set(target, comp.Location{Space: desk, Pos: units.V3(52*units.Centimeter, 32*units.Centimeter, 0)})
	ecs.Get[colony.ExpeditionTarget](r).Set(target, colony.ExpeditionTarget{
		Colony: nest, Clue: clue, MinMembers: 1, MaxMembers: 1, SurveyTicks: 20, MaxDuration: 2000,
	})

	spawnTiny(w, desk, nest, 10*units.Centimeter, 10*units.Centimeter, colony.RoleGeneralist, 0.9, 900, 900)
	spawnTiny(w, desk, nest, 8*units.Centimeter, 8*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)

	var snap *world.Snapshot
	completed := false
	for range 2500 {
		w.Step()
		if snap == nil {
			if _, exp := colony.ActiveExpedition(w, nest); exp != nil && exp.Phase == colony.ExpReturning {
				var err error
				snap, err = w.Save()
				if err != nil {
					t.Fatal(err)
				}
			}
		}
		for _, ev := range w.Events() {
			if _, ok := ev.Event.(colony.ExpeditionCompleted); ok {
				completed = true
			}
		}
		if completed {
			break
		}
	}
	if snap == nil || !completed {
		t.Fatalf("no se completó o no se capturó snapshot en ExpReturning: snap=%v completed=%v", snap != nil, completed)
	}
	w2, err := engine.Restore(snap)
	if err != nil {
		t.Fatal(err)
	}
	for w2.Tick() < w.Tick() {
		w2.Step()
	}
	if w2.Hash() != w.Hash() {
		t.Fatalf("divergencia al continuar desde retirada: %#x != %#x", w2.Hash(), w.Hash())
	}
}

// M53 (18.8.4): Expedición al bolsillo del sofá: completa el reconocimiento
// incluso con el lugar vacío, y si hay una migaja presente la reporta y genera
// acarreo.
func TestSofaPocketExpeditionEmptyAndWithCrumb(t *testing.T) {
	for _, withCrumb := range []bool{false, true} {
		w := engine.MustNew(world.Config{Seed: 15})
		r := w.ECS()
		root := r.Spawn()
		ecs.Get[space.Space](r).Set(root, space.Space{Kind: space.Static})

		sofa := r.Spawn()
		ecs.Get[space.Space](r).Set(sofa, space.Space{Kind: space.Elevated, Extent: units.Rect{Max: units.V2(40*units.Centimeter, 40*units.Centimeter)}})
		comp.Of(r).Location.Set(sofa, comp.Location{Space: root, Pos: units.V3(0, 0, 45*units.Centimeter)})
		g1, _ := nav.NewGrid(units.V2(0, 0), 2*units.Centimeter, 20, 20)
		ecs.Get[nav.Grid](r).Set(sofa, *g1)

		pocket := r.Spawn()
		ecs.Get[space.Space](r).Set(pocket, space.Space{Kind: space.Elevated, Extent: units.Rect{Max: units.V2(12*units.Centimeter, 8*units.Centimeter)}})
		comp.Of(r).Location.Set(pocket, comp.Location{Space: sofa})
		g2, _ := nav.NewGrid(units.V2(0, 0), 2*units.Centimeter, 6, 4)
		ecs.Get[nav.Grid](r).Set(pocket, *g2)

		entry := r.Spawn()
		ecs.Get[nav.Portal](r).Set(entry, nav.Portal{
			A:     nav.PortalEnd{Space: sofa, Pos: units.V2(20*units.Centimeter, 20*units.Centimeter)},
			B:     nav.PortalEnd{Space: pocket, Pos: units.V2(6*units.Centimeter, 6*units.Centimeter)},
			Scale: "tiny",
		})

		nest := r.Spawn()
		comp.Of(r).Location.Set(nest, comp.Location{Space: sofa, Pos: units.V3(6*units.Centimeter, 6*units.Centimeter, 0)})
		ecs.Get[colony.Colony](r).Set(nest, colony.Colony{Radius: 4 * units.Centimeter, Stock: 2})
		ecs.Get[ai.Portions](r).Set(nest, ai.Portions{Left: 20})

		target := r.Spawn()
		comp.Of(r).Location.Set(target, comp.Location{Space: pocket, Pos: units.V3(6*units.Centimeter, 3*units.Centimeter, 0)})
		ecs.Get[colony.ExpeditionTarget](r).Set(target, colony.ExpeditionTarget{
			Colony: nest, Clue: entry, MinMembers: 2, MaxMembers: 2, SurveyTicks: 20, MaxDuration: 2000,
		})

		var crumb ids.ID
		if withCrumb {
			crumb = spawnCrumb(w, pocket, 6*units.Centimeter, 3*units.Centimeter)
		}

		for i := range 2 {
			e := spawnTiny(w, sofa, nest, 8*units.Centimeter+units.Length(i)*units.Centimeter, 8*units.Centimeter, colony.RoleGeneralist, 0.9, 900, 900)
			a := ecs.Get[ai.Agent](r).Ptr(e)
			a.Known = append(a.Known, ai.Known{Entity: entry})
		}
		spawnTiny(w, sofa, nest, 6*units.Centimeter, 8*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)
		spawnTiny(w, sofa, nest, 6*units.Centimeter, 10*units.Centimeter, colony.RoleGeneralist, 0, 900, 900)

		completed, haulPosted := false, false
		for range 2500 {
			w.Step()
			for _, ev := range w.Events() {
				if _, ok := ev.Event.(colony.ExpeditionCompleted); ok {
					completed = true
				}
				if x, ok := ev.Event.(jobs.JobPosted); ok && x.Type == "haul" && x.Target == crumb {
					haulPosted = true
				}
			}
			if completed {
				break
			}
		}
		if !completed {
			t.Fatalf("withCrumb=%v: la expedición al bolsillo del sofá no terminó", withCrumb)
		}
		if withCrumb && !haulPosted {
			t.Fatal("con migaja en el bolsillo del sofá, al reportar debería publicarse un acarreo")
		}
		if tgt, _ := ecs.Get[colony.ExpeditionTarget](r).Get(target); !tgt.Visited {
			t.Fatalf("withCrumb=%v: el destino debería quedar marcado como visitado", withCrumb)
		}
	}
}
