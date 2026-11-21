// Package content carga y valida las definiciones de objetos (gtsandbox/content):
// cada archivo objects/<nombre>.json es un spawn parcial (name, body,
// components...) que los escenarios usan como base con "object".
package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"
	"sync"

	root "gtsandbox/content"
	"gtsandbox/internal/sim/cmds"
	"gtsandbox/internal/sim/engine"
	"gtsandbox/internal/sim/world"
)

// Library son las definiciones cargadas.
type Library struct {
	objects map[string]json.RawMessage
}

// Load lee y valida objects/*.json de fsys: campos y componentes
// desconocidos son error, y el nombre debe coincidir con el del archivo.
func Load(fsys fs.FS) (*Library, error) {
	files, err := fs.Glob(fsys, "objects/*.json")
	if err != nil {
		return nil, err
	}
	probe, err := engine.New(world.Config{})
	if err != nil {
		return nil, err
	}
	lib := &Library{objects: map[string]json.RawMessage{}}
	for _, f := range files {
		raw, err := fs.ReadFile(fsys, f)
		if err != nil {
			return nil, err
		}
		sp, err := decode(probe, raw)
		if err != nil {
			return nil, fmt.Errorf("content: %s: %w", f, err)
		}
		if want := strings.TrimSuffix(path.Base(f), ".json"); string(sp.Name) != want {
			return nil, fmt.Errorf("content: %s: name %q no coincide con el archivo", f, sp.Name)
		}
		lib.objects[string(sp.Name)] = raw
	}
	return lib, nil
}

func decode(probe *world.World, raw []byte) (cmds.Spawn, error) {
	var sp cmds.Spawn
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&sp); err != nil {
		return sp, err
	}
	for _, n := range slices.Sorted(maps.Keys(sp.Components)) {
		if _, err := probe.ECS().DecodeComponent(n, sp.Components[n]); err != nil {
			return sp, err
		}
	}
	return sp, nil
}

var (
	once   sync.Once
	def    *Library
	defErr error
)

// Default es la biblioteca empaquetada en el binario (se carga una vez).
func Default() (*Library, error) {
	once.Do(func() { def, defErr = Load(root.FS) })
	return def, defErr
}

// Names devuelve los nombres de los objetos, ordenados.
func (l *Library) Names() []string {
	return slices.Sorted(maps.Keys(l.objects))
}

// Spawn combina la definición de object con overlay (un spawn parcial en
// JSON): los campos de overlay reemplazan a los de la definición, y sus
// components se agregan o reemplazan uno por uno.
func (l *Library) Spawn(object string, overlay json.RawMessage) (json.RawMessage, error) {
	base, ok := l.objects[object]
	if !ok {
		return nil, fmt.Errorf("content: objeto desconocido %q", object)
	}
	var b, o map[string]json.RawMessage
	if err := json.Unmarshal(base, &b); err != nil {
		return nil, err
	}
	if len(overlay) > 0 {
		if err := json.Unmarshal(overlay, &o); err != nil {
			return nil, err
		}
	}
	for k, v := range o {
		if k != "components" {
			b[k] = v
			continue
		}
		var bc, oc map[string]json.RawMessage
		if raw, ok := b[k]; ok {
			if err := json.Unmarshal(raw, &bc); err != nil {
				return nil, err
			}
		}
		if err := json.Unmarshal(v, &oc); err != nil {
			return nil, err
		}
		if bc == nil {
			bc = map[string]json.RawMessage{}
		}
		for ck, cv := range oc {
			bc[ck] = cv
		}
		merged, err := json.Marshal(bc)
		if err != nil {
			return nil, err
		}
		b[k] = merged
	}
	return json.Marshal(b)
}
