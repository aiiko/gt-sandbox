package scenario

import (
	"fmt"
	"os"
	"strings"
	"sync"

	contentfs "gtsandbox/content"
	"gtsandbox/internal/core/rng"
)

var (
	defaultNamesOnce sync.Once
	defaultNamesList []string
)

// ParseNames lee un nombre por línea (ignora líneas vacías, comentarios con #
// y duplicados, conservando el orden).
func ParseNames(text string) []string {
	seen := map[string]bool{}
	var out []string
	for _, line := range strings.Split(text, "\n") {
		name := strings.TrimSpace(line)
		if name == "" || strings.HasPrefix(name, "#") || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

// LoadNamesFile carga una lista de nombres desde un archivo de texto.
func LoadNamesFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseNames(string(data)), nil
}

// DefaultTinyNames devuelve la lista embebida de content/names.txt.
func DefaultTinyNames() []string {
	defaultNamesOnce.Do(func() {
		if data, err := contentfs.FS.ReadFile("names.txt"); err == nil {
			defaultNamesList = ParseNames(string(data))
		}
	})
	return defaultNamesList
}

// PickNames elige n nombres al azar de pool usando rnd, sin repetir hasta
// agotar la lista; si n supera len(pool), vuelve a barajar agregando un
// sufijo numérico («Aiiko 2», «Aiiko 3»…) para diferenciarlos.
func PickNames(pool []string, n int, rnd *rng.Stream) []string {
	out := make([]string, n)
	if len(pool) == 0 {
		for i := range n {
			out[i] = fmt.Sprintf("Tiny %02d", i+1)
		}
		return out
	}
	k := len(pool)
	var perm []int
	for i := range n {
		if i%k == 0 {
			perm = rnd.Perm(k)
		}
		round := i / k
		base := pool[perm[i%k]]
		if round == 0 {
			out[i] = base
		} else {
			out[i] = fmt.Sprintf("%s %d", base, round+1)
		}
	}
	return out
}
