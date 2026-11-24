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
	"gtsandbox/internal/sim/needs"
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
// Si el base le fija a un agente su cuidado (agent.care, la giantess que mira
// antes de apoyar la mano), también se sortea. La semilla del mundo pasa a ser
// seed y se quitan las expectativas: una partida no tiene veredicto. Con la
// misma semilla, la misma partida. Al final sortea la diversión de los
// tinies y la despensa inicial (ver randomLife).
func Randomize(base *Spec, seed uint64) (*Spec, error) {
	return RandomizeCount(base, seed, 0)
}

// RandomizeCount es como Randomize pero permite fijar la cantidad de tinies
// cuando count > 0 (con count <= 0 sortea entre RandomMinTinies y
// RandomMaxTinies).
func RandomizeCount(base *Spec, seed uint64, count int) (*Spec, error) {
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
	var cells, restCells []units.Vec2
	for y := range grid.H {
		for x := range grid.W {
			p := grid.Center(x, y)
			d := p.Sub(center)
			if grid.At(x, y) == nav.Blocked {
				continue
			}
			if max(d.X, -d.X) < randomNestGap && max(d.Y, -d.Y) < randomNestGap {
				continue
			}
			if max(d.X, -d.X) > randomSpread || max(d.Y, -d.Y) > randomSpread {
				restCells = append(restCells, p)
				continue
			}
			cells = append(cells, p)
		}
	}
	if count > len(cells) {
		cells = append(cells, restCells...)
	}
	if len(cells) == 0 {
		return nil, fmt.Errorf("scenario: no hay celdas libres cerca del nido")
	}
	n := count
	if n <= 0 {
		n = min(RandomMinTinies+rnd.IntN(RandomMaxTinies-RandomMinTinies+1), len(cells))
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
	names := PickNames(DefaultTinyNames(), n, rng.NewStream(seed, "names"))
	for i := range n {
		spawn := deepCopy(mold).(map[string]any)
		spawn["name"] = names[i]
		p := cells[order[i%len(cells)]]
		if jitter > 0 {
			p.X += units.Length(rnd.Int64N(2*jitter+1) - jitter)
			p.Y += units.Length(rnd.Int64N(2*jitter+1) - jitter)
		}
		spawn["location"] = map[string]any{"space": surfaceRef, "pos": map[string]any{"x": int64(p.X), "y": int64(p.Y), "z": 0}}
		c := spawn["components"].(map[string]any)
		if needs, ok := c["needs"].([]any); ok {
			for _, need := range needs {
				if nd, ok := need.(map[string]any); ok && !isLateNeed(nd["name"]) {
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
	commands, err := randomCare(base.Commands, rnd)
	if err != nil {
		return nil, err
	}
	out.Commands = commands
	if base.Random != nil && len(base.Random.Starts) > 0 {
		start := base.Random.Starts[rnd.IntN(len(base.Random.Starts))]
		if err := applyStart(out.Entities, strings.TrimPrefix(base.Random.Agent, "$"), start); err != nil {
			return nil, err
		}
		out.Profile = start.Name
	}
	if err := randomLife(out.Entities[len(rest):], out.Entities, strings.TrimPrefix(colony, "$"), rnd); err != nil {
		return nil, err
	}
	if err := randomTrades(out.Entities[len(rest):], out.Entities, rnd); err != nil {
		return nil, err
	}
	if err := randomActivities(out.Entities[len(rest):], rnd); err != nil {
		return nil, err
	}
	out.Random = nil
	return &out, nil
}

// lateNeed es la necesidad de los tinies que se sortea al final (la
// diversión llegó después que el resto: así no cambia lo ya sorteado).
const (
	lateNeed   = "fun"
	thirstNeed = "thirst"
)

func isLateNeed(name any) bool {
	return name == lateNeed || name == thirstNeed
}

// randomActivities sortea las necesidades del slice de actividades (thirst) al
// final de Randomize, para no alterar lo ya sorteado.
func randomActivities(tinies []EntitySpec, rnd *rng.Stream) error {
	for i, e := range tinies {
		var spawn map[string]any
		if err := json.Unmarshal(e.Spawn, &spawn); err != nil {
			return err
		}
		c, _ := spawn["components"].(map[string]any)
		list, _ := c["needs"].([]any)
		changed := false
		for _, need := range list {
			if nd, ok := need.(map[string]any); ok && nd["name"] == thirstNeed {
				if v, ok := nd["value"].(float64); ok {
					nd["value"] = min(1000, int64(v*(0.6+0.5*rnd.Float64())))
					changed = true
				}
			}
		}
		if changed {
			raw, err := json.Marshal(spawn)
			if err != nil {
				return err
			}
			tinies[i].Spawn = raw
		}
	}
	return nil
}

func hasFoodStock(stock any) bool {
	switch s := stock.(type) {
	case float64:
		return s > 0
	case map[string]any:
		v, _ := s["food"].(float64)
		return v > 0
	case []any:
		for _, it := range s {
			if m, ok := it.(map[string]any); ok && m["resource"] == "food" {
				if v, _ := m["amount"].(float64); v > 0 {
					return true
				}
			}
		}
	}
	return false
}

// randomLife sortea, después de todo lo demás, la diversión de los tinies
// y la despensa inicial del nido (de 0 a una porción por tiny, por debajo
// de su meta si tiene stock).
func randomLife(tinies, entities []EntitySpec, nest string, rnd *rng.Stream) error {
	for i, e := range tinies {
		var spawn map[string]any
		if err := json.Unmarshal(e.Spawn, &spawn); err != nil {
			return err
		}
		c, _ := spawn["components"].(map[string]any)
		list, _ := c["needs"].([]any)
		changed := false
		for _, need := range list {
			if nd, ok := need.(map[string]any); ok && nd["name"] == lateNeed {
				if v, ok := nd["value"].(float64); ok {
					nd["value"] = min(1000, int64(v*(0.6+0.5*rnd.Float64())))
					changed = true
				}
			}
		}
		if changed {
			raw, err := json.Marshal(spawn)
			if err != nil {
				return err
			}
			tinies[i].Spawn = raw
		}
	}
	for i, e := range entities {
		if e.Ref != nest {
			continue
		}
		var spawn map[string]any
		if err := json.Unmarshal(e.Spawn, &spawn); err != nil {
			return err
		}
		c, _ := spawn["components"].(map[string]any)
		col, _ := c["colony"].(map[string]any)
		portions, _ := c["portions"].(map[string]any)
		if !hasFoodStock(col["stock"]) || portions == nil {
			return nil
		}
		n := rnd.IntN(len(tinies) + 1)
		if items, ok := portions["items"].([]any); ok {
			updated := false
			for _, it := range items {
				if m, ok := it.(map[string]any); ok && m["resource"] == "food" {
					m["amount"] = n
					updated = true
				}
			}
			if !updated && n > 0 {
				portions["items"] = append(items, map[string]any{"resource": "food", "amount": n})
			}
		} else {
			portions["left"] = n
		}
		raw, err := json.Marshal(spawn)
		if err != nil {
			return err
		}
		entities[i].Spawn = raw
	}
	return nil
}

// randomTrades sortea oficios iniciales (mezcla equilibrada con generalistas)
// e inclinación por la aventura independiente, solo cuando la partida incluye
// destinos de expedición o puestos de vigía (activación por contenido).
func randomTrades(tinies, entities []EntitySpec, rnd *rng.Stream) error {
	hasAdventureMap := false
	for _, e := range entities {
		if strings.Contains(string(e.Spawn), `"expedition_target"`) || strings.Contains(string(e.Spawn), `"lookout_post"`) {
			hasAdventureMap = true
			break
		}
	}
	if !hasAdventureMap {
		return nil
	}
	roles := []string{"gatherer", "porter", "builder", "lookout", "scout", "", ""}
	for i, e := range tinies {
		var spawn map[string]any
		if err := json.Unmarshal(e.Spawn, &spawn); err != nil {
			return err
		}
		c, _ := spawn["components"].(map[string]any)
		m, _ := c["member"].(map[string]any)
		if m == nil {
			continue
		}
		if i < 5 {
			m["role"] = roles[i]
		} else {
			m["role"] = roles[rnd.IntN(len(roles))]
		}
		if rnd.Float64() < 0.6 {
			m["adventure"] = round2(0.4 + 0.5*rnd.Float64())
		} else {
			m["adventure"] = 0.0
		}
		if mv, _ := c["mover"].(map[string]any); mv != nil {
			if _, has := mv["scale"]; !has {
				mv["scale"] = "tiny"
			}
		}
		raw, err := json.Marshal(spawn)
		if err != nil {
			return err
		}
		tinies[i].Spawn = raw
	}
	return nil
}

// RandomSpec son los perfiles de arranque de un agente (la giantess) entre
// los que Randomize elige uno: cada perfil reemplaza valor y decaimiento de
// sus necesidades. Así la rutina cambia entre partidas.
type RandomSpec struct {
	Agent  string        `json:"agent"` // "$ref" del agente
	Starts []RandomStart `json:"starts"`
}

// RandomStart es un perfil de arranque: con hambre, cansada…
type RandomStart struct {
	Name  string       `json:"name"`
	Needs []needs.Need `json:"needs"`
}

// applyStart reemplaza en la entidad ref las necesidades del perfil (las que
// no tenía se agregan al final).
func applyStart(entities []EntitySpec, ref string, start RandomStart) error {
	for i, e := range entities {
		if e.Ref != ref {
			continue
		}
		var spawn map[string]any
		if err := json.Unmarshal(e.Spawn, &spawn); err != nil {
			return fmt.Errorf("scenario: entidad %q: %w", ref, err)
		}
		comps, _ := spawn["components"].(map[string]any)
		if comps == nil {
			comps = map[string]any{}
			spawn["components"] = comps
		}
		var list []needs.Need
		if raw, ok := comps["needs"]; ok {
			b, _ := json.Marshal(raw)
			if err := json.Unmarshal(b, &list); err != nil {
				return fmt.Errorf("scenario: entidad %q: needs: %w", ref, err)
			}
		}
		for _, n := range start.Needs {
			if j := slices.IndexFunc(list, func(m needs.Need) bool { return m.Name == n.Name }); j >= 0 {
				list[j] = n
			} else {
				list = append(list, n)
			}
		}
		comps["needs"] = list
		raw, err := json.Marshal(spawn)
		if err != nil {
			return err
		}
		entities[i].Spawn = raw
		return nil
	}
	return fmt.Errorf("scenario: el perfil de arranque es de %q, que no existe", ref)
}

// RandomMaxCare es el cuidado máximo que se sortea para la giantess.
const RandomMaxCare = 0.9

// randomCare sortea, entre 0 y RandomMaxCare, el cuidado (agent.care) que
// los comandos set_components del base le fijan a un agente: cuánto mira
// antes de apoyar la mano. Los demás comandos quedan igual.
func randomCare(commands []CommandSpec, rnd *rng.Stream) ([]CommandSpec, error) {
	out := slices.Clone(commands)
	for i, c := range out {
		if c.Kind != "set_components" {
			continue
		}
		var data map[string]any
		if err := json.Unmarshal(c.Data, &data); err != nil {
			return nil, fmt.Errorf("scenario: comando %d: %w", i, err)
		}
		comps, _ := data["components"].(map[string]any)
		agent, _ := comps["agent"].(map[string]any)
		if _, ok := agent["care"]; !ok {
			continue
		}
		agent["care"] = round2(RandomMaxCare * rnd.Float64())
		raw, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		out[i].Data = raw
	}
	return out, nil
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
