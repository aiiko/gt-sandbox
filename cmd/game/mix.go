package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"time"

	"gtsandbox/internal/sim/clock"
	"gtsandbox/internal/sound"
)

// mixManifest es el recorrido de la mezcla de referencia: qué escenario,
// cómo se mira en cada momento y el resumen de lo que sonó.
const mixManifest = "docs/sonido/mezcla.json"

// tourStep cambia la vista a partir de un tick del escenario: "macro" vuelve
// al encuadre general, Micro abre la vista micro de ese Space y Follow sigue
// a esa entidad.
type tourStep struct {
	Tick   clock.Tick `json:"tick"`
	View   string     `json:"view,omitempty"`
	Micro  string     `json:"micro,omitempty"`
	Follow string     `json:"follow,omitempty"`
}

type mixSpec struct {
	Scenario string         `json:"scenario"`
	Seed     uint64         `json:"seed"`
	Size     string         `json:"size"`
	Ticks    clock.Tick     `json:"ticks"`
	Tour     []tourStep     `json:"tour"`
	Frames   int64          `json:"frames,omitempty"`
	Digest   string         `json:"digest,omitempty"`
	Counts   map[string]int `json:"counts,omitempty"`
}

func readMix(path string) (*mixSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m mixSpec
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &m, nil
}

// offlineMix es un Backend que mezcla en memoria, frame a frame, lo que el
// juego haría sonar. Guarda además cada pedido para el resumen.
type offlineMix struct {
	rate    int
	render  bool
	pcm     map[string][]float32
	out     []float32 // estéreo intercalado
	voices  []voice
	loops   map[string]*loopVoice
	frame   int64
	plays   []string // "frame id volumen panorama"
	levels  []string // niveles de los bucles cada segundo
	counts  map[string]int
	loopVol map[string]float64
	// continuidad del ambiente
	lastVol  map[string]float64
	maxJump  float64 // mayor cambio de volumen de un bucle entre dos frames
	silences int64   // frames sin ningún bucle sonando, pasado el primer fundido
}

type voice struct {
	pcm       []float32
	pos       int
	vol, l, r float32
}

type loopVoice struct {
	pcm []float32
	pos int
	vol float32
}

func newOfflineMix(lib *sound.Library, render bool) (*offlineMix, error) {
	m := &offlineMix{rate: lib.Rate, render: render, pcm: map[string][]float32{}, loops: map[string]*loopVoice{},
		counts: map[string]int{}, loopVol: map[string]float64{}, lastVol: map[string]float64{}}
	for _, d := range lib.Sounds {
		pcm, err := readMono(lib.Path(d))
		if err != nil {
			return nil, err
		}
		m.pcm[d.ID] = pcm
	}
	return m, nil
}

// readMono lee un WAV PCM mono de 16 bits como lo exporta tools/audio.
func readMono(path string) ([]float32, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 44 || string(data[:4]) != "RIFF" {
		return nil, fmt.Errorf("%s no es WAV", path)
	}
	samples := data[44:]
	out := make([]float32, len(samples)/2)
	for i := range out {
		out[i] = float32(int16(binary.LittleEndian.Uint16(samples[2*i:]))) / 32768
	}
	return out, nil
}

func (m *offlineMix) Play(id string, volume, pan float64) {
	m.plays = append(m.plays, fmt.Sprintf("%d %s %.3f %.3f", m.frame, id, volume, pan))
	m.counts[id]++
	left, right := min(1, 1-pan), min(1, 1+pan)
	m.voices = append(m.voices, voice{pcm: m.pcm[id], vol: float32(volume), l: float32(left), r: float32(right)})
}

func (m *offlineMix) Loop(id string, volume float64) {
	m.loopVol[id] = volume
	l, ok := m.loops[id]
	if !ok {
		l = &loopVoice{pcm: m.pcm[id]}
		m.loops[id] = l
	}
	l.vol = float32(volume)
}

func (m *offlineMix) StopAll() { m.voices = nil }

// next mezcla el frame actual (1/60 s) y pasa al siguiente.
func (m *offlineMix) next() {
	if m.frame%sound.FPS == 0 {
		line := fmt.Sprintf("%d", m.frame)
		for _, id := range sound.Ambients {
			line += fmt.Sprintf(" %.3f", m.loopVol[id])
		}
		m.levels = append(m.levels, line)
	}
	total := 0.0
	for _, id := range sound.Ambients {
		v := m.loopVol[id]
		if m.frame > 0 { // el primer frame junta la observación inicial
			m.maxJump = max(m.maxJump, math.Abs(v-m.lastVol[id]))
		}
		m.lastVol[id] = v
		total += v
	}
	if total == 0 && m.frame > sound.FadeFrames {
		m.silences++
	}
	m.frame++
	if !m.render {
		return
	}
	n := m.rate / sound.FPS
	base := len(m.out)
	m.out = append(m.out, make([]float32, 2*n)...)
	buf := m.out[base:]
	for _, id := range sound.Ambients {
		l := m.loops[id]
		if l == nil || l.vol <= 0 || len(l.pcm) == 0 {
			if l != nil && len(l.pcm) > 0 {
				l.pos = (l.pos + n) % len(l.pcm) // el bucle sigue corriendo en silencio
			}
			continue
		}
		for i := range n {
			s := l.pcm[l.pos] * l.vol
			buf[2*i] += s
			buf[2*i+1] += s
			l.pos = (l.pos + 1) % len(l.pcm)
		}
	}
	alive := m.voices[:0]
	for _, v := range m.voices {
		for i := 0; i < n && v.pos < len(v.pcm); i++ {
			s := v.pcm[v.pos] * v.vol
			buf[2*i] += s * v.l
			buf[2*i+1] += s * v.r
			v.pos++
		}
		if v.pos < len(v.pcm) {
			alive = append(alive, v)
		}
	}
	m.voices = alive
}

// digest resume la mezcla: cada pedido y los niveles del ambiente por
// segundo. No depende de los samples, así que cambiar cómo suena un sonido
// no lo cambia; cambiar qué suena y cuándo, sí.
func (m *offlineMix) digest() string {
	h := sha256.New()
	for _, line := range slices.Concat(m.plays, []string{"--"}, m.levels) {
		h.Write([]byte(line + "\n"))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// writeWAV guarda la mezcla como WAV estéreo de 16 bits, con un limitador
// suave para que nada recorte.
func (m *offlineMix) writeWAV(path string) error {
	var buf bytes.Buffer
	n := len(m.out)
	header := []any{
		[4]byte{'R', 'I', 'F', 'F'}, uint32(36 + 2*n), [4]byte{'W', 'A', 'V', 'E'},
		[4]byte{'f', 'm', 't', ' '}, uint32(16), uint16(1), uint16(2), uint32(m.rate), uint32(m.rate * 4), uint16(4), uint16(16),
		[4]byte{'d', 'a', 't', 'a'}, uint32(2 * n),
	}
	for _, v := range header {
		if err := binary.Write(&buf, binary.LittleEndian, v); err != nil {
			return err
		}
	}
	for _, s := range m.out {
		v := math.Tanh(float64(s))
		binary.Write(&buf, binary.LittleEndian, int16(v*32767))
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// tourView aplica un paso del recorrido.
func (g *game) tourView(s tourStep) {
	switch {
	case s.View == "macro":
		g.stopFollowing()
		if g.art != nil && !g.art.focus.IsNil() {
			g.toggleArtMicro()
		}
		g.sel = 0
		g.fit()
	case s.Micro != "":
		g.stopFollowing()
		if g.art != nil && !g.art.focus.IsNil() {
			g.toggleArtMicro()
		}
		for _, sp := range g.run.Cur.Spaces {
			if sp.Name == s.Micro {
				g.sel = sp.ID
				g.toggleMicro()
			}
		}
	case s.Follow != "":
		g.stopFollowing()
		g.sel = 0
		g.startFollowing(s.Follow)
	}
}

// renderMix arma la mezcla de referencia sin pantalla y la guarda como WAV.
func renderMix(path string) error {
	spec, err := readMix(mixManifest)
	if err != nil {
		return err
	}
	g, err := newGame(spec.Scenario, 0, 0, "")
	if err != nil {
		return err
	}
	if err := g.enableArt(""); err != nil {
		return err
	}
	var w, h int
	if _, err := fmt.Sscanf(spec.Size, "%dx%d", &w, &h); err != nil {
		return fmt.Errorf("%s: size %q", mixManifest, spec.Size)
	}
	g.w, g.h = (pixelGame{g}).Layout(w, h)
	g.fit()
	lib, err := sound.Load(audioManifest)
	if err != nil {
		return err
	}
	out, err := runMix(g, spec, lib, true)
	if err != nil {
		return err
	}
	if d := out.digest(); d != spec.Digest {
		fmt.Fprintf(os.Stderr, "aviso: el resumen de la mezcla es %s y el registrado %s\n", d, spec.Digest)
	}
	if err := out.writeWAV(path); err != nil {
		return err
	}
	fmt.Printf("mezcla de %.1f s escrita en %s (resumen %s)\n", float64(out.frame)/sound.FPS, path, out.digest())
	return nil
}

// runMix recorre el escenario a x1, un frame de 1/60 s por vez, con el
// sonido mezclado sin pantalla.
func runMix(g *game, spec *mixSpec, lib *sound.Library, render bool) (*offlineMix, error) {
	out, err := newOfflineMix(lib, render)
	if err != nil {
		return nil, err
	}
	g.sound = sound.NewMixer(lib, out)
	start := g.run.World.Tick()
	next := 0
	g.observeSound()
	for g.run.World.Tick() < start+spec.Ticks {
		for next < len(spec.Tour) && g.run.World.Tick() >= start+spec.Tour[next].Tick {
			g.tourView(spec.Tour[next])
			next++
		}
		g.run.Advance(time.Second / sound.FPS)
		if g.run.Err != nil {
			return nil, g.run.Err
		}
		g.observeSound()
		out.next()
	}
	return out, nil
}
