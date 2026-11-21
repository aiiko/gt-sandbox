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

El juego usa **pixel art top-down ortográfico de estética anime**. El suelo se proyecta en planta, con el norte arriba en ambas vistas. Los sprites muestran caras y frentes en tres cuartos mediante el dibujo. Las posiciones, alturas y contactos provienen de la simulación.

| Vista | Qué se lee | Densidad inicial de arte | Escena de referencia |
| --- | --- | --- | --- |
| Macro | Habitaciones, puertas, muebles, rutina y desplazamiento de la giantess | 16 px por 40 cm. La referencia de la giantess de pie usa un lienzo de 31 × 38 px | La casa completa y el viaje del dormitorio al living |
| Micro | Tinies individuales, migajas, bordes y pasos entre superficies | 16 px por 10 mm. El tiny tiene una silueta visible de 12–16 px de alto | El nido y las migajas sobre el escritorio, el borde de la palma y el descenso a la mesa del living |

La escena se dibuja en un lienzo lógico de **480 × 270 px**. La ventana de escritorio admite ese lienzo a partir de 1×. Se usa el mayor factor entero que entre, con muestreo por vecino más cercano (nearest-neighbor) y bandas centradas cuando sobra espacio. La posición se interpola primero y se alinea a la cuadrícula de píxeles al dibujar. Las densidades son parámetros para la primera prueba de arte y se ajustan con capturas de la casa y la palma. La cuadrícula visual es independiente de las celdas de navegación.

Macro muestra a los tinies mediante un marcador del nido y un indicador para el tiny seleccionado. Micro muestra el sprite de cada tiny, su carga y suficiente textura para distinguir madera, cerámica, papel y piel. De la giantess se dibuja la parte relevante, como la palma y la manga que entran por un borde del encuadre. La taza conserva su color, el lado del asa y su posición de apoyo en las dos representaciones.

Entrar en micro centra la superficie elegida o al actor en seguimiento. Al volver a macro se recupera el encuadre anterior, salvo que haya una entidad en seguimiento. En ese caso se conserva el seguimiento. Un corte con fundido breve de unos 180 ms confirma el cambio de escala en el rótulo de ubicación. La opción de movimiento reducido usa un corte directo. Mientras un tiny viaja sobre la palma se lo sigue en el mismo modo micro. Si baja a otra superficie, cambia el contexto de ubicación y el encuadre continúa siguiendo al actor.

### Personajes y lenguaje visual

La dirección toma del anime cotidiano las expresiones claras, el vestuario reconocible y la luz de mañana sobre interiores habitados. En el estudio, el cabello violeta, el cárdigan lila y el acento coral identifican a la giantess. Al acercarse a la palma, la misma manga, su puño claro y el botón coral mantienen la continuidad entre escalas.

Los personajes usan proporciones chibi para dar espacio al rostro y al peinado. En macro, la giantess se lee por el contorno del cabello y la separación entre cabeza, ropa y pies. En micro, el tiny se distingue por sus mechones azules, su bufanda menta y unos pocos píxeles de expresión. Cada postura conserva el apoyo de los pies o la mano sobre la superficie.

El sombreado es plano, con bordes definidos como en la animación con celdas. Pelo, piel y ropa se dibujan con un color base, una sombra y uno o dos brillos agrupados. El contorno usa ciruela en cabello y ropa y malva cálido en piel. Los bordes que reciben luz pueden usar un tono del propio material. La madera reserva sus vetas más finas para micro y la cerámica concentra sus reflejos en el borde y el asa.

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

Las superficies tienen tres valores principales de luz y sombra. Los tinies se distinguen por silueta y contraste de valor sobre madera y piel. Un paso abierto se marca con dos extremos unidos y un paso cerrado pierde esa conexión. El icono y su forma permiten distinguir ambos estados con la imagen en escala de grises.

Cada pieza de arte declara un pivote de apoyo y, cuando corresponda, un anclaje de mano o carga. La presentación dibuja suelo y sombras, superficies y frentes, y ocupantes por sus puntos de apoyo. Los frentes de primer plano pueden tapar parte de los sprites. La altura y el parentesco separan superficies superpuestas, y el orden dentro de cada superficie se resuelve con la posición de apoyo en pantalla y un desempate estable. Si un borde oculta a la entidad seleccionada, su contorno o marcador sigue visible. El punto de selección coincide con la entidad que se ve bajo el cursor.

### Personajes, poses y animación

| Elemento | Animación mínima para la escena |
| --- | --- |
| Giantess | Reposo y caminata en cuatro direcciones, mano apoyada y recogida, comer sentada y mirar TV. Escribir, dormir y tomar un snack pueden reutilizar poses coherentes cuando el motor muestra esas acciones |
| Palma | Palma elevada y apoyada, con sombras y continuidad de manga. El dibujo se ancla al Space de la mano |
| Tiny | Reposo, caminata en cuatro direcciones y acarreo. La carga se mueve con el tiny y su silueta sigue visible |
| Objetos | Taza en mano y sobre la mesa, migajas sueltas y depósito del nido. Los cambios de estado se reflejan en la vista correspondiente |

Se parte de ciclos de caminata de cuatro cuadros y una cadencia visual de 6–10 cuadros/s. Las direcciones se eligen entre variantes dibujadas y el pivote queda estable entre frames. El inicio de una pose o animación sale del estado o los eventos de la simulación. El paso entre mesa y mano solo se muestra transitable mientras el portal correspondiente está activo. Una animación puede suavizar el dibujo del gesto, con el apoyo, la posición de los pasajeros y el indicador de paso alineados al estado del mundo.

### Interfaz y alcance de arte

La interfaz de partida muestra hora, pausa y velocidad, ubicación, selección y seguimiento. El panel de historia destaca pocos cambios relevantes de la rutina y permite localizar al actor o la superficie cuando siguen disponibles. En micro, el rótulo identifica el contexto con una jerarquía como «Dormitorio › Escritorio» o «Giantess › Palma». La intervención del jugador se limita a los controles existentes de cámara, tiempo, inspección y órdenes de destino donde la simulación las acepta. Los overlays técnicos y el inspector de componentes quedan disponibles en el visor de debug.

El primer lote de arte cubre suelo, paredes, puertas y mobiliario del dormitorio, cocina y living. El lote micro cubre escritorio, palma, mesa del living, nido, tinies, taza y migajas. Las exportaciones que consume el juego son PNG transparentes y metadatos JSON con identificador visual, escala, rectángulos del atlas, pivotes, capas y clips. La asignación de arte usa identificadores visuales estables independientes del nombre mostrado. La paleta y los tamaños se calibran con la primera escena antes de extender el lote.

La [muestra de presentación](specs/presentacion.md#prueba-de-arte) reúne las referencias macro y micro de esta dirección. El [plan de hitos](plan-motor.md) lleva esas reglas al recorrido completo.

## \*\*12. Riesgos que resolvería temprano\*\*

| Riesgo                          | Consecuencia                                                        | Estrategia                                                |
| ------------------------------- | ------------------------------------------------------------------- | --------------------------------------------------------- |
| Escalas incompatibles           | Personajes atravesando objetos o apareciendo en lugares incorrectos | Unidades físicas canónicas y transformaciones jerárquicas |
| Navegación dinámica             | Tinies atrapados o teletransportados                                | Portales, grafo de superficies y rutas invalidables       |
| Explosión de complejidad        | Proyecto interminable                                               | Vertical slices pequeños                                  |
| IA errática                     | NPC que cambian de opinión constantemente                           | Utility AI con persistencia de objetivos                  |
| Simulación costosa              | Caídas de FPS con muchos agentes                                    | Scheduling, LOD y profiling                               |
| Arte de escala extrema          | Necesidad de sprites enormes                                        | Representaciones visuales específicas por escala          |
| Bugs emergentes                 | Situaciones difíciles de reproducir                                 | Replay determinista y escenarios de prueba                |
| Demasiadas interacciones únicas | Alto costo de mantenimiento                                         | Smart Objects y acciones reutilizables                    |

## 13. Alcance posterior

El foco inmediato es presentar una casa que ya funciona como ecosistema vivo. La giantess puede seguir con su rutina aunque los tinies la observen o interactúen con su entorno. El jugador puede pasar de una escala a la otra y reconocer cómo una acción cotidiana cambia las posibilidades de la colonia.

La expansión posterior puede sumar más superficies explorables, construcción tiny, relaciones, personalidad y eventos ambientales. Fluidos, simulación anatómica continua y economías más amplias quedan para etapas posteriores.
