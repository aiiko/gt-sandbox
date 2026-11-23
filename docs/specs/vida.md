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
