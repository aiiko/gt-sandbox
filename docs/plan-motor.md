# Plan de desarrollo

Los slices anteriores están cerrados y documentados en sus specs: la mañana ([presentación](specs/presentacion.md), [sonido](specs/sonido.md)), el aplastamiento ([crush](specs/crush.md)), los encuentros ([encuentros](specs/encuentros.md)) y los asientos ([asientos](specs/asientos.md)). **Vida cotidiana de los tinies** (M39 a M44) también está cerrado: la [spec de vida](specs/vida.md) cubre lo implementado y su aceptación; el análisis y la justificación de cada hito siguen abajo.

## Hitos de Vida cotidiana de los tinies

El slice **Vida cotidiana de los tinies** (M39 a M44) les da a los tinies motivos para salir, volver y quedarse en el nido: una despensa con meta, puntos de búsqueda, un refugio auxiliar que se construye con fibras de papel y un punto de reunión. El diseño está en la [propuesta](propuesta.md#16-vida-cotidiana-de-los-tinies). Cada hito abre o amplía una spec nueva, `docs/specs/vida.md`, y actualiza la [spec de IA](specs/ai.md) cuando cambia un contrato.

| Hito | Contenido | Cierre |
| --- | --- | --- |
| M39 Despensa y rutina del nido (hecho, [spec](specs/vida.md#despensa-y-rutina-del-nido-m39)) | Meta de porciones por miembro en `colony` (`colony.Stocked`). Necesidades de los tinies ajustadas para que coman y descansen en momentos distintos. Hambre o cansancio urgentes que cortan un acarreo. Alternativa cuando el nido no está al alcance | `tiny_returns_to_eat`, `tiny_rests_at_nest`, `nest_unreachable_eats_crumb` |
| M40 Puntos de búsqueda (hecho, [spec](specs/vida.md#puntos-de-búsqueda-m40)) | Store `landmark` en `ai` (la percepción aprende cosas sin ofertas). Componente `search_spot` y acción `search` en `colony`, con reserva exclusiva y enfriamiento si el punto estaba vacío. Utilidad según el déficit de la despensa. `encounter` cuenta `search` como buscar comida | `colony_searches_spots` (búsqueda → hallazgo → acarreo → depósito → consumo), `stocked_colony_searches_less`, `search_spot_reserved_once`, `search_portal_closed` |
| M41 Fibras y refugio auxiliar (hecho, [spec](specs/vida.md#fibras-y-refugio-auxiliar-m41)) | Servilleta (`fiber_source`), fibra (`material`) y obra (`build_site`). Tipos de trabajo `pluck` y `build`; el acarreo reutiliza `haul`. Sistema `colony.build` que cuenta recibidas, en viaje y sueltas antes de pedir otra. Al terminar, la obra recibe `shelter` con el mismo ID | `colony_builds_refuge`, `fiber_dropped_is_recovered`, `refuge_shelters_from_palm` |
| M42 Reunión (hecho, [spec](specs/vida.md#reunión-m42)) | Necesidad `fun` en los tinies. Componente `gathering` (dos plazas, plazo y pausa) y acción `chat` | `tinies_chat`, `chat_partner_never_comes`, `chat_interrupted_by_threat` |
| M43 Partida libre (hecho, [spec](specs/vida.md#partida-libre-m43)) | Puntos de búsqueda, servilleta, obra, punto de reunión y comida cerca de la silla en `content/partidas/casa.json`. `Randomize` sortea `fun` y la despensa inicial. Ajuste de utilidades sobre las semillas de referencia | Prueba de rutina y ocio (tres actividades distintas en 5 min por semilla) y continuidad desde snapshot a mitad de una búsqueda y de la obra |
| M44 Presentación y cierre (hecho, [spec](specs/vida.md#presentación-y-cierre-m44)) | Textos de actividad con motivo y destino en la ficha, símbolos en micro, arte de servilleta, fibra, obra y refugio, historia, sonidos y nota del menú | Prueba de aceptación de [16.4](propuesta.md#164-cierre-de-la-iteración). Capturas y mezcla de referencia sin cambios o actualizadas a propósito |

El orden sigue las dependencias: la meta de la despensa (M39) es la entrada de la utilidad de `search` (M40) y de cuándo se pide material (M41). La reunión (M42) es independiente, pero necesita que las prioridades de M39 a M41 estén ajustadas para no tapar al resto. M43 es el primer hito que toca la partida libre.

### Punto de partida: qué hay y qué falta

Los cambios salen de comparar la propuesta con lo que el motor hace hoy:

- **La despensa no motiva nada.** `colony.store` guarda la comida que llega al nido, pero ninguna decisión lee cuántas porciones hay. `forage` vale `member.forage` (0,02 en la partida) esté el nido lleno o vacío, y elige un punto al azar a 30 cm con `WanderPoint`: no recuerda adónde fue ni deja de volver a un lugar vacío.
- **La percepción solo ve ofertas y portales.** `ai.look` aprende entidades con `offers` y extremos de portal. Un punto de búsqueda, una servilleta o una fibra no tienen ofertas, así que nadie los vería ni la colonia los compartiría.
- **Las reservas son de un solo dueño.** `reservation` es un componente de la cosa reservada; sirve para un punto de búsqueda o una fibra, pero no para un punto de reunión con dos plazas.
- **Los trabajos solo acarrean.** `jobs.Kinds` tiene solo `haul`, y `colony.share` publica acarreos solo de `food` hacia el nido. No hay trabajos que creen algo (desprender una fibra) ni que cambien a otra entidad (armar la obra).
- **El refugio ya funciona.** `crush.Shelter` protege a quien está a menos de `Radius` de un aplastador con `Force ≤ Strength`, y `ai.escape` ya elige un refugio del mismo Space si queda más cerca que la salida libre. Basta con que la obra reciba `shelter` al terminar.
- **Los tinies no tienen diversión.** En `casa.json` tienen `hunger` (decae 2/s) y `energy` (1/s). El nido ofrece `eat` (+300) y `rest` (+400). Con `curiosity: 0,8`, la exploración le gana a casi todo mientras haya Spaces por (re)visitar.
- **La galleta no es comida de la colonia.** Las galletas del escritorio tienen `offers` y `portions` pero no `food`: se mordisquean donde están y nunca se acarrean. Las migajas sí tienen `food`.
- **Ya hay piezas para las interrupciones.** `flee` corta cualquier acción; `work` suelta la carga al interrumpirse (`job_unassigned`); una acción fallida evita su objetivo 30 s (`agent.avoid`), así que un portal cerrado no se reintenta en cada decisión.
- **La ficha no dice adónde va.** `view.activity` devuelve acción y paso, pero el `Frame` no lleva el destino ni el motivo.

### M39 Despensa y rutina del nido

**Meta de la despensa.** `colony` suma `stock` (porciones por miembro vivo). `colony.Stocked(w, col)` devuelve el déficit en [0, 1]: 0 con el depósito en la meta o por encima, 1 vacío. Se calcula al vuelo contando miembros: no es estado nuevo.
*Por qué:* es la única entrada que necesitan M40 y M41 para decidir cuándo buscar o construir. Calcularlo en vez de guardarlo evita mantener un contador sincronizado con muertes y depósitos.

**Hambre y cansancio que interrumpen.** Se verifica con escenarios que `use:eat` le gane a `work` con la histéresis actual cuando el hambre baja de ~300, y que `rest` haga lo mismo con la energía. Si la curva no alcanza, el ajuste es de contenido (`weights` o `amount`), no del selector.
*Por qué:* con `d(v)` cuadrática, comer desde 300 vale ~0,33 y un acarreo 0,2 + prioridad; la regla de inercia (0,15) deja un margen fino. Medirlo antes de tocar código evita cambiar reglas que sostienen la aceptación de los slices anteriores.

**Alternativa al nido.** Si `eat` falla (portal cerrado), el nido queda en `avoid` 30 s y el selector ya ofrece las migajas y galletas conocidas. Si no hay ninguna, el tiny espera quieto donde está hasta que el nido vuelve a ser una opción. (Esperar en un refugio cercano quedó fuera de la entrega: la huida ya lleva al refugio ante una amenaza.)
*Por qué:* la propuesta pide buscar «una alternativa conocida y accesible» y no insistir con la ruta cerrada; los reintentos y las ofertas existentes ya lo hacen. No hizo falta código nuevo.

**Cierre.** `tiny_returns_to_eat` (el hambre corta un acarreo, la carga queda en el piso, come y consume una porción del depósito), `tiny_rests_at_nest` (la energía sube) y `nest_unreachable_eats_crumb` (con la pata del escritorio cerrada, come una migaja conocida y no reintenta el nido antes de 30 s). Cambia el formato de `colony`: `HashVersion` y `SnapshotSchema`.

### M40 Puntos de búsqueda

**Percepción de lugares.** `ai` suma un store `landmark` (marcador vacío). `ai.look` aprende también las entidades con `landmark` a la vista, y `colony.share` las comparte como hoy.
*Por qué:* la propuesta pide que los tinies descubran puntos, servilleta y fibras con la percepción existente. Un marcador explícito evita darles `offers` falsas, que `use` intentaría usar.

**Componente `search_spot`.** Lleva `look_ticks` (cuánto tarda en revisarse), `cooldown` y `empty_until`. Va en una entidad ubicada en su superficie (rincón del escritorio, borde de la silla) con `landmark`.

**Acción `search`** (de `colony`, instalada como acción extra de `ai`, igual que `forage`):

1. Opciones: los puntos conocidos, no reservados, fuera de `empty_until` y sin peligro. Utilidad `member.search × déficit − distance_cost × distancia`, con `member.search` por debajo de `work` (0,2). Así un acarreo conocido siempre gana, como pide la propuesta («si ninguna está al alcance»).
2. `Start` reserva el punto (`jobs.Reserve`). Si no hay ruta, `nav.Approach` falla en el primer paso y el punto queda evitado 30 s.
3. Camina al punto, renovando la reserva, y espera `look_ticks`.
4. Termina al notar comida (`noticed` de algo con `food`, del propio actor, desde que empezó) o al acabar la espera. Emite `searched {entity, spot, found}`. Si no encontró nada, fija `empty_until`.

*Por qué:* la reserva es la que ya existe para migajas, así que «solo un tiny revisa cada punto» queda cubierto por el invariante de reservas. El hallazgo genera el acarreo con `colony.share`, sin cambios. Dejar que falle `nav.Approach` en vez de buscar la ruta al calcular opciones mantiene barato el selector (las opciones se calculan cada 10 ticks por tiny).

**Forrajeo.** `forage` sigue igual y con utilidad baja: es la exploración que descubre puntos nuevos. `encounter.foraging` cuenta también `search`, para que la giantess siga reconociendo a los tinies que buscan comida.

**Cierre.** `colony_searches_spots` (el ciclo completo hasta `stored` y un `eat`), `stocked_colony_searches_less` (mismo escenario con el depósito lleno: menos `searched`), `search_spot_reserved_once` (dos tinies, un punto) y `search_portal_closed` (se cierra la pata a mitad de la salida: falla una vez, libera la reserva y evita el punto). Store nuevo: `HashVersion` y `SnapshotSchema`.

### M41 Fibras y refugio auxiliar

**Entidades.**

- **Servilleta:** `fiber_source {left, item}`, con `landmark`. `item` es la definición de la fibra, expandida por `content` como en los pasos `spawn` (`"item": "fiber"`).
- **Fibra** (`content/objects/fiber.json`): `material {site}` y `landmark`; se reserva y se carga como cualquier objeto.
- **Obra:** `build_site {required, delivered, radius, done}`, con `landmark`.

**Trabajos.** `jobs` expone `RegisterKind` (en vez de escribir el mapa `Kinds` desde otro paquete) y `colony` registra:

- `pluck` (objetivo: la servilleta): ir, esperar `PluckTicks`, gastar una unidad y crear la fibra en el mismo paso, con el mismo mecanismo que el paso `spawn`. El que la desprende la aprende.
- `build` (objetivo: la obra): ir, esperar `BuildTicks` y terminar.

El acarreo es un `haul` con destino en la obra.
*Por qué:* la propuesta pide que cada encargo sea una entidad `job` que reserva su objetivo al asignarse. Reutilizar `haul` hereda la suspensión, la carga que cae al huir y la vuelta a la cola.

**Sistema `colony.build`** (fase Decide, junto a `share`). Para cada obra sin terminar cuya colonia conoce la servilleta:

- si la despensa está en la meta (`déficit == 0`), pide fibras mientras `required − delivered − fibras existentes de la obra − pluck abiertos > 0`;
- publica `haul` para cada fibra suelta sin trabajo;
- con `delivered == required`, publica un `build`.

**Entrega** (fase Consequences, como `colony.store`). Una fibra suelta, sin reservar y a menos de `radius` de su obra suma 1 a `delivered`, desaparece y emite `fiber_delivered {site, item, total}`.
*Por qué:* contar recién al soltar y desaparecer en el mismo paso hace que una fibra cuente una sola vez y que el total no dependa de quién la llevó. Contar las existentes antes de pedir otra cubre las fibras que quedaron en el piso.

**Terminar.** Al completar `build`, la obra recibe `shelter` (resistencia de la palma, radio de la obra), `done: true` y se emite `refuge_built {site, by}`. Desde ahí `escape` ya la elige.

**Invariantes.** `delivered ≤ required`; una obra con `done` tiene `shelter`; una fibra con `material` apunta a una obra viva.

**Cierre.** `colony_builds_refuge` (de servilleta a refugio, mismo ID), `fiber_dropped_is_recovered` (una sombra hace soltar una fibra en viaje; otro miembro la entrega y el total no cambia) y `refuge_shelters_from_palm` (un tiny dentro del refugio emite `sheltered`, sin `crushed`). Stores nuevos: `HashVersion` y `SnapshotSchema`.

### M42 Reunión

**Diversión.** Los tinies de los escenarios del slice suman `fun` con decaimiento bajo. El punto de reunión es su única fuente.

**Componente `gathering`.** Lleva `wait` (plazo para que llegue el segundo), `talk_ticks`, `pause`, `next` (cuándo vuelve a ofrecerse) y `seats []{by, arrived}` con a lo sumo dos.
*Por qué:* `reservation` admite un solo dueño. Las plazas viven en el componente del punto, que es estado de simulación (entra en el hash y el snapshot), con su invariante: cada plaza es de un agente vivo cuya acción es `chat` sobre ese punto.

**Acción `chat`** (de `colony`):

- Opciones: con hambre y energía por encima de un umbral, sin trabajo asignado, el punto conocido, con plaza libre y fuera de la pausa. La utilidad sale de la curva de `fun`, con el mismo cálculo que una oferta.
- Pasos: tomar plaza → caminar → esperar al compañero hasta `wait` → conversar `talk_ticks` → `fun` sube a los dos.
- La conversación termina por el tick guardado en el punto, no por cada agente, así los dos terminan en el mismo tick sin depender del orden del store.
- Si el compañero no llega, se va o muere, la acción termina y libera su plaza. Con la última plaza libre empieza la pausa.

Eventos: `chat_started {gathering, a, b}`, `chat_done {gathering, a, b}` y `chat_abandoned {gathering, entity, reason}`.

**Cierre.** `tinies_chat` (la diversión de los dos sube), `chat_partner_never_comes` (uno espera solo, vence el plazo y libera el punto) y `chat_interrupted_by_threat` (una sombra corta la conversación, los dos huyen y el punto queda libre). Store nuevo: `HashVersion` y `SnapshotSchema`.

### M43 Partida libre

**Contenido.** Un script nuevo, `tools/scenarios/vida.py`, que corre después de `partida.py`, agrega a `content/partidas/casa.json`:

- dos o tres puntos de búsqueda (rincón del escritorio, borde de la silla, uno en el piso del dormitorio);
- una servilleta en el escritorio;
- la obra junto al extremo del portal `leg1_top` (la ruta de salida), lejos del radio del nido;
- el punto de reunión junto al nido;
- migajas con `food` cerca de la silla, sin que nadie las conozca;
- pocas porciones en el nido.

El mismo script genera los escenarios del slice.
*Por qué:* el contenido sigue la regla de escenarios generados por script, y `partida.py` queda como está.

**Sorteo.** `Randomize` agrega al final del sorteo de la colonia la diversión inicial y la despensa.
*Por qué:* agregar al final no cambia lo ya sorteado de las partidas existentes (ver [encuentros](specs/encuentros.md#partida-libre)).

**Ajuste.** Se ajustan `curiosity`, `member.search`, el umbral de `chat` y los decaimientos sobre las semillas de referencia (1, 7, 21, 42 y 99). Es probable que la curiosidad actual (0,8) tape todo lo demás.

**Compatibilidad.** `TestSeatsAcceptance` y la aceptación de encuentros corren sobre `casa.json`. Si cambian, se revisa que sigan cumpliendo sus criterios y se actualizan a propósito, sin relajar el criterio. La mañana y el descuido no cambian.

**Cierre.** Prueba de rutina sobre las semillas: en los primeros 6000 ticks hay al menos tres actividades distintas completadas entre `search`, acarreo, `eat`, `rest`, `pluck`/`build` y `chat`. La continuidad reutiliza `playChecked` y elige snapshots que caen a mitad de un `search` y de la obra (con `delivered` entre 1 y `required − 1`).

### M44 Presentación y cierre

- **Ficha.** `view.Thing` suma el destino (Space y punto) y el motivo de la acción en curso: qué punto revisa, adónde lleva lo que carga, y si va al nido por hambre o por cansancio (`use:eat`/`use:rest` en el paso `goto`). `cmd/game/hud.go` los traduce: «Revisa el rincón del escritorio», «Lleva una fibra a la obra», «Vuelve al nido a comer», «Conversa con Tiny 04».
- **Micro.** Símbolos sobre el tiny para buscar, volver con hambre o cansancio y conversar (clips de emote en `content/art/src/casa/`). Arte de servilleta, fibra, obra en etapas y refugio terminado: piezas de código en `tools/art/` o cuadros en `content/art/src/casa/` y `go run ./cmd/artpack`.
- **Historia.** `view.storyText` suma `searched` con hallazgo, `fiber_delivered` (agrupado), `refuge_built` y `chat_done`. Una búsqueda vacía no aparece.
- **Sonido.** Roce de papel al desprender o entregar una fibra y murmullo breve en la conversación, en `tools/audio/sounds.json`. Solo presentación.
- **Menú.** La nota de «Nueva partida» menciona la vida del nido.
- **Referencias.** Las capturas y la mezcla usan la mañana y no deberían cambiar. Si cambian, se regeneran a propósito.

**Aceptación.** Una prueba recorre los criterios de [16.4](propuesta.md#164-cierre-de-la-iteración):

- abastecimiento: el ciclo completo y menos salidas con la despensa llena;
- rutina y ocio: tres actividades por semilla, el descanso que sube la energía, la conversación que sube la diversión y la espera que vence;
- coordinación: ninguna reserva doble de punto, migaja o fibra; cada fibra cuenta una vez; la obra conserva su ID y protege;
- interrupciones: la huida en un acarreo o una conversación, el portal cerrado a mitad de una salida y las reservas liberadas;
- lectura y continuidad: los textos de actividad distinguen búsqueda, acarreo, regreso y reunión, y el mismo log y hash al continuar desde snapshot.

## Contratos de integración

- Todo se activa por contenido: sin `search_spot`, `build_site` ni `gathering`, una colonia se comporta como hoy. La mañana, el descuido y sus capturas no cambian.
- Las acciones nuevas siguen el contrato de `ai.Action`: sin estado propio, todo en `agent.doing` o en el componente del objetivo. Mover con `nav.Approach`/`nav.Halt` y reservar con `jobs.Reserve`.
- Orden estable (D5): las plazas y las listas de los componentes nuevos se guardan ordenadas. Si hay que sortear, se usa un stream propio (`colony.search`).
- Cada hito que agrega o cambia un store incrementa `HashVersion` y `SnapshotSchema`, cierra con escenarios en `tests/scenarios/` y deja la spec `vida.md` al día. Antes de cada commit: `gofmt -l .`, `go vet ./...` y `go test ./...`.
- La presentación lee `view.Frame` y los eventos. Nada en `render`/`ui` lee el mundo.

Quedan fuera, como dice la propuesta: guardia y avisos entre tinies, depósitos en otras habitaciones, expediciones en la palma y gotas junto a la taza.
