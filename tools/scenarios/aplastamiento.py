"""Genera los escenarios de aplastamiento (M23 y M24) en tests/scenarios/.

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


# Tick en que el primer tiny está levantando la migaja (entre job_assigned y
# el fin del paso de levantar, 1 s). Medido con simrun -events.
PICKUP_TICK = 60


def main():
    for make in (palm_crushes_tiny, tiny_on_palm_is_safe, tiny_between_fingers_survives,
                 nest_shelters_from_palm, cup_lands_on_tiny, haul_requeued_after_death,
                 step_crushes_tiny_on_floor, step_spares_tiny_on_desk):
        d = make()
        with open(os.path.join(SCEN, d['name'] + '.json'), 'w', encoding='utf-8', newline='\n') as f:
            f.write(json.dumps(d, ensure_ascii=False, indent=2) + '\n')


if __name__ == '__main__':
    main()
