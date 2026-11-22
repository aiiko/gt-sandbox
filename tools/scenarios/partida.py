"""Genera la base de la partida libre y el escenario de la merienda (M28).

    python tools/scenarios/descuido.py   # antes, si cambió la mañana peligrosa
    python tools/scenarios/partida.py

Parte de «Un descuido en el escritorio» (un_descuido_en_el_escritorio.json):

- content/partidas/casa.json es la base de «Nueva partida». Suma perfiles de
  arranque de la giantess (bloque "random": con hambre, cansada o con ganas
  de entretenerse), con hambre y diversión que decaen, y cambia la merienda
  de la heladera por una en la mesa del living que deja migajas.
  scenario.Randomize sortea la colonia, el cuidado y el perfil.
- tests/scenarios/giantess_snacks_in_living.json verifica la merienda: con
  algo de hambre, la giantess lleva un plato a la mesa del living, deja
  migajas y come ahí.
- tests/scenarios/giantess_logs_sightings.json: la giantess registra dónde ve
  tinies (componente "encounter") y olvida un avistamiento viejo.
- tests/scenarios/giantess_observes_desk.json: sin necesidades que atender,
  se detiene a mirar el escritorio y no vuelve a mirarlo enseguida.
- tests/scenarios/giantess_shares_crumbs.json: deja migajas donde los tinies
  buscan comida y no hay; giantess_skips_sharing.json: si al llegar ya hay
  migajas a la vista, no comparte.

Los escenarios de referencia (la mañana y el descuido) no cambian.
"""
import copy
import json
import os

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
SCENARIOS = os.path.join(ROOT, 'tests', 'scenarios')

# Merienda: sacar un plato de la heladera, llevarlo a la mesa del living,
# dejarlo con unas migajas alrededor y comer con la mano apoyada.
SNACK_LIVING = {
    'action': 'snack_living', 'scale': 'giantess', 'effects': [{'need': 'hunger', 'amount': 350}],
    'approach': {'x': 0, 'y': -500000},
    'steps': [
        {'do': 'goto'},
        {'do': 'wait', 'ticks': 30},
        {'do': 'take', 'item': 'plate'},
        {'do': 'goto', 'place': 'living_table'},
        {'do': 'pose', 'pose': 'set_table'},
        {'do': 'drop', 'offset': {'x': 40000, 'y': 120000}},
        {'do': 'spawn', 'item': 'crumb', 'count': 3, 'offset': {'x': 100000, 'y': 0}, 'spread': 120000},
        {'do': 'wait', 'ticks': 12},
        {'do': 'pose', 'pose': 'hand_rest'},
        {'do': 'use', 'ticks': 250},
    ],
}

# Encuentros de la giantess en la partida libre (utilidades de cada acción).
OBJECTS = os.path.join(ROOT, 'content', 'objects')


def crumb():
    """La migaja que deja al compartir (definición completa: la simulación no
    conoce la biblioteca de objetos)."""
    return json.load(open(os.path.join(OBJECTS, 'crumb.json'), encoding='utf-8'))


ENCOUNTER = {'observe': 0.12, 'share': 0.2}


def satisfied(d):
    """La giantess sin necesidades que atender (para probar un encuentro solo)."""
    entity(d, 'giantess')['spawn']['components']['needs'] = [
        {'name': 'hunger', 'value': 950, 'decay': 0},
        {'name': 'energy', 'value': 950, 'decay': 0},
        {'name': 'fun', 'value': 950, 'decay': 0}]

# Perfiles de arranque. Hambre y diversión bajan 2 por segundo (de llenas a
# vacías en unos 8 minutos) y la energía 1, así la rutina vuelve a decidir.
STARTS = [
    {'name': 'con hambre', 'needs': [
        {'name': 'hunger', 'value': 180, 'decay': 2},
        {'name': 'energy', 'value': 900, 'decay': 1},
        {'name': 'fun', 'value': 600, 'decay': 2}]},
    {'name': 'cansada', 'needs': [
        {'name': 'hunger', 'value': 600, 'decay': 2},
        {'name': 'energy', 'value': 180, 'decay': 1},
        {'name': 'fun', 'value': 650, 'decay': 2}]},
    {'name': 'con ganas de entretenerse', 'needs': [
        {'name': 'hunger', 'value': 650, 'decay': 2},
        {'name': 'energy', 'value': 850, 'decay': 1},
        {'name': 'fun', 'value': 150, 'decay': 2}]},
]


def entity(d, ref):
    return next(e for e in d['entities'] if e['ref'] == ref)


def with_living_snack(d):
    fridge = entity(d, 'fridge')
    offers = fridge['spawn']['components']['offers']
    offers[:] = [o for o in offers if o['action'] != 'snack'] + [copy.deepcopy(SNACK_LIVING)]


def write(path, d):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w', encoding='utf-8', newline='\n') as f:
        f.write(json.dumps(d, ensure_ascii=False, indent=2) + '\n')
    print('escrito', os.path.relpath(path, ROOT))


def main():
    base = json.load(open(os.path.join(SCENARIOS, 'un_descuido_en_el_escritorio.json'), encoding='utf-8'))

    d = copy.deepcopy(base)
    d['name'] = 'Partida libre'
    d['notes'] = ('Base de «Nueva partida»: la mañana peligrosa con perfiles de arranque de la giantess y la '
                  'merienda en el living. scenario.Randomize sortea la colonia, el cuidado y el perfil.')
    d['atlas'] = '../art/casa.json'
    d['expect'] = []
    with_living_snack(d)
    d['random'] = {'agent': '$giantess', 'starts': copy.deepcopy(STARTS)}
    entity(d, 'giantess')['spawn']['components']['encounter'] = dict(copy.deepcopy(ENCOUNTER), portion=crumb())
    write(os.path.join(ROOT, 'content', 'partidas', 'casa.json'), d)

    d = copy.deepcopy(base)
    d['name'] = 'giantess_snacks_in_living'
    d['notes'] = ('Con algo de hambre (no tanta como para desayunar), la giantess saca un plato de la '
                  'heladera, lo lleva a la mesa del living, deja migajas alrededor y come ahí.')
    d['ticks'] = 900
    del d['atlas']  # prueba del motor: no se ofrece en el menú
    with_living_snack(d)
    entity(d, 'giantess')['spawn']['components']['needs'] = [
        {'name': 'hunger', 'value': 750, 'decay': 0},
        {'name': 'energy', 'value': 950, 'decay': 0},
        {'name': 'fun', 'value': 950, 'decay': 0}]
    d['expect'] = [
        {'event': 'action_started', 'match': {'entity': '$giantess', 'what': 'snack_living'}},
        {'event': 'item_dropped', 'match': {'entity': '$giantess', 'on': '$desk2'}},
        {'event': 'entity_fell', 'match': {'to': '$desk2'}},
        {'event': 'action_done', 'match': {'entity': '$giantess', 'what': 'snack_living'}},
        {'no_event': 'action_started', 'match': {'entity': '$giantess', 'what': 'breakfast'}},
        {'no_event': 'command_rejected'},
    ]
    write(os.path.join(SCENARIOS, 'giantess_snacks_in_living.json'), d)

    d = copy.deepcopy(base)
    d['name'] = 'giantess_logs_sightings'
    d['notes'] = ('La giantess registra en qué superficie ve tinies y cuántos buscan comida; un avistamiento '
                  'viejo (la cocina, en el tick 0) se olvida pasados dos minutos sin volver a ver a nadie ahí.')
    d['ticks'] = 2600
    del d['atlas']
    entity(d, 'giantess')['spawn']['components']['encounter'] = {
        'sightings': [{'surface': '$kitchen', 'last': 0, 'seen': 5}]}
    d['expect'] = [{'no_event': 'command_rejected'}]
    write(os.path.join(SCENARIOS, 'giantess_logs_sightings.json'), d)

    d = copy.deepcopy(base)
    d['name'] = 'giantess_observes_desk'
    d['notes'] = ('Sin necesidades que atender, la giantess va a mirar a los tinies del escritorio con la mano '
                  'en reposo (no se asustan) y, por el enfriamiento, no vuelve a mirarlo enseguida.')
    d['ticks'] = 1500
    del d['atlas']
    satisfied(d)
    entity(d, 'giantess')['spawn']['components']['encounter'] = {'observe': 0.12}
    d['expect'] = [
        {'event': 'action_started', 'match': {'entity': '$giantess', 'action': 'observe', 'target': '$desk1'}},
        {'event': 'observed', 'match': {'entity': '$giantess', 'surface': '$desk1'}},
        {'event': 'action_started', 'match': {'entity': '$giantess', 'action': 'observe'}, 'count': 1},
        {'no_event': 'alarmed'},
        {'no_event': 'crushed'},
        {'no_event': 'command_rejected'},
    ]
    write(os.path.join(SCENARIOS, 'giantess_observes_desk.json'), d)

    # Sin galletas en el escritorio, los tinies salen a buscar comida.
    hungry = copy.deepcopy(base)
    del hungry['atlas']
    hungry['entities'] = [e for e in hungry['entities'] if not e['ref'].startswith('cookie')]
    satisfied(hungry)
    entity(hungry, 'giantess')['spawn']['components']['encounter'] = {'share': 0.3, 'portion': crumb()}

    d = copy.deepcopy(hungry)
    d['name'] = 'giantess_shares_crumbs'
    d['notes'] = ('En el escritorio no hay comida y los tinies salen a buscar: la giantess se acerca, suspende la '
                  'mano y deja dos migajas, que la colonia acarrea al nido.')
    d['ticks'] = 3000
    d['expect'] = [
        {'event': 'action_started', 'match': {'entity': '$giantess', 'action': 'share', 'target': '$desk1'}},
        {'event': 'items_left', 'match': {'entity': '$giantess', 'motive': 'share'}},
        {'event': 'shared', 'match': {'entity': '$giantess', 'surface': '$desk1'}},
        {'event': 'stored'},
        {'no_event': 'crushed'},
        {'no_event': 'command_rejected'},
    ]
    write(os.path.join(SCENARIOS, 'giantess_shares_crumbs.json'), d)

    d = copy.deepcopy(hungry)
    d['name'] = 'giantess_skips_sharing'
    d['notes'] = ('Los tinies buscan comida y la giantess va a compartir, pero mientras camina caen migajas en el '
                  'escritorio: al llegar ve que hay suficientes y no deja nada (encounter_skipped).')
    d['ticks'] = 1200
    for i, (x, y) in enumerate([(150000, 200000), (200000, 120000), (520000, 100000)]):
        spawn = dict(crumb(), location={'space': '$desk1', 'pos': {'x': x, 'y': y, 'z': 0}})
        d['commands'].append({'tick': 15, 'kind': 'spawn', 'data': spawn})
    d['expect'] = [
        {'event': 'action_started', 'match': {'entity': '$giantess', 'action': 'share'}, 'count': 1},
        {'event': 'encounter_skipped', 'match': {'entity': '$giantess', 'action': 'share', 'reason': 'enough'}},
        {'no_event': 'items_left', 'match': {'motive': 'share'}},
        {'no_event': 'shared'},
        {'no_event': 'command_rejected'},
    ]
    write(os.path.join(SCENARIOS, 'giantess_skips_sharing.json'), d)


if __name__ == '__main__':
    main()
