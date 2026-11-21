// Package fixedmath implementa ángulos y transformaciones en punto fijo.
//
// Todo es entero y determinista (D1, D5): los ángulos se miden en pasos de
// 1/4096 de vuelta (~0,088°) y el seno sale de una tabla precalculada.
// Las rotaciones de múltiplos de 90° son exactas; las demás redondean al µm.
//
// Rango: las coordenadas deben mantenerse por debajo de ±2·10⁹ µm (2 km) para
// que las multiplicaciones por 2^30 no desborden int64.
package fixedmath

//go:generate go run ./gen
