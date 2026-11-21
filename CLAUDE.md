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

## Convenciones del núcleo (ver `docs/specs/nucleo.md`)

- Crear mundos siempre con `engine.New(...)`: define la "forma" del mundo, y save/load/replay requieren la misma forma. Los sistemas nuevos del núcleo se instalan ahí.
- Componentes: structs de datos puros con campos exportados y tags JSON; el hash y los snapshots salen de su JSON. Nombre de store estable (`ecs.Register[T](r, "nombre")`).
- Comandos y eventos: tipo con `Kind()` estable, registrados con `world.RegisterCommand[T]` / `world.RegisterEvent[T]`. Un `Apply` que falla no debe haber modificado nada.
- Sistemas con estado propio implementan `world.Stateful`.
- Jerarquía espacial: un Space es una entidad con `space.Space`; todo cuelga por `comp.Location` (ver `docs/specs/space.md`). Posiciones de mundo siempre vía `space.Of(reg)` (`Frame`, `WorldPose`, `Reparent`), nunca sumando offsets a mano.
- Ángulos con `fixedmath.Angle`; nunca `math.Sin`/`math.Cos` en la simulación. La tabla se regenera con `go generate ./internal/core/fixedmath`.
- Cada subsistema registra sus invariantes con `world.AddInvariant`.
- Estado derivado (cachés, índices) va en recursos (`world.SetResource`/`world.Resource`), nunca en componentes: no se hashea ni se guarda.
- Navegación (ver `docs/specs/nav.md`): grilla `nav_grid` en el Space, portales como entidades con `portal`, componente `mover` con ruta en tramos; comandos `move_to`/`stop`/`set_cells`/`set_portal`. Para generar escenarios grandes, escribir un script que emita el JSON.
- Contactos, portales dinámicos y poses en `sim/attach` (ver `docs/specs/attach.md`); caídas en `space.falls`. Una entidad con `comp.Body` debe estar apoyada en su Space (dentro de la extensión, Z≈0) o caerá.
- Cambiar qué se hashea → incrementar `world.HashVersion`. Cambio incompatible en snapshots → incrementar `world.SnapshotSchema`.
- `simrun`: `go run ./cmd/simrun -seed 7 -ticks 300 -spawn 50 -save s.json` y `-load s.json`; `-scenario tests/scenarios/x.json` corre un escenario.
- Escenarios de integración en `tests/scenarios/*.json` (los corre `TestScenarios`). Escribir archivos con UTF-8 y LF.
- Rendimiento: `go run ./cmd/simbench -n 200`.
- IA (ver `docs/specs/ai.md`): `needs`, `agent` + `offers` (smart objects), `job`/`reservation`/`carried`. Las acciones (`ai.Action`) no guardan estado propio: todo va en `agent.doing`. Desde un sistema, mover con `nav.Approach`/`nav.Halt` y reservar con `jobs.Reserve`. Objetos en `content/objects/*.json`, usados en escenarios con `"object"`.
- Visor de debug (ver `docs/specs/visor.md`): `go run ./cmd/game` (escenario de M5), `-scenario x.json`, `-crowd 200`, `-load s.json`; `-shot out.png -shot-tick N [-micro palm]` guarda una captura y sale (sirve para revisar cambios visuales sin interactuar). El renderer solo dibuja `view.Frame`; nunca leer el mundo desde `render`/`ui`.
- Go: `go.mod` fija `toolchain go1.27.4` (lo descarga el comando `go`).

## Capas

```
core  ←  sim/*  ←  content, scenario  ←  view, camera  ←  render/ui  ←  cmd/*
```

`internal/archtest` lo verifica. `internal/sim` nunca importa Ebitengine ni `render`; `view` y `camera` tampoco importan Ebitengine. Si el test falla, mover el código, no relajar la regla.
