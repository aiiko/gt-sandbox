"""Sprites de mobiliario en vista micro a 400 px/m (4 px/cm, renderizados a scale=4 -> 1600 px/m).

Cada pieza coincide milímetro a milímetro con su variante macro (80 px/m × 5 = 400 px/m,
y × 4 en pantalla = 1600 px/m) para eliminar cualquier inconsistencia entre macro y micro.
"""
import math
import os

from shapes import Sprite

SRC = None

INK = '#3F365C'
INK_SOFT = (0x3F, 0x36, 0x5C, 65)
WOOD, WOOD_D, WOOD_L = '#C89474', '#9E6E5C', '#E2B896'


def save(s, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    s.save(path)


def noise(x, y, seed=0):
    n = (x * 374761393 + y * 668265263 + seed * 1442695041) & 0xFFFFFFFF
    n = ((n ^ (n >> 13)) * 1274126177) & 0xFFFFFFFF
    return (n ^ (n >> 16)) & 0xFF


# ------------------------------------------------------------ Dormitorio

def low_closet():
    """Placard bajo (500 × 260, pivote [10, 210])."""
    s = Sprite(500, 260)
    s.rect(20, 190, 495, 250, INK_SOFT)
    # tapa superior de madera con vetas
    s.rrect(10, 10, 485, 170, '#D5A07A', INK, r=6)
    s.rect(14, 14, 481, 20, '#EFC2A1')
    s.rect(14, 14, 20, 166, '#EFC2A1')
    for y in range(28, 164, 14):
        s.hline(16, 480, y, '#C8926E')
        for k in range(5):
            gx = 24 + ((y * 7 + k * 97) % 380)
            s.hline(gx, min(gx + 60, 476), y + 4, '#B9825E')
    # frente con dos puertas corredizas
    s.rect(10, 170, 485, 230, '#B57B5A')
    s.rrect(10, 170, 485, 230, None, INK, r=1)
    s.rect(14, 174, 481, 178, '#D19B76')
    s.rect(245, 174, 249, 226, '#8A5A44')
    s.rect(250, 174, 253, 226, '#D19B76')
    # marcos interiores y tiradores embutidos
    s.rrect(22, 182, 236, 220, '#AC7352', '#8A5A44', r=3)
    s.rrect(262, 182, 474, 220, '#AC7352', '#8A5A44', r=3)
    s.rrect(208, 192, 224, 212, '#6E4E46', INK, r=2)
    s.rrect(274, 192, 290, 212, '#6E4E46', INK, r=2)
    save(s, 'low_closet/micro_idle.png')


def bed():
    """Futón legado con almohada (340 × 490, pivote [160, 350])."""
    s = Sprite(340, 490)
    s.rect(36, 56, 326, 474, (0x6E, 0x5F, 0x86, 80))
    s.rrect(20, 40, 306, 446, '#FFF1D8', INK, r=10)
    s.rect(28, 48, 298, 56, '#FFFBF2')
    s.rect(296, 48, 302, 440, '#D9CBB9')
    s.rect(28, 434, 298, 442, '#D9CBB9')
    # almohada
    s.rrect(80, 60, 250, 130, '#F7EEDD', '#7C6A95', r=14)
    s.rect(90, 90, 240, 102, '#9A8BE3')
    s.hline(90, 240, 92, '#B9ADEB')
    # edredón
    s.rrect(30, 150, 296, 436, '#9A8BE3', INK, r=10)
    s.rect(38, 158, 288, 178, '#C3B9F0')
    for y in range(210, 415, 56):
        for x in range(64, 275, 56):
            ox = 28 if ((y - 210) // 56) % 2 else 0
            fx, fy = x + ox, y
            if fx < 270:
                s.ell(fx - 10, fy - 10, fx + 10, fy + 10, '#FFC2CD')
                s.ell(fx - 4, fy - 4, fx + 4, fy + 4, '#FF9CAD')
    save(s, 'bed/micro_idle.png')


# ------------------------------------------------------------ Living

def rug():
    """Alfombra seigaiha tejida (710 × 510, pivote [15, 495])."""
    s = Sprite(710, 510)
    x, y, w, h = 15, 15, 680, 480
    base, ring, ring2 = '#B997C2', '#D6BCD8', '#9F7FAE'
    s.rrect(x, y, x + w - 1, y + h - 1, '#8E6E9F', '#6F5684', r=8)
    s.rect(x + 14, y + 14, x + w - 15, y + h - 15, base)
    # patrón de olas seigaiha a resolución micro
    for row in range(0, (h - 32) // 48 + 2):
        cy = y + 20 + row * 48
        off = 32 if row % 2 else 0
        for cx in range(x + 20 + off, x + w - 20, 64):
            for r, col in ((28, ring2), (20, ring), (12, ring2)):
                for a in range(180, 361, 8):
                    rad = math.radians(a)
                    px_ = int(round(cx + math.cos(rad) * r))
                    py_ = int(round(cy + math.sin(rad) * r))
                    if x + 16 <= px_ < x + w - 16 and y + 16 <= py_ < y + h - 16:
                        s.rect(px_, py_, px_ + 1, py_ + 1, col)
    s.rrect(x + 14, y + 14, x + w - 15, y + h - 15, None, '#A889B5', r=4)
    s.rrect(x, y, x + w - 1, y + h - 1, None, '#6F5684', r=8)
    # flecos finos en bordes norte y sur
    for i in range(x + 16, x + w - 16, 12):
        s.rect(i, y - 10, i + 3, y - 1, '#D6BCD8')
        s.rect(i, y + h, i + 3, y + h + 9, '#D6BCD8')
    save(s, 'rug/micro_idle.png')


def tv_stand():
    """Mueble bajo de TV (470 × 220, pivote [15, 180])."""
    s = Sprite(470, 220)
    s.rect(24, 170, 460, 210, INK_SOFT)
    s.rrect(15, 20, 450, 140, '#C897AD', INK, r=6)
    s.rect(20, 25, 445, 32, '#E2B8C7')
    for y in range(44, 134, 16):
        s.hline(22, 444, y, '#BD8CA2')
    # frente con puertas y hueco central
    s.rect(15, 140, 450, 190, '#A67A91')
    s.rrect(15, 140, 450, 190, None, INK, r=1)
    s.rrect(36, 152, 162, 180, '#966C83', '#78526A', r=3)
    s.rrect(304, 152, 430, 180, '#966C83', '#78526A', r=3)
    s.rrect(178, 152, 288, 180, '#4E4163', INK, r=3)
    s.rect(86, 162, 114, 168, '#E2B8C7')
    s.rect(352, 162, 380, 168, '#E2B8C7')
    save(s, 'tv_stand/micro_idle.png')


def apt_tv():
    """Televisor independiente (330 × 190, pivote [165, 160])."""
    s = Sprite(330, 190)
    s.ell(86, 148, 244, 180, INK_SOFT)
    s.rrect(118, 138, 212, 166, '#585979', INK, r=4)
    s.rect(124, 142, 206, 148, '#7A7C9E')
    s.rrect(14, 10, 316, 146, '#3F365C', r=10)
    s.rrect(26, 22, 304, 134, '#698BA1', r=4)
    # bandas de reflejo en pantalla
    for i in range(60):
        s.hline(max(28, 170 - i), min(302, 210 - i), 28 + i, '#8CABBA')
    s.rect(32, 28, 140, 36, '#BDD7DB')
    s.ell(284, 122, 294, 132, '#9CE0D6')
    save(s, 'apt_tv/micro_idle.png')


def tv_legacy():
    """Televisor con mueble legado (360 × 290, pivote [180, 240])."""
    s = Sprite(360, 290)
    s.rect(20, 245, 352, 282, INK_SOFT)
    s.rrect(10, 170, 346, 260, '#B2839A', INK, r=6)
    s.rect(16, 176, 340, 184, '#D5A5AF')
    s.rrect(28, 200, 168, 246, '#AC859D', '#7C5D72', r=4)
    s.rrect(188, 200, 328, 246, '#AC859D', '#7C5D72', r=4)
    s.rect(140, 150, 216, 172, '#585979')
    s.rrect(30, 10, 326, 154, '#3F365C', r=10)
    s.rrect(44, 24, 312, 140, '#698BA1', r=4)
    s.ell(296, 128, 306, 138, '#9CE0D6')
    save(s, 'tv/micro_idle.png')


def plant():
    """Planta en maceta (230 × 280, pivote [115, 250])."""
    s = Sprite(230, 280)
    cx, cy = 115, 250
    s.ell(cx - 72, cy - 10, cx + 72, cy + 22, INK_SOFT)
    # maceta de terracota
    s.rrect(cx - 50, cy - 82, cx + 50, cy + 6, '#D98E7E', INK, r=10)
    s.rect(cx - 42, cy - 76, cx - 28, cy - 8, '#F0B3A2')
    s.rect(cx + 34, cy - 76, cx + 44, cy - 8, '#B9706A')
    s.rrect(cx - 60, cy - 96, cx + 60, cy - 78, '#D98E7E', INK, r=6)
    s.ell(cx - 46, cy - 96, cx + 46, cy - 72, '#6E4E46')
    # hojas grandes con nervadura
    leaves = [
        (-16, -230, 34, 118), (-72, -192, 36, 92), (34, -200, 36, 102),
        (-98, -142, 36, 68), (62, -150, 36, 74),
    ]
    for lx, ly, lw, lh in leaves:
        s.rrect(cx + lx, cy + ly, cx + lx + lw, cy + ly + lh, '#6FB5A3', '#3E6E73', r=12)
        s.rect(cx + lx + lw // 2 - 2, cy + ly + 10, cx + lx + lw // 2 + 2, cy + ly + lh - 12, '#9CE0D6')
        s.ell(cx + lx + 6, cy + ly + 8, cx + lx + 16, cy + ly + 22, '#C6F0EA')
    save(s, 'plant/micro_idle.png')


# ------------------------------------------------------------ Cocina

def sink():
    """Bacha encastrada de acero (220 × 220, pivote [10, 210])."""
    s = Sprite(220, 220)
    s.rrect(10, 10, 209, 209, '#C5D8DB', '#6F8E96', r=10)
    s.rrect(26, 26, 193, 193, '#9CB8BE', '#5D7C86', r=8)
    s.rect(32, 32, 187, 38, '#D5E7EA')
    s.rect(32, 32, 38, 187, '#D5E7EA')
    # cepillado metálico
    for y in range(44, 186, 8):
        s.hline(40, 184, y, '#A9C3C9')
    # desagüe con rejilla
    s.ell(92, 92, 128, 128, '#58707A')
    s.ring(92, 92, 128, 128, INK)
    for dx in (-10, 0, 10):
        s.vline(110 + dx, 98, 122, '#3F365C')
    # canilla cromada a la izquierda
    s.rrect(16, 82, 36, 138, '#7A96A0', INK, r=4)
    s.rrect(22, 98, 86, 122, '#9FB8C0', INK, r=6)
    s.rect(28, 102, 80, 108, '#E9F8FB')
    save(s, 'sink/micro_idle.png')


def stove():
    """Anafe de 2 hornallas con pava menta (220 × 260, pivote [10, 250])."""
    s = Sprite(220, 260)
    s.rrect(10, 10, 209, 249, '#E9E5E0', '#7C7686', r=10)
    s.rect(18, 18, 201, 24, '#FFFFFF')
    for cy in (70, 180):
        s.ell(52, cy - 46, 148, cy + 46, '#484456')
        s.ell(64, cy - 34, 136, cy + 34, '#2E2A3A')
        s.ring(52, cy - 46, 148, cy + 46, INK)
        s.rect(48, cy - 3, 152, cy + 3, '#3E3A4A')
        s.rect(97, cy - 48, 103, cy + 48, '#3E3A4A')
        s.ell(84, cy - 16, 116, cy + 16, '#6E687C')
    # pava menta sobre la hornalla superior
    s.ell(62, 32, 138, 108, '#9CE0D6')
    s.ring(62, 32, 138, 108, INK)
    s.ell(78, 46, 122, 90, '#C6F0EA')
    s.ell(94, 62, 106, 74, '#3F365C')
    s.rrect(134, 62, 162, 78, '#78BEB4', INK, r=3)  # pico
    # perillas a la derecha
    for ky in (56, 118, 180):
        s.ell(168, ky - 12, 192, ky + 12, '#6C6678')
        s.ring(168, ky - 12, 192, ky + 12, INK)
        s.vline(180, ky - 10, ky + 10, '#E9E5E0')
    save(s, 'stove/micro_idle.png')


def rice_cooker():
    """Arrocera crema con panel lila (130 × 130, pivote [65, 100])."""
    s = Sprite(130, 130)
    s.ell(14, 92, 116, 124, INK_SOFT)
    s.rrect(14, 14, 115, 115, '#FFF1D8', INK, r=24)
    s.rect(32, 22, 96, 30, '#FFFFFF')
    s.ell(38, 34, 92, 78, '#EFE0C6')
    s.ring(38, 34, 92, 78, '#D6C4A9')
    # válvula de vapor
    s.ell(58, 48, 72, 62, '#8E7D74')
    s.ell(62, 52, 68, 58, '#5C4F4A')
    # panel frontal lila con botones
    s.rrect(38, 84, 92, 106, '#9A8BE3', INK, r=5)
    s.rect(46, 90, 68, 100, '#D6CEF5')
    s.ell(74, 91, 84, 101, '#FF9CAD')
    save(s, 'rice_cooker/micro_idle.png')


def dish_shelf():
    """Repisa de vajilla en pared oeste (110 × 320, pivote [20, 160])."""
    s = Sprite(110, 320)
    s.rect(20, 20, 102, 310, INK_SOFT)
    s.rrect(10, 10, 88, 308, WOOD, INK, r=8)
    s.rect(18, 18, 28, 300, WOOD_L)
    s.rect(76, 18, 84, 300, WOOD_D)
    for cy, col, hi in (
        (60, '#FF9CAD', '#FFD0D9'),
        (130, '#9CE0D6', '#D2F5F0'),
        (200, '#9A8BE3', '#CBC2F5'),
        (264, '#FFF1D8', '#FFFFFF'),
    ):
        s.ell(24, cy - 26, 76, cy + 26, col)
        s.ring(24, cy - 26, 76, cy + 26, INK)
        s.ell(34, cy - 16, 66, cy + 16, hi)
    save(s, 'dish_shelf/micro_idle.png')


def fridge():
    """Heladera con imanes (230 × 350, pivote [110, 160])."""
    s = Sprite(230, 350)
    s.rect(28, 310, 224, 344, INK_SOFT)
    s.rrect(10, 10, 208, 324, '#C0D7CB', INK, r=14)
    s.rect(18, 18, 36, 314, '#D5E4D9')
    s.rect(192, 20, 202, 314, '#9CB5B2')
    # división freezer / heladera
    s.rect(14, 140, 204, 146, '#75989D')
    s.rect(14, 147, 204, 152, '#ADCCC3')
    # manijas
    s.rrect(164, 48, 180, 114, '#587084', INK, r=4)
    s.rrect(164, 178, 180, 272, '#587084', INK, r=4)
    # rejilla inferior
    s.rect(36, 296, 182, 312, '#678B94')
    # imanes y nota
    s.rrect(44, 36, 64, 56, '#FF9CAD', INK, r=3)
    s.rrect(84, 54, 104, 74, '#9A8BE3', INK, r=3)
    s.rrect(42, 178, 104, 232, '#F7EEDD', '#DDCBB4', r=4)
    s.rect(52, 194, 88, 198, '#9A8BE3')
    s.rect(52, 208, 78, 212, '#ECABAF')
    s.ell(118, 228, 142, 252, '#FFC86B')
    save(s, 'fridge/micro_idle.png')


# ------------------------------------------------------------ Baño

def washbasin():
    """Lavamanos de cerámica (240 × 300, pivote [10, 290])."""
    s = Sprite(240, 300)
    s.rect(2, 26, 16, 290, INK_SOFT)
    s.rrect(10, 10, 229, 289, '#F5F9FA', INK, r=12)
    s.rect(18, 18, 220, 26, '#FFFFFF')
    # pileta ovalada
    s.ell(34, 48, 178, 242, '#D5EAEE')
    s.ring(34, 48, 178, 242, '#8FB2BC')
    s.ell(96, 132, 124, 160, '#6A8B96')
    # canilla cromada contra la pared este
    s.rrect(170, 132, 218, 158, '#8DA4AC', INK, r=5)
    s.rect(176, 138, 212, 144, '#FFFFFF')
    # vaso rosa y jabón
    s.ell(174, 38, 208, 72, '#FF9CAD')
    s.ring(174, 38, 208, 72, INK)
    s.rrect(174, 226, 208, 256, '#FFF1D8', '#C9B59E', r=6)
    save(s, 'washbasin/micro_idle.png')


def toilet():
    """Inodoro japonés (220 × 260, pivote [10, 250])."""
    s = Sprite(220, 260)
    s.ell(20, 36, 210, 252, INK_SOFT)
    # tanque
    s.rrect(20, 170, 198, 248, '#EDF4F6', INK, r=10)
    s.rect(28, 178, 190, 188, '#FFFFFF')
    s.rrect(170, 190, 190, 206, '#9CB5BC', INK, r=3)
    # taza y tapa ovalada
    s.ell(34, 20, 184, 190, '#F5F9FA')
    s.ring(34, 20, 184, 190, INK)
    s.ell(54, 40, 164, 168, '#E1EEF2')
    s.ring(54, 40, 164, 168, '#B8CDD4')
    # panel del washlet
    s.rrect(174, 108, 204, 162, '#DCE8EC', INK, r=5)
    s.ell(182, 120, 194, 132, '#9CE0D6')
    s.ell(182, 140, 194, 152, '#FF9CAD')
    save(s, 'toilet/micro_idle.png')


def shower_tray():
    """Plato de ducha en zona húmeda (400 × 400, pivote [10, 390])."""
    s = Sprite(400, 400)
    s.rrect(10, 10, 389, 389, '#CBE0E6', INK, r=14)
    s.rrect(28, 28, 371, 371, '#B7D5DE', '#7FA4B0', r=10)
    # cuadrícula antideslizante sutil
    for k in range(48, 360, 24):
        s.hline(34, 365, k, '#ABCAD4')
        s.vline(k, 34, 365, '#ABCAD4')
    # líneas de pendiente hacia el desagüe
    s.line([(36, 36), (170, 170)], '#8FB4C0')
    s.line([(363, 36), (230, 170)], '#8FB4C0')
    s.line([(36, 363), (170, 230)], '#8FB4C0')
    s.line([(363, 363), (230, 230)], '#8FB4C0')
    # rejilla del desagüe
    s.rrect(174, 174, 226, 226, '#7E9AA6', INK, r=4)
    for y in range(184, 220, 10):
        s.rect(182, y, 218, y + 4, '#4E6773')
    # duchador en la pared este
    s.rrect(328, 170, 376, 224, '#8DA4AC', INK, r=5)
    s.ell(288, 174, 338, 220, '#D5E7EA')
    s.ring(288, 174, 338, 220, '#58707A')
    save(s, 'shower_tray/micro_idle.png')


# ------------------------------------------------------------ Entrada (genkan)

def shoe_rack():
    """Zapatero en pared oeste (140 × 340, pivote [10, 330])."""
    s = Sprite(140, 340)
    s.rect(120, 28, 136, 334, INK_SOFT)
    s.rrect(10, 10, 126, 326, '#C89474', INK, r=8)
    s.rect(18, 18, 28, 318, '#E2B896')
    s.rect(112, 18, 120, 318, '#9E6E5C')
    s.rect(18, 166, 118, 172, '#9E6E5C')
    s.rect(18, 173, 118, 178, '#E2B896')
    # bandeja de llaves y pequeña maceta
    s.rrect(34, 48, 98, 112, '#F7EEDD', '#B9A490', r=10)
    s.rrect(56, 72, 76, 86, '#FFCA8A', '#B98842', r=3)
    s.ell(40, 218, 92, 270, '#D98E7E')
    s.ring(40, 218, 92, 270, INK)
    s.ell(50, 226, 82, 258, '#6FB5A3')
    s.ell(58, 234, 74, 250, '#9CE0D6')
    save(s, 'shoe_rack/micro_idle.png')


# ------------------------------------------------------------ Muebles transitables (micro completo)

def chair():
    """Silla de madera con cojín capitoné y respaldo (200 × 240, pivote [20, 180])."""
    s = Sprite(200, 240)
    wood_d, wood, wood_l = '#8A5A44', '#B57B5A', '#D19B76'
    seat_d, seat, seat_l = '#C48A6A', '#E2AA86', '#F2C5A4'
    s.ell(10, 188, 190, 236, INK_SOFT)
    # patas visibles abajo
    for x in (30, 150):
        s.rrect(x, 160, x + 16, 224, wood_d, INK, r=2)
        s.rect(x + 3, 162, x + 7, 220, wood)
    # asiento tapizado (160 × 160: x 20..179, y 20..179)
    s.rrect(20, 20, 179, 179, seat, INK, r=10)
    s.rect(28, 28, 171, 42, seat_l)
    s.rect(28, 28, 38, 171, seat_l)
    s.rect(28, 164, 171, 173, seat_d)
    # trama tejida suave y costuras del tapizado
    for y in range(44, 162, 16):
        s.hline(32, 168, y, '#D69C78')
    for x in range(44, 162, 16):
        s.vline(x, 36, 164, '#D69C78')
    # botones del capitoné
    for bx, by in ((60, 68), (138, 68), (60, 128), (138, 128)):
        s.ell(bx - 6, by - 6, bx + 6, by + 6, seat_d)
        s.ell(bx - 3, by - 3, bx + 3, by + 3, '#A66F52')
        s.px(bx - 1, by - 1, seat_l)
    # respaldo de madera en el borde sur (mira al norte, hacia el escritorio)
    s.rrect(18, 142, 181, 190, wood, INK, r=8)
    s.rect(26, 148, 173, 156, wood_l)
    s.rect(26, 178, 173, 184, wood_d)
    for x in (56, 98, 140):
        s.rrect(x, 154, x + 10, 178, wood_d, '#6E4532', r=2)
    save(s, 'chair/micro_idle.png')


def apt_sofa():
    """Sofá de 3 cuerpos mirando al sur (540 × 280, pivote [30, 230]): asiento 480 × 200."""
    s = Sprite(540, 280)
    fab_d, fab, fab_l = '#5A8578', '#78A899', '#99C7B8'
    s.ell(20, 212, 520, 272, INK_SOFT)
    for x in (50, 470):
        s.rrect(x, 228, x + 16, 258, '#8A5A44', INK, r=2)
    # respaldo al norte (contra pared norte)
    s.rrect(30, 10, 509, 78, fab_d, INK, r=12)
    s.rect(42, 18, 496, 30, fab)
    s.rect(42, 18, 496, 24, fab_l)
    # asiento de 3 almohadones (480 × 200: x 30..509, y 30..229)
    s.rrect(34, 58, 505, 229, fab, INK, r=12)
    s.rect(46, 66, 494, 78, fab_l)
    s.rect(44, 214, 496, 224, fab_d)
    # textura textil diagonal sutil
    for y in range(80, 214, 10):
        s.hline(46, 494, y, '#6F9F90')
    # divisiones de los 3 almohadones y botones
    for x in (190, 348):
        s.rect(x - 2, 66, x + 2, 222, fab_d)
        s.vline(x + 3, 66, 222, fab_l)
    for cx in (112, 270, 428):
        s.ell(cx - 5, 140, cx + 5, 150, fab_d)
        s.px(cx - 1, 143, fab_l)
    # almohadones decorativos (rosa y crema)
    s.rrect(62, 74, 138, 142, '#FF9CAD', '#C66F86', r=10)
    s.rect(72, 82, 128, 92, '#FFC2CD')
    s.rrect(396, 74, 472, 142, '#FFF1D8', '#C9B59E', r=10)
    s.rect(406, 82, 462, 92, '#FFFFFF')
    # apoyabrazos laterales
    s.rrect(10, 36, 52, 232, fab_d, INK, r=10)
    s.rect(18, 46, 28, 222, fab)
    s.rrect(486, 36, 528, 232, fab_d, INK, r=10)
    s.rect(494, 46, 504, 222, fab)
    save(s, 'apt_sofa/micro_idle.png')


def sofa_legacy():
    """Sofá de 3 cuerpos legado mirando al norte (560 × 300, pivote [40, 220])."""
    s = Sprite(560, 300)
    fab_d, fab, fab_l = '#5A8578', '#78A899', '#99C7B8'
    s.ell(20, 236, 540, 294, INK_SOFT)
    for x in (52, 490):
        s.rrect(x, 236, x + 16, 274, '#8A5A44', INK, r=2)
    s.rrect(40, 20, 519, 220, fab, INK, r=12)
    s.rect(52, 28, 506, 40, fab_l)
    for x in (200, 360):
        s.rect(x - 2, 28, x + 2, 210, fab_d)
        s.vline(x + 3, 28, 210, fab_l)
    for cx in (120, 280, 440):
        s.ell(cx - 5, 114, cx + 5, 124, fab_d)
    s.rrect(10, 46, 56, 236, fab_d, INK, r=10)
    s.rect(18, 56, 28, 226, fab)
    s.rrect(502, 46, 548, 236, fab_d, INK, r=10)
    s.rect(510, 56, 520, 226, fab)
    # respaldo al sur
    s.rrect(30, 188, 529, 258, fab_d, INK, r=10)
    s.rect(42, 196, 516, 208, fab)
    s.rect(42, 196, 516, 202, fab_l)
    save(s, 'sofa/micro_idle.png')


def apt_futon():
    """Futón completo sin almohada (520 × 760, pivote [20, 740]): 480 × 720."""
    s = Sprite(520, 760)
    s.rect(38, 38, 510, 750, (0x6E, 0x5F, 0x86, 75))
    # shikibuton (colchón crema)
    s.rrect(20, 20, 499, 739, '#FFF1D8', INK, r=14)
    s.rect(30, 30, 488, 42, '#FFFBF2')
    s.rect(484, 30, 494, 728, '#D9CBB9')
    s.rect(30, 724, 488, 734, '#D9CBB9')
    # costuras del colchón en cabecera
    for x in range(80, 460, 80):
        s.vline(x, 44, 204, '#E8DAC2')
    # kakebuton (edredón lila con doblez superior y flores de cerezo)
    s.rrect(34, 210, 484, 724, '#9A8BE3', INK, r=14)
    s.rect(44, 220, 474, 260, '#C3B9F0')
    s.rect(44, 258, 474, 266, '#B3A7EC')
    s.rect(468, 266, 478, 714, '#7F70C9')
    s.rect(44, 708, 474, 718, '#7F70C9')
    # costuras acolchadas del edredón
    for y in range(290, 700, 64):
        s.hline(48, 466, y, '#8D7ED6')
    for x in range(110, 450, 88):
        s.vline(x, 268, 706, '#8D7ED6')
    # flores de cerezo bordadas
    for y in range(306, 680, 80):
        for x in range(80, 440, 80):
            ox = 40 if ((y - 306) // 80) % 2 else 0
            fx, fy = x + ox, y
            if fx < 440:
                for dx, dy in ((-8, 0), (8, 0), (0, -8), (0, 8), (-6, -6), (6, 6)):
                    s.ell(fx + dx - 4, fy + dy - 4, fx + dx + 4, fy + dy + 4, '#FFC2CD')
                s.ell(fx - 4, fy - 4, fx + 4, fy + 4, '#FF9CAD')
    save(s, 'apt_futon/micro_idle.png')


def apt_pillow():
    """Almohada makura independiente (260 × 180, pivote [10, 170]): 240 × 160."""
    s = Sprite(260, 180)
    s.rrect(18, 22, 252, 174, INK_SOFT, r=20)
    s.rrect(10, 10, 249, 169, '#F7EEDD', INK, r=24)
    s.rect(28, 18, 230, 30, '#FFFFFF')
    s.rect(28, 152, 230, 162, '#DDCBB4')
    # trama de lino sutil
    for y in range(34, 150, 10):
        s.hline(20, 238, y, '#EFE4D0')
    # banda lila central del makura
    s.rect(24, 66, 234, 104, '#9A8BE3')
    s.rect(24, 70, 234, 76, '#B9ADEB')
    s.rect(24, 98, 234, 104, '#7C6A95')
    # costuras laterales
    s.vline(34, 24, 154, '#D9CBB4')
    s.vline(224, 24, 154, '#D9CBB4')
    save(s, 'apt_pillow/micro_idle.png')


def counter():
    """Mesada completa de cocina (280 × 1020, pivote [20, 980]): 240 × 960."""
    s = Sprite(280, 1020)
    # sombra en el piso hacia el este y sur
    s.rect(260, 36, 276, 1012, INK_SOFT)
    s.rect(32, 990, 276, 1016, INK_SOFT)
    # tablero de piedra clara (240 × 960: x 20..259, y 20..979)
    s.rrect(20, 20, 259, 979, '#EFF3EE', INK, r=8)
    s.rect(28, 28, 251, 971, '#EFF3EE')
    s.rect(28, 28, 36, 971, '#FFFFFF')
    s.rect(28, 28, 251, 36, '#FFFFFF')
    s.rect(248, 28, 256, 971, '#D1DBD4')
    # motas sutiles de granito claro
    for y in range(40, 964, 12):
        for x in range(40, 244, 12):
            n = noise(x, y, 53)
            if n < 24:
                s.rect(x, y, x + 2, y + 2, '#DCE3DC')
            elif n > 236:
                s.rect(x, y, x + 2, y + 2, '#FFFFFF')
    # frente del bajo mesada en el canto sur
    s.rect(20, 979, 259, 1002, '#7FB0A8')
    s.rrect(20, 979, 259, 1002, None, INK, r=1)
    s.rect(24, 982, 255, 986, '#5E8E88')
    # tiradores de cajones insinuados en el borde este
    for y in range(140, 950, 160):
        s.rect(236, y, 254, y + 8, '#BFD0C8')
    # cable lateral en el borde norte (acceso tiny desde el zócalo)
    s.rect(228, 0, 236, 22, '#5D566E')
    s.vline(230, 0, 22, '#7A728F')
    save(s, 'counter/micro_idle.png')


def main(src):
    global SRC
    SRC = src
    chair()
    apt_sofa()
    sofa_legacy()
    apt_futon()
    apt_pillow()
    counter()
    low_closet()
    bed()
    rug()
    tv_stand()
    apt_tv()
    tv_legacy()
    plant()
    sink()
    stove()
    rice_cooker()
    dish_shelf()
    fridge()
    washbasin()
    toilet()
    shower_tray()
    shoe_rack()
