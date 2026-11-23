"""Arte micro de silla común, sofá y sombra al sentarse (M37; el macro está en macro.py).

- chair/micro_tile.png (160 × 160 px): tapizado de tela tramada del asiento (10 × 10 cm).
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
    chair_micro_tile()
    sofa_micro_tile()
    seat_shadows()
