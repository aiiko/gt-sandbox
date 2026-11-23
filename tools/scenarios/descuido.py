"""Genera el escenario de cierre del aplastamiento (M27) en tests/scenarios/.

    python tools/scenarios/descuido.py

Parte de «Una mañana en la casa» (una_manana_en_la_casa.json) y la vuelve
peligrosa: los tinies son frágiles y se asustan (reflex), la palma aplasta,
la giantess pisa al caminar (stride) y una pata del escritorio baja al tatami
del dormitorio, como en floor_to_desk_by_leg. Sale en dos versiones: la
giantess distraída (care 0) y la cuidadosa (care 0,9).
"""
import copy
import json
import os

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
DIR = os.path.join(ROOT, 'tests', 'scenarios')

PALM_CRUSHER = {'source': 'hand', 'force': 10, 'thickness': 20000}
STRIDE = {'length': 350000, 'side': 100000, 'foot_length': 240000, 'foot_width': 90000, 'force': 600}
REFLEX = 4

# La pata delantera izquierda del escritorio: en el dormitorio, el escritorio
# está en (1,5 m, 1 m) y mide 60 × 40 cm.
DESK = (1500000, 1000000)
LEG = {'ref': 'leg1', 'visual': 'climb', 'spawn': {
    'name': 'pata del escritorio',
    'space': {'kind': 'elevated', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 20000, 'y': 740000}}},
    'location': {'space': '$bedroom', 'pos': {'x': DESK[0], 'y': DESK[1], 'z': 0}},
    'components': {'nav_grid': {'origin': {'x': 0, 'y': 0}, 'cell': 20000, 'version': 0, 'rows': ['.'] * 37}}}}
PORTALS = [
    {'ref': 'leg1_bottom', 'spawn': {'name': 'leg1_bottom', 'components': {'portal': {
        'a': {'space': '$bedroom', 'pos': {'x': DESK[0] - 25000, 'y': DESK[1] - 25000}},
        'b': {'space': '$leg1', 'pos': {'x': 10000, 'y': 10000}}, 'cost': 0, 'version': 0}}}},
    {'ref': 'leg1_top', 'spawn': {'name': 'leg1_top', 'components': {'portal': {
        'a': {'space': '$leg1', 'pos': {'x': 10000, 'y': 730000}},
        'b': {'space': '$desk1', 'pos': {'x': 10000, 'y': 10000}}, 'cost': 0, 'version': 0}}}},
]


def scatter_crumbs(offers):
    """Desparrama las migajas del desayuno: cuatro en un anillo de 18 cm
    alrededor del punto de siempre, así una cae detrás del escritorio, al
    tatami."""
    for o in offers:
        for st in o.get('steps', []):
            if st['do'] == 'spawn' and st['item'] == 'crumb':
                st['count'], st['spread'] = 4, 180000


# Donde la giantess apoya la mano para desayunar (alrededor de (30 cm, 10 cm)
# del escritorio) quedaron galletas de anoche, demasiado grandes para acarrear
# (no son `food`): dos bajo el medio de la palma, de donde no se sale a tiempo,
# y una junto al borde de la huella, entre dos dedos. Tres tinies con hambre
# empiezan al lado y las mordisquean un buen rato.
def cookie(ref, x, y):
    return {'ref': ref, 'visual': 'crumb', 'spawn': {
        'name': 'galleta', 'location': {'space': '$desk1', 'pos': {'x': x, 'y': y, 'z': 0}},
        'body': {'radius': 1500, 'height': 1000, 'mass': 0},
        'components': {'portions': {'left': 6}, 'offers': [
            {'action': 'nibble', 'scale': 'tiny', 'effects': [{'need': 'hunger', 'amount': 150}],
             'steps': [{'do': 'goto'}, {'do': 'use', 'ticks': 400}, {'do': 'consume'}]}]}}}


COOKIES = [cookie('cookie1', 260000, 100000), cookie('cookie2', 300000, 90000), cookie('cookie3', 340000, 130000)]
NEAR_HAND = {'tiny16': (250000, 120000), 'tiny17': (300000, 110000), 'tiny18': (350000, 147000)}
HUNGRY = 80
# Tiny 21 bajó por la pata antes de que empiece la mañana: está en el tatami,
# en el camino de la giantess cuando vuelve con el desayuno.
FLOOR_TINY = (1550000, 750000)


def build(name, title, care, notes, expect):
    d = json.load(open(os.path.join(DIR, 'una_manana_en_la_casa.json'), encoding='utf-8'))
    d = copy.deepcopy(d)
    d['name'], d['notes'] = title, notes
    for e in d['entities']:
        comps = e.get('spawn', {}).setdefault('components', {})
        if e['ref'].startswith('tiny'):
            comps['fragile'] = {}
            comps['agent']['reflex'] = REFLEX
            if e['ref'] in NEAR_HAND:
                x, y = NEAR_HAND[e['ref']]
                e['spawn']['location']['pos'].update(x=x, y=y)
                for n in comps['needs']:
                    if n['name'] == 'hunger':
                        n['value'] = HUNGRY
        elif e['ref'] == 'giantess':
            comps['stride'] = dict(STRIDE)
        elif e['ref'] == 'palm':
            comps['crusher'] = dict(PALM_CRUSHER)
        elif e['ref'] == 'fridge':
            scatter_crumbs(comps['offers'])
    i = next(i for i, e in enumerate(d['entities']) if e['ref'] == 'desk1')
    d['entities'][i + 1:i + 1] = [LEG] + PORTALS
    j = next(j for j, e in enumerate(d['entities']) if e['ref'] == 'nest')
    d['entities'][j + 1:j + 1] = copy.deepcopy(COOKIES)
    last = max(i for i, e in enumerate(d['entities']) if e['ref'].startswith('tiny'))
    floor = copy.deepcopy(d['entities'][last])
    floor['ref'], floor['spawn']['name'] = 'tiny21', 'Tiny 21'
    floor['spawn']['location'] = {'space': '$bedroom', 'pos': {'x': FLOOR_TINY[0], 'y': FLOOR_TINY[1], 'z': 0}}
    d['entities'].insert(last + 1, floor)
    d['commands'][0]['data']['components']['agent']['care'] = care
    d['expect'] = expect + [{'no_event': 'command_rejected'}]
    out = os.path.join(DIR, name + '.json')
    with open(out, 'w', encoding='utf-8', newline='\n') as f:
        f.write(json.dumps(d, ensure_ascii=False, indent=2) + '\n')
    print('escrito', os.path.relpath(out, ROOT))
    return d


def main():
    build('un_descuido_en_el_escritorio', 'Un descuido en el escritorio', 0,
          'La mañana de la casa con tinies frágiles. La giantess, distraída, desayuna en el escritorio '
          'y apoya la mano sin mirar sobre los que mordisquean galletas: alguno muere, alguno se salva por '
          'poco y la colonia marca la zona. Pisa al caminar, una pata del escritorio baja al tatami y una '
          'migaja cae al piso.',
          [{'event': 'crushed', 'match': {'by': '$palm', 'space': '$desk1'}},
           {'event': 'member_lost', 'match': {'colony': '$nest'}},
           {'event': 'near_miss', 'match': {'by': '$palm'}},
           {'event': 'entity_fell', 'match': {'to': '$bedroom'}},
           {'no_event': 'hesitated'}])
    build('un_descuido_cuidadosa', 'Un descuido en el escritorio (cuidadosa)', 0.9,
          'La misma mañana con una giantess cuidadosa (care 0,9): sostiene la mano más tiempo, mira '
          'si hay tinies debajo y titubea antes de apoyarla.',
          [{'event': 'hesitated', 'match': {'entity': '$giantess', 'surface': '$desk1'}},
           {'event': 'pressed', 'match': {'by': '$palm', 'surface': '$desk1'}}])


if __name__ == '__main__':
    main()
