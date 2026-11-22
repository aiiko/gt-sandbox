"""Arte de silla común, sofá y sombra al sentarse (M37).

- chair/macro_idle.png (20 × 24 px, pivote [2, 18]): silla de madera con cojín cálido.
- chair/micro_tile.png (160 × 160 px): tapizado de tela tramada del asiento (10 × 10 cm).
- sofa/macro_idle.png (56 × 30 px, pivote [4, 22]): sofá de tres lugares con apoyabrazos.
- sofa/micro_tile.png (160 × 160 px): tapizado mullido en tono salvia (10 × 10 cm).
- seat/micro_shadow_<dir>.png (384 × 384 px, pivote [192, 192]): sombra al sentarse.
"""
import math
import os

from px import Canvas, hexc

SRC = None  # lo fija main
INK = '#3F365C'


def save(c, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    c.save(path)


def chair_macro():
    c = Canvas(20, 24)
    wood_d, wood, wood_l = '#8A5A44', '#B57B5A', '#D19B76'
    seat_d, seat, seat_l = '#C48A6A', '#E2AA86', '#F2C5A4'
    # sombra en el piso
    c.ellipse(10, 21, 8, 2, hexc(INK, 65))
    # patas traseras y delanteras
    for x in (3, 15):
        c.rect(x, 16, 2, 6, wood_d)
        c.vline(x, 16, 5, wood)
    # asiento (16 × 12)
    c.rrect(2, 6, 16, 12, seat, INK)
    c.hline(4, 7, 12, seat_l)
    c.hline(3, 16, 14, seat_d)
    # respaldo (borde sur de la silla, mirando al norte hacia el escritorio)
    c.rrect(2, 14, 16, 5, wood, INK)
    c.hline(4, 15, 12, wood_l)
    c.hline(3, 17, 14, wood_d)
    save(c, 'chair/macro_idle.png')


def chair_micro_tile():
    size = 160
    base, light, dark = hexc('#E2AA86'), hexc('#ECC09E'), hexc('#CE9370')
    c = Canvas(size, size)
    px = c.im.load()
    for y in range(size):
        for x in range(size):
            # trama tejida suave
            if (x // 4 + y // 4) % 2 == 0:
                col = light if (x % 4 == 0 or y % 4 == 0) else base
            else:
                col = dark if (x % 4 == 3 or y % 4 == 3) else base
            px[x, y] = col
    # costura sutil
    for k in range(0, size, 40):
        c.hline(0, k, size, '#C08360')
        c.vline(k, 0, size, '#C08360')
    save(c, 'chair/micro_tile.png')


def sofa_macro():
    c = Canvas(56, 30)
    fab_d, fab, fab_l = '#5A8578', '#78A899', '#99C7B8'
    wood = '#8A5A44'
    # sombra
    c.ellipse(28, 27, 25, 2, hexc(INK, 65))
    # patas cortas
    for x in (5, 49):
        c.rect(x, 24, 2, 4, wood)
    # cuerpo del asiento (48 × 20)
    c.rrect(4, 4, 48, 20, fab, INK)
    c.hline(6, 5, 44, fab_l)
    # divisiones de los tres almohadones
    for x in (20, 36):
        c.vline(x, 5, 16, fab_d)
        c.vline(x + 1, 5, 16, fab_l)
    # apoyabrazos laterales
    c.rrect(1, 5, 5, 19, fab_d, INK)
    c.vline(2, 7, 15, fab)
    c.rrect(50, 5, 5, 19, fab_d, INK)
    c.vline(51, 7, 15, fab)
    # respaldo (borde sur, mirando al norte hacia la mesa y TV)
    c.rrect(3, 19, 50, 7, fab_d, INK)
    c.hline(5, 20, 46, fab)
    save(c, 'sofa/macro_idle.png')


def sofa_micro_tile():
    size = 160
    base, light, dark = hexc('#78A899'), hexc('#8AB8A9'), hexc('#649284')
    c = Canvas(size, size)
    px = c.im.load()
    for y in range(size):
        for x in range(size):
            band = (x + y) % 8
            if band == 0:
                col = light
            elif band == 4:
                col = dark
            else:
                col = base
            px[x, y] = col
    # capitoné/costura suave
    c.hline(0, 0, size, '#568072')
    c.vline(0, 0, size, '#568072')
    save(c, 'sofa/micro_tile.png')


def seat_shadows():
    size = 384
    cx, cy = size // 2, size // 2
    col = hexc(INK, 255)
    c = Canvas(size, size)
    px = c.im.load()
    # silueta redondeada de la sombra al sentarse
    for y in range(size):
        for x in range(size):
            dx = (x - cx) / 168.0
            dy = (y - cy) / 152.0
            if dx * dx + dy * dy <= 1.0:
                px[x, y] = col
    for d in ('north', 'east', 'south', 'west'):
        save(c, f'seat/micro_shadow_{d}.png')


def main(src):
    global SRC
    SRC = src
    chair_macro()
    chair_micro_tile()
    sofa_macro()
    sofa_micro_tile()
    seat_shadows()
