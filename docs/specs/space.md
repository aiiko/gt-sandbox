# Spec: Spaces y transformaciones (M2)

Paquetes: `core/fixedmath`, `sim/space`.

## Modelo

- Un **Space** es una entidad con el componente `space.Space` (`kind`, `extent`). Kinds: `static`, `elevated`, `rigid`, `articulated`, `body`, `built`.
- **La jerarquía se arma solo con `comp.Location`** (`space`, `pos`, `facing`). Un Space con Location está montado en otro y se mueve con él. Un Space sin Location es una raíz.
- Ejemplo: `casa (raíz) > habitación > giantess (body) > palma (articulated) > tiny`. Mover a la giantess cambia solo su Location; la palma y el tiny no se tocan.
- Una pose (M5) consistirá en cambiar las Locations de los Spaces del cuerpo.

## Matemática

- `fixedmath.Angle`: 1/4096 de vuelta, antihorario sobre Z. Seno por tabla generada (`go generate ./internal/core/fixedmath`), nunca `math.Sin` en runtime.
- `Transform{Pos, Rot}`: rota sobre Z y traslada. Z solo se traslada (no hay inclinación todavía).
- Rotaciones de múltiplos de 90° exactas; las demás redondean al µm.

## API (`space.Of(registry)`)

| Función | Qué hace |
|---|---|
| `Frame(e)` | Transform de las coordenadas locales de `e` a las de su raíz. |
| `ToWorld(s, p)` / `FromWorld(s, p)` | Conversión de puntos entre un Space y su raíz. |
| `WorldPose(e)` | Posición y orientación de `e` en su raíz. |
| `CanPlace(e, target)` | `target` es un Space vivo y no es `e` ni está montado sobre `e`. |
| `Reparent(e, target)` | Mueve `e` a `target` conservando posición y orientación en el mundo. |
| `Root(e)`, `Parent(e)`, `IsSpace(e)` | Consultas de jerarquía. |

Las transformaciones se calculan bajo demanda, en O(profundidad). Todavía no hay caché.

## Invariantes (registradas en `World.CheckInvariants`)

1. Toda Location apunta a un Space vivo.
2. La jerarquía no tiene ciclos (profundidad ≤ `MaxDepth`).
3. Reparent entre raíces distintas es error.

## Destrucción de un Space

El sistema `space.orphans` corre último en la fase Flush. Mueve los ocupantes de cada Space que va a destruirse al ancestro vivo más cercano, sin moverlos en el mundo, y emite `entity_reparented` con `cause: "space_destroyed"`. Si no queda ningún ancestro (se destruye la raíz), los ocupantes se destruyen en cascada. Que después caigan es asunto de M5.

## Comandos y eventos

- `spawn` acepta `space` y valida `location.space`.
- `set_location` valida destino y ciclos.
- `reparent {entity, space}`.
- Evento `entity_reparented {entity, from, to, cause}`.

## Tolerancia de Reparent

Exacto con ángulos de 90°. Con ángulos arbitrarios, unos pocos µm (el test acota 3 µm), despreciable frente a un tiny de 8 mm.
