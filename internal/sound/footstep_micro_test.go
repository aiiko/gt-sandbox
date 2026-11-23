package sound

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/view"
)

func TestFootstepMicroOnlyInMicro(t *testing.T) {
	roomID := ids.ID(1)
	deskID := ids.ID(2)
	otherRoomID := ids.ID(3)
	giantessID := ids.ID(4)

	prev := &view.Frame{
		Tick: 1,
		Spaces: []view.Space{
			{ID: roomID, Kind: space.Static},
			{ID: deskID, Kind: space.Elevated, Parent: roomID},
			{ID: otherRoomID, Kind: space.Static},
		},
	}
	cur := &view.Frame{
		Tick:   2,
		Spaces: prev.Spaces,
	}
	ev := []view.Published{{
		Seq:  1,
		Tick: 2,
		Event: crush.Footstep{
			Entity: giantessID,
			Space:  roomID,
			Pos:    units.V2(units.Meter, units.Meter),
		},
	}}
	baseView := func(micro bool, focus ids.ID) View {
		return View{
			Micro:      micro,
			Focus:      focus,
			Locate:     func(units.Vec3) (float64, bool) { return 0, true },
			Visual:     func(ids.ID) string { return "" },
			ShowsEmote: func(ids.ID) bool { return false },
		}
	}

	// Macro: no debe disparar FootStepMicro.
	var scMacro Scene
	scMacro.Sounds(prev, nil, baseView(false, ids.Nil))
	if reqs := scMacro.Sounds(cur, ev, baseView(false, ids.Nil)); len(reqs) != 0 {
		t.Fatalf("en macro no debe sonar FootStepMicro, sonó %+v", reqs)
	}

	// Micro enfocado en el piso de la pisada: volumen completo.
	var scFloor Scene
	scFloor.Sounds(prev, nil, baseView(true, roomID))
	reqs := scFloor.Sounds(cur, ev, baseView(true, roomID))
	if len(reqs) != 1 || reqs[0].ID != FootStepMicro || reqs[0].Volume != 1.0 {
		t.Fatalf("micro en el piso esperaba FootStepMicro vol=1.0, obtuvo %+v", reqs)
	}

	// Micro enfocado en un mueble dentro de la misma habitación: retumbo atenuado.
	var scDesk Scene
	scDesk.Sounds(prev, nil, baseView(true, deskID))
	reqs = scDesk.Sounds(cur, ev, baseView(true, deskID))
	if len(reqs) != 1 || reqs[0].ID != FootStepMicro || reqs[0].Volume != 0.55 {
		t.Fatalf("micro en escritorio de la misma habitación esperaba FootStepMicro vol=0.55, obtuvo %+v", reqs)
	}

	// Micro enfocado en otra habitación: silencio.
	var scOther Scene
	scOther.Sounds(prev, nil, baseView(true, otherRoomID))
	if reqs := scOther.Sounds(cur, ev, baseView(true, otherRoomID)); len(reqs) != 0 {
		t.Fatalf("micro en otra habitación no debía sonar, obtuvo %+v", reqs)
	}
}

func TestSitResonanceInMicro(t *testing.T) {
	roomID := ids.ID(1)
	chairID := ids.ID(2)
	deskID := ids.ID(3)
	giantessID := ids.ID(4)

	prev := &view.Frame{
		Tick: 1,
		Spaces: []view.Space{
			{ID: roomID, Kind: space.Static},
			{ID: chairID, Kind: space.Elevated, Parent: roomID},
			{ID: deskID, Kind: space.Elevated, Parent: roomID},
			{ID: giantessID, Kind: space.Body, Parent: roomID},
		},
	}
	cur := &view.Frame{Tick: 2, Spaces: prev.Spaces}
	ev := []view.Published{{
		Seq:   1,
		Tick:  2,
		Event: attach.PoseChanged{Entity: giantessID, Pose: "sit"},
	}}
	visuals := map[ids.ID]string{chairID: "chair", deskID: "desk", roomID: "bedroom"}
	mkView := func(micro bool, focus ids.ID) View {
		return View{
			Micro:      micro,
			Focus:      focus,
			Locate:     func(units.Vec3) (float64, bool) { return 0, true },
			Visual:     func(id ids.ID) string { return visuals[id] },
			ShowsEmote: func(ids.ID) bool { return false },
		}
	}

	// Micro en la silla: suena chair_creak + sit_thud_micro a volumen completo.
	var scChair Scene
	scChair.Sounds(prev, nil, mkView(true, chairID))
	reqs := scChair.Sounds(cur, ev, mkView(true, chairID))
	hasThud := false
	for _, r := range reqs {
		if r.ID == SitThudMicro && r.Volume == 1.0 {
			hasThud = true
		}
	}
	if !hasThud {
		t.Fatalf("micro en la silla debía incluir SitThudMicro vol=1.0, obtuvo %+v", reqs)
	}

	// Micro en otro mueble de la misma habitación (escritorio): retumbo atenuado.
	var scDesk Scene
	scDesk.Sounds(prev, nil, mkView(true, deskID))
	reqs = scDesk.Sounds(cur, ev, mkView(true, deskID))
	if len(reqs) != 1 || reqs[0].ID != SitThudMicro || reqs[0].Volume != 0.65 {
		t.Fatalf("micro en escritorio debía sonar SitThudMicro vol=0.65, obtuvo %+v", reqs)
	}

	// Macro: no debe sonar SitThudMicro.
	var scMacro Scene
	scMacro.Sounds(prev, nil, mkView(false, ids.Nil))
	for _, r := range scMacro.Sounds(cur, ev, mkView(false, ids.Nil)) {
		if r.ID == SitThudMicro {
			t.Fatalf("en macro no debía sonar SitThudMicro: %+v", r)
		}
	}
}
