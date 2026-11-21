// Package sound decide qué suena y cómo, sin reproducir nada: traduce el
// estado de la presentación (interfaz, eventos, Frame) en pedidos de sonido
// para un Backend. Es presentación pura: no lee ni modifica el mundo, y no
// importa Ebitengine, así que se prueba sin dispositivo de audio.
//
// Los sonidos salen de tools/audio (sintetizados) y se describen en el
// manifiesto content/audio/sounds.json con un identificador estable.
package sound

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

// Category agrupa sonidos con un volumen propio.
type Category string

const (
	Ambient   Category = "ambiente"
	Effects   Category = "efectos"
	Interface Category = "interfaz"
)

// Categories en orden estable.
var Categories = []Category{Ambient, Effects, Interface}

// Def describe un sonido del manifiesto.
type Def struct {
	ID         string   `json:"id"`
	File       string   `json:"file"`
	Category   Category `json:"category"`
	MS         int      `json:"ms"`          // duración
	Voices     int      `json:"voices"`      // copias simultáneas como máximo
	CooldownMS int      `json:"cooldown_ms"` // tiempo mínimo entre comienzos
}

// Library es el manifiesto cargado.
type Library struct {
	Rate   int   `json:"rate"`
	Sounds []Def `json:"sounds"`
	Dir    string
	byID   map[string]Def
}

// Load lee un manifiesto; los archivos se resuelven junto a él.
func Load(path string) (*Library, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lib Library
	if err := json.Unmarshal(data, &lib); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	lib.Dir = filepath.Dir(path)
	lib.byID = map[string]Def{}
	for _, d := range lib.Sounds {
		if !slices.Contains(Categories, d.Category) {
			return nil, fmt.Errorf("%s: %s tiene categoría %q", path, d.ID, d.Category)
		}
		if d.Voices < 1 {
			return nil, fmt.Errorf("%s: %s necesita al menos una voz", path, d.ID)
		}
		if _, dup := lib.byID[d.ID]; dup {
			return nil, fmt.Errorf("%s: %s repetido", path, d.ID)
		}
		lib.byID[d.ID] = d
	}
	return &lib, nil
}

// Def devuelve la descripción de un sonido.
func (l *Library) Def(id string) (Def, bool) {
	d, ok := l.byID[id]
	return d, ok
}

// Path es el archivo WAV de un sonido.
func (l *Library) Path(d Def) string { return filepath.Join(l.Dir, d.File) }

// Backend reproduce los pedidos. El juego usa el audio de Ebitengine; las
// pruebas, un grabador.
type Backend interface {
	Play(id string, volume, pan float64)
	StopAll()
}

// Played es un pedido aceptado por el mezclador.
type Played struct {
	Frame  int64
	ID     string
	Volume float64
	Pan    float64
}
