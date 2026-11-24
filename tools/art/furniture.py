"""Sprites independientes de mobiliario en macro (80 px/m) sobre fondo transparente.

Cada pieza de la tabla 20.2 tiene su propio recurso de arte con pivote alineado
a su huella física en tools/scenarios/planta.py (0,8 px por cm).
Luz desde arriba a la izquierda, contorno INK (#3F365C) y sombras suaves INK_SOFT.
"""
import math
import os

from shapes import Sprite

SRC = None

INK = '#3F365C'
INK_SOFT = (0x3F, 0x36, 0x5C, 70)
WOOD, WOOD_D, WOOD_L = '#C89474', '#9E6E5C', '#E2B896'


def save(s, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    s.save(path)


# ------------------------------------------------------------ Dormitorio

def apt_futon():
    """Futón sin almohada (104 × 152 px, pivote [4, 148]): huella 96 × 144 px (1,2 × 1,8 m)."""
    s = Sprite(104, 152)
    s.rect(8, 8, 102, 150, (0x6E, 0x5F, 0x86, 80))
    # shikibuton (colchón crema)
    s.rrect(4, 4, 99, 147, '#FFF1D8', INK, r=3)
    s.hline(6, 97, 6, '#FFFBF2')
    s.vline(98, 6, 145, '#D9CBB9')
    s.hline(6, 97, 145, '#D9CBB9')
    # kakebuton (edredón lila con flores de cerezo, dejando cabecera libre para la almohada)
    s.rrect(7, 42, 96, 144, '#9A8BE3', INK, r=3)
    s.rect(9, 44, 94, 52, '#C3B9F0')
    s.hline(9, 94, 53, '#B3A7EC')
    for y in range(62, 138, 16):
        for x in range(16, 90, 16):
            ox = 8 if ((y - 62) // 16) % 2 else 0
            fx, fy = x + ox, y
            if fx < 88:
                for dx, dy in ((0, 0), (-1, 0), (1, 0), (0, -1), (0, 1)):
                    s.px(fx + dx, fy + dy, '#FF9CAD' if (dx, dy) == (0, 0) else '#FFC2CD')
                for dx, dy in ((-2, 0), (2, 0), (0, -2), (0, 2)):
                    s.px(fx + dx, fy + dy, '#FFC2CD')
    s.vline(94, 54, 142, '#7F70C9')
    s.hline(9, 94, 142, '#7F70C9')
    save(s, 'apt_futon/macro_idle.png')


def apt_pillow():
    """Almohada independiente (52 × 36 px, pivote [2, 34]): huella 48 × 32 px (60 × 40 cm)."""
    s = Sprite(52, 36)
    s.rrect(4, 5, 50, 35, INK_SOFT, r=4)
    s.rrect(2, 2, 49, 33, '#F7EEDD', INK, r=5)
    s.hline(6, 45, 4, '#FFFFFF')
    s.hline(6, 45, 5, '#FFFBF2')
    s.hline(6, 45, 30, '#DDCBB4')
    s.hline(6, 45, 31, '#C9B59E')
    # banda lila central del makura
    s.rect(5, 13, 46, 20, '#9A8BE3')
    s.hline(5, 46, 14, '#B9ADEB')
    s.hline(5, 46, 20, '#7C6A95')
    save(s, 'apt_pillow/macro_idle.png')


def low_closet():
    """Placard bajo con puertas corredizas (100 × 52 px, pivote [2, 42]): huella 96 × 40 px."""
    s = Sprite(100, 52)
    s.rect(4, 38, 99, 50, INK_SOFT)
    # tapa superior
    s.rrect(2, 2, 97, 34, '#D5A07A', INK, r=2)
    s.hline(4, 95, 4, '#EFC2A1')
    s.vline(3, 4, 32, '#EFC2A1')
    for k, y in enumerate(range(10, 32, 7)):
        gx = 10 + (k * 23) % 48
        s.hline(gx, gx + 22, y, '#C08A66')
    # frente con dos puertas corredizas
    s.rect(2, 34, 97, 46, '#B57B5A')
    s.hline(2, 97, 34, INK)
    s.hline(2, 97, 46, INK)
    s.vline(2, 34, 46, INK)
    s.vline(97, 34, 46, INK)
    s.vline(49, 35, 45, '#8A5A44')
    s.vline(50, 35, 45, '#D19B76')
    # tiradores embutidos
    s.rect(42, 38, 45, 42, '#6E4E46')
    s.rect(54, 38, 57, 42, '#6E4E46')
    save(s, 'low_closet/macro_idle.png')


# ------------------------------------------------------------ Living

def apt_sofa():
    """Sofá de 3 cuerpos mirando al sur (108 × 56 px, pivote [6, 46]): asiento 96 × 40 px."""
    s = Sprite(108, 56)
    fab_d, fab, fab_l = '#5A8578', '#78A899', '#99C7B8'
    s.ell(4, 42, 104, 54, INK_SOFT)
    for x in (10, 94):
        s.rect(x, 46, x + 3, 51, '#8A5A44')
    # respaldo en el borde norte (contra la pared norte, mirando al sur)
    s.rrect(6, 2, 101, 16, fab_d, INK, r=3)
    s.hline(9, 98, 4, fab)
    s.hline(9, 98, 5, fab_l)
    # asiento de tres almohadones
    s.rrect(8, 13, 99, 46, fab, INK, r=3)
    s.hline(11, 96, 15, fab_l)
    s.hline(10, 97, 44, fab_d)
    for x in (38, 69):
        s.vline(x, 15, 44, fab_d)
        s.vline(x + 1, 15, 44, fab_l)
    # almohadones decorativos (rosa y crema)
    s.rrect(13, 16, 27, 28, '#FF9CAD', '#C66F86', r=2)
    s.hline(15, 25, 17, '#FFC2CD')
    s.rrect(79, 16, 93, 28, '#FFF1D8', '#C9B59E', r=2)
    s.hline(81, 91, 17, '#FFFFFF')
    # apoyabrazos laterales
    s.rrect(2, 8, 11, 46, fab_d, INK, r=2)
    s.vline(4, 10, 44, fab)
    s.rrect(96, 8, 105, 46, fab_d, INK, r=2)
    s.vline(98, 10, 44, fab)
    save(s, 'apt_sofa/macro_idle.png')


def tv_stand():
    """Mueble bajo de TV (94 × 44 px, pivote [3, 36]): huella 88 × 32 px."""
    s = Sprite(94, 44)
    s.rect(5, 34, 92, 42, INK_SOFT)
    # tapa
    s.rrect(3, 4, 90, 28, '#C897AD', INK, r=2)
    s.hline(5, 88, 6, '#E2B8C7')
    s.vline(4, 6, 26, '#E2B8C7')
    # frente con dos puertas y estante central
    s.rect(3, 28, 90, 38, '#A67A91')
    s.hline(3, 90, 28, INK)
    s.hline(3, 90, 38, INK)
    s.vline(3, 28, 38, INK)
    s.vline(90, 28, 38, INK)
    s.rect(8, 31, 32, 36, '#966C83')
    s.rect(61, 31, 85, 36, '#966C83')
    s.rect(36, 31, 57, 36, '#584B6E')  # hueco para reproductor
    s.hline(18, 22, 33, '#E2B8C7')
    s.hline(71, 75, 33, '#E2B8C7')
    save(s, 'tv_stand/macro_idle.png')


def apt_tv():
    """Televisor independiente sobre el mueble bajo (66 × 38 px, pivote [33, 32])."""
    s = Sprite(66, 38)
    s.ell(18, 30, 48, 36, INK_SOFT)
    # pie
    s.rect(24, 28, 41, 33, '#585979')
    s.hline(25, 40, 28, '#7A7C9E')
    # marco y pantalla
    s.rrect(3, 2, 62, 29, '#3F365C', r=3)
    s.rrect(6, 5, 59, 26, '#698BA1', r=1)
    for i in range(13):
        s.hline(36 - i, 42 - i, 6 + i, '#8CABBA')
    s.rect(7, 6, 26, 7, '#BDD7DB')
    s.px(57, 25, '#9CE0D6')
    save(s, 'apt_tv/macro_idle.png')


def rug():
    """Alfombra seigaiha independiente (142 × 102 px, pivote [3, 99]): 136 × 96 px (1,7 × 1,2 m)."""
    s = Sprite(142, 102)
    x, y, w, h = 3, 3, 136, 96
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
    for i in range(x + 4, x + w - 4, 4):
        s.vline(i, y - 2, y - 1, '#D6BCD8')
        s.vline(i, y + h, y + h + 1, '#D6BCD8')
    save(s, 'rug/macro_idle.png')


def plant():
    """Planta en maceta independiente (46 × 56 px, pivote [23, 50])."""
    s = Sprite(46, 56)
    x, y = 23, 50
    s.ell(x - 14, y - 1, x + 14, y + 4, INK_SOFT)
    s.rrect(x - 10, y - 16, x + 10, y + 1, '#D98E7E', INK, r=2)
    s.rect(x - 8, y - 15, x - 6, y - 2, '#F0B3A2')
    s.vline(x + 8, y - 14, y - 1, '#B9706A')
    s.rect(x - 12, y - 19, x + 12, y - 16, '#B9706A')
    s.hline(x - 12, x + 12, y - 19, INK)
    s.hline(x - 11, x + 11, y - 18, '#D98E7E')
    s.ell(x - 9, y - 19, x + 9, y - 15, '#6E4E46')
    leaves = [(-3, -46, 7, 23), (-14, -38, 7, 17), (7, -40, 7, 19), (-20, -28, 7, 12), (13, -30, 7, 13)]
    for lx, ly, lw, lh in leaves:
        s.rrect(x + lx, y + ly, x + lx + lw, y + ly + lh, '#6FB5A3', '#3E6E73', r=2)
        s.vline(x + lx + lw // 2, y + ly + 2, y + ly + lh - 3, '#9CE0D6')
        s.px(x + lx + 1, y + ly + 2, '#C6F0EA')
    save(s, 'plant/macro_idle.png')


# ------------------------------------------------------------ Cocina

def counter():
    """Mesada con bajo mesada contra la pared oeste (56 × 204 px, pivote [4, 196]): 48 × 192 px."""
    s = Sprite(56, 204)
    # sombra en el piso hacia el este y sur
    s.rect(52, 8, 55, 202, INK_SOFT)
    s.rect(6, 198, 55, 203, INK_SOFT)
    # tablero de piedra clara (48 × 192 px: x 4..51, y 4..195)
    s.rrect(4, 4, 51, 195, '#EFF3EE', INK, r=2)
    s.rect(6, 6, 49, 193, '#EFF3EE')
    s.vline(6, 6, 193, '#FFFFFF')
    s.hline(6, 49, 6, '#FFFFFF')
    s.vline(50, 6, 193, '#D1DBD4')
    # frente del bajo mesada (canto este y sur)
    s.rect(4, 195, 51, 200, '#7FB0A8')
    s.hline(4, 51, 195, '#5E8E88')
    s.hline(4, 51, 200, INK)
    s.vline(4, 195, 200, INK)
    s.vline(51, 195, 200, INK)
    # juntas de puertas y tiradores en el borde este
    for y in range(28, 190, 32):
        s.hline(46, 50, y, '#BFD0C8')
    # cable lateral en el extremo norte (acceso tiny desde el zócalo)
    s.vline(46, 0, 5, '#5D566E')
    s.vline(47, 0, 5, '#7A728F')
    save(s, 'counter/macro_idle.png')


def sink():
    """Bacha encastrada de acero (44 × 44 px, pivote [2, 42]): 40 × 40 px."""
    s = Sprite(44, 44)
    s.rrect(2, 2, 41, 41, '#BCD3D6', '#6F8E96', r=3)
    s.rrect(6, 6, 37, 37, '#9CB8BE', '#6F8E96', r=2)
    s.hline(8, 35, 8, '#D5E7EA')
    s.vline(8, 8, 35, '#D5E7EA')
    # desagüe
    s.ell(19, 19, 25, 25, '#58707A')
    s.px(22, 22, '#3F365C')
    # canilla sobre el lado oeste
    s.rect(4, 19, 15, 23, '#8DA4AC')
    s.hline(5, 14, 20, '#E9F8FB')
    s.rect(4, 17, 7, 25, '#6F8E96')
    save(s, 'sink/macro_idle.png')


def stove():
    """Anafe encastrado de 2 hornallas con pava menta (44 × 52 px, pivote [2, 50]): 40 × 48 px."""
    s = Sprite(44, 52)
    s.rrect(2, 2, 41, 49, '#E9E5E0', '#7C7686', r=2)
    s.hline(4, 39, 4, '#FFFFFF')
    for cy in (14, 36):
        s.ring(12, cy - 8, 28, cy + 8, '#3E3A4A')
        s.ring(15, cy - 5, 25, cy + 5, '#5C576A')
        s.hline(11, 29, cy, '#3E3A4A')
        s.vline(20, cy - 9, cy + 9, '#3E3A4A')
    # pava menta sobre la hornalla superior
    s.ell(13, 6, 27, 20, '#9CE0D6')
    s.ring(13, 6, 27, 20, INK)
    s.ell(17, 9, 23, 15, '#C6F0EA')
    s.px(20, 12, '#3F365C')
    s.rect(27, 11, 31, 14, INK)  # pico
    # perillas a la derecha
    for ky in (12, 24, 36):
        s.rect(34, ky, 37, ky + 3, '#6C6678')
    save(s, 'stove/macro_idle.png')


def rice_cooker():
    """Arrocera crema con panel lila (26 × 26 px, pivote [13, 20]): 24 × 24 px."""
    s = Sprite(26, 26)
    s.ell(3, 18, 23, 25, INK_SOFT)
    s.rrect(3, 3, 22, 22, '#FFF1D8', INK, r=5)
    s.hline(7, 18, 5, '#FFFFFF')
    s.ell(8, 7, 17, 15, '#EFE0C6')
    s.px(12, 10, '#8E7D74')  # válvula de vapor
    s.rect(9, 17, 16, 20, '#9A8BE3')  # panel
    s.px(15, 18, '#FF9CAD')
    save(s, 'rice_cooker/macro_idle.png')


def dish_shelf():
    """Repisa de vajilla anclada a la pared oeste (22 × 64 px, pivote [4, 32])."""
    s = Sprite(22, 64)
    s.rect(4, 4, 20, 62, INK_SOFT)
    s.rrect(2, 2, 17, 61, WOOD, INK, r=2)
    s.vline(4, 4, 59, WOOD_L)
    s.vline(16, 4, 59, WOOD_D)
    # cuencos y tazas apilados
    for cy, col in ((12, '#FF9CAD'), (26, '#9CE0D6'), (40, '#9A8BE3'), (52, '#FFF1D8')):
        s.ell(5, cy - 5, 15, cy + 5, col)
        s.ring(5, cy - 5, 15, cy + 5, INK)
        s.px(8, cy - 2, '#FFFFFF')
    save(s, 'dish_shelf/macro_idle.png')


# ------------------------------------------------------------ Baño

def washbasin():
    """Lavamanos de cerámica contra la pared este (48 × 60 px, pivote [2, 58]): 44 × 56 px."""
    s = Sprite(48, 60)
    s.rect(0, 6, 4, 58, INK_SOFT)
    s.rrect(2, 2, 45, 57, '#F5F9FA', INK, r=3)
    s.hline(4, 43, 4, '#FFFFFF')
    # pileta ovalada
    s.ell(7, 10, 35, 48, '#D5EAEE')
    s.ring(7, 10, 35, 48, '#8FB2BC')
    s.ell(19, 26, 25, 32, '#6A8B96')
    # canilla contra la pared este
    s.rect(34, 27, 43, 31, '#8DA4AC')
    s.hline(35, 42, 28, '#FFFFFF')
    # vaso y cepillo en la esquina
    s.ell(35, 8, 41, 14, '#FF9CAD')
    s.px(38, 10, '#FFFFFF')
    save(s, 'washbasin/macro_idle.png')


def toilet():
    """Inodoro japonés mirando al norte (44 × 52 px, pivote [2, 50]): 40 × 48 px."""
    s = Sprite(44, 52)
    s.ell(4, 8, 42, 50, INK_SOFT)
    # tanque al sur
    s.rrect(4, 34, 39, 49, '#EDF4F6', INK, r=2)
    s.hline(6, 37, 36, '#FFFFFF')
    s.rect(34, 38, 37, 41, '#9CB5BC')  # palanca
    # taza y tapa ovalada hacia el norte
    s.ell(7, 4, 36, 38, '#F5F9FA')
    s.ring(7, 4, 36, 38, INK)
    s.ell(11, 8, 32, 33, '#E1EEF2')
    s.hline(16, 27, 9, '#FFFFFF')
    # panel lateral del washlet
    s.rect(35, 22, 40, 32, '#DCE8EC')
    s.px(37, 25, '#9CE0D6')
    s.px(37, 29, '#FF9CAD')
    save(s, 'toilet/macro_idle.png')


def shower_tray():
    """Plato de ducha en zona húmeda (80 × 80 px, pivote [2, 78]): 76 × 76 px."""
    s = Sprite(80, 80)
    # reborde elevado que delimita el piso seco
    s.rrect(2, 2, 77, 77, '#CBE0E6', INK, r=3)
    s.rrect(6, 6, 73, 73, '#B7D5DE', '#7FA4B0', r=2)
    s.hline(8, 71, 8, '#E4F2F5')
    s.vline(8, 8, 71, '#E4F2F5')
    # líneas de pendiente hacia el desagüe
    s.line([(8, 8), (34, 34)], '#9FC2CE')
    s.line([(71, 8), (45, 34)], '#9FC2CE')
    s.line([(8, 71), (34, 45)], '#9FC2CE')
    s.line([(71, 71), (45, 45)], '#9FC2CE')
    # rejilla del desagüe
    s.rrect(35, 35, 44, 44, '#7E9AA6', '#4E6773', r=1)
    for y in (37, 40, 42):
        s.hline(37, 42, y, '#4E6773')
    # duchador en la pared este
    s.rect(66, 34, 75, 44, '#8DA4AC')
    s.ell(58, 35, 67, 43, '#D5E7EA')
    s.ring(58, 35, 67, 43, '#58707A')
    save(s, 'shower_tray/macro_idle.png')


# ------------------------------------------------------------ Entrada (genkan)

def shoe_rack():
    """Zapatero poco profundo contra la pared oeste (28 × 68 px, pivote [2, 66]): 24 × 64 px."""
    s = Sprite(28, 68)
    s.rect(24, 6, 27, 67, INK_SOFT)
    s.rrect(2, 2, 25, 65, '#C89474', INK, r=2)
    s.vline(4, 4, 63, '#E2B896')
    s.hline(4, 23, 4, '#E2B896')
    s.vline(23, 4, 63, '#9E6E5C')
    # divisiones de puertas
    s.hline(4, 23, 33, '#9E6E5C')
    s.hline(4, 23, 34, '#E2B896')
    # bandeja de llaves y pequeña maceta arriba
    s.rrect(7, 10, 19, 22, '#F7EEDD', '#B9A490', r=2)
    s.px(13, 16, '#FFCA8A')
    s.ell(8, 44, 18, 54, '#D98E7E')
    s.ell(10, 45, 16, 51, '#6FB5A3')
    save(s, 'shoe_rack/macro_idle.png')


def main(src):
    global SRC
    SRC = src
    apt_futon()
    apt_pillow()
    low_closet()
    apt_sofa()
    tv_stand()
    apt_tv()
    rug()
    plant()
    counter()
    sink()
    stove()
    rice_cooker()
    dish_shelf()
    washbasin()
    toilet()
    shower_tray()
    shoe_rack()
