// artpack arma el atlas de presentación a partir de sus fuentes por pieza.
//
//	go run ./cmd/artpack -src content/art/src/casa -out content/art
//	go run ./cmd/artpack -src content/art/src/casa -out content/art -check
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gtsandbox/internal/artpack"
)

func main() {
	src := flag.String("src", "content/art/src/casa", "carpeta de fuentes con atlas.json")
	out := flag.String("out", "content/art", "carpeta donde se escriben el PNG y el manifiesto")
	check := flag.Bool("check", false, "solo verificar que el atlas exportado coincide con las fuentes")
	flag.Parse()
	if err := run(*src, *out, *check); err != nil {
		fmt.Fprintln(os.Stderr, "artpack:", err)
		os.Exit(1)
	}
}

func run(src, out string, check bool) error {
	r, err := artpack.Pack(os.DirFS(src))
	if err != nil {
		return err
	}
	pngFile := filepath.Join(out, r.Image)
	jsonFile := filepath.Join(out, strings.TrimSuffix(r.Image, ".png")+".json")
	files := map[string][]byte{pngFile: r.PNG, jsonFile: r.Manifest}
	for _, name := range []string{pngFile, jsonFile} {
		if check {
			old, err := os.ReadFile(name)
			if err != nil {
				return err
			}
			if !bytes.Equal(old, files[name]) {
				return fmt.Errorf("%s no coincide con las fuentes; correr artpack sin -check", name)
			}
			continue
		}
		if err := os.WriteFile(name, files[name], 0o644); err != nil {
			return err
		}
		fmt.Println("escrito", name)
	}
	return nil
}
