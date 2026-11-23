# gt-sandbox

Simulador de vida multi-escala (giantess + tinies) en Go. El diseño está en `docs/propuesta.md`.

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
- `simrun`: `go run ./cmd/simrun -seed 7 -ticks 300 -spawn 50 -save s.json` y `-load s.json`; `-scenario tests/scenarios/x.json` corre un escenario (`-events` imprime su log de eventos).
- Escenarios de integración en `tests/scenarios/*.json` (los corre `TestScenarios`). Escribir archivos con UTF-8 y LF.
- Rendimiento: `go run ./cmd/simbench -n 200` (simulación) y `go run ./cmd/game -size 1920x1080 -bench 20s` (presentación); resultados en `docs/specs/presentacion.md#rendimiento`.
- IA (ver `docs/specs/ai.md`): `needs`, `agent` + `offers` (smart objects), `job`/`reservation`/`carried`. Las acciones (`ai.Action`) no guardan estado propio: todo va en `agent.doing`. Desde un sistema, mover con `nav.Approach`/`nav.Halt` y reservar con `jobs.Reserve`. Percepción con `agent.sight`/`known` (sin `sight` conoce todo), `watch` para notar agentes de otra escala; exploración con `curiosity`/`visited` (acción `explore` por portales); una oferta con mano apoyada o algo en la mano solo la corta algo urgente (`UrgentUtility`); colonias en `sim/colony` (nido = smart object con depósito en `portions`). Objetos en `content/objects/*.json`, usados en escenarios con `"object"`; las ofertas de varios pasos usan `steps`, y `"item": "<objeto>"` se expande al cargar.
- Aplastamiento (ver `docs/specs/crush.md`): `crusher` (palma con `thickness`, objetos al aterrizar), `stride` (pisadas al caminar, solo en el Space donde se camina), `fragile`, `shelter`, `remains`; solo un apoyo nuevo aplasta (`crusher.pressing`). Amenazas con `crush.Threats(w)` (sombra y próxima pisada); un agente con `reflex` se alarma y huye (`flee`, en `ai`). `agent.care` hace esperar y titubear a la giantess antes de apoyar la mano; `colony.mourn` crea zonas `danger` (store de `ai`) donde muere un miembro. La muerte es un despawn normal y los restos son una entidad nueva. Escenarios generados con `python tools/scenarios/aplastamiento.py`; el de cierre (`un_descuido_en_el_escritorio`, con la variante cuidadosa) con `python tools/scenarios/descuido.py`. En la presentación, restos (`remains`), próxima pisada (`Frame.Footprints`) y espíritu (`spirit`).
- Rutinas y encuentros (ver `docs/specs/encuentros.md`): la base de «Nueva partida» es `content/partidas/casa.json` (perfiles de arranque en el bloque `random`, merienda en el living); se genera con `python tools/scenarios/partida.py` después de `descuido.py` y, después, `python tools/scenarios/vida.py`, que le suma la vida del nido; encuentros en `sim/encounter` (componente `encounter` en la giantess: avistamientos por superficie, enfriamientos y las acciones `observe`, `share` y `offer_palm`, que hacen pasos de oferta con `ai.RunSteps`); `agent.revisit` hace que un Space visitado vuelva a interesar. Escenarios de encuentros generados por `partida.py`.
- Vida cotidiana de los tinies (ver `docs/specs/vida.md`): en `sim/colony`, meta de la despensa (`colony.stock`, `colony.Deficit`), puntos de búsqueda (`search_spot` + acción `search`), refugio auxiliar (`fiber_source`, `material`, `build_site`; trabajos `pluck`/`build` con `jobs.RegisterKind`) y reunión (`gathering` + acción `chat`, necesidad `fun`). Lugares que se ven de lejos con `landmark` (store de `ai`). Escenarios generados con `python tools/scenarios/vida.py`.
- Juego y visor de debug (ver `docs/specs/visor.md`): `go run ./cmd/game` abre el menú principal (nueva partida al azar con `scenario.Randomize`, escenarios con `atlas`, modo debug); `-new-game [-seed N]`, `-scenario x.json`, `-crowd 200`, `-load s.json` abren directo. Siempre en pixel art si el mundo tiene visuales; `-debug-view` (o un mundo sin visuales) usa el visor vectorial; `-shot out.png -shot-tick N [-micro palma] [-follow "Tiny 01"] [-pan dx,dy]` guarda una captura y sale; `-bench 20s` mide la presentación. El lienzo llena la ventana (`render.CanvasSize`, mínimo 768 × 432, factor hasta 3) y la escena va en una capa aparte cuyo factor entero cambia la rueda (`-zoom N` para arrancar con uno); sin `-size` la ventana se recuerda entre sesiones y `F11` alterna pantalla completa. Capturas de referencia en `docs/capturas/capturas.json`: regenerar con `go run ./cmd/capturas` y, si el dibujo cambió a propósito, `go test ./cmd/game -run TestReferenceCaptures -update`. El renderer solo dibuja `view.Frame`; nunca leer el mundo desde `render`/`ui`. Con la casa se abre la interfaz de partida (`ui.HUD`, armada en `cmd/game/hud.go`); la historia es `view.Chronicle` (`Runner.Chronicle`: crónica de hitos, rutina y resúmenes, y un diario por entidad y lugar con el detalle); `F3`/`-debug` superpone los overlays técnicos.
- Arte (ver `docs/specs/presentacion.md`): editar las fuentes en `content/art/src/casa/` (un PNG por cuadro + `atlas.json`) y regenerar con `go run ./cmd/artpack`; nunca tocar `content/art/casa.{png,json}` a mano. Las piezas dibujadas por código salen de `tools/art/` (`python tools/art/build.py`, con `-check` para verificar): ajustarlas en el script, no en el PNG. Todo el macro (80 px/m, densidad declarada en `atlas.json`) sale de `macro.py` y `giantess.py`.
- Sonido (ver `docs/specs/sonido.md`): solo presentación, nunca cambia el hash. Los sonidos se sintetizan desde `tools/audio/sounds.json` con `python tools/audio/build.py` (`-check` para verificar); nunca tocar `content/audio/` a mano. `internal/sound` decide qué suena (`Mixer`, `UIWatcher`, `Scene` a partir de `Runner.Events` y el `Frame`, con perspectiva por escala) sin importar Ebitengine; la reproducción está en `cmd/game/audio.go`. Ambiente por habitación con `sound.Ambience` (bucles `"loop": true`). Mezcla de referencia en `docs/sonido/mezcla.json`: `go run ./cmd/game -mix out.wav` la renderiza y `TestReferenceMix` verifica su resumen (`-update` si cambió a propósito). `M` silencia; `-mute`, `-volume`.
- Go: `go.mod` fija `toolchain go1.27.4` (lo descarga el comando `go`).

## Capas

```
core  ←  sim/*  ←  content, scenario  ←  view, camera  ←  render/ui, sound  ←  cmd/*
```

`internal/archtest` lo verifica. `internal/sim` nunca importa Ebitengine ni `render`; `view`, `camera` y `sound` tampoco importan Ebitengine. Si el test falla, mover el código, no relajar la regla.
