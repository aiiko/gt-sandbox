"""Genera el mapa corporal completo de la giantess (M58, propuesta 19),
integra futón/almohada y el mapa corporal en content/partidas/casa.json,
y genera los escenarios de integración de 19.6.3 en tests/scenarios/.

    python tools/scenarios/partida.py
    python tools/scenarios/vida.py
    python tools/scenarios/oficios.py
    python tools/scenarios/cuerpo.py

Es idempotente sobre content/partidas/casa.json.
"""
import copy
import json
import os

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
CASA = os.path.join(ROOT, 'content', 'partidas', 'casa.json')
SCEN = os.path.join(ROOT, 'tests', 'scenarios')

COMMON = [{'no_event': 'command_rejected'}]

BODY_REFS = [
    # Superficies navegables de descanso en el dormitorio
    'futon', 'pillow', 'futon_step', 'pillow_step',
    # Cabeza y cuello
    'body_head', 'body_neck',
    # Torso y caderas
    'body_shoulder_l', 'body_shoulder_r',
    'body_chest', 'body_back', 'body_side_l', 'body_side_r', 'body_hips',
    # Brazos y manos (palm ya existe como mano derecha)
    'body_hand_r_back', 'body_forearm_r', 'body_upper_arm_r',
    'body_hand_l', 'body_forearm_l', 'body_upper_arm_l',
    # Piernas y pies
    'body_thigh_l', 'body_calf_l', 'body_foot_l',
    'body_thigh_r', 'body_calf_r', 'body_foot_r',
    # Portales internos (articulaciones y bordes)
    'joint_hand_r_edge', 'joint_wrist_r', 'joint_elbow_r', 'joint_shoulder_r',
    'joint_wrist_l', 'joint_elbow_l', 'joint_shoulder_l',
    'joint_shoulder_r_chest', 'joint_shoulder_l_chest',
    'joint_shoulder_r_back', 'joint_shoulder_l_back',
    'joint_neck_chest', 'joint_neck_back', 'joint_neck_head',
    'joint_chest_side_l', 'joint_chest_side_r',
    'joint_back_side_l', 'joint_back_side_r',
    'joint_side_l_hips', 'joint_side_r_hips',
    'joint_chest_hips', 'joint_back_hips',
    'joint_hip_l', 'joint_knee_l', 'joint_ankle_l',
    'joint_hip_r', 'joint_knee_r', 'joint_ankle_r',
]


def _grid(w_cm, h_cm, cell_cm=2):
    cw = max(1, w_cm // cell_cm)
    ch = max(1, h_cm // cell_cm)
    return {
        'origin': {'x': 0, 'y': 0},
        'cell': cell_cm * 10000,
        'version': 0,
        'rows': ['.' * cw] * ch,
    }


def _rect(w_cm, h_cm):
    return {'min': {'x': 0, 'y': 0}, 'max': {'x': w_cm * 10000, 'y': h_cm * 10000}}


def _anchor_proj(a, b, climb=True):
    return {
        'anchor_a': {'x': a[0], 'y': a[1], 'z': a[2]},
        'anchor_b': {'x': b[0], 'y': b[1], 'z': b[2]},
        'climb': climb,
    }


def _portal(ref, name, sp_a, pa, sp_b, pb, cost=20000):
    return {
        'ref': ref,
        'spawn': {
            'name': name,
            'components': {
                'portal': {
                    'a': {'space': sp_a, 'pos': {'x': pa[0], 'y': pa[1]}},
                    'b': {'space': sp_b, 'pos': {'x': pb[0], 'y': pb[1]}},
                    'cost': cost,
                    'scale': 'tiny',
                    'version': 0,
                },
            },
        },
    }


def futon_entities():
    """Superficies navegables del futón y la almohada en el dormitorio."""
    return [
        {
            'ref': 'futon',
            'spawn': {
                'name': 'futón',
                'space': {'kind': 'elevated', 'extent': _rect(120, 180)},
                'location': {'space': '$bedroom', 'pos': {'x': 300000, 'y': 1200000, 'z': 150000}},
                'components': {
                    'nav_grid': _grid(120, 180, cell_cm=10),
                    'contact_surface': {'tolerance': 50000},
                },
            },
        },
        {
            'ref': 'pillow',
            'spawn': {
                'name': 'almohada',
                'space': {'kind': 'elevated', 'extent': _rect(60, 40)},
                'location': {'space': '$bedroom', 'pos': {'x': 600000, 'y': 2500000, 'z': 250000}},
                'components': {
                    'nav_grid': _grid(60, 40, cell_cm=5),
                    'contact_surface': {'tolerance': 50000},
                },
            },
        },
        _portal('futon_step', 'borde del futón', '$bedroom', (350000, 1200000), '$futon', (100000, 100000)),
        _portal('pillow_step', 'borde de la almohada', '$futon', (500000, 1500000), '$pillow', (200000, 100000)),
    ]


def body_entities(giantess_ref='$giantess', palm_ref='$palm'):
    """Red completa de regiones corporales y portales internos de la giantess."""
    ents = [
        # --- Cabeza y cuello ---
        {
            'ref': 'body_head',
            'spawn': {
                'name': 'cabello y rostro',
                'space': {'kind': 'articulated', 'extent': _rect(20, 22)},
                'location': {'space': giantess_ref, 'pos': {'x': -100000, 'y': -100000, 'z': 1550000}},
                'components': {
                    'nav_grid': _grid(20, 22, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'cabello y rostro', 'occluded': True,
                        'grips': [
                            {'pos': {'x': 40000, 'y': 100000}, 'radius': 50000, 'label': 'mechón izquierdo'},
                            {'pos': {'x': 160000, 'y': 100000}, 'radius': 50000, 'label': 'mechón derecho'},
                        ],
                    },
                    'contact_surface': {
                        'tolerance': 60000, 'scale': 'tiny', 'stable_ticks': 2,
                        'support_offset': -20000,
                        'access': [{'x': 100000, 'y': 180000}, {'x': 40000, 'y': 100000}],
                        'access_reach': 80000, 'body': giantess_ref,
                    },
                },
            },
        },
        {
            'ref': 'body_neck',
            'spawn': {
                'name': 'cuello y nuca',
                'space': {
                    'kind': 'articulated', 'extent': _rect(8, 12),
                    'projection': _anchor_proj((0, 0, 1420000), (0, 0, 1550000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(8, 12, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'cuello y nuca', 'require_grip': True,
                        'grips': [{'pos': {'x': 40000, 'y': 60000}, 'radius': 50000, 'label': 'cuello de la ropa'}],
                    },
                },
            },
        },
        # --- Hombros, torso y caderas ---
        {
            'ref': 'body_shoulder_r',
            'spawn': {
                'name': 'hombro derecho',
                'space': {'kind': 'articulated', 'extent': _rect(12, 12)},
                'location': {'space': giantess_ref, 'pos': {'x': 140000, 'y': -60000, 'z': 1400000}},
                'components': {
                    'nav_grid': _grid(12, 12, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'hombro derecho',
                        'grips': [{'pos': {'x': 60000, 'y': 60000}, 'radius': 50000, 'label': 'costura del hombro derecho'}],
                    },
                },
            },
        },
        {
            'ref': 'body_shoulder_l',
            'spawn': {
                'name': 'hombro izquierdo',
                'space': {'kind': 'articulated', 'extent': _rect(12, 12)},
                'location': {'space': giantess_ref, 'pos': {'x': -260000, 'y': -60000, 'z': 1400000}},
                'components': {
                    'nav_grid': _grid(12, 12, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'hombro izquierdo',
                        'grips': [{'pos': {'x': 60000, 'y': 60000}, 'radius': 50000, 'label': 'costura del hombro izquierdo'}],
                    },
                },
            },
        },
        {
            'ref': 'body_chest',
            'spawn': {
                'name': 'pecho',
                'space': {
                    'kind': 'articulated', 'extent': _rect(28, 36),
                    'projection': _anchor_proj((0, 100000, 1040000), (0, 100000, 1400000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(28, 36, cell_cm=4),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'pecho', 'require_grip': True,
                        'grips': [{'pos': {'x': 140000, 'y': 180000}, 'radius': 80000, 'label': 'borde del bolsillo'}],
                    },
                },
            },
        },
        {
            'ref': 'body_back',
            'spawn': {
                'name': 'espalda',
                'space': {
                    'kind': 'articulated', 'extent': _rect(28, 36),
                    'projection': _anchor_proj((0, -100000, 1040000), (0, -100000, 1400000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(28, 36, cell_cm=4),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'espalda', 'require_grip': True,
                        'grips': [{'pos': {'x': 140000, 'y': 180000}, 'radius': 80000, 'label': 'costura de la espalda'}],
                    },
                    'contact_surface': {
                        'tolerance': 60000, 'scale': 'tiny', 'stable_ticks': 2,
                        'support_offset': -20000,
                        'access': [{'x': 20000, 'y': 180000}, {'x': 260000, 'y': 180000}],
                        'access_reach': 80000, 'body': giantess_ref,
                    },
                },
            },
        },
        {
            'ref': 'body_side_r',
            'spawn': {
                'name': 'costado derecho',
                'space': {
                    'kind': 'articulated', 'extent': _rect(10, 34),
                    'projection': _anchor_proj((180000, 0, 1040000), (180000, 0, 1380000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(10, 34, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'costado derecho', 'require_grip': True,
                        'grips': [{'pos': {'x': 50000, 'y': 160000}, 'radius': 60000, 'label': 'costura lateral derecha'}],
                    },
                },
            },
        },
        {
            'ref': 'body_side_l',
            'spawn': {
                'name': 'costado izquierdo',
                'space': {
                    'kind': 'articulated', 'extent': _rect(10, 34),
                    'projection': _anchor_proj((-180000, 0, 1040000), (-180000, 0, 1380000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(10, 34, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'costado izquierdo', 'require_grip': True,
                        'grips': [{'pos': {'x': 50000, 'y': 160000}, 'radius': 60000, 'label': 'costura lateral izquierda'}],
                    },
                    'contact_surface': {
                        'tolerance': 60000, 'scale': 'tiny', 'stable_ticks': 2,
                        'support_offset': -20000,
                        'access': [{'x': 50000, 'y': 160000}],
                        'access_reach': 80000, 'body': giantess_ref,
                    },
                },
            },
        },
        {
            'ref': 'body_hips',
            'spawn': {
                'name': 'cintura y caderas',
                'space': {'kind': 'articulated', 'extent': _rect(32, 20)},
                'location': {'space': giantess_ref, 'pos': {'x': -160000, 'y': -100000, 'z': 950000}},
                'components': {
                    'nav_grid': _grid(32, 20, cell_cm=4),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'cintura y caderas',
                        'grips': [{'pos': {'x': 160000, 'y': 100000}, 'radius': 60000, 'label': 'cintura'}],
                    },
                },
            },
        },
        # --- Brazo y mano derechos (conectados a $palm) ---
        {
            'ref': 'body_hand_r_back',
            'spawn': {
                'name': 'dorso de mano derecha',
                'space': {'kind': 'articulated', 'extent': _rect(10, 14)},
                'location': {'space': giantess_ref, 'pos': {'x': 300000, 'y': -70000, 'z': 1400000}},
                'components': {
                    'nav_grid': _grid(10, 14, cell_cm=2),
                    'body_zone': {'body': giantess_ref, 'region': 'mano derecha'},
                },
            },
        },
        {
            'ref': 'body_forearm_r',
            'spawn': {
                'name': 'antebrazo derecho',
                'space': {
                    'kind': 'articulated', 'extent': _rect(6, 12),
                    'projection': _anchor_proj((300000, 0, 1100000), (260000, 0, 1250000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(6, 12, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'antebrazo derecho', 'require_grip': True, 'occluded': True,
                        'grips': [{'pos': {'x': 30000, 'y': 20000}, 'radius': 50000, 'label': 'puño de la manga derecha'}],
                    },
                    'contact_surface': {
                        'tolerance': 50000, 'scale': 'tiny', 'stable_ticks': 2,
                        'support_offset': -20000,
                        'access': [{'x': 30000, 'y': 60000}],
                        'access_reach': 70000, 'body': giantess_ref,
                    },
                },
            },
        },
        {
            'ref': 'body_upper_arm_r',
            'spawn': {
                'name': 'brazo derecho',
                'space': {
                    'kind': 'articulated', 'extent': _rect(6, 12),
                    'projection': _anchor_proj((260000, 0, 1250000), (200000, 0, 1400000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(6, 12, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'brazo derecho', 'require_grip': True,
                        'grips': [{'pos': {'x': 30000, 'y': 60000}, 'radius': 50000, 'label': 'costura del brazo derecho'}],
                    },
                },
            },
        },
        # --- Brazo y mano izquierdos ---
        {
            'ref': 'body_hand_l',
            'spawn': {
                'name': 'mano izquierda',
                'space': {'kind': 'articulated', 'extent': _rect(10, 14)},
                'location': {'space': giantess_ref, 'pos': {'x': -320000, 'y': -70000, 'z': 950000}},
                'components': {
                    'nav_grid': _grid(10, 14, cell_cm=2),
                    'body_zone': {'body': giantess_ref, 'region': 'mano izquierda'},
                    'contact_surface': {
                        'tolerance': 50000, 'scale': 'tiny', 'stable_ticks': 2,
                        'support_offset': -20000,
                        'access': [{'x': 20000, 'y': 70000}],
                        'access_reach': 70000, 'body': giantess_ref, 'self_contact': True,
                    },
                },
            },
        },
        {
            'ref': 'body_forearm_l',
            'spawn': {
                'name': 'antebrazo izquierdo',
                'space': {
                    'kind': 'articulated', 'extent': _rect(6, 12),
                    'projection': _anchor_proj((-280000, 0, 950000), (-260000, 0, 1180000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(6, 12, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'antebrazo izquierdo', 'require_grip': True, 'occluded': True,
                        'grips': [{'pos': {'x': 30000, 'y': 20000}, 'radius': 50000, 'label': 'puño de la manga izquierda'}],
                    },
                    'contact_surface': {
                        'tolerance': 50000, 'scale': 'tiny', 'stable_ticks': 2,
                        'support_offset': -20000,
                        'access': [{'x': 30000, 'y': 60000}],
                        'access_reach': 70000, 'body': giantess_ref,
                    },
                },
            },
        },
        {
            'ref': 'body_upper_arm_l',
            'spawn': {
                'name': 'brazo izquierdo',
                'space': {
                    'kind': 'articulated', 'extent': _rect(6, 12),
                    'projection': _anchor_proj((-260000, 0, 1180000), (-200000, 0, 1400000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(6, 12, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'brazo izquierdo', 'require_grip': True,
                        'grips': [{'pos': {'x': 30000, 'y': 60000}, 'radius': 50000, 'label': 'costura del brazo izquierdo'}],
                    },
                },
            },
        },
        # --- Pierna y pie izquierdos ---
        {
            'ref': 'body_thigh_l',
            'spawn': {
                'name': 'muslo izquierdo',
                'space': {
                    'kind': 'articulated', 'extent': _rect(14, 34),
                    'projection': _anchor_proj((-100000, 0, 500000), (-100000, 0, 950000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(14, 34, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'muslo izquierdo', 'require_grip': True,
                        'grips': [{'pos': {'x': 70000, 'y': 160000}, 'radius': 60000, 'label': 'pliegue del muslo izquierdo'}],
                    },
                    'contact_surface': {
                        'tolerance': 60000, 'scale': 'tiny', 'stable_ticks': 2,
                        'support_offset': -20000,
                        'access': [{'x': 20000, 'y': 160000}, {'x': 120000, 'y': 160000}],
                        'access_reach': 80000, 'body': giantess_ref, 'self_contact': True,
                    },
                },
            },
        },
        {
            'ref': 'body_calf_l',
            'spawn': {
                'name': 'pantorrilla izquierda',
                'space': {
                    'kind': 'articulated', 'extent': _rect(8, 16),
                    'projection': _anchor_proj((-100000, 0, 60000), (-100000, 0, 500000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(8, 16, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'pantorrilla izquierda', 'require_grip': True,
                        'grips': [{'pos': {'x': 40000, 'y': 80000}, 'radius': 50000, 'label': 'costura de la pantorrilla izquierda'}],
                    },
                },
            },
        },
        {
            'ref': 'body_foot_l',
            'spawn': {
                'name': 'pie izquierdo',
                'space': {'kind': 'articulated', 'extent': _rect(10, 24)},
                'location': {'space': giantess_ref, 'pos': {'x': -150000, 'y': -80000, 'z': 20000}},
                'components': {
                    'nav_grid': _grid(10, 24, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'pie izquierdo',
                        'grips': [{'pos': {'x': 20000, 'y': 120000}, 'radius': 50000, 'label': 'borde del calzado izquierdo'}],
                    },
                    'contact_surface': {
                        'tolerance': 40000, 'scale': 'tiny', 'stable_ticks': 2, 'require_still': True,
                        'support_offset': -20000,
                        'access': [{'x': 20000, 'y': 120000}],
                        'access_reach': 80000, 'body': giantess_ref,
                    },
                },
            },
        },
        # --- Pierna y pie derechos ---
        {
            'ref': 'body_thigh_r',
            'spawn': {
                'name': 'muslo derecho',
                'space': {
                    'kind': 'articulated', 'extent': _rect(14, 34),
                    'projection': _anchor_proj((100000, 0, 500000), (100000, 0, 950000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(14, 34, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'muslo derecho', 'require_grip': True,
                        'grips': [{'pos': {'x': 70000, 'y': 160000}, 'radius': 60000, 'label': 'pliegue del muslo derecho'}],
                    },
                    'contact_surface': {
                        'tolerance': 60000, 'scale': 'tiny', 'stable_ticks': 2,
                        'support_offset': -20000,
                        'access': [{'x': 20000, 'y': 160000}, {'x': 120000, 'y': 160000}],
                        'access_reach': 80000, 'body': giantess_ref, 'self_contact': True,
                    },
                },
            },
        },
        {
            'ref': 'body_calf_r',
            'spawn': {
                'name': 'pantorrilla derecha',
                'space': {
                    'kind': 'articulated', 'extent': _rect(8, 16),
                    'projection': _anchor_proj((100000, 0, 60000), (100000, 0, 500000), climb=True),
                },
                'location': {'space': giantess_ref, 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {
                    'nav_grid': _grid(8, 16, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'pantorrilla derecha', 'require_grip': True,
                        'grips': [{'pos': {'x': 40000, 'y': 80000}, 'radius': 50000, 'label': 'costura de la pantorrilla derecha'}],
                    },
                },
            },
        },
        {
            'ref': 'body_foot_r',
            'spawn': {
                'name': 'pie derecho',
                'space': {'kind': 'articulated', 'extent': _rect(10, 24)},
                'location': {'space': giantess_ref, 'pos': {'x': 50000, 'y': -80000, 'z': 20000}},
                'components': {
                    'nav_grid': _grid(10, 24, cell_cm=2),
                    'body_zone': {
                        'body': giantess_ref, 'region': 'pie derecho',
                        'grips': [{'pos': {'x': 80000, 'y': 120000}, 'radius': 50000, 'label': 'borde del calzado derecho'}],
                    },
                    'contact_surface': {
                        'tolerance': 40000, 'scale': 'tiny', 'stable_ticks': 2, 'require_still': True,
                        'support_offset': -20000,
                        'access': [{'x': 80000, 'y': 120000}],
                        'access_reach': 80000, 'body': giantess_ref,
                    },
                },
            },
        },
    ]

    # --- Portales internos del grafo corporal ---
    # Nota: en $palm (casa.json), el centro (0, 0) y (-10000, -70000) son celdas transitables.
    joints = [
        _portal('joint_hand_r_edge', 'borde de mano derecha', palm_ref, (-10000, -70000), '$body_hand_r_back', (50000, 20000)),
        _portal('joint_wrist_r', 'muñeca derecha', palm_ref, (-10000, -80000), '$body_forearm_r', (30000, 10000)),
        _portal('joint_elbow_r', 'codo derecho', '$body_forearm_r', (30000, 110000), '$body_upper_arm_r', (30000, 10000)),
        _portal('joint_shoulder_r', 'hombro derecho', '$body_upper_arm_r', (30000, 110000), '$body_shoulder_r', (60000, 60000)),

        _portal('joint_wrist_l', 'muñeca izquierda', '$body_hand_l', (50000, 120000), '$body_forearm_l', (30000, 10000)),
        _portal('joint_elbow_l', 'codo izquierdo', '$body_forearm_l', (30000, 110000), '$body_upper_arm_l', (30000, 10000)),
        _portal('joint_shoulder_l', 'hombro izquierdo', '$body_upper_arm_l', (30000, 110000), '$body_shoulder_l', (60000, 60000)),

        _portal('joint_shoulder_r_chest', 'hombro derecho a pecho', '$body_shoulder_r', (20000, 60000), '$body_chest', (240000, 320000)),
        _portal('joint_shoulder_l_chest', 'hombro izquierdo a pecho', '$body_shoulder_l', (100000, 60000), '$body_chest', (40000, 320000)),
        _portal('joint_shoulder_r_back', 'hombro derecho a espalda', '$body_shoulder_r', (20000, 20000), '$body_back', (240000, 320000)),
        _portal('joint_shoulder_l_back', 'hombro izquierdo a espalda', '$body_shoulder_l', (100000, 20000), '$body_back', (40000, 320000)),

        _portal('joint_neck_chest', 'cuello a pecho', '$body_chest', (140000, 340000), '$body_neck', (40000, 20000)),
        _portal('joint_neck_back', 'nuca a espalda', '$body_back', (140000, 340000), '$body_neck', (40000, 20000)),
        _portal('joint_neck_head', 'nuca a cabello', '$body_neck', (40000, 100000), '$body_head', (100000, 20000)),

        _portal('joint_chest_side_l', 'pecho a costado izquierdo', '$body_chest', (20000, 180000), '$body_side_l', (50000, 160000)),
        _portal('joint_chest_side_r', 'pecho a costado derecho', '$body_chest', (260000, 180000), '$body_side_r', (50000, 160000)),
        _portal('joint_back_side_l', 'espalda a costado izquierdo', '$body_back', (20000, 180000), '$body_side_l', (50000, 180000)),
        _portal('joint_back_side_r', 'espalda a costado derecho', '$body_back', (260000, 180000), '$body_side_r', (50000, 180000)),
        _portal('joint_side_l_hips', 'costado izquierdo a cadera', '$body_side_l', (50000, 20000), '$body_hips', (40000, 100000)),
        _portal('joint_side_r_hips', 'costado derecho a cadera', '$body_side_r', (50000, 20000), '$body_hips', (280000, 100000)),
        _portal('joint_chest_hips', 'pecho a cintura', '$body_chest', (140000, 20000), '$body_hips', (160000, 160000)),
        _portal('joint_back_hips', 'espalda a cintura', '$body_back', (140000, 20000), '$body_hips', (160000, 40000)),

        _portal('joint_hip_l', 'cadera a muslo izquierdo', '$body_hips', (60000, 100000), '$body_thigh_l', (70000, 320000)),
        _portal('joint_knee_l', 'rodilla izquierda', '$body_thigh_l', (70000, 20000), '$body_calf_l', (40000, 140000)),
        _portal('joint_ankle_l', 'tobillo izquierdo', '$body_calf_l', (40000, 20000), '$body_foot_l', (50000, 120000)),

        _portal('joint_hip_r', 'cadera a muslo derecho', '$body_hips', (260000, 100000), '$body_thigh_r', (70000, 320000)),
        _portal('joint_knee_r', 'rodilla derecha', '$body_thigh_r', (70000, 20000), '$body_calf_r', (40000, 140000)),
        _portal('joint_ankle_r', 'tobillo derecho', '$body_calf_r', (40000, 20000), '$body_foot_r', (50000, 120000)),
    ]
    return ents + joints


def body_joints_component():
    return [
        {'portal': '$joint_elbow_r', 'closed_poses': ['elbow_bent']},
        {'portal': '$joint_elbow_l', 'closed_poses': ['elbow_bent']},
    ]


def body_pose_parts(mode):
    """Partes corporales (sin $palm ni $seat, que se conservan de cada pose base)."""
    if mode == 'sit':
        return [
            {'part': '$body_hips', 'pos': {'x': -160000, 'y': -100000, 'z': 470000}, 'facing': 0},
            {'part': '$body_thigh_l', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
             'projection': {'origin': {'x': -170000, 'y': -100000, 'z': 470000},
                            'axis_u': {'x': 1000000, 'y': 0, 'z': 0},
                            'axis_v': {'x': 0, 'y': 1000000, 'z': 0}}},
            {'part': '$body_thigh_r', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
             'projection': {'origin': {'x': 30000, 'y': -100000, 'z': 470000},
                            'axis_u': {'x': 1000000, 'y': 0, 'z': 0},
                            'axis_v': {'x': 0, 'y': 1000000, 'z': 0}}},
            {'part': '$body_calf_l', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
             'projection': _anchor_proj((-100000, 240000, 60000), (-100000, 240000, 470000), climb=True)},
            {'part': '$body_calf_r', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
             'projection': _anchor_proj((100000, 240000, 60000), (100000, 240000, 470000), climb=True)},
            {'part': '$body_foot_l', 'pos': {'x': -150000, 'y': 160000, 'z': 20000}, 'facing': 0},
            {'part': '$body_foot_r', 'pos': {'x': 50000, 'y': 160000, 'z': 20000}, 'facing': 0},
            {'part': '$body_forearm_l', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
             'projection': {'origin': {'x': -280000, 'y': -60000, 'z': 470000},
                            'axis_u': {'x': 1000000, 'y': 0, 'z': 0},
                            'axis_v': {'x': 0, 'y': 1000000, 'z': 0}}},
            {'part': '$body_hand_l', 'pos': {'x': -150000, 'y': 120000, 'z': 490000}, 'facing': 0},
            {'part': '$body_head', 'pos': {'x': -100000, 'y': -100000, 'z': 1200000}, 'facing': 0},
        ]
    if mode == 'lie_down':
        return [
            {'part': '$body_hips', 'pos': {'x': -160000, 'y': -100000, 'z': 170000}, 'facing': 0},
            {'part': '$body_back', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
             'projection': {'origin': {'x': -140000, 'y': -180000, 'z': 170000},
                            'axis_u': {'x': 1000000, 'y': 0, 'z': 0},
                            'axis_v': {'x': 0, 'y': 1000000, 'z': 0}}},
            {'part': '$body_side_l', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
             'projection': {'origin': {'x': -240000, 'y': -170000, 'z': 170000},
                            'axis_u': {'x': 1000000, 'y': 0, 'z': 0},
                            'axis_v': {'x': 0, 'y': 1000000, 'z': 0}}},
            {'part': '$body_head', 'pos': {'x': -100000, 'y': 420000, 'z': 270000}, 'facing': 0},
            {'part': '$body_neck', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
             'projection': _anchor_proj((0, 300000, 200000), (0, 420000, 270000), climb=False)},
            {'part': '$body_foot_l', 'pos': {'x': -150000, 'y': -800000, 'z': 170000}, 'facing': 0},
            {'part': '$body_foot_r', 'pos': {'x': 50000, 'y': -800000, 'z': 170000}, 'facing': 0},
        ]
    # stand / hand_rest
    forearm_r = _anchor_proj((300000, 0, 1100000), (260000, 0, 1250000), climb=True)
    if mode == 'hand_rest':
        forearm_r = _anchor_proj((400000, 0, 770000), (260000, 0, 1120000), climb=True)
    return [
        {'part': '$body_hips', 'pos': {'x': -160000, 'y': -100000, 'z': 950000}, 'facing': 0},
        {'part': '$body_thigh_l', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
         'projection': _anchor_proj((-100000, 0, 500000), (-100000, 0, 950000), climb=True)},
        {'part': '$body_thigh_r', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
         'projection': _anchor_proj((100000, 0, 500000), (100000, 0, 950000), climb=True)},
        {'part': '$body_calf_l', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
         'projection': _anchor_proj((-100000, 0, 60000), (-100000, 0, 500000), climb=True)},
        {'part': '$body_calf_r', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
         'projection': _anchor_proj((100000, 0, 60000), (100000, 0, 500000), climb=True)},
        {'part': '$body_foot_l', 'pos': {'x': -150000, 'y': -80000, 'z': 20000}, 'facing': 0},
        {'part': '$body_foot_r', 'pos': {'x': 50000, 'y': -80000, 'z': 20000}, 'facing': 0},
        {'part': '$body_forearm_r', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0, 'projection': forearm_r},
        {'part': '$body_forearm_l', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
         'projection': _anchor_proj((-280000, 0, 950000), (-260000, 0, 1180000), climb=True)},
        {'part': '$body_hand_l', 'pos': {'x': -320000, 'y': -70000, 'z': 950000}, 'facing': 0},
        {'part': '$body_back', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
         'projection': _anchor_proj((0, -100000, 1040000), (0, -100000, 1400000), climb=True)},
        {'part': '$body_side_l', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
         'projection': _anchor_proj((-180000, 0, 1040000), (-180000, 0, 1380000), climb=True)},
        {'part': '$body_head', 'pos': {'x': -100000, 'y': -100000, 'z': 1550000}, 'facing': 0},
        {'part': '$body_neck', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
         'projection': _anchor_proj((0, 0, 1420000), (0, 0, 1550000), climb=True)},
    ]


def with_body_map(d):
    """Integra idempotentemente el mapa corporal completo y futón/almohada en casa.json."""
    d['entities'] = [e for e in d['entities'] if e['ref'] not in BODY_REFS]
    # Agregar contact_surface a bedroom y living para que los pies puedan conectar con el piso.
    for ref in ('bedroom', 'living'):
        for e in d['entities']:
            if e['ref'] == ref:
                e['spawn'].setdefault('components', {}).setdefault('contact_surface', {'tolerance': 40000})
    # Agregar futón, almohada y el mapa corporal al final para conservar los IDs
    # de las entidades existentes de la casa y la colonia.
    d['entities'].extend(futon_entities())
    d['entities'].extend(body_entities())

    # Actualizar poses y agregar body_joints + body_state en el comando de setup de giantess.
    setup = next(c for c in d['commands'] if c['kind'] == 'set_components' and c['data'].get('entity') == '$giantess')
    comps = setup['data']['components']
    poses = comps['poses']
    for name, parts in list(poses.items()):
        # Quitar partes corporales viejas si se re-ejecuta el script.
        base = [p for p in parts if not p['part'].startswith('$body_')]
        if name.startswith('sit'):
            poses[name] = base + body_pose_parts('sit')
        elif 'hand_rest' in name:
            poses[name] = base + body_pose_parts('hand_rest')
        else:
            poses[name] = base + body_pose_parts('stand')
    poses['elbow_bent'] = [
        {'part': '$palm', 'pos': {'x': 300000, 'y': 0, 'z': 1300000}, 'facing': 0},
        {'part': '$seat', 'pos': {'x': 0, 'y': 0, 'z': 950000}, 'facing': 0},
    ] + body_pose_parts('stand')
    poses['lie_down'] = [
        {'part': '$palm', 'pos': {'x': 300000, 'y': 0, 'z': 200000}, 'facing': 0},
        {'part': '$seat', 'pos': {'x': 0, 'y': 0, 'z': 170000}, 'facing': 0},
    ] + body_pose_parts('lie_down')

    comps['body_joints'] = body_joints_component()
    comps['body_state'] = {'pose': 'idle'}

    # Actualizar la oferta sleep de bed para que use la pose lie_down (abre pelo->almohada y espalda->futón).
    bed = next(e for e in d['entities'] if e['ref'] == 'bed')
    bed['spawn'].setdefault('components', {})['offers'] = [{
        'action': 'sleep', 'scale': 'giantess',
        'effects': [{'need': 'energy', 'amount': 800}],
        'steps': [
            {'do': 'goto'},
            {'do': 'pose', 'pose': 'lie_down'},
            {'do': 'use', 'ticks': 600},
        ],
    }]

    d['notes'] = d['notes'].split(' Mapa corporal:')[0] + (
        ' Mapa corporal: regiones completas de la giantess, articulaciones, apoyos, '
        'futón y almohada (tools/scenarios/cuerpo.py).')


def desk_to_pillow_scenario():
    """19.6.3a: escritorio -> palma -> antebrazo -> brazo -> hombro -> cuello -> cabello -> almohada,
    con regreso al cambiar el apoyo (bajada por costado al futón)."""
    palm_ent = {
        'ref': 'palm',
        'spawn': {
            'name': 'palma',
            'space': {'kind': 'articulated', 'extent': {'min': {'x': -50000, 'y': -90000}, 'max': {'x': 50000, 'y': 90000}}},
            'location': {'space': '$giantess', 'pos': {'x': 400000, 'y': 0, 'z': 770000}},
            'components': {
                'contact_surface': {'tolerance': 40000, 'min_overlap': 10000, 'scale': 'tiny'},
            },
        },
    }
    poses = {
        'hand_rest': [
            {'part': '$palm', 'pos': {'x': 400000, 'y': 0, 'z': 770000}, 'facing': 0},
        ] + body_pose_parts('hand_rest'),
        'idle': [
            {'part': '$palm', 'pos': {'x': 300000, 'y': 0, 'z': 1400000}, 'facing': 0},
        ] + body_pose_parts('stand'),
        'lie_down': [
            {'part': '$palm', 'pos': {'x': 300000, 'y': 0, 'z': 200000}, 'facing': 0},
        ] + body_pose_parts('lie_down'),
    }
    return {
        'name': 'body_desk_to_pillow',
        'seed': 191,
        'ticks': 1400,
        'notes': (
            '19.6.3: ruta completa escritorio -> palma -> brazo -> hombro -> cuello -> cabello -> almohada '
            'al recostarse en el futón, y regreso hacia el futón.'
        ),
        'entities': [
            {'ref': 'house', 'spawn': {'name': 'house', 'space': {'kind': 'static', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 0, 'y': 0}}}}},
            {'ref': 'bedroom', 'spawn': {
                'name': 'dormitorio',
                'space': {'kind': 'static', 'extent': _rect(300, 300)},
                'location': {'space': '$house', 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {'nav_grid': _grid(300, 300, cell_cm=10), 'contact_surface': {'tolerance': 40000}},
            }},
            {'ref': 'desk1', 'spawn': {
                'name': 'escritorio',
                'space': {'kind': 'elevated', 'extent': _rect(80, 60)},
                'location': {'space': '$bedroom', 'pos': {'x': 1500000, 'y': 1000000, 'z': 750000}},
                'components': {'nav_grid': _grid(80, 60, cell_cm=5), 'contact_surface': {'tolerance': 40000}},
            }},
            *futon_entities(),
            {'ref': 'giantess', 'spawn': {
                'name': 'giantess',
                'space': {'kind': 'body', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 0, 'y': 0}}},
                'location': {'space': '$bedroom', 'pos': {'x': 1400000, 'y': 1200000, 'z': 0}},
                'components': {'mover': {'speed': 1000000, 'scale': 'giantess'}},
            }},
            palm_ent,
            *body_entities(),
            {'ref': 'tiny1', 'spawn': {
                'name': 'Tiny 01',
                'location': {'space': '$desk1', 'pos': {'x': 300000, 'y': 200000, 'z': 0}},
                'body': {'radius': 2000, 'height': 8000, 'mass': 1},
                'components': {'mover': {'speed': 150000, 'scale': 'tiny'}},
            }},
        ],
        'commands': [
            {'tick': 0, 'kind': 'set_components', 'data': {
                'entity': '$giantess',
                'components': {'poses': poses, 'body_joints': body_joints_component(), 'body_state': {'pose': 'hand_rest'}},
            }},
            # 1. Tiny sube desde el escritorio por la palma hasta el cabello.
            {'tick': 10, 'kind': 'move_to', 'data': {'entity': '$tiny1', 'space': '$body_head', 'goal': {'x': 100000, 'y': 100000}}},
            # 2. Giantess levanta la mano, camina al futón y se recuesta (cabeza sobre almohada).
            {'tick': 450, 'kind': 'set_pose', 'data': {'entity': '$giantess', 'pose': 'idle'}},
            {'tick': 455, 'kind': 'move_to', 'data': {'entity': '$giantess', 'space': '$bedroom', 'goal': {'x': 850000, 'y': 2200000}}},
            {'tick': 650, 'kind': 'set_pose', 'data': {'entity': '$giantess', 'pose': 'lie_down'}},
            # 3. Tiny desembarca desde el cabello a la almohada.
            {'tick': 670, 'kind': 'move_to', 'data': {'entity': '$tiny1', 'space': '$pillow', 'goal': {'x': 300000, 'y': 200000}}},
            # 4. Luego baja desde la almohada al futón.
            {'tick': 900, 'kind': 'move_to', 'data': {'entity': '$tiny1', 'space': '$futon', 'goal': {'x': 500000, 'y': 1000000}}},
        ],
        'expect': COMMON + [
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$palm'}},
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$body_shoulder_r'}},
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$body_head'}},
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$pillow'}},
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$futon'}},
            {'at': '$tiny1', 'space': '$futon', 'pos': {'x': 500000, 'y': 1000000}, 'tol': 30000},
        ],
    }


def floor_to_sofa_scenario():
    """19.6.3b: piso -> pie izquierdo -> pantorrilla -> muslo -> sofá al sentarse,
    más una caída que aterriza en un soporte visible (el sofá)."""
    palm_ent = {
        'ref': 'palm',
        'spawn': {
            'name': 'palma',
            'space': {'kind': 'articulated', 'extent': {'min': {'x': -50000, 'y': -90000}, 'max': {'x': 50000, 'y': 90000}}},
            'location': {'space': '$giantess', 'pos': {'x': 300000, 'y': 0, 'z': 1400000}},
            'components': {'contact_surface': {'tolerance': 40000, 'scale': 'tiny'}},
        },
    }
    poses = {
        'idle': [{'part': '$palm', 'pos': {'x': 300000, 'y': 0, 'z': 1400000}, 'facing': 0}] + body_pose_parts('stand'),
        'sit': [{'part': '$palm', 'pos': {'x': 300000, 'y': 0, 'z': 1400000}, 'facing': 0}] + body_pose_parts('sit'),
        'jolt_drop': [
            {'part': '$body_thigh_r', 'pos': {'x': 0, 'y': 0, 'z': 0}, 'facing': 0,
             'projection': {'origin': {'x': 30000, 'y': -100000, 'z': 550000},
                            'axis_u': {'x': 1000000, 'y': 0, 'z': 0},
                            'axis_v': {'x': 0, 'y': 1000000, 'z': 0}}},
        ],
    }
    return {
        'name': 'body_floor_to_sofa',
        'seed': 192,
        'ticks': 1200,
        'notes': (
            '19.6.3: ruta completa piso -> pie izquierdo -> pantorrilla -> muslo -> sofá al sentarse; '
            'y un segundo tiny sin agarre cae desde el muslo derecho y aterriza en el sofá.'
        ),
        'entities': [
            {'ref': 'house', 'spawn': {'name': 'house', 'space': {'kind': 'static', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 0, 'y': 0}}}}},
            {'ref': 'living', 'spawn': {
                'name': 'living',
                'space': {'kind': 'static', 'extent': _rect(300, 300)},
                'location': {'space': '$house', 'pos': {'x': 0, 'y': 0, 'z': 0}},
                'components': {'nav_grid': _grid(300, 300, cell_cm=10), 'contact_surface': {'tolerance': 40000}},
            }},
            {'ref': 'sofa', 'spawn': {
                'name': 'sofá',
                'space': {'kind': 'elevated', 'extent': _rect(140, 70)},
                'location': {'space': '$living', 'pos': {'x': 1000000, 'y': 400000, 'z': 450000}},
                'components': {'nav_grid': _grid(140, 70, cell_cm=5), 'contact_surface': {'tolerance': 60000}},
            }},
            {'ref': 'giantess', 'spawn': {
                'name': 'giantess',
                'space': {'kind': 'body', 'extent': {'min': {'x': 0, 'y': 0}, 'max': {'x': 0, 'y': 0}}},
                'location': {'space': '$living', 'pos': {'x': 1500000, 'y': 1500000, 'z': 0}},
                'components': {'mover': {'speed': 1000000, 'scale': 'giantess'}},
            }},
            palm_ent,
            *body_entities(),
            {'ref': 'tiny1', 'spawn': {
                'name': 'Tiny 01',
                'location': {'space': '$living', 'pos': {'x': 1300000, 'y': 1500000, 'z': 0}},
                'body': {'radius': 2000, 'height': 8000, 'mass': 1},
                'components': {'mover': {'speed': 150000, 'scale': 'tiny'}},
            }},
            # Segundo tiny en el muslo derecho, lejos del agarre, para probar caída al sofá.
            {'ref': 'tiny2', 'spawn': {
                'name': 'Tiny 02',
                'location': {'space': '$body_thigh_r', 'pos': {'x': 120000, 'y': 300000, 'z': 0}},
                'body': {'radius': 2000, 'height': 8000, 'mass': 1},
            }},
        ],
        'commands': [
            {'tick': 0, 'kind': 'set_components', 'data': {
                'entity': '$giantess',
                'components': {'poses': poses, 'body_joints': body_joints_component(), 'body_state': {'pose': 'idle'}},
            }},
            # 1. Tiny 01 sube al pie izquierdo plantado y trepa hasta el muslo izquierdo.
            {'tick': 10, 'kind': 'move_to', 'data': {'entity': '$tiny1', 'space': '$body_thigh_l', 'goal': {'x': 70000, 'y': 160000}}},
            # 2. Giantess camina al sofá y se sienta (muslo izquierdo contacta con el sofá).
            {'tick': 450, 'kind': 'move_to', 'data': {'entity': '$giantess', 'space': '$living', 'goal': {'x': 1500000, 'y': 650000}}},
            {'tick': 600, 'kind': 'set_pose', 'data': {'entity': '$giantess', 'pose': 'sit'}},
            # 3. Tiny 01 desembarca del muslo al sofá.
            {'tick': 620, 'kind': 'move_to', 'data': {'entity': '$tiny1', 'space': '$sofa', 'goal': {'x': 200000, 'y': 300000}}},
            # 4. Sacudida en el muslo derecho: Tiny 02 (sin agarre) pierde pie y cae sobre el sofá.
            {'tick': 700, 'kind': 'set_pose', 'data': {'entity': '$giantess', 'pose': 'jolt_drop'}},
        ],
        'expect': COMMON + [
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$body_foot_l'}},
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$body_calf_l'}},
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$body_thigh_l'}},
            {'event': 'entity_reparented', 'match': {'entity': '$tiny1', 'to': '$sofa'}},
            {'event': 'entity_lost_grip', 'match': {'entity': '$tiny2', 'space': '$body_thigh_r'}},
            {'event': 'entity_fell', 'match': {'entity': '$tiny2', 'to': '$sofa'}},
            {'at': '$tiny1', 'space': '$sofa', 'pos': {'x': 200000, 'y': 300000}, 'tol': 30000},
        ],
    }


def write(path, d):
    with open(path, 'w', encoding='utf-8', newline='\n') as f:
        f.write(json.dumps(d, ensure_ascii=False, indent=2) + '\n')
    print('escrito', os.path.relpath(path, ROOT))


def main():
    casa = json.load(open(CASA, encoding='utf-8'))
    with_body_map(casa)
    write(CASA, casa)
    write(os.path.join(SCEN, 'body_desk_to_pillow.json'), desk_to_pillow_scenario())
    write(os.path.join(SCEN, 'body_floor_to_sofa.json'), floor_to_sofa_scenario())


if __name__ == '__main__':
    main()
