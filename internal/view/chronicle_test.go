package view

import (
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/sim/clock"
)

func emptyChronicle() *Chronicle {
	c := &Chronicle{seeded: true}
	c.diaries, c.firsts, c.routine, c.empty = map[ids.ID][]StoryEntry{}, map[ids.ID]bool{}, map[ids.ID]int{}, map[ids.ID]bool{}
	return c
}

// Los detalles van a los diarios, no a la crónica; los de un mismo tipo y
// lugar que llegan juntos se agrupan en el diario del lugar, y cada actor
// guarda el suyo.
func TestChronicleDetailsGroupByPlace(t *testing.T) {
	c := emptyChronicle()
	a, b, kitchen := ids.Make(1, 1), ids.Make(2, 1), ids.Make(3, 1)
	for i, who := range []ids.ID{a, b} {
		c.add(StoryEntry{Tick: clock.Tick(100 + 5*i), Text: "x explora cocina.", Actor: who, Place: kitchen, plural: "%d tinies exploran cocina."})
	}
	c.add(StoryEntry{Tick: 900, Text: "Tiny 01 explora cocina.", Actor: a, Place: kitchen, plural: "%d tinies exploran cocina."})
	if len(c.Lines()) != 0 {
		t.Fatalf("la crónica tiene detalles: %+v", c.Lines())
	}
	d := c.Diary(kitchen)
	if len(d) != 2 || d[0].Text != "2 tinies exploran cocina." || d[0].Count != 2 || !d[0].Actor.IsNil() {
		t.Fatalf("diario de la cocina: %+v", d)
	}
	if len(c.Diary(a)) != 2 || len(c.Diary(b)) != 1 {
		t.Fatalf("diarios %d %d", len(c.Diary(a)), len(c.Diary(b)))
	}
}

// El resumen de la colonia sale cada media hora del reloj, solo si pasó
// algo, con la hora en que se escribe.
func TestChronicleSummary(t *testing.T) {
	c := emptyChronicle()
	c.tally[tallyStored], c.tally[tallyChat] = 2, 1
	c.Flush(SummaryTicks - 1)
	if len(c.Lines()) != 0 {
		t.Fatal("resumen antes de la media hora")
	}
	c.Flush(SummaryTicks)
	l := c.Lines()
	if len(l) != 1 || l[0].Level != LevelSummary || l[0].Tick != SummaryTicks ||
		l[0].Text != "Desde las 07:00, la colonia: 2 migajas al nido y 1 conversación." {
		t.Fatalf("resumen %+v", l)
	}
	c.Flush(2 * SummaryTicks)
	if len(c.Lines()) != 1 {
		t.Fatalf("resumen vacío: %+v", c.Lines())
	}
}

// La línea de una actividad se completa al terminar, en la crónica y en el
// diario, sin cambiar su Seq.
func TestChronicleRoutineCloses(t *testing.T) {
	c := emptyChronicle()
	g := ids.Make(9, 1)
	e := c.add(StoryEntry{Tick: 0, Text: "Desayuna en cocina.", Actor: g, Level: LevelRoutine})
	c.routine[g] = e.Seq
	c.closeRoutine(g, e.Seq, 24*20)
	for _, got := range []StoryEntry{c.Last(), c.Diary(g)[0]} {
		if got.Seq != e.Seq || got.End != 24*20 || got.Text != "Desayuna en cocina, hasta las 07:24." {
			t.Fatalf("rutina %+v", got)
		}
	}
}
