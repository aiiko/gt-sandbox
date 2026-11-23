"""Genera los escenarios de Vida cotidiana de los tinies (M39 a M44).

    python tools/scenarios/partida.py   # antes: rehace content/partidas/casa.json
    python tools/scenarios/vida.py

Además suma a content/partidas/casa.json (la base de «Nueva partida») la vida
del nido: meta de la despensa, puntos de búsqueda, migajas por descubrir en
la silla, la servilleta, la obra del refugio junto a la pata del escritorio
y el punto de reunión, y ajusta el carácter de los tinies. Es idempotente.

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
         agent=None, member=None, extra=None, fun=None):
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
    if fun is not None:
        comps['needs'].append({'name': 'fun', 'value': fun, 'decay': decay[2] if len(decay) > 2 else 0})
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


# --- M40: puntos de búsqueda ---

def spot(ref, space, x, y, look=60, cooldown=2400, seen_from=1000000):
    return {'ref': ref, 'spawn': {
        'name': ref, 'location': {'space': '$' + space, 'pos': {'x': x, 'y': y, 'z': 0}},
        'components': {'search_spot': {'look_ticks': look, 'cooldown': cooldown},
                       'landmark': {'range': seen_from}}}}


SEARCHER = {'search': 0.15}


def colony_searches_spots():
    d = world('colony_searches_spots', 3600,
              'M40: el nido está vacío y su meta es 2 porciones por miembro. Tres tinies que no ven comida '
              'eligen puntos de búsqueda que se ven de lejos: el borde vacío (queda sin ofrecerse un rato) y '
              'el rincón, donde hay migajas. El hallazgo genera acarreos, la comida se guarda y se come. '
              'Nunca dos revisan el mismo punto (TestSearchSpotReservedOnce); con la despensa llena no sale '
              'nadie (TestStockedColonySearchesLess).')
    d['entities'] += [
        nest(50000, 50000, portions=0, stock=2),
        spot('corner', 'desk1', 550000, 350000),
        spot('edge', 'desk1', 550000, 50000),
        crumb('c1', 'desk1', 540000, 340000),
        crumb('c2', 'desk1', 560000, 360000),
    ] + [tiny('t%d' % i, 'desk1', 70000 + 10000 * i, 50000, hunger=700, decay=(4, 0), sight=60000,
              member=SEARCHER) for i in range(3)]
    d['expect'] = [
        {'event': 'searched', 'match': {'spot': '$edge', 'found': False}},
        {'event': 'searched', 'match': {'spot': '$corner', 'found': True}},
        {'event': 'job_posted', 'match': {'type': 'haul', 'target': '$c1'}},
        {'event': 'stored', 'match': {'colony': '$nest'}},
        {'event': 'action_done', 'match': {'what': 'eat', 'target': '$nest'}},
    ] + COMMON
    return d


def search_portal_closed():
    d = world('search_portal_closed', 1400,
              'M40: un tiny baja por la pata a revisar un punto del piso y la pata se cierra abajo mientras '
              'baja. Falla, libera el punto y no vuelve a intentarlo enseguida.')
    d['entities'] += [
        nest(300000, 200000, portions=0, stock=1),
        spot('floor', 'roomA', 1300000, 900000),
        tiny('t', 'desk1', 20000, 20000, sight=60000, member=SEARCHER,
             agent={'known': [{'entity': '$floor', 'since': 0}]}),
    ]
    d['commands'] += [{'tick': 200, 'kind': 'set_portal', 'data': {'portal': '$leg1_bottom', 'enabled': False}}]
    d['expect'] = [
        {'event': 'action_started', 'match': {'entity': '$t', 'action': 'search', 'target': '$floor'},
         'by': 20},
        {'event': 'entity_reparented', 'match': {'entity': '$t', 'to': '$leg1'}, 'by': 200},
        {'event': 'action_failed', 'match': {'entity': '$t', 'action': 'search', 'target': '$floor'}},
        {'event': 'reservation_released', 'match': {'target': '$floor', 'by': '$t'}},
        {'event': 'action_started', 'match': {'entity': '$t', 'action': 'search'}, 'by': 800, 'count': 1},
        {'no_event': 'searched'},
    ] + COMMON
    return d


# --- M41: fibras y refugio auxiliar ---

def napkin(x, y, left=12):
    return {'ref': 'napkin', 'object': 'napkin', 'spawn': {
        'name': 'servilleta', 'location': {'space': '$desk1', 'pos': {'x': x, 'y': y, 'z': 0}},
        'components': {'fiber_source': {'left': left, 'item': 'fiber'}}}}


def site(x, y, required=3, delivered=0):
    return {'ref': 'site', 'object': 'refuge_site', 'spawn': {
        'name': 'obra', 'location': {'space': '$desk1', 'pos': {'x': x, 'y': y, 'z': 0}},
        'components': {'build_site': {
            'colony': '$nest', 'source': '$napkin', 'required': required, 'delivered': delivered,
            'radius': 15000, 'shelter': {'strength': 20, 'radius': 15000}}}}}


def fiber(ref, x, y):
    return {'ref': ref, 'object': 'fiber', 'spawn': {
        'name': ref, 'location': {'space': '$desk1', 'pos': {'x': x, 'y': y, 'z': 0}},
        'components': {'material': {'site': '$site'}}}}


def colony_builds_refuge():
    d = world('colony_builds_refuge', 5000,
              'M41: la despensa está en la meta, así que la colonia arma el refugio auxiliar junto a la '
              'salida del escritorio. Los tinies desprenden fibras de la servilleta de a una, las llevan a la '
              'obra y, con las tres, uno la arma: la obra conserva su ID y recibe shelter.')
    d['entities'] += [
        nest(450000, 300000, portions=10, stock=1),
        napkin(350000, 100000),
        site(60000, 40000),
        tiny('t0', 'desk1', 400000, 280000, sight=60000),
        tiny('t1', 'desk1', 420000, 280000, sight=60000),
    ]
    d['expect'] = [
        {'event': 'job_posted', 'match': {'type': 'pluck', 'target': '$napkin'}, 'by': 20},
        {'event': 'fiber_plucked', 'match': {'source': '$napkin'}, 'count': 3},
        {'event': 'fiber_delivered', 'match': {'site': '$site', 'total': 3}, 'count': 1},
        {'event': 'job_posted', 'match': {'type': 'build', 'target': '$site'}, 'count': 1},
        {'event': 'refuge_built', 'match': {'site': '$site'}, 'count': 1},
        {'no_event': 'entity_despawned', 'match': {'entity': '$site'}},
    ] + COMMON
    return d


def fiber_dropped_is_recovered():
    d = world('fiber_dropped_is_recovered', 1600,
              'M41: un tiny lleva una fibra a la obra y la sombra de la palma lo hace huir: suelta la fibra. '
              'Otro miembro (que no huye de sombras) la recoge y la entrega; cuenta una sola vez.')
    d['entities'] += [
        nest(500000, 350000, portions=10, stock=1),
        napkin(550000, 50000, left=0),
        site(50000, 200000, required=1),
        fiber('f', 280000, 200000),
        tiny('t0', 'desk1', 300000, 200000, sight=60000, agent={'reflex': 4}),
        tiny('t1', 'desk1', 490000, 340000, hunger=300, sight=60000),
    ]
    d['commands'] += [pose(150, 'hover'), pose(700, 'raised')]
    d['expect'] = [
        {'event': 'job_assigned', 'match': {'by': '$t0'}, 'by': 20},
        {'event': 'alarmed', 'match': {'entity': '$t0'}},
        {'event': 'job_unassigned', 'match': {'by': '$t0', 'reason': 'stopped'}},
        {'event': 'job_assigned', 'match': {'by': '$t1'}},
        {'event': 'fiber_delivered', 'match': {'site': '$site', 'item': '$f', 'total': 1}, 'count': 1},
        {'event': 'refuge_built', 'match': {'site': '$site'}},
    ] + COMMON
    return d


def fragile_tiny(ref, x, y):
    return {'ref': ref, 'spawn': {
        'name': ref, 'location': {'space': '$desk1', 'pos': {'x': x, 'y': y, 'z': 0}},
        'body': {'radius': 2000, 'height': 8000, 'mass': 0}, 'components': {'fragile': {}}}}


def refuge_shelters_from_palm():
    d = world('refuge_shelters_from_palm', 400,
              'M41: la obra ya tiene sus fibras y un tiny la arma bajo donde después se apoya la palma. El '
              'refugio terminado (fuerza 20 contra la palma de 10) salva a quien está dentro; el que quedó '
              'afuera, a 6 cm, muere.')
    d['entities'] += [
        nest(500000, 350000, portions=10, stock=1),
        napkin(550000, 50000, left=0),
        site(200000, 200000, required=3, delivered=3),
        tiny('builder', 'desk1', 210000, 220000, sight=60000, extra={'fragile': {}}),
        fragile_tiny('inside', 205000, 200000),
        fragile_tiny('outside', 200000, 260000),
    ]
    d['commands'] += [pose(300, 'hover'), pose(330, 'rest')]
    d['expect'] = [
        {'event': 'refuge_built', 'match': {'site': '$site', 'by': '$builder'}, 'by': 300},
        {'event': 'sheltered', 'match': {'entity': '$inside', 'shelter': '$site'}, 'after': 330, 'by': 331},
        {'event': 'sheltered', 'match': {'entity': '$builder', 'shelter': '$site'}},
        {'no_event': 'crushed', 'match': {'entity': '$inside'}},
        {'event': 'crushed', 'match': {'entity': '$outside'}},
    ] + COMMON
    return d


# --- M42: reunión ---

def gathering(x, y, wait=400, talk=160, pause=600):
    return {'ref': 'meet', 'spawn': {
        'name': 'punto de reunión', 'location': {'space': '$desk1', 'pos': {'x': x, 'y': y, 'z': 0}},
        'components': {'gathering': {'wait': wait, 'talk_ticks': talk, 'pause': pause, 'gain': 300},
                       'landmark': {'range': 1000000}}}}


def tinies_chat():
    d = world('tinies_chat', 1200,
              'M42: dos tinies bien alimentados, descansados y aburridos van al punto de reunión junto al '
              'nido, conversan y a los dos les sube la diversión (TestChatRaisesFun). El punto queda en '
              'pausa y, al vencer, vuelven a reunirse.')
    d['entities'] += [
        nest(450000, 300000, portions=5),
        gathering(400000, 300000),
        tiny('a', 'desk1', 300000, 200000, fun=200),
        tiny('b', 'desk1', 500000, 150000, fun=250),
    ]
    d['expect'] = [
        {'event': 'chat_started', 'match': {'gathering': '$meet'}, 'by': 900, 'count': 1},
        {'event': 'chat_done', 'match': {'gathering': '$meet'}, 'by': 900, 'count': 1},
        {'event': 'chat_started', 'match': {'gathering': '$meet'}, 'after': 900},
        {'event': 'action_done', 'match': {'entity': '$a', 'action': 'chat'}},
        {'event': 'action_done', 'match': {'entity': '$b', 'action': 'chat'}},
        {'no_event': 'chat_abandoned'},
    ] + COMMON
    return d


def chat_partner_never_comes():
    d = world('chat_partner_never_comes', 900,
              'M42: un tiny aburrido va al punto de reunión, pero el otro tiny no tiene ganas de conversar. '
              'Espera solo hasta que vence el plazo, se retira y el punto queda libre.')
    d['entities'] += [
        nest(450000, 300000, portions=5),
        gathering(400000, 300000, wait=200),
        tiny('a', 'desk1', 300000, 200000, fun=200),
        tiny('b', 'desk1', 500000, 150000, fun=950),
    ]
    d['expect'] = [
        {'event': 'chat_abandoned', 'match': {'gathering': '$meet', 'entity': '$a', 'reason': 'alone'}},
        {'event': 'action_failed', 'match': {'entity': '$a', 'action': 'chat'}},
        {'no_event': 'chat_started'},
    ] + COMMON
    return d


def chat_interrupted_by_threat():
    d = world('chat_interrupted_by_threat', 700,
              'M42: dos tinies conversan bajo donde la giantess sostiene la mano. La sombra corta la '
              'conversación: los dos huyen y el punto queda libre, sin diversión ganada.')
    d['entities'] += [
        nest(500000, 350000, portions=5),
        gathering(200000, 200000, talk=400),
        tiny('a', 'desk1', 180000, 160000, fun=200, agent={'reflex': 4}),
        tiny('b', 'desk1', 230000, 240000, fun=250, agent={'reflex': 4}),
    ]
    d['commands'] += [pose(200, 'hover')]
    d['expect'] = [
        {'event': 'chat_started', 'match': {'gathering': '$meet'}, 'by': 200},
        {'event': 'action_interrupted', 'match': {'entity': '$a', 'action': 'chat', 'by': 'flee'}},
        {'event': 'action_started', 'match': {'entity': '$b', 'action': 'flee'}},
        {'no_event': 'chat_done'},
    ] + COMMON
    return d


# --- M43: partida libre ---

CASA = os.path.join(ROOT, 'content', 'partidas', 'casa.json')
LIFE_REFS = ['spot_corner', 'spot_back', 'spot_chair', 'spot_floor', 'chair_crumb1', 'chair_crumb2',
             'chair_crumb3', 'napkin', 'site', 'meet']
# Carácter de los tinies en la partida libre (los demás campos quedan).
TINY_AGENT = {'curiosity': 0.3}
TINY_MEMBER = {'search': 0.15}
TINY_FUN = {'name': 'fun', 'value': 700, 'decay': 2}


def on(space, ref, x, y, comps, name=None, obj=None):
    e = {'ref': ref, 'spawn': {'name': name or ref,
                               'location': {'space': '$' + space, 'pos': {'x': x, 'y': y, 'z': 0}},
                               'components': comps}}
    if obj:
        e['object'] = obj
    return e


def life_entities():
    def spot_on(space, ref, x, y, name):
        return on(space, ref, x, y, {'search_spot': {}, 'landmark': {'range': 1000000}}, name)

    def chair_crumb(ref, x, y):
        e = on('chair', ref, x, y, {}, 'crumb', obj='crumb')
        del e['spawn']['components']
        return e

    return [
        spot_on('desk1', 'spot_corner', 560000, 50000, 'rincón del escritorio'),
        spot_on('desk1', 'spot_back', 80000, 350000, 'fondo del escritorio'),
        spot_on('chair', 'spot_chair', 340000, 340000, 'borde de la silla'),
        spot_on('bedroom', 'spot_floor', 1350000, 900000, 'piso junto al escritorio'),
        chair_crumb('chair_crumb1', 330000, 330000),
        chair_crumb('chair_crumb2', 355000, 350000),
        chair_crumb('chair_crumb3', 340000, 365000),
        on('desk1', 'napkin', 200000, 300000, {'fiber_source': {'left': 12, 'item': 'fiber'}},
           'servilleta', obj='napkin'),
        on('desk1', 'site', 60000, 60000, {'build_site': {
            'colony': '$nest', 'source': '$napkin', 'required': 4, 'radius': 15000,
            'shelter': {'strength': 20, 'radius': 15000}}}, 'refugio', obj='refuge_site'),
        on('desk1', 'meet', 400000, 330000, {
            'gathering': {}, 'landmark': {'range': 1000000}}, 'punto de reunión'),
    ]


def with_life(d):
    d['entities'] = [e for e in d['entities'] if e['ref'] not in LIFE_REFS]
    for e in d['entities']:
        c = e['spawn'].setdefault('components', {})
        if e['ref'] == 'nest':
            c['colony'] = {'radius': 30000, 'stock': 2}
            c['portions'] = {'left': 0}
        if 'member' in c:
            c['agent'].update(TINY_AGENT)
            c['member'].update(TINY_MEMBER)
            c['needs'] = [n for n in c['needs'] if n['name'] != 'fun'] + [dict(TINY_FUN)]
    # la servilleta antes que la obra (la obra la nombra)
    i = next(i for i, e in enumerate(d['entities']) if e['ref'] == 'nest') + 1
    d['entities'][i:i] = life_entities()
    d['notes'] = d['notes'].split(' Vida del nido:')[0] + (
        ' Vida del nido: despensa con meta, puntos de búsqueda, migajas en la silla, servilleta, '
        'refugio por armar y punto de reunión (tools/scenarios/vida.py).')


def main():
    d = json.load(open(CASA, encoding='utf-8'))
    with_life(d)
    with open(CASA, 'w', encoding='utf-8', newline='\n') as out:
        out.write(json.dumps(d, ensure_ascii=False, indent=2) + '\n')

    for f in [tiny_returns_to_eat, tiny_rests_at_nest, nest_unreachable_eats_crumb,
              colony_searches_spots, search_portal_closed,
              colony_builds_refuge, fiber_dropped_is_recovered, refuge_shelters_from_palm,
              tinies_chat, chat_partner_never_comes, chat_interrupted_by_threat]:
        d = f()
        with open(os.path.join(SCEN, d['name'] + '.json'), 'w', encoding='utf-8', newline='\n') as out:
            json.dump(d, out, ensure_ascii=False, indent=2)
            out.write('\n')


if __name__ == '__main__':
    main()
