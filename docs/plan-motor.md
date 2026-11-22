# Plan de desarrollo

**Una mañana en la casa** está cerrada en imagen y sonido. El [cierre visual](specs/presentacion.md#cierre-de-la-entrega-visual) reúne las capturas de referencia, la prueba de continuidad y las mediciones de rendimiento. La [spec de sonido](specs/sonido.md) cubre el sonido sintetizado y sin música (M20 a M22): la base de audio, los sonidos de la mañana con perspectiva por escala, el ambiente por habitación y la mezcla de referencia. El [escenario de referencia](specs/manana-en-la-casa.md) fija el recorrido con el que se integró cada hito.

**Un descuido en el escritorio** (M23 a M27) también está cerrado: la [spec de aplastamiento](specs/crush.md) cubre la muerte bajo la palma, la pisada o un objeto, la huida y el cuidado de la giantess. La partida libre del menú (`scenario.Randomize`) parte de ese escenario, con la colonia y el cuidado de la giantess sorteados.

## Hitos pendientes

El slice **Rutinas y encuentros en la casa** hace que la giantess varíe su rutina entre partidas y que lo que hacen los tinies le abra acciones nuevas: observar, compartir migajas y ofrecer la palma. El diseño está en la [propuesta](propuesta.md#15-rutinas-y-encuentros-en-la-casa). Cada hito abre o amplía una spec nueva, `docs/specs/encuentros.md`, y actualiza la [spec de IA](specs/ai.md) cuando cambia un contrato.

| Hito | Contenido | Cierre |
| --- | --- | --- |
| M28 Rutina variable (hecho, [spec](specs/encuentros.md#rutina-variable-m28)) | Perfiles de arranque de la giantess sorteados por `scenario.Randomize`. Hambre y diversión que decaen en la partida libre. Merienda en el living como oferta de varios pasos que deja migajas en la mesa del living | `giantess_snacks_in_living`. Semillas de referencia con al menos dos órdenes de actividades. Continuación desde snapshot con las mismas decisiones |
| M29 Avistamientos (hecho, [spec](specs/encuentros.md#avistamientos-m29)) | Paquete `sim/encounter`: registro de actividad de tinies por superficie (`sightings`) y enfriamiento por superficie y acción (`cooldowns`). Opciones de encuentro con utilidad moderada en el selector. Comprobación al llegar | `giantess_logs_sightings` (registro, olvido y orden estable). Incrementar `HashVersion` y `SnapshotSchema` |
| M30 Observar | Acción `observe`: ir a un punto de observación, esperar y seguir registrando. Enfriamiento por superficie | `giantess_observes_desk` (se detiene a mirar y no repite en el mismo lugar hasta que vence el enfriamiento) |
| M31 Compartir migajas | Acción `share` sobre la maquinaria de pasos de `use`: dejar una porción donde hay tinies buscando comida y pocas migajas a la vista. Evento con motivo para distinguir compartir de merendar | `giantess_shares_crumbs` (porción rastreable hasta `stored`), `giantess_skips_sharing` (con migajas suficientes no es opción) |
| M32 Ofrecer la palma | Acción `offer_palm`: anuncio con la sombra, apoyo que se cancela si la huella sigue ocupada, espera con el portal de contacto abierto y retiro. Curiosidad que se recupera para volver a subir | `tiny_boards_offered_palm`, `nobody_boards_offered_palm`, `offer_canceled_when_occupied` |
| M33 Presentación y cierre | Actividades, símbolos y retrato de los encuentros. Historia con motivo y desenlace. Sonidos y nota del menú. Aceptación de la propuesta sobre las semillas de referencia | Prueba de aceptación de [15.3](propuesta.md#153-criterios-de-aceptación). Capturas y mezcla de referencia sin cambios o actualizadas a propósito |

### Punto de partida: qué hay y qué falta

Los cambios salen de comparar la propuesta con lo que el motor hace hoy:

- **La giantess arranca siempre igual y deja de decidir.** En `un_descuido_en_el_escritorio.json` (la base de la partida libre) tiene hambre 220 y diversión 460, las dos con `decay: 0`; solo la energía baja. `scenario.Randomize` sortea la colonia y `agent.care`, pero no toca sus necesidades. Después de desayunar y mirar televisión no le queda ninguna necesidad que la mueva, así que todas las semillas producen el mismo orden de actividades. Esto incumple el primer criterio de aceptación.
- **No hay merienda que deje migajas.** La oferta `snack` de la heladera es un `use` de 200 ticks junto a la heladera, sin pasos. La mesa del living tiene `place: living_table` pero `offers: []`. Las migajas solo aparecen en el desayuno, siempre en el escritorio.
- **La giantess nota a cada tiny una sola vez.** `ai.look.watch` agrega a `agent.known` a los agentes de otra escala a menos de `watch` y no los vuelve a registrar mientras estén en la memoria. No guarda en qué superficie estaban, qué hacían ni cuándo los vio por última vez. Para la giantess, `known` es solo un registro que no cambia su plan (ver la [spec de IA](specs/ai.md#percepción-y-memoria)).
- **Las acciones salen de necesidades o de la colonia.** El selector junta opciones de `use` (ofertas de objetos), `work`, `explore`, `flee` y `forage`. Ninguna surge de lo que la giantess vio. `ai.Install` admite acciones extra (así entra `colony.Forage`), así que hay un punto de extensión.
- **Anunciar el apoyo es una amenaza.** La mano a menos de `ThreatHeight` (20 cm) sobre una superficie es una sombra en `crush.Threats`. Los tinies con `reflex` se alarman y huyen. El paso `pose` con `care` espera mientras haya tinies debajo, pero al llegar a `CareWaitMax` apoya igual. La propuesta pide cancelar si la huella sigue ocupada.
- **Un Space visitado no vuelve a interesar.** `explore` solo ofrece portales hacia Spaces fuera de `agent.visited`, y esa lista nunca se olvida. Un tiny que ya estuvo en la palma no vuelve a subir aunque se la ofrezcan. En la partida libre eso pasa pronto: la palma se apoya en el desayuno.
- **La historia no distingue motivos.** `view.storyText` escribe «Aparecen migajas.» para todo `entity_spawned` de una migaja, y `action_started` solo tiene texto para `breakfast`, `watch` y `sleep`.

### M28 Rutina variable

**Perfiles de arranque en la partida libre.** `scenario.Spec` suma un bloque opcional `random` con la referencia del agente (`"agent": "$giantess"`) y una lista de perfiles (`starts`): por ejemplo, con hambre, cansada y con ganas de entretenerse. Cada perfil reemplaza `value` y `decay` de sus necesidades. `Randomize` elige uno con su stream `partida`, después de sortear la colonia y el cuidado, así que las partidas ya existentes no cambian de colonia.
*Por qué:* la propuesta pide que la semilla elija entre unos pocos puntos de partida. Si los perfiles viven en el escenario base, `Randomize` sigue siendo un generador de inputs fuera de la simulación (D4 y D7 se mantienen) y el contenido se ajusta sin tocar código. `Spec` rechaza campos desconocidos, así que el bloque necesita su campo con su validación.

**Decaimiento de hambre y diversión.** Los perfiles traen un `decay` bajo para hambre y diversión (del orden de 1 por segundo, como la energía).
*Por qué:* sin decaimiento la rutina se agota en la primera media hora simulada. Se activa solo en la partida libre: `una_manana_en_la_casa` y `un_descuido_en_el_escritorio` conservan sus valores. Así se mantienen su aceptación, las capturas de referencia y la mezcla, que dependen de esos escenarios.

**Merienda en el living.** La heladera de la base de la partida libre suma una oferta `snack_living` de varios pasos, con la misma forma que `breakfast`: ir a la heladera, sacar un plato (`take`), ir a `living_table`, apoyar la mano, dejar el plato y migajas (`drop`, `spawn` con `spread`) y comer (`use`). Necesita un objeto nuevo `content/objects/plate.json` y su arte en `content/art/src/casa/` (regenerar con `go run ./cmd/artpack`).
*Por qué:* la propuesta pide reutilizar la colocación de comida, y `breakfast` ya resuelve todos esos pasos. Las migajas en la mesa del living abren un segundo lugar de recolección: la colonia puede descubrirlo con `explore` (cruzando la palma apoyada) y acarrearlo con `haul`. El `snack` actual, que se come junto a la heladera, se quita de esa base. Si no, gana siempre por distancia (`distance_cost`) y la merienda en el living no ocurre.

**Cierre.** `giantess_snacks_in_living`: con hambre, la giantess merienda en la mesa del living y deja migajas ahí. Una prueba sobre las semillas de referencia (1, 7, 21, 42 y 99, como en M27) verifica que hay al menos dos órdenes distintos de `action_started` de la giantess. La continuación desde snapshot reutiliza `playChecked` (guardar y cargar cada 113 ticks) y compara hashes y eventos.

### M29 Avistamientos

**Paquete `sim/encounter`.** El componente `encounter` va en la giantess y tiene dos partes:

- `sightings`: por superficie, el último tick en que vio tinies, cuántos vio y cuántos buscaban comida (`forage` o un `haul` en curso).
- `cooldowns`: por superficie y acción, hasta qué tick no la repite.

Las dos listas se guardan ordenadas por ID de superficie (D5) y los avistamientos viejos se olvidan pasado `SightingTTL`. El sistema `encounter.look` corre en Decide después de `ai.look`. Cada `LookPeriod` recorre los agentes de otra escala a menos de `watch` y actualiza el registro.
*Por qué:* `agent.known` responde a «¿lo vi alguna vez?» y la propuesta necesita «¿hay actividad reciente en esta superficie, y qué hacen?». Un paquete aparte sigue el patrón de `colony` (estado propio, acción extra instalada con `ai.Install`). `sim/ai` queda genérico y el comportamiento se activa por contenido: sin `encounter`, la giantess de las escenas de referencia no cambia, igual que sin `fragile` no hay aplastamiento. El registro es estado que afecta decisiones, así que va en un componente y no en un recurso: entra en el hash y en el snapshot.

**Opciones de encuentro en el selector.** Cada acción de encuentro ofrece una opción por superficie con avistamiento reciente y sin enfriamiento vigente. La utilidad es moderada y fija por acción (`encounter.utility` en el componente, por debajo de `UrgentUtility`) y se le resta `distance_cost`.
*Por qué:* la propuesta pide que necesidades y encuentros compitan en el mismo selector con una utilidad independiente de la diversión. Así se mantienen la histéresis (`inertia`), el compromiso de las ofertas con la mano apoyada y la interrupción por urgencia. Ninguna regla del selector cambia.

**Comprobación al llegar.** Cada acción de encuentro revalida su condición en `Valid` y en el paso previo a dejar comida o apoyar la mano. Si ya no se cumple, termina con `encounter_skipped {entity, action, surface, reason}`.
*Por qué:* lo pide la propuesta, y los tinies se mueven mientras ella camina. Además, un evento explícito permite verificar el criterio de que compartir deje de ser opción con migajas suficientes.

**Cierre.** `giantess_logs_sightings` verifica el registro, el olvido y el orden. Se incrementa `HashVersion` porque cambia lo que se hashea (store nuevo) y `SnapshotSchema` porque un snapshot anterior no trae el store.

### M30 Observar

**Acción `observe`.** Va a un punto de observación junto a la superficie, espera `ObserveTicks` y, mientras espera, `encounter.look` sigue actualizando el registro (puede notar actividad nueva). Al terminar fija el enfriamiento de esa superficie. Eventos: `action_started` con `what: "observe"` y `observed {entity, surface, seen}`.
El punto de observación es un `place` con etiqueta `observe` en los muebles (`desk.json`, la mesa del living) y el centro del piso en las habitaciones.
*Por qué:* `place` y el paso `goto` ya resuelven llegar y orientarse. Una etiqueta propia evita que observar compita por el mismo `spot` que desayunar. La mano queda en la pose de reposo (a 140 cm, fuera de `ThreatHeight`), así que observar no es una amenaza. Los tinies siguen con lo suyo, como pide la propuesta.

**Límite de la entrega.** Que un tiny se aleje o se refugie *porque lo miran* requiere que los tinies perciban a la giantess, y hoy solo perciben amenazas. Queda en Acciones posteriores. Los criterios de 15.3 no lo exigen, y las pisadas al acercarse ya pueden espantar a los del piso.

### M31 Compartir migajas

**Acción `share`.** Está disponible si la superficie tiene un avistamiento reciente con tinies buscando comida y hay menos de `ShareBelow` migajas a la vista en ella (`food` suelta, sin reserva ni carga). Ejecuta una oferta sintética con los pasos existentes: ir al `place` de la superficie, mano suspendida, `spawn` de una porción y volver a la pose de reposo.
*Por qué:* la propuesta pide reutilizar la colocación de comida. El ejecutor de pasos vive en `ai/use.go`; se lo expone (`ai.RunSteps` o similar) para que una acción de otro paquete ejecute pasos sin duplicar `goto`, `pose`, `spawn` ni el cierre ante interrupciones. Una migaja que cae no es `crusher`, así que no aplasta a nadie.

**Motivo en los eventos.** `item_dropped` suma `motive` (la acción que lo produjo: `breakfast`, `snack_living`, `share`), y el paso `spawn` emite `items_left {entity, items, motive}` junto a los `entity_spawned`. `encounter` emite además `shared {entity, surface, items}` con los IDs de lo que dejó.
*Por qué:* el criterio pide distinguir en la historia una merienda que deja migajas de un gesto deliberado. También pide rastrear la porción hasta que la colonia la guarda. Con los IDs, una prueba sigue la cadena `shared` → `noticed` → `job_posted` (`haul`) → `stored`.

**Cierre.** `giantess_shares_crumbs` (porción rastreable hasta `stored`) y `giantess_skips_sharing` (con migajas suficientes, `share` no es opción o termina en `encounter_skipped`).

### M32 Ofrecer la palma

**Acción `offer_palm`.** Está disponible si un tiny avistado está cerca de un punto de apoyo alcanzable: una superficie con `contact_surface` y un `place` donde la giantess tiene pose de apoyo. Los pasos son:

1. ir al `place`;
2. `hand_hover` (la sombra anuncia el apoyo);
3. esperar `OfferAnnounce`;
4. apoyar solo si la huella está libre;
5. esperar `OfferWait` con el portal de contacto abierto;
6. volver a la pose de reposo.

Eventos: `palm_offered {entity, surface}` y `palm_withdrawn {entity, surface, boarded}`, donde `boarded` lista quién cruzó.
*Por qué:* las poses y los portales de contacto (`attach`) ya existen y el portal se abre solo al apoyar. Que la sombra sea una amenaza es lo que hace segura la oferta: los tinies con `reflex` salen de la huella antes de que baje la mano.

**Apoyo que se cancela.** El paso `pose` suma un modo `cancel_if_occupied`: si al terminar el anuncio hay frágiles bajo la huella (la misma prueba que usa `care`), la acción termina sin apoyar y emite `encounter_skipped` con `reason: "occupied"`.
*Por qué:* con `care`, al vencer la espera apoya igual. Una oferta que aplastara a un tiny contradiría la propuesta («si la huella sigue ocupada, cancela el apoyo») y su tercer desenlace verificable.

**Curiosidad que se recupera.** `agent.visited` pasa a guardar el tick de la visita, y `explore` vuelve a ofrecer un Space visitado pasado `RevisitTicks`, con la utilidad reducida.
*Por qué:* hoy un tiny que ya estuvo en la palma no vuelve a subir nunca, y en la partida libre casi todos la visitan en el desayuno. Sin este cambio, «un tiny cruza por decisión propia» solo ocurriría la primera vez. Ya figura como pendiente en la [spec de IA](specs/ai.md#pendiente). Cambia el formato de `agent`: `SnapshotSchema` (y `HashVersion` si no se incrementó en M29 dentro de la misma versión).

**Pasajeros.** Si un tiny queda en la palma al retirarla, sigue ahí como hoy y baja en el próximo apoyo (desayuno, merienda o TV). Llevarlo a un destino es «Llevar a un pasajero», de Acciones posteriores.

**Cierre.** `tiny_boards_offered_palm` (un tiny curioso cruza por el portal), `nobody_boards_offered_palm` (sin curiosidad nadie sube y la mano se retira al acabar la espera) y `offer_canceled_when_occupied` (un tiny atrapado en la huella hace cancelar el apoyo, sin `crushed`).

### M33 Presentación y cierre

- **Interfaz.** `cmd/game/hud.go` suma las actividades («Merienda en el living», «Mira a los tinies», «Comparte migajas», «Ofrece la palma»). Observar muestra un símbolo sobre la giantess y el retrato pasa a curiosa. Hacen falta un clip de emote y una expresión en `content/art/src/casa/`.
- **Historia.** `view.storyText` suma textos para `observed`, `shared` (distinto de las migajas de una comida, gracias a `motive`), `palm_offered` y `palm_withdrawn`, y para `encounter_skipped` cuando importa. La historia solo cuenta un cruce o un acarreo cuando ocurrió, porque `boarded` y `stored` salen de eventos.
- **Sonido.** Cambios menores en `tools/audio/sounds.json`, por ejemplo un roce al dejar migajas. Solo presentación.
- **Menú.** La nota de «Nueva partida» menciona que la giantess cambia de rutina.
- **Referencias.** Las capturas (`docs/capturas`) y la mezcla (`docs/sonido/mezcla.json`) usan la mañana y no deberían cambiar. Si cambian, se regeneran a propósito.

**Aceptación.** Una prueba recorre los criterios de [15.3](propuesta.md#153-criterios-de-aceptación) sobre las semillas de referencia:

- órdenes de actividad distintos entre semillas, y los mismos con snapshot;
- observar y su enfriamiento;
- compartir rastreable y su descarte;
- los desenlaces de la palma;
- la historia distingue merienda de compartir.

## Contratos de integración

- La presentación lee `view.Frame` y los eventos publicados. Puede ampliar los datos que captura `view` para resolver apariencia, pose u oclusión. El estado lógico se modifica mediante comandos.
- El formato del atlas, el escalado del lienzo y las referencias visuales están en la [spec de presentación](specs/presentacion.md).
- Cámara, filtros, selección, animación y transiciones pertenecen a la presentación. No afectan rutas, portales, tiempos lógicos ni snapshots. Se mantienen [las reglas del motor](../CLAUDE.md) y [los contratos de simulación](specs/nucleo.md).
- Cada hito cierra con una prueba reproducible sobre el escenario de referencia, una captura cuando cambia el dibujo y la spec actualizada cuando cambia un contrato. Ejecutar `go test ./...` y `go vet ./...` antes del cierre.
- Los encuentros se activan por contenido (`encounter` en la giantess) y los perfiles de rutina solo en la partida libre. `una_manana_en_la_casa` y `un_descuido_en_el_escritorio` conservan su comportamiento y su aceptación.

El alcance de la mañana termina en las tres habitaciones y el viaje en la palma. El slice de aplastamiento se activa por contenido (`fragile`), así que la mañana no cambia. Quedan fuera las heridas, las caídas letales, el aturdimiento sin muerte, sentarse o acostarse sobre tinies, las alarmas entre tinies y las relaciones entre la giantess y la colonia. Del slice de encuentros quedan para después ordenar el escritorio, llevar a un pasajero (ver [15.4](propuesta.md#154-acciones-posteriores)) y que los tinies reaccionen a ser observados. El sonido es solo presentación: sale de eventos y del `Frame` y nunca cambia la simulación. La música queda para una etapa posterior. Las superficies corporales adicionales, la construcción de colonia, el LOD de simulación y la adopción de editores externos quedan para etapas posteriores.
