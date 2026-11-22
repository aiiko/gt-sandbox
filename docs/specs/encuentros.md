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
