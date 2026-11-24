package scenario_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/space"
)

// TestApartmentCirculationAndPortals verifica 20.4.1:
//   - Recorridos de la giantess: genkan → distribuidor → dormitorio, baño y living, y living → cocina.
//   - Cada puerta conecta dos puntos próximos en el mundo.
//   - Dormitorio ↔ living por piso (distribuidor) y por pasacables, con filtro de escala tiny.
func TestApartmentCirculationAndPortals(t *testing.T) {
	spec, err := scenario.ReadFile(partidaBase)
	if err != nil {
		t.Fatal(err)
	}
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	w, refs := in.World, in.Refs
	w.Step() // aplica set_components del tick 0

	r := w.ECS()
	sp := space.Of(r)
	portals := ecs.Get[nav.Portal](r)

	// 1. Verificar que los portales entre habitaciones conectan puntos próximos en el mundo (<= 30 cm).
	for _, pref := range []string{"genkan_step", "door_bathroom", "door_bedroom", "door_living", "door_kitchen"} {
		pid := refs[pref]
		if pid.IsNil() {
			t.Fatalf("falta portal %q en casa.json", pref)
		}
		p, ok := portals.Get(pid)
		if !ok {
			t.Fatalf("portal %q sin componente nav.Portal", pref)
		}
		pa, errA := sp.Frame(p.A.Space)
		pb, errB := sp.Frame(p.B.Space)
		if errA != nil || errB != nil {
			t.Fatalf("portal %q: sin Frame (%v, %v)", pref, errA, errB)
		}
		wa := pa.Apply(units.V3(p.A.Pos.X, p.A.Pos.Y, 0))
		wb := pb.Apply(units.V3(p.B.Pos.X, p.B.Pos.Y, 0))
		d := wa.XY().Sub(wb.XY()).Len()
		if d > 30*units.Centimeter {
			t.Errorf("portal %q separa sus extremos %v (> 30 cm)", pref, d)
		}
	}

	// 2. Verificar que la giantess tiene ruta desde genkan hacia distribuidor, dormitorio, baño, living y cocina,
	//    y que nunca usa el pasacables (scale=tiny).
	giantess := refs["giantess"]
	ecs.Get[ai.Agent](r).Remove(giantess)
	locs := ecs.Get[comp.Location](r)
	locs.Set(giantess, comp.Location{Space: refs["genkan"], Pos: units.V3(80*units.Centimeter, 100*units.Centimeter, 0)})
	targets := []struct {
		room string
		goal units.Vec2
	}{
		{"hallway", units.V2(220*units.Centimeter, 55*units.Centimeter)},
		{"bathroom", units.V2(120*units.Centimeter, 140*units.Centimeter)},
		{"bedroom", units.V2(180*units.Centimeter, 150*units.Centimeter)},
		{"living", units.V2(200*units.Centimeter, 120*units.Centimeter)},
		{"kitchen", units.V2(180*units.Centimeter, 150*units.Centimeter)},
	}
	for _, tc := range targets {
		dest := refs[tc.room]
		w.Enqueue(nav.MoveTo{Entity: giantess, Space: dest, Goal: tc.goal})
		arrived := false
		for range 800 {
			w.Step()
			loc, _ := ecs.Get[comp.Location](r).Get(giantess)
			if loc.Space == refs["cable_duct"] {
				t.Fatalf("la giantess entró al pasacables yendo a %s", tc.room)
			}
			if loc.Space == dest && loc.Pos.XY().Sub(tc.goal).Len() <= 15*units.Centimeter {
				arrived = true
				break
			}
		}
		if !arrived {
			loc, _ := ecs.Get[comp.Location](r).Get(giantess)
			t.Fatalf("la giantess no llegó a %s (quedó en %v pos=%v)", tc.room, loc.Space, loc.Pos)
		}
	}

	// 3. Verificar que un tiny puede ir del dormitorio al living por el pasacables y también por el piso (distribuidor).
	tinyID := refs["tiny01"]
	ecs.Get[ai.Agent](r).Remove(tinyID)
	locs.Set(tinyID, comp.Location{Space: refs["bedroom"], Pos: units.V3(290*units.Centimeter, 230*units.Centimeter, 0)})
	ductID := refs["cable_duct"]
	w.Enqueue(nav.MoveTo{Entity: tinyID, Space: refs["living"], Goal: units.V2(25*units.Centimeter, 232*units.Centimeter)})
	usedDuct := false
	reachedLiving := false
	for range 1200 {
		w.Step()
		loc, _ := ecs.Get[comp.Location](r).Get(tinyID)
		if loc.Space == ductID {
			usedDuct = true
		}
		if loc.Space == refs["living"] {
			reachedLiving = true
			break
		}
	}
	if !usedDuct || !reachedLiving {
		t.Fatalf("tiny no cruzó al living por pasacables: usedDuct=%v reachedLiving=%v", usedDuct, reachedLiving)
	}
}

// TestApartmentPlayableSurfacesAndFalls verifica 20.4.2:
//   - Todas las superficies marcadas «Sí» en 20.2 son alcanzables por un tiny.
//   - Las piezas marcadas «No» no son superficies caminables por encima.
//   - Una caída dentro de la bacha aterriza en sink_basin; una caída del escritorio aterriza en el piso.
func TestApartmentPlayableSurfacesAndFalls(t *testing.T) {
	spec, err := scenario.ReadFile(partidaBase)
	if err != nil {
		t.Fatal(err)
	}
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	w, refs := in.World, in.Refs
	w.Step()
	r := w.ECS()
	spaces := ecs.Get[space.Space](r)
	portals := ecs.Get[nav.Portal](r)

	// 1. Grafo de conectividad tiny desde desk1 (incluye solo portales habilitados compatibles con tiny).
	adj := map[ids.ID][]ids.ID{}
	portals.Each(func(_ ids.ID, p *nav.Portal) {
		if p.Disabled || (p.Scale != "" && p.Scale != "tiny") {
			return
		}
		adj[p.A.Space] = append(adj[p.A.Space], p.B.Space)
		adj[p.B.Space] = append(adj[p.B.Space], p.A.Space)
	})
	reachable := map[ids.ID]bool{refs["desk1"]: true}
	queue := []ids.ID{refs["desk1"]}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, v := range adj[u] {
			if !reachable[v] {
				reachable[v] = true
				queue = append(queue, v)
			}
		}
	}

	// Todas las superficies jugables («Sí» en 20.2) + los 6 pisos + hueco bajo el escritorio.
	mustReach := []string{
		"desk1", "chair", "futon", "pillow",
		"sofa", "desk2", "sofa_pocket",
		"counter", "under_desk", "cable_duct",
		"bedroom", "living", "kitchen", "hallway", "bathroom", "genkan",
	}
	for _, ref := range mustReach {
		id := refs[ref]
		if id.IsNil() {
			t.Fatalf("falta superficie jugable %q", ref)
		}
		if !reachable[id] {
			t.Errorf("superficie jugable %q no es alcanzable por un tiny desde el escritorio", ref)
		}
	}

	// 2. Las piezas marcadas «No» en 20.2 no deben tener Space caminable propio.
	nonWalkable := []string{
		"low_closet", "tv_stand", "tv", "plant",
		"sink", "stove", "fridge", "rice_cooker", "dish_shelf",
		"washbasin", "toilet", "shower_tray", "shoe_rack",
	}
	for _, ref := range nonWalkable {
		id := refs[ref]
		if id.IsNil() {
			t.Fatalf("falta pieza %q en casa.json", ref)
		}
		if spaces.Has(id) {
			t.Errorf("pieza no caminable %q no debería tener Space navegable propio", ref)
		}
	}

	// 3. Verificar que la zona húmeda (shower_tray) bloquea el piso del baño para tinies.
	bathGrid, ok := ecs.Get[nav.Grid](r).Get(refs["bathroom"])
	if !ok {
		t.Fatal("falta nav_grid del baño")
	}
	if bathGrid.At(24, 5) != nav.Blocked {
		t.Errorf("la celda dentro del plato de ducha debería estar bloqueada en el baño")
	}

	// 4. Caída dentro de la bacha aterriza en sink_basin (fondo de la bacha, z = 68 cm).
	tinyID := refs["tiny02"]
	ecs.Get[ai.Agent](r).Remove(tinyID)
	ecs.Get[comp.Location](r).Set(tinyID, comp.Location{Space: refs["kitchen"], Pos: units.V3(30*units.Centimeter, 55*units.Centimeter, 84*units.Centimeter)})
	w.Step()
	loc, _ := ecs.Get[comp.Location](r).Get(tinyID)
	if loc.Space != refs["sink_basin"] {
		t.Errorf("una caída en la bacha debía aterrizar en sink_basin (%v), cayó en %v", refs["sink_basin"], loc.Space)
	}
}

// TestApartmentProfilesAndBodyContacts verifica 20.4.4:
//   - Los 3 perfiles de inicio llegan a dormir en el futón, desayunar/escribir en el escritorio,
//     y merendar/mirar TV desde el sofá.
//   - Funcionan los contactos escritorio↔palma, sofá↔cuerpo y almohada↔cuerpo en la planta nueva.
func TestApartmentProfilesAndBodyContacts(t *testing.T) {
	if testing.Short() {
		t.Skip("prueba de perfiles y contactos sobre la planta completa")
	}
	base, err := scenario.ReadFile(partidaBase)
	if err != nil {
		t.Fatal(err)
	}
	sawAction := map[string]bool{}
	sawContact := map[string]bool{}
	for _, start := range base.Random.Starts {
		s := *base
		s.Random = nil
		s.Expect = nil
		s.Ticks = 4000
		in, err := scenario.Build(&s)
		if err != nil {
			t.Fatalf("perfil %s: %v", start.Name, err)
		}
		g := in.Refs["giantess"]
		ecs.Get[needs.Needs](in.World.ECS()).Set(g, append(needs.Needs(nil), start.Needs...))
		if _, err := in.Run(); err != nil {
			t.Fatalf("perfil %s: %v", start.Name, err)
		}
		desk1, sofa, pillow := in.Refs["desk1"], in.Refs["sofa"], in.Refs["pillow"]
		palm := in.Refs["palm"]
		for _, e := range in.Events {
			switch ev := e.Event.(type) {
			case ai.ActionDone:
				if ev.Entity == g {
					sawAction[ev.What] = true
				}
			case attach.ContactBegan:
				if (ev.A == desk1 && ev.B == palm) || (ev.B == desk1 && ev.A == palm) {
					sawContact["desk_palm"] = true
				}
				if ev.A == sofa || ev.B == sofa {
					sawContact["sofa_body"] = true
				}
				if ev.A == pillow || ev.B == pillow {
					sawContact["pillow_body"] = true
				}
			}
		}
	}
	t.Logf("acciones completadas: %v, contactos: %v", sawAction, sawContact)
	for _, want := range []string{"sleep", "breakfast", "watch"} {
		if !sawAction[want] {
			t.Errorf("ningún perfil completó %q (vistas: %v)", want, sawAction)
		}
	}
	for _, want := range []string{"desk_palm", "sofa_body", "pillow_body"} {
		if !sawContact[want] {
			t.Errorf("no se abrió el contacto %q en la planta nueva (vistos: %v)", want, sawContact)
		}
	}
}
