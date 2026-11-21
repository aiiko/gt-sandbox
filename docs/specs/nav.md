# Spec: Movimiento y navegación local (M3)

Paquetes: `sim/nav`, `sim/spatial`, `scenario`. Herramientas: `simrun -scenario`, `cmd/simbench`.

## Grilla (`nav_grid`)

- Componente de un Space: grilla regular en coordenadas locales del Space. `Origin` es la esquina de la celda (0,0); x crece con X, y con Y.
- Costo por celda: `0` bloqueada, `1` normal, `2`–`9` más lenta o menos deseable. Fuera de la grilla = bloqueada.
- JSON: `{"origin", "cell", "version", "rows": ["..#.."]}`. `#` bloqueada, `.` costo 1, dígitos 2–9. La fila 0 es la de menor Y (en el archivo, el escritorio aparece "de cabeza").
- `Version` sube con cada cambio de celda. Máximo `MaxCells` (2²²) celdas.
- Un Space sin grilla es campo libre: los movers van en línea recta.

## A*

- 8 direcciones, sin cortar esquinas (un paso diagonal requiere las dos ortogonales transitables).
- Costos enteros: 10 recto, 14 diagonal, × costo de la celda destino. Heurística octil (admisible).
- Desempate total `(f, h, índice)`: misma grilla y extremos ⇒ misma ruta.
- La ruta devuelta son los puntos donde cambia la dirección (centros de celda) y termina exactamente en el destino. Saltear celdas colineales es seguro: el segmento recto atraviesa solo celdas de la ruta o las ortogonales de un paso diagonal.
- `Searcher` reutiliza memoria; no es estado de la simulación.

## Mover

Componente `mover`: `speed` (µm/s), `status` (`idle`/`planning`/`moving`), `space`, `goal`, `path`, `path_version`.

| Sistema | Fase | Qué hace |
|---|---|---|
| `nav.plan` | Act | Corre A* para los movers en `planning`, hasta `PlanBudget` (64) por tick en orden del store. |
| `nav.move` | Move | Avanza `speed/TPS` µm por la ruta. Si la grilla cambió (`version != path_version`), se queda quieto y vuelve a `planning`. |

Al llegar: `move_arrived {entity}`. Al fallar: `move_failed {entity, reason}` con `no_path`, `not_walkable`, `search_budget` o `space_changed` (la entidad cambió de Space durante el movimiento; la navegación entre Spaces es M4).

## Comandos

- `move_to {entity, goal}`: goal en coordenadas locales del Space actual. Reemplaza el movimiento en curso.
- `stop {entity}`: vuelve a `idle` sin evento.
- `set_cells {space, cells: [{x, y, cost}]}`: atómico; se rechaza si alguna celda está fuera de rango o si bloquearía una celda ocupada por un mover.
- `spawn.components`: componentes de otros subsistemas por nombre de store (p. ej. `nav_grid`, `mover`).

## Invariantes

Un mover en `moving` tiene ruta no vacía y está en una celda transitable de su grilla.

## Índice espacial

`spatial.Of(w)` es un recurso derivado (no se hashea ni se guarda) con buckets de 2 cm por Space, sobre posiciones locales. Se reconstruye en `spatial.pre` (inicio de los sistemas, después de los comandos) y `spatial.post` (fin de Move). `Near(space, centro, radio, fn)` itera en orden determinista.

## Escenarios

`tests/scenarios/*.json` (formato en el doc de `internal/scenario`): entidades con `ref`, strings `"$ref"` reemplazados por IDs, comandos por tick, expectativas `event` (con `entity` y `by` opcionales), `no_event` y `at` (posición final ± `tol`). `TestScenarios` corre todos dos veces y compara hashes; las invariantes se chequean en cada tick.

## Rendimiento

`go run ./cmd/simbench -n 200`: ~0,03 ms/tick con 200 tinies en un escritorio de 120×80 celdas con obstáculos (objetivo: < 1 ms/tick).
