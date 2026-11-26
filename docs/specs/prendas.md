# Spec: Prendas explorables (mueble ↔ cuerpo)

Paquetes afectados: `sim/attach`, `sim/crush`, `sim/ai`, `sim/colony`, `content`, `scenario`, `view`, `render`, `sound`, `cmd/game`.

## 1. Propósito y experiencia de juego

Una **prenda explorable** es una entidad `Space` (`space.Rigid`) que alterna entre dos modos físicos dentro de la misma simulación:

1. **Apoyada en la casa (`placed`)**: descansa sobre un mueble o el piso (p. ej. pantuflas junto al futón o en el baño, cárdigan sobre la silla del escritorio). Los tinies pueden entrar a explorarla, buscar fibras/tesoros o refugiarse en su interior.
2. **Puesta en la giantess (`worn`)**: se reengancha (`space.Reparent`) a una parte del cuerpo de la giantess (pie, torso, brazo). Todo tiny u objeto que quedó dentro conserva sus coordenadas locales y viaja con ella por toda la casa sin haber tenido que subir por la mano.

### Situaciones emergentes buscadas
- **Viaje accidental**: un tiny entra a explorar una pantufla junto al futón; la giantess se despierta, se la calza y camina a la cocina o al baño; cuando se la saca para ducharse, el tiny desembarca en una habitación nueva.
- **Aviso y escape al calzarse**: ponerse una prenda no aplasta de golpe; abre una ventana de aviso (`donning`) durante la cual un tiny puede huir por la abertura hacia el piso o meterse en una zona segura (puntera, borde del empeine o bolsillo).
- **Cuidado de la giantess**: si la giantess tiene `agent.care` alto y nota actividad dentro de la prenda antes de ponérsela, titubea y espera (mismo mecanismo que con la palma).
- **Conexión con el mapa corporal**: mientras está puesta, la prenda habilita portales internos hacia el cuerpo (pantufla ↔ tobillo, cárdigan ↔ hombro/manga), integrándose con las expediciones corporales existentes (`docs/specs/cuerpo.md`).

Todo se activa por contenido: los escenarios sin entidades con `garment` conservan exactamente su comportamiento actual.

---

## 2. Modelo de simulación (`sim/attach`)

Se integra en `sim/attach` (junto a `BodyState`, `BodyJoints`, `BodyZone` y `contact_surface`) para mantener acoplada la lógica de poses, contactos y jerarquía corporal sin abrir un paquete nuevo innecesario.

### 2.1. Componente `Garment` (store `"garment"`)

Vive en la entidad `Space` de la prenda (que además lleva `space.Space{Kind: space.Rigid}`, `comp.Location`, `nav.Grid`, `attach.Surface` y `attach.BodyZone`).

```go
type GarmentState string

const (
    GarmentPlaced  GarmentState = "placed"  // apoyada en un mueble o piso
    GarmentDonning GarmentState = "donning" // poniéndosela: ventana de aviso previa
    GarmentWorn    GarmentState = "worn"    // montada sobre MountPart del cuerpo
    GarmentDoffing GarmentState = "doffing" // sacándosela (opcional, breve)
)

// InnerPressure define la región interior de la prenda que se comprime
// al completarse el calce (donning -> worn). Todo lo que queda fuera de
// Zone (borde, puntera, bolsillo) es seguro durante el calce.
type InnerPressure struct {
    Zone   units.Rect `json:"zone"`
    Force  int32      `json:"force"`
    Source string     `json:"source,omitempty"` // por defecto "garment"
}

type Garment struct {
    Owner ids.ID       `json:"owner"`
    Slot  string       `json:"slot"` // único por Owner en estado worn (ej. "slipper_left", "cardigan")
    State GarmentState `json:"state"`

    // Ventana de transición (donning / doffing).
    StateUntil clock.Tick `json:"state_until,omitempty"`
    // Destino pendiente cuando State == "doffing".
    DoffSpace  ids.ID          `json:"doff_space,omitempty"`
    DoffPos    units.Vec3      `json:"doff_pos,omitempty"`
    DoffFacing fixedmath.Angle `json:"doff_facing,omitempty"`

    // Configuración al estar puesta (mounted sobre MountPart).
    MountPart      ids.ID            `json:"mount_part"`
    WornPos        units.Vec3        `json:"worn_pos"`
    WornFacing     fixedmath.Angle   `json:"worn_facing,omitempty"`
    WornProjection *space.Projection `json:"worn_projection,omitempty"`

    // Proyección al estar apoyada (nil = plano XY).
    PlacedProjection *space.Projection `json:"placed_projection,omitempty"`

    // Portales internos prenda <-> zonas del cuerpo (p. ej. pantufla <-> tobillo).
    // Habilitados solo cuando State == "worn"; se alterna Disabled sin recrearlos.
    BodyPortals []ids.ID `json:"body_portals,omitempty"`

    // Presión interior al completar el calce (nil para prendas/bolsillos sin compresión).
    InnerPressure *InnerPressure `json:"inner_pressure,omitempty"`

    // Aviso base en ticks antes de completar el calce (0 = DefaultDonWarnTicks).
    DonWarnTicks clock.Tick `json:"don_warn_ticks,omitempty"`
}
```

### 2.2. Comandos

Registrados en `attach.Install`:

| Comando | Campos | Validación atómica (si falla, no muta nada) | Efecto |
| --- | --- | --- | --- |
| `DonGarment` (`"don_garment"`) | `Actor`, `Garment`, `WarnTicks` (opcional) | `Actor` y `Garment` vivos; `Garment` tiene `garment` con `Owner == Actor`, `State == "placed"`, `MountPart` vivo y montado sobre `Actor`; ninguna otra prenda del mismo `Owner` + `Slot` está en `donning`/`worn`. | Pasa a `State = "donning"`, fija `StateUntil = w.Tick() + warnTicks` y emite `GarmentDonning`. |
| `CancelDon` (`"cancel_don"`) | `Actor`, `Garment` | `Garment` en `State == "donning"` y `Owner == Actor`. | Vuelve a `State = "placed"`, limpia `StateUntil` y emite `GarmentDonCancelled`. |
| `DoffGarment` (`"doff_garment"`) | `Actor`, `Garment`, `TargetSpace`, `Pos`, `Facing` | `Actor` y `Garment` vivos; `State == "worn"`, `Owner == Actor`; `TargetSpace` es un `Space` vivo que no cuelga de `Garment` (sin ciclos). | Deshabilita `BodyPortals` (`Version++`), reparenta `Garment` a `TargetSpace` en `(Pos, Facing)` con `PlacedProjection`, pasa a `State = "placed"` y emite `GarmentDoffed`. |

> **Nota sobre `DonGarment` inmediato en tests**: si `WarnTicks < 0` (o en escenarios que prueban solo el viaje), el comando puede usar un mínimo de `1` tick o permitir completar en el próximo `PhaseTopology` para mantener siempre el mismo orden de fases.

### 2.3. Eventos

- `GarmentDonning {Actor, Garment, Slot, Until}` (`"garment_donning"`)
- `GarmentDonCancelled {Actor, Garment, Slot}` (`"garment_don_cancelled"`)
- `GarmentWorn {Actor, Garment, Slot, MountPart}` (`"garment_worn"`)
- `GarmentDoffed {Actor, Garment, Slot, Surface}` (`"garment_doffed"`)

### 2.4. Sistema `attach.garments`

- **Fase**: `clock.PhaseAct`, `Order: -5` (antes de `attach.contacts` en `PhaseTopology` y antes de `crush.resolve` en `PhaseConsequences`).
- **Iteración determinista (D5)**: recorre `ecs.Get[Garment](r).IDs()` en orden ascendente de `ids.ID`.
- **Al vencerse `donning` (`w.Tick() >= g.StateUntil`)**:
  1. Si `g.InnerPressure != nil`, registra un evento/marca interna para que `crush.resolve` en `PhaseConsequences` de ese mismo tick evalúe la compresión interior sobre los ocupantes frágiles dentro de `InnerPressure.Zone` (respetando `shelter`).
  2. Ejecuta `space.Reparent` de la prenda a `g.MountPart` (actualizando `Location.Pos = g.WornPos`, `Location.Facing = g.WornFacing` y `Space.Projection = g.WornProjection`). **No modifica las coordenadas locales de los ocupantes dentro de la prenda.**
  3. Cambia `g.State = GarmentWorn`, limpia `g.StateUntil`.
  4. Habilita todos los `g.BodyPortals` vivos (`p.Disabled = false; p.Version++`).
  5. Llama a `checkGripOnPose(w, []ids.ID{garmentID})` si la prenda tiene `RequireGrip` o `WornProjection != nil`, de modo que un cambio brusco de inclinación permita agarrarse a un `GripPoint` o provoque una caída física real.
  6. Emite `GarmentWorn`.
- **Limpieza**: si `Owner` o `MountPart` dejan de estar vivos (`!r.Alive`), una prenda en `donning` vuelve a `placed`; una prenda en `worn` se deja que `space.orphans` la reubique y se fuerza `State = GarmentPlaced` deshabilitando `BodyPortals`.

### 2.5. Invariantes (`world.AddInvariant("attach.garment", ...)`)

Verificadas al final de cada tick:
1. Toda prenda viva tiene `Slot != ""` y `State` válido (`placed`, `donning`, `worn`, `doffing`).
2. Si `State == "worn"`, `Owner` y `MountPart` están vivos, `Location.Space == MountPart`, y no existen dos prendas vivas con mismo `(Owner, Slot)` en estado `worn` o `donning`.
3. Si `State == "placed"` o `State == "donning"`, todos sus `BodyPortals` vivos tienen `Disabled == true`.
4. Si `State == "donning"`, `StateUntil > 0`.

---

## 3. Amenazas, aplastamiento interior y contactos (`sim/crush`, `sim/attach`, `sim/nav`)

### 3.1. Ventana de aviso y huida (`crush.Threats`)
- Mientras una prenda está en `State == "donning"` y tiene `InnerPressure != nil`, `crush.threats` agrega una amenaza sobre el `Space` de la prenda cubriendo `InnerPressure.Zone`.
- Los tinies dentro de `InnerPressure.Zone` la perciben por el canal de sombra/vibración existente en `ai.look` y reaccionan tras `agent.reflex` ticks con `flee`.
- Como `nav.NearestFree` busca la celda transitable más cercana fuera de la huella dentro del mismo `Space`, un tiny puede:
  - correr hacia la **zona segura** de la prenda (borde del empeine, puntera o bolsillo), o
  - escapar por el portal abierto hacia el piso/mueble si está cerca de la entrada.
- Si el tiny queda dentro de `InnerPressure.Zone` al completarse el calce y no está protegido por un `shelter` con `Strength >= InnerPressure.Force`, `crush.resolve` lo aplasta con `Source = "garment"` (mismas reglas: despawn + entidad `remains`, reservas liberadas).
- Si está en la zona segura de la prenda, **sobrevive y viaja montado en ella**.

### 3.2. Contactos con la casa (`attach.Surface`)
- **En estado `placed`**: la prenda tiene `attach.Surface` con `Scale: "tiny"`. Al estar a `Z = 0` sobre el piso o un mueble, `attach.contacts` abre automáticamente un portal de entrada/salida entre el soporte y la prenda.
- **En estado `worn`**: al reparentarse a `MountPart`, la prenda pasa a formar parte del cuerpo de `Owner`.
  - Si es una pantufla en `body_foot_l`, su `attach.Surface` sigue las reglas de `RequireStill: true` y `Access` al costado (igual que el pie): cuando la giantess está quieta con el pie plantado, el borde exterior de la pantufla puede abrir paso al piso; al caminar se cierra.
  - Si es un cárdigan en el torso, al sentarse en el sofá o recostarse en el futón, el borde del bolsillo/costado puede abrir paso al mueble si cumple solapamiento y estabilidad.
- **Validación en el tick de cruce**: `attach.PortalCrossingValid` ya revalida pose y movimiento del soporte antes de cruzar. Se extiende para rechazar el cruce hacia el interior de una prenda que acaba de entrar en `worn` si el portal externo dejó de estar en contacto en ese tick.

### 3.3. Navegación y expediciones (`sim/nav`, `sim/colony`)
- Al pasar de `placed` a `worn`, los portales de contacto con el mueble anterior se cierran en `attach.contacts` y cualquier ruta que dependía de ellos se invalida y replanea.
- Un tiny que queda adentro durante el viaje entra en su flujo habitual: puede explorar la prenda, cruzar por un `BodyPortal` al tobillo/torso si es explorador o aventurero (`Adventure >= 0.5`), o quedar `Stranded` esperando en un `GripPoint` a que la prenda vuelva a apoyarse (`placed`) o abra un apoyo para bajar.
- Cuando un tiny desembarca de una prenda en otra habitación y regresa al nido, `colony` registra el hallazgo en la crónica y en el diario del tiny.

---

## 4. Rutinas de la giantess (`sim/ai`)

### 4.1. Nuevos pasos en `ai.Step` (`internal/sim/ai/use.go`)

```go
const (
    StepDon  = "don"  // ponerse la prenda del Slot indicado
    StepDoff = "doff" // sacarse la prenda del Slot indicado y dejarla en Offset
)
```

Campos reutilizados/nuevos en `ai.Step`:
- `Slot string` (`json:"slot,omitempty"`): ranura objetivo (`"slipper_left"`, `"slipper_right"`, `"cardigan"`, o prefijo/lista si se ponen de a pares).
- `Offset units.Vec2`: desplazamiento relativo a la giantess al ejecutar `doff` (se proyecta sobre el `Space` actual de la giantess en `Z = 0`, verificando que la celda esté dentro de `Extent` y transitable).

### 4.2. Ejecución y cuidado (`agent.care`)
- **`StepDon`**:
  1. Busca la prenda viva del actor con ese `Slot`. Si no existe o ya está `worn`, el paso termina al instante (no-op, retrocompatible).
  2. Si la prenda está `placed` en otro lugar alcanzable, el actor camina hasta ella (`nav.Approach`).
  3. Antes de disparar `DonGarment`, si el actor tiene `agent.care > 0` y `watch` detecta tinies dentro de la prenda, hace la misma tirada de atención (`w.RNG("ai.attention")`) que al apoyar la mano: si los ve, espera hasta `CareWaitMax` con emote `?` antes de calzarse.
  4. Ejecuta `DonGarment` con `WarnTicks = baseWarn + care * HoverExtra` y espera a que la prenda alcance `GarmentWorn`.
- **`StepDoff`**:
  1. Busca la prenda viva del actor con ese `Slot` en estado `worn`. Si no hay ninguna, es no-op.
  2. Calcula la posición de apoyo en el `Space` actual del actor usando `Offset` (clampeada a celdas transitables) y ejecuta `DoffGarment`.
- **Interrupciones**: si una necesidad urgente interrumpe `StepDon` mientras `State == "donning"`, se emite `CancelDon` para no dejar la prenda trabada en `donning`.

### 4.3. Rutinas concretas en la casa (`content/objects/` y `content/partidas/casa.json`)

| Rutina | Oferta | Cambio con prendas |
| --- | --- | --- |
| **Dormir en el futón** | `bed` → `sleep` | Al llegar al futón: `doff` de pantuflas junto al borde; duerme; al despertar: `don` de pantuflas. |
| **Ducharse en el baño** | `shower_tray` → `shower` | En el tramo seco frente a la ducha: `doff` de pantuflas (y cárdigan si lo lleva); se ducha; al terminar: `don` de pantuflas (y cárdigan). |
| **Trabajar / merendar** *(Fase 4)* | `desk` / `chair` | Al sentarse en el escritorio puede hacer `doff` del cárdigan sobre la silla y `don` al levantarse. |

---

## 5. Contenido inicial: catálogo de prendas

### 5.1. Pantuflas (`slipper_left`, `slipper_right`) — Vertical slice principal
- **Geometría**: `Space` de `10 cm × 24 cm` (`Extent: [-5 cm, +5 cm] × [-12 cm, +12 cm]`), `nav_grid` con celdas de `1 cm` (10 × 24).
- **Zonas internas**:
  - `InnerPressure.Zone`: plantilla central `[-4 cm, +4 cm] × [-8 cm, +7 cm]`, `Force: 50`, `Source: "garment"`.
  - **Zonas seguras**: fondo de la puntera (`y > 7 cm`, con `GripPoint{Label: "puntera"}`) y borde del talón (`y < -8 cm`, con `GripPoint{Label: "borde del talón"}`).
- **`BodyZone`**: `Occluded: true` sobre la puntera/empeine para que el renderer recorte la tela superior alrededor del tiny seguido.
- **Conexiones**:
  - `placed`: `contact_surface` con `Scale: "tiny"`, borde de acceso en el talón (`y = -11 cm`) abierto hacia el piso.
  - `worn`: montada en `body_foot_l` / `body_foot_r`, `BodyPortal` abierto hacia el tobillo (`body_ankle_l` / `body_ankle_r`).
- **Atractivo para la colonia**: un `search_spot` interior (fibras sueltas o botín menor) o curiosidad natural por ser un `Space` nuevo con `landmark` cuando queda apoyada cerca del futón o en el baño.

### 5.2. Cárdigan y bolsillo (`cardigan_pocket`) — Expansión
- **Geometría**: `Space` de bolsillo `12 cm × 14 cm`, `Occluded: true`, `InnerPressure: nil` (refugio móvil seguro sin compresión al ponérselo).
- **Conexiones**: `placed` sobre la silla o el futón; `worn` montado en el torso con `BodyPortal` hacia el costado/pecho y `GripPoint{Label: "borde del bolsillo"}`.

---

## 6. Presentación, arte y sonido

Todo es presentación pura derivada de `view.Frame` y `Runner.Events` (nunca altera el hash).

### 6.1. `view` y `cmd/game/hud.go`
- `view.Space` suma `GarmentState string` (opcional) para distinguir visualmente una prenda apoyada de una en transición o puesta.
- `view.Frame.Threats` incluye automáticamente la `InnerPressure.Zone` cuando `State == "donning"` (se dibuja como sombra creciente dentro de la pantufla en micro).
- **Rótulo jerárquico**:
  - Apoyada: `«Dormitorio › Pantufla izquierda»` o `«Baño › Pantufla izquierda»`.
  - Puesta: `«Giantess › Pie izquierdo › Pantufla izquierda»`.
- **Crónica (`view.Chronicle`)**:
  - Hito cuando un tiny viaja dentro de una prenda a otra habitación: *«Tiny 04 viajó en la pantufla izquierda hasta el baño.»*
  - Hito de escape: *«Tiny 02 salió de la pantufla izquierda justo antes de que se la pusiera.»*
  - Hito de aplastamiento: *«Tiny 07 quedó dentro de la pantufla izquierda.»*

### 6.2. Arte (`tools/art/`)
- **Macro (80 px/m)**:
  - `slipper/macro_idle.png`: pantufla apoyada en el piso (solo visible cuando `State == "placed"` o `"donning"`; oculta o acoplada al sprite de la giantess cuando `State == "worn"`).
  - En `giantess.py`: opcionalmente variante de pies descalzos cuando no lleva pantuflas (o cambio sutil de color de pie).
- **Micro (1600 px/m)**:
  - `slipper/micro_idle.png`: vista superior del interior acolchado, borde del talón y cubierta del empeine (capa frontal con recorte `Occluded` alrededor del tiny seguido, reutilizando el cutaway de `render/pixels.go`).
  - Animación o sombra de pie entrando durante `donning`.

### 6.3. Sonido (`tools/audio/sounds.json`, `internal/sound`)
- `slipper_slide`: sonido suave de tela/calce al emitirse `GarmentDonning` / `GarmentWorn` (más grave y envolvente en vista micro).
- `garment_drop`: golpe blando al emitirse `GarmentDoffed`.
- `garment_rustle`: roce de tela al ponerse/sacarse el cárdigan.

---

## 7. Determinismo y versionado

- **D1**: todas las posiciones, `InnerPressure.Zone` y offsets en `units.Length` (µm).
- **D3**: `Owner`, `MountPart`, `DoffSpace` y `BodyPortals` son `ids.ID` generacionales verificados con `Alive` antes de usarse.
- **D4**: la tirada de atención al calzarse usa `w.RNG("ai.attention")`.
- **D5**: `attach.garments` y la invariante iteran `ecs.Get[Garment](r).IDs()` en orden ascendente.
- **D7**: ningún sistema fuera de `World.Step()` muta `Garment`.
- **Versiones**: al registrar el store `"garment"` en `engine.New`, incrementar `world.HashVersion` (`15 → 16`) y `world.SnapshotSchema` (`18 → 19`).

---

## 8. Plan de implementación en vertical slices

Cada fase cierra con `gofmt -l .`, `go vet ./...`, `go test ./...` y tests de continuidad desde snapshot.

| Fase | Alcance | Escenarios y tests de cierre |
| --- | --- | --- |
| **P1 — Núcleo `garment` y presión interior** | Store `"garment"`, comandos `don_garment`/`cancel_don`/`doff_garment`, sistema `attach.garments`, invariante `attach.garment`, amenaza y resolución en `crush`, validación en `attach.PortalCrossingValid`, bump de `HashVersion`/`SnapshotSchema`. | `garment_ride_and_doff` (tiny sube a pantufla apoyada, se calza, camina a otro cuarto, se la saca y el tiny baja), `garment_donning_flee_and_crush` (tiny en plantilla huye a puntera segura o muere si no llega), test de continuidad desde snapshot en medio de `donning` y en viaje `worn`. |
| **P2 — Rutinas IA (`don` / `doff`) y cuidado** | Pasos `StepDon`/`StepDoff` en `sim/ai`, tirada de atención con `agent.care`, cancelación limpia ante urgencia, integración en `bed` (`sleep`) y `shower_tray` (`shower`). | `giantess_slipper_routine` (duerme sin pantuflas, se las calza, va al baño, se las saca para ducharse), `careful_giantess_checks_slipper` (con `care` alto espera y no aplasta). |
| **P3 — Casa, arte micro/macro, crónica y audio** | Generador `tools/scenarios/prendas.py` (idempotente sobre `casa.json`), sprites macro/micro en `tools/art/` (`go run ./cmd/artpack`), rótulo jerárquico, crónica de viajes en prenda y sonidos en `tools/audio/sounds.json`. | Captura de referencia micro dentro de la pantufla (con cutaway y sombra de aviso), actualización de `TestReferenceCaptures` y `TestReferenceMix`. |
| **P4 — Cárdigan con bolsillo móvil** *(opcional)* | Segunda prenda sobre la silla del escritorio y torso, con bolsillo seguro (`InnerPressure: nil`) y conexión `BodyPortal` al pecho/hombro. | `cardigan_pocket_ride` (tiny entra al bolsillo en la silla, la giantess se lo pone y lo lleva al living/baño, el tiny sale al hombro o baja al sacárselo). |
