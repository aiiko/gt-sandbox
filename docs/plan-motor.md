# Plan de desarrollo

La entrega visual de **Una mañana en la casa** está cerrada; el [cierre](specs/presentacion.md#cierre-de-la-entrega-visual) reúne sus capturas de referencia, la prueba de continuidad y las mediciones de rendimiento. La próxima entrega le suma **sonido sintetizado, sin música**, según la [propuesta de sonido](propuesta.md#sonido). El [escenario de referencia](specs/manana-en-la-casa.md) sigue fijando el recorrido con el que se integra cada hito.

## Hitos pendientes

### M20 — Base de audio

- Sintetizador en `tools/audio/` que lee los parámetros versionados de cada sonido y escribe WAV cortos en `content/audio/`, con un manifiesto de identificadores estables y un modo `-check` que verifica que lo exportado coincide con las fuentes, como el arte.
- Paquete `internal/sound`, sin Ebitengine: traduce eventos, `Frame`, vista y cámara en pedidos de sonido (identificador, categoría, volumen, panorama, prioridad). `archtest` verifica que `sim` y `view` no lo importan.
- Reproducción en `cmd/game` con el paquete `audio` de Ebitengine detrás de una interfaz, con un backend nulo para los tests. Categorías ambiente, efectos e interfaz con volumen propio, silencio con `M` y `-mute` al arrancar.
- Primeros sonidos: los de la interfaz (elegir, seguir, cambio de escala con el fundido, pausa, entrada nueva en la historia).

**Aceptación.** Las acciones de la interfaz suenan y `M` silencia. Recorrer la mañana con sonido mantiene el hash de la ejecución headless (`TestPresentationMatchesHeadless`). El sintetizador regenera los WAV de forma reproducible.

### M21 — Sonidos de la mañana

- Sonidos de la giantess, la mano, los tinies, los símbolos de expresión y la señal de paso, con perspectiva por escala: el mismo evento elige otro sonido o volumen en macro y en micro.
- Panorama y volumen según la posición en pantalla; silencio fuera de cuadro.
- Límite de voces por sonido, tiempo mínimo entre repeticiones y variación leve de tono. A x4 y x16 se omiten los sonidos menores; la pausa silencia los efectos.

**Aceptación.** Una prueba recorre la mañana con el backend nulo y verifica que cada momento clave dispara su sonido en el tick correcto: el desayuno, el descubrimiento de las migajas, la mano apoyada, el ascenso, el viaje, la llegada al living y la caída. Con veinte tinies a la vez no se supera el límite de voces.

### M22 — Ambiente y mezcla

- Ambiente por habitación (dormitorio, cocina, living) y tono de sala en micro, con fundidos al cambiar de habitación seguida o de escala. La TV encendida suma su murmullo al living.
- Mezcla de referencia: un comando renderiza la mañana completa a un WAV sin pantalla, a partir de los mismos pedidos de sonido. Un resumen de esa mezcla (pedidos por tick) queda registrado como las capturas de referencia.
- Indicador de silencio y volumen en la interfaz de partida.

**Aceptación.** La mañana completa se escucha con ambiente continuo y sin cortes al cambiar de vista. La mezcla de referencia es reproducible y su resumen coincide con el registrado. Pasan los escenarios y se conserva la continuidad de la simulación.

## Contratos de integración

- La presentación lee `view.Frame` y los eventos publicados. Puede ampliar los datos que captura `view` para resolver apariencia, pose u oclusión. El estado lógico se modifica mediante comandos.
- El formato del atlas, el escalado del lienzo y las referencias visuales están en la [spec de presentación](specs/presentacion.md).
- Cámara, filtros, selección, animación y transiciones pertenecen a la presentación. No afectan rutas, portales, tiempos lógicos ni snapshots. Se mantienen [las reglas del motor](../CLAUDE.md) y [los contratos de simulación](specs/nucleo.md).
- Cada hito cierra con una prueba reproducible sobre el escenario de referencia, una captura cuando cambia el dibujo y la spec actualizada cuando cambia un contrato. Ejecutar `go test ./...` y `go vet ./...` antes del cierre.

El alcance de esta etapa termina en las tres habitaciones y el viaje en la palma. El sonido es solo presentación: sale de eventos y del `Frame` y nunca cambia la simulación. La música queda para una etapa posterior. Las superficies corporales adicionales, la construcción de colonia, el LOD de simulación y la adopción de editores externos quedan para etapas posteriores.
