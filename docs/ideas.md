# Ideas de contenido — Temática Giantess + Tinies (por impacto / costo)

Banco de ideas para profundizar la experiencia emergente **giantess + tinies** sobre la base actual del simulador ([propuesta](propuesta.md), [mapa corporal](specs/cuerpo.md), [prendas explorables](specs/prendas.md), [encuentros](specs/encuentros.md), [aplastamiento](specs/crush.md), [actividades](specs/actividades.md), [oficios y expediciones](specs/oficios.md) y [planta del departamento](../tools/scenarios/planta.py)).

Las ideas están ordenadas por **mayor impacto temático al menor costo técnico**, midiendo el costo según cuánto reutilizan los sistemas ya construidos (`sim/encounter`, `attach.Garment`, `BodyZone`/`Gripping`, `expire`, `colony`, `ai.Offer` y vistas macro/micro).

---

## Tabla resumen de priorización

| Nivel | Estado | # | Idea | Categoría | Impacto | Costo | Qué reutiliza del motor |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 · Quick win | `[x]` | 1 | Transporte deliberado («Llevar a un pasajero») | Encuentros | Muy alto | Bajo | `encounter/offer.go`, `doing.items`, `nav.Approach`, `ai.Step` |
| 1 · Quick win | `[x]` | 2 | Más prendas y accesorios (`medias`, `coletero`, `genkan`) | Cuerpo y ropa | Alto | Bajo | `attach.Garment`, `BodyZone`, `BodyPortal`, `search_spot` |
| 1 · Quick win | `[x]` | 3 | Sensibilidad táctil y reacciones («Cosquillas») | Cuerpo vivo | Muy alto | Bajo | `BodyZone`, `Gripping`, poses y eventos de `attach` |
| 1 · Quick win | `[x]` | 4 | Vínculo dinámico (confianza ↔ miedo) y reacción a `observe` | Relación | Alto | Bajo | `encounter.look`, `agent.care`, `near_miss`, `gift_spot` |
| 1 · Quick win | `[x]` | 5 | Micro-eventos con `expire` y `spawn` (ducha, arrocera, planta) | Rutinas / biomas | Alto | Bajo | `expire.Expires`, `water_drop`, `puddle`, `ai.Offer` |
| 2 · Gran salto | `[ ]` | 6 | Recoger e inspeccionar al rostro + viñeta UI (*SAEKO*) | Encuentros / UI | Muy alto | Medio | `encounter`, poses, `view.Frame`, retrato de UI |
| 2 · Gran salto | `[ ]` | 7 | Fuerzas no letales (viento, secador, celular vibrando, barrido) | Física cotidiana | Muy alto | Medio | `Gripping`, `crush.Threats`, `ai.Offer`, `space.falls` |
| 2 · Gran salto | `[ ]` | 8 | Ingeniería tiny con tesoros («préstamos» funcionales) | Colonia | Alto | Medio | `colony.Trinket`, `build_site`, `nav.Portal`, `crush.Shelter` |
| 2 · Gran salto | `[ ]` | 9 | Cadena de alerta, rescate de aturdidos y memoria de rutinas | Colonia | Alto | Medio | `lookout_post`, `crush.Threats`, `colony.KnownSupports` |
| 2 · Gran salto | `[ ]` | 10 | Correr el punto de apoyo / hacer de puente | Encuentros / cuidado | Alto | Medio | `agent.care`, `crush.Threats`, `attach.Surface` |
| 2 · Gran salto | `[ ]` | 11 | Respiración al dormir y calor corporal | Cuerpo vivo | Medio-alto | Medio | `BodyState` (`lie_down`), `Garment`, `needs` |
| 3 · Expansión | `[ ]` | 12 | Ciclo día / tarde / noche | Ecosistema | Muy alto | Medio-alto | `clock.Tick`, iluminación/render, utilidad de `ai` |
| 3 · Expansión | `[ ]` | 13 | El *kotatsu* y muebles transformables (libros, futón) | Biomas / casa | Alto | Medio-alto | `space.Projection`, `BodyZone` (`occluded`), `planta.py` |
| 3 · Expansión | `[ ]` | 14 | Sub-nidos y puestos avanzados por habitación | Colonia | Alto | Medio-alto | `colony.Colony`, `nest`, `haul`, `expedition` |
| 3 · Expansión | `[ ]` | 15 | Una visita en el departamento (segunda giantess) | Variedad | Alto | Medio-alto | Multi-agente nativo, `encounter`, `planta.py` |
| 4 · Largo plazo | `[ ]` | 16 | Intervención suave de sandbox para el jugador | Control / UI | Medio-alto | Alto | `cmds`, `ui.HUD`, `ai.danger` / `landmark` |
| 4 · Largo plazo | `[ ]` | 17 | Micro-fauna a escala tiny (hormigas / escarabajos) | Ecosistema | Medio-alto | Alto | Nuevo tipo de agente, sprites, `ai`, `crush` |

---

## Nivel 1 — Quick wins (Alto impacto · Costo bajo)

Cambios que casi no requieren infraestructura nueva porque se apoyan directo en los sistemas ya cerrados.

### 1. [x] Transporte deliberado («Llevar a un pasajero»)
- **Qué es**: hoy, si un tiny sube a la palma durante `offer_palm` (`internal/sim/encounter/offer.go`), queda en `doing.items` y viaja de polizón hasta la siguiente rutina. Con una extensión directa de `encounter`, si la giantess nota pasajeros sobre su palma (o un tiny `stranded` en el piso/baño), los traslada intencionalmente al escritorio o a la mesa donde acaba de merendar y espera a que bajen antes de retirar la mano.
- **Por qué rinde tanto**: convierte la interacción más icónica del género en un gesto consciente y legible en la crónica, reutilizando por completo `boarded`, `nav.Approach` y las poses de apoyo.
- **Implementado**: `CarryPassenger` en `internal/sim/encounter/carry.go`, narrativa/HUD en `internal/view/{cues,story,chronicle}.go` y `cmd/game/hud.go`, sprite `ui/emote_carry.png` y escenarios `giantess_carries_passenger_to_desk.json` y `giantess_cancels_carry_when_occupied.json`.

### 2. [x] Más prendas y accesorios (`medias`, `coletero`, `genkan`)
- **Qué es**: sumar nuevas entidades de contenido sobre `attach.Garment` y `BodyZone` sin tocar el motor:
  - **Medias / calcetines**: interior muy oscuro (`occluded`), trampa clásica si se los calza con alguien adentro (`InnerPressure`), y fuente rica de fibras/pelusa si quedan tirados junto al futón.
  - **Coletero / hebilla en el cabello**: un refugio fijo y seguro dentro de `body_head` que los scouts pueden usar como «campamento móvil» mientras ella recorre la casa.
  - **Zapatos de calle y bolso/mochila en el `genkan`**: activa la entrada (`genkan`, ya definida en [planta.py](../tools/scenarios/planta.py)) como zona de expedición de alto riesgo y botín único (monedas, envoltorios, llaves, migas dulces).
- **Por qué rinde tanto**: casi todo es JSON, script de escenario y sprites, aprovechando al 100 % el trabajo reciente de [prendas.md](specs/prendas.md).
- **Implementado**: `sock_left`/`sock_right`, `hair_tie` con refugio en `internal/sim/crush/resolve.go`, `street_shoe_l`/`r`, `genkan_bag` y botín `trinket_coin` (`internal/sim/colony/trinket.go`, `content/objects/trinket_coin.json`), atenuación de exploración en `colony/{search,scout}.go`, sprites `sock`, `hair_tie`, `street_shoe`, `genkan_bag` y `trinket_coin`, audio `SockRustle`, y escenarios `genkan_expedition_and_loot.json` y `sock_donning_and_hair_tie_shelter.json`.

### 3. [x] Sensibilidad táctil y reacciones involuntarias («Cosquillas»)
- **Qué es**: añadir un campo de sensibilidad a `BodyZone` (alta en cuello, rostro, palma y planta del pie descalzo; media en brazos; nula en ropa y cabello). Al caminar o correr sobre piel sensible se acumula estímulo hasta disparar una reacción breve de la giantess: **detenerse a mirar, rascarse, sacudir la extremidad o acomodarse la ropa**, obligando a los tinies a sujetarse a un `GripPoint` o caer.
- **Por qué rinde tanto**: con un contador simple sobre eventos de movimiento en `attach`, el cuerpo deja de ser un escenario pasivo y vuelve táctico elegir ropa/cabello frente a piel desnuda.
- **Implementado**: `BodyZone.Sensitivity` y sistema de cosquillas en `internal/sim/attach/{body,pose,tickle}.go` (`BodyTickled`), reacciones visuales (`EmoteTickle`, `FaceSurprised`) en `internal/view`, audio `BodyTickle` en `internal/sound`, sprite `ui/emote_tickle.png` y escenario `tiny_tickles_giantess_neck.json`.

### 4. [x] Vínculo dinámico (confianza ↔ miedo) y reacción tiny a `observe` / `offer_palm`
- **Qué es**:
  - **Del lado tiny**: registrar afinidad/confianza simple hacia la giantess. Los tinies confiados se acercan a la palma ofrecida o saludan cuando ella los mira (`observe`); los que sufrieron un `near_miss` o perdieron compañeros se esconden o retroceden apenas notan que los observa.
  - **Del lado giantess**: perfil de personalidad sobre `agent.care` (*cuidadosa, curiosa, distraída, ordenada, juguetona*) y afinidad evolutiva con la colonia (los `gift_spot` y encuentros pacíficos suben paciencia/atención; presenciar `remains` en alguien cuidadosa dispara cautela temporal).
- **Por qué rinde tanto**: transforma el sentido de `observe` y `offer_palm` (hoy los tinies ignoran que los están mirando) usando contadores mínimos y señales ya existentes.
- **Implementado**: componente `Bond`, `TinyReacted` y evolución de vínculo en `internal/sim/encounter/bond.go`, integración en `internal/scenario/random.go`, HUD/crónica en `internal/view` y `cmd/game/hud.go`, sprites `ui/emote_wave.png` y `ui/emote_shy.png`, y escenarios `tinies_react_to_observe_by_trust.json` y `bond_evolves_from_share_and_near_miss.json`.

### 5. [x] Micro-eventos con `expire` y `spawn` (post-ducha, arrocera, regar la planta)
- **Qué es**: nuevas ofertas cotidianas en objetos ya presentes o pequeños de cocina/living que reutilizan `expire.Expires`, `water_drop`, `puddle` y `spawn`:
  - **Humedad post-ducha**: al salir de `shower_tray`, piel y piso del baño quedan con gotas temporales cosechables para tinies valientes.
  - **Arrocera / preparar té**: vapor cálido y **granos de arroz** (una unidad equivale a muchas porciones de comida para el nido).
  - **Regar la planta del living**: para la giantess dura 3 segundos; para los tinies en la maceta es una lluvia torrencial que deja gotas y charcos.
- **Por qué rinde tanto**: agrega variedad ambiental y recursos temáticos en macro/micro usando exactamente el patrón de [actividades.md](specs/actividades.md).
- **Implementado**: `content/objects/{rice_grain,rice_cooker,potted_plant}.json`, `bath_mat` y `spot_living_plant` en `tools/scenarios/quickwins.py` y `content/partidas/casa.json`, sprites `bath_mat`, `rice_cooker`, `rice_grain` y `potted_plant`, audio `RiceSteam` y `PlantRain`, escenarios `giantess_cooks_rice_and_colony_hauls.json` y `giantess_waters_plant_and_tinies_drink.json`, y balance multi-semilla verificado en `internal/scenario/quickwins_balance_test.go`.

---

## Nivel 2 — Gran salto temático (Impacto muy alto · Costo medio)

Requieren lógica nueva acotada, pero elevan muchísimo la fantasía giantess + tinies.

### 6. Recoger e inspeccionar al rostro (`inspect_tiny`) + viñeta de perspectiva (*SAEKO*)
- **Qué es**:
  - **Simulación**: acción de encuentro para una giantess curiosa/afectuosa que recoge con cuidado a un tiny en la palma o entre los dedos y lo eleva a la altura de su rostro para observarlo de cerca, reaccionando según su personalidad antes de volver a apoyarlo en una superficie segura.
  - **UI (guiño a *SAEKO*)**: sin cambiar la simulación, cuando se sigue en micro a un tiny bajo la sombra de la mano, siendo observado (`observe`/`inspect_tiny`) o sobre su palma/hombro, mostrar una viñeta o retrato contextual que refuerce la sensación de inmensidad desde abajo.
- **Por qué compensa el costo**: combina la escena más icónica del género con una lectura visual inmediata en micro.

### 7. Fuerzas no letales: ráfagas de aire, secador de pelo, celular vibrando y barridos
- **Qué es**: introducir impulsos/desplazamientos horizontales sobre una superficie sin matar al instante:
  - **Viento y corrientes**: abrir un `fusuma`, sacudir el edredón al levantarse, ventilador en el living, estornudo o el **secador de pelo en el baño**. Arrastran tinies sin `Gripping` y desparraman migajas/fibras, obligando a parapetarse detrás de objetos pesados.
  - **Limpieza de superficies**: pasar un paño por el escritorio o barrer el piso empuja todo lo suelto hacia el borde salvo lo que esté en refugios.
  - **Celular sobre el escritorio**: al sonar/vibrar genera una onda sísmica local sobre la superficie que alerta a los tinies e ilumina en la oscuridad.
- **Por qué compensa el costo**: aporta amenazas físicas espectaculares y dinámicas cotidianas que no dependen de matar tinies por aplastamiento.

### 8. Ingeniería tiny con tesoros («préstamos» funcionales)
- **Qué es**: dar utilidad física real a los `trinkets` de [actividades.md](specs/actividades.md) además de `fun` y regalos:
  - **Hilo (`thread`)**: tender **cuerdas/tirolesas** para bajar rápido de muebles o armar un **arnés** para scouts que mejora su agarre sobre el cuerpo.
  - **Clip (`clip`)**: construir una **escalera o puente portátil** (subir el escalón del `genkan`, cruzar entre silla y escritorio).
  - **Botón (`button`)**: reforzar el techo de un refugio para resistir mayor fuerza de aplastamiento (taza o mano).
- **Por qué compensa el costo**: conecta la recolección de tesoros (*Arrietty* / *The Borrowers*) directamente con la supervivencia y la navegación en la casa.

### 9. Cadena de alerta, pánico colectivo y rescate de aturdidos
- **Qué es**:
  - **Cadena de alerta**: el vigía (`lookout_post`) y tinies alarmados propagan aviso a compañeros cercanos para evacuar en grupo antes de una pisada o mano.
  - **Estado aturdido / atrapado (no letal)**: por caída alta, ráfaga o roce de tela; un tiny rescatista puede ayudarlo a volver al nido.
  - **Memoria de rutinas**: extender `KnownSupports` para recordar qué superficies son peligrosas o ricas según los hábitos de la giantess.
- **Por qué compensa el costo**: cierra pendientes de [crush.md](specs/crush.md) y hace que la colonia actúe de verdad como grupo coordinado frente a la escala gigante.

### 10. Correr el punto de apoyo y hacer de puente
- **Qué es**:
  - **Correr el apoyo**: una giantess con `care` alto que detecta tinies o `remains` bajo la huella no solo titubea: desplaza el apoyo de la mano hacia una zona libre (cerrando el pendiente explícito de [crush.md](specs/crush.md)).
  - **Hacer de puente**: si nota tinies mirando el borde de un mueble o varados en un rincón, apoya la mano o un objeto (cuaderno/pantufla) como puente temporal.
- **Por qué compensa el costo**: hace creíble e inteligente el comportamiento de una giantess cuidadosa.

### 11. Respiración al dormir y calor corporal (microclimas simples)
- **Qué es**:
  - **Respiración**: en `lie_down` cerca de `body_head` y la almohada, ciclos periódicos de brisa suave que mecen o empujan a tinies sin agarre.
  - **Calor corporal**: acurrucarse en el bolsillo del cárdigan o sobre el hombro recupera energía más rápido que un refugio común.
- **Por qué compensa el costo**: suma inmersión sensorial al mapa corporal con reglas periódicas simples.

---

## Nivel 3 — Expansiones grandes (Alto impacto · Costo medio-alto)

Ideas estructurales que abren capas nuevas de juego una vez consolidado lo anterior.

### 12. Ciclo día / tarde / noche
- **Qué es**: dividir la partida en fases horarias con iluminación, ambiente sonoro y pesos de rutina:
  - **De día**: manchas de sol sobre el tatami, giantess activa (desayuno, estudio, TV, ducha), rutas peligrosas pero abundancia de migajas frescas y viajes.
  - **De noche**: al dormir en el futón, la luz se apaga y la casa queda silenciosa — ventana ideal para **expediciones largas por el piso**, incursiones al baño/cocina o trepar hasta la almohada sin riesgo de pisadas.

### 13. El *kotatsu* y muebles transformables (libros abiertos, futón)
- **Qué es**:
  - **El *kotatsu*** (ya insinuado en [propuesta.md](propuesta.md)): debajo de la manta se forma un bioma oscuro (`occluded`), cálido y lleno de migajas, compartido con las piernas y pies de la giantess mientras mira TV.
  - **Estudiar / leer**: apoyar libros abiertos crea temporalmente muros, rampas y puentes de papel sobre el escritorio.

### 14. Sub-nidos y puestos avanzados por habitación
- **Qué es**: permitir que la colonia establezca escondites secundarios (bajo la mesada de la cocina, en el bolsillo del sofá o en el `genkan`) con despensa local, reduciendo viajes mortales por pisos abiertos y habilitando **caravanas** aprovechando las pantuflas o el cárdigan.

### 15. Una visita en el departamento (segunda giantess)
- **Qué es**: aprovechar que el motor no asume una única giantess para introducir un escenario o evento con una **amiga de visita** de personalidad opuesta (por ejemplo, dueña cuidadosa + visita distraída que apoya su bolso en el escritorio y se sienta a tomar té), multiplicando los contrastes y el drama de escala.

---

## Nivel 4 — Apuestas de largo plazo (Costo alto / fase posterior)

Conviene dejarlas para después de los niveles anteriores por su costo en arte, IA o UX.

### 16. Intervención suave de sandbox para el jugador
- **Qué es**: órdenes indirectas que respeten la autonomía de la simulación: marcar zonas de «evitar / explorar» para la colonia, activar alerta de refugio, o darle un «empujoncito» de curiosidad a la giantess hacia un cuarto o superficie para observar qué emerge.

### 17. Micro-fauna a escala tiny
- **Qué es**: un insecto doméstico ocasional (hormiga exploradora o escarabajo chico) que compite por migajas en la cocina o el `genkan`: para los tinies es una bestia enorme que deben evitar o distraer, mientras la giantess puede espantarlo sin saber que salvó a la colonia.
