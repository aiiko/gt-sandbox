"""Lote de sprites de M15: tiny en cuatro direcciones, palma elevada y taza
en mano (la giantess macro está en giantess.py)."""
import os
from PIL import Image, ImageOps
from px import Canvas, load, hexc
import hand

SRC = BASE = None  # los fija main
INK = '#3F365C'


def save(img, rel):
    path = os.path.join(SRC, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    img.save(path, optimize=True)


# ---------------------------------------------------------------- tiny

TINY_PALETTES = {
    'tiny': {'b': '#456A90', 'c': '#89BED7', 'd': '#507AA0', 'h': '#3D9295', 'i': '#9CE0D6'},
    'tiny_lavender': None,  # se lee del original
    'tiny_ochre': {'b': '#855449', 'c': '#E2A57A', 'd': '#A56954', 'h': '#537F73', 'i': '#C9DB98'},
}
COMMON = {'a': INK, 'e': '#FFD8B9', 'f': '#D4898C', 'g': '#FF9CAD', 'j': '#F8B99D',
          'k': '#6C628E', 'n': '#2E5E86', 'o': '#FFFFFF'}


def palette_from(path):
    """Lee los colores de pelo y bufanda del sprite original."""
    im = load(path)
    roles = {'b': (4, 2), 'c': (7, 2), 'd': (6, 5), 'h': (5, 9), 'i': (6, 10)}
    return {k: '#%02X%02X%02X' % im.getpixel(xy)[:3] for k, xy in roles.items()}


UPPER = {
    'south': [
        "................",
        ".....aaaaaa.....",
        "....abbccbba....",
        "....abcbbccba...",
        "...abcbbdbbbba..",
        "...abdddddeeba..",
        "...abdddaeaeba..",
        "...abdeenenfba..",
        "...abeegeeegaa..",
        "....ahaaaaaha...",
        "....hiiiiiih....",
        "...ahhhhhhhhh...",
        "..ahjhhhhhhjh...",
    ],
    'north': [
        "................",
        ".....aaaaaa.....",
        "....abbccbba....",
        "....abcbbccba...",
        "...abcbbbbbbba..",
        "...abbdbbbdbba..",
        "...abdbbbbbdba..",
        "...abbdbbbdbba..",
        "...aabbbbbbbaa..",
        "....ahaaaaaha...",
        "....hhhhiihh....",
        "...ahhhhhiihh...",
        "..ahjhhhhihhja..",
    ],
    'east': [
        "................",
        ".....aaaaaa.....",
        "....abbccbba....",
        "...abcbbccbba...",
        "...abbbbdbbbba..",
        "...abbbddeeeba..",
        "...abbbdeeaeea..",
        "...abbbdeeneea..",
        "...aabbbeeegea..",
        "....ahaaaaaha...",
        "...ihhiiiiih....",
        "..iahhhhhhhhh...",
        "....ahhhjhhha...",
    ],
}

# Piernas: filas 13 a 15. Cuatro cuadros de caminata y una pose parada.
LEGS_FRONT = {
    'stand': [".....akkkka.....", ".....akaaka.....", ".....aaaaaa....."],
    'left':  [".....akkkka.....", ".....aaaaka.....", "........aaa....."],
    'right': [".....akkkka.....", ".....akaaaa.....", ".....aaa........"],
}
LEGS_SIDE = {
    'stand': [".....akkkka.....", "......akka......", "......aaaaa....."],
    'apart': [".....akkkka.....", ".....ak..ka.....", "....aaa..aaa...."],
    'pass':  [".....akkkka.....", "......akka......", ".......aaaa....."],
}


def grid_to_canvas(c, rows, pal, dx=0, dy=0):
    for y, row in enumerate(rows):
        assert len(row) == 16, (row, len(row))
        for x, ch in enumerate(row):
            if ch != '.':
                c.px(x + dx, y + dy, pal[ch])


def shadow(c, w=10):
    c.ellipse(8, 16, w // 2, 1, hexc(INK, 70))


def tiny_frame(pal, facing, legs, bob=0, arms_up=False, fall=None):
    c = Canvas(16, 18)
    if not fall:
        shadow(c, 10)
    side = facing in ('east', 'west')
    upper = UPPER['east' if side else facing]
    table = LEGS_SIDE if side else LEGS_FRONT
    grid_to_canvas(c, table[legs], pal, dy=13)
    if bob:
        grid_to_canvas(c, [".....akkkka....."], pal, dy=12)
    rows = [r for r in upper]
    if arms_up or fall:
        # brazos arriba: las manos suben a los costados de la cabeza
        rows[12] = rows[12].replace('j', 'h')
        rows = [r for r in rows]
    grid_to_canvas(c, rows, pal, dy=-bob)
    if arms_up or fall:
        lift = 1 if fall == 2 else 0
        for x in (2, 13) if not side else (11, 13):
            c.vline(x, 8 - bob - lift, 4, pal['h'])
            c.px(x, 7 - bob - lift, pal['j'])
            c.px(x, 6 - bob - lift, INK)
    if fall:
        # boca abierta y gota de sudor
        if facing != 'north':
            mx = 8 if not side else 11
            c.px(mx, 8, INK)
        c.px(13, 2, '#9CE0D6')
        c.px(13, 3, '#FFFFFF')
    im = c.im
    if facing == 'west':
        im = ImageOps.mirror(im)
    return im


def tinies():
    for name, pal in TINY_PALETTES.items():
        if pal is None:
            pal = palette_from(os.path.join(BASE, name, 'micro_idle.png'))
        full = dict(COMMON, **pal)
        for facing in ('south', 'north', 'east', 'west'):
            side = facing in ('east', 'west')
            stand = 'stand'
            save(tiny_frame(full, facing, stand), f'{name}/micro_idle_{facing}.png')
            seq = ['apart', 'pass', 'apart', 'pass'] if side else ['left', 'stand', 'right', 'stand']
            bobs = [0, 1, 0, 1]
            for i, (legs, bob) in enumerate(zip(seq, bobs)):
                save(tiny_frame(full, facing, legs, bob), f'{name}/micro_walk_{facing}_{i + 1}.png')
                save(tiny_frame(full, facing, legs, bob, arms_up=True), f'{name}/micro_carry_{facing}_{i + 1}.png')
        for i in (1, 2):
            save(tiny_frame(full, 'south', 'left' if i == 1 else 'right', 0, fall=i), f'{name}/micro_fall_{i}.png')


# ------------------------------------------------------------ palma y taza

PALM_SIZE = 480  # cuadro cuadrado: el pivote queda en el centro en las cuatro orientaciones
# Rotación de la imagen original (dibujada mirando al norte) para cada orientación.
PALM_TURNS = {'north': 0, 'east': -90, 'south': 180, 'west': 90}


def palm_shadowless():
    return hand.draw()


def palm_square(turn):
    """Mano con antebrazo, con el centro de la palma en el centro del cuadro."""
    sq = Image.new('RGBA', (PALM_SIZE, PALM_SIZE), (0, 0, 0, 0))
    sq.alpha_composite(palm_shadowless(), (PALM_SIZE // 2 - hand.W // 2, PALM_SIZE // 2 - hand.H // 2))
    return sq.rotate(turn) if turn else sq


def with_shadow(img, offset, alpha):
    out = Image.new('RGBA', img.size, (0, 0, 0, 0))
    mask = img.split()[3].point(lambda a: alpha if a else 0)
    sh = Image.new('RGBA', img.size, hexc(INK))
    sh.putalpha(mask)
    out.alpha_composite(sh, offset)
    out.alpha_composite(img, (0, 0))
    return out


def palm():
    for facing, turn in PALM_TURNS.items():
        sq = palm_square(turn)
        # La luz viene siempre de arriba a la izquierda: la sombra no gira con la mano.
        save(with_shadow(sq, (2, 4), 120), f'palm/micro_rested_{facing}.png')
        save(with_shadow(sq, (10, 16), 80), f'palm/micro_raised_{facing}.png')


def cup():
    micro = load(os.path.join(BASE, 'cup/micro_idle.png'))
    save(micro, 'cup/micro_idle.png')
    held = micro.copy()
    # en la mano no hay sombra de contacto: se quitan los tonos de sombra bajo la taza
    for y in range(micro.height):
        for x in range(micro.width):
            p = micro.getpixel((x, y))
            if p[3] and p[:3] in {(0xBD, 0x89, 0x7F), (0xA7, 0x72, 0x79), (0xA9, 0x77, 0x79)} and y > 60:
                held.putpixel((x, y), (0, 0, 0, 0))
    save(held, 'cup/micro_held.png')


def plate_micro(c, cx, cy, shadow):
    """Plato de postre visto desde arriba con una galleta y migas."""
    if shadow:
        c.ellipse(cx + 3, cy + 5, 45, 27, '#BD897F')
    c.ellipse(cx, cy, 45, 27, INK)
    c.ellipse(cx, cy, 44, 26, '#E8DCCB')
    c.ellipse(cx, cy - 1, 43, 25, '#FFF6EA')
    c.ellipse(cx, cy + 1, 32, 18, '#F2E6D6')
    c.ellipse(cx, cy, 31, 17, '#FFF9F0')
    # galleta
    c.ellipse(cx - 6, cy - 2, 12, 8, INK)
    c.ellipse(cx - 6, cy - 2, 11, 7, '#D9A066')
    c.ellipse(cx - 7, cy - 3, 9, 5, '#E8B77D')
    for dx, dy in ((-10, -4), (-4, -6), (-2, 0), (-9, 1)):
        c.px(cx + dx, cy + dy, '#8F5A3C')
    for dx, dy in ((10, 4), (14, -3), (8, 8), (18, 2)):
        c.px(cx + dx, cy + dy, '#D9A066')


def plate():
    c = Canvas(98, 64)
    plate_micro(c, 47, 29, True)
    save(c.im, 'plate/micro_idle.png')
    c = Canvas(98, 64)
    plate_micro(c, 47, 29, False)
    save(c.im, 'plate/micro_held.png')


def tiny_shadow():
    """Sombra sola del tiny, opaca: la presentación le da la transparencia."""
    c = Canvas(16, 18)
    c.ellipse(8, 16, 5, 1, INK)
    for name in TINY_PALETTES:
        save(c.im, f'{name}/micro_shadow.png')


def soft_silhouette(img):
    """Silueta opaca con el borde punteado, para una sombra difusa."""
    a = img.split()[3]
    w, h = img.size
    out = Image.new('RGBA', img.size, (0, 0, 0, 0))
    src, dst = a.load(), out.load()
    ink = hexc(INK)
    for y in range(h):
        for x in range(w):
            if not src[x, y]:
                continue
            edge = 0
            for r in (1, 2, 3):
                if any(not (0 <= x + dx < w and 0 <= y + dy < h) or not src[x + dx, y + dy]
                       for dx, dy in ((r, 0), (-r, 0), (0, r), (0, -r))):
                    edge = r
                    break
            if edge == 0 or (edge == 2 and (x + y) % 2 == 0):
                dst[x, y] = ink
    return out


BAYER4 = [[0, 8, 2, 10], [12, 4, 14, 6], [3, 11, 1, 9], [15, 7, 13, 5]]


def fade_forearm(img, start=hand.W - 24, length=110):
    """Desvanece la sombra a lo largo del antebrazo con un tramado ordenado."""
    px = img.load()
    for y in range(img.height):
        for x in range(start, img.width):
            keep = 1 - (x - start) / length
            if px[x, y][3] and BAYER4[y % 4][x % 4] / 16 >= keep:
                px[x, y] = (0, 0, 0, 0)
    return img


def palm_shadows():
    sil = fade_forearm(soft_silhouette(palm_shadowless()))
    for facing, turn in PALM_TURNS.items():
        sq = Image.new('RGBA', (PALM_SIZE, PALM_SIZE), (0, 0, 0, 0))
        sq.alpha_composite(sil, (PALM_SIZE // 2 - hand.W // 2, PALM_SIZE // 2 - hand.H // 2))
        save(sq.rotate(turn) if turn else sq, f'palm/micro_shadow_{facing}.png')


def pass_sign():
    """Señal del paso: dos anillos unidos (abierto) o separados (cerrado)."""
    teal, teal_hi, teal_sh = '#76D8CB', '#D8F6F0', '#3D9295'
    muted, muted_sh = '#B3A7EC', '#7969A4'

    def ring(c, cx, fill, shade):
        c.ellipse(cx, 5, 4, 4, INK)
        c.ellipse(cx, 5, 3, 3, fill)
        c.ellipse(cx, 5, 1, 1, INK)
        c.px(cx - 2, 3, teal_hi if fill == teal else '#E4DEFA')
        c.px(cx + 2, 7, shade)

    for name, fill, shade, bar in (('open_1', teal, teal_sh, teal), ('open_2', teal, teal_sh, teal_hi),
                                   ('closed', muted, muted_sh, None)):
        c = Canvas(22, 11)
        c.ellipse(11, 9, 9, 1, hexc(INK, 70))
        if bar:
            c.rect(6, 3, 10, 5, INK)
            c.rect(6, 4, 10, 3, bar)
            c.hline(6, 4, 10, teal_hi)
        else:
            # cabos cortados: la unión falta
            c.rect(7, 4, 2, 3, INK)
            c.rect(13, 4, 2, 3, INK)
        ring(c, 5, fill, shade)
        ring(c, 16, fill, shade)
        save(c.im, f'pass/micro_{name}.png')


def remains():
    """Restos de un aplastado: silueta gris aplanada, con la bufanda menta."""
    c = Canvas(18, 8)
    c.ellipse(9, 5, 8, 2, hexc(INK, 70))
    grey, grey_hi, grey_sh = '#A9A3BC', '#C9C4D8', '#7D7794'
    c.ellipse(9, 4, 7, 2, INK)
    c.ellipse(9, 4, 6, 1, grey)
    c.hline(4, 3, 7, grey_hi)
    c.ellipse(4, 4, 2, 2, INK)  # cabeza aplastada
    c.ellipse(4, 4, 1, 1, grey)
    c.px(3, 3, grey_hi)
    c.hline(6, 5, 2, '#9CE0D6')  # bufanda
    c.px(8, 6, '#9CE0D6')
    c.px(13, 5, grey_sh)
    c.px(15, 4, grey_sh)
    save(c.im, 'remains/micro_idle.png')
    m = Canvas(4, 2)
    m.hline(0, 0, 4, INK)
    m.hline(1, 1, 2, grey)
    save(m.im, 'remains/macro_idle.png')


def main(src, base):
    global SRC, BASE
    SRC, BASE = src, base
    remains()
    tinies()
    tiny_shadow()
    palm()
    palm_shadows()
    pass_sign()
    cup()
    plate()
