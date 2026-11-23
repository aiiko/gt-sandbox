"""Arte de Vida cotidiana de los tinies (M41–M44), solo vista micro.

- napkin/micro_idle.png (64 × 56 px, pivote [32, 30]): servilleta de papel doblada, con fibras sueltas en
  un borde.
- fiber/micro_idle.png (14 × 6 px, pivote [7, 3]): una fibra de papel suelta.
- refuge_site/micro_idle.png (28 × 18 px, pivote [14, 13]): la obra, unas fibras apiladas.
- refuge_site/micro_built.png (30 × 24 px, pivote [15, 19]): el refugio terminado, una carpa de fibras.
"""
import os

from px import Canvas, hexc

SRC = None  # lo fija main
INK = '#3F365C'
PAPER, PAPER_HI, PAPER_SH = '#FBF6EC', '#FFFFFF', '#DCD3C4'
FIBER, FIBER_SH = '#F4EEE2', '#C9BFAE'


def save(c, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    c.save(path)


def napkin():
    c = Canvas(64, 56)
    c.ellipse(33, 50, 28, 4, hexc(INK, 50))
    c.rrect(4, 4, 56, 46, PAPER, INK)
    # doblez en diagonal y relieve del papel
    for i in range(40):
        c.px(10 + i, 8 + i * 38 // 40, PAPER_SH)
    for y in range(8, 46, 6):
        c.hline(8, y, 3, PAPER_HI)
        c.hline(48, y + 3, 4, PAPER_SH)
    c.hline(6, 5, 52, PAPER_HI)
    # borde deshilachado (de ahí salen las fibras)
    for x in range(8, 56, 5):
        c.vline(x, 49, 2, FIBER_SH)
        c.px(x + 1, 51, FIBER)
    save(c, 'napkin/micro_idle.png')


def fiber():
    c = Canvas(14, 6)
    c.ellipse(7, 4, 6, 1, hexc(INK, 45))
    pts = [(1, 3), (2, 2), (3, 2), (4, 3), (5, 3), (6, 2), (7, 2), (8, 3), (9, 3), (10, 2), (11, 2), (12, 3)]
    for x, y in pts:
        c.px(x, y - 1, INK)
        c.px(x, y, FIBER)
        c.px(x, y + 1, FIBER_SH)
    save(c, 'fiber/micro_idle.png')


def site():
    c = Canvas(28, 18)
    c.ellipse(14, 14, 12, 3, hexc(INK, 55))
    for k, (x, y) in enumerate([(4, 11), (9, 9), (14, 11), (7, 7), (13, 6), (18, 9)]):
        for i in range(9):
            yy = y + (1 if (i + k) % 4 < 2 else 0)
            c.px(x + i, yy - 1, INK)
            c.px(x + i, yy, FIBER if k % 2 else PAPER)
            c.px(x + i, yy + 1, FIBER_SH)
    save(c, 'refuge_site/micro_idle.png')


def built():
    c = Canvas(30, 24)
    c.ellipse(15, 20, 14, 3, hexc(INK, 60))
    # carpa: dos faldones de fibras que se apoyan arriba
    for y in range(3, 20):
        half = (y - 2) * 12 // 17
        for x in range(15 - half, 15 + half + 1):
            band = (x + y // 2) % 4
            c.px(x, y, FIBER_SH if band == 0 else (PAPER if band == 1 else FIBER))
        c.px(15 - half - 1, y, INK)
        c.px(15 + half + 1, y, INK)
    c.hline(2, 20, 27, INK)
    # entrada oscura
    for y in range(12, 20):
        w = (y - 11) * 4 // 8
        c.hline(15 - w, y, 2 * w + 1, hexc(INK, 200))
    c.px(15, 2, INK)
    save(c, 'refuge_site/micro_built.png')


def main(src):
    global SRC
    SRC = src
    napkin()
    fiber()
    site()
    built()
