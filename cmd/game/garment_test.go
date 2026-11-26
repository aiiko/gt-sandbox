package main

import (
	"slices"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/render"
)

func TestGarmentLocationLabelsAndCasaArt(t *testing.T) {
	g := artGame(t, randomBase)
	f := g.run.Cur
	g.w, g.h = 480, 270

	var bedroom, giantess, footL, slipperL, cardigan ids.ID
	for _, s := range f.Spaces {
		switch s.Name {
		case "dormitorio":
			bedroom = s.ID
		case "Giantess", "giantess":
			giantess = s.ID
		case "pie izquierdo":
			footL = s.ID
		case "pantufla izquierda":
			slipperL = s.ID
		case "bolsillo del cárdigan":
			cardigan = s.ID
		}
	}
	if bedroom.IsNil() || giantess.IsNil() || footL.IsNil() || slipperL.IsNil() || cardigan.IsNil() {
		t.Fatalf("faltan entidades en casa.json: bed=%v g=%v foot=%v slip=%v card=%v", bedroom, giantess, footL, slipperL, cardigan)
	}

	if got := g.location(f, slipperL); !slices.Equal(got, []string{"Dormitorio", "Pantufla izquierda"}) {
		t.Fatalf("rótulo colocada: %v", got)
	}

	// Verificar arte micro propio para ambas prendas (sin basicSurface).
	for _, spID := range []ids.ID{slipperL, cardigan} {
		g.art.focus = spID
		g.fitArt()
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		scene := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, spID)
		if scene.HasBackdrop() {
			t.Errorf("prenda %s cayó a vista micro básica", g.name(spID))
		}
		if scene.Clip(spID) == "" {
			t.Errorf("prenda %s no dibujó clip en micro", g.name(spID))
		}
	}

	// Simular prenda puesta sobre footL en el frame y verificar rótulo jerárquico.
	fCopy := *f
	fCopy.Spaces = slices.Clone(f.Spaces)
	for i := range fCopy.Spaces {
		if fCopy.Spaces[i].ID == slipperL {
			fCopy.Spaces[i].Parent = footL
			fCopy.Spaces[i].GarmentState = "worn"
		}
	}
	if got := g.location(&fCopy, slipperL); !slices.Equal(got, []string{"Giantess", "Pie izquierdo", "Pantufla izquierda"}) {
		t.Fatalf("rótulo puesta: %v", got)
	}
}
