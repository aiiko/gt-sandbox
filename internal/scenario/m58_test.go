package scenario_test

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/space"
)

// TestBodyMapCoverageAndReachableExits valida el contenido corporal en
// content/partidas/casa.json (propuesta 19.5):
//   - Toda zona corporal tiene extensión no vacía (vista micro disponible).
//   - Toda zona corporal tiene una salida alcanzable hacia la casa (vía grafo
//     interno + alguna zona con contact_surface externa) en al menos una pose.
func TestBodyMapCoverageAndReachableExits(t *testing.T) {
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
	spaces := ecs.Get[space.Space](r)
	zones := ecs.Get[attach.BodyZone](r)
	surfaces := ecs.Get[attach.Surface](r)
	portals := ecs.Get[nav.Portal](r)

	giantess := refs["giantess"]
	if giantess.IsNil() {
		t.Fatal("falta giantess en casa.json")
	}

	// Recolectar todas las zonas corporales (BodyZone + palm).
	bodyZones := map[ids.ID]bool{}
	if palm := refs["palm"]; !palm.IsNil() {
		bodyZones[palm] = true
	}
	zones.Each(func(id ids.ID, z *attach.BodyZone) {
		if z.Body == giantess {
			bodyZones[id] = true
		}
	})
	if len(bodyZones) < 15 {
		t.Fatalf("se esperaba mapa corporal completo (>= 15 zonas), hay %d", len(bodyZones))
	}

	// 1. Toda zona corporal debe tener extensión no vacía (para vista micro).
	for id := range bodyZones {
		sp, ok := spaces.Get(id)
		if !ok || sp.Extent.Empty() {
			t.Errorf("zona corporal %v sin extensión no vacía para vista micro", id)
		}
	}

	// 2. Grafo interno de portales (en pose idle): toda zona debe alcanzar alguna
	//    zona corporal con contact_surface (salida a la casa en alguna pose).
	adj := map[ids.ID][]ids.ID{}
	portals.Each(func(_ ids.ID, p *nav.Portal) {
		if p.Disabled {
			return
		}
		adj[p.A.Space] = append(adj[p.A.Space], p.B.Space)
		adj[p.B.Space] = append(adj[p.B.Space], p.A.Space)
	})

	visited := map[ids.ID]bool{}
	var queue []ids.ID
	for id := range bodyZones {
		if surfaces.Has(id) {
			visited[id] = true
			queue = append(queue, id)
		}
	}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, v := range adj[u] {
			if bodyZones[v] && !visited[v] {
				visited[v] = true
				queue = append(queue, v)
			}
		}
	}
	for id := range bodyZones {
		if !visited[id] {
			t.Errorf("zona corporal %v no tiene salida alcanzable a ninguna superficie de contacto", id)
		}
	}
}
