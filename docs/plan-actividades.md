# Plan de desarrollo: actividades de los tinies en la casa

La reescritura de la casa (M60) sumó cocina, baño, genkan, futón, alfombra y maceta. Los tinies casi no los usan: todo su trabajo pasa entre el escritorio, la silla y el sofá. Este slice (**M61 a M71**) les da **diez actividades nuevas** repartidas por toda la planta. Cada hallazgo en una habitación sirve para algo en otra, así que hay motivos para ir lejos y volver.

Cada hito trae **simulación, contenido, sprites, textos y sonido**: al cerrarlo, la actividad se ve y se entiende en la partida, sin esperar a un hito final de arte. La spec del slice es `docs/specs/actividades.md` y cada hito le suma una sección. El diseño de juego va en una sección nueva de la propuesta (§21), que se escribe antes de M61.

## Hitos

| Hito | Actividad | Qué agrega al motor | Sprites | Cierre |
| --- | --- | --- | --- | --- |
| M61 | Agua del baño y la cocina | `expires`, necesidad `thirst`, `water_drop`, depósito por tipo en el nido, rutinas `wash_hands`/`wash_dishes` | Gota, charco, tapita del nido, emote de sed, la giantess lavándose | `tiny_drinks_drop`, `drop_evaporates`, `water_hauled_to_nest` |
| M62 | Granos de la arrocera | Acarreo cooperativo (`heavy`, trabajo con varios portadores), rutina `cook_rice` | Grano suelto y cargado entre dos, arrocera con vapor, la giantess cocinando, emote de espera | `two_tinies_carry_grain`, `grain_partner_leaves`, `grain_split_portions` |
| M63 | Tesoros perdidos y regalos | `trinket`, `hoard`, botín en `search_spot`, `gift_spot` y reacción en `encounter` | Botón, clip, hilo, mina de lápiz, pila del nido en etapas, emote de la giantess al encontrar un regalo | `tiny_finds_trinket`, `hoard_raises_fun`, `giantess_notices_gift` |
| M64 | Tender una cuerda | `anchor_site`: una obra que termina como portal (`set_portal`) | Nudo, cuerda colgando en macro y en vista de costado, obra en etapas | `colony_ties_rope`, `rope_route_used`, `rope_cut_by_palm` |
| M65 | Marcar el camino | `waymark`, `graphite_source`, acción `mark`, percepción de marcas, oficio de cartógrafo | Marcas de paso, peligro y refugio, lápiz en el escritorio, desgaste | `mark_teaches_portal`, `danger_mark_avoided`, `mark_worn_by_step` |
| M66 | Depósito avanzado en la cocina | `depot`, meta lejana, relevos depósito → nido | Depósito en obra, vacío, a medias y lleno; marcador macro | `colony_founds_depot`, `depot_relay`, `depot_starves_when_far` |
| M67 | Lo que traen los zapatos | Rutinas `go_out`/`come_home`, la giantess fuera de la casa (`away`), `shoe` como Space y refugio, `debris` sorteado | Par de zapatos en macro, interior en micro, semilla, arenilla, pétalo, la giantess saliendo y entrando, puerta exterior | `giantess_goes_out`, `debris_after_return`, `tiny_hides_in_shoe` |
| M68 | Hora del día y calor | Reloj del día (`daytime`), necesidad `warmth`, `sunpatch` móvil, `warm_spot` en el futón y la giantess que se da vuelta dormida | Manchas de sol, tinte de luz por hora, la giantess dándose vuelta, emotes de frío y de calor | `sunpatch_moves`, `tiny_warms_in_sun`, `rollover_threat_flee` |
| M69 | Huerta en la maceta | Tierra de la maceta como Space, `plot` con etapas y humedad, `seed`, `leaf` como material alternativo | Tierra en micro, parcela en cinco etapas, semilla, hoja, refugio de hojas | `colony_plants_seed`, `plot_dries_without_water`, `leaf_refuge_stronger` |
| M70 | Ver la tele y contar historias | `viewing_spot` con plazas ligado a una fuente activa, `tale` y acción `listen` | Reflejo de la pantalla en micro, tinies sentados mirando, burbuja de relato, grupo en la alfombra en macro | `tinies_watch_tv`, `tv_off_ends_viewing`, `expedition_tale_heard` |
| M71 | Partida e integración | Generador `actividades.py`, sorteo, ajuste sobre semillas, aceptación del slice | Íconos de la ficha de colonia, marcas en el miniplano, nota del menú | Aceptación de §21 y continuidad desde snapshot |

### Por qué este orden

- **M61 primero**: trae tres piezas que usa todo el resto: los recursos con vencimiento (`expires`), las rutinas de la giantess que dejan cosas con `spawn` y el depósito del nido por tipo de recurso. El agua además es insumo de la huerta (M69).
- **M62 antes que M66**: una carga pesada vuelve caros los viajes largos. Ese es el problema que resuelve el depósito avanzado, y la cocina queda calibrada como destino antes de fundar algo allí.
- **M63 antes que M64 y M65**: el hilo y la mina de lápiz son tesoros. La cuerda y las marcas consumen lo que trae la búsqueda.
- **M64 y M65 juntos**: son las dos formas en que los tinies dejan conocimiento en el mapa: una ruta nueva y un aviso. Las marcas pueden señalar la cuerda.
- **M67 antes que M69**: el genkan es la fuente de semillas. Además, la salida de la giantess abre una ventana sin pisadas que conviene probar antes de que haya proyectos largos expuestos a su rutina.
- **M68 antes que M69**: la parcela crece más rápido al sol y la maceta está junto a la ventana. Con el reloj del día, el crecimiento y las manchas de sol comparten la misma hora.
- **M70 al final**: es independiente, pero los relatos necesitan expediciones con hallazgos de este slice (zapatos, huerta) para tener algo que contar.

## Punto de partida: qué hay y qué falta

- **Las rutinas ya pueden dejar cosas.** Los pasos de oferta tienen `spawn` (el desayuno deja migajas con `{"do": "spawn", "item": "crumb"}`). Lavarse, cocinar y volver de la calle se escriben como ofertas con pasos en `content/objects/*.json`, sin código nuevo en `encounter`.
- **No hay vencimiento.** Nada desaparece solo. Las gotas, las manchas de sol, las marcas y los restos necesitan un sistema común de vencimiento con un evento de despawn.
- **El depósito del nido es de un solo tipo.** `portions` cuenta comida. El agua, los tesoros y luego las semillas necesitan cuentas separadas con su propia meta (`colony.stock` por recurso).
- **El acarreo es de un portador.** `carried` y `jobs.Reserve` asumen un solo dueño. Ya existe un patrón de dos plazas en `gathering` (M42) que se puede reutilizar para dos portadores.
- **Los portales nuevos solo los crea el escenario o `attach`.** `set_portal` existe como comando, pero ninguna obra de la colonia lo emite. M64 lo hace desde un sistema (D7).
- **No hay hora del día.** `clock` mide ticks, fases y escalonado. Las manchas de sol, el crecimiento y una salida de la giantess «a la mañana» necesitan una hora derivada del tick y de un inicio guardado en el mundo.
- **La giantess nunca sale.** Siempre ocupa la casa. Estar fuera (`away`) significa quitarle cuerpo, pisadas y percepción durante un plazo sin hacer despawn, porque el ID debe seguir vivo (D3).
- **Las superficies «No» del §20.2** (maceta, zapatero, lavamanos) no tienen Space. La maceta y el interior de los zapatos pasan a ser superficies tiny. El lavamanos sigue sin serlo: las gotas caen al piso del baño y al borde del plato de ducha.
- **Arte.** Las piezas micro de la colonia se dibujan por código (`tools/art/vida.py`, `oficios.py`) y los emotes salen de `tools/art/ui.py` (`emote(name)`). El slice suma `tools/art/actividades.py`, registrado en `build.py`, y amplía `furniture.py`/`furniture_micro.py` para las piezas de casa que cambian (maceta con tierra, zapatos, arrocera con vapor). Densidades: macro a 80 px/m, micro a 1600 px/m (1 mm ≈ 1,6 px).

## Rutina de cada hito

Cada hito sigue los mismos pasos, en este orden:

1. **Simulación**: stores, sistemas, acciones y eventos en `internal/sim/*`, con invariantes (`world.AddInvariant`) y `HashVersion`/`SnapshotSchema` incrementados si cambia un store.
2. **Contenido**: objetos en `content/objects/*.json` y escenarios del hito generados por `tools/scenarios/actividades.py`, que corre después de `cuerpo.py`.
3. **Sprites**: piezas en `tools/art/actividades.py` (o en el script de mueble que corresponda), entradas en `content/art/src/casa/atlas.json`, `python tools/art/build.py` y `go run ./cmd/artpack`. Toda entidad nueva visible tiene variante **micro**. Las que se ven desde lejos tienen también **macro** (o `hidden: true` explícito). Las piezas que tapan a un tiny llevan `foreground`, y las que se acarrean llevan `carry`.
4. **Lectura**: texto de actividad con motivo y destino (`cmd/game/hud.go`), entrada de crónica (`view.storyText`) y emote si la acción no se reconoce por el sprite.
5. **Sonido**: los que correspondan en `tools/audio/sounds.json` y `python tools/audio/build.py`. Solo presentación.
6. **Cierre**: escenarios en `tests/scenarios/`, sección de la spec y captura del hito con `-shot` en `docs/specs/img/`. Antes del commit: `gofmt -l .`, `go vet ./...`, `go test ./...`, `python tools/art/build.py -check` y `go run ./cmd/artpack -check`.

## M61 Agua del baño y la cocina

**Vencimiento genérico.** Un componente `expires {at}` y un sistema `expire` (paquete `sim/expire`, instalado en `engine.New`, en la última fase) que hace despawn de lo vencido en orden de ID y emite `expired {entity, kind}`.
*Por qué:* las gotas son el primer caso, pero las manchas de sol, las marcas y los restos lo reutilizan. Un solo sistema evita un contador por tipo.

**Sed.** Necesidad `thirst` en los tinies, con un decaimiento más lento que `hunger`. `Randomize` la sortea al final, así no cambia lo que ya se sortea.

**Gotas.** `water_drop {volume}` con `offers` (`drink`, escala tiny, efecto en `thirst`) y un componente `water` (el equivalente de `food` para el agua) para que la colonia pueda acarrearla. Tomar una gota la achica. Al quedar vacía hace despawn.

**Depósito por tipo.** `portions` pasa a un mapa ordenado por recurso (`food`, `water`), serializado como lista ordenada por nombre (D5). `colony.stock` y `colony.Deficit` reciben el recurso. El nido ofrece `drink` mientras haya agua.

**Rutinas.** `wash_hands` en el lavamanos (necesidad de la giantess `hygiene`, nueva y lenta) y `wash_dishes` en la bacha después de `snack`/`breakfast` si hay vajilla en la cocina. Las dos terminan con `spawn` de 2–4 gotas sobre el piso del baño y la mesada, con dispersión.

**Sprites.**
- `water_drop`, micro: gota de 6 × 6 px con brillo, clip `idle` de 2 cuadros (destello) y clip `low` más chico para cuando le queda poco. Sin macro (`hidden`).
- `puddle`, micro: charco de 14 × 6 px, la variante de una gota grande en el borde de la ducha.
- Tapita del nido (`nest_water`), micro: tapa de botella de 18 × 10 px vacía, a medias y llena, compuesta sobre el nido según `water`.
- `ui/emote_thirst.png` en `ui.py`.
- La giantess macro: clip `wash` (de pie frente al lavamanos o la bacha, 4 cuadros, manos moviéndose), en `giantess.py`.

**Lectura y sonido.** «Toma agua en el baño», «Lleva agua al nido». Goteo al aparecer gotas y canilla abierta durante `wash_*`.

**Cierre.** `tiny_drinks_drop`, `drop_evaporates` (nadie llega y la gota vence), `water_hauled_to_nest` (sube `water` del nido y otro tiny toma ahí).

## M62 Granos de la arrocera

**Carga pesada.** `heavy {bearers}` en el objeto. Un trabajo `haul` sobre algo `heavy` tiene `bearers` plazas, reservadas como las de `gathering`. El grano solo se mueve cuando están todos los portadores. La velocidad es la del más lento y el grano va en el punto medio. Si uno se va (urgencia, huida), el grano queda en el piso y la plaza vuelve a la cola.
*Por qué:* reutilizar el patrón de plazas evita una segunda forma de reservar. El grano queda en un solo lugar autoritativo (D1) aunque lo lleven dos.

**Rutina.** `cook_rice` en la arrocera: tiempo de espera largo con la tapa cerrada y, al servir, `spawn` de 3–6 granos alrededor de la arrocera, en la mesada.

**Granos.** `rice_grain` con `food` y `portions: 3`. Al depositarse suma tres porciones.

**Decisión.** Un tiny que ve un grano y no tiene compañero libre compara esperar (curva de `patience`, en `agent.doing`) contra una migaja cercana. El emote de espera muestra que está esperando a alguien.

**Sprites.**
- `rice_grain`, micro: 10 × 5 px, idle y variante `carried` con la orientación del recorrido. `carry` desplaza el grano a la altura de los brazos de los dos.
- Tiny, micro: clip `carry_heavy` (brazos arriba, paso corto), en `sprites.py` para las cuatro direcciones y las tres paletas.
- Arrocera macro: clip `cooking` con vapor de 3 cuadros, en `furniture.py`. Micro: vista de la base y la tapa sobre la mesada, en `furniture_micro.py`.
- La giantess macro: clip `cook` frente a la mesada (reutiliza las partes de `reach`).
- `ui/emote_wait_partner.png`.

**Lectura y sonido.** «Espera a alguien para cargar un grano», «Cargan un grano con Tiny 03». Burbujeo de la arrocera y el clic de la tapa.

**Cierre.** `two_tinies_carry_grain`, `grain_partner_leaves` (una sombra hace huir a uno, el grano queda y otro lo retoma), `grain_split_portions` (tres porciones al depositarlo).

## M63 Tesoros perdidos y regalos

**Tesoros.** `trinket {kind, uses}` con los tipos `button`, `clip`, `thread`, `graphite`. Llevan `landmark` para que se aprendan sin ofertas. La colonia publica `haul` hacia el `hoard` del nido cuando no hay déficit de comida ni de agua.

**Botín.** `search_spot` suma `loot`: una lista de tesoros posibles con peso, que se sortea con `w.RNG("colony.loot")` al revisar el punto. Se colocan puntos nuevos bajo el futón, frente al placard bajo, junto al zapatero y debajo de la alfombra.
*Por qué:* reutiliza `search` (M40): un punto que a veces da comida y a veces un tesoro le da motivo a revisar rincones lejos de la comida.

**Colección.** `hoard {items}` en el nido (lista ordenada de IDs). Cada tesoro distinto sube el piso de `fun` de la colonia (mientras haya un tipo nuevo, la diversión decae más despacio). Los tesoros se pueden sacar para usarlos (M64, M65).

**Regalos.** Acción `gift`: un tiny con mucha `fun` y un tesoro repetido lo deja en un `gift_spot` (borde del escritorio, mesa del living). Si la giantess lo ve, `encounter` suma la reacción `notice_gift`: se detiene, mira y baja su enfriamiento de `share` en esa superficie.

**Sprites.**
- Micro: `trinket_button` (16 × 16 px, cuatro agujeros), `trinket_clip` (20 × 8 px, metálico), `trinket_thread` (madeja de 14 × 10 px y un cabo suelto), `trinket_graphite` (12 × 4 px). Variante `carried` de cada uno con `carry`.
- Nido micro: capa `hoard` en tres etapas (un tesoro, varios, pila) compuesta sobre la caja de fósforos.
- Macro: punto de 2 px para los tesoros sueltos, visible solo con `selected_only`.
- `ui/emote_treasure.png` (tiny) y `ui/emote_curious.png` en la giantess para `notice_gift`.

**Lectura y sonido.** «Revisa debajo del futón», «Encontró un botón». Crónica: «La colonia encontró su primer botón», «Ella encontró el regalo de Tiny 05». Tintineo del clip y un sonido suave de hallazgo.

**Cierre.** `tiny_finds_trinket` (sorteo reproducible), `hoard_raises_fun`, `giantess_notices_gift`.

## M64 Tender una cuerda

**Anclajes.** `anchor_site {to, length, required}`: un punto en el borde de una superficie elevada (escritorio → futón, mesada → piso de la cocina, asiento del sofá → alfombra) con el extremo de destino. Tiene `landmark`, así que hay que haberlo visto explorando.

**Obra.** Pide un `thread` del `hoard` (job `haul`) y un `build` (M41). Al terminar, el sistema `colony.rope` emite `set_portal` con escala `tiny`, tramo de trepada con longitud 3D real (`Projection`) y la cuerda como entidad (`rope`) dueña del portal.
*Por qué:* el estado cambia desde un sistema (D7) y el portal queda igual que uno de escenario: la navegación no necesita saber que lo construyeron.

**Fragilidad.** La cuerda tiene `fragile`: un apoyo de palma o una pisada en el anclaje la corta. El portal se cierra (los que trepan caen con `space.falls`) y el hilo vuelve como tesoro suelto.

**Sprites.**
- `rope`, macro: línea de 1 px del anclaje al piso, con `pivot` en el nudo (se dibuja estirada entre los dos extremos del `Frame`).
- `rope`, micro: nudo en el borde (12 × 10 px) y vista de costado de la cuerda para `Space.Climb`, como tile vertical de 6 px de ancho.
- `anchor_site`, micro: marca en el borde, hilo enrollado listo y nudo atado.
- Tiny, micro: clip `climb_rope` (mano sobre mano), en `sprites.py`.
- `rope_cut`, micro: cabos sueltos y deshilachados, 2 cuadros.

**Lectura y sonido.** «Ata una cuerda al borde del escritorio», «Baja por la cuerda». Crónica: «La colonia tendió una cuerda al futón». Tensión de hilo al trepar y chasquido al cortarse.

**Cierre.** `colony_ties_rope`, `rope_route_used` (con la cuerda hecha, un acarreo elige el camino más corto), `rope_cut_by_palm` (la palma corta la cuerda, quien trepaba cae y la ruta se recalcula).

## M65 Marcar el camino

**Material.** `graphite_source` (el lápiz del escritorio) ofrece `scrape`, que crea un `trinket_graphite`. Los tesoros de grafito también sirven.

**Marcas.** `waymark {kind, target, wear}` con los tipos `path` (un portal conocido), `danger` (una zona `danger`) y `refuge` (un `shelter`). Acción `mark`: un tiny con grafito que cruza un paso que la colonia no tiene marcado lo marca al pasar.

**Lectura de marcas.** `ai.look` trata la marca como una nota: un tiny que la ve aprende su `target` (el portal en `known`, la zona en `danger`, el refugio) en ese mismo tick, sin volver al nido.
*Por qué:* hoy el conocimiento solo viaja por `colony.share` en el nido. Las marcas permiten que una expedición o un tiny varado aprendan en el camino, que es justamente lo que falta para explorar lejos.

**Desgaste.** Cada pisada o apoyo encima sube `wear`. Al llegar al máximo, la marca vence (`expires` de M61 con `at` inmediato).

**Cartógrafo.** Nuevo `member.role = "mapper"`, con afinidad por `mark` y `explore`, en el reparto de `ai.SetJobScorer`.

**Sprites.**
- `waymark`, micro: tres símbolos de 10 × 10 px en trazo de grafito (flecha, cruz de peligro y techito de refugio), con dos cuadros de desgaste (entero, borroso).
- Macro: `hidden`, salvo con `selected_only` en el miniplano.
- Lápiz: macro (12 × 2 px sobre el escritorio, en `macro.py`) y micro (80 × 10 px con punta raspada y polvo).
- Tiny, micro: clip `draw` agachado (2 cuadros) en `sprites.py`.
- `ui/emote_mark.png`.

**Lectura y sonido.** «Marca el paso del pasacables», «Leyó una marca de peligro». Rasguido de grafito.

**Cierre.** `mark_teaches_portal` (un tiny que nunca cruzó aprende el paso por la marca), `danger_mark_avoided` (la ruta evita la zona), `mark_worn_by_step`.

## M66 Depósito avanzado en la cocina

**Depósito.** `depot {resources, target}`: un `build_site` (M41) en el zócalo de la cocina, junto al pie de la mesada, que al terminar recibe `depot` y `shelter` débil. No tiene `rest`.

**Meta lejana.** La colonia publica acarreos de origen → depósito cuando el origen está más cerca del depósito que del nido (coste de ruta con `nav`) y el depósito está por debajo de su meta. Hay relevos depósito → nido cuando el nido tiene déficit. Los granos (M62) y las gotas (M61) son los primeros clientes.

**Riesgo.** Un depósito con comida en la línea de pisadas de la cocina atrae a la giantess (`encounter` lo cuenta como actividad). Si un apoyo destruye el depósito, la comida queda suelta como migajas.

**Sprites.**
- `depot`, micro: obra (tapa de yogur volcada, fibras) y depósito terminado en tres etapas de llenado (vacío, a medias, lleno), 34 × 22 px, con `foreground` en el borde.
- Macro: marcador de 4 × 3 px junto al zócalo, para que el depósito se vea en el plano completo.
- `depot_ruined`, micro: tapa aplastada y migajas desparramadas.

**Lectura y sonido.** «Lleva un grano al depósito de la cocina», «Trae comida del depósito al nido». Ficha de colonia: porciones en el nido y en el depósito.

**Cierre.** `colony_founds_depot`, `depot_relay` (con el nido en déficit, un relevo vacía el depósito), `depot_starves_when_far` (sin fuentes cerca, la colonia no lo llena y no gasta viajes).

## M67 Lo que traen los zapatos

**Salir y volver.** La giantess suma la necesidad `outside` (lenta). La oferta `go_out` en la puerta exterior la lleva al genkan, se pone los zapatos y queda `away {until}`: sin cuerpo, sin `stride` y sin `encounter`, con el ID vivo (D3). `come_home` la devuelve por la misma puerta. El escenario fija la duración y el sorteo va por `w.RNG("encounter.outing")`.
*Por qué:* es la primera ventana larga sin amenazas, buena para trabajos lejanos. Probarla aquí, antes de los proyectos largos, asegura que todos los sistemas toleran a una giantess ausente.

**Zapatos.** `shoe` (dos entidades en el zapatero o en el genkan, según la rutina) con un Space interior tiny y `shelter`. Mientras la giantess los tiene puestos no están en el genkan: un tiny adentro tiene que salir antes de que ella se los ponga. `agent.care` hace que los sacuda antes (lo que obliga a los tinies a salir).

**Restos.** Al volver, `come_home` hace `spawn` de `debris {loot}` en el genkan: semillas (`seed`, M69), arenilla, un pétalo (tesoro), con `expires` largo. Se sortean con `w.RNG("encounter.debris")`.

**Sprites.**
- Zapatos macro: par de 10 × 6 px en el genkan y en el zapatero, en `furniture.py`. Micro: exterior en el piso (`foreground` en la puntera) e interior como tile del Space del zapato.
- Puerta exterior macro: clips `open`/`closed`, en `apartment.py`.
- La giantess macro: clips `shoes_on` (agachada, 4 cuadros) y `step_out`. Fuera de la casa no se dibuja; el HUD muestra «Salió».
- `debris`, micro: semilla (8 × 5 px), arenilla (puntos dispersos, 12 × 8 px), pétalo (12 × 8 px, tesoro). Variantes `carried`.
- `ui/emote_shake.png` para el titubeo al sacudir.

**Lectura y sonido.** «Ella salió de casa» en la crónica y el HUD, «Revisa lo que quedó en el genkan», «Se esconde en un zapato». Llaves, puerta y pasos en el genkan. El ambiente de la casa baja mientras ella no está.

**Cierre.** `giantess_goes_out` (sale, vuelve y la continuidad desde snapshot durante `away` da el mismo hash), `debris_after_return`, `tiny_hides_in_shoe` (y sale cuando ella lo sacude).

## M68 Hora del día y calor

**Reloj del día.** Un recurso derivado `daytime` (hora en [0, 24) calculada con aritmética entera a partir del tick y de `start_hour`, guardado en una entidad de mundo y hasheado). Un día dura lo que fije el escenario (por defecto 20 minutos reales). Nunca se lee el reloj real (D2).

**Calor.** Necesidad `warmth` en los tinies. Baja más rápido de noche y en el piso del genkan y el baño.

**Manchas de sol.** `sunpatch {window, area}` por ventana. Un sistema mueve su rectángulo sobre el piso según la hora (proyección entera de la ventana con `fixedmath.Angle`) y lo apaga de noche. Estar dentro sube `warmth` (acción `bask`, con plazas de hecho limitadas por el área).

**El futón de noche.** `warm_spot` en el edredón mientras la giantess duerme: mucho `warmth`, pero ella se da vuelta cada tanto (sorteo con `w.RNG("encounter.sleep")`). Antes de darse vuelta proyecta una amenaza del cuerpo, como las pisadas de `crush.Threats`, y el `reflex` de los tinies los hace huir.

**Sprites.**
- `sunpatch`, macro y micro: rectángulo de luz cálida semitransparente con borde suave, dibujado por la presentación sobre el piso a partir del `Frame` (sin cuadro fijo) y una textura de polvo en suspensión (2 cuadros) en micro.
- Tinte global por hora (madrugada, mañana, tarde, noche) en la capa de escena. Solo presentación: el `Frame` lleva la hora.
- La giantess macro: clip `sleep_turn` sobre el futón (3 cuadros), en `giantess.py`.
- Tiny, micro: clip `bask` (estirado al sol) y `shiver`, en `sprites.py`.
- `ui/emote_cold.png` y `ui/emote_warm.png`.

**Lectura y sonido.** «Toma sol junto a la ventana», «Duerme al calor del futón». Hora en el HUD. Ambiente que cambia entre el día y la noche (pájaros, silencio nocturno) y el roce del edredón al darse vuelta.

**Cierre.** `sunpatch_moves` (la mancha recorre el piso y se apaga de noche, con el mismo hash al continuar), `tiny_warms_in_sun`, `rollover_threat_flee`.

## M69 Huerta en la maceta

**Superficie.** La tierra de la maceta pasa a ser un Space tiny, con acceso trepable por el borde del tiesto desde el piso del living. La planta en sí sigue sin ser transitable.

**Parcela.** `plot {stage, moisture, growth}`: acción `plant` con una `seed` (M67), `water` con agua del nido o de una gota (M61), y crecimiento por ticks (más rápido con una `sunpatch` encima, M68). Sin agua, la humedad baja y el crecimiento se frena. Llega a cinco etapas (semilla, brote, tallo, planta, fruto). En la última ofrece `harvest`, que crea porciones de comida y vuelve a la etapa planta.

**Hojas.** La planta deja caer `leaf` cada tanto (sorteo con `w.RNG("colony.plant")`). Una hoja es un material que cuenta como tres fibras para `build_site`, y un refugio con hojas recibe un `shelter.strength` mayor.

**Sprites.**
- Tierra de la maceta, micro: tile de 1600 px/m con terrones y raíces. Borde del tiesto con `foreground`. Macro: la maceta de `furniture.py` suma un borde visible de tierra.
- `plot`, micro: cinco etapas (12 × 6 a 40 × 48 px) con variante seca (tono amarillento) para cuando falta agua.
- `seed` (de M67) y `leaf`, micro: hoja de 30 × 16 px, idle y `carried`.
- Refugio de hojas, micro: variante de `refuge_site/micro_built` con techo de hoja.
- Tiny, micro: clip `dig` (2 cuadros) y `pour` (vertiendo la tapita).
- `ui/emote_sprout.png`.

**Lectura y sonido.** «Riega la parcela», «Cosecha la planta». Crónica: «Brotó la semilla que trajo Tiny 02». Tierra removida y el chorro de agua.

**Cierre.** `colony_plants_seed` (de la semilla a la cosecha), `plot_dries_without_water`, `leaf_refuge_stronger` (un refugio de hojas resiste una palma que uno de fibras no aguanta).

## M70 Ver la tele y contar historias

**Función.** `viewing_spot {source, seats}` en la alfombra y el apoyabrazos del sofá, ligado al televisor. Solo se ofrece mientras alguien usa la oferta de la TV (la giantess mirando). La acción `watch_tv` ocupa una plaza, sube `fun` a todos los presentes y termina cuando se apaga la tele o por urgencia. Las plazas generalizan `gathering` (M42) a N plazas.

**Riesgo.** Los lugares están cerca de los pies de la giantess sentada. Si se levanta o se acomoda, `stride` y `reflex` los hacen huir.

**Relatos.** Al cerrar una expedición con hallazgos, el que regresa crea un `tale {expedition, findings, until}` en el nido. La acción `listen` (con `gathering`, plazas para varios) sube `fun` a quienes escuchan y los hallazgos les llegan como si los hubieran visto: refuerza `colony.share` con un momento legible.

**Sprites.**
- Reflejo de la pantalla, micro: luz azulada que parpadea sobre la alfombra y los tinies (3 cuadros, capa aditiva), según la oferta activa del `Frame`.
- `viewing_spot`, micro: marcas de plaza solo con `selected_only`.
- Tiny, micro: clips `sit_watch` y `cheer` (brazos arriba), en `sprites.py`.
- Grupo en macro: los tinies ya se dibujan; se agrega un brillo de TV de 2 px en la alfombra en macro.
- `ui/emote_tv.png` y `ui/emote_tale.png` (burbuja con una pisada adentro).

**Lectura y sonido.** «Mira la tele con Tiny 01 y Tiny 04», «Cuenta lo que vio en el genkan». El audio de la TV ya existe; suma una risa chiquita en `cheer` y un murmullo de relato.

**Cierre.** `tinies_watch_tv`, `tv_off_ends_viewing` (se apaga la tele y se liberan todas las plazas), `expedition_tale_heard` (quien escucha aprende el hallazgo).

## M71 Partida e integración

**Contenido.** `tools/scenarios/actividades.py` corre después de `cuerpo.py` y agrega a `content/partidas/casa.json` todo lo del slice: puntos con botín, gotas iniciales, anclajes, el lápiz, el sitio del depósito, los zapatos, las manchas de sol, la parcela y los lugares para ver la tele. El orden de regeneración queda `partida.py` → `vida.py` → `oficios.py` → `cuerpo.py` → `actividades.py`, y se documenta en `CLAUDE.md`.

**Sorteo.** `Randomize` agrega al final `thirst`, `warmth`, la hora inicial y el agua inicial del nido, para no cambiar lo que ya se sortea.

**Ajuste.** Utilidades, decaimientos y plazos sobre las semillas de referencia (1, 7, 21, 42 y 99). Con diez actividades nuevas, el riesgo es que una tape a las demás (como la curiosidad en M43). Se mide con un histograma de acciones completadas por semilla.

**Compatibilidad.** Las aceptaciones de asientos, encuentros, vida, oficios y cuerpo corren sobre `casa.json`. Si cambian, se revisa que sigan cumpliendo sus criterios y se actualizan a propósito, sin relajarlos. La mañana y el descuido no cambian.

**Sprites e interfaz.**
- Íconos de la ficha de colonia para agua, tesoros, depósito y huerta (12 × 12 px, en `ui.py`), con sus cantidades.
- Miniplano: marcas de cuerdas, depósito, parcela y marcas de camino conocidas.
- Nota de «Nueva partida» en el menú.
- Revisión de legibilidad en `cmd/game/readability.go` para los sprites nuevos más chicos (gotas, grafito, semillas).

**Aceptación.**
- **Cobertura del mapa.** En la primera hora de juego de cada semilla de referencia, los tinies completan acciones en al menos cinco de las seis habitaciones y en al menos seis de las diez actividades.
- **Cadenas.** En al menos una semilla se completa una cadena entre habitaciones: semilla del genkan → parcela regada con agua del baño → cosecha depositada en el nido.
- **Coordinación.** Ninguna reserva doble de plaza, tesoro, grano o anclaje. Cada recurso entregado cuenta una vez.
- **Interrupciones.** La huida corta cualquier actividad nueva y las reservas quedan libres.
- **Continuidad.** El mismo log y el mismo hash al continuar desde snapshots a mitad de un acarreo cooperativo, con la giantess fuera de casa y con la parcela en crecimiento.
- **Referencias.** Las capturas (`go run ./cmd/capturas`) y la mezcla de referencia sin cambios o actualizadas a propósito.

## Contratos de integración

- Todo se activa por contenido: sin las entidades nuevas, una colonia y la giantess se comportan como hoy. La mañana, el descuido y sus capturas no cambian.
- Las acciones nuevas siguen el contrato de `ai.Action`: sin estado propio, todo en `agent.doing` o en el componente del objetivo. Mover con `nav.Approach`/`nav.Halt` y reservar con `jobs.Reserve`.
- Cada sorteo usa su propio stream (`colony.loot`, `colony.plant`, `encounter.outing`, `encounter.debris`, `encounter.sleep`). Listas y mapas se guardan ordenados (D5).
- Ningún float en estado espacial: el movimiento de las manchas de sol, la cuerda y el grano entre dos portadores se calcula en `units.Length` con `fixedmath` (D1, D6).
- Cada hito que agrega o cambia un store incrementa `HashVersion` y `SnapshotSchema` y cierra con escenarios, sprites y la sección de spec.
- La presentación lee `view.Frame` y los eventos. Nada en `render`/`ui`/`sound` lee el mundo, y `internal/archtest` sigue pasando.
- Las piezas de arte salen de scripts (`tools/art/*.py`). Nunca se tocan `content/art/casa.{png,json}` ni `content/audio/` a mano.
