"""Genera la integración de prendas explorables en content/partidas/casa.json
y los escenarios de cierre en tests/scenarios/ (docs/specs/prendas.md).

    python tools/scenarios/partida.py
    python tools/scenarios/vida.py
    python tools/scenarios/oficios.py
    python tools/scenarios/cuerpo.py
    python tools/scenarios/prendas.py

Es idempotente sobre content/partidas/casa.json.
"""
import json
import os

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
CASA = os.path.join(ROOT, 'content', 'partidas', 'casa.json')
SCEN = os.path.join(ROOT, 'tests', 'scenarios')

COMMON = [{'no_event': 'command_rejected'}]

GARMENT_REFS = [
    'slipper_left', 'slipper_right', 'slipper_fibers',
    'portal_slipper_l_ankle', 'portal_slipper_l_foot',
    'portal_slipper_r_ankle', 'portal_slipper_r_foot',
    'cardigan_pocket',
    'portal_cardigan_chest', 'portal_cardigan_shoulder',
]

DEFAULT_LOOT = [
    {'kind': 'button', 'weight': 2},
    {'kind': 'clip', 'weight': 2},
    {'kind': 'thread', 'weight': 4},
    {'kind': 'graphite', 'weight': 1},
]


def _grid(w_cm, h_cm, cell_cm=1, origin=(0, 0)):
    cw = max(1, w_cm // cell_cm)
    ch = max(1, h_cm // cell_cm)
    return {
        'origin': {'x': origin[0] * 10000, 'y': origin[1] * 10000},
        'cell': cell_cm * 10000,
        'version': 0,
        'rows': ['.' * cw] * ch,
    }


def _rect_centered(half_w_cm, half_h_cm):
    return {
        'min': {'x': -half_w_cm * 10000, 'y': -half_h_cm * 10000},
        'max': {'x': half_w_cm * 10000, 'y': half_h_cm * 10000},
    }


def _rect(w_cm, h_cm):
    return {'min': {'x': 0, 'y': 0}, 'max': {'x': w_cm * 10000, 'y': h_cm * 10000}}


def _portal(ref, name, sp_a, pa, sp_b, pb, disabled=False, cost=20000):
    p = {
        'a': {'space': sp_a, 'pos': {'x': pa[0], 'y': pa[1]}},
        'b': {'space': sp_b, 'pos': {'x': pb[0], 'y': pb[1]}},
        'cost': cost,
        'scale': 'tiny',
        'version': 0,
    }
    if disabled:
        p['disabled'] = True
    return {
        'ref': ref,
        'spawn': {
            'name': name,
            'components': {'portal': p},
        },
    }


def _slipper_entity(ref, name, slot, mount_part, space_ref, pos, giantess_ref='$giantess'):
    return {
        'ref': ref,
        'visual': 'slipper',
        'spawn': {
            'name': name,
            'space': {'kind': 'rigid', 'extent': _rect_centered(5, 12)},
            'location': {'space': space_ref, 'pos': {'x': pos[0], 'y': pos[1], 'z': pos[2]}},
            'components': {
                'nav_grid': _grid(10, 24, cell_cm=1, origin=(-5, -12)),
                'body_zone': {
                    'body': giantess_ref,
                    'region': name,
                    'occluded': True,
                    'grips': [
                        {'pos': {'x': 0, 'y': 95000}, 'radius': 40000, 'label': 'puntera'},
                        {'pos': {'x': 0, 'y': -100000}, 'radius': 40000, 'label': 'borde del talón'},
                    ],
                },
                'contact_surface': {
                    'tolerance': 40000,
                    'scale': 'tiny',
                    'stable_ticks': 2,
                    'require_still': True,
                    'body': giantess_ref,
                    'access': [{'x': 0, 'y': -110000}, {'x': -40000, 'y': -40000}],
                    'access_reach': 80000,
                },
                'garment': {
                    'owner': giantess_ref,
                    'slot': slot,
                    'state': 'placed',
                    'mount_part': mount_part,
                    'worn_pos': {'x': 50000, 'y': 120000, 'z': -10000},
                    'inner_pressure': {
                        'zone': {'min': {'x': -40000, 'y': -80000}, 'max': {'x': 40000, 'y': 70000}},
                        'force': 50,
                        'source': 'garment',
                    },
                    'don_warn_ticks': 20,
                },
                'landmark': {},
            },
        },
    }


def _cardigan_pocket_entity(space_ref, pos, mount_part='$body_hips', giantess_ref='$giantess'):
    return {
        'ref': 'cardigan_pocket',
        'visual': 'cardigan_pocket',
        'spawn': {
            'name': 'bolsillo del cárdigan',
            'space': {'kind': 'rigid', 'extent': _rect_centered(6, 7)},
            'location': {'space': space_ref, 'pos': {'x': pos[0], 'y': pos[1], 'z': pos[2]}},
            'components': {
                'nav_grid': _grid(12, 14, cell_cm=1, origin=(-6, -7)),
                'body_zone': {
                    'body': giantess_ref,
                    'region': 'bolsillo del cárdigan',
                    'occluded': True,
                    'grips': [
                        {'pos': {'x': 0, 'y': 55000}, 'radius': 40000, 'label': 'borde del bolsillo'},
                    ],
                },
                'contact_surface': {
                    'tolerance': 50000,
                    'scale': 'tiny',
                    'stable_ticks': 2,
                    'body': giantess_ref,
                    'access': [{'x': 0, 'y': 60000}, {'x': -50000, 'y': 0}],
                    'access_reach': 80000,
                },
                'garment': {
                    'owner': giantess_ref,
                    'slot': 'cardigan',
                    'state': 'placed',
                    'mount_part': mount_part,
                    'worn_pos': {'x': 240000, 'y': 120000, 'z': 120000},
                    'don_warn_ticks': 20,
                },
                'landmark': {},
            },
        },
    }


def with_garments(d):
    """Integra idempotentemente las prendas explorables en content/partidas/casa.json."""
    d['entities'] = [e for e in d['entities'] if e.get('ref') not in GARMENT_REFS]
    d['commands'] = [
        c for c in d['commands']
        if not (c['kind'] == 'set_components' and c['data'].get('entity') in (
            '$slipper_left', '$slipper_right', '$cardigan_pocket'))
    ]

    # Asegurar contact_surface en bedroom, bathroom, chair y futon.
    for ref in ('bedroom', 'bathroom', 'living', 'chair', 'futon'):
        for e in d['entities']:
            if e.get('ref') == ref:
                e['spawn'].setdefault('components', {}).setdefault('contact_surface', {'tolerance': 50000})

    d['entities'].extend([
        _slipper_entity('slipper_left', 'pantufla izquierda', 'slipper_left', '$body_foot_l',
                        '$bedroom', (1550000, 1450000, 0)),
        _slipper_entity('slipper_right', 'pantufla derecha', 'slipper_right', '$body_foot_r',
                        '$bedroom', (1700000, 1450000, 0)),
        {
            'ref': 'slipper_fibers',
            'spawn': {
                'name': 'fibras de pantufla',
                'location': {'space': '$slipper_left', 'pos': {'x': 0, 'y': 90000, 'z': 0}},
                'components': {
                    'search_spot': {'look_ticks': 40, 'loot': DEFAULT_LOOT},
                    'landmark': {},
                },
            },
        },
        _portal('portal_slipper_l_ankle', 'pantufla izquierda a tobillo',
                '$slipper_left', (0, -90000), '$body_calf_l', (40000, 20000), disabled=True),
        _portal('portal_slipper_l_foot', 'pantufla izquierda a pie',
                '$slipper_left', (0, 0), '$body_foot_l', (50000, 120000), disabled=True),
        _portal('portal_slipper_r_ankle', 'pantufla derecha a tobillo',
                '$slipper_right', (0, -90000), '$body_calf_r', (40000, 20000), disabled=True),
        _portal('portal_slipper_r_foot', 'pantufla derecha a pie',
                '$slipper_right', (0, 0), '$body_foot_r', (50000, 120000), disabled=True),

        _cardigan_pocket_entity('$chair', (200000, 200000, 0)),
        _portal('portal_cardigan_chest', 'bolsillo a pecho',
                '$cardigan_pocket', (0, 50000), '$body_chest', (200000, 140000), disabled=True),
        _portal('portal_cardigan_shoulder', 'bolsillo a hombro izquierdo',
                '$cardigan_pocket', (-40000, 50000), '$body_shoulder_l', (60000, 60000), disabled=True),
    ])

    d['commands'].extend([
        {'tick': 0, 'kind': 'set_components', 'data': {
            'entity': '$slipper_left',
            'components': {'garment': {
                'owner': '$giantess',
                'slot': 'slipper_left',
                'state': 'placed',
                'mount_part': '$body_foot_l',
                'worn_pos': {'x': 50000, 'y': 120000, 'z': -10000},
                'body_portals': ['$portal_slipper_l_ankle', '$portal_slipper_l_foot'],
                'inner_pressure': {
                    'zone': {'min': {'x': -40000, 'y': -80000}, 'max': {'x': 40000, 'y': 70000}},
                    'force': 50,
                    'source': 'garment',
                },
                'don_warn_ticks': 20,
            }},
        }},
        {'tick': 0, 'kind': 'set_components', 'data': {
            'entity': '$slipper_right',
            'components': {'garment': {
                'owner': '$giantess',
                'slot': 'slipper_right',
                'state': 'placed',
                'mount_part': '$body_foot_r',
                'worn_pos': {'x': 50000, 'y': 120000, 'z': -10000},
                'body_portals': ['$portal_slipper_r_ankle', '$portal_slipper_r_foot'],
                'inner_pressure': {
                    'zone': {'min': {'x': -40000, 'y': -80000}, 'max': {'x': 40000, 'y': 70000}},
                    'force': 50,
                    'source': 'garment',
                },
                'don_warn_ticks': 20,
            }},
        }},
        {'tick': 0, 'kind': 'set_components', 'data': {
            'entity': '$cardigan_pocket',
            'components': {'garment': {
                'owner': '$giantess',
                'slot': 'cardigan',
                'state': 'placed',
                'mount_part': '$body_hips',
                'worn_pos': {'x': 240000, 'y': 120000, 'z': 120000},
                'body_portals': ['$portal_cardigan_chest', '$portal_cardigan_shoulder'],
                'don_warn_ticks': 20,
            }},
        }},
    ])

    # Actualizar sleep en bed para sacarse pantuflas al acostarse y calzarlas al despertar.
    for e in d['entities']:
        if e.get('ref') == 'bed':
            e['spawn'].setdefault('components', {})['offers'] = [{
                'action': 'sleep',
                'scale': 'giantess',
                'effects': [{'need': 'energy', 'amount': 800}],
                'steps': [
                    {'do': 'goto'},
                    {'do': 'doff', 'slot': 'slipper_left', 'offset': {'x': -180000, 'y': -250000}},
                    {'do': 'doff', 'slot': 'slipper_right', 'offset': {'x': 180000, 'y': -250000}},
                    {'do': 'pose', 'pose': 'lie_down'},
                    {'do': 'use', 'ticks': 600},
                    {'do': 'pose', 'pose': 'idle'},
                    {'do': 'don', 'slot': 'slipper_left'},
                    {'do': 'don', 'slot': 'slipper_right'},
                ],
            }]
        elif e.get('ref') == 'desk1':
            offers = e['spawn'].get('components', {}).get('offers', [])
            for o in offers:
                if o.get('action') == 'write':
                    o['steps'] = [
                        {'do': 'goto'},
                        {'do': 'doff', 'slot': 'cardigan', 'place': 'chair'},
                        {'do': 'pose', 'pose': 'sit_hover'},
                        {'do': 'wait', 'ticks': 16},
                        {'do': 'pose', 'pose': 'sit'},
                        {'do': 'pose', 'pose': 'sit_hand_hover'},
                        {'do': 'wait', 'ticks': 12},
                        {'do': 'pose', 'pose': 'sit_hand_rest'},
                        {'do': 'use', 'ticks': 400},
                        {'do': 'don', 'slot': 'cardigan'},
                    ]

    d['notes'] = d['notes'].split(' Prendas:')[0] + (
        ' Prendas: pantuflas explorables junto al futón y bolsillo del cárdigan (tools/scenarios/prendas.py).')


def _house_two_rooms():
    return [
        {'ref': 'house', 'spawn': {'name': 'house', 'space': {'kind': 'static', 'extent': _rect(0, 0)}}},
        {'ref': 'bedroom', 'visual': 'bedroom', 'spawn': {
            'name': 'dormitorio',
            'space': {'kind': 'static', 'extent': _rect(300, 300)},
            'location': {'space': '$house', 'pos': {'x': 0, 'y': 0, 'z': 0}},
            'components': {'nav_grid': _grid(300, 300, cell_cm=10), 'contact_surface': {'tolerance': 40000}},
        }},
        {'ref': 'bathroom', 'visual': 'apt_bathroom', 'spawn': {
            'name': 'baño',
            'space': {'kind': 'static', 'extent': _rect(300, 300)},
            'location': {'space': '$house', 'pos': {'x': 3000000, 'y': 0, 'z': 0}},
            'components': {'nav_grid': _grid(300, 300, cell_cm=10), 'contact_surface': {'tolerance': 40000}},
        }},
        {'ref': 'door_bed_bath', 'spawn': {
            'name': 'puerta al baño',
            'components': {
                'portal': {
                    'a': {'space': '$bedroom', 'pos': {'x': 2900000, 'y': 1500000}},
                    'b': {'space': '$bathroom', 'pos': {'x': 100000, 'y': 1500000}},
                    'cost': 10000,
                    'version': 0,
                },
            },
        }},
    ]


def _simple_giantess(space_ref='$bedroom', pos=(1500000, 1500000, 0), care=0.0):
    agent = {'scale': 'giantess', 'distance_cost': 0.05, 'watch': 2500000}
    if care > 0:
        agent['care'] = care
    return [
        {'ref': 'giantess', 'visual': 'giantess', 'spawn': {
            'name': 'Giantess',
            'space': {'kind': 'body', 'extent': _rect(0, 0)},
            'location': {'space': space_ref, 'pos': {'x': pos[0], 'y': pos[1], 'z': pos[2]}},
            'components': {
                'mover': {'speed': 1000000, 'scale': 'giantess'},
                'agent': agent,
                'needs': [
                    {'name': 'energy', 'value': 900, 'decay': 0},
                    {'name': 'hygiene', 'value': 900, 'decay': 0},
                    {'name': 'hunger', 'value': 900, 'decay': 0},
                ],
            },
        }},
        {'ref': 'body_foot_l', 'visual': 'body_foot_l', 'spawn': {
            'name': 'pie izquierdo',
            'space': {'kind': 'articulated', 'extent': _rect(10, 24)},
            'location': {'space': '$giantess', 'pos': {'x': -150000, 'y': -80000, 'z': 20000}},
            'components': {'nav_grid': _grid(10, 24, cell_cm=2)},
        }},
        {'ref': 'body_foot_r', 'visual': 'body_foot_r', 'spawn': {
            'name': 'pie derecho',
            'space': {'kind': 'articulated', 'extent': _rect(10, 24)},
            'location': {'space': '$giantess', 'pos': {'x': 50000, 'y': -80000, 'z': 20000}},
            'components': {'nav_grid': _grid(10, 24, cell_cm=2)},
        }},
        {'ref': 'body_calf_l', 'visual': 'body_calf_l', 'spawn': {
            'name': 'pantorrilla izquierda',
            'space': {'kind': 'articulated', 'extent': _rect(8, 16)},
            'location': {'space': '$giantess', 'pos': {'x': -140000, 'y': -40000, 'z': 120000}},
            'components': {'nav_grid': _grid(8, 16, cell_cm=2)},
        }},
    ]


def garment_ride_and_doff():
    return {
        'name': 'garment_ride_and_doff',
        'seed': 201,
        'ticks': 500,
        'notes': 'P1: un tiny sube a la pantufla apoyada, la giantess se la calza, camina al baño y se la saca; el tiny desembarca.',
        'entities': _house_two_rooms() + _simple_giantess() + [
            _slipper_entity('slipper_left', 'pantufla izquierda', 'slipper_left', '$body_foot_l',
                            '$bedroom', (1300000, 1500000, 0)),
            {'ref': 'tiny1', 'visual': 'tiny', 'spawn': {
                'name': 'Tiny 01',
                'location': {'space': '$bedroom', 'pos': {'x': 1200000, 'y': 1400000, 'z': 0}},
                'body': {'radius': 2000, 'height': 8000, 'mass': 1},
                'components': {'mover': {'speed': 80000, 'scale': 'tiny'}, 'fragile': {}},
            }},
        ],
        'commands': [
            {'tick': 10, 'kind': 'move_to', 'data': {'entity': '$tiny1', 'space': '$slipper_left', 'goal': {'x': 0, 'y': 90000}}},
            {'tick': 120, 'kind': 'don_garment', 'data': {'actor': '$giantess', 'garment': '$slipper_left', 'warn_ticks': 15}},
            {'tick': 150, 'kind': 'move_to', 'data': {'entity': '$giantess', 'space': '$bathroom', 'goal': {'x': 1200000, 'y': 1500000}}},
            {'tick': 320, 'kind': 'doff_garment', 'data': {
                'actor': '$giantess', 'garment': '$slipper_left', 'target_space': '$bathroom',
                'pos': {'x': 1000000, 'y': 1500000, 'z': 0},
            }},
            {'tick': 350, 'kind': 'move_to', 'data': {'entity': '$tiny1', 'space': '$bathroom', 'goal': {'x': 1300000, 'y': 1500000}}},
        ],
        'expect': COMMON + [
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$slipper_left'}},
            {'event': 'garment_donning', 'match': {'garment': '$slipper_left'}},
            {'event': 'garment_worn', 'match': {'garment': '$slipper_left'}},
            {'event': 'garment_doffed', 'match': {'garment': '$slipper_left', 'surface': '$bathroom'}},
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$bathroom'}},
            {'at': '$tiny1', 'space': '$bathroom', 'pos': {'x': 1300000, 'y': 1500000}, 'tol': 40000},
        ],
    }


def garment_donning_flee_and_crush():
    return {
        'name': 'garment_donning_flee_and_crush',
        'atlas': '../../content/art/casa.json',
        'seed': 202,
        'ticks': 100,
        'notes': 'P1/P3: durante donning, un tiny con reflejos huye de la plantilla hacia la puntera segura y otro sin reflejos muere aplastado.',
        'entities': [
            {'ref': 'house', 'spawn': {'name': 'house', 'space': {'kind': 'static', 'extent': _rect(0, 0)}}},
            {'ref': 'bedroom', 'visual': 'bedroom', 'spawn': {
                'name': 'dormitorio',
                'space': {'kind': 'static', 'extent': _rect(300, 300)},
                'location': {'space': '$house', 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {'nav_grid': _grid(300, 300, cell_cm=10), 'contact_surface': {'tolerance': 40000}},
            }},
        ] + _simple_giantess() + [
            _slipper_entity('slipper_left', 'pantufla izquierda', 'slipper_left', '$body_foot_l',
                            '$bedroom', (1400000, 1450000, 0)),
            {'ref': 'tiny1', 'visual': 'tiny', 'spawn': {
                'name': 'Tiny 01',
                'location': {'space': '$slipper_left', 'pos': {'x': 0, 'y': 45000, 'z': 0}},
                'body': {'radius': 2000, 'height': 8000, 'mass': 1},
                'components': {
                    'mover': {'speed': 40000, 'scale': 'tiny'},
                    'agent': {'scale': 'tiny', 'reflex': 3},
                    'needs': [{'name': 'energy', 'value': 900, 'decay': 0}],
                    'fragile': {},
                },
            }},
            {'ref': 'tiny2', 'visual': 'tiny_lavender', 'spawn': {
                'name': 'Tiny 02',
                'location': {'space': '$slipper_left', 'pos': {'x': 0, 'y': -20000, 'z': 0}},
                'body': {'radius': 2000, 'height': 8000, 'mass': 1},
                'components': {'fragile': {}},
            }},
        ],
        'commands': [
            {'tick': 10, 'kind': 'don_garment', 'data': {'actor': '$giantess', 'garment': '$slipper_left', 'warn_ticks': 25}},
        ],
        'expect': COMMON + [
            {'event': 'alarmed', 'match': {'entity': '$tiny1', 'source': 'garment'}},
            {'event': 'garment_worn', 'match': {'garment': '$slipper_left'}},
            {'event': 'crushed', 'match': {'entity': '$tiny2', 'source': 'garment'}},
            {'no_event': 'crushed', 'match': {'entity': '$tiny1'}},
        ],
    }


def giantess_slipper_routine():
    return {
        'name': 'giantess_slipper_routine',
        'seed': 203,
        'ticks': 1500,
        'notes': 'P2: la giantess duerme sin pantuflas, se las calza al despertar, va al baño y se las saca antes de ducharse.',
        'entities': _house_two_rooms() + _simple_giantess(pos=(1000000, 1500000, 0)) + [
            _slipper_entity('slipper_left', 'pantufla izquierda', 'slipper_left', '$body_foot_l',
                            '$bedroom', (1150000, 1450000, 0)),
            _slipper_entity('slipper_right', 'pantufla derecha', 'slipper_right', '$body_foot_r',
                            '$bedroom', (1250000, 1450000, 0)),
            {'ref': 'bed', 'object': 'bed', 'visual': 'bed', 'spawn': {
                'name': 'cama',
                'location': {'space': '$bedroom', 'pos': {'x': 1000000, 'y': 1500000, 'z': 0}},
            }},
            {'ref': 'shower_tray', 'object': 'shower_tray', 'visual': 'shower_tray', 'spawn': {
                'name': 'plato de ducha',
                'location': {'space': '$bathroom', 'pos': {'x': 2000000, 'y': 1000000, 'z': 0}},
            }},
        ],
        'commands': [
            {'tick': 0, 'kind': 'set_need', 'data': {'entity': '$giantess', 'need': 'energy', 'value': 150}},
            {'tick': 750, 'kind': 'set_need', 'data': {'entity': '$giantess', 'need': 'hygiene', 'value': 100}},
        ],
        'expect': COMMON + [
            {'event': 'action_done', 'match': {'entity': '$giantess', 'what': 'sleep'}},
            {'event': 'garment_worn', 'match': {'garment': '$slipper_left'}},
            {'event': 'garment_worn', 'match': {'garment': '$slipper_right'}},
            {'event': 'garment_doffed', 'match': {'garment': '$slipper_left', 'surface': '$bathroom'}},
            {'event': 'action_done', 'match': {'entity': '$giantess', 'what': 'shower'}},
        ],
    }


def careful_giantess_checks_slipper():
    return {
        'name': 'careful_giantess_checks_slipper',
        'seed': 204,
        'ticks': 400,
        'notes': 'P2: con agent.care alto, la giantess nota al tiny en la pantufla, titubea y extiende el aviso para que escape sin morir.',
        'entities': [
            {'ref': 'house', 'spawn': {'name': 'house', 'space': {'kind': 'static', 'extent': _rect(0, 0)}}},
            {'ref': 'bedroom', 'visual': 'bedroom', 'spawn': {
                'name': 'dormitorio',
                'space': {'kind': 'static', 'extent': _rect(300, 300)},
                'location': {'space': '$house', 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {'nav_grid': _grid(300, 300, cell_cm=10), 'contact_surface': {'tolerance': 40000}},
            }},
        ] + _simple_giantess(pos=(1000000, 1500000, 0), care=1.0) + [
            _slipper_entity('slipper_left', 'pantufla izquierda', 'slipper_left', '$body_foot_l',
                            '$bedroom', (1150000, 1500000, 0)),
            {'ref': 'bed', 'visual': 'bed', 'spawn': {
                'name': 'cama',
                'location': {'space': '$bedroom', 'pos': {'x': 1000000, 'y': 1500000, 'z': 0}},
                'components': {'offers': [{
                    'action': 'sleep',
                    'scale': 'giantess',
                    'effects': [{'need': 'energy', 'amount': 700}],
                    'steps': [
                        {'do': 'goto'},
                        {'do': 'use', 'ticks': 15},
                        {'do': 'don', 'slot': 'slipper_left'},
                    ],
                }]},
            }},
            {'ref': 'tiny1', 'visual': 'tiny', 'spawn': {
                'name': 'Tiny 01',
                'location': {'space': '$slipper_left', 'pos': {'x': 0, 'y': 20000, 'z': 0}},
                'body': {'radius': 2000, 'height': 8000, 'mass': 1},
                'components': {
                    'mover': {'speed': 40000, 'scale': 'tiny'},
                    'agent': {'scale': 'tiny', 'reflex': 4},
                    'needs': [{'name': 'energy', 'value': 900, 'decay': 0}],
                    'fragile': {},
                },
            }},
        ],
        'commands': [
            {'tick': 0, 'kind': 'set_need', 'data': {'entity': '$giantess', 'need': 'energy', 'value': 150}},
        ],
        'expect': COMMON + [
            {'event': 'hesitated', 'match': {'entity': '$giantess', 'surface': '$slipper_left', 'source': 'garment'}},
            {'event': 'garment_donning', 'match': {'garment': '$slipper_left'}},
            {'event': 'garment_worn', 'match': {'garment': '$slipper_left'}},
            {'no_event': 'crushed'},
        ],
    }


def cardigan_pocket_ride():
    return {
        'name': 'cardigan_pocket_ride',
        'seed': 205,
        'ticks': 500,
        'notes': 'P4: un tiny entra al bolsillo del cárdigan en la silla, la giantess se lo pone y lo lleva al baño; otro tiny sale por el portal al pecho.',
        'entities': _house_two_rooms() + _simple_giantess(pos=(1500000, 1500000, 0)) + [
            {'ref': 'body_hips', 'visual': 'body_hips', 'spawn': {
                'name': 'cintura y caderas',
                'space': {'kind': 'articulated', 'extent': _rect(32, 20)},
                'location': {'space': '$giantess', 'pos': {'x': -160000, 'y': -100000, 'z': 950000}},
                'components': {'nav_grid': _grid(32, 20, cell_cm=4)},
            }},
            {'ref': 'body_chest', 'visual': 'body_chest', 'spawn': {
                'name': 'pecho',
                'space': {'kind': 'articulated', 'extent': _rect(28, 36)},
                'location': {'space': '$giantess', 'pos': {'x': -140000, 'y': -180000, 'z': 1040000}},
                'components': {
                    'nav_grid': _grid(28, 36, cell_cm=4),
                    'body_zone': {'body': '$giantess', 'region': 'pecho'},
                },
            }},
            {'ref': 'chair', 'object': 'chair', 'visual': 'chair', 'spawn': {
                'name': 'silla',
                'location': {'space': '$bedroom', 'pos': {'x': 1100000, 'y': 1400000, 'z': 450000}},
            }},
            _cardigan_pocket_entity('$chair', (200000, 200000, 0)),
            _portal('portal_cardigan_chest', 'bolsillo a pecho',
                    '$cardigan_pocket', (0, 50000), '$body_chest', (200000, 140000), disabled=True),
            {'ref': 'tiny1', 'visual': 'tiny', 'spawn': {
                'name': 'Tiny 01',
                'location': {'space': '$chair', 'pos': {'x': 120000, 'y': 150000, 'z': 0}},
                'body': {'radius': 2000, 'height': 8000, 'mass': 1},
                'components': {'mover': {'speed': 80000, 'scale': 'tiny'}, 'fragile': {}},
            }},
            {'ref': 'tiny2', 'visual': 'tiny_ochre', 'spawn': {
                'name': 'Tiny 02',
                'location': {'space': '$cardigan_pocket', 'pos': {'x': 0, 'y': -20000, 'z': 0}},
                'body': {'radius': 2000, 'height': 8000, 'mass': 1},
                'components': {'mover': {'speed': 80000, 'scale': 'tiny'}, 'fragile': {}},
            }},
        ],
        'commands': [
            {'tick': 0, 'kind': 'set_components', 'data': {
                'entity': '$cardigan_pocket',
                'components': {'garment': {
                    'owner': '$giantess',
                    'slot': 'cardigan',
                    'state': 'placed',
                    'mount_part': '$body_hips',
                    'worn_pos': {'x': 240000, 'y': 120000, 'z': 120000},
                    'body_portals': ['$portal_cardigan_chest'],
                    'don_warn_ticks': 10,
                }},
            }},
            {'tick': 10, 'kind': 'move_to', 'data': {'entity': '$tiny1', 'space': '$cardigan_pocket', 'goal': {'x': 0, 'y': 0}}},
            {'tick': 100, 'kind': 'don_garment', 'data': {'actor': '$giantess', 'garment': '$cardigan_pocket', 'warn_ticks': 10}},
            {'tick': 130, 'kind': 'move_to', 'data': {'entity': '$tiny2', 'space': '$body_chest', 'goal': {'x': 200000, 'y': 140000}}},
            {'tick': 140, 'kind': 'move_to', 'data': {'entity': '$giantess', 'space': '$bathroom', 'goal': {'x': 1200000, 'y': 1500000}}},
            {'tick': 320, 'kind': 'doff_garment', 'data': {
                'actor': '$giantess', 'garment': '$cardigan_pocket', 'target_space': '$bathroom',
                'pos': {'x': 1000000, 'y': 1500000, 'z': 0},
            }},
            {'tick': 350, 'kind': 'move_to', 'data': {'entity': '$tiny1', 'space': '$bathroom', 'goal': {'x': 1300000, 'y': 1500000}}},
        ],
        'expect': COMMON + [
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$cardigan_pocket'}},
            {'event': 'garment_worn', 'match': {'garment': '$cardigan_pocket'}},
            {'event': 'entity_reparented', 'match': {'entity': '$tiny2', 'to': '$body_chest'}},
            {'event': 'garment_doffed', 'match': {'garment': '$cardigan_pocket', 'surface': '$bathroom'}},
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$bathroom'}},
            {'no_event': 'crushed'},
        ],
    }


def _write(path, d):
    with open(path, 'w', encoding='utf-8', newline='\n') as f:
        f.write(json.dumps(d, ensure_ascii=False, indent=2) + '\n')
    print('escrito', os.path.relpath(path, ROOT))


def main():
    with open(CASA, encoding='utf-8') as f:
        casa = json.load(f)
    with_garments(casa)
    _write(CASA, casa)
    for fn in (
        garment_ride_and_doff,
        garment_donning_flee_and_crush,
        giantess_slipper_routine,
        careful_giantess_checks_slipper,
        cardigan_pocket_ride,
    ):
        d = fn()
        _write(os.path.join(SCEN, d['name'] + '.json'), d)


if __name__ == '__main__':
    main()
