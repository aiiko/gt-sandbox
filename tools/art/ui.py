"""Piezas de la interfaz de partida: retrato de la giantess y símbolos de expresión.

El retrato es un busto de 48 × 48 px con seis expresiones (neutral, contenta,
hambrienta, sorprendida, somnolienta, curiosa) que comparten cabeza, pelo y ropa; solo
cambian cejas, ojos, boca y rubor. Los símbolos son plantillas de texto de
11 × 11 px con contorno ciruela, uno por señal, de formas distintas para que se
lean sin color.
"""
import math

from PIL import Image

INK = (0x3F, 0x36, 0x5C)
HAIR = (0x5D, 0x50, 0x8D)
HAIR_HI = (0x86, 0x77, 0xB7)
HAIR_SH = (0x49, 0x3F, 0x71)
SKIN = (0xFF, 0xD8, 0xB9)
SKIN_SH = (0xF4, 0xB4, 0x9A)
BLUSH = (0xFF, 0x9C, 0xAD)
LILAC = (0x9A, 0x8B, 0xE3)
LILAC_HI = (0xC8, 0xB9, 0xF5)
LILAC_SH = (0x79, 0x69, 0xA4)
CORAL = (0xFF, 0x9C, 0xAD)
CORAL_SH = (0xCE, 0x81, 0x90)
IVORY = (0xFF, 0xF1, 0xD8)
TEAL = (0x3D, 0x92, 0x95)
TEAL_HI = (0x76, 0xD8, 0xCB)
MOUTH = (0xB0, 0x52, 0x6A)
TONGUE = (0xFF, 0x9C, 0xAD)
DROP = (0x9C, 0xD8, 0xF0)
DROP_HI = (0xF0, 0xFB, 0xFF)
RED = (0xE8, 0x5A, 0x6E)

S = 48


def ellipse_in(x, y, cx, cy, rx, ry):
    return ((x + 0.5 - cx) / rx) ** 2 + ((y + 0.5 - cy) / ry) ** 2 <= 1


class Pic:
    def __init__(self, w, h):
        self.im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
        self.p = self.im.load()
        self.w, self.h = w, h

    def put(self, x, y, c):
        if 0 <= x < self.w and 0 <= y < self.h:
            self.p[x, y] = c + (255,)

    def get(self, x, y):
        if 0 <= x < self.w and 0 <= y < self.h:
            return self.p[x, y]
        return (0, 0, 0, 0)

    def fill(self, test, c):
        for y in range(self.h):
            for x in range(self.w):
                if test(x, y):
                    self.put(x, y, c)

    def outline(self, c=INK):
        """Contorno de un píxel alrededor de todo lo opaco."""
        pts = []
        for y in range(self.h):
            for x in range(self.w):
                if self.get(x, y)[3]:
                    continue
                if any(self.get(x + dx, y + dy)[3] for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1))):
                    pts.append((x, y))
        for x, y in pts:
            self.put(x, y, c)


def head(expr):
    """Busto de la giantess mirando al frente, con la expresión dada."""
    p = Pic(S, S)
    # pelo largo por detrás de los hombros
    p.fill(lambda x, y: ellipse_in(x, y, 24, 22, 17, 18) or (7 <= x <= 40 and 22 <= y <= 40), HAIR_SH)
    # hombros y cárdigan
    p.fill(lambda x, y: y >= 38 and ellipse_in(x, y, 24, 50, 20, 14), LILAC)
    p.fill(lambda x, y: y >= 38 and ellipse_in(x, y, 24, 50, 20, 14) and x < 14, LILAC_HI)
    p.fill(lambda x, y: y >= 44 and ellipse_in(x, y, 24, 50, 20, 14) and x > 34, LILAC_SH)
    # cuello y cuello de la blusa coral
    p.fill(lambda x, y: 21 <= x <= 26 and 32 <= y <= 39, SKIN_SH)
    for x in range(17, 31):
        y0 = 38 + abs(x - 23.5) // 4
        for y in range(int(y0), int(y0) + 3):
            p.put(x, y, CORAL if y < y0 + 2 else CORAL_SH)
    # cara
    p.fill(lambda x, y: ellipse_in(x, y, 24, 22, 11.5, 12.5) and y < 34, SKIN)
    p.fill(lambda x, y: ellipse_in(x, y, 24, 22, 11.5, 12.5) and y < 34 and not ellipse_in(x, y, 23, 20, 11, 12.5), SKIN_SH)
    # orejas
    for cx in (12, 36):
        p.fill(lambda x, y, cx=cx: ellipse_in(x, y, cx, 23, 2, 3), SKIN_SH)
    # flequillo con mechones
    p.fill(lambda x, y: ellipse_in(x, y, 24, 17, 15, 11) and y < 17, HAIR)
    for cx, w, depth in ((15, 4, 21), (20, 4, 19), (25, 5, 20), (30, 4, 18), (34, 3, 22)):
        for y in range(12, depth):
            half = w * (depth - y) / (depth - 12)
            for x in range(int(cx - half), int(cx + half) + 1):
                p.put(x, y, HAIR)
    # mechones laterales sobre las mejillas
    for side in (-1, 1):
        for y in range(14, 34):
            x0 = 24 + side * (11 + (y - 14) // 10)
            for k in range(3 if y < 30 else 2):
                p.put(x0 + side * k, y, HAIR)
    # brillo del pelo
    for x in range(15, 31):
        y = 9 + int(((x - 23) / 8) ** 2 * 2)
        if (x // 2) % 2 == 0:
            p.put(x, y, HAIR_HI)
    # lazo coral
    for (x, y) in ((33, 9), (34, 8), (35, 8), (36, 9), (35, 10), (34, 10), (37, 10), (38, 9), (39, 8), (39, 10), (38, 11)):
        p.put(x, y, CORAL)
    p.put(36, 10, CORAL_SH)
    face(p, expr)
    p.outline()
    return p.im


def eye_open(p, cx, look=0, wide=False):
    """Ojo anime: pestaña gruesa, iris alto con pupila y brillo."""
    top = 19 if wide else 20
    outer = -1 if cx < 24 else 1
    x0, x1 = cx - 3, cx + 2  # cinco columnas
    for y in range(top + 2, 28):
        for x in range(x0, x1):
            p.put(x, y, IVORY)
    for x in range(x0 - (1 if outer < 0 else 0), x1 + (1 if outer > 0 else 0)):
        p.put(x, top, INK)
        p.put(x, top + 1, INK)
    p.put(x0 - 1 if outer < 0 else x1, top + 2, INK)
    ix = x0 + 1 + look
    for y in range(top + 2, 28):
        for x in range(ix - (0 if wide else 1), ix + (2 if wide else 3)):
            if x0 <= x < x1:
                p.put(x, y, TEAL if y < 25 else TEAL_HI)
    if wide:
        p.put(ix, top + 4, INK)
        p.put(ix + 1, top + 4, INK)
    else:
        for x in range(ix, ix + 2):
            p.put(x, top + 2, INK)
            p.put(x, top + 3, (0x2C, 0x5E, 0x66))
    p.put(ix - (0 if wide else 1), top + 2, (0xFF, 0xFF, 0xFF))
    p.put(ix - (0 if wide else 1), top + 3, (0xFF, 0xFF, 0xFF))
    for x in range(x0, x1):
        p.put(x, 28, SKIN_SH)


def eye_closed_happy(p, cx):
    for x, y in ((cx - 3, 25), (cx - 2, 24), (cx - 1, 23), (cx, 23), (cx + 1, 24), (cx + 2, 25)):
        p.put(x, y, INK)
        p.put(x, y + 1, INK if x in (cx - 1, cx) else SKIN)


def eye_sleepy(p, cx):
    for x in range(cx - 4, cx + 3):
        p.put(x, 24, INK)
        p.put(x, 25, INK)
    for x in range(cx - 3, cx + 2):
        p.put(x, 26, TEAL)
        p.put(x, 27, TEAL_HI)
    p.put(cx - 2, 26, IVORY)


def brows(p, kind):
    for cx, side in ((19, -1), (29, 1)):
        for k in range(-2, 2):
            x = cx + k
            if kind == 'raised':
                y = 16
            elif kind == 'worried':
                y = 17 + (1 if (k * side) > 0 else 0)
            else:
                y = 17
            if p.get(x, y)[:3] == HAIR:
                continue
            p.put(x, y, HAIR_SH)


def blush(p, strong=False):
    for cx in (16, 32):
        for x in range(cx - 1, cx + 2):
            if x % 2 == 0 or strong:
                p.put(x, 29, BLUSH)


def face(p, expr):
    if expr == 'happy':
        brows(p, 'normal')
        eye_closed_happy(p, 19)
        eye_closed_happy(p, 29)
        blush(p, strong=True)
        for x in range(22, 27):
            p.put(x, 30, MOUTH)
        for x in range(23, 26):
            p.put(x, 31, MOUTH)
        p.put(24, 31, TONGUE)
    elif expr == 'hungry':
        brows(p, 'worried')
        eye_open(p, 19, look=1)
        eye_open(p, 29, look=1)
        blush(p)
        for x, y in ((22, 31), (23, 30), (24, 31), (25, 30), (26, 31)):
            p.put(x, y, MOUTH)
        # gota de baba en la comisura
        p.put(27, 32, DROP)
        p.put(27, 33, DROP)
    elif expr == 'surprised':
        brows(p, 'raised')
        eye_open(p, 19, wide=True)
        eye_open(p, 29, wide=True)
        for x, y in ((23, 30), (24, 30), (22, 31), (25, 31), (23, 32), (24, 32)):
            p.put(x, y, MOUTH)
        p.put(23, 31, TONGUE)
        p.put(24, 31, TONGUE)
    elif expr == 'curious':
        # mira hacia un costado con una ceja alzada y media sonrisa
        brows(p, 'normal')
        for k in range(-2, 2):
            if p.get(29 + k, 16)[:3] != HAIR:
                p.put(29 + k, 16, HAIR_SH)
            p.put(29 + k, 17, SKIN)
        eye_open(p, 19, look=-1)
        eye_open(p, 29, look=-1)
        blush(p)
        for x, y in ((22, 30), (23, 31), (24, 31), (25, 31), (26, 30)):
            p.put(x, y, MOUTH)
    elif expr == 'sleepy':
        brows(p, 'normal')
        eye_sleepy(p, 19)
        eye_sleepy(p, 29)
        blush(p)
        for x in range(23, 26):
            p.put(x, 31, MOUTH)
    else:
        brows(p, 'normal')
        eye_open(p, 19)
        eye_open(p, 29)
        blush(p)
        for x in range(23, 26):
            p.put(x, 31, MOUTH)


# Símbolos: 'X' tinta, 'o' color principal, 'h' brillo, 's' sombra.
EMOTES = {
    'spirit': ('''
...XXXXX...
..XhooooX..
.XhoXoXooX.
.XooooooosX
.XooXXXoosX
.XoooooossX
..XoooossX.
..XooossX..
...XossX...
....XsX....
.....X.....''', (0xF4, 0xF1, 0xFF), (0xFF, 0xFF, 0xFF), (0xC8, 0xB9, 0xF5)),
    'discover': ('''
....XXX....
....XoX....
....XoX....
....XoX....
....XoX....
....XsX....
....XXX....
...........
....XXX....
....XoX....
....XXX....''', RED, (0xFF, 0xC8, 0xD2), (0xB0, 0x40, 0x58)),
    'explore': ('''
..XXXXXX...
.XoooooX...
.XoXXXooX..
..X..XooX..
....XooX...
...XooX....
...XoX.....
...XXX.....
...XXX.....
...XoX.....
...XXX.....''', TEAL_HI, (0xD8, 0xF6, 0xF0), TEAL),
    'sweat': ('''
.....X.....
....XoX....
....XoX....
...XoooX...
...XhooX...
..XhoooX...
..XhooooX..
..XoooosX..
..XoooosX..
...XsssX...
....XXX....''', DROP, DROP_HI, (0x6A, 0xA8, 0xD0)),
    'sleep': ('''
.....XXXXX.
.....XoooX.
.......XoX.
.XXXXX.XXX.
.XoooXXoooX
...XoX.XXXX
..XoX......
.XoooX.....
.XXXXX.....
...........
...........''', LILAC_HI, IVORY, LILAC),
    'note': ('''
.....XXXX..
.....XooXX.
.....XoXoX.
.....XoXXX.
.....XoX...
.....XoX...
..XXXXoX...
.XoooooX...
.XooooXX...
..XXXX.....
...........''', CORAL, (0xFF, 0xC8, 0xD2), CORAL_SH),
    'look': ('''
...........
...........
...XXXXX...
.XXhooooXX.
XhooXXXoosX
XooXXhXXosX
XhooXXXoosX
.XXoooossX.
...XXXXX...
...........
...........''', IVORY, (0xFF, 0xFF, 0xFF), (0xD8, 0xC8, 0xB0)),
    'vein': ('''
...........
.XXX...XXX.
.XoX...XoX.
.XooX.XooX.
..XX...XX..
...........
..XX...XX..
.XooX.XooX.
.XoX...XoX.
.XXX...XXX.
...........''', RED, (0xFF, 0xC8, 0xD2), (0xB0, 0x40, 0x58)),
    'chat': ('''
...........
.XXXXXXXXX.
XhooooooooX
XoXoXoXoosX
XoooooooosX
.XXXoosXXX.
...XosX....
...XsX.....
...XX......
...........
...........''', IVORY, (0xFF, 0xFF, 0xFF), (0xD8, 0xC8, 0xB0)),
}


def emote(name):
    text, main, hi, sh = EMOTES[name]
    rows = text.strip('\n').split('\n')
    p = Pic(11, 11)
    cols = {'X': INK, 'o': main, 'h': hi, 's': sh}
    for y, row in enumerate(rows):
        for x, ch in enumerate(row):
            if ch in cols:
                p.put(x, y, cols[ch])
    return p.im


EXPRESSIONS = ['neutral', 'happy', 'hungry', 'surprised', 'sleepy', 'curious']


def main(save):
    for e in EXPRESSIONS:
        save(head(e), f'ui/portrait_{e}.png')
    for name in EMOTES:
        save(emote(name), f'ui/emote_{name}.png')


if __name__ == '__main__':
    import sys
    out = Image.new('RGBA', (S * 5 + 4 * 4, S + 16), (213, 160, 122, 255))
    for i, e in enumerate(EXPRESSIONS):
        out.alpha_composite(head(e), (i * (S + 4), 0))
    for i, name in enumerate(EMOTES):
        out.alpha_composite(emote(name), (i * 14, S + 3))
    out = out.resize((out.width * 4, out.height * 4), Image.NEAREST)
    out.save(sys.argv[1] if len(sys.argv) > 1 else 'ui.png')
