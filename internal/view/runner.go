package view

import (
	"encoding/json"
	"slices"
	"time"

	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Speeds son las velocidades que ofrece el visor (ticks simulados por tick
// real).
var Speeds = []int{1, 4, 16}

// MaxStepsPerFrame limita cuánto se pone al día un frame lento: el tiempo
// que no entra se descarta (la simulación se ralentiza en vez de trabarse).
const MaxStepsPerFrame = 64

// LogSize es la cantidad de eventos que guarda el Runner.
const LogSize = 200

// Logged es un evento publicado, ya serializado para mostrarlo.
type Logged struct {
	Tick clock.Tick
	Kind string
	Data string
}

// Runner avanza un mundo según el tiempo real, con pausa, paso a paso y
// velocidad. Solo decide cuántas veces llamar a Step: la simulación es la
// misma a cualquier velocidad o tasa de frames.
type Runner struct {
	World  *world.World
	step   func() error
	cap    Capturer
	Paused bool
	Speed  int // multiplicador; uno de Speeds
	acc    time.Duration

	Prev, Cur *Frame       // estado antes y después del último Step
	Log       []Logged     // últimos LogSize eventos, del más viejo al más nuevo
	Story     []StoryEntry // últimos LogSize resúmenes, en orden de publicación
	Err       error        // primer error de step; detiene el Runner
	falls     []Fall       // caídas recientes, para animarlas

	hashed   bool
	hashTick clock.Tick
	hash     uint64
}

// NewRunner crea un Runner. step avanza el mundo un tick (w.Step, o el de un
// escenario, que además encola sus comandos).
func NewRunner(w *world.World, step func() error) *Runner {
	r := &Runner{World: w, step: step, Speed: 1}
	r.Cur = r.cap.Capture(w)
	r.Prev = r.Cur
	return r
}

// Advance suma dt de tiempo real y corre los Steps que correspondan.
// Devuelve cuántos corrió.
func (r *Runner) Advance(dt time.Duration) int {
	if r.Paused || r.Err != nil {
		return 0
	}
	r.acc += dt * time.Duration(r.Speed)
	n := int(r.acc / clock.TickDuration)
	r.acc -= time.Duration(n) * clock.TickDuration
	if n > MaxStepsPerFrame {
		n, r.acc = MaxStepsPerFrame, 0
	}
	r.steps(n)
	return n
}

// StepOnce corre un solo tick (para el paso a paso en pausa).
func (r *Runner) StepOnce() {
	if r.Err == nil {
		r.acc = 0
		r.steps(1)
	}
}

func (r *Runner) steps(n int) {
	for i := range n {
		if i == n-1 {
			if n > 1 {
				r.Prev = r.cap.Capture(r.World)
			} else {
				r.Prev = r.Cur
			}
		}
		if err := r.step(); err != nil {
			r.Err = err
			r.Paused = true
			break
		}
		r.record()
	}
	if n > 0 {
		r.Cur = r.cap.Capture(r.World)
		r.attachFalls(r.Cur)
		if n > 1 {
			r.attachFalls(r.Prev)
		}
	}
}

// attachFalls copia al frame las caídas que siguen animándose en su tick.
func (r *Runner) attachFalls(f *Frame) {
	r.falls = slices.DeleteFunc(r.falls, func(fl Fall) bool { return r.Cur.Tick-fl.Tick >= FallTicks })
	f.Falls = nil
	for _, fl := range r.falls {
		if fl.Tick <= f.Tick && f.Tick-fl.Tick < FallTicks {
			f.Falls = append(f.Falls, fl)
		}
	}
}

func (r *Runner) record() {
	first := len(r.Story)
	for _, e := range r.World.Events() {
		if fell, ok := e.Event.(space.EntityFell); ok {
			// cuenta desde el primer frame que ya la muestra resuelta
			r.falls = append(r.falls, Fall{Entity: fell.Entity, From: fell.From, To: fell.To, Height: fell.Height, Tick: r.World.Tick()})
		}
		data, err := json.Marshal(e.Event)
		if err != nil {
			data = []byte(err.Error())
		}
		r.Log = append(r.Log, Logged{Tick: e.Tick, Kind: e.Event.Kind(), Data: string(data)})
		if summary := storyText(r.World.ECS(), e.Event); summary != "" {
			entry := StoryEntry{Tick: e.Tick, Text: summary}
			if !slices.Contains(r.Story[first:], entry) {
				r.Story = append(r.Story, entry)
			}
		}
	}
	if over := len(r.Log) - LogSize; over > 0 {
		r.Log = append(r.Log[:0], r.Log[over:]...)
	}
	if over := len(r.Story) - LogSize; over > 0 {
		r.Story = append(r.Story[:0], r.Story[over:]...)
	}
}

// Alpha es la fracción de tick transcurrida desde el último Step, para
// interpolar entre Prev y Cur. En pausa es 1 (se dibuja Cur).
func (r *Runner) Alpha() float64 {
	if r.Paused {
		return 1
	}
	return min(float64(r.acc)/float64(clock.TickDuration), 1)
}

// Hash devuelve el hash del estado, calculado una vez por tick.
func (r *Runner) Hash() uint64 {
	if t := r.World.Tick(); !r.hashed || t != r.hashTick {
		r.hashed, r.hashTick, r.hash = true, t, r.World.Hash()
	}
	return r.hash
}

// Enqueue manda un comando del jugador (se aplica en el próximo Step).
func (r *Runner) Enqueue(c world.Command) error {
	_, err := r.World.Enqueue(c)
	return err
}
