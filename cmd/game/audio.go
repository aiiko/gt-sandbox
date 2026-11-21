package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sound"
)

// audioManifest es el manifiesto de sonidos, relativo a la raíz del repo.
const audioManifest = "content/audio/sounds.json"

// ebitenAudio reproduce los pedidos del mezclador con el audio de Ebitengine.
// Decodifica cada WAV una vez (PCM de 16 bits estéreo) y aplica el panorama
// copiando los bytes con una ganancia por canal.
type ebitenAudio struct {
	ctx     *audio.Context
	pcm     map[string][]byte
	players []*audio.Player
}

func newEbitenAudio(lib *sound.Library) (*ebitenAudio, error) {
	a := &ebitenAudio{ctx: audio.NewContext(lib.Rate), pcm: map[string][]byte{}}
	for _, d := range lib.Sounds {
		data, err := os.ReadFile(lib.Path(d))
		if err != nil {
			return nil, err
		}
		s, err := wav.DecodeWithSampleRate(lib.Rate, bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", d.File, err)
		}
		pcm, err := io.ReadAll(s)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", d.File, err)
		}
		a.pcm[d.ID] = pcm
	}
	return a, nil
}

func (a *ebitenAudio) Play(id string, volume, pan float64) {
	pcm, ok := a.pcm[id]
	if !ok {
		return
	}
	alive := a.players[:0]
	for _, p := range a.players {
		if p.IsPlaying() {
			alive = append(alive, p)
		} else {
			p.Close()
		}
	}
	a.players = alive
	if pan != 0 {
		pcm = panned(pcm, pan)
	}
	p := a.ctx.NewPlayerFromBytes(pcm)
	p.SetVolume(volume)
	p.Play()
	a.players = append(a.players, p)
}

func (a *ebitenAudio) StopAll() {
	for _, p := range a.players {
		p.Pause()
		p.Close()
	}
	a.players = nil
}

// panned copia PCM estéreo de 16 bits bajando el canal opuesto al panorama.
func panned(pcm []byte, pan float64) []byte {
	left, right := min(1, 1-pan), min(1, 1+pan)
	out := make([]byte, len(pcm))
	for i := 0; i+3 < len(pcm); i += 4 {
		l := int16(binary.LittleEndian.Uint16(pcm[i:]))
		r := int16(binary.LittleEndian.Uint16(pcm[i+2:]))
		binary.LittleEndian.PutUint16(out[i:], uint16(int16(float64(l)*left)))
		binary.LittleEndian.PutUint16(out[i+2:], uint16(int16(float64(r)*right)))
	}
	return out
}

// enableSound carga los sonidos y abre el dispositivo de audio.
func (g *game) enableSound(manifest string, volume float64, muted bool) error {
	lib, err := sound.Load(manifest)
	if err != nil {
		return err
	}
	out, err := newEbitenAudio(lib)
	if err != nil {
		return err
	}
	g.sound = sound.NewMixer(lib, out)
	g.sound.Master = max(0, min(1, volume))
	g.sound.SetMuted(muted)
	return nil
}

// uiState resume lo que muestra la interfaz para los sonidos de interfaz.
func (g *game) uiState() sound.UIState {
	micro := !g.cam.Anchor.IsNil()
	if g.art != nil {
		micro = !g.artFocus(g.run.Cur).IsNil()
	}
	st := sound.UIState{Selected: g.sel, Following: g.follow, Micro: micro, Paused: g.run.Paused, Story: g.story}
	if n := len(g.run.Story); n > 0 {
		st.Last = g.run.Story[n-1]
	}
	return st
}

// observeSound hace sonar los cambios de la interfaz y avanza el mezclador un
// frame. No toca la simulación.
func (g *game) observeSound() {
	if g.sound == nil {
		return
	}
	g.uiSounds.Observe(g.sound, g.uiState())
	g.scene.Observe(g.sound, g.run.Cur, g.run.Events, g.soundView())
	g.sound.Advance()
}

// soundView le dice al sonido qué se ve: la escala, la velocidad, dónde cae
// cada punto en pantalla y de quién se dibuja el símbolo (como drawEmotes).
func (g *game) soundView() sound.View {
	f := g.run.Cur
	focus := g.cam.Anchor
	if g.art != nil {
		focus = g.artFocus(f)
	}
	micro := !focus.IsNil()
	proj := g.cam.Project(g.anchor(f), g.w, g.h)
	w, h := float32(g.w), float32(g.h)
	giantess := g.giantessID()
	return sound.View{
		Micro: micro,
		Focus: focus,
		Fast:  g.run.Speed > 1,
		Locate: func(p units.Vec3) (float64, bool) {
			x, y := proj.ToScreen(p)
			if w <= 0 || x < -w/10 || x > w*1.1 || y < -h/10 || y > h*1.1 {
				return 0, false
			}
			return max(-1, min(1, float64(x/w)*2-1)) * soundPanWidth, true
		},
		Visual: func(id ids.ID) string { return g.visuals[id] },
		ShowsEmote: func(id ids.ID) bool {
			if micro {
				_, ok := f.ThingByID(id)
				return ok
			}
			return !id.IsNil() && (id == giantess || id == g.sel || id == g.follow)
		},
	}
}

// soundPanWidth abre el panorama sin llevar nada del todo a un lado.
const soundPanWidth = 0.7

// toggleMute alterna el silencio; al volver suena el clic de interfaz.
func (g *game) toggleMute() {
	if g.sound == nil {
		return
	}
	g.sound.SetMuted(!g.sound.Muted())
	if !g.sound.Muted() {
		g.sound.Play(sound.UISelect, 1, 0)
	}
}
