// Package artpack arma un atlas de presentación a partir de fuentes por pieza.
//
// La carpeta de fuentes contiene un `atlas.json` y los PNG de cada cuadro. El
// manifiesto de fuentes tiene la misma forma que el que lee render.LoadAtlas,
// salvo que cada cuadro y cada primer plano nombra archivos en lugar de
// rectángulos. Un cuadro puede ser una lista de capas del mismo tamaño que se
// componen de abajo hacia arriba. El resultado es determinista: las mismas
// fuentes producen los mismos bytes de PNG y JSON. `density` (píxeles por
// metro de cada escala) pasa tal cual al manifiesto exportado.
package artpack

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/fs"
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"
)

// Gap es la separación en píxeles entre cuadros del atlas.
const Gap = 1

type sourceManifest struct {
	Image   string                              `json:"image"`
	Width   int                                 `json:"width"`
	Scales  []string                            `json:"scales,omitempty"`
	Density map[string]int                      `json:"density,omitempty"`
	Sprites map[string]map[string]sourceVariant `json:"sprites"`
}

type sourceVariant struct {
	Pivot        []int                 `json:"pivot,omitempty"`
	Layer        string                `json:"layer,omitempty"`
	DepthOffset  int                   `json:"depth_offset,omitempty"`
	Foreground   layers                `json:"foreground,omitempty"`
	Focus        []int                 `json:"focus,omitempty"`
	Carry        []int                 `json:"carry,omitempty"`
	Hidden       bool                  `json:"hidden,omitempty"`
	SelectedOnly bool                  `json:"selected_only,omitempty"`
	Tile         bool                  `json:"tile,omitempty"`
	Clips        map[string]sourceClip `json:"clips,omitempty"`
}

type sourceClip struct {
	FrameTicks uint64   `json:"frame_ticks"`
	Frames     []layers `json:"frames"`
}

// layers acepta un archivo o una lista de capas.
type layers []string

func (l *layers) UnmarshalJSON(b []byte) error {
	var one string
	if err := json.Unmarshal(b, &one); err == nil {
		*l = layers{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(b, &many); err != nil {
		return fmt.Errorf("se esperaba un archivo o una lista de capas")
	}
	if len(many) == 0 {
		return fmt.Errorf("lista de capas vacía")
	}
	*l = many
	return nil
}

type outManifest struct {
	Image   string                           `json:"image"`
	Scales  []string                         `json:"scales,omitempty"`
	Density map[string]int                   `json:"density,omitempty"`
	Sprites map[string]map[string]outVariant `json:"sprites"`
}

type outVariant struct {
	Pivot        []int              `json:"pivot,omitempty"`
	Layer        string             `json:"layer,omitempty"`
	DepthOffset  int                `json:"depth_offset,omitempty"`
	Foreground   []int              `json:"foreground,omitempty"`
	Focus        []int              `json:"focus,omitempty"`
	Carry        []int              `json:"carry,omitempty"`
	Hidden       bool               `json:"hidden,omitempty"`
	SelectedOnly bool               `json:"selected_only,omitempty"`
	Tile         bool               `json:"tile,omitempty"`
	Clips        map[string]outClip `json:"clips,omitempty"`
}

type outClip struct {
	FrameTicks uint64  `json:"frame_ticks"`
	Frames     [][]int `json:"frames"`
}

// Result es el atlas armado.
type Result struct {
	Image    string // nombre del PNG declarado en el manifiesto
	PNG      []byte
	Manifest []byte
}

type piece struct {
	key string // capas unidas, para ordenar y reutilizar
	img *image.NRGBA
	sum [32]byte
	at  image.Point
}

// Pack lee `atlas.json` en la raíz de src y arma el atlas.
func Pack(src fs.FS) (*Result, error) {
	f, err := src.Open("atlas.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var m sourceManifest
	d := json.NewDecoder(f)
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return nil, fmt.Errorf("atlas.json: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("atlas.json: contenido extra después del manifiesto")
	}
	if m.Image == "" || strings.ContainsAny(m.Image, `/\`) || path.Ext(m.Image) != ".png" {
		return nil, fmt.Errorf("atlas.json: nombre de imagen inválido %q", m.Image)
	}
	if m.Width <= 0 {
		return nil, fmt.Errorf("atlas.json: falta el ancho del atlas")
	}
	for _, scale := range slices.Sorted(maps.Keys(m.Density)) {
		if m.Density[scale] <= 0 {
			return nil, fmt.Errorf("atlas.json: densidad inválida para %s: %d", scale, m.Density[scale])
		}
	}

	pieces := map[string]*piece{}
	load := func(l layers, context string) (*piece, error) {
		key := strings.Join(l, "+")
		if p, ok := pieces[key]; ok {
			return p, nil
		}
		var out *image.NRGBA
		for _, name := range l {
			img, err := readPNG(src, name)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", context, err)
			}
			if out == nil {
				out = image.NewNRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
			} else if img.Bounds().Size() != out.Bounds().Size() {
				return nil, fmt.Errorf("%s: la capa %s mide %v y la base %v", context, name, img.Bounds().Size(), out.Bounds().Size())
			}
			draw.Draw(out, out.Bounds(), img, img.Bounds().Min, draw.Over)
		}
		if out.Bounds().Dx() > m.Width {
			return nil, fmt.Errorf("%s: %s es más ancho que el atlas", context, key)
		}
		p := &piece{key: key, img: out, sum: sha256.Sum256(out.Pix)}
		pieces[key] = p
		return p, nil
	}

	type ref struct {
		sprite, scale, clip string
		frame               int // -1 para el primer plano
		p                   *piece
	}
	var refs []ref
	for _, id := range slices.Sorted(maps.Keys(m.Sprites)) {
		for _, scale := range slices.Sorted(maps.Keys(m.Sprites[id])) {
			v := m.Sprites[id][scale]
			context := fmt.Sprintf("sprite %q, variante %s", id, scale)
			if v.Foreground != nil {
				p, err := load(v.Foreground, context+", primer plano")
				if err != nil {
					return nil, err
				}
				refs = append(refs, ref{id, scale, "", -1, p})
			}
			for _, name := range slices.Sorted(maps.Keys(v.Clips)) {
				for i, l := range v.Clips[name].Frames {
					p, err := load(l, fmt.Sprintf("%s, clip %q, cuadro %d", context, name, i+1))
					if err != nil {
						return nil, err
					}
					refs = append(refs, ref{id, scale, name, i, p})
				}
			}
		}
	}

	// Cuadros con los mismos píxeles comparten lugar en el atlas.
	type identity struct {
		sum  [32]byte
		size image.Point
	}
	id := func(p *piece) identity { return identity{p.sum, p.img.Bounds().Size()} }
	unique := map[identity]*piece{}
	var placed []*piece
	for _, key := range slices.Sorted(maps.Keys(pieces)) {
		p := pieces[key]
		if _, ok := unique[id(p)]; ok {
			continue
		}
		unique[id(p)] = p
		placed = append(placed, p)
	}
	slices.SortStableFunc(placed, func(a, b *piece) int {
		if c := b.img.Bounds().Dy() - a.img.Bounds().Dy(); c != 0 {
			return c
		}
		if c := b.img.Bounds().Dx() - a.img.Bounds().Dx(); c != 0 {
			return c
		}
		return strings.Compare(a.key, b.key)
	})
	x, y, shelf := 0, 0, 0
	for _, p := range placed {
		w, h := p.img.Bounds().Dx(), p.img.Bounds().Dy()
		if x > 0 && x+w > m.Width {
			x, y, shelf = 0, y+shelf+Gap, 0
		}
		p.at = image.Pt(x, y)
		x += w + Gap
		shelf = max(shelf, h)
	}
	atlas := image.NewNRGBA(image.Rect(0, 0, m.Width, y+shelf))
	for _, p := range placed {
		draw.Draw(atlas, p.img.Bounds().Add(p.at), p.img, image.Point{}, draw.Src)
	}
	rect := func(p *piece) []int {
		q := unique[id(p)]
		return []int{q.at.X, q.at.Y, p.img.Bounds().Dx(), p.img.Bounds().Dy()}
	}

	out := outManifest{Image: m.Image, Scales: m.Scales, Density: m.Density, Sprites: map[string]map[string]outVariant{}}
	for name, variants := range m.Sprites {
		out.Sprites[name] = map[string]outVariant{}
		for scale, v := range variants {
			o := outVariant{
				Pivot: v.Pivot, Layer: v.Layer, DepthOffset: v.DepthOffset, Focus: v.Focus,
				Carry: v.Carry, Hidden: v.Hidden, SelectedOnly: v.SelectedOnly, Tile: v.Tile,
			}
			if len(v.Clips) > 0 {
				o.Clips = map[string]outClip{}
				for clip, c := range v.Clips {
					o.Clips[clip] = outClip{FrameTicks: c.FrameTicks, Frames: make([][]int, len(c.Frames))}
				}
			}
			out.Sprites[name][scale] = o
		}
	}
	for _, r := range refs {
		o := out.Sprites[r.sprite][r.scale]
		if r.frame < 0 {
			o.Foreground = rect(r.p)
		} else {
			o.Clips[r.clip].Frames[r.frame] = rect(r.p)
		}
		out.Sprites[r.sprite][r.scale] = o
	}

	var img bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&img, atlas); err != nil {
		return nil, err
	}
	js, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	return &Result{Image: m.Image, PNG: img.Bytes(), Manifest: append(compactNumbers(js), '\n')}, nil
}

func readPNG(src fs.FS, name string) (image.Image, error) {
	if !fs.ValidPath(name) || path.Ext(name) != ".png" {
		return nil, fmt.Errorf("ruta de capa inválida %q", name)
	}
	f, err := src.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return img, nil
}

var numberArray = regexp.MustCompile(`\[\s*(-?\d+(?:,\s*-?\d+)*)\s*\]`)
var space = regexp.MustCompile(`,\s+`)

// compactNumbers deja en una línea las listas de números (pivotes y cuadros).
func compactNumbers(js []byte) []byte {
	return numberArray.ReplaceAllFunc(js, func(b []byte) []byte {
		inner := numberArray.FindSubmatch(b)[1]
		return append(append([]byte("["), space.ReplaceAll(inner, []byte(", "))...), ']')
	})
}
