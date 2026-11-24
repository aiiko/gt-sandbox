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


def shower_fx():
    # 1) Sobre la giantess en macro (62 × 76 px, pivote [30, 68]):
    #    chorros de agua cayendo, salpicaduras en los pies y nubes suaves de vapor.
    steam_soft = hexc('#F2FAFF', 135)
    steam_mid = hexc('#FFFFFF', 175)
    spray_a = hexc('#89D1F2', 200)
    spray_b = hexc('#E8FAFF', 225)
    splash = hexc('#76C8F0', 190)

    for frame in range(4):
        c = Canvas(62, 76)
        # Nubes de vapor subiendo alrededor de cabeza y hombros
        rise = frame * 2
        for bx, by, rx, ry in (
            (14, 26 - rise, 7, 5),
            (46, 22 - (rise + 3) % 8, 8, 5),
            (22, 14 - (rise + 5) % 8, 6, 4),
            (40, 12 - (rise + 1) % 8, 6, 4),
            (10, 44 - (rise + 2) % 8, 6, 4),
            (50, 42 - (rise + 4) % 8, 6, 4),
        ):
            c.ellipse(bx, by, rx, ry, steam_soft)
            c.ellipse(bx - 1, by - 1, max(2, rx - 2), max(2, ry - 2), steam_mid)

        # Líneas de agua cayendo desde arriba
        for col_i, x0 in enumerate((16, 22, 28, 34, 40, 46)):
            y_off = (frame * 5 + col_i * 7) % 16
            for y in range(4 + y_off, 66, 14):
                col = spray_b if (col_i + frame) % 2 == 0 else spray_a
                for dy in range(6):
                    yy = y + dy
                    if 4 <= yy <= 66:
                        c.px(x0 + (dy // 4), yy, col)

        # Salpicaduras en la base (pies)
        ring_r = 12 + (frame % 2) * 3
        for dx in range(-ring_r, ring_r + 1, 3):
            c.px(30 + dx, 67 + ((dx + frame) % 2), splash)
        for sx, sy in ((18, 64 - frame % 3), (42, 63 - (frame + 1) % 3), (26, 65 - (frame + 2) % 3), (36, 64 - frame % 2)):
            c.px(sx, sy, spray_b)
            c.px(sx, sy + 1, spray_a)
        save(c, f'shower_fx/macro_idle_{frame + 1}.png')

    # 2) Sobre el plato de ducha en macro (80 × 80 px, pivote [2, 78]):
    #    chorro saliendo de la ducha, ondas en el piso y vapor.
    for frame in range(4):
        c = Canvas(80, 80)
        # Vapor suave sobre el plato
        rise = frame * 2
        for bx, by, rx, ry in (
            (32, 24 - rise, 11, 6),
            (52, 30 - (rise + 4) % 8, 10, 6),
            (40, 50 - (rise + 2) % 8, 12, 6),
        ):
            c.ellipse(bx, by, rx, ry, hexc('#F2FAFF', 110))
            c.ellipse(bx - 1, by - 1, max(3, rx - 3), max(2, ry - 2), hexc('#FFFFFF', 145))

        # Chorro desde la esquina superior derecha hacia el centro del plato
        for k in range(7):
            x0 = 66 - k * 5 - (frame * 2) % 6
            y0 = 10 + k * 4 + (frame * 3) % 7
            col = spray_b if (k + frame) % 2 == 0 else spray_a
            for d in range(7):
                xx, yy = x0 - d, y0 + d
                if 6 <= xx < 74 and 6 <= yy < 74:
                    c.px(xx, yy, col)

        # Ondas y salpicaduras en el plato
        r1 = 6 + (frame * 3) % 10
        r2 = 5 + ((frame + 2) * 3) % 10
        for cx, cy, rad in ((38, 48, r1), (50, 58, r2)):
            for step in range(-rad, rad + 1, 2):
                c.px(cx + step, cy - rad // 3, spray_b)
                c.px(cx + step, cy + rad // 3, spray_a)
        save(c, f'shower_tray_fx/macro_idle_{frame + 1}.png')

    # 3) Variante micro (48 × 48 px, pivote [24, 24]): gotitas y bruma fina.
    for frame in range(4):
        c = Canvas(48, 48)
        for bx, by in ((14, 16), (34, 20), (22, 34)):
            c.ellipse(bx, by - (frame % 3), 8, 5, hexc('#F2FAFF', 75))
        for i in range(8):
            x = 6 + (i * 5 + frame * 3) % 36
            y = 4 + (i * 7 + frame * 6) % 38
            c.px(x, y, spray_b)
            c.px(x, y + 1, spray_a)
            c.px(x, y + 2, spray_a)
        save(c, f'shower_fx/micro_idle_{frame + 1}.png')


def main(src):
    global SRC
    SRC = src
    water_drop()
    puddle()
    nest_water()
    shower_fx()
