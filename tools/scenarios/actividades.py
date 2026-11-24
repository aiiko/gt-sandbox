"""Genera los escenarios del slice de actividades en la casa (M61–M71).

    python tools/scenarios/actividades.py

Los escenarios de cierre se escriben en tests/scenarios/.
"""
import json
import os

from aplastamiento import base

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
SCEN = os.path.join(ROOT, 'tests', 'scenarios')

COMMON = [{'no_event': 'command_rejected'}]


def _write(d):
    path = os.path.join(SCEN, d['name'] + '.json')
    with open(path, 'w', encoding='utf-8', newline='\n') as f:
        json.dump(d, f, indent=2, ensure_ascii=False)
        f.write('\n')
    print('escrito', os.path.relpath(path, ROOT))


def _nest(space, x, y, portions=None, stock=None):
    comps = {'colony': {'radius': 30000}}
    if stock is not None:
        comps['colony']['stock'] = stock
    if portions is not None:
        comps['portions'] = portions
    return {
        'ref': 'nest',
        'object': 'nest',
        'visual': 'nest',
        'spawn': {
            'name': 'nido',
            'location': {'space': '$' + space, 'pos': {'x': x, 'y': y, 'z': 0}},
            'components': comps,
        },
    }


def _tiny(ref, space, x, y, hunger=900, thirst=900, energy=900, decay=(0, 0, 0),
          work=0.2, sight=250000, visual='tiny', member=None):
    m = {'colony': '$nest'}
    m.update(member or {})
    return {
        'ref': ref,
        'visual': visual,
        'spawn': {
            'name': ref,
            'location': {'space': '$' + space, 'pos': {'x': x, 'y': y, 'z': 0}},
            'body': {'radius': 2000, 'height': 8000, 'mass': 0},
            'components': {
                'mover': {'speed': 20000},
                'agent': {'scale': 'tiny', 'work': work, 'distance_cost': 0.2, 'sight': sight},
                'member': m,
                'needs': [
                    {'name': 'hunger', 'value': hunger, 'decay': decay[0]},
                    {'name': 'thirst', 'value': thirst, 'decay': decay[1]},
                    {'name': 'energy', 'value': energy, 'decay': decay[2]},
                ],
            },
        },
    }


def _drop(ref, space, x, y, volume=3, at=None, ttl=2400):
    ex = {'kind': 'water_drop'}
    if at is not None:
        ex['at'] = at
    else:
        ex['ttl'] = ttl
    return {
        'ref': ref,
        'object': 'water_drop',
        'visual': 'water_drop',
        'spawn': {
            'name': ref,
            'location': {'space': '$' + space, 'pos': {'x': x, 'y': y, 'z': 0}},
            'components': {
                'water_drop': {'volume': volume},
                'portions': {'items': [{'resource': 'water', 'amount': volume}]},
                'expires': ex,
            },
        },
    }


# --- M61: Agua del baño y la cocina ---

def tiny_drinks_drop():
    d = base('tiny_drinks_drop', 600,
             'M61: un tiny sediento encuentra una gota de agua, toma de ella, su sed sube y la gota se achica.')
    d['atlas'] = '../../content/art/casa.json'
    for e in d['entities']:
        if e['ref'] == 'desk1':
            e['visual'] = 'desk'
        elif e['ref'] == 'roomA':
            e['visual'] = 'bedroom'
        elif e['ref'] == 'giantess':
            e['visual'] = 'giantess'
        elif e['ref'] == 'palm':
            e['visual'] = 'palm'
    d['entities'] += [
        _nest('desk1', 80000, 80000, portions={'items': []}, stock={'food': 2, 'water': 2}),
        _drop('drop', 'desk1', 160000, 120000, volume=2),
        _tiny('t', 'desk1', 110000, 100000, thirst=250, work=0.05),
    ]
    d['expect'] = [
        {'event': 'action_done', 'match': {'entity': '$t', 'what': 'drink', 'target': '$drop'}},
        {'event': 'consumed', 'match': {'entity': '$t', 'target': '$drop', 'resource': 'water', 'left': 1}},
    ] + COMMON
    return d


def drop_evaporates():
    d = base('drop_evaporates', 200,
             'M61: nadie llega a una gota de agua y al alcanzar su vencimiento desaparece emitiendo expired.')
    d['entities'] += [
        _drop('drop', 'desk1', 200000, 200000, volume=2, at=80),
    ]
    d['expect'] = [
        {'event': 'expired', 'match': {'entity': '$drop', 'kind': 'water_drop'}, 'by': 90},
    ] + COMMON
    return d


def water_hauled_to_nest():
    d = base('water_hauled_to_nest', 1800,
             'M61: un tiny acarrea una gota de agua al nido (sube water del depósito) y otro tiny sediento toma ahí.')
    d['atlas'] = '../../content/art/casa.json'
    for e in d['entities']:
        if e['ref'] == 'desk1':
            e['visual'] = 'desk'
        elif e['ref'] == 'roomA':
            e['visual'] = 'bedroom'
        elif e['ref'] == 'giantess':
            e['visual'] = 'giantess'
        elif e['ref'] == 'palm':
            e['visual'] = 'palm'
    d['entities'] += [
        _nest('desk1', 80000, 80000, portions={'items': []}, stock={'food': 2, 'water': 2}),
        _drop('drop', 'desk1', 240000, 120000, volume=3, ttl=4000),
        _tiny('porter', 'desk1', 210000, 120000, thirst=950, work=0.5, sight=120000, visual='tiny'),
        _tiny('drinker', 'desk1', 60000, 80000, thirst=300, work=0.0, sight=80000, visual='tiny_lavender'),
    ]
    d['expect'] = [
        {'event': 'stored', 'match': {'colony': '$nest', 'item': '$drop', 'resource': 'water'}},
        {'event': 'action_done', 'match': {'entity': '$drinker', 'what': 'drink', 'target': '$nest'}},
        {'event': 'consumed', 'match': {'entity': '$drinker', 'target': '$nest', 'resource': 'water'}},
    ] + COMMON
    return d


def giantess_washes_hands():
    d = base('giantess_washes_hands', 600,
             'M61: la giantess con higiene baja va al lavamanos, hace wash_hands y deja gotas de agua en el piso.')
    # Ajustar la giantess con necesidad hygiene baja y colocar lavamanos en roomA.
    for e in d['entities']:
        if e['ref'] == 'giantess':
            e['spawn']['components']['agent'] = {'scale': 'giantess', 'distance_cost': 0.05}
            e['spawn']['components']['needs'] = [
                {'name': 'hygiene', 'value': 200, 'decay': 0},
            ]
    d['entities'].append({
        'ref': 'washbasin',
        'object': 'washbasin',
        'visual': 'washbasin',
        'spawn': {
            'name': 'lavamanos',
            'location': {'space': '$roomA', 'pos': {'x': 2400000, 'y': 1600000, 'z': 0}},
        },
    })
    d['expect'] = [
        {'event': 'action_done', 'match': {'entity': '$giantess', 'what': 'wash_hands', 'target': '$washbasin'}},
        {'event': 'items_left', 'match': {'entity': '$giantess', 'motive': 'wash_hands'}},
    ] + COMMON
    return d


def main():
    for fn in (tiny_drinks_drop, drop_evaporates, water_hauled_to_nest, giantess_washes_hands):
        _write(fn())


if __name__ == '__main__':
    main()
