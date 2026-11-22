# Spec: Rutinas y encuentros (M28–M33)

Diseño y alcance del slice en la [propuesta](../propuesta.md#15-rutinas-y-encuentros-en-la-casa); hitos y justificación en el [plan](../plan-motor.md#hitos-pendientes).

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
