# Plan de desarrollo del motor base

Este documento baja la propuesta (`docs/propuesta.md`) a un plan concreto para construir **el núcleo de simulación** sobre el que después se van a montar la IA, la colonia, el body map y la presentación.

El objetivo no es el vertical slice todavía. Es tener una base que:

1. Sea **determinista y testeable headless** desde el primer día.
2. Tenga los **contratos estables** (IDs, unidades, ubicación, eventos, comandos, guardado) que todo lo demás va a usar. Cambiarlos después es caro.
3. Deje **puntos de extensión** claros para agregar sistemas sin tocar el núcleo.

---

## 1. Decisiones fundacionales (tomarlas ahora y no tocarlas)

Estas son las decisiones que, si salen mal, obligan a reescribir. Cada una debería quedar registrada como un ADR corto en `docs/adr/`.

| # | Decisión | Elección | Motivo |
|---|----------|----------|--------|
| D1 | Unidad espacial canónica | **Micrómetros (µm) en `int64`** | Con mm no alcanza: un tiny de 8 mm a 20 TPS se mueve fracciones de mm por tick. En µm una casa de 20 m son 2·10⁷, muy lejos del límite de `int64`. Enteros = determinismo y sin drift. |
| D2 | Tiempo | **Tick fijo de 50 ms (20 TPS)**, contador `uint64`. Nada en la simulación lee el reloj real. | Replays, tests y simulación fuera de cámara dependen de esto. |
| D3 | Identidad | **IDs generacionales** (`index uint32 + gen uint32`). Nunca se reutiliza un ID vivo; referencias viejas se detectan. | Los tinies mueren, los objetos se destruyen; sin generación aparecen referencias colgantes silenciosas. |
| D4 | Aleatoriedad | **PRNG explícito y sembrado** (`math/rand/v2` PCG), un stream por sistema derivado de la semilla del mundo. Prohibido `rand` global. | Agregar un sistema nuevo no debe cambiar los resultados de los demás. |
| D5 | Orden de iteración | **Siempre estable**: slices ordenadas por ID, nunca `range` sobre `map` en lógica que afecte estado. | Es la fuente nº1 de no-determinismo en Go. |
| D6 | Floats | Permitidos para **puntuaciones de IA y curvas**, nunca para estado espacial autoritativo. | Pragmático: utility AI con enteros es incómodo; las posiciones sí deben ser exactas. |
| D7 | Mutación del estado | Solo dentro del tick, por sistemas. Afuera entran **comandos**; afuera salen **eventos** y **snapshots**. | Es la separación sim/render de la propuesta, hecha cumplir por la API. |
| D8 | Concurrencia | **Simulación secuencial.** Paralelizar más tarde, por sistema, si el profiler lo pide. | Determinismo primero. |
| D9 | Entidades | Stores propios por componente (no ECS externo todavía), detrás de una interfaz que permita migrar a Ark si hace falta. | Ver §3.2. |

---

## 2. Estructura del repositorio

Ligeramente ajustada respecto a la propuesta para que las dependencias vayan en una sola dirección:

```
gt-sandbox/
├── go.mod                     # module gtsandbox (o el nombre que elijas)
├── cmd/
│   ├── game/                  # Ebitengine (llega en M6)
│   ├── simrun/                # Corre escenarios headless, imprime log/hash
│   └── simbench/              # Benchmarks con N agentes
├── internal/
│   ├── core/                  # SIN dependencias internas
│   │   ├── units/             # µm, Vec3, ángulos, conversiones
│   │   ├── ids/               # IDs generacionales, allocator
│   │   ├── rng/               # Streams deterministas
│   │   └── fixedmath/         # Transform 2D+Z, rotaciones discretas
│   ├── sim/
│   │   ├── ecs/               # Entity store + component stores
│   │   ├── clock/             # Tick, scheduler, fases, staggering
│   │   ├── events/            # Bus de eventos diferido
│   │   ├── commands/          # Cola de comandos externos
│   │   ├── world/             # World: agrupa todo, Step()
│   │   ├── space/             # Spaces, jerarquía, transforms, superficies
│   │   ├── spatial/           # Índice espacial por Space
│   │   ├── nav/               # Grid local + grafo de portales
│   │   ├── attach/            # Anclajes y contactos dinámicos
│   │   ├── ai/  needs/  jobs/ # Esqueletos en M7, contenido después
│   │   └── save/              # Snapshot versionado + hash de estado
│   ├── content/               # Carga y validación de definiciones JSON
│   ├── scenario/              # Formato y runner de escenarios de test
│   ├── render/  camera/  ui/  # Solo leen snapshots
├── content/                   # Datos: objetos, actores, mapas
├── tests/scenarios/           # Escenarios .json + golden files
└── docs/
    ├── propuesta.md
    ├── plan-motor.md
    ├── adr/                   # Decisiones registradas
    └── specs/                 # Una spec corta por sistema
```

**Regla de dependencias** (verificable con un test que analice imports):

```
core  ←  sim/*  ←  content, scenario  ←  render/ui  ←  cmd/*
```

`sim/*` nunca importa Ebitengine. `render` nunca escribe en `sim`.

---

## 3. Componentes del núcleo

### 3.1. El loop de simulación

```go
type World struct { /* stores, spaces, scheduler, bus, rngs, tick */ }

func (w *World) Enqueue(cmd commands.Command)   // thread-safe, desde input/UI
func (w *World) Step()                          // avanza exactamente 1 tick
func (w *World) Snapshot() *Snapshot            // vista inmutable para el render
func (w *World) Events() []events.Event         // eventos del último tick
func (w *World) Hash() uint64                   // hash del estado lógico
```

Fases fijas dentro de un `Step()`:

1. **ApplyCommands** — comandos encolados, en orden de llegada con número de secuencia.
2. **Perception / Decide** — IA (solo para entidades a las que les toca este tick).
3. **Act** — ejecución de acciones/jobs en curso; producen *intenciones* de movimiento.
4. **Move** — resolución de movimiento local, transforms de spaces, attachments.
5. **Contacts / Topology** — detección de contactos, alta/baja de portales dinámicos.
6. **Consequences** — reacciones a eventos físicos (caídas, aplastamiento, derrames…).
7. **Slow** — necesidades, metabolismo, procesos lentos (por intervalo).
8. **FlushEvents** — los eventos emitidos durante el tick se publican al final.

Los sistemas se registran declarando fase, prioridad dentro de la fase y frecuencia. El scheduler distribuye entidades con `(entityIndex + tick) % period` para no recalcular todo en el mismo tick.

### 3.2. Entidades y componentes

- `EntityStore`: alloc/free de IDs generacionales, lista de vivas ordenada.
- `Store[T]`: genérico, *sparse set* (array denso + índice disperso). Iteración en orden de inserción estable o por ID; `Get`, `Set`, `Remove`, `Each`.
- Destrucción diferida: `Despawn` marca y se ejecuta al final del tick para no invalidar iteraciones.
- Toda la API pasa por un paquete `ecs` propio. Si más adelante Ark conviene, se reemplaza la implementación sin tocar los sistemas.

Componentes iniciales (solo datos, sin lógica):

| Componente | Contenido |
|------------|-----------|
| `Location` | `Space SpaceID`, `Pos units.Vec3` (local, µm), `Facing` |
| `Body` | Tamaño/radio de colisión en µm, masa aproximada, escala |
| `Mover` | Velocidad máx., intención de movimiento actual, ruta |
| `Attachment` | Si está agarrado/sujeto a algo: a qué y con qué fuerza |
| `SpaceOwner` | Spaces que esta entidad posee (mano, escritorio, zapato) |
| `Name`, `Tags` | Debug y consultas de contenido |

### 3.3. Spaces (el corazón multi-escala)

Un **Space** es una superficie navegable con su propio sistema de coordenadas local.

```go
type Space struct {
    ID        SpaceID          // también generacional
    Kind      SpaceKind        // Static, Elevated, MovingRigid, Articulated, Built
    Parent    ParentRef        // Mundo, otro Space, o (Entity, Anchor)
    Local     fixedmath.Transform // relativo al padre
    Extent    units.Rect       // límites en coords locales
    Grid      *nav.Grid        // resolución propia (5 mm micro, 10 cm macro…)
    Gravity   GravityRule      // "arriba" del space, inclinación, si se puede caer
}
```

Puntos clave:

- **Transform mundial = composición de la cadena de padres** (`T_giantess · T_parte · P_local`). Se cachea por tick con flag *dirty*: si nadie movió la mano, no se recalcula.
- Un tiny en `Space=LeftPalm` no se toca cuando la giantess camina. Solo cambian transforms.
- `space.ToWorld(loc)`, `space.FromWorld(space, pos)` y `space.Reparent(entity, newSpace)` (preserva la posición mundial) son las tres operaciones que todo lo demás usa.
- La **pose** de una entidad articulada es simplemente un conjunto de transforms locales de sus spaces hijos; cambiar de pose = interpolar/reemplazar esos transforms. No hace falta esqueleto.
- **Invariante crítico:** toda entidad con `Location` está en exactamente un Space vivo. Si un Space se destruye, sus ocupantes se reubican (caen) en el mismo tick. Se chequea en modo debug cada tick.

### 3.4. Navegación en dos niveles

- **Local:** grid por Space (celdas con altura, transitabilidad, material). A* con heurística octil. Resolución definida por el Space, no global.
- **Global:** grafo donde los nodos son Spaces y las aristas **Portales** (`From`, `To`, segmento de entrada en cada lado, costo, condición). Dijkstra/A* sobre el grafo, luego A* local por tramo.
- **Portales estáticos** vienen del mapa (pata del escritorio → superficie).
- **Portales dinámicos** los crea/destruye el sistema de contactos (mano apoyada sobre el escritorio). Cada portal tiene versión; las rutas guardan las versiones que usaron y se invalidan perezosamente cuando cambian.
- El pathfinding expone `RequestPath` y devuelve un handle; así después se puede hacer asincrónico o presupuestado por tick sin cambiar la API.

### 3.5. Attachments y contactos

Es lo que vuelve a la giantess un "mapa móvil":

- Un **Anchor** es un punto con nombre en un Space (`"palm_center"`, `"sole"`).
- Un **Contact** es una relación temporal entre dos Spaces (o entidad↔Space) con área de solapamiento, calculada en la fase *Contacts*.
- Reglas de contenido traducen contactos a portales: `si palma toca superficie Elevated con solapamiento > X → portal bidireccional`.
- Eventos emitidos: `ContactBegan`, `ContactEnded`, `PortalOpened`, `PortalClosed`, `EntityFell`, `EntityReparented`.
- Movimiento brusco de un Space (aceleración entre ticks) produce un evento que los sistemas de equilibrio de los tinies consumirán más adelante.

### 3.6. Eventos y comandos

- **Comandos:** structs tipados (`SpawnEntity`, `MoveTo`, `SetPose`, `DebugTeleport`…) con `Apply(*World) error`. Todo lo que el jugador, la UI o un test hace al mundo pasa por acá. Se pueden grabar → replay.
- **Eventos:** structs tipados con `Tick` y `Seq`. Se acumulan durante el tick y se publican al final. Los suscriptores internos (sistemas) los leen en la fase correspondiente del tick siguiente → no hay reentrancia ni orden accidental.
- Los eventos son también la materia prima del **historial emergente** (Dwarf Fortress) y del log de debug.

### 3.7. Guardado y hash

- `Snapshot` versionado (`SchemaVersion` + migraciones), serializado con un encoder propio a JSON en dev (legible, diffeable) y a binario después.
- `Hash()` recorre stores en orden estable y hashea (FNV/xxhash) solo estado lógico.
- Test permanente: `save → load → N ticks` produce el mismo hash que `N ticks` directos.

### 3.8. Contenido

- Definiciones JSON (`content/objects/*.json`, `content/actors/*.json`) cargadas a un `Registry` con validación estricta al iniciar (campos desconocidos = error).
- Los mapas iniciales en JSON propio simple; LDtk se integra después como *importador* hacia ese formato, para que el motor no dependa de LDtk.
- Los Smart Objects (acciones ofrecidas) se declaran acá en M7.

### 3.9. Escenarios (la herramienta de desarrollo principal)

```json
{
  "name": "tiny_rides_moving_hand",
  "seed": 42,
  "load": "content/maps/test_room.json",
  "spawn": [ ... ],
  "commands": [ { "tick": 10, "cmd": "SetPose", "entity": "g1", "pose": "hand_raised" } ],
  "run_ticks": 200,
  "expect": [
    { "at": 200, "entity": "t1", "space": "g1.left_palm", "local_pos_unchanged": true }
  ]
}
```

- `cmd/simrun scenario.json` imprime log de eventos, hash final y falla si alguna expectativa no se cumple.
- Cada escenario en `tests/scenarios/` se corre desde `go test` y tiene un *golden* del log de eventos.
- Esto es lo que permite iterar con agentes (Opus/Codex): especificás comportamiento como escenario, el agente implementa hasta que pasa.

---

## 4. Hitos

Cada hito termina con: tests verdes, escenarios nuevos que pasan, y la spec del sistema en `docs/specs/`. No se empieza el siguiente sin cerrar el anterior.

### M0 — Esqueleto y determinismo
- `go.mod`, layout de carpetas, CI local (`go vet`, `staticcheck`, `go test -race`).
- `core/units`, `core/ids`, `core/rng`.
- `World` vacío con `Step()`, tick counter, scheduler con fases (sin sistemas reales).
- Test de arquitectura: `sim/*` no importa Ebitengine ni `render`.

**Aceptación:** `go test ./...` corre; un test de 10.000 ticks vacíos con misma semilla da el mismo hash dos veces.

**Estado: completado.** Módulo `gtsandbox`. Además del alcance: `core/statehash`, `cmd/simrun`, test que prohíbe `math/rand`/`time` en `sim`/`core`, y `CLAUDE.md`. Pendiente: `staticcheck` y `-race` (no hay gcc instalado para cgo).

### M1 — Entidades, comandos, eventos
- `ecs` con `Store[T]`, despawn diferido, IDs generacionales.
- Cola de comandos con secuencia, bus de eventos diferido.
- Componentes `Location`, `Body`, `Name`.
- `Hash()` + snapshot JSON + `Load`.

**Aceptación:** spawn/despawn masivo sin referencias colgantes (test con fuzzing); save/load/continuar = mismo hash; replay de comandos grabados reproduce el estado.

### M2 — Spaces y transforms
- `space` con jerarquía, transforms fijos 2D + Z + rotación, caché dirty.
- `ToWorld`, `FromWorld`, `Reparent`.
- Spaces adjuntos a entidades vía anchors. Una entidad "giantess" mínima con un Space `palm` hijo.
- Invariante "toda entidad está en un Space vivo" verificado en debug.

**Aceptación (tests de la propuesta):**
- Tiny sobre mano en movimiento conserva su posición local.
- Giantess cambia de habitación → sus spaces siguen siendo hijos y las posiciones mundiales de sus ocupantes son consistentes.
- `Reparent` preserva la posición mundial con error 0 µm (enteros).

### M3 — Movimiento y navegación local
- `nav.Grid` por Space, A* local, componente `Mover`, sistema de movimiento con intención → paso por tick.
- `spatial`: índice por Space (grid hash) para consultas de vecinos.
- `cmd/simrun` funcional con escenarios.

**Aceptación:** tiny cruza el escritorio esquivando obstáculos; 200 tinies con rutas en un Space bajo X ms/tick (fijar X tras primer benchmark).

### M4 — Grafo de spaces y portales
- Portales estáticos desde el mapa, pathfinding jerárquico (grafo → local).
- Versionado de portales e invalidación de rutas.
- Mapa de prueba: suelo → pata → escritorio.

**Aceptación:** tiny va del suelo a la superficie del escritorio por la pata; si se deshabilita el portal a mitad de camino, la ruta se recalcula o el movimiento se cancela con evento.

### M5 — Contactos y portales dinámicos
- Sistema de contactos entre Spaces, reglas de contacto→portal.
- Pose como conjunto de transforms; comando `SetPose`.
- Caída: si el Space del tiny desaparece o se inclina más allá del límite, se resuelve aterrizaje en el Space de abajo.

**Aceptación (el núcleo de la idea):**
- Mano toca mesa → aparece ruta válida; mano se levanta → la ruta se invalida.
- Escenario completo: tiny camina del escritorio a la palma apoyada, la giantess levanta la mano, se mueve a otra habitación, la apoya, el tiny baja. **Sin ningún script específico**: solo comandos de pose/movimiento sobre la giantess y una orden `MoveTo` al tiny.

### M6 — Renderer de debug (Ebitengine)
- `cmd/game`: loop Ebitengine que llama `Step()` a 20 TPS e interpola snapshots a 60 FPS.
- Vista macro (rectángulos por Space y entidad) y vista micro de un Space elegido. Sin arte: formas y colores.
- Overlay: grids, portales, contactos, rutas; panel con el log de eventos; click para inspeccionar entidad.
- Controles de tiempo: pausa, paso a paso, x1/x4/x16.

**Aceptación:** cambiar de cámara o de velocidad no cambia el hash; el escenario de M5 se ve funcionando.

### M7 — Esqueleto de IA (sin contenido rico)
- `needs`: valores con decaimiento por intervalo.
- `ai`: interfaz `Action` (precondiciones, utilidad, ejecución por pasos, interrumpible), selector utility con histéresis.
- Smart Objects: los objetos declaran en JSON qué acciones ofrecen y a qué escala.
- `jobs`: cola de trabajos, reservas con expiración, liberación al morir/cancelar.

**Aceptación:** dos tinies no reservan la misma migaja; una giantess con 3 necesidades y 3 acciones elige de forma estable (no oscila); un job interrumpido por un evento se retoma.

**Al terminar M7 el motor base está listo** y se puede atacar el vertical slice ("Una mañana en la casa") de la propuesta como trabajo de contenido y comportamiento, no de infraestructura.

---

## 5. Qué queda explícitamente fuera del motor base

Para no caer en la "explosión de complejidad" de la propuesta:

- LOD de simulación (se diseña la API del scheduler para permitirlo, pero no se implementa).
- Percepción con canales (visión/sonido/vibración) — llega con el vertical slice.
- Fluidos, temperatura, materiales.
- Construcción tiny, relaciones, personalidad.
- Arte, animación, LDtk, EbitenUI.
- Paralelismo.

Cada uno de estos debería poder agregarse como **un sistema nuevo registrado en una fase**, más componentes nuevos, sin modificar `world`, `space` ni `nav`. Si alguno lo requiere, es señal de que falta un punto de extensión y se resuelve con un ADR.

---

## 6. Prácticas de trabajo

- **Spec antes que código:** `docs/specs/<sistema>.md` de una página — responsabilidades, API, invariantes, casos límite, escenarios de aceptación.
- **Todo se trabaja en `master`**, con commits chicos y cada uno con sus tests.
- **Invariantes en debug:** `world.CheckInvariants()` corre cada tick en tests y en builds de debug.
- **Benchmarks desde M3:** `cmd/simbench` con 20/100/500/1000 entidades; guardar resultados para comparar regresiones.
- **CLAUDE.md** en la raíz con las reglas D1–D9 y la regla de dependencias, para que cualquier agente las respete sin volver a explicarlas.

---

## 7. Cosas a verificar antes de arrancar

- La propuesta afirma que **Ebitengine 2.10** salió en septiembre de 2026 con builds en Go puro y virtualización para tests con agentes. Conviene confirmar la versión actual y su requisito de Go (localmente hay Go 1.24.13; probablemente haya que actualizar a la versión que pida Ebitengine). No bloquea nada hasta M6.
- Elegir el nombre del módulo Go.
