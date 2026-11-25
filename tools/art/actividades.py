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


def trinkets():
    # M63: Tesoros perdidos (micro idle + held, macro punto de 2x2 px) y pila del nido.
    btn, btn_hi, btn_sh = '#E8859A', '#FFC2D0', '#B5556D'
    metal, metal_hi, metal_sh = '#C6D4E0', '#FFFFFF', '#7E90A4'
    thread, thread_hi, thread_sh = '#76C8B8', '#B8F0E4', '#4A9688'
    graph, graph_hi, graph_sh = '#555A6E', '#8E95AC', '#313544'

    # Punto macro 2x2 px (selected_only) para cada tesoro
    for name, col in (
        ('trinket_button', btn),
        ('trinket_clip', metal_hi),
        ('trinket_thread', thread),
        ('trinket_graphite', graph_hi),
    ):
        m = Canvas(2, 2)
        m.px(0, 0, col)
        m.px(1, 0, col)
        m.px(0, 1, INK)
        m.px(1, 1, col)
        save(m, f'{name}/macro_idle.png')

    # 1) Botón (16 x 16 px, cuatro agujeros)
    c = Canvas(16, 16)
    c.ellipse(8, 12, 6, 3, hexc(INK, 55))
    c.ellipse(8, 8, 6, 6, INK)
    c.ellipse(8, 8, 5, 5, btn)
    c.ellipse(7, 7, 4, 4, btn_hi)
    c.ellipse(8, 8, 3, 3, btn)
    for hx, hy in ((6, 7), (10, 7), (6, 10), (10, 10)):
        c.px(hx, hy, INK)
    c.px(5, 4, '#FFFFFF')
    save(c, 'trinket_button/micro_idle.png')
    # Variante carried (held) - mismo tamaño que idle para compartir pivote
    h = Canvas(16, 16)
    h.ellipse(8, 10, 5, 5, INK)
    h.ellipse(8, 10, 4, 4, btn)
    h.px(6, 9, INK)
    h.px(10, 9, INK)
    h.px(6, 11, INK)
    h.px(10, 11, INK)
    h.px(6, 7, btn_hi)
    save(h, 'trinket_button/micro_held.png')

    # 2) Clip metálico (20 x 8 px)
    c = Canvas(20, 8)
    c.ellipse(10, 6, 8, 1, hexc(INK, 45))
    # lazo exterior
    c.rrect(2, 1, 16, 6, metal, INK)
    c.rrect(4, 2, 12, 4, (0, 0, 0, 0), metal_sh)
    c.hline(5, 3, 9, metal_hi)
    c.hline(3, 1, 14, metal_hi)
    save(c, 'trinket_clip/micro_idle.png')
    h = Canvas(20, 8)
    h.rrect(3, 2, 14, 4, metal, INK)
    h.hline(5, 3, 10, metal_hi)
    save(h, 'trinket_clip/micro_held.png')

    # 3) Madeja de hilo con cabo suelto (14 x 10 px)
    c = Canvas(14, 10)
    c.ellipse(7, 8, 5, 2, hexc(INK, 45))
    c.ellipse(6, 5, 5, 4, INK)
    c.ellipse(6, 5, 4, 3, thread)
    c.hline(4, 3, 4, thread_hi)
    c.hline(4, 6, 5, thread_sh)
    # cabo suelto hacia la derecha
    for x, y in ((10, 6), (11, 7), (12, 7), (13, 6)):
        c.px(x, y, INK)
        c.px(x, y - 1, thread)
    save(c, 'trinket_thread/micro_idle.png')
    h = Canvas(14, 10)
    h.ellipse(7, 6, 4, 3, INK)
    h.ellipse(7, 6, 3, 2, thread)
    h.hline(5, 5, 3, thread_hi)
    save(h, 'trinket_thread/micro_held.png')

    # 4) Mina de grafito (12 x 4 px)
    c = Canvas(12, 4)
    c.hline(1, 3, 10, hexc(INK, 45))
    for x in range(1, 11):
        c.px(x, 0, INK)
        c.px(x, 1, graph_hi if x < 4 else graph)
        c.px(x, 2, graph_sh)
    c.px(0, 1, INK)
    c.px(11, 1, INK)
    save(c, 'trinket_graphite/micro_idle.png')
    h = Canvas(12, 4)
    for x in range(2, 10):
        h.px(x, 0, INK)
        h.px(x, 1, graph_hi if x < 4 else graph)
        h.px(x, 2, INK)
    save(h, 'trinket_graphite/micro_held.png')

    # 5) Capa hoard del nido en 3 etapas (28 x 14 px, pivote [14, 10])
    for stage in ('one', 'few', 'pile'):
        c = Canvas(28, 14)
        # etapa 1: un botón
        c.ellipse(9, 9, 4, 3, INK)
        c.ellipse(9, 9, 3, 2, btn)
        c.px(8, 8, btn_hi)
        if stage in ('few', 'pile'):
            # etapa 2: suma un clip y un hilo
            c.rrect(13, 8, 9, 4, metal, INK)
            c.hline(14, 9, 6, metal_hi)
            c.ellipse(18, 6, 4, 3, INK)
            c.ellipse(18, 6, 3, 2, thread)
            c.px(17, 5, thread_hi)
        if stage == 'pile':
            # etapa 3: pila completa con grafito y brillo extra
            for x in range(6, 16):
                c.px(x, 5, INK)
                c.px(x, 6, graph_hi if x < 10 else graph)
            c.ellipse(14, 5, 3, 2, INK)
            c.ellipse(14, 5, 2, 1, '#FFD866')
            c.px(13, 4, '#FFFFFF')
        save(c, f'nest_hoard/micro_{stage}.png')


def main(src):
    global SRC
    SRC = src
    water_drop()
    puddle()
    nest_water()
    shower_fx()
    trinkets()
