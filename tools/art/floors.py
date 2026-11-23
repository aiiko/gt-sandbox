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

from PIL import Image, ImageDraw

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
    """Calzado violeta con media crema visto desde arriba mirando al este
    (talón a la izquierda, puntera redondeada a la derecha), 384 × 384 px."""
    cy = FOOT // 2
    ink = hexc(INK)
    shoe, shoe_l, shoe_hi, shoe_d, shoe_d2 = (
        hexc('#76659B'), hexc('#9A8AC2'), hexc('#B8AAD9'), hexc('#554675'), hexc('#3D3157')
    )
    sock, sock_d = hexc('#FFF1D8'), hexc('#DCCFE0')
    buckle, white = hexc('#FFCA8A'), hexc('#FFFFFF')

    im = Image.new('RGBA', (FOOT, FOOT), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    # Silueta del calzado mirando al este: talón a la izquierda, puntera redondeada a la derecha
    shoe_pts = [
        (42, cy - 42), (108, cy - 56), (206, cy - 58), (288, cy - 66),
        (338, cy - 46), (354, cy - 16), (354, cy + 20), (336, cy + 50),
        (284, cy + 68), (202, cy + 58), (108, cy + 60), (42, cy + 44),
    ]
    d.polygon(shoe_pts, fill=shoe)
    d.ellipse([236, cy - 68, 360, cy + 68], fill=shoe)
    d.ellipse([26, cy - 52, 128, cy + 56], fill=shoe)

    if not skin:
        px = im.load()
        for y in range(FOOT):
            for x in range(FOOT):
                if px[x, y][3] > 0:
                    px[x, y] = ink
        return im

    # Sombreado cel-shading (luz desde arriba a la izquierda)
    px = im.load()
    for y in range(FOOT):
        for x in range(FOOT):
            if px[x, y][3] == 0:
                continue
            v = (y - (cy - 68)) / 136.0
            if v < 0.20:
                px[x, y] = shoe_l
            elif v > 0.80:
                px[x, y] = shoe_d2
            elif v > 0.62:
                px[x, y] = shoe_d

    # Brillo lustroso en la puntera derecha y pespunte perimetral
    d.arc([240, cy - 62, 354, cy + 62], start=275, end=85, fill=shoe_hi, width=3)
    d.ellipse([282, cy - 40, 336, cy - 2], fill=shoe_hi)
    d.ellipse([296, cy - 30, 326, cy - 10], fill=white)

    # Abertura del zapato mostrando la media crema acanalada
    d.ellipse([66, cy - 42, 224, cy + 44], fill=sock, outline=ink, width=3)
    d.arc([70, cy - 38, 220, cy + 40], start=20, end=170, fill=sock_d, width=8)
    for y in range(cy - 30, cy + 34, 8):
        d.line([(82, y), (208, y)], fill=sock_d, width=2)

    # Correa Mary-Jane cruzando el empeine con hebilla dorada
    d.rounded_rectangle([196, cy - 62, 224, cy + 64], radius=6, fill=shoe, outline=ink, width=2)
    d.line([(202, cy - 56), (202, cy + 58)], fill=shoe_l, width=3)
    d.rounded_rectangle([200, cy - 58, 220, cy - 42], radius=3, fill=buckle, outline=ink, width=2)

    # Contorno exterior de 2 px
    src = im.load()
    out = im.copy()
    dst = out.load()
    for y in range(FOOT):
        for x in range(FOOT):
            if src[x, y][3] > 0:
                continue
            for dy in (-2, -1, 0, 1, 2):
                for dx in (-2, -1, 0, 1, 2):
                    if dx * dx + dy * dy <= 5:
                        nx, ny = x + dx, y + dy
                        if 0 <= nx < FOOT and 0 <= ny < FOOT and src[nx, ny][3] > 200:
                            dst[x, y] = ink
                            break
                else:
                    continue
                break
    return out


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
