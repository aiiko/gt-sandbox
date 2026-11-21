# Spec: Navegación (M3 local, M4 entre Spaces)

Paquetes: `sim/nav`, `sim/spatial`, `scenario`. Herramientas: `simrun -scenario`, `cmd/simbench`.

## Grilla (`nav_grid`)

- Componente de un Space: grilla regular en coordenadas locales del Space. `Origin` es la esquina de la celda (0,0); x crece con X, y con Y.
- Costo por celda: `0` bloqueada, `1` normal, `2`–`9` más lenta o menos deseable. Fuera de la grilla = bloqueada.
- JSON: `{"origin", "cell", "version", "rows": ["..#.."]}`. `#` bloqueada, `.` costo 1, dígitos 2–9. La fila 0 es la de menor Y (en el archivo, el escritorio aparece "de cabeza").
- `Version` sube con cada cambio de celda. Máximo `MaxCells` (2²²) celdas.
- Un Space sin grilla es campo libre: los movers van en línea recta.

## A* local

- 8 direcciones, sin cortar esquinas (un paso diagonal requiere las dos ortogonales transitables).
- Costos enteros: 10 recto, 14 diagonal, × costo de la celda destino. Heurística octil (admisible). El costo devuelto se expresa como distancia equivalente en µm, comparable entre Spaces de distinta resolución.
- Desempate total `(f, h, índice)`: misma grilla y extremos ⇒ misma ruta.
- La ruta devuelta son los puntos donde cambia la dirección (centros de celda) y termina exactamente en el destino. Saltear celdas colineales es seguro: el segmento recto atraviesa solo celdas de la ruta o las ortogonales de un paso diagonal.
- `Searcher` reutiliza memoria; no es estado de la simulación.

## Portales (`portal`)

- Componente de una entidad propia (no es un Space): `{a: {space, pos}, b: {space, pos}, cost, disabled, version}`. Bidireccional; los extremos son puntos en coordenadas locales de cada Space.
- `cost` es un costo extra de cruce como distancia equivalente. Un recorrido físico (trepar una pata) se modela mejor como un Space propio con su grilla, unido por dos portales: así el tiny tarda lo que corresponde y puede quedar a mitad de camino.
- `set_portal {portal, enabled}` cambia el estado y sube `version`.
- El sistema `nav.portals` (último del Flush, después de `space.orphans`) destruye los portales cuyos extremos están en Spaces que se destruyen.
- Los portales estáticos vienen del mapa (escenario o contenido); los dinámicos los creará el sistema de contactos (M5).

## Rutas entre Spaces (`Router`)

Grafo: nodos = origen, destino y los dos extremos de cada portal habilitado. Aristas = cruce de portal (costo `cost`) y tramos dentro de un mismo Space entre nodos de ese Space.

Búsqueda perezosa: los tramos empiezan estimados en línea recta; se corre Dijkstra (desempate por índice de nodo), se calcula con A* cada tramo estimado de la mejor ruta y se repite hasta que la mejor ruta sea toda exacta. A* corre solo en los tramos candidatos. Tope: `MaxSearches` (32) búsquedas por planificación.

La ruta resultante es una lista de **tramos** (`legs`): `{space, path, grid_version, portal, reverse, portal_version}`. Al terminar el `path` de un tramo con `portal`, el mover cruza al otro extremo (cambia su `Location`) y sigue con el tramo siguiente en el mismo tick.

## Mover

Componente `mover`: `speed` (µm/s), `status` (`idle`/`planning`/`moving`), `goal_space`, `goal`, `legs`.

| Sistema | Fase | Qué hace |
|---|---|---|
| `nav.plan` | Act | Planifica para los movers en `planning`, hasta `PlanBudget` (64) por tick en orden del store. |
| `nav.move` | Move | Valida la ruta y avanza `speed/TPS` µm, cruzando portales. |

**Invalidación perezosa:** antes de mover, `nav.move` verifica que la ruta siga vigente: el mover está en el Space del primer tramo, cada grilla tiene la versión con la que se planeó, y cada portal sigue vivo, habilitado, con su versión y conectando los mismos Spaces. Si algo cambió, el mover se queda quieto ese tick y vuelve a `planning`; la ruta se recalcula desde donde está (puede volver sobre sus pasos).

Eventos: `move_arrived {entity}`; `move_failed {entity, reason}` con `no_path`, `not_walkable` o `search_budget`; al cruzar un portal, `entity_reparented` con `cause: "portal"`.

## Comandos

- `move_to {entity, space?, goal}`: goal en coordenadas locales de `space` (por defecto, el Space actual). Reemplaza el movimiento en curso.
- `stop {entity}`: vuelve a `idle` sin evento.
- `set_cells {space, cells: [{x, y, cost}]}`: atómico; se rechaza si alguna celda está fuera de rango o si bloquearía una celda ocupada por un mover.
- `set_portal {portal, enabled}`.
- `spawn.components`: componentes de otros subsistemas por nombre de store (`nav_grid`, `mover`, `portal`).

## Invariantes

- Un mover en `moving` tiene ruta no vacía y, si está en el Space de su primer tramo, en una celda transitable.
- Todo portal tiene sus dos extremos en Spaces vivos.

## Índice espacial

`spatial.Of(w)` es un recurso derivado (no se hashea ni se guarda) con buckets de 2 cm por Space, sobre posiciones locales. Se reconstruye en `spatial.pre` (inicio de los sistemas, después de los comandos) y `spatial.post` (fin de Move). `Near(space, centro, radio, fn)` itera en orden determinista.

## Escenarios

`tests/scenarios/*.json` (formato en el doc de `internal/scenario`): entidades con `ref`, strings `"$ref"` reemplazados por IDs, comandos por tick, expectativas `event` / `no_event` (con `entity`, `match` por campos del evento y `by`), y `at` (posición final ± `tol`, opcionalmente en un `space`). `TestScenarios` corre todos dos veces y compara hashes; las invariantes se chequean en cada tick.

Mapa de M4 (`floor_to_desk_by_leg`, `portal_disabled_midway`, `portal_disabled_no_route`): habitación 2×2 m (celdas de 5 cm), escritorio 60×40 cm a 75 cm (celdas de 2 cm) y dos patas de 2×74 cm, cada una con un portal abajo (suelo) y otro arriba (escritorio).

## Pendiente

- La inclinación de los Spaces no existe todavía: una pata es una superficie lógica; su orientación en el mundo no es vertical.
- Extremos de portal como segmentos, portales de un solo sentido, condiciones por tamaño del tiny.
- Con muchos portales el grafo es grande (todos los pares de extremos de un Space): habrá que cachear costos entre portales o agrupar por componente conexa.

## Rendimiento

`go run ./cmd/simbench -n 200`: ~0,03 ms/tick con 200 tinies en un escritorio de 120×80 celdas con obstáculos (objetivo: < 1 ms/tick).
