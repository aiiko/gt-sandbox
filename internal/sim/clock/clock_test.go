package clock

import (
	"testing"
	"time"
)

func TestTickDuration(t *testing.T) {
	if TickDuration != 50*time.Millisecond {
		t.Fatalf("TickDuration = %v", TickDuration)
	}
	if got := TicksFor(time.Second); got != 20 {
		t.Fatalf("TicksFor(1s) = %d", got)
	}
	if got := TicksFor(time.Millisecond); got != 1 {
		t.Fatalf("una duración positiva nunca es 0 ticks, vino %d", got)
	}
	if got := Tick(40).Duration(); got != 2*time.Second {
		t.Fatalf("Duration = %v", got)
	}
}

func TestPhasesNamed(t *testing.T) {
	for p := Phase(0); p.Valid(); p++ {
		if p.String() == "" {
			t.Fatalf("fase %d sin nombre", p)
		}
	}
	if len(phaseNames) != PhaseCount {
		t.Fatal("phaseNames desincronizado con las fases")
	}
}

func TestStaggeredSpreadsLoad(t *testing.T) {
	const period = 10
	const entities = 1000
	for tick := Tick(0); tick < period*3; tick++ {
		n := 0
		for e := uint64(0); e < entities; e++ {
			if Staggered(tick, period, e) {
				n++
			}
		}
		if n != entities/period {
			t.Fatalf("tick %d: %d entidades, esperado %d", tick, n, entities/period)
		}
	}
	// Cada entidad corre exactamente una vez por período.
	for e := uint64(0); e < 50; e++ {
		runs := 0
		for tick := Tick(0); tick < period; tick++ {
			if Staggered(tick, period, e) {
				runs++
			}
		}
		if runs != 1 {
			t.Fatalf("entidad %d corrió %d veces en un período", e, runs)
		}
	}
}

func TestDueEveryTick(t *testing.T) {
	for tick := Tick(0); tick < 5; tick++ {
		if !Due(tick, 0, 0) || !Due(tick, 1, 3) {
			t.Fatal("period 0/1 debe correr siempre")
		}
	}
}
