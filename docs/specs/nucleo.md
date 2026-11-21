# Spec: núcleo de simulación (M0–M1)

Paquetes: `core/{units,ids,rng,statehash}`, `sim/{clock,ecs,world,comp,cmds,engine}`.

## Responsabilidades

| Paquete | Responsabilidad |
|---|---|
| `world` | Estado autoritativo, loop `Step()`, cola de comandos, bus de eventos, hash, snapshot, replay. |
| `ecs` | Entidades (IDs generacionales) y stores de componentes (sparse sets). Despawn diferido. |
| `comp` | Componentes básicos: `Location`, `Body`, `Name`. |
| `cmds` | Comandos genéricos: `spawn`, `despawn`, `set_location`; evento `entity_spawned`. |
| `engine` | Arma la forma estándar del mundo. Único punto de construcción. |

## Un tick

```
Step(t):
  1. aplicar comandos encolados (orden de secuencia) → pueden emitir eventos
     · si Apply falla → evento command_rejected
  2. sistemas por (fase, Order, registro), si Due(t, Period, Offset)
  3. ecs.Flush(): destruir entidades marcadas → evento entity_despawned
  4. publicar eventos de t (visibles vía Events() hasta el próximo Step)
  5. t++
```

Los sistemas leen en `Events()` los eventos del tick anterior. No hay reentrancia ni suscriptores con callbacks.

## Contratos

- **Comandos:** única vía de entrada externa. `Enqueue` es seguro entre goroutines; aplica en el próximo tick. Un `Apply` con error no modifica nada.
- **Eventos:** solo se emiten dentro de `Step`, el tipo debe estar registrado.
- **Componentes:** datos puros serializables a JSON; el hash se deriva del JSON (ningún campo queda afuera por olvido).
- **Iteración:** `Store.Each` en orden denso, determinista. Durante `Each` se puede mutar el valor, marcar despawns y hacer `Set` sobre entidades que ya tienen el componente; agregar/quitar de ese store hace panic.
- **Sistemas con estado:** implementan `world.Stateful`; su estado entra en hash y snapshot.

## Invariantes

1. Ningún store contiene componentes de entidades muertas (`CheckInvariants`).
2. Un ID viejo (otra generación) nunca accede a datos de la entidad que reutilizó su slot.
3. Misma forma + semilla + comandos ⇒ mismo `Hash()` en cada tick.
4. `Save` → `Restore` ⇒ mismo `Hash()`; continuar ⇒ mismo `Hash()` que sin guardar.
5. `Save` solo entre ticks (sin despawns pendientes).

## Snapshot

JSON con `schema`, semilla, tick, estado de RNG, entidades (`gens`/`alive`/`free`), componentes por store (en orden denso), estado de sistemas, eventos publicados del último tick, comandos pendientes. Restaurar requiere un mundo recién creado con la misma forma; diferencias de forma son error (las migraciones van en `ReadSnapshot` cuando cambie `SnapshotSchema`).

## Replay

`SetRecording(true)` registra cada comando aplicado con su tick. `world.Replay(w, log, until)` lo reproduce; saltea registros anteriores al tick actual (para continuar desde un snapshot) y exige que no haya comandos encolados.

## Casos límite cubiertos por tests

Despawn doble, despawn durante `Each`, `Set` sobre muerta (panic), wrap de generación, snapshot con comandos pendientes, campos JSON desconocidos, stores/sistemas faltantes o sobrantes, log desordenado.
