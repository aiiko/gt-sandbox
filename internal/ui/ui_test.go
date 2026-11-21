package ui

import (
	"slices"
	"testing"
)

func TestWrap(t *testing.T) {
	got := Wrap([]string{"abcdefghijklmn", "", "corto"}, 8)
	want := []string{"abcdefgh", "  ijklmn", "", "corto"}
	if !slices.Equal(got, want) {
		t.Fatalf("%q", got)
	}
}
