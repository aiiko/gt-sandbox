"""Genera los escenarios de Vida cotidiana de los tinies (M39 a M44).

    python tools/scenarios/vida.py

Parte de la base de aplastamiento (aplastamiento.py: dormitorio, escritorio
de 60 x 40 cm a 75 cm, giantess y palma) y le suma una pata con portales
entre el piso y el escritorio. Los escenarios del slice se escriben en
tests/scenarios/.
"""
import json
import os

from aplastamiento import base, pose

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
SCEN = os.path.join(ROOT, 'tests', 'scenarios')

COMMON = [{'no_event': 'command_rejected'}]

# El escritorio desk1 está en roomA en (150, 100, 75) cm y mide 60 x 40 cm.
LEG = [
    {'ref': 'leg1', 'spawn': {
        'name': 'leg1',
        'space': {'kind': 'elevated', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 20000, 'y': 740000}}},
        'location': {'space': '$roomA', 'pos': {'x': 1500000, 'y': 1000000, 'z': 0}},
        'components': {'nav_grid': {'origin': {'x': 0, 'y': 0}, 'cell': 20000, 'version': 0,
                                    'rows': ['.'] * 37}}}},
    {'ref': 'leg1_bottom', 'spawn': {'name': 'leg1_bottom', 'components': {'portal': {
        'a': {'space': '$roomA', 'pos': {'x': 1475000, 'y': 975000}},
        'b': {'space': '$leg1', 'pos': {'x': 10000, 'y': 10000}}, 'cost': 0, 'version': 0}}}},
    {'ref': 'leg1_top', 'spawn': {'name': 'leg1_top', 'components': {'portal': {
        'a': {'space': '$leg1', 'pos': {'x': 10000, 'y': 730000}},
        'b': {'space': '$desk1', 'pos': {'x': 10000, 'y': 10000}}, 'cost': 0, 'version': 0}}}},
]


def world(name, ticks, notes):
    d = base(name, ticks, notes)
    d['entities'] += json.loads(json.dumps(LEG))
    return d


def nest(x, y, portions=0, **colony):
    comps = {'portions': {'left': portions}}
    if colony:
        comps['colony'] = dict({'radius': 30000}, **colony)
    return {'ref': 'nest', 'object': 'nest', 'spawn': {
        'name': 'nido', 'location': {'space': '$desk1', 'pos': {'x': x, 'y': y, 'z': 0}},
        'components': comps}}


def crumb(ref, space, x, y):
    return {'ref': ref, 'object': 'crumb', 'spawn': {
        'name': ref, 'location': {'space': '$' + space, 'pos': {'x': x, 'y': y, 'z': 0}}}}


def tiny(ref, space, x, y, hunger=900, energy=900, decay=(0, 0), work=0.2, sight=200000,
         agent=None, member=None, extra=None):
    a = {'scale': 'tiny', 'work': work, 'distance_cost': 0.2, 'sight': sight}
    a.update(agent or {})
    m = {'colony': '$nest'}
    m.update(member or {})
    comps = {
        'mover': {'speed': 20000},
        'agent': a,
        'member': m,
        'needs': [{'name': 'hunger', 'value': hunger, 'decay': decay[0]},
                  {'name': 'energy', 'value': energy, 'decay': decay[1]}],
    }
    comps.update(extra or {})
    return {'ref': ref, 'spawn': {
        'name': ref,
        'location': {'space': '$' + space, 'pos': {'x': x, 'y': y, 'z': 0}},
        'body': {'radius': 2000, 'height': 8000, 'mass': 0},
        'components': comps,
    }}


def post_haul(tick, target, dest_x, dest_y):
    return {'tick': tick, 'kind': 'post_job', 'data': {
        'type': 'haul', 'target': '$' + target, 'dest_space': '$desk1', 'dest': {'x': dest_x, 'y': dest_y}}}


# --- M39: despensa y rutina del nido ---

def tiny_returns_to_eat():
    d = world('tiny_returns_to_eat', 2400,
              'M39: un tiny acarrea una migaja lejana al nido. El hambre baja rápido y, antes de terminar, '
              'le gana al acarreo: suelta la migaja, vuelve al nido, come y gasta una porción del depósito. '
              'Después retoma el trabajo.')
    d['entities'] += [
        nest(60000, 60000, portions=3),
        crumb('far', 'desk1', 560000, 360000),
        tiny('t', 'desk1', 80000, 60000, hunger=700, decay=(12, 0), sight=50000),
    ]
    d['commands'] += [post_haul(1, 'far', 60000, 60000)]
    d['expect'] = [
        {'event': 'job_assigned', 'match': {'by': '$t'}, 'by': 20},
        {'event': 'action_interrupted', 'match': {'entity': '$t', 'action': 'work', 'by': 'eat'}},
        {'event': 'job_unassigned', 'match': {'by': '$t', 'reason': 'stopped'}},
        {'event': 'action_done', 'match': {'entity': '$t', 'what': 'eat', 'target': '$nest'}},
        {'event': 'consumed', 'match': {'entity': '$t', 'target': '$nest', 'left': 2}},
        {'event': 'stored', 'match': {'colony': '$nest', 'item': '$far'}},
    ] + COMMON
    return d


def tiny_rests_at_nest():
    d = world('tiny_rests_at_nest', 3000,
              'M39: un tiny cansado que acarrea una migaja deja el trabajo para descansar en el nido y su '
              'energía sube (lo verifica TestRestRaisesEnergy).')
    d['entities'] += [
        nest(60000, 60000),
        crumb('far', 'desk1', 560000, 360000),
        tiny('t', 'desk1', 80000, 60000, energy=700, decay=(0, 6), sight=50000),
    ]
    d['commands'] += [post_haul(1, 'far', 60000, 60000)]
    d['expect'] = [
        {'event': 'action_interrupted', 'match': {'entity': '$t', 'action': 'work', 'by': 'rest'}},
        {'event': 'action_done', 'match': {'entity': '$t', 'what': 'rest', 'target': '$nest'}},
    ] + COMMON
    return d


def nest_unreachable_eats_crumb():
    d = world('nest_unreachable_eats_crumb', 900,
              'M39: un tiny con hambre está en el piso y el nido está en el escritorio, pero la pata está '
              'cerrada. Intenta volver, falla y no insiste: come la migaja que conoce en el piso y no '
              'vuelve a intentar el nido antes de 30 s.')
    d['entities'] += [
        nest(60000, 60000, portions=5),
        crumb('near', 'roomA', 1300000, 900000),
        tiny('t', 'roomA', 1350000, 900000, hunger=250, sight=100000),
    ]
    d['commands'] += [{'tick': 0, 'kind': 'set_portal', 'data': {'portal': '$leg1_top', 'enabled': False}}]
    d['expect'] = [
        {'event': 'action_failed', 'match': {'entity': '$t', 'what': 'eat', 'target': '$nest'}},
        {'event': 'action_done', 'match': {'entity': '$t', 'what': 'nibble', 'target': '$near'}},
        {'event': 'action_started', 'match': {'entity': '$t', 'what': 'eat'}, 'by': 600, 'count': 1},
    ] + COMMON
    return d


def main():
    for f in [tiny_returns_to_eat, tiny_rests_at_nest, nest_unreachable_eats_crumb]:
        d = f()
        with open(os.path.join(SCEN, d['name'] + '.json'), 'w', encoding='utf-8', newline='\n') as out:
            json.dump(d, out, ensure_ascii=False, indent=2)
            out.write('\n')


if __name__ == '__main__':
    main()
