"""Pisos de la casa vistos de cerca (micro) y el pie de la giantess.

La vista micro dibuja 1,6 px por mm. Cada piso es una baldosa que se repite
sobre la habitación (variante "tile" del atlas):

- dormitorio: una esterilla de tatami de 90 cm (1440 px) con su borde (heri);
- living: tablas de madera, 60 × 30 cm (960 × 480 px);
- cocina: baldosas en damero de 7,5 cm, 30 × 30 cm (480 × 480 px).

El pie (24 × 9 cm) va en un cuadro de 384 × 384 px, centrado, en las cuatro
direcciones: apoyado (foot_<dir>) y su sombra opaca (shadow_<dir>; la
presentación le da la transparencia).
"""
import math
import os

from PIL import Image

from px import Canvas, hexc

SRC = None  # lo fija main
INK = '#3F365C'


def noise(x, y, seed=0):
    """Ruido entero determinista en [0, 1)."""
    n = (x * 374761393 + y * 668265263 + seed * 2147483647) & 0xFFFFFFFF
    n = ((n ^ (n >> 13)) * 1274126177) & 0xFFFFFFFF
    return ((n ^ (n >> 16)) & 0xFFFF) / 65536


def tatami():
    size, heri = 1440, 40
    base, light, dark, deep = hexc('#C8C98F'), hexc('#D3D49E'), hexc('#B7B97F'), hexc('#A9AB72')
    im = Image.new('RGBA', (size, size), base)
    px = im.load()
    # trama: hileras horizontales de 8 px, con puntadas que alternan
    for y in range(size):
        row, k = y // 8, y % 8
        for x in range(size):
            c = base
            if k == 0:
                c = dark
            elif k == 1:
                c = light
            elif (x // 6 + row) % 2 == 0 and k in (3, 4):
                c = light
            if noise(x // 3, y // 2, 1) < 0.06:
                c = deep if c is not light else base
            px[x, y] = c
    c = Canvas(size, size)
    c.im = im
    # heri: borde de tela a los dos lados largos (arriba y abajo de la baldosa)
    navy, navy_l, navy_d = '#5D4F7A', '#76699A', '#463B60'
    for y0 in (0, size - heri):
        c.rect(0, y0, size, heri, navy)
        c.hline(0, y0 + 3, size, navy_l)
        c.hline(0, y0 + heri - 4, size, navy_d)
        for x in range(0, size, 24):  # dibujo del heri
            c.rect(x + 8, y0 + 16, 8, 8, navy_l)
    # junta con la esterilla vecina
    c.vline(0, 0, size, '#8F9061')
    c.vline(size - 1, 0, size, '#E1E2B4')
    c.save(os.path.join(SRC, 'bedroom/micro_tile.png'))


def planks():
    w, h, plank = 960, 480, 120
    tones = ['#D2A685', '#D9B08E', '#CFA07E', '#DDB694']
    c = Canvas(w, h)
    for i in range(h // plank):
        y0 = i * plank
        seam = (i * 360) % w  # juntas escalonadas
        tone = hexc(tones[i % len(tones)])
        for y in range(y0, y0 + plank):
            for x in range(w):
                grain = math.sin((x + i * 97) / 37.0 + math.sin(y / 9.0) * 1.7)
                col = tone
                if grain > 0.85:
                    col = hexc('#BF9076')
                elif grain < -0.9:
                    col = hexc('#E6C4A2')
                c.im.putpixel((x, y), col)
        c.hline(0, y0, w, '#9E6E5C')
        c.hline(0, y0 + 1, w, '#E2B896')
        c.vline(seam, y0, plank, '#9E6E5C')
        c.vline((seam + 1) % w, y0, plank, '#E2B896')
    c.save(os.path.join(SRC, 'living/micro_tile.png'))


def checker():
    size, sq = 480, 120
    c = Canvas(size, size)
    for j in range(size // sq):
        for i in range(size // sq):
            fill = '#CDE0D2' if (i + j) % 2 == 0 else '#E6E4CC'
            c.rect(i * sq, j * sq, sq, sq, fill)
            c.hline(i * sq, j * sq, sq, '#B9CBBE')
            c.vline(i * sq, j * sq, sq, '#B9CBBE')
            c.hline(i * sq + 2, j * sq + 2, sq - 4, '#F4F3E2')
    c.save(os.path.join(SRC, 'kitchen/micro_tile.png'))


FOOT = 384
TURNS = {'east': 0, 'north': 90, 'west': 180, 'south': 270}


def foot_shape(skin=True):
    """Pie derecho visto desde arriba mirando al este: talón a la izquierda,
    dedos a la derecha. Es una unión de elipses: talón, arco, planta y dedos."""
    c = Canvas(FOOT, FOOT)
    cy = FOOT // 2
    body = [(74, cy + 14, 48, 42), (150, cy + 20, 64, 38), (236, cy + 2, 70, 64)]
    toes = [(312, cy - 36, 30, 26), (328, cy - 2, 21, 18), (326, cy + 26, 18, 16), (316, cy + 48, 15, 13),
            (300, cy + 64, 12, 11)]

    def ell(x, y, e):
        ex, ey, rx, ry = e
        return ((x - ex) / rx) ** 2 + ((y - ey) / ry) ** 2 <= 1

    shape = {}
    for y in range(FOOT):
        for x in range(FOOT):
            if any(ell(x, y, e) for e in body):
                shape[x, y] = 'b'
            for i, t in enumerate(toes):
                if ell(x, y, t):
                    shape[x, y] = i
    if not skin:
        for x, y in shape:
            c.im.putpixel((x, y), hexc(INK))
        return c.im
    skin_c, shade, light, nail = hexc('#FFD8B9'), hexc('#F2B9A0'), hexc('#FFE9DA'), hexc('#FFF1E8')
    for (x, y), part in shape.items():
        col = skin_c
        if part == 'b':
            if (x, y + 12) not in shape:
                col = shade  # lado de afuera, en sombra
            elif (x, y - 8) not in shape and 100 < x < 290:
                col = light  # empeine iluminado
        c.im.putpixel((x, y), col)
    ink = hexc(INK)
    for (x, y), part in shape.items():
        edge = any(shape.get((x + dx, y + dy)) is None for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)))
        # surco entre cada dedo y lo demás
        seam = isinstance(part, int) and any(
            shape.get((x + dx, y + dy)) not in (None, part) for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)))
        if edge or seam:
            c.im.putpixel((x, y), ink)
    for tx, ty, rx, ry in toes:  # uñas
        c.ellipse(tx + rx // 2 - 2, ty, max(rx // 3, 3), max(ry // 3, 3), nail)
    return c.im


def foot():
    sole, sil = foot_shape(True), foot_shape(False)
    for facing, turn in TURNS.items():
        for name, img in (('foot', sole), ('shadow', sil)):
            out = img.rotate(turn) if turn else img
            out.save(os.path.join(SRC, f'foot/micro_{name}_{facing}.png'), optimize=True)


def main(src):
    global SRC
    SRC = src
    for d in ('bedroom', 'living', 'kitchen', 'foot'):
        os.makedirs(os.path.join(src, d), exist_ok=True)
    tatami()
    planks()
    checker()
    foot()
