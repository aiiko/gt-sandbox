# Spec: Contactos, portales dinámicos, poses y caídas (M5)

Paquetes: `sim/attach` (contactos, poses), `sim/space` (caídas).

## Superficies de contacto (`contact_surface`)

Componente de un Space: `{tolerance, min_overlap}` (0 = por defecto: 3 cm y 1 cm).

Dos superficies **se tocan** si:
1. están en la misma raíz;
2. sus alturas en el mundo (Z del marco del Space) difieren a lo sumo en la mayor de sus `tolerance`;
3. las cajas alineadas a los ejes de sus extensiones, llevadas al mundo, se solapan.

Con rotaciones arbitrarias la caja es una sobreestimación; el punto del portal sí se verifica contra ambas extensiones.

## Contactos (`contact`)

El sistema `attach.contacts` (fase Topology) recalcula todos los pares cada tick (O(n²) en superficies, que son pocas).

- Cada contacto activo es una **entidad** con `contact {a, b, since}` (`a < b` por ID). `contact_began` / `contact_ended`.
- Si el solapamiento mide al menos `min_overlap` en cada lado y hay un punto del solapamiento dentro de ambas extensiones y transitable en ambas grillas, **la misma entidad** lleva un `nav.Portal` entre esos puntos: `portal_opened` / `portal_closed`. Se prueba el centro y luego 25 muestras, de la más cercana al centro a la más lejana.
- El portal es estable: solo se recalcula (y sube `version`, lo que invalida las rutas que lo usan) si sus extremos se separan en el mundo más de `PortalDrift` (5 mm), p. ej. si la mano se desliza.
- Cuando el contacto termina, la entidad se destruye: las rutas que la usaban se replanean y, si no hay alternativa, fallan con `move_failed`.
- `attach.cleanup` (último del Flush) termina, con sus eventos, los contactos cuyas superficies destruyó un sistema posterior a Topology.

## Poses (`poses`, `set_pose`)

- `poses` es un componente del cuerpo: `{"nombre": [{part, pos, facing}]}`. Una pose es solo un conjunto de Locations de partes (sin esqueleto).
- `set_pose {entity, pose}` cambia la `pos`/`facing` de cada parte respecto de su Space contenedor, que no cambia. Todas las partes deben estar montadas, directa o indirectamente, sobre el cuerpo; si no, se rechaza sin modificar nada. Emite `pose_changed`. Es instantáneo (la interpolación llega con la animación).
- `set_components {entity, components}` (en `cmds`) agrega componentes por nombre de store a una entidad existente; sirve para referencias circulares en escenarios (las poses nombran partes que cuelgan del cuerpo).

## Caídas (`space.falls`)

Fase Consequences. Una entidad física (con `body`, que no sea un Space) está **apoyada** si su posición local está dentro de la extensión de su Space y a `|Z| ≤ 1 mm`. Si no lo está, aterriza en la superficie más alta que tenga debajo: un Space de la misma raíz, no montado sobre ella, con extensión no vacía que contenga su vertical, y cuya superficie esté a su altura o más abajo. Se reubica con `Z = 0` y conserva su posición en el plano y su orientación en el mundo. Emite `entity_fell {entity, from, to, height}`.

Si no hay nada debajo, se queda donde está. La caída es instantánea; el daño y la animación son de otros sistemas.

Caso típico: se destruye la palma → `space.orphans` mueve al tiny al marco de la giantess (sin extensión, así que no lo apoya) → al tick siguiente cae sobre el escritorio o el suelo.

## Escenarios

- `giantess_carries_tiny` (aceptación): el tiny camina del escritorio a la palma apoyada; la giantess levanta la mano, camina a la otra habitación (cruzando la puerta), la apoya en otro escritorio y el tiny baja. Solo comandos `set_pose`, `move_to` a la giantess y dos `move_to` al tiny (ir a la palma, bajar al escritorio).
- `contact_opens_and_closes_route`: sin contacto no hay ruta; al apoyar la mano aparece; si se levanta mientras el tiny va hacia ella, el movimiento falla con `no_path`. Destruir la palma apoyada termina el contacto.
- `palm_destroyed_tiny_falls`: caída de 65 cm sobre el escritorio.

## Pendiente

- Inclinación de Spaces (y caída por inclinación más allá de un límite): las transformaciones todavía solo rotan sobre Z.
- Anchors con nombre; evento de movimiento brusco de un Space para el equilibrio de los tinies.
- Reglas de contacto→portal por contenido (hoy la regla es única: tocar + solapamiento mínimo).
