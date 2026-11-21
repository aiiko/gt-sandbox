"""Genera tests/scenarios/tiny_cae_de_la_palma.json.

Parte de palm_destroyed_tiny_falls.json (mismas habitaciones, muebles,
giantess y palma) y le agrega el atlas de la casa: un comando deja al tiny
fuera del borde de la palma levantada sobre el escritorio y cae 65 cm.
Uso: python tools/scenarios/tiny_cae_de_la_palma.py
"""
import json
import os

ROOT = os.path.join(os.path.dirname(__file__), '..', '..')
SRC = os.path.join(ROOT, 'tests', 'scenarios', 'palm_destroyed_tiny_falls.json')
OUT = os.path.join(ROOT, 'tests', 'scenarios', 'tiny_cae_de_la_palma.json')

VISUALS = {'roomA': 'bedroom', 'roomB': 'living', 'desk1': 'desk', 'desk2': 'table',
           'giantess': 'giantess', 'palm': 'palm', 'tiny': 'tiny', 'door': 'door'}


def main():
    d = json.load(open(SRC, encoding='utf-8'))
    d['name'] = 'tiny_cae_de_la_palma'
    d['ticks'] = 40
    d['atlas'] = '../../content/art/casa.json'
    d['notes'] = ('El tiny está en la palma levantada sobre el escritorio y un comando lo deja fuera del borde: '
                  'cae 65 cm sobre el escritorio. En el visor, la caída se anima y el seguimiento '
                  'pasa de la palma al escritorio.')
    for e in d['entities']:
        if e['ref'] in VISUALS:
            e['visual'] = VISUALS[e['ref']]
        if e['ref'] == 'tiny':
            e['spawn']['name'] = 'Tiny 01'
    d['commands'] = [{
        'tick': 10,
        'kind': 'set_location',
        'data': {'entity': '$tiny', 'location': {'space': '$palm', 'pos': {'x': 0, 'y': 100000, 'z': 0}}},
    }]
    d['expect'] = [
        {'event': 'entity_fell',
         'match': {'entity': '$tiny', 'from': '$palm', 'to': '$desk1', 'height': 650000}, 'by': 12},
        {'no_event': 'command_rejected'},
        {'at': '$tiny', 'space': '$desk1', 'pos': {'x': 100000, 'y': 300000}, 'tol': 0},
    ]
    with open(OUT, 'w', encoding='utf-8', newline='\n') as f:
        f.write(json.dumps(d, ensure_ascii=False, indent=2) + '\n')


if __name__ == '__main__':
    main()
