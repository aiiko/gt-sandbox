# Spec: Aplastamiento (M23–M24)

Paquete: `sim/crush`. Diseño y alcance del slice en la [propuesta](../propuesta.md#14-próximo-slice-un-descuido-en-el-escritorio).

Lo que baja sobre una superficie mata a los tinies frágiles que quedan debajo, salvo que estén en un refugio que resista. El aplastamiento es instantáneo: no hay daño parcial y las caídas no matan.

## Componentes

| Store | Datos | De quién |
| --- | --- | --- |
| `crusher` | `{source, force, thickness, radius, pressing}` | Aplastador: la palma (`source: "hand"`), un objeto (`"object"`). `pressing` es estado |
| `fragile` | `{}` | Quien muere aplastado (los tinies) |
| `shelter` | `{strength, radius}` | Refugio: el nido (`content/objects/nest.json`, fuerza 50, 2 cm) |
| `remains` | `{name, of, by, source, tick}` | Restos de un aplastado |
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
- `TestCrushOnlyOnNewPress`: un tiny camina debajo de la palma apoyada y no muere, tampoco después de guardar y cargar. El mundo restaurado tiene el mismo hash.

## Pendiente

- Debajo de una mano apoyada se puede caminar: la grilla del escritorio no sabe que la mano ocupa ese lugar.
- Los restos no saben a qué colonia pertenecía la víctima. Llega en M26 con `member_lost`.
- Amenaza y huida (M25), cuidado de la giantess (M26).
- Pisadas: sin volumen barrido, sin ritmo propio (dependen solo de la distancia) y sin pisar cuando la giantess gira en el lugar. La casa todavía no tiene `stride` ni ruta al suelo: llega con el escenario de cierre (M27).
- Fuerza de los objetos por masa y altura de caída: hoy es un valor fijo por objeto.
