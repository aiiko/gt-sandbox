"""Genera la integración de los 5 quick wins en content/partidas/casa.json
y sus escenarios de cierre en tests/scenarios/ (docs/ideas.md).

    python tools/scenarios/partida.py
    python tools/scenarios/vida.py
    python tools/scenarios/oficios.py
    python tools/scenarios/cuerpo.py
    python tools/scenarios/prendas.py
    python tools/scenarios/quickwins.py

Es idempotente sobre content/partidas/casa.json.
"""
import copy
import json
import os

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
CASA = os.path.join(ROOT, 'content', 'partidas', 'casa.json')
SCEN = os.path.join(ROOT, 'tests', 'scenarios')

COMMON = [{'no_event': 'command_rejected'}]

QW2_REFS = {
    'sock_left', 'sock_right', 'sock_fibers',
    'portal_sock_l_ankle', 'portal_sock_l_foot',
    'portal_sock_r_ankle', 'portal_sock_r_foot',
    'hair_tie', 'portal_hair_tie_head',
    'street_shoe_l', 'street_shoe_r', 'street_shoe_spot',
    'genkan_bag', 'genkan_bag_spot', 'genkan_bag_target',
    'genkan_bag_crumb_1', 'genkan_bag_crumb_2',
    'spot_bedroom_door', 'spot_hallway_bath',
}

QW2_GARMENT_ENTITIES = {'$sock_left', '$sock_right', '$hair_tie'}

GENKAN_LOOT = [
    {'kind': 'coin', 'weight': 4},
    {'kind': 'clip', 'weight': 3},
    {'kind': 'thread', 'weight': 3},
    {'kind': 'button', 'weight': 2},
]


def _entity(d, ref):
    return next(e for e in d['entities'] if e.get('ref') == ref)


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
    return {'ref': ref, 'spawn': {'name': name, 'components': {'portal': p}}}


def _sock_entity(ref, name, slot, mount_part, pos, state='worn'):
    is_worn = state == 'worn'
    loc = {'space': mount_part, 'pos': {'x': 50000, 'y': 120000, 'z': -5000}} if is_worn else {'space': '$bedroom', 'pos': {'x': pos[0], 'y': pos[1], 'z': pos[2]}}
    return {
        'ref': ref,
        'visual': 'sock',
        'spawn': {
            'name': name,
            'space': {'kind': 'rigid', 'extent': _rect_centered(4, 11)},
            'location': loc,
            'components': {
                'nav_grid': _grid(8, 22, cell_cm=1, origin=(-4, -11)),
                'body_zone': {
                    'body': '$giantess',
                    'region': name,
                    'occluded': True,
                    'grips': [
                        {'pos': {'x': 0, 'y': 90000}, 'radius': 35000, 'label': 'puntera de la media'},
                        {'pos': {'x': 0, 'y': -95000}, 'radius': 35000, 'label': 'puño de la media'},
                    ],
                },
                'contact_surface': {
                    'tolerance': 40000,
                    'scale': 'tiny',
                    'stable_ticks': 2,
                    'require_still': True,
                    'body': '$giantess',
                    'access': [{'x': 0, 'y': -100000}, {'x': -35000, 'y': -30000}],
                    'access_reach': 80000,
                },
                'garment': {
                    'owner': '$giantess',
                    'slot': slot,
                    'state': state,
                    'mount_part': mount_part,
                    'worn_pos': {'x': 50000, 'y': 120000, 'z': -5000},
                    'inner_pressure': {
                        'zone': {'min': {'x': -35000, 'y': -75000}, 'max': {'x': 35000, 'y': 65000}},
                        'force': 35,
                        'source': 'garment',
                    },
                    'don_warn_ticks': 15,
                },
            },
        },
    }


def _street_shoe_entity(ref, name, pos):
    return {
        'ref': ref,
        'visual': 'street_shoe',
        'spawn': {
            'name': name,
            'space': {'kind': 'rigid', 'extent': _rect_centered(5, 13)},
            'location': {'space': '$genkan', 'pos': {'x': pos[0], 'y': pos[1], 'z': pos[2]}},
            'components': {
                'nav_grid': _grid(10, 26, cell_cm=1, origin=(-5, -13)),
                'body_zone': {
                    'region': name,
                    'occluded': True,
                    'grips': [
                        {'pos': {'x': 0, 'y': -100000}, 'radius': 40000, 'label': 'talón del zapato'},
                    ],
                },
                'contact_surface': {
                    'tolerance': 40000,
                    'scale': 'tiny',
                    'stable_ticks': 2,
                    'access': [{'x': 0, 'y': -120000}, {'x': -40000, 'y': -40000}],
                    'access_reach': 80000,
                },
                'landmark': {},
            },
        },
    }


def _genkan_bag_entity():
    return {
        'ref': 'genkan_bag',
        'visual': 'genkan_bag',
        'spawn': {
            'name': 'bolso de la entrada',
            'space': {'kind': 'rigid', 'extent': _rect_centered(13, 9)},
            'location': {'space': '$genkan', 'pos': {'x': 1150000, 'y': 600000, 'z': 0}},
            'components': {
                'nav_grid': _grid(26, 18, cell_cm=1, origin=(-13, -9)),
                'body_zone': {
                    'region': 'bolso de la entrada',
                    'occluded': True,
                    'grips': [
                        {'pos': {'x': 0, 'y': 75000}, 'radius': 45000, 'label': 'solapa del bolso'},
                    ],
                },
                'contact_surface': {
                    'tolerance': 40000,
                    'scale': 'tiny',
                    'stable_ticks': 2,
                    'access': [{'x': -120000, 'y': 0}, {'x': 0, 'y': 80000}],
                    'access_reach': 80000,
                },
                'landmark': {},
            },
        },
    }


def with_quickwins(d):
    """Integra idempotentemente los quick wins en content/partidas/casa.json."""
    # QW1: transporte deliberado de pasajeros en la palma.
    g = _entity(d, 'giantess')
    enc = g['spawn']['components'].setdefault('encounter', {})
    enc['carry'] = 0.18

    # QW2: limpiar entidades y comandos previos de QW2 para idempotencia.
    d['entities'] = [e for e in d['entities'] if e.get('ref') not in QW2_REFS]
    d['commands'] = [
        c for c in d['commands']
        if not (c['kind'] == 'set_components' and c['data'].get('entity') in QW2_GARMENT_ENTITIES)
    ]

    # Sumar moneda al botín del zapatero del genkan si existe, y hacer visible el charco del baño dentro del cuarto.
    for e in d['entities']:
        if e.get('ref') == 'spot_shoes':
            sp = e['spawn']['components'].get('search_spot', {})
            loot = sp.get('loot', [])
            if not any(it.get('kind') == 'coin' for it in loot):
                loot.append({'kind': 'coin', 'weight': 3})
        elif e.get('ref') == 'bath_puddle':
            e['spawn'].setdefault('components', {})['landmark'] = {'range': 1800000}

    d['entities'].extend([
        _sock_entity('sock_left', 'media izquierda', 'sock_left', '$body_foot_l', (1550000, 1750000, 0)),
        _sock_entity('sock_right', 'media derecha', 'sock_right', '$body_foot_r', (1700000, 1750000, 0)),
        {
            'ref': 'sock_fibers',
            'spawn': {
                'name': 'pelusa de la media',
                'location': {'space': '$sock_left', 'pos': {'x': 0, 'y': 80000, 'z': 0}},
                'components': {
                    'search_spot': {'look_ticks': 40, 'loot': [
                        {'kind': 'thread', 'weight': 4},
                        {'kind': 'button', 'weight': 2},
                    ]},
                },
            },
        },
        _portal('portal_sock_l_ankle', 'media izquierda a tobillo',
                '$sock_left', (0, -85000), '$body_calf_l', (40000, 20000), disabled=False),
        _portal('portal_sock_l_foot', 'media izquierda a pie',
                '$sock_left', (0, 0), '$body_foot_l', (50000, 120000), disabled=False),
        _portal('portal_sock_r_ankle', 'media derecha a tobillo',
                '$sock_right', (0, -85000), '$body_calf_r', (40000, 20000), disabled=False),
        _portal('portal_sock_r_foot', 'media derecha a pie',
                '$sock_right', (0, 0), '$body_foot_r', (50000, 120000), disabled=False),

        {
            'ref': 'hair_tie',
            'visual': 'hair_tie',
            'spawn': {
                'name': 'coletero del cabello',
                'space': {'kind': 'rigid', 'extent': _rect_centered(3, 3)},
                'location': {'space': '$body_head', 'pos': {'x': 100000, 'y': 110000, 'z': 20000}},
                'components': {
                    'nav_grid': _grid(6, 6, cell_cm=1, origin=(-3, -3)),
                    'body_zone': {
                        'body': '$giantess',
                        'region': 'coletero del cabello',
                        'occluded': True,
                        'grips': [
                            {'pos': {'x': 0, 'y': 0}, 'radius': 30000, 'label': 'lazo del coletero'},
                        ],
                    },
                    'shelter': {'strength': 60, 'radius': 30000},
                    'contact_surface': {
                        'tolerance': 40000,
                        'scale': 'tiny',
                        'stable_ticks': 2,
                        'body': '$giantess',
                        'access': [{'x': 0, 'y': -25000}, {'x': -25000, 'y': 0}],
                        'access_reach': 60000,
                    },
                    'garment': {
                        'owner': '$giantess',
                        'slot': 'hair_tie',
                        'state': 'worn',
                        'mount_part': '$body_head',
                        'worn_pos': {'x': 100000, 'y': 110000, 'z': 20000},
                        'don_warn_ticks': 5,
                    },
                },
            },
        },
        _portal('portal_hair_tie_head', 'coletero a cabello',
                '$hair_tie', (0, 0), '$body_head', (100000, 110000), disabled=False),

        _street_shoe_entity('street_shoe_l', 'zapato de calle izquierdo', (650000, 900000, 0)),
        _street_shoe_entity('street_shoe_r', 'zapato de calle derecho', (820000, 900000, 0)),
        {
            'ref': 'street_shoe_spot',
            'spawn': {
                'name': 'puntera del zapato de calle',
                'location': {'space': '$street_shoe_l', 'pos': {'x': 0, 'y': 70000, 'z': 0}},
                'components': {
                    'search_spot': {'look_ticks': 40, 'loot': GENKAN_LOOT},
                    'landmark': {},
                },
            },
        },
        _genkan_bag_entity(),
        {
            'ref': 'genkan_bag_spot',
            'spawn': {
                'name': 'interior del bolso',
                'location': {'space': '$genkan_bag', 'pos': {'x': 30000, 'y': 0, 'z': 0}},
                'components': {
                    'search_spot': {'look_ticks': 40, 'loot': GENKAN_LOOT},
                    'landmark': {},
                },
            },
        },
        {
            'ref': 'genkan_bag_target',
            'spawn': {
                'name': 'bolso de la entrada',
                'location': {'space': '$genkan_bag', 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'expedition_target': {
                        'colony': '$nest',
                        'clue': '$genkan_climb_top',
                        'min_members': 2,
                        'max_members': 3,
                        'waypoints': ['$threshold_refuge'],
                        'max_duration': 7200,
                    },
                    'landmark': {'range': 250000},
                },
            },
        },
        {
            'ref': 'genkan_bag_crumb_1',
            'object': 'crumb',
            'visual': 'crumb',
            'spawn': {
                'name': 'migaja del bolso',
                'location': {'space': '$genkan_bag', 'pos': {'x': -50000, 'y': 20000, 'z': 0}},
            },
        },
        {
            'ref': 'genkan_bag_crumb_2',
            'object': 'crumb',
            'visual': 'crumb',
            'spawn': {
                'name': 'migaja del bolso',
                'location': {'space': '$genkan_bag', 'pos': {'x': 60000, 'y': -30000, 'z': 0}},
            },
        },
        {
            'ref': 'spot_bedroom_door',
            'spawn': {
                'name': 'umbral del dormitorio',
                'location': {'space': '$bedroom', 'pos': {'x': 2150000, 'y': 150000, 'z': 0}},
                'components': {
                    'search_spot': {},
                    'landmark': {'range': 2800000},
                },
            },
        },
        {
            'ref': 'spot_hallway_bath',
            'spawn': {
                'name': 'umbral del baño',
                'location': {'space': '$hallway', 'pos': {'x': 2200000, 'y': 150000, 'z': 0}},
                'components': {
                    'search_spot': {},
                    'landmark': {'range': 2500000},
                },
            },
        },
    ])

    d['commands'].extend([
        {'tick': 0, 'kind': 'set_components', 'data': {
            'entity': '$sock_left',
            'components': {'garment': {
                'owner': '$giantess',
                'slot': 'sock_left',
                'state': 'worn',
                'mount_part': '$body_foot_l',
                'worn_pos': {'x': 50000, 'y': 120000, 'z': -5000},
                'body_portals': ['$portal_sock_l_ankle', '$portal_sock_l_foot'],
                'inner_pressure': {
                    'zone': {'min': {'x': -35000, 'y': -75000}, 'max': {'x': 35000, 'y': 65000}},
                    'force': 35,
                    'source': 'garment',
                },
                'don_warn_ticks': 15,
            }},
        }},
        {'tick': 0, 'kind': 'set_components', 'data': {
            'entity': '$sock_right',
            'components': {'garment': {
                'owner': '$giantess',
                'slot': 'sock_right',
                'state': 'worn',
                'mount_part': '$body_foot_r',
                'worn_pos': {'x': 50000, 'y': 120000, 'z': -5000},
                'body_portals': ['$portal_sock_r_ankle', '$portal_sock_r_foot'],
                'inner_pressure': {
                    'zone': {'min': {'x': -35000, 'y': -75000}, 'max': {'x': 35000, 'y': 65000}},
                    'force': 35,
                    'source': 'garment',
                },
                'don_warn_ticks': 15,
            }},
        }},
        {'tick': 0, 'kind': 'set_components', 'data': {
            'entity': '$hair_tie',
            'components': {'garment': {
                'owner': '$giantess',
                'slot': 'hair_tie',
                'state': 'worn',
                'mount_part': '$body_head',
                'worn_pos': {'x': 100000, 'y': 110000, 'z': 20000},
                'body_portals': ['$portal_hair_tie_head'],
                'don_warn_ticks': 5,
            }},
        }},
    ])


def _carry_base(casa, name, ticks, notes):
    d = copy.deepcopy(casa)
    d['name'] = name
    d['notes'] = notes
    d['ticks'] = ticks
    d.pop('atlas', None)
    d.pop('random', None)

    d.entities if False else None
    tinies = [e for e in d['entities'] if e['spawn'].get('components', {}).get('member')]
    keep_tiny = copy.deepcopy(tinies[0])
    d['entities'] = [e for e in d['entities'] if not e['spawn'].get('components', {}).get('member')]

    keep_tiny['ref'] = 'passenger'
    keep_tiny['spawn']['name'] = 'Tiny 01'
    keep_tiny['spawn']['location'] = {'space': '$palm', 'pos': {'x': 0, 'y': 0, 'z': 0}}
    keep_tiny['spawn']['components']['agent'].update({'curiosity': 0.7, 'reflex': 4, 'work': 0.2, 'revisit': 200})
    keep_tiny['spawn']['components']['needs'] = [
        {'name': 'hunger', 'value': 180, 'decay': 0},
        {'name': 'thirst', 'value': 900, 'decay': 0},
        {'name': 'energy', 'value': 900, 'decay': 0},
        {'name': 'fun', 'value': 900, 'decay': 0},
    ]
    d['entities'].append(keep_tiny)

    g = _entity(d, 'giantess')
    g['spawn']['location'] = {'space': '$living', 'pos': {'x': 1800000, 'y': 1500000, 'z': 0}}
    g['spawn']['components']['needs'] = [
        {'name': 'hunger', 'value': 950, 'decay': 0},
        {'name': 'energy', 'value': 950, 'decay': 0},
        {'name': 'fun', 'value': 950, 'decay': 0},
        {'name': 'hygiene', 'value': 950, 'decay': 0},
    ]
    g['spawn']['components']['encounter'] = {'carry': 0.35}
    return d


def giantess_carries_passenger_to_desk(casa):
    d = _carry_base(
        casa,
        'giantess_carries_passenger_to_desk',
        1200,
        'QW1: un tiny viaja sobre la palma en el living; la giantess lo lleva intencionalmente hasta el escritorio y espera a que baje.',
    )
    d['expect'] = COMMON + [
        {'event': 'action_started', 'match': {'entity': '$giantess', 'action': 'carry_passenger', 'target': '$desk1'}},
        {'event': 'carry_started', 'match': {'entity': '$giantess', 'surface': '$desk1'}},
        {'event': 'entity_reparented', 'match': {'entity': '$passenger', 'from': '$palm', 'to': '$desk1'}},
        {'event': 'passenger_delivered', 'match': {'entity': '$giantess', 'surface': '$desk1'}},
        {'no_event': 'crushed'},
    ]
    return d


def giantess_cancels_carry_when_occupied(casa):
    d = _carry_base(
        casa,
        'giantess_cancels_carry_when_occupied',
        450,
        'QW1: al llegar al escritorio para bajar al pasajero, un tiny sin reflejos está bajo la huella; cancela el apoyo sin aplastarlo.',
    )
    _entity(d, 'giantess')['spawn']['components']['encounter']['cooldowns'] = [
        {'surface': '$desk2', 'action': 'carry_passenger', 'until': 10000},
    ]
    blocker = {
        'ref': 'blocker',
        'visual': 'tiny_ochre',
        'spawn': {
            'name': 'Tiny 02',
            'location': {'space': '$desk1', 'pos': {'x': 300000, 'y': 110000, 'z': 0}},
            'body': {'radius': 2000, 'height': 8000, 'mass': 1},
            'components': {
                'mover': {'speed': 20000, 'scale': 'tiny'},
                'fragile': {},
                'agent': {'scale': 'tiny', 'reflex': 0, 'curiosity': 0, 'work': 0},
                'needs': [
                    {'name': 'hunger', 'value': 950, 'decay': 0},
                    {'name': 'energy', 'value': 950, 'decay': 0},
                ],
            },
        },
    }
    d['entities'].append(blocker)
    d['expect'] = COMMON + [
        {'event': 'action_started', 'match': {'entity': '$giantess', 'action': 'carry_passenger', 'target': '$desk1'}},
        {'event': 'encounter_skipped', 'match': {'entity': '$giantess', 'action': 'carry_passenger', 'reason': 'occupied'}},
        {'no_event': 'passenger_delivered'},
        {'no_event': 'crushed'},
    ]
    return d


def genkan_expedition_and_loot(casa):
    d = copy.deepcopy(casa)
    d['name'] = 'genkan_expedition_and_loot'
    d['notes'] = 'QW2: un tiny explora el genkan bajando por genkan_climb, entra al bolso de la entrada y encuentra una moneda.'
    d['ticks'] = 1500
    d.pop('atlas', None)
    d.pop('random', None)

    d.entities if False else None
    tinies = [e for e in d['entities'] if e['spawn'].get('components', {}).get('member')]
    scout = copy.deepcopy(tinies[0])
    d['entities'] = [e for e in d['entities'] if not e['spawn'].get('components', {}).get('member')]

    # Nido con stock cumplido para evitar acarreos que distraigan la búsqueda del bolso.
    nest = _entity(d, 'nest')
    nest['spawn']['components']['portions'] = {'items': [{'resource': 'food', 'amount': 6}, {'resource': 'water', 'amount': 6}]}

    # Solo moneda en los puntos de búsqueda del genkan para test determinista de TrinketKind=="coin".
    _entity(d, 'street_shoe_spot')['spawn']['components']['search_spot']['loot'] = [{'kind': 'coin', 'weight': 1}]
    _entity(d, 'genkan_bag_spot')['spawn']['components']['search_spot']['loot'] = [{'kind': 'coin', 'weight': 1}]

    scout['ref'] = 'scout_genkan'
    scout['spawn']['name'] = 'Tiny 01'
    scout['spawn']['location'] = {'space': '$hallway', 'pos': {'x': 450000, 'y': 100000, 'z': 0}}
    scout['spawn']['components']['mover'] = {'speed': 80000, 'scale': 'tiny'}
    scout['spawn']['components']['agent'].update({'curiosity': 0.9, 'work': 0.05, 'sight': 1800000})
    scout['spawn']['components']['member'] = {'colony': '$nest', 'role': '', 'search': 0.7, 'adventure': 0.5}
    scout['spawn']['components']['needs'] = [
        {'name': 'hunger', 'value': 950, 'decay': 0},
        {'name': 'thirst', 'value': 950, 'decay': 0},
        {'name': 'energy', 'value': 950, 'decay': 0},
        {'name': 'fun', 'value': 950, 'decay': 0},
    ]
    d['entities'].append(scout)

    g = _entity(d, 'giantess')
    g['spawn']['components']['needs'] = [
        {'name': 'hunger', 'value': 950, 'decay': 0},
        {'name': 'energy', 'value': 950, 'decay': 0},
        {'name': 'fun', 'value': 950, 'decay': 0},
        {'name': 'hygiene', 'value': 950, 'decay': 0},
    ]
    d['expect'] = COMMON + [
        {'event': 'entity_reparented', 'match': {'entity': '$scout_genkan', 'to': '$genkan'}},
        {'event': 'entity_reparented', 'match': {'entity': '$scout_genkan', 'to': '$genkan_bag'}},
        {'event': 'trinket_found', 'match': {'entity': '$scout_genkan', 'kind': 'coin'}},
        {'no_event': 'crushed'},
    ]
    return d


def sock_donning_and_hair_tie_shelter(casa):
    d = copy.deepcopy(casa)
    d['name'] = 'sock_donning_and_hair_tie_shelter'
    d['notes'] = 'QW2: un tiny dentro de la media huye a la puntera segura al calzarse, y otro dentro del coletero queda protegido de una caída.'
    d['ticks'] = 300
    d.pop('atlas', None)
    d.pop('random', None)

    tinies = [e for e in d['entities'] if e['spawn'].get('components', {}).get('member')]
    t1 = copy.deepcopy(tinies[0])
    t2 = copy.deepcopy(tinies[1])
    d['entities'] = [e for e in d['entities'] if not e['spawn'].get('components', {}).get('member')]

    # En este test la media izquierda arranca apoyada en el dormitorio para calzarla en tick 10.
    _entity(d, 'sock_left')['spawn']['location'] = {'space': '$bedroom', 'pos': {'x': 1550000, 'y': 1750000, 'z': 0}}
    _entity(d, 'sock_left')['spawn']['components']['garment']['state'] = 'placed'
    for pref in ('portal_sock_l_ankle', 'portal_sock_l_foot'):
        _entity(d, pref)['spawn']['components']['portal']['disabled'] = True
    for c in d['commands']:
        if c['kind'] == 'set_components' and c['data'].get('entity') == '$sock_left':
            c['data']['components']['garment']['state'] = 'placed'

    t1['ref'] = 'tiny_sock'
    t1['spawn']['name'] = 'Tiny 01'
    t1['spawn']['location'] = {'space': '$sock_left', 'pos': {'x': 0, 'y': 50000, 'z': 0}}
    t1['spawn']['components']['mover'] = {'speed': 60000, 'scale': 'tiny'}
    t1['spawn']['components']['agent'].update({'reflex': 2, 'curiosity': 0, 'work': 0})
    t1['spawn']['components']['needs'] = [
        {'name': 'hunger', 'value': 950, 'decay': 0},
        {'name': 'energy', 'value': 950, 'decay': 0},
    ]

    t2['ref'] = 'tiny_hair'
    t2['spawn']['name'] = 'Tiny 02'
    t2['spawn']['location'] = {'space': '$hair_tie', 'pos': {'x': 0, 'y': 0, 'z': 0}}
    t2['spawn']['components']['mover'] = {'speed': 40000, 'scale': 'tiny'}
    t2['spawn']['components']['agent'].update({'reflex': 0, 'curiosity': 0, 'work': 0})
    t2['spawn']['components']['needs'] = [
        {'name': 'hunger', 'value': 950, 'decay': 0},
        {'name': 'energy', 'value': 950, 'decay': 0},
    ]

    d['entities'].extend([t1, t2])
    g = _entity(d, 'giantess')
    g['spawn']['components']['needs'] = [
        {'name': 'hunger', 'value': 950, 'decay': 0},
        {'name': 'energy', 'value': 950, 'decay': 0},
        {'name': 'fun', 'value': 950, 'decay': 0},
        {'name': 'hygiene', 'value': 950, 'decay': 0},
    ]
    g['spawn']['components'].pop('encounter', None)

    # Soltar taza sobre el coletero (en $body_head) y calzar media izquierda en tick 10.
    cup = json.load(open(os.path.join(ROOT, 'content', 'objects', 'cup.json'), encoding='utf-8'))
    cup.pop('visual', None)
    cup['location'] = {'space': '$hair_tie', 'pos': {'x': 0, 'y': 0, 'z': 150000}}
    d['commands'].extend([
        {'tick': 10, 'kind': 'don_garment', 'data': {'actor': '$giantess', 'garment': '$sock_left', 'warn_ticks': 20}},
        {'tick': 50, 'kind': 'spawn', 'data': cup},
        {'tick': 80, 'kind': 'move_to', 'data': {'entity': '$giantess', 'space': '$living', 'goal': {'x': 1200000, 'y': 1500000}}},
    ])
    d['expect'] = COMMON + [
        {'event': 'garment_donning', 'match': {'garment': '$sock_left'}},
        {'event': 'garment_worn', 'match': {'garment': '$sock_left'}},
        {'event': 'sheltered', 'match': {'entity': '$tiny_hair', 'shelter': '$hair_tie'}},
        {'no_event': 'crushed'},
    ]
    return d


def _write(path, d):
    with open(path, 'w', encoding='utf-8', newline='\n') as f:
        f.write(json.dumps(d, ensure_ascii=False, indent=2) + '\n')
    print('escrito', os.path.relpath(path, ROOT))


def main():
    with open(CASA, encoding='utf-8') as f:
        casa = json.load(f)
    with_quickwins(casa)
    _write(CASA, casa)
    for fn in (
        giantess_carries_passenger_to_desk,
        giantess_cancels_carry_when_occupied,
        genkan_expedition_and_loot,
        sock_donning_and_hair_tie_shelter,
    ):
        d = fn(casa)
        _write(os.path.join(SCEN, d['name'] + '.json'), d)


if __name__ == '__main__':
    main()
