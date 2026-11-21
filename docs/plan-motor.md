# Plan de desarrollo

La próxima entrega es **Una mañana en la casa en pixel art top-down**. La [propuesta visual](propuesta.md#11-dirección-visual) define la cámara, las dos escalas, el arte y la interfaz. El [escenario de referencia](specs/manana-en-la-casa.md) fija el recorrido con el que se integra cada hito.

## Hitos pendientes

### M15 — Fuentes de arte y ambientación japonesa

- Pasar el arte a fuentes editables por pieza, versionadas junto al atlas, con un comando reproducible que arme `casa.png` y `casa.json`. El atlas exportado deja de editarse a mano.
- Redibujar la casa con la [ambientación japonesa](propuesta.md#ambientación): washitsu con tatami y futón, puertas corredizas, cocina compacta y living de madera. Los muebles conservan posición, altura y ofertas.
- Producir el lote que piden los hitos siguientes: tiny con reposo, caminata y acarreo en cuatro direcciones y caída; giantess sentada comiendo, con la mano apoyada y mirando TV; palma elevada y apoyada; taza en mano y apoyada.
- Agregar un test que recorra los `visual` del escenario y del contenido y verifique que el atlas tiene las variantes y clips que la presentación va a pedir para cada uno.

**Aceptación.** El comando de exportación reproduce el atlas versionado byte a byte. Las capturas de la casa y del escritorio muestran la nueva ambientación con la paleta de la propuesta. El test de cobertura del atlas pasa y el hash del escenario no cambia.

### M16 — Animación y poses

- Vincular poses, caminata, reposo, comer y acarreo al estado que recibe la presentación. Elegir la dirección del tiny por la ruta igual que en la giantess.
- Alinear manga, palma y carga con los anclajes de cada vista. La taza y las migajas cambian de representación al ocurrir su evento.

**Aceptación.** En macro se reconoce cuándo la giantess camina, come sentada, apoya la mano y mira TV. En micro el tiny camina y acarrea en la dirección de su ruta con la carga alineada. Las capturas por tick coinciden con los eventos del log.

### M17 — Contacto, viaje y caídas

- Mostrar cuándo el paso entre mano y mueble está disponible. Diferenciar mano elevada y apoyada mediante silueta, sombra y señal del paso. La sombra de la mano anticipa su llegada a la superficie.
- Mantener el seguimiento micro durante el viaje y actualizar el contexto cuando el tiny baja de la mano.
- Mostrar las caídas de `space.falls`: el tiny que cae conserva su marcador hasta apoyarse y el seguimiento lo acompaña a la superficie de destino.

**Aceptación.** Se puede seguir visualmente al mismo tiny al subir del escritorio a la mano, viajar en la palma y bajar en el living. La señal del paso coincide con el portal real en ambos apoyos y durante el trayecto. Un escenario de prueba con un tiny que cae de la palma muestra la caída y mantiene el seguimiento.

### M18 — Experiencia de partida

- Preservar un modo de debug e integrar una interfaz de partida con reloj, pausa, velocidad, cambio de escala, selección, seguimiento e inspector resumido, con el estilo de interfaz de la propuesta.
- Presentar el panel de historia con el retrato de la giantess y permitir enfocar desde una entrada al actor o la superficie disponibles. Exponer las órdenes de destino existentes donde correspondan.
- Dibujar los [símbolos de expresión](propuesta.md#expresiones-y-señales) a partir de eventos y estado, y cambiar la expresión del retrato según necesidades y acción.
- Hacer legibles los estados con texto o forma además del color, mostrar los controles en la interfaz y usar el rótulo jerárquico de ubicación («Dormitorio › Escritorio»).

**Aceptación.** Se abre la escena en la interfaz de partida y se puede observar la secuencia completa, cambiar de escala, pausar, seguir a un tiny y volver al encuadre general. El descubrimiento de las migajas, el ascenso a la palma y el descanso se ven con su símbolo. El modo de debug sigue permitiendo inspeccionar rutas, portales y contactos.

### M19 — Cierre de la entrega visual

- Conservar capturas de referencia de la casa, las migajas y la taza en el escritorio, la mano apoyada, el viaje y la llegada al living. Registrar semilla, vista, encuadre y tick para reproducir cada captura.
- Reunir en una sola prueba las comparaciones de hash con la ejecución headless que hoy reparten los tests de velocidad, vistas y seguimiento. Reejecutar la aceptación de cinco semillas y la continuación desde snapshots.
- Medir simulación y renderer por separado con el escenario completo y con `-crowd 200`. Registrar resolución y equipo de referencia. La meta es menos de 2 ms/tick de simulación con 200 tinies y 60 FPS de presentación a 1920 × 1080.

**Aceptación.** El recorrido de principio a fin se entiende en macro y micro. Pasan los escenarios, se conserva la continuidad de la simulación y las capturas verifican nitidez, orden de dibujo y correspondencia entre animación y contacto.

## Contratos de integración

- La presentación lee `view.Frame` y los eventos publicados. Puede ampliar los datos que captura `view` para resolver apariencia, pose u oclusión. El estado lógico se modifica mediante comandos.
- El formato del atlas, el escalado del lienzo y las referencias visuales están en la [spec de presentación](specs/presentacion.md).
- Cámara, filtros, selección, animación y transiciones pertenecen a la presentación. No afectan rutas, portales, tiempos lógicos ni snapshots. Se mantienen [las reglas del motor](../CLAUDE.md) y [los contratos de simulación](specs/nucleo.md).
- Cada hito cierra con una prueba reproducible sobre el escenario de referencia, una captura cuando cambia el dibujo y la spec actualizada cuando cambia un contrato. Ejecutar `go test ./...` y `go vet ./...` antes del cierre.

El alcance de esta etapa termina en las tres habitaciones y el viaje en la palma. El audio queda para después de la entrega visual. Las superficies corporales adicionales, la construcción de colonia, el LOD de simulación y la adopción de editores externos quedan para etapas posteriores.
