"""Formas sin antialias para dibujar piezas por código (macro a 80 px/m).

Complementa a px.Canvas: polígonos y elipses con coordenadas inclusivas,
contorno automático de la silueta y espejado.
"""
from PIL import Image, ImageDraw, ImageOps

from px import hexc


def rgba(c):
    return hexc(c) if isinstance(c, str) else c


class Sprite:
    """Lienzo RGBA con primitivas de pixel art (coordenadas inclusivas)."""

    def __init__(self, w, h):
        self.w, self.h = w, h
        self.im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
        self.d = ImageDraw.Draw(self.im)

    def px(self, x, y, c):
        if 0 <= x < self.w and 0 <= y < self.h:
            c = rgba(c)
            if c[3] < 255:
                self.blend(x, y, c)
            else:
                self.im.putpixel((x, y), c)

    def blend(self, x, y, c):
        if not (0 <= x < self.w and 0 <= y < self.h):
            return
        r, g, b, a = self.im.getpixel((x, y))
        c = rgba(c)
        t = c[3] / 255
        if a == 0:
            self.im.putpixel((x, y), c)
            return
        self.im.putpixel((x, y), (round(r + (c[0] - r) * t), round(g + (c[1] - g) * t),
                                  round(b + (c[2] - b) * t), a))

    def get(self, x, y):
        if 0 <= x < self.w and 0 <= y < self.h:
            return self.im.getpixel((x, y))
        return (0, 0, 0, 0)

    def rect(self, x0, y0, x1, y1, c):
        self.d.rectangle((x0, y0, x1, y1), fill=rgba(c))

    def hline(self, x0, x1, y, c):
        self.d.line((x0, y, x1, y), fill=rgba(c))

    def vline(self, x, y0, y1, c):
        self.d.line((x, y0, x, y1), fill=rgba(c))

    def line(self, pts, c):
        self.d.line(pts, fill=rgba(c))

    def ell(self, x0, y0, x1, y1, c):
        self.d.ellipse((x0, y0, x1, y1), fill=rgba(c))

    def ring(self, x0, y0, x1, y1, c):
        self.d.ellipse((x0, y0, x1, y1), outline=rgba(c))

    def poly(self, pts, c):
        self.d.polygon(pts, fill=rgba(c))

    def rrect(self, x0, y0, x1, y1, fill, outline=None, r=1):
        """Rectángulo con esquinas recortadas r píxeles."""
        if fill:
            self.rect(x0 + r, y0, x1 - r, y1, fill)
            self.rect(x0, y0 + r, x1, y1 - r, fill)
            if r > 1:
                self.rect(x0 + 1, y0 + 1, x1 - 1, y1 - 1, fill)
        if outline:
            self.hline(x0 + r, x1 - r, y0, outline)
            self.hline(x0 + r, x1 - r, y1, outline)
            self.vline(x0, y0 + r, y1 - r, outline)
            self.vline(x1, y0 + r, y1 - r, outline)
            if r > 1:
                self.px(x0 + 1, y0 + 1, outline)
                self.px(x1 - 1, y0 + 1, outline)
                self.px(x0 + 1, y1 - 1, outline)
                self.px(x1 - 1, y1 - 1, outline)

    def paste(self, img, x, y):
        self.im.alpha_composite(img, (x, y))

    def outline(self, c, keep=()):
        """Pinta c en cada píxel transparente vecino (4 direcciones) de la
        silueta, salvo los vecinos de colores en keep (sombras)."""
        c = rgba(c)
        keep = {rgba(k) for k in keep}
        src = self.im.copy().load()
        out = self.im.load()
        for y in range(self.h):
            for x in range(self.w):
                if src[x, y][3]:
                    continue
                for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)):
                    nx, ny = x + dx, y + dy
                    if 0 <= nx < self.w and 0 <= ny < self.h:
                        p = src[nx, ny]
                        if p[3] == 255 and p not in keep:
                            out[x, y] = c
                            break

    def under(self, other):
        """Dibuja other debajo de lo que ya hay (sombras de contacto)."""
        base = other.im.copy()
        base.alpha_composite(self.im)
        self.im = base
        self.d = ImageDraw.Draw(self.im)

    def recolor(self, src, dst):
        src, dst = rgba(src), rgba(dst)
        pix = self.im.load()
        for y in range(self.h):
            for x in range(self.w):
                if pix[x, y] == src:
                    pix[x, y] = dst

    def mirrored(self):
        s = Sprite(self.w, self.h)
        s.im = ImageOps.mirror(self.im)
        s.d = ImageDraw.Draw(s.im)
        return s

    def save(self, path):
        import os
        os.makedirs(os.path.dirname(path), exist_ok=True)
        self.im.save(path, optimize=True)
