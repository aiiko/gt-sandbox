package main

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/render"
	"gtsandbox/internal/scenario"
)

func compassOf(a fixedmath.Angle) string {
	q := fixedmath.QuarterTurn
	return [...]string{"east", "north", "west", "south"}[((a.Norm()+q/2)/q)%4]
}

// En cada tick de la mañana, el clip de la giantess, la palma, la taza y los
// tinies sale del estado que publica la simulación: la caminata, la acción en
// uso, el contacto de la palma y la carga.
func TestMorningAnimationFollowsState(t *testing.T) {
	g, err := newGame("../../"+morningScenario, 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.enableArt(""); err != nil {
		t.Fatal(err)
	}
	g.w, g.h = (pixelGame{g}).Layout(960, 540)
	g.fit()
	byVisual := func(visual string) ids.ID {
		for _, id := range slices.Sorted(maps.Keys(g.visuals)) {
			if g.visuals[id] == visual {
				return id
			}
		}
		t.Fatalf("falta el visual %s", visual)
		return ids.Nil
	}
	giantess, palm := byVisual("giantess"), byVisual("palm")
	spec, err := scenario.ReadFile("../../" + morningScenario)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for range spec.Ticks {
		g.run.StepOnce()
		if g.run.Err != nil {
			t.Fatal(g.run.Err)
		}
		f := g.run.Cur
		proj := g.cam.Project(g.anchor(f), g.w, g.h)
		macro := g.art.atlas.Arrange(f, proj, g.visuals, render.Macro, ids.Nil)

		contact := map[ids.ID]bool{}
		for _, c := range f.Contacts {
			contact[c.A], contact[c.B] = true, true
		}
		gs, _ := f.SpaceByID(giantess)
		ps, _ := f.SpaceByID(palm)
		want := "idle_" + compassOf(gs.Pose.Rot)
		switch {
		case gs.Mover != nil && gs.Mover.Status == "moving":
			want = "walk_"
		case gs.Activity == "use:breakfast" && gs.Step == "use":
			want = "eat_" + compassOf(gs.Pose.Rot)
		case gs.Activity == "use:watch" && gs.Step == "use":
			want = "sit_" + compassOf(gs.Pose.Rot)
		case contact[palm]:
			want = "reach_" + compassOf(gs.Pose.Rot)
		}
		got := macro.Clip(giantess)
		if !strings.HasPrefix(got, want) {
			t.Fatalf("tick %d: giantess %q (%s, paso %q, contacto %v), quería %s", f.Tick, got, gs.Activity, gs.Step, contact[palm], want)
		}
		seen[strings.Split(got, "_")[0]]++

		for _, th := range f.Things {
			if g.visuals[th.ID] != "cup" {
				continue
			}
			want := "idle"
			if th.CarriedBy == giantess {
				want = "held"
			}
			if got := macro.Clip(th.ID); got != want {
				t.Fatalf("tick %d: taza %q, quería %q (porteador %s)", f.Tick, got, want, th.CarriedBy)
			}
			seen["cup_"+want]++
		}

		micro := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, ps.ID)
		state := "raised_"
		if contact[palm] {
			state = "rested_"
		}
		if got, want := micro.Clip(palm), state+compassOf(ps.Pose.Rot); got != want {
			t.Fatalf("tick %d: palma %q, quería %q", f.Tick, got, want)
		}
		seen["palm_"+strings.TrimSuffix(state, "_")]++

		desk := g.art.atlas.Arrange(f, proj, g.visuals, render.Micro, g.artDesk)
		carriers := map[ids.ID]bool{}
		for _, th := range f.Things {
			if !th.CarriedBy.IsNil() {
				carriers[th.CarriedBy] = true
			}
		}
		for _, th := range f.Things {
			if th.Space != g.artDesk || !strings.HasPrefix(g.visuals[th.ID], "tiny") || th.Mover == nil || th.Mover.Status != "moving" {
				continue
			}
			prefix := "walk_"
			if carriers[th.ID] {
				prefix = "carry_"
			}
			if got := desk.Clip(th.ID); !strings.HasPrefix(got, prefix) {
				t.Fatalf("tick %d: %s %q, quería %s", f.Tick, th.Name, got, prefix)
			}
			seen["tiny_"+strings.TrimSuffix(prefix, "_")]++
		}
	}
	for _, k := range []string{"walk", "eat", "sit", "idle", "cup_held", "cup_idle", "palm_rested", "palm_raised", "tiny_walk", "tiny_carry"} {
		if seen[k] == 0 {
			t.Errorf("el recorrido nunca mostró %s: %v", k, seen)
		}
	}
}
