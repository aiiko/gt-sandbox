package sound

import (
	"maps"
	"slices"
)

// FPS es la cadencia de Advance: el juego llama una vez por Update.
const FPS = 60

// Mixer aplica las reglas de mezcla antes de pasar un sonido al Backend:
// silencio, volumen general y por categoría, voces simultáneas y tiempo
// mínimo entre comienzos. Cuenta el tiempo en frames de presentación, nunca
// con el reloj real, así que es reproducible.
type Mixer struct {
	Lib     *Library
	Out     Backend
	Master  float64
	Volumes map[Category]float64
	// Log guarda los últimos pedidos aceptados (para pruebas y depuración).
	Log []Played

	muted  bool
	frame  int64
	loops  map[string]float64 // volumen final de cada bucle
	last   map[string]int64   // último comienzo de cada sonido
	voices map[string][]int64 // fin de cada voz activa
}

// DefaultVolumes son los volúmenes por categoría al arrancar.
var DefaultVolumes = map[Category]float64{Ambient: 0.5, Effects: 0.8, Interface: 0.6}

// LogSize es cuántos pedidos recientes guarda Log.
const LogSize = 64

// NewMixer arma un mezclador con los volúmenes por defecto.
func NewMixer(lib *Library, out Backend) *Mixer {
	m := &Mixer{Lib: lib, Out: out, Master: 1, Volumes: map[Category]float64{},
		last: map[string]int64{}, voices: map[string][]int64{}, loops: map[string]float64{}}
	for c, v := range DefaultVolumes {
		m.Volumes[c] = v
	}
	return m
}

// Frame es el frame de presentación actual.
func (m *Mixer) Frame() int64 { return m.frame }

// Advance pasa al frame siguiente.
func (m *Mixer) Advance() { m.frame++ }

// Muted dice si el sonido está silenciado.
func (m *Mixer) Muted() bool { return m.muted }

// SetMuted silencia (cortando lo que suena) o vuelve a habilitar el sonido.
func (m *Mixer) SetMuted(muted bool) {
	if muted && !m.muted && m.Out != nil {
		m.Out.StopAll()
	}
	m.muted = muted
	if muted {
		clear(m.voices)
	}
}

func frames(ms int) int64 { return int64(ms) * FPS / 1000 }

// Loop fija el nivel (0..1) de un bucle de ambiente; el mezclador le aplica
// el volumen general, el de su categoría y el silencio.
func (m *Mixer) Loop(id string, level float64) {
	if m.Lib == nil {
		return
	}
	d, ok := m.Lib.Def(id)
	if !ok {
		return
	}
	v := level * m.Master * m.Volumes[d.Category]
	if m.muted {
		v = 0
	}
	if old, ok := m.loops[id]; ok && old == v {
		return
	}
	m.loops[id] = v
	if m.Out != nil {
		m.Out.Loop(id, v)
	}
}

// Hush corta lo que suena y apaga los bucles de ambiente, por ejemplo al
// dejar una partida para volver al menú. El silencio no cambia.
func (m *Mixer) Hush() {
	for _, id := range slices.Sorted(maps.Keys(m.loops)) {
		if m.loops[id] != 0 {
			m.loops[id] = 0
			if m.Out != nil {
				m.Out.Loop(id, 0)
			}
		}
	}
	if m.Out != nil {
		m.Out.StopAll()
	}
	clear(m.voices)
}

// LoopVolume es el volumen final de un bucle.
func (m *Mixer) LoopVolume(id string) float64 { return m.loops[id] }

// Play pide un sonido con un volumen relativo (0..1) y un panorama (-1
// izquierda, 0 centro, 1 derecha). Devuelve si sonó.
func (m *Mixer) Play(id string, volume, pan float64) bool {
	if m.muted || m.Lib == nil {
		return false
	}
	d, ok := m.Lib.Def(id)
	if !ok {
		return false
	}
	if last, ok := m.last[id]; ok && m.frame-last < frames(d.CooldownMS) {
		return false
	}
	active := m.voices[id][:0]
	for _, end := range m.voices[id] {
		if end > m.frame {
			active = append(active, end)
		}
	}
	m.voices[id] = active
	if len(active) >= d.Voices {
		return false
	}
	v := volume * m.Master * m.Volumes[d.Category]
	if v <= 0 {
		return false
	}
	pan = max(-1, min(1, pan))
	m.last[id] = m.frame
	m.voices[id] = append(active, m.frame+max(1, frames(d.MS)))
	if m.Out != nil {
		m.Out.Play(id, v, pan)
	}
	m.Log = append(m.Log, Played{Frame: m.frame, ID: id, Volume: v, Pan: pan})
	if over := len(m.Log) - LogSize; over > 0 {
		m.Log = append(m.Log[:0], m.Log[over:]...)
	}
	return true
}
