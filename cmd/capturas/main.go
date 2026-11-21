// capturas regenera las capturas de referencia de docs/capturas a partir de
// su manifiesto: compila cmd/game una vez y saca cada captura con -shot.
//
//	go run ./cmd/capturas              # todas
//	go run ./cmd/capturas -only viaje  # una
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type capture struct {
	Name     string    `json:"name"`
	Scenario string    `json:"scenario"`
	Tick     int       `json:"tick"`
	Micro    string    `json:"micro,omitempty"`
	Follow   string    `json:"follow,omitempty"`
	Pan      []float64 `json:"pan,omitempty"`
}

type manifest struct {
	Size     string    `json:"size"`
	Captures []capture `json:"captures"`
}

func main() {
	path := flag.String("manifest", "docs/capturas/capturas.json", "manifiesto de capturas")
	only := flag.String("only", "", "regenerar solo la captura con este nombre")
	flag.Parse()
	if err := run(*path, *only); err != nil {
		fmt.Fprintln(os.Stderr, "capturas:", err)
		os.Exit(1)
	}
}

func run(path, only string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "capturas")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	bin := filepath.Join(tmp, "game.exe")
	build := exec.Command("go", "build", "-o", bin, "./cmd/game")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("compilar cmd/game: %w", err)
	}
	dir := filepath.Dir(path)
	done := 0
	for _, c := range m.Captures {
		if only != "" && c.Name != only {
			continue
		}
		out := filepath.Join(dir, c.Name+".png")
		args := []string{"-art", "-scenario", c.Scenario, "-size", m.Size, "-shot", out, "-shot-tick", strconv.Itoa(c.Tick)}
		if c.Micro != "" {
			args = append(args, "-micro", c.Micro)
		}
		if c.Follow != "" {
			args = append(args, "-follow", c.Follow)
		}
		if len(c.Pan) == 2 {
			args = append(args, "-pan", fmt.Sprintf("%g,%g", c.Pan[0], c.Pan[1]))
		}
		cmd := exec.Command(bin, args...)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", c.Name, err)
		}
		fmt.Println("escrita", out)
		done++
	}
	if done == 0 {
		return fmt.Errorf("ninguna captura se llama %q", only)
	}
	return nil
}
