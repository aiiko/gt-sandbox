# Proyecto: Giantess Life Simulator

Análisis de diseño, simulación y arquitectura técnica

Tu idea tiene una combinación bastante particular: tomar las situaciones y diferencias de tamaño de los juegos Giantess clásicos, pero reemplazar su estructura de RPG por un **sandbox de simulación emergente**, donde los personajes tienen necesidades, relaciones, rutinas y comportamientos autónomos.

La diferencia fundamental sería esta:

**En un RPG tradicional, la giantess hace algo porque un evento programado lo indica. En tu juego, lo hace porque la simulación la llevó a tomar esa decisión.**

Por ejemplo, imaginemos una casa habitada por una giantess y una pequeña civilización de tinies que vive debajo de su escritorio.

La giantess se despierta, tiene hambre y decide preparar el desayuno. Se levanta de la cama, camina hasta la cocina y mueve algunos objetos. Mientras tanto, los tinies continúan con su propia rutina: recolectan comida, construyen refugios, exploran y reaccionan a los cambios en el entorno.

Un tiny podría subirse accidentalmente a su zapatilla. Cuando ella empieza a caminar, el tiny pasa a desplazarse junto con ella. Otro podría estar explorando su cabello, mientras un grupo aprovecha que salió de la habitación para recuperar recursos.

Ninguna de esas situaciones necesitaría una misión escrita a mano.

El juego produciría historias mediante la interacción de sus sistemas.

## La combinación que propongo

- **RimWorld** — Colonias, tareas, supervivencia y consecuencias
- **The Sims** — Necesidades, autonomía, personalidad y relaciones
- **Dwarf Fortress** — Mundo persistente, causalidad y profundidad de simulación
- **Giantess RPGs** — Perspectiva, exploración, interacciones entre escalas y situaciones especiales

Mi propuesta es diseñarlo como **un simulador de ecosistemas a distintas escalas**, no como un RPG con más NPC ni como un clon completo de RimWorld.

Esto importa porque las mecánicas más interesantes de tu juego aparecen precisamente donde se cruzan ambos mundos: la misma acción cotidiana puede ser trivial para una giantess y un acontecimiento enorme para los tinies.

## 1. Qué podemos extraer de los juegos de referencia

### 1.1. Los juegos Giantess: conservar la temática, cambiar la estructura

#### Size Matters — Uru

Sus mapas de tamaños diferentes, zonas como ciudades diminutas y situaciones de interacción entre personajes de distintas escalas son referencias útiles. La documentación comunitaria de sus versiones muestra una gran cantidad de áreas y eventos particulares. [aryion.com](https://aryion.com) · [Imagen de Size Matters](https://images.openai.com/static-rsc-4/pCrNRjhE4TrfVIklpLLSluxOmSgwlL4g3fxqojxT8R4jRu89ZNe6Zb9heTB5_2D_CQiQ7uwWGF1XY4i_t4ZwHNthuQVzq4XjYV-iO5o3Ax1UTsjM_82f60MM9pZFvXTfz6TGYEyZzUIm6Ae5J7xV4cl0x-pOGX8iu4UJqty1BqA?purpose=inline)

**Adoptaría** la exploración ambiental y la variedad de situaciones que surgen de las diferencias de tamaño. Cambiaría la dependencia de mapas y eventos independientes para representar cada interacción.

#### Shrink High / 縮小学園

Destaca por la vulnerabilidad de los personajes diminutos, la exploración y el peligro asociado a encuentros con personajes enormes. Un equipo que desarrolló *Astra's Plaything* en 2026 también identificó esa sensación de desventaja como uno de los grandes atractivos del género. [furaffinity.net](https://www.furaffinity.net) · [Shrink High Download Game - excellentlasopa](https://images.openai.com/static-rsc-4/rPA2KOeqozjjHMyVdXmymVfc9OPY7YlAq1vwDfqNvyZ-6VHIvti60wpDovm8CzX059eOEykcXGo1q8j9PY2R6Houf_xRSFaiK9FwcilgB9BtuqCTh7IvAB9WrAyqdm7YGeqYSXdUhYhTe3mzeHbwpA7BnuqsvLYhJ5GOw6enGsU?purpose=inline)

**Adoptaría** que el tamaño afecte realmente la exploración, el desplazamiento, la percepción y la supervivencia.

#### Terada Ochiko — obras y juegos asociados

Las referencias relacionadas con Ochiko aportan situaciones de escala extrema y exploración de espacios asociados a personajes gigantes. Hay juegos comunitarios atribuidos a su trabajo, como *President Matsudo's Ambition*. También figura como artista colaborador en *Shrink High Gaiden*, no como su único desarrollador. [aryion.com](https://aryion.com) · [PULSE - ゲームを記録・管理できる究極のゲームライブラリ](https://images.openai.com/static-rsc-4/WB2SyeNoBmdBQrNN8_-K1ab0rZol2HlSyYqlOSNOzKZ-oTwem5oiD8T6CC4iOAWTgHdNOugBCeHsny4tuJWl2HKCXwyCIsnlrhT4iQy3emPwsf1HNY2DlK0AN0RmVs2yRCIobHUh8y-cfDW7hsg1Zg7YzwuH54ghzuDSlKSf3UM?purpose=inline)

**Adoptaría** el cuerpo de una giantess como entorno explorable, con sus propios obstáculos, lugares y eventos, pero convertido en un sistema reutilizable.

#### SAEKO: Giantess Dating Sim — referencia adicional

Es especialmente relevante porque utiliza dos estilos de pixel art y distintos sistemas de juego para representar la vida de humanos diminutos bajo el control de una gigante. Sin embargo, su experiencia sigue estando organizada alrededor de una historia y decisiones narrativas. [SAEKO: Giantess Dating Sim](https://saekogame.com) · [小人となって死を潜り抜ける会話型アドベンチャー『SAEKO: Giantess Dating Sim』。家主の巨大少女の機嫌を損ねると…？【電撃インディー#670】 - 電撃オンライン](https://images.openai.com/static-rsc-4/I4lkcyrEwXi9GeZB7QzGR_f3mOBEanIBklVj2bZf3aArRcafciPk24h2yJDHB-yJ9i46tZrXV2lx06f4lhJ8V12DJ201rADtewoE0OQc00bWDfCShOGgmkcSRx0DwRO2u4n9brao9lrNGFqdfUOowjc5JsabOIUQXp039F4itsc?purpose=inline)

**Adoptaría** distintas representaciones visuales del mismo universo según la perspectiva y la distancia.

La oportunidad está en convertir lo que estos juegos presentan mediante escenas diseñadas individualmente en **reglas generales que producen muchas situaciones distintas**.

Por ejemplo:

| Situación                    | RPG tradicional          | Tu simulador                                          |
| ---------------------------- | ------------------------ | ----------------------------------------------------- |
| Una giantess camina          | Secuencia predeterminada | Acción autónoma con consecuencias espaciales          |
| Un tiny queda atrapado       | Evento de una misión     | Resultado de movimiento, obstáculos y entorno         |
| Un tiny sube a una persona   | Mapa especial            | Transición a una superficie móvil explorable          |
| Una giantess descubre tinies | Diálogo predefinido      | Reacción según percepción, personalidad y relación    |
| Una casa cambia              | Evento argumental        | Consecuencia de actividades e interacción con objetos |

### 1.2. Los simuladores: qué sistema tomar de cada uno

| Juego               | Sistema más valioso                                         | Cómo se adapta                                                |
| ------------------- | ----------------------------------------------------------- | ------------------------------------------------------------- |
| RimWorld            | Jobs, prioridades, colonias y emergencias                   | Los tinies trabajan, construyen y se organizan                |
| The Sims            | Necesidades, autonomía, personalidad y objetos interactivos | Las giantesses tienen una vida independiente                  |
| Dwarf Fortress      | Persistencia, causalidad, materiales, historia emergente    | Los acontecimientos dejan consecuencias duraderas             |
| Oxygen Not Included | Recursos, condiciones ambientales y redes                   | Las colonias tinies aprovechan agua, comida, calor y refugios |
| Kenshi              | Mundo autónomo y libertad de roles                          | Los personajes existen independientemente del jugador         |

Hay dos referencias técnicas especialmente interesantes.

**RimWorld:** su arquitectura distingue la selección de trabajos de su ejecución y distribuye evaluaciones de IA entre distintos ticks. Esa separación evita tener que reconsiderar todas las decisiones de todos los personajes en cada frame. [GitHub](https://github.com)

**The Sims:** Robert Zubek, que trabajó en ese ámbito, describió un sistema de inteligencia artificial basado en necesidades y objetos que ofrecen comportamientos. En vez de programar cada actividad completa dentro del personaje, parte de la información sobre qué puede hacerse está en los propios objetos del mundo. [robert.zubek.net](https://robert.zubek.net)

Esto último encaja particularmente bien en una casa llena de objetos.

## 2. Qué tipo de juego debería ser

Veo tres direcciones posibles.

| Enfoque                   | Experiencia                                                     | Complejidad |
| ------------------------- | --------------------------------------------------------------- | ----------- |
| Colony Sim                | Gestionás una sociedad de tinies dentro de la casa              | Alta        |
| Life Sim                  | Controlás o influís en una giantess que vive su rutina          | Media-alta  |
| Sandbox emergente híbrido | Alternás perspectivas e intervenís sobre un ecosistema autónomo | Muy alta    |

**Elegiría el tercero como objetivo final, pero desarrollaría primero el núcleo de un Colony Sim.**

Esto no significa descartar jugar como giantess. Significa construir primero una simulación capaz de funcionar aunque no controles directamente a nadie.

La estructura podría tener tres perspectivas:

- **Vista de la casa.** Observás las habitaciones, rutinas y desplazamientos de las giantesses. Similar a The Sims.
- **Vista de los tinies.** Gestionás colonias, exploración, recursos, peligros y trabajos. Similar a RimWorld.
- **Vista del cuerpo.** Explorás superficies asociadas a una giantess, como manos, brazos, ropa o cabello, mientras ella sigue moviéndose y actuando.

La tercera perspectiva es lo que más diferenciaría tu proyecto.

Pero también introduce el problema técnico más importante de todo el diseño: **un personaje que simultáneamente es un agente autónomo, un objeto físico y un mundo que otros personajes pueden explorar**.

## 3. El sistema central: un mundo con varias escalas

No construiría tres mundos aislados. Construiría un único estado de simulación con múltiples espacios conectados.

### 3.1. La casa es una jerarquía de espacios

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

## 4. La giantess como mapa móvil

Este es el sistema que diseñaría primero en un prototipo técnico, incluso antes de construir una IA compleja.

La idea es asignarle a cada giantess un conjunto de espacios explorables, cada uno asociado a una parte de su cuerpo o a un objeto que lleva puesto.

**Ejemplo de superficies explorables**

| Parte | Superficies |
| --- | --- |
| Cabeza | Cabello, rostro |
| Brazos | Antebrazos, codos |
| Manos | Palmas, dedos |
| Torso | Ropa, hombros |
| Piernas | Rodillas, ropa |
| Pies | Calzado, superficies |

Cada superficie tendría su propio mapa, puntos de conexión y reglas físicas.

Por ejemplo:

```text
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
```

Un tiny podría desplazarse entre estas regiones, siempre que sean accesibles según la pose actual.

### 4.1. Cómo sería el movimiento

Imaginemos que un tiny está sobre el antebrazo.

La giantess levanta la mano para tomar una taza. Técnicamente ocurren varios cambios:

1. Su IA decide ejecutar la acción `TakeCup`.
2. El sistema de acciones cambia la pose del brazo.
3. El espacio asociado al antebrazo actualiza su transformación.
4. El tiny conserva su posición local respecto al antebrazo, si sigue sujeto a él.
5. Las nuevas condiciones del movimiento pueden provocar deslizamiento, pérdida de equilibrio o la necesidad de agarrarse.
6. Si cae, pasa del espacio corporal a otra superficie del mundo, resolviendo dónde aterriza.

La posición del tiny sería relativa a su superficie:

$$
P_{\text{mundo}} = T_{\text{giantess}} \cdot T_{\text{parte}} \cdot P_{\text{local}}
$$

En este modelo, las transformaciones convierten la posición local del tiny a una posición global. En una simulación top-down con altura discreta, se puede implementar con transformaciones 2D y un componente adicional de elevación.

El detalle importante es que **la giantess puede continuar ejecutando su rutina mientras los tinies exploran su cuerpo**.

### 4.2. El problema de las poses

No alcanza con dibujar el cuerpo como un mapa fijo.

Una giantess parada, sentada, acostada o caminando tiene distintas superficies accesibles.

| Pose                  | Efecto sobre los tinies                                                    |
| --------------------- | -------------------------------------------------------------------------- |
| Parada                | Acceso principalmente por zapatos, ropa cercana al suelo y objetos         |
| Sentada               | Nuevas rutas mediante silla, piernas o manos apoyadas                      |
| Acostada              | Más superficies corporales quedan cercanas y accesibles                    |
| Caminando             | Los espacios corporales se desplazan y algunas rutas se vuelven inestables |
| Sosteniendo un objeto | El objeto puede convertirse en una ruta de transporte                      |

No empezaría con un esqueleto físico articulado completo. Para pixel art top-down, sería mucho más económico definir **poses discretas, puntos de anclaje y transiciones controladas**.

Eso ofrece buena parte de la experiencia sin implementar una simulación anatómica continua.

### 4.3. Dos sistemas de navegación conectados

El recorrido de un tiny podría ser:

```text
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
```

El motor primero calcularía una ruta general por los espacios conectados y después rutas locales dentro de cada mapa.

Se parece al principio de pathfinding jerárquico `HPA*`, que reduce el problema de navegación mediante clusters y conexiones de nivel superior. [theory.stanford.edu](https://theory.stanford.edu)

La parte nueva sería que algunas conexiones aparecen, desaparecen o cambian de posición según lo que hace la giantess.

Por ejemplo, cuando coloca su mano sobre el escritorio, se habilita temporalmente un camino entre ambos espacios.

**Esto es muchísimo más flexible que teletransportar al tiny a un mapa especial mediante un evento.**

## 5. Sistemas de simulación: cuáles combinaría y cómo

La clave es que los sistemas puedan interactuar entre sí sin que cada uno necesite conocer todos los detalles de los demás.

Mi selección sería la siguiente:

| Sistema                          | Inspiración               | Prioridad         | Dificultad |
| -------------------------------- | ------------------------- | ----------------- | ---------- |
| Mundo multi-escala               | Diseño propio             | Crítica           | Muy alta   |
| Necesidades y rutinas            | The Sims                  | Crítica           | Media      |
| Trabajos y prioridades           | RimWorld                  | Alta              | Media      |
| Navegación entre superficies     | RimWorld / `HPA*`         | Crítica           | Muy alta   |
| Percepción y conocimiento        | The Sims / IA de agentes  | Alta              | Alta       |
| Relaciones y personalidad        | The Sims                  | Media             | Media      |
| Interacciones físicas            | Simulaciones emergentes   | Alta              | Alta       |
| Construcción tiny                | RimWorld / Dwarf Fortress | Media             | Alta       |
| Objetos interactivos             | The Sims                  | Crítica           | Media      |
| Historia emergente               | Dwarf Fortress            | Alta              | Media      |
| Simulación de fluidos y ambiente | Oxygen Not Included       | Baja inicialmente | Muy alta   |

### 5.1. IA de las giantesses: Utility AI + acciones

En lugar de escribir una secuencia rígida como:

`Despertar → Ir al baño → Desayunar → Mirar televisión`

utilizaría una IA basada en necesidades y puntuaciones de utilidad (*Utility AI*).

Por ejemplo, una giantess podría tener estos valores:

**Estado de una giantess**

| Necesidad | Ejemplo |
| --- | --- |
| Hambre | 25% |
| Energía | 75% |
| Higiene | 50% |
| Diversión | 40% |
| Social | 65% |

La IA evalúa cuáles de sus acciones disponibles mejoran estas necesidades y considera distancia, disponibilidad, personalidad y riesgos.

Una versión simple de la puntuación sería:

$$
U(a) = \sum_i w_i \cdot \Delta N_i(a) - C(a) - R(a)
$$

- $U(a)$ es la utilidad de la acción.
- $w_i$ representa la importancia de cada necesidad para el personaje.
- $\Delta N_i(a)$ es el beneficio esperado para esa necesidad.
- $C(a)$ representa el costo de tiempo, esfuerzo y desplazamiento.
- $R(a)$ es la penalización por riesgo.

La fórmula es una propuesta simplificada; en un sistema más avanzado utilizaríamos curvas de utilidad no lineales y un mecanismo que impida cambiar de decisión continuamente. La literatura sobre *Needs-Based AI* describe precisamente estas estrategias de selección y arbitraje. [robert.zubek.net](https://robert.zubek.net)

Pero agregaría otra dimensión fundamental: **la personalidad**.

Dos giantesses podrían tener exactamente la misma hambre y, aun así, elegir actividades distintas.

Una puede ser ordenada, cuidadosa y observadora. Otra, distraída, impulsiva y despreocupada. Esas diferencias no deberían limitarse al diálogo: modificarían la frecuencia con que revisan el suelo, mueven objetos, investigan ruidos o reaccionan ante los tinies.

### 5.2. Smart Objects: los objetos ofrecen comportamientos

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

### 5.3. IA de tinies: Jobs + Utility AI

Los tinies necesitan un sistema adicional: trabajo colectivo.

Un tiny hambriento puede querer buscar comida. Pero también podría tener asignada la construcción de un refugio.

Propongo esta jerarquía de decisiones:

```text
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
```

Cada trabajo podría tener una secuencia de pasos, similar al enfoque de jobs y ejecutores de RimWorld.

Por ejemplo, un trabajo `CollectCrumb`:

`Buscar recurso → Reservarlo → Calcular ruta → Llegar → Recoger → Transportar → Depositar`

La reserva es importante: evita que diez tinies decidan recoger exactamente la misma migaja.

También permite interrumpir trabajos cuando ocurre una emergencia y retomarlos después.

### 5.4. Percepción: los tinies no deberían ser omniscientes

Este es un sistema fácil de subestimar y uno de los que más mejoraría la simulación.

Cada agente necesitaría distinguir tres cosas:

- Qué existe realmente en el mundo.
- Qué puede percibir en este momento.
- Qué cree o recuerda acerca del mundo.

Por ejemplo, una giantess podría caminar sobre una alfombra donde viven varios tinies, pero no necesariamente saber que están allí.

Los tinies, en cambio, podrían percibir sus pasos mediante vibraciones antes de verla.

Mi diseño incluiría canales de percepción como visión, sonido, vibración y contacto. Su sensibilidad dependería de la escala, la distancia, el material y las capacidades del personaje.

Así pueden aparecer situaciones emergentes como una giantess que encuentra rastros de actividad diminuta, investiga y descubre gradualmente la existencia de una colonia.

### 5.5. Física: simular consecuencias, no átomos

No intentaría reproducir físicamente todas las interacciones de una giantess con millones de partículas.

Usaría un enfoque de *simulación basada en eventos físicos*.

Por ejemplo, cuando la giantess apoya una taza sobre la mesa, el motor produce un evento de contacto que contiene la posición, superficie y características físicas relevantes.

El sistema de consecuencias puede entonces evaluar qué ocurre con los objetos y agentes próximos.

Un vaso que se derrama podría producir un volumen de líquido simplificado que se propaga sobre unas cuantas celdas. Los tinies cercanos reaccionarían según sus capacidades y ubicación.

Para movimientos grandes, como caminar, usaría volúmenes de colisión y trayectorias barridas, no una búsqueda exhaustiva sobre todas las celdas diminutas.

Esto permite tener consecuencias espaciales incluso cuando la cámara está observando otra habitación.

## 6. El gran desafío: que la simulación siga funcionando fuera de cámara

Una giantess puede estar en la cocina mientras el jugador observa una colonia debajo de la cama.

**El mundo no debe detenerse ni depender de lo que está dibujado en pantalla.**

Pero simular cada personaje con el máximo detalle todo el tiempo sería innecesariamente costoso.

Por eso recomiendo *Simulation Level of Detail* (LOD de simulación).

| Nivel        | Qué se simula                                 | Ejemplo                               |
| ------------ | --------------------------------------------- | ------------------------------------- |
| Detallado    | Movimiento, contactos, acciones y navegación  | Tinies cerca de la cámara             |
| Intermedio   | Trabajos, rutas resumidas, necesidades        | Tinies activos en otra zona           |
| Agregado     | Recursos, actividad y progreso por intervalos | Colonia distante sin eventos críticos |
| Event-driven | Actualizaciones cuando sucede algo relevante  | Objetos inactivos y procesos lentos   |

La decisión de qué nivel usar **no debería depender solamente de la cámara**.

Una zona con un acontecimiento importante necesita mantenerse suficientemente detallada aunque el jugador no la esté mirando. Lo mismo sucede si una giantess está trasladando tinies entre habitaciones.

Para evitar contradicciones, mantendría siempre las colisiones y transferencias críticas bajo un sistema lógico autoritativo.

La filosofía general es compatible con una recomendación del propio Tarn Adams: construir modelos comprensibles, evitar complejidad innecesaria y apoyarse en analogías del mundo real. [gameaipro.com](https://www.gameaipro.com)

## 7. Enfoque técnico: qué tecnología utilizaría

Tomando Go como lenguaje principal, que además conocés por tu trabajo, elegiría una arquitectura donde toda la simulación pueda ejecutarse independientemente del motor visual.

### 7.1. Comparación de motores y frameworks

Evaluaciones relativas para este proyecto, no benchmarks medidos.

| Stack                  | Simulación compleja | Pixel art | Desarrollo con IA             | Riesgo técnico |
| ---------------------- | ------------------- | --------- | ----------------------------- | -------------- |
| **Go + Ebitengine**    | Excelente           | Excelente | Excelente para lógica y tests | Bajo-medio     |
| Go + raylib-go         | Excelente           | Muy bueno | Muy bueno                     | Medio          |
| Godot + GDScript/C#    | Muy bueno           | Excelente | Muy bueno                     | Medio          |
| Godot + Go/GDExtension | Muy bueno           | Excelente | Medio                         | Alto           |

**Mi elección: Go + Ebitengine.**

#### Ebitengine

Está especializado en 2D y ofrece rendering acelerado por GPU, batching automático cuando es posible y una API bastante reducida. Desde la versión 2.10, publicada en septiembre de 2026, soporta builds de escritorio en Go puro sin requerir compilador C. [ebitengine.org](https://ebitengine.org) · [Ebitengine Tutorial: Boids Flocking in Golang - BioErrorLog Tech Blog](https://images.openai.com/static-rsc-4/QLn1O7m54PPnYR5OosaJ7pmCy2Ff8W4VXX6cAR038c5gBiAApjHdMyW86uLjYjIGAN9Y-8gPJL2SI4XdNQQUkYuQaEBhFU9sdNzSJuwWINw1t5ActSMe3czk4ICTUyyc2q7LdLA2Jeq6P9Ok5Y4k9ysvxWJBuEU2_QLDKOVJZms?purpose=inline)

Eso es muy conveniente para la asistencia con IA: menos capas de integración, más código Go convencional y una separación clara entre lógica y presentación.

También hay una novedad reciente especialmente útil para tu workflow.

**Ebitengine 2.10: pruebas visuales automatizadas para agentes**

Ebitengine incorporó un mecanismo experimental de virtualización de aplicaciones. Un programa externo puede controlar los inputs, avanzar ticks y recuperar imágenes renderizadas, permitiendo probar juegos mediante agentes de código.

Los desarrolladores publicaron además un recurso llamado `run-ebitengine-app-headless` destinado a estos workflows. Requiere habilitación explícita y un entorno gráfico en el host, pero permite pruebas sin mostrar la ventana normal del juego. [Ebitengine](https://ebitengine.org)

Para un proyecto desarrollado con Opus o Codex, esto representa una ventaja muy interesante: **el agente puede comprobar visualmente si sus cambios funcionaron, además de ejecutar tests Go**.

Como alternativa, raylib-go también permite desarrollar una simulación independiente y conectar un renderer sencillo. Su principal desventaja en este caso es depender de bindings de raylib y su capa de compilación C. [GitHub](https://github.com)

Godot ofrece un editor visual más completo, mejor tooling incorporado y un desarrollo de UI probablemente más rápido. Sin embargo, usar Go dentro de Godot mediante `godot-go` agrega una integración experimental. Su propio repositorio advierte que sigue en desarrollo activo y que puede tener problemas de memoria. [GitHub](https://github.com)

Si priorizáramos crear rápidamente escenas, animaciones y herramientas visuales por encima de Go puro, Godot sería el competidor más fuerte.

### 7.2. Stack recomendado

| Capa                      | Tecnología                                                   |
| ------------------------- | ------------------------------------------------------------ |
| Lenguaje principal        | Go                                                           |
| Motor 2D                  | Ebitengine 2.10+                                             |
| Simulación                | Paquetes Go puros                                            |
| Entidades                 | Structs y stores propios inicialmente; Ark ECS si hace falta |
| Mapas diseñados a mano    | JSON del proyecto. LDtk queda como posible importador        |
| Sprites y animación       | PNG y metadatos JSON. Aseprite es una opción de autoría       |
| UI                        | Componentes propios sobre Ebitengine                         |
| Navegación                | `A*` local + grafo jerárquico                                 |
| Definiciones de contenido | JSON                                                         |
| Tests                     | `go test`, benchmarks, fuzzing                               |
| Profiling                 | `pprof`                                                      |
| Guardado                  | Snapshots versionados + registro de eventos relevantes       |

El renderer de la próxima etapa cargará las exportaciones de arte y los datos del proyecto directamente. La elección del editor se puede ajustar durante la producción.

#### ¿Usaría ECS?

No empezaría con un ECS complejo.

En la primera versión utilizaría entidades identificadas por IDs estables, structs Go y sistemas independientes. Es más sencillo de depurar y de mantener para un proyecto de aprendizaje.

Si el profiling muestra que las iteraciones sobre miles de entidades son un cuello de botella, incorporaría un ECS basado en arquetipos.

Hoy consideraría especialmente [**Ark ECS**](https://github.com/mlange-42/ark?utm_source=chatgpt.com), sucesor recomendado por los autores de Arche. Proporciona consultas tipadas, relaciones entre entidades, operaciones por lotes y herramientas de serialización. [GitHub](https://github.com)

No hay necesidad de resolver una hipotética carga de 100.000 agentes antes de que existan veinte.

## 8. Arquitectura propuesta

```text
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
```

La separación más importante es esta:

```text
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
```

El renderer únicamente interpreta estados producidos por la simulación.

No debería modificar necesidades, decidir trabajos, resolver contactos ni alterar la posición lógica de los personajes.

Esto permite ejecutar miles de ticks en un test sin abrir una ventana.

### 8.1. Modelo de ubicación

Lo diseñaría alrededor de una estructura similar a esta:

```go
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
```

Para el prototipo, cada `Space` tendría una transformación local hacia su entidad padre o hacia el mundo global. En producción también necesitaríamos IDs versionados, geometrías de portales y reglas de accesibilidad.

La ventaja de este modelo es que un tiny puede encontrarse en `SpaceID=LeftHand` mientras la mano se mueve.

Su posición local no necesita actualizarse con cada desplazamiento global de la giantess. Solamente cambian las transformaciones de la jerarquía.

Ese diseño reduce trabajo y simplifica muchísimo la lógica de transporte.

### 8.2. Simulación con fixed ticks

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

### 8.3. Navegación y colisiones

Usaría tres estructuras complementarias:

1. **Índice espacial global.** Permite encontrar entidades y objetos próximos sin revisar todas las entidades del juego.
2. **Navegación local.** Usa `A*` sobre tiles, regiones o celdas de una superficie.
3. **Grafo de espacios.** Conecta habitaciones, muebles, superficies elevadas y partes del cuerpo de las giantesses.

Una ruta larga primero recorrería el grafo y luego resolvería trayectorias locales.

Cuando una giantess se mueve, el sistema actualiza sus puntos de conexión y solo invalida las rutas afectadas.

El mayor desafío no es ejecutar `A*`, sino **gestionar los cambios topológicos**: una mano deja de tocar una mesa, un zapato es levantado del suelo o un objeto bloquea una entrada.

Por eso los portales, superficies y contactos deben ser conceptos independientes del renderizado.

## 9. Cómo lo desarrollaría con IA

No le pediría a Opus que construyera todo el juego de una sola vez.

Haría que implementara sistemas aislados, con interfaces y criterios de aceptación verificables. Así evitamos que una modificación del sistema corporal termine rompiendo las necesidades o el guardado.

### Workflow

1. **Especificaciones pequeñas**

   Crear documentos breves para `WorldSpace`, `TinyMovement`, `Jobs`, `BodySpace` y otros sistemas. Cada documento define responsabilidades, contratos, invariantes y casos límite.
2. **Simulación headless primero**

   Antes de dibujar un personaje, probar que pueda cambiar de superficie, recibir una orden y completar un trabajo. El resultado debe poder comprobarse con `go test`.
3. **Escenarios reproducibles**

   Un escenario incluiría un estado inicial, una semilla aleatoria, acciones externas y cantidad de ticks. Después compararíamos resultados e invariantes, no solamente imágenes.
4. **Integración visual y pruebas automáticas**

   Una vez que la simulación funciona, conectar el renderer Ebitengine. Aprovechar su mecanismo de virtualización para verificar visualmente inputs, cámara, animaciones y UI.
5. **Benchmarks y profiling**

   Repetir escenarios con 20, 100, 500 y 1.000 tinies. Medir CPU, memoria, latencia por tick, búsquedas de caminos y allocaciones con las herramientas oficiales de Go. [The Go Programming Language](https://go.dev)

### Tests que consideraría imprescindibles

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

Este recorrido sirve para probar la dirección visual. En macro se debe entender qué hace la giantess y adónde va. En micro se debe poder seguir al mismo tiny desde el escritorio hasta la palma y la mesa del living. La secuencia se tiene que entender por lo que se ve en ambas vistas, con la historia como apoyo.

## 11. Dirección visual

**Anime cotidiano vívido en pixel art.** La casa se ve como un interior de anime en una mañana soleada, con manchas de sol bien definidas, materiales de colores reconocibles y personajes expresivos. Cada cuarto tiene su propia combinación de color. Los tonos más saturados se concentran en ropa, textiles y objetos pequeños sobre superficies claras. La luz de ventana y los detalles de uso cotidiano hacen que el departamento se sienta habitado.

### Cámara y escala

El suelo se proyecta en planta ortográfica, con el norte arriba en ambas vistas. Los sprites muestran caras y frentes en tres cuartos mediante el dibujo. Las posiciones, alturas y contactos provienen de la simulación.

| Vista | Qué se lee | Densidad inicial de arte | Escena de referencia |
| --- | --- | --- | --- |
| Macro | Habitaciones, puertas, muebles, rutina y desplazamiento de la giantess | 16 px por 20 cm (80 px/m, [17.3](#173-más-detalle-en-la-vista-macro)). La giantess de pie usa un lienzo de 62 × 76 px | La casa completa y el viaje del dormitorio al living |
| Micro | Tinies individuales, migajas, bordes y pasos entre superficies | 16 px por 10 mm. El tiny tiene una silueta visible de 12–16 px de alto | El nido y las migajas sobre el escritorio, el borde de la palma y el descenso a la mesa del living |

La escena se dibuja en un lienzo lógico de **480 × 270 px**. La ventana de escritorio admite ese lienzo a partir de 1×. Se usa el mayor factor entero que entre, con muestreo por vecino más cercano (nearest-neighbor) y bandas centradas cuando sobra espacio. La posición se interpola primero y se alinea a la cuadrícula de píxeles al dibujar. Las densidades son parámetros para la primera prueba de arte y se ajustan con capturas de la casa y la palma. La cuadrícula visual es independiente de las celdas de navegación.

> **Revisión prevista.** El lienzo fijo, la densidad macro y el zoom se replantean en la [sección 17](#17-interfaz-y-lectura-en-pc): lienzo que llena la ventana, escena e interfaz en capas con su propio factor, macro a 80 px/m y zoom con la rueda.

Macro muestra a los tinies mediante un marcador del nido y un indicador para el tiny seleccionado. Micro muestra el sprite de cada tiny, su carga y suficiente textura para distinguir madera, cerámica, papel y piel. De la giantess se dibuja la parte relevante, como la palma y la manga que entran por un borde del encuadre. La taza conserva su color, el lado del asa y su posición de apoyo en las dos representaciones.

Entrar en micro centra la superficie elegida o al actor en seguimiento. Al volver a macro se recupera el encuadre anterior, salvo que haya una entidad en seguimiento. En ese caso se conserva el seguimiento. Un corte con fundido breve de unos 180 ms confirma el cambio de escala en el rótulo de ubicación. La opción de movimiento reducido usa un corte directo. Mientras un tiny viaja sobre la palma se lo sigue en el mismo modo micro. Si baja a otra superficie, cambia el contexto de ubicación y el encuadre continúa siguiendo al actor.

### Ambientación

La casa es un departamento japonés contemporáneo. Las telas, la vajilla y los objetos de uso diario dan a cada cuarto una identidad que se reconoce aun en la vista general. El dibujo respeta la posición y la altura que tienen los muebles en el escenario.

| Espacio | Color y luz | Detalles que lo identifican |
| --- | --- | --- |
| Entrada y umbrales | Madera clara y papel blanco cálido, con sombras cortas bajo los marcos | Puertas corredizas, fusuma decorado y shoji según el ambiente |
| Dormitorio | Tatami verde junco con bordes verde oscuro, edredón azul cielo y una franja dorada de sol junto al escritorio | Futón, shoji y un escritorio de madera para estudiar. El escritorio mantiene sus 75 cm porque ahí vive la colonia |
| Cocina | Azulejo celeste, blancos fríos y un paño rojo tomate | Anafe de dos hornallas, arrocera, heladera chica con imanes, noren en el paso y vajilla de cerámica |
| Living | Madera miel, zabuton mandarina, hojas verdes y una alfombra azul petróleo | Planta en maceta, televisor bajo y mesa clara a la altura del escenario |

En micro, la caja de fósforos del nido lleva una etiqueta con trazos de kana, la taza puede ser un yunomi o llevar una flor esmaltada y la mesa puede tener un posavasos de bambú. Los rótulos y la interfaz están en castellano. Un kotatsu queda como variante posterior porque requeriría bajar la superficie en el escenario.

### Personajes y lenguaje visual

La propuesta de color para la giantess parte de cabello castaño oscuro con reflejos miel, cárdigan azul cobalto y botón mandarina. El retrato y el sprite macro comparten la forma del flequillo y los colores de la ropa. Al pasar a la palma, la manga azul, el puño claro y el botón mantienen la continuidad.

Los sprites usan proporciones chibi para dar espacio al rostro y al peinado. En macro, la giantess se lee por el perfil del pelo y la separación entre cabeza, ropa y pies. En el retrato, la mirada, las cejas y la boca llevan el gesto. En micro, cada tiny se reconoce por la silueta del pelo, el color de su bufanda y unos pocos píxeles de expresión. Las bufandas pueden alternar rojo tomate, verde hoja y amarillo cálido, con un borde que se lea tanto sobre madera como sobre piel. Cada postura conserva el apoyo de los pies o la mano sobre la superficie.

El sombreado de celda usa manchas limpias y bordes definidos. Pelo, piel y ropa parten de un color base, un plano de sombra y brillos agrupados. El borde del cabello es castaño profundo, el de la ropa azul oscuro y el de la piel terracota. En el lado iluminado, el contorno puede fundirse con un tono más claro del material. El brillo del ojo se resuelve con un píxel claro, también en el tiny de 16 px.

### Expresiones y señales

Los personajes muestran su estado con símbolos de manga sobre la cabeza: «!» al descubrir algo, «?» al explorar, gota de sudor ante un peligro o al ser levantado, «zzz» al descansar, nota musical al divertirse y vena de enojo cuando se interrumpe una acción. Cada símbolo sale de un evento o del estado de la simulación, dura un tiempo fijo de presentación y no modifica el mundo. En micro se dibujan sobre cada tiny y en macro solo sobre la giantess y el tiny seleccionado. Cada símbolo tiene una forma distinta, así que también sirve como señal legible sin color.

La giantess tiene un retrato de busto con expresiones neutral, contenta, hambrienta, sorprendida y somnolienta. El retrato acompaña el panel de historia y el inspector resumido, y la expresión sigue sus necesidades y su acción en curso. En micro nunca se le ve la cara, así que el retrato mantiene el vínculo con ella mientras se sigue a la colonia.

En micro, la sombra anticipa la llegada de la mano. El sombreado se intensifica a medida que baja, y desde el aviso se debe poder ver el límite de la zona de riesgo. El tono depende del material que recibe la sombra, con contraste suficiente para distinguirla de las sombras de contacto habituales. La huella sale de la pose y la altura de la mano en el `Frame`.

### Paleta, materiales y capas

La luz entra desde arriba a la izquierda en manchas amplias, con bordes propios del pixel art. Cada material tiene su propia gama de sombra. La madera vira a ámbar tostado, el tatami a verde oliva, el blanco de la cerámica a azul grisáceo y la piel a terracota. Las sombras de contacto son cortas. El canto del mueble, la separación de su sombra y el solapamiento indican altura.

| Familia | Base | Sombra | Uso |
| --- | --- | --- | --- |
| Blanco cálido | `#FFF7E5` | `#C9DCE4` | Papel del shoji y cerámica esmaltada |
| Madera miel | `#DDA86E` | `#9F663D` | Pisos, escritorio y mesa del living |
| Verde junco | `#B6C56B` | `#6A874E` | Tatami y bordes de tela |
| Celeste | `#B9E5EF` | `#68A9BE` | Azulejos y reflejos de ventana |
| Verde hoja | `#62AD52` | `#2C6D43` | Plantas y accesorios tiny |
| Piel cálida | `#FFC8A1` | `#C97760` | Rostro y palma bajo luz de mañana |
| Castaño | `#50312C` | `#251A17` | Cabello de la giantess |
| Azul cobalto | `#326DE7` | `#1A3D98` | Cárdigan y manga de la giantess |
| Rojo tomate | `#EB5547` | `#A93434` | Paño de cocina y bufandas tiny |
| Mandarina | `#FF963D` | `#C65324` | Botón del cárdigan y textiles del living |

Las superficies usan tres valores principales de luz y sombra. En macro se leen como planos amplios. En micro la madera gana vetas espaciadas, el papel muestra sus bordes y la cerámica concentra el brillo en el borde y el asa. Las texturas alrededor de rutas, migajas y zonas de actividad mantienen un contraste menor que el de los tinies y lo que cargan. Un paso abierto se marca con dos extremos unidos y un paso cerrado pierde esa conexión. La forma permite distinguir ambos estados en escala de grises.

Cada pieza de arte declara un pivote de apoyo y, cuando corresponda, un anclaje de mano o carga. La presentación dibuja suelo y sombras, superficies y frentes, y ocupantes por sus puntos de apoyo. Los frentes de primer plano pueden tapar parte de los sprites. La altura y el parentesco separan superficies superpuestas, y el orden dentro de cada superficie se resuelve con la posición de apoyo en pantalla y un desempate estable. Si un borde oculta a la entidad seleccionada, su contorno o marcador sigue visible. El punto de selección coincide con la entidad que se ve bajo el cursor.

### Personajes, poses y animación

| Elemento | Animación mínima para la escena |
| --- | --- |
| Giantess | Reposo y caminata en cuatro direcciones, mano apoyada y recogida, comer sentada y mirar TV. Escribir, dormir y tomar un snack pueden reutilizar poses coherentes cuando el motor muestra esas acciones |
| Palma | Palma elevada y apoyada, con sombras y continuidad de manga. El dibujo se ancla al Space de la mano |
| Tiny | Reposo, caminata y acarreo en cuatro direcciones, y caída. La carga se mueve con el tiny y su silueta sigue visible |
| Objetos | Taza en mano y sobre la mesa, migajas sueltas y depósito del nido. Los cambios de estado se reflejan en la vista correspondiente |
| Expresiones | Símbolos de manga para giantess y tinies, y retrato de la giantess con cinco expresiones |

Se parte de ciclos de caminata de cuatro cuadros y una cadencia visual de 6–10 cuadros/s. El movimiento del pelo y la ropa acompaña la zancada. Al recoger algo, comer o apoyarse, la pose de preparación y la pausa al final hacen legible el gesto. Las direcciones se eligen entre variantes dibujadas y el pivote queda estable entre cuadros. El inicio de cada pose o animación sale del estado o los eventos de la simulación. El paso entre mesa y mano se muestra transitable mientras el portal está activo, con el apoyo, los pasajeros y el indicador alineados al estado del mundo.

### Interfaz y alcance de arte

La interfaz de partida muestra hora, pausa y velocidad, ubicación, selección y seguimiento. El panel de historia destaca pocos cambios relevantes de la rutina y permite localizar al actor o la superficie cuando siguen disponibles. En micro, el rótulo identifica el contexto con una jerarquía como «Dormitorio › Escritorio» o «Giantess › Palma». La intervención del jugador se limita a los controles existentes de cámara, tiempo, inspección y órdenes de destino donde la simulación las acepta. Los overlays técnicos y el inspector de componentes quedan disponibles en el visor de debug.

La interfaz de partida usa paneles casi blancos, un borde fino azul grisáceo y texto azul tinta. La tipografía pixel incluye tildes y eñes. Los globos del historial llevan una cola hacia el actor cuando se lo puede ubicar, y el retrato aporta expresión al panel sin tapar la acción. El cian vivo identifica selección y pasos disponibles. El ámbar identifica avisos de peligro. Cada señal se reconoce también por su forma o icono.

El primer lote de arte cubre suelo, paredes, puertas y mobiliario del dormitorio, cocina y living con la ambientación japonesa. El lote micro cubre escritorio, palma, mesa del living, nido, tinies, taza y migajas. Cada pieza tiene un archivo fuente editable con capas, versionado junto al atlas, y un comando reproducible arma el PNG y el manifiesto a partir de esas fuentes. El atlas exportado no se edita a mano. Las exportaciones que consume el juego son PNG transparentes y metadatos JSON con identificador visual, escala, rectángulos del atlas, pivotes, capas y clips. La asignación de arte usa identificadores visuales estables independientes del nombre mostrado. La paleta y los tamaños se calibran con la primera escena antes de extender el lote.

La revisión de arte se calibra con la casa en macro y con el escritorio y la palma en micro. A 1× se deben distinguir los tres cuartos por sus superficies y acentos, reconocer a la giantess entre sprite, retrato y manga, y seguir a un tiny con su carga tanto sobre madera como sobre piel. En gris tienen que seguir leyéndose el borde de la mano, la silueta del tiny y el estado del paso. Las capturas actuales sirven como referencia de escala y encuadre. [Muestra de presentación](specs/presentacion.md#prueba-de-arte)

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

## 12. Riesgos que resolvería temprano

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

La casa ya cuenta con [sonido](#sonido), las consecuencias de [Un descuido en el escritorio](#14-un-descuido-en-el-escritorio), [rutinas y encuentros](#15-rutinas-y-encuentros-en-la-casa) y [asientos transitables](specs/asientos.md). La [próxima entrega](#16-vida-cotidiana-de-los-tinies) se centra en las decisiones de los tinies y en cómo su actividad cambia el nido y sus alrededores. Para entregas posteriores quedan más superficies explorables, construcción libre para la colonia, relaciones, personalidad, eventos ambientales, fluidos, simulación anatómica continua, economías más amplias y una banda sonora.

## 14. Un descuido en el escritorio

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

- La mano suspendida sobre un mueble proyecta su huella sobre él mientras está a menos de `ThreatHeight` de la superficie. Esa huella guía la sombra que dibuja la vista micro.
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
- **Restos.** Una silueta gris y plana conserva el color de la bufanda del tiny para que se pueda reconocer en micro. En macro, una marca sobre el nido indica cuántos miembros perdió la colonia.
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

## 15. Rutinas y encuentros en la casa

La giantess alterna su rutina doméstica con acciones que surgen de lo que hacen los tinies. La partida libre sortea sus necesidades iniciales y su cuidado. El hambre y la diversión vuelven a bajar con el tiempo, y notar a los tinies puede llevarla a observarlos, compartir comida u ofrecer la palma. La [spec de encuentros](specs/encuentros.md) describe la implementación.

### 15.1. Cómo decide la giantess

- **La rutina depende del estado inicial.** Al crear una partida libre, la semilla elige entre unos pocos puntos de partida, como una giantess con hambre, cansada o con ganas de entretenerse. Hambre y diversión decaen suavemente durante la partida para que vuelvan a generar decisiones. Las actividades domésticas se configuran como ofertas de los muebles y compiten en el selector existente por necesidad, distancia y disponibilidad.
- **Los encuentros abren opciones locales.** La percepción de otra escala registra actividad reciente por superficie. Un avistamiento puede motivar que se acerque a mirar. Ver a los tinies buscando comida donde quedan pocas migajas puede motivar que comparta. La frecuencia de cada acción se limita por superficie, incluso cuando los tinies se esconden o nadie se acerca. La situación se comprueba de nuevo al llegar, antes de dejar comida o apoyar la mano.
- **Las necesidades y los encuentros compiten en el mismo selector.** Las opciones de encuentro tienen una utilidad moderada, independiente de que a la giantess le falte diversión. Se mantienen las reglas de compromiso e interrupción del selector. Los tinies eligen su respuesta con sus tareas, exploración y reflejos. El historial distingue qué motivó a la giantess, qué hizo y cómo respondieron los tinies.

### 15.2. Acciones de rutina y encuentro

| Acción | Cuándo está disponible | Qué cambia en la escena |
| --- | --- | --- |
| **Merendar en el living** | Tiene algo de hambre y la mesa del living está disponible. | Lleva la merienda desde la cocina y deja migajas en la mesa. Si los tinies llegan a esa superficie, pueden descubrir otro lugar donde recolectar. |
| **Observar a los tinies** | Nota actividad reciente en una superficie donde hace rato que no se detiene a mirar. | Se acerca a un punto de observación y espera brevemente. Los tinies pueden seguir con el acarreo, alejarse o refugiarse. Ella puede notar nueva actividad durante la pausa. |
| **Compartir migajas** | Ve tinies buscando comida y quedan pocas migajas visibles en esa superficie. | Deja una porción al alcance de los tinies. Cuando la descubren, la colonia puede asignar un acarreo al nido con el sistema existente. |
| **Ofrecer la palma** | Ve un tiny cerca de un punto de apoyo alcanzable. | Anuncia el apoyo con la sombra y abre un portal de contacto durante una espera breve. Si la huella sigue ocupada, cancela el apoyo. Los tinies deciden si cruzan a la palma. |

Merendar y compartir reutilizan la colocación de comida. Ofrecer la palma reutiliza poses y portales de contacto. La [IA de la casa](specs/ai.md) sigue resolviendo las necesidades, la navegación y los trabajos de recolección.

### 15.3. Criterios de aceptación

- En la misma casa, semillas de referencia distintas producen al menos dos órdenes de actividades de la giantess. Con la misma semilla, la partida y su continuación desde un snapshot reproducen las mismas decisiones y eventos.
- La giantess puede detener su rutina para observar una superficie donde vio tinies. Una vez terminada la observación, deja pasar un tiempo antes de volver a observar en el mismo lugar.
- Si la giantess ve a tinies buscando comida en una superficie con pocas migajas visibles, puede compartir allí. Se puede rastrear la porción desde que la deja hasta que la colonia la descubre y almacena. Si ya hay suficientes migajas a la vista, compartir deja de ser una opción.
- Ofrecer la palma tiene dos desenlaces verificables. En uno, un tiny cruza por el portal por decisión propia. En otro, nadie sube y la giantess retira la mano al acabar la espera.
- El historial permite distinguir una merienda que deja migajas de un gesto deliberado de compartir. Solo muestra un cruce o un acarreo cuando efectivamente ocurrió.

### 15.4. Acciones posteriores

| Acción | Cuándo surge | Qué agrega |
| --- | --- | --- |
| **Ordenar el escritorio** | Terminó de comer y quedó una taza sobre el escritorio. | Recoge esa misma taza y la lleva a la cocina. Requiere mover un objeto que ya existe. |
| **Llevar a un pasajero** | Un tiny subió a la palma y hay una superficie de desembarco accesible. | Elige la mesa del living o el escritorio como destino y vuelve a ofrecer contacto. El pasajero decide si baja. |

## 16. Vida cotidiana de los tinies

En **Nueva partida**, los tinies ya pueden descubrir y acarrear migajas, comer, descansar en el nido, cruzar portales y huir de amenazas. Cuando ninguna tarea gana prioridad, el forrajeo elige un punto al azar en la superficie y la curiosidad puede atraerlos a un portal. La próxima iteración les da motivos para salir, volver y quedarse un rato en el nido. Al seguir a un tiny en vista micro, se debería poder ver qué intenta hacer, adónde va y qué cambia cuando termina. [Sistemas de IA y colonia](specs/ai.md)

### 16.1. Lluvia de ideas

| Idea | Qué decisión agrega | Alcance |
| --- | --- | --- |
| Buscar comida por sectores | Si escasea la comida, decidir qué rincón cercano revisar o a qué fuente conocida volver. | Próxima iteración |
| Volver al nido a comer o descansar | Decidir entre seguir con la tarea o volver antes de que el hambre o el cansancio se vuelvan urgentes. | Próxima iteración |
| Acondicionar un refugio con fibras de papel | Dedicar un viaje a traer material para que las próximas salidas tengan dónde resguardarse. | Próxima iteración |
| Juntarse a conversar | Buscar a otro tiny en una pausa y pasar un rato en el nido cuando hay margen. | Próxima iteración |
| Hacer guardia y avisar | Quedarse cerca de un paso para advertirle a otro tiny que vienen pisadas o que va a bajar la palma. | Más adelante |
| Preparar un depósito en otra habitación | Acortar los viajes con una reserva intermedia y decidir cuánta comida dejar lejos del nido. | Más adelante |
| Hacer una expedición en la palma | Aprovechar un contacto breve para alcanzar una fuente lejana y buscar una ruta de regreso. | Más adelante |
| Juntar gotas junto a una taza | Aprovechar un recurso que aparece con la rutina de la giantess, con un motivo distinto del hambre. | Más adelante |

### 16.2. Objetivos de la próxima iteración

1. **Hacer que la despensa motive las salidas.** Si el nido está por debajo de una meta de porciones por miembro, los tinies dan prioridad a acarrear desde fuentes de comida conocidas. Si ninguna está al alcance, un tiny elige un punto de búsqueda que ya vio y al que puede llegar, como un rincón del escritorio o un borde de la silla. Solo un tiny revisa cada punto a la vez. La salida termina al descubrir comida o al revisar el punto. Un punto vacío pierde prioridad por un tiempo. El hallazgo de comida se comparte y genera trabajo con el sistema actual. Con la despensa abastecida, baja la prioridad del forrajeo. La exploración libre sigue sirviendo para descubrir nuevos puntos y pasos.
2. **Dar al nido una rutina visible.** Comer consume una porción y descansar recupera energía. Se ajustan las necesidades y prioridades de la partida libre para que ambos ocurran durante una sesión normal y en momentos distintos según el tiny. El hambre o el cansancio urgentes pueden interrumpir una tarea de colonia. Si el nido está momentáneamente fuera de alcance, el tiny busca una alternativa conocida y accesible, como comer una migaja cercana o esperar en una zona segura hasta poder volver.
3. **Construir un refugio auxiliar junto a una ruta de salida.** En un punto fijo cerca del borde del escritorio, los tinies pueden juntar fibras de papel de una servilleta y llevarlas de a una. Estos trabajos ganan prioridad cuando la despensa está abastecida. Al reunir las fibras requeridas, un tiny termina el refugio. Desde entonces, los tinies pueden elegirlo para huir y quedar protegidos de un apoyo de palma que el refugio resista. Las fibras entregadas quedan en el sitio y los encargos pendientes los puede tomar otro miembro.
4. **Darles algo que hacer juntos.** Cuando están bien alimentados y descansados y no tienen un encargo prioritario, los tinies pueden acercarse a un punto de reunión junto al nido. Si llega otro tiny disponible, conversan brevemente y a ambos les sube la diversión al completar la actividad. Si el compañero tarda o surge una urgencia, la espera termina y el lugar queda disponible. Cada reunión ocupa a dos tinies y tiene una pausa antes de volver a ofrecerse. Para esta actividad alcanza con una necesidad de diversión y una animación o símbolo legible en micro.

La huida conserva la prioridad ante una amenaza inmediata. Si se cierra un paso de la ruta, aparece peligro o desaparece el objetivo, el tiny reevalúa la tarea y libera las reservas cuando la abandona. Los trabajos pendientes se pueden retomar cuando vuelve a haber un objetivo accesible. Los avistamientos de la giantess siguen reconociendo la búsqueda de comida, así que un recorrido nuevo puede acabar en un gesto de compartir migajas. La ficha del tiny muestra su motivo y destino. El historial registra hallazgos, entregas y reuniones completadas.

### 16.3. Entidades nuevas

Cada punto puede ser destino de una acción y guardar su propio estado. La partida coloca los puntos de búsqueda, una servilleta, un sitio de obra y un punto de reunión en las superficies existentes. Las fibras aparecen durante la simulación.

| Entidad | Datos y comportamiento | Ciclo de vida |
| --- | --- | --- |
| **Punto de búsqueda** | Marca un lugar para revisar, como un rincón del escritorio o el borde de la silla. Tiene ubicación, una reserva exclusiva y un plazo antes de volver a ofrecerse si se revisó sin encontrar comida. | Permanece en su superficie. Un tiny que lo descubre y puede llegar lo reserva mientras lo revisa. El punto queda libre al terminar o abandonar la salida. |
| **Servilleta** | Fuente de fibras en el escritorio. Guarda cuántas quedan y se puede reservar mientras un tiny desprende una. | Al completar la acción, gasta una unidad de su reserva y crea una fibra suelta en el mismo paso. Al agotarse, deja de ofrecer material. |
| **Fibra de papel** | Una unidad de material con su propia ubicación. Se reserva y se lleva con el mecanismo de acarreo existente. | Nace al desprenderse de la servilleta. Si el tiny la suelta al huir, queda en el lugar para que alguien la recoja. Cuando se confirma su entrega en la obra, cuenta una vez y la fibra se retira de la escena. |
| **Obra del refugio auxiliar** | Sitio fijo junto a la ruta de salida del escritorio. Guarda cuántas fibras requiere, cuántas recibió y si terminó el armado. Un tiny reserva el sitio para armarlo cuando están todos los materiales. | Conserva el mismo ID al pasar de obra a refugio. Al terminar, recibe `shelter`, el componente que ya define la resistencia y el radio de protección del nido. Desde ese momento puede ser destino de huida. |
| **Punto de reunión** | Lugar junto al nido con dos plazas, un plazo para que llegue el segundo tiny y una pausa entre conversaciones. Su ocupación es independiente del uso del nido para comer o descansar. | Permanece en el escritorio. Al terminar la conversación o cancelarse por falta de compañero, peligro o pérdida de un participante, libera las plazas. |

**Entidades existentes.** El nido mantiene su depósito y recibe la meta de porciones por miembro. Cada tiny guarda su diversión en `needs` y su actividad en `agent.doing`. La comida cerca de la silla se representa con migajas. Desprender una fibra, acarrearla y armar el refugio son encargos sobre la servilleta, la fibra y la obra respectivamente. Cada encargo se crea como una entidad con el componente `job` y reserva su objetivo cuando se asigna. [Reservas y trabajos](specs/ai.md#reservas-y-trabajos-jobs)

**Descubrimiento y coordinación.** La búsqueda usa la memoria de los tinies. La percepción incorpora los puntos de búsqueda, la servilleta y las fibras cercanas para que puedan descubrirlos y compartir el hallazgo con la colonia. Antes de pedir otra fibra, la obra cuenta las que ya recibió, las que están en viaje y las que quedaron sueltas y se pueden recuperar. El total entregado se conserva aunque cambie el encargado. [Percepción y memoria](specs/ai.md#percepción-y-memoria)

### 16.4. Cierre de la iteración

El escenario de referencia parte de la casa existente. El nido tiene pocas porciones, queda una fuente de comida por descubrir cerca de la silla y hay una servilleta con fibras en el escritorio. Los tinies deciden por su cuenta quién busca, quién acarrea y cuándo vuelve. Una vez abastecida la despensa, la colonia puede terminar el refugio mientras otros miembros comen, descansan o se reúnen. Una amenaza durante un acarreo pone a prueba la huida y la posibilidad de retomar el trabajo.

- **Abastecimiento.** Un escenario controlado verifica el ciclo de búsqueda, hallazgo, acarreo, depósito y consumo. Con la despensa abastecida, se inician menos salidas de búsqueda que en la misma situación con pocas porciones.
- **Rutina y ocio.** En los primeros cinco minutos de simulación de cada semilla de referencia, los tinies completan al menos tres actividades distintas entre buscar, acarrear, comer, descansar, preparar el refugio y reunirse. En escenarios controlados, descansar aumenta la energía y una conversación completada aumenta la diversión de los dos participantes. Si un tiny espera solo hasta que vence el plazo, se retira y libera el punto de reunión.
- **Coordinación y progreso.** Dos tinies no se reservan el mismo punto, migaja o fibra. Cada fibra entregada cuenta una vez y el total se conserva si cambia el encargado. Al terminar el armado, la obra conserva su ID y recibe la protección. Un tiny dentro queda protegido ante un apoyo de palma que el refugio pueda resistir.
- **Interrupciones.** Una amenaza interrumpe un acarreo o una conversación y el tiny intenta huir. Si un portal se cierra a mitad de una salida, el tiny vuelve a evaluar sus opciones sin insistir continuamente en la ruta cerrada. Las tareas abandonadas dejan sus reservas disponibles.
- **Lectura y continuidad.** En micro se distinguen búsqueda, acarreo, regreso por hambre o cansancio y reunión. Al continuar desde un snapshot a mitad de una salida o de la preparación del refugio, se obtienen los mismos eventos y el mismo hash que en la ejecución continua.

## 17. Interfaz y lectura en PC

El juego corre en PC y todavía se ve como si fuera para una pantalla chica. La escena y la interfaz comparten un lienzo fijo de 480 × 270, la casa ocupa una franja angosta y la historia se llena de movimientos sin importancia. Esta iteración reorganiza la presentación para aprovechar la ventana y aclarar qué está pasando. Nada de esto cambia la simulación: el hash, los snapshots y los escenarios siguen iguales.

### 17.1. Diagnóstico

| Problema | Causa actual | Efecto |
| --- | --- | --- |
| La ventana es chica | El juego abre a 960 × 540 (lienzo 2×). Se usa el mayor factor entero que entra y el resto se rellena con bandas | En un monitor de 1920 × 1080 queda mucho escritorio sin usar. A pantalla completa el factor sube a 4 y todo crece por igual, incluida la interfaz, así que no se ve más mundo |
| La casa se ve comprimida | Las tres habitaciones miden 9 × 3 m. A 40 px/m ocupan 360 × 120 px del lienzo | La casa usa menos de la mitad del alto. La giantess mide 31 × 38 px y se le distinguen pocos gestos |
| No se puede hacer zoom sobre el arte | La rueda solo funciona en el visor vectorial. En pixel art, `fitArt` fija una densidad por escala | Entre la casa y el escritorio no hay niveles intermedios: la escala salta 40 veces |
| La interfaz tapa la escena | Barras, paneles y rótulos se dibujan sobre el mismo lienzo que el mundo | El panel de historia tapa los rótulos de Cocina y Dormitorio (captura `casa.png`). A pantalla completa, la fuente de 12 px se ve a 48 px |
| La historia no cuenta nada | `storyText` convierte cada evento suelto en una línea y el panel muestra las últimas seis | En 5 minutos de `casa.json` hay 63 `explored`, 81 cambios de Space, 15 cambios de pose y 9 acarreos asignados, contra 1 aplastamiento, 7 sustos y 1 palma ofrecida. A x4, un hito desaparece del panel en segundos |

### 17.2. Un lienzo que llena la ventana

La escena deja de tener tamaño fijo. Se elige un **factor de escena** entero *N* y el lienzo pasa a medir `ventana / N`, redondeado hacia arriba. Así la imagen llena la ventana sin bandas. El mínimo de 480 × 270 lógicos sigue siendo el tamaño más chico admitido. La proyección ya trabaja con un ancho y un alto cualesquiera (`camera.Project`), y `render.PixelViewport` pasa a calcular el factor y el tamaño del lienzo en lugar de centrar uno fijo.

- **Ventana al arrancar.** El juego abre maximizado, o al 90 % del monitor si maximizar no está disponible. `F11` alterna pantalla completa. El tamaño y el modo se guardan en una configuración local de presentación, separada de las partidas y de la simulación.
- **Factor por defecto.** Se usa el mayor *N*, con un máximo de 3, que permite ver la casa entera con margen para los paneles. Con la densidad del punto 17.3, eso da *N* = 2 en 1080p y *N* = 3 en 1440p.
- **Capturas reproducibles.** `-shot`, `capturas.json` y `-bench` fijan el tamaño con `-size`, así que las referencias siguen siendo deterministas. Las capturas de referencia se regeneran una vez.

**Estado.** Implementado. Con la densidad de 17.3, el lienzo mínimo es 768 × 432 (la casa de 720 × 240 más los paneles) y *N* tiene un tope de 3 para que en pantallas grandes la interfaz no se agrande: en 1080p, 960 × 540 a ×2; en 1440p, 854 × 480 a ×3. Las capturas de referencia se toman a 960 × 540, el lienzo de 1080p a 1×. [Presentación](specs/presentacion.md)

### 17.3. Más detalle en la vista macro

La casa es una sola zona, así que macro puede trabajar con el doble de píxeles. La densidad pasa de **16 px por 40 cm a 16 px por 20 cm (80 px/m)**. La casa ocupa 720 × 240 px, una proporción adecuada para un lienzo de 960 × 540 con paneles. La giantess de pie pasa a medir unos 62 × 76 px, suficiente para que se le lean la cara, las manos y lo que lleva. Micro mantiene 16 px por 10 mm.

| Pieza macro | Qué gana con el doble de densidad |
| --- | --- |
| Giantess | Expresión en el sprite (no solo en el retrato), mano y objeto en la mano, poses de sentada y de mano apoyada distinguibles |
| Muebles con uso | Escritorio, mesa, cama, televisor y anafe muestran su estado: taza encima, migajas, refugio, pantalla encendida |
| Suelos y paredes | Tatami, azulejo y madera con textura propia. Puertas y noren se distinguen abiertos o cerrados |
| Tinies | Siguen midiendo menos de un píxel a esta escala, así que se mantienen como marcadores y globos (punto 17.7) |

**Migración de una vez.** El atlas declara su densidad por escala (`density` en el manifiesto) y la cámara la toma de ahí, sin constantes en `art.go`. No hay escalado intermedio: todas las fuentes macro se redibujaron por código a 80 px/m (`tools/art/macro.py` para la casa y los muebles, `tools/art/giantess.py` para la giantess), con la misma ubicación en la escena (tamaño, pivote, profundidad y anclaje de carga al doble) y más detalle: ojos con brillo, flequillo y moño, cárdigan con blusa y lazo, falda tableada y zabuton en la giantess; vetas, tatami con heri, shoji, azulejos, bacha, anafe y alfombra seigaiha en la casa. Los marcadores de tinies y restos, los símbolos y el retrato son interfaz y conservan su tamaño.

**Estado.** Implementado. Ya no quedan piezas macro dibujadas a mano: `tools/art/base/` solo guarda originales micro.

### 17.4. Zoom con la rueda: escena e interfaz en capas separadas

**Sí se puede.** La condición es que la escena y la interfaz dejen de compartir lienzo. Se dibujan en dos imágenes y `DrawFinalScreen` las compone:

- **Capa de escena**: sprites, sombras, señales de paso y símbolos. Se escala por el factor de escena *N*, que la rueda modifica.
- **Capa de interfaz**: barras, paneles, ficha, avisos, rótulos de habitaciones, globos de tinies y marcadores. Usa un **factor de interfaz** *U* que no depende del zoom. Por defecto `U = max(1, alto / 450)` (2 en 1080p, 3 en 1440p, 4 en 4K) y se puede cambiar en Opciones con Automático o de 1 a 4.

El costo es que los píxeles de la escena y los de la interfaz pueden tener tamaños distintos. Es una convención habitual en juegos de pixel art con zoom, y se acepta si *U* se mantiene fijo durante la partida. Los elementos que señalan algo del mundo, como rótulos, globos, marcadores de selección y colas de los globos de historia, se proyectan desde la escena y se dibujan en la capa de interfaz. Así mantienen un tamaño legible con cualquier zoom.

| Regla | Decisión |
| --- | --- |
| Pasos de zoom | Solo enteros: *N* de 1 a 4 en macro y de 2 a 4 en micro. Cada paso mantiene fijo el punto bajo el cursor (`Camera.ZoomAt`). No se re-muestrea con filtro, así que el arte nunca se ve borroso. Un zoom fraccionario con filtro *sharp bilinear* queda como opción posterior |
| Zoom semántico | Acercar más allá del máximo de macro sobre una superficie con vista micro entra en micro en ese punto, con el mismo fundido de 180 ms. Alejar más allá del mínimo de micro vuelve a la casa con el encuadre anterior. `Tab` mantiene su función |
| Paneo | Se arrastra con el botón izquierdo a partir de 4 px de umbral, como en el visor vectorial, o con el del medio. WASD y las flechas siguen funcionando. Clic derecho sigue siendo la orden de destino |
| Picking | El cursor se prueba primero contra la interfaz y después contra la escena, cada uno con su propia transformación. `ArtScene.Pick` sigue trabajando en píxeles de la escena |
| Límites del encuadre | La cámara no deja que la casa o la superficie salgan por completo de la pantalla. `F` vuelve al encuadre general |

La cámara es presentación pura. `Camera.Zoom` sigue en px/µm y el factor de escena es solo cómo se muestra.

**Estado.** Implementado en parte, con el foco en alejar: capas separadas, rueda por los factores ½, 1, 2, 3, 4, 6 y 8 en las dos escalas (½ reduce promediando con mipmaps; el resto es entero) y punto fijo bajo el cursor, `F` vuelve al factor del lienzo. La interfaz sigue con el factor del lienzo *N* (todavía sin *U* configurable en Opciones). Faltan el zoom semántico, el paneo con arrastre y los límites del encuadre. [Presentación](specs/presentacion.md)

### 17.5. Historia: una crónica de hitos y un diario por entidad

La historia tiene dos niveles. La **crónica** cuenta la partida solo con lo que importa. El **diario** de cada entidad conserva el detalle, porque «Tiny 07 explora la cocina» no aporta nada en la crónica general, pero sí sirve para ver qué hizo ese tiny en particular.

| Nivel | Qué entra | Ejemplos |
| --- | --- | --- |
| **Hito** (crónica, con icono y, si es grave, aviso) | Muertes, sustos, tinies atrapados, encuentros, primeras veces y cambios de estado de la colonia | «Tiny 03 quedó bajo la palma en el escritorio». «La giantess ofrece la palma; suben 2 tinies». «Por primera vez un tiny llega a la cocina». «La despensa se vació». «Terminan el refugio» |
| **Rutina de la giantess** (crónica, una línea por actividad) | Cada actividad en una sola línea que se completa al terminar | «07:00–07:24 Desayuna en la cocina» |
| **Resumen de la colonia** (crónica, cada 30 minutos de reloj o cuando la giantess cambia de actividad) | Lo que hizo la colonia en ese período, en una sola línea | «07:00–07:30 · 6 migajas al nido · 2 fibras · 1 conversación». No se escribe si no pasó nada |
| **Detalle** (solo en el diario de la entidad y del lugar) | Lo que hoy llena el panel | Explora, sube o baja, encargo asignado, entrega, búsqueda vacía, come, descansa, conversa |

**Cómo se arma.** `view.Chronicle` reemplaza a `storyText`. Clasifica cada evento publicado como hito, rutina, aporte al resumen o detalle. Las primeras veces se registran por pareja de quién y dónde. Los detalles de un mismo tipo y lugar que llegan juntos se agrupan («3 tinies exploran la cocina»). Cada entrada guarda su actor y su lugar, como hoy, para poder enfocarlos. Cada diario conserva las últimas entradas de su entidad. Cuando muere un tiny, los restos heredan su diario y su nombre. La crónica sale de los eventos y no del estado, así que sigue siendo presentación. Al cargar una partida empieza vacía, salvo que más adelante se guarde junto al snapshot, en un archivo aparte que no se hashea.

**Criterio de aceptación.** En los primeros 5 minutos de `casa.json` con la semilla de referencia, la crónica no tiene ninguna línea de «explora», «sube a» ni «asigna un acarreo». Muestra todos los aplastamientos, sustos y encuentros. La ficha de un tiny que exploró muestra esas exploraciones en su diario.

**Estado.** Implementado (`view.Chronicle`, verificado por `TestChronicleAcceptance`). La crónica reemplaza a `Runner.Story`. Los hitos se distinguen por el color de la hora: coral los graves y turquesa los demás. Los iconos por tipo y el panel con pestañas quedan para 17.6. Los avisos siguen siendo los de los decesos. [Presentación](specs/presentacion.md#interfaz-de-partida)

### 17.6. Interfaz de partida y menú de pausa

Con la capa de interfaz a 960 × 540 lógicos en 1080p hay espacio para paneles que no tapen la casa.

| Zona | Contenido |
| --- | --- |
| **Barra superior** | Reloj, pausa y velocidad, ubicación jerárquica, población, despensa (icono y porciones sobre la meta), seguimiento y sonido |
| **Panel lateral izquierdo** (se pliega con `H`) | Tres pestañas. **Crónica**: la lista completa con desplazamiento; cada hito lleva icono y color (ámbar para peligro, cian para encuentros) y se enfoca con un clic. **Colonia**: barra de la despensa, avance del refugio, cuántos tinies hacen cada actividad (un clic elige a uno) y la lista de tinies con su actividad y alertas de hambre o cansancio. **Giantess**: retrato, necesidades, actividad actual y destino |
| **Ficha** (derecha, al elegir algo) | Nombre, ubicación, motivo y destino, carga, necesidades, las últimas entradas del diario y los botones Seguir e Ir allí |
| **Avisos** (arriba a la derecha) | Decesos, como hoy, más amenazas a la colonia, palma ofrecida y refugio terminado. Un clic lleva al lugar |
| **Barra inferior** | Solo tres o cuatro controles según el contexto. `F1` o `?` abre la ayuda completa con la leyenda de símbolos |

**Menú de pausa (`Esc`).** Pausa la partida y ofrece: Continuar, Guardar partida, Cargar partida, Opciones (volumen, escala de interfaz, pantalla completa, movimiento reducido y pausa ante amenazas), Controles y Volver al menú principal. Reemplaza a la pregunta actual de volver al menú. `Esc` sigue soltando primero la selección y el seguimiento, como hoy. Guardar y cargar usan el snapshot existente de `-save`/`-load`.

### 17.7. Otras mejoras para entender qué pasa

| Mejora | Qué aclara |
| --- | --- |
| **Información al pasar el cursor** | Nombre y actividad de lo que está debajo, sin hacer clic. En macro, sobre un globo de tinies, la lista de quiénes están ahí y qué hacen |
| **Intención de la giantess** | En macro, un globo con el icono de su actividad y una línea punteada hasta su destino mientras camina. Se entiende adónde va antes de que llegue |
| **Amenazas visibles desde macro** | Si una mano o una pisada va a bajar sobre una superficie con tinies, el globo de esa superficie se pone ámbar y aparece un aviso. Con «pausa ante amenazas», el juego se pausa o baja a x1 |
| **Globos de tinies con actividad** | El globo de cada superficie muestra la actividad dominante con un icono (comen, acarrean, descansan, huyen), además de la cantidad |
| **Miniplano en micro** | Un recuadro chico de la casa con la superficie actual y la posición de la giantess, para no perder la orientación al seguir a un tiny |
| **Rótulos que no se tapan** | Los rótulos de habitaciones se colocan en la capa de interfaz y evitan los paneles. Si una habitación queda tapada, el rótulo se pega al borde visible |
| **Atajos de selección** | Doble clic sigue a la entidad. Al pasar el cursor por la etiqueta de `Tab`, se resalta la superficie a la que lleva |

### 17.8. Orden y aceptación

1. **Capas y lienzo variable.** Escena e interfaz separadas, ventana grande, `F11`, zoom entero con la rueda, paneo con arrastre y capturas regeneradas. Todavía sin arte nuevo.
2. **Crónica y diarios.** `view.Chronicle`, resumen de la colonia y diario en la ficha.
3. **Interfaz de partida.** Panel lateral con pestañas, ficha ampliada, avisos y menú de pausa con Opciones, Guardar y Cargar.
4. **Macro a 80 px/m.** Densidad declarada en el atlas, arte provisorio escalado y repintado de la giantess y los muebles con uso.
5. **Lectura.** Zoom semántico, intención de la giantess, amenazas en macro, miniplano e información al pasar el cursor.

En todos los pasos, el hash y los eventos de los escenarios no cambian (`TestScenarios`, `hash_test`). En 1920 × 1080, la casa ocupa al menos el 60 % del ancho, sin bandas, y ningún panel tapa los rótulos de las habitaciones. El paso entre macro y micro con la rueda conserva el punto bajo el cursor. Los tests de `ui` verifican que el diseño de la interfaz entra en 1×, 2× y 3×. `-bench` mantiene los fps de referencia con dos capas.
