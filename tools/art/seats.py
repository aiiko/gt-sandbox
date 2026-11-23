"""Arte micro de silla común, sofá, asiento sentado y sombra al sentarse (M37; el macro está en macro.py).

- chair/micro_tile.png (160 × 160 px): tapizado de tela tramada del asiento (10 × 10 cm).
- sofa/micro_tile.png (160 × 160 px): tapizado mullido en tono salvia (10 × 10 cm).
- seat/micro_idle.png (384 × 384 px, pivote [192, 192]): falda tableada y regazo al sentarse.
- seat/micro_shadow_<dir>.png (384 × 384 px, pivote [192, 192]): sombra al sentarse.
"""
import math
import os

from PIL import Image, ImageDraw
from px import Canvas, hexc

SRC = None  # lo fija main
INK = '#3F365C'


def save(c, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    c.save(path)


def chair_micro_tile():
    size = 160
    base, light, dark = hexc('#E2AA86'), hexc('#ECC09E'), hexc('#CE9370')
    c = Canvas(size, size)
    px = c.im.load()
    for y in range(size):
        for x in range(size):
            # trama tejida suave
            if (x // 4 + y // 4) % 2 == 0:
                col = light if (x % 4 == 0 or y % 4 == 0) else base
            else:
                col = dark if (x % 4 == 3 or y % 4 == 3) else base
            px[x, y] = col
    # costura sutil
    for k in range(0, size, 40):
        c.hline(0, k, size, '#C08360')
        c.vline(k, 0, size, '#C08360')
    save(c, 'chair/micro_tile.png')


def sofa_micro_tile():
    size = 160
    base, light, dark = hexc('#78A899'), hexc('#8AB8A9'), hexc('#649284')
    c = Canvas(size, size)
    px = c.im.load()
    for y in range(size):
        for x in range(size):
            band = (x + y) % 8
            if band == 0:
                col = light
            elif band == 4:
                col = dark
            else:
                col = base
            px[x, y] = col
    # capitoné/costura suave
    c.hline(0, 0, size, '#568072')
    c.vline(0, 0, size, '#568072')
    save(c, 'sofa/micro_tile.png')


def seat_idle():
    """Asiento/regazo de la giantess (384 × 384 px): falda tableada violeta y caderas al sentarse."""
    size = 384
    cx, cy = size // 2, size // 2
    ink = hexc(INK)
    skirt, skirt_l, skirt_d, skirt_d2 = hexc('#67588F'), hexc('#8070AD'), hexc('#504273'), hexc('#3E335C')
    card, card_l, card_d = hexc('#9A8BE3'), hexc('#C8B9F5'), hexc('#7A6BC6')

    im = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    # Silueta orgánica de caderas y falda apoyada al sentarse (dos curvas de muslos/caderas abajo)
    d.ellipse([cx - 166, cy - 134, cx + 166, cy + 126], fill=skirt)
    d.ellipse([cx - 154, cy - 30, cx + 14, cy + 150], fill=skirt)
    d.ellipse([cx - 14, cy - 30, cx + 154, cy + 150], fill=skirt)

    # Sombreado cel-shading suave y tablas verticales/abiertas de la falda
    px = im.load()
    for y in range(size):
        x0, x1 = -1, -1
        for x in range(size):
            if px[x, y][3] > 0:
                if x0 < 0:
                    x0 = x
                x1 = x
        if x0 < 0 or x1 <= x0:
            continue
        span = float(x1 - x0)
        for x in range(x0, x1 + 1):
            if px[x, y][3] == 0:
                continue
            u = (x - x0) / span
            v = (y - (cy - 134)) / 284.0
            if u < 0.20 and v < 0.65:
                col = skirt_l
            elif u > 0.80 or v > 0.84:
                col = skirt_d2
            elif u > 0.62 or v > 0.68:
                col = skirt_d
            else:
                col = skirt
            # tablas suaves que se abren desde la cintura (arriba) hacia el ruedo (abajo)
            t = max(0.0, min(1.0, (y - (cy - 100)) / 240.0))
            pleat_w = 24.0 + 10.0 * t
            rel = (x - cx) / pleat_w
            frac = rel - math.floor(rel)
            if frac < 0.10:
                col = skirt_d2
            elif frac < 0.24 and u < 0.72 and v < 0.80:
                col = skirt_l
            px[x, y] = col

    # Detalles internos con alpha_composite (brillos suaves y pliegue central)
    ov = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    do = ImageDraw.Draw(ov)
    # Borde inferior del cárdigan lila en la cintura trasera (arriba)
    do.pieslice([cx - 166, cy - 142, cx + 166, cy - 10], start=185, end=355, fill=card)
    do.arc([cx - 162, cy - 138, cx + 162, cy - 14], start=195, end=295, fill=card_l, width=10)
    do.arc([cx - 166, cy - 134, cx + 166, cy - 6], start=188, end=352, fill=card_d, width=6)
    # Brillo suave sobre la curva izquierda y derecha de las caderas/muslos
    do.arc([cx - 130, cy - 36, cx - 22, cy + 72], start=195, end=325, fill=(*skirt_l[:3], 130), width=16)
    do.arc([cx + 22, cy - 36, cx + 124, cy + 72], start=205, end=320, fill=(*skirt_l[:3], 95), width=14)
    # Pliegue central suave entre los muslos/glúteos
    do.line([(cx, cy - 20), (cx, cy + 144)], fill=(*skirt_d2[:3], 210), width=4)
    mask = im.getchannel('A')
    ov.putalpha(Image.composite(ov.getchannel('A'), Image.new('L', (size, size), 0), mask))
    im.alpha_composite(ov)

    # Contorno limpio de 2 px
    src = im.load()
    out = im.copy()
    dst = out.load()
    for y in range(size):
        for x in range(size):
            if src[x, y][3] > 0:
                continue
            for dy in (-2, -1, 0, 1, 2):
                for dx in (-2, -1, 0, 1, 2):
                    if dx * dx + dy * dy <= 5:
                        nx, ny = x + dx, y + dy
                        if 0 <= nx < size and 0 <= ny < size and src[nx, ny][3] > 200:
                            dst[x, y] = ink
                            break
                else:
                    continue
                break
    path = os.path.join(SRC, 'seat/micro_idle.png')
    os.makedirs(os.path.dirname(path), exist_ok=True)
    out.save(path)


def seat_shadows():
    size = 384
    cx, cy = size // 2, size // 2
    col = hexc(INK, 255)
    c = Canvas(size, size)
    px = c.im.load()
    # silueta redondeada de la sombra al sentarse
    for y in range(size):
        for x in range(size):
            dx = (x - cx) / 168.0
            dy = (y - cy) / 152.0
            if dx * dx + dy * dy <= 1.0:
                px[x, y] = col
    for d in ('north', 'east', 'south', 'west'):
        save(c, f'seat/micro_shadow_{d}.png')


def main(src):
    global SRC
    SRC = src
    chair_micro_tile()
    sofa_micro_tile()
    seat_idle()
    seat_shadows()
