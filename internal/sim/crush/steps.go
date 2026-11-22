package crush

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// Stride es el componente "stride" de quien pisa al caminar (la giantess).
// Cada Length recorridos en su Space apoya un pie, alternando izquierda y
// derecha a Side del centro de la marcha. La huella es un rectángulo de
// FootLength × FootWidth orientado según la dirección del desplazamiento
// (no según facing: el mover no la orienta).
//
// Travel, Left, Space y Last son estado.
type Stride struct {
	Length     units.Length `json:"length"`
	Side       units.Length `json:"side"`
	FootLength units.Length `json:"foot_length"`
	FootWidth  units.Length `json:"foot_width"`
	Force      int32        `json:"force"`

	Travel units.Length `json:"travel,omitempty"` // recorrido desde la última pisada
	Right  bool         `json:"right,omitempty"`  // la próxima pisada es del pie derecho
	Space  ids.ID       `json:"space,omitempty"`  // Space de Last
	Last   units.Vec2   `json:"last"`             // posición del tick anterior
}

// Footstep se emite con cada pisada. Pos es el centro del pie en Space.
type Footstep struct {
	Entity ids.ID     `json:"entity"`
	Space  ids.ID     `json:"space"`
	Pos    units.Vec2 `json:"pos"`
	Right  bool       `json:"right"`
}

func (Footstep) Kind() string { return "footstep" }

// SourceFoot es el source de los aplastamientos por pisada.
const SourceFoot = "foot"

// steps (fase Move, después de nav.move) avanza el recorrido de cada Stride
// y apoya los pies. Una pisada aplasta en el acto lo frágil que está en el
// mismo Space (el suelo de la habitación), dentro de la huella: lo que está
// sobre un mueble está en otro Space y no lo alcanza.
type steps struct{}

func (steps) Name() string { return "crush.steps" }

type footfall struct {
	e      ids.ID
	s      Stride
	space  ids.ID
	center units.Vec2
	dir    units.Vec2 // dirección de la marcha (no nula)
}

func (steps) Run(w *world.World) {
	r := w.ECS()
	locs := comp.Of(r).Location
	var falls []footfall
	ecs.Get[Stride](r).Each(func(e ids.ID, st *Stride) {
		l, ok := locs.Get(e)
		if !ok || r.Despawning(e) || st.Length <= 0 {
			return
		}
		pos := l.Pos.XY()
		if st.Space != l.Space { // cambió de Space (una puerta): empieza de nuevo
			st.Space, st.Last, st.Travel = l.Space, pos, 0
			return
		}
		d := pos.Sub(st.Last)
		st.Last = pos
		n := d.Len()
		if n == 0 {
			return
		}
		st.Travel += n
		for st.Travel >= st.Length {
			st.Travel -= st.Length
			side := st.Side
			if st.Right {
				side = -side
			}
			// izquierda = la marcha girada 90° antihorario
			off := units.V2(-d.Y, d.X)
			center := pos.Add(units.V2(off.X.MulDiv(int64(side), int64(n)), off.Y.MulDiv(int64(side), int64(n))))
			falls = append(falls, footfall{e: e, s: *st, space: l.Space, center: center, dir: d})
			st.Right = !st.Right
		}
	})
	g := space.Of(r)
	for _, f := range falls {
		w.Emit(Footstep{Entity: f.e, Space: f.space, Pos: f.center, Right: f.s.Right})
		im := impact{by: f.e, c: Crusher{Source: SourceFoot, Force: f.s.Force}, surface: f.space, inside: footprint(g, f)}
		for _, v := range victims(w, g, im) {
			kill(w, g, v, im)
		}
	}
}

// footprint devuelve la prueba de la huella de un pie, en coordenadas del
// mundo.
func footprint(g space.Graph, f footfall) func(units.Vec3) bool {
	n := int64(f.dir.Len())
	halfL, halfW := int64(f.s.FootLength)/2, int64(f.s.FootWidth)/2
	return func(p units.Vec3) bool {
		l, err := g.FromWorld(f.space, p)
		if err != nil {
			return false
		}
		v := l.XY().Sub(f.center)
		along := int64(v.X)*int64(f.dir.X) + int64(v.Y)*int64(f.dir.Y)  // · dir
		across := int64(v.Y)*int64(f.dir.X) - int64(v.X)*int64(f.dir.Y) // × dir
		return abs(along) <= halfL*n && abs(across) <= halfW*n
	}
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
