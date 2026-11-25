"""Genera los escenarios del slice de actividades en la casa (M61–M71).

    python tools/scenarios/actividades.py

Los escenarios de cierre se escriben en tests/scenarios/.
"""
import json
import os

from aplastamiento import base

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
SCEN = os.path.join(ROOT, 'tests', 'scenarios')
CASA = os.path.join(ROOT, 'content', 'partidas', 'casa.json')
OBJECTS = os.path.join(ROOT, 'content', 'objects')

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
        _tiny('drinker', 'desk1', 60000, 80000, thirst=760, decay=(0, 2, 0), work=0.0, sight=80000, visual='tiny_lavender'),
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


def giantess_takes_shower():
    d = base('giantess_takes_shower', 800,
             'M61: la giantess con higiene baja usa la ducha y deja muchas gotas de agua y un charco.')
    for e in d['entities']:
        if e['ref'] == 'giantess':
            e['spawn']['components']['agent'] = {'scale': 'giantess', 'distance_cost': 0.05}
            e['spawn']['components']['needs'] = [
                {'name': 'hygiene', 'value': 150, 'decay': 0},
            ]
    d['entities'].append({
        'ref': 'shower_tray',
        'object': 'shower_tray',
        'visual': 'shower_tray',
        'spawn': {
            'name': 'plato de ducha',
            'location': {'space': '$roomA', 'pos': {'x': 2400000, 'y': 1400000, 'z': 0}},
        },
    })
    d['expect'] = [
        {'event': 'action_done', 'match': {'entity': '$giantess', 'what': 'shower', 'target': '$shower_tray'}},
        {'event': 'items_left', 'match': {'entity': '$giantess', 'motive': 'shower'}},
    ] + COMMON
    return d


# --- M63: Tesoros perdidos y regalos ---

DEFAULT_LOOT = [
    {'kind': 'button', 'weight': 3},
    {'kind': 'clip', 'weight': 3},
    {'kind': 'thread', 'weight': 2},
    {'kind': 'graphite', 'weight': 2},
]


def tiny_finds_trinket():
    d = base('tiny_finds_trinket', 1200,
             'M63: un tiny revisa un punto de búsqueda con botín, sortea un tesoro con colony.loot y lo acarrea al hoard del nido.')
    d['entities'] += [
        _nest('desk1', 80000, 80000,
              portions={'items': [{'resource': 'food', 'amount': 4}, {'resource': 'water', 'amount': 4}]},
              stock={'food': 2, 'water': 2}),
        {
            'ref': 'loot_spot',
            'spawn': {
                'name': 'debajo del futón',
                'location': {'space': '$desk1', 'pos': {'x': 220000, 'y': 160000, 'z': 0}},
                'components': {
                    'search_spot': {'look_ticks': 40, 'loot': DEFAULT_LOOT},
                    'landmark': {},
                },
            },
        },
        _tiny('t', 'desk1', 120000, 100000, hunger=950, thirst=950, energy=950,
              work=0.3, sight=250000, member={'search': 0.4}),
    ]
    d['expect'] = [
        {'event': 'trinket_found', 'match': {'entity': '$t', 'spot': '$loot_spot'}},
        {'event': 'trinket_stored', 'match': {'colony': '$nest', 'first_of_kind': True}},
    ] + COMMON
    return d


def hoard_raises_fun():
    d = base('hoard_raises_fun', 1600,
             'M63: dos tesoros distintos en el hoard del nido elevan el piso de diversión (fun >= 300) y frenan su decaimiento.')
    d['entities'] += [
        _nest('desk1', 80000, 80000,
              portions={'items': [{'resource': 'food', 'amount': 6}, {'resource': 'water', 'amount': 6}]},
              stock={'food': 2, 'water': 2}),
        {
            'ref': 'btn',
            'object': 'trinket_button',
            'visual': 'trinket_button',
            'spawn': {
                'name': 'trinket_button',
                'location': {'space': '$desk1', 'pos': {'x': 82000, 'y': 80000, 'z': 0}},
            },
        },
        {
            'ref': 'clp',
            'object': 'trinket_clip',
            'visual': 'trinket_clip',
            'spawn': {
                'name': 'trinket_clip',
                'location': {'space': '$desk1', 'pos': {'x': 78000, 'y': 82000, 'z': 0}},
            },
        },
        {
            'ref': 't',
            'visual': 'tiny',
            'spawn': {
                'name': 't',
                'location': {'space': '$desk1', 'pos': {'x': 120000, 'y': 100000, 'z': 0}},
                'body': {'radius': 2000, 'height': 8000, 'mass': 0},
                'components': {
                    'mover': {'speed': 20000},
                    'agent': {'scale': 'tiny', 'work': 0.1, 'distance_cost': 0.2, 'sight': 200000},
                    'member': {'colony': '$nest'},
                    'needs': [
                        {'name': 'hunger', 'value': 950, 'decay': 0},
                        {'name': 'thirst', 'value': 950, 'decay': 0},
                        {'name': 'energy', 'value': 950, 'decay': 0},
                        {'name': 'fun', 'value': 450, 'decay': 5},
                    ],
                },
            },
        },
    ]
    d['expect'] = COMMON
    return d


def giantess_notices_gift():
    d = base('giantess_notices_gift', 1600,
             'M63: un tiny con mucha diversión deja un tesoro repetido en un gift_spot; la giantess lo nota (notice_gift) y baja su enfriamiento de share.')
    for e in d['entities']:
        if e['ref'] == 'giantess':
            e['spawn']['components']['agent'] = {
                'scale': 'giantess', 'distance_cost': 0.05, 'watch': 3000000,
            }
            e['spawn']['components']['needs'] = [
                {'name': 'hunger', 'value': 950, 'decay': 0},
                {'name': 'energy', 'value': 950, 'decay': 0},
            ]
            e['spawn']['components']['encounter'] = {
                'observe': 0.12, 'share': 0.2, 'notice_gift': 0.4,
                'cooldowns': [{'surface': '$desk1', 'action': 'share', 'until': 6000}],
            }
        elif e['ref'] == 'desk1':
            e['spawn'].setdefault('components', {})['place'] = {
                'tags': ['desk'], 'spot': {'x': 300000, 'y': -300000}, 'facing': 1024,
            }
    d['entities'] += [
        _nest('desk1', 80000, 80000,
              portions={'items': [{'resource': 'food', 'amount': 6}, {'resource': 'water', 'amount': 6}]},
              stock={'food': 2, 'water': 2}),
        {
            'ref': 'btn1',
            'object': 'trinket_button',
            'visual': 'trinket_button',
            'spawn': {
                'name': 'trinket_button',
                'location': {'space': '$desk1', 'pos': {'x': 80000, 'y': 80000, 'z': 0}},
            },
        },
        {
            'ref': 'btn2',
            'object': 'trinket_button',
            'visual': 'trinket_button',
            'spawn': {
                'name': 'trinket_button',
                'location': {'space': '$desk1', 'pos': {'x': 82000, 'y': 80000, 'z': 0}},
            },
        },
        {
            'ref': 'gift_desk',
            'spawn': {
                'name': 'borde del escritorio',
                'location': {'space': '$desk1', 'pos': {'x': 240000, 'y': 60000, 'z': 0}},
                'components': {
                    'gift_spot': {},
                    'landmark': {},
                },
            },
        },
        {
            'ref': 't',
            'visual': 'tiny',
            'spawn': {
                'name': 't',
                'location': {'space': '$desk1', 'pos': {'x': 100000, 'y': 80000, 'z': 0}},
                'body': {'radius': 2000, 'height': 8000, 'mass': 0},
                'components': {
                    'mover': {'speed': 20000},
                    'agent': {'scale': 'tiny', 'work': 0.1, 'distance_cost': 0.2, 'sight': 300000},
                    'member': {'colony': '$nest'},
                    'needs': [
                        {'name': 'hunger', 'value': 950, 'decay': 0},
                        {'name': 'thirst', 'value': 950, 'decay': 0},
                        {'name': 'energy', 'value': 950, 'decay': 0},
                        {'name': 'fun', 'value': 900, 'decay': 0},
                    ],
                },
            },
        },
    ]
    d['expect'] = [
        {'event': 'gift_placed', 'match': {'entity': '$t', 'kind': 'button', 'spot': '$gift_desk'}},
        {'event': 'gift_noticed', 'match': {'entity': '$giantess', 'surface': '$desk1', 'kind': 'button', 'giver': '$t'}},
    ] + COMMON
    return d


def _set_need(needs_list, name, value, decay):
    for n in needs_list:
        if n.get('name') == name:
            n['value'] = value
            n['decay'] = decay
            return
    needs_list.append({'name': name, 'value': value, 'decay': decay})


def apply_to_casa():
    """Integra M61 en content/partidas/casa.json de forma idempotente:
    - necesidad hygiene en la giantess (y perfiles de random.starts);
    - ofertas wash_hands (lavamanos), shower (plato de ducha) y wash_dishes (mesada/bacha);
    - plato en la mesada para habilitar wash_dishes;
    - necesidad thirst en los tinies y meta de agua en el nido.
    """
    with open(CASA, encoding='utf-8') as f:
        d = json.load(f)

    by_ref = {e.get('ref'): e for e in d['entities']}

    # 1. Giantess: necesidad de higiene (decay 2 para que use ducha además de lavarse las manos).
    if 'giantess' in by_ref:
        g_needs = by_ref['giantess']['spawn'].setdefault('components', {}).setdefault('needs', [])
        _set_need(g_needs, 'hygiene', 500, 2)

    if 'random' in d and 'starts' in d['random']:
        hygiene_by_start = {'con hambre': 500, 'cansada': 420, 'con ganas de entretenerse': 350}
        for st in d['random']['starts']:
            val = hygiene_by_start.get(st.get('name'), 450)
            _set_need(st.setdefault('needs', []), 'hygiene', val, 2)

    # 2. Baño: lavamanos y plato de ducha con sus objetos (wash_hands y shower),
    #    un punto de búsqueda en el baño y un charco inicial junto a la ducha.
    if 'washbasin' in by_ref:
        by_ref['washbasin']['object'] = 'washbasin'
    if 'shower_tray' in by_ref:
        by_ref['shower_tray']['object'] = 'shower_tray'
        if 'bath_puddle' not in by_ref:
            idx = next(i for i, e in enumerate(d['entities']) if e.get('ref') == 'shower_tray') + 1
            d['entities'].insert(idx, {
                'ref': 'bath_puddle',
                'object': 'puddle',
                'visual': 'puddle',
                'spawn': {
                    'name': 'charco de la ducha',
                    'location': {'space': '$bathroom', 'pos': {'x': 1750000, 'y': 800000, 'z': 0}},
                },
            })
        if 'spot_bath' not in by_ref:
            idx = next(i for i, e in enumerate(d['entities']) if e.get('ref') == 'shower_tray') + 1
            d['entities'].insert(idx, {
                'ref': 'spot_bath',
                'spawn': {
                    'name': 'borde de la ducha',
                    'location': {'space': '$bathroom', 'pos': {'x': 1750000, 'y': 1100000, 'z': 0}},
                    'components': {
                        'search_spot': {},
                        'landmark': {},
                    },
                },
            })

    # 3. Cocina: oferta wash_dishes en counter (en el piso de la cocina frente a la bacha)
    #    y un plato en la mesada para habilitarla.
    if 'counter' in by_ref:
        with open(os.path.join(OBJECTS, 'water_drop.json'), encoding='utf-8') as f:
            drop_item = json.load(f)
        drop_item.pop('visual', None)
        c_offers = by_ref['counter']['spawn'].setdefault('components', {}).setdefault('offers', [])
        wash_dishes = {
            'action': 'wash_dishes',
            'scale': 'giantess',
            'effects': [{'need': 'hygiene', 'amount': 180}],
            'requires': ['cup', 'plate'],
            'approach': {'x': 950000, 'y': 400000},
            'facing': 2048,
            'steps': [
                {'do': 'goto'},
                {'do': 'use', 'ticks': 110},
                {'do': 'spawn', 'item': drop_item, 'count': 6,
                 'offset': {'x': 650000, 'y': -400000}, 'spread': 90000},
            ],
        }
        c_offers[:] = [o for o in c_offers if o.get('action') != 'wash_dishes']
        c_offers.append(wash_dishes)

        if 'kitchen_plate' not in by_ref:
            idx = next(i for i, e in enumerate(d['entities']) if e.get('ref') == 'counter') + 1
            d['entities'].insert(idx, {
                'ref': 'kitchen_plate',
                'object': 'plate',
                'visual': 'plate',
                'spawn': {
                    'name': 'plate',
                    'location': {'space': '$counter', 'pos': {'x': 350000, 'y': 800000, 'z': 0}},
                },
            })

    # 4. Nido: meta de agua y reserva inicial de agua para que no empiecen todos con sed.
    if 'nest' in by_ref:
        nest_comps = by_ref['nest']['spawn'].setdefault('components', {})
        col = nest_comps.setdefault('colony', {})
        stock = col.get('stock', 4)
        if isinstance(stock, int):
            col['stock'] = {'food': stock, 'water': 3}
        elif isinstance(stock, dict):
            stock.setdefault('food', 4)
            stock['water'] = 3
        nest_comps['portions'] = {'items': [{'resource': 'water', 'amount': 6}]}
        nest_comps.setdefault('hoard', {})

    # 5. Tinies: sed (decaimiento más lento que hambre).
    for e in d['entities']:
        comps = e.get('spawn', {}).get('components', {})
        if 'member' in comps and 'needs' in comps:
            _set_need(comps['needs'], 'thirst', 950, 1)

    # 6. M63: puntos de botín (bajo el futón, frente al placard bajo, junto al
    #    zapatero y debajo de la alfombra) y puntos de regalo (escritorio y mesa).
    m63_spots = [
        ('spot_futon', 'debajo del futón', '$bedroom', (150000, 1200000),
         {'search_spot': {'loot': DEFAULT_LOOT}, 'landmark': {}}),
        ('spot_closet', 'frente al placard bajo', '$bedroom', (700000, 700000),
         {'search_spot': {'loot': DEFAULT_LOOT}, 'landmark': {}}),
        ('spot_shoes', 'junto al zapatero', '$genkan', (500000, 1100000),
         {'search_spot': {'loot': DEFAULT_LOOT}, 'landmark': {}}),
        ('spot_rug', 'debajo de la alfombra', '$living', (1600000, 1800000),
         {'search_spot': {'loot': DEFAULT_LOOT}, 'landmark': {}}),
        ('gift_desk', 'borde del escritorio', '$desk1', (520000, 60000),
         {'gift_spot': {}, 'landmark': {}}),
        ('gift_table', 'mesa del living', '$desk2', (450000, 60000),
         {'gift_spot': {}, 'landmark': {}}),
    ]
    for ref, name, sp, (x, y), comps in m63_spots:
        if ref not in by_ref:
            d['entities'].append({
                'ref': ref,
                'spawn': {
                    'name': name,
                    'location': {'space': sp, 'pos': {'x': x, 'y': y, 'z': 0}},
                    'components': comps,
                },
            })

    with open(CASA, 'w', encoding='utf-8', newline='\n') as f:
        json.dump(d, f, indent=2, ensure_ascii=False)
        f.write('\n')
    print('actualizado', os.path.relpath(CASA, ROOT))


def main():
    for fn in (tiny_drinks_drop, drop_evaporates, water_hauled_to_nest,
               giantess_washes_hands, giantess_takes_shower,
               tiny_finds_trinket, hoard_raises_fun, giantess_notices_gift):
        _write(fn())
    apply_to_casa()


if __name__ == '__main__':
    main()
