"""Arte macro y micro de las prendas explorables (docs/specs/prendas.md):

- slipper/macro_idle.png (14x24 px, pivote [7, 12])
- slipper/micro_idle.png (208x432 px, pivote [104, 216]): 10x24 cm a 16 px/cm
- slipper/micro_donning.png (208x432 px, pivote [104, 216]): pie descendiendo
- cardigan_pocket/macro_idle.png (18x16 px, pivote [9, 8])
- cardigan_pocket/micro_idle.png (240x272 px, pivote [120, 136]): 12x14 cm a 16 px/cm
"""
import os

from PIL import Image, ImageDraw
from px import hexc

SRC = None  # lo fija main

INK = hexc('#3F365C')
SLIP_OUT = hexc('#D6879E')
SLIP_HI = hexc('#F7BCCB')
SLIP_D = hexc('#AD5E77')
SLIP_D2 = hexc('#85435B')
SOLE = hexc('#FFF3E0')
SOLE_D = hexc('#DFCEB8')
SOLE_STITCH = hexc('#C9AF96')
SOCK = hexc('#FFF1D8')
SOCK_D = hexc('#DCCFE0')

CARD = hexc('#9A8BE3')
CARD_L = hexc('#C8B9F5')
CARD_D = hexc('#7A6BC6')
CARD_D2 = hexc('#5C4D9C')
BOW = hexc('#FF9CAD')


def _save(im, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    im.save(path)


def _slipper_macro():
    im = Image.new('RGBA', (14, 24), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    # Sombra suave
    d.ellipse([2, 3, 12, 22], fill=(0x3F, 0x36, 0x5C, 80))
    # Cuerpo de pantufla rosa acolchada
    d.rounded_rectangle([2, 2, 11, 20], radius=4, fill=SLIP_OUT, outline=INK)
    # Plantilla interior en talón
    d.rounded_rectangle([4, 11, 9, 18], radius=2, fill=SOLE)
    # Empeine abombado en puntera
    d.rounded_rectangle([3, 3, 10, 11], radius=3, fill=SLIP_HI, outline=SLIP_D)
    d.line([(4, 4), (9, 4)], fill=SOLE)
    _save(im, 'slipper/macro_idle.png')


def _draw_slipper_micro(donning=False):
    w, h = 208, 432
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)

    # Sombra bajo la pantufla
    d.rounded_rectangle([20, 22, w - 16, h - 14], radius=72, fill=(0x3F, 0x36, 0x5C, 55))

    # Borde exterior mullido (cubre [-5..5] cm x [-12..12] cm -> [24..184] x [24..408])
    d.rounded_rectangle([18, 18, w - 18, h - 18], radius=70, fill=SLIP_OUT, outline=INK, width=3)
    d.rounded_rectangle([24, 24, w - 24, h - 24], radius=64, fill=SLIP_HI)
    d.rounded_rectangle([28, 30, w - 28, h - 26], radius=60, fill=SLIP_OUT)

    # Plantilla interior acolchada (talón y centro)
    d.rounded_rectangle([38, 36, w - 38, h - 36], radius=48, fill=SOLE, outline=SLIP_D, width=3)
    # Pespunte romboidal suave de la plantilla
    for yy in range(60, h - 46, 28):
        for xx in range(52, w - 52, 24):
            d.line([(xx - 8, yy), (xx, yy - 8), (xx + 8, yy), (xx, yy + 8), (xx - 8, yy)], fill=SOLE_D, width=1)
    for yy in range(56, h - 48, 18):
        d.point((44, yy), fill=SOLE_STITCH)
        d.point((w - 44, yy), fill=SOLE_STITCH)

    # Borde del talón (y < -8 cm -> parte superior en pantalla si +Y es norte,
    # o y < -8 cm -> py > 216 + 128 = 344 porque Y mundo crece hacia arriba y pantalla hacia abajo)
    # Nota: en camera.ToScreen, Y positivo apunta arriba (pantalla Y menor).
    # Puntera: y > +7 cm -> py < 216 - 112 = 104 (parte superior del sprite).
    # Talón / abertura: y < -8 cm -> py > 216 + 128 = 344 (parte inferior del sprite).
    # Cubierta del empeine y puntera (arriba, py = 22..196):
    vamp = [
        (24, 196),
        (24, 96),
        (46, 36),
        (w // 2, 22),
        (w - 46, 36),
        (w - 24, 96),
        (w - 24, 196),
        (w - 40, 174),
        (w // 2, 160),
        (40, 174),
    ]
    d.polygon(vamp, fill=SLIP_OUT, outline=INK)
    d.arc([28, 28, w - 28, 170], start=195, end=345, fill=SLIP_HI, width=12)
    # Costura del borde abierto del empeine
    d.arc([32, 152, w - 32, 208], start=190, end=350, fill=SLIP_D2, width=5)
    d.arc([36, 158, w - 36, 204], start=195, end=345, fill=SOLE, width=3)

    # Moñito decorativo en el empeine
    cx = w // 2
    d.polygon([(cx - 4, 126), (cx - 26, 114), (cx - 26, 138)], fill=BOW, outline=INK)
    d.polygon([(cx + 4, 126), (cx + 26, 114), (cx + 26, 138)], fill=BOW, outline=INK)
    d.ellipse([cx - 6, 120, cx + 6, 132], fill=SLIP_D, outline=INK)

    # Borde acogedor del talón (parte inferior, entrada)
    d.arc([32, h - 108, w - 32, h - 22], start=10, end=170, fill=SLIP_D, width=8)
    d.arc([38, h - 98, w - 38, h - 30], start=15, end=165, fill=SOLE_STITCH, width=3)

    if donning:
        # Sombra de pie / media descendiendo sobre la plantilla central
        overlay = Image.new('RGBA', (w, h), (0, 0, 0, 0))
        od = ImageDraw.Draw(overlay)
        od.rounded_rectangle([44, 108, w - 44, h - 84], radius=40, fill=(0x3F, 0x36, 0x5C, 125))
        od.rounded_rectangle([54, 124, w - 54, h - 102], radius=34, fill=(*SOCK[:3], 150), outline=(*SOCK_D[:3], 180), width=3)
        im.alpha_composite(overlay)

    return im


def _cardigan_pocket_macro():
    im = Image.new('RGBA', (18, 16), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    d.rounded_rectangle([2, 2, 16, 14], radius=3, fill=CARD, outline=INK)
    d.rounded_rectangle([4, 3, 14, 7], radius=1, fill=CARD_L)
    d.rounded_rectangle([5, 7, 13, 12], radius=2, fill=CARD_D, outline=CARD_D2)
    d.point((9, 5), fill=BOW)
    _save(im, 'cardigan_pocket/macro_idle.png')


def _cardigan_pocket_micro():
    w, h = 240, 272
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)

    # Fondo tejido lila del cárdigan (12x14 cm -> 192x224 px en [24..216] x [24..248])
    d.rounded_rectangle([18, 18, w - 18, h - 18], radius=24, fill=CARD, outline=INK, width=3)
    px = im.load()
    for y in range(24, h - 24):
        for x in range(24, w - 24):
            if px[x, y][3] == 0:
                continue
            ux, uy = x % 8, y % 8
            if ux in (1, 2) and uy in (1, 2, 3):
                px[x, y] = CARD_L
            elif ux == 5 and uy in (2, 3, 4):
                px[x, y] = CARD_D

    # Interior del bolsillo + pretina acanalada superior (entrada en parte superior)
    d.rounded_rectangle([32, 38, w - 32, h - 30], radius=18, fill=CARD_D, outline=CARD_D2, width=4)
    d.rounded_rectangle([32, 26, w - 32, 58], radius=8, fill=CARD_L, outline=INK, width=2)
    for x in range(40, w - 38, 8):
        d.line([(x, 30), (x, 54)], fill=CARD_D, width=2)
    # Pespunte lateral del bolsillo
    for y in range(66, h - 38, 12):
        d.line([(38, y), (38, y + 6)], fill=CARD_L, width=2)
        d.line([(w - 38, y), (w - 38, y + 6)], fill=CARD_L, width=2)

    _save(im, 'cardigan_pocket/micro_idle.png')


def main(src):
    global SRC
    SRC = src
    _slipper_macro()
    _save(_draw_slipper_micro(donning=False), 'slipper/micro_idle.png')
    _save(_draw_slipper_micro(donning=True), 'slipper/micro_donning.png')
    _cardigan_pocket_macro()
    _cardigan_pocket_micro()
