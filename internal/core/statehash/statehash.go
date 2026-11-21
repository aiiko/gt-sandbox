// Package statehash produce hashes del estado lógico de la simulación.
//
// El hash se usa para verificar determinismo (mismas semilla e inputs => mismo
// hash), save/load y replays. No es criptográfico. Cada valor se escribe con
// tamaño fijo y little-endian para que el resultado no dependa de la
// plataforma.
package statehash

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"math"
)

// Hashable lo implementa cualquier parte del estado que contribuya al hash.
// La implementación debe recorrer su estado en orden estable (D5).
type Hashable interface {
	HashState(h *Hasher)
}

// Hasher acumula valores tipados en un hash FNV-1a de 64 bits.
type Hasher struct {
	h   hash.Hash64
	buf [8]byte
}

func New() *Hasher { return &Hasher{h: fnv.New64a()} }

func (h *Hasher) U64(v uint64) {
	binary.LittleEndian.PutUint64(h.buf[:], v)
	h.h.Write(h.buf[:])
}

func (h *Hasher) U32(v uint32) {
	binary.LittleEndian.PutUint32(h.buf[:4], v)
	h.h.Write(h.buf[:4])
}

func (h *Hasher) I64(v int64) { h.U64(uint64(v)) }
func (h *Hasher) Int(v int)   { h.U64(uint64(int64(v))) }

func (h *Hasher) Bool(v bool) {
	if v {
		h.h.Write([]byte{1})
	} else {
		h.h.Write([]byte{0})
	}
}

// F64 escribe los bits exactos del float.
func (h *Hasher) F64(v float64) { h.U64(math.Float64bits(v)) }

// Bytes escribe un bloque prefijado por su longitud (evita ambigüedades).
func (h *Hasher) Bytes(b []byte) {
	h.Int(len(b))
	h.h.Write(b)
}

// String escribe un string prefijado por su longitud.
func (h *Hasher) String(s string) { h.Bytes([]byte(s)) }

// Sum devuelve el hash acumulado.
func (h *Hasher) Sum() uint64 { return h.h.Sum64() }
