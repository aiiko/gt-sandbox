"""Arte micro del mapa corporal de la giantess, futón y almohada (propuesta 19).

Piezas generadas en content/art/src/casa/ (solo vista micro):
- futon/micro_tile.png (64 × 64 px): algodón acolchado crema del futón.
- pillow/micro_tile.png (64 × 64 px): lino lavanda-crema de la almohada.
- body_hair/micro_tile.png (64 × 64 px): mechones de cabello violeta.
- body_skin/micro_tile.png (64 × 64 px): piel cálida (cuello y nuca).
- body_hand/micro_tile.png (64 × 64 px): piel de mano con pliegues suaves.
- body_cardigan/micro_tile.png (64 × 64 px): tejido de punto lila del cárdigan.
- body_sleeve/micro_tile.png (48 × 64 px): manga acanalada con costura vertical.
- body_chest/micro_tile.png (64 × 64 px): frente del cárdigan con blusa y botones.
- body_skirt/micro_tile.png (64 × 64 px): falda tableada violeta (cintura/caderas).
- body_thigh/micro_tile.png (64 × 64 px): pliegues de falda en muslos.
- body_sock/micro_tile.png (48 × 64 px): media crema acanalada con costura.
- body_shoe/micro_tile.png (64 × 64 px): calzado/pantufla violeta con pespunte.
- grip_point/micro_idle.png (16 × 16 px, pivote [8, 8]): punto de agarre.
"""
import math
import os

from px import Canvas, hexc

SRC = None  # lo fija main
INK = '#3F365C'


def save(c, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    c.save(path)


def futon_micro_tile():
    """Futón de algodón crema acolchado con pespuntes en rombo."""
    w, h = 64, 64
    base, hi, sh, stitch = hexc('#FFF4E2'), hexc('#FFFBF2'), hexc('#F0DEC6'), hexc('#D2BCA2')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            # trama suave de algodón
            weave = (x + y) % 4
            if weave == 0:
                col = hi
            elif weave == 2:
                col = sh
            else:
                col = base
            # líneas diagonales de acolchado en rombo cada 32 px
            d1 = (x + y) % 32
            d2 = (x - y) % 32
            if (d1 == 0 or d2 == 0) and (x % 4 < 2):
                col = stitch
            elif d1 == 1 or d2 == 1:
                col = sh
            px[x, y] = col
    # botón de acolchado en las intersecciones
    for cx, cy in ((0, 0), (32, 0), (16, 16), (48, 16), (0, 32), (32, 32), (16, 48), (48, 48)):
        for dy in (-1, 0, 1):
            for dx in (-1, 0, 1):
                if abs(dx) + abs(dy) <= 1:
                    px[(cx + dx) % w, (cy + dy) % h] = stitch
    save(c, 'futon/micro_tile.png')


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
    # costura sutil horizontal
    for x in range(0, w, 6):
        c.hline(x, 31, 3, seam)
    save(c, 'pillow/micro_tile.png')


def body_hair_micro_tile():
    """Cabello largo violeta de la giantess: mechones ondulados y brillo."""
    w, h = 64, 64
    hair, dark, light, hi = hexc('#5D508D'), hexc('#4A3F75'), hexc('#8677B7'), hexc('#9D86C9')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        wave = round(math.sin(y * 2 * math.pi / h) * 2.2)
        for x in range(w):
            u = (x + wave) % 16
            if u in (0, 1):
                col = dark
            elif u in (5, 6):
                col = hi if (10 <= y <= 20 or 42 <= y <= 50) else light
            elif u in (7, 8):
                col = light
            elif u == 14:
                col = dark
            else:
                col = hair
            px[x, y] = col
    save(c, 'body_hair/micro_tile.png')


def body_skin_micro_tile():
    """Piel cálida de la giantess (cuello y nuca) con sombreado suave."""
    w, h = 64, 64
    base, hi, sh, crease = hexc('#FFD8B9'), hexc('#FFEAD6'), hexc('#F4B49A'), hexc('#E8A696')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            # gradiente cilíndrico suave (luz desde la izquierda)
            if x < 14:
                col = hi if (x + y) % 5 != 0 else base
            elif x >= 50:
                col = sh if (x + y) % 4 != 0 else base
            else:
                col = hi if (x + y * 2) % 23 == 0 else base
            px[x, y] = col
    # pliegue sutil del cuello
    for x in range(8, w - 8):
        if x % 3 != 0:
            px[x, 32] = crease
    save(c, 'body_skin/micro_tile.png')


def body_hand_micro_tile():
    """Piel de mano (mano izquierda / dorso derecho) con pliegues delicados."""
    w, h = 64, 64
    base, palm, hi, sh, crease = (
        hexc('#FFD8B9'), hexc('#FCCDAE'), hexc('#FFEAD6'), hexc('#F4B49A'), hexc('#E8A696')
    )
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            if x < 10 or y < 8:
                col = hi
            elif x >= 54 or y >= 56:
                col = sh
            elif (x + y) % 11 == 0:
                col = palm
            else:
                col = base
            px[x, y] = col
    # líneas suaves de pliegue de nudillos/palma
    for x in range(6, w - 6):
        y1 = 22 + int(round(math.sin(x * math.pi / w) * 3))
        y2 = 44 - int(round(math.sin(x * math.pi / w) * 2))
        if x % 2 == 0:
            px[x, y1] = crease
            px[x, y2] = crease
    save(c, 'body_hand/micro_tile.png')


def body_cardigan_micro_tile():
    """Tejido de punto lila del cárdigan (hombros, espalda y costados)."""
    w, h = 64, 64
    base, light, dark, seam = hexc('#9A8BE3'), hexc('#B4A6F0'), hexc('#7A6BC6'), hexc('#67588F')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            # punto jersey en V (repetición 8 × 8)
            ux, uy = x % 8, y % 8
            if (ux == 1 and uy in (1, 2, 3)) or (ux == 2 and uy in (3, 4, 5)):
                col = light
            elif (ux == 5 and uy in (1, 2, 3)) or (ux == 4 and uy in (3, 4, 5)):
                col = dark
            elif ux == 3 and uy == 6:
                col = dark
            else:
                col = base
            px[x, y] = col
    # costura vertical sutil de prenda
    for y in range(0, h, 4):
        px[31, y] = seam
        px[32, (y + 2) % h] = seam
    save(c, 'body_cardigan/micro_tile.png')


def body_sleeve_micro_tile():
    """Manga lila acanalada para trepar (antebrazos y brazos)."""
    w, h = 48, 64
    base, light, dark, deep, seam = (
        hexc('#9A8BE3'), hexc('#C8B9F5'), hexc('#7A6BC6'), hexc('#67588F'), hexc('#544777')
    )
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            if x <= 1 or x >= w - 2:
                col = deep
            elif x < 8:
                col = light if (y % 4 != 0) else base
            elif x >= w - 8:
                col = dark if (y % 4 != 0) else deep
            else:
                # canalé vertical + puntadas horizontales cada 8 px
                rib = x % 6
                if rib == 0:
                    col = light
                elif rib == 3:
                    col = dark
                else:
                    col = base
                if y % 8 == 0 and rib not in (0, 3):
                    col = dark
            px[x, y] = col
    # costura central trepable y marcas de puño cada 32 px
    cx = w // 2
    for y in range(h):
        px[cx - 1, y] = seam
        px[cx, y] = light if y % 4 < 2 else dark
    for y in (0, 32):
        c.hline(3, y, w - 6, hexc('#67588F', 140))
    save(c, 'body_sleeve/micro_tile.png')


def body_chest_micro_tile():
    """Frente del cárdigan lila con tapeta de blusa crema y botones."""
    w, h = 64, 64
    card, card_l, card_d = hexc('#9A8BE3'), hexc('#B4A6F0'), hexc('#7A6BC6')
    blouse, blouse_d = hexc('#FFF1D8'), hexc('#E3D3C9')
    btn, btn_hi = hexc('#FF963D'), hexc('#FFC88A')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            if 26 <= x <= 37:
                # tapeta central crema de la blusa
                if x in (26, 37):
                    col = card_d
                elif x in (27, 36):
                    col = blouse_d
                else:
                    col = blouse
            else:
                ux, uy = x % 8, y % 8
                if (ux == 1 and uy in (1, 2, 3)) or (ux == 2 and uy in (3, 4)):
                    col = card_l
                elif ux == 5 and uy in (2, 3, 4):
                    col = card_d
                else:
                    col = card
            px[x, y] = col
    # botón mandarina del cárdigan
    c.ellipse(32, 32, 3, 3, btn)
    px[31, 31] = btn_hi
    px[33, 33] = hexc(INK)
    save(c, 'body_chest/micro_tile.png')


def body_skirt_micro_tile():
    """Falda tableada violeta oscuro (cintura y caderas)."""
    w, h = 64, 64
    base, light, dark, ink = hexc('#67588F'), hexc('#7969A4'), hexc('#544777'), hexc(INK)
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            pleat = x % 16
            if pleat == 0:
                col = ink
            elif pleat in (1, 2):
                col = dark
            elif pleat in (3, 4, 5):
                col = light
            elif pleat >= 14:
                col = dark
            else:
                col = base
            # sutil textura diagonal de lana
            if (x + y) % 12 == 0 and pleat not in (0, 1):
                col = light if col == base else col
            px[x, y] = col
    # pespunte de cintura
    for x in range(2, w, 6):
        c.hline(x, 6, 3, light)
    save(c, 'body_skirt/micro_tile.png')


def body_thigh_micro_tile():
    """Pliegues de falda/muslo con sombreado vertical suave y empalme continuo."""
    w, h = 64, 64
    base, light, dark, deep = hexc('#67588F'), hexc('#8373B0'), hexc('#544777'), hexc('#453A63')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        sway = round(math.sin(y * 2 * math.pi / h) * 1.2)
        for x in range(w):
            pleat = (x + sway) % 16
            if pleat == 0:
                col = deep
            elif pleat in (1, 2):
                col = dark
            elif pleat in (4, 5, 6):
                col = light
            elif pleat == 15:
                col = dark
            else:
                col = base
            px[x, y] = col
    save(c, 'body_thigh/micro_tile.png')


def body_sock_micro_tile():
    """Media crema acanalada (pantorrillas) con costura lateral trepable."""
    w, h = 48, 64
    base, hi, sh, seam = hexc('#FFF1D8'), hexc('#FFF9EE'), hexc('#DCCFE0'), hexc('#B8A8C2')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            if x <= 1 or x >= w - 2:
                col = seam
            else:
                rib = x % 4
                if rib == 0:
                    col = hi
                elif rib == 2:
                    col = sh
                else:
                    col = base
                if y % 6 == 0 and rib == 1:
                    col = sh
            px[x, y] = col
    # costura vertical central para trepar
    cx = w // 2
    for y in range(h):
        px[cx, y] = seam if y % 3 != 0 else sh
    save(c, 'body_sock/micro_tile.png')


def body_shoe_micro_tile():
    """Calzado/pantufla violeta de la giantess con pespunte perimetral."""
    w, h = 64, 64
    base, light, dark, welt = hexc('#76659B'), hexc('#9586BC'), hexc('#5C4D7D'), hexc('#453861')
    stitch = hexc('#C8B9F5')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            if x < 6 or x >= w - 6:
                col = dark
            elif (x + y) % 14 == 0:
                col = light
            elif (x - y) % 16 == 0:
                col = dark
            else:
                col = base
            px[x, y] = col
    # ribete cosido de la puntera/lateral
    c.vline(6, 0, h, welt)
    c.vline(w - 7, 0, h, welt)
    for y in range(2, h, 6):
        c.vline(8, y, 3, stitch)
        c.vline(w - 9, y, 3, stitch)
    save(c, 'body_shoe/micro_tile.png')


def grip_point():
    """Marcador de punto de agarre (costura, puño, mechón) en vista micro."""
    c = Canvas(16, 16)
    ring, hi, core = '#76D8CB', '#CFF7F0', '#FFF7E5'
    # anillo turquesa con borde de tinta y centro cálido
    for y in range(16):
        for x in range(16):
            d = math.hypot(x - 7.5, y - 7.5)
            if 5.2 <= d <= 6.8:
                c.px(x, y, INK)
            elif 3.6 <= d < 5.2:
                c.px(x, y, hi if y < 7 else ring)
            elif 2.3 <= d < 3.6:
                c.px(x, y, INK)
            elif d <= 1.8:
                c.px(x, y, core)
    save(c, 'grip_point/micro_idle.png')


def main(src):
    global SRC
    SRC = src
    futon_micro_tile()
    pillow_micro_tile()
    body_hair_micro_tile()
    body_skin_micro_tile()
    body_hand_micro_tile()
    body_cardigan_micro_tile()
    body_sleeve_micro_tile()
    body_chest_micro_tile()
    body_skirt_micro_tile()
    body_thigh_micro_tile()
    body_sock_micro_tile()
    body_shoe_micro_tile()
    grip_point()
