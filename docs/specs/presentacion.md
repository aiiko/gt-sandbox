# Base de presentación

`go run ./cmd/game -art` abre la prueba de arte del escritorio. Usa el mismo runner, simulación y cámara del visor. El escenario asigna un `visual` por entidad. El código de presentación vincula ese identificador al ID generacional que resolvió el escenario. El vínculo queda fuera del estado de simulación, el hash y los snapshots.

La presentación dibuja en un lienzo de 480 × 270 píxeles. Usa el mayor múltiplo entero disponible en la pantalla final y centra el lienzo con bandas. El cursor usa la transformación inversa de ese mismo rectángulo. Las bandas quedan fuera de la escena.

Cada variante se dibuja a escala nativa con el pivote en la posición de mundo interpolada, redondeada a un píxel del lienzo. El hit test recorre los sprites visibles en orden inverso al dibujo y consulta el alfa del mismo cuadro dibujado. El modo de prueba usa densidades de 16 px por 40 cm en macro y 16 px por 10 mm en micro.

El manifiesto JSON declara el PNG y las variantes por escala de cada sprite. Cada variante declara un pivote `[x, y]` relativo al cuadro y clips con duración en ticks y cuadros `[x, y, ancho, alto]` en el atlas. El clip `idle` es obligatorio. La carga identifica el manifiesto y el sprite, variante, clip o cuadro ante una exportación inválida.

## La casa en macro

El escenario de «Una mañana en la casa» enlaza su atlas por la ruta `atlas`, relativa al archivo del escenario. El manifiesto de la casa declara `scales: ["macro"]`. Un atlas sin esa lista ofrece las dos escalas. `-atlas` permite reemplazar el manifiesto elegido por el escenario.

El dibujo macro coloca primero las variantes `layer: "background"`, después los umbrales y detalles del suelo de `layer: "ground"`, luego las piezas por su línea de apoyo en pantalla y por último su `foreground` opcional. El primer plano usa el mismo pivote que el cuadro base. `depth_offset` desplaza la línea de apoyo vertical en píxeles. Los ocupantes de un mueble comparten su orden de profundidad y se dibujan encima del mueble. El picking usa el orden final y el alfa de cada capa.

Los sprites que ofrecen clips `idle_north` o `walk_north` y sus variantes para las otras tres direcciones los eligen a partir de la orientación o el tramo activo de la ruta. Las variantes sin esos clips usan `idle` y, si está disponible, `walk`. Las puertas se dibujan en el punto medio de los portales estáticos asociados a un visual.

La casa se abre con este comando. Hacé clic en la giantess para elegirla y pulsá `T` para seguirla. `F` recupera el encuadre completo. `Espacio` pausa y `1/2/3` elige la velocidad.

```sh
go run ./cmd/game -art -scenario tests/scenarios/una_manana_en_la_casa.json
```

![Vista macro de la casa con la giantess en la cocina](img/m13-casa.png)

La prueba del visor verifica el picking sobre la giantess y los umbrales, el seguimiento del dormitorio a la cocina y de regreso al dormitorio antes de llegar al living, y la igualdad del hash con la ejecución headless.

## Prueba de arte

`Tab` alterna entre la vista del estudio y la superficie seleccionada. El foco inicial de micro es el escritorio. El clic elige un sprite, `T` sigue la selección y `F` vuelve a centrar. Las dos vistas usan la misma taza y el mismo tiny. El escritorio y la palma representan madera y piel con la densidad de cada vista.

| Macro | Micro |
| --- | --- |
| ![Estudio con escritorio en macro](img/m12-macro.png) | ![Escritorio con taza, palma y tiny en micro](img/m12-micro.png) |

Las referencias muestran la [dirección anime](../propuesta.md#11-dirección-visual) en el tick inicial del [escenario de prueba](../../tests/scenarios/muestra_pixel_art.json), a 1×. El cárdigan lila vincula el sprite macro de la giantess con la manga de la vista micro. Se puede cargar otra exportación con `-atlas ruta/al/manifiesto.json` o usar `-size` para revisar otros tamaños de ventana.

```sh
go run ./cmd/game -art -size 480x270 -shot macro.png
go run ./cmd/game -art -micro Escritorio -size 480x270 -shot micro.png
```
