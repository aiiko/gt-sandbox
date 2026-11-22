# Plan de desarrollo

**Una mañana en la casa** está cerrada en imagen y sonido. El [cierre visual](specs/presentacion.md#cierre-de-la-entrega-visual) reúne las capturas de referencia, la prueba de continuidad y las mediciones de rendimiento. La [spec de sonido](specs/sonido.md) cubre el sonido sintetizado y sin música (M20 a M22): la base de audio, los sonidos de la mañana con perspectiva por escala, el ambiente por habitación y la mezcla de referencia. El [escenario de referencia](specs/manana-en-la-casa.md) fija el recorrido con el que se integró cada hito.

## Hitos pendientes

El slice **Un descuido en el escritorio** suma la muerte por aplastamiento: los tinies pueden morir bajo la palma, una pisada o un objeto que cae. El diseño está en la [propuesta](propuesta.md#14-próximo-slice-un-descuido-en-el-escritorio). Cada hito abre o amplía la spec `docs/specs/crush.md`.

| Hito | Contenido | Cierre |
| --- | --- | --- |
| M23 Aplastamiento | Paquete `sim/crush` con `crusher`, `fragile`, `shelter` y `remains`. Aplastan la palma al apoyarse (contacto nuevo de `attach`) y los objetos al aterrizar (`space.falls`/`space.Land`). La huella de la palma sale de su grilla. Muerte como despawn con restos como entidad nueva. Trabajos y reservas se liberan por los caminos existentes. `crushed`. Incrementar `HashVersion` y `SnapshotSchema` | `palm_crushes_tiny`, `tiny_on_palm_is_safe`, `tiny_between_fingers_survives`, `nest_shelters_from_palm`, `cup_lands_on_tiny`, `haul_requeued_after_death`. La mañana conserva su aceptación |
| M24 Pisadas | `stride` y pisadas discretas derivadas del `mover` (`crush.steps`, `footstep`). La casa suma el portal por la pata del escritorio al tatami y migajas que pueden caer al piso | `step_crushes_tiny_on_floor`, `step_spares_tiny_on_desk` |
| M25 Amenaza y huida | Zonas de amenaza como recurso (`crush.threats`). Sombra y vibración en `ai.look`. Reevaluación inmediata con `agent.reflex`. Acción `flee` por encima de `UrgentUtility`, con `nav.NearestFree`. `near_miss` y `trapped` | `tiny_flees_shadow` (se salva desde el borde y muere en el centro), `tiny_drops_crumb_to_flee`, `tiny_flees_footsteps` |
| M26 Cuidado y consecuencias | `agent.care`: demora del paso `pose` y tirada de atención (`ai.attention`) con espera acotada. `member_lost` y zonas `danger` que evitan `forage`, `explore` y `WanderPoint` hasta expirar | `careful_giantess_waits`, `colony_avoids_danger`, `danger_expires` |
| M27 Presentación y cierre | `Frame` con amenazas, pisadas, restos y estados. Sprites de tiny aplastado y restos, sombra de pisada, historia con causa, sonidos («plaf», «tum», susto). Escenario `un_descuido_en_el_escritorio` con giantess distraída y cuidadosa | Aceptación de la [propuesta](propuesta.md#147-alcance-y-hitos) sobre las semillas de referencia, capturas y mezcla de referencia actualizadas |

## Contratos de integración

- La presentación lee `view.Frame` y los eventos publicados. Puede ampliar los datos que captura `view` para resolver apariencia, pose u oclusión. El estado lógico se modifica mediante comandos.
- El formato del atlas, el escalado del lienzo y las referencias visuales están en la [spec de presentación](specs/presentacion.md).
- Cámara, filtros, selección, animación y transiciones pertenecen a la presentación. No afectan rutas, portales, tiempos lógicos ni snapshots. Se mantienen [las reglas del motor](../CLAUDE.md) y [los contratos de simulación](specs/nucleo.md).
- Cada hito cierra con una prueba reproducible sobre el escenario de referencia, una captura cuando cambia el dibujo y la spec actualizada cuando cambia un contrato. Ejecutar `go test ./...` y `go vet ./...` antes del cierre.

El alcance de la mañana termina en las tres habitaciones y el viaje en la palma. El slice de aplastamiento se activa por contenido (`fragile`), así que la mañana no cambia. Quedan fuera las heridas, las caídas letales, el aturdimiento sin muerte, sentarse o acostarse sobre tinies, las alarmas entre tinies y las relaciones entre la giantess y la colonia. El sonido es solo presentación: sale de eventos y del `Frame` y nunca cambia la simulación. La música queda para una etapa posterior. Las superficies corporales adicionales, la construcción de colonia, el LOD de simulación y la adopción de editores externos quedan para etapas posteriores.
