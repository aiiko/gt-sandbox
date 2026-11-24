"""Arquitectura del departamento (propuesta 20) en macro (80 px/m) y baldosas micro (1600 px/m).

Importa medidas, ventanas y vanos directamente desde tools/scenarios/planta.py
para mantener una fuente única de geometría (20.3.1).
"""
import importlib.util
import math
import os

from PIL import Image

HERE = os.path.dirname(os.path.abspath(__file__))
_planta_path = os.path.join(os.path.dirname(HERE), 'scenarios', 'planta.py')
_spec = importlib.util.spec_from_file_location('planta', _planta_path)
planta = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(planta)

from px import Canvas, hexc  # noqa: E402
from shapes import Sprite  # noqa: E402

SRC = None

INK = '#3F365C'
INK_SOFT = (0x3F, 0x36, 0x5C, 70)
WOOD, WOOD_D, WOOD_L = '#C89474', '#9E6E5C', '#E2B896'
PAPER, PAPER_L = '#F7EEDD', '#DDCBB4'

NORTH_WALL_H = 32
SOUTH_WALL_H = 10
PART_T = 6  # espesor de un tabique interior compartido


def cm_to_px(cm):
    return int(round(cm * 0.8))


def save(s, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    s.save(path)


def noise(x, y, seed=0):
    n = (x * 374761393 + y * 668265263 + seed * 1442695041) & 0xFFFFFFFF
    n = ((n ^ (n >> 13)) * 1274126177) & 0xFFFFFFFF
    return (n ^ (n >> 16)) & 0xFF


# ------------------------------------------------------------ texturas de piso

def tatami_mat(s, x, y, w, h, horizontal):
    base, light, dark = '#C8C98F', '#D6D7A2', '#B4B67C'
    for j in range(h):
        for i in range(w):
            k = j if horizontal else i
            m = k % 4
            col = light if m == 0 else dark if m == 3 else base
            if noise(x + i, y + j, 7) < 10:
                col = dark
            s.px(x + i, y + j, col)
    edge, edge_hi = '#5D4F7A', '#7C6E9C'
    if horizontal:
        s.rect(x, y, x + w - 1, y + 1, edge)
        s.hline(x, x + w - 1, y, edge_hi)
        s.rect(x, y + h - 2, x + w - 1, y + h - 1, edge)
        s.vline(x, y, y + h - 1, '#A9AB72')
    else:
        s.rect(x, y, x + 1, y + h - 1, edge)
        s.vline(x, y, y + h - 1, edge_hi)
        s.rect(x + w - 2, y, x + w - 1, y + h - 1, edge)
        s.hline(x, x + w - 1, y, '#A9AB72')


def plank_floor(s, w, y0, y1, colors=('#D9B08E', '#D2A685', '#DDB694'), seam='#BF9076', grain='#C99C7C'):
    row = 0
    for y in range(y0, y1, 12):
        col = colors[row % len(colors)]
        s.rect(0, y, w - 1, min(y + 11, y1 - 1), col)
        s.hline(0, w - 1, y, seam)
        off = (row * 74) % 120
        for x in range(off, w, 120):
            s.vline(x, y, min(y + 11, y1 - 1), seam)
            if x + 1 < w:
                s.vline(x + 1, y + 1, min(y + 11, y1 - 1), colors[(row + 1) % len(colors)])
        for k in range(max(3, w // 80)):
            gx = (row * 53 + k * 81) % w
            gy = y + 3 + (k * 3) % 7
            if gy < y1:
                s.hline(gx, min(gx + 18 + (k % 3) * 6, w - 1), gy, grain)
        row += 1


def checker_floor(s, w, y0, y1, a, b, grout, tile=16):
    for y in range(y0, y1):
        for x in range(w):
            light = ((x // tile) + ((y - y0) // tile)) % 2 == 0
            col = a if light else b
            if x % tile == 0 or (y - y0) % tile == 0:
                col = grout
            s.px(x, y, col)


def slate_floor(s, w, y0, y1):
    """Baldosas de piedra gris cálido para el genkan."""
    tile = 20
    for y in range(y0, y1):
        for x in range(w):
            tx, ty = x // tile, (y - y0) // tile
            tone = '#B9B4BC' if (tx + ty) % 2 == 0 else '#C3BEC5'
            if noise(x, y, 19) < 14:
                tone = '#A9A3AE'
            if x % tile == 0 or (y - y0) % tile == 0:
                tone = '#8E8894'
            s.px(x, y, tone)


# ------------------------------------------------------------ muros y ventanas

def north_wall(s, w, tint):
    wall, wall_hi, wall_d, base, base_d = tint
    s.rect(0, 0, w - 1, 4, INK)
    s.hline(0, w - 1, 5, wall_hi)
    s.hline(0, w - 1, 6, wall_hi)
    s.rect(0, 7, w - 1, 27, wall)
    for y in range(8, 27):
        for x in range(w):
            if noise(x, y, 3) < 6:
                s.px(x, y, wall_d)
    s.hline(0, w - 1, 8, wall_hi)
    s.rect(0, 28, w - 1, 31, base)
    s.hline(0, w - 1, 31, base_d)
    for x in (6, w - 10):
        s.rect(x, 5, x + 3, 27, WOOD)
        s.vline(x, 5, 27, WOOD_L)
        s.vline(x + 3, 5, 27, WOOD_D)


def shoji_window(s, x, w):
    s.rect(x, 6, x + w - 1, 27, WOOD_D)
    s.rect(x + 2, 8, x + w - 3, 25, PAPER)
    for i in range(x + 9, x + w - 3, 8):
        s.vline(i, 8, 25, PAPER_L)
    for y in (13, 19):
        s.hline(x + 2, x + w - 3, y, PAPER_L)
    s.hline(x, x + w - 1, 6, WOOD)
    s.hline(x - 2, x + w + 1, 26, WOOD)
    s.hline(x - 2, x + w + 1, 27, WOOD_D)
    for y in range(9, 13):
        s.hline(x + 3, x + 6, y, '#FFFBF2')


def glass_window(s, x, w):
    s.rect(x, 5, x + w - 1, 26, '#D5E4E9')
    s.rect(x + 2, 7, x + w - 3, 24, '#B9E5EF')
    s.vline(x + w // 2, 7, 24, '#D5E4E9')
    s.vline(x + w // 2 + 1, 7, 24, '#8DB7C6')
    for i in range(6):
        s.px(x + 5 + i, 18 - i, '#E9F8FB')
        s.px(x + 6 + i, 18 - i, '#E9F8FB')
    s.hline(x - 2, x + w + 1, 25, '#F2E4E4')
    s.hline(x - 2, x + w + 1, 26, '#C9AEB5')
    s.hline(x + 2, x + w - 3, 7, '#8DB7C6')


def light_shaft(s, x, w, y0, y1, alpha=32):
    for y in range(y0, y1):
        off = (y - y0) // 2
        for i in range(w):
            px_x = x + off + i
            if 0 <= px_x < s.w:
                s.blend(px_x, y, (0xFF, 0xF6, 0xDE, alpha))


def draw_ew_partition(s, w, y0, gaps_px):
    """Un solo tabique horizontal en el borde superior de un cuarto interior."""
    s.rect(0, y0, w - 1, y0 + PART_T - 1, '#9C7D93')
    s.hline(0, w - 1, y0, '#E2BCC3')
    s.hline(0, w - 1, y0 + 1, '#C7A1B2')
    s.hline(0, w - 1, y0 + PART_T - 1, INK)
    for g0, g1 in gaps_px:
        # Dejar el piso visible en el vano y dibujar jambas de madera.
        s.rect(max(0, g0 - 2), y0, g0 - 1, y0 + PART_T - 1, WOOD)
        s.rect(g1, y0, min(w - 1, g1 + 1), y0 + PART_T - 1, WOOD)


def draw_south_exterior_wall(s, w, y0, door_span_px=None):
    """Muro exterior sur con remate limpio."""
    s.rect(0, y0, w - 1, y0 + 2, '#E7C8CD')
    s.hline(0, w - 1, y0, '#F6E1E3')
    s.rect(0, y0 + 3, w - 1, y0 + SOUTH_WALL_H - 2, '#A68AA2')
    for x in range(0, w, 30):
        s.vline(x, y0 + 3, y0 + SOUTH_WALL_H - 2, '#957A92')
    s.hline(0, w - 1, y0 + SOUTH_WALL_H - 1, INK)
    if door_span_px:
        x0, x1 = door_span_px
        s.rect(x0, y0, x1, y0 + SOUTH_WALL_H - 1, '#584B73')
        s.hline(x0 + 2, x1 - 2, y0 + 1, '#7D6E9E')
        s.rect(x0 + 2, y0 + 3, x1 - 2, y0 + SOUTH_WALL_H - 2, '#453A5E')
        s.rect(x1 - 8, y0 + 4, x1 - 5, y0 + 6, '#FFCA8A')  # picaporte


def side_wall_fg(s, w, h, y0, y1, west_mode, east_mode, south_y=None, south_door=None, east_window=None):
    """Primer plano de muros laterales y sur (un solo tabique por límite compartido).

    mode: 'outer' (6 px exterior), 'shared' (4 px medio tabique), None.
    """
    outer_cols = ['#E2BCC3', '#C7A1B2', '#AA899E', '#8C7189', '#776681', '#5D4E6B']
    shared_cols = ['#D5B0BA', '#AA899E', '#8C7189', '#6B5B76']
    if west_mode == 'outer':
        for i, col in enumerate(outer_cols):
            s.vline(i, y0, y1, col)
    elif west_mode == 'shared':
        for i, col in enumerate(shared_cols):
            s.vline(i, y0, y1, col)
    if east_mode == 'outer':
        for i, col in enumerate(reversed(outer_cols)):
            s.vline(w - len(outer_cols) + i, y0, y1, col)
        if east_window:
            wy0, wy1 = east_window
            s.rect(w - 6, wy0, w - 1, wy1, '#B9E5EF')
            s.vline(w - 6, wy0, wy1, '#D5E4E9')
            s.vline(w - 1, wy0, wy1, '#6E8E9E')
            s.hline(w - 6, w - 1, wy0, '#8DB7C6')
            s.hline(w - 6, w - 1, wy1, '#8DB7C6')
    elif east_mode == 'shared':
        for i, col in enumerate(reversed(shared_cols)):
            s.vline(w - len(shared_cols) + i, y0, y1, col)
    if south_y is not None:
        draw_south_exterior_wall(s, w, south_y, south_door)


def _room_ew_gaps(room_key, wall_y_cm):
    """Devuelve los vanos (x0_px, x1_px) locales en el borde superior de room_key."""
    rx = planta.ROOMS[room_key]['origin_cm'][0]
    gaps = []
    for dw in planta.DOORWAYS:
        if dw['axis'] == 'horizontal' and dw['y_cm'] == wall_y_cm:
            if dw['room_a'] == room_key or dw['room_b'] == room_key:
                g0 = cm_to_px(dw['span_cm'][0] - rx)
                g1 = cm_to_px(dw['span_cm'][1] - rx)
                gaps.append((g0, g1))
    return gaps


# ------------------------------------------------------------ 6 habitaciones

def apt_bedroom():
    r = planta.ROOMS['bedroom']
    w, fh = cm_to_px(r['size_cm'][0]), cm_to_px(r['size_cm'][1])
    h = NORTH_WALL_H + fh + 2
    s = Sprite(w, h)
    # Esterillas de tatami sobre 256 × 264 px (fila superior, medias, inferior).
    mats = [
        (0, 32, 128, 66, True), (128, 32, 128, 66, True),
        (0, 98, 64, 132, False), (192, 98, 64, 132, False),
        (64, 98, 128, 66, True), (64, 164, 128, 66, True),
        (0, 230, 128, 66, True), (128, 230, 128, 66, True),
    ]
    for x, y, mw, mh, hor in mats:
        tatami_mat(s, x, y, mw, mh, hor)
    north_wall(s, w, r['wall_tint'])
    for win in r['windows']:
        wx, ww = cm_to_px(win['pos_cm']), cm_to_px(win['width_cm'])
        shoji_window(s, wx, ww)
        light_shaft(s, wx + 4, ww - 8, NORTH_WALL_H, NORTH_WALL_H + 84, 32)
    # Kakejiku con rama de ciruelo sobre el futón
    s.hline(64, 86, 7, WOOD_D)
    s.rect(66, 8, 84, 26, '#7C6A95')
    s.rect(68, 10, 82, 24, PAPER)
    s.line([(71, 22), (75, 16), (79, 13)], '#6E4E46')
    s.line([(75, 16), (77, 19)], '#6E4E46')
    for px_, py_ in ((79, 12), (72, 20), (77, 19)):
        s.px(px_, py_, '#FF9CAD')
    save(s, 'apt_bedroom/macro_idle.png')

    fg = Sprite(w, h)
    side_wall_fg(fg, w, h, 5, NORTH_WALL_H + fh - 1, 'outer', 'shared')
    save(fg, 'apt_bedroom/macro_fg.png')


def apt_living():
    r = planta.ROOMS['living']
    w, fh = cm_to_px(r['size_cm'][0]), cm_to_px(r['size_cm'][1])
    h = NORTH_WALL_H + fh + 2
    s = Sprite(w, h)
    plank_floor(s, w, NORTH_WALL_H, NORTH_WALL_H + fh)
    north_wall(s, w, r['wall_tint'])
    for win in r['windows']:
        wx, ww = cm_to_px(win['pos_cm']), cm_to_px(win['width_cm'])
        glass_window(s, wx, ww)
        light_shaft(s, wx + 4, ww - 8, NORTH_WALL_H, NORTH_WALL_H + 80, 28)
    save(s, 'apt_living/macro_idle.png')

    fg = Sprite(w, h)
    side_wall_fg(fg, w, h, 5, NORTH_WALL_H + fh - 1, 'shared', 'outer')
    save(fg, 'apt_living/macro_fg.png')


def apt_hallway():
    r = planta.ROOMS['hallway']
    w, fh = cm_to_px(r['size_cm'][0]), cm_to_px(r['size_cm'][1])
    h = fh + 2
    s = Sprite(w, h)
    plank_floor(s, w, 0, fh, colors=('#D6AD8B', '#CFA382', '#DAB290'))
    # Tabique compartido norte (y = 310 cm) con vanos a dormitorio y living.
    gaps = _room_ew_gaps('hallway', 310)
    # Dibujar el tabique solo fuera de los vanos.
    s.rect(0, 0, w - 1, PART_T - 1, '#9C7D93')
    s.hline(0, w - 1, 0, '#E2BCC3')
    s.hline(0, w - 1, 1, '#C7A1B2')
    s.hline(0, w - 1, PART_T - 1, INK)
    for g0, g1 in gaps:
        # Reponer piso de madera en el vano.
        s.rect(g0, 0, g1 - 1, PART_T - 1, '#D6AD8B')
        s.hline(g0, g1 - 1, 0, '#BF9076')
        s.rect(max(0, g0 - 2), 0, g0 - 1, PART_T - 1, WOOD)
        s.rect(g1, 0, min(w - 1, g1 + 1), PART_T - 1, WOOD)
    save(s, 'apt_hallway/macro_idle.png')

    fg = Sprite(w, h)
    side_wall_fg(fg, w, h, 0, fh - 1, 'outer', 'shared')
    save(fg, 'apt_hallway/macro_fg.png')


def apt_kitchen():
    r = planta.ROOMS['kitchen']
    w, fh = cm_to_px(r['size_cm'][0]), cm_to_px(r['size_cm'][1])
    h = fh + SOUTH_WALL_H
    s = Sprite(w, h)
    checker_floor(s, w, 0, fh, '#E6E4CC', '#CDE0D2', '#D7DCCB', tile=16)
    # Tabique compartido norte (y = 310 cm) con vano del noren al living.
    gaps = _room_ew_gaps('kitchen', 310)
    s.rect(0, 0, w - 1, PART_T - 1, '#9C7D93')
    s.hline(0, w - 1, 0, '#E2BCC3')
    s.hline(0, w - 1, 1, '#C7A1B2')
    s.hline(0, w - 1, PART_T - 1, INK)
    for g0, g1 in gaps:
        s.rect(g0, 0, g1 - 1, PART_T - 1, '#E6E4CC')
        s.rect(max(0, g0 - 2), 0, g0 - 1, PART_T - 1, WOOD)
        s.rect(g1, 0, min(w - 1, g1 + 1), PART_T - 1, WOOD)
    # Mancha de sol desde la ventana este
    east_win = r['windows'][0]
    wy1 = fh - cm_to_px(east_win['pos_cm'])
    wy0 = wy1 - cm_to_px(east_win['width_cm'])
    for y in range(wy0, wy1):
        for i in range(64):
            px_x = w - 8 - i - (y - wy0) // 4
            if 0 <= px_x < w:
                s.blend(px_x, y, (0xFF, 0xF6, 0xDE, 26))
    save(s, 'apt_kitchen/macro_idle.png')

    fg = Sprite(w, h)
    side_wall_fg(fg, w, h, 0, fh - 1, 'shared', 'outer', south_y=fh, east_window=(wy0, wy1))
    save(fg, 'apt_kitchen/macro_fg.png')


def apt_bathroom():
    r = planta.ROOMS['bathroom']
    w, fh = cm_to_px(r['size_cm'][0]), cm_to_px(r['size_cm'][1])
    h = fh + SOUTH_WALL_H
    s = Sprite(w, h)
    checker_floor(s, w, 0, fh, '#EAF4F6', '#DCECEF', '#C2D9DF', tile=14)
    # Tabique compartido norte (y = 200 cm) con puerta del baño.
    gaps = _room_ew_gaps('bathroom', 200)
    s.rect(0, 0, w - 1, PART_T - 1, '#9C7D93')
    s.hline(0, w - 1, 0, '#E2BCC3')
    s.hline(0, w - 1, 1, '#C7A1B2')
    s.hline(0, w - 1, PART_T - 1, INK)
    for g0, g1 in gaps:
        s.rect(g0, 0, g1 - 1, PART_T - 1, '#EAF4F6')
        s.rect(max(0, g0 - 2), 0, g0 - 1, PART_T - 1, WOOD)
        s.rect(g1, 0, min(w - 1, g1 + 1), PART_T - 1, WOOD)
    save(s, 'apt_bathroom/macro_idle.png')

    fg = Sprite(w, h)
    side_wall_fg(fg, w, h, 0, fh - 1, 'shared', 'shared', south_y=fh)
    save(fg, 'apt_bathroom/macro_fg.png')


def apt_genkan():
    r = planta.ROOMS['genkan']
    w, fh = cm_to_px(r['size_cm'][0]), cm_to_px(r['size_cm'][1])
    h = fh + SOUTH_WALL_H
    s = Sprite(w, h)
    slate_floor(s, w, 0, fh)
    # Agarikamachi: escalón de madera de 8 cm en el borde norte hacia el distribuidor.
    s.rect(0, 0, w - 1, 6, WOOD)
    s.hline(0, w - 1, 0, WOOD_L)
    s.hline(0, w - 1, 5, WOOD_D)
    s.hline(0, w - 1, 6, INK)
    s.rect(0, 7, w - 1, 10, INK_SOFT)
    # Felpudo junto a la puerta exterior
    d0, d1 = cm_to_px(planta.EXTERIOR_DOOR_SPAN_CM[0]), cm_to_px(planta.EXTERIOR_DOOR_SPAN_CM[1])
    s.rrect(d0 + 6, fh - 30, d1 - 6, fh - 8, '#A98A78', '#6E564C', r=2)
    s.rrect(d0 + 9, fh - 27, d1 - 9, fh - 11, None, '#C8AD9C', r=1)
    save(s, 'apt_genkan/macro_idle.png')

    fg = Sprite(w, h)
    side_wall_fg(fg, w, h, 0, fh - 1, 'outer', 'shared', south_y=fh, south_door=(d0, d1))
    save(fg, 'apt_genkan/macro_fg.png')


# ------------------------------------------------------------ puertas horizontales

def door_h():
    """Fusuma horizontal entre distribuidor y dormitorio (72 × 16, pivote [36, 8])."""
    s = Sprite(72, 16)
    s.rect(0, 2, 71, 13, WOOD_D)
    s.hline(0, 71, 2, WOOD_L)
    s.rect(2, 4, 34, 11, PAPER)
    s.rect(37, 4, 69, 11, PAPER)
    # ola lila sutil sobre el papel
    for x in range(3, 34):
        s.px(x, 8 + round(math.sin(x / 3.0)), '#B9ADEB')
    for x in range(38, 69):
        s.px(x, 8 + round(math.sin(x / 3.0)), '#B9ADEB')
    s.rect(31, 6, 33, 9, '#7C6A95')
    s.rect(38, 6, 40, 9, '#7C6A95')
    s.hline(0, 71, 14, INK_SOFT)
    save(s, 'door_h/macro_idle.png')


def door_noren_h():
    """Cortina noren horizontal entre living y cocina (72 × 20, pivote [36, 10])."""
    s = Sprite(72, 20)
    s.rect(0, 1, 71, 3, WOOD_D)
    s.hline(0, 71, 1, WOOD_L)
    for x0 in (4, 26, 48):
        s.rect(x0, 4, x0 + 19, 15, '#6E64B8')
        s.hline(x0, x0 + 19, 4, '#5A5199')
        s.hline(x0, x0 + 19, 15, '#5A5199')
        s.vline(x0 + 2, 5, 14, '#7F75C9')
    # flor blanca en el paño central
    for dx, dy in ((0, -2), (2, 0), (0, 2), (-2, 0)):
        s.px(36 + dx, 9 + dy, '#FFF1D8')
    s.px(36, 9, '#FF9CAD')
    s.rect(4, 16, 67, 19, INK_SOFT)
    save(s, 'door_noren_h/macro_idle.png')


def door_bath():
    """Umbral y hoja entreabierta del baño (64 × 14, pivote [32, 7])."""
    s = Sprite(64, 14)
    s.rect(0, 3, 63, 10, '#C8D6DC')
    s.hline(0, 63, 3, '#EAF4F6')
    s.hline(0, 63, 10, '#8DA6B2')
    s.rect(4, 4, 24, 9, '#F2ECE4')
    s.hline(4, 24, 4, '#FFFFFF')
    s.rect(20, 6, 22, 8, '#7C8F9B')
    save(s, 'door_bath/macro_idle.png')


# ------------------------------------------------------------ baldosas micro nuevas

def micro_bath_tile():
    size, sq = 320, 80
    c = Canvas(size, size)
    for j in range(size // sq):
        for i in range(size // sq):
            fill = '#EAF4F6' if (i + j) % 2 == 0 else '#DCECEF'
            c.rect(i * sq, j * sq, sq, sq, fill)
            c.hline(i * sq, j * sq, sq, '#BBD2D9')
            c.vline(i * sq, j * sq, sq, '#BBD2D9')
            c.hline(i * sq + 2, j * sq + 2, sq - 4, '#F7FCFD')
    save(c, 'apt_bathroom/micro_tile.png')


def micro_genkan_tile():
    size, sq = 320, 160
    c = Canvas(size, size)
    px = c.im.load()
    for j in range(size // sq):
        for i in range(size // sq):
            base = hexc('#B9B4BC' if (i + j) % 2 == 0 else '#C3BEC5')
            speck = hexc('#A59FAA')
            for y in range(j * sq, (j + 1) * sq):
                for x in range(i * sq, (i + 1) * sq):
                    px[x, y] = speck if noise(x // 2, y // 2, 23) < 18 else base
            c.hline(i * sq, j * sq, sq, '#8E8894')
            c.vline(i * sq, j * sq, sq, '#8E8894')
    save(c, 'apt_genkan/micro_tile.png')


def micro_counter_tile():
    """Superficie de mesada de piedra clara (160 × 160 px = 10 × 10 cm)."""
    size = 160
    c = Canvas(size, size)
    px = c.im.load()
    base, light, dark = hexc('#EFF3EE'), hexc('#F8FAF7'), hexc('#DCE3DC')
    for y in range(size):
        for x in range(size):
            n = noise(x // 2, y // 2, 31)
            if n < 14:
                px[x, y] = dark
            elif n > 240:
                px[x, y] = light
            else:
                px[x, y] = base
    c.hline(0, 0, size, '#E2E8E2')
    save(c, 'counter/micro_tile.png')


def micro_sink_basin_tile():
    """Acero inoxidable cepillado del fondo de la bacha (160 × 160 px)."""
    size = 160
    c = Canvas(size, size)
    px = c.im.load()
    base, light, dark = hexc('#B8CDD2'), hexc('#CEE0E4'), hexc('#9FB8BF')
    for y in range(size):
        tone = light if y % 6 == 0 else dark if y % 6 == 3 else base
        for x in range(size):
            px[x, y] = dark if noise(x // 4, y, 41) < 12 else tone
    save(c, 'sink_basin/micro_tile.png')


def main(src):
    global SRC
    SRC = src
    apt_bedroom()
    apt_living()
    apt_hallway()
    apt_kitchen()
    apt_bathroom()
    apt_genkan()
    door_h()
    door_noren_h()
    door_bath()
    micro_bath_tile()
    micro_genkan_tile()
    micro_counter_tile()
    micro_sink_basin_tile()
