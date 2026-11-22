"""Genera los escenarios de Asientos en la casa (M34–M36) en tests/scenarios/.

    python tools/scenarios/asientos.py

Genera:
- tiny_reaches_chair_and_sofa.json (M34): acceso a silla y sofá por pata desde
  el piso, puente desde la mesa/escritorio y caída de migaja al asiento.
- giantess_sits_and_relaxes.json (M35): la giantess se sienta en el sofá con su
  oferta propia `relax` (sit_hover -> sit -> idle).
- chair_sit_crushes_tiny.json (M36): al sentarse en la silla sin cuidado, muere
  el tiny del centro (source: seat) y el del borde huye y se salva.
- sofa_sit_spares_far_side.json (M36): al sentarse en el sofá, muere el tiny del
  centro y sobrevive el del extremo opuesto.
- careful_sit_needs_high_care.json (M36): con cuidado alto la giantess titubea
  antes de sentarse (hesitated con source: seat) y da tiempo a huir.
"""
import copy
import json
import os

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
DIR = os.path.join(ROOT, 'tests', 'scenarios')

SEAT_CRUSHER = {'source': 'seat', 'force': 450, 'thickness': 20000}

# Poses combinadas simples: mueven tanto la palma como el asiento ($seat).
SEAT_POSES = {
    'idle': [
        {'part': '$palm', 'pos': {'x': 300000, 'y': 0, 'z': 1400000}, 'facing': 0},
        {'part': '$seat', 'pos': {'x': 0, 'y': 0, 'z': 950000}, 'facing': 0},
    ],
    'hand_hover': [
        {'part': '$palm', 'pos': {'x': 400000, 'y': 0, 'z': 880000}, 'facing': 0},
        {'part': '$seat', 'pos': {'x': 0, 'y': 0, 'z': 950000}, 'facing': 0},
    ],
    'hand_rest': [
        {'part': '$palm', 'pos': {'x': 400000, 'y': 0, 'z': 770000}, 'facing': 0},
        {'part': '$seat', 'pos': {'x': 0, 'y': 0, 'z': 950000}, 'facing': 0},
    ],
    'set_table': [
        {'part': '$palm', 'pos': {'x': 400000, 'y': 0, 'z': 900000}, 'facing': 0},
        {'part': '$seat', 'pos': {'x': 0, 'y': 0, 'z': 950000}, 'facing': 0},
    ],
    'sit_hover': [
        {'part': '$palm', 'pos': {'x': 300000, 'y': 0, 'z': 1400000}, 'facing': 0},
        {'part': '$seat', 'pos': {'x': 0, 'y': 0, 'z': 600000}, 'facing': 0},
    ],
    'sit': [
        {'part': '$palm', 'pos': {'x': 300000, 'y': 0, 'z': 1400000}, 'facing': 0},
        {'part': '$seat', 'pos': {'x': 0, 'y': 0, 'z': 470000}, 'facing': 0},
    ],
    'sit_set_table': [
        {'part': '$palm', 'pos': {'x': 400000, 'y': 0, 'z': 900000}, 'facing': 0},
        {'part': '$seat', 'pos': {'x': 0, 'y': 0, 'z': 470000}, 'facing': 0},
    ],
    'sit_hand_hover': [
        {'part': '$palm', 'pos': {'x': 400000, 'y': 0, 'z': 880000}, 'facing': 0},
        {'part': '$seat', 'pos': {'x': 0, 'y': 0, 'z': 470000}, 'facing': 0},
    ],
    'sit_hand_rest': [
        {'part': '$palm', 'pos': {'x': 400000, 'y': 0, 'z': 770000}, 'facing': 0},
        {'part': '$seat', 'pos': {'x': 0, 'y': 0, 'z': 470000}, 'facing': 0},
    ],
    'sit_hand_sofa': [
        {'part': '$palm', 'pos': {'x': 0, 'y': -280000, 'z': 470000}, 'facing': 0},
        {'part': '$seat', 'pos': {'x': 0, 'y': 0, 'z': 470000}, 'facing': 0},
    ],
}


def seat_entity():
    return {
        'ref': 'seat',
        'visual': 'seat',
        'spawn': {
            'name': 'asiento',
            'space': {'kind': 'articulated', 'extent': {'min': {'x': -180000, 'y': -180000}, 'max': {'x': 180000, 'y': 180000}}},
            'location': {'space': '$giantess', 'pos': {'x': 0, 'y': 0, 'z': 950000}},
            'components': {'crusher': dict(SEAT_CRUSHER)},
        },
    }


def chair_entities():
    """Silla común fija en el dormitorio frente a desk1 + pata + puente desde desk1."""
    return [
        {
            'ref': 'chair',
            'object': 'chair',
            'visual': 'chair',
            'spawn': {
                'name': 'silla',
                'location': {'space': '$bedroom', 'pos': {'x': 1600000, 'y': 550000, 'z': 450000}},
            },
        },
        {
            'ref': 'chair_leg',
            'spawn': {
                'name': 'pata de la silla',
                'space': {'kind': 'elevated', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 20000, 'y': 450000}}},
                'location': {'space': '$bedroom', 'pos': {'x': 1600000, 'y': 550000, 'z': 0}},
                'components': {'nav_grid': {'origin': {'x': 0, 'y': 0}, 'cell': 20000, 'version': 0, 'rows': ['.'] * 23}},
            },
        },
        {
            'ref': 'chair_leg_bottom',
            'spawn': {'name': 'chair_leg_bottom', 'components': {'portal': {
                'a': {'space': '$bedroom', 'pos': {'x': 1575000, 'y': 525000}},
                'b': {'space': '$chair_leg', 'pos': {'x': 10000, 'y': 10000}},
                'cost': 0, 'version': 0,
            }}},
        },
        {
            'ref': 'chair_leg_top',
            'spawn': {'name': 'chair_leg_top', 'components': {'portal': {
                'a': {'space': '$chair_leg', 'pos': {'x': 10000, 'y': 440000}},
                'b': {'space': '$chair', 'pos': {'x': 10000, 'y': 10000}},
                'cost': 0, 'version': 0,
            }}},
        },
        {
            'ref': 'chair_bridge',
            'spawn': {
                'name': 'respaldo de la silla',
                'space': {'kind': 'elevated', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 20000, 'y': 300000}}},
                'location': {'space': '$bedroom', 'pos': {'x': 1650000, 'y': 950000, 'z': 450000}},
                'components': {'nav_grid': {'origin': {'x': 0, 'y': 0}, 'cell': 20000, 'version': 0, 'rows': ['.'] * 15}},
            },
        },
        {
            'ref': 'chair_bridge_top',
            'spawn': {'name': 'chair_bridge_top', 'components': {'portal': {
                'a': {'space': '$desk1', 'pos': {'x': 150000, 'y': 10000}},
                'b': {'space': '$chair_bridge', 'pos': {'x': 10000, 'y': 290000}},
                'cost': 0, 'version': 0,
            }}},
        },
        {
            'ref': 'chair_bridge_bottom',
            'spawn': {'name': 'chair_bridge_bottom', 'components': {'portal': {
                'a': {'space': '$chair_bridge', 'pos': {'x': 10000, 'y': 10000}},
                'b': {'space': '$chair', 'pos': {'x': 50000, 'y': 390000}},
                'cost': 0, 'version': 0,
            }}},
        },
    ]


def sofa_entities():
    """Sofá fijo en el living frente a desk2 y tv + pata + puente desde desk2."""
    return [
        {
            'ref': 'sofa',
            'object': 'sofa',
            'visual': 'sofa',
            'spawn': {
                'name': 'sofá',
                'location': {'space': '$living', 'pos': {'x': 1200000, 'y': 450000, 'z': 450000}},
            },
        },
        {
            'ref': 'sofa_leg',
            'spawn': {
                'name': 'pata del sofá',
                'space': {'kind': 'elevated', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 20000, 'y': 450000}}},
                'location': {'space': '$living', 'pos': {'x': 1200000, 'y': 450000, 'z': 0}},
                'components': {'nav_grid': {'origin': {'x': 0, 'y': 0}, 'cell': 20000, 'version': 0, 'rows': ['.'] * 23}},
            },
        },
        {
            'ref': 'sofa_leg_bottom',
            'spawn': {'name': 'sofa_leg_bottom', 'components': {'portal': {
                'a': {'space': '$living', 'pos': {'x': 1175000, 'y': 425000}},
                'b': {'space': '$sofa_leg', 'pos': {'x': 10000, 'y': 10000}},
                'cost': 0, 'version': 0,
            }}},
        },
        {
            'ref': 'sofa_leg_top',
            'spawn': {'name': 'sofa_leg_top', 'components': {'portal': {
                'a': {'space': '$sofa_leg', 'pos': {'x': 10000, 'y': 440000}},
                'b': {'space': '$sofa', 'pos': {'x': 10000, 'y': 10000}},
                'cost': 0, 'version': 0,
            }}},
        },
        {
            'ref': 'sofa_bridge',
            'spawn': {
                'name': 'borde del sofá',
                'space': {'kind': 'elevated', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 20000, 'y': 300000}}},
                'location': {'space': '$living', 'pos': {'x': 1550000, 'y': 950000, 'z': 450000}},
                'components': {'nav_grid': {'origin': {'x': 0, 'y': 0}, 'cell': 20000, 'version': 0, 'rows': ['.'] * 15}},
            },
        },
        {
            'ref': 'sofa_bridge_top',
            'spawn': {'name': 'sofa_bridge_top', 'components': {'portal': {
                'a': {'space': '$desk2', 'pos': {'x': 50000, 'y': 10000}},
                'b': {'space': '$sofa_bridge', 'pos': {'x': 10000, 'y': 290000}},
                'cost': 0, 'version': 0,
            }}},
        },
        {
            'ref': 'sofa_bridge_bottom',
            'spawn': {'name': 'sofa_bridge_bottom', 'components': {'portal': {
                'a': {'space': '$sofa_bridge', 'pos': {'x': 10000, 'y': 10000}},
                'b': {'space': '$sofa', 'pos': {'x': 350000, 'y': 490000}},
                'cost': 0, 'version': 0,
            }}},
        },
    ]


def room(ref, name, ox):
    return {
        'ref': ref,
        'spawn': {
            'name': name,
            'space': {'kind': 'static', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 3000000, 'y': 3000000}}},
            'location': {'space': '$house', 'pos': {'x': ox, 'y': 0, 'z': 0}},
            'components': {'nav_grid': {'origin': {'x': 0, 'y': 0}, 'cell': 100000, 'version': 0, 'rows': ['.' * 30] * 30}},
        },
    }


def tiny(ref, name, space_ref, x, y, reflex=4, speed=20000, fragile=True, hunger=900):
    comps = {
        'mover': {'speed': speed},
        'agent': {'scale': 'tiny', 'reflex': reflex},
        'needs': [{'name': 'hunger', 'value': hunger}, {'name': 'energy', 'value': 900}],
    }
    if fragile:
        comps['fragile'] = {}
    return {
        'ref': ref,
        'spawn': {
            'name': name,
            'location': {'space': space_ref, 'pos': {'x': x, 'y': y, 'z': 0}},
            'body': {'radius': 2000, 'height': 8000, 'mass': 1},
            'components': comps,
        },
    }


def giantess_entities(space_ref, x, y, needs_list):
    return [
        {
            'ref': 'giantess',
            'spawn': {
                'name': 'Giantess',
                'space': {'kind': 'body', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 0, 'y': 0}}},
                'location': {'space': space_ref, 'pos': {'x': x, 'y': y, 'z': 0}},
                'components': {
                    'mover': {'speed': 1000000},
                    'needs': needs_list,
                },
            },
        },
        {
            'ref': 'palm',
            'spawn': {
                'name': 'palma',
                'space': {'kind': 'articulated', 'extent': {'min': {'x': -50000, 'y': -90000}, 'max': {'x': 50000, 'y': 90000}}},
                'location': {'space': '$giantess', 'pos': {'x': 300000, 'y': 0, 'z': 1400000}},
                'components': {
                    'contact_surface': {'tolerance': 0, 'min_overlap': 0},
                    'crusher': {'source': 'hand', 'force': 10, 'thickness': 20000},
                },
            },
        },
        seat_entity(),
    ]


def giantess_setup_cmd(care=0):
    return {
        'tick': 0,
        'kind': 'set_components',
        'data': {
            'entity': '$giantess',
            'components': {
                'poses': copy.deepcopy(SEAT_POSES),
                'agent': {
                    'scale': 'giantess',
                    'distance_cost': 0.01,
                    'hand': '$palm',
                    'idle_pose': 'idle',
                    'watch': 2000000,
                    'care': care,
                },
            },
        },
    }


def base_house():
    return [
        {'ref': 'house', 'spawn': {'name': 'house', 'space': {'kind': 'static', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 0, 'y': 0}}}}},
        room('bedroom', 'dormitorio', 3000000),
        room('living', 'living', 6000000),
        {'ref': 'door', 'spawn': {'name': 'door', 'components': {'portal': {
            'a': {'space': '$bedroom', 'pos': {'x': 2950000, 'y': 1500000}},
            'b': {'space': '$living', 'pos': {'x': 50000, 'y': 1500000}},
            'cost': 0, 'version': 0,
        }}}},
        {'ref': 'desk1', 'object': 'desk', 'spawn': {'name': 'escritorio', 'location': {'space': '$bedroom', 'pos': {'x': 1500000, 'y': 1000000, 'z': 750000}}}},
        {'ref': 'desk2', 'object': 'desk', 'spawn': {'name': 'mesa del living', 'location': {'space': '$living', 'pos': {'x': 1500000, 'y': 1000000, 'z': 750000}}, 'components': {'offers': []}}},
    ] + chair_entities() + sofa_entities()


def write(name, d):
    path = os.path.join(DIR, name + '.json')
    with open(path, 'w', encoding='utf-8', newline='\n') as f:
        f.write(json.dumps(d, ensure_ascii=False, indent=2) + '\n')
    print('escrito', os.path.relpath(path, ROOT))


def main():
    # M34: tinies llegan a silla y sofá por pata desde el piso, puente desde mesa/escritorio y cae migaja al asiento.
    m34 = {
        'name': 'tiny_reaches_chair_and_sofa',
        'notes': ('Cuatro tinies alcanzan la silla y el sofá por pata desde el piso y por puente desde el escritorio '
                  'o la mesa del living; una migaja que cae sobre la silla aterriza en el asiento.'),
        'seed': 7,
        'ticks': 900,
        'entities': base_house() + [
            tiny('t_floor_chair', 'Tiny Piso Silla', '$bedroom', 1550000, 500000, fragile=False, hunger=1000),
            tiny('t_desk_chair', 'Tiny Mesa Silla', '$desk1', 150000, 50000, fragile=False, hunger=1000),
            tiny('t_floor_sofa', 'Tiny Piso Sofa', '$living', 1150000, 400000, fragile=False, hunger=1000),
            tiny('t_desk_sofa', 'Tiny Mesa Sofa', '$desk2', 50000, 50000, fragile=False, hunger=1000),
        ],
        'commands': [
            {'tick': 1, 'kind': 'spawn', 'data': {
                'name': 'migaja',
                'location': {'space': '$bedroom', 'pos': {'x': 1800000, 'y': 750000, 'z': 750000}},
                'body': {'radius': 1500, 'height': 1000, 'mass': 0},
            }},
            {'tick': 1, 'kind': 'move_to', 'data': {'entity': '$t_floor_chair', 'space': '$chair', 'goal': {'x': 200000, 'y': 200000}}},
            {'tick': 1, 'kind': 'move_to', 'data': {'entity': '$t_desk_chair', 'space': '$chair', 'goal': {'x': 250000, 'y': 250000}}},
            {'tick': 1, 'kind': 'move_to', 'data': {'entity': '$t_floor_sofa', 'space': '$sofa', 'goal': {'x': 200000, 'y': 200000}}},
            {'tick': 1, 'kind': 'move_to', 'data': {'entity': '$t_desk_sofa', 'space': '$sofa', 'goal': {'x': 400000, 'y': 300000}}},
        ],
        'expect': [
            {'event': 'entity_fell', 'match': {'from': '$bedroom', 'to': '$chair'}},
            {'event': 'entity_reparented', 'match': {'entity': '$t_floor_chair', 'to': '$chair', 'cause': 'portal'}},
            {'event': 'entity_reparented', 'match': {'entity': '$t_desk_chair', 'to': '$chair', 'cause': 'portal'}},
            {'event': 'entity_reparented', 'match': {'entity': '$t_floor_sofa', 'to': '$sofa', 'cause': 'portal'}},
            {'event': 'entity_reparented', 'match': {'entity': '$t_desk_sofa', 'to': '$sofa', 'cause': 'portal'}},
            {'at': '$t_floor_chair', 'space': '$chair', 'pos': {'x': 200000, 'y': 200000}, 'tol': 10000},
            {'at': '$t_desk_chair', 'space': '$chair', 'pos': {'x': 250000, 'y': 250000}, 'tol': 10000},
            {'at': '$t_floor_sofa', 'space': '$sofa', 'pos': {'x': 200000, 'y': 200000}, 'tol': 10000},
            {'at': '$t_desk_sofa', 'space': '$sofa', 'pos': {'x': 400000, 'y': 300000}, 'tol': 10000},
            {'no_event': 'command_rejected'},
        ],
    }
    write('tiny_reaches_chair_and_sofa', m34)

    # M35: la giantess descansa en el sofá usando su oferta propia `relax`.
    m35 = {
        'name': 'giantess_sits_and_relaxes',
        'notes': ('Con poca energía, la giantess va al sofá, baja a sentarse (sit_hover -> sit), '
                  'descansa con la oferta relax y vuelve a ponerse de pie (idle).'),
        'seed': 7,
        'ticks': 500,
        'entities': base_house() + giantess_entities('$living', 1000000, 1500000, [
            {'name': 'hunger', 'value': 950, 'decay': 0},
            {'name': 'energy', 'value': 300, 'decay': 0},
            {'name': 'fun', 'value': 400, 'decay': 0},
        ]),
        'commands': [giantess_setup_cmd(care=0)],
        'expect': [
            {'event': 'action_started', 'match': {'entity': '$giantess', 'what': 'relax', 'target': '$sofa'}},
            {'event': 'pose_changed', 'match': {'entity': '$giantess', 'pose': 'sit_hover'}},
            {'event': 'pose_changed', 'match': {'entity': '$giantess', 'pose': 'sit'}},
            {'event': 'pressed', 'match': {'by': '$seat', 'source': 'seat', 'surface': '$sofa'}},
            {'event': 'action_done', 'match': {'entity': '$giantess', 'what': 'relax', 'target': '$sofa'}},
            {'event': 'pose_changed', 'match': {'entity': '$giantess', 'pose': 'idle'}},
            {'no_event': 'command_rejected'},
        ],
    }
    write('giantess_sits_and_relaxes', m35)

    # M36a: aplastamiento al sentarse en la silla: el del centro muere, el del borde huye y se salva.
    chair_write = {
        'action': 'write',
        'scale': 'giantess',
        'effects': [{'need': 'fun', 'amount': 250}],
        'approach': {'x': 300000, 'y': -300000},
        'facing': 1024,
        'steps': [
            {'do': 'goto'},
            {'do': 'pose', 'pose': 'sit_hover'},
            {'do': 'wait', 'ticks': 16},
            {'do': 'pose', 'pose': 'sit'},
            {'do': 'use', 'ticks': 100},
        ],
    }
    entities_chair = base_house() + giantess_entities('$bedroom', 1800000, 700000, [
        {'name': 'hunger', 'value': 950, 'decay': 0},
        {'name': 'energy', 'value': 950, 'decay': 0},
        {'name': 'fun', 'value': 200, 'decay': 0},
    ]) + [
        # chair está en (1.6, 0.55) -> $seat cubre local X [2..38 cm], Y [0..33 cm].
        tiny('tiny_center', 'Tiny Centro', '$chair', 200000, 150000, reflex=4),
        tiny('tiny_edge', 'Tiny Borde', '$chair', 200000, 326000, reflex=4),
    ]
    for e in entities_chair:
        if e['ref'] == 'desk1':
            e['spawn']['components'] = {'offers': [chair_write]}
    m36_chair = {
        'name': 'chair_sit_crushes_tiny',
        'notes': ('La giantess distraída (care 0) se sienta en la silla: la sombra de sit_hover alarma a los tinies; '
                  'el del borde huye hacia el frente y se salva por poco, el del centro muere aplastado (source: seat).'),
        'seed': 7,
        'ticks': 250,
        'entities': entities_chair,
        'commands': [giantess_setup_cmd(care=0)],
        'expect': [
            {'event': 'alarmed', 'match': {'entity': '$tiny_center', 'source': 'seat', 'channel': 'shadow'}},
            {'event': 'alarmed', 'match': {'entity': '$tiny_edge', 'source': 'seat', 'channel': 'shadow'}},
            {'event': 'pressed', 'match': {'by': '$seat', 'source': 'seat', 'surface': '$chair'}},
            {'event': 'crushed', 'match': {'entity': '$tiny_center', 'by': '$seat', 'source': 'seat', 'space': '$chair'}},
            {'no_event': 'crushed', 'match': {'entity': '$tiny_edge'}},
            {'event': 'near_miss', 'match': {'entity': '$tiny_edge', 'by': '$seat', 'source': 'seat'}},
            {'no_event': 'command_rejected'},
        ],
    }
    write('chair_sit_crushes_tiny', m36_chair)

    # M36b: aplastamiento en el sofá: muere el del centro, sobrevive el del extremo lejano sin susto.
    m36_sofa = {
        'name': 'sofa_sit_spares_far_side',
        'notes': ('La giantess se sienta en el centro del sofá: el tiny debajo muere aplastado y el del extremo '
                  'opuesto sobrevive intacto.'),
        'seed': 7,
        'ticks': 250,
        'entities': base_house() + giantess_entities('$living', 1800000, 700000, [
            {'name': 'hunger', 'value': 950, 'decay': 0},
            {'name': 'energy', 'value': 250, 'decay': 0},
            {'name': 'fun', 'value': 400, 'decay': 0},
        ]) + [
            tiny('tiny_mid', 'Tiny Medio', '$sofa', 600000, 250000, reflex=4),
            tiny('tiny_far', 'Tiny Lejos', '$sofa', 150000, 250000, reflex=4),
        ],
        'commands': [giantess_setup_cmd(care=0)],
        'expect': [
            {'event': 'pressed', 'match': {'by': '$seat', 'source': 'seat', 'surface': '$sofa'}},
            {'event': 'crushed', 'match': {'entity': '$tiny_mid', 'by': '$seat', 'source': 'seat', 'space': '$sofa'}},
            {'no_event': 'crushed', 'match': {'entity': '$tiny_far'}},
            {'no_event': 'near_miss', 'match': {'entity': '$tiny_far'}},
            {'no_event': 'command_rejected'},
        ],
    }
    write('sofa_sit_spares_far_side', m36_sofa)

    # M36c: cuidado al sentarse (fuera de la línea de vista): con care alto (0,95) titubea y salva tinies cerca del borde.
    entities_care = copy.deepcopy(entities_chair)
    entities_care = [e for e in entities_care if not e['ref'].startswith('tiny_')] + [
        # Tres tinies a 2-3 cm del borde delantero: sin titubeo no llegan (12 ticks = 1,2 cm);
        # con care 0,95 la giantess titubea (hesitated, source: seat) y alcanzan a salir.
        tiny('tiny_a', 'Tiny A', '$chair', 180000, 310000, reflex=4),
        tiny('tiny_b', 'Tiny B', '$chair', 200000, 310000, reflex=4),
        tiny('tiny_c', 'Tiny C', '$chair', 220000, 310000, reflex=4),
    ]
    m36_care = {
        'name': 'careful_sit_needs_high_care',
        'notes': ('Con cuidado alto (0,95), la giantess nota a los tinies en la silla antes de sentarse, '
                  'titubea (hesitated, source: seat) y les da tiempo a salir del borde sin aplastar a nadie.'),
        'seed': 7,
        'ticks': 300,
        'entities': entities_care,
        'commands': [giantess_setup_cmd(care=0.95)],
        'expect': [
            {'event': 'hesitated', 'match': {'entity': '$giantess', 'surface': '$chair', 'source': 'seat'}},
            {'event': 'pressed', 'match': {'by': '$seat', 'source': 'seat', 'surface': '$chair'}},
            {'no_event': 'crushed'},
            {'no_event': 'command_rejected'},
        ],
    }
    write('careful_sit_needs_high_care', m36_care)


if __name__ == '__main__':
    main()
