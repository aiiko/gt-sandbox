# Spec: Asientos en la casa (M34–M38)

Paquetes: `content/objects` (`chair`, `sofa`), `sim/crush` (`SourceSeat`), `sim/ai` (`SeatCareScale`, `Hesitated.Source`), `internal/view`, `internal/render`, `internal/sound`, `cmd/game`.

Diseño y alcance en el [plan](../plan-motor.md#hitos-de-asientos-en-la-casa-planificados).

## Superficies y accesos (M34)

- **Silla común (`content/objects/chair.json`):** `Space` elevado de `40 × 40 cm` (`nav_grid` de 2 cm, `contact_surface`), fija en el dormitorio frente a `desk1` a 45 cm de altura (`(1,60 m, 0,55 m, 0,45 m)`).
- **Sofá (`content/objects/sofa.json`):** `Space` elevado de `120 × 50 cm` (`nav_grid` de 2 cm, `contact_surface`), fijo en el living frente a `desk2` y la `tv` a 45 cm de altura (`(1,20 m, 0,45 m, 0,45 m)`).
- **Accesos para los tinies:**
  1. **Pata desde el piso:** `chair_leg` (`bedroom ↔ chair_leg ↔ chair`) y `sofa_leg` (`living ↔ sofa_leg ↔ sofa`), cada una como `Space` elevado de `2 × 45 cm` con dos portales estáticos.
  2. **Puente desde la mesa/escritorio:** `chair_bridge` (`desk1 ↔ chair_bridge ↔ chair`) y `sofa_bridge` (`desk2 ↔ sofa_bridge ↔ sofa`), cada uno de `2 × 30 cm`, para llegar sin bajar al piso.
  3. **Contacto dinámico (`contact_surface`):** si la giantess apoya la palma sobre el asiento (p. ej. `sit_hand_sofa`), `attach.contacts` abre un portal dinámico `palma ↔ asiento`.
  4. **Caída desde arriba (`space.falls`):** las migajas del desayuno (`breakfast`) y de la merienda (`snack_living`) pueden caer al asiento de la silla o del sofá, generando recolección (`haul` / `forage`) además de exploración (`explore`).

## Sentarse y rutina (M35)

- **Parte articulada `$seat`:** `Space` `articulated` de `36 × 36 cm` montado sobre `$giantess`, con `crusher {source: "seat", force: 450, thickness: 20000}`.
- **Poses combinadas simples:** el componente `poses` mueve `$palm` y `$seat` en cada pose:
  - `idle`: de pie (`$seat` a 95 cm, fuera de `ThreatHeight`).
  - `sit_hover`: bajando a sentarse (`$seat` a 60 cm, cara inferior a 13 cm del asiento → amenaza de sombra).
  - `sit`: sentada (`$seat` a 47 cm, cara inferior a ras de los 45 cm del asiento).
  - `sit_set_table`, `sit_hand_hover`, `sit_hand_rest`, `sit_hand_sofa`: sentada combinada con la posición de la mano.
- **Oferta propia del sofá (`relax`):** `energy +250`, `fun +150`, con pasos `goto → sit_hover → wait 16 → sit → use 250`. Al terminar o interrumpirse vuelve a `idle`.
- **Rutina en `content/partidas/casa.json`:** `breakfast` y `write` se sientan en la silla; `watch`, `snack_living` y `relax` se sientan en el sofá. Los escenarios de referencia (`una_manana_en_la_casa`, `un_descuido_en_el_escritorio`) no cambian.

## Aplastamiento al sentarse y cuidado (M36)

- `crush.resolve` y `crush.Threats` tratan `$seat` como cualquier `Space` aplastador (`crush.SourceSeat = "seat"`):
  - en `sit_hover`, los tinies con `reflex` reciben `alarmed` (`channel: "shadow"`, `source: "seat"`) y huyen (`flee`);
  - en `sit`, `crush.resolve` emite `pressed`, mata a los frágiles bajo la huella (`crushed` con `source: "seat"`), da `near_miss` a los cercanos y `colony.mourn` marca zona `danger`.
- **Contraste silla vs. sofá:** en la silla (`40 × 40 cm`) la huella de `36 × 36 cm` cubre casi todo el asiento y solo deja margen estrecho hacia el frente; en el sofá (`120 × 50 cm`) cubre solo el lugar central, dejando a salvo los extremos.
- **Cuidado fuera de la línea de vista (`ai.SeatCareScale` = 0,45):** sentarse ocurre de espaldas, fuera del campo visual directo. `careful` aplica `care_eff = care² · SeatCareScale` para la tirada de atención y el tiempo extra de espera (`HoverExtra`), de modo que un `care` moderado suele fallar y hace falta un `care` alto para notar tinies y emitir `hesitated {source: "seat"}`.

## Presentación y sonido (M37)

- **Arte (`tools/art/seats.py`, `content/art/src/casa/atlas.json`):** sprites macro y baldosas micro (`tile: true`) para `chair` y `sofa`, y sombra `seat` (`shadow_<dir>`) en vista micro; `poseClips` usa `sit` para `relax`.
- **HUD e historia (`cmd/game/hud.go`, `internal/view/story.go`, `internal/view/cues.go`):** actividad «Descansa en el sofá», `FaceHappy` durante `relax`, causa de restos «Aplastado al sentarse», y textos de historia para `relax`, `sit`, `crushed` (`source: "seat"`) y `hesitated` (`source: "seat"`).
- **Sonido (`tools/audio/sounds.json`, `internal/sound/scene.go`):** `chair_creak` y `sofa_sit` al sentarse (`pose: "sit"`), y `crush_tum` / `crush_far` para `source: "seat"`.

## Pruebas y aceptación (M34–M38)

Generador de escenarios: `python tools/scenarios/asientos.py` y `python tools/scenarios/partida.py`.

- `tiny_reaches_chair_and_sofa`: acceso por pata y puente en silla y sofá, y caída de migaja al asiento.
- `giantess_sits_and_relaxes`: uso de `relax` en el sofá (`sit_hover → sit → idle`).
- `chair_sit_crushes_tiny`: centro aplastado (`source: "seat"`), borde huye y se salva (`near_miss`).
- `sofa_sit_spares_far_side`: centro aplastado, extremo lejano intacto.
- `careful_sit_needs_high_care` y `TestSitRequiresMoreCareThanHand`: el cuidado alto titubea (`hesitated`, `source: "seat"`) y reduce muertes frente al cuidado moderado.
- `TestPalmContactWithSofa`: `sit_hand_sofa` abre contacto dinámico `palma ↔ sofá`.
- `TestSeatsScenariosSnapshots`: continuación desde snapshots cada 37 ticks con mismo hash e invariantes.
- `TestSeatsStoryAndExpression`: textos de historia y `FaceHappy` en `internal/view`.
- `TestSeatsAcceptance` (`internal/scenario/m38_test.go`): aceptación en `content/partidas/casa.json` sobre las semillas de referencia (1, 7, 21, 42, 99).
