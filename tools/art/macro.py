"""Casa en macro a 80 px/m (16 px por 20 cm), dibujada por código.

Cada habitación mide 3 × 3 m: 240 × 258 px con el pivote en [0, 244] (la
esquina suroeste del piso). La franja de arriba es la pared norte en tres
cuartos; el primer plano (`macro_fg`) son los muros este, oeste y sur, con el
hueco de las puertas a 1,5 m. Los muebles conservan la ubicación de la
escena: su pivote es el punto de la entidad y su tapa coincide con la
extensión del Space. La luz entra desde arriba a la izquierda.
"""
import math
import os

from shapes import Sprite

SRC = None  # lo fija main

INK = '#3F365C'
INK_SOFT = (0x3F, 0x36, 0x5C, 70)
WOOD, WOOD_D, WOOD_L = '#C89474', '#9E6E5C', '#E2B896'
PAPER, PAPER_L = '#F7EEDD', '#DDCBB4'
SUN = (0xFF, 0xF6, 0xDE, 34)

ROOM_W, ROOM_H, ROOM_Y0 = 240, 258, 244
WALL_H = 32           # pared norte
FLOOR_Y0 = WALL_H     # primera fila de piso
FLOOR_Y1 = 240        # última fila de piso + 1 (después viene el muro sur)
DOOR_Y = (112, 139)   # hueco de las puertas en los muros este y oeste


def save(s, rel):
    s.save(os.path.join(SRC, rel))


def noise(x, y, seed=0):
    """Ruido entero determinista para vetas y motas."""
    n = (x * 374761393 + y * 668265263 + seed * 1442695041) & 0xFFFFFFFF
    n = ((n ^ (n >> 13)) * 1274126177) & 0xFFFFFFFF
    return (n ^ (n >> 16)) & 0xFF


# ------------------------------------------------------------ pared norte

def north_wall(s, wall, wall_hi, wall_d, base, base_d):
    s.rect(0, 0, ROOM_W - 1, 4, INK)
    s.hline(0, ROOM_W - 1, 5, wall_hi)
    s.hline(0, ROOM_W - 1, 6, wall_hi)
    s.rect(0, 7, ROOM_W - 1, 27, wall)
    # textura de revoque: motas sueltas
    for y in range(8, 27):
        for x in range(ROOM_W):
            if noise(x, y, 3) < 6:
                s.px(x, y, wall_d)
    s.hline(0, ROOM_W - 1, 8, wall_hi)               # nageshi
    s.rect(0, 28, ROOM_W - 1, 31, base)               # zócalo
    s.hline(0, ROOM_W - 1, 31, base_d)
    for x in (6, 232):                                # columnas
        s.rect(x, 5, x + 3, 27, WOOD)
        s.vline(x, 5, 27, WOOD_L)
        s.vline(x + 3, 5, 27, WOOD_D)


def shoji_window(s, x, w):
    s.rect(x, 6, x + w - 1, 27, WOOD_D)
    s.rect(x + 2, 8, x + w - 3, 25, PAPER)
    for i in range(x + 2 + 7, x + w - 3, 8):
        s.vline(i, 8, 25, PAPER_L)
    for y in (13, 19):
        s.hline(x + 2, x + w - 3, y, PAPER_L)
    s.hline(x, x + w - 1, 6, WOOD)
    s.hline(x - 2, x + w + 1, 26, WOOD)
    s.hline(x - 2, x + w + 1, 27, WOOD_D)
    # luz que atraviesa el papel
    for y in range(9, 13):
        s.hline(x + 3, x + 6, y, '#FFFBF2')


def glass_window(s, x, w):
    """Ventana corrediza con vidrio celeste y reflejo."""
    s.rect(x, 5, x + w - 1, 26, '#D5E4E9')
    s.rect(x + 2, 7, x + w - 3, 24, '#B9E5EF')
    s.vline(x + w // 2, 7, 24, '#D5E4E9')
    s.vline(x + w // 2 + 1, 7, 24, '#8DB7C6')
    for i in range(6):
        s.px(x + 5 + i, 18 - i, '#E9F8FB')
        s.px(x + 6 + i, 18 - i, '#E9F8FB')
    s.hline(x - 2, x + w + 1, 25, '#F2E4E4')
    s.hline(x - 2, x + w + 1, 26, '#C9AEB5')
    s.hline(x + 2, x + w - 3, 7, '#8DB7C6')


def light_shaft(s, x, w, y0, y1, alpha=34):
    """Mancha de sol en diagonal desde una ventana."""
    for y in range(y0, y1):
        off = (y - y0) // 2
        for i in range(w):
            s.blend(x + off + i, y, (0xFF, 0xF6, 0xDE, alpha))


def plant(s, x, y):
    """Planta en maceta con el pie en (x, y)."""
    s.ell(x - 14, y - 1, x + 14, y + 4, INK_SOFT)
    s.rrect(x - 10, y - 16, x + 10, y + 1, '#D98E7E', INK, r=2)
    s.rect(x - 8, y - 15, x - 6, y - 2, '#F0B3A2')
    s.vline(x + 8, y - 14, y - 1, '#B9706A')
    s.rect(x - 12, y - 19, x + 12, y - 16, '#B9706A')
    s.hline(x - 12, x + 12, y - 19, INK)
    s.hline(x - 11, x + 11, y - 18, '#D98E7E')
    s.ell(x - 9, y - 19, x + 9, y - 15, '#6E4E46')
    leaves = [(-2, -48, 6, 24), (-14, -40, 7, 18), (8, -42, 7, 20), (-20, -30, 7, 12), (14, -32, 7, 14)]
    for lx, ly, w, h in leaves:
        s.rrect(x + lx, y + ly, x + lx + w, y + ly + h, '#6FB5A3', '#3E6E73', r=2)
        s.vline(x + lx + w // 2, y + ly + 2, y + ly + h - 3, '#9CE0D6')
        s.px(x + lx + 1, y + ly + 2, '#C6F0EA')


# ------------------------------------------------------------ pisos

def tatami_mat(s, x, y, w, h, horizontal):
    base, light, dark = '#C8C98F', '#D6D7A2', '#B4B67C'
    for j in range(h):
        for i in range(w):
            k = j if horizontal else i
            m = k % 4
            col = light if m == 0 else dark if m == 3 else base
            if noise(x + i, y + j, 7) < 10:
                col = dark
            s.px(x + i, y + j, col)
    # heri: bordes de tela en los lados largos
    edge, edge_hi = '#5D4F7A', '#7C6E9C'
    if horizontal:
        s.rect(x, y, x + w - 1, y + 1, edge)
        s.hline(x, x + w - 1, y, edge_hi)
        s.rect(x, y + h - 2, x + w - 1, y + h - 1, edge)
        s.vline(x, y, y + h - 1, '#A9AB72')
    else:
        s.rect(x, y, x + 1, y + h - 1, edge)
        s.vline(x, y, y + h - 1, edge_hi)
        s.rect(x + w - 2, y, x + w - 1, y + h - 1, edge)
        s.hline(x, x + w - 1, y, '#A9AB72')


def plank_floor(s, y0, y1, colors, seam, grain):
    row = 0
    for y in range(y0, y1, 12):
        col = colors[row % len(colors)]
        s.rect(0, y, ROOM_W - 1, min(y + 11, y1 - 1), col)
        s.hline(0, ROOM_W - 1, y, seam)
        off = (row * 74) % 120
        for x in range(off, ROOM_W, 120):
            s.vline(x, y, min(y + 11, y1 - 1), seam)
            s.vline(x + 1, y + 1, min(y + 11, y1 - 1), colors[(row + 1) % len(colors)])
        # vetas largas
        for k in range(3):
            gx = (row * 53 + k * 81) % ROOM_W
            gy = y + 3 + (k * 3) % 7
            if gy < y1:
                s.hline(gx, min(gx + 18 + k * 6, ROOM_W - 1), gy, grain)
        row += 1


def checker_floor(s, y0, y1, a, b, grout):
    for y in range(y0, y1):
        for x in range(ROOM_W):
            light = ((x // 16) + (y // 16)) % 2 == 0
            col = a if light else b
            if x % 16 == 0 or y % 16 == 0:
                col = grout
            s.px(x, y, col)


# ------------------------------------------------------------ habitaciones

def bedroom():
    s = Sprite(ROOM_W, ROOM_H)
    floor = [(0, 32, 120, 60, True), (120, 32, 120, 60, True),
             (0, 92, 60, 120, False), (180, 92, 60, 120, False),
             (60, 92, 120, 60, True), (60, 152, 120, 60, True),
             (0, 212, 120, 28, True), (120, 212, 120, 28, True)]
    for x, y, w, h, hor in floor:
        tatami_mat(s, x, y, w, h, hor)
    north_wall(s, '#EBD7C9', '#F5E6DA', '#DFC8B9', '#B98A6E', '#8F6650')
    shoji_window(s, 84, 64)
    # kakejiku con una rama de ciruelo
    s.hline(178, 200, 7, WOOD_D)
    s.rect(180, 8, 198, 26, '#7C6A95')
    s.rect(182, 10, 196, 24, PAPER)
    s.line([(185, 22), (189, 16), (193, 13)], '#6E4E46')
    s.line([(189, 16), (191, 19)], '#6E4E46')
    for px_, py_ in ((193, 12), (186, 20), (191, 19)):
        s.px(px_, py_, '#FF9CAD')
    s.rect(188, 26, 190, 27, '#7C6A95')
    light_shaft(s, 88, 56, 32, 112)
    plant(s, 212, 80)
    save(s, 'bedroom/macro_idle.png')


def living():
    s = Sprite(ROOM_W, ROOM_H)
    plank_floor(s, 32, 240, ['#D9B08E', '#D2A685', '#DDB694'], '#BF9076', '#C99C7C')
    north_wall(s, '#E8D3D3', '#F4E4E2', '#DCC4C6', '#B98A6E', '#8F6650')
    glass_window(s, 32, 50)
    glass_window(s, 160, 50)
    light_shaft(s, 40, 52, 32, 104, 30)
    light_shaft(s, 168, 52, 32, 104, 30)
    seigaiha_rug(s, 52, 104, 168, 112)
    low_sofa(s, 12, 24)
    plant(s, 216, 84)
    for zx, zy in ((92, 148), (196, 148)):             # zabuton junto a la mesa
        s.rrect(zx - 10, zy - 8, zx + 10, zy + 9, '#FF9CAD', '#C66F86', r=2)
        s.hline(zx - 8, zx + 8, zy - 6, '#FFC2CD')
        s.rect(zx - 1, zy - 1, zx + 1, zy + 1, '#C66F86')
        for cx, cy in ((zx - 10, zy - 8), (zx + 10, zy - 8), (zx - 10, zy + 9), (zx + 10, zy + 9)):
            s.px(cx, cy, '#C66F86')
    save(s, 'living/macro_idle.png')


def seigaiha_rug(s, x, y, w, h):
    base, ring, ring2 = '#B997C2', '#D6BCD8', '#9F7FAE'
    s.rrect(x, y, x + w - 1, y + h - 1, '#8E6E9F', r=2)
    s.rect(x + 3, y + 3, x + w - 4, y + h - 4, base)
    for j in range(y + 5, y + h - 5):
        for i in range(x + 5, x + w - 5):
            u, v = (i - x) % 16, (j - y) % 12
            cx = 8 if ((j - y) // 12) % 2 == 0 else 0
            d = math.hypot(u - cx, v - 12)
            if 7 <= d < 8:
                s.px(i, j, ring)
            elif 10 <= d < 11:
                s.px(i, j, ring2)
    s.rrect(x, y, x + w - 1, y + h - 1, None, '#6F5684', r=2)
    s.rrect(x + 3, y + 3, x + w - 4, y + h - 4, None, '#A889B5', r=1)
    for i in range(x + 4, x + w - 4, 4):           # flecos
        s.vline(i, y - 2, y - 1, '#D6BCD8')
        s.vline(i, y + h, y + h + 1, '#D6BCD8')


def low_sofa(s, x, y):
    """Sofá bajo junto a la pared norte del living."""
    s.rect(x + 4, y + 36, x + 84, y + 41, INK_SOFT)
    s.rrect(x, y, x + 84, y + 38, '#9A8BE3', INK, r=2)
    s.rect(x + 3, y + 3, x + 81, y + 13, '#B3A7EC')
    s.hline(x + 3, x + 81, y + 4, '#C8BEF2')
    for cx in (x + 5, x + 44):
        s.rrect(cx, y + 14, cx + 35, y + 34, '#8577D2', '#6E64B8', r=2)
        s.hline(cx + 2, cx + 33, y + 15, '#9A8BE3')
    s.rrect(x + 10, y + 16, x + 24, y + 28, '#FF9CAD', '#C66F86', r=2)
    s.hline(x + 12, x + 22, y + 17, '#FFC2CD')
    s.rrect(x + 62, y + 16, x + 76, y + 28, '#9CE0D6', '#5FA89C', r=2)
    s.hline(x + 64, x + 74, y + 17, '#C6F0EA')


def kitchen():
    s = Sprite(ROOM_W, ROOM_H)
    checker_floor(s, 32, 240, '#E6E4CC', '#CDE0D2', '#D7DCCB')
    north_wall(s, '#D8EAE4', '#E8F4F0', '#C9DED8', '#9DBDB5', '#789B94')
    # azulejos celestes sobre la mesada
    for y in range(8, 27):
        for x in list(range(10, 78)) + list(range(162, 230)):
            col = '#B9E5EF' if ((x // 6) + (y // 6)) % 2 else '#CBEDF3'
            if x % 6 == 0 or y % 6 == 0:
                col = '#E9F8FB'
            s.px(x, y, col)
    glass_window(s, 86, 68)
    # mesada con bacha (oeste) y anafe (este)
    counter(s, 10, 76)
    sink(s, 24, 34)
    s.rrect(56, 32, 72, 48, '#FFF1D8', INK, r=2)          # arrocera
    s.hline(58, 70, 38, '#D9CBB9')
    s.rect(62, 42, 66, 44, '#9A8BE3')
    s.px(64, 34, '#C9B9A6')
    counter(s, 162, 228)
    stove(s, 168, 34)
    # estante con cuencos sobre el anafe
    s.hline(168, 204, 16, WOOD_D)
    s.hline(168, 204, 17, '#6E4E46')
    for bx, col in ((172, '#FF9CAD'), (182, '#9CE0D6'), (192, '#9A8BE3')):
        s.rrect(bx, 11, bx + 7, 15, col, INK, r=1)
        s.hline(bx + 1, bx + 6, 12, '#FFFFFF')
    # paño rojo tomate colgado
    s.rect(214, 36, 222, 50, '#EB5547')
    s.hline(214, 222, 36, '#A93434')
    for y in (40, 44, 48):
        s.hline(214, 222, y, '#FFFFFF')
    # alfombra de cocina larga con asanoha simplificado
    s.rrect(60, 192, 148, 219, '#E9B49A', '#B87A72', r=2)
    s.rrect(63, 195, 145, 216, None, '#F3CDB9', r=1)
    for i in range(68, 142, 12):
        for dx, dy in ((4, 6), (2, 8), (6, 8), (4, 10), (4, 16), (2, 18), (6, 18)):
            s.px(i + dx, 192 + dy, '#FFF1D8')
    save(s, 'kitchen/macro_idle.png')


def counter(s, x0, x1):
    s.rect(x0, 30, x1, 33, '#F2F7F4')                     # tapa
    s.hline(x0, x1, 30, '#FFFFFF')
    s.rect(x0, 34, x1, 58, '#7FB0A8')                     # frente
    s.hline(x0, x1, 34, '#5E8E88')
    for x in range(x0 + 16, x1, 18):
        s.vline(x, 35, 57, '#5E8E88')
    for x in range(x0 + 6, x1 - 4, 18):
        s.hline(x, x + 5, 40, '#D5E8E2')                  # tiradores
    s.hline(x0, x1, 58, INK)
    s.rect(x0, 59, x1, 61, INK_SOFT)


def sink(s, x, y):
    s.rrect(x, y - 3, x + 22, y + 5, '#BCD3D6', '#6F8E96', r=2)
    s.rect(x + 3, y - 1, x + 19, y + 3, '#9CB8BE')
    s.vline(x + 11, y - 10, y - 3, '#8DA4AC')             # canilla
    s.hline(x + 9, x + 14, y - 10, '#8DA4AC')
    s.px(x + 12, y - 8, '#E9F8FB')


def stove(s, x, y):
    s.rrect(x, y - 4, x + 40, y + 8, '#E9E5E0', '#7C7686', r=2)
    for cx in (x + 10, x + 30):
        s.ring(cx - 6, y - 2, cx + 6, y + 6, '#3E3A4A')
        s.ring(cx - 4, y, cx + 4, y + 4, '#5C576A')
    # pava menta
    s.ell(x + 24, y - 6, x + 36, y + 2, '#9CE0D6')
    s.hline(x + 26, x + 34, y - 7, INK)
    s.px(x + 37, y - 3, INK)
    s.px(x + 28, y - 4, '#FFFFFF')
    s.rect(x + 2, y + 6, x + 38, y + 7, '#D3CEC8')


# ------------------------------------------------------------ primer plano

WALL_SIDE = ['#E2BCC3', '#C7A1B2', '#AA899E', '#AA899E', '#9C7D93', '#8C7189', '#776681', '#6B5B76']


def room_fg(name, west_door, east_door):
    s = Sprite(ROOM_W, ROOM_H)
    for x0, door, mirror in ((0, west_door, False), (ROOM_W - 8, east_door, True)):
        cols = list(reversed(WALL_SIDE)) if mirror else WALL_SIDE
        for i, col in enumerate(cols):
            s.vline(x0 + i, 6, 239, col)
        if door:
            s.im.paste((0, 0, 0, 0), (x0, DOOR_Y[0], x0 + 8, DOOR_Y[1] + 1))
            for y in (DOOR_Y[0] - 2, DOOR_Y[0] - 1, DOOR_Y[1] + 1, DOOR_Y[1] + 2):
                s.hline(x0, x0 + 7, y, '#F2CDBF')
        # remate superior del muro
        s.hline(x0, x0 + 7, 6, '#F2CDBF')
    # muro sur
    s.rect(0, 240, ROOM_W - 1, 243, '#E7C8CD')
    s.hline(0, ROOM_W - 1, 240, '#F6E1E3')
    s.rect(0, 244, ROOM_W - 1, 251, '#A68AA2')
    for x in range(0, ROOM_W, 30):                        # juntas del zócalo
        s.vline(x, 245, 251, '#957A92')
    s.rect(0, 252, ROOM_W - 1, 253, INK)
    save(s, f'{name}/macro_fg.png')


# ------------------------------------------------------------ puertas

def fusuma():
    """Puerta corrediza decorada con una ola lila (28 × 36, pivote [14, 18])."""
    s = Sprite(28, 36)
    s.rect(0, 0, 27, 35, WOOD_D)
    s.rect(2, 2, 25, 33, PAPER)
    s.vline(13, 2, 33, WOOD)
    s.vline(14, 2, 33, WOOD_D)
    for x in range(2, 26):
        y = 26 + round(2 * math.sin(x / 2.2))
        s.vline(x, y, 33, '#B9ADEB')
        s.px(x, y, '#D6CEF5')
    for x in range(2, 26, 5):
        s.px(x, 30, '#9A8BE3')
    # tiradores redondos
    for cx in (9, 18):
        s.ring(cx - 2, 15, cx + 2, 19, '#7C6A95')
    s.px(20, 8, '#FF9CAD')
    s.px(21, 9, '#C66F86')
    s.hline(0, 27, 0, WOOD_L)
    s.hline(0, 27, 1, WOOD)
    save(s, 'door/macro_idle.png')


def noren():
    """Cortina noren en el paso a la cocina (28 × 36, pivote [14, 18])."""
    s = Sprite(28, 36)
    s.rect(0, 0, 27, 3, WOOD_D)
    s.hline(0, 27, 0, WOOD_L)
    for x0 in (3, 14):                                   # dos telas que se tocan al centro
        s.rect(x0, 4, x0 + 10, 27, '#6E64B8')
        s.vline(x0, 4, 27, '#5A5199')
        s.hline(x0, x0 + 10, 27, '#5A5199')
        s.vline(x0 + 2, 5, 25, '#7F75C9')
    s.vline(24, 4, 27, '#5A5199')
    # flor blanca centrada sobre las dos telas
    for dx, dy in ((0, -3), (3, 0), (0, 3), (-3, 0), (2, -2), (-2, 2), (2, 2), (-2, -2)):
        s.px(14 + dx, 13 + dy, '#FFF1D8')
    s.px(14, 13, '#FF9CAD')
    s.rect(0, 28, 27, 35, INK_SOFT)
    save(s, 'door_noren/macro_idle.png')


# ------------------------------------------------------------ muebles

def futon():
    """Futón con almohada y edredón doblado (68 × 98, pivote [32, 70])."""
    s = Sprite(68, 98)
    s.rect(8, 12, 65, 94, (0x6E, 0x5F, 0x86, 90))
    s.rrect(4, 8, 61, 89, '#FFF1D8', INK, r=2)           # shikibuton
    s.vline(60, 10, 88, '#D9CBB9')
    s.hline(6, 59, 87, '#D9CBB9')
    s.hline(6, 59, 10, '#FFFBF2')
    # makura
    s.rrect(16, 12, 50, 26, '#F7EEDD', '#7C6A95', r=3)
    s.hline(18, 48, 18, '#9A8BE3')
    s.hline(18, 48, 19, '#B3A7EC')
    s.hline(19, 47, 14, '#FFFFFF')
    # kakebuton
    s.rrect(6, 30, 59, 87, '#9A8BE3', INK, r=2)
    s.rect(8, 32, 57, 35, '#C3B9F0')
    s.hline(8, 57, 36, '#B3A7EC')
    for y in range(42, 84, 12):
        for x in range(13, 56, 12):
            o = 6 if (y // 12) % 2 else 0
            fx, fy = x + o, y
            if fx < 55:
                for dx, dy in ((0, 0), (-1, 0), (1, 0), (0, -1), (0, 1)):
                    s.px(fx + dx, fy + dy, '#FF9CAD' if (dx, dy) == (0, 0) else '#FFC2CD')
                s.px(fx - 2, fy, '#FFC2CD')
                s.px(fx + 2, fy, '#FFC2CD')
                s.px(fx, fy - 2, '#FFC2CD')
                s.px(fx, fy + 2, '#FFC2CD')
    s.vline(57, 37, 85, '#7F70C9')
    s.hline(8, 57, 85, '#7F70C9')
    save(s, 'bed/macro_idle.png')


def desk_like(top, top_l, top_d, front, front_d, leg, rel, detail):
    """Mesa baja de 0,6 × 0,4 m (62 × 54, pivote [6, 34]): tapa de 48 × 32."""
    s = Sprite(62, 54)
    s.ell(8, 40, 61, 53, INK_SOFT)                        # sombra en el piso
    s.rect(10, 42, 13, 50, leg)                           # patas
    s.rect(47, 42, 50, 50, leg)
    s.vline(10, 42, 50, front)
    s.rrect(4, 0, 55, 36, top, INK, r=2)                  # tapa
    s.hline(6, 53, 2, top_l)
    s.hline(6, 53, 3, top_l)
    s.vline(5, 3, 34, top_l)
    for k, y in enumerate(range(8, 34, 6)):               # vetas
        x = 8 + (k * 13) % 30
        s.hline(x, x + 14, y, top_d)
    s.rect(4, 36, 55, 43, front)                          # canto
    s.hline(5, 54, 37, top_l)
    s.hline(4, 55, 43, front_d)
    s.vline(4, 36, 43, INK)
    s.vline(55, 36, 43, INK)
    s.hline(5, 54, 44, INK)
    detail(s)
    save(s, rel)


def desk():
    def notebook(s):
        s.rrect(8, 6, 22, 18, '#FBDFCB', INK, r=1)
        for y in (9, 12, 15):
            s.hline(10, 20, y, '#C7908C')
        s.vline(23, 8, 16, '#7C67A1')                     # lápiz
        s.px(23, 17, '#3F365C')
    desk_like('#D5A07A', '#EFC2A1', '#C08A66', '#A6869F', '#875E72', '#8C6077', 'desk/macro_idle.png', notebook)


def table():
    def placemat(s):
        s.ell(16, 8, 44, 24, '#BA9DB9')
        s.ell(18, 10, 42, 22, '#CCB3C1')
        s.ring(18, 10, 42, 22, '#A6869F')
        s.hline(24, 36, 13, '#E2D2DC')
    desk_like('#D5A07A', '#EFC2A1', '#C08A66', '#A6869F', '#875E72', '#8C6077', 'table/macro_idle.png', placemat)


def fridge():
    """Heladera chica con imanes (46 × 70, pivote [22, 32])."""
    s = Sprite(46, 70)
    s.rect(6, 62, 45, 69, INK_SOFT)
    s.rrect(2, 0, 41, 64, '#C0D7CB', INK, r=3)
    s.rect(4, 2, 7, 62, '#D5E4D9')                        # luz desde la izquierda
    s.vline(39, 3, 61, '#9CB5B2')
    s.hline(3, 40, 28, '#75989D')                         # división del freezer
    s.hline(3, 40, 29, '#ADCCC3')
    s.rect(33, 10, 35, 22, '#587084')                     # manijas
    s.rect(33, 36, 35, 54, '#587084')
    s.vline(33, 10, 22, '#9CB5B2')
    s.vline(33, 36, 54, '#9CB5B2')
    s.hline(4, 39, 59, '#75989D')
    s.rect(8, 60, 36, 62, '#678B94')                      # rejilla
    # imanes y una nota
    s.rect(9, 7, 12, 10, '#FF9CAD')
    s.rect(17, 11, 20, 14, '#9A8BE3')
    s.rrect(9, 36, 20, 46, '#F7EEDD', '#DDCBB4', r=1)
    s.hline(11, 17, 39, '#9A8BE3')
    s.hline(11, 15, 42, '#ECABAF')
    s.ell(24, 46, 28, 50, '#FFC86B')
    save(s, 'fridge/macro_idle.png')


def tv():
    """Televisor bajo sobre un mueble (72 × 58, pivote [36, 48])."""
    s = Sprite(72, 58)
    s.rect(4, 50, 71, 57, INK_SOFT)
    # mueble
    s.rrect(2, 34, 69, 52, '#B2839A', INK, r=2)
    s.hline(4, 67, 36, '#D5A5AF')
    s.rect(6, 41, 33, 49, '#AC859D')
    s.rect(38, 41, 65, 49, '#AC859D')
    s.hline(17, 22, 45, '#6C5579')
    s.hline(49, 54, 45, '#6C5579')
    # pantalla
    s.rect(28, 30, 43, 34, '#585979')
    s.rrect(6, 2, 65, 31, '#3F365C', r=3)
    s.rrect(9, 5, 62, 28, '#698BA1', r=1)
    for i in range(14):                                   # reflejo diagonal
        s.hline(40 - i, 46 - i, 6 + i, '#8CABBA')
    s.rect(10, 6, 30, 7, '#BDD7DB')
    s.px(60, 27, '#9CE0D6')                               # luz de encendido
    save(s, 'tv/macro_idle.png')


def nest():
    """Caja de fósforos del nido con la etiqueta (18 × 20, pivote [8, 14])."""
    s = Sprite(18, 20)
    s.ell(1, 13, 17, 19, INK_SOFT)
    s.rrect(1, 4, 16, 16, '#76D8CB', INK, r=1)
    s.hline(2, 15, 5, '#B5EEE6')
    s.rect(2, 13, 15, 15, '#458F99')
    s.rrect(4, 7, 13, 12, '#FFF1D8', None)                # etiqueta
    s.ell(5, 8, 8, 11, '#FF6F86')
    for x in (10, 12):
        s.vline(x, 8, 11, INK)
    # techo de la caja entreabierta
    s.poly([(3, 4), (8, 0), (14, 4)], '#5FB8AC')
    s.line([(3, 4), (8, 0), (14, 4)], INK)
    save(s, 'nest/macro_idle.png')


def cup(held):
    """Yunomi menta con flor (16 × 10, pivote [12, 6])."""
    s = Sprite(16, 10)
    if not held:
        s.ell(3, 6, 15, 9, INK_SOFT)
    s.rrect(4, 1, 13, 8, '#9CE0D6', INK, r=1)
    s.rect(5, 1, 12, 2, '#FFF1D8')                        # borde con el té
    s.hline(6, 11, 2, '#C98F5A')
    s.vline(5, 3, 7, '#C6F0EA')
    s.px(8, 5, '#FF9CAD')
    s.px(10, 7, '#3D9295')
    s.vline(14, 3, 5, INK)                                # asa
    s.px(15, 4, INK)
    if held:
        s.rect(13, 3, 15, 6, '#FFD8B9')                   # dedos sobre el asa
        s.px(15, 6, '#F2B79E')
    save(s, f'cup/macro_{"held" if held else "idle"}.png')


def plate(held):
    """Plato de postre con una galleta (12 × 8 o 14 × 8, pivote [6, 4])."""
    s = Sprite(14 if held else 12, 8)
    s.ell(0, 1, 11, 7, INK)
    s.ell(1, 2, 10, 6, '#FFF6EA')
    s.hline(3, 8, 2, '#FFFFFF')
    s.ell(3, 3, 7, 5, '#D9A066')
    s.px(4, 3, '#E8B77D')
    s.px(6, 4, '#8F5A3C')
    if held:
        s.rect(10, 2, 12, 5, '#FFD8B9')
        s.px(13, 4, '#FFD8B9')
    save(s, f'plate/macro_{"held" if held else "idle"}.png')


def chair():
    """Silla de madera con cojín (40 × 48, pivote [4, 36]): asiento de 32 × 32."""
    s = Sprite(40, 48)
    wood_d, wood, wood_l = '#8A5A44', '#B57B5A', '#D19B76'
    seat_d, seat, seat_l = '#C48A6A', '#E2AA86', '#F2C5A4'
    s.ell(2, 38, 38, 47, INK_SOFT)
    for x in (6, 30):                                     # patas
        s.rect(x, 32, x + 3, 45, wood_d)
        s.vline(x, 32, 44, wood)
    s.rrect(4, 4, 35, 36, seat, INK, r=2)                 # cojín
    s.rect(6, 6, 33, 8, seat_l)
    s.hline(6, 33, 34, seat_d)
    for x, y in ((12, 14), (27, 14), (12, 26), (27, 26)):
        s.px(x, y, seat_d)                                # botones del capitoné
    # respaldo en el borde sur (mira al norte, hacia el escritorio)
    s.rrect(4, 28, 35, 38, wood, INK, r=2)
    s.hline(7, 32, 30, wood_l)
    s.hline(6, 33, 36, wood_d)
    for x in (12, 20, 28):
        s.vline(x, 31, 35, wood_d)
    save(s, 'chair/macro_idle.png')


def sofa():
    """Sofá de tres cuerpos (112 × 60, pivote [8, 44]): asiento de 96 × 40."""
    s = Sprite(112, 60)
    fab_d, fab, fab_l = '#5A8578', '#78A899', '#99C7B8'
    wood = '#8A5A44'
    s.ell(4, 48, 108, 59, INK_SOFT)
    for x in (10, 98):
        s.rect(x, 48, x + 3, 55, wood)
    s.rrect(8, 6, 103, 48, fab, INK, r=3)
    s.rect(11, 9, 100, 11, fab_l)
    for x in (40, 72):                                    # almohadones
        s.vline(x, 9, 42, fab_d)
        s.vline(x + 1, 9, 42, fab_l)
    for cx in (24, 56, 88):
        s.px(cx, 24, fab_d)
    s.rrect(2, 10, 12, 48, fab_d, INK, r=2)               # apoyabrazos
    s.vline(4, 13, 45, fab)
    s.rrect(99, 10, 109, 48, fab_d, INK, r=2)
    s.vline(101, 13, 45, fab)
    # respaldo en el borde sur (mira al norte, hacia la mesa y la TV)
    s.rrect(6, 38, 105, 52, fab_d, INK, r=2)
    s.hline(9, 102, 40, fab)
    s.hline(9, 102, 41, fab_l)
    save(s, 'sofa/macro_idle.png')


def footprint():
    """Sombra de la próxima pisada: un óvalo opaco; la presentación lo aclara."""
    s = Sprite(14, 8)
    s.ell(0, 1, 13, 6, INK)
    save(s, 'footprint/macro_idle.png')


def main(src):
    global SRC
    SRC = src
    bedroom()
    living()
    kitchen()
    room_fg('kitchen', False, True)
    room_fg('bedroom', True, True)
    room_fg('living', True, False)
    fusuma()
    noren()
    futon()
    desk()
    table()
    fridge()
    tv()
    nest()
    cup(False)
    cup(True)
    plate(False)
    plate(True)
    chair()
    sofa()
    footprint()
