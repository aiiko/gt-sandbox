"""Sintetiza los sonidos del juego a partir de tools/audio/sounds.json.

    python tools/audio/build.py            # escribe content/audio/*.wav y sounds.json
    python tools/audio/build.py -check     # verifica que lo exportado coincide

Solo usa la biblioteca estándar. Cada sonido es una suma de capas; cada capa
tiene forma de onda, barrido de frecuencia, envolvente, filtro y retardo:

    wave      sine | square | triangle | saw | noise
    f0, f1    frecuencia al empezar y al terminar (Hz, barrido exponencial)
    dur       duración de la capa (s)
    delay     cuándo empieza dentro del sonido (s)
    attack    subida lineal (s)
    decay     constante de caída exponencial (s); 0 = sostenido hasta el final
    gain      volumen de la capa (0..1)
    lp        corte del filtro pasabajos de un polo (Hz); 0 = sin filtro
    duty      ancho de pulso de square (0..1, por defecto 0.5)
    vibrato   [Hz, profundidad en semitonos]
    seed      semilla del ruido (por defecto, derivada del nombre)

Todo es determinista: el ruido sale de un generador congruencial propio y el
resultado se redondea a 16 bits, mono, a la frecuencia de muestreo del
manifiesto. El manifiesto exportado (content/audio/sounds.json) nombra cada
sonido con un identificador estable, su categoría, duración, voces
simultáneas y tiempo mínimo entre repeticiones.
"""
import io
import json
import math
import os
import struct
import sys
import wave
import zlib

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(os.path.dirname(HERE))
SPEC = os.path.join(HERE, 'sounds.json')
OUT = os.path.join(ROOT, 'content', 'audio')

CATEGORIES = ('ambiente', 'efectos', 'interfaz')
FADE = 0.004  # fundido final contra clics (s)


class Noise:
    """Generador congruencial de 32 bits: el mismo ruido en cualquier equipo."""

    def __init__(self, seed):
        self.state = seed & 0xFFFFFFFF or 1

    def next(self):
        self.state = (self.state * 1664525 + 1013904223) & 0xFFFFFFFF
        return self.state / 0x80000000 - 1.0


def oscillator(wave_name, phase, duty, noise):
    p = phase % 1.0
    if wave_name == 'sine':
        return math.sin(2 * math.pi * p)
    if wave_name == 'square':
        return 1.0 if p < duty else -1.0
    if wave_name == 'triangle':
        return 4 * p - 1 if p < 0.5 else 3 - 4 * p
    if wave_name == 'saw':
        return 2 * p - 1
    if wave_name == 'noise':
        return noise.next()
    raise ValueError('forma de onda desconocida: ' + wave_name)


def render_layer(layer, rate, name, index):
    dur = layer['dur']
    n = int(round(dur * rate))
    f0 = layer.get('f0', 440.0)
    f1 = layer.get('f1', f0)
    attack = layer.get('attack', 0.002)
    decay = layer.get('decay', 0.0)
    gain = layer.get('gain', 0.5)
    duty = layer.get('duty', 0.5)
    vib_rate, vib_depth = layer.get('vibrato', [0, 0])
    seed = layer.get('seed', zlib.crc32(('%s/%d' % (name, index)).encode()))
    noise = Noise(seed)
    lp = layer.get('lp', 0)
    alpha = 1 - math.exp(-2 * math.pi * lp / rate) if lp else 1.0
    out = []
    phase = 0.0
    held = 0.0
    filtered = 0.0
    # el ruido cambia a la frecuencia de la capa (ruido «de muestreo»), así
    # f0/f1 le dan color; con f0 >= rate es ruido blanco.
    for i in range(n):
        t = i / rate
        frac = i / max(1, n - 1)
        freq = f0 * (f1 / f0) ** frac if f0 > 0 and f1 > 0 else f0
        if vib_depth:
            freq *= 2 ** (vib_depth * math.sin(2 * math.pi * vib_rate * t) / 12)
        if layer['wave'] == 'noise':
            before = phase
            phase += freq / rate
            if int(phase) != int(before) or i == 0:
                held = noise.next()
            v = held
        else:
            v = oscillator(layer['wave'], phase, duty, noise)
            phase += freq / rate
        env = min(1.0, t / attack) if attack > 0 else 1.0
        if decay > 0:
            env *= math.exp(-max(0.0, t - attack) / decay)
        remaining = dur - t
        if remaining < FADE:
            env *= max(0.0, remaining / FADE)
        filtered += alpha * (v - filtered)
        out.append(filtered * env * gain)
    return out


def render(name, sound, rate):
    layers = sound['layers']
    length = max(layer.get('delay', 0) + layer['dur'] for layer in layers)
    mix = [0.0] * int(round(length * rate))
    for index, layer in enumerate(layers):
        start = int(round(layer.get('delay', 0) * rate))
        for i, v in enumerate(render_layer(layer, rate, name, index)):
            if start + i < len(mix):
                mix[start + i] += v
    volume = sound.get('volume', 1.0)
    pcm = bytearray()
    for v in mix:
        s = max(-1.0, min(1.0, v * volume))
        pcm += struct.pack('<h', int(round(s * 32767)))
    buf = io.BytesIO()
    with wave.open(buf, 'wb') as w:
        w.setnchannels(1)
        w.setsampwidth(2)
        w.setframerate(rate)
        w.writeframes(bytes(pcm))
    return buf.getvalue(), int(round(length * 1000))


def build():
    with open(SPEC, encoding='utf-8') as f:
        spec = json.load(f)
    rate = spec['rate']
    files = {}
    manifest = {'rate': rate, 'sounds': []}
    for name in sorted(spec['sounds']):
        sound = spec['sounds'][name]
        if sound['category'] not in CATEGORIES:
            raise ValueError('%s: categoría %r' % (name, sound['category']))
        data, ms = render(name, sound, rate)
        files[name + '.wav'] = data
        manifest['sounds'].append({
            'id': name,
            'file': name + '.wav',
            'category': sound['category'],
            'ms': ms,
            'voices': sound.get('voices', 1),
            'cooldown_ms': sound.get('cooldown_ms', 0),
        })
    text = json.dumps(manifest, ensure_ascii=False, indent=2) + '\n'
    files['sounds.json'] = text.encode('utf-8')
    return files


def main():
    files = build()
    if '-check' in sys.argv[1:]:
        bad = []
        for name, data in sorted(files.items()):
            path = os.path.join(OUT, name)
            if not os.path.exists(path) or open(path, 'rb').read() != data:
                bad.append(name)
        extra = sorted(set(os.listdir(OUT)) - set(files)) if os.path.isdir(OUT) else []
        if bad or extra:
            for name in bad:
                print('distinto:', name)
            for name in extra:
                print('sobra:', name)
            sys.exit(1)
        print('%d archivos de audio al día' % len(files))
        return
    os.makedirs(OUT, exist_ok=True)
    for name in os.listdir(OUT):
        if name not in files:
            os.remove(os.path.join(OUT, name))
    for name, data in sorted(files.items()):
        with open(os.path.join(OUT, name), 'wb') as f:
            f.write(data)
    print('%d archivos escritos en %s' % (len(files), OUT))


if __name__ == '__main__':
    main()
