# gt-sandbox

Simulador de vida multi-escala (giantess + tinies) en Go. Diseño: `docs/propuesta.md`. Plan del motor e hitos: `docs/plan-motor.md`.

## Flujo de trabajo

- Se trabaja siempre en `master`, con commits chicos y cada uno con sus tests.
- Antes de commitear: `gofmt -l .` (debe salir vacío), `go vet ./...`, `go test ./...`.
- `-race` requiere cgo (gcc), que no está instalado en esta máquina.
- Correr headless: `go run ./cmd/simrun -seed 42 -ticks 10000`.

## Reglas del motor (no negociables sin ADR en `docs/adr/`)

- **D1** Toda magnitud espacial autoritativa es `units.Length` (µm, `int64`). Nada de floats para posiciones.
- **D2** El tiempo es `clock.Tick` (20 TPS). La simulación nunca lee el reloj real (`time.Now` prohibido en `internal/sim` e `internal/core`).
- **D3** Referencias entre entidades con `ids.ID` generacional; siempre verificar `Alive` antes de usar un ID guardado.
- **D4** Aleatoriedad solo vía `w.RNG("<nombre-del-sistema>")`. Nunca `math/rand` global.
- **D5** Orden de iteración estable: nunca `range` sobre un `map` cuando afecta estado o hash. Ordenar claves.
- **D6** Floats permitidos en puntuaciones de IA/curvas, no en estado espacial.
- **D7** El estado solo cambia dentro de `World.Step()`, desde sistemas. Afuera entran comandos, salen eventos y snapshots.
- **D8** Simulación secuencial, sin goroutines dentro de `Step`.
- **D9** Entidades con stores propios detrás del paquete `ecs` (no ECS externo por ahora).

Todo sistema con estado propio implementa `statehash.Hashable`. Si cambia qué se hashea, incrementar `world.HashVersion`.

## Capas

```
core  ←  sim/*  ←  content, scenario  ←  render/camera/ui  ←  cmd/*
```

`internal/archtest` lo verifica. `internal/sim` nunca importa Ebitengine ni `render`. Si el test falla, mover el código, no relajar la regla.
