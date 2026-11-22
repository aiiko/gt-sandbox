package ai

import (
	"testing"

	"gtsandbox/internal/core/ids"
)

// Un Space visitado vuelve a interesar pasado Revisit, con menos curiosidad;
// sin Revisit, nunca.
func TestCuriosityRecovers(t *testing.T) {
	palm := ids.Make(3, 1)
	a := &Agent{Curiosity: 0.8}
	if got := a.curiosityFor(palm, 10); got != 0.8 {
		t.Fatalf("Space nuevo: %v", got)
	}
	a.visit(palm, 100)
	a.visit(palm, 200) // sigue ahí: se actualiza la última vez
	if len(a.Visited) != 1 || a.Visited[0].At != 200 {
		t.Fatalf("visitas: %+v", a.Visited)
	}
	if got := a.curiosityFor(palm, 100000); got != 0 {
		t.Fatalf("sin Revisit volvió a interesarle: %v", got)
	}
	a.Revisit = 1000
	if got := a.curiosityFor(palm, 1199); got != 0 {
		t.Fatalf("antes de Revisit: %v", got)
	}
	if got := a.curiosityFor(palm, 1200); got != 0.8*RevisitFactor {
		t.Fatalf("pasado Revisit: %v", got)
	}
}
