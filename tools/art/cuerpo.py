"""Arte micro detallado del cuerpo de la giantess, futón y almohada (propuesta 19).

Cada región corporal es un sprite anatómico a escala real (16 px/cm) con
silueta orgánica, contorno, sombreado cel-shading (luz desde arriba a la
izquierda) y prendas coherentes con giantess.py y hand.py:
- Cabello violeta largo con flequillo, brillos y moño rosa.
- Piel cálida con rubor y sombras terracota suaves.
- Cárdigan lila tejido con blusa crema, lazo rosa, botones y bolsillo.
- Falda tableada violeta oscuro.
- Medias crema acanaladas y calzado violeta con correa.
"""
import math
import os

from PIL import Image, ImageDraw
from px import Canvas, hexc

SRC = None  # lo fija main

# Paleta coherente con giantess.py y hand.py
INK = hexc('#3F365C')
OUTLINE = hexc('#6E4C68')

HAIR = hexc('#5D508D')
HAIR_D = hexc('#4A3F75')
HAIR_D2 = hexc('#392F5E')
HAIR_L = hexc('#8677B7')
HAIR_HI = hexc('#A994D6')

SKIN = hexc('#FFD8B9')
SKIN_HI = hexc('#FFEAD6')
SKIN_SH = hexc('#F4B49A')
SKIN_SH2 = hexc('#E09888')
CREASE = hexc('#CF8C88')
BLUSH = hexc('#F49AAB')
NAIL = hexc('#FFE8E2')

EYE = hexc('#3D9295')
EYE_D = hexc('#2C6E78')
EYE_L = hexc('#76D8CB')
WHITE = hexc('#FFFFFF')
MOUTH = hexc('#B8748B')

CARD = hexc('#9A8BE3')
CARD_L = hexc('#C8B9F5')
CARD_D = hexc('#7A6BC6')
CARD_D2 = hexc('#6152A3')

BLOUSE = hexc('#FFF1D8')
BLOUSE_HI = hexc('#FFFBF0')
BLOUSE_D = hexc('#E3D3C9')

BOW = hexc('#FF9CAD')
BOW_HI = hexc('#FFC4D0')
BOW_D = hexc('#CE788A')

SKIRT = hexc('#67588F')
SKIRT_L = hexc('#8070AD')
SKIRT_D = hexc('#504273')
SKIRT_D2 = hexc('#3E335C')

SOCK = hexc('#FFF1D8')
SOCK_HI = hexc('#FFFBF2')
SOCK_D = hexc('#DCCFE0')
SOCK_D2 = hexc('#BEB0C9')

SHOE = hexc('#76659B')
SHOE_L = hexc('#9A8AC2')
SHOE_HI = hexc('#B8AAD9')
SHOE_D = hexc('#554675')
SHOE_D2 = hexc('#3D3157')

BTN = hexc('#FF963D')
BTN_HI = hexc('#FFCA8A')


def save_img(im, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    im.save(path)


def save_canvas(c, rel):
    save_img(c.im, rel)


def blend(c1, c2, t):
    t = max(0.0, min(1.0, t))
    return (
        int(c1[0] + (c2[0] - c1[0]) * t),
        int(c1[1] + (c2[1] - c1[1]) * t),
        int(c1[2] + (c2[2] - c1[2]) * t),
        255,
    )


def add_outline(im, col=INK, thickness=2):
    """Agrega contorno limpio alrededor de los píxeles opacos de im."""
    w, h = im.size
    src = im.load()
    out = im.copy()
    dst = out.load()
    offsets = []
    for dy in range(-thickness, thickness + 1):
        for dx in range(-thickness, thickness + 1):
            if dx * dx + dy * dy <= thickness * thickness + 1 and (dx or dy):
                offsets.append((dx, dy))
    for y in range(h):
        for x in range(w):
            if src[x, y][3] > 0:
                continue
            for dx, dy in offsets:
                nx, ny = x + dx, y + dy
                if 0 <= nx < w and 0 <= ny < h and src[nx, ny][3] > 200:
                    dst[x, y] = col
                    break
    return out


def shade_cylinder(im, base, light, dark, deep, knit=False, rib=0, seam_x=None):
    """Aplica sombreado cilíndrico (luz desde la izquierda) a los píxeles no transparentes."""
    w, h = im.size
    px = im.load()
    for y in range(h):
        # encontrar límites horizontales del cuerpo en esta fila
        x0 = -1
        x1 = -1
        for x in range(w):
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
            u = (x - x0) / span  # 0 = izq (luz), 1 = der (sombra)
            if u < 0.16:
                col = light
            elif u < 0.62:
                col = base
            elif u < 0.86:
                col = dark
            else:
                col = deep
            if knit:
                ux, uy = x % 8, y % 8
                if (ux == 1 and uy in (1, 2, 3)) or (ux == 2 and uy in (3, 4)):
                    col = blend(col, light, 0.45)
                elif ux == 5 and uy in (2, 3, 4):
                    col = blend(col, deep, 0.35)
            elif rib > 0:
                r = x % rib
                if r == 0:
                    col = blend(col, light, 0.45)
                elif r == rib // 2:
                    col = blend(col, dark, 0.35)
            px[x, y] = col
        if seam_x is not None:
            sx = int(round(x0 + span * seam_x))
            if x0 + 3 <= sx <= x1 - 3:
                px[sx, y] = deep
                if y % 6 < 3 and sx + 1 <= x1 - 2:
                    px[sx + 1, y] = light


def draw_bow(draw, cx, cy, scale=1.0):
    """Lazo rosa característico de la giantess."""
    s = scale
    # lazos laterales
    left = [
        (cx - 2 * s, cy),
        (cx - 24 * s, cy - 12 * s),
        (cx - 28 * s, cy - 2 * s),
        (cx - 22 * s, cy + 12 * s),
    ]
    right = [
        (cx + 2 * s, cy),
        (cx + 24 * s, cy - 12 * s),
        (cx + 28 * s, cy - 2 * s),
        (cx + 22 * s, cy + 12 * s),
    ]
    draw.polygon(left, fill=BOW, outline=INK)
    draw.polygon(right, fill=BOW, outline=INK)
    # brillos y pliegues
    draw.polygon([(cx - 6 * s, cy - 2 * s), (cx - 21 * s, cy - 9 * s), (cx - 20 * s, cy - 3 * s)], fill=BOW_HI)
    draw.polygon([(cx + 6 * s, cy - 2 * s), (cx + 21 * s, cy - 9 * s), (cx + 20 * s, cy - 3 * s)], fill=BOW_HI)
    draw.line([(cx - 5 * s, cy + 2 * s), (cx - 19 * s, cy + 6 * s)], fill=BOW_D, width=max(1, int(2 * s)))
    draw.line([(cx + 5 * s, cy + 2 * s), (cx + 19 * s, cy + 6 * s)], fill=BOW_D, width=max(1, int(2 * s)))
    # cintas colgantes
    tail_l = [(cx - 4 * s, cy + 4 * s), (cx - 16 * s, cy + 28 * s), (cx - 8 * s, cy + 26 * s), (cx - 1 * s, cy + 6 * s)]
    tail_r = [(cx + 4 * s, cy + 4 * s), (cx + 16 * s, cy + 28 * s), (cx + 8 * s, cy + 26 * s), (cx + 1 * s, cy + 6 * s)]
    draw.polygon(tail_l, fill=BOW, outline=INK)
    draw.polygon(tail_r, fill=BOW_D, outline=INK)
    # nudo central
    draw.rounded_rectangle([cx - 6 * s, cy - 6 * s, cx + 6 * s, cy + 6 * s], radius=int(3 * s), fill=BOW_D, outline=INK)
    draw.ellipse([cx - 3 * s, cy - 4 * s, cx + 2 * s, cy], fill=BOW_HI)


# ------------------------------------------------------------ muebles nuevos

def futon_micro_tile():
    """Futón de algodón crema acolchado con pespuntes en rombo."""
    w, h = 64, 64
    base, hi, sh, stitch = hexc('#FFF4E2'), hexc('#FFFBF2'), hexc('#F0DEC6'), hexc('#D2BCA2')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            weave = (x + y) % 4
            col = hi if weave == 0 else (sh if weave == 2 else base)
            d1 = (x + y) % 32
            d2 = (x - y) % 32
            if (d1 == 0 or d2 == 0) and (x % 4 < 2):
                col = stitch
            elif d1 == 1 or d2 == 1:
                col = sh
            px[x, y] = col
    for cx, cy in ((0, 0), (32, 0), (16, 16), (48, 16), (0, 32), (32, 32), (16, 48), (48, 48)):
        for dy in (-1, 0, 1):
            for dx in (-1, 0, 1):
                if abs(dx) + abs(dy) <= 1:
                    px[(cx + dx) % w, (cy + dy) % h] = stitch
    save_canvas(c, 'futon/micro_tile.png')


def pillow_micro_tile():
    """Almohada de lino lavanda muy claro con trama delicada."""
    w, h = 64, 64
    base, hi, sh, seam = hexc('#F5EFFA'), hexc('#FFFFFF'), hexc('#E5DAF0'), hexc('#CBBBD9')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            if (x % 8 == 0 and y % 2 == 0) or (y % 8 == 0 and x % 2 == 0):
                col = sh
            elif (x + y) % 6 == 0:
                col = hi
            else:
                col = base
            px[x, y] = col
    for x in range(0, w, 6):
        c.hline(x, 31, 3, seam)
    save_canvas(c, 'pillow/micro_tile.png')


def grip_point():
    """Marcador de punto de agarre (costura, puño, mechón) en vista micro."""
    c = Canvas(16, 16)
    ring, hi, core = '#76D8CB', '#CFF7F0', '#FFF7E5'
    for y in range(16):
        for x in range(16):
            d = math.hypot(x - 7.5, y - 7.5)
            if 5.2 <= d <= 6.8:
                c.px(x, y, '#3F365C')
            elif 3.6 <= d < 5.2:
                c.px(x, y, hi if y < 7 else ring)
            elif 2.3 <= d < 3.6:
                c.px(x, y, '#3F365C')
            elif d <= 1.8:
                c.px(x, y, core)
    save_canvas(c, 'grip_point/micro_idle.png')


# ------------------------------------------------------------ cuerpo anatómico

def body_head():
    """Cabello y rostro (20×22 cm -> 320×352 px, sprite 368×400, pivote [24, 376])."""
    w, h = 368, 400
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    cx = w // 2

    # Cabello posterior voluminoso (óvalo + mechones largos a los lados)
    d.ellipse([20, 14, w - 20, 340], fill=HAIR_D)
    d.rounded_rectangle([24, 160, w - 24, h - 18], radius=46, fill=HAIR_D)
    d.ellipse([20, 10, w - 28, 330], fill=HAIR)
    d.rounded_rectangle([24, 156, w - 32, h - 22], radius=44, fill=HAIR)

    # Orejas asomando a los costados
    d.ellipse([54, 194, 88, 250], fill=SKIN_SH, outline=OUTLINE, width=2)
    d.ellipse([w - 88, 194, w - 54, 250], fill=SKIN_SH, outline=OUTLINE, width=2)

    # Rostro (óvalo suave con mentón delicado)
    face_poly = [
        (76, 150), (w - 76, 150), (w - 78, 256), (w - 102, 316),
        (cx + 26, 348), (cx - 26, 348), (102, 316), (78, 256),
    ]
    d.ellipse([74, 108, w - 74, 312], fill=SKIN)
    d.polygon(face_poly, fill=SKIN)
    # Sombra suave en el lado derecho del rostro y bajo el mentón
    d.arc([74, 108, w - 74, 312], start=330, end=80, fill=SKIN_SH, width=10)
    d.line([(w - 84, 256), (w - 104, 314), (cx + 24, 346), (cx - 24, 346)], fill=SKIN_SH, width=5)

    # Ojos suaves estilo anime
    for side, ex in ((-1, cx - 54), (1, cx + 54)):
        # pestaña superior curva
        d.arc([ex - 26, 200, ex + 26, 236], start=190, end=350, fill=INK, width=5)
        # iris turquesa
        d.ellipse([ex - 18, 208, ex + 18, 254], fill=EYE)
        d.ellipse([ex - 14, 212, ex + 14, 240], fill=EYE_D)
        d.ellipse([ex - 8, 218, ex + 8, 238], fill=INK)
        d.arc([ex - 12, 232, ex + 12, 252], start=10, end=170, fill=EYE_L, width=5)
        # brillos
        d.ellipse([ex - 12, 214, ex - 2, 224], fill=WHITE)
        d.ellipse([ex + 6, 238, ex + 12, 244], fill=WHITE)
        # ceja delicada
        d.arc([ex - 20, 182, ex + 20, 204], start=200, end=340, fill=HAIR_D, width=3)

    # Rubor en mejillas
    d.ellipse([96, 260, 134, 278], fill=BLUSH)
    d.ellipse([w - 134, 260, w - 96, 278], fill=BLUSH)

    # Nariz diminuta y sonrisa cálida
    d.line([(cx, 264), (cx + 3, 272)], fill=SKIN_SH2, width=2)
    d.arc([cx - 14, 286, cx + 14, 304], start=15, end=165, fill=MOUTH, width=3)

    # Flequillo con puntas y mechones laterales (sobre los puntos de agarre)
    bangs = [
        (40, 96), (w - 40, 96), (w - 38, 216), (w - 64, 174),
        (w - 86, 202), (w - 114, 158), (cx + 34, 198), (cx, 152),
        (cx - 34, 198), (114, 158), (86, 202), (64, 174), (38, 216),
    ]
    d.polygon(bangs, fill=HAIR)
    # Mechones laterales largos (izquierdo y derecho)
    d.polygon([(36, 150), (88, 176), (86, 326), (54, 376), (26, 286)], fill=HAIR)
    d.polygon([(w - 36, 150), (w - 88, 176), (w - 86, 326), (w - 54, 376), (w - 26, 286)], fill=HAIR)

    # Líneas de separación de mechones y brillo en arco sobre la coronilla
    for x0, x1 in ((94, 102), (144, 148), (cx, cx), (w - 144, w - 148), (w - 94, w - 102)):
        d.line([(x0, 94), (x1, 166)], fill=HAIR_D, width=3)
    d.line([(62, 192), (56, 356)], fill=HAIR_D, width=3)
    d.line([(w - 62, 192), (w - 56, 356)], fill=HAIR_D, width=3)
    # Brillo del cabello
    d.arc([72, 46, w - 96, 146], start=195, end=330, fill=HAIR_L, width=14)
    d.arc([88, 54, w - 120, 134], start=205, end=315, fill=HAIR_HI, width=6)

    # Moño rosa en la parte superior derecha
    draw_bow(d, w - 96, 86, scale=1.35)

    im = add_outline(im, INK, thickness=2)
    save_img(im, 'body_head/micro_idle.png')


def body_neck():
    """Cuello y nuca (8×12 cm -> 128×192 px, sprite 192×224, pivote [32, 208])."""
    w, h = 192, 224
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    cx = w // 2

    # Cilindro del cuello con suave curvatura hacia los hombros abajo
    neck_pts = [
        (cx - 50, 16), (cx + 50, 16), (cx + 52, 136), (cx + 78, 186),
        (cx + 80, h - 14), (cx - 80, h - 14), (cx - 78, 186), (cx - 52, 136),
    ]
    d.polygon(neck_pts, fill=SKIN)
    shade_cylinder(im, SKIN, SKIN_HI, SKIN_SH, SKIN_SH2)

    # Sombra bajo el mentón/nuca arriba
    d = ImageDraw.Draw(im)
    d.pieslice([cx - 56, -22, cx + 56, 62], start=0, end=180, fill=SKIN_SH)
    d.arc([cx - 46, 20, cx + 46, 56], start=10, end=170, fill=SKIN_SH2, width=4)

    # Mechones de cabello cayendo por los costados arriba
    d.polygon([(cx - 56, 14), (cx - 38, 14), (cx - 44, 96), (cx - 62, 74)], fill=HAIR)
    d.polygon([(cx + 56, 14), (cx + 38, 14), (cx + 44, 96), (cx + 62, 74)], fill=HAIR_D)

    # Cuello de la blusa crema y escote del cárdigan lila abajo (agarre en y=6cm -> py=112)
    d.polygon([(cx - 82, 136), (cx - 42, 128), (cx, 174), (cx + 42, 128), (cx + 82, 136), (cx + 84, h - 12), (cx - 84, h - 12)], fill=CARD)
    # Puntas de la blusa crema
    d.polygon([(cx - 54, 124), (cx - 12, 164), (cx - 46, 174), (cx - 64, 146)], fill=BLOUSE, outline=INK)
    d.polygon([(cx + 54, 124), (cx + 12, 164), (cx + 46, 174), (cx + 64, 146)], fill=BLOUSE_D, outline=INK)
    # Escote en V y moño pequeño
    d.line([(cx - 68, 134), (cx, 198), (cx + 68, 134)], fill=CARD_D2, width=5)
    draw_bow(d, cx, 184, scale=0.75)

    im = add_outline(im, INK, thickness=2)
    save_img(im, 'body_neck/micro_idle.png')


def body_shoulders():
    """Hombro derecho e izquierdo (12×12 cm -> 192×192 px, sprite 240×240, pivote [24, 216])."""
    w, h = 240, 240
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)

    # Hombro derecho: lado interno (cuello) a la izquierda, caída del brazo a la derecha
    pts = [
        (26, 36), (118, 22), (186, 44), (218, 102), (220, 214),
        (26, 214),
    ]
    d.polygon(pts, fill=CARD)
    shade_cylinder(im, CARD, CARD_L, CARD_D, CARD_D2, knit=True)

    d = ImageDraw.Draw(im)
    # Borde de blusa crema y piel en el lado interno (izquierda)
    d.polygon([(24, 34), (72, 26), (56, 98), (24, 118)], fill=BLOUSE, outline=INK)
    d.polygon([(24, 34), (56, 28), (44, 68), (24, 78)], fill=SKIN)
    # Costura curva del hombro pasando por el centro (agarre)
    d.arc([78, 24, 206, 214], start=195, end=335, fill=CARD_D2, width=5)
    for ang in range(205, 330, 12):
        rad = math.radians(ang)
        sx = 142 + int(64 * math.cos(rad))
        sy = 119 + int(95 * math.sin(rad))
        d.line([(sx - 3, sy), (sx + 3, sy)], fill=CARD_L, width=2)
    # Mechón de cabello violeta cayendo sobre la parte superior trasera
    d.polygon([(46, 22), (118, 20), (132, 62), (96, 84), (74, 58), (52, 76)], fill=HAIR)
    d.line([(76, 24), (88, 66)], fill=HAIR_D, width=3)
    d.line([(98, 24), (112, 62)], fill=HAIR_L, width=3)

    im_r = add_outline(im, INK, thickness=2)
    save_img(im_r, 'body_shoulder_r/micro_idle.png')
    save_img(im_r.transpose(Image.FLIP_LEFT_RIGHT), 'body_shoulder_l/micro_idle.png')


def body_chest():
    """Pecho (28×36 cm -> 448×576 px, sprite 496×624, pivote [24, 600])."""
    w, h = 496, 624
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    cx = w // 2

    # Silueta femenina de joven adulta: busto prominente, cintura entallada, cadera suave
    torso = [
        (42, 34), (146, 22), (w - 146, 22), (w - 42, 34),
        (w - 28, 118), (w - 14, 198), (w - 18, 276), (w - 42, 346),
        (w - 76, 444), (w - 64, 520), (w - 48, 574), (w - 54, h - 22),
        (54, h - 22), (48, 574), (64, 520), (76, 444),
        (42, 346), (18, 276), (14, 198), (28, 118),
    ]
    d.polygon(torso, fill=CARD)
    shade_cylinder(im, CARD, CARD_L, CARD_D, CARD_D2, knit=True)

    # Volumen 3D del busto sobre el tejido del cárdigan (cel-shading suave, sin cortes rectos)
    bust = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    db = ImageDraw.Draw(bust)
    # Sombra proyectada continua bajo la línea del busto (forma de W suave)
    under_shadow = [
        (36, 266), (72, 316), (138, 338), (cx - 24, 322), (cx, 298),
        (cx + 24, 322), (w - 138, 338), (w - 72, 316), (w - 36, 266),
        (w - 44, 344), (w - 92, 368), (cx + 28, 356), (cx, 336),
        (cx - 28, 356), (92, 368), (44, 344),
    ]
    db.polygon(under_shadow, fill=(*CARD_D2[:3], 130))
    for side in (-1, 1):
        bx = cx + side * 96
        # media luna suave en la cara inferior del busto
        db.arc([bx - 88, 178, bx + 88, 322], start=18, end=162, fill=(*CARD_D[:3], 190), width=28)
        db.arc([bx - 88, 186, bx + 88, 324], start=22, end=158, fill=(*CARD_D2[:3], 220), width=8)
        # brillo suave en la pendiente superior (luz desde arriba a la izquierda)
        hl_dx = -18 if side == -1 else -10
        db.arc([bx - 76 + hl_dx, 166, bx + 68 + hl_dx, 260], start=205, end=330, fill=(*CARD_L[:3], 150), width=22)
        db.arc([bx - 64 + hl_dx, 174, bx + 52 + hl_dx, 244], start=210, end=320, fill=(*CARD_L[:3], 200), width=10)
    # valle central y pliegues de tensión de la tela hacia los botones
    db.polygon([(cx - 22, 214), (cx + 22, 214), (cx + 16, 326), (cx - 16, 326)], fill=(*CARD_D2[:3], 110))
    for dy, spread in ((240, 54), (266, 68), (290, 44)):
        db.line([(cx - 16, dy), (cx - 16 - spread, dy + 10)], fill=(*CARD_D2[:3], 150), width=3)
        db.line([(cx + 16, dy), (cx + 16 + spread, dy + 10)], fill=(*CARD_D2[:3], 150), width=3)
    # recortar el busto a la silueta del torso
    mask = im.getchannel('A')
    bust.putalpha(Image.composite(bust.getchannel('A'), Image.new('L', (w, h), 0), mask))
    im.alpha_composite(bust)

    d = ImageDraw.Draw(im)
    # Escote en V con blusa crema y cuello
    v_poly = [(cx - 102, 22), (cx + 102, 22), (cx, 212)]
    d.polygon(v_poly, fill=BLOUSE)
    d.polygon([(cx + 12, 22), (cx + 102, 22), (cx, 212)], fill=BLOUSE_D)
    # Borde tejido del escote en V
    d.line([(cx - 104, 22), (cx, 216), (cx + 104, 22)], fill=CARD_D2, width=10)
    d.line([(cx - 96, 22), (cx, 204), (cx + 96, 22)], fill=CARD_L, width=3)

    # Tapeta central de botones
    d.rectangle([cx - 16, 212, cx + 16, 540], fill=CARD_D)
    d.line([(cx - 16, 212), (cx - 16, 540)], fill=CARD_D2, width=3)
    d.line([(cx + 16, 212), (cx + 16, 540)], fill=CARD_D2, width=3)
    for by in (280, 386, 488):
        d.ellipse([cx - 12, by - 12, cx + 12, by + 12], fill=BTN, outline=INK, width=2)
        d.ellipse([cx - 6, by - 7, cx, by - 1], fill=BTN_HI)

    # Bolsillo cosido en el frente bajo el busto (alrededor del agarre)
    pk = [cx + 42, 344, cx + 146, 444]
    d.rounded_rectangle(pk, radius=12, fill=CARD_D, outline=CARD_D2, width=4)
    d.line([(pk[0] + 6, pk[1] + 6), (pk[2] - 6, pk[1] + 6)], fill=CARD_L, width=3)

    # Moño rosa en el escote
    draw_bow(d, cx, 128, scale=1.7)

    # Pretina acanalada inferior del cárdigan + borde de la falda violeta
    d.rectangle([52, 536, w - 52, 578], fill=CARD_D)
    for x in range(56, w - 56, 8):
        d.line([(x, 538), (x, 576)], fill=CARD_D2, width=2)
    d.polygon([(54, 578), (w - 54, 578), (w - 48, h - 22), (48, h - 22)], fill=SKIRT)
    for x in range(74, w - 60, 44):
        d.line([(x, 578), (x + (x - cx) // 10, h - 22)], fill=SKIRT_D2, width=4)

    im = add_outline(im, INK, thickness=2)
    save_img(im, 'body_chest/micro_idle.png')


def body_back():
    """Espalda (28×36 cm -> 448×576 px, sprite 496×624, pivote [24, 600])."""
    w, h = 496, 624
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    cx = w // 2

    torso = [
        (36, 36), (146, 22), (w - 146, 22), (w - 36, 36),
        (w - 22, 140), (w - 36, 290), (w - 64, 446), (w - 48, 566),
        (w - 58, h - 22), (58, h - 22), (48, 566), (64, 446),
        (36, 290), (22, 140),
    ]
    d.polygon(torso, fill=CARD)
    shade_cylinder(im, CARD, CARD_L, CARD_D, CARD_D2, knit=True, seam_x=0.5)

    d = ImageDraw.Draw(im)
    # Cabello largo cayendo en puntas sobre la parte superior de la espalda
    hair_tips = [
        (64, 22), (w - 64, 22), (w - 58, 198),
        (w - 94, 242), (w - 126, 206), (cx + 56, 256), (cx + 18, 212),
        (cx - 18, 258), (cx - 64, 214), (114, 248), (78, 204),
    ]
    d.polygon(hair_tips, fill=HAIR_D)
    d.polygon([(x, max(22, y - 8)) for x, y in hair_tips], fill=HAIR)
    for hx in range(96, w - 90, 38):
        d.line([(hx, 26), (hx + (hx - cx) // 8, 210)], fill=HAIR_D2, width=4)
        d.line([(hx + 10, 44), (hx + 12, 130)], fill=HAIR_L, width=3)

    # Pretina inferior del cárdigan + borde de falda
    d.rectangle([52, 536, w - 52, 578], fill=CARD_D)
    for x in range(56, w - 56, 8):
        d.line([(x, 538), (x, 576)], fill=CARD_D2, width=2)
    d.polygon([(54, 578), (w - 54, 578), (w - 48, h - 22), (48, h - 22)], fill=SKIRT)
    for x in range(74, w - 60, 44):
        d.line([(x, 578), (x, h - 22)], fill=SKIRT_D2, width=4)

    im = add_outline(im, INK, thickness=2)
    save_img(im, 'body_back/micro_idle.png')


def body_sides():
    """Costado derecho e izquierdo (10×34 cm -> 160×544 px, sprite 208×592, pivote [24, 568])."""
    w, h = 208, 592
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)

    # Perfil lateral de joven adulta: busto proyectado al frente (der), cintura marcada y cadera
    side_poly = [
        (42, 24), (154, 28), (186, 132), (198, 214), (188, 284), (152, 342),
        (146, 412), (176, 514), (174, h - 22), (34, h - 22), (26, 514),
        (46, 404), (34, 240), (28, 130),
    ]
    d.polygon(side_poly, fill=CARD)
    shade_cylinder(im, CARD, CARD_L, CARD_D, CARD_D2, knit=True, seam_x=0.45)

    # Detalles internos recortados a la silueta (sisa y sombra curva bajo el busto)
    det = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    dd = ImageDraw.Draw(det)
    dd.arc([44, 14, 156, 140], start=10, end=170, fill=(*CARD_D2[:3], 235), width=5)
    dd.polygon([(112, 286), (186, 284), (154, 344), (146, 366), (108, 326)], fill=(*CARD_D2[:3], 145))
    dd.arc([104, 224, 192, 322], start=35, end=145, fill=(*CARD_D2[:3], 220), width=6)
    mask = im.getchannel('A')
    det.putalpha(Image.composite(det.getchannel('A'), Image.new('L', (w, h), 0), mask))
    im.alpha_composite(det)

    d = ImageDraw.Draw(im)
    # Pretina y borde de falda abajo
    d.rectangle([32, 512, 176, 548], fill=CARD_D)
    for x in range(36, 174, 8):
        d.line([(x, 514), (x, 546)], fill=CARD_D2, width=2)
    d.polygon([(34, 548), (174, 548), (178, h - 22), (30, h - 22)], fill=SKIRT)

    im_r = add_outline(im, INK, thickness=2)
    save_img(im_r, 'body_side_r/micro_idle.png')
    save_img(im_r.transpose(Image.FLIP_LEFT_RIGHT), 'body_side_l/micro_idle.png')


def body_hips():
    """Cintura y caderas (32×20 cm -> 512×320 px, sprite 560×368, pivote [24, 344])."""
    w, h = 560, 368
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    cx = w // 2

    # Silueta acampanada de cintura a falda tableada
    skirt_poly = [
        (66, 22), (w - 66, 22), (w - 46, 108), (w - 20, h - 24), (20, h - 24), (46, 108),
    ]
    d.polygon(skirt_poly, fill=SKIRT)
    shade_cylinder(im, SKIRT, SKIRT_L, SKIRT_D, SKIRT_D2)

    d = ImageDraw.Draw(im)
    # Tablas de la falda abriéndose hacia abajo con luz y sombra por pliegue
    n_pleats = 10
    for i in range(n_pleats):
        t0 = i / n_pleats
        t1 = (i + 1) / n_pleats
        tx0 = 66 + (w - 132) * t0
        tx1 = 66 + (w - 132) * t1
        bx0 = 22 + (w - 44) * t0
        bx1 = 22 + (w - 44) * t1
        # borde iluminado del pliegue
        d.polygon([(tx0, 82), (tx0 + 12, 82), (bx0 + 14, h - 24), (bx0, h - 24)], fill=SKIRT_L)
        # sombra profunda en la arista derecha del pliegue
        d.polygon([(tx1 - 10, 82), (tx1, 82), (bx1, h - 24), (bx1 - 12, h - 24)], fill=SKIRT_D2)
        d.line([(tx0, 82), (bx0, h - 24)], fill=INK, width=2)

    # Pretina de la falda + borde inferior del cárdigan lila arriba
    d.polygon([(66, 22), (w - 66, 22), (w - 56, 68), (56, 68)], fill=CARD)
    for x in range(66, w - 66, 8):
        d.line([(x, 24), (x, 66)], fill=CARD_D, width=2)
    d.polygon([(56, 68), (w - 56, 68), (w - 52, 92), (52, 92)], fill=SKIRT_D, outline=INK)
    # Pespunte de la cintura
    for x in range(64, w - 64, 14):
        d.line([(x, 80), (x + 7, 80)], fill=SKIRT_L, width=2)

    im = add_outline(im, INK, thickness=2)
    save_img(im, 'body_hips/micro_idle.png')


def _draw_hand_vertical(back=True):
    """Mano vertical anatómica (10×14 cm -> 160×224 px, sprite 208×272, pivote [24, 248])."""
    w, h = 208, 272
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)

    # Cuatro dedos apuntando hacia arriba + pulgar al costado
    fingers = [
        (46, 48, 66, 134),   # meñique
        (70, 26, 92, 128),   # anular
        (96, 18, 118, 126),  # medio
        (122, 32, 144, 130), # índice
    ]
    for fx0, fy0, fx1, fy1 in fingers:
        d.rounded_rectangle([fx0, fy0, fx1, fy1], radius=10, fill=SKIN)
    # Pulgar
    thumb = [(140, 122), (176, 86), (186, 102), (156, 162)]
    d.polygon(thumb, fill=SKIN)
    d.ellipse([166, 82, 188, 104], fill=SKIN)
    # Palma / dorso + muñeca
    d.rounded_rectangle([42, 110, 152, 212], radius=28, fill=SKIN)
    d.rounded_rectangle([58, 196, 138, 248], radius=12, fill=SKIN)

    shade_cylinder(im, SKIN, SKIN_HI, SKIN_SH, SKIN_SH2)

    d = ImageDraw.Draw(im)
    # Separación entre dedos
    for fx in (68, 94, 120):
        d.line([(fx, 56), (fx, 126)], fill=CREASE, width=2)
    if back:
        # Uñas delicadas y nudillos sutiles (dorso de mano derecha)
        for fx0, fy0, fx1, _ in fingers:
            d.rounded_rectangle([fx0 + 4, fy0 + 4, fx1 - 4, fy0 + 18], radius=5, fill=NAIL, outline=CREASE)
        for nx in (56, 81, 107, 133):
            d.arc([nx - 8, 122, nx + 8, 134], start=190, end=350, fill=CREASE, width=2)
    else:
        # Pliegues de palma (mano izquierda)
        d.arc([54, 134, 142, 166], start=15, end=165, fill=CREASE, width=3)
        d.arc([62, 156, 136, 194], start=200, end=340, fill=CREASE, width=3)

    # Puño de la manga lila en la muñeca abajo
    d.rounded_rectangle([52, 224, 144, 254], radius=8, fill=CARD, outline=INK, width=2)
    for x in range(58, 140, 8):
        d.line([(x, 228), (x, 250)], fill=CARD_D, width=2)

    return add_outline(im, INK, thickness=2)


def body_hands():
    save_img(_draw_hand_vertical(back=True), 'body_hand_r_back/micro_idle.png')
    save_img(_draw_hand_vertical(back=False).transpose(Image.FLIP_LEFT_RIGHT), 'body_hand_l/micro_idle.png')


def body_forearms():
    """Antebrazo derecho e izquierdo (6×12 cm -> 96×192 px, sprite 144×240, pivote [24, 216])."""
    w, h = 144, 240
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    cx = w // 2

    # Muñeca de piel asomando abajo
    d.rounded_rectangle([cx - 32, 198, cx + 32, h - 14], radius=10, fill=SKIN)
    shade_cylinder(im, SKIN, SKIN_HI, SKIN_SH, SKIN_SH2)

    # Manga lila cónica (más angosta en el puño abajo, más ancha hacia el codo arriba)
    sleeve = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    ds = ImageDraw.Draw(sleeve)
    pts = [
        (cx - 50, 20), (cx + 50, 20), (cx + 46, 110), (cx + 40, 206),
        (cx - 40, 206), (cx - 46, 110),
    ]
    ds.polygon(pts, fill=CARD)
    shade_cylinder(sleeve, CARD, CARD_L, CARD_D, CARD_D2, knit=True, seam_x=0.5)
    im.alpha_composite(sleeve)

    d = ImageDraw.Draw(im)
    # Puño acanalado en y=2cm (py=184..208)
    d.rounded_rectangle([cx - 42, 174, cx + 42, 208], radius=6, fill=CARD_D, outline=CARD_D2, width=2)
    for x in range(cx - 36, cx + 38, 6):
        d.line([(x, 178), (x, 204)], fill=CARD_L if x < cx else CARD_D2, width=2)
    # Pliegues de tela junto al codo arriba
    d.arc([cx - 42, 28, cx + 42, 56], start=10, end=170, fill=CARD_D2, width=3)
    d.arc([cx - 38, 46, cx + 38, 72], start=15, end=165, fill=CARD_D, width=2)

    im_r = add_outline(im, INK, thickness=2)
    save_img(im_r, 'body_forearm_r/micro_idle.png')
    save_img(im_r.transpose(Image.FLIP_LEFT_RIGHT), 'body_forearm_l/micro_idle.png')


def body_upper_arms():
    """Brazo derecho e izquierdo (6×12 cm -> 96×192 px, sprite 144×240, pivote [24, 216])."""
    w, h = 144, 240
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    cx = w // 2

    # Silueta de brazo superior en manga lila (curva del hombro arriba, codo abajo)
    pts = [
        (cx - 54, 34), (cx - 34, 18), (cx + 34, 18), (cx + 54, 34),
        (cx + 52, 130), (cx + 46, h - 18), (cx - 46, h - 18), (cx - 52, 130),
    ]
    d.polygon(pts, fill=CARD)
    shade_cylinder(im, CARD, CARD_L, CARD_D, CARD_D2, knit=True, seam_x=0.5)

    d = ImageDraw.Draw(im)
    # Costura superior del hombro y pliegue del codo abajo
    d.arc([cx - 50, 20, cx + 50, 62], start=195, end=345, fill=CARD_D2, width=4)
    d.arc([cx - 40, h - 54, cx + 40, h - 22], start=15, end=165, fill=CARD_D2, width=3)

    im_r = add_outline(im, INK, thickness=2)
    save_img(im_r, 'body_upper_arm_r/micro_idle.png')
    save_img(im_r.transpose(Image.FLIP_LEFT_RIGHT), 'body_upper_arm_l/micro_idle.png')


def body_thighs():
    """Muslo izquierdo y derecho (14×34 cm -> 224×544 px, sprite 272×592, pivote [24, 568])."""
    w, h = 272, 592
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    cx = w // 2

    # Pierna de piel (muslo afinándose hacia la rodilla abajo)
    leg = [
        (cx - 100, 180), (cx + 100, 180), (cx + 88, 370), (cx + 76, 496),
        (cx + 72, h - 22), (cx - 72, h - 22), (cx - 76, 496), (cx - 88, 370),
    ]
    d.polygon(leg, fill=SKIN)
    shade_cylinder(im, SKIN, SKIN_HI, SKIN_SH, SKIN_SH2)

    # Media crema asomando en la parte inferior (bajo la rodilla)
    sock = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    dsk = ImageDraw.Draw(sock)
    dsk.polygon([(cx - 76, 520), (cx + 76, 520), (cx + 72, h - 22), (cx - 72, h - 22)], fill=SOCK)
    shade_cylinder(sock, SOCK, SOCK_HI, SOCK_D, SOCK_D2, rib=6)
    im.alpha_composite(sock)

    d = ImageDraw.Draw(im)
    # Contorno sutil de la rótula (rodilla)
    d.arc([cx - 38, 434, cx + 38, 494], start=20, end=160, fill=SKIN_SH, width=4)
    d.ellipse([cx - 24, 444, cx + 10, 470], fill=SKIN_HI)
    # Elástico superior de la media
    d.line([(cx - 76, 520), (cx + 76, 520)], fill=INK, width=2)
    d.line([(cx - 76, 534), (cx + 76, 534)], fill=SOCK_D2, width=2)

    # Sombra proyectada por la falda sobre la piel del muslo
    d.polygon([(cx - 102, 286), (cx + 102, 286), (cx + 96, 322), (cx - 96, 322)], fill=SKIN_SH2)

    # Falda tableada cubriendo la mitad superior (ruedo cerca del agarre en y=16cm -> py=312)
    skirt = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    ds = ImageDraw.Draw(skirt)
    ds.polygon([(cx - 108, 22), (cx + 108, 22), (cx + 116, 294), (cx - 116, 294)], fill=SKIRT)
    shade_cylinder(skirt, SKIRT, SKIRT_L, SKIRT_D, SKIRT_D2)
    ds = ImageDraw.Draw(skirt)
    for px_x in range(cx - 88, cx + 92, 44):
        ds.line([(px_x, 22), (px_x + (px_x - cx) // 10, 294)], fill=SKIRT_D2, width=4)
        ds.line([(px_x + 8, 22), (px_x + 8 + (px_x - cx) // 10, 294)], fill=SKIRT_L, width=3)
    ds.line([(cx - 116, 294), (cx + 116, 294)], fill=INK, width=3)
    im.alpha_composite(skirt)

    im_l = add_outline(im, INK, thickness=2)
    save_img(im_l, 'body_thigh_l/micro_idle.png')
    save_img(im_l.transpose(Image.FLIP_LEFT_RIGHT), 'body_thigh_r/micro_idle.png')


def body_calves():
    """Pantorrilla izquierda y derecha (8×16 cm -> 128×256 px, sprite 176×304, pivote [24, 280])."""
    w, h = 176, 304
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    cx = w // 2

    # Piel de la rodilla arriba
    d.polygon([(cx - 60, 20), (cx + 60, 20), (cx + 64, 58), (cx - 64, 58)], fill=SKIN)
    shade_cylinder(im, SKIN, SKIN_HI, SKIN_SH, SKIN_SH2)

    # Curva anatómica de la pantorrilla en media crema acanalada (ancha arriba, tobillo fino abajo)
    sock = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    ds = ImageDraw.Draw(sock)
    calf_pts = [
        (cx - 62, 52), (cx + 62, 52), (cx + 66, 114), (cx + 56, 188),
        (cx + 42, 260), (cx + 44, h - 18), (cx - 44, h - 18), (cx - 42, 260),
        (cx - 56, 188), (cx - 66, 114),
    ]
    ds.polygon(calf_pts, fill=SOCK)
    shade_cylinder(sock, SOCK, SOCK_HI, SOCK_D, SOCK_D2, rib=6, seam_x=0.5)
    im.alpha_composite(sock)

    d = ImageDraw.Draw(im)
    # Puño elástico superior de la media
    d.rectangle([cx - 62, 52, cx + 62, 74], fill=SOCK_D)
    for x in range(cx - 58, cx + 60, 6):
        d.line([(x, 54), (x, 72)], fill=SOCK_HI if x < cx else SOCK_D2, width=2)
    d.line([(cx - 62, 52), (cx + 62, 52)], fill=INK, width=2)
    d.line([(cx - 63, 74), (cx + 63, 74)], fill=SOCK_D2, width=2)

    # Borde del calzado violeta envolviendo el tobillo abajo
    d.polygon([(cx - 48, 256), (cx + 48, 256), (cx + 52, h - 16), (cx - 52, h - 16)], fill=SHOE)
    d.line([(cx - 48, 256), (cx + 48, 256)], fill=SHOE_L, width=3)

    im_l = add_outline(im, INK, thickness=2)
    save_img(im_l, 'body_calf_l/micro_idle.png')
    save_img(im_l.transpose(Image.FLIP_LEFT_RIGHT), 'body_calf_r/micro_idle.png')


def body_feet():
    """Pie izquierdo y derecho (10×24 cm -> 160×384 px, sprite 208×432, pivote [24, 408])."""
    w, h = 208, 432
    im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    cx = w // 2

    # Silueta de zapato/pantufla visto desde arriba: puntera redondeada arriba, arco, talón abajo
    shoe_pts = [
        (cx - 52, 46), (cx - 24, 24), (cx + 28, 24), (cx + 62, 56),
        (cx + 76, 134), (cx + 64, 246), (cx + 66, 344), (cx + 46, h - 26),
        (cx - 42, h - 26), (cx - 60, 344), (cx - 56, 246), (cx - 68, 134),
    ]
    d.polygon(shoe_pts, fill=SHOE)
    d.ellipse([cx - 68, 22, cx + 74, 152], fill=SHOE)
    d.ellipse([cx - 58, 326, cx + 64, h - 20], fill=SHOE)
    shade_cylinder(im, SHOE, SHOE_L, SHOE_D, SHOE_D2)

    d = ImageDraw.Draw(im)
    # Pespunte de la suela alrededor
    d.arc([cx - 62, 28, cx + 68, 150], start=185, end=355, fill=SHOE_HI, width=3)
    # Brillo lustroso en la puntera
    d.ellipse([cx - 42, 48, cx - 2, 102], fill=SHOE_HI)
    d.ellipse([cx - 32, 56, cx - 10, 86], fill=WHITE)

    # Abertura del zapato mostrando la media crema adentro (zona del empeine/tobillo)
    d.ellipse([cx - 44, 174, cx + 48, 358], fill=SOCK, outline=INK, width=3)
    # Sombreado de la media dentro del zapato
    d.arc([cx - 40, 178, cx + 44, 354], start=310, end=120, fill=SOCK_D, width=10)
    for x in range(cx - 30, cx + 36, 8):
        d.line([(x, 192), (x, 344)], fill=SOCK_D, width=2)

    # Correa Mary-Jane cruzando el empeine con hebilla lateral
    d.rounded_rectangle([cx - 64, 176, cx + 68, 206], radius=6, fill=SHOE, outline=INK, width=2)
    d.line([(cx - 58, 182), (cx + 62, 182)], fill=SHOE_L, width=3)
    d.rounded_rectangle([cx - 60, 180, cx - 44, 202], radius=3, fill=BTN_HI, outline=INK, width=2)

    im_l = add_outline(im, INK, thickness=2)
    save_img(im_l, 'body_foot_l/micro_idle.png')
    save_img(im_l.transpose(Image.FLIP_LEFT_RIGHT), 'body_foot_r/micro_idle.png')


def main(src):
    global SRC
    SRC = src
    futon_micro_tile()
    pillow_micro_tile()
    grip_point()
    body_head()
    body_neck()
    body_shoulders()
    body_chest()
    body_back()
    body_sides()
    body_hips()
    body_hands()
    body_forearms()
    body_upper_arms()
    body_thighs()
    body_calves()
    body_feet()
