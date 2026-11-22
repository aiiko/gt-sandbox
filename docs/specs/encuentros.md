# Spec: Rutinas y encuentros (M28–M33)

Diseño y alcance del slice en la [propuesta](../propuesta.md#15-rutinas-y-encuentros-en-la-casa); hitos y justificación en el [plan](../plan-motor.md#hitos-de-rutinas-y-encuentros-cerrados).

## Rutina variable (M28)

### Partida libre

«Nueva partida» sale de `content/partidas/casa.json`, que genera `python tools/scenarios/partida.py` a partir de «Un descuido en el escritorio». `scenario.Randomize` la convierte en una partida concreta con el stream `partida` de la semilla, en este orden:

1. la colonia (cantidad, lugar, color, necesidades y carácter de los tinies);
2. el cuidado de la giantess (`agent.care`, de 0 a 0,9);
3. el perfil de arranque.

Agregar algo al final no cambia lo ya sorteado de las partidas existentes.

### Perfiles de arranque

El bloque opcional `random` de un escenario dice qué agente tiene perfiles y cuáles son:

```json
"random": {"agent": "$giantess", "starts": [
  {"name": "con hambre", "needs": [{"name": "hunger", "value": 180, "decay": 2}, …]},
  {"name": "cansada", …},
  {"name": "con ganas de entretenerse", …}
]}
```

- `Randomize` elige uno y reemplaza en ese agente `value` y `decay` de cada necesidad del perfil (las que no tenía se agregan).
- El nombre queda en `Spec.Profile` y el bloque `random` no pasa a la partida.
- `Build` ignora `random`: un escenario con perfiles se puede correr tal cual.

En la partida libre, hambre y diversión bajan 2 por segundo y la energía 1, así que la rutina vuelve a decidir (desayuno, merienda, televisión, escritura, dormir). Los escenarios de referencia conservan sus necesidades sin decaimiento.

### Merienda en el living

La heladera de la partida libre cambia `snack` (comer junto a la heladera) por `snack_living`. Es una oferta de varios pasos como `breakfast`, con hambre +350:

1. `goto` a la heladera y `wait`;
2. `take` del plato (`content/objects/plate.json`, visual `plate`, aplasta como la taza);
3. `goto` a `living_table` y `pose: set_table`;
4. `drop` del plato y `spawn` de tres migajas en un anillo de 12 cm;
5. `pose: hand_rest` y `use` durante 250 ticks.

Las migajas sobre la mesa del living son un segundo lugar de recolección: los tinies llegan por la palma apoyada y la colonia las acarrea con `haul`. Historia: «Giantess decide merendar en el living.» y «… termina de merendar.»; la interfaz la muestra como «Merienda en el living».

### Pruebas

- `giantess_snacks_in_living`: con hambre 750 (desayunar desperdiciaría más de la mitad), la giantess merienda en la mesa del living, deja el plato y migajas allí y no desayuna.
- `TestPartidaProfiles`: cada perfil aparece y se aplica tal cual; la base no cambia.
- `TestPartidaRoutineVaries`: sobre las semillas de referencia (1, 7, 21, 42 y 99), cinco minutos de partida dan al menos dos órdenes de actividades de la giantess, con al menos tres actividades cada uno.
- `TestPartidaSnapshotContinuation`: la partida y su continuación desde snapshots cada 113 ticks dan los mismos eventos y el mismo hash.

## Avistamientos (M29)

Paquete `sim/encounter`. Se activa por contenido: solo quien tiene el componente `encounter` (la giantess de la partida libre) registra avistamientos y tiene acciones de encuentro. Los escenarios de referencia no lo usan y no cambian.

### Componente `encounter`

```json
"encounter": {"observe": 0.12, "share": 0.2, "offer": 0.16,
  "sightings": [{"surface": "$desk1", "last": 340, "seen": 6, "foraging": 2}],
  "cooldowns": [{"surface": "$desk1", "action": "observe", "until": 2740}]}
```

- `observe`, `share` y `offer` son las utilidades de cada acción. Con 0 no la hace.
- `sightings`, por superficie (ordenados por ID), guarda lo último que vio: el tick (`last`), cuántos tinies (`seen`) y cuántos buscaban comida (`foraging`: los que estaban en `forage` o en un trabajo, que en la colonia es acarrear).
- `cooldowns`, ordenados por superficie y acción, indica hasta qué tick no repite una acción en una superficie.

El registro es estado que cambia decisiones, así que va en un componente: entra en el hash y en el snapshot (`HashVersion` 6, `SnapshotSchema` 10).

### Sistema `encounter.look`

Corre en Decide después de `ai.look` (Order −45), cada `ai.LookPeriod` ticks repartidos por agente. Para cada agente con `encounter` y `watch`:

1. cuenta, por superficie, los agentes de otra escala a menos de `watch` en el mundo, salvo los que van en su propia mano;
2. actualiza `last`, `seen` y `foraging` de esas superficies;
3. olvida los avistamientos sin novedad hace más de `SightingTTL` (2 minutos) y los enfriamientos vencidos.

A diferencia de `agent.known`, que solo dice si notó a un agente alguna vez, el registro responde dónde hay actividad reciente y qué hacen los tinies. Una acción de encuentro se abre con avistamientos de menos de `Recent` (20 s).

### Encuentro descartado

`encounter_skipped {entity, action, surface, reason}`: una acción de encuentro llegó y su condición ya no se cumple. Termina sin efecto y sin marcar la superficie como fallida, porque `ai` evita durante un rato los objetivos fallidos y el escritorio también es objetivo de `write`.

### Pruebas

`giantess_logs_sightings` (`TestGiantessLogsSightings`): la giantess registra a los tinies del escritorio, mantiene el orden y olvida a tiempo un avistamiento viejo de la cocina, guardando y cargando cada 113 ticks.

## Acciones de encuentro

Las acciones de encuentro son acciones extra de `ai` (`ai.Install(w, encounter.Actions()...)`), así que compiten en el mismo selector que las necesidades. Se mantienen la inercia, el compromiso con la mano apoyada y la interrupción por urgencia.

- **Opciones.** Hay una por superficie con avistamiento reciente (menos de `Recent`), sin enfriamiento de esa acción y con lugar donde pararse. Ese lugar es el `place` del mueble, que es la superficie misma: el escritorio, la mesa del living. Los pisos no tienen `place`, así que no abren encuentros.
- **Utilidad.** Es la fija de la acción en `encounter` menos `distance_cost` por la distancia. Las utilidades van por debajo de `UrgentUtility`, y la diversión no interviene.
- **Comprobación al llegar.** Si la condición ya no se cumple, la acción emite `encounter_skipped` y termina.
- **Enfriamiento.** La superficie se enfría para la acción al empezarla, así que una acción interrumpida no se repite enseguida.

`What` de las acciones de encuentro es su nombre, así que aparecen en `action_started` como cualquier oferta.

## Observar (M30)

`observe` va al lugar de la superficie, se orienta y se queda mirando `ObserveTicks` (10 s) con la mano en la pose de reposo, a 140 cm. La mano no está a menos de `ThreatHeight` de nada, así que no es una amenaza y los tinies siguen con lo suyo. Mientras mira, `encounter.look` sigue registrando.

- Al llegar, si el avistamiento ya no es reciente: `encounter_skipped` con `reason: "gone"`.
- Al terminar: `observed {entity, surface, seen}`, con cuántos tinies veía al final.
- Enfriamiento por superficie: `ObserveCooldown` (3 minutos).

En la partida libre `encounter.observe` vale 0,12. Le gana a quedarse sin hacer nada, pero no interrumpe una actividad en curso: la inercia es 0,15.

Que un tiny se aleje o se esconda porque lo miran queda fuera de esta entrega (ver el [plan](../plan-motor.md#m30-observar)).

Prueba: `giantess_observes_desk`. Sin necesidades que atender, la giantess va a mirar el escritorio, emite `observed`, no asusta a nadie y no vuelve a mirarlo en 75 s.

## Compartir migajas (M31)

`share` queda disponible cuando hay un avistamiento reciente con tinies buscando comida (`foraging > 0`) y menos de `ShareBelow` (2) porciones de comida a la vista en esa superficie. Comida a la vista es lo que tiene `food` y no va cargado; una migaja reservada para acarrear sigue sobre la mesa.

La acción reutiliza los pasos de oferta de `ai`:

- `ai.RunSteps(c, d, offer, loc)` avanza los pasos de una oferta sintética desde `d.Stage`, igual que `use`.
- `ai.UndoSteps(c, d)` deshace lo que quedó a medias (movimiento, lo cargado, la pose). `use.Stop` también los usa.

La oferta de `share` tiene estos pasos:

1. ir al lugar de la superficie (lo hace la acción, para comprobar al llegar);
2. `pose` con `encounter.drop_pose` (`set_table` si falta: la mano suspendida sobre la mesa);
3. `spawn` de `ShareCount` (2) copias de `encounter.portion`, a 10 cm delante de la mano en un anillo de 5 cm;
4. `wait` de 12 ticks.

`Stop` vuelve a la pose de reposo. Una migaja que cae no aplasta. La mano suspendida sí es una sombra, así que los tinies que estén justo debajo se apartan y vuelven después.

- **Al llegar.** Si ya no hay tinies buscando comida, emite `encounter_skipped` con `reason: "gone"`. Si ya hay comida suficiente a la vista, con `reason: "enough"`.
- **Enfriamiento.** Por superficie, `ShareCooldown` (4 minutos).
- **`encounter.portion`.** Es la definición completa del objeto: la simulación no conoce la biblioteca. `partida.py` copia `content/objects/crumb.json`.

### Motivo de lo que se deja

Todo paso `spawn` emite `items_left {entity, items, motive}`, donde `motive` es la oferta que lo produjo: `breakfast`, `snack_living` o `share`. Los IDs quedan además en `doing.items`. Al terminar, `share` emite `shared {entity, surface, items}`. Así la historia distingue las migajas de una comida de un gesto deliberado, y una prueba sigue cada porción hasta el depósito. `doing.items` entra en el hash (`HashVersion` 7).

### Pruebas

- `giantess_shares_crumbs` y `TestSharedCrumbsAreTraceable`: sin comida en el escritorio, los tinies salen a buscar y la giantess deja dos migajas. Cada una se sigue por `shared` → `noticed` → `job_posted` (`haul`) → `stored`.
- `giantess_skips_sharing`: caen migajas mientras ella camina, y al llegar emite `encounter_skipped` (`enough`) sin dejar nada.
- `TestNoSharingWithCrumbsInSight`: con migajas desde el principio, compartir no es opción.

## Ofrecer la palma (M32)

`offer_palm` queda disponible cuando hay un avistamiento reciente en una superficie con `contact_surface` (el escritorio, la mesa del living) y la mano también es superficie de contacto. Los pasos son:

1. ir al lugar de la superficie (lo hace la acción); si al llegar no hay nadie, emite `encounter_skipped` con `reason: "gone"`;
2. `pose` con `encounter.hover_pose` (`hand_hover`): la mano suspendida a menos de `ThreatHeight` es una sombra, y los tinies con `reflex` se apartan de la huella;
3. `wait` durante `OfferAnnounce` (3 s);
4. **apoyo que se cancela:** si quedan frágiles bajo la huella (`crush.Threat.Under` sin margen), emite `encounter_skipped` con `reason: "occupied"` y retira la mano sin apoyarla. Si no, `pose` con `encounter.rest_pose` (`hand_rest`). El apoyo de una giantess con `care` puede demorarse igual que en las ofertas, y la huella se revisa en cada tick hasta apoyar;
5. al apoyar, `palm_offered {entity, surface}`. `attach` abre el portal de contacto como en cualquier apoyo;
6. `wait` durante `OfferWait` (20 s). Los agentes de otra escala que llegan a la mano se anotan en `doing.items`;
7. `Stop` vuelve a la pose de reposo y, si la mano llegó a apoyarse, emite `palm_withdrawn {entity, surface, boarded}` con quienes subieron.

- **Enfriamiento.** Por superficie, `OfferCooldown` (5 minutos).
- **Compromiso.** Con la mano suspendida o apoyada (`doing.posed`), solo algo urgente la interrumpe.
- **Pasajeros.** Un tiny que sigue en la palma al retirarla queda ahí como pasajero y baja en el próximo apoyo. Llevarlo a un destino queda para «Llevar a un pasajero».

### Curiosidad que se recupera

`agent.visited` guarda cada Space con la última vez que el agente estuvo ahí (`[{space, at}]`). `agent.revisit` es cuánto tarda un Space visitado en volver a interesarle; con 0 (por defecto), nunca. Pasado ese tiempo, `explore` lo ofrece con `curiosity × RevisitFactor` (0,5), y el doble si el portal es dinámico.

En la partida libre los tinies tienen `revisit` de 90 s, así que pueden volver a subir a una palma que ya conocen. Los escenarios de referencia no lo usan y no cambian. Cambia el formato de `agent` (`HashVersion` 8, `SnapshotSchema` 11).

### Pruebas

- `tiny_boards_offered_palm` y `TestTinyBoardsOfferedPalm`: algún tiny curioso cruza por el portal, y `boarded` coincide con los `explored` hacia la palma.
- `nobody_boards_offered_palm` y `TestNobodyBoardsOfferedPalm`: sin curiosidad nadie sube, y la mano se retira al terminar la espera con `boarded: []`.
- `offer_canceled_when_occupied` y `TestOfferCanceledWhenOccupied`: un tiny sin reflejos se queda bajo la huella; la giantess cancela (`occupied`), no apoya y no muere nadie.
- `TestCuriosityRecovers`: la curiosidad por un Space visitado vuelve pasado `revisit`.

Todas guardan y cargan cada 113 ticks.

## Presentación (M33)

- **Actividades.** La interfaz muestra «Mira a los tinies», «Comparte migajas» y «Ofrece la palma». Mientras va hacia la superficie, cada una se muestra con «Va a …»: `view` marca `goto` en el primer tramo de las acciones de encuentro, según `encounter.IsAction`.
- **Símbolo.** Mientras observa, la giantess muestra el ojo (`view.EmoteLook`, `ui/emote_look.png`), con su sonido `emote_look`, que es menor: se omite a x4 y x16.
- **Retrato.** Pasa a curioso (`view.FaceCurious`, `ui/portrait_curious.png`) mientras observa, comparte u ofrece la palma, salvo en el camino. La merienda en el living lo pone contento, como el desayuno.
- **Historia.** Cuenta qué pasó, no qué podría pasar:

| Evento | Texto |
| --- | --- |
| `action_started` `observe` | «Giantess se acerca a mirar a los tinies de escritorio.» |
| `observed` | «… se queda mirando a N tinies en …» |
| `items_left` (`breakfast`, `snack_living`) | «Caen migajas del desayuno.» o «… de la merienda.» Reemplaza a «Aparecen migajas.» |
| `shared` | «… deja unas migajas para los tinies de …» |
| `palm_offered` | «… ofrece la palma abierta a los tinies de …» |
| `palm_withdrawn` | «Nadie sube a la palma; … la retira.», «Tiny 03 subió a la palma; …» o «Subieron N tinies a la palma; …» |
| `encounter_skipped` `occupied` / `enough` | «… ve un tiny donde iba a apoyar la mano y no la apoya.» / «… ve que ya hay migajas en …» |

- **Menú.** La nota de «Nueva partida» dice que la giantess cambia de rutina y que, si ve tinies, se detiene a mirarlos, les deja migajas o les ofrece la palma.
- **Referencias.** Las capturas de referencia quedan idénticas. Su resumen cambia porque el atlas se reempaquetó y la historia de la mañana tiene menos entradas (una de migajas en vez de una por migaja). La mezcla de referencia no cambia.

## Aceptación

| Criterio de [15.3](../propuesta.md#153-criterios-de-aceptación) | Prueba |
| --- | --- |
| Semillas distintas, al menos dos órdenes de actividad; con snapshot, las mismas decisiones | `TestPartidaRoutineVaries`, `TestPartidaSnapshotContinuation` |
| Se detiene a observar y no repite en la misma superficie hasta que vence el enfriamiento | `giantess_observes_desk`, `TestEncountersAcceptance` |
| Comparte migajas rastreables hasta `stored`; con suficientes, no es opción | `TestSharedCrumbsAreTraceable`, `giantess_skips_sharing`, `TestNoSharingWithCrumbsInSight` |
| Palma ofrecida: alguien sube, nadie sube, se cancela con la huella ocupada | `TestTinyBoardsOfferedPalm`, `TestNobodyBoardsOfferedPalm`, `TestOfferCanceledWhenOccupied` |
| La historia distingue merienda de compartir y cuenta los desenlaces que ocurrieron | `TestEncounterStory` (en `view`) |

`TestEncountersAcceptance` corre diez minutos de partida libre con cada semilla de referencia (1, 7, 21, 42 y 99). En cada partida verifica dos cosas: que ningún encuentro se repita en la misma superficie antes de su enfriamiento, y que ningún aplastamiento ocurra durante una palma ofrecida. Entre todas las semillas deben aparecer todos los desenlaces: mirar, compartir con alguna porción guardada, alguien que sube y una oferta sin pasajeros o cancelada.
