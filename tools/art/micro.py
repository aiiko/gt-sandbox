"""Detalles japoneses en las piezas micro."""
import math
import os
from px import Canvas, load, hexc

SRC = BASE = None  # los fija main
INK = '#3F365C'


def nest_label():
    old = load(os.path.join(BASE, 'nest/micro_fg.png'))
    c = Canvas(old.width, old.height)
    c.paste(old, 0, 0)
    # buscar el labio: primera fila opaca
    ys = [y for y in range(old.height) if any(old.getpixel((x, y))[3] for x in range(old.width))]
    y0 = ys[0] + 6
    c.rrect(14, y0, 22, 9, '#FFF1D8', INK)
    c.ellipse(19, y0 + 4, 2, 2, '#FF6F86')
    for i, (dx, dy, h) in enumerate([(25, 2, 4), (28, 2, 5), (31, 3, 3)]):
        c.vline(dx, y0 + dy, h, INK)
    c.px(26, y0 + 3, INK)
    c.px(30, y0 + 4, INK)
    c.save(os.path.join(SRC, 'nest/micro_fg.png'))


def coaster():
    old = load(os.path.join(BASE, 'table/micro_idle.png'))
    c = Canvas(old.width, old.height)
    c.paste(old, 0, 0)
    cx, cy, r = 650, 378, 44
    for y in range(cy - r - 4, cy + r + 6):
        for x in range(cx - r - 4, cx + r + 6):
            d = math.hypot(x - cx - 3, y - cy - 4)
            if d <= r:
                c.blend(x, y, hexc('#7A5A6E', 70))
    for y in range(cy - r, cy + r + 1):
        for x in range(cx - r, cx + r + 1):
            d = math.hypot(x - cx, y - cy)
            if d > r:
                continue
            slat = ((x - cx + r) // 9) % 2
            col = '#E3C98F' if slat == 0 else '#D4B77B'
            if (x - cx + r) % 9 == 0:
                col = '#B89A62'
            if d > r - 4:
                col = '#C9AA6E' if d <= r - 1 else '#8F7447'
            c.px(x, y, col)
    # nudos del bambú
    for k in range(-r + 4, r, 9):
        for y in (cy - 14, cy + 17):
            if math.hypot(k + 4, y - cy) < r - 5:
                c.hline(cx + k + 1, y, 7, '#B89A62')
    # brillo de luz desde arriba a la izquierda
    for y in range(cy - r + 6, cy - r + 9):
        for x in range(cx - 20, cx + 8):
            if math.hypot(x - cx, y - cy) < r - 4:
                c.blend(x, y, hexc('#FFF6DE', 80))
    c.save(os.path.join(SRC, 'table/micro_idle.png'))


def main(src, base):
    global SRC, BASE
    SRC, BASE = src, base
    nest_label()
    coaster()
