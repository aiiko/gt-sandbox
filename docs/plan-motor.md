# Plan de desarrollo

**Una mañana en la casa** está cerrada en imagen y sonido. El [cierre visual](specs/presentacion.md#cierre-de-la-entrega-visual) reúne las capturas de referencia, la prueba de continuidad y las mediciones de rendimiento. La [spec de sonido](specs/sonido.md) cubre el sonido sintetizado y sin música (M20 a M22): la base de audio, los sonidos de la mañana con perspectiva por escala, el ambiente por habitación y la mezcla de referencia. El [escenario de referencia](specs/manana-en-la-casa.md) fija el recorrido con el que se integró cada hito.

## Hitos pendientes

No hay hitos pendientes. Los próximos saldrán del [alcance posterior](propuesta.md#13-alcance-posterior).

## Contratos de integración

- La presentación lee `view.Frame` y los eventos publicados. Puede ampliar los datos que captura `view` para resolver apariencia, pose u oclusión. El estado lógico se modifica mediante comandos.
- El formato del atlas, el escalado del lienzo y las referencias visuales están en la [spec de presentación](specs/presentacion.md).
- Cámara, filtros, selección, animación y transiciones pertenecen a la presentación. No afectan rutas, portales, tiempos lógicos ni snapshots. Se mantienen [las reglas del motor](../CLAUDE.md) y [los contratos de simulación](specs/nucleo.md).
- Cada hito cierra con una prueba reproducible sobre el escenario de referencia, una captura cuando cambia el dibujo y la spec actualizada cuando cambia un contrato. Ejecutar `go test ./...` y `go vet ./...` antes del cierre.

El alcance de esta etapa termina en las tres habitaciones y el viaje en la palma. El sonido es solo presentación: sale de eventos y del `Frame` y nunca cambia la simulación. La música queda para una etapa posterior. Las superficies corporales adicionales, la construcción de colonia, el LOD de simulación y la adopción de editores externos quedan para etapas posteriores.
