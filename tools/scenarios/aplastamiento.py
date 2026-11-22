"""Genera los escenarios de aplastamiento (M23 a M26) en tests/scenarios/.

Parte de contact_opens_and_closes_route.json: dormitorio, escritorio a 75 cm,
giantess y palma. La palma aplasta (crusher "hand", 2 cm de grosor) y los
tinies son frágiles. Con la pose "rest" la palma queda a ras del escritorio,
centrada en (20 cm, 20 cm) del escritorio: su huella cubre x en [15, 25] cm e
y en [11, 29] cm. "hover" la sostiene 11 cm más arriba.

Uso: python tools/scenarios/aplastamiento.py
"""
import copy
import json
import os

ROOT = os.path.join(os.path.dirname(__file__), '..', '..')
SCEN = os.path.join(ROOT, 'tests', 'scenarios')
BASE = os.path.join(SCEN, 'contact_opens_and_closes_route.json')
OBJECTS = os.path.join(ROOT, 'content', 'objects')

KEEP = ['house', 'roomA', 'desk1', 'giantess', 'palm']

PALM_CRUSHER = {'source': 'hand', 'force': 10, 'thickness': 20000}

POSES = {
    'raised': [{'part': '$palm', 'pos': {'x': 300000, 'y': 0, 'z': 1400000}, 'facing': 0}],
    'hover': [{'part': '$palm', 'pos': {'x': 400000, 'y': 0, 'z': 880000}, 'facing': 0}],
    'rest': [{'part': '$palm', 'pos': {'x': 400000, 'y': 0, 'z': 770000}, 'facing': 0}],
}


def base(name, ticks, notes, palm_rows=None):
    d = json.load(open(BASE, encoding='utf-8'))
    d['name'] = name
    d['ticks'] = ticks
    d['notes'] = notes
    d['entities'] = [e for e in d['entities'] if e['ref'] in KEEP]
    for e in d['entities']:
        if e['ref'] == 'palm':
            e['spawn']['components']['crusher'] = dict(PALM_CRUSHER)
            if palm_rows:
                e['spawn']['components']['nav_grid']['rows'] = palm_rows
    d['commands'] = [{'tick': 0, 'kind': 'set_components',
                      'data': {'entity': '$giantess', 'components': {'poses': POSES}}}]
    d['expect'] = []
    return d


def tiny(ref, space, x, y, agent=None):
    comps = {'mover': {'speed': 20000}, 'fragile': {}}
    if agent:
        comps['agent'] = agent
    return {'ref': ref, 'spawn': {
        'name': ref,
        'location': {'space': '$' + space, 'pos': {'x': x, 'y': y, 'z': 0}},
        'body': {'radius': 2000, 'height': 8000, 'mass': 0},
        'components': comps,
    }}


def pose(tick, name):
    return {'tick': tick, 'kind': 'set_pose', 'data': {'entity': '$giantess', 'pose': name}}


def crushed(ref, after, by, source='hand', crusher='$palm'):
    return {'event': 'crushed', 'match': {'entity': '$' + ref, 'by': crusher, 'source': source},
            'after': after, 'by': by}


def spared(ref):
    return {'no_event': 'crushed', 'match': {'entity': '$' + ref}}


COMMON = [{'no_event': 'command_rejected'}]


def palm_crushes_tiny():
    d = base('palm_crushes_tiny', 120,
             'La giantess sostiene la mano sobre el escritorio y la apoya: el tiny que está debajo muere '
             'aplastado y quedan sus restos; el que está a 5 cm del borde no. Mientras la mano está en el '
             'aire no aplasta, y volver a apoyarla no repite el impacto sobre nadie.')
    d['entities'] += [tiny('under', 'desk1', 200000, 200000), tiny('beside', 'desk1', 300000, 200000)]
    d['commands'] += [pose(10, 'hover'), pose(20, 'rest'), pose(60, 'raised'), pose(80, 'rest')]
    d['expect'] = [
        {'no_event': 'crushed', 'by': 20},
        crushed('under', 20, 21),
        {'event': 'crushed', 'count': 1},
        spared('beside'),
        {'event': 'pressed', 'match': {'by': '$palm', 'surface': '$desk1', 'source': 'hand'}, 'count': 2},
        {'event': 'entity_despawned', 'entity': '$under', 'after': 20, 'by': 21},
        {'at': '$beside', 'space': '$desk1', 'pos': {'x': 300000, 'y': 200000}, 'tol': 0},
    ] + COMMON
    return d


def tiny_on_palm_is_safe():
    d = base('tiny_on_palm_is_safe', 100,
             'Un tiny viaja en la palma levantada y otro espera sobre el escritorio, debajo. Al apoyarla, '
             'muere el del escritorio; el de la palma está montado sobre el aplastador y no muere, ni al '
             'primer apoyo ni al segundo.')
    d['entities'] += [tiny('rider', 'palm', 0, 0), tiny('under', 'desk1', 200000, 220000)]
    d['commands'] += [pose(20, 'rest'), pose(50, 'raised'), pose(70, 'rest')]
    d['expect'] = [crushed('under', 20, 21), spared('rider'), {'no_event': 'entity_fell'},
                   {'at': '$rider', 'space': '$palm', 'pos': {'x': 0, 'y': 0}, 'tol': 0}] + COMMON
    return d


# Palma con dedos: las filas 12 a 17 (y de 3 a 9 cm) alternan dedo y hueco.
# Las columnas 2, 5 y 8 (x en [-3, -2), [0, 1) y [3, 4) cm) son huecos.
FINGERS = ['..........'] * 12 + ['..#..#..#.'] * 6


def tiny_between_fingers_survives():
    d = base('tiny_between_fingers_survives', 60,
             'La huella de la palma sale de su grilla: entre los dedos no hay mano. El tiny que queda en '
             'un hueco entre dos dedos sobrevive al apoyo; el que queda bajo un dedo, no.',
             palm_rows=FINGERS)
    # hueco: palma (-2,5 cm, 5,5 cm) = escritorio (17,5 cm, 25,5 cm); dedo: 1 cm a la izquierda
    d['entities'] += [tiny('gap', 'desk1', 175000, 255000), tiny('finger', 'desk1', 165000, 255000)]
    d['commands'] += [pose(20, 'rest')]
    d['expect'] = [crushed('finger', 20, 21), spared('gap')] + COMMON
    return d


def nest_shelters_from_palm():
    d = base('nest_shelters_from_palm', 60,
             'El nido (caja de fósforos, refugio de fuerza 50) está bajo la palma. El tiny que está en el '
             'nido se salva (sheltered); el que está a 6 cm, fuera del refugio, muere.')
    d['entities'] += [
        {'ref': 'nest', 'object': 'nest', 'spawn': {
            'name': 'nido', 'location': {'space': '$desk1', 'pos': {'x': 200000, 'y': 200000, 'z': 0}}}},
        tiny('inside', 'desk1', 210000, 200000),
        tiny('outside', 'desk1', 200000, 260000),
    ]
    d['commands'] += [pose(20, 'rest')]
    d['expect'] = [
        {'event': 'sheltered', 'match': {'entity': '$inside', 'by': '$palm', 'shelter': '$nest'},
         'after': 20, 'by': 21},
        spared('inside'),
        crushed('outside', 20, 21),
    ] + COMMON
    return d


def cup_lands_on_tiny():
    d = base('cup_lands_on_tiny', 40,
             'Una taza aparece a 20 cm sobre el escritorio y cae: aplasta al tiny que está debajo (radio '
             '4 cm) y no al que está a 5 cm.')
    cup = json.load(open(os.path.join(OBJECTS, 'cup.json'), encoding='utf-8'))
    cup = copy.deepcopy(cup)
    cup.pop('visual', None)
    cup['location'] = {'space': '$desk1', 'pos': {'x': 400000, 'y': 200000, 'z': 200000}}
    d['entities'] += [tiny('under', 'desk1', 400000, 200000), tiny('beside', 'desk1', 400000, 250000)]
    d['commands'] += [{'tick': 5, 'kind': 'spawn', 'data': cup}]
    d['expect'] = [
        {'event': 'entity_fell', 'match': {'to': '$desk1', 'height': 200000}, 'after': 5, 'by': 6},
        {'event': 'crushed', 'match': {'entity': '$under', 'source': 'object'}, 'after': 6, 'by': 7},
        spared('beside'),
    ] + COMMON
    return d


def haul_requeued_after_death():
    work = {'scale': 'tiny', 'work': 1, 'distance_cost': 1}
    d = base('haul_requeued_after_death', 900,
             'El primer tiny toma el acarreo de una migaja que está bajo la palma suspendida y la mano '
             'baja mientras la levanta. Muere, la reserva se libera (holder_gone), el trabajo vuelve a la '
             'cola y la migaja queda en el lugar. Con la mano otra vez arriba, el segundo tiny lo termina.')
    d['entities'] += [
        {'ref': 'crumb', 'object': 'crumb', 'spawn': {
            'name': 'migaja', 'location': {'space': '$desk1', 'pos': {'x': 200000, 'y': 200000, 'z': 0}}}},
        tiny('first', 'desk1', 150000, 200000, agent=dict(work)),
        tiny('second', 'desk1', 500000, 350000, agent=dict(work)),
    ]
    d['commands'] += [
        {'tick': 0, 'kind': 'post_job', 'data': {'type': 'haul', 'target': '$crumb', 'dest_space': '$desk1',
                                                  'dest': {'x': 500000, 'y': 100000}}},
        pose(5, 'hover'), pose(PICKUP_TICK, 'rest'), pose(PICKUP_TICK + 40, 'raised'),
    ]
    d['expect'] = [
        {'event': 'job_assigned', 'match': {'by': '$first'}, 'by': 20},
        crushed('first', PICKUP_TICK, PICKUP_TICK + 1),
        {'event': 'reservation_released', 'match': {'by': '$first', 'reason': 'holder_gone'},
         'after': PICKUP_TICK, 'by': PICKUP_TICK + 1},
        {'event': 'job_unassigned', 'after': PICKUP_TICK, 'by': PICKUP_TICK + 1},
        {'event': 'job_assigned', 'match': {'by': '$second'}, 'after': PICKUP_TICK},
        {'event': 'job_done', 'match': {'by': '$second'}},
        spared('second'),
        {'at': '$crumb', 'space': '$desk1', 'pos': {'x': 500000, 'y': 100000}, 'tol': 1000},
    ] + COMMON
    return d


STRIDE = {'length': 350000, 'side': 100000, 'foot_length': 240000, 'foot_width': 90000, 'force': 600}


def walking(name, ticks, notes):
    """La giantess camina 1 m hacia el este por el dormitorio, de (1 m, 1,5 m)
    a (2 m, 1,5 m). Pisa cada 35 cm alternando pies a 10 cm del centro:
    izquierdo en (1,35 m, 1,6 m) y derecho en (1,7 m, 1,4 m)."""
    d = base(name, ticks, notes)
    for e in d['entities']:
        if e['ref'] == 'giantess':
            e['spawn']['location']['pos'] = {'x': 1000000, 'y': 1500000, 'z': 0}
            e['spawn']['components']['stride'] = dict(STRIDE)
    d['commands'] += [{'tick': 1, 'kind': 'move_to',
                       'data': {'entity': '$giantess', 'space': '$roomA', 'goal': {'x': 2000000, 'y': 1500000}}}]
    return d


def foot(ref, after, by):
    return crushed(ref, after, by, source='foot', crusher='$giantess')


def step_crushes_tiny_on_floor():
    d = walking('step_crushes_tiny_on_floor', 80,
                'La giantess cruza el dormitorio y pisa dos veces. Muere el tiny que está bajo el pie '
                'izquierdo y también el que está en un nido en el suelo bajo el pie derecho: la caja de '
                'fósforos resiste una mano (fuerza 10), no un pie (600). Sobrevive el que quedó en la línea '
                'del pie derecho donde ese pie no se apoyó, y el que está a un costado de la marcha.')
    d['entities'] += [
        tiny('left_step', 'roomA', 1350000, 1600000),
        {'ref': 'nest', 'object': 'nest', 'spawn': {
            'name': 'nido', 'location': {'space': '$roomA', 'pos': {'x': 1700000, 'y': 1400000, 'z': 0}}}},
        tiny('in_nest', 'roomA', 1705000, 1400000),
        tiny('between', 'roomA', 1350000, 1400000),
        tiny('aside', 'roomA', 1500000, 1800000),
    ]
    d['expect'] = [
        {'event': 'footstep', 'match': {'entity': '$giantess', 'space': '$roomA'}, 'count': 2},
        {'event': 'footstep', 'match': {'right': False, 'pos': {'x': 1350000, 'y': 1600000}}, 'count': 1},
        {'event': 'footstep', 'match': {'right': True, 'pos': {'x': 1700000, 'y': 1400000}}, 'count': 1},
        foot('left_step', 1, 20),
        foot('in_nest', 1, 20),
        {'no_event': 'sheltered'},
        spared('between'), spared('aside'),
        {'event': 'crushed', 'count': 2},
    ] + COMMON
    return d


def step_spares_tiny_on_desk():
    d = walking('step_spares_tiny_on_desk', 80,
                'El escritorio está sobre la marcha de la giantess. Un tiny sobre el escritorio, justo '
                'encima del pie izquierdo, no muere: la pisada solo alcanza el suelo. Otro en el suelo, '
                'debajo del escritorio y del mismo pie, sí.')
    # escritorio en (1,5 m, 1,0 m) con 60 × 40 cm: (1,35 m, 1,6 m) queda fuera. Se lo corre.
    for e in d['entities']:
        if e['ref'] == 'desk1':
            e['spawn']['location']['pos'] = {'x': 1200000, 'y': 1400000, 'z': 750000}
    d['entities'] += [tiny('on_desk', 'desk1', 150000, 200000), tiny('on_floor', 'roomA', 1350000, 1600000)]
    d['expect'] = [
        {'event': 'footstep', 'count': 2},
        spared('on_desk'),
        foot('on_floor', 1, 20),
        {'at': '$on_desk', 'space': '$desk1', 'pos': {'x': 150000, 'y': 200000}, 'tol': 0},
    ] + COMMON
    return d


SCARED = {'scale': 'tiny', 'reflex': 4}


def desk_rows(blocked):
    """Grilla del escritorio (30 × 20 celdas de 2 cm) con celdas bloqueadas."""
    rows = [['.'] * 30 for _ in range(20)]
    for x, y in blocked:
        rows[y][x] = '#'
    return [''.join(r) for r in rows]


def alarmed(ref, after, by, crusher='$palm', channel='shadow'):
    return {'event': 'alarmed', 'match': {'entity': '$' + ref, 'by': crusher, 'channel': channel},
            'after': after, 'by': by}


def fled(ref, after, by):
    return {'event': 'action_started', 'match': {'entity': '$' + ref, 'action': 'flee'}, 'after': after, 'by': by}


def tiny_flees_shadow():
    d = base('tiny_flees_shadow', 80,
             'La mano se sostiene sobre el escritorio 20 ticks antes de apoyarse. Los tinies bajo su sombra '
             'se alarman y, 4 ticks después, buscan salida. El que estaba a 5 mm del borde llega afuera y se '
             'lleva un susto (near_miss); el del centro no llega y muere; el que está encerrado entre celdas '
             'bloqueadas queda atrapado y muere. El que estaba lejos ni se entera.')
    # bolsillo: la celda (8, 6) (centro en 17 cm, 13 cm) rodeada de celdas bloqueadas
    pocket = [(8 + dx, 6 + dy) for dx in (-1, 0, 1) for dy in (-1, 0, 1) if (dx, dy) != (0, 0)]
    for e in d['entities']:
        if e['ref'] == 'desk1':
            e['spawn']['components']['nav_grid']['rows'] = desk_rows(pocket)
    d['entities'] += [
        tiny('edge', 'desk1', 245000, 200000, agent=dict(SCARED)),
        tiny('center', 'desk1', 200000, 200000, agent=dict(SCARED)),
        tiny('pocket', 'desk1', 170000, 130000, agent=dict(SCARED)),
        tiny('far', 'desk1', 500000, 300000, agent=dict(SCARED)),
    ]
    d['commands'] += [pose(10, 'hover'), pose(30, 'rest')]
    d['expect'] = [
        alarmed('edge', 10, 11), alarmed('center', 10, 11), alarmed('pocket', 10, 11),
        {'no_event': 'alarmed', 'match': {'entity': '$far'}},
        fled('edge', 14, 15), fled('center', 14, 15),
        {'event': 'action_done', 'match': {'entity': '$edge', 'action': 'flee'}, 'by': 30},
        {'event': 'trapped', 'match': {'entity': '$pocket', 'by': '$palm'}, 'after': 14, 'by': 15},
        spared('edge'), crushed('center', 30, 31), crushed('pocket', 30, 31),
        {'event': 'near_miss', 'match': {'entity': '$edge', 'by': '$palm'}, 'after': 30, 'by': 31},
        {'no_event': 'action_failed'},
    ] + COMMON
    return d


def tiny_drops_crumb_to_flee():
    work = {'scale': 'tiny', 'work': 1, 'distance_cost': 1, 'reflex': 4}
    d = haul_requeued_after_death()
    d['name'] = 'tiny_drops_crumb_to_flee'
    d['ticks'] = 1200
    d['notes'] = ('Como haul_requeued_after_death, pero los tinies tienen reflejos. La mano aparece sobre el '
                  'tiny mientras levanta la migaja: suelta la migaja, deja el trabajo y huye. Mientras la mano '
                  'sigue arriba nadie vuelve a buscar la migaja bajo su sombra. La mano se apoya sin víctimas '
                  'y el tiny, que volvía a buscar la migaja, se asusta (near_miss) y suspende el trabajo; lo '
                  'retoma y la migaja termina en su destino.')
    for e in d['entities']:
        if e['ref'] in ('first', 'second'):
            e['spawn']['components']['agent'] = dict(work)
    d['commands'] = d['commands'][:2] + [pose(55, 'hover'), pose(130, 'rest'), pose(170, 'raised')]
    d['expect'] = [
        alarmed('first', 55, 56),
        {'event': 'action_interrupted', 'match': {'entity': '$first', 'action': 'work', 'by': 'flee'},
         'after': 59, 'by': 60},
        {'event': 'job_unassigned', 'match': {'by': '$first', 'reason': 'stopped'}, 'after': 59, 'by': 60},
        {'no_event': 'job_assigned', 'after': 60, 'by': 130},
        {'event': 'near_miss', 'match': {'entity': '$first'}, 'after': 130, 'by': 131},
        {'event': 'action_suspended', 'match': {'entity': '$first', 'cause': 'near_miss'}, 'after': 131, 'by': 132},
        {'no_event': 'crushed'},
        {'event': 'job_done'},
        {'at': '$crumb', 'space': '$desk1', 'pos': {'x': 500000, 'y': 100000}, 'tol': 1000},
    ] + COMMON
    return d


def tiny_flees_footsteps():
    d = walking('tiny_flees_footsteps', 60,
                'La giantess camina despacio (0,5 m/s, una pisada cada 14 ticks). Después de la primera '
                'pisada, la vibración anuncia la segunda: el tiny que está a 0,5 mm del borde del pie derecho '
                'huye y se salva por poco; el que está en el centro no alcanza a salir.')
    for e in d['entities']:
        if e['ref'] == 'giantess':
            e['spawn']['components']['mover'] = {'speed': 500000}
    scared = {'scale': 'tiny', 'reflex': 2}
    d['entities'] += [
        tiny('edge', 'roomA', 1700000, 1444500, agent=dict(scared)),
        tiny('center', 'roomA', 1700000, 1400000, agent=dict(scared)),
    ]
    d['expect'] = [
        {'event': 'footstep', 'match': {'right': True, 'pos': {'x': 1700000, 'y': 1400000}}, 'after': 28, 'by': 29},
        alarmed('edge', 15, 16, crusher='$giantess', channel='vibration'),
        {'event': 'action_done', 'match': {'entity': '$edge', 'action': 'flee'}, 'by': 28},
        spared('edge'), foot('center', 28, 29),
        {'event': 'near_miss', 'match': {'entity': '$edge', 'source': 'foot'}, 'after': 28, 'by': 29},
    ] + COMMON
    return d


def giantess_at_desk(name, ticks, notes, care):
    """La giantess quiere escribir en un cuaderno sobre el escritorio: sostiene
    la mano 12 ticks y la apoya sobre la huella de siempre."""
    d = base(name, ticks, notes)
    for e in d['entities']:
        if e['ref'] == 'giantess':
            e['spawn']['components']['needs'] = [{'name': 'fun', 'value': 200, 'decay': 0}]
    d['commands'][0]['data']['components']['agent'] = {
        'scale': 'giantess', 'hand': '$palm', 'idle_pose': 'raised', 'watch': 2000000, 'care': care}
    d['entities'] += [
        {'ref': 'notebook', 'spawn': {
            'name': 'cuaderno', 'location': {'space': '$desk1', 'pos': {'x': 500000, 'y': 300000, 'z': 0}},
            'components': {'offers': [{
                'action': 'write', 'scale': 'giantess', 'effects': [{'need': 'fun', 'amount': 300}],
                'steps': [{'do': 'pose', 'pose': 'hover'}, {'do': 'wait', 'ticks': 12},
                          {'do': 'pose', 'pose': 'rest'}, {'do': 'use', 'ticks': 100}]}]}}},
        tiny('a', 'desk1', 200000, 200000, agent=dict(SCARED)),
        tiny('b', 'desk1', 210000, 220000, agent=dict(SCARED)),
    ]
    return d


def careless_giantess_crushes():
    d = giantess_at_desk('careless_giantess_crushes', 200,
                         'Una giantess distraída (care 0) sostiene la mano 12 ticks y la apoya: los dos tinies '
                         'del centro de la huella se alarman pero no llegan a salir.', 0)
    d['expect'] = [
        {'event': 'action_started', 'match': {'entity': '$giantess', 'what': 'write'}},
        {'no_event': 'hesitated'},
        {'event': 'crushed', 'match': {'by': '$palm'}, 'count': 2},
    ] + COMMON
    return d


def careful_giantess_waits():
    d = giantess_at_desk('careful_giantess_waits', 300,
                         'La misma escena con una giantess cuidadosa (care 0,9): sostiene la mano 36 ticks más, '
                         've a los tinies debajo y espera a que se aparten (hesitated). Nadie muere.', 0.9)
    d['expect'] = [
        {'event': 'hesitated', 'match': {'entity': '$giantess', 'surface': '$desk1', 'seen': 2}, 'by': 20},
        {'event': 'action_done', 'match': {'entity': '$giantess', 'what': 'write'}},
        {'event': 'pressed', 'match': {'by': '$palm'}},
        {'no_event': 'crushed'},
    ] + COMMON
    return d


def colony_avoids_danger():
    d = base('colony_avoids_danger', 1200,
             'La mano se apoya en el tick 0 sobre un miembro de la colonia: muere y la colonia marca una zona de '
             'peligro de 10 cm por 300 ticks alrededor de sus restos. La giantess, que mira, nota los restos. '
             'Otro miembro ve una migaja junto a los restos y la colonia publica el acarreo, pero nadie lo toma '
             'mientras dura la zona. Cuando vence, el trabajo se toma y la migaja llega al nido.')
    d['commands'][0]['data']['components']['agent'] = {
        'scale': 'giantess', 'hand': '$palm', 'idle_pose': 'raised', 'watch': 2000000}
    member = {'colony': '$nest'}
    d['entities'] += [
        {'ref': 'nest', 'object': 'nest', 'spawn': {
            'name': 'nido', 'location': {'space': '$desk1', 'pos': {'x': 450000, 'y': 300000, 'z': 0}},
            'components': {'colony': {'radius': 30000, 'danger_ticks': 300}}}},
        {'ref': 'crumb', 'object': 'crumb', 'spawn': {
            'name': 'migaja', 'location': {'space': '$desk1', 'pos': {'x': 220000, 'y': 220000, 'z': 0}}}},
    ]
    victim = tiny('victim', 'desk1', 200000, 200000, agent={'scale': 'tiny'})
    victim['spawn']['components']['member'] = dict(member)
    worker = tiny('worker', 'desk1', 400000, 300000,
                  agent={'scale': 'tiny', 'work': 1, 'distance_cost': 1, 'sight': 300000})
    worker['spawn']['components']['member'] = dict(member)
    d['entities'] += [victim, worker]
    d['commands'] += [pose(0, 'rest'), pose(20, 'raised')]
    d['expect'] = [
        crushed('victim', 0, 1),
        {'event': 'member_lost', 'match': {'colony': '$nest', 'entity': '$victim', 'cause': 'hand'},
         'after': 0, 'by': 1},
        {'event': 'noticed', 'match': {'entity': '$giantess'}, 'by': 20},
        {'event': 'job_posted', 'match': {'target': '$crumb'}, 'by': 30},
        {'no_event': 'job_assigned', 'by': 300},
        {'event': 'danger_expired', 'after': 300, 'by': 301},
        {'event': 'job_assigned', 'match': {'by': '$worker'}, 'after': 300},
        {'event': 'stored', 'match': {'colony': '$nest', 'item': '$crumb'}},
    ] + COMMON
    return d


# Tick en que el primer tiny está levantando la migaja (entre job_assigned y
# el fin del paso de levantar, 1 s). Medido con simrun -events.
PICKUP_TICK = 60


def main():
    for make in (palm_crushes_tiny, tiny_on_palm_is_safe, tiny_between_fingers_survives,
                 nest_shelters_from_palm, cup_lands_on_tiny, haul_requeued_after_death,
                 step_crushes_tiny_on_floor, step_spares_tiny_on_desk,
                 tiny_flees_shadow, tiny_drops_crumb_to_flee, tiny_flees_footsteps,
                 careless_giantess_crushes, careful_giantess_waits, colony_avoids_danger):
        d = make()
        with open(os.path.join(SCEN, d['name'] + '.json'), 'w', encoding='utf-8', newline='\n') as f:
            f.write(json.dumps(d, ensure_ascii=False, indent=2) + '\n')


if __name__ == '__main__':
    main()
