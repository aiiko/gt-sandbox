package colony

import (
	"errors"
	"fmt"
	"slices"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/world"
)

// Reunión. Un punto de reunión (gathering) junto al nido tiene dos plazas.
// Un tiny bien alimentado y descansado, sin trabajo, toma una plaza y
// espera al compañero hasta Wait; cuando están los dos conversan TalkTicks
// y a ambos les sube "fun". Al quedar vacío, el punto no se ofrece hasta
// pasada la pausa.

// Valores por defecto y umbrales de la reunión.
const (
	ChatAction                  = "chat"
	ChatNeed                    = "fun"
	ChatReady                   = 500 // hambre y energía mínimas para ir a conversar
	DefaultChatWait  clock.Tick = 20 * clock.TPS
	DefaultTalkTicks clock.Tick = 8 * clock.TPS
	DefaultChatPause clock.Tick = 30 * clock.TPS
	DefaultChatGain             = 300
	// SeatOffset es cuánto se separa cada plaza del centro del punto.
	SeatOffset = 15 * units.Millimeter
	chatSeats  = 2
)

// Seat es una plaza ocupada: quién y si ya llegó.
type Seat struct {
	By      ids.ID `json:"by"`
	Arrived bool   `json:"arrived,omitempty"`
}

// Gathering es el componente "gathering" del punto de reunión. Talk es el
// tick en que termina la conversación en curso (0 = nadie conversa) y
// Next, desde cuándo vuelve a ofrecerse.
type Gathering struct {
	Wait  clock.Tick `json:"wait,omitempty"`
	Talk  clock.Tick `json:"talk_ticks,omitempty"`
	Pause clock.Tick `json:"pause,omitempty"`
	Gain  int32      `json:"gain,omitempty"`
	Next  clock.Tick `json:"next,omitempty"`
	Until clock.Tick `json:"until,omitempty"`
	Seats []Seat     `json:"seats,omitempty"`
	// Pair son los que conversan (hasta que se cuenta la conversación).
	Pair []ids.ID `json:"pair,omitempty"`
}

func (g *Gathering) wait() clock.Tick  { return cmpOr(g.Wait, DefaultChatWait) }
func (g *Gathering) talk() clock.Tick  { return cmpOr(g.Talk, DefaultTalkTicks) }
func (g *Gathering) pause() clock.Tick { return cmpOr(g.Pause, DefaultChatPause) }
func (g *Gathering) gain() int32       { return cmpOr(g.Gain, DefaultChatGain) }

func (g *Gathering) seat(e ids.ID) int {
	return slices.IndexFunc(g.Seats, func(s Seat) bool { return s.By == e })
}

// Eventos de la reunión.
type (
	ChatStarted struct {
		Gathering ids.ID `json:"gathering"`
		A         ids.ID `json:"a"`
		B         ids.ID `json:"b"`
	}
	ChatDone struct {
		Gathering ids.ID `json:"gathering"`
		A         ids.ID `json:"a"`
		B         ids.ID `json:"b"`
	}
	// ChatAbandoned: Reason es "alone" (no llegó nadie) o "partner_left".
	ChatAbandoned struct {
		Gathering ids.ID `json:"gathering"`
		Entity    ids.ID `json:"entity"`
		Reason    string `json:"reason"`
	}
)

func (ChatStarted) Kind() string   { return "chat_started" }
func (ChatDone) Kind() string      { return "chat_done" }
func (ChatAbandoned) Kind() string { return "chat_abandoned" }

// Chat es la acción "chat". Doing: Target es el punto; Stage 0 camina a su
// plaza, 1 espera al compañero (Progress), 2 conversa.
type Chat struct{}

func (Chat) Name() string { return ChatAction }

// funGain es la utilidad de subir "fun" en amount, con la curva de las
// necesidades (como una oferta). 0 si se desperdiciaría.
func funGain(c *ai.Ctx, amount int32) float64 {
	v, ok := c.Needs.Get(ChatNeed)
	if !ok || v+amount/2 > needs.Max {
		return 0
	}
	d := func(v int32) float64 {
		x := float64(needs.Max-min(v, needs.Max)) / needs.Max
		return x * x
	}
	return c.Agent.Weight(ChatNeed) * (d(v) - d(v+amount))
}

// ready: con hambre y energía por encima de ChatReady (las que tenga).
func ready(c *ai.Ctx) bool {
	for _, n := range []string{"hunger", "energy"} {
		if v, ok := c.Needs.Get(n); ok && v < ChatReady {
			return false
		}
	}
	return true
}

func (Chat) Options(c *ai.Ctx) []ai.Option {
	r := c.W.ECS()
	if c.Needs == nil || !ecs.Get[Member](r).Has(c.Actor) || !ready(c) || !jobs.AssignedTo(c.W, c.Actor).IsNil() {
		return nil
	}
	locs := comp.Of(r).Location
	var out []ai.Option
	ecs.Get[Gathering](r).Each(func(e ids.ID, g *Gathering) {
		l, ok := locs.Get(e)
		if !ok || r.Despawning(e) || !c.Agent.Knows(e) || c.W.Tick() < g.Next || g.Until > 0 ||
			len(g.Seats) >= chatSeats || g.seat(c.Actor) >= 0 {
			return
		}
		u := funGain(c, g.gain())
		if u <= 0 {
			return
		}
		out = append(out, ai.Option{Target: e, Utility: u - c.Agent.DistanceCost*ai.Distance(c, l.Space, l.Pos.XY())})
	})
	return out
}

func (Chat) Utility(c *ai.Ctx, d *ai.Doing) float64 {
	g, ok := ecs.Get[Gathering](c.W.ECS()).Get(d.Target)
	if !ok {
		return 0
	}
	return funGain(c, g.gain())
}

func (Chat) Start(c *ai.Ctx, d *ai.Doing) bool {
	g := ecs.Get[Gathering](c.W.ECS()).Ptr(d.Target)
	if g == nil || g.Until > 0 || len(g.Seats) >= chatSeats || g.seat(c.Actor) >= 0 {
		return false
	}
	g.Seats = append(g.Seats, Seat{By: c.Actor})
	return true
}

func (Chat) Step(c *ai.Ctx, d *ai.Doing) ai.Result {
	w := c.W
	g := ecs.Get[Gathering](w.ECS()).Ptr(d.Target)
	l, ok := comp.Of(w.ECS()).Location.Get(d.Target)
	if g == nil || !ok {
		return ai.Failed
	}
	i := g.seat(c.Actor)
	if i < 0 {
		return ai.Failed
	}
	switch d.Stage {
	case 0: // a la plaza
		off := units.V2(-SeatOffset, 0)
		if i == 1 {
			off = units.V2(SeatOffset, 0)
		}
		switch nav.Approach(w, c.Actor, l.Space, l.Pos.XY().Add(off), &d.Moving) {
		case nav.Failed:
			return ai.Failed
		case nav.Underway:
			return ai.Running
		}
		g.Seats[i].Arrived = true
		d.Stage, d.Moving, d.Progress = 1, false, 0
		fallthrough
	case 1: // esperar al compañero
		if g.Until == 0 && len(g.Seats) == chatSeats && g.Seats[0].Arrived && g.Seats[1].Arrived {
			g.Until = w.Tick() + g.talk()
			g.Pair = []ids.ID{g.Seats[0].By, g.Seats[1].By}
			w.Emit(ChatStarted{Gathering: d.Target, A: g.Seats[0].By, B: g.Seats[1].By})
		}
		if g.Until > 0 {
			d.Stage = 2
			return ai.Running
		}
		if d.Progress++; d.Progress >= g.wait() {
			w.Emit(ChatAbandoned{Gathering: d.Target, Entity: c.Actor, Reason: "alone"})
			return ai.Failed
		}
		return ai.Running
	case 2: // conversar
		if w.Tick() >= g.Until { // terminan los dos en el mismo tick
			if n := ecs.Get[needs.Needs](w.ECS()).Ptr(c.Actor); n != nil {
				n.Add(ChatNeed, g.gain())
			}
			if len(g.Pair) == chatSeats { // el primero cuenta la conversación
				w.Emit(ChatDone{Gathering: d.Target, A: g.Pair[0], B: g.Pair[1]})
				g.Pair = nil
			}
			return ai.Done
		}
		if len(g.Seats) < chatSeats {
			w.Emit(ChatAbandoned{Gathering: d.Target, Entity: c.Actor, Reason: "partner_left"})
			return ai.Failed
		}
		return ai.Running
	}
	return ai.Failed
}

// Stop libera la plaza. Con el punto vacío empieza la pausa.
func (Chat) Stop(c *ai.Ctx, d *ai.Doing, _ bool) {
	if d.Moving {
		nav.Halt(c.W, c.Actor)
	}
	g := ecs.Get[Gathering](c.W.ECS()).Ptr(d.Target)
	if g == nil {
		return
	}
	leave(c.W, g, c.Actor)
}

func leave(w *world.World, g *Gathering, e ids.ID) {
	if i := g.seat(e); i >= 0 {
		g.Seats = slices.Delete(g.Seats, i, i+1)
	}
	if len(g.Seats) == 0 {
		g.Seats, g.Pair = nil, nil
		g.Until = 0
		g.Next = w.Tick() + g.pause()
	}
}

func (Chat) Valid(c *ai.Ctx, d *ai.Doing) bool {
	g, ok := ecs.Get[Gathering](c.W.ECS()).Get(d.Target)
	return ok && g.seat(c.Actor) >= 0
}

func (Chat) Interruptible(*ai.Ctx, *ai.Doing) bool { return true }

// gather libera las plazas de quien murió (Decide, antes del selector).
type gather struct{}

func (gather) Name() string { return "colony.gather" }

func (gather) Run(w *world.World) {
	r := w.ECS()
	ecs.Get[Gathering](r).Each(func(e ids.ID, g *Gathering) {
		for _, s := range slices.Clone(g.Seats) {
			if !r.Alive(s.By) || r.Despawning(s.By) {
				leave(w, g, s.By)
			}
		}
	})
}

func checkGatherings(w *world.World) error {
	r := w.ECS()
	agents := ecs.Get[ai.Agent](r)
	var err error
	ecs.Get[Gathering](r).Each(func(e ids.ID, g *Gathering) {
		if len(g.Seats) > chatSeats {
			err = errors.Join(err, fmt.Errorf("colony: reunión %v con %d plazas", e, len(g.Seats)))
		}
		for i, s := range g.Seats {
			if g.seat(s.By) != i {
				err = errors.Join(err, fmt.Errorf("colony: reunión %v con %v dos veces", e, s.By))
			}
			if !r.Alive(s.By) || r.Despawning(s.By) {
				continue // se libera en el próximo tick
			}
			a := agents.Ptr(s.By)
			chatting := func(d *ai.Doing) bool { return d != nil && d.Action == ChatAction && d.Target == e }
			if a == nil || !chatting(a.Doing) && !chatting(a.Suspended) {
				err = errors.Join(err, fmt.Errorf("colony: %v ocupa la reunión %v sin conversar", s.By, e))
			}
		}
	})
	return err
}

// registerGather registra el store, los eventos, el sistema y el invariante
// de la reunión.
func registerGather(w *world.World) error {
	if err := register[Gathering](w.ECS(), "gathering"); err != nil {
		return err
	}
	w.AddInvariant("colony.gather", checkGatherings)
	return errors.Join(
		world.RegisterEvent[ChatStarted](w),
		world.RegisterEvent[ChatDone](w),
		world.RegisterEvent[ChatAbandoned](w),
		w.Register(world.SystemSpec{Phase: clock.PhaseDecide, Order: -38}, gather{}),
	)
}
