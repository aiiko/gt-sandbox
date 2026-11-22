# Spec: Necesidades, IA, smart objects, trabajos, colonias y exploración (M7–M10)

Paquetes: `sim/needs`, `sim/jobs`, `sim/ai`, `sim/colony`, `content` (datos) e `internal/content` (carga). Es un esqueleto: la infraestructura está y el contenido es mínimo.

## Necesidades (`needs`)

- Componente `needs`: lista `[{name, value, decay}]`. `value` es un entero en [0, 1000] (1000 = satisfecha). Baja `decay` cada `Interval` (1 s) en `needs.decay` (fase Slow), sin pasar de 0.
- `set_need {entity, need, value}` para debug y escenarios.

## Smart objects (`offers`)

Componente de un objeto: lista de ofertas `{action, scale, effects: [{need, amount}], duration, approach, facing, steps}`.

- `scale` dice qué agentes pueden usarla (`giantess`, `tiny`).
- `approach` y `facing` son el punto de uso y la orientación, relativos a la posición del objeto en su Space.
- Sin `steps`, la oferta es "ir al punto de uso y usar durante `duration`".

Las definiciones viven en `content/objects/<nombre>.json` (embebidas en el binario) y los escenarios las usan con `"object": "<nombre>"`; el `spawn` de la entidad reemplaza campos y componentes de la definición. Campos, componentes o pasos desconocidos son error al cargar.

### Pasos (`steps`)

Una oferta de varios pasos se describe con datos. Los pasos instantáneos se encadenan en el mismo tick.

| Paso | Qué hace |
|---|---|
| `goto` | Camina al punto de uso, o con `place: "<tag>"` al lugar con esa etiqueta más cercano (componente `place {tags, spot, facing}`). Al llegar toma `facing`. |
| `wait` | Espera `ticks`. |
| `use` | Aplica los efectos de a poco durante `ticks`; al terminar suman exactamente `amount`. Tiene que haber exactamente uno. |
| `pose` | Pone al actor en la pose `pose` (`attach`). Al terminar o interrumpirse vuelve a `agent.idle_pose`. |
| `take` | Crea `item` y lo carga (`carried` + reserva). |
| `drop` | Deja lo cargado en `offset` (`item_dropped`). |
| `spawn` | Crea `count` copias de `item` en un anillo de radio `spread` alrededor de `offset`. |
| `consume` | Gasta `amount` porciones del objeto (`portions {left}`); en 0, o si no tiene porciones, el objeto desaparece (`consumed`), salvo con `keep` (un depósito). Si no quedan porciones suficientes, la oferta no es una opción. |

- `offset` es relativo a la mano del actor (`agent.hand`, un Space montado sobre él) o, si no tiene, al actor, y gira con él. Lo que se deja o se crea cae enseguida a la superficie de abajo (`space.Land`, que emite `entity_fell`): así la taza y las migajas que suelta la mano apoyada terminan sobre el escritorio.
- En el JSON de `content`, `"item": "<nombre>"` se reemplaza por la definición completa de ese objeto (con detección de ciclos). La simulación recibe datos autocontenidos y no conoce la biblioteca.
- Si un paso `goto` necesita un lugar que no existe, la oferta no es una opción.
- `Offers.Validate` verifica los pasos; lo usan el cargador de `content` y el invariante de `ai`.

Ejemplo: `breakfast` de la heladera va a la heladera, espera, saca una taza, va al escritorio, apoya la mano, come, deja la taza y tres migajas, y vuelve a la pose de reposo.

## Agentes (`agent`) y selector

Componente `agent`: `scale`, `weights` (importancia de cada necesidad, por defecto 1), `work` (ganas de trabajar), `distance_cost` (utilidad que se pierde por metro), `inertia` (por defecto 0,15), `hand` y `idle_pose`, `sight`/`known`/`watch` (percepción), `curiosity`/`visited` (exploración), y el estado `doing` (con el lugar elegido, lo que carga y si cambió de pose), `suspended` y `recover_until`.

- **Utilidad de usar una oferta:** `Σ w · (d(v) − d(v + amount)) − distance_cost · distancia` (al primer punto de uso), con `d(v) = ((1000 − v)/1000)²`. La curva cuadrática hace que una necesidad baja pese mucho más. Seguir con una acción en curso vale lo que falta de su efecto.
- **Sin desperdicio:** una oferta cuyos efectos se desperdiciarían todos en más de la mitad no es una opción (no se come sin hambre aunque no haya nada mejor que hacer).
- **Compromiso:** una oferta con la mano apoyada (`pose`) o algo en la mano (`take`) solo la interrumpe una opción urgente (utilidad ≥ `UrgentUtility`, 0,5). Sin esto, la ganancia que le falta baja mientras se usa y cualquier otra cosa la corta a medias (taza en el piso). Una acción lo declara con el método opcional `Committed`.
- **Cierre:** una oferta se puede interrumpir hasta terminar su paso `use`; lo que sigue (dejar cosas) se completa siempre. Al interrumpirse, el actor deja lo que carga a sus pies y vuelve a la pose de reposo.
- **Utilidad de un trabajo:** `work + 0,01 · priority − distance_cost · distancia`.
- **Decisión** (`ai.decide`, fase Decide): cada agente decide cada `DecidePeriod` (10 ticks), repartido por índice de entidad. Junta las opciones de todas las acciones y elige la mejor; los empates se resuelven por acción, objetivo y oferta. Si no hace nada, empieza la mejor siempre que llegue a `MinUtility`. Si ya hace algo, solo cambia cuando la nueva supera a la actual por más que `inertia` (histéresis). La acción reemplazada se termina con `action_interrupted`.
- **Ejecución** (`ai.act`, fase Act, antes de `nav.plan`): un paso por tick. Emite `action_started`, `action_done` y `action_failed`. Los eventos de acción llevan `what` (la oferta, para `use`), y `action_interrupted.by` dice qué la reemplazó.
- **Interfaz `Action`:** `Options`, `Utility`, `Start`, `Step`, `Stop`, `Valid` e `Interruptible`. Las implementaciones no guardan estado propio: todo va en `Doing`, que se guarda y entra en el hash.
- **Acciones del núcleo:**
  - `use`: reserva el objeto, camina al punto de uso y aplica los efectos de a poco durante `duration`; al terminar suman exactamente `amount`.
  - `work`: toma un trabajo abierto y lo ejecuta.
  - `explore`: cruza un portal conocido hacia un Space donde nunca estuvo y da una vuelta allí (ver Exploración).
- **Movimiento desde sistemas:** `nav.Approach` pide caminar y avisa si el agente llegó o si falló. `nav.Halt` lo detiene.

## Percepción y memoria

- Un agente con `sight > 0` solo usa los smart objects que conoce (`agent.known`). Sin `sight` conoce todo (la giantess, por ahora).
- `ai.look` (fase Decide, antes del selector) mira cada `LookPeriod` (10 ticks) con el índice espacial: los objetos con `offers` a distancia ≤ `sight` en su mismo Space. Algo nuevo se agrega con `noticed {entity, target}`. Olvida lo que ya no existe y, pasado `MaxKnown` (64), lo más viejo.
- `ai.Learn` hace que un agente conozca algo sin verlo (se lo contaron).
- `ai.look` también registra el Space donde está el agente (`agent.visited`) y aprende los portales que tienen un extremo en su Space a distancia ≤ `sight`.
- Un agente con `watch > 0` nota (`noticed`) a los agentes de otra escala a menos de `watch` en el mundo, aunque estén en otro Space. Para la giantess, que conoce todo, `known` funciona como registro de a quién vio; por ahora no cambia su plan.
- Los trabajos no dependen de la memoria: la cola es conocimiento compartido.

## Exploración

- `agent.curiosity` es la utilidad de explorar. Las opciones de `explore` son los portales conocidos y habilitados con un extremo en el Space del agente y el otro en un Space no visitado. Si el portal es de contacto (`attach`, por ejemplo la mano apoyada) la utilidad se multiplica por `DynamicPortalX` (2). Se le resta `distance_cost` por la distancia al extremo propio.
- La acción camina al otro extremo (cruzando el portal). Al llegar registra el Space, emite `explored {entity, space}` y da una vuelta a un punto transitable al azar a menos de `ExploreRange` (10 cm, RNG `ai.explore`). Si la vuelta no se puede, igual termina bien.
- Los portales de contacto se crean de nuevo en cada apoyo y se borran al levantar la mano; la memoria olvida los borrados.
- Si la mano se levanta con el tiny encima, el tiny sigue en la palma: es su Space, ya visitado, y espera a que aparezca un portal hacia un lugar nuevo.
- `ai.WanderPoint` (punto transitable al azar cercano) lo comparten `explore` y `forage`.

## Huida

Ver `docs/specs/crush.md#huida-m25-en-simai`. Un agente con `reflex` se alarma bajo una amenaza de `crush.Threats` (`alarmed`), huye fuera de la huella (`flee`) o queda atrapado (`trapped`). `near_miss` interrumpe como una caída. Estado: `agent.reflex`, `alarm` y `trapped`.

## Cuidado y zonas de peligro

Ver `docs/specs/crush.md`. `agent.care` hace que un paso `pose` que apoya la mano espere y pueda titubear (`hesitated`, `Doing.wary`). El store `danger` marca zonas que los agentes de una escala evitan como destino (opciones, `WanderPoint`, `explore`) hasta que vencen.

## Reintentos

Cuando una acción falla, el agente evita ese objetivo durante `RetryTicks` (30 s): `agent.avoid`. Vale para cualquier acción; un trabajo inalcanzable vuelve a la cola y el que falló no lo retoma enseguida.

## Colonias (`colony`)

- La colonia es una entidad en el nido con `colony {radius}`. También es un smart object: `portions` es su depósito, y sus ofertas son comer del depósito (`consume` con `keep`) y descansar. Definición en `content/objects/nest.json`.
- Los miembros tienen `member {colony, forage}`.
- `colony.share` (fase Decide, entre `ai.look` y el selector):
  - cada miembro conoce su nido;
  - lo que un miembro notó en el tick anterior lo aprenden los demás;
  - si es comida (`food`), no está en el nido y nadie la lleva ni tiene un trabajo sobre ella, la colonia publica un `haul` al nido.
- `colony.store` (fase Consequences): la comida suelta, sin reservar y a menos de `radius` del nido se guarda; suma sus porciones (o 1) al depósito y desaparece (`stored {colony, item, portions, total}`).
- Acción `forage` (de `colony`, instalada como acción extra de `ai`): caminar a un punto transitable al azar a menos de `ForageRange` (30 cm), con el RNG `colony.forage`. Vale `member.forage`, así que solo gana cuando no hay nada mejor. Así se encuentran cosas fuera de la vista.

## Suspensión y reanudación

Hay eventos que interrumpen: `entity_fell` y `near_miss`. Cuando le pasa a un agente, el selector lo lee en el tick siguiente, detiene su movimiento y suspende la acción (`action_suspended {cause}`). El agente se recupera durante `RecoverTicks` (2 s) y después la retoma (`action_resumed`) si sigue siendo válida. La etapa del trabajo vive en el trabajo, así que continúa donde quedó, y la reserva se mantiene porque dura más que la recuperación.

## Reservas y trabajos (`jobs`)

- **Reservas:** `reservation {by, until}` es un componente de la cosa reservada, así que no puede haber dos. `Reserve` reserva o renueva por `TTL` (10 s); quien la usa la renueva en cada paso. `jobs.upkeep` (fase Consequences) libera las que vencieron (`expired`) o cuyo dueño murió (`holder_gone`). Eventos: `reserved` y `reservation_released {reason}`.
- **Trabajos:** cada trabajo es una entidad con `job {type, target, dest_space, dest, priority, assignee, stage, wait}`.
  - `post_job` (si no se indica `dest_space`, se usa el Space del objetivo) y `cancel_job`.
  - Tomar un trabajo exige reservar su objetivo.
  - Si el objetivo desaparece, el trabajo se cancela (`job_canceled`).
  - Si el asignado muere o pierde la reserva, el trabajo vuelve a la cola (`job_unassigned`), soltando lo que llevaba.
  - Otros eventos: `job_posted`, `job_assigned` y `job_done`.
- **Tipos de trabajo:** son `Executor` registrados en `jobs.Kinds`. `haul` va al objetivo, lo levanta (1 s), lo lleva al destino y lo deja (1 s).
- **Carga:** `carried {by}` hace que el objeto viaje con quien lo lleva (`jobs.carry`, al final de Move).

## Invariantes

- Un trabajo asignado tiene su objetivo reservado por el asignado.
- Lo que se carga está reservado por quien lo carga.
- Las necesidades están en rango, con nombres únicos.
- Toda acción tiene nombre.

## Escenarios

- `tinies_share_crumbs` (aceptación): tres tinies y dos migajas con trabajos de acarreo. Cada migaja se reserva una sola vez y el tercer tiny no hace nada (`TestCrumbReservedOnce`).
- `giantess_needs_stable` (aceptación): giantess con hambre, energía y diversión, y una heladera, una cama y una tele. Usa las tres. `TestStableChoice` verifica que no oscila y que, con `inertia: 0`, sí oscila; eso comprueba que el test detecta la oscilación.
- `job_resumes_after_fall` (aceptación): el tiny toma el trabajo, la palma desaparece y el tiny cae. El trabajo se suspende, se retoma y termina, sin que la reserva venza.
- `giantess_breakfast` (aceptación de M8): casa con cocina, dormitorio y living. Con hambre alta, la giantess desayuna en el escritorio con la mano apoyada (portal mano↔escritorio), deja la taza y tres migajas sobre el escritorio, levanta la mano y no repite el desayuno sin hambre. `TestBreakfastInterrupted` la interrumpe mientras come: la mano se levanta y la taza queda en el piso, sin carga ni reserva.
- `tiny_eats_crumb` (aceptación de M8): un tiny come una migaja de tres porciones hasta que desaparece.
- `colony_gathers` (aceptación de M9): seis tinies, un nido y ocho migajas fuera de su vista. Salen a buscar, la colonia acarrea lo que ven y comen del depósito. `TestColonyFeeds` verifica que no hay trabajos sobre algo que nadie vio, que nadie usa algo que no conoce, que se guarda comida y que en 10 minutos nadie llega a hambre 0.
- `unreachable_food_retry` (aceptación de M9): una migaja encerrada; el tiny toma el trabajo, falla y lo vuelve a tomar recién 30 s después (exactamente dos asignaciones en 50 s).
- `tiny_rides_hand` (aceptación de M10): dormitorio, living y cocina, cada uno con su escritorio salvo la cocina. La giantess escribe con la mano apoyada en el dormitorio y un tiny curioso sube a la palma. Después ella desayuna y come en el escritorio del living; el tiny baja y lo explora. Ella nota al tiny una vez y no cambia de plan; el tiny no cae. `TestTinyRidesHand` verifica que recuerda el escritorio nuevo y que sin curiosidad no sube. `TestBreakfastCommitted` arranca con más hambre: el desayuno corta la escritura, pero no al revés.

## Pendiente

- Percepción: solo un radio y solo el propio Space (no se ve de la palma al escritorio). Lo que se ve se comparte con la colonia de inmediato, no al volver al nido. Personalidad más allá de pesos. Curvas de utilidad por contenido.
- Más eventos que interrumpan, y prioridades de emergencia (huir) por encima de la utilidad.
- Compromiso: hoy es un umbral fijo (`UrgentUtility`). Retomar una oferta a medias (la taza que quedó en el piso) en vez de empezar de nuevo.
- Exploración: la curiosidad no decae con lo visitado ni se cansa; un Space visitado no vuelve a interesar nunca. El tiny no comparte con la colonia lo que explora.
- Lugares reservables (hoy dos agentes pueden elegir el mismo escritorio).
- `spawn` en anillo fijo: variar posiciones con RNG cuando importe.
