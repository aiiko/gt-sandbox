# Sonido

El sonido es presentación, como los símbolos de expresión (ver la [propuesta](../propuesta.md#sonido)). Sale del estado de la interfaz, de los eventos y del `Frame`. Nunca lee ni modifica el mundo, y el hash no cambia (`TestPresentationMatchesHeadless` recorre la mañana con el mezclador observando la interfaz en cada frame).

## Sonidos sintetizados

Los sonidos se generan por código con `tools/audio/build.py`, que solo usa la biblioteca estándar de Python:

```sh
python tools/audio/build.py          # escribe content/audio/*.wav y sounds.json
python tools/audio/build.py -check   # verifica que lo exportado coincide
```

Los parámetros están en `tools/audio/sounds.json`. Cada sonido es una suma de capas, y cada capa define estos parámetros:

- forma de onda: `sine`, `square`, `triangle`, `saw` o `noise`;
- barrido de frecuencia exponencial (`f0` → `f1`);
- retardo y duración;
- ataque lineal y caída exponencial;
- ganancia;
- pasabajos de un polo (`lp`);
- ancho de pulso;
- vibrato.

El ruido sale de un generador congruencial propio, así que la salida es reproducible. Los WAV son PCM mono de 16 bits a 44,1 kHz. Para cambiar un sonido se editan los parámetros y se regenera; los WAV no se tocan a mano.

El manifiesto exportado, `content/audio/sounds.json`, describe cada sonido con estos campos:

| Campo | Contenido |
| --- | --- |
| `id` | Identificador estable |
| `file` | Archivo WAV |
| `category` | `ambiente`, `efectos` o `interfaz` |
| `ms` | Duración |
| `voices` | Copias simultáneas como máximo |
| `cooldown_ms` | Tiempo mínimo entre comienzos |

`TestManifestMatchesFiles` (en `internal/sound`) verifica que los WAV existen, tienen el formato correcto y duran lo declarado.

## Capas

- **`internal/sound`** decide qué suena, sin reproducir nada y sin importar Ebitengine. `archtest` verifica que `core`, `sim`, `content`, `scenario`, `view` y `camera` no lo importan.
  - `Library` carga el manifiesto.
  - `Mixer` aplica las reglas de mezcla y pasa los pedidos a un `Backend`.
  - `UIWatcher` traduce los cambios de la interfaz en sonidos.
  - `Scene` traduce los eventos publicados y los cambios entre frames en sonidos de la escena.
- **`cmd/game/audio.go`** es el `Backend` real. Usa el paquete `audio` de Ebitengine, decodifica cada WAV una vez y aplica el panorama copiando el PCM con una ganancia por canal. Las pruebas usan un grabador en su lugar.

## Mezclador

`Mixer.Play(id, volumen, panorama)` acepta un pedido solo si se cumplen todas estas condiciones:

- el sonido no está silenciado;
- no se superan las voces del sonido;
- pasó su tiempo mínimo entre comienzos;
- el volumen final no es cero.

El volumen final es el volumen pedido × el general (`-volume`) × el de la categoría (ambiente 0,5, efectos 0,8, interfaz 0,6). El tiempo se cuenta en frames de presentación (`Advance`, una vez por `Update`), nunca con el reloj real. Silenciar corta lo que está sonando.

## Sonidos de interfaz

`UIWatcher` compara en cada frame el estado de la interfaz (`UIState`) con el del frame anterior. Por eso suena igual sin importar si el cambio vino del teclado, el mouse o la historia. El primer estado solo se registra.

| Cambio | Sonido | Qué es |
| --- | --- | --- |
| Elegir algo | `ui_select` | Clic corto |
| Empezar a seguir (en lugar del clic) | `ui_follow` | Campanita de dos notas |
| Pasar a micro / volver a macro | `ui_zoom_in` / `ui_zoom_out` | Barrido que sube o baja, junto con el fundido |
| Pausar / reanudar | `ui_pause` / `ui_resume` | Cinta que frena o arranca |
| Entrada nueva con la historia visible | `ui_story` | Hoja que pasa |

`TestInterfaceSounds` recorre cada acción, más el silencio con `M`.

## Sonidos de la escena

`Scene` recibe tres cosas:

- el `Frame` actual;
- los eventos publicados con su tipo (`Runner.Events`, cada uno con un `Seq` creciente, así que se distinguen los nuevos aunque pasen varios ticks en un frame);
- un `View`, con la escala, la superficie enfocada, la velocidad, dónde cae cada punto en pantalla (`Locate`, que da el panorama) y de quién se dibuja el símbolo.

Compara el frame con el anterior que observó y devuelve los pedidos. La primera observación solo registra el estado, y un frame sin tick nuevo no suena.

**Perspectiva por escala.** En macro suenan la giantess y la casa, y los tinies no se oyen salvo el símbolo de lo elegido o seguido. En micro suenan los tinies, la mano de cerca y la señal de paso.

| Momento | Fuente | Macro | Micro |
| --- | --- | --- | --- |
| La giantess camina | su pose cambia, cada 10 ticks | pasos según el suelo: tatami, madera, baldosa | — |
| Pasa una puerta | `entity_reparented` por portal | fusuma o noren (cocina); al salir de la cocina, la heladera se cierra | — |
| Llega a la heladera | `move_arrived` en la cocina | la heladera se abre | — |
| Desayuna | actividad `use:breakfast` en `use`, cada 30 ticks | mordiscos | — |
| Llega al living | `move_arrived` en el living | zabuton y la TV se enciende | — |
| Deja la taza | `item_dropped` | tintineo | tintineo, si cae en la superficie enfocada |
| La mano se acerca | `pose_changed` a `hand_hover` o `set_table`, sobre la superficie enfocada | — | soplo que crece |
| La mano se apoya | `contact_began` | toque suave | golpe grave, si toca la superficie enfocada |
| La mano se levanta | `contact_ended` | — | soplo hacia arriba |
| Se abre o cierra el paso | `portal_opened` / `portal_closed` en la superficie enfocada | — | dos notas que suben / bajan |
| Un tiny camina | su pose cambia, cada 8 ticks desfasados por tiny | — | pasitos, una variante por paleta |
| Levanta una migaja | `carried_by` aparece | — | pop |
| Una migaja cae | `entity_fell` de una migaja | — | tic |
| Sube o baja de la mano | `entity_reparented` por portal, desde o hacia la superficie enfocada | — | saltito |
| Cae | `entity_fell` de un tiny | — | silbido que baja |
| Aterriza | al terminar la animación de caída (`FallTicks`) | — | «puf» |
| El nido guarda | `stored` | — | campanita |
| Aparece un símbolo | `Frame.Emote` cambia, si se dibuja | sí (giantess, elegido, seguido) | sí (todos los tinies) |

**Fuera de cuadro.** Lo que no se ve no suena (`Locate` deja un margen del 10 %). La mano es enorme, así que sus sonidos, la señal de paso, el saltito y la taza se oyen cuando tocan la superficie enfocada aunque el punto quede fuera de cuadro. El panorama abre hasta 0,7 hacia cada lado.

**Densidad.** Cada sonido tiene sus voces y su tiempo mínimo entre comienzos. Por ejemplo, cada variante de pasitos tiene 2 voces y 110 ms, así que veinte tinies no saturan. A x4 y x16 se omiten los sonidos que se repiten: pasos, mordiscos, pasitos, migajas, «?», «zzz» y la nota. Quedan la mano, las caídas, los descubrimientos y las puertas. En pausa no avanza ningún tick, así que la escena no suena.

**Pruebas.**

- `TestMorningSounds` recorre la mañana en macro, en el micro del escritorio y siguiendo a Tiny 01, más el escenario de caída. Verifica qué suena y qué no en cada escala, y que cada evento con sonido suene en el mismo frame en que se publica. También verifica el orden: la mano se acerca antes de apoyarse, y el descubrimiento llega antes del ascenso. Comprueba el viaje (levantar, apoyar en el living, bajar), que el aterrizaje llegue `FallTicks − 1` ticks después de la caída, y que los pasitos respeten su límite.
- `TestFastAndPausedSounds` cubre las velocidades rápidas y la pausa.

## Controles

- `M` silencia o devuelve el sonido; al volver suena el clic.
- `-mute` arranca en silencio.
- `-volume 0.5` fija el volumen general.

`-shot` y `-bench` no abren el dispositivo de audio.
