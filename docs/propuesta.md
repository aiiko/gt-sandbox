# \*\*Proyecto: Giantess Life Simulator\*\*

Análisis de diseño, simulación y arquitectura técnica

Tu idea tiene una combinación bastante particular: tomar las situaciones y diferencias de tamaño de los juegos Giantess clásicos, pero reemplazar su estructura de RPG por un \*\*sandbox de simulación emergente\*\*, donde los personajes tienen necesidades, relaciones, rutinas y comportamientos autónomos.

La diferencia fundamental sería esta:

\*\*En un RPG tradicional, la giantess hace algo porque un evento programado lo indica. En tu juego, lo hace porque la simulación la llevó a tomar esa decisión.\*\*

Por ejemplo, imaginemos una casa habitada por una giantess y una pequeña civilización de tinies que vive debajo de su escritorio.

La giantess se despierta, tiene hambre y decide preparar el desayuno. Se levanta de la cama, camina hasta la cocina y mueve algunos objetos. Mientras tanto, los tinies continúan con su propia rutina: recolectan comida, construyen refugios, exploran y reaccionan a los cambios en el entorno.

Un tiny podría subirse accidentalmente a su zapatilla. Cuando ella empieza a caminar, el tiny pasa a desplazarse junto con ella. Otro podría estar explorando su cabello, mientras un grupo aprovecha que salió de la habitación para recuperar recursos.

Ninguna de esas situaciones necesitaría una misión escrita a mano.

El juego produciría historias mediante la interacción de sus sistemas.

LA COMBINACIÓN QUE PROPONGO

\*\*RimWorld\*\*

Colonias, tareas, supervivencia y consecuencias

\*\*The Sims\*\*

Necesidades, autonomía, personalidad y relaciones

\*\*Dwarf Fortress\*\*

Mundo persistente, causalidad y profundidad de simulación

\*\*Giantess RPGs\*\*

Perspectiva, exploración, interacciones entre escalas y situaciones especiales

Mi propuesta es diseñarlo como \*\*un simulador de ecosistemas a distintas escalas\*\*, no como un RPG con más NPC ni como un clon completo de RimWorld.

Esto importa porque las mecánicas más interesantes de tu juego aparecen precisamente donde se cruzan ambos mundos: la misma acción cotidiana puede ser trivial para una giantess y un acontecimiento enorme para los tinies.

## \*\*1. Qué podemos extraer de los juegos de referencia\*\*

### \*\*1.1. Los juegos Giantess: conservar la temática, cambiar la estructura\*\*

\[X\](https://images.openai.com/static-rsc-4/pCrNRjhE4TrfVIklpLLSluxOmSgwlL4g3fxqojxT8R4jRu89ZNe6Zb9heTB5\_2D\_CQiQ7uwWGF1XY4i\_t4ZwHNthuQVzq4XjYV-iO5o3Ax1UTsjM\_82f60MM9pZFvXTfz6TGYEyZzUIm6Ae5J7xV4cl0x-pOGX8iu4UJqty1BqA?purpose=inline)

\*\*Size Matters — Uru\*\*

Sus mapas de tamaños diferentes, zonas como ciudades diminutas y situaciones de interacción entre personajes de distintas escalas son referencias útiles. La documentación comunitaria de sus versiones muestra una gran cantidad de áreas y eventos particulares.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://aryion.com\\&sz=32)

aryion.com+2



\*\*Lo que adoptaría:\*\* la exploración ambiental y la variedad de situaciones que surgen de las diferencias de tamaño. Lo que cambiaría: depender de mapas y eventos independientes para representar cada interacción.

\[Shrink High Download Game - excellentlasopa\](https://images.openai.com/static-rsc-4/rPA2KOeqozjjHMyVdXmymVfc9OPY7YlAq1vwDfqNvyZ-6VHIvti60wpDovm8CzX059eOEykcXGo1q8j9PY2R6Houf\_xRSFaiK9FwcilgB9BtuqCTh7IvAB9WrAyqdm7YGeqYSXdUhYhTe3mzeHbwpA7BnuqsvLYhJ5GOw6enGsU?purpose=inline)

\*\*Shrink High / 縮小学園\*\*

Destaca por la vulnerabilidad de los personajes diminutos, la exploración y el peligro asociado a encuentros con personajes enormes. Un equipo que desarrolló \*Astra's Plaything\* en 2026 también identificó esa sensación de desventaja como uno de los grandes atractivos del género.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://www.furaffinity.net\\&sz=32)

furaffinity.net+1



\*\*Lo que adoptaría:\*\* que el tamaño afecte realmente la exploración, el desplazamiento, la percepción y la supervivencia.

\[PULSE - ゲームを記録・管理できる究極のゲームライブラリ\](https://images.openai.com/static-rsc-4/WB2SyeNoBmdBQrNN8\_-K1ab0rZol2HlSyYqlOSNOzKZ-oTwem5oiD8T6CC4iOAWTgHdNOugBCeHsny4tuJWl2HKCXwyCIsnlrhT4iQy3emPwsf1HNY2DlK0AN0RmVs2yRCIobHUh8y-cfDW7hsg1Zg7YzwuH54ghzuDSlKSf3UM?purpose=inline)

\*\*Terada Ochiko — obras y juegos asociados\*\*

Las referencias relacionadas con Ochiko aportan situaciones de escala extrema y exploración de espacios asociados a personajes gigantes. Hay juegos comunitarios atribuidos a su trabajo, como \*President Matsudo's Ambition\*. También figura como artista colaborador en \*Shrink High Gaiden\*, no como su único desarrollador.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://aryion.com\\&sz=32)

aryion.com+2



\*\*Lo que adoptaría:\*\* el cuerpo de una giantess como entorno explorable, con sus propios obstáculos, lugares y eventos, pero convertido en un sistema reutilizable.

\[小人となって死を潜り抜ける会話型アドベンチャー『SAEKO: Giantess Dating Sim』。家主の巨大少女の機嫌を損ねると…？【電撃インディー#670】 - 電撃オンライン\](https://images.openai.com/static-rsc-4/I4lkcyrEwXi9GeZB7QzGR\_f3mOBEanIBklVj2bZf3aArRcafciPk24h2yJDHB-yJ9i46tZrXV2lx06f4lhJ8V12DJ201rADtewoE0OQc00bWDfCShOGgmkcSRx0DwRO2u4n9brao9lrNGFqdfUOowjc5JsabOIUQXp039F4itsc?purpose=inline)

\*\*SAEKO: Giantess Dating Sim — referencia adicional\*\*

Es especialmente relevante porque utiliza dos estilos de pixel art y distintos sistemas de juego para representar la vida de humanos diminutos bajo el control de una gigante. Sin embargo, su experiencia sigue estando organizada alrededor de una historia y decisiones narrativas.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://saekogame.com\\&sz=32)

SAEKO: Giantess Dating Sim+1



\*\*Lo que adoptaría:\*\* distintas representaciones visuales del mismo universo según la perspectiva y la distancia.

La oportunidad está en convertir lo que estos juegos presentan mediante escenas diseñadas individualmente en \*\*reglas generales que producen muchas situaciones distintas\*\*.

Por ejemplo:

| Situación                    | RPG tradicional          | Tu simulador                                          |
| ---------------------------- | ------------------------ | ----------------------------------------------------- |
| Una giantess camina          | Secuencia predeterminada | Acción autónoma con consecuencias espaciales          |
| Un tiny queda atrapado       | Evento de una misión     | Resultado de movimiento, obstáculos y entorno         |
| Un tiny sube a una persona   | Mapa especial            | Transición a una superficie móvil explorable          |
| Una giantess descubre tinies | Diálogo predefinido      | Reacción según percepción, personalidad y relación    |
| Una casa cambia              | Evento argumental        | Consecuencia de actividades e interacción con objetos |

### \*\*1.2. Los simuladores: qué sistema tomar de cada uno\*\*

| Juego               | Sistema más valioso                                         | Cómo se adapta                                                |
| ------------------- | ----------------------------------------------------------- | ------------------------------------------------------------- |
| RimWorld            | Jobs, prioridades, colonias y emergencias                   | Los tinies trabajan, construyen y se organizan                |
| The Sims            | Necesidades, autonomía, personalidad y objetos interactivos | Las giantesses tienen una vida independiente                  |
| Dwarf Fortress      | Persistencia, causalidad, materiales, historia emergente    | Los acontecimientos dejan consecuencias duraderas             |
| Oxygen Not Included | Recursos, condiciones ambientales y redes                   | Las colonias tinies aprovechan agua, comida, calor y refugios |
| Kenshi              | Mundo autónomo y libertad de roles                          | Los personajes existen independientemente del jugador         |

Hay dos referencias técnicas especialmente interesantes.

\*\*RimWorld:\*\* su arquitectura distingue la selección de trabajos de su ejecución y distribuye evaluaciones de IA entre distintos ticks. Esa separación evita tener que reconsiderar todas las decisiones de todos los personajes en cada frame.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://github.com\\&sz=32)

GitHub+1



\*\*The Sims:\*\* Robert Zubek, que trabajó en ese ámbito, describió un sistema de inteligencia artificial basado en necesidades y objetos que ofrecen comportamientos. En vez de programar cada actividad completa dentro del personaje, parte de la información sobre qué puede hacerse está en los propios objetos del mundo.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://robert.zubek.net\\&sz=32)

robert.zubek.net+1



Esto último encaja particularmente bien en una casa llena de objetos.

## \*\*2. Qué tipo de juego debería ser\*\*

Veo tres direcciones posibles.

| Enfoque                   | Experiencia                                                     | Complejidad |
| ------------------------- | --------------------------------------------------------------- | ----------- |
| Colony Sim                | Gestionás una sociedad de tinies dentro de la casa              | Alta        |
| Life Sim                  | Controlás o influís en una giantess que vive su rutina          | Media-alta  |
| Sandbox emergente híbrido | Alternás perspectivas e intervenís sobre un ecosistema autónomo | Muy alta    |

\*\*Elegiría el tercero como objetivo final, pero desarrollaría primero el núcleo de un Colony Sim.\*\*

Esto no significa descartar jugar como giantess. Significa construir primero una simulación capaz de funcionar aunque no controles directamente a nadie.

La estructura podría tener tres perspectivas:

\*\*Vista de la casa\*\*

Observás las habitaciones, rutinas y desplazamientos de las giantesses. Similar a The Sims.

\*\*Vista de los tinies\*\*

Gestionás colonias, exploración, recursos, peligros y trabajos. Similar a RimWorld.

\*\*Vista del cuerpo\*\*

Explorás superficies asociadas a una giantess, como manos, brazos, ropa o cabello, mientras ella sigue moviéndose y actuando.

La tercera perspectiva es lo que más diferenciaría tu proyecto.

Pero también introduce el problema técnico más importante de todo el diseño: \*\*un personaje que simultáneamente es un agente autónomo, un objeto físico y un mundo que otros personajes pueden explorar\*\*.

## \*\*3. El sistema central: un mundo con varias escalas\*\*

No construiría tres mundos aislados. Construiría un único estado de simulación con múltiples espacios conectados.

### \*\*3.1. La casa es una jerarquía de espacios\*\*

Arquitectura conceptual: los espacios tienen posiciones, conexiones y relaciones de pertenencia. Los cuerpos son espacios móviles dentro del mismo mundo.

En esta estructura existirían varios tipos de superficies:

| Superficie       | Ejemplo                           | Comportamiento                       |
| ---------------- | --------------------------------- | ------------------------------------ |
| Estática         | Suelo de la habitación            | No se desplaza                       |
| Elevada          | Escritorio, mesa, cama            | Tiene altura y puntos de acceso      |
| Móvil rígida     | Zapato, taza, objeto transportado | Se desplaza con una entidad          |
| Móvil articulada | Mano, brazo, cabeza               | Depende de la pose de la giantess    |
| Construida       | Refugio o puente tiny             | Puede modificarse durante la partida |

Esto permite agregar espacios nuevos sin reprogramar todo el juego.

### 3.2. Dos escalas visuales y una sola simulación

Una giantess de 170 cm mide unas 212 veces lo que un tiny de 8 mm. La vista macro encuadra habitaciones y permite seguir la rutina de la giantess. La vista micro encuadra una superficie navegable y permite leer a cada tiny, su carga y los bordes del terreno.

Las dos vistas representan las mismas entidades, poses y estados. Cada superficie tiene su grilla de navegación y su arte puede usar una densidad de píxeles distinta. La cámara cambia el encuadre y la representación visual mientras el mundo conserva sus coordenadas físicas. Las reglas de proyección, densidad y transición se detallan en [Dirección visual](#11-dirección-visual).

## \*\*4. La giantess como mapa móvil\*\*

Este es el sistema que diseñaría primero en un prototipo técnico, incluso antes de construir una IA compleja.

La idea es asignarle a cada giantess un conjunto de espacios explorables, cada uno asociado a una parte de su cuerpo o a un objeto que lleva puesto.

Ejemplo de superficies explorables

Cabeza

Cabello, rostro

Brazos

Antebrazos, codos

Manos

Palmas, dedos

Torso

Ropa, hombros

Piernas

Rodillas, ropa

Pies

Calzado, superficies

Cada superficie tendría su propio mapa, puntos de conexión y reglas físicas.

Por ejemplo:

\`\`\`

Mano izquierda
  ├── Palma
  │    ├── Dedo índice
  │    ├── Dedo pulgar
  │    └── Otros dedos
  └── Muñeca
       └── Antebrazo
            └── Brazo
                 └── Hombro
                      └── Torso

\`\`\`

Un tiny podría desplazarse entre estas regiones, siempre que sean accesibles según la pose actual.

### \*\*4.1. Cómo sería el movimiento\*\*

Imaginemos que un tiny está sobre el antebrazo.

La giantess levanta la mano para tomar una taza. Técnicamente ocurren varios cambios:

1. Su IA decide ejecutar la acción \`TakeCup\`.
2. El sistema de acciones cambia la pose del brazo.
3. El espacio asociado al antebrazo actualiza su transformación.
4. El tiny conserva su posición local respecto al antebrazo, si sigue sujeto a él.
5. Las nuevas condiciones del movimiento pueden provocar deslizamiento, pérdida de equilibrio o la necesidad de agarrarse.
6. Si cae, pasa del espacio corporal a otra superficie del mundo, resolviendo dónde aterriza.

La posición del tiny sería relativa a su superficie:

\\\\\[ P\\\_{\\text{mundo}} = T\\\_{\\text{giantess}}\\cdot T\\\_{\\text{parte}}\\cdot P\\\_{\\text{local}} \\\\\]

En este modelo, las transformaciones convierten la posición local del tiny a una posición global. En una simulación top-down con altura discreta, se puede implementar con transformaciones 2D y un componente adicional de elevación.

El detalle importante es que \*\*la giantess puede continuar ejecutando su rutina mientras los tinies exploran su cuerpo\*\*.

### \*\*4.2. El problema de las poses\*\*

No alcanza con dibujar el cuerpo como un mapa fijo.

Una giantess parada, sentada, acostada o caminando tiene distintas superficies accesibles.

| Pose                  | Efecto sobre los tinies                                                    |
| --------------------- | -------------------------------------------------------------------------- |
| Parada                | Acceso principalmente por zapatos, ropa cercana al suelo y objetos         |
| Sentada               | Nuevas rutas mediante silla, piernas o manos apoyadas                      |
| Acostada              | Más superficies corporales quedan cercanas y accesibles                    |
| Caminando             | Los espacios corporales se desplazan y algunas rutas se vuelven inestables |
| Sosteniendo un objeto | El objeto puede convertirse en una ruta de transporte                      |

No empezaría con un esqueleto físico articulado completo. Para pixel art top-down, sería mucho más económico definir \*\*poses discretas, puntos de anclaje y transiciones controladas\*\*.

Eso ofrece buena parte de la experiencia sin implementar una simulación anatómica continua.

### \*\*4.3. Dos sistemas de navegación conectados\*\*

El recorrido de un tiny podría ser:

\`\`\`

Suelo de habitación
    ↓ Escalar
Pata de escritorio
    ↓
Superficie del escritorio
    ↓ Contacto con mano
Palma de la giantess
    ↓ Gate de conexión
Muñeca
    ↓
Antebrazo

\`\`\`

El motor primero calcularía una ruta general por los espacios conectados y después rutas locales dentro de cada mapa.

Se parece al principio de pathfinding jerárquico HPA\\\*, que reduce el problema de navegación mediante clusters y conexiones de nivel superior.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://theory.stanford.edu\\&sz=32)

theory.stanford.edu+1



La parte nueva sería que algunas conexiones aparecen, desaparecen o cambian de posición según lo que hace la giantess.

Por ejemplo, cuando coloca su mano sobre el escritorio, se habilita temporalmente un camino entre ambos espacios.

\*\*Esto es muchísimo más flexible que teletransportar al tiny a un mapa especial mediante un evento.\*\*

## \*\*5. Sistemas de simulación: cuáles combinaría y cómo\*\*

La clave es que los sistemas puedan interactuar entre sí sin que cada uno necesite conocer todos los detalles de los demás.

Mi selección sería la siguiente:

| Sistema                          | Inspiración               | Prioridad         | Dificultad |
| -------------------------------- | ------------------------- | ----------------- | ---------- |
| Mundo multi-escala               | Diseño propio             | Crítica           | Muy alta   |
| Necesidades y rutinas            | The Sims                  | Crítica           | Media      |
| Trabajos y prioridades           | RimWorld                  | Alta              | Media      |
| Navegación entre superficies     | RimWorld / HPA\\\*          | Crítica           | Muy alta   |
| Percepción y conocimiento        | The Sims / IA de agentes  | Alta              | Alta       |
| Relaciones y personalidad        | The Sims                  | Media             | Media      |
| Interacciones físicas            | Simulaciones emergentes   | Alta              | Alta       |
| Construcción tiny                | RimWorld / Dwarf Fortress | Media             | Alta       |
| Objetos interactivos             | The Sims                  | Crítica           | Media      |
| Historia emergente               | Dwarf Fortress            | Alta              | Media      |
| Simulación de fluidos y ambiente | Oxygen Not Included       | Baja inicialmente | Muy alta   |

### \*\*5.1. IA de las giantesses: Utility AI + acciones\*\*

En lugar de escribir una secuencia rígida como:

\`Despertar → Ir al baño → Desayunar → Mirar televisión\`

utilizaría una IA basada en necesidades y puntuaciones de utilidad (\*Utility AI\*).

Por ejemplo, una giantess podría tener estos valores:

\*\*Estado de una giantess\*\*

Ejemplo

Hambre

25%

Energía

75%

Higiene

50%

Diversión

40%

Social

65%

La IA evalúa cuáles de sus acciones disponibles mejoran estas necesidades y considera distancia, disponibilidad, personalidad y riesgos.

Una versión simple de la puntuación sería:

\\\\\[ U(a)=\\sum\_i w\_i\\cdot \\Delta N\_i(a)-C(a)-R(a) \\\\\]

Donde:

- \\\\(U(a)\\\\): utilidad de la acción.
- \\\\(w\_i\\\\): importancia de cada necesidad para el personaje.
- \\\\(\\Delta N\_i(a)\\\\): beneficio esperado para esa necesidad.
- \\\\(C(a)\\\\): costo de tiempo, esfuerzo y desplazamiento.
- \\\\(R(a)\\\\): penalización por riesgo.

La fórmula es una propuesta simplificada; en un sistema más avanzado utilizaríamos curvas de utilidad no lineales y un mecanismo que impida cambiar de decisión continuamente. La literatura sobre \*Needs-Based AI\* describe precisamente estas estrategias de selección y arbitraje.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://robert.zubek.net\\&sz=32)

robert.zubek.net+1



Pero agregaría otra dimensión fundamental: \*\*la personalidad\*\*.

Dos giantesses podrían tener exactamente la misma hambre y, aun así, elegir actividades distintas.

Una puede ser ordenada, cuidadosa y observadora. Otra, distraída, impulsiva y despreocupada. Esas diferencias no deberían limitarse al diálogo: modificarían la frecuencia con que revisan el suelo, mueven objetos, investigan ruidos o reaccionan ante los tinies.

### \*\*5.2. Smart Objects: los objetos ofrecen comportamientos\*\*

Este patrón de The Sims resulta particularmente apropiado para tu juego.

El sofá no sería solamente un sprite y una colisión. También ofrecería acciones.

| Objeto       | Acción giantess                    | Acción tiny                            |
| ------------ | ---------------------------------- | -------------------------------------- |
| Sofá         | Sentarse, dormir, mirar televisión | Esconderse, explorar, construir debajo |
| Taza         | Llenar, beber, transportar         | Escalar, recolectar gotas, refugiarse  |
| Escritorio   | Trabajar, apoyar objetos           | Construir una colonia, explorar        |
| Zapato       | Equipar, quitar, mover             | Entrar, esconderse, viajar sujeto a él |
| Mano         | Agarrar objetos, gesticular        | Sujetarse, caminar, interactuar        |
| Refrigerador | Guardar y retirar comida           | Buscar recursos, explorar alrededores  |

El sistema de selección consulta las acciones que ofrecen los objetos accesibles, filtra las imposibles y evalúa las restantes.

Así, al agregar un objeto nuevo, podemos introducir comportamientos sin modificar cada agente individualmente.

### \*\*5.3. IA de tinies: Jobs + Utility AI\*\*

Los tinies necesitan un sistema adicional: trabajo colectivo.

Un tiny hambriento puede querer buscar comida. Pero también podría tener asignada la construcción de un refugio.

Propongo esta jerarquía de decisiones:

\`\`\`

Tiny Agent
  |
  +-- Emergency Response
  |    +-- Escape danger
  |    +-- Seek shelter
  |    +-- Recover balance
  |
  +-- Personal Needs
  |    +-- Eat
  |    +-- Sleep
  |    +-- Rest
  |
  +-- Colony Jobs
  |    +-- Gather resources
  |    +-- Build
  |    +-- Transport items
  |    +-- Explore
  |
  +-- Social / Idle
       +-- Interact
       +-- Wander
       +-- Relax

\`\`\`

Cada trabajo podría tener una secuencia de pasos, similar al enfoque de jobs y ejecutores de RimWorld.

Por ejemplo, un trabajo \`CollectCrumb\`:

\`Buscar recurso → Reservarlo → Calcular ruta → Llegar → Recoger → Transportar → Depositar\`

La reserva es importante: evita que diez tinies decidan recoger exactamente la misma migaja.

También permite interrumpir trabajos cuando ocurre una emergencia y retomarlos después.

### \*\*5.4. Percepción: los tinies no deberían ser omniscientes\*\*

Este es un sistema fácil de subestimar y uno de los que más mejoraría la simulación.

Cada agente necesitaría distinguir tres cosas:

- Qué existe realmente en el mundo.
- Qué puede percibir en este momento.
- Qué cree o recuerda acerca del mundo.

Por ejemplo, una giantess podría caminar sobre una alfombra donde viven varios tinies, pero no necesariamente saber que están allí.

Los tinies, en cambio, podrían percibir sus pasos mediante vibraciones antes de verla.

Mi diseño incluiría canales de percepción como visión, sonido, vibración y contacto. Su sensibilidad dependería de la escala, la distancia, el material y las capacidades del personaje.

Así pueden aparecer situaciones emergentes como una giantess que encuentra rastros de actividad diminuta, investiga y descubre gradualmente la existencia de una colonia.

### \*\*5.5. Física: simular consecuencias, no átomos\*\*

No intentaría reproducir físicamente todas las interacciones de una giantess con millones de partículas.

Usaría un enfoque de \*simulación basada en eventos físicos\*.

Por ejemplo, cuando la giantess apoya una taza sobre la mesa, el motor produce un evento de contacto que contiene la posición, superficie y características físicas relevantes.

El sistema de consecuencias puede entonces evaluar qué ocurre con los objetos y agentes próximos.

Un vaso que se derrama podría producir un volumen de líquido simplificado que se propaga sobre unas cuantas celdas. Los tinies cercanos reaccionarían según sus capacidades y ubicación.

Para movimientos grandes, como caminar, usaría volúmenes de colisión y trayectorias barridas, no una búsqueda exhaustiva sobre todas las celdas diminutas.

Esto permite tener consecuencias espaciales incluso cuando la cámara está observando otra habitación.

## \*\*6. El gran desafío: que la simulación siga funcionando fuera de cámara\*\*

Una giantess puede estar en la cocina mientras el jugador observa una colonia debajo de la cama.

\*\*El mundo no debe detenerse ni depender de lo que está dibujado en pantalla.\*\*

Pero simular cada personaje con el máximo detalle todo el tiempo sería innecesariamente costoso.

Por eso recomiendo \*Simulation Level of Detail\* (LOD de simulación).

| Nivel        | Qué se simula                                 | Ejemplo                               |
| ------------ | --------------------------------------------- | ------------------------------------- |
| Detallado    | Movimiento, contactos, acciones y navegación  | Tinies cerca de la cámara             |
| Intermedio   | Trabajos, rutas resumidas, necesidades        | Tinies activos en otra zona           |
| Agregado     | Recursos, actividad y progreso por intervalos | Colonia distante sin eventos críticos |
| Event-driven | Actualizaciones cuando sucede algo relevante  | Objetos inactivos y procesos lentos   |

La decisión de qué nivel usar \*\*no debería depender solamente de la cámara\*\*.

Una zona con un acontecimiento importante necesita mantenerse suficientemente detallada aunque el jugador no la esté mirando. Lo mismo sucede si una giantess está trasladando tinies entre habitaciones.

Para evitar contradicciones, mantendría siempre las colisiones y transferencias críticas bajo un sistema lógico autoritativo.

La filosofía general es compatible con una recomendación del propio Tarn Adams: construir modelos comprensibles, evitar complejidad innecesaria y apoyarse en analogías del mundo real.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://www.gameaipro.com\\&sz=32)

gameaipro.com+1



# \*\*7. Enfoque técnico: qué tecnología utilizaría\*\*

Tomando Go como lenguaje principal, que además conocés por tu trabajo, elegiría una arquitectura donde toda la simulación pueda ejecutarse independientemente del motor visual.

## \*\*7.1. Comparación de motores y frameworks\*\*

Evaluaciones relativas para este proyecto, no benchmarks medidos.

| Stack                  | Simulación compleja | Pixel art | Desarrollo con IA             | Riesgo técnico |
| ---------------------- | ------------------- | --------- | ----------------------------- | -------------- |
| \*\*Go + Ebitengine\*\*    | Excelente           | Excelente | Excelente para lógica y tests | Bajo-medio     |
| Go + raylib-go         | Excelente           | Muy bueno | Muy bueno                     | Medio          |
| Godot + GDScript/C#    | Muy bueno           | Excelente | Muy bueno                     | Medio          |
| Godot + Go/GDExtension | Muy bueno           | Excelente | Medio                         | Alto           |

\*\*Mi elección: Go + Ebitengine.\*\*

\[Ebitengine Tutorial: Boids Flocking in Golang - BioErrorLog Tech Blog\](https://images.openai.com/static-rsc-4/QLn1O7m54PPnYR5OosaJ7pmCy2Ff8W4VXX6cAR038c5gBiAApjHdMyW86uLjYjIGAN9Y-8gPJL2SI4XdNQQUkYuQaEBhFU9sdNzSJuwWINw1t5ActSMe3czk4ICTUyyc2q7LdLA2Jeq6P9Ok5Y4k9ysvxWJBuEU2\_QLDKOVJZms?purpose=inline)

\*\*Ebitengine\*\*

Está especializado en 2D y ofrece rendering acelerado por GPU, batching automático cuando es posible y una API bastante reducida. Desde la versión 2.10, publicada en septiembre de 2026, soporta builds de escritorio en Go puro sin requerir compilador C.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://ebitengine.org\\&sz=32)

ebitengine.org+1



Eso es muy conveniente para la asistencia con IA: menos capas de integración, más código Go convencional y una separación clara entre lógica y presentación.

También hay una novedad reciente especialmente útil para tu workflow.

\*\*Ebitengine 2.10: pruebas visuales automatizadas para agentes\*\*

Ebitengine incorporó un mecanismo experimental de virtualización de aplicaciones. Un programa externo puede controlar los inputs, avanzar ticks y recuperar imágenes renderizadas, permitiendo probar juegos mediante agentes de código.

Los desarrolladores publicaron además un recurso llamado \`run-ebitengine-app-headless\` destinado a estos workflows. Requiere habilitación explícita y un entorno gráfico en el host, pero permite pruebas sin mostrar la ventana normal del juego.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://ebitengine.org\\&sz=32)

Ebitengine+1



Para un proyecto desarrollado con Opus o Codex, esto representa una ventaja muy interesante: \*\*el agente puede comprobar visualmente si sus cambios funcionaron, además de ejecutar tests Go\*\*.

Como alternativa, raylib-go también permite desarrollar una simulación independiente y conectar un renderer sencillo. Su principal desventaja en este caso es depender de bindings de raylib y su capa de compilación C.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://github.com\\&sz=32)

GitHub



Godot ofrece un editor visual más completo, mejor tooling incorporado y un desarrollo de UI probablemente más rápido. Sin embargo, usar Go dentro de Godot mediante \`godot-go\` agrega una integración experimental. Su propio repositorio advierte que sigue en desarrollo activo y que puede tener problemas de memoria.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://github.com\\&sz=32)

GitHub



Si priorizáramos crear rápidamente escenas, animaciones y herramientas visuales por encima de Go puro, Godot sería el competidor más fuerte.

## \*\*7.2. Stack recomendado\*\*

| Capa                      | Tecnología                                                   |
| ------------------------- | ------------------------------------------------------------ |
| Lenguaje principal        | Go                                                           |
| Motor 2D                  | Ebitengine 2.10+                                             |
| Simulación                | Paquetes Go puros                                            |
| Entidades                 | Structs y stores propios inicialmente; Ark ECS si hace falta |
| Mapas diseñados a mano    | JSON del proyecto. LDtk queda como posible importador        |
| Sprites y animación       | PNG y metadatos JSON. Aseprite es una opción de autoría       |
| UI                        | Componentes propios sobre Ebitengine                         |
| Navegación                | A\\\* local + grafo jerárquico                                 |
| Definiciones de contenido | JSON                                                         |
| Tests                     | \`go test\`, benchmarks, fuzzing                               |
| Profiling                 | \`pprof\`                                                      |
| Guardado                  | Snapshots versionados + registro de eventos relevantes       |

El renderer de la próxima etapa cargará las exportaciones de arte y los datos del proyecto directamente. La elección del editor se puede ajustar durante la producción.

### \*\*¿Usaría ECS?\*\*

No empezaría con un ECS complejo.

En la primera versión utilizaría entidades identificadas por IDs estables, structs Go y sistemas independientes. Es más sencillo de depurar y de mantener para un proyecto de aprendizaje.

Si el profiling muestra que las iteraciones sobre miles de entidades son un cuello de botella, incorporaría un ECS basado en arquetipos.

Hoy consideraría especialmente \[\*\*Ark ECS\*\*\](https://github.com/mlange-42/ark?utm\_source=chatgpt.com), sucesor recomendado por los autores de Arche. Proporciona consultas tipadas, relaciones entre entidades, operaciones por lotes y herramientas de serialización.&amp;#x20;

\[image\](https://www.google.com/s2/favicons?domain=https://github.com\\&sz=32)

GitHub+1



No hay necesidad de resolver una hipotética carga de 100.000 agentes antes de que existan veinte.

## \*\*8. Arquitectura propuesta\*\*

\`\`\`

giantess-sim/
│
├── cmd/
│   ├── game/            # Juego Ebitengine
│   └── simbench/        # Benchmarks headless
│
├── internal/
│   ├── sim/
│   │   ├── world/       # Estado y espacios
│   │   ├── entity/      # IDs y componentes
│   │   ├── spatial/     # Índices y contactos
│   │   ├── navigation/  # Rutas y portales
│   │   ├── ai/          # Utility AI
│   │   ├── jobs/        # Tareas y reservas
│   │   ├── needs/       # Necesidades
│   │   ├── body/        # Espacios corporales
│   │   ├── physics/     # Interacciones físicas
│   │   ├── events/      # Eventos del mundo
│   │   └── save/        # Persistencia
│   │
│   ├── renderer/        # Dibujo Ebitengine
│   ├── camera/          # Perspectivas y zoom
│   ├── ui/              # Inspector, órdenes
│   └── assets/          # Carga de recursos
│
├── content/
│   ├── actors/
│   ├── objects/
│   ├── actions/
│   ├── maps/
│   └── sprites/
│
└── tests/
    ├── scenarios/
    └── fixtures/

\`\`\`

La separación más importante es esta:

\`\`\`

             PLAYER INPUT
                  |
                  v
          Commands / Orders
                  |
                  v
        +-------------------+
        |  SIMULATION CORE  |
        |     Pure Go       |
        |                   |
        | World / AI / Jobs |
        | Body / Physics    |
        +-------------------+
             |         |
             v         v
           Events   Snapshot
             |         |
             v         v
          History    Renderer
                      Ebitengine
                         |
                         v
                      Screen

\`\`\`

El renderer únicamente interpreta estados producidos por la simulación.

No debería modificar necesidades, decidir trabajos, resolver contactos ni alterar la posición lógica de los personajes.

Esto permite ejecutar miles de ticks en un test sin abrir una ventana.

### \*\*8.1. Modelo de ubicación\*\*

Lo diseñaría alrededor de una estructura similar a esta:

\`\`\`

type EntityID uint64
type SpaceID uint32

// Coordenadas locales expresadas en milímetros.
type Location struct {
    Space SpaceID

    X int64
    Y int64
    Z int64
}

// Un espacio puede ser fijo o estar sujeto
// a un personaje u objeto móvil.
type Space struct {
    ID     SpaceID
    Parent EntityID
    Anchor string
}

// Conexión navegable entre espacios.
type Portal struct {
    From SpaceID
    To   SpaceID

    Enabled bool
    Cost    float64
}

\`\`\`

Para el prototipo, cada \`Space\` tendría una transformación local hacia su entidad padre o hacia el mundo global. En producción también necesitaríamos IDs versionados, geometrías de portales y reglas de accesibilidad.

La ventaja de este modelo es que un tiny puede encontrarse en \`SpaceID=LeftHand\` mientras la mano se mueve.

Su posición local no necesita actualizarse con cada desplazamiento global de la giantess. Solamente cambian las transformaciones de la jerarquía.

Ese diseño reduce trabajo y simplifica muchísimo la lógica de transporte.

### \*\*8.2. Simulación con fixed ticks\*\*

Recomendaría comenzar con una frecuencia lógica de 20 TPS y renderizar a 60 FPS.

| Sistema                        | Frecuencia inicial sugerida |
| ------------------------------ | --------------------------- |
| Renderizado                    | Hasta 60 FPS                |
| Movimiento y contactos activos | 20 Hz                       |
| Reevaluaciones urgentes de IA  | 5–10 Hz                     |
| Selección normal de trabajos   | 1–2 Hz o por eventos        |
| Necesidades y metabolismo      | Cada 1–5 segundos           |
| Eventos ambientales lentos     | Por intervalos o eventos    |
| Simulación distante            | Agregada según actividad    |

Son valores de diseño iniciales, no cifras universalmente óptimas.

También distribuiría los cálculos caros entre distintos ticks utilizando offsets por entidad, para no recalcular todas las decisiones en el mismo momento.

Es fundamental que el reloj del mundo sea único, independientemente de cuántas cámaras o espacios tengamos abiertos.

### \*\*8.3. Navegación y colisiones\*\*

Usaría tres estructuras complementarias:

\*\*1. Índice espacial global:\*\* permite encontrar entidades y objetos próximos sin revisar todas las entidades del juego.

\*\*2. Navegación local:\*\* A\\\* sobre tiles, regiones o celdas de una superficie.

\*\*3. Grafo de espacios:\*\* conecta habitaciones, muebles, superficies elevadas y partes del cuerpo de las giantesses.

Una ruta larga primero recorrería el grafo y luego resolvería trayectorias locales.

Cuando una giantess se mueve, el sistema actualiza sus puntos de conexión y solo invalida las rutas afectadas.

El mayor desafío no es ejecutar A\\\*, sino \*\*gestionar los cambios topológicos\*\*: una mano deja de tocar una mesa, un zapato es levantado del suelo o un objeto bloquea una entrada.

Por eso los portales, superficies y contactos deben ser conceptos independientes del renderizado.

# \*\*9. Cómo lo desarrollaría con IA\*\*

No le pediría a Opus que construyera todo el juego de una sola vez.

Haría que implementara sistemas aislados, con interfaces y criterios de aceptación verificables. Así evitamos que una modificación del sistema corporal termine rompiendo las necesidades o el guardado.

## \*\*Workflow\*\*

1. \*\*Especificaciones pequeñas\*\*

   Crear documentos breves para \`WorldSpace\`, \`TinyMovement\`, \`Jobs\`, \`BodySpace\` y otros sistemas. Cada documento define responsabilidades, contratos, invariantes y casos límite.
2. \*\*Simulación headless primero\*\*

   Antes de dibujar un personaje, probar que pueda cambiar de superficie, recibir una orden y completar un trabajo. El resultado debe poder comprobarse con \`go test\`.
3. \*\*Escenarios reproducibles\*\*

   Un escenario incluiría un estado inicial, una semilla aleatoria, acciones externas y cantidad de ticks. Después compararíamos resultados e invariantes, no solamente imágenes.
4. \*\*Integración visual y pruebas automáticas\*\*

   Una vez que la simulación funciona, conectar el renderer Ebitengine. Aprovechar su mecanismo de virtualización para verificar visualmente inputs, cámara, animaciones y UI.
5. \*\*Benchmarks y profiling\*\*

   Repetir escenarios con 20, 100, 500 y 1.000 tinies. Medir CPU, memoria, latencia por tick, búsquedas de caminos y allocaciones con las herramientas oficiales de Go.&amp;#x20;

   \[image\](https://www.google.com/s2/favicons?domain=https://go.dev\\&sz=32)

   The Go Programming Language+1



### \*\*Tests que consideraría imprescindibles\*\*

| Test                              | Resultado esperado                                                         |
| --------------------------------- | -------------------------------------------------------------------------- |
| Tiny sobre mano en movimiento     | Conserva su ubicación local                                                |
| Giantess cambia de habitación     | Sus superficies móviles siguen conectadas a ella                           |
| Mano toca mesa                    | Aparece una ruta válida                                                    |
| Mano deja de tocar mesa           | La ruta se invalida                                                        |
| Dos tinies buscan un mismo objeto | Se respeta la reserva                                                      |
| Se cierra un camino               | La IA recalcula o cancela el trabajo                                       |
| Cambia la cámara                  | No cambia el resultado lógico                                              |
| Guardar y cargar                  | Se recuperan posiciones, relaciones y trabajos                             |
| Misma semilla e inputs            | Produce el mismo estado lógico, bajo condiciones deterministas controladas |

Este último punto requiere disciplina: orden estable de ejecución, control de aleatoriedad y cuidado con la concurrencia.

No utilizaría goroutines indiscriminadamente dentro de cada agente. Inicialmente ejecutaría la simulación secuencialmente. Después paralelizaría sistemas concretos si los benchmarks lo justifican.

## 10. Una mañana en la casa

El escenario de referencia reúne un dormitorio con cama y escritorio, una cocina y un living. Una giantess autónoma convive con una colonia de 20 tinies en el escritorio. El simulador ya reproduce este recorrido a partir de necesidades, trabajos, percepción y contactos. [Spec del escenario](specs/manana-en-la-casa.md)

1. La giantess decide desayunar porque su hambre es alta.
2. Deja una taza y unas migajas sobre el escritorio.
3. La colonia tiny detecta nuevos recursos y asigna recolectores.
4. Un tiny alcanza su mano mientras está apoyada sobre el escritorio.
5. La giantess levanta la mano y se dirige a otra habitación.
6. El tiny viaja sobre una superficie móvil y termina explorando una nueva zona.

La siguiente etapa lleva el recorrido al pixel art. En macro se debe entender qué hace la giantess y adónde va. En micro se debe poder seguir al mismo tiny desde el escritorio hasta la palma y la mesa del living. La secuencia se tiene que entender por lo que se ve en ambas vistas, con la historia como apoyo.

## 11. Dirección visual

### Cámara y escala

El juego usa **pixel art top-down ortográfico de estética anime japonesa**, ambientado en un departamento japonés. El suelo se proyecta en planta, con el norte arriba en ambas vistas. Los sprites muestran caras y frentes en tres cuartos mediante el dibujo. Las posiciones, alturas y contactos provienen de la simulación.

| Vista | Qué se lee | Densidad inicial de arte | Escena de referencia |
| --- | --- | --- | --- |
| Macro | Habitaciones, puertas, muebles, rutina y desplazamiento de la giantess | 16 px por 40 cm. La referencia de la giantess de pie usa un lienzo de 31 × 38 px | La casa completa y el viaje del dormitorio al living |
| Micro | Tinies individuales, migajas, bordes y pasos entre superficies | 16 px por 10 mm. El tiny tiene una silueta visible de 12–16 px de alto | El nido y las migajas sobre el escritorio, el borde de la palma y el descenso a la mesa del living |

La escena se dibuja en un lienzo lógico de **480 × 270 px**. La ventana de escritorio admite ese lienzo a partir de 1×. Se usa el mayor factor entero que entre, con muestreo por vecino más cercano (nearest-neighbor) y bandas centradas cuando sobra espacio. La posición se interpola primero y se alinea a la cuadrícula de píxeles al dibujar. Las densidades son parámetros para la primera prueba de arte y se ajustan con capturas de la casa y la palma. La cuadrícula visual es independiente de las celdas de navegación.

Macro muestra a los tinies mediante un marcador del nido y un indicador para el tiny seleccionado. Micro muestra el sprite de cada tiny, su carga y suficiente textura para distinguir madera, cerámica, papel y piel. De la giantess se dibuja la parte relevante, como la palma y la manga que entran por un borde del encuadre. La taza conserva su color, el lado del asa y su posición de apoyo en las dos representaciones.

Entrar en micro centra la superficie elegida o al actor en seguimiento. Al volver a macro se recupera el encuadre anterior, salvo que haya una entidad en seguimiento. En ese caso se conserva el seguimiento. Un corte con fundido breve de unos 180 ms confirma el cambio de escala en el rótulo de ubicación. La opción de movimiento reducido usa un corte directo. Mientras un tiny viaja sobre la palma se lo sigue en el mismo modo micro. Si baja a otra superficie, cambia el contexto de ubicación y el encuadre continúa siguiendo al actor.

### Ambientación

La casa es un departamento japonés contemporáneo, como los de un anime de la vida cotidiana. La ambientación cambia el dibujo y no la simulación: cada mueble conserva su posición, su altura y sus ofertas del escenario.

| Espacio | Rasgos de dibujo |
| --- | --- |
| Entrada y umbrales | Puertas corredizas con marco de madera clara. Las puertas entre ambientes alternan fusuma decorado y shoji de papel según el cuarto |
| Dormitorio | Washitsu con tatami y bordes de tela, futón con edredón estampado, ventana con shoji y un escritorio de madera para estudiar. El escritorio mantiene sus 75 cm porque ahí vive la colonia |
| Cocina | Cocina compacta de departamento con anafe de dos hornallas, arrocera, heladera chica con imanes, noren en el paso y vajilla de cerámica |
| Living | Piso de madera, zabuton, planta en maceta, televisor bajo y una mesa de madera clara a la altura del escenario. Un kotatsu requeriría bajar la superficie en el escenario y queda como variante posterior |

Los detalles que se leen a escala micro también son japoneses: la caja de fósforos del nido tiene una etiqueta con kana dibujada como textura sin texto legible, la taza es un yunomi o una taza con flor esmaltada y la mesa puede tener un posavasos de bambú. Los rótulos y la interfaz están en castellano.

### Personajes y lenguaje visual

La dirección toma del anime cotidiano las expresiones claras, el vestuario reconocible y la luz de mañana sobre interiores habitados. En el estudio, el cabello violeta, el cárdigan lila y el acento coral identifican a la giantess. Al acercarse a la palma, la misma manga, su puño claro y el botón coral mantienen la continuidad entre escalas.

Los personajes usan proporciones chibi para dar espacio al rostro y al peinado. En macro, la giantess se lee por el contorno del cabello y la separación entre cabeza, ropa y pies. En micro, el tiny se distingue por sus mechones azules, su bufanda menta y unos pocos píxeles de expresión. Cada postura conserva el apoyo de los pies o la mano sobre la superficie.

El sombreado es plano, con bordes definidos como en la animación con celdas. Pelo, piel y ropa se dibujan con un color base, una sombra y uno o dos brillos agrupados. El contorno usa ciruela en cabello y ropa y malva cálido en piel. Los bordes que reciben luz pueden usar un tono del propio material. La madera reserva sus vetas más finas para micro y la cerámica concentra sus reflejos en el borde y el asa. Los ojos de los personajes llevan siempre un píxel de brillo, también en el tiny de 16 px.

### Expresiones y señales

Los personajes muestran su estado con símbolos de manga sobre la cabeza: «!» al descubrir algo, «?» al explorar, gota de sudor ante un peligro o al ser levantado, «zzz» al descansar, nota musical al divertirse y vena de enojo cuando se interrumpe una acción. Cada símbolo sale de un evento o del estado de la simulación, dura un tiempo fijo de presentación y no modifica el mundo. En micro se dibujan sobre cada tiny y en macro solo sobre la giantess y el tiny seleccionado. Cada símbolo tiene una forma distinta, así que también sirve como señal legible sin color.

La giantess tiene un retrato de busto con expresiones neutral, contenta, hambrienta, sorprendida y somnolienta. El retrato acompaña el panel de historia y el inspector resumido, y la expresión sigue sus necesidades y su acción en curso. En micro nunca se le ve la cara, así que el retrato mantiene el vínculo con ella mientras se sigue a la colonia.

En micro, la presencia de la giantess se anticipa con su sombra. Antes de que la mano entre en cuadro, una sombra violeta difusa crece sobre la superficie y se define al apoyarse. La sombra sale de la pose y la altura de la mano en el `Frame`.

### Paleta, materiales y capas

La casa combina luz cálida de arriba a la izquierda con sombras violetas. El decorado usa contrastes suaves y reserva los acentos más vivos para personajes, objetos interactivos y señales. Las sombras de contacto son cortas y la altura se aprecia por el canto del mueble, la separación de la sombra y el solapamiento.

| Familia | Muestra de partida | Uso |
| --- | --- | --- |
| Marfil | `#FFF1D8` | Luz en paredes, cerámica y acabados de ropa |
| Madera durazno | `#D5A07A` | Escritorio, con veta visible en micro |
| Ciruela | `#3F365C` | Cabello, contornos y sombras profundas |
| Lila | `#9A8BE3` | Cárdigan, manga y detalles de vestuario |
| Coral | `#FF9CAD` | Lazo, botón y flor esmaltada de la taza |
| Menta | `#9CE0D6` | Accesorios del tiny y cerámica |
| Turquesa claro | `#76D8CB` | Selección y paso disponible, con señal por forma |
| Tatami | `#C8C98F` | Tatami del dormitorio, con borde de tela ciruela |
| Papel shoji | `#F7EEDD` | Shoji y fusuma, con marco de madera durazno |

Las superficies tienen tres valores principales de luz y sombra. Los tinies se distinguen por silueta y contraste de valor sobre madera y piel. Un paso abierto se marca con dos extremos unidos y un paso cerrado pierde esa conexión. El icono y su forma permiten distinguir ambos estados con la imagen en escala de grises.

Cada pieza de arte declara un pivote de apoyo y, cuando corresponda, un anclaje de mano o carga. La presentación dibuja suelo y sombras, superficies y frentes, y ocupantes por sus puntos de apoyo. Los frentes de primer plano pueden tapar parte de los sprites. La altura y el parentesco separan superficies superpuestas, y el orden dentro de cada superficie se resuelve con la posición de apoyo en pantalla y un desempate estable. Si un borde oculta a la entidad seleccionada, su contorno o marcador sigue visible. El punto de selección coincide con la entidad que se ve bajo el cursor.

### Personajes, poses y animación

| Elemento | Animación mínima para la escena |
| --- | --- |
| Giantess | Reposo y caminata en cuatro direcciones, mano apoyada y recogida, comer sentada y mirar TV. Escribir, dormir y tomar un snack pueden reutilizar poses coherentes cuando el motor muestra esas acciones |
| Palma | Palma elevada y apoyada, con sombras y continuidad de manga. El dibujo se ancla al Space de la mano |
| Tiny | Reposo, caminata y acarreo en cuatro direcciones, y caída. La carga se mueve con el tiny y su silueta sigue visible |
| Objetos | Taza en mano y sobre la mesa, migajas sueltas y depósito del nido. Los cambios de estado se reflejan en la vista correspondiente |
| Expresiones | Símbolos de manga para giantess y tinies, y retrato de la giantess con cinco expresiones |

Se parte de ciclos de caminata de cuatro cuadros y una cadencia visual de 6–10 cuadros/s. Las direcciones se eligen entre variantes dibujadas y el pivote queda estable entre frames. El inicio de una pose o animación sale del estado o los eventos de la simulación. El paso entre mesa y mano solo se muestra transitable mientras el portal correspondiente está activo. Una animación puede suavizar el dibujo del gesto, con el apoyo, la posición de los pasajeros y el indicador de paso alineados al estado del mundo.

### Interfaz y alcance de arte

La interfaz de partida muestra hora, pausa y velocidad, ubicación, selección y seguimiento. El panel de historia destaca pocos cambios relevantes de la rutina y permite localizar al actor o la superficie cuando siguen disponibles. En micro, el rótulo identifica el contexto con una jerarquía como «Dormitorio › Escritorio» o «Giantess › Palma». La intervención del jugador se limita a los controles existentes de cámara, tiempo, inspección y órdenes de destino donde la simulación las acepta. Los overlays técnicos y el inspector de componentes quedan disponibles en el visor de debug.

La interfaz de partida tiene su propio estilo, distinto del visor de debug: paneles marfil con esquinas redondeadas y borde ciruela, una tipografía pixel con tildes y eñes, globos de historia con la cola apuntando al actor y el retrato de la giantess junto al panel. Las barras oscuras y la fuente monoespaciada quedan para el modo de debug.

El primer lote de arte cubre suelo, paredes, puertas y mobiliario del dormitorio, cocina y living con la ambientación japonesa. El lote micro cubre escritorio, palma, mesa del living, nido, tinies, taza y migajas. Cada pieza tiene un archivo fuente editable con capas, versionado junto al atlas, y un comando reproducible arma el PNG y el manifiesto a partir de esas fuentes. El atlas exportado no se edita a mano. Las exportaciones que consume el juego son PNG transparentes y metadatos JSON con identificador visual, escala, rectángulos del atlas, pivotes, capas y clips. La asignación de arte usa identificadores visuales estables independientes del nombre mostrado. La paleta y los tamaños se calibran con la primera escena antes de extender el lote.

La [muestra de presentación](specs/presentacion.md#prueba-de-arte) reúne las referencias macro y micro de esta dirección. El [plan de hitos](plan-motor.md) lleva esas reglas al recorrido completo.

### Sonido

El sonido es presentación, como los símbolos de expresión. Sale de los eventos publicados y del `Frame`, nunca lee ni modifica el mundo, y no cambia el hash. Todo sonido tiene un equivalente visual que ya existe (símbolo, señal de paso, sombra), así que la escena se entiende en silencio.

**Perspectiva por escala.** El mismo hecho suena distinto según la vista. En macro, la giantess camina con pantuflas suaves y los tinies casi no se oyen. En micro, la mano que baja es un soplo de aire que crece con la sombra y un golpe grave y cálido sobre la madera, y los pasitos de los tinies se oyen claros.

**Sin música.** No hay banda sonora: la casa suena por su ambiente. Cada habitación tiene un fondo propio (gorriones y un fūrin en el dormitorio, el zumbido de la heladera en la cocina, el tic del reloj y el murmullo de la TV mientras la giantess la mira en el living), y la vista micro suma un tono de sala grave, como si todo estuviera amplificado.

**Sonidos sintetizados.** Todos los sonidos se generan por código, como el arte dibujado por código: un sintetizador simple lee parámetros versionados (forma de onda, envolvente, tono, ruido, filtro) y escribe WAV cortos. Encaja con el pixel art, cada sonido se ajusta editando números y no hay licencias que seguir. Un manifiesto nombra cada sonido con un identificador estable, como el atlas con los visuales.

| Grupo | Sonidos |
| --- | --- |
| Ambiente | Fondo por habitación, tono de sala en micro |
| Giantess (macro) | Pasos según el suelo (tatami, madera, baldosa), noren y fusuma, heladera, taza sobre la mesa, mordiscos, zabuton, encender la TV |
| La mano | En micro: soplo al acercarse, golpe al apoyarse, soplo al levantarse. En macro: un toque suave sobre el mueble |
| Tinies (micro) | Pasitos de tono distinto por paleta, pop al recoger una migaja, plink al dejarla en el nido, saltito al subir a la mano, silbido al caer y «puf» al aterrizar |
| Símbolos | Un sonido por símbolo: blip («!»), blip que sube («?»), gota, ronquidito («zzz»), tres notas, zumbido gruñón (vena) |
| Señal de paso | Dos notas que suben al abrirse el paso y dos que bajan al cerrarse |
| Interfaz | Clic al elegir, campanita al seguir, zoom al cambiar de escala con el fundido, cinta al pausar, hoja al sumar una entrada a la historia |

**Mezcla.** El volumen y el panorama siguen la posición en pantalla, y lo que queda fuera de cuadro no suena, salvo el ambiente. Cada sonido tiene un límite de voces simultáneas, un tiempo mínimo entre repeticiones y una leve variación de tono, para que veinte tinies no saturen. A x4 y x16 se omiten los sonidos menores y quedan los importantes (la mano, las caídas, los descubrimientos). La pausa silencia los efectos y baja el ambiente. Hay volúmenes separados para ambiente, efectos e interfaz, y una tecla para silenciar.

## \*\*12. Riesgos que resolvería temprano\*\*

| Riesgo                          | Consecuencia                                                        | Estrategia                                                |
| ------------------------------- | ------------------------------------------------------------------- | --------------------------------------------------------- |
| Escalas incompatibles           | Personajes atravesando objetos o apareciendo en lugares incorrectos | Unidades físicas canónicas y transformaciones jerárquicas |
| Navegación dinámica             | Tinies atrapados o teletransportados                                | Portales, grafo de superficies y rutas invalidables       |
| Explosión de complejidad        | Proyecto interminable                                               | Vertical slices pequeños                                  |
| IA errática                     | NPC que cambian de opinión constantemente                           | Utility AI con persistencia de objetivos                  |
| Simulación costosa              | Caídas de FPS con muchos agentes                                    | Scheduling, LOD y profiling                               |
| Arte de escala extrema          | Necesidad de sprites enormes                                        | Representaciones visuales específicas por escala          |
| Arte sin fuente editable        | Cada sprite nuevo o corrección obliga a rehacer el atlas            | Fuentes por pieza y exportación reproducible del atlas    |
| Bugs emergentes                 | Situaciones difíciles de reproducir                                 | Replay determinista y escenarios de prueba                |
| Demasiadas interacciones únicas | Alto costo de mantenimiento                                         | Smart Objects y acciones reutilizables                    |
| Muerte injusta o arbitraria     | El jugador no entiende por qué murió un tiny                        | Aviso previo con sombra o vibración, y causa en la historia |
| Muerte en cascada               | La colonia se vacía en pocos minutos y la partida pierde interés    | Reacción de huida, refugios y una giantess que puede tener cuidado |
| IDs colgados tras una muerte    | Trabajos, reservas o seguimiento apuntan a un tiny que ya no existe | Muerte como despawn normal y restos como entidad nueva (D3) |

## 13. Alcance posterior

El foco inmediato es presentar una casa que ya funciona como ecosistema vivo. La giantess puede seguir con su rutina aunque los tinies la observen o interactúen con su entorno. El jugador puede pasar de una escala a la otra y reconocer cómo una acción cotidiana cambia las posibilidades de la colonia.

El próximo slice vertical es [Un descuido en el escritorio](#14-próximo-slice-un-descuido-en-el-escritorio): los tinies pueden morir aplastados. Después, la expansión puede sumar más superficies explorables, construcción tiny, relaciones, personalidad y eventos ambientales. Fluidos, simulación anatómica continua y economías más amplias quedan para etapas posteriores. El [sonido](#sonido) se suma a la escena cerrada sin música; una banda sonora queda para una etapa posterior.

## 14. Próximo slice: Un descuido en el escritorio

La mañana ya muestra que una acción cotidiana de la giantess cambia las posibilidades de la colonia. Este slice agrega la otra cara: **la misma acción puede costarle la vida a un tiny**. Apoyar la mano para desayunar, pisar el tatami o dejar la taza son gestos triviales en macro y catástrofes en micro.

La muerte tiene que salir de la simulación, igual que todo lo demás. Ningún evento escrito dice "acá muere un tiny". Muere porque estaba debajo de una superficie que bajó, no huyó a tiempo y no tenía refugio. Y la giantess no lo hizo a propósito: no lo vio o no tuvo cuidado.

Tres reglas de diseño:

1. **Toda muerte se anuncia.** Antes de cada aplastamiento hay una señal que el tiny puede percibir y el jugador puede ver: la sombra de la mano que baja o la vibración de los pasos. Si un tiny muere, la historia dice qué lo aplastó.
2. **Escapar depende de la situación.** El tiempo de aviso, la velocidad del tiny, la distancia al borde y los refugios cercanos deciden quién se salva. Con los números actuales (tiny a 2 cm/s, palma de 10 × 18 cm, 12 ticks de mano suspendida) solo se salvan los que están cerca del borde. Una giantess cuidadosa sostiene la mano más tiempo, y eso amplía el margen.
3. **La muerte deja consecuencias.** Quedan restos en el lugar, la colonia pierde un miembro y sus trabajos, y la zona queda marcada como peligrosa durante un tiempo.

### 14.1. Cómo funciona

**Aplastadores.** Un aplastador es algo que puede bajar sobre una superficie: la palma, el pie de la giantess, un objeto que cae. Se define con el componente `crusher {force, footprint}`. La huella sale de la grilla de navegación del Space cuando existe. La palma ya tiene una grilla que sigue la silueta de la mano, así que un tiny que queda entre dos dedos sobrevive. Los pies y los objetos usan una huella rectangular o elíptica.

**Momento del impacto.** El aplastamiento ocurre en un instante preciso, cuando la huella toca la superficie:

| Fuente | Cuándo aplasta | Cómo se detecta |
| --- | --- | --- |
| Palma | Al apoyarse sobre un mueble | Su cara de abajo queda a ras del mueble y no lo estaba en el tick anterior |
| Pie | En cada pisada mientras la giantess camina por el suelo | Pisadas discretas derivadas del `mover` (componente `stride`) |
| Objeto | Al aterrizar sobre una superficie | Aterrizaje en `space.falls` o `space.Land` (taza, migajas, objetos soltados) |

Las pisadas son discretas. El sistema avanza un acumulador de distancia y, cada `stride.length`, estampa la huella del pie alternando izquierda y derecha, orientada según la dirección de la marcha. No se calcula un volumen barrido continuo. Una pisada solo afecta a lo que está en el suelo de la habitación: los tinies sobre el escritorio están en otro Space, a otra altura.

**Víctimas.** El aplastamiento afecta a las entidades con `fragile {}` cuyo punto de apoyo cae dentro de la huella, en la misma superficie que la recibe. Lo que viaja sobre el aplastador queda excluido, porque está montado sobre él: el tiny que va en la palma no muere cuando ella la apoya.

**Refugios.** Un objeto con `shelter {strength, extent}` protege a quien está dentro de su extensión si `strength ≥ force`. La caja de fósforos del nido resiste una mano que se apoya, pero no un pie. Por eso el nido es un lugar seguro en el escritorio y en el suelo no lo sería.

**Resultado.** Cada víctima produce el evento `crushed {entity, by, source, space, pos}`. El tiny se destruye con un despawn normal y en su lugar aparecen restos: una entidad nueva con `remains {name, cause, colony, tick}` y el nombre del tiny. Usar una entidad nueva evita IDs colgados (D3). Los sistemas existentes ya resuelven lo demás: `jobs.upkeep` libera las reservas con `holder_gone`, el trabajo vuelve a la cola con `job_unassigned` y lo que el tiny cargaba queda suelto en el lugar.

**Caídas.** Las caídas siguen sin matar. Un tiny de 8 mm cae como un insecto: su velocidad terminal es baja. Es una decisión de escala, coherente con apoyarse en analogías del mundo real, y evita que viajar en la palma sea una trampa.

### 14.2. Amenaza, percepción y huida

**Zonas de amenaza.** Cada tick se calcula la proyección de los aplastadores que están por bajar:

- La mano suspendida sobre un mueble proyecta su huella sobre él mientras está a menos de `ThreatHeight` de la superficie. Es la misma sombra violeta que ya dibuja la vista micro.
- La próxima pisada se conoce por la ruta del `mover` y su huella se publica unos ticks antes de bajar.

Las zonas se derivan del estado, así que viven en un recurso (`world.SetResource`): no se guardan ni entran en el hash.

**Percepción.** `ai.look` suma dos canales. La sombra la percibe cualquier tiny que esté debajo, sin depender de `sight`. La vibración la percibe un tiny en el suelo a menos de `vibration` de una pisada. Es el canal de vibración que describe la sección 5.4.

**Reacción inmediata.** Hoy un agente decide cada 10 ticks, demasiado lento para huir. Cuando un tiny entra en una zona de amenaza, `ai.decide` lo reevalúa después de `agent.reflex` ticks de reacción. Es un rasgo y no un valor fijo, para que la personalidad influya.

**Acción `flee`.** Es la capa de emergencia de la sección 5.3. Su utilidad supera `UrgentUtility`, así que interrumpe incluso las acciones comprometidas. Al interrumpirse, el tiny suelta lo que carga y la migaja queda sobre el escritorio. El destino es la celda transitable más cercana fuera de la huella más un margen, o un refugio si queda más cerca. Si no hay salida, el tiny queda atrapado (`trapped`) y lo muestra la gota de sudor.

**Sustos.** Un tiny que estaba en la zona y salió con vida, o que quedó cerca de la huella, recibe `near_miss`: se suspende como después de una caída y muestra «!».

### 14.3. La giantess y el cuidado

La giantess conoce todos los objetos, pero no ve todos los tinies. Este slice le da un rasgo nuevo, `agent.care` en [0, 1]:

- **Más aviso.** El paso `pose` que apoya la mano se demora `care · HoverExtra` ticks además de su `wait`. Una giantess cuidadosa da más tiempo para huir.
- **Atención.** Antes de apoyar, hace una tirada de atención con el RNG `ai.attention` por los tinies que nota con `watch` dentro de la huella. La probabilidad crece con `care` y con la cantidad de tinies en movimiento. Si los ve, espera hasta `CareWaitMax` a que se aparten, con el símbolo «?». Si se cumple el plazo, apoya igual.
- **Después.** Si nota restos, emite `noticed` y cambia su expresión a sorprendida. Su plan no cambia en este slice. La culpa, la evitación duradera y la relación con la colonia quedan para el sistema de relaciones.

Así se cumple lo que pide la sección 5.1: la personalidad cambia con qué frecuencia la giantess revisa la superficie, no solo el diálogo.

### 14.4. Consecuencias en la colonia

- `colony.share` aprende los restos como cualquier otra cosa notada y publica `member_lost {colony, entity, cause}`.
- La colonia marca una zona de peligro `danger {space, center, radius, until}` alrededor de los restos. `forage`, `explore` y `ai.WanderPoint` evitan los puntos dentro de una zona activa, y las ofertas que quedan dentro pierden utilidad. La zona expira, para que la colonia vuelva a recolectar sobre el escritorio.
- La colonia no reemplaza miembros en este slice. Si mueren todos, el nido queda vacío con su depósito.

### 14.5. Qué sistemas hay que adaptar

| Paquete | Cambio |
| --- | --- |
| `sim/crush` (nuevo) | Componentes `crusher`, `fragile`, `shelter`, `remains` y `stride`. Sistemas `crush.steps` (Move, después del `mover`), `crush.threats` (Topology, después de `attach.contacts`) y `crush.resolve` (Consequences, antes de `space.falls`). Eventos `crushed`, `sheltered`, `pressed`, `near_miss` y `footstep`. Invariantes: el apoyo en curso está ordenado y los restos tienen nombre y causa |
| `sim/space` | Sin cambios: `crush.resolve` lee los `entity_fell` del tick anterior |
| `sim/attach` | Sin cambios: el apoyo de la palma se detecta por su altura, no por el contacto (cuya tolerancia es de 3 cm) |
| `sim/nav` | `nav.NearestFree(space, from, avoid)` para elegir la celda de huida, sin cruzar portales |
| `sim/ai` | Canales de sombra y vibración en `ai.look`. Reevaluación inmediata con `reflex`. Acción `flee`. Rasgo `care`, demora del paso `pose` y tirada de atención |
| `sim/colony` | `member_lost`, zonas `danger` y filtro en `forage` y `WanderPoint` |
| `sim/jobs` | Sin cambios de contrato. Un test nuevo cubre la muerte del asignado durante un acarreo |
| `engine` | Instalar `crush` después de `attach` y antes de `ai`. Incrementar `world.HashVersion` y `world.SnapshotSchema` |
| `content` | `crusher` en palma, pies y taza. `fragile` en los tinies. `shelter` en el nido |
| `view` | El `Frame` suma zonas de amenaza, pisadas, restos y los estados de huida y atrapado. El seguimiento de un tiny que muere pasa a sus restos |
| `render` y `ui` | Sprites de tiny aplastado y restos, sombra de pisada y entrada de historia con causa y lugar |
| `sound` | «Plaf» de la palma y «tum» de la pisada en micro, golpe sordo en macro y blip de susto, sintetizados desde `tools/audio/sounds.json` |

La mañana en la casa no cambia. Sus entidades no llevan `fragile`, así que `TestUnaMananaEnLaCasa` y sus capturas se mantienen; solo se regeneran los hashes por el incremento de versión. El aplastamiento se activa por contenido.

### 14.6. Presentación

El tono sigue siendo el del anime cotidiano: la muerte se cuenta sin gore y se entiende sin texto.

- **Aviso.** La sombra de la mano ya existe. La pisada suma una sombra ovalada que crece sobre el tatami. Los tinies en zona de amenaza muestran la gota de sudor.
- **Impacto.** Un cuadro con el tiny aplanado como en una caricatura y un pequeño espíritu que sube, al estilo manga, durante un tiempo fijo de presentación.
- **Restos.** Una silueta gris y plana con la bufanda menta del tiny, legible en micro. En macro, una marca sobre el nido indica cuántos miembros perdió la colonia.
- **Historia.** «Tiny 07 quedó bajo la palma en el escritorio», con la opción de localizar los restos. Las huidas exitosas también entran en la historia, para que se vea quién se salvó y no solo quién murió.

Cada señal sale de un evento o del `Frame` y no modifica el mundo, igual que los símbolos y el sonido.

### 14.7. Alcance y hitos

| Hito | Contenido | Cierre |
| --- | --- | --- |
| M23 Aplastamiento | `crusher`, `fragile`, `shelter` y `remains`. Palma y objetos que caen. Muerte con restos y propagación a trabajos y reservas | Escenarios `palm_crushes_tiny`, `tiny_on_palm_is_safe`, `tiny_between_fingers_survives`, `nest_shelters_from_palm`, `cup_lands_on_tiny` y `haul_requeued_after_death` |
| M24 Pisadas | `stride` y pisadas discretas | `step_crushes_tiny_on_floor` y `step_spares_tiny_on_desk` |
| M25 Amenaza y huida | Zonas de amenaza, sombra y vibración, `reflex`, `flee`, `near_miss` y `trapped` | `tiny_flees_shadow` (se salva desde el borde y muere en el centro), `tiny_drops_crumb_to_flee` y `tiny_flees_footsteps` |
| M26 Cuidado y consecuencias | `care`, demora y atención de la giantess. `member_lost` y zonas `danger` | `careless_giantess_crushes` y `careful_giantess_waits` (con `care` alto nadie muere y con 0 sí), `colony_avoids_danger` (con el vencimiento de la zona) |
| M27 Presentación | Sprites, sombras, historia, sonidos, capturas de referencia y escenario de cierre. La casa suma pisadas, el portal por la pata del escritorio al tatami, como en `floor_to_desk_by_leg`, y migajas que caen al piso | Capturas en `docs/capturas/` y mezcla de referencia actualizada |

**Escenario de cierre: Un descuido en el escritorio.** Parte de la mañana y le suma `fragile` a los tinies, la pata del escritorio hacia el tatami y migajas que caen al piso. Se juega con dos variantes de la misma giantess: distraída (`care` 0) y cuidadosa (`care` 0,9). La aceptación pide, para las semillas de referencia:

- con la distraída, al menos un tiny aplastado por la palma o por una pisada, al menos una huida exitosa con `near_miss` y la colonia evitando la zona hasta que expira;
- con la cuidadosa, menos muertes que con la distraída en el total de semillas y al menos una espera;
- en ambas, invariantes en cada tick, ninguna reserva ni trabajo a nombre de un tiny muerto, y el mismo log y hash al continuar desde snapshots;
- un tiny que viaja en la palma nunca muere por el apoyo de esa misma palma.

**Fuera de alcance.** Heridas y daño parcial, caídas letales, sentarse o acostarse sobre tinies, pasos con volumen barrido continuo, objetos que ruedan o se arrastran, alarmas entre tinies, duelo, nacimientos o reemplazo de miembros, relaciones entre la giantess y la colonia y una variante de aturdimiento sin muerte. Esa variante es la candidata natural para un hito posterior, porque reutiliza la suspensión de las caídas.
