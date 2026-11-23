package sound

import (
	"gtsandbox/internal/core/ids"
	"gtsandbox/internal/view"
)

// Sonidos de la interfaz.
const (
	UISelect  = "ui_select"   // elegir algo
	UIFollow  = "ui_follow"   // empezar a seguir
	UIZoomIn  = "ui_zoom_in"  // pasar a micro
	UIZoomOut = "ui_zoom_out" // volver a macro
	UIPause   = "ui_pause"    // pausar
	UIResume  = "ui_resume"   // reanudar
	UIStory   = "ui_story"    // entrada nueva en la historia visible
)

// UIState es lo que la interfaz muestra en un frame. Los sonidos de interfaz
// salen de comparar dos estados seguidos, así suenan igual sin importar si el
// cambio vino del teclado, el mouse o la historia.
type UIState struct {
	Selected  ids.ID
	Following ids.ID
	Micro     bool
	Paused    bool
	Story     bool            // el panel de historia está visible
	Last      view.StoryEntry // última entrada de la historia
}

// UISounds devuelve los sonidos que corresponden al pasar de prev a cur, en
// orden estable.
func UISounds(prev, cur UIState) []string {
	var out []string
	switch {
	case cur.Following != prev.Following && !cur.Following.IsNil():
		out = append(out, UIFollow)
	case cur.Selected != prev.Selected && !cur.Selected.IsNil():
		out = append(out, UISelect)
	}
	if cur.Micro != prev.Micro {
		if cur.Micro {
			out = append(out, UIZoomIn)
		} else {
			out = append(out, UIZoomOut)
		}
	}
	if cur.Paused != prev.Paused {
		if cur.Paused {
			out = append(out, UIPause)
		} else {
			out = append(out, UIResume)
		}
	}
	if cur.Story && cur.Last.Seq != prev.Last.Seq && cur.Last.Text != "" {
		out = append(out, UIStory)
	}
	return out
}

// UIWatcher sigue el estado de la interfaz entre frames y hace sonar sus
// cambios. El primer estado observado solo se registra.
type UIWatcher struct {
	prev UIState
	seen bool
}

// Observe compara con el frame anterior y pide los sonidos al mezclador.
func (i *UIWatcher) Observe(m *Mixer, cur UIState) {
	if i.seen && m != nil {
		for _, id := range UISounds(i.prev, cur) {
			m.Play(id, 1, 0)
		}
	}
	i.prev, i.seen = cur, true
}
