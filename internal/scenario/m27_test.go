package scenario_test

import (
	"fmt"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/scenario"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/world"
)

// descuido resume una corrida de «Un descuido en el escritorio».
type descuido struct {
	crushed, escapes, hesitations int // escapes: sobrevivientes que se salvaron por poco o huyeron
}

// playDescuido corre una versión del escenario de cierre con una semilla,
// guardando y cargando a cada rato, y verifica lo que vale siempre: las
// invariantes en cada tick, que una palma no aplaste a quien va sobre ella,
// que cada muerte de un miembro deje una zona de peligro que la colonia no
// elige como destino mientras dura, y que al final ningún muerto quede con
// trabajos, reservas ni cargas.
func playDescuido(t *testing.T, file string, seed uint64) descuido {
	t.Helper()
	spec, err := scenario.ReadFile("../../tests/scenarios/" + file + ".json")
	if err != nil {
		t.Fatal(err)
	}
	spec.Seed = seed
	in, err := scenario.Build(spec)
	if err != nil {
		t.Fatal(err)
	}
	w := in.World
	r := w.ECS()
	zones := map[ids.ID]bool{}
	playChecked(t, in, 113, func() {
		if err := w.CheckInvariants(); err != nil {
			t.Fatalf("tick %d: %v", w.Tick(), err)
		}
		for _, e := range w.Events() {
			switch ev := e.Event.(type) {
			case colony.MemberLost:
				zones[ev.Danger] = true
			case ai.ActionStarted:
				a, ok := ecs.Get[ai.Agent](r).Get(ev.Entity)
				l, ok2 := comp.Of(r).Location.Get(ev.Target)
				if !ok || !ok2 || a.Scale != "tiny" {
					continue
				}
				ecs.Get[ai.Danger](r).Each(func(z ids.ID, d *ai.Danger) {
					if zones[z] && d.Space == l.Space && d.Until > w.Tick() &&
						l.Pos.XY().DistSq(d.Center) <= int64(d.Radius)*int64(d.Radius) && ev.Action != ai.FleeAction {
						t.Errorf("tick %d: %v eligió %s en una zona de peligro", w.Tick(), ev.Entity, ev.Action)
					}
				})
			}
		}
	})

	var out descuido
	crushed := map[ids.ID]bool{}
	lost := 0
	for _, e := range in.Events {
		switch ev := e.Event.(type) {
		case world.CommandRejected:
			t.Fatalf("tick %d: comando rechazado: %s", e.Tick, ev.Error)
		case crush.Crushed:
			out.crushed++
			crushed[ev.Entity] = true
			if ev.Space == ev.By {
				t.Errorf("tick %d: la palma aplastó a quien iba sobre ella: %+v", e.Tick, ev)
			}
		case colony.MemberLost:
			lost++
		case ai.Hesitated:
			out.hesitations++
		}
	}
	if lost != out.crushed {
		t.Errorf("%d muertes y %d miembros perdidos", out.crushed, lost)
	}
	escaped := map[ids.ID]bool{}
	for _, e := range in.Events {
		switch ev := e.Event.(type) {
		case crush.NearMiss:
			if !crushed[ev.Entity] && ecs.Get[ai.Agent](r).Has(ev.Entity) {
				escaped[ev.Entity] = true
			}
		case ai.ActionDone:
			if ev.Action == ai.FleeAction && !crushed[ev.Entity] {
				escaped[ev.Entity] = true
			}
		}
	}
	out.escapes = len(escaped)
	ecs.Get[jobs.Job](r).Each(func(id ids.ID, j *jobs.Job) {
		if !j.Assignee.IsNil() && !r.Alive(j.Assignee) {
			t.Errorf("el trabajo %v quedó asignado a un muerto", id)
		}
	})
	ecs.Get[jobs.Reservation](r).Each(func(id ids.ID, v *jobs.Reservation) {
		if !r.Alive(v.By) {
			t.Errorf("%v quedó reservado por un muerto", id)
		}
	})
	ecs.Get[jobs.Carried](r).Each(func(id ids.ID, c *jobs.Carried) {
		if !r.Alive(c.By) {
			t.Errorf("%v quedó en manos de un muerto", id)
		}
	})
	return out
}

// Aceptación de M27 sobre las semillas de referencia: la giantess distraída
// aplasta al menos a un tiny y al menos otro se salva por poco; la cuidadosa
// titubea y mata menos en total.
func TestUnDescuidoEnElEscritorio(t *testing.T) {
	careless, careful := 0, 0
	for _, seed := range []uint64{1, 7, 21, 42, 99} {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			a := playDescuido(t, "un_descuido_en_el_escritorio", seed)
			if a.crushed < 1 || a.escapes < 1 {
				t.Errorf("distraída: %d aplastados y %d salvados por poco; se espera al menos uno de cada", a.crushed, a.escapes)
			}
			if a.hesitations != 0 {
				t.Errorf("distraída: titubeó %d veces", a.hesitations)
			}
			b := playDescuido(t, "un_descuido_cuidadosa", seed)
			if b.hesitations < 1 {
				t.Error("cuidadosa: no titubeó")
			}
			careless += a.crushed
			careful += b.crushed
			t.Logf("distraída: %d aplastados, %d salvados; cuidadosa: %d aplastados, %d titubeos",
				a.crushed, a.escapes, b.crushed, b.hesitations)
		})
	}
	if careful >= careless {
		t.Errorf("la cuidadosa aplastó %d y la distraída %d", careful, careless)
	}
}
