"""Planta del departamento y catálogo paramétrico de muebles (propuesta 20).

Fuente única de geometría para:
- la generación de content/partidas/casa.json (tools/scenarios/partida.py);
- el arte de arquitectura (tools/art/apartment.py);
- las pruebas de edición y circulación (20.4.1 y 20.4.2).

Todas las medidas autoritativas se convierten a units.Length (µm, int).
1 cm = 10_000 µm, 1 m = 1_000_000 µm.
Origen en la esquina sudoeste (0, 0); X crece al este, Y crece al norte.
"""
import copy

CM = 10_000

# ---------------------------------------------------------------------------
# 20.1 Planta del departamento (7,4 × 6,4 m)
# ---------------------------------------------------------------------------

APARTMENT_W_CM = 740
APARTMENT_H_CM = 640

# Cada habitación define su origen (x, y, z) en cm, su tamaño (w, h) en cm,
# el material de piso, ventanas exteriores y vanos compartidos.
ROOMS = {
    'bedroom': {
        'ref': 'bedroom',
        'name': 'dormitorio',
        'visual': 'apt_bedroom',
        'origin_cm': (0, 310, 0),
        'size_cm': (320, 330),
        'floor': 'tatami',
        'wall_tint': ('#EBD7C9', '#F5E6DA', '#DFC8B9', '#B98A6E', '#8F6650'),
        'windows': [
            {'wall': 'north', 'pos_cm': 200, 'width_cm': 90, 'kind': 'shoji'},
        ],
    },
    'living': {
        'ref': 'living',
        'name': 'living',
        'visual': 'apt_living',
        'origin_cm': (320, 310, 0),
        'size_cm': (420, 330),
        'floor': 'wood',
        'wall_tint': ('#E8D3D3', '#F4E4E2', '#DCC4C6', '#B98A6E', '#8F6650'),
        'windows': [
            {'wall': 'north', 'pos_cm': 40, 'width_cm': 75, 'kind': 'glass'},
            {'wall': 'north', 'pos_cm': 275, 'width_cm': 95, 'kind': 'glass'},
        ],
    },
    'kitchen': {
        'ref': 'kitchen',
        'name': 'cocina',
        'visual': 'apt_kitchen',
        'origin_cm': (440, 0, 0),
        'size_cm': (300, 310),
        'floor': 'kitchen_tile',
        'wall_tint': ('#D8EAE4', '#E8F4F0', '#C9DED8', '#9DBDB5', '#789B94'),
        'windows': [
            {'wall': 'east', 'pos_cm': 140, 'width_cm': 90, 'kind': 'glass'},
        ],
    },
    'hallway': {
        'ref': 'hallway',
        'name': 'distribuidor',
        'visual': 'apt_hallway',
        'origin_cm': (0, 200, 0),
        'size_cm': (440, 110),
        'floor': 'wood',
        'wall_tint': ('#EDE0D6', '#F6ECE4', '#E0D1C5', '#B98A6E', '#8F6650'),
        'windows': [],
    },
    'bathroom': {
        'ref': 'bathroom',
        'name': 'baño',
        'visual': 'apt_bathroom',
        'origin_cm': (150, 0, 0),
        'size_cm': (290, 200),
        'floor': 'bath_tile',
        'wall_tint': ('#D7E8EE', '#E8F4F8', '#C5DCE4', '#8EB0BD', '#6D8E9B'),
        'windows': [],
    },
    'genkan': {
        'ref': 'genkan',
        'name': 'entrada',
        'visual': 'apt_genkan',
        'origin_cm': (0, 0, -8),
        'size_cm': (150, 200),
        'floor': 'genkan_tile',
        'wall_tint': ('#E6DCD4', '#F2EAE3', '#D9CDC3', '#9D8A7E', '#78685E'),
        'windows': [],
    },
}

# Vanos de circulación (coordenadas de mundo en cm). Compartidos entre el
# generador de escenarios y el arte de arquitectura (propuesta 20.3.1).
DOORWAYS = [
    {
        'ref': 'genkan_step',
        'name': 'escalón de la entrada',
        'room_a': 'genkan',
        'room_b': 'hallway',
        'axis': 'horizontal',
        'y_cm': 200,
        'span_cm': (45, 135),  # 90 cm libres
        'a_local_cm': (90, 195),
        'b_local_cm': (90, 5),
        'scale': 'giantess',
        'visual': None,
    },
    {
        'ref': 'door_bathroom',
        'name': 'puerta del baño',
        'room_a': 'hallway',
        'room_b': 'bathroom',
        'axis': 'horizontal',
        'y_cm': 200,
        'span_cm': (180, 260),  # 80 cm libres
        'a_local_cm': (220, 5),
        'b_local_cm': (70, 195),
        'scale': '',
        'visual': 'door_bath',
    },
    {
        'ref': 'door_bedroom',
        'name': 'fusuma del dormitorio',
        'room_a': 'hallway',
        'room_b': 'bedroom',
        'axis': 'horizontal',
        'y_cm': 310,
        'span_cm': (170, 260),  # 90 cm libres
        'a_local_cm': (215, 105),
        'b_local_cm': (215, 5),
        'scale': '',
        'visual': 'door_h',
    },
    {
        'ref': 'door_living',
        'name': 'paso al living',
        'room_a': 'hallway',
        'room_b': 'living',
        'axis': 'horizontal',
        'y_cm': 310,
        'span_cm': (335, 425),  # 90 cm libres
        'a_local_cm': (380, 105),
        'b_local_cm': (60, 5),
        'scale': '',
        'visual': None,
    },
    {
        'ref': 'door_kitchen',
        'name': 'noren de la cocina',
        'room_a': 'living',
        'room_b': 'kitchen',
        'axis': 'horizontal',
        'y_cm': 310,
        'span_cm': (590, 680),  # 90 cm libres
        'a_local_cm': (315, 5),
        'b_local_cm': (195, 305),
        'scale': '',
        'visual': 'door_noren_h',
    },
]

# Puerta exterior en la pared sur del genkan (solo arquitectura).
EXTERIOR_DOOR_SPAN_CM = (45, 125)

# Pasacables del zócalo entre dormitorio y living (coordenada Y de mundo en cm).
CABLE_DUCT_WORLD_Y_CM = 542

# ---------------------------------------------------------------------------
# 20.2 y 20.3.3 Catálogo de piezas de mobiliario (21 piezas)
# ---------------------------------------------------------------------------

PIECES = {
    'desk': {
        'width_cm': 60, 'depth_cm': 40, 'height_cm': 75,
        'blocks_floor': True, 'walkable': True, 'visual': 'desk',
        'pivot': (6, 34), 'occludes': ('north', 'east', 'west'),
        'cell_cm': 2, 'under_clearance_cm': 72,
    },
    'chair': {
        'width_cm': 40, 'depth_cm': 40, 'height_cm': 45,
        'blocks_floor': False, 'walkable': True, 'visual': 'chair',
        'pivot': (4, 36), 'occludes': ('south',),
        'cell_cm': 2, 'under_clearance_cm': 42,
    },
    'futon': {
        'width_cm': 120, 'depth_cm': 180, 'height_cm': 15,
        'blocks_floor': False, 'walkable': True, 'visual': 'apt_futon',
        'pivot': (4, 148), 'occludes': (),
        'cell_cm': 10,
    },
    'pillow': {
        'width_cm': 60, 'depth_cm': 40, 'height_cm': 25,
        'blocks_floor': False, 'walkable': True, 'visual': 'apt_pillow',
        'pivot': (2, 34), 'occludes': (),
        'cell_cm': 5,
    },
    'low_closet': {
        'width_cm': 120, 'depth_cm': 50, 'height_cm': 90,
        'blocks_floor': True, 'walkable': False, 'visual': 'low_closet',
        'pivot': (2, 42), 'occludes': ('north', 'south', 'east', 'west'),
    },
    'sofa': {
        'width_cm': 120, 'depth_cm': 50, 'height_cm': 45,
        'blocks_floor': False, 'walkable': True, 'visual': 'apt_sofa',
        'pivot': (6, 46), 'occludes': ('north', 'east', 'west'),
        'cell_cm': 2,
    },
    'living_table': {
        'width_cm': 60, 'depth_cm': 40, 'height_cm': 65,
        'blocks_floor': True, 'walkable': True, 'visual': 'table',
        'pivot': (6, 34), 'occludes': ('north', 'east', 'west'),
        'cell_cm': 2, 'under_clearance_cm': 62,
    },
    'tv_stand': {
        'width_cm': 110, 'depth_cm': 40, 'height_cm': 45,
        'blocks_floor': True, 'walkable': False, 'visual': 'tv_stand',
        'pivot': (3, 36), 'occludes': ('north', 'south', 'east', 'west'),
    },
    'tv': {
        'width_cm': 80, 'depth_cm': 20, 'height_cm': 105,
        'blocks_floor': False, 'walkable': False, 'visual': 'apt_tv',
        'pivot': (33, 32), 'occludes': ('south',),
    },
    'rug': {
        'width_cm': 170, 'depth_cm': 120, 'height_cm': 0,
        'blocks_floor': False, 'walkable': True, 'visual': 'rug',
        'pivot': (3, 99), 'occludes': (),
    },
    'plant': {
        'width_cm': 30, 'depth_cm': 30, 'height_cm': 70,
        'blocks_floor': True, 'walkable': False, 'visual': 'plant',
        'pivot': (23, 50), 'occludes': ('north', 'south', 'east', 'west'),
    },
    'counter': {
        'width_cm': 60, 'depth_cm': 240, 'height_cm': 85,
        'blocks_floor': True, 'walkable': True, 'visual': 'counter',
        'pivot': (4, 196), 'occludes': ('west', 'east', 'south'),
        'cell_cm': 5,
    },
    'sink': {
        'width_cm': 50, 'depth_cm': 50, 'height_cm': 85,
        'blocks_floor': False, 'walkable': False, 'visual': 'sink',
        'pivot': (2, 42), 'occludes': (),
        'basin_depth_cm': 17,
    },
    'stove': {
        'width_cm': 50, 'depth_cm': 60, 'height_cm': 85,
        'blocks_floor': False, 'walkable': False, 'visual': 'stove',
        'pivot': (2, 50), 'occludes': (),
    },
    'fridge': {
        'width_cm': 60, 'depth_cm': 65, 'height_cm': 180,
        'blocks_floor': True, 'walkable': False, 'visual': 'fridge',
        'pivot': (22, 32), 'occludes': ('north', 'south', 'east', 'west'),
    },
    'rice_cooker': {
        'width_cm': 30, 'depth_cm': 30, 'height_cm': 25,
        'blocks_floor': False, 'walkable': False, 'visual': 'rice_cooker',
        'pivot': (13, 20), 'occludes': ('north', 'south', 'east', 'west'),
    },
    'dish_shelf': {
        'width_cm': 25, 'depth_cm': 80, 'height_cm': 145,
        'blocks_floor': False, 'walkable': False, 'visual': 'dish_shelf',
        'pivot': (4, 32), 'occludes': (),
    },
    'washbasin': {
        'width_cm': 55, 'depth_cm': 70, 'height_cm': 85,
        'blocks_floor': True, 'walkable': False, 'visual': 'washbasin',
        'pivot': (2, 58), 'occludes': ('east', 'north', 'south'),
    },
    'toilet': {
        'width_cm': 50, 'depth_cm': 60, 'height_cm': 75,
        'blocks_floor': True, 'walkable': False, 'visual': 'toilet',
        'pivot': (2, 50), 'occludes': ('south', 'east', 'west'),
    },
    'shower_tray': {
        'width_cm': 95, 'depth_cm': 95, 'height_cm': 6,
        'blocks_floor': True, 'walkable': False, 'visual': 'shower_tray',
        'pivot': (2, 78), 'occludes': (),
    },
    'shoe_rack': {
        'width_cm': 30, 'depth_cm': 80, 'height_cm': 95,
        'blocks_floor': True, 'walkable': False, 'visual': 'shoe_rack',
        'pivot': (2, 66), 'occludes': ('west', 'north', 'south'),
    },
}

# ---------------------------------------------------------------------------
# Planta por defecto: ubicación de cada pieza en coordenadas locales de su
# habitación (cm, esquina SO de la huella salvo piezas centradas).
# ---------------------------------------------------------------------------

DEFAULT_LAYOUT = [
    # --- Dormitorio (320 × 330 cm) ---
    {'ref': 'futon', 'piece': 'futon', 'room': 'bedroom', 'pos_cm': (15, 140), 'facing': 1024, 'name': 'futón'},
    {'ref': 'pillow', 'piece': 'pillow', 'room': 'bedroom', 'pos_cm': (45, 270), 'facing': 1024, 'name': 'almohada'},
    {'ref': 'low_closet', 'piece': 'low_closet', 'room': 'bedroom', 'pos_cm': (10, 0), 'facing': 1024, 'name': 'placard bajo'},
    {'ref': 'desk1', 'piece': 'desk', 'room': 'bedroom', 'pos_cm': (220, 260), 'facing': 1024, 'name': 'escritorio'},
    {'ref': 'chair', 'piece': 'chair', 'room': 'bedroom', 'pos_cm': (230, 215), 'facing': 1024, 'name': 'silla'},

    # --- Living (420 × 330 cm) ---
    {'ref': 'rug', 'piece': 'rug', 'room': 'living', 'pos_cm': (105, 155), 'facing': 0, 'name': 'alfombra'},
    {'ref': 'sofa', 'piece': 'sofa', 'room': 'living', 'pos_cm': (130, 265), 'facing': 3072, 'name': 'sofá'},
    {'ref': 'desk2', 'piece': 'living_table', 'room': 'living', 'pos_cm': (160, 210), 'facing': 3072, 'name': 'mesa del living'},
    {'ref': 'tv_stand', 'piece': 'tv_stand', 'room': 'living', 'pos_cm': (135, 5), 'facing': 1024, 'name': 'mueble de TV'},
    {'ref': 'tv', 'piece': 'tv', 'room': 'living', 'pos_cm': (190, 25), 'facing': 1024, 'name': 'televisor'},
    {'ref': 'plant', 'piece': 'plant', 'room': 'living', 'pos_cm': (385, 285), 'facing': 0, 'name': 'planta'},

    # --- Cocina (300 × 310 cm) ---
    {'ref': 'counter', 'piece': 'counter', 'room': 'kitchen', 'pos_cm': (0, 20), 'facing': 0, 'name': 'mesada'},
    {'ref': 'sink', 'piece': 'sink', 'room': 'kitchen', 'pos_cm': (5, 30), 'facing': 0, 'name': 'bacha'},
    {'ref': 'stove', 'piece': 'stove', 'room': 'kitchen', 'pos_cm': (5, 185), 'facing': 0, 'name': 'anafe'},
    {'ref': 'rice_cooker', 'piece': 'rice_cooker', 'room': 'kitchen', 'pos_cm': (20, 110), 'facing': 0, 'name': 'arrocera'},
    {'ref': 'dish_shelf', 'piece': 'dish_shelf', 'room': 'kitchen', 'pos_cm': (0, 100), 'facing': 0, 'name': 'repisa de vajilla'},
    {'ref': 'fridge', 'piece': 'fridge', 'room': 'kitchen', 'pos_cm': (260, 85), 'facing': 2048, 'name': 'heladera'},

    # --- Baño (290 × 200 cm) ---
    {'ref': 'toilet', 'piece': 'toilet', 'room': 'bathroom', 'pos_cm': (40, 10), 'facing': 1024, 'name': 'inodoro'},
    {'ref': 'washbasin', 'piece': 'washbasin', 'room': 'bathroom', 'pos_cm': (235, 120), 'facing': 2048, 'name': 'lavamanos'},
    {'ref': 'shower_tray', 'piece': 'shower_tray', 'room': 'bathroom', 'pos_cm': (195, 5), 'facing': 2048, 'name': 'plato de ducha'},

    # --- Entrada / genkan (150 × 200 cm) ---
    {'ref': 'shoe_rack', 'piece': 'shoe_rack', 'room': 'genkan', 'pos_cm': (5, 85), 'facing': 0, 'name': 'zapatero'},
]


def _make_grid(w_cm, h_cm, cell_cm, blocked_rects_cm=()):
    """Construye un nav_grid bloqueando los rectángulos (x0, y0, x1, y1) en cm."""
    cw = max(1, w_cm // cell_cm)
    ch = max(1, h_cm // cell_cm)
    rows = []
    for r in range(ch):
        cy0, cy1 = r * cell_cm, (r + 1) * cell_cm
        line = []
        for c in range(cw):
            cx0, cx1 = c * cell_cm, (c + 1) * cell_cm
            blocked = False
            for bx0, by0, bx1, by1 in blocked_rects_cm:
                if cx0 < bx1 and cx1 > bx0 and cy0 < by1 and cy1 > by0:
                    blocked = True
                    break
            line.append('#' if blocked else '.')
        rows.append(''.join(line))
    return {
        'origin': {'x': 0, 'y': 0},
        'cell': cell_cm * CM,
        'version': 0,
        'rows': rows,
    }


def _rect_um(w_cm, h_cm):
    return {'min': {'x': 0, 'y': 0}, 'max': {'x': w_cm * CM, 'y': h_cm * CM}}


def _portal(ref, name, sp_a, pa_cm, sp_b, pb_cm, cost=0, scale='', visual=None):
    p = {
        'a': {'space': sp_a, 'pos': {'x': int(round(pa_cm[0] * CM)), 'y': int(round(pa_cm[1] * CM))}},
        'b': {'space': sp_b, 'pos': {'x': int(round(pb_cm[0] * CM)), 'y': int(round(pb_cm[1] * CM))}},
        'cost': cost,
        'version': 0,
    }
    if scale:
        p['scale'] = scale
    ent = {'ref': ref, 'spawn': {'name': name, 'components': {'portal': p}}}
    if visual:
        ent['visual'] = visual
    return ent


def _climb_strip(ref, name, parent_ref, pos_cm, height_cm, z_cm=0):
    cell_cm = 2
    steps = max(1, (height_cm + cell_cm - 1) // cell_cm)
    return {
        'ref': ref,
        'visual': 'climb',
        'spawn': {
            'name': name,
            'space': {'kind': 'elevated', 'extent': _rect_um(cell_cm, steps * cell_cm)},
            'location': {'space': parent_ref, 'pos': {'x': pos_cm[0] * CM, 'y': pos_cm[1] * CM, 'z': z_cm * CM}},
            'components': {'nav_grid': _make_grid(cell_cm, steps * cell_cm, cell_cm)},
        },
    }


def _find_free_adjacent_cm(room_w, room_h, blocked_rects, px, py, dx=-5, dy=-5):
    """Devuelve un punto de acceso transitable junto a (px, py) dentro de la habitación."""
    candidates = [
        (px + dx, py + dy),
        (px - 5, py + 5),
        (px + 5, py - 5),
        (px + 5, py + 5),
        (px - 10, py),
        (px, py - 10),
        (px + 10, py),
        (px, py + 10),
    ]
    for cx, cy in candidates:
        if cx < 5 or cx >= room_w - 5 or cy < 5 or cy >= room_h - 5:
            continue
        cell_x0, cell_y0 = (cx // 10) * 10, (cy // 10) * 10
        hit = False
        for bx0, by0, bx1, by1 in blocked_rects:
            if cell_x0 < bx1 and cell_x0 + 10 > bx0 and cell_y0 < by1 and cell_y0 + 10 > by0:
                hit = True
                break
        if not hit:
            return (cx, cy)
    return (max(5, min(room_w - 5, px + dx)), max(5, min(room_h - 5, py + dy)))


def _rects_close_cm(a, b, max_gap_cm):
    """True si dos rectángulos (x0, y0, x1, y1) están a <= max_gap_cm en ambos ejes."""
    gx = max(0, max(a[0] - b[2], b[0] - a[2]))
    gy = max(0, max(a[1] - b[3], b[1] - a[3]))
    return gx <= max_gap_cm and gy <= max_gap_cm


def build_apartment(layout=None):
    """Genera las entidades de arquitectura, vanos y muebles del departamento.

    Acepta un layout editado (lista de colocaciones como DEFAULT_LAYOUT) para
    verificar recolocaciones: actualiza automáticamente el bloqueo del piso,
    las superficies hijas y las conexiones entre muebles (desactivando puentes
    que quedaron sin contacto).

    Devuelve (entities, meta).
    """
    items = copy.deepcopy(DEFAULT_LAYOUT if layout is None else layout)
    by_ref = {it['ref']: it for it in items}

    # 1. Recolectar huellas sólidas por habitación.
    blocked_by_room = {r: [] for r in ROOMS}
    for it in items:
        spec = PIECES[it['piece']]
        if not spec.get('blocks_floor'):
            continue
        x, y = it['pos_cm']
        w, d = spec['width_cm'], spec['depth_cm']
        if it['piece'] == 'fridge':
            # fridge usa pivote centrado en su cuerpo (35 cm radio ≈ 60×65 cm).
            blocked_by_room[it['room']].append((x - w // 2, y - d // 2, x + w // 2, y + d // 2))
        elif it['piece'] == 'plant':
            blocked_by_room[it['room']].append((x - w // 2, y - d // 2, x + w // 2, y + d // 2))
        else:
            blocked_by_room[it['room']].append((x, y, x + w, y + d))

    entities = [
        {'ref': 'house', 'spawn': {
            'name': 'house',
            'space': {'kind': 'static', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 0, 'y': 0}}},
        }},
    ]

    # 2. Habitaciones con su nav_grid y contact_surface.
    for key in ('bedroom', 'living', 'kitchen', 'hallway', 'bathroom', 'genkan'):
        r = ROOMS[key]
        ox, oy, oz = r['origin_cm']
        w, h = r['size_cm']
        entities.append({
            'ref': r['ref'],
            'visual': r['visual'],
            'spawn': {
                'name': r['name'],
                'space': {'kind': 'static', 'extent': _rect_um(w, h)},
                'location': {'space': '$house', 'pos': {'x': ox * CM, 'y': oy * CM, 'z': oz * CM}},
                'components': {
                    'nav_grid': _make_grid(w, h, 10, blocked_by_room[key]),
                    'contact_surface': {'tolerance': 50000},
                },
            },
        })

    # 3. Vanos de circulación entre habitaciones y acceso trepable del genkan.
    for dw in DOORWAYS:
        entities.append(_portal(
            dw['ref'], dw['name'],
            '$' + dw['room_a'], dw['a_local_cm'],
            '$' + dw['room_b'], dw['b_local_cm'],
            cost=0, scale=dw['scale'], visual=dw['visual'],
        ))

    # Desnivel de 8 cm del genkan: acceso trepable lateral solo para tinies.
    entities.append(_climb_strip('genkan_climb', 'lateral del escalón', '$genkan', (45, 196), 8, z_cm=0))
    entities.append(_portal('genkan_climb_bottom', 'base del escalón', '$genkan', (45, 190), '$genkan_climb', (1, 1), scale='tiny'))
    entities.append(_portal('genkan_climb_top', 'borde del escalón', '$genkan_climb', (1, 7), '$hallway', (45, 5), scale='tiny'))

    # 4. Piezas de mobiliario, superficies hijas y conexiones.
    meta = {
        'rooms': ROOMS,
        'blocked_by_room': blocked_by_room,
        'by_ref': by_ref,
        'bridges': {},
    }

    # --- Escritorio (desk1) ---
    if 'desk1' in by_ref:
        it = by_ref['desk1']
        dx, dy = it['pos_cm']
        spec = PIECES['desk']
        room = ROOMS[it['room']]
        desk_ent = {
            'ref': 'desk1',
            'object': 'desk',
            'visual': 'desk',
            'spawn': {
                'name': it['name'],
                'location': {'space': '$' + it['room'], 'pos': {'x': dx * CM, 'y': dy * CM, 'z': spec['height_cm'] * CM}},
            },
        }
        entities.append(desk_ent)
        # Pata del escritorio (acceso desde el piso del dormitorio).
        leg_floor = _find_free_adjacent_cm(room['size_cm'][0], room['size_cm'][1], blocked_by_room[it['room']], dx, dy, -5, -5)
        entities.append(_climb_strip('leg1', 'pata del escritorio', '$' + it['room'], (dx, dy), spec['height_cm']))
        entities.append(_portal('leg1_bottom', 'base de la pata del escritorio', '$' + it['room'], leg_floor, '$leg1', (1, 1)))
        entities.append(_portal('leg1_top', 'cima de la pata del escritorio', '$leg1', (1, spec['height_cm'] - 1), '$desk1', (1, 1)))
        meta['leg1_floor_cm'] = leg_floor

        # Hueco bajo el escritorio: espacio local exclusivo de tinies (20.2).
        entities.append({
            'ref': 'under_desk',
            'spawn': {
                'name': 'hueco bajo el escritorio',
                'space': {'kind': 'elevated', 'extent': _rect_um(spec['width_cm'], spec['depth_cm'])},
                'location': {'space': '$' + it['room'], 'pos': {'x': dx * CM, 'y': dy * CM, 'z': 0}},
                'components': {
                    'nav_grid': _make_grid(spec['width_cm'], spec['depth_cm'], 5),
                },
            },
        })
        entities.append(_portal('under_desk_entry', 'entrada bajo el escritorio', '$' + it['room'], leg_floor, '$under_desk', (5, 5), scale='tiny'))

    # --- Silla (chair) ---
    if 'chair' in by_ref:
        it = by_ref['chair']
        cx, cy = it['pos_cm']
        spec = PIECES['chair']
        room = ROOMS[it['room']]
        entities.append({
            'ref': 'chair',
            'object': 'chair',
            'visual': 'chair',
            'spawn': {
                'name': it['name'],
                'location': {'space': '$' + it['room'], 'pos': {'x': cx * CM, 'y': cy * CM, 'z': spec['height_cm'] * CM}},
                'components': {'contact_surface': {'tolerance': 50000}},
            },
        })
        chair_floor = _find_free_adjacent_cm(room['size_cm'][0], room['size_cm'][1], blocked_by_room[it['room']], cx, cy, -5, -5)
        entities.append(_climb_strip('chair_leg', 'pata de la silla', '$' + it['room'], (cx, cy), spec['height_cm']))
        entities.append(_portal('chair_leg_bottom', 'base de la pata de la silla', '$' + it['room'], chair_floor, '$chair_leg', (1, 1)))
        entities.append(_portal('chair_leg_top', 'cima de la pata de la silla', '$chair_leg', (1, spec['height_cm'] - 1), '$chair', (1, 1)))
        meta['chair_floor_cm'] = chair_floor

        # Puente respaldo de la silla <-> escritorio solo si están en contacto (<= 15 cm).
        if 'desk1' in by_ref and by_ref['desk1']['room'] == it['room']:
            d_it = by_ref['desk1']
            dx, dy = d_it['pos_cm']
            d_rect = (dx, dy, dx + PIECES['desk']['width_cm'], dy + PIECES['desk']['depth_cm'])
            c_rect = (cx, cy, cx + spec['width_cm'], cy + spec['depth_cm'])
            if _rects_close_cm(d_rect, c_rect, 15):
                # Punto de contacto sobre el borde compartido.
                ov_x0 = max(d_rect[0], c_rect[0])
                ov_x1 = min(d_rect[2], c_rect[2])
                bx = (ov_x0 + ov_x1) // 2 if ov_x1 > ov_x0 else cx + 5
                by = min(cy + spec['depth_cm'], dy)
                desk_local_x = max(2, min(PIECES['desk']['width_cm'] - 2, bx - dx))
                chair_local_x = max(2, min(spec['width_cm'] - 2, bx - cx))
                bridge_h = PIECES['desk']['height_cm'] - spec['height_cm']
                entities.append(_climb_strip('chair_bridge', 'respaldo de la silla', '$' + it['room'], (bx, by), bridge_h, z_cm=spec['height_cm']))
                entities.append(_portal('chair_bridge_top', 'borde del escritorio (silla)', '$desk1', (desk_local_x, 1), '$chair_bridge', (1, bridge_h - 1)))
                entities.append(_portal('chair_bridge_bottom', 'borde del respaldo', '$chair_bridge', (1, 1), '$chair', (chair_local_x, spec['depth_cm'] - 1)))
                meta['bridges']['chair_desk'] = True
            else:
                meta['bridges']['chair_desk'] = False

    # --- Futón, almohada y punto de sueño (bed) ---
    if 'futon' in by_ref:
        it = by_ref['futon']
        fx, fy = it['pos_cm']
        spec = PIECES['futon']
        room = ROOMS[it['room']]
        entities.append({
            'ref': 'futon',
            'visual': 'apt_futon',
            'spawn': {
                'name': it['name'],
                'space': {'kind': 'elevated', 'extent': _rect_um(spec['width_cm'], spec['depth_cm'])},
                'location': {'space': '$' + it['room'], 'pos': {'x': fx * CM, 'y': fy * CM, 'z': spec['height_cm'] * CM}},
                'components': {
                    'nav_grid': _make_grid(spec['width_cm'], spec['depth_cm'], spec['cell_cm']),
                    'contact_surface': {'tolerance': 50000},
                },
            },
        })
        # Subida al futón por el borde este (libre hacia el dormitorio).
        step_x = min(room['size_cm'][0] - 5, fx + spec['width_cm'] + 5)
        entities.append(_portal('futon_step', 'borde del futón', '$' + it['room'], (step_x, fy + 10), '$futon', (spec['width_cm'] - 5, 10), cost=20000, scale='tiny'))
        # Ancla de la oferta sleep (la giantess se recuesta centrada sobre el futón).
        entities.append({
            'ref': 'bed',
            'object': 'bed',
            'spawn': {
                'name': 'futón',
                'location': {'space': '$' + it['room'], 'pos': {'x': (fx + spec['width_cm'] // 2) * CM, 'y': (fy + 90) * CM, 'z': 0}},
                'components': {'offers': [{
                    'action': 'sleep', 'scale': 'giantess',
                    'effects': [{'need': 'energy', 'amount': 800}],
                    'steps': [
                        {'do': 'goto'},
                        {'do': 'pose', 'pose': 'lie_down'},
                        {'do': 'use', 'ticks': 600},
                    ],
                }]},
            },
        })

    if 'pillow' in by_ref:
        it = by_ref['pillow']
        px, py = it['pos_cm']
        spec = PIECES['pillow']
        entities.append({
            'ref': 'pillow',
            'visual': 'apt_pillow',
            'spawn': {
                'name': it['name'],
                'space': {'kind': 'elevated', 'extent': _rect_um(spec['width_cm'], spec['depth_cm'])},
                'location': {'space': '$' + it['room'], 'pos': {'x': px * CM, 'y': py * CM, 'z': spec['height_cm'] * CM}},
                'components': {
                    'nav_grid': _make_grid(spec['width_cm'], spec['depth_cm'], spec['cell_cm']),
                    'contact_surface': {'tolerance': 50000},
                },
            },
        })
        if 'futon' in by_ref:
            fx, fy = by_ref['futon']['pos_cm']
            futon_local_x = max(5, min(PIECES['futon']['width_cm'] - 5, (px + spec['width_cm'] // 2) - fx))
            futon_local_y = max(5, min(PIECES['futon']['depth_cm'] - 5, py - fy))
            entities.append(_portal('pillow_step', 'costura de la almohada', '$futon', (futon_local_x, futon_local_y), '$pillow', (spec['width_cm'] // 2, 5), cost=20000, scale='tiny'))

    # --- Placard bajo (low_closet) ---
    if 'low_closet' in by_ref:
        it = by_ref['low_closet']
        x, y = it['pos_cm']
        entities.append({
            'ref': 'low_closet',
            'visual': 'low_closet',
            'spawn': {
                'name': it['name'],
                'location': {'space': '$' + it['room'], 'pos': {'x': x * CM, 'y': y * CM, 'z': 0}},
            },
        })

    # --- Alfombra del living (rug) ---
    if 'rug' in by_ref:
        it = by_ref['rug']
        x, y = it['pos_cm']
        entities.append({
            'ref': 'rug',
            'visual': 'rug',
            'spawn': {
                'name': it['name'],
                'location': {'space': '$' + it['room'], 'pos': {'x': x * CM, 'y': y * CM, 'z': 0}},
            },
        })

    # --- Sofá (sofa) ---
    if 'sofa' in by_ref:
        it = by_ref['sofa']
        sx, sy = it['pos_cm']
        spec = PIECES['sofa']
        room = ROOMS[it['room']]
        sofa_ent = {
            'ref': 'sofa',
            'object': 'sofa',
            'visual': 'apt_sofa',
            'spawn': {
                'name': it['name'],
                'location': {'space': '$' + it['room'], 'pos': {'x': sx * CM, 'y': sy * CM, 'z': spec['height_cm'] * CM}},
                'components': {
                    'contact_surface': {'tolerance': 50000},
                    'offers': [{
                        'action': 'relax', 'scale': 'giantess',
                        'effects': [{'need': 'energy', 'amount': 250}, {'need': 'fun', 'amount': 150}],
                        'approach': {'x': (spec['width_cm'] // 2) * CM, 'y': (spec['depth_cm'] // 2) * CM},
                        'facing': it['facing'],
                        'steps': [
                            {'do': 'goto'},
                            {'do': 'pose', 'pose': 'sit_hover'},
                            {'do': 'wait', 'ticks': 16},
                            {'do': 'pose', 'pose': 'sit'},
                            {'do': 'use', 'ticks': 250},
                        ],
                    }],
                },
            },
        }
        entities.append(sofa_ent)
        sofa_floor = _find_free_adjacent_cm(room['size_cm'][0], room['size_cm'][1], blocked_by_room[it['room']], sx, sy, -5, -5)
        entities.append(_climb_strip('sofa_leg', 'pata del sofá', '$' + it['room'], (sx, sy), spec['height_cm']))
        entities.append(_portal('sofa_leg_bottom', 'base de la pata del sofá', '$' + it['room'], sofa_floor, '$sofa_leg', (1, 1)))
        entities.append(_portal('sofa_leg_top', 'cima de la pata del sofá', '$sofa_leg', (1, spec['height_cm'] - 1), '$sofa', (1, 1)))
        meta['sofa_floor_cm'] = sofa_floor

    # --- Mesa auxiliar del living (desk2) ---
    if 'desk2' in by_ref:
        it = by_ref['desk2']
        tx, ty = it['pos_cm']
        spec = PIECES['living_table']
        room = ROOMS[it['room']]
        # Punto de uso desde el sofá (si hay sofá en el mismo ambiente, la giantess
        # se sienta en el centro del sofá mirando hacia la mesa).
        if 'sofa' in by_ref and by_ref['sofa']['room'] == it['room']:
            sx, sy = by_ref['sofa']['pos_cm']
            spot_x = (sx + PIECES['sofa']['width_cm'] // 2) - tx
            spot_y = (sy + PIECES['sofa']['depth_cm'] // 2) - ty
            spot_facing = by_ref['sofa']['facing']
        else:
            spot_x, spot_y, spot_facing = spec['width_cm'] // 2, spec['depth_cm'] + 30, 3072
        entities.append({
            'ref': 'desk2',
            'object': 'desk',
            'visual': 'table',
            'spawn': {
                'name': it['name'],
                'location': {'space': '$' + it['room'], 'pos': {'x': tx * CM, 'y': ty * CM, 'z': spec['height_cm'] * CM}},
                'components': {
                    'place': {'tags': ['living_table'], 'spot': {'x': spot_x * CM, 'y': spot_y * CM}, 'facing': spot_facing},
                    'offers': [],
                },
            },
        })
        table_floor = _find_free_adjacent_cm(room['size_cm'][0], room['size_cm'][1], blocked_by_room[it['room']], tx, ty, -5, -5)
        entities.append(_climb_strip('desk2_leg', 'pata de la mesa del living', '$' + it['room'], (tx, ty), spec['height_cm']))
        entities.append(_portal('desk2_leg_bottom', 'base de la pata de la mesa', '$' + it['room'], table_floor, '$desk2_leg', (1, 1)))
        entities.append(_portal('desk2_leg_top', 'cima de la pata de la mesa', '$desk2_leg', (1, spec['height_cm'] - 1), '$desk2', (1, 1)))

        # Puente sofá <-> mesa auxiliar si están a <= 25 cm.
        if 'sofa' in by_ref and by_ref['sofa']['room'] == it['room']:
            sx, sy = by_ref['sofa']['pos_cm']
            s_spec = PIECES['sofa']
            s_rect = (sx, sy, sx + s_spec['width_cm'], sy + s_spec['depth_cm'])
            t_rect = (tx, ty, tx + spec['width_cm'], ty + spec['depth_cm'])
            if _rects_close_cm(s_rect, t_rect, 25):
                ov_x0 = max(s_rect[0], t_rect[0])
                ov_x1 = min(s_rect[2], t_rect[2])
                bx = (ov_x0 + ov_x1) // 2 if ov_x1 > ov_x0 else tx + 10
                by = min(sy, ty + spec['depth_cm'])
                t_local_x = max(2, min(spec['width_cm'] - 2, bx - tx))
                s_local_x = max(2, min(s_spec['width_cm'] - 2, bx - sx))
                bridge_h = spec['height_cm'] - s_spec['height_cm']
                entities.append(_climb_strip('sofa_bridge', 'borde del sofá', '$' + it['room'], (bx, by), bridge_h, z_cm=s_spec['height_cm']))
                entities.append(_portal('sofa_bridge_top', 'borde de la mesa (sofá)', '$desk2', (t_local_x, spec['depth_cm'] - 1), '$sofa_bridge', (1, bridge_h - 1)))
                entities.append(_portal('sofa_bridge_bottom', 'asiento del sofá (mesa)', '$sofa_bridge', (1, 1), '$sofa', (s_local_x, 1)))
                meta['bridges']['sofa_table'] = True
            else:
                meta['bridges']['sofa_table'] = False

    # --- Mueble de TV y Televisor ---
    if 'tv_stand' in by_ref:
        it = by_ref['tv_stand']
        x, y = it['pos_cm']
        entities.append({
            'ref': 'tv_stand',
            'visual': 'tv_stand',
            'spawn': {
                'name': it['name'],
                'location': {'space': '$' + it['room'], 'pos': {'x': x * CM, 'y': y * CM, 'z': 0}},
            },
        })

    if 'tv' in by_ref:
        it = by_ref['tv']
        vx, vy = it['pos_cm']
        if 'sofa' in by_ref and by_ref['sofa']['room'] == it['room']:
            sx, sy = by_ref['sofa']['pos_cm']
            app_x = (sx + PIECES['sofa']['width_cm'] // 2) - vx
            app_y = (sy + PIECES['sofa']['depth_cm'] // 2) - vy
            watch_facing = by_ref['sofa']['facing']
        else:
            app_x, app_y, watch_facing = 0, 200, 3072
        entities.append({
            'ref': 'tv',
            'object': 'tv',
            'visual': 'apt_tv',
            'spawn': {
                'name': it['name'],
                'location': {'space': '$' + it['room'], 'pos': {'x': vx * CM, 'y': vy * CM, 'z': 0}},
                'components': {'offers': [{
                    'action': 'watch', 'scale': 'giantess',
                    'effects': [{'need': 'fun', 'amount': 500}],
                    'approach': {'x': app_x * CM, 'y': app_y * CM},
                    'facing': watch_facing,
                    'steps': [
                        {'do': 'goto'},
                        {'do': 'pose', 'pose': 'sit_hover'},
                        {'do': 'wait', 'ticks': 16},
                        {'do': 'pose', 'pose': 'sit'},
                        {'do': 'pose', 'pose': 'sit_hand_hover'},
                        {'do': 'wait', 'ticks': 12},
                        {'do': 'pose', 'pose': 'sit_hand_rest'},
                        {'do': 'use', 'ticks': 300},
                    ],
                }]},
            },
        })

    # --- Planta en maceta ---
    if 'plant' in by_ref:
        it = by_ref['plant']
        x, y = it['pos_cm']
        entities.append({
            'ref': 'plant',
            'visual': 'plant',
            'spawn': {
                'name': it['name'],
                'location': {'space': '$' + it['room'], 'pos': {'x': x * CM, 'y': y * CM, 'z': 0}},
            },
        })

    # --- Cocina: mesada, bacha (recorte + fondo para caídas), anafe, arrocera, repisa, heladera ---
    if 'counter' in by_ref:
        it = by_ref['counter']
        kx, ky = it['pos_cm']
        spec = PIECES['counter']
        room = ROOMS[it['room']]
        # Recortes en el tablero de la mesada (coordenadas locales a counter, en cm):
        # bacha, anafe y arrocera quedan fuera de la grilla transitable de la mesada.
        cutouts = []
        if 'sink' in by_ref:
            sx, sy = by_ref['sink']['pos_cm']
            ss = PIECES['sink']
            cutouts.append((sx - kx, sy - ky, sx - kx + ss['width_cm'], sy - ky + ss['depth_cm']))
        if 'stove' in by_ref:
            stx, sty = by_ref['stove']['pos_cm']
            sts = PIECES['stove']
            # Hornallas bloqueadas dejando un borde seco al este (x >= 50) para paso tiny.
            cutouts.append((stx - kx, sty - ky, min(spec['width_cm'] - 10, stx - kx + sts['width_cm']), sty - ky + sts['depth_cm']))
        if 'rice_cooker' in by_ref:
            rx, ry = by_ref['rice_cooker']['pos_cm']
            rs = PIECES['rice_cooker']
            cutouts.append((rx - kx - rs['width_cm'] // 2, ry - ky - rs['depth_cm'] // 2,
                            rx - kx + rs['width_cm'] // 2, ry - ky + rs['depth_cm'] // 2))

        entities.append({
            'ref': 'counter',
            'visual': 'counter',
            'spawn': {
                'name': it['name'],
                'space': {'kind': 'elevated', 'extent': _rect_um(spec['width_cm'], spec['depth_cm'])},
                'location': {'space': '$' + it['room'], 'pos': {'x': kx * CM, 'y': ky * CM, 'z': spec['height_cm'] * CM}},
                'components': {
                    'nav_grid': _make_grid(spec['width_cm'], spec['depth_cm'], spec['cell_cm'], cutouts),
                    'contact_surface': {'tolerance': 50000},
                },
            },
        })
        # Cable lateral sujeto al mueble desde el zócalo (acceso tiny a la mesada).
        cable_y = ky + spec['depth_cm']
        cable_floor = _find_free_adjacent_cm(room['size_cm'][0], room['size_cm'][1], blocked_by_room[it['room']], kx + spec['width_cm'] - 5, cable_y, 0, 10)
        entities.append(_climb_strip('counter_cable', 'cable de la mesada', '$' + it['room'], (kx + spec['width_cm'] - 5, cable_y), spec['height_cm']))
        entities.append(_portal('counter_cable_bottom', 'base del cable de la mesada', '$' + it['room'], cable_floor, '$counter_cable', (1, 1), scale='tiny'))
        entities.append(_portal('counter_cable_top', 'borde de la mesada', '$counter_cable', (1, spec['height_cm'] - 1), '$counter', (spec['width_cm'] - 5, spec['depth_cm'] - 5), scale='tiny'))

    if 'sink' in by_ref:
        it = by_ref['sink']
        sx, sy = it['pos_cm']
        ss = PIECES['sink']
        # Visual de la bacha encastrada + fondo de la bacha (apoyo ante una caída, 20.2).
        if 'counter' in by_ref:
            kx, ky = by_ref['counter']['pos_cm']
            sink_loc = {'space': '$counter', 'pos': {'x': (sx - kx) * CM, 'y': (sy - ky) * CM, 'z': 0}}
        else:
            sink_loc = {'space': '$' + it['room'], 'pos': {'x': sx * CM, 'y': sy * CM, 'z': ss['height_cm'] * CM}}
        entities.append({
            'ref': 'sink',
            'visual': 'sink',
            'spawn': {'name': it['name'], 'location': sink_loc},
        })
        basin_z = ss['height_cm'] - ss['basin_depth_cm']
        entities.append({
            'ref': 'sink_basin',
            'visual': 'sink_basin',
            'spawn': {
                'name': 'fondo de la bacha',
                'space': {'kind': 'elevated', 'extent': _rect_um(ss['width_cm'], ss['depth_cm'])},
                'location': {'space': '$' + it['room'], 'pos': {'x': sx * CM, 'y': sy * CM, 'z': basin_z * CM}},
                'components': {
                    'nav_grid': _make_grid(ss['width_cm'], ss['depth_cm'], 5),
                },
            },
        })

    if 'stove' in by_ref:
        it = by_ref['stove']
        x, y = it['pos_cm']
        if 'counter' in by_ref:
            kx, ky = by_ref['counter']['pos_cm']
            stove_loc = {'space': '$counter', 'pos': {'x': (x - kx) * CM, 'y': (y - ky) * CM, 'z': 0}}
        else:
            stove_loc = {'space': '$' + it['room'], 'pos': {'x': x * CM, 'y': y * CM, 'z': PIECES['stove']['height_cm'] * CM}}
        entities.append({
            'ref': 'stove',
            'visual': 'stove',
            'spawn': {'name': it['name'], 'location': stove_loc},
        })

    if 'rice_cooker' in by_ref:
        it = by_ref['rice_cooker']
        x, y = it['pos_cm']
        # Cuelga de la mesada si existe, si no del piso de cocina.
        if 'counter' in by_ref:
            kx, ky = by_ref['counter']['pos_cm']
            loc = {'space': '$counter', 'pos': {'x': (x - kx) * CM, 'y': (y - ky) * CM, 'z': 0}}
        else:
            loc = {'space': '$' + it['room'], 'pos': {'x': x * CM, 'y': y * CM, 'z': 0}}
        entities.append({
            'ref': 'rice_cooker',
            'visual': 'rice_cooker',
            'spawn': {'name': it['name'], 'location': loc},
        })

    if 'dish_shelf' in by_ref:
        it = by_ref['dish_shelf']
        x, y = it['pos_cm']
        if 'counter' in by_ref:
            kx, ky = by_ref['counter']['pos_cm']
            shelf_loc = {'space': '$counter', 'pos': {'x': (x - kx) * CM, 'y': (y - ky) * CM, 'z': (PIECES['dish_shelf']['height_cm'] - PIECES['counter']['height_cm']) * CM}}
        else:
            shelf_loc = {'space': '$' + it['room'], 'pos': {'x': x * CM, 'y': y * CM, 'z': PIECES['dish_shelf']['height_cm'] * CM}}
        entities.append({
            'ref': 'dish_shelf',
            'visual': 'dish_shelf',
            'spawn': {'name': it['name'], 'location': shelf_loc},
        })

    if 'fridge' in by_ref:
        it = by_ref['fridge']
        fx, fy = it['pos_cm']
        entities.append({
            'ref': 'fridge',
            'object': 'fridge',
            'visual': 'fridge',
            'spawn': {
                'name': it['name'],
                'location': {'space': '$' + it['room'], 'pos': {'x': fx * CM, 'y': fy * CM, 'z': 0}},
            },
        })

    # --- Baño: inodoro, lavamanos, plato de ducha ---
    for ref in ('toilet', 'washbasin', 'shower_tray'):
        if ref in by_ref:
            it = by_ref[ref]
            x, y = it['pos_cm']
            entities.append({
                'ref': ref,
                'visual': PIECES[it['piece']]['visual'],
                'spawn': {
                    'name': it['name'],
                    'location': {'space': '$' + it['room'], 'pos': {'x': x * CM, 'y': y * CM, 'z': 0}},
                },
            })

    # --- Entrada (genkan): zapatero ---
    if 'shoe_rack' in by_ref:
        it = by_ref['shoe_rack']
        x, y = it['pos_cm']
        entities.append({
            'ref': 'shoe_rack',
            'visual': 'shoe_rack',
            'spawn': {
                'name': it['name'],
                'location': {'space': '$' + it['room'], 'pos': {'x': x * CM, 'y': y * CM, 'z': 0}},
            },
        })

    return entities, meta


def verify_edit_positions():
    """Verifica 20.4.2: mover escritorio/mesa y silla en dos posiciones de edición.

    Comprueba que:
    1. En la posición por defecto, silla y escritorio están en contacto (chair_bridge activo)
       y sofá y mesa auxiliar están en contacto (sofa_bridge activo).
    2. Al mover la silla lejos del escritorio y la mesa auxiliar lejos del sofá:
       - el piso que ocupaba la posición anterior queda libre y utilizable;
       - la huella sólida y las superficies hijas se trasladan con la pieza;
       - los puentes (chair_bridge, sofa_bridge) se desactivan al perder contacto.
    """
    ents1, meta1 = build_apartment()
    assert meta1['bridges'].get('chair_desk') is True, 'chair_bridge debía estar activo en posición 1'
    assert meta1['bridges'].get('sofa_table') is True, 'sofa_bridge debía estar activo en posición 1'
    assert any(e['ref'] == 'chair_bridge' for e in ents1)
    assert any(e['ref'] == 'sofa_bridge' for e in ents1)

    edited = copy.deepcopy(DEFAULT_LAYOUT)
    for it in edited:
        if it['ref'] == 'chair':
            # Mover la silla lejos del escritorio (en el extremo SE del dormitorio).
            it['pos_cm'] = (240, 80)
        elif it['ref'] == 'desk2':
            # Mover la mesa auxiliar lejos del sofá (en el este del living).
            it['pos_cm'] = (300, 120)
    ents2, meta2 = build_apartment(edited)
    assert meta2['bridges'].get('chair_desk') is False, 'chair_bridge debía desactivarse al perder contacto'
    assert meta2['bridges'].get('sofa_table') is False, 'sofa_bridge debía desactivarse al perder contacto'
    assert not any(e['ref'] == 'chair_bridge' for e in ents2)
    assert not any(e['ref'] == 'sofa_bridge' for e in ents2)

    # Verificar que la huella anterior de desk2 en living (160..220, 210..250) quedó libre
    # y la nueva huella (300..360, 120..160) quedó bloqueada.
    living2 = next(e for e in ents2 if e['ref'] == 'living')
    rows = living2['spawn']['components']['nav_grid']['rows']
    assert rows[22][18] == '.', 'la posición anterior de desk2 debía quedar libre'
    assert rows[13][32] == '#', 'la nueva posición de desk2 debía bloquear el piso'

    # Verificar que las patas y superficies se movieron con las piezas.
    chair2 = next(e for e in ents2 if e['ref'] == 'chair')
    chair_leg2 = next(e for e in ents2 if e['ref'] == 'chair_leg')
    assert chair2['spawn']['location']['pos']['x'] == 240 * CM
    assert chair_leg2['spawn']['location']['pos']['x'] == 240 * CM
    desk2_leg2 = next(e for e in ents2 if e['ref'] == 'desk2_leg')
    assert desk2_leg2['spawn']['location']['pos']['x'] == 300 * CM
    print('verify_edit_positions: OK')


if __name__ == '__main__':
    verify_edit_positions()
