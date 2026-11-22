# Spec: Visor de debug (M6)

Paquetes: `view` (Frame, Runner, picking), `camera` (proyección), `render` (dibujo), `ui` (paneles de texto), `cmd/game` (loop de Ebitengine).

## Frame (`view.Frame`)

Foto plana y serializable de un tick, armada por `Capturer.Capture(w)` (solo lee el mundo):

- `spaces`: id, nombre, kind, padre, profundidad, pose de mundo, extensión, copia de la grilla (cacheada por versión) y si es superficie de contacto. Los Spaces móviles pueden incluir el estado y la ruta del mover. Ordenados de abajo hacia arriba (Z de mundo, luego ID).
- Los Spaces con `agent` (la giantess) incluyen `activity` y `step`, igual que las Things.
- `things`: entidades con `location` que no son Spaces: pose de mundo, cuerpo, porteador para los objetos cargados y, si tienen `mover`, su estado y la ruta pendiente en puntos de mundo.
- `portals` (extremos en el mundo y el Space de cada uno; `dynamic` si los creó un contacto) y `contacts`.
- Los agentes incluyen sus necesidades (`needs`). `cues`: las señales de expresión de los últimos `CueTicks`, que completa el `Runner` a partir de eventos; `Emote(id)` y `Expression(id)` combinan señales y estado (ver [presentación](presentacion.md#interfaz-de-partida)).
- `falls`: las caídas (`entity_fell`) de los últimos `FallTicks`, que completa el `Runner` para animarlas. `FallOf(id)` devuelve la caída en curso de una entidad.

El renderer dibuja solo desde el Frame. `view.Inspect(w, id)` es la única lectura directa del mundo (el panel del inspector: JSON de cada componente vía `ecs.Registry.Components`).

## Tiempo (`view.Runner`)

- `Advance(dt)` acumula `dt × velocidad` y corre los `Step` que entran (máximo 64 por frame; el resto se descarta y la simulación se ralentiza). `StepOnce` es el paso a paso. Velocidades: x1, x4, x16.
- Guarda `Prev` y `Cur` (Frames antes y después del último Step) y `Alpha` (fracción del tick siguiente). `Interpolate(prev, cur, alpha)` interpola poses de Spaces y Things para dibujar a 60 FPS; saltos de más de 1 m no se interpolan.
- Guarda los últimos 200 eventos ya serializados y hasta 200 resúmenes de historia, cada uno con su actor y su lugar. Resume los eventos al publicarse, incluso en un frame con varios ticks simulados. `Hash()` se calcula una vez por tick.
- El paso de un escenario es `scenario.Player.Step` (encola los comandos del tick); pasado el final, el mundo sigue sin comandos y el HUD muestra el resultado de las expectativas.

**Garantía:** el visor solo decide *cuántas* veces llamar a `Step` y encola comandos del jugador. Cámara, velocidad, pausa, capturas, interpolación y picking no cambian el hash (`TestPresentationMatchesHeadless`, en `cmd/game`, la única prueba que compara con la ejecución headless).

## Cámara (`camera`)

Sigue a un Space ancla (`Nil` = mundo): el centro está en coordenadas locales del ancla y la vista gira con él. Vista macro = ancla en el mundo; vista micro = ancla en un Space (p. ej. la palma, que sigue a la mano). Las coordenadas pasan a float relativas al ancla. +Y local va hacia arriba en pantalla. `ZoomAt` mantiene el punto bajo el cursor; `Follow` cambia de ancla sin mover la vista.

## Dibujo

Spaces como polígonos por kind (los marcos sin extensión, como el cuerpo de la giantess, como una cruz con su orientación); celdas bloqueadas y lentas; líneas de grilla si la celda mide ≥ 10 px. Portales cian (estáticos), magenta (de contacto) o gris (deshabilitados). Contactos como línea entre centros. Rutas en amarillo. Things como círculos (mínimo 3 px) coloreados por estado del mover. Con más de 40 things solo se rotula la elegida.

Con `T`, el visor centra la cámara en la entidad seleccionada y usa su posición de mundo interpolada como origen. La vista mantiene su orientación aunque la entidad cambie de Space. Si la entidad desaparece, se conserva el último centro de la cámara.

## Menú principal y modos

Sin flags, `go run ./cmd/game` abre el menú principal (`cmd/game/menu.go`, dibujado con `ui.Menu`) en el lienzo de pixel art:

- **Nueva partida**: la mañana en la casa con la colonia sorteada por `scenario.Randomize` (de 8 a 20 tinies en celdas libres del escritorio cerca del nido, con color, necesidades y carácter al azar; sin expectativas). La semilla queda en el título de la ventana; `-new-game -seed N` repite esa partida sin pasar por el menú.
- **Escenarios**: los de `tests/scenarios` que declaran `atlas` (los que se juegan con la interfaz de partida). Los demás son pruebas del motor y solo se abren con `-scenario`.
- **Modo debug**: arranca las partidas con los overlays técnicos (`F3`) sobre el arte. `-debug` lo deja encendido al abrir el menú.

En una partida abierta desde el menú, `Esc` sin nada elegido ni seguido pausa y pregunta si volver (`Enter` vuelve, `Esc` sigue jugando). El mezclador de sonido es uno solo por proceso: al volver, `Mixer.Hush` apaga el ambiente.

El pixel art es el modo por defecto también con `-scenario`, `-load`, `-crowd` y `-new-game`. Si el mundo no tiene visuales (los escenarios sin arte, `-crowd`, `-load`), o con `-debug-view`, se abre el visor de debug vectorial que describe esta spec. `-art` se acepta por compatibilidad y no hace nada.

## Controles

Espacio pausa · N paso · 1/2/3 velocidad · Tab vista micro del Space elegido (o del Space de la entidad elegida) / macro · F encuadrar · rueda zoom · arrastrar, flechas o WASD para mover · click elegir · click derecho `move_to` de la entidad elegida a la superficie más alta bajo el cursor · G/P/C/R/L overlays · Esc soltar.

`T` sigue la entidad seleccionada o libera la cámara si ya la sigue. `H` alterna entre el log y la historia. `RePág` y `AvPág` recorren la historia. «Una mañana en la casa», la partida al azar y lo que se abre desde el menú arrancan en modo historia. `-story` lo activa al iniciar con otros escenarios.

`-shot out.png -shot-tick N` corre a x16 hasta el tick N del escenario, guarda la pantalla y sale; con `-micro <nombre>` arranca en la vista micro de ese Space y con `-follow <nombre>` siguiendo a esa entidad; `-pan dx,dy` desplaza el encuadre inicial. `-bench 20s` mide la presentación sin vsync y sale. `M` silencia el sonido; `-mute` arranca en silencio y `-volume` fija el volumen general; `-mix out.wav` renderiza la mezcla de referencia sin pantalla (ver [sonido](sonido.md)). En pixel art, `F3` o `-debug` superponen al arte rutas, portales y contactos.

## Pendiente

- Vista lateral (Z): hoy todo es planta; las alturas se ven solo en el inspector.
- Los IDs del log se reemplazan por nombres; el inspector muestra IDs.
- Comandos desde la UI más allá de `move_to` (poses, `set_cells`, `set_portal`).
