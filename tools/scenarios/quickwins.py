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


def _entity(d, ref):
    return next(e for e in d['entities'] if e.get('ref') == ref)


def with_quickwins(d):
    """Integra idempotentemente los quick wins en content/partidas/casa.json."""
    g = _entity(d, 'giantess')
    enc = g['spawn']['components'].setdefault('encounter', {})
    enc['carry'] = 0.18


def _carry_base(casa, name, ticks, notes):
    d = copy.deepcopy(casa)
    d['name'] = name
    d['notes'] = notes
    d['ticks'] = ticks
    d.pop('atlas', None)
    d.pop('random', None)

    # Dejar solo un tiny molde como pasajero en la palma y satisfacer necesidades de la giantess.
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
    ):
        d = fn(casa)
        _write(os.path.join(SCEN, d['name'] + '.json'), d)


if __name__ == '__main__':
    main()
