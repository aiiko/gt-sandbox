# Spec: Visor de debug (M6)

Paquetes: `view` (Frame, Runner, picking), `camera` (proyección), `render` (dibujo), `ui` (paneles de texto), `cmd/game` (loop de Ebitengine).

## Frame (`view.Frame`)

Foto plana y serializable de un tick, armada por `Capturer.Capture(w)` (solo lee el mundo):

- `spaces`: id, nombre, kind, padre, profundidad, pose de mundo, extensión, copia de la grilla (cacheada por versión) y si es superficie de contacto. Ordenados de abajo hacia arriba (Z de mundo, luego ID).
- `things`: entidades con `location` que no son Spaces: pose de mundo, cuerpo y, si tienen `mover`, su estado y la ruta pendiente en puntos de mundo.
- `portals` (extremos en el mundo; `dynamic` si los creó un contacto) y `contacts`.

El renderer dibuja solo desde el Frame. `view.Inspect(w, id)` es la única lectura directa del mundo (el panel del inspector: JSON de cada componente vía `ecs.Registry.Components`).

## Tiempo (`view.Runner`)

- `Advance(dt)` acumula `dt × velocidad` y corre los `Step` que entran (máximo 64 por frame; el resto se descarta y la simulación se ralentiza). `StepOnce` es el paso a paso. Velocidades: x1, x4, x16.
- Guarda `Prev` y `Cur` (Frames antes y después del último Step) y `Alpha` (fracción del tick siguiente). `Interpolate(prev, cur, alpha)` interpola poses de Spaces y Things para dibujar a 60 FPS; saltos de más de 1 m no se interpolan.
- Guarda los últimos 200 eventos ya serializados. `Hash()` se calcula una vez por tick.
- El paso de un escenario es `scenario.Player.Step` (encola los comandos del tick); pasado el final, el mundo sigue sin comandos y el HUD muestra el resultado de las expectativas.

**Garantía:** el visor solo decide *cuántas* veces llamar a `Step` y encola comandos del jugador. Cámara, velocidad, pausa, capturas, interpolación y picking no cambian el hash (`TestSpeedDoesNotChangeHash`).

## Cámara (`camera`)

Sigue a un Space ancla (`Nil` = mundo): el centro está en coordenadas locales del ancla y la vista gira con él. Vista macro = ancla en el mundo; vista micro = ancla en un Space (p. ej. la palma, que sigue a la mano). Las coordenadas pasan a float relativas al ancla. +Y local va hacia arriba en pantalla. `ZoomAt` mantiene el punto bajo el cursor; `Follow` cambia de ancla sin mover la vista.

## Dibujo

Spaces como polígonos por kind (los marcos sin extensión, como el cuerpo de la giantess, como una cruz con su orientación); celdas bloqueadas y lentas; líneas de grilla si la celda mide ≥ 10 px. Portales cian (estáticos), magenta (de contacto) o gris (deshabilitados). Contactos como línea entre centros. Rutas en amarillo. Things como círculos (mínimo 3 px) coloreados por estado del mover. Con más de 40 things solo se rotula la elegida.

## Controles

Espacio pausa · N paso · 1/2/3 velocidad · Tab vista micro del Space elegido (o del Space de la entidad elegida) / macro · F encuadrar · rueda zoom · arrastrar o flechas para mover · click elegir · click derecho `move_to` de la entidad elegida a la superficie más alta bajo el cursor · G/P/C/R/L overlays · Esc soltar.

`-shot out.png -shot-tick N` corre a x16 hasta el tick N del escenario, guarda la pantalla y sale; con `-micro <nombre>` arranca en la vista micro de ese Space.

## Pendiente

- Vista lateral (Z): hoy todo es planta; las alturas se ven solo en el inspector.
- Los IDs del log se reemplazan por nombres; el inspector muestra IDs.
- Comandos desde la UI más allá de `move_to` (poses, `set_cells`, `set_portal`).
