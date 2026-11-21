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

## Controles

- `M` silencia o devuelve el sonido; al volver suena el clic.
- `-mute` arranca en silencio.
- `-volume 0.5` fija el volumen general.

`-shot` y `-bench` no abren el dispositivo de audio.
