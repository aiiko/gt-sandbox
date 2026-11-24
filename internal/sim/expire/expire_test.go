package expire_test

import (
	"bytes"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/expire"
	"gtsandbox/internal/sim/world"
)

func TestExpireOrderAndEvents(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 1})
	r := w.ECS()
	exp := ecs.Get[expire.Expires](r)

	a := r.Spawn()
	b := r.Spawn()
	c := r.Spawn()
	exp.Set(a, expire.Expires{At: 5, Kind: "water_drop"})
	exp.Set(b, expire.Expires{At: 5, Kind: "waymark"})
	exp.Set(c, expire.Expires{At: 10, Kind: "water_drop"})

	w.StepN(5) // ticks 0..4
	if !r.Alive(a) || !r.Alive(b) || !r.Alive(c) {
		t.Fatalf("vencieron antes de tiempo en tick %v", w.Tick())
	}

	w.Step() // tick 5: vencen a y b en orden de ID
	if r.Alive(a) || r.Alive(b) {
		t.Fatalf("a y b debían desaparecer en tick 5")
	}
	if !r.Alive(c) {
		t.Fatalf("c no debía desaparecer todavía")
	}
	var got []expire.Expired
	for _, ev := range w.Events() {
		if ex, ok := ev.Event.(expire.Expired); ok {
			got = append(got, ex)
		}
	}
	if len(got) != 2 || got[0].Entity != a || got[0].ItemKind != "water_drop" || got[1].Entity != b || got[1].ItemKind != "waymark" {
		t.Fatalf("eventos inesperados: %+v", got)
	}
	if err := w.CheckInvariants(); err != nil {
		t.Fatalf("invariantes: %v", err)
	}
}

func TestExpireTTLAndSnapshot(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 7})
	w.StepN(10)

	r := w.ECS()
	e := r.Spawn()
	comp.Of(r).Name.Set(e, "gota")
	ecs.Get[expire.Expires](r).Set(e, expire.Expires{TTL: 5})

	w.Step() // tick 10: fija At = 15
	got, ok := ecs.Get[expire.Expires](r).Get(e)
	if !ok || got.At != 15 || got.TTL != 0 {
		t.Fatalf("TTL no resuelto: %+v", got)
	}

	snap, err := w.Save()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := world.WriteSnapshot(&buf, snap); err != nil {
		t.Fatal(err)
	}
	w2, err := engine.Load(&buf)
	if err != nil {
		t.Fatal(err)
	}

	w.StepN(5)  // ticks 11..15
	w2.StepN(5) // ticks 11..15

	if w.Hash() != w2.Hash() {
		t.Fatalf("hash distinto tras snapshot: %x vs %x", w.Hash(), w2.Hash())
	}
	if w.ECS().Alive(e) {
		t.Fatalf("la entidad %v debía vencer en tick 15", e)
	}
	var seen []ids.ID
	for _, ev := range w.Events() {
		if ex, ok := ev.Event.(expire.Expired); ok && ex.ItemKind == "gota" {
			seen = append(seen, ex.Entity)
		}
	}
	if len(seen) != 1 || seen[0] != e {
		t.Fatalf("esperaba Expired con kind fallback 'gota', obtuve %+v", seen)
	}
}

func TestExpireImmediateAtZero(t *testing.T) {
	w := engine.MustNew(world.Config{Seed: 2})
	e := w.ECS().Spawn()
	ecs.Get[expire.Expires](w.ECS()).Set(e, expire.Expires{At: clock.Tick(0), Kind: "mark"})
	w.Step()
	if w.ECS().Alive(e) {
		t.Fatalf("At=0 debía vencer inmediatamente")
	}
}
