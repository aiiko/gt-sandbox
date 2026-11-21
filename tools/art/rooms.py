"""Redibuja la casa con ambientación japonesa (macro)."""
import os
from px import Canvas, load, hexc

SRC = BASE = None  # los fija main

INK = '#3F365C'
WOOD = '#C89474'
WOOD_D = '#9E6E5C'
WOOD_L = '#E2B896'
PAPER = '#F7EEDD'
PAPER_L = '#DDCBB4'


def top_wall(c, wall, wall_hi, base):
    c.rect(0, 0, 120, 3, INK)
    c.hline(0, 3, 120, wall_hi)
    c.rect(0, 4, 120, 10, wall)
    c.rect(0, 14, 120, 2, base)


def shoji_window(c, x, w):
    c.rect(x, 3, w, 11, WOOD_D)
    c.rect(x + 1, 4, w - 2, 9, PAPER)
    for i in range(x + 1 + 4, x + w - 1, 5):
        c.vline(i, 4, 9, PAPER_L)
    c.hline(x + 1, 8, w - 2, PAPER_L)
    c.hline(x, 3, w, WOOD)
    c.hline(x - 1, 13, w + 2, WOOD)


def light_shaft(c, x, w, y0, y1, alpha=34):
    for y in range(y0, y1):
        off = (y - y0) // 2
        for i in range(w):
            c.blend(x + off + i, y, hexc('#FFF6DE', alpha))


def plant(c, x, y):
    """Planta en maceta con el pie en (x, y)."""
    c.ellipse(x, y + 1, 7, 2, hexc('#3F365C', 60))
    c.rrect(x - 5, y - 8, 11, 9, '#D98E7E', INK)
    c.hline(x - 4, y - 7, 9, '#F0B3A2')
    c.hline(x - 6, y - 9, 13, INK)
    c.hline(x - 5, y - 8, 11, '#B9706A')
    leaves = [(-1, -24, 3, 12), (-7, -20, 4, 9), (4, -21, 4, 10), (-10, -15, 4, 6), (7, -16, 4, 7)]
    for lx, ly, w, h in leaves:
        c.rrect(x + lx, y + ly, w, h, '#6FB5A3', '#3E6E73')
        c.vline(x + lx + w // 2, y + ly + 1, h - 2, '#9CE0D6')


def low_sofa(c, x, y):
    c.rect(x + 2, y + 18, 40, 3, hexc('#3F365C', 60))
    c.rrect(x, y, 42, 19, '#9A8BE3', INK)
    c.rect(x + 2, y + 2, 38, 5, '#B3A7EC')
    c.rrect(x + 3, y + 7, 17, 10, '#8577D2', '#6E64B8')
    c.rrect(x + 22, y + 7, 17, 10, '#8577D2', '#6E64B8')
    c.rrect(x + 5, y + 8, 7, 6, '#FF9CAD', '#C66F86')
    c.rrect(x + 31, y + 8, 7, 6, '#9CE0D6', '#5FA89C')


def tatami_mat(c, x, y, w, h, horizontal):
    base, light, dark = hexc('#C8C98F'), hexc('#D3D49E'), hexc('#B7B97F')
    for j in range(h):
        for i in range(w):
            k = (j if horizontal else i)
            col = light if k % 3 == 0 else (dark if k % 3 == 2 else base)
            c.px(x + i, y + j, col)
    # heri: bordes de tela en los lados largos
    if horizontal:
        c.hline(x, y, w, '#5D4F7A')
        c.hline(x, y + h - 1, w, '#5D4F7A')
        c.vline(x, y, h, '#A9AB72')
    else:
        c.vline(x, y, h, '#5D4F7A')
        c.vline(x + w - 1, y, h, '#5D4F7A')
        c.hline(x, y, w, '#A9AB72')


def bedroom():
    c = Canvas(120, 129)
    top_wall(c, '#EBD7C9', '#F5E6DA', '#B98A6E')
    for x in (3, 116):
        c.rect(x, 3, 2, 11, WOOD)
    c.hline(0, 4, 120, WOOD_L)  # nageshi
    shoji_window(c, 42, 32)
    # kakejiku
    c.rect(90, 4, 9, 10, '#7C6A95')
    c.rect(91, 5, 7, 8, PAPER)
    c.vline(94, 6, 4, INK)
    c.px(93, 7, INK)
    c.px(95, 8, INK)
    c.px(96, 11, '#FF9CAD')
    c.hline(89, 4, 11, WOOD_D)
    # tatami
    floor = [(0, 16, 60, 30, True), (60, 16, 60, 30, True),
             (0, 46, 30, 60, False), (90, 46, 30, 60, False),
             (30, 46, 60, 30, True), (30, 76, 60, 30, True),
             (0, 106, 60, 14, True), (60, 106, 60, 14, True)]
    for x, y, w, h, hor in floor:
        tatami_mat(c, x, y, w, h, hor)
    light_shaft(c, 44, 28, 16, 56)
    plant(c, 106, 40)
    c.save(os.path.join(SRC, 'bedroom/macro_idle.png'))


def plank_floor(c, y0, y1, colors, seams):
    row = 0
    for y in range(y0, y1, 6):
        col = colors[row % len(colors)]
        c.rect(0, y, 120, min(6, y1 - y), col)
        c.hline(0, y, 120, seams)
        off = (row * 37) % 60
        for x in range(off, 120, 60):
            c.vline(x, y, min(6, y1 - y), seams)
        row += 1
    for x, y in [(17, 30), (84, 47), (40, 71), (100, 90), (12, 103), (66, 113)]:
        c.hline(x, y, 3, seams)


def seigaiha_rug(c, x, y, w, h):
    base, ring, ring2 = hexc('#B997C2'), hexc('#D6BCD8'), hexc('#9F7FAE')
    c.rrect(x, y, w, h, '#8E6E9F')
    c.rect(x + 2, y + 2, w - 4, h - 4, base)
    for j in range(y + 3, y + h - 3):
        for i in range(x + 3, x + w - 3):
            u, v = (i - x) % 8, (j - y) % 6
            cx = 4 if ((j - y) // 6) % 2 == 0 else 0
            d = abs(u - cx) + v
            if d in (3, 4) and v > 0:
                c.px(i, j, ring if d == 3 else ring2)
    c.rrect(x, y, w, h, None, '#6F5684')


def living():
    old = load(os.path.join(BASE, 'living/macro_idle.png'))
    c = Canvas(120, 129)
    plank_floor(c, 16, 120, [hexc('#D9B08E'), hexc('#D2A685'), hexc('#DDB694')], hexc('#BF9076'))
    c.paste(old.crop((0, 0, 120, 16)), 0, 0)  # pared y ventanas
    for x in (3, 116):
        c.rect(x, 3, 2, 11, WOOD)
    light_shaft(c, 20, 26, 16, 52, 30)
    light_shaft(c, 84, 26, 16, 52, 30)
    seigaiha_rug(c, 26, 52, 84, 56)
    # sofá bajo y planta del original
    low_sofa(c, 6, 12)
    plant(c, 108, 42)
    # zabuton junto a la mesa
    for zx, zy in [(46, 74), (98, 74)]:
        c.rrect(zx - 5, zy - 4, 11, 9, '#FF9CAD', '#C66F86')
        c.hline(zx - 4, zy - 3, 9, '#FFC2CD')
        c.px(zx, zy, '#C66F86')
    c.save(os.path.join(SRC, 'living/macro_idle.png'))


def kitchen():
    old = load(os.path.join(BASE, 'kitchen/macro_idle.png'))
    c = Canvas(120, 129)
    c.paste(old, 0, 0)
    # piso vinílico a cuadros crema y menta
    for y in range(32, 120):
        for x in range(0, 120):
            if y < 34 and (x < 40 or x > 79):
                continue
            a = ((x // 8) + (y // 8)) % 2 == 0
            c.px(x, y, '#E6E4CC' if a else '#CDE0D2')
    c.paste(old.crop((0, 16, 120, 32)), 0, 16)
    for x in (3, 116):
        c.rect(x, 3, 2, 11, WOOD)
    # arrocera sobre la mesada
    c.rrect(28, 17, 11, 9, '#FFF1D8', INK)
    c.hline(29, 20, 9, '#D9CBB9')
    c.rect(32, 22, 3, 2, '#9A8BE3')
    c.px(33, 18, '#C9B9A6')
    # pava sobre la hornalla
    c.ellipse(93, 21, 4, 3, '#9CE0D6')
    c.hline(91, 18, 5, INK)
    c.px(98, 21, INK)
    c.px(93, 20, '#FFFFFF')
    # estante con cuencos
    c.hline(84, 9, 18, WOOD_D)
    for bx, col in [(86, '#FF9CAD'), (91, '#9CE0D6'), (96, '#9A8BE3')]:
        c.rect(bx, 7, 4, 2, col)
        c.hline(bx, 6, 4, INK)
    # alfombra de cocina larga con asanoha simplificado
    c.rrect(30, 96, 44, 14, '#E9B49A', '#B87A72')
    for i in range(33, 71, 6):
        c.px(i + 2, 100, '#FFF1D8')
        c.px(i + 1, 101, '#FFF1D8')
        c.px(i + 3, 101, '#FFF1D8')
        c.px(i + 2, 102, '#FFF1D8')
        c.px(i + 2, 105, '#FFF1D8')
    c.save(os.path.join(SRC, 'kitchen/macro_idle.png'))


def futon():
    c = Canvas(34, 49)
    # sombra corta: el futón está casi al ras del tatami
    c.rect(4, 6, 29, 41, hexc('#6E5F86', 90))
    # shikibuton
    c.rrect(2, 4, 29, 41, '#FFF1D8', INK)
    c.vline(30, 5, 39, '#D9CBB9')
    c.hline(3, 43, 27, '#D9CBB9')
    # makura
    c.rrect(8, 6, 17, 7, '#F7EEDD', '#7C6A95')
    c.hline(9, 9, 15, '#9A8BE3')
    # kakebuton doblado hacia abajo
    c.rrect(3, 15, 27, 29, '#9A8BE3', INK)
    c.hline(4, 16, 25, '#C3B9F0')
    c.hline(4, 17, 25, '#B3A7EC')
    for y in range(20, 42, 6):
        for x in range(6, 28, 6):
            o = 3 if (y // 6) % 2 else 0
            px, py = x + o, y
            if px < 28:
                c.px(px, py, '#FF9CAD')
                c.px(px - 1, py, '#FFC2CD')
                c.px(px + 1, py, '#FFC2CD')
                c.px(px, py - 1, '#FFC2CD')
                c.px(px, py + 1, '#FFC2CD')
    c.vline(29, 18, 25, '#7F70C9')
    c.hline(4, 42, 25, '#7F70C9')
    c.save(os.path.join(SRC, 'bed/macro_idle.png'))


def fusuma():
    c = Canvas(14, 18)
    c.rect(0, 0, 14, 18, WOOD_D)
    c.rect(1, 1, 12, 16, PAPER)
    # ola lila en la parte inferior del panel
    for x in range(1, 13):
        y = 13 + (1 if x % 4 in (0, 3) else 0)
        c.vline(x, y, 17 - y, '#B9ADEB')
    c.px(10, 8, '#FF9CAD')
    c.px(10, 9, '#C66F86')
    c.hline(0, 0, 14, WOOD_L)
    c.vline(6, 1, 16, WOOD)
    c.save(os.path.join(SRC, 'door/macro_idle.png'))


def noren():
    c = Canvas(14, 18)
    c.rect(0, 0, 14, 2, WOOD_D)
    c.hline(0, 0, 14, WOOD_L)
    for x0 in (1, 7):
        c.rect(x0, 2, 6, 12, '#6E64B8')
        c.vline(x0, 2, 12, '#5A5199')
        c.hline(x0, 13, 6, '#5A5199')
    # motivo blanco centrado entre las dos telas
    c.px(5, 6, '#FFF1D8')
    c.px(6, 5, '#FFF1D8')
    c.px(8, 5, '#FFF1D8')
    c.px(9, 6, '#FFF1D8')
    c.px(5, 7, '#FFF1D8')
    c.px(9, 7, '#FFF1D8')
    c.px(6, 8, '#FFF1D8')
    c.px(8, 8, '#FFF1D8')
    c.rect(0, 14, 14, 4, hexc('#3F365C', 50))
    os.makedirs(os.path.join(SRC, 'door_noren'), exist_ok=True)
    c.save(os.path.join(SRC, 'door_noren/macro_idle.png'))


def fridge():
    old = load(os.path.join(BASE, 'fridge/macro_idle.png'))
    c = Canvas(old.width, old.height)
    c.paste(old, 0, 0)
    c.rect(4, 4, 2, 2, '#FF9CAD')
    c.rect(8, 6, 2, 2, '#9A8BE3')
    c.rect(4, 19, 4, 3, PAPER)
    c.hline(5, 20, 2, '#9A8BE3')
    c.save(os.path.join(SRC, 'fridge/macro_idle.png'))


def main(src, base):
    global SRC, BASE
    SRC, BASE = src, base
    bedroom()
    living()
    kitchen()
    futon()
    fusuma()
    noren()
    fridge()
