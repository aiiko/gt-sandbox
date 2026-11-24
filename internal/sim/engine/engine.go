// Package engine arma un World con todo lo que el motor provee: componentes,
// comandos, eventos y sistemas del núcleo.
//
// Es el único lugar que conoce la "forma" completa del mundo. Guardar,
// cargar y reproducir replays dependen de que el mundo tenga siempre la misma
// forma, así que todo código que cree mundos de juego debería pasar por acá.
package engine

import (
	"fmt"
	"io"

	"gtsandbox/internal/sim/ai"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/colony"
	"gtsandbox/internal/sim/comp"
	"gtsandbox/internal/sim/crush"
	"gtsandbox/internal/sim/encounter"
	"gtsandbox/internal/sim/expire"
	"gtsandbox/internal/sim/jobs"
	"gtsandbox/internal/sim/nav"
	"gtsandbox/internal/sim/needs"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/sim/spatial"
	"gtsandbox/internal/sim/world"
)

// Option permite agregar cosas a la forma del mundo (sistemas de test,
// contenido experimental). Las mismas opciones deben pasarse al cargar.
type Option func(*world.World) error

// New crea un mundo con todo el núcleo instalado.
func New(cfg world.Config, opts ...Option) (*world.World, error) {
	w := world.New(cfg)
	if _, err := comp.Register(w.ECS()); err != nil {
		return nil, fmt.Errorf("engine: componentes: %w", err)
	}
	if err := space.Install(w); err != nil {
		return nil, fmt.Errorf("engine: spaces: %w", err)
	}
	if err := nav.Install(w); err != nil {
		return nil, fmt.Errorf("engine: nav: %w", err)
	}
	if err := spatial.Install(w); err != nil {
		return nil, fmt.Errorf("engine: spatial: %w", err)
	}
	if err := attach.Install(w); err != nil {
		return nil, fmt.Errorf("engine: attach: %w", err)
	}
	if err := crush.Install(w); err != nil {
		return nil, fmt.Errorf("engine: crush: %w", err)
	}
	if err := needs.Install(w); err != nil {
		return nil, fmt.Errorf("engine: needs: %w", err)
	}
	if err := jobs.Install(w); err != nil {
		return nil, fmt.Errorf("engine: jobs: %w", err)
	}
	if err := ai.Install(w, append(colony.Actions(), encounter.Actions()...)...); err != nil {
		return nil, fmt.Errorf("engine: ai: %w", err)
	}
	if err := colony.Install(w); err != nil {
		return nil, fmt.Errorf("engine: colony: %w", err)
	}
	if err := encounter.Install(w); err != nil {
		return nil, fmt.Errorf("engine: encounter: %w", err)
	}
	if err := expire.Install(w); err != nil {
		return nil, fmt.Errorf("engine: expire: %w", err)
	}
	if err := cmds.Register(w); err != nil {
		return nil, fmt.Errorf("engine: comandos: %w", err)
	}
	for _, o := range opts {
		if err := o(w); err != nil {
			return nil, err
		}
	}
	return w, nil
}

// MustNew es New con panic. Para tests y herramientas.
func MustNew(cfg world.Config, opts ...Option) *world.World {
	w, err := New(cfg, opts...)
	if err != nil {
		panic(err)
	}
	return w
}

// Restore crea un mundo con la forma estándar y le carga el snapshot.
func Restore(s *world.Snapshot, opts ...Option) (*world.World, error) {
	w, err := New(world.Config{Seed: s.Seed}, opts...)
	if err != nil {
		return nil, err
	}
	if err := w.Restore(s); err != nil {
		return nil, err
	}
	return w, nil
}

// Load lee un snapshot JSON y lo restaura.
func Load(in io.Reader, opts ...Option) (*world.World, error) {
	s, err := world.ReadSnapshot(in)
	if err != nil {
		return nil, err
	}
	return Restore(s, opts...)
}
