"""Giantess en macro a 80 px/m (62 × 76 px, pivote [30, 68]).

Chibi dibujado por partes: pelo largo con flequillo y moño, cárdigan lila
con blusa y lazo, falda tableada, medias y zapatos. Las poses (reposo,
caminata de cuatro cuadros, sentada en seiza, alcanzar y comer) mueven las
mismas partes, así que el cuerpo es el mismo en todos los cuadros. El oeste
es el espejo del este.
"""
import os

from shapes import Sprite

SRC = None  # lo fija main

W, H = 62, 76
PX, PY = 30, 68  # pivote: entre los pies

INK = '#3F365C'
HAIR, HAIR_D, HAIR_L, HAIR_HI = '#5D508D', '#4A3F75', '#8677B7', '#9D86C9'
SKIN, SKIN_D, BLUSH = '#FFD8B9', '#F2B79E', '#F49AAB'
EYE, EYE_D, EYE_L, WHITE = '#3D9295', '#2C6E78', '#76D8CB', '#FFFFFF'
MOUTH = '#A96F86'
CARD, CARD_L, CARD_D = '#9A8BE3', '#C8B9F5', '#7A6BC6'
BLOUSE, BLOUSE_D = '#FFF1D8', '#E3D3C9'
BOW, BOW_D = '#FF9CAD', '#CE8190'
SKIRT, SKIRT_D, SKIRT_L = '#67588F', '#544777', '#7969A4'
SOCK, SOCK_D = '#FFF1D8', '#DCCFE0'
SHOE, SHOE_L = '#76659B', '#9586BC'
SHADOW = (0x91, 0x6B, 0x8F, 255)
RICE, NORI = '#FFFFFF', '#2F3B4C'

# Cuánto baja el torso al sentarse en seiza.
SEAT_DROP = 14


# ------------------------------------------------------------------ partes

def shadow(s):
    sh = Sprite(W, H)
    sh.ell(PX - 21, PY - 4, PX + 21, PY + 4, SHADOW)
    s.under(sh)


def legs_front(s, dy, left, right):
    """Piernas de frente o de espaldas. left/right: cuánto sube cada pie."""
    for x0, lift in ((23, left), (33, right)):
        y1 = PY - lift
        s.rect(x0, 57 + dy, x0 + 4, y1 - 4, SKIN)          # rodilla
        s.rect(x0, 60 + dy, x0 + 4, y1 - 3, SOCK)          # media
        s.vline(x0 + 4, 59 + dy, y1 - 3, SOCK_D)
        s.rrect(x0 - 1, y1 - 3, x0 + 5, y1, SHOE)          # zapato
        s.hline(x0, x0 + 2, y1 - 3, SHOE_L)


def legs_side(s, dy, front, back):
    """Piernas de perfil (mirando al este): front/back desplazan cada pie."""
    for fx, lift, far in ((back, 0, True), (front, 0, False)):
        x0 = 28 + fx
        y1 = PY - lift
        sock = SOCK_D if far else SOCK
        s.rect(x0, 56 + dy, x0 + 4, y1 - 4, SKIN_D if far else SKIN)
        s.rect(x0, 59 + dy, x0 + 4, y1 - 3, sock)
        s.rrect(x0 - 1, y1 - 3, x0 + 6, y1, SHOE)
        s.hline(x0 + 1, x0 + 4, y1 - 3, SHOE_L)


def skirt(s, dy, narrow=0):
    top, bot = 50 + dy, 59 + dy
    s.poly([(19 + narrow, top), (41 - narrow, top), (45 - narrow, bot), (15 + narrow, bot)], SKIRT)
    # tablas: líneas que se abren hacia el ruedo
    for x in range(22 + narrow, 40 - narrow, 4):
        s.line([(x, top + 2), (x + (x - 30) // 5, bot - 1)], SKIRT_D)
    s.hline(16 + narrow, 44 - narrow, bot, SKIRT_D)
    s.hline(20 + narrow, 26, top + 1, SKIRT_L)


def torso_front(s, dy, back=False):
    s.rrect(18, 39 + dy, 42, 52 + dy, CARD, r=2)
    s.rect(19, 41 + dy, 20, 50 + dy, CARD_L)            # luz desde la izquierda
    s.vline(41, 41 + dy, 51 + dy, CARD_D)
    s.hline(19, 41, 52 + dy, CARD_D)
    if back:
        s.vline(30, 43 + dy, 51 + dy, CARD_D)             # costura de la espalda
        return
    # blusa en V, lazo y botones
    s.poly([(25, 39 + dy), (35, 39 + dy), (30, 47 + dy)], BLOUSE)
    s.line([(25, 39 + dy), (30, 47 + dy)], CARD_D)
    s.line([(35, 39 + dy), (30, 47 + dy)], CARD_D)
    s.vline(30, 48 + dy, 52 + dy, CARD_D)
    for y in (49, 51):
        s.px(29, y + dy, BLOUSE)
    bow(s, 30, 43 + dy)


def bow(s, cx, cy):
    s.poly([(cx - 4, cy - 2), (cx - 1, cy), (cx - 4, cy + 2)], BOW)
    s.poly([(cx + 4, cy - 2), (cx + 1, cy), (cx + 4, cy + 2)], BOW)
    s.px(cx - 4, cy, BOW_D)
    s.px(cx + 4, cy, BOW_D)
    s.rect(cx - 1, cy - 1, cx + 1, cy + 1, BOW_D)
    s.px(cx - 1, cy + 3, BOW)
    s.px(cx + 1, cy + 3, BOW)


def arm_side(s, x, dy, swing=0, hand=True):
    """Brazo colgando a un costado; swing lo adelanta (baja la mano)."""
    y0, y1 = 41 + dy, 51 + dy + swing
    s.rrect(x, y0, x + 4, y1, CARD)
    s.vline(x, y0 + 1, y1 - 1, CARD_L if x < 30 else CARD_D)
    s.hline(x, x + 4, y1 - 1, CARD_L)                    # puño
    if hand:
        s.ell(x, y1, x + 4, y1 + 4, SKIN)
        s.px(x + 3, y1 + 3, SKIN_D)


def hair_back_front(s, dy, long=48):
    # media luna más oscura abajo a la derecha: volumen con luz desde la izquierda
    s.ell(10, 3 + dy, 50, 42 + dy, HAIR_D)
    s.rrect(11, 22 + dy, 49, long + dy, HAIR_D, r=3)
    s.ell(10, 2 + dy, 48, 41 + dy, HAIR)
    s.rrect(11, 22 + dy, 47, long - 1 + dy, HAIR, r=3)


def eye(s, ex, ey, flick):
    """Ojo de 5 × 7: pestaña gruesa con la cola hacia afuera, iris alto en
    degradé, pupila y dos brillos (grande arriba a la izquierda, chico abajo)."""
    s.hline(ex, ex + 4, ey - 1, INK)
    s.px(ex + 5 if flick > 0 else ex - 1, ey - 1, INK)    # cola de la pestaña
    s.px(ex + 5 if flick > 0 else ex - 1, ey, INK)
    s.rect(ex, ey, ex + 4, ey + 5, EYE)
    s.hline(ex, ex + 4, ey, EYE_D)
    s.rect(ex + 1, ey + 1, ex + 3, ey + 3, EYE_D)
    s.rect(ex + 2, ey + 1, ex + 3, ey + 3, INK)            # pupila
    s.hline(ex + 1, ex + 3, ey + 5, EYE_L)
    s.px(ex + 2, ey + 4, EYE_L)
    s.rect(ex, ey + 1, ex + 1, ey + 2, WHITE)              # brillo grande
    s.px(ex + 4, ey + 4, WHITE)                            # brillo chico
    s.hline(ex + 1, ex + 3, ey + 6, SKIN_D)                # párpado inferior suave


def face_front(s, dy, expression='calm'):
    # cara redonda arriba y mentón en punta suave
    s.ell(18, 16 + dy, 42, 37 + dy, SKIN)
    s.poly([(18, 30 + dy), (42, 30 + dy), (38, 38 + dy), (33, 41 + dy), (27, 41 + dy), (22, 38 + dy)], SKIN)
    s.line([(22, 38 + dy), (27, 41 + dy)], SKIN_D)
    s.line([(38, 38 + dy), (33, 41 + dy)], SKIN_D)
    s.hline(28, 32, 41 + dy, SKIN_D)
    eye(s, 22, 28 + dy, -1)
    eye(s, 34, 28 + dy, 1)
    # cejas finas y relajadas bajo el flequillo
    s.hline(23, 25, 25 + dy, HAIR_D)
    s.hline(35, 37, 25 + dy, HAIR_D)
    for bx in (20, 22, 38, 40):
        s.px(bx, 35 + dy, BLUSH)
    s.px(21, 36 + dy, BLUSH)
    s.px(39, 36 + dy, BLUSH)
    if expression == 'eat':
        s.ell(28, 36 + dy, 32, 39 + dy, MOUTH)
    else:
        s.hline(29, 31, 37 + dy, MOUTH)
        s.px(30, 38 + dy, BOW_D)


def bangs_front(s, dy):
    pts = [(15, 14), (45, 14), (45, 27), (43, 22), (40, 24), (37, 19), (34, 23), (31, 18),
           (28, 23), (25, 19), (22, 24), (19, 21), (15, 27)]
    s.poly([(x, y + dy) for x, y in pts], HAIR)
    # mechones a los lados de la cara
    s.poly([(15, 22 + dy), (19, 24 + dy), (19, 40 + dy), (16, 45 + dy), (14, 34 + dy)], HAIR)
    s.poly([(45, 22 + dy), (41, 24 + dy), (41, 40 + dy), (44, 45 + dy), (46, 34 + dy)], HAIR)
    # separaciones entre mechones
    for x0, x1 in ((24, 25), (30, 30), (36, 35)):
        s.line([(x0, 16 + dy), (x1, 20 + dy)], HAIR_D)
    s.line([(18, 30 + dy), (17, 40 + dy)], HAIR_D)
    s.line([(42, 30 + dy), (43, 40 + dy)], HAIR_D)


def hair_shine(s, dy, cx=30):
    """Brillo en arco sobre la coronilla (luz desde arriba a la izquierda)."""
    for x in range(cx - 13, cx + 6):
        t = (x - (cx - 13)) / 19
        y = round(10 - 3 * (1 - (2 * t - 1) ** 2)) + dy
        s.px(x, y, HAIR_L)
        if x % 3 != 0:
            s.px(x, y + 1, HAIR_L)
    for x, y in ((cx - 11, 8), (cx - 7, 6), (cx - 3, 6)):
        s.px(x, y + dy, HAIR_HI)


def ribbon(s, cx, cy):
    """Moño rosa del pelo."""
    s.poly([(cx - 5, cy - 3), (cx - 1, cy), (cx - 5, cy + 3)], BOW)
    s.poly([(cx + 5, cy - 3), (cx + 1, cy), (cx + 5, cy + 3)], BOW)
    s.hline(cx - 4, cx - 3, cy - 2, '#FFC2CD')
    s.rect(cx - 1, cy - 1, cx + 1, cy + 1, BOW_D)
    s.line([(cx - 1, cy + 2), (cx - 3, cy + 6)], BOW)
    s.line([(cx + 1, cy + 2), (cx + 3, cy + 6)], BOW)


# ------------------------------------------------------------ vistas

def south(dy=0, legs=(0, 0), arms=(0, 0), expression='calm', seated=False, hands=True):
    s = Sprite(W, H)
    hair_back_front(s, dy)
    if not seated:  # sentada, el asiento lo pone el mueble
        legs_front(s, 0, *legs)
        skirt(s, dy)
    arm_side(s, 14, dy, arms[0], hands and not seated)
    arm_side(s, 42, dy, arms[1], hands and not seated)
    torso_front(s, dy)
    face_front(s, dy, expression)
    bangs_front(s, dy)
    hair_shine(s, dy)
    ribbon(s, 44, 9 + dy)
    if seated:
        knees(s, hands)
    return s


def north(dy=0, legs=(0, 0), arms=(0, 0), seated=False):
    s = Sprite(W, H)
    if not seated:  # sentada, el asiento lo pone el mueble
        legs_front(s, 0, *legs)
        skirt(s, dy)
    arm_side(s, 14, dy, arms[0], not seated)
    arm_side(s, 42, dy, arms[1], not seated)
    torso_front(s, dy, back=True)
    # pelo largo que cubre la cabeza y cae sobre la espalda
    s.ell(10, 3 + dy, 50, 42 + dy, HAIR_D)
    s.poly([(12, 24 + dy), (48, 24 + dy), (46, 51 + dy), (40, 48 + dy), (34, 52 + dy),
            (28, 48 + dy), (22, 52 + dy), (14, 49 + dy)], HAIR_D)
    s.ell(10, 2 + dy, 48, 41 + dy, HAIR)
    s.poly([(12, 24 + dy), (46, 24 + dy), (44, 49 + dy), (40, 46 + dy), (34, 50 + dy),
            (28, 46 + dy), (22, 50 + dy), (14, 47 + dy)], HAIR)
    for x in (20, 26, 33, 39):
        s.line([(x, 26 + dy), (x + (x - 30) // 6, 47 + dy)], HAIR_D)
    hair_shine(s, dy)
    ribbon(s, 30, 22 + dy)
    if seated:
        # de espaldas en seiza asoman las plantas de los pies
        for x0 in (20, 33):
            s.rrect(x0, 64, x0 + 7, 70, SOCK, INK, r=2)
            s.hline(x0 + 1, x0 + 6, 69, SOCK_D)
    return s


def east(dy=0, stride=(0, 0), arm=0, expression='calm', seated=False, hands=True):
    s = Sprite(W, H)
    # pelo largo detrás, cae hacia la espalda (oeste)
    s.ell(9, 3 + dy, 47, 42 + dy, HAIR_D)
    s.poly([(10, 22 + dy), (30, 22 + dy), (28, 51 + dy), (20, 48 + dy), (12, 51 + dy)], HAIR_D)
    s.ell(9, 2 + dy, 45, 41 + dy, HAIR)
    s.poly([(10, 22 + dy), (28, 22 + dy), (26, 49 + dy), (20, 46 + dy), (12, 49 + dy)], HAIR)
    s.line([(16, 26 + dy), (15, 47 + dy)], HAIR_D)
    s.line([(22, 26 + dy), (22, 47 + dy)], HAIR_D)
    if not seated:  # sentada, el asiento lo pone el mueble
        legs_side(s, 0, *stride)
        skirt(s, dy, narrow=2)
    # torso de perfil
    s.rrect(20, 39 + dy, 40, 52 + dy, CARD, r=2)
    s.rect(21, 41 + dy, 22, 50 + dy, CARD_L)
    s.hline(21, 39, 52 + dy, CARD_D)
    s.poly([(34, 39 + dy), (40, 39 + dy), (38, 46 + dy)], BLOUSE)
    s.line([(34, 39 + dy), (38, 46 + dy)], CARD_D)
    bow(s, 38, 43 + dy)
    # cara de tres cuartos hacia el este
    s.ell(22, 16 + dy, 46, 37 + dy, SKIN)
    s.poly([(22, 30 + dy), (46, 30 + dy), (44, 37 + dy), (39, 41 + dy), (31, 41 + dy), (25, 37 + dy)], SKIN)
    s.line([(44, 37 + dy), (39, 41 + dy)], SKIN_D)
    s.hline(32, 38, 41 + dy, SKIN_D)
    # los dos ojos del mismo tamaño, corridos hacia donde mira
    eye(s, 28, 28 + dy, -1)
    eye(s, 38, 28 + dy, 1)
    s.hline(29, 31, 25 + dy, HAIR_D)
    s.hline(39, 41, 25 + dy, HAIR_D)
    for bx in (28, 30, 40, 42):
        s.px(bx, 35 + dy, BLUSH)
    s.px(29, 36 + dy, BLUSH)
    s.px(41, 36 + dy, BLUSH)
    if expression == 'eat':
        s.ell(34, 36 + dy, 37, 39 + dy, MOUTH)
    else:
        s.hline(35, 37, 37 + dy, MOUTH)
        s.px(36, 38 + dy, BOW_D)
    # flequillo hacia el este y mechón sobre la oreja
    pts = [(14, 14), (47, 14), (47, 26), (45, 21), (42, 24), (39, 19), (36, 23), (33, 18),
           (30, 23), (27, 19), (25, 27), (22, 34), (18, 30)]
    s.poly([(x, y + dy) for x, y in pts], HAIR)
    s.poly([(21, 24 + dy), (26, 26 + dy), (26, 42 + dy), (22, 46 + dy)], HAIR)
    for x0, x1 in ((33, 33), (39, 38)):
        s.line([(x0, 16 + dy), (x1, 19 + dy)], HAIR_D)
    s.line([(23, 30 + dy), (23, 43 + dy)], HAIR_D)
    hair_shine(s, dy, cx=28)
    ribbon(s, 14, 12 + dy)
    if seated:
        knees(s, hands, side=True)
    elif arm is not None:
        # brazo cercano delante del cuerpo
        arm_side(s, 27 + arm, dy, 0, hands)
    return s


def knees(s, hands, side=False):
    """Rodillas en seiza delante del torso (la falda tableada encima de las
    piernas plegadas), con las manos apoyadas."""
    if side:
        s.rrect(15, 64, 23, 70, SOCK, INK, r=2)          # pies bajo la cadera
        s.hline(16, 22, 69, SOCK_D)
        s.poly([(20, 61), (46, 61), (51, 64), (51, 67), (46, 70), (20, 70)], INK)
        s.poly([(21, 62), (46, 62), (50, 65), (50, 66), (46, 69), (21, 69)], SKIRT)
        s.hline(23, 46, 63, SKIRT_L)
        s.hline(23, 45, 68, SKIRT_D)
    else:
        for x0 in (16, 31):                              # dos rodillas redondas
            s.ell(x0, 61, x0 + 14, 70, INK)
            s.ell(x0 + 1, 62, x0 + 13, 69, SKIRT)
            s.hline(x0 + 4, x0 + 9, 63, SKIRT_L)
            s.hline(x0 + 3, x0 + 11, 68, SKIRT_D)
    if hands:
        for hx in ((24, 33) if not side else (37,)):
            s.ell(hx, 60, hx + 5, 64, SKIN)
            s.px(hx + 4, 63, SKIN_D)


def finish(s, seated=False):
    s.outline(INK, keep=(SHADOW,))
    if not seated:
        shadow(s)
    return s


# --------------------------------------------------------------- poses

def idle(facing):
    if facing == 'south':
        return finish(south())
    if facing == 'north':
        return finish(north())
    return finish(east())


# Caminata: (bob, pie izquierdo, pie derecho, brazo izquierdo, brazo derecho)
WALK_FRONT = [(0, 3, 0, 2, -1), (-1, 0, 0, 0, 0), (0, 0, 3, -1, 2), (-1, 0, 0, 0, 0)]
WALK_SIDE = [(0, 6, -6, 3), (-1, 1, -1, 0), (0, -6, 6, -3), (-1, -1, 1, 0)]


def walk(facing, i):
    if facing in ('south', 'north'):
        bob, l, r, al, ar = WALK_FRONT[i]
        if facing == 'south':
            return finish(south(bob, (l, r), (al, ar)))
        return finish(north(bob, (r, l), (ar, al)))
    bob, front, back, arm = WALK_SIDE[i]
    return finish(east(bob, (front, back), arm))


def sit(facing):
    d = SEAT_DROP
    if facing == 'south':
        return finish(south(d, seated=True), seated=True)
    if facing == 'north':
        return finish(north(d, seated=True), seated=True)
    return finish(east(d, seated=True), seated=True)


def reach(facing):
    if facing == 'south':
        s = south(arms=(0, 0))
        # antebrazo hacia la cámara: la mano baja delante del cuerpo
        s.rrect(33, 45, 40, 52, CARD)
        s.hline(33, 40, 51, CARD_L)
        s.ell(33, 51, 40, 58, SKIN)
        s.px(38, 56, SKIN_D)
        return finish(s)
    if facing == 'north':
        s = north()
        # de espaldas: el brazo sube por delante, por encima del hombro
        s.rrect(42, 18, 46, 42, CARD)
        s.vline(46, 19, 41, CARD_D)
        s.ell(41, 12, 47, 19, SKIN)
        return finish(s)
    s = east(arm=None)
    s.rrect(30, 40, 54, 44, CARD)
    s.hline(31, 53, 40, CARD_L)
    s.vline(53, 40, 44, CARD_L)
    s.ell(54, 39, 59, 45, SKIN)
    s.px(58, 44, SKIN_D)
    return finish(s)


def onigiri(s, x, y):
    """Onigiri de 7 × 7 con la punta en (x, y)."""
    s.poly([(x, y), (x + 4, y + 5), (x - 4, y + 5)], RICE)
    s.rect(x - 3, y + 4, x + 3, y + 6, NORI)
    s.px(x - 1, y + 2, BLOUSE_D)


def eat(facing, frame):
    up = frame == 1
    d = SEAT_DROP
    if facing == 'north':
        s = north(d, seated=True)
        y = 30 if up else 36
        s.rrect(41, y + d - 6, 47, y + d + 2, CARD)       # codo que sube y baja
        return finish(s, seated=True)
    if facing == 'south':
        s = south(d, expression='eat' if up else 'calm', seated=True, hands=True)
        y = (38 if up else 44) + d
        s.rrect(33, y + 4, 40, y + 12, CARD)
        s.ell(33, y, 39, y + 6, SKIN)
        onigiri(s, 34, y - 4)
        return finish(s, seated=True)
    s = east(d, expression='eat' if up else 'calm', seated=True, arm=None)
    y = (38 if up else 44) + d
    s.rrect(34, y + 4, 41, y + 12, CARD)
    s.ell(38, y, 44, y + 6, SKIN)
    onigiri(s, 44, y - 4)
    return finish(s, seated=True)


def sleep(frame):
    """Acostada en el futón, con la cabeza en la almohada y el edredón hasta
    los hombros. Se dibuja con el pivote en el punto de la cama: el cuadro
    calza sobre el futón (pivote [32, 70]) corrido (-2, -2)."""
    s = Sprite(W, H)
    rise = 1 if frame == 2 else 0                         # respira
    # pelo desparramado sobre la almohada
    s.ell(13, 5, 49, 31, HAIR_D)
    s.ell(14, 4, 47, 30, HAIR)
    s.poly([(14, 18), (8, 30), (16, 31)], HAIR)
    s.poly([(47, 18), (53, 30), (45, 31)], HAIR_D)
    hair_shine(s, -1, cx=31)
    # cara dormida
    s.ell(20, 12, 42, 33, SKIN)
    s.hline(24, 38, 33, SKIN_D)
    for ex, d in ((23, 1), (34, -1)):                     # ojos cerrados
        s.hline(ex + 1, ex + 4, 24, INK)
        s.px(ex, 23, INK)
        s.px(ex + 5, 23, INK)
        s.px(ex + (5 if d < 0 else 0), 25, HAIR_D)        # pestañas
    for bx in (22, 24, 38, 40):
        s.px(bx, 27, BLUSH)
    s.px(23, 28, BLUSH)
    s.px(39, 28, BLUSH)
    s.hline(30, 32, 29, MOUTH)
    # flequillo
    pts = [(18, 12), (44, 12), (44, 21), (41, 17), (38, 20), (35, 15), (32, 19), (29, 15),
           (26, 19), (23, 16), (19, 21)]
    s.poly(pts, HAIR)
    for x0, x1 in ((26, 27), (32, 32), (38, 37)):
        s.line([(x0, 13), (x1, 16)], HAIR_D)
    ribbon(s, 46, 11)
    # edredón hasta los hombros (mismos colores que el futón)
    top = 32 - rise
    s.rrect(4, top, 57, 75, '#9A8BE3', r=2)
    s.rect(6, top + 1, 55, top + 4, '#C3B9F0')            # borde doblado
    s.hline(6, 55, top + 5, '#B3A7EC')
    s.hline(4, 57, top, INK)
    # pliegues que marcan el cuerpo debajo
    s.line([(18, top + 8), (16, top + 40)], '#8A7BD6')
    s.line([(44, top + 8), (46, top + 40)], '#8A7BD6')
    s.line([(19, top + 8), (17, top + 40)], '#B3A7EC')
    # flores del edredón, en el mismo lugar que en el futón
    for y in range(42, 84, 12):
        for x in range(13, 56, 12):
            fx, fy = x + (6 if (y // 12) % 2 else 0) - 2, y - 2
            if fx < 53 and fy > top + 6:
                s.px(fx, fy, '#FF9CAD')
                for dx, dy in ((-1, 0), (1, 0), (0, -1), (0, 1), (-2, 0), (2, 0), (0, -2), (0, 2)):
                    s.px(fx + dx, fy + dy, '#FFC2CD')
    s.rrect(40, top - 1, 46, top + 4, SKIN, INK, r=2)     # mano sobre el borde
    s.vline(57, top + 6, 75, '#7F70C9')
    s.outline(INK)
    return s


def wash(facing, frame):
    """De pie frente al lavamanos o la bacha, manos moviéndose (4 cuadros)."""
    dx = (0, 2, 0, -2)[frame]
    dy = (0, -1, 1, -1)[frame]
    if facing == 'north':
        s = north(0)
        # codos que se mueven a los lados
        s.rrect(13 + dx // 2, 42 + dy, 18, 49 + dy, CARD)
        s.rrect(44, 42 - dy, 49 + dx // 2, 49 - dy, CARD)
        return finish(s)
    if facing == 'south':
        s = south(0, expression='calm', hands=True)
        # manos juntas al frente frotándose
        s.rrect(21, 43, 27, 50 + dy, CARD)
        s.rrect(35, 43, 41, 50 - dy, CARD)
        s.ell(25 + dx, 48 + dy, 31 + dx, 53 + dy, SKIN)
        s.ell(31 - dx, 48 - dy, 37 - dx, 53 - dy, SKIN)
        if frame in (1, 3):
            s.px(30, 47, '#89D1F2')
            s.px(32, 53, '#E8FAFF')
        return finish(s)
    # east (west se espeja)
    s = east(arm=None)
    s.rrect(30, 41, 46 + dx, 46 + dy, CARD)
    s.hline(31, 45 + dx, 41, CARD_L)
    s.ell(46 + dx, 40 + dy, 52 + dx, 46 + dy, SKIN)
    s.ell(44 - dx // 2, 42 - dy, 50 - dx // 2, 47 - dy, SKIN_D)
    if frame in (1, 3):
        s.px(52 + dx, 45 + dy, '#89D1F2')
        s.px(50, 48, '#E8FAFF')
    return finish(s)


def frames():
    """Todos los cuadros: nombre de archivo → Sprite."""
    out = {}
    for facing in ('south', 'north', 'east'):
        out[f'idle_{facing}'] = idle(facing)
        for i in range(4):
            out[f'walk_{facing}_{i + 1}'] = walk(facing, i)
        out[f'sit_{facing}'] = sit(facing)
        out[f'reach_{facing}'] = reach(facing)
        for i in (1, 2):
            out[f'eat_{facing}_{i}'] = eat(facing, i)
        for i in range(4):
            out[f'wash_{facing}_{i + 1}'] = wash(facing, i)
    for i in (1, 2):
        out[f'sleep_{i}'] = sleep(i)
    for name in [n for n in out if n.endswith('east') or '_east_' in n]:
        out[name.replace('east', 'west')] = out[name].mirrored()
    return out


def main(src):
    global SRC
    SRC = src
    for name, s in frames().items():
        s.save(os.path.join(SRC, 'giantess', f'macro_{name}.png'))
