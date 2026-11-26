"""Arte macro y micro para los quick wins (QW2 y QW5).

QW2:
- sock/{macro_idle,micro_idle,micro_donning}.png
- hair_tie/{macro_idle,micro_idle}.png
- street_shoe/{macro_idle,micro_idle}.png
- genkan_bag/{macro_idle,micro_idle}.png
- trinket_coin/{macro_idle,micro_idle,micro_held}.png
"""
import os

from PIL import Image, ImageDraw
from px import Canvas, hexc

SRC = None  # lo fija main

INK = '#3F365C'
INK_T = hexc(INK)

SOCK = hexc('#FFF1D8')
SOCK_HI = hexc('#FFFFFF')
SOCK_SH = hexc('#E0D2BE')
SOCK_STRIPE = hexc('#9A8BE3')
SOCK_STRIPE_D = hexc('#7A6BC6')

TIE = hexc('#FF9CAD')
TIE_HI = hexc('#FFD2DC')
TIE_SH = hexc('#CE7589')

SHOE = hexc('#5C4540')
SHOE_HI = hexc('#85655E')
SHOE_D = hexc('#3B2A26')
SHOE_IN = hexc('#D8C4AA')

BAG = hexc('#6A988E')
BAG_HI = hexc('#96C2B8')
BAG_SH = hexc('#496E66')
BAG_STRAP = hexc('#D9A066')

COIN = '#FFD25A'
COIN_HI = '#FFF6C8'
COIN_SH = '#C88D2A'

MAT = hexc('#76C8B8')
MAT_HI = hexc('#A9E5D8')
MAT_SH = hexc('#4C998B')

RICE = '#FFF7E8'
RICE_HI = '#FFFFFF'
RICE_SH = '#DCC9AE'

POT = hexc('#C97B56')
POT_HI = hexc('#E59E79')
POT_SH = hexc('#945235')
SOIL = hexc('#5C4338')
SOIL_D = hexc('#443028')
LEAF = hexc('#5FA86D')
LEAF_HI = hexc('#8ED299')
LEAF_D = hexc('#3F7A4D')


def _save(im, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    im.save(path)


def _socks():
    # Macro (12x22, pivote [6, 11])
    im = Image.new('RGBA', (12, 22), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    d.ellipse([2, 3, 10, 20], fill=(0x3F, 0x36, 0x5C, 65))
    d.rounded_rectangle([2, 2, 9, 19], radius=3, fill=SOCK, outline=INK_T)
    d.rectangle([3, 15, 8, 17], fill=SOCK_STRIPE)
    d.rounded_rectangle([3, 3, 8, 7], radius=2, fill=SOCK_SH)
    _save(im, 'sock/macro_idle.png')

    # Micro (192x400, pivote [96, 200] -> 9x22 cm a 16 px/cm + margen 24 px)
    for donning in (False, True):
        w, h = 192, 400
        im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
        d = ImageDraw.Draw(im)
        d.rounded_rectangle([20, 20, w - 16, h - 14], radius=64, fill=(0x3F, 0x36, 0x5C, 50))
        d.rounded_rectangle([18, 18, w - 18, h - 18], radius=64, fill=SOCK, outline=INK_T, width=3)
        d.rounded_rectangle([28, 28, w - 28, h - 28], radius=54, fill=SOCK_SH)
        d.rounded_rectangle([36, 36, w - 36, h - 36], radius=46, fill=SOCK)
        # Rayas tejidas en el puño (abajo = talón/abertura)
        for yy in (h - 84, h - 62):
            d.rectangle([30, yy, w - 30, yy + 10], fill=SOCK_STRIPE, outline=SOCK_STRIPE_D)
        # Costura tejida en la puntera (arriba)
        d.arc([34, 28, w - 34, 150], start=195, end=345, fill=SOCK_HI, width=6)
        d.line([(48, 74), (w - 48, 74)], fill=SOCK_STRIPE_D, width=3)
        # Textura suave de punto
        for yy in range(104, h - 104, 24):
            for xx in range(52, w - 52, 22):
                d.line([(xx, yy), (xx + 5, yy + 6), (xx + 10, yy)], fill=SOCK_SH, width=2)
        if donning:
            ov = Image.new('RGBA', (w, h), (0, 0, 0, 0))
            od = ImageDraw.Draw(ov)
            od.rounded_rectangle([42, 92, w - 42, h - 92], radius=36, fill=(0x3F, 0x36, 0x5C, 120))
            im.alpha_composite(ov)
        _save(im, 'sock/micro_donning.png' if donning else 'sock/micro_idle.png')


def _hair_tie():
    # Macro (10x10, pivote [5, 5])
    im = Image.new('RGBA', (10, 10), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    d.ellipse([1, 2, 8, 9], outline=INK_T, fill=TIE)
    d.ellipse([3, 4, 6, 7], fill=(0, 0, 0, 0), outline=TIE_SH)
    d.polygon([(6, 2), (9, 0), (9, 4)], fill=TIE_HI, outline=INK_T)
    _save(im, 'hair_tie/macro_idle.png')

    # Micro (128x128, pivote [64, 64] -> 6x6 cm a 16 px/cm + margen 16 px)
    w, h = 128, 128
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    # Fondo de mechón suave bajo el coletero
    d.rounded_rectangle([14, 14, w - 14, h - 14], radius=46, fill=TIE_SH, outline=INK_T, width=3)
    d.ellipse([20, 20, w - 20, h - 20], fill=TIE, outline=INK_T, width=3)
    d.ellipse([36, 36, w - 36, h - 36], fill=TIE_HI, outline=TIE_SH, width=2)
    # Lazo coral de la giantess en el centro
    cx, cy = w // 2, h // 2
    d.polygon([(cx - 4, cy), (cx - 34, cy - 18), (cx - 34, cy + 18)], fill=TIE_HI, outline=INK_T)
    d.polygon([(cx + 4, cy), (cx + 34, cy - 18), (cx + 34, cy + 18)], fill=TIE_HI, outline=INK_T)
    d.ellipse([cx - 10, cy - 10, cx + 10, cy + 10], fill=TIE_SH, outline=INK_T, width=2)
    _save(im, 'hair_tie/micro_idle.png')


def _street_shoe():
    # Macro (16x26, pivote [8, 13])
    im = Image.new('RGBA', (16, 26), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    d.ellipse([2, 3, 14, 24], fill=(0x3F, 0x36, 0x5C, 80))
    d.rounded_rectangle([2, 2, 13, 23], radius=5, fill=SHOE, outline=INK_T)
    d.rounded_rectangle([4, 11, 11, 20], radius=3, fill=SHOE_IN)
    d.rounded_rectangle([3, 3, 12, 11], radius=4, fill=SHOE_HI, outline=SHOE_D)
    _save(im, 'street_shoe/macro_idle.png')

    # Micro (224x464, pivote [112, 232] -> 11x26 cm a 16 px/cm + margen 24 px)
    w, h = 224, 464
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    d.rounded_rectangle([20, 20, w - 16, h - 14], radius=76, fill=(0x3F, 0x36, 0x5C, 60))
    d.rounded_rectangle([18, 18, w - 18, h - 18], radius=72, fill=SHOE_D, outline=INK_T, width=3)
    d.rounded_rectangle([26, 26, w - 26, h - 26], radius=64, fill=SHOE)
    d.rounded_rectangle([40, 38, w - 40, h - 38], radius=48, fill=SHOE_IN, outline=SHOE_D, width=3)
    # Lengüeta y cordones superiores (occluded)
    d.rounded_rectangle([28, 24, w - 28, 200], radius=56, fill=SHOE, outline=INK_T, width=3)
    d.arc([34, 32, w - 34, 175], start=195, end=345, fill=SHOE_HI, width=10)
    for yy in (110, 136, 162):
        d.line([(68, yy), (w - 68, yy)], fill=SOCK, width=4)
    _save(im, 'street_shoe/micro_idle.png')


def _genkan_bag():
    # Macro (26x20, pivote [13, 10])
    im = Image.new('RGBA', (26, 20), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    d.rounded_rectangle([2, 3, 24, 18], radius=4, fill=(0x3F, 0x36, 0x5C, 75))
    d.rounded_rectangle([2, 2, 23, 17], radius=4, fill=BAG, outline=INK_T)
    d.rectangle([3, 3, 22, 8], fill=BAG_HI)
    d.rectangle([8, 7, 17, 10], fill=BAG_STRAP, outline=INK_T)
    _save(im, 'genkan_bag/macro_idle.png')

    # Micro (464x336, pivote [232, 168] -> 26x18 cm a 16 px/cm + margen 24 px)
    w, h = 464, 336
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    d.rounded_rectangle([20, 22, w - 16, h - 14], radius=42, fill=(0x3F, 0x36, 0x5C, 60))
    d.rounded_rectangle([18, 18, w - 18, h - 18], radius=40, fill=BAG_SH, outline=INK_T, width=4)
    d.rounded_rectangle([28, 28, w - 28, h - 28], radius=34, fill=BAG)
    # Interior del bolsillo abierto
    d.rounded_rectangle([46, 56, w - 46, h - 42], radius=24, fill=BAG_SH, outline=INK_T, width=3)
    d.rounded_rectangle([54, 64, w - 54, h - 50], radius=20, fill=BAG_HI)
    # Solapa superior y correa
    d.rounded_rectangle([32, 24, w - 32, 92], radius=20, fill=BAG, outline=INK_T, width=3)
    d.rounded_rectangle([w // 2 - 42, 72, w // 2 + 42, 102], radius=8, fill=BAG_STRAP, outline=INK_T, width=3)
    _save(im, 'genkan_bag/micro_idle.png')


def _coin():
    m = Canvas(2, 2)
    m.px(0, 0, COIN_HI)
    m.px(1, 0, COIN)
    m.px(0, 1, INK)
    m.px(1, 1, COIN_SH)
    m.save(os.path.join(SRC, 'trinket_coin/macro_idle.png'))

    c = Canvas(14, 14)
    c.ellipse(7, 11, 5, 2, hexc(INK, 55))
    c.ellipse(7, 7, 5, 5, INK)
    c.ellipse(7, 7, 4, 4, COIN)
    c.ellipse(6, 6, 3, 3, COIN_HI)
    c.ellipse(7, 7, 2, 2, COIN)
    c.vline(7, 5, 4, COIN_SH)
    c.px(5, 4, '#FFFFFF')
    c.save(os.path.join(SRC, 'trinket_coin/micro_idle.png'))

    h = Canvas(14, 14)
    h.ellipse(7, 8, 4, 4, INK)
    h.ellipse(7, 8, 3, 3, COIN)
    h.px(6, 7, COIN_HI)
    h.px(8, 9, COIN_SH)
    h.save(os.path.join(SRC, 'trinket_coin/micro_held.png'))


def _bath_mat():
    # Macro (28x20, pivote [14, 10])
    im = Image.new('RGBA', (28, 20), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    d.rounded_rectangle([2, 3, 26, 18], radius=3, fill=(0x3F, 0x36, 0x5C, 65))
    d.rounded_rectangle([2, 2, 25, 17], radius=3, fill=MAT, outline=INK_T)
    d.rounded_rectangle([4, 4, 23, 15], radius=2, outline=SOCK)
    d.line([(6, 6), (21, 6)], fill=MAT_HI)
    d.line([(6, 13), (21, 13)], fill=MAT_SH)
    _save(im, 'bath_mat/macro_idle.png')

    # Micro (scale 4 -> 140x100, pivote [70, 50])
    w, h = 140, 100
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    d.rounded_rectangle([8, 10, w - 6, h - 6], radius=14, fill=(0x3F, 0x36, 0x5C, 50))
    d.rounded_rectangle([6, 6, w - 8, h - 8], radius=14, fill=MAT, outline=INK_T, width=3)
    d.rounded_rectangle([14, 14, w - 16, h - 16], radius=10, fill=MAT_SH, outline=SOCK, width=3)
    d.rounded_rectangle([20, 20, w - 22, h - 22], radius=8, fill=MAT)
    for yy in range(28, h - 24, 12):
        for xx in range(28, w - 24, 14):
            d.ellipse([xx, yy, xx + 4, yy + 4], fill=MAT_HI)
    _save(im, 'bath_mat/micro_idle.png')


def _rice_cooker_use():
    # Macro use (26x26, pivote [13, 20]): base igual a rice_cooker + vapor blanco encima
    im = Image.new('RGBA', (26, 26), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    d.ellipse([3, 18, 23, 25], fill=(0x3F, 0x36, 0x5C, 70))
    d.rounded_rectangle([3, 5, 22, 23], radius=5, fill=SOCK, outline=INK_T)
    d.ellipse([8, 8, 17, 16], fill=SOCK_SH, outline=(0xD6, 0xC4, 0xA9, 255))
    d.rectangle([9, 17, 16, 20], fill=SOCK_STRIPE)
    d.point((15, 18), fill=TIE)
    # Nubes de vapor tibio saliendo de la válvula
    d.ellipse([9, 2, 14, 7], fill=(255, 255, 255, 225), outline=INK_T)
    d.ellipse([13, 0, 18, 5], fill=(255, 255, 255, 235), outline=INK_T)
    d.ellipse([11, 5, 15, 9], fill=(255, 255, 255, 210))
    _save(im, 'rice_cooker/macro_use.png')


def _rice_grain():
    # Macro (3x2, selected_only, pivote [1, 1])
    m = Canvas(3, 2)
    m.px(0, 0, RICE_HI)
    m.px(1, 0, RICE)
    m.px(2, 0, RICE_SH)
    m.px(1, 1, INK)
    m.save(os.path.join(SRC, 'rice_grain/macro_idle.png'))

    # Micro (16x14, pivote [8, 11])
    c = Canvas(16, 14)
    c.ellipse(8, 11, 6, 2, hexc(INK, 55))
    for gx, gy, col in ((5, 9, RICE_SH), (10, 9, RICE_SH), (7, 7, RICE), (9, 8, RICE), (6, 8, RICE_HI), (8, 6, RICE_HI)):
        c.ellipse(gx, gy, 2, 1, INK)
        c.hline(gx - 1, gx + 1, gy, col)
    c.save(os.path.join(SRC, 'rice_grain/micro_idle.png'))

    h = Canvas(14, 14)
    h.ellipse(7, 8, 3, 2, INK)
    h.ellipse(7, 8, 2, 1, RICE_HI)
    h.px(6, 7, RICE)
    h.save(os.path.join(SRC, 'rice_grain/micro_held.png'))


def _potted_plant():
    # Macro (46x54, pivote [23, 50])
    im = Image.new('RGBA', (46, 54), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    d.ellipse([7, 41, 39, 52], fill=(0x3F, 0x36, 0x5C, 80))
    d.polygon([(12, 26), (34, 26), (30, 48), (16, 48)], fill=POT, outline=INK_T)
    d.rounded_rectangle([10, 22, 36, 28], radius=3, fill=POT_HI, outline=INK_T)
    d.ellipse([12, 22, 34, 27], fill=SOIL, outline=INK_T)
    # Hojas alrededor dejando ver la maceta y la tierra
    for box, col in (
        ([3, 12, 20, 24], LEAF_D),
        ([26, 12, 43, 24], LEAF_D),
        ([8, 4, 24, 18], LEAF),
        ([22, 4, 38, 18], LEAF),
        ([15, 1, 31, 14], LEAF_HI),
    ):
        d.ellipse(box, fill=col, outline=INK_T)
    _save(im, 'potted_plant/macro_idle.png')

    # Micro (scale 4 -> 184x184, pivote [92, 92] -> 36x36 cm = 144x144 px + 20 px hojas)
    w, h = 184, 184
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    d.ellipse([8, 10, w - 6, h - 6], fill=(0x3F, 0x36, 0x5C, 55))
    d.ellipse([10, 10, w - 10, h - 10], fill=POT_SH, outline=INK_T, width=3)
    d.ellipse([16, 16, w - 16, h - 16], fill=POT_HI, outline=INK_T, width=2)
    d.ellipse([24, 24, w - 24, h - 24], fill=SOIL, outline=INK_T, width=3)
    # Textura orgánica de tierra húmeda en el interior caminable
    for xx, yy, r in ((54, 58, 8), (118, 62, 10), (76, 118, 9), (126, 112, 7), (92, 88, 12)):
        d.ellipse([xx - r, yy - r // 2, xx + r, yy + r // 2], fill=SOIL_D)
    # Hojas decorando el borde exterior sin tapar el suelo
    for box, col in (
        ([2, 64, 34, 116], LEAF),
        ([w - 34, 64, w - 2, 116], LEAF),
        ([64, 2, 116, 32], LEAF_HI),
        ([64, h - 32, 116, h - 2], LEAF_D),
    ):
        d.ellipse(box, fill=col, outline=INK_T, width=3)
    _save(im, 'potted_plant/micro_idle.png')


def main(src):
    global SRC
    SRC = src
    _socks()
    _hair_tie()
    _street_shoe()
    _genkan_bag()
    _coin()
    _bath_mat()
    _rice_cooker_use()
    _rice_grain()
    _potted_plant()
