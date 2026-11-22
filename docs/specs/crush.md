# Spec: Aplastamiento (M23–M27)

Paquete: `sim/crush`. Diseño y alcance del slice en la [propuesta](../propuesta.md#14-un-descuido-en-el-escritorio).

Lo que baja sobre una superficie mata a los tinies frágiles que quedan debajo, salvo que estén en un refugio que resista. El aplastamiento es instantáneo: no hay daño parcial y las caídas no matan.

## Componentes

| Store | Datos | De quién |
| --- | --- | --- |
| `crusher` | `{source, force, thickness, radius, pressing}` | Aplastador: la palma (`source: "hand"`), un objeto (`"object"`). `pressing` es estado |
| `fragile` | `{}` | Quien muere aplastado (los tinies) |
| `shelter` | `{strength, radius}` | Refugio: el nido (`content/objects/nest.json`, fuerza 50, 2 cm) |
| `remains` | `{name, of, by, source, tick, colony}` | Restos de un aplastado. `colony` la completa `colony.mourn` |
| `stride` | `{length, side, foot_length, foot_width, force, travel, right, space, last}` | Quien pisa al caminar (la giantess). Desde `travel` es estado |

El aplastamiento se activa por contenido. Un escenario sin `fragile` no cambia: la mañana en la casa sigue igual.

## Cuándo aplasta

`crush.resolve` corre en Consequences antes de `space.falls` y `jobs.upkeep`.

**Un Space aplastador** (la palma) se apoya sobre una superficie `s` cuando:

1. `s` es un Space de la misma raíz, con extensión, que no es el aplastador ni está montado sobre él;
2. la cara de abajo del aplastador (su Z en el mundo menos `thickness`) queda entre `PressBelow` (1 cm) por debajo y `PressAbove` (5 mm) por encima de la superficie de `s`;
3. las cajas de ambas extensiones en el mundo se solapan.

Solo un apoyo **nuevo** aplasta: `pressing` guarda las superficies del tick anterior, y entra en el hash y el snapshot. Un tiny que camina por debajo de una mano ya apoyada no muere (ver Pendiente). La palma de los escenarios tiene 2 cm de grosor: con la pose de apoyo a 77 cm queda a ras de un escritorio de 75 cm, y la de mano suspendida a 88 cm no aplasta.

La huella es la extensión del aplastador, recortada por su grilla de navegación si tiene. La palma con dedos bloquea los huecos de su grilla, así que un tiny entre dos dedos sobrevive.

**Un objeto aplastador** (la taza, `content/objects/cup.json`, fuerza 3) aplasta al aterrizar: `crush.resolve` lee los `entity_fell` del tick anterior y aplasta en el Space donde quedó, en un disco de radio `radius` (o el de su `body`). Hay un tick de demora entre el aterrizaje y el aplastamiento.

Cada apoyo emite `pressed {by, source, surface}`, haya víctimas o no.

**Una pisada** (M24) la da `crush.steps`, en Move después de `nav.move`. Cada tick suma a `travel` lo que el que camina avanzó en su Space; cada `length` apoya un pie, empezando por el izquierdo y alternando, a `side` del centro de la marcha. La huella es un rectángulo de `foot_length × foot_width` orientado según el desplazamiento del tick, no según `facing`, porque el `mover` no orienta a nadie. Al cambiar de Space (una puerta) el recorrido empieza de nuevo. Cada pisada emite `footstep {entity, space, pos, right}` y aplasta en el acto, con `source: "foot"` y `by` igual a quien pisa. Solo alcanza al Space donde camina, el suelo de la habitación: un tiny sobre un mueble está en otro Space y no muere aunque esté encima del pie. Las pisadas son discretas: entre dos apoyos no hay barrido, así que un tiny puede quedar entre dos pasos.

## Víctimas

Una entidad `fragile` muere si:

- está en el Space que recibe el apoyo (lo que viaja sobre el aplastador está en otro Space: el tiny en la palma no muere cuando ella se apoya);
- su posición en el mundo cae en la huella;
- su `body.height` supera el hueco entre la cara de abajo y la superficie.

Una pisada no tiene hueco: alcanza a cualquier altura. Un refugio la salva si está en su mismo Space, a menos de `radius`, y `strength ≥ force`: emite `sheltered {entity, by, shelter}`.

Al morir, la víctima se destruye con un despawn normal y se crean restos: una entidad nueva en el mismo lugar con `location`, `body` (el radio de la víctima, 1 mm de alto), `name` y `remains`. Se emite `crushed {entity, by, source, space, pos, remains}`. El resto sale de los sistemas existentes en el mismo tick: `jobs.upkeep` libera las reservas (`holder_gone`), devuelve el trabajo a la cola (`job_unassigned`, `assignee_gone`) y lo que cargaba queda suelto.

## Sustos

Después de cada apoyo o pisada, los frágiles vivos de esa superficie que quedaron dentro de la huella (a cubierto) o a menos de `NearMissRange` (2 cm) reciben `near_miss {entity, by, source}`. Para `ai` es un evento que interrumpe, como una caída: suspende la acción en curso y el agente se recupera `RecoverTicks`.

## Amenazas (M25)

`crush.Threats(w)` devuelve lo que está por aplastar, calculado del estado en el momento (sin caché: no hay nada que guardar ni reconstruir al cargar):

- **Sombra:** un Space aplastador con su cara de abajo entre `PressAbove` y `ThreatHeight` (20 cm) sobre una superficie. La mano suspendida a 88 cm sobre el escritorio de 75 cm es una amenaza; la levantada a 140 cm, no.
- **Vibración:** la próxima pisada de quien camina, prevista con la dirección del último tick (`stride.dir`) y lo que falta recorrer hasta el paso.

`Threat.Covers(p, margin)` prueba un punto del mundo y ocho a `margin` alrededor.

## Huida (M25, en `sim/ai`)

Un agente con `agent.reflex > 0` reacciona a una amenaza que lo cubre en su propio Space. `ai.decide` lo revisa en cada tick, fuera del período de decisión:

1. Al quedar bajo la amenaza, emite `alarmed {entity, by, source, channel}` y fija `agent.alarm` en `tick + reflex`. Mientras espera, sigue con lo suyo.
2. Al llegar `alarm`, interrumpe lo que hace (`action_interrupted` con `by: "flee"`; un trabajo se suelta con lo que cargaba) y empieza `flee` hacia la salida más cercana. La salida se busca en 16 direcciones, en pasos de 2 mm hasta `FleeRange` (30 cm): un punto de su Space alcanzable por la grilla desde donde está (4 vecinos), fuera de todas las amenazas de ese Space con un margen de su radio más 1 mm. Un refugio que resista la amenaza gana si queda más cerca.
3. Si no hay salida, emite `trapped {entity, by}`, suelta lo que hace y se queda quieto. Si la huida falla en el camino, también queda atrapado (sin evento).
4. Al salir de toda amenaza se borran `alarm` y `trapped`.

`flee` no se interrumpe, está comprometida y no se retoma después de una suspensión. Un agente con `reflex` no elige opciones cuyo objetivo (o el objetivo de su trabajo) esté bajo una amenaza, así que no vuelve a meterse debajo de la mano que lo espantó.

Con los números de los escenarios (tiny a 1 mm por tick, `reflex` de 4 ticks), un tiny tiene que estar a pocos milímetros del borde para salvarse de una mano que espera 20 ticks. De una pisada a 1 m/s no se salva casi nadie: el aviso dura 7 ticks.

## Cuidado de la giantess (M26, en `sim/ai`)

`agent.care` (0 a 1) cambia cómo apoya la mano. Un paso `pose` de un agente con `care` cuya mano (`agent.hand`) está suspendida sobre una superficie (una amenaza de `crush.Threats`) es el que la apoya, y antes de hacerlo:

- espera `care · HoverExtra` ticks más (`HoverExtra` = 40), lo que da más tiempo para huir;
- en el primer tick, si hay frágiles bajo la huella (con margen de 2 cm) a menos de su `watch`, hace una tirada de atención con el RNG `ai.attention`: los nota con probabilidad `1 − (1 − care)^n`, con `n` tinies a la vista. Si los nota emite `hesitated {entity, surface, seen}` (`Doing.wary`) y sigue esperando mientras haya alguno debajo, hasta `CareWaitMax` (100 ticks). Después apoya igual.

`watch` también nota los restos (`noticed`). Su plan no cambia.

## Consecuencias en la colonia (M26, en `sim/colony`)

`colony.mourn` corre en Consequences después de `crush.resolve`, cuando la víctima todavía tiene su `member` (se destruye en el Flush). Por cada restos nuevos de un miembro:

- completa `remains.colony`;
- crea una zona de peligro: una entidad con `danger {scale, space, center, radius, until}` (store de `ai`) en el lugar de los restos, de `colony.danger_radius` (10 cm) por `colony.danger_ticks` (2 min);
- emite `member_lost {colony, entity, cause, remains, danger}`.

La colonia se entera en el acto, como de lo que ven sus miembros. Mientras la zona está vigente, los agentes de su escala no eligen opciones cuyo objetivo (o el de su trabajo) esté adentro, `ai.WanderPoint` (forage y la vuelta de explore) no elige puntos adentro y `explore` no cruza portales que terminan adentro. Al vencer, `colony.mourn` la borra con `danger_expired {danger, space}`.

## Presentación (M27)

Todo sale del `Frame` y de los eventos; nada cambia el hash.

- `view.Thing` suma `remains` (la causa: `hand`, `foot`, `object`), `alarmed` y `trapped`. `Frame.Footprints` trae la próxima pisada de cada uno que camina (`crush.NextSteps`): centro de mundo, dirección y tamaño del pie.
- Señales: `alarmed` y `trapped` dan la gota, que además manda sobre cualquier otra mientras el tiny huye o está atrapado; `near_miss` da «!»; `hesitated` da «?» sobre la giantess; `crushed` deja el espíritu (`spirit`) sobre los restos durante `CueTicks` (`Frame.Fresh`). La giantess se sorprende en el retrato al notar unos restos. Huir no enoja: la acción interrumpida por `flee` no da la vena.
- Historia: «Tiny 16 quedó bajo la palma en escritorio.» (o «bajo un pie», o el objeto), con los restos como actor, así que la entrada enfoca los restos; «se salvó por poco», «quedó atrapado sin salida» y «ve tinies bajo su mano y espera».
- Arte: visual `remains` (silueta gris aplanada con la bufanda menta; en macro solo si está elegida) y `footprint` (óvalo de la próxima pisada en macro, tenue, mientras la giantess camina), emote `spirit`. El escenario asigna `remains` a los restos al publicarse `crushed`.
- El piso de cerca: las habitaciones tienen vista micro en baldosas (tatami, madera, damero; ver `docs/specs/presentacion.md#pisos-de-cerca`). Ahí el visual `foot` muestra la sombra del pie que va a bajar (`shadow_<dir>`, más oscura cuanto menos le falta: `Footprint.Remaining` sobre `Footprint.Stride`) y el pie apoyado (`foot_<dir>`) durante `view.StepTicks` después de cada `footstep` (`Frame.Steps`), por encima de lo que pisa.
- Seguir a un tiny que muere pasa el seguimiento (y la selección) a sus restos.
- Sonido: `crush_plaf` (palma u objeto, micro), `crush_tum` (pisada, micro, se oye aunque quede fuera de cuadro), `crush_far` (macro), `emote_spirit`.

## Escenario de cierre (M27)

`python tools/scenarios/descuido.py` genera `un_descuido_en_el_escritorio` (giantess distraída, `care` 0) y `un_descuido_cuidadosa` (`care` 0,9) a partir de «Una mañana en la casa»:

- los tinies son `fragile` y tienen `reflex` 4; la palma es `crusher` y la giantess tiene `stride`;
- la pata del escritorio baja al tatami del dormitorio (portales `leg1_bottom` y `leg1_top`, como en `floor_to_desk_by_leg`);
- las migajas del desayuno se desparraman en un anillo de 18 cm: una cae detrás del escritorio, al piso, y otra a la palma;
- donde la giantess apoya la mano para desayunar quedaron tres galletas (no son `food`, así que la colonia no las acarrea) y tres tinies con hambre las mordisquean: dos bajo el medio de la palma y uno junto al borde;
- Tiny 21 empieza en el tatami, en el camino de la giantess cuando vuelve con el desayuno: en la semilla 42 la pisada le pasa al lado.

`TestUnDescuidoEnElEscritorio` corre las dos variantes con las semillas 1, 7, 21, 42 y 99, guardando y cargando cada 113 ticks, y verifica: invariantes en cada tick; ninguna palma aplasta a quien va sobre ella; cada muerte de un miembro da `member_lost`; ningún tiny elige como destino algo dentro de una zona de peligro vigente; al final nadie muerto tiene trabajos, reservas ni cargas. Con la distraída hay en cada semilla al menos un aplastado y al menos uno que se salva (por poco o huyendo) y ningún titubeo; con la cuidadosa hay al menos un titubeo y, en el total, menos muertes (hoy 12 contra 2).

## Invariantes

- `crusher.pressing` está ordenado.
- Los restos tienen nombre y causa.

## Escenarios

Generados por `python tools/scenarios/aplastamiento.py` a partir de `contact_opens_and_closes_route.json`.

- `palm_crushes_tiny`: la mano suspendida no aplasta. Al apoyarla muere el tiny de debajo y no el que está a 5 cm. Volver a apoyarla no repite el impacto sobre nadie.
- `tiny_on_palm_is_safe`: el tiny que viaja en la palma no muere al apoyarla, ni la primera vez ni la segunda.
- `tiny_between_fingers_survives`: la huella sale de la grilla de la palma.
- `nest_shelters_from_palm`: el tiny en el nido se salva y el que está afuera no.
- `cup_lands_on_tiny`: la taza cae 20 cm y aplasta al de abajo, no al que está a 5 cm.
- `haul_requeued_after_death`: la mano baja mientras el tiny levanta la migaja. La reserva se libera, el trabajo vuelve a la cola y otro tiny lo termina.
- `step_crushes_tiny_on_floor` (M24): la giantess cruza el dormitorio y pisa dos veces. Muere el tiny bajo el pie izquierdo y el que está en un nido en el suelo (la caja resiste una mano, fuerza 10, y no un pie, 600). Sobreviven el que quedó donde ese pie no se apoyó y el que está a un costado.
- `step_spares_tiny_on_desk` (M24): el tiny sobre el escritorio, encima del pie, no muere; el del suelo debajo, sí.
- `TestStrideSurvivesSave` (M24): guardar entre dos pisadas y continuar da la misma pisada y el mismo hash.
- `tiny_flees_shadow` (M25): la mano espera 20 ticks antes de apoyarse. El tiny a 5 mm del borde huye y se salva con un susto, el del centro no llega, el encerrado entre celdas bloqueadas queda atrapado y el lejano no se entera.
- `tiny_drops_crumb_to_flee` (M25): el tiny que levanta la migaja la suelta, deja el trabajo y huye. Nadie vuelve a buscarla mientras la sombra sigue ahí. Al apoyarse la mano, el que volvía se asusta, suspende el trabajo y lo termina después.
- `tiny_flees_footsteps` (M25): con la giantess a 0,5 m/s, la vibración anuncia la segunda pisada. El tiny a 0,5 mm del borde se salva por poco y el del centro no.
- `careless_giantess_crushes` y `careful_giantess_waits` (M26): la misma giantess escribe en el escritorio con dos tinies bajo la mano. Con `care` 0 los aplasta; con 0,9 espera más, los ve, titubea y nadie muere.
- `colony_avoids_danger` (M26): la mano mata a un miembro, la colonia marca la zona, la giantess nota los restos y nadie toma el acarreo de una migaja junto a ellos hasta que la zona vence; después la migaja llega al nido.
- `TestCrushScenariosSnapshots` (M25): los escenarios de aplastamiento dan el mismo log y hash guardando y cargando cada 3 ticks.
- `TestCrushOnlyOnNewPress`: un tiny camina debajo de la palma apoyada y no muere, tampoco después de guardar y cargar. El mundo restaurado tiene el mismo hash.

## Pendiente

- Debajo de una mano apoyada se puede caminar: la grilla del escritorio no sabe que la mano ocupa ese lugar.
- Cuidado: la giantess no corre la mano a otro lugar ni recuerda dónde aplastó; los restos no cambian su plan.
- Zonas de peligro: se evita el destino, no el camino (una ruta puede cruzar una zona).
- Huida: sin pánico colectivo ni aviso entre tinies; la salida no considera amenazas que todavía no existen.
- Pisadas: sin volumen barrido, sin ritmo propio (dependen solo de la distancia) y sin pisar cuando la giantess gira en el lugar. La mañana original no tiene `stride`; solo el escenario de cierre.
- Presentación: el tiny aplastado no tiene un cuadro propio (los restos aparecen con el espíritu); en macro no hay marca sobre el nido con los miembros perdidos; la giantess no se dibuja de cerca (solo su pie al pisar y la mano), y los muebles del piso (cama, heladera) tampoco; un tiny que trepa la pata se ve, de cerca, caminando sobre el tatami.
- Fuerza de los objetos por masa y altura de caída: hoy es un valor fijo por objeto.
