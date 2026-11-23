package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCanvasLayoutFillsWindow(t *testing.T) {
	for _, tc := range []struct{ w, h, cw, ch int }{
		{960, 540, 480, 270},   // capturas de referencia: el lienzo de siempre
		{1920, 1009, 640, 337}, // maximizada en 1080p: más mundo, sin bandas
		{1031, 601, 516, 301},
	} {
		if cw, ch := canvasLayout(tc.w, tc.h); cw != tc.cw || ch != tc.ch {
			t.Errorf("%dx%d: lienzo %dx%d, quería %dx%d", tc.w, tc.h, cw, ch, tc.cw, tc.ch)
		}
	}
	defer func(s func() float64) { deviceScale = s }(deviceScale)
	deviceScale = func() float64 { return 1.5 } // 1280×720 lógicos = 1920×1080 físicos
	if cw, ch := canvasLayout(1280, 720); cw != 640 || ch != 360 {
		t.Errorf("con escala 1,5: lienzo %dx%d", cw, ch)
	}
}

func TestDefaultWindowFitsMonitor(t *testing.T) {
	for _, tc := range []struct{ mw, mh, w, h int }{
		{1920, 1080, 1440, 810},
		{2560, 1440, 1920, 1080},
		{1366, 768, 960, 540},
		{800, 600, 480, 270},
	} {
		if w, h := defaultWindow(tc.mw, tc.mh); w != tc.w || h != tc.h {
			t.Errorf("monitor %dx%d: ventana %dx%d, quería %dx%d", tc.mw, tc.mh, w, h, tc.w, tc.h)
		}
	}
}

func TestWindowConfigRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gt-sandbox", "ventana.json")
	if _, ok := loadWindow(path); ok {
		t.Fatal("sin archivo no hay ventana guardada")
	}
	want := windowConfig{Width: 1440, Height: 810, Maximized: true}
	if err := saveWindow(path, want); err != nil {
		t.Fatal(err)
	}
	if got, ok := loadWindow(path); !ok || got != want {
		t.Fatalf("ventana %+v (%v), quería %+v", got, ok, want)
	}
	// Una ventana más chica que el lienzo mínimo, o un archivo roto, se ignoran.
	if err := saveWindow(path, windowConfig{Width: 100, Height: 50}); err != nil {
		t.Fatal(err)
	}
	if _, ok := loadWindow(path); ok {
		t.Fatal("se aceptó una ventana más chica que el lienzo mínimo")
	}
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := loadWindow(path); ok {
		t.Fatal("se aceptó un archivo roto")
	}
}
