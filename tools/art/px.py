"""Primitivas de pixel art para las fuentes de content/art/src/casa."""
import os

from PIL import Image


def hexc(s, a=255):
    s = s.lstrip('#')
    return (int(s[0:2], 16), int(s[2:4], 16), int(s[4:6], 16), a)


class Canvas:
    def __init__(self, w, h):
        self.im = Image.new('RGBA', (w, h), (0, 0, 0, 0))
        self.w, self.h = w, h

    def px(self, x, y, c):
        if 0 <= x < self.w and 0 <= y < self.h:
            if isinstance(c, str):
                c = hexc(c)
            if len(c) == 4 and c[3] < 255:
                self.blend(x, y, c)
            else:
                self.im.putpixel((x, y), c)

    def blend(self, x, y, c):
        if not (0 <= x < self.w and 0 <= y < self.h):
            return
        r, g, b, a = self.im.getpixel((x, y))
        t = c[3] / 255
        if a == 0:
            self.im.putpixel((x, y), c)
            return
        self.im.putpixel((x, y), (round(r + (c[0] - r) * t), round(g + (c[1] - g) * t), round(b + (c[2] - b) * t), a))

    def rect(self, x, y, w, h, c):
        for j in range(y, y + h):
            for i in range(x, x + w):
                self.px(i, j, c)

    def hline(self, x, y, w, c):
        self.rect(x, y, w, 1, c)

    def vline(self, x, y, h, c):
        self.rect(x, y, 1, h, c)

    def frame(self, x, y, w, h, c):
        self.hline(x, y, w, c)
        self.hline(x, y + h - 1, w, c)
        self.vline(x, y, h, c)
        self.vline(x + w - 1, y, h, c)

    def rrect(self, x, y, w, h, fill, outline=None):
        """Rectángulo con esquinas de un píxel recortadas."""
        if fill:
            self.rect(x + 1, y, w - 2, h, fill)
            self.rect(x, y + 1, w, h - 2, fill)
        if outline:
            self.hline(x + 1, y, w - 2, outline)
            self.hline(x + 1, y + h - 1, w - 2, outline)
            self.vline(x, y + 1, h - 2, outline)
            self.vline(x + w - 1, y + 1, h - 2, outline)

    def ellipse(self, cx, cy, rx, ry, c):
        for j in range(-ry, ry + 1):
            for i in range(-rx, rx + 1):
                if (i * i) / (rx * rx + 0.25) + (j * j) / (ry * ry + 0.25) <= 1:
                    self.px(cx + i, cy + j, c)

    def paste(self, img, x, y):
        self.im.alpha_composite(img, (x, y))

    def save(self, path):
        os.makedirs(os.path.dirname(path), exist_ok=True)
        self.im.save(path, optimize=True)


def load(path):
    return Image.open(path).convert('RGBA')
