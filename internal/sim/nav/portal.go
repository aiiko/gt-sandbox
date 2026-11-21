package nav

import (
	"fmt"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/ecs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/world"
)

// PortalEnd es un extremo de un portal: un punto en coordenadas locales de
// un Space.
type PortalEnd struct {
	Space ids.ID     `json:"space"`
	Pos   units.Vec2 `json:"pos"`
}

// Portal conecta dos Spaces en ambos sentidos: un mover parado en un extremo
// pasa al otro. Es el componente de una entidad propia (store "portal"); los
// portales estáticos vienen del mapa, los dinámicos llegarán con los
// contactos (M5).
//
// Por ahora los extremos son puntos; si hace falta, serán segmentos.
type Portal struct {
	A PortalEnd `json:"a"`
	B PortalEnd `json:"b"`
	// Cost es el costo extra de cruzar, como distancia equivalente. El
	// recorrido físico (trepar una pata) se modela mejor como un Space propio.
	Cost     units.Length `json:"cost"`
	Disabled bool         `json:"disabled,omitempty"`
	// Version sube con cada cambio; las rutas la recuerdan.
	Version uint64 `json:"version"`
}

// ends devuelve los extremos de entrada y salida al cruzar desde A
// (reverse=false) o desde B (reverse=true).
func (p *Portal) ends(reverse bool) (in, out PortalEnd) {
	if reverse {
		return p.B, p.A
	}
	return p.A, p.B
}

// SetPortal habilita o deshabilita un portal.
type SetPortal struct {
	Portal  ids.ID `json:"portal"`
	Enabled bool   `json:"enabled"`
}

func (SetPortal) Kind() string { return "set_portal" }

func (c SetPortal) Apply(w *world.World) error {
	p := ecs.Get[Portal](w.ECS()).Ptr(c.Portal)
	if p == nil {
		return fmt.Errorf("set_portal %v: no es un portal", c.Portal)
	}
	if p.Disabled == c.Enabled {
		p.Disabled = !c.Enabled
		p.Version++
	}
	return nil
}

// portalCleanup destruye los portales cuyos extremos están en Spaces que se
// destruyen. Corre al final del Flush, después de space.orphans.
type portalCleanup struct{}

func (portalCleanup) Name() string { return "nav.portals" }

func (portalCleanup) Run(w *world.World) {
	r := w.ECS()
	g := space.Of(r)
	gone := func(s ids.ID) bool { return !g.IsSpace(s) || r.Despawning(s) }
	ecs.Get[Portal](r).Each(func(e ids.ID, p *Portal) {
		if !r.Despawning(e) && (gone(p.A.Space) || gone(p.B.Space)) {
			r.Despawn(e)
		}
	})
}

func checkPortals(w *world.World) error {
	g := space.Of(w.ECS())
	var err error
	ecs.Get[Portal](w.ECS()).Each(func(e ids.ID, p *Portal) {
		if err == nil && (!g.IsSpace(p.A.Space) || !g.IsSpace(p.B.Space)) {
			err = fmt.Errorf("portal %v con extremo fuera de un Space vivo", e)
		}
	})
	return err
}
