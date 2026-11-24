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
- tests/scenarios/tiny_boards_offered_palm.json, nobody_boards_offered_palm.json
  y offer_canceled_when_occupied.json: los tres desenlaces de ofrecer la
  palma (alguien sube, nadie sube, la huella ocupada cancela el apoyo).

Los escenarios de referencia (la mañana y el descuido) no cambian.
"""
import copy
import json
import os

from asientos import SEAT_POSES, chair_entities, seat_entity, sofa_entities
from planta import build_apartment

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
SCENARIOS = os.path.join(ROOT, 'tests', 'scenarios')

# Poses de la partida libre: suma las poses para la mesa auxiliar del living
# (65 cm de alto, delante del sofá mirando al sur).
APARTMENT_POSES = copy.deepcopy(SEAT_POSES)
APARTMENT_POSES.update({
    'sit_table_set': [
        {'part': '$palm', 'pos': {'x': 550000, 'y': 0, 'z': 800000}, 'facing': 0},
        {'part': '$body_rear', 'pos': {'x': 0, 'y': 0, 'z': 470000}, 'facing': 0},
    ],
    'sit_table_hover': [
        {'part': '$palm', 'pos': {'x': 550000, 'y': 0, 'z': 780000}, 'facing': 0},
        {'part': '$body_rear', 'pos': {'x': 0, 'y': 0, 'z': 470000}, 'facing': 0},
    ],
    'sit_table_rest': [
        {'part': '$palm', 'pos': {'x': 550000, 'y': 0, 'z': 670000}, 'facing': 0},
        {'part': '$body_rear', 'pos': {'x': 0, 'y': 0, 'z': 470000}, 'facing': 0},
    ],
})

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

# Variantes con asientos para la partida libre (content/partidas/casa.json):
# desayunar/escribir sentada en la silla; merendar/mirar TV sentada en el sofá.
SNACK_LIVING_SEAT = {
    'action': 'snack_living', 'scale': 'giantess', 'effects': [{'need': 'hunger', 'amount': 350}],
    'approach': {'x': -600000, 'y': 0}, 'facing': 0,
    'steps': [
        {'do': 'goto'},
        {'do': 'wait', 'ticks': 30},
        {'do': 'take', 'item': 'plate'},
        {'do': 'goto', 'place': 'living_table'},
        {'do': 'pose', 'pose': 'sit_hover'},
        {'do': 'wait', 'ticks': 16},
        {'do': 'pose', 'pose': 'sit'},
        {'do': 'pose', 'pose': 'sit_table_set'},
        {'do': 'drop', 'offset': {'x': -100000, 'y': 40000}},
        {'do': 'spawn', 'item': 'crumb', 'count': 4, 'offset': {'x': 0, 'y': 0}, 'spread': 160000},
        {'do': 'wait', 'ticks': 12},
        {'do': 'pose', 'pose': 'sit_table_rest'},
        {'do': 'use', 'ticks': 250},
    ],
}

WRITE_SEAT = {
    'action': 'write', 'scale': 'giantess', 'effects': [{'need': 'fun', 'amount': 250}],
    'approach': {'x': 300000, 'y': -300000}, 'facing': 1024,
    'steps': [
        {'do': 'goto'},
        {'do': 'pose', 'pose': 'sit_hover'},
        {'do': 'wait', 'ticks': 16},
        {'do': 'pose', 'pose': 'sit'},
        {'do': 'pose', 'pose': 'sit_hand_hover'},
        {'do': 'wait', 'ticks': 12},
        {'do': 'pose', 'pose': 'sit_hand_rest'},
        {'do': 'use', 'ticks': 400},
    ],
}

WATCH_SEAT = {
    'action': 'watch', 'scale': 'giantess', 'effects': [{'need': 'fun', 'amount': 500}],
    'approach': {'x': 0, 'y': -1500000}, 'facing': 1024,
    'steps': [
        {'do': 'goto'},
        {'do': 'pose', 'pose': 'sit_hover'},
        {'do': 'wait', 'ticks': 16},
        {'do': 'pose', 'pose': 'sit'},
        {'do': 'pose', 'pose': 'sit_hand_hover'},
        {'do': 'wait', 'ticks': 12},
        {'do': 'pose', 'pose': 'sit_hand_rest'},
        {'do': 'use', 'ticks': 300},
    ],
}


def with_seats(d):
    """Agrega silla, sofá, sus accesos y las poses/rutinas de sentarse (M34–M38)."""
    i_palm = next(i for i, e in enumerate(d['entities']) if e['ref'] == 'palm')
    d['entities'].insert(i_palm + 1, seat_entity())
    i_leg = next(i for i, e in enumerate(d['entities']) if e['ref'] == 'leg1_top')
    d['entities'][i_leg + 1:i_leg + 1] = chair_entities()
    i_desk2 = next(i for i, e in enumerate(d['entities']) if e['ref'] == 'desk2')
    d['entities'][i_desk2 + 1:i_desk2 + 1] = sofa_entities()

    d['commands'][0]['data']['components']['poses'] = copy.deepcopy(SEAT_POSES)

    # Desayuno y merienda sentadas, con migajas que también caen a la silla y al sofá.
    fridge = entity(d, 'fridge')
    offers = fridge['spawn']['components']['offers']
    for o in offers:
        if o['action'] == 'breakfast':
            o['steps'] = [
                {'do': 'goto'},
                {'do': 'wait', 'ticks': 40},
                {'do': 'take', 'item': 'cup'},
                {'do': 'goto', 'place': 'desk'},
                {'do': 'pose', 'pose': 'sit_hover'},
                {'do': 'wait', 'ticks': 16},
                {'do': 'pose', 'pose': 'sit'},
                {'do': 'pose', 'pose': 'sit_set_table'},
                {'do': 'drop', 'offset': {'x': 50000, 'y': 150000}},
                {'do': 'spawn', 'item': 'crumb', 'count': 4, 'offset': {'x': 60000, 'y': 0}, 'spread': 220000},
                {'do': 'wait', 'ticks': 12},
                {'do': 'pose', 'pose': 'sit_hand_rest'},
                {'do': 'use', 'ticks': 300},
            ]
    offers[:] = [copy.deepcopy(SNACK_LIVING_SEAT) if o['action'] == 'snack_living' else o for o in offers]

    entity(d, 'desk1')['spawn'].setdefault('components', {})['offers'] = [copy.deepcopy(WRITE_SEAT)]
    entity(d, 'tv')['spawn'].setdefault('components', {})['offers'] = [copy.deepcopy(WATCH_SEAT)]

# Encuentros de la giantess en la partida libre (utilidades de cada acción).
OBJECTS = os.path.join(ROOT, 'content', 'objects')


def crumb():
    """La migaja que deja al compartir (definición completa: la simulación no
    conoce la biblioteca de objetos)."""
    return json.load(open(os.path.join(OBJECTS, 'crumb.json'), encoding='utf-8'))


ENCOUNTER = {'observe': 0.12, 'share': 0.2, 'offer': 0.16}

# Cuánto tarda un tiny en volver a interesarse por un Space donde ya estuvo
# (la palma, sobre todo): 90 segundos.
REVISIT = 90 * 20


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


OLD_HOUSE_REFS = {
    'house', 'kitchen', 'bedroom', 'living', 'door1', 'door2', 'door_kb', 'door_bl',
    'bed', 'desk1', 'leg1', 'leg1_bottom', 'leg1_top', 'fridge', 'desk2', 'tv',
}


def build_casa_apartment(base):
    """Arma content/partidas/casa.json sobre la nueva planta del departamento (propuesta 20)."""
    d = copy.deepcopy(base)
    d['name'] = 'Partida libre'
    d['notes'] = (
        'Base de «Nueva partida»: departamento de 7,4 × 6,4 m con distribuidor, dormitorio, living, '
        'cocina, baño y genkan, y 21 piezas de mobiliario independientes (propuesta 20). '
        'scenario.Randomize sortea la colonia, el cuidado y el perfil.'
    )
    d['atlas'] = '../art/casa.json'
    d['expect'] = []

    apt_ents, _ = build_apartment()
    kept = [e for e in d['entities'] if e['ref'] not in OLD_HOUSE_REFS]
    # Ubicar a la giantess en zona despejada del dormitorio nuevo e insertar body_rear tras palm.
    for e in kept:
        if e['ref'] == 'giantess':
            e['spawn']['location'] = {'space': '$bedroom', 'pos': {'x': 1800000, 'y': 1500000, 'z': 0}}
            gc = e['spawn'].setdefault('components', {})
            gc.setdefault('mover', {})['scale'] = 'giantess'
            gc['needs'] = copy.deepcopy(STARTS[0]['needs'])
    i_palm = next(i for i, e in enumerate(kept) if e['ref'] == 'palm')
    kept.insert(i_palm + 1, seat_entity())
    d['entities'] = apt_ents + kept

    d['commands'][0]['data']['components']['poses'] = copy.deepcopy(APARTMENT_POSES)

    # Ofertas de heladera (frente despejado al oeste) y escritorio.
    fridge = entity(d, 'fridge')
    fridge['spawn'].setdefault('components', {})['offers'] = [
        {
            'action': 'breakfast', 'scale': 'giantess',
            'effects': [{'need': 'hunger', 'amount': 600}],
            'approach': {'x': -600000, 'y': 0}, 'facing': 0,
            'steps': [
                {'do': 'goto'},
                {'do': 'wait', 'ticks': 40},
                {'do': 'take', 'item': 'cup'},
                {'do': 'goto', 'place': 'desk'},
                {'do': 'pose', 'pose': 'sit_hover'},
                {'do': 'wait', 'ticks': 16},
                {'do': 'pose', 'pose': 'sit'},
                {'do': 'pose', 'pose': 'sit_set_table'},
                {'do': 'drop', 'offset': {'x': 50000, 'y': 150000}},
                {'do': 'spawn', 'item': 'crumb', 'count': 4, 'offset': {'x': 60000, 'y': 0}, 'spread': 220000},
                {'do': 'wait', 'ticks': 12},
                {'do': 'pose', 'pose': 'sit_hand_rest'},
                {'do': 'use', 'ticks': 300},
            ],
        },
        copy.deepcopy(SNACK_LIVING_SEAT),
    ]
    entity(d, 'desk1')['spawn'].setdefault('components', {})['offers'] = [copy.deepcopy(WRITE_SEAT)]

    d['random'] = {'agent': '$giantess', 'starts': copy.deepcopy(STARTS)}
    entity(d, 'giantess')['spawn']['components']['encounter'] = dict(copy.deepcopy(ENCOUNTER), portion=crumb())
    for e in d['entities']:
        a = e['spawn'].get('components', {}).get('agent')
        if a and a.get('scale') == 'tiny':
            a['revisit'] = REVISIT
    return d


def main():
    base = json.load(open(os.path.join(SCENARIOS, 'un_descuido_en_el_escritorio.json'), encoding='utf-8'))

    d = build_casa_apartment(base)
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

    # Sin galletas: la huella de la mano (junto al borde del escritorio) queda
    # libre una vez que la sombra aparta a los que pasaban por ahí.
    offer = copy.deepcopy(base)
    del offer['atlas']
    offer['entities'] = [e for e in offer['entities'] if not e['ref'].startswith('cookie')]
    satisfied(offer)
    entity(offer, 'giantess')['spawn']['components']['encounter'] = {'offer': 0.3}
    common = [
        {'event': 'action_started', 'match': {'entity': '$giantess', 'action': 'offer_palm', 'target': '$desk1'}},
        {'no_event': 'crushed'},
        {'no_event': 'command_rejected'},
    ]

    d = copy.deepcopy(offer)
    d['name'] = 'tiny_boards_offered_palm'
    d['notes'] = ('La giantess apoya la palma abierta junto a la colonia y espera: algún tiny curioso cruza por '
                  'el portal de contacto. Después retira la mano.')
    d['ticks'] = 900
    d['expect'] = common + [
        {'event': 'palm_offered', 'match': {'entity': '$giantess', 'surface': '$desk1'}},
        {'event': 'explored', 'match': {'space': '$palm'}},
        {'event': 'palm_withdrawn', 'match': {'entity': '$giantess', 'surface': '$desk1'}},
    ]
    write(os.path.join(SCENARIOS, 'tiny_boards_offered_palm.json'), d)

    d = copy.deepcopy(offer)
    d['name'] = 'nobody_boards_offered_palm'
    d['notes'] = ('Ofrece la palma a tinies sin curiosidad: nadie sube y la mano se retira al terminar la espera.')
    d['ticks'] = 900
    for e in d['entities']:
        a = e['spawn'].get('components', {}).get('agent')
        if a and a.get('scale') == 'tiny':
            a['curiosity'] = 0
    d['expect'] = common + [
        {'event': 'palm_offered', 'match': {'entity': '$giantess', 'surface': '$desk1'}},
        {'event': 'palm_withdrawn', 'match': {'entity': '$giantess', 'surface': '$desk1', 'boarded': []}},
        {'no_event': 'explored', 'match': {'space': '$palm'}},
    ]
    write(os.path.join(SCENARIOS, 'nobody_boards_offered_palm.json'), d)

    d = copy.deepcopy(offer)
    d['name'] = 'offer_canceled_when_occupied'
    d['notes'] = ('Un tiny sin reflejos se queda quieto donde la giantess iba a apoyar la mano: la sombra no lo '
                  'aparta, así que al terminar el anuncio cancela el apoyo en vez de aplastarlo.')
    d['ticks'] = 400
    still = entity(d, 'tiny17')['spawn']
    still['location']['pos'] = {'x': 300000, 'y': 110000, 'z': 0}
    still['components']['agent'].update({'reflex': 0, 'curiosity': 0, 'work': 0})
    still['components']['member']['forage'] = 0
    still['components']['needs'] = [{'name': 'hunger', 'value': 900}, {'name': 'energy', 'value': 900}]
    d['expect'] = common + [
        {'event': 'encounter_skipped', 'match': {'entity': '$giantess', 'action': 'offer_palm', 'reason': 'occupied'}},
        {'no_event': 'palm_offered'},
        {'no_event': 'pressed', 'match': {'by': '$palm'}},
    ]
    write(os.path.join(SCENARIOS, 'offer_canceled_when_occupied.json'), d)


if __name__ == '__main__':
    main()
