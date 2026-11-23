# Oficios de la colonia, expediciones y cobertura micro

Implementación de la [sección 18 de la propuesta](../propuesta.md#18-oficios-de-la-colonia-y-expediciones) (hitos M50 a M55).

## Arquitectura y separación de responsabilidades

Todo se activa por contenido: las colonias sin oficios (`member.role`), sin puestos de vigía (`lookout_post`) ni destinos (`expedition_target`) conservan su comportamiento previo.

| Capa / Paquete | Responsabilidad |
| --- | --- |
| `internal/sim/ai` | Hook `ai.SetJobScorer(w, fn)` (recurso del mundo, no serializado) para ajustar el puntaje de trabajos sin acoplar `ai` a `colony`. |
| `internal/sim/nav` | Filtro de escala `Portal.Scale` / `Mover.Scale` (`""` = cualquiera, `"tiny"` = pasajes estrechos), `Mover.WaitUntil` (pausa sin cancelar `Approach`) y hook `nav.SetPortalPenalty(w, fn)` para desvíos ante peligro o avisos. |
| `internal/sim/colony/roles.go` | Oficios preferidos (`gatherer`, `porter`, `builder`, `lookout`, o `""` generalista), prioridad de despensa crítica (`CriticalFood`), espera breve por especialista (`CoverDelay`) y suplencias. |
| `internal/sim/colony/lookout.go` | Componente `lookout_post`, acción `watch`, evento `lookout_warned`, sistema `colony.lookout` (vence avisos y detiene o replanea a miembros cercanos) y penalización de portales en zonas peligrosas o avisadas. |
| `internal/sim/colony/provisions.go` | Componente `provisions`, acción `eat_ration`, traslado atómico en el mismo tick (`equipRation` / `returnRations`) e invariante `colony.provisions`. |
| `internal/sim/colony/expedition.go` | Componentes `expedition_target`, `expedition` (máx. 1 activa por colonia, 1 a 4 miembros ordenados por ID) y `stranded`, acción `expedition`, sistema `colony.expeditions` e invariante `colony.expedition`. `colony.share` retiene los hallazgos de expedicionarios activos hasta que regresan al nido. |
| `internal/scenario/random.go` | `randomTrades` sortea una mezcla inicial de oficios (con generalistas) e inclinación por la aventura al final del sorteo, solo si el escenario incluye mapa de aventuras. |
| `tools/scenarios/oficios.py` | Agrega a `content/partidas/casa.json` el pasacables del zócalo (`cable_duct`), el refugio del umbral (`threshold_refuge`), el bolsillo del sofá (`sofa_pocket`), el puesto de vigía (`lookout_leg`) y los tres destinos (`backrest_lookout`, `duct_target`, `sofa_pocket_target`). |
| `internal/render/pixels.go` y `cmd/game/art.go` | Cobertura micro en todo `Space` con extensión no vacía (incluidas patas, respaldos, pasacables y bolsillo del sofá), con vista micro básica (`basicSurface`) y proyección vertical sin aplanar cuando el foco es una tira `climb`. |
| `internal/view`, `internal/ui`, `cmd/game/hud.go` | `Frame` expone `Role`, `Why`, `Rations`, `Stranded`, `Colony.Expedition` y `Colony.NoExpReason`; la crónica registra salidas, hallazgos, avisos y varados; la ficha y la pestaña Colonia muestran oficio, motivo, expedición y qué impide partir. |

## Invariantes

- `colony.provisions`: `Provisions.Rations >= 0`. Cada porción se descuenta al equipar y se devuelve al regresar vivo en el mismo tick.
- `colony.expedition`: como máximo 1 `Expedition` viva por colonia; cada tiny participa en como máximo 1 expedición activa; `len(Members) <= 4` ordenados por ID (D5).
- `nav.portal` / `nav.mover`: un portal con `Scale != ""` solo es planificable por un `Mover` con el mismo `Scale`.

## Tests y aceptación

- `internal/sim/colony/oficios_test.go`: oficios y suplencias (`TestRolesAndCoverage`, `TestGeneralistColonyFeedsItself`), filtro de escala (`TestNavPortalScaleFilter`), vigía (`TestLookoutWarnsAndReleasesPost`), expedición completa + conservación + snapshot en convocatoria (`TestExpeditionLifecycleAndConservation`), cierre de portal y regreso de varado (`TestExpeditionStrandedAndRecovery`), reemplazo de guía (`TestExpeditionGuideReplacement`) y cancelación por falta de provisiones (`TestExpeditionCanceledWhenLowProvisions`).
- `tests/scenarios/duct_detour_and_giantess_door.json`: ida por la puerta, vuelta por el pasacables ante peligro y giantess usando siempre la puerta.
- `cmd/game/zoom_test.go` (`TestFollowStaysInMicroOnClimb`) y `cmd/game/micro_coverage_test.go` (`TestMicroCoverageFullJourney`): seguimiento continuo en micro desde el escritorio, por la pata y el pasacables, hasta el bolsillo del sofá.
