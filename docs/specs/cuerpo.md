# Spec: La giantess como mapa explorable (M56–M59, propuesta 19)

Paquetes: `sim/space`, `sim/nav`, `sim/attach`, `sim/crush`, `sim/colony`, `view`, `render`, `cmd/game`. Contenido: `tools/scenarios/cuerpo.py`.

## Proyección 3D de superficies (`space.Projection`, M56)

Cada `space.Space` admite un campo opcional `projection` (`nil` = plano XY clásico, sin impacto en superficies planas de la casa):

- **Tramo angosto** (`anchor_a`, `anchor_b`, `width`, `climb`): la coordenada `Y` local en `[Extent.Min.Y, Extent.Max.Y]` interpola linealmente de `AnchorA` a `AnchorB`; `X` local se proyecta perpendicularmente en el plano XY.
- **Superficie inclinada** (`origin`, `axis_u`, `axis_v`): marco afín en µm con escala `space.BasisScale` (1 m) y aritmética entera de 128 bits (`mulDiv64`, sin floats, determinista D1/D5).
- `Graph.ToWorld`, `Graph.FromWorld`, `Graph.WorldPose` y `Graph.Reparent` componen la proyección con la jerarquía `comp.Location`.
- `Space.SegmentLen3D` mide la longitud real en 3D: `nav.Router` y `nav.mover` la usan para costos de ruta y tiempo de trepada.
- `Graph.Below` / `BelowExcluding` excluyen los tramos verticales (`IsClimb`) como posibles aterrizajes e intersectan la vertical sobre superficies inclinadas para calcular la altura exacta de caída.

## Poses completas, articulaciones y agarre (`sim/attach`, M56)

- `PartPose.Projection` permite actualizar la proyección 3D de una parte al aplicar `set_pose` (validada atómicamente).
- `BodyState` (store `"body_state"`) registra la pose activa y el tick del último cambio.
- `BodyJoints` (store `"body_joints"`) declara portales internos con `closed_poses` / `open_poses`: `set_pose` alterna `nav.Portal.Disabled` y sube `Version` sin recrear la entidad del portal.
- `BodyZone` (store `"body_zone"`) declara `body`, `region`, `require_grip`, puntos de agarre `grips` (`pos`, `radius`, `label`) y `occluded` (ropa o pelo que tapan al tiny).
- `Gripping` (store `"gripping"`) guarda el estado de sujeción de un tiny tras un cambio de pose; eventos `entity_gripped` y `entity_lost_grip`. Si una zona con `require_grip` (no `climb`) cambia de pose sin agarre al alcance, el tiny pierde apoyo y cae sobre la superficie debajo (`space.LandExcluding`).

## Apoyos con bordes de acceso y validación al cruzar (M57)

`attach.Surface` (`contact_surface`) suma campos opcionales compatibles hacia atrás:

- `scale`: fija `nav.Portal.Scale` (p. ej. `"tiny"`).
- `stable_ticks`: ticks mínimos de contacto continuo antes de abrir portal (evita aperturas por un roce breve).
- `require_still`: exige que el cuerpo dueño no esté caminando (p. ej. pie plantado en el piso).
- `support` / `support_offset`: cara de apoyo que toca muebles y define la huella de presión en `crush.resolve`.
- `access` / `access_reach`: puntos locales de borde de acceso por donde entra o sale un tiny (pueden quedar al costado de la huella de presión).
- `body` / `self_contact`: evita autocontactos accidentales entre partes del mismo cuerpo salvo cuando se declaran (p. ej. mano sobre rodilla).

**Validación exacta en el tick de cruce:** en `nav.move`, los soportes móviles (Spaces con `nav.Mover`) avanzan antes que los ocupantes; justo antes de cruzar un portal se consulta `nav.SetPortalValidator` (`attach.PortalCrossingValid`), que revalida la pose, quietud, transitabilidad y alcance del tick actual. Si el apoyo se retiró en ese tick, quien aún no cruzó queda de su lado y replanea; quien ya cruzó viaja con el soporte.

## Mapa corporal completo y escenarios (M58)

`python tools/scenarios/cuerpo.py` integra idempotentemente en `content/partidas/casa.json` (después de `oficios.py`) y genera los escenarios de integración:

- **Regiones:** cabello y rostro (`body_head`), cuello y nuca (`body_neck`), hombros, pecho, espalda, costados, cintura y caderas, brazos/antebrazos/manos (ambos lados, aprovechando `$palm` en el lado derecho) y piernas/pantorrillas/pies (ambos lados), más las superficies navegables `futon` y `pillow` en el dormitorio y la pose `lie_down` en `bed`.
- **Escenarios:** `tests/scenarios/body_desk_to_pillow.json` (escritorio → palma → brazo → hombro → cuello → cabello → almohada → futón) y `tests/scenarios/body_floor_to_sofa.json` (piso → pie izquierdo → pantorrilla → muslo → sofá + caída al sofá).
- **Validación de contenido:** `TestBodyMapCoverageAndReachableExits` verifica que toda zona corporal tenga extensión no vacía (vista micro disponible) y una salida alcanzable a alguna superficie de contacto.

## Aventuras en el cuerpo, vista micro y continuidad (M59)

- **Memoria de apoyos y expediciones (`sim/colony`):** `Colony.KnownSupports` guarda `SupportNote {body_space, house_space, pose, tick}` reportados al volver al nido (`support_discovered`). Durante una expedición sobre zonas corporales, si una articulación se cierra temporalmente el integrante aguarda en el punto de agarre más cercano antes de retirarse o quedar `Stranded` esperando un apoyo abierto. Los exploradores (`RoleScout`) y suplentes aventureros (`Adventure >= 0.5`) pueden cruzar portales de contacto hacia zonas corporales articuladas; la curiosidad casual (`ai.explore`) sube a la palma apoyada pero no encadena recorridos internos por el cuerpo.
- **Vista micro, rótulos y crónica (`view`, `render`, `cmd/game`):** `view.Space` expone `Climb` y `Occluded`; `view.Thing` expone `GripLabel`. En `render/pixels.go`, las zonas con `Climb` usan vista básica de trepada de costado y las zonas con `Occluded` recortan la capa frontal alrededor del tiny seguido. `cmd/game/hud.go` muestra el rótulo jerárquico desde el cuerpo (`«Giantess › Antebrazo izquierdo › Puño de la manga izquierda»`). La crónica registra los apoyos descubiertos (`«Desde cabello y rostro se puede bajar a almohada cuando se recuesta.»`).
- **Continuidad y pruebas:** `TestHandToShoulderClimbAndElbowJoint`, `TestSupportAccessRules`, `TestSupportWithdrawnAtCrossingTickAndCrushProtection`, `TestBodyMapCoverageAndReachableExits`, `TestBodyExpeditionSupportMemoryAndContinuity`, `TestBodyFrameFlagsAndSupportChronicle` y `TestMicroBodyClimbAndOccludedCutaway`.
