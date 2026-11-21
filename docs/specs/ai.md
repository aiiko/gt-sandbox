# Spec: Necesidades, IA, smart objects y trabajos (M7)

Paquetes: `sim/needs`, `sim/jobs`, `sim/ai`, `content` (datos) e `internal/content` (carga). Es un esqueleto: la infraestructura está y el contenido es mínimo.

## Necesidades (`needs`)

- Componente `needs`: lista `[{name, value, decay}]`. `value` es un entero en [0, 1000] (1000 = satisfecha). Baja `decay` cada `Interval` (1 s) en `needs.decay` (fase Slow), sin pasar de 0.
- `set_need {entity, need, value}` para debug y escenarios.

## Smart objects (`offers`)

Componente de un objeto: lista de ofertas `{action, scale, effects: [{need, amount}], duration, approach}`. `scale` dice qué agentes pueden usarla (`giantess`, `tiny`). `approach` es el punto de uso, relativo a la posición del objeto en su Space. Las definiciones viven en `content/objects/<nombre>.json` (embebidas en el binario) y los escenarios las usan con `"object": "<nombre>"`; el `spawn` de la entidad reemplaza campos y componentes de la definición. Campos o componentes desconocidos son error al cargar.

## Agentes (`agent`) y selector

Componente `agent`: `scale`, `weights` (importancia de cada necesidad, por defecto 1), `work` (ganas de trabajar), `distance_cost` (utilidad que se pierde por metro), `inertia` (por defecto 0,15), y el estado `doing`, `suspended` y `recover_until`.

- **Utilidad de usar una oferta:** `Σ w · (d(v) − d(v + amount)) − distance_cost · distancia`, con `d(v) = ((1000 − v)/1000)²`. La curva cuadrática hace que una necesidad baja pese mucho más. Seguir con una acción en curso vale lo que falta de su efecto.
- **Utilidad de un trabajo:** `work + 0,01 · priority − distance_cost · distancia`.
- **Decisión** (`ai.decide`, fase Decide): cada agente decide cada `DecidePeriod` (10 ticks), repartido por índice de entidad. Junta las opciones de todas las acciones y elige la mejor; los empates se resuelven por acción, objetivo y oferta. Si no hace nada, empieza la mejor siempre que llegue a `MinUtility`. Si ya hace algo, solo cambia cuando la nueva supera a la actual por más que `inertia` (histéresis). La acción reemplazada se termina con `action_interrupted`.
- **Ejecución** (`ai.act`, fase Act, antes de `nav.plan`): un paso por tick. Emite `action_started`, `action_done` y `action_failed`.
- **Interfaz `Action`:** `Options`, `Utility`, `Start`, `Step`, `Stop`, `Valid` e `Interruptible`. Las implementaciones no guardan estado propio: todo va en `Doing`, que se guarda y entra en el hash.
- **Acciones del núcleo:**
  - `use`: reserva el objeto, camina al punto de uso y aplica los efectos de a poco durante `duration`; al terminar suman exactamente `amount`.
  - `work`: toma un trabajo abierto y lo ejecuta.
- **Movimiento desde sistemas:** `nav.Approach` pide caminar y avisa si el agente llegó o si falló. `nav.Halt` lo detiene.

## Suspensión y reanudación

Hay eventos que interrumpen: por ahora `entity_fell`. Cuando le pasa a un agente, el selector lo lee en el tick siguiente, detiene su movimiento y suspende la acción (`action_suspended {cause}`). El agente se recupera durante `RecoverTicks` (2 s) y después la retoma (`action_resumed`) si sigue siendo válida. La etapa del trabajo vive en el trabajo, así que continúa donde quedó, y la reserva se mantiene porque dura más que la recuperación.

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

## Pendiente

- Percepción: hoy los agentes conocen todas las ofertas y trabajos. Personalidad más allá de pesos. Curvas de utilidad por contenido.
- Más eventos que interrumpan, y prioridades de emergencia (huir) por encima de la utilidad.
- Reintentos con espera: un trabajo inalcanzable vuelve a la cola y puede volver a elegirse enseguida.
- Ofertas que consumen el objeto (comerse la migaja) y efectos que no son necesidades.
