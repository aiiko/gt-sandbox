// Package rng provee aleatoriedad determinista por streams (decisión D4).
//
// Cada sistema pide su propio Stream por nombre. La semilla de cada stream se
// deriva de la semilla del mundo y del nombre, así que:
//   - el orden en que se crean los streams no importa;
//   - agregar un sistema nuevo (un stream nuevo) no altera los números que
//     reciben los demás.
//
// Prohibido usar math/rand global dentro de la simulación.
package rng

import (
	"hash/fnv"
	"math/rand/v2"
	"slices"

	"gtsandbox/internal/core/statehash"
)

// Stream es una fuente PCG con nombre.
type Stream struct {
	name string
	pcg  *rand.PCG
	*rand.Rand
}

// NewStream crea el stream `name` derivado de worldSeed.
func NewStream(worldSeed uint64, name string) *Stream {
	h := fnv.New64a()
	h.Write([]byte(name))
	s1 := splitmix64(worldSeed ^ h.Sum64())
	s2 := splitmix64(s1)
	pcg := rand.NewPCG(s1, s2)
	return &Stream{name: name, pcg: pcg, Rand: rand.New(pcg)}
}

func (s *Stream) Name() string { return s.name }

// MarshalBinary serializa el estado interno (para guardado).
func (s *Stream) MarshalBinary() ([]byte, error) { return s.pcg.MarshalBinary() }

// UnmarshalBinary restaura el estado interno.
func (s *Stream) UnmarshalBinary(b []byte) error { return s.pcg.UnmarshalBinary(b) }

func (s *Stream) HashState(h *statehash.Hasher) {
	b, err := s.pcg.MarshalBinary()
	if err != nil {
		panic(err) // PCG.MarshalBinary nunca falla
	}
	h.String(s.name)
	h.Bytes(b)
}

// Set agrupa los streams de un mundo.
type Set struct {
	seed    uint64
	streams map[string]*Stream
}

func NewSet(worldSeed uint64) *Set {
	return &Set{seed: worldSeed, streams: map[string]*Stream{}}
}

func (s *Set) Seed() uint64 { return s.seed }

// Get devuelve el stream `name`, creándolo si no existe.
func (s *Set) Get(name string) *Stream {
	if st, ok := s.streams[name]; ok {
		return st
	}
	st := NewStream(s.seed, name)
	s.streams[name] = st
	return st
}

// Names devuelve los nombres de los streams existentes, ordenados.
func (s *Set) Names() []string {
	names := make([]string, 0, len(s.streams))
	for n := range s.streams {
		names = append(names, n)
	}
	slices.Sort(names)
	return names
}

// HashState hashea todos los streams en orden de nombre.
func (s *Set) HashState(h *statehash.Hasher) {
	h.U64(s.seed)
	names := s.Names()
	h.Int(len(names))
	for _, n := range names {
		s.streams[n].HashState(h)
	}
}

func splitmix64(x uint64) uint64 {
	x += 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}
