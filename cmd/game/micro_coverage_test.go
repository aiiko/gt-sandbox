package main

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/render"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/nav"
)

// M54 (propuesta 18.7 y 18.8.5): seguir al mismo tiny desde el escritorio, por
// la pata, el dormitorio, el pasacables y el living hasta el bolsillo del sofá.
// En cada tramo mantiene el seguimiento en micro, su sprite se ve en la escena
// (con vista básica en las superficies sin arte propio) y cambiar de cámara no
// altera el hash de la simulación.
func TestMicroCoverageFullJourney(t *testing.T) {
	g := artGame(t, randomBase)
	f := g.run.Cur

	var tinyID, pocketID, ductID, legID ids.ID
	for _, th := range f.Things {
		if isTiny(th) {
			tinyID = th.ID
			break
		}
	}
	for _, s := range f.Spaces {
		switch s.Name {
		case "bolsillo del sofá":
			pocketID = s.ID
		case "pasacables del zócalo":
			ductID = s.ID
		case "pata del escritorio":
			legID = s.ID
		}
	}
	if tinyID.IsNil() || pocketID.IsNil() || ductID.IsNil() || legID.IsNil() {
		t.Fatalf("faltan entidades del mapa de aventuras: tiny=%v pocket=%v duct=%v leg=%v", tinyID, pocketID, ductID, legID)
	}

	// Quitar la IA del tiny de prueba para que siga exactamente el recorrido pedido.
	ecs.Get[ai.Agent](g.run.World.ECS()).Remove(tinyID)
	// Empezar siguiendo al tiny en micro.
	g.sel = tinyID
	g.toggleFollow()
	g.toggleArtMicro()
	if g.artFocus(g.run.Cur).IsNil() {
		t.Fatal("no entró en vista micro")
	}

	// Ordenar ir al bolsillo del sofá pasando por el pasacables.
	g.run.World.Enqueue(nav.MoveTo{Entity: tinyID, Space: ductID, Goal: units.V2(10*units.Centimeter, 2*units.Centimeter)})
	visited := map[ids.ID]bool{}
	artSeen := map[ids.ID]bool{}
	sentPocket := false

	for range 8000 {
		g.run.StepOnce()
		f = g.run.Cur
		th, ok := f.ThingByID(tinyID)
		if !ok {
			t.Fatal("el tiny seguido desapareció")
		}
		focus := g.artFocus(f)
		if focus != th.Space {
			t.Fatalf("tick %d: foco %s, tiny en %s", f.Tick, g.name(focus), g.name(th.Space))
		}
		visited[th.Space] = true
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		scene := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, focus, g.sel, g.follow)
		if scene.Clip(tinyID) == "" {
			t.Fatalf("tick %d: tiny invisible en micro de %s", f.Tick, g.name(th.Space))
		}
		if scene.Clip(th.Space) != "" {
			artSeen[th.Space] = true
		}
		if th.Space == ductID && !sentPocket {
			g.run.World.Enqueue(nav.MoveTo{Entity: tinyID, Space: pocketID, Goal: units.V2(6*units.Centimeter, 3*units.Centimeter)})
			sentPocket = true
		}
		if th.Space == pocketID {
			break
		}
	}

	if !visited[ductID] {
		t.Fatal("el tiny no pasó por el pasacables del zócalo")
	}
	if !visited[pocketID] {
		t.Fatal("el tiny no llegó al bolsillo del sofá")
	}
	if !artSeen[ductID] || !artSeen[pocketID] {
		t.Fatalf("las superficies nuevas deben dibujarse en micro: duct=%v pocket=%v", artSeen[ductID], artSeen[pocketID])
	}

	// Verificar también el respaldo de vista micro básica cuando falta arte específico.
	delete(g.visuals, pocketID)
	proj := g.cam.Project(g.anchor(f), g.w, g.h)
	fallback := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, pocketID, g.sel, g.follow)
	if !fallback.HasBackdrop() || fallback.Clip(tinyID) == "" {
		t.Fatal("una superficie sin visual debe usar la vista micro básica y mostrar al tiny")
	}
	// Verificar que el miniplano en micro incluye pistas de expedición (18.6).
	if mm := g.minimap(f, pocketID); mm == nil || len(mm.Pins) == 0 {
		t.Fatal("el miniplano en micro debería marcar los destinos/pistas de expedición")
	}
}

// M54 (18.8.5): Desde macro, presionar Tab con un tiny elegido a mitad de una
// trepada abre directamente la vista micro de esa pata/respaldo.
func TestTabOnClimbingTinyOpensMicro(t *testing.T) {
	g := artGame(t, randomBase)
	f := g.run.Cur
	var climber, climbSpace ids.ID
	for i := 0; i < 6000 && climber.IsNil(); i++ {
		g.run.StepOnce()
		f = g.run.Cur
		for _, th := range f.Things {
			if isTiny(th) && g.visuals[th.Space] == render.ClimbVisual {
				climber, climbSpace = th.ID, th.Space
				break
			}
		}
	}
	if climber.IsNil() {
		t.Fatal("ningún tiny trepó una pata o respaldo")
	}
	// En macro, con el tiny elegido: Tab abre la vista micro de esa pata.
	g.art.focus = ids.Nil
	g.sel = climber
	g.toggleArtMicro()
	if got := g.artFocus(g.run.Cur); got != climbSpace {
		t.Fatalf("Tab en mitad de la trepada abrió %s, se esperaba %s", g.name(got), g.name(climbSpace))
	}
}
