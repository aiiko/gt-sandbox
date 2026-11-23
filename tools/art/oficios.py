"""Arte de Oficios, expediciones y cobertura micro (M53–M55, propuesta 18).

Piezas generadas en content/art/src/casa/ (solo vista micro):
- climb/micro_tile.png (32 × 160 px): textura lateral de pata/respaldo de madera.
- cable_duct/micro_tile.png (64 × 64 px): pasacables del zócalo con cable y borde.
- sofa_pocket/micro_tile.png (64 × 64 px): interior de tela salvia del bolsillo del sofá.
- threshold_refuge/micro_idle.png (44 × 32 px, pivote [22, 24]): refugio del umbral.
- lookout_post/micro_idle.png (24 × 24 px, pivote [12, 19]): puesto de vigía con banderín.
- expedition_marker/micro_idle.png (20 × 20 px, pivote [10, 15]): hito de expedición.
"""
import os

from px import Canvas, hexc

SRC = None  # lo fija main
INK = '#3F365C'


def save(c, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    c.save(path)


def climb_micro_tile():
    """Tira vertical de madera miel para patas y respaldos vistos de costado."""
    w, h = 32, 160
    base, light, dark, edge = hexc('#DDA86E'), hexc('#ECC08A'), hexc('#B8804C'), hexc('#8D5A32')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            if x == 0 or x == w - 1:
                col = edge
            elif x == 1 or x == 2:
                col = light
            elif x == w - 2 or x == w - 3:
                col = dark
            else:
                # vetas verticales suaves
                grain = (x * 3 + (y // 12) + ((x * y) // 97)) % 9
                if grain == 0:
                    col = dark
                elif grain == 4:
                    col = light
                else:
                    col = base
            px[x, y] = col
    # marcas horizontales suaves de altura (cada 32 px = 2 cm)
    for y in range(0, h, 32):
        c.hline(2, y, w - 4, hexc('#9F663D', 110))
    save(c, 'climb/micro_tile.png')


def cable_duct_micro_tile():
    """Pasacables del zócalo: canal estrecho con cable oscuro y borde de pared."""
    w, h = 64, 64
    wall, floor, shade = hexc('#5C5368'), hexc('#8C8396'), hexc('#6E657A')
    cable, cable_hi = hexc('#2D2A3A'), hexc('#524C66')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            if y < 6 or y >= h - 6:
                col = wall
            elif y < 12:
                col = shade
            elif 12 <= y <= 20:
                # cable grueso pegado a la pared superior
                col = cable_hi if y in (14, 15) and (x // 4) % 2 == 0 else cable
            else:
                col = shade if (x + y) % 16 == 0 else floor
            px[x, y] = col
    # junta de zócalo
    c.hline(0, 6, w, INK)
    c.hline(0, h - 6, w, INK)
    c.vline(32, 6, h - 12, hexc(INK, 90))
    save(c, 'cable_duct/micro_tile.png')


def sofa_pocket_micro_tile():
    """Interior del bolsillo del sofá: tela salvia más profunda y ribete cosido."""
    w, h = 64, 64
    base, light, dark, seam = hexc('#5F8C7E'), hexc('#709E90'), hexc('#4C7568'), hexc('#3C5F54')
    c = Canvas(w, h)
    px = c.im.load()
    for y in range(h):
        for x in range(w):
            band = (x + y) % 6
            if band == 0:
                col = light
            elif band == 3:
                col = dark
            else:
                col = base
            px[x, y] = col
    # costura de ribete
    for x in range(2, w, 6):
        c.hline(x, 4, 3, seam)
        c.hline(x, h - 5, 3, seam)
    save(c, 'sofa_pocket/micro_tile.png')


def threshold_refuge():
    """Refugio del umbral: pequeño toldo de tela y madera junto al zócalo."""
    c = Canvas(44, 32)
    c.ellipse(22, 26, 19, 4, hexc(INK, 65))
    # techo a dos aguas en tonos mandarina y crema
    cloth, cloth_hi, cloth_sh = '#E89B4C', '#F7BC74', '#C47530'
    wood, wood_sh = '#C48B54', '#8D5A32'
    for y in range(6, 22):
        half = 6 + (y - 6) * 12 // 16
        for x in range(22 - half, 22 + half + 1):
            stripe = ((x - 22) // 4) % 2 == 0
            col = cloth_hi if y < 10 and stripe else (cloth if stripe else cloth_sh)
            c.px(x, y, col)
        c.px(22 - half - 1, y, INK)
        c.px(22 + half + 1, y, INK)
    c.hline(16, 5, 13, INK)
    # postes y abertura oscura acogedora
    for px_x in (9, 34):
        c.rect(px_x, 22, 3, 5, wood)
        c.vline(px_x, 22, 5, INK)
        c.vline(px_x + 2, 22, 5, INK)
    c.rect(14, 15, 16, 11, hexc(INK, 220))
    c.hline(14, 25, 16, wood_sh)
    save(c, 'threshold_refuge/micro_idle.png')


def lookout_post():
    """Puesto de vigía: atalaya de fósforos con banderín azul y mandarina."""
    c = Canvas(24, 24)
    c.ellipse(12, 20, 8, 2, hexc(INK, 55))
    wood, wood_hi, wood_sh = '#DDA86E', '#F0C692', '#9F663D'
    # base de madera + mástil
    c.rrect(5, 15, 14, 5, wood, INK)
    c.hline(6, 16, 12, wood_hi)
    c.hline(6, 18, 12, wood_sh)
    c.rect(11, 4, 2, 12, wood)
    c.vline(10, 4, 12, INK)
    c.vline(13, 11, 5, INK)
    # banderín azul cobalto con punta mandarina
    for y in range(5, 11):
        span = 8 - abs(y - 7) * 2
        if span > 0:
            c.hline(13, y, span, '#326DE7' if y < 9 else '#FF963D')
            c.px(13 + span, y, INK)
    c.hline(13, 4, 5, INK)
    c.hline(13, 11, 5, INK)
    save(c, 'lookout_post/micro_idle.png')


def expedition_marker():
    """Hito de expedición: mojón de piedras y cinta turquesa."""
    c = Canvas(20, 20)
    c.ellipse(10, 16, 7, 2, hexc(INK, 50))
    stone, stone_hi, stone_sh = '#C9C4CE', '#ECE8F0', '#8F889A'
    # tres piedritas apiladas
    c.ellipse(10, 14, 5, 3, stone)
    c.ellipse(10, 11, 4, 2, stone_hi)
    c.ellipse(10, 8, 2, 2, stone)
    for x in range(6, 15):
        c.px(x, 16, stone_sh)
    # cinta turquesa arriba
    c.hline(8, 7, 5, '#76D8CB')
    c.px(13, 8, '#76D8CB')
    c.px(10, 5, INK)
    save(c, 'expedition_marker/micro_idle.png')


def main(src):
    global SRC
    SRC = src
    climb_micro_tile()
    cable_duct_micro_tile()
    sofa_pocket_micro_tile()
    threshold_refuge()
    lookout_post()
    expedition_marker()
