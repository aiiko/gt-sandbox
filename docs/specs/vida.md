# Spec: Vida cotidiana de los tinies (M39–M44)

Diseño y alcance del slice en la [propuesta](../propuesta.md#16-vida-cotidiana-de-los-tinies); plan en [plan-motor](../plan-motor.md). Los escenarios del slice los genera `python tools/scenarios/vida.py` (parte de la base de `aplastamiento.py` y le suma la pata `leg1` entre el piso y el escritorio).

## Despensa y rutina del nido (M39)

### Meta de la despensa

- `colony.stock` es la meta del depósito en porciones por miembro vivo (0 = sin meta, como hasta ahora).
- `colony.Deficit(w, col)` devuelve cuánto falta para la meta, en [0, 1]: 1 con el depósito vacío, 0 en la meta o por encima, y 0 sin meta. Se calcula al vuelo (miembros vivos × `stock` contra `portions.left`), así que no es estado: no cambia el hash ni el snapshot de las colonias sin meta.
- Lo leen `search` (M40) y la obra (M41).

### Comer y descansar

No hacen falta reglas nuevas en el selector. Con la curva cuadrática de las necesidades, comer en el nido (+300) vale más que un acarreo (`work` 0,2) más la inercia (0,15) cuando el hambre baja de unos 230; lo mismo con descansar (+400) y la energía. El acarreo se interrumpe, la carga queda en el piso (`job_unassigned` con `stopped`) y el trabajo vuelve a la cola: el mismo tiny u otro lo retoma después.

El nido se reserva mientras alguien come o descansa (como cualquier smart object), así que lo usa un tiny a la vez.

### Nido fuera de alcance

Si el nido no se alcanza (la pata cerrada), `eat` falla y el nido queda en `agent.avoid` durante `RetryTicks` (30 s). Mientras tanto el selector ofrece lo que el tiny conoce: una migaja o una galleta cercanas. Si no conoce nada, espera donde está hasta que el nido vuelve a ser una opción.

### Escenarios

- `tiny_returns_to_eat`: el hambre corta un acarreo; el tiny come en el nido, gasta una porción y después retoma el trabajo hasta guardar la migaja.
- `tiny_rests_at_nest`: el cansancio corta un acarreo y descansa en el nido. `TestRestRaisesEnergy` verifica que la energía sube.
- `nest_unreachable_eats_crumb`: con la pata cerrada, falla una vez al volver al nido, come una migaja del piso y no vuelve a intentar el nido antes de 30 s.
- `TestPantryDeficit` verifica el déficit (sin meta, vacío, parcial, en la meta y por encima).

## Puntos de búsqueda (M40)

### Lugares que se ven

`ai` suma el store `landmark {range}`: un marcador para lugares que los agentes con vista aprenden aunque no ofrezcan nada (puntos de búsqueda, servilleta, obra, punto de reunión). `ai.look` los aprende como a los smart objects (a distancia ≤ `sight`) y además desde `range` si es mayor: un rincón del escritorio se ve desde lejos aunque una migaja no. Como todo lo que se nota, la colonia lo comparte (`colony.share`).

### `search_spot` y la acción `search`

- Componente `search_spot {look_ticks, cooldown, empty_until}` (por defecto 3 s y 120 s). Va con `landmark` en una entidad ubicada en su superficie.
- `member.search` es la utilidad de revisar un punto con la despensa vacía. La opción vale `search × Deficit − distance_cost × distancia`, así que con la despensa en la meta no hay opción. Con `search` por debajo de `agent.work`, un acarreo conocido siempre le gana.
- Opciones: puntos conocidos, sin reserva ajena, fuera de `empty_until` y donde el tiny no sabe ya de comida suelta a menos de su vista (de eso se encarga la cola de trabajos). El selector descarta, como siempre, los objetivos evitados, amenazados o en peligro.
- `Start` reserva el punto (`jobs.Reserve`); `Step` renueva la reserva, camina hasta el punto y mira `look_ticks`. Termina antes si ve comida suelta. Emite `searched {entity, spot, found}`; sin hallazgo fija `empty_until = ahora + cooldown`. `Stop` libera la reserva.
- La ruta no se comprueba al empezar: si no hay camino, `nav.Approach` falla, la acción termina en `action_failed` y el punto queda evitado `RetryTicks` (30 s). Lo mismo si se cierra un portal a mitad de la salida.
- El hallazgo no necesita nada nuevo: `ai.look` nota la comida y `colony.share` publica el acarreo al nido.
- `forage` sigue igual (salir a un punto al azar): es la exploración que encuentra puntos nuevos. `encounter` cuenta `search` como buscar comida en sus avistamientos.

Stores nuevos (`landmark`, `search_spot`): `HashVersion` 9 y `SnapshotSchema` 12.

### Escenarios

- `colony_searches_spots`: tres tinies, nido vacío con meta 2. Uno revisa el borde vacío y otro el rincón con migajas; los acarreos llevan la comida al nido y se come. `TestSearchSpotReservedOnce` verifica en cada tick que dos tinies no revisan el mismo punto, y `TestStockedColonySearchesLess` que con el depósito en la meta se inician menos búsquedas.
- `search_portal_closed`: la pata se cierra abajo mientras el tiny baja a un punto del piso; la búsqueda falla, libera el punto y no se reintenta antes de 30 s.

## Fibras y refugio auxiliar (M41)

### Entidades

| Entidad | Componentes | Objeto |
| --- | --- | --- |
| Servilleta | `fiber_source {left, item}` y `landmark` | `content/objects/napkin.json` (`item: "fiber"` se expande como en los pasos `spawn`) |
| Fibra | `material {site}`, `landmark` y un cuerpo chato | `content/objects/fiber.json` |
| Obra | `build_site {colony, source, required, delivered, radius, shelter, done}` y `landmark` | `content/objects/refuge_site.json` |

`build_site.shelter` es el `crush.Shelter` que la obra recibe al terminar (por defecto fuerza 20, más que la palma, y el radio de la obra).

### Trabajos

`jobs.RegisterKind` agrega tipos de trabajo desde otros paquetes (en `init`, así el mapa no cambia durante la simulación). `colony` registra:

- `pluck` (objetivo: la servilleta): ir, esperar `PluckTicks` (2 s), crear una fibra a los pies del tiny en el mismo paso, con `material.site` de la obra que saca de esa servilleta, y gastar una unidad. El tiny la aprende. Emite `fiber_plucked {entity, source, fiber, left}`.
- `build` (objetivo: la obra): con todas las fibras, ir, esperar `BuildTicks` (5 s), marcar `done`, poner `shelter` en la obra (el mismo ID) y emitir `refuge_built {site, by}`. Desde ahí la huida (`escape`) y el aplastamiento la usan como cualquier refugio.
- El acarreo de una fibra es un `haul` con destino en la obra, así que hereda la carga que cae al huir y la vuelta a la cola.

### `colony.site` y la entrega

`colony.site` (fase Decide, después de `colony.share`) revisa cada obra sin terminar, en orden de store:

1. publica un `haul` para cada fibra suya que esté suelta, lejos de la obra y sin trabajo;
2. con `delivered == required`, publica un `build` si no hay uno;
3. si no, y la despensa de su colonia está en la meta (`Deficit == 0`) y algún miembro conoce la servilleta, publica un `pluck` cuando `required − delivered − fibras existentes > 0` y no hay otro `pluck` abierto (la servilleta se reserva, así que se desprende de a una);
4. con la servilleta agotada, cancela el `pluck` abierto.

`colony.deliver` (fase Consequences, junto a `colony.store`): una fibra suelta, sin reserva y a menos de `radius` de su obra suma 1 a `delivered`, desaparece y emite `fiber_delivered {site, item, total}`. Contar al desaparecer hace que cada fibra cuente una vez, la lleve quien la lleve; contar las existentes antes de pedir otra recupera las que quedaron en el piso.

Invariantes: `0 ≤ delivered ≤ required` y una obra con `done` tiene `shelter`.

Stores nuevos (`fiber_source`, `material`, `build_site`): `HashVersion` 10 y `SnapshotSchema` 13.

### Arte

`tools/art/vida.py` dibuja la servilleta, la fibra y la obra (cuadros `idle` y `built`) en vista micro; en macro están ocultas. Se regeneran con `python tools/art/build.py` y `go run ./cmd/artpack`.

### Escenarios

- `colony_builds_refuge`: con la despensa en la meta, dos tinies desprenden tres fibras, las llevan y uno arma el refugio. `TestColonyBuildsRefuge` verifica que cada fibra cuenta una vez, que la obra conserva su ID y recibe `shelter`, y la continuidad desde snapshots (uno a mitad de la obra).
- `fiber_dropped_is_recovered`: la sombra de la palma hace huir al que lleva la fibra y la suelta; otro miembro la entrega y el total es 1.
- `refuge_shelters_from_palm`: un tiny arma el refugio y la palma se apoya encima: los de dentro emiten `sheltered` y el de afuera muere.
