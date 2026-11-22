package scenario

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"gtsandbox/internal/content"
	"gtsandbox/internal/core/rng"
	"gtsandbox/internal/core/units"
	"gtsandbox/internal/sim/nav"
)

// Límites de una partida al azar.
const (
	RandomMinTinies = 8
	RandomMaxTinies = 20
	// randomSpread es hasta dónde del nido se reparten los tinies.
	randomSpread = 15 * units.Centimeter
	// randomNestGap deja libre el lugar del nido.
	randomNestGap = 2 * units.Centimeter
)

// Randomize arma una partida nueva a partir de un escenario base, como un
// jugador que prepara la mesa: es un generador de inputs, fuera de la
// simulación. Toma como molde a los miembros de colonia del base (entidades
// con componente "member") y los reemplaza por entre RandomMinTinies y
// RandomMaxTinies tinies repartidos por celdas libres de la superficie donde
// estaban, cerca de su nido, con colores, necesidades y carácter sorteados.
// La semilla del mundo pasa a ser seed y se quitan las expectativas: una
// partida no tiene veredicto. Con la misma semilla, la misma partida.
func Randomize(base *Spec, seed uint64) (*Spec, error) {
	rnd := rng.NewStream(seed, "partida")
	var (
		molds []map[string]any
		rest  []EntitySpec
		gone  []string
	)
	visuals := map[string]bool{}
	for _, e := range base.Entities {
		var spawn map[string]any
		if err := json.Unmarshal(e.Spawn, &spawn); err != nil {
			return nil, fmt.Errorf("scenario: entidad %q: %w", e.Ref, err)
		}
		if comps, _ := spawn["components"].(map[string]any); comps["member"] != nil {
			molds = append(molds, spawn)
			gone = append(gone, e.Ref)
			if e.Visual != "" {
				visuals[e.Visual] = true
			}
			continue
		}
		rest = append(rest, e)
	}
	if len(molds) == 0 {
		return nil, fmt.Errorf("scenario: %q no tiene miembros de colonia para sortear", base.Name)
	}
	for _, c := range base.Commands {
		for _, ref := range gone {
			if ref != "" && strings.Contains(string(c.Data), `"$`+ref+`"`) {
				return nil, fmt.Errorf("scenario: el comando %q usa a %q, que se sortea", c.Kind, ref)
			}
		}
	}
	mold := molds[0]
	loc, _ := mold["location"].(map[string]any)
	surfaceRef, _ := loc["space"].(string)
	grid, err := entityGrid(rest, strings.TrimPrefix(surfaceRef, "$"))
	if err != nil {
		return nil, err
	}
	comps := mold["components"].(map[string]any)
	member, _ := comps["member"].(map[string]any)
	colony, _ := member["colony"].(string)
	center, ok := entityPos(rest, strings.TrimPrefix(colony, "$"), surfaceRef)
	if !ok {
		return nil, fmt.Errorf("scenario: el nido %q no está en %q", colony, surfaceRef)
	}

	// Celdas libres cerca del nido, en orden de grilla y después barajadas.
	var cells []units.Vec2
	for y := range grid.H {
		for x := range grid.W {
			p := grid.Center(x, y)
			d := p.Sub(center)
			if grid.At(x, y) == nav.Blocked || max(d.X, -d.X) > randomSpread || max(d.Y, -d.Y) > randomSpread {
				continue
			}
			if max(d.X, -d.X) < randomNestGap && max(d.Y, -d.Y) < randomNestGap {
				continue
			}
			cells = append(cells, p)
		}
	}
	n := min(RandomMinTinies+rnd.IntN(RandomMaxTinies-RandomMinTinies+1), len(cells))
	if n == 0 {
		return nil, fmt.Errorf("scenario: no hay celdas libres cerca del nido")
	}
	order := rnd.Perm(len(cells))
	colors := slices.Sorted(maps.Keys(visuals))
	body, _ := mold["body"].(map[string]any)
	radius, _ := body["radius"].(float64)
	jitter := max(0, int64(grid.Cell)/2-int64(radius)-int64(units.Millimeter))

	out := *base
	out.Seed = seed
	out.Expect = nil
	out.Entities = rest
	for i := range n {
		spawn := deepCopy(mold).(map[string]any)
		spawn["name"] = fmt.Sprintf("Tiny %02d", i+1)
		p := cells[order[i]]
		if jitter > 0 {
			p.X += units.Length(rnd.Int64N(2*jitter+1) - jitter)
			p.Y += units.Length(rnd.Int64N(2*jitter+1) - jitter)
		}
		spawn["location"] = map[string]any{"space": surfaceRef, "pos": map[string]any{"x": int64(p.X), "y": int64(p.Y), "z": 0}}
		c := spawn["components"].(map[string]any)
		if needs, ok := c["needs"].([]any); ok {
			for _, need := range needs {
				if nd, ok := need.(map[string]any); ok {
					if v, ok := nd["value"].(float64); ok {
						nd["value"] = min(1000, int64(v*(0.6+0.5*rnd.Float64())))
					}
				}
			}
		}
		if agent, ok := c["agent"].(map[string]any); ok {
			for _, trait := range []string{"curiosity", "work"} {
				if v, ok := agent[trait].(float64); ok {
					agent[trait] = round2(min(1, v*(0.6+0.7*rnd.Float64())))
				}
			}
		}
		raw, err := json.Marshal(spawn)
		if err != nil {
			return nil, err
		}
		e := EntitySpec{Ref: fmt.Sprintf("tiny%02d", i+1), Spawn: raw}
		if len(colors) > 0 {
			e.Visual = colors[rnd.IntN(len(colors))]
		}
		out.Entities = append(out.Entities, e)
	}
	return &out, nil
}

// entityGrid devuelve la grilla de navegación de la entidad ref, propia o de
// su objeto de content.
func entityGrid(entities []EntitySpec, ref string) (*nav.Grid, error) {
	for _, e := range entities {
		if e.Ref != ref {
			continue
		}
		spawn := e.Spawn
		if e.Object != "" {
			lib, err := content.Default()
			if err != nil {
				return nil, err
			}
			if spawn, err = lib.Spawn(e.Object, spawn); err != nil {
				return nil, err
			}
		}
		var s struct {
			Components struct {
				Grid *nav.Grid `json:"nav_grid"`
			} `json:"components"`
		}
		if err := json.Unmarshal(spawn, &s); err != nil {
			return nil, err
		}
		if s.Components.Grid == nil || s.Components.Grid.W == 0 {
			return nil, fmt.Errorf("scenario: %q no tiene nav_grid", ref)
		}
		return s.Components.Grid, nil
	}
	return nil, fmt.Errorf("scenario: no existe la entidad %q", ref)
}

// entityPos devuelve la posición de la entidad ref si está en surface.
func entityPos(entities []EntitySpec, ref, surface string) (units.Vec2, bool) {
	for _, e := range entities {
		if e.Ref != ref {
			continue
		}
		var s struct {
			Location struct {
				Space string     `json:"space"`
				Pos   units.Vec2 `json:"pos"`
			} `json:"location"`
		}
		if json.Unmarshal(e.Spawn, &s) != nil || s.Location.Space != surface {
			return units.Vec2{}, false
		}
		return s.Location.Pos, true
	}
	return units.Vec2{}, false
}

func deepCopy(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = deepCopy(e)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = deepCopy(e)
		}
		return out
	}
	return v
}

func round2(v float64) float64 { return float64(int64(v*100+0.5)) / 100 }
