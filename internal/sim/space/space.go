// Package space implementa la jerarquía de superficies multi-escala.
//
// Un Space es una entidad con el componente Space: una superficie con su
// propio sistema de coordenadas (el suelo de una habitación, un escritorio,
// la palma de una giantess). Toda entidad con comp.Location está dentro de
// exactamente un Space vivo.
//
// La jerarquía se arma solo con Location: un Space ubicado en otro se mueve
// con él. Así, mover a la giantess mueve su palma y a todo lo que esté
// encima, sin tocar sus posiciones locales:
//
//	P_mundo = T_habitación · T_giantess · T_palma · P_local
//
// Las transformaciones se calculan bajo demanda recorriendo la cadena
// (O(profundidad)); si el profiling lo pide se agregará una caché.
package space

import (
	"errors"
	"fmt"

	"gtsandbox/internal/core/fixedmath"
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/ecs"
)

// Kind clasifica un Space según cómo se comporta.
type Kind uint8

const (
	Static      Kind = iota // suelo de una habitación: no se mueve
	Elevated                // escritorio, mesa, cama: altura y accesos
	Rigid                   // zapato, taza: objeto móvil rígido
	Articulated             // mano, brazo: depende de la pose
	Body                    // marco de referencia de un personaje
	Built                   // construcción tiny, modificable
	kindCount
)

var kindNames = [...]string{"static", "elevated", "rigid", "articulated", "body", "built"}

func (k Kind) String() string {
	if k >= kindCount {
		return fmt.Sprintf("Kind(%d)", k)
	}
	return kindNames[k]
}

// MarshalText serializa el kind por nombre (snapshots legibles y estables).
func (k Kind) MarshalText() ([]byte, error) {
	if k >= kindCount {
		return nil, fmt.Errorf("space: kind inválido %d", k)
	}
	return []byte(kindNames[k]), nil
}

func (k *Kind) UnmarshalText(b []byte) error {
	for i, n := range kindNames {
		if n == string(b) {
			*k = Kind(i)
			return nil
		}
	}
	return fmt.Errorf("space: kind desconocido %q", b)
}

// Space es el componente de una superficie.
type Space struct {
	Kind       Kind        `json:"kind"`
	Extent     units.Rect  `json:"extent"`               // límites de la superficie, en coordenadas locales
	Projection *Projection `json:"projection,omitempty"` // proyección 3D opcional (nil = plano XY)
}

// MaxDepth limita la profundidad de la jerarquía; superarlo indica un ciclo.
const MaxDepth = 32

var (
	ErrNotSpace       = errors.New("space: no es un Space vivo")
	ErrNoFrame        = errors.New("space: la entidad no tiene Location ni es un Space")
	ErrCycle          = errors.New("space: la jerarquía tendría un ciclo")
	ErrTooDeep        = errors.New("space: jerarquía demasiado profunda")
	ErrDifferentRoots = errors.New("space: los Spaces pertenecen a raíces distintas")
)

// Graph da acceso a la jerarquía de un registry.
type Graph struct {
	reg    *ecs.Registry
	loc    *ecs.Store[comp.Location]
	spaces *ecs.Store[Space]
}

// Of devuelve el Graph de un registry con los stores ya registrados.
func Of(r *ecs.Registry) Graph {
	return Graph{reg: r, loc: ecs.Get[comp.Location](r), spaces: ecs.Get[Space](r)}
}

// IsSpace indica si e es un Space vivo.
func (g Graph) IsSpace(e ids.ID) bool { return g.reg.Alive(e) && g.spaces.Has(e) }

// Parent devuelve el Space que contiene a e, o Nil si e no tiene Location.
func (g Graph) Parent(e ids.ID) ids.ID {
	if l, ok := g.loc.Get(e); ok {
		return l.Space
	}
	return ids.Nil
}

// Frame devuelve la transformación de e respecto de su raíz: lleva puntos en
// coordenadas locales de e a coordenadas de la raíz. Para un Space raíz es la
// identidad.
func (g Graph) Frame(e ids.ID) (fixedmath.Transform, error) {
	t, _, err := g.walk(e)
	return t, err
}

// Root devuelve el Space raíz de la jerarquía que contiene a e.
func (g Graph) Root(e ids.ID) (ids.ID, error) {
	_, root, err := g.walk(e)
	return root, err
}

func (g Graph) walk(e ids.ID) (fixedmath.Transform, ids.ID, error) {
	if !g.reg.Alive(e) {
		return fixedmath.Identity, ids.Nil, fmt.Errorf("%w: %v", ErrNotSpace, e)
	}
	t := fixedmath.Identity
	cur := e
	for depth := 0; ; depth++ {
		l, ok := g.loc.Get(cur)
		if !ok {
			if !g.spaces.Has(cur) {
				return t, ids.Nil, fmt.Errorf("%w: %v", ErrNoFrame, cur)
			}
			return t, cur, nil
		}
		if depth >= MaxDepth {
			return t, ids.Nil, ErrTooDeep
		}
		if !g.IsSpace(l.Space) {
			return t, ids.Nil, fmt.Errorf("%w: %v (contenedor de %v)", ErrNotSpace, l.Space, cur)
		}
		t = l.Transform().Mul(t)
		cur = l.Space
	}
}

// ToWorld convierte un punto en coordenadas locales de s a coordenadas de su raíz.
func (g Graph) ToWorld(s ids.ID, p units.Vec3) (units.Vec3, error) {
	t, err := g.Frame(s)
	if err != nil {
		return units.Vec3{}, err
	}
	if sp, ok := g.spaces.Get(s); ok && sp.Projection != nil {
		p = sp.Project(p)
	}
	return t.Apply(p), nil
}

// FromWorld convierte un punto en coordenadas de la raíz a coordenadas locales de s.
func (g Graph) FromWorld(s ids.ID, p units.Vec3) (units.Vec3, error) {
	t, err := g.Frame(s)
	if err != nil {
		return units.Vec3{}, err
	}
	local := t.ApplyInv(p)
	if sp, ok := g.spaces.Get(s); ok && sp.Projection != nil {
		local = sp.Unproject(local)
	}
	return local, nil
}

// WorldPose devuelve posición y orientación de e en coordenadas de su raíz.
func (g Graph) WorldPose(e ids.ID) (units.Vec3, fixedmath.Angle, error) {
	if !g.spaces.Has(e) {
		if l, ok := g.loc.Get(e); ok {
			if sp, ok2 := g.spaces.Get(l.Space); ok2 && sp.Projection != nil {
				t, err := g.Frame(l.Space)
				if err != nil {
					return units.Vec3{}, 0, err
				}
				return t.Apply(sp.Project(l.Pos)), t.Rot.Add(l.Facing), nil
			}
		}
	}
	t, err := g.Frame(e)
	return t.Pos, t.Rot, err
}

// CanPlace verifica que e pueda ubicarse dentro de target: target es un Space
// vivo y no es e ni está montado (directa o indirectamente) sobre e.
// e puede ser Nil (entidad todavía no creada).
func (g Graph) CanPlace(e, target ids.ID) error {
	if !g.IsSpace(target) {
		return fmt.Errorf("%w: %v", ErrNotSpace, target)
	}
	if e.IsNil() {
		return nil
	}
	cur := target
	for depth := 0; !cur.IsNil(); depth++ {
		if cur == e {
			return fmt.Errorf("%w: %v dentro de %v", ErrCycle, e, target)
		}
		if depth > MaxDepth {
			return ErrTooDeep
		}
		cur = g.Parent(cur)
	}
	return nil
}

// Reparent mueve e al Space target conservando su posición y orientación en
// el mundo. Devuelve el Space anterior.
//
// Error de redondeo: exacto si todos los ángulos involucrados son múltiplos de
// 90°; si no, a lo sumo unos µm (despreciable frente a un tiny de 8 mm).
func (g Graph) Reparent(e, target ids.ID) (ids.ID, error) {
	old, ok := g.loc.Get(e)
	if !ok {
		return ids.Nil, fmt.Errorf("%w: %v", ErrNoFrame, e)
	}
	if err := g.CanPlace(e, target); err != nil {
		return ids.Nil, err
	}
	srcPos, srcRot, err := g.WorldPose(e)
	if err != nil {
		return ids.Nil, err
	}
	_, srcRoot, err := g.walk(e)
	if err != nil {
		return ids.Nil, err
	}
	dst, dstRoot, err := g.walk(target)
	if err != nil {
		return ids.Nil, err
	}
	if srcRoot != dstRoot {
		return ids.Nil, ErrDifferentRoots
	}
	dstPos := dst.ApplyInv(srcPos)
	if !g.spaces.Has(e) {
		if sp, ok := g.spaces.Get(target); ok && sp.Projection != nil {
			dstPos = sp.Unproject(dstPos)
		}
	}
	g.loc.Set(e, comp.Location{
		Space:  target,
		Pos:    dstPos,
		Facing: srcRot.Sub(dst.Rot),
	})
	return old.Space, nil
}

// CheckInvariants verifica que toda Location apunte a un Space vivo y que la
// jerarquía no tenga ciclos.
func (g Graph) CheckInvariants() error {
	var err error
	g.loc.Each(func(e ids.ID, _ *comp.Location) {
		if err == nil {
			_, _, err = g.walk(e)
		}
	})
	return err
}
