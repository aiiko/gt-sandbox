"""Genera los escenarios de Oficios y expediciones (M50–M55, propuesta 18) y
amplía content/partidas/casa.json con el mapa de aventuras.

    python tools/scenarios/partida.py   # rehace casa.json base
    python tools/scenarios/vida.py      # suma la vida del nido
    python tools/scenarios/oficios.py   # suma oficios, vigía, pasacables y bolsillo

Es idempotente sobre content/partidas/casa.json.
"""
import copy
import json
import os

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
CASA = os.path.join(ROOT, 'content', 'partidas', 'casa.json')
SCEN = os.path.join(ROOT, 'tests', 'scenarios')

COMMON = [{'no_event': 'command_rejected'}]

ADVENTURE_REFS = [
    'cable_duct', 'duct_bedroom', 'duct_living', 'duct_shelter',
    'threshold_refuge',
    'sofa_pocket', 'sofa_pocket_entry', 'pocket_shelter',
    'lookout_leg',
    'backrest_lookout', 'duct_target', 'sofa_pocket_target',
]


def adventure_entities(d=None):
    # Derivar posiciones del pasacables, refugio del umbral y vigía a partir de
    # la geometría real de dormitorio, living y leg1_bottom (propuesta 20.4.3).
    bed_w = 3000000
    bed_ox, bed_oy = 3000000, 0
    duct_local_y = 270000
    refuge_x = 150000
    watch_x, watch_y = 1475000, 975000
    lookout_x, lookout_y = 1400000, 950000
    if d is not None:
        bed = next((e for e in d['entities'] if e['ref'] == 'bedroom'), None)
        if bed is not None:
            bed_ox = bed['spawn']['location']['pos']['x']
            bed_oy = bed['spawn']['location']['pos']['y']
            bed_w = bed['spawn']['space']['extent']['max']['x']
            if bed_oy > 0:
                # Nueva planta: pasacables cerca del sofá, lejos de la puerta del living.
                duct_local_y = 2320000
                refuge_x = 250000
        leg_b = next((e for e in d['entities'] if e['ref'] == 'leg1_bottom'), None)
        if leg_b is not None:
            p = leg_b['spawn']['components']['portal']['a']['pos']
            watch_x, watch_y = p['x'], p['y']
            lookout_x, lookout_y = max(100000, watch_x - 80000), max(100000, watch_y - 30000)
    shared_x = bed_ox + bed_w
    duct_world_y = bed_oy + duct_local_y

    return [
        # Pasacables del zócalo entre dormitorio y living (solo tinies).
        {
            'ref': 'cable_duct',
            'visual': 'cable_duct',
            'spawn': {
                'name': 'pasacables del zócalo',
                'space': {'kind': 'static', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 200000, 'y': 40000}}},
                'location': {'space': '$house', 'pos': {'x': shared_x - 100000, 'y': duct_world_y - 20000, 'z': 0}},
                'components': {'nav_grid': {
                    'origin': {'x': 0, 'y': 0}, 'cell': 20000, 'version': 0,
                    'rows': ['.' * 10] * 2,
                }},
            },
        },
        {
            'ref': 'duct_shelter',
            'spawn': {
                'name': 'cobertura del zócalo',
                'location': {'space': '$cable_duct', 'pos': {'x': 100000, 'y': 20000, 'z': 0}},
                'components': {'shelter': {'strength': 500, 'radius': 110000}},
            },
        },
        {
            'ref': 'duct_bedroom',
            'spawn': {
                'name': 'boca del pasacables (dormitorio)',
                'components': {'portal': {
                    'a': {'space': '$bedroom', 'pos': {'x': bed_w - 50000, 'y': duct_local_y}},
                    'b': {'space': '$cable_duct', 'pos': {'x': 10000, 'y': 20000}},
                    'cost': 50000, 'scale': 'tiny', 'version': 0,
                }},
            },
        },
        {
            'ref': 'duct_living',
            'spawn': {
                'name': 'boca del pasacables (living)',
                'components': {'portal': {
                    'a': {'space': '$cable_duct', 'pos': {'x': 190000, 'y': 20000}},
                    'b': {'space': '$living', 'pos': {'x': 50000, 'y': duct_local_y}},
                    'cost': 0, 'scale': 'tiny', 'version': 0,
                }},
            },
        },
        # Refugio del umbral, lado del living (fuera de la trayectoria de la puerta).
        {
            'ref': 'threshold_refuge',
            'visual': 'threshold_refuge',
            'spawn': {
                'name': 'refugio del umbral',
                'location': {'space': '$living', 'pos': {'x': refuge_x, 'y': duct_local_y, 'z': 0}},
                'components': {
                    'shelter': {'strength': 500, 'radius': 60000},
                    'landmark': {'range': 1000000},
                    'offers': [{
                        'action': 'rest', 'scale': 'tiny',
                        'effects': [{'need': 'energy', 'amount': 300}],
                        'duration': 100,
                    }],
                },
            },
        },
        # Bolsillo del sofá: cavidad protegida conectada al borde del asiento.
        {
            'ref': 'sofa_pocket',
            'visual': 'sofa_pocket',
            'spawn': {
                'name': 'bolsillo del sofá',
                'space': {'kind': 'elevated', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 120000, 'y': 80000}}},
                'location': {'space': '$sofa', 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {'nav_grid': {
                    'origin': {'x': 0, 'y': 0}, 'cell': 20000, 'version': 0,
                    'rows': ['.' * 6] * 4,
                }},
            },
        },
        {
            'ref': 'pocket_shelter',
            'spawn': {
                'name': 'abrigo del bolsillo',
                'location': {'space': '$sofa_pocket', 'pos': {'x': 60000, 'y': 40000, 'z': 0}},
                'components': {'shelter': {'strength': 500, 'radius': 80000}},
            },
        },
        {
            'ref': 'sofa_pocket_entry',
            'spawn': {
                'name': 'abertura del bolsillo',
                'components': {'portal': {
                    'a': {'space': '$sofa', 'pos': {'x': 50000, 'y': 10000}},
                    'b': {'space': '$sofa_pocket', 'pos': {'x': 60000, 'y': 70000}},
                    'cost': 0, 'scale': 'tiny', 'version': 0,
                }},
            },
        },
        # Puesto de vigía junto a la pata del escritorio (lado del dormitorio).
        {
            'ref': 'lookout_leg',
            'visual': 'lookout_post',
            'spawn': {
                'name': 'puesto de vigía',
                'location': {'space': '$bedroom', 'pos': {'x': lookout_x, 'y': lookout_y, 'z': 0}},
                'components': {
                    'lookout_post': {
                        'colony': '$nest',
                        'portal': '$leg1_bottom',
                        'watch_point': {'x': watch_x, 'y': watch_y},
                        'watch_radius': 120000,
                        'alert_radius': 500000,
                    },
                    'landmark': {'range': 1000000},
                },
            },
        },
        # Destino 1: Mirador del respaldo (salida solitaria).
        {
            'ref': 'backrest_lookout',
            'visual': 'expedition_marker',
            'spawn': {
                'name': 'mirador del respaldo',
                'location': {'space': '$chair', 'pos': {'x': 350000, 'y': 380000, 'z': 0}},
                'components': {
                    'expedition_target': {
                        'colony': '$nest',
                        'clue': '$chair_bridge_top',
                        'min_members': 1,
                        'max_members': 1,
                        'extra_clues': ['$chair_leg_bottom'],
                    },
                    'landmark': {'range': 300000},
                },
            },
        },
        # Destino 2: Pasacables del zócalo (compañía de 2 a 4).
        {
            'ref': 'duct_target',
            'visual': 'expedition_marker',
            'spawn': {
                'name': 'salida del pasacables',
                'location': {'space': '$living', 'pos': {'x': refuge_x, 'y': duct_local_y, 'z': 0}},
                'components': {
                    'expedition_target': {
                        'colony': '$nest',
                        'clue': '$duct_bedroom',
                        'min_members': 2,
                        'max_members': 4,
                        'max_duration': 4800,
                    },
                    'landmark': {'range': 300000},
                },
            },
        },
        # Destino 3: Bolsillo del sofá (compañía de 2 a 4 con reagrupación en el umbral).
        {
            'ref': 'sofa_pocket_target',
            'visual': 'expedition_marker',
            'spawn': {
                'name': 'bolsillo del sofá',
                'location': {'space': '$sofa_pocket', 'pos': {'x': 60000, 'y': 30000, 'z': 0}},
                'components': {
                    'expedition_target': {
                        'colony': '$nest',
                        'clue': '$sofa_leg_bottom',
                        'min_members': 2,
                        'max_members': 4,
                        'waypoints': ['$threshold_refuge'],
                        'max_duration': 7200,
                    },
                    'landmark': {'range': 200000},
                },
            },
        },
    ]


def with_adventures(d):
    d['entities'] = [e for e in d['entities'] if e['ref'] not in ADVENTURE_REFS]
    for e in d['entities']:
        c = e['spawn'].setdefault('components', {})
        mv = c.get('mover')
        ag = c.get('agent')
        if mv and ag and ag.get('scale'):
            mv['scale'] = ag['scale']
    # Insertar después de sofa_bridge_bottom (así bedroom, living, chair y sofa ya existen).
    i = next(i for i, e in enumerate(d['entities']) if e['ref'] == 'sofa_bridge_bottom') + 1
    d['entities'][i:i] = adventure_entities(d)
    d['notes'] = d['notes'].split(' Oficios y expediciones:')[0] + (
        ' Oficios y expediciones: oficios flexibles, puesto de vigía, mirador del respaldo, '
        'pasacables del zócalo, refugio del umbral y bolsillo del sofá (tools/scenarios/oficios.py).')


def duct_detour_scenario():
    """M53: Un tiny cruza al living por la puerta cuando es segura y vuelve por
    el pasacables cuando la puerta tiene una zona de peligro; la giantess sigue
    usando la puerta con el pasacables abierto."""
    d = {
        'name': 'duct_detour_and_giantess_door',
        'seed': 18,
        'ticks': 1800,
        'notes': (
            'M53 (propuesta 18.8.4): con la puerta sin peligro el tiny va al living por door1. '
            'Cuando una zona de peligro cubre door1, el regreso usa cable_duct, mientras la '
            'giantess (scale=giantess) sigue cruzando por door1.'
        ),
        'entities': [
            {'ref': 'house', 'spawn': {'name': 'house', 'space': {'kind': 'static', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 0, 'y': 0}}}}},
            {'ref': 'bedroom', 'spawn': {
                'name': 'dormitorio',
                'space': {'kind': 'static', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 1000000, 'y': 1000000}}},
                'location': {'space': '$house', 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {'nav_grid': {'origin': {'x': 0, 'y': 0}, 'cell': 50000, 'version': 0, 'rows': ['.' * 20] * 20}},
            }},
            {'ref': 'living', 'spawn': {
                'name': 'living',
                'space': {'kind': 'static', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 1000000, 'y': 1000000}}},
                'location': {'space': '$house', 'pos': {'x': 1000000, 'y': 0, 'z': 0}},
                'components': {'nav_grid': {'origin': {'x': 0, 'y': 0}, 'cell': 50000, 'version': 0, 'rows': ['.' * 20] * 20}},
            }},
            {'ref': 'cable_duct', 'spawn': {
                'name': 'pasacables',
                'space': {'kind': 'static', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 100000, 'y': 40000}}},
                'location': {'space': '$house', 'pos': {'x': 950000, 'y': 150000, 'z': 0}},
                'components': {'nav_grid': {'origin': {'x': 0, 'y': 0}, 'cell': 20000, 'version': 0, 'rows': ['.' * 5] * 2}},
            }},
            {'ref': 'door1', 'spawn': {'name': 'door1', 'components': {'portal': {
                'a': {'space': '$bedroom', 'pos': {'x': 950000, 'y': 500000}},
                'b': {'space': '$living', 'pos': {'x': 50000, 'y': 500000}},
                'cost': 0, 'version': 0,
            }}}},
            {'ref': 'duct_a', 'spawn': {'name': 'duct_a', 'components': {'portal': {
                'a': {'space': '$bedroom', 'pos': {'x': 950000, 'y': 170000}},
                'b': {'space': '$cable_duct', 'pos': {'x': 10000, 'y': 20000}},
                'cost': 50000, 'scale': 'tiny', 'version': 0,
            }}}},
            {'ref': 'duct_b', 'spawn': {'name': 'duct_b', 'components': {'portal': {
                'a': {'space': '$cable_duct', 'pos': {'x': 90000, 'y': 20000}},
                'b': {'space': '$living', 'pos': {'x': 50000, 'y': 170000}},
                'cost': 0, 'scale': 'tiny', 'version': 0,
            }}}},
            {'ref': 't', 'spawn': {
                'name': 'explorador',
                'location': {'space': '$bedroom', 'pos': {'x': 750000, 'y': 500000, 'z': 0}},
                'body': {'radius': 2000, 'height': 8000, 'mass': 0},
                'components': {
                    'mover': {'speed': 40000, 'scale': 'tiny'},
                    'agent': {'scale': 'tiny'},
                },
            }},
            {'ref': 'g', 'spawn': {
                'name': 'giantess',
                'location': {'space': '$bedroom', 'pos': {'x': 500000, 'y': 170000, 'z': 0}},
                'body': {'radius': 100000, 'height': 1700000, 'mass': 60},
                'components': {
                    'mover': {'speed': 500000, 'scale': 'giantess'},
                    'agent': {'scale': 'giantess'},
                },
            }},
        ],
        'commands': [
            # Ida del tiny por la puerta (más corta y segura).
            {'tick': 1, 'kind': 'move_to', 'data': {'entity': '$t', 'space': '$living', 'goal': {'x': 200000, 'y': 500000}}},
            # Giantess cruza al living cerca del pasacables: debe usar door1, no cable_duct.
            {'tick': 1, 'kind': 'move_to', 'data': {'entity': '$g', 'space': '$living', 'goal': {'x': 200000, 'y': 170000}}},
            # Aparece peligro en door1 y el tiny vuelve al dormitorio: debe usar cable_duct.
            {'tick': 300, 'kind': 'spawn', 'data': {'name': 'door_danger', 'components': {'danger': {
                'scale': 'tiny', 'space': '$living', 'center': {'x': 50000, 'y': 500000},
                'radius': 100000, 'until': 1800,
            }}}},
            {'tick': 310, 'kind': 'move_to', 'data': {'entity': '$t', 'space': '$bedroom', 'goal': {'x': 750000, 'y': 500000}}},
        ],
        'expect': [
            # Ida por la puerta directa (bedroom -> living).
            {'event': 'entity_reparented', 'match': {'entity': '$t', 'from': '$bedroom', 'to': '$living'}, 'by': 280},
            # Vuelta por el pasacables (living -> cable_duct -> bedroom).
            {'event': 'entity_reparented', 'match': {'entity': '$t', 'from': '$living', 'to': '$cable_duct'}, 'after': 300},
            {'event': 'entity_reparented', 'match': {'entity': '$t', 'from': '$cable_duct', 'to': '$bedroom'}, 'after': 300},
            # La giantess nunca entra al pasacables y llega al living por la puerta.
            {'event': 'entity_reparented', 'match': {'entity': '$g', 'from': '$bedroom', 'to': '$living'}},
            {'no_event': 'entity_reparented', 'match': {'entity': '$g', 'to': '$cable_duct'}},
        ] + COMMON,
    }
    return d


def main():
    d = json.load(open(CASA, encoding='utf-8'))
    with_adventures(d)
    with open(CASA, 'w', encoding='utf-8', newline='\n') as out:
        out.write(json.dumps(d, ensure_ascii=False, indent=2) + '\n')
    print('actualizado', os.path.relpath(CASA, ROOT))

    for f in [duct_detour_scenario]:
        s = f()
        p = os.path.join(SCEN, s['name'] + '.json')
        with open(p, 'w', encoding='utf-8', newline='\n') as out:
            json.dump(s, out, ensure_ascii=False, indent=2)
            out.write('\n')
        print('escrito', os.path.relpath(p, ROOT))


if __name__ == '__main__':
    main()
