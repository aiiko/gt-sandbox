// Package archtest verifica la regla de dependencias entre capas:
//
//	core  ←  sim/*  ←  content, scenario  ←  view, camera  ←  render/ui  ←  cmd/*
//
// view y camera no importan Ebitengine: son los datos y la matemática del
// visor, testeables sin ventana. camera solo depende de core.
// Si este test falla, no lo "arregles" relajando la regla: mové el código a la
// capa correcta o registrá un ADR que justifique el cambio.
package archtest

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const module = "gtsandbox"

// rules: paquetes bajo la clave no pueden importar nada con estos prefijos.
var rules = map[string][]string{
	"internal/core": {
		module + "/internal/sim",
		module + "/internal/content",
		module + "/internal/scenario",
		module + "/internal/view",
		module + "/internal/render",
		module + "/internal/camera",
		module + "/internal/ui",
		"github.com/hajimehoshi/ebiten",
	},
	"internal/sim": {
		module + "/internal/content",
		module + "/internal/scenario",
		module + "/internal/view",
		module + "/internal/render",
		module + "/internal/camera",
		module + "/internal/ui",
		"github.com/hajimehoshi/ebiten",
	},
	"internal/content": {
		module + "/internal/scenario",
		module + "/internal/view",
		module + "/internal/render",
		module + "/internal/camera",
		module + "/internal/ui",
		"github.com/hajimehoshi/ebiten",
	},
	"internal/scenario": {
		module + "/internal/view",
		module + "/internal/render",
		module + "/internal/camera",
		module + "/internal/ui",
		"github.com/hajimehoshi/ebiten",
	},
	"internal/view": {
		module + "/internal/render",
		module + "/internal/ui",
		"github.com/hajimehoshi/ebiten",
	},
	"internal/camera": {
		module + "/internal/sim",
		module + "/internal/content",
		module + "/internal/scenario",
		module + "/internal/view",
		module + "/internal/render",
		module + "/internal/ui",
		"github.com/hajimehoshi/ebiten",
	},
}

// Además, dentro de la simulación están prohibidas fuentes de no-determinismo.
var simForbidden = []string{
	"math/rand",    // usar core/rng (D4)
	"math/rand/v2", // idem; solo core/rng puede importarlo
	"time",         // salvo clock, que solo define duraciones (D2)
}

var simForbiddenExceptions = map[string][]string{
	"internal/core/rng":  {"math/rand/v2"},
	"internal/sim/clock": {"time"},
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod no encontrado")
		}
		dir = parent
	}
}

type fileImports struct {
	pkgDir  string // relativo al root, con '/'
	file    string
	imports []string
}

func collect(t *testing.T, root string) []fileImports {
	t.Helper()
	var out []fileImports
	fset := token.NewFileSet()
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, filepath.Dir(path))
		fi := fileImports{pkgDir: filepath.ToSlash(rel), file: filepath.Base(path)}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			fi.imports = append(fi.imports, p)
		}
		out = append(out, fi)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func under(dir, prefix string) bool {
	return dir == prefix || strings.HasPrefix(dir, prefix+"/")
}

func TestLayerDependencies(t *testing.T) {
	files := collect(t, moduleRoot(t))
	if len(files) == 0 {
		t.Fatal("no se encontraron archivos Go")
	}
	for _, f := range files {
		for layer, forbidden := range rules {
			if !under(f.pkgDir, layer) {
				continue
			}
			for _, imp := range f.imports {
				for _, bad := range forbidden {
					if imp == bad || strings.HasPrefix(imp, bad+"/") {
						t.Errorf("%s/%s importa %q: prohibido en %s", f.pkgDir, f.file, imp, layer)
					}
				}
			}
		}
	}
}

func TestNoNondeterministicImportsInSim(t *testing.T) {
	for _, f := range collect(t, moduleRoot(t)) {
		if !under(f.pkgDir, "internal/sim") && !under(f.pkgDir, "internal/core") {
			continue
		}
		if strings.HasSuffix(f.file, "_test.go") {
			continue // los tests pueden usar rand/time libremente
		}
		for _, imp := range f.imports {
			for _, bad := range simForbidden {
				if imp != bad || allowed(f.pkgDir, imp) {
					continue
				}
				t.Errorf("%s/%s importa %q: fuente de no-determinismo (ver D2/D4)", f.pkgDir, f.file, imp)
			}
		}
	}
}

func allowed(pkgDir, imp string) bool {
	for _, a := range simForbiddenExceptions[pkgDir] {
		if a == imp {
			return true
		}
	}
	return false
}
