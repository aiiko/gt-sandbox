"""Regenera las piezas dibujadas por código en content/art/src/casa.

    python tools/art/build.py            # escribe en content/art/src/casa
    python tools/art/build.py -check     # verifica que las fuentes coinciden
    go run ./cmd/artpack                 # después, rearmar el atlas

Requiere Pillow. Las imágenes de base/ son las piezas originales de M14 que
los scripts modifican (paredes, heladera, nido, mesa, giantess, palma, taza).
"""
import os
import sys
import tempfile

from PIL import Image

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)

import micro  # noqa: E402
import rooms  # noqa: E402
import sprites  # noqa: E402
import ui  # noqa: E402

ROOT = os.path.dirname(os.path.dirname(HERE))
SRC = os.path.join(ROOT, 'content', 'art', 'src', 'casa')
BASE = os.path.join(HERE, 'base')


def build(out):
    rooms.main(out, BASE)
    micro.main(out, BASE)
    sprites.main(out, BASE)
    ui.main(sprites.save)


def pixels(path):
    return Image.open(path).convert('RGBA').tobytes()


def check():
    """Compara píxeles, no bytes: el PNG de Pillow puede variar entre versiones."""
    with tempfile.TemporaryDirectory() as tmp:
        build(tmp)
        bad = []
        for dirpath, _, files in os.walk(tmp):
            for f in sorted(files):
                rel = os.path.relpath(os.path.join(dirpath, f), tmp)
                want = os.path.join(SRC, rel)
                if not os.path.exists(want) or pixels(want) != pixels(os.path.join(tmp, rel)):
                    bad.append(rel.replace(os.sep, '/'))
        for rel in bad:
            print('distinto:', rel)
        print(f'{len(bad)} piezas distintas')
        return 1 if bad else 0


if __name__ == '__main__':
    if sys.argv[1:] == ['-check']:
        sys.exit(check())
    build(SRC)
    print('fuentes regeneradas en', SRC)
