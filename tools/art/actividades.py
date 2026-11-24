"""Arte del slice de actividades en la casa (M61–M71), vista micro.

M61:
- water_drop/micro_idle_1.png, micro_idle_2.png (6 × 6 px, pivote [3, 4]): gota con brillo.
- water_drop/micro_low.png (6 × 6 px, pivote [3, 4]): gota chica cuando le queda poco.
- puddle/micro_idle_1.png, micro_idle_2.png (14 × 6 px, pivote [7, 4]): charco en el borde de la ducha.
- puddle/micro_low.png (14 × 6 px, pivote [7, 4]): charco reducido.
- nest_water/micro_empty.png, micro_half.png, micro_full.png (18 × 10 px, pivote [9, 7]):
  tapita de botella del nido (vacía, a medias y llena).
"""
import os

from px import Canvas, hexc

SRC = None  # lo fija main
INK = '#3F365C'
WATER, WATER_HI, WATER_SH = '#89D1F2', '#E8FAFF', '#4D9AC4'
CAP, CAP_HI, CAP_SH, CAP_IN = '#E66E7C', '#FFA5B0', '#B64959', '#C95A69'


def save(c, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    c.save(path)


def water_drop():
    # Gota de 6 × 6 px, 2 cuadros de destello + variante low más chica.
    for frame in (1, 2):
        c = Canvas(6, 6)
        c.ellipse(3, 4, 2, 1, hexc(INK, 45))
        # contorno y cuerpo
        for x, y in [(2, 1), (3, 1), (1, 2), (4, 2), (1, 3), (4, 3), (2, 4), (3, 4)]:
            c.px(x, y, INK)
        for x, y in [(2, 2), (3, 2), (2, 3), (3, 3)]:
            c.px(x, y, WATER)
        c.px(3, 3, WATER_SH)
        # destello alternado
        c.px(2, 2, WATER_HI)
        if frame == 2:
            c.px(3, 2, WATER_HI)
            c.px(1, 1, hexc(WATER_HI, 180))
        save(c, f'water_drop/micro_idle_{frame}.png')

    # Variante low: más chiquita
    c = Canvas(6, 6)
    c.ellipse(3, 4, 1, 1, hexc(INK, 35))
    for x, y in [(2, 2), (3, 2), (2, 4), (3, 4), (1, 3), (4, 3)]:
        c.px(x, y, INK)
    c.px(2, 3, WATER_HI)
    c.px(3, 3, WATER)
    save(c, 'water_drop/micro_low.png')


def puddle():
    # Charco de 14 × 6 px, 2 cuadros de brillo + variante low.
    for frame in (1, 2):
        c = Canvas(14, 6)
        c.ellipse(7, 3, 6, 2, INK)
        c.ellipse(7, 3, 5, 1, WATER)
        for x in range(4, 11):
            c.px(x, 4, WATER_SH)
        if frame == 1:
            c.hline(4, 2, 3, WATER_HI)
            c.px(9, 3, WATER_HI)
        else:
            c.hline(5, 2, 2, WATER_HI)
            c.hline(8, 3, 2, WATER_HI)
        save(c, f'puddle/micro_idle_{frame}.png')

    c = Canvas(14, 6)
    c.ellipse(7, 3, 4, 1, INK)
    c.hline(4, 3, 7, WATER)
    c.hline(5, 3, 2, WATER_HI)
    save(c, 'puddle/micro_low.png')


def nest_water():
    # Tapita de botella de 18 × 10 px: vacía, a medias y llena.
    for state in ('empty', 'half', 'full'):
        c = Canvas(18, 10)
        # sombra
        c.ellipse(9, 8, 8, 2, hexc(INK, 55))
        # pared exterior de la tapita (ranurada)
        c.ellipse(9, 6, 8, 3, INK)
        c.ellipse(9, 6, 7, 2, CAP_SH)
        for x in range(3, 16, 2):
            c.px(x, 7, CAP)
        # borde superior
        c.ellipse(9, 4, 8, 3, INK)
        c.ellipse(9, 4, 7, 2, CAP)
        c.hline(4, 2, 10, CAP_HI)
        # interior de la tapita
        c.ellipse(9, 4, 5, 1, CAP_IN)
        if state == 'half':
            c.hline(5, 4, 9, WATER_SH)
            c.hline(6, 4, 7, WATER)
            c.hline(6, 4, 2, WATER_HI)
        elif state == 'full':
            c.ellipse(9, 4, 5, 1, WATER)
            c.hline(5, 3, 8, WATER)
            c.hline(6, 3, 4, WATER_HI)
            c.px(11, 4, WATER_SH)
        save(c, f'nest_water/micro_{state}.png')


def main(src):
    global SRC
    SRC = src
    water_drop()
    puddle()
    nest_water()
