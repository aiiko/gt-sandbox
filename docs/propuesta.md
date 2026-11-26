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

La [spec del mapa corporal](specs/cuerpo.md) detalla las rutas por el cuerpo entero, los apoyos que abren accesos a la casa y el seguimiento en micro.

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
| Macro | Habitaciones, puertas, muebles, rutina y desplazamiento de la giantess | 16 px por 20 cm (80 px/m). La giantess de pie usa un lienzo de 62 × 76 px | La casa completa y el viaje del dormitorio al living |
| Micro | Tinies individuales, migajas, bordes y pasos entre superficies | 16 px por 10 mm. El tiny tiene una silueta visible de 12–16 px de alto | El nido y las migajas sobre el escritorio, el borde de la palma y el descenso a la mesa del living |

La escena y la interfaz se dibujan en capas separadas sobre un lienzo que llena la ventana (mínimo 768 × 432 px lógicos, con factores de zoom en escena e interfaz independiente; ver [Presentación](specs/presentacion.md)). La posición se interpola primero y se alinea a la cuadrícula de píxeles al dibujar. La cuadrícula visual es independiente de las celdas de navegación.

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

El núcleo ya funciona como ecosistema vivo: la giantess sigue su rutina mientras los tinies exploran, trabajan y reaccionan a su entorno, y el jugador puede pasar entre macro y micro leyendo la historia de la partida. Los sistemas implementados están documentados en `docs/specs/` ([núcleo](specs/nucleo.md), [espacios](specs/space.md), [navegación](specs/nav.md), [contactos](specs/attach.md), [IA](specs/ai.md), [aplastamiento](specs/crush.md), [encuentros](specs/encuentros.md), [asientos](specs/asientos.md), [vida cotidiana](specs/vida.md), [oficios y expediciones](specs/oficios.md), [mapa corporal](specs/cuerpo.md), [actividades](specs/actividades.md), [presentación](specs/presentacion.md), [visor](specs/visor.md) y [sonido](specs/sonido.md)).

Para entregas posteriores quedan líneas abiertas como [prendas explorables](specs/prendas.md), fuerzas físicas no letales (viento/barridos), micro-fauna, construcción libre para la colonia, relaciones y personalidad, eventos ambientales, fluidos, simulación anatómica continua, economías más amplias y una banda sonora.
