"""Mano derecha de la giantess, palma hacia arriba, vista desde arriba.

Orientación norte: la muñeca y la manga entran por el este, los dedos apuntan
al oeste y el pulgar al norte. El cuadro mide 288 × 160 px, es decir
18 × 10 cm a 16 px por cm, igual que la extensión de la palma.

Cada parte (dedos, palma, pulgar, muñeca) es una forma con distancia con signo.
De esa distancia salen el contorno, las costuras entre partes y el sombreado de
celda: luz desde arriba a la izquierda, un tono de sombra en los bordes que
miran al sudeste y un brillo en los que miran al noroeste.
"""
import math

from PIL import Image

W, H = 288, 160

SKIN = (0xFF, 0xD8, 0xB9)
SKIN_PALM = (0xFC, 0xCD, 0xAE)
SKIN_HI = (0xFF, 0xEA, 0xD6)
SKIN_SH = (0xF4, 0xB4, 0x9A)
SKIN_SH2 = (0xE2, 0x9A, 0x8C)
CREASE = (0xE8, 0xA6, 0x96)
SEAM = (0xCF, 0x8C, 0x88)
OUTLINE = (0x84, 0x54, 0x6D)
INK = (0x3F, 0x36, 0x5C)
LILAC = (0x9A, 0x8B, 0xE3)
LILAC_HI = (0xC8, 0xB9, 0xF5)
LILAC_SH = (0x79, 0x69, 0xA4)
LILAC_SH2 = (0x67, 0x58, 0x8F)
CORAL = (0xFF, 0x9C, 0xAD)
CORAL_SH = (0xC6, 0x6F, 0x86)

LIGHT = (-0.6, -0.8)  # hacia la luz


class Capsule:
    def __init__(self, a, b, r):
        self.a, self.b, self.r = a, b, r

    def closest(self, x, y):
        ax, ay = self.a
        bx, by = self.b
        dx, dy = bx - ax, by - ay
        t = max(0.0, min(1.0, ((x - ax) * dx + (y - ay) * dy) / (dx * dx + dy * dy)))
        return ax + t * dx, ay + t * dy, t

    def sdf(self, x, y):
        cx, cy, _ = self.closest(x, y)
        return math.hypot(x - cx, y - cy) - self.r

    def normal(self, x, y):
        cx, cy, _ = self.closest(x, y)
        d = math.hypot(x - cx, y - cy) or 1
        return (x - cx) / d, (y - cy) / d


class RoundRect:
    def __init__(self, x0, y0, x1, y1, r):
        self.c = ((x0 + x1) / 2, (y0 + y1) / 2)
        self.h = ((x1 - x0) / 2 - r, (y1 - y0) / 2 - r)
        self.r = r

    def _q(self, x, y):
        return abs(x - self.c[0]) - self.h[0], abs(y - self.c[1]) - self.h[1]

    def sdf(self, x, y):
        qx, qy = self._q(x, y)
        return math.hypot(max(qx, 0), max(qy, 0)) + min(max(qx, qy), 0) - self.r

    def normal(self, x, y):
        e = 0.5
        gx = self.sdf(x + e, y) - self.sdf(x - e, y)
        gy = self.sdf(x, y + e) - self.sdf(x, y - e)
        d = math.hypot(gx, gy) or 1
        return gx / d, gy / d


class Ellipse:
    def __init__(self, cx, cy, rx, ry):
        self.cx, self.cy, self.rx, self.ry = cx, cy, rx, ry

    def sdf(self, x, y):
        u, v = (x - self.cx) / self.rx, (y - self.cy) / self.ry
        return (math.hypot(u, v) - 1) * min(self.rx, self.ry)

    def normal(self, x, y):
        u, v = (x - self.cx) / self.rx ** 2, (y - self.cy) / self.ry ** 2
        d = math.hypot(u, v) or 1
        return u / d, v / d


class Union:
    def __init__(self, *parts):
        self.parts = parts

    def sdf(self, x, y):
        return min(p.sdf(x, y) for p in self.parts)

    def normal(self, x, y):
        return min(self.parts, key=lambda p: p.sdf(x, y)).normal(x, y)


# Partes en orden de pintura: lo que va después tapa lo anterior.
FINGERS = [
    ('pinky', Capsule((118, 121), (54, 127), 9.5)),
    ('ring', Capsule((114, 101), (30, 104), 11)),
    ('middle', Capsule((112, 78), (17, 78), 11.5)),
    ('index', Capsule((114, 55), (34, 51), 11)),
]
PALM = Union(RoundRect(104, 46, 196, 134, 20), Ellipse(176, 68, 38, 26), Ellipse(176, 116, 34, 20),
             RoundRect(150, 56, 214, 128, 22))
WRIST = RoundRect(196, 58, 250, 124, 12)
THUMB = Union(Capsule((186, 58), (150, 30), 13), Capsule((150, 30), (122, 14), 11))
PARTS = FINGERS + [('wrist', WRIST), ('palm', PALM), ('thumb', THUMB)]
# La palma y la muñeca se sombrean con el contorno de toda la mano, así su
# borde contra los dedos no se ve como el canto de una pieza suelta.
HAND = Union(*[shape for _, shape in PARTS])

# El antebrazo sigue más allá del cuadro de la palma para que su corte quede
# fuera de pantalla cuando la palma está centrada.
EXTRA = 96
SLEEVE = RoundRect(226, 26, W + EXTRA + 40, 152, 12)


def bezier(p0, p1, p2, n=80):
    for i in range(n + 1):
        t = i / n
        yield ((1 - t) ** 2 * p0[0] + 2 * (1 - t) * t * p1[0] + t * t * p2[0],
               (1 - t) ** 2 * p0[1] + 2 * (1 - t) * t * p1[1] + t * t * p2[1])


def draw():
    """Mano y manga en un cuadro de (W + EXTRA) × H; la mano ocupa los primeros W px."""
    label = [[None] * W for _ in range(H)]
    for y in range(H):
        for x in range(W):
            for name, shape in PARTS:
                if shape.sdf(x + 0.5, y + 0.5) <= 0:
                    label[y][x] = name
    shapes = dict(PARTS)
    im = Image.new('RGBA', (W, H), (0, 0, 0, 0))
    px = im.load()

    def inside(x, y):
        return 0 <= x < W and 0 <= y < H and label[y][x] is not None

    for y in range(H):
        for x in range(W):
            name = label[y][x]
            if name is None:
                continue
            shape = HAND if name in ('palm', 'wrist') else shapes[name]
            d = -shape.sdf(x + 0.5, y + 0.5)
            nx, ny = shape.normal(x + 0.5, y + 0.5)
            facing = nx * LIGHT[0] + ny * LIGHT[1]
            col = SKIN_PALM if name == 'palm' else SKIN
            # la palma y la muñeca tienen una franja de sombra de celda más ancha
            band = 9 if name in ('palm', 'wrist') else 4
            if d < band and facing < -0.35:
                col = SKIN_SH2 if d < 1.6 else SKIN_SH
            elif d < 2.5 and facing > 0.55:
                col = SKIN_HI
            px[x, y] = col + (255,)

    # Brillo en la yema de cada dedo y del pulgar.
    for cap in [c for _, c in FINGERS] + [Capsule((150, 30), (122, 14), 11)]:
        bx, by = cap.b
        for y in range(int(by - cap.r) + 2, int(by - cap.r) + 4):
            for x in range(int(bx) - 2, int(bx) + 3):
                if inside(x, y):
                    px[x, y] = SKIN_HI + (255,)

    # Pliegues de las falanges, sin tocar los bordes del dedo.
    for name, cap in FINGERS:
        for t in (0.30, 0.62):
            cx = cap.a[0] + (cap.b[0] - cap.a[0]) * t
            cy = cap.a[1] + (cap.b[1] - cap.a[1]) * t
            for k in range(-int(cap.r) + 4, int(cap.r) - 3):
                x, y = int(cx), int(cy + k)
                if inside(x, y) and label[y][x] == name:
                    px[x, y] = CREASE + (255,)

    # Líneas de la palma: la de la vida rodea el pulgar, la del corazón baja hacia el meñique.
    for path in (bezier((150, 62), (170, 92), (206, 90)), bezier((118, 70), (126, 108), (152, 124)),
                 bezier((126, 64), (150, 74), (176, 100))):
        for x, y in path:
            x, y = int(x), int(y)
            if inside(x, y) and label[y][x] == 'palm' and -PALM.sdf(x + 0.5, y + 0.5) > 4:
                px[x, y] = CREASE + (255,)

    # Costuras entre partes y contorno exterior.
    out = im.copy()
    opx = out.load()
    order = {name: i for i, (name, _) in enumerate(PARTS)}
    for y in range(H):
        for x in range(W):
            name = label[y][x]
            if name is None:
                continue
            for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)):
                nx, ny = x + dx, y + dy
                if not inside(nx, ny):
                    opx[x, y] = OUTLINE + (255,)
                    break
                other = label[ny][nx]
                if other == name or order[other] < order[name]:
                    continue
                # entre dedos y bajo el pulgar, costura marcada; entre dedo y
                # palma, el pliegue suave de la base del dedo
                fingers = {n for n, _ in FINGERS}
                if name in fingers and other in fingers or other == 'thumb':
                    opx[x, y] = SEAM + (255,)
                elif name in fingers and other == 'palm':
                    opx[x, y] = CREASE + (255,)
    wide = Image.new('RGBA', (W + EXTRA, H), (0, 0, 0, 0))
    wide.alpha_composite(out, (0, 0))
    sleeve(wide)
    return wide


def sleeve(im):
    px = im.load()
    for y in range(H):
        for x in range(im.width):
            d = -SLEEVE.sdf(x + 0.5, y + 0.5)
            if d < 0:
                continue
            col = LILAC
            if y > 124:
                col = LILAC_SH
            if y > 144:
                col = LILAC_SH2
            if d < 4 and y < 40:
                col = LILAC_HI
            # puño acanalado
            if x < 248:
                col = LILAC_HI if (x - 226) % 4 else LILAC
                if y > 128:
                    col = LILAC if (x - 226) % 4 else LILAC_SH
            if d < 1.2:
                col = INK
            px[x, y] = col + (255,)
    # sombra del puño sobre la manga y un pliegue de la tela
    for y in range(28, 150):
        if -SLEEVE.sdf(249.5, y + 0.5) > 1.2:
            px[248, y] = LILAC_SH + (255,)
    for t in range(40):
        x, y = 256 + t // 2, 50 + t
        if -SLEEVE.sdf(x + 0.5, y + 0.5) > 2:
            px[x, y] = LILAC_SH + (255,)
    cx, cy = 270, 106
    for y in range(cy - 5, cy + 6):
        for x in range(cx - 5, cx + 6):
            r = math.hypot(x - cx, y - cy)
            if r <= 4.6:
                col = CORAL
                if r > 3.4:
                    col = CORAL_SH
                elif x < cx and y < cy - 1:
                    col = (0xFF, 0xC8, 0xD2)
                px[x, y] = col + (255,)


def shadowless():
    return draw()


if __name__ == '__main__':
    import sys
    img = draw()
    img.save(sys.argv[1] if len(sys.argv) > 1 else 'hand.png')


def grid_rows(img=None, threshold=128):
    """Filas de la nav_grid de la palma (celdas de 1 cm) que siguen la silueta.

    La fila 0 es la de menor Y local (el extremo de la muñeca) y la columna 0
    la de menor X local (el lado del meñique). Una celda es transitable si al
    menos `threshold` de sus 256 píxeles son de la mano o la manga.
    """
    img = img or draw()
    rows = []
    for j in range(18):
        row = ''
        for i in range(10):
            x0, y0 = W - 16 * (j + 1), H - 16 * (i + 1)
            n = sum(1 for y in range(y0, y0 + 16) for x in range(x0, x0 + 16) if img.getpixel((x, y))[3])
            row += '.' if n >= threshold else '#'
        rows.append(row)
    return rows
