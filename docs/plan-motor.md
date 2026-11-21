# Plan de desarrollo

La próxima entrega es **Una mañana en la casa en pixel art top-down**. La [propuesta visual](propuesta.md#11-dirección-visual) define la cámara, las dos escalas, el arte y la interfaz. El [escenario de referencia](specs/manana-en-la-casa.md) fija el recorrido con el que se integra cada hito.

## Hitos pendientes

### M13 — La casa en macro

- Producir tiles y sprites para dormitorio, cocina y living. Dibujar suelo, paredes, puertas, cama, escritorio, heladera, mesa y televisor.
- Representar a la giantess con su sprite direccional y a la colonia con el marcador del nido. Añadir profundidad por punto de apoyo y tratamiento de paredes de primer plano.
- Encuadrar la casa y permitir seleccionar y seguir a la giantess durante el recorrido.

**Aceptación.** En una ejecución del escenario se distinguen las tres habitaciones y sus conexiones, se identifica la ida a la cocina y el regreso al escritorio, y la giantess puede seguirse hasta el living con las superposiciones coherentes.

### M14 — Superficies y colonia en micro

- Producir el arte de escritorio, nido, palma, mesa del living, taza, migajas y tinies. Dibujar superficies con su contorno útil y detalles ajustados a la vista micro.
- Mostrar el tiny, su carga y su ubicación con puntos de apoyo compartidos con la escena macro. Resolver oclusión en bordes y conservar selección y seguimiento al cambiar de vista.
- Incorporar el cambio macro/micro y la vuelta al encuadre anterior, con fundido breve y una opción de corte directo.

**Aceptación.** Al enfocar el escritorio se distinguen nido, recursos y tinies sin overlays técnicos. La misma taza aparece en el lugar correcto en ambas vistas. Cambiar de escala conserva la identidad del tiny seleccionado y su punto de selección coincide con su posición en pantalla.

### M15 — Animación, contacto y viaje

- Vincular poses, caminata, reposo, comer y acarreo al estado que recibe la presentación. Alinear manga, palma y carga con los anclajes de cada vista.
- Mostrar cuándo el paso entre mano y mueble está disponible. Diferenciar mano elevada y apoyada mediante silueta, sombra y señal del paso.
- Mantener el seguimiento micro durante el viaje y actualizar el contexto cuando el tiny baja de la mano.

**Aceptación.** Se puede seguir visualmente al mismo tiny al subir del escritorio a la mano, viajar en la palma y bajar en el living. La señal del paso coincide con el portal real en ambos apoyos y durante el trayecto. La taza y los recursos cambian de estado al ocurrir su evento.

### M16 — Experiencia de partida

- Preservar un modo de debug e integrar una interfaz de partida con reloj, pausa, velocidad, cambio de escala, selección, seguimiento e inspector resumido.
- Presentar el panel de historia y permitir enfocar desde una entrada al actor o la superficie disponibles. Exponer las órdenes de destino existentes donde correspondan.
- Hacer legibles los estados con texto o forma además del color y mostrar los controles en la interfaz.

**Aceptación.** Se abre la escena en la interfaz de partida y se puede observar la secuencia completa, cambiar de escala, pausar, seguir a un tiny y volver al encuadre general. El modo de debug sigue permitiendo inspeccionar rutas, portales y contactos.

### M17 — Cierre de la entrega visual

- Conservar capturas de referencia de la casa, las migajas en el escritorio, la mano apoyada, el viaje y la llegada al living. Registrar semilla, vista, encuadre y tick para reproducir cada captura.
- Comparar el hash al mismo tick con una ejecución headless al alternar vistas, seguimiento y velocidad. Reejecutar la aceptación de cinco semillas y la continuación desde snapshots.
- Medir simulación y renderer por separado con el escenario completo. Registrar resolución y equipo de referencia. La meta es menos de 2 ms/tick de simulación y 60 FPS de presentación a 1920 × 1080.

**Aceptación.** El recorrido de principio a fin se entiende en macro y micro. Pasan los escenarios, se conserva la continuidad de la simulación y las capturas verifican nitidez, orden de dibujo y correspondencia entre animación y contacto.

## Contratos de integración

- La presentación lee `view.Frame` y los eventos publicados. Puede ampliar los datos que captura `view` para resolver apariencia, pose u oclusión. El estado lógico se modifica mediante comandos.
- El formato del atlas, el escalado del lienzo y las referencias visuales están en la [spec de presentación](specs/presentacion.md).
- Cámara, filtros, selección, animación y transiciones pertenecen a la presentación. No afectan rutas, portales, tiempos lógicos ni snapshots. Se mantienen [las reglas del motor](../CLAUDE.md) y [los contratos de simulación](specs/nucleo.md).
- Cada hito cierra con una prueba reproducible sobre el escenario de referencia, una captura cuando cambia el dibujo y la spec actualizada cuando cambia un contrato. Ejecutar `go test ./...` y `go vet ./...` antes del cierre.

El alcance de esta etapa termina en las tres habitaciones y el viaje en la palma. Las superficies corporales adicionales, la construcción de colonia, el LOD de simulación y la adopción de editores externos quedan para etapas posteriores.
