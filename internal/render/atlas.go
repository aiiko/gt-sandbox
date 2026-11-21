package render

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// Scale indica qué variante del sprite se usa.
type Scale string

const (
	Macro Scale = "macro"
	Micro Scale = "micro"
)

type atlasManifest struct {
	Image   string                       `json:"image"`
	Scales  []Scale                      `json:"scales,omitempty"`
	Sprites map[string]map[Scale]variant `json:"sprites"`
}

type variant struct {
	Pivot        []int           `json:"pivot"`
	Clips        map[string]clip `json:"clips"`
	Layer        string          `json:"layer,omitempty"`
	DepthOffset  int             `json:"depth_offset,omitempty"`
	Foreground   []int           `json:"foreground,omitempty"`
	Focus        []int           `json:"focus,omitempty"`
	Carry        []int           `json:"carry,omitempty"`
	Hidden       bool            `json:"hidden,omitempty"`
	SelectedOnly bool            `json:"selected_only,omitempty"`
}

type clip struct {
	FrameTicks uint64  `json:"frame_ticks"`
	Frames     [][]int `json:"frames"`
}

// Atlas conserva el PNG para consultar el alfa y crea la textura al dibujar.
type Atlas struct {
	source  string
	pixels  image.Image
	texture *ebiten.Image
	scales  []Scale
	sprites map[string]map[Scale]variant
}

// LoadAtlas valida todas las variantes y cuadros antes de crear la textura.
func LoadAtlas(fsys fs.FS, file string) (*Atlas, error) {
	f, err := fsys.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var manifest atlasManifest
	d := json.NewDecoder(f)
	d.DisallowUnknownFields()
	if err := d.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("atlas %s: %w", file, err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("atlas %s: contenido extra después del manifiesto", file)
	}
	if !fs.ValidPath(manifest.Image) || manifest.Image == "." {
		return nil, fmt.Errorf("atlas %s: ruta de imagen inválida %q", file, manifest.Image)
	}
	pngPath := path.Join(path.Dir(file), manifest.Image)
	img, err := fsys.Open(pngPath)
	if err != nil {
		return nil, fmt.Errorf("atlas %s: %w", file, err)
	}
	defer img.Close()
	pixels, err := png.Decode(img)
	if err != nil {
		return nil, fmt.Errorf("atlas %s, PNG %s: %w", file, pngPath, err)
	}
	if len(manifest.Sprites) == 0 {
		return nil, fmt.Errorf("atlas %s: falta al menos un sprite", file)
	}
	if manifest.Scales == nil {
		manifest.Scales = []Scale{Macro, Micro}
	}
	if len(manifest.Scales) == 0 {
		return nil, fmt.Errorf("atlas %s: falta al menos una escala", file)
	}
	for i, scale := range manifest.Scales {
		if (scale != Macro && scale != Micro) || slices.Contains(manifest.Scales[:i], scale) {
			return nil, fmt.Errorf("atlas %s: escala inválida o duplicada %q", file, scale)
		}
	}
	for _, id := range slices.Sorted(maps.Keys(manifest.Sprites)) {
		if id == "" {
			return nil, fmt.Errorf("atlas %s: identificador de sprite vacío", file)
		}
		variants := manifest.Sprites[id]
		for _, scale := range slices.Sorted(maps.Keys(variants)) {
			if !slices.Contains(manifest.Scales, scale) {
				return nil, fmt.Errorf("atlas %s, sprite %q: variante desconocida %q", file, id, scale)
			}
		}
		for _, scale := range manifest.Scales {
			v, ok := variants[scale]
			context := fmt.Sprintf("atlas %s, sprite %q, variante %s", file, id, scale)
			if !ok {
				return nil, fmt.Errorf("%s: falta la variante", context)
			}
			if err := validateVariant(v, pixels.Bounds(), context); err != nil {
				return nil, err
			}
		}
	}
	return &Atlas{source: file, pixels: pixels, scales: manifest.Scales, sprites: manifest.Sprites}, nil
}

func validateVariant(v variant, bounds image.Rectangle, context string) error {
	if v.Hidden {
		if len(v.Pivot) != 0 || len(v.Clips) != 0 || v.Layer != "" || v.DepthOffset != 0 ||
			v.Foreground != nil || v.Focus != nil || v.Carry != nil || v.SelectedOnly {
			return fmt.Errorf("%s: la variante oculta tiene datos de dibujo", context)
		}
		return nil
	}
	if len(v.Pivot) != 2 || v.Pivot[0] < 0 || v.Pivot[1] < 0 {
		return fmt.Errorf("%s: pivote inválido %v", context, v.Pivot)
	}
	if _, ok := v.Clips["idle"]; !ok {
		return fmt.Errorf("%s: falta el clip idle", context)
	}
	if v.Layer != "" && v.Layer != "background" && v.Layer != "ground" {
		return fmt.Errorf("%s: capa desconocida %q", context, v.Layer)
	}
	if v.Foreground != nil {
		if !validFrame(v.Foreground, bounds) || v.Pivot[0] >= v.Foreground[2] || v.Pivot[1] >= v.Foreground[3] {
			return fmt.Errorf("%s: rectángulo o pivote de primer plano inválido", context)
		}
	}
	if v.Carry != nil && len(v.Carry) != 2 {
		return fmt.Errorf("%s: se requieren dos coordenadas para carry", context)
	}
	if v.Focus != nil && (len(v.Focus) != 2 || v.Focus[0] < 0 || v.Focus[1] < 0) {
		return fmt.Errorf("%s: foco inválido %v", context, v.Focus)
	}
	for _, name := range slices.Sorted(maps.Keys(v.Clips)) {
		c := v.Clips[name]
		if name == "" || c.FrameTicks == 0 || len(c.Frames) == 0 {
			return fmt.Errorf("%s, clip %q: se requieren nombre, duración y cuadros", context, name)
		}
		for i, frame := range c.Frames {
			if !validFrame(frame, bounds) {
				return fmt.Errorf("%s, clip %q, cuadro %d: rectángulo fuera del PNG o inválido %v", context, name, i+1, frame)
			}
			if v.Pivot[0] >= frame[2] || v.Pivot[1] >= frame[3] {
				return fmt.Errorf("%s, clip %q, cuadro %d: pivote %v fuera del cuadro", context, name, i+1, v.Pivot)
			}
			if v.DepthOffset < -frame[3] || v.DepthOffset > frame[3] {
				return fmt.Errorf("%s, clip %q, cuadro %d: desplazamiento de profundidad fuera del cuadro", context, name, i+1)
			}
			if v.Focus != nil && (v.Focus[0] >= frame[2] || v.Focus[1] >= frame[3]) {
				return fmt.Errorf("%s, clip %q, cuadro %d: foco fuera del cuadro", context, name, i+1)
			}
		}
	}
	return nil
}

func validFrame(r []int, bounds image.Rectangle) bool {
	return len(r) == 4 && r[0] >= 0 && r[1] >= 0 && r[2] > 0 && r[3] > 0 &&
		r[0] <= bounds.Dx()-r[2] && r[1] <= bounds.Dy()-r[3]
}

// Require verifica que el atlas contenga un identificador usado por la escena.
func (a *Atlas) Require(id string) error {
	if _, ok := a.sprites[id]; !ok {
		return fmt.Errorf("atlas %s: falta el sprite %q", a.source, id)
	}
	return nil
}

func (a *Atlas) Supports(scale Scale) bool { return slices.Contains(a.scales, scale) }

func (a *Atlas) Visible(id string, scale Scale) bool {
	v, ok := a.sprites[id][scale]
	return ok && !v.Hidden
}

// MicroFocus devuelve el desplazamiento del centro de interés respecto del pivote.
func (a *Atlas) MicroFocus(id string) (image.Point, bool) {
	v := a.sprites[id][Micro]
	if v.Focus == nil {
		return image.Point{}, false
	}
	return image.Pt(v.Focus[0]-v.Pivot[0], v.Focus[1]-v.Pivot[1]), true
}

func (v variant) frame(name string, tick uint64) image.Rectangle {
	c, ok := v.Clips[name]
	if !ok {
		base, _, _ := strings.Cut(name, "_")
		c, ok = v.Clips[base]
	}
	if !ok {
		c = v.Clips["idle"]
	}
	return frameRect(c.Frames[(tick/c.FrameTicks)%uint64(len(c.Frames))])
}

func frameRect(r []int) image.Rectangle { return image.Rect(r[0], r[1], r[0]+r[2], r[1]+r[3]) }
