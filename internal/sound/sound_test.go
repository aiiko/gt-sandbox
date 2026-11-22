package sound

import (
	"encoding/binary"
	"os"
	"slices"
	"testing"

	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/attach"
	"gtsandbox/internal/sim/space"
	"gtsandbox/internal/view"
)

const manifest = "../../content/audio/sounds.json"

func library(t *testing.T) *Library {
	t.Helper()
	lib, err := Load(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return lib
}

// El manifiesto exportado por tools/audio coincide con sus WAV: existen, son
// PCM mono de 16 bits a la frecuencia del manifiesto y duran lo declarado.
// Todos los sonidos que pide la interfaz están.
func TestManifestMatchesFiles(t *testing.T) {
	lib := library(t)
	for _, d := range lib.Sounds {
		data, err := os.ReadFile(lib.Path(d))
		if err != nil {
			t.Fatal(err)
		}
		if len(data) < 44 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
			t.Fatalf("%s no es WAV", d.File)
		}
		channels := binary.LittleEndian.Uint16(data[22:24])
		rate := binary.LittleEndian.Uint32(data[24:28])
		bits := binary.LittleEndian.Uint16(data[34:36])
		if channels != 1 || int(rate) != lib.Rate || bits != 16 {
			t.Errorf("%s: %d canales, %d Hz, %d bits", d.File, channels, rate, bits)
		}
		samples := (len(data) - 44) / 2
		if ms := samples * 1000 / lib.Rate; abs(ms-d.MS) > 1 {
			t.Errorf("%s dura %d ms, el manifiesto dice %d", d.File, ms, d.MS)
		}
	}
	for _, id := range []string{UISelect, UIFollow, UIZoomIn, UIZoomOut, UIPause, UIResume, UIStory} {
		if d, ok := lib.Def(id); !ok || d.Category != Interface {
			t.Errorf("falta el sonido de interfaz %s", id)
		}
	}
	scene := []string{StepTatami, StepWood, StepTile, DoorFusuma, DoorNoren, FridgeOpen, FridgeClose,
		CupClink, Bite, Zabuton, TVOn, HandNear, HandLand, HandLift, HandTap, CrumbPop, CrumbTick,
		NestStore, TinyHop, TinyFall, TinyLand, PassOpen, PassClose}
	for _, step := range tinySteps {
		scene = append(scene, step)
	}
	for _, e := range []string{view.EmoteDiscover, view.EmoteExplore, view.EmoteSweat, view.EmoteSleep, view.EmoteNote, view.EmoteVein} {
		scene = append(scene, Emote+e)
	}
	for _, id := range scene {
		if d, ok := lib.Def(id); !ok || d.Category != Effects {
			t.Errorf("falta el efecto %s", id)
		}
	}
	for id := range minor {
		if _, ok := lib.Def(id); !ok {
			t.Errorf("el sonido menor %s no existe", id)
		}
	}
}

// La primera observación solo registra: los eventos anteriores no suenan, y
// un frame sin tick nuevo tampoco.
func TestSceneStartsSilent(t *testing.T) {
	everywhere := View{
		Micro:      true,
		Locate:     func(units.Vec3) (float64, bool) { return 0, true },
		Visual:     func(ids.ID) string { return "" },
		ShowsEmote: func(ids.ID) bool { return true },
	}
	hand := ids.Make(2, 1)
	f0 := &view.Frame{Tick: 10, Spaces: []view.Space{{ID: hand, Kind: space.Articulated}}}
	old := []view.Published{{Seq: 1, Tick: 9, Event: attach.ContactBegan{A: ids.Make(1, 1), B: hand}}}
	var s Scene
	if got := s.Sounds(f0, old, everywhere); got != nil {
		t.Fatalf("la primera observación sonó %v", got)
	}
	everywhere.Focus = ids.Make(1, 1)
	again := append(old, view.Published{Seq: 2, Tick: 10, Event: attach.ContactBegan{A: ids.Make(1, 1), B: hand}})
	f1 := &view.Frame{Tick: 11, Spaces: f0.Spaces}
	got := s.Sounds(f1, again, everywhere)
	if len(got) != 1 || got[0].ID != HandLand {
		t.Fatalf("solo el evento nuevo suena: %v", got)
	}
	if got := s.Sounds(f1, again, everywhere); got != nil {
		t.Fatalf("el mismo tick no vuelve a sonar: %v", got)
	}
}

func abs(v int) int { return max(v, -v) }

type recorder struct {
	played  []string
	stopped int
	loops   map[string]float64
}

func (r *recorder) Play(id string, _, _ float64) { r.played = append(r.played, id) }
func (r *recorder) StopAll()                     { r.stopped++ }
func (r *recorder) Loop(id string, v float64) {
	if r.loops == nil {
		r.loops = map[string]float64{}
	}
	r.loops[id] = v
}

func testLibrary() *Library {
	lib := &Library{byID: map[string]Def{}}
	for _, d := range []Def{
		{ID: "a", Category: Effects, MS: 500, Voices: 2},
		{ID: "b", Category: Interface, MS: 100, Voices: 1, CooldownMS: 200},
	} {
		lib.Sounds = append(lib.Sounds, d)
		lib.byID[d.ID] = d
	}
	return lib
}

// Las voces simultáneas y el tiempo entre comienzos limitan los pedidos; el
// silencio corta lo que suena y rechaza todo.
func TestMixerRules(t *testing.T) {
	r := &recorder{}
	m := NewMixer(testLibrary(), r)
	for range 3 {
		m.Play("a", 1, 0)
	}
	if len(r.played) != 2 {
		t.Fatalf("con 2 voces sonaron %d", len(r.played))
	}
	for range frames(500) {
		m.Advance()
	}
	if !m.Play("a", 1, 0) {
		t.Fatal("al terminar una voz debería volver a sonar")
	}
	if !m.Play("b", 1, 0) || m.Play("b", 1, 0) {
		t.Fatal("b tiene una voz")
	}
	for range frames(150) {
		m.Advance()
	}
	if m.Play("b", 1, 0) {
		t.Fatal("b no debería repetirse antes de 200 ms")
	}
	m.SetMuted(true)
	if r.stopped != 1 || m.Play("a", 1, 0) {
		t.Fatal("el silencio corta y rechaza")
	}
	m.SetMuted(false)
	if !m.Play("a", 1, 0) {
		t.Fatal("al quitar el silencio vuelve a sonar")
	}
	m.Volumes[Effects] = 0
	for range frames(1000) {
		m.Advance()
	}
	if m.Play("a", 1, 0) {
		t.Fatal("con volumen 0 no se pide nada")
	}
	if got := m.Log[len(m.Log)-1]; got.ID != "a" || got.Volume != DefaultVolumes[Effects] {
		t.Errorf("registro %+v", got)
	}
}

// Cada cambio de la interfaz tiene su sonido; empezar a seguir suena como
// seguimiento y no además como selección.
func TestUISounds(t *testing.T) {
	a, b := ids.Make(1, 1), ids.Make(2, 1)
	entry := view.StoryEntry{Tick: 10, Text: "giantess apoya la mano."}
	for _, c := range []struct {
		name      string
		prev, cur UIState
		want      []string
	}{
		{"nada", UIState{}, UIState{}, nil},
		{"elegir", UIState{}, UIState{Selected: a}, []string{UISelect}},
		{"deselegir", UIState{Selected: a}, UIState{}, nil},
		{"seguir", UIState{Selected: a}, UIState{Selected: b, Following: b}, []string{UIFollow}},
		{"micro", UIState{}, UIState{Micro: true}, []string{UIZoomIn}},
		{"macro y pausa", UIState{Micro: true}, UIState{Paused: true}, []string{UIZoomOut, UIPause}},
		{"reanudar", UIState{Paused: true}, UIState{}, []string{UIResume}},
		{"historia oculta", UIState{}, UIState{Last: entry}, nil},
		{"historia", UIState{Story: true}, UIState{Story: true, Last: entry}, []string{UIStory}},
	} {
		if got := UISounds(c.prev, c.cur); !slices.Equal(got, c.want) {
			t.Errorf("%s: %v, quería %v", c.name, got, c.want)
		}
	}
	r := &recorder{}
	m := NewMixer(library(t), r)
	var i UIWatcher
	i.Observe(m, UIState{Paused: true}) // el primer estado no suena
	i.Observe(m, UIState{})
	if !slices.Equal(r.played, []string{UIResume}) {
		t.Errorf("sonó %v", r.played)
	}
}

// Hush apaga los bucles y corta lo que suena sin silenciar el mezclador.
func TestMixerHush(t *testing.T) {
	r := &recorder{}
	m := NewMixer(testLibrary(), r)
	m.Loop("a", 1)
	m.Play("a", 1, 0)
	m.Play("a", 1, 0)
	m.Hush()
	if r.loops["a"] != 0 || m.LoopVolume("a") != 0 || r.stopped != 1 || m.Muted() {
		t.Fatalf("bucles %v, cortes %d, silencio %v", r.loops, r.stopped, m.Muted())
	}
	if !m.Play("a", 1, 0) {
		t.Fatal("después de Hush las voces quedan libres")
	}
	m.Loop("a", 1)
	if r.loops["a"] == 0 {
		t.Fatal("el bucle vuelve a sonar al pedirlo")
	}
}
