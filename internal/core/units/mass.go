package units

import "fmt"

// Mass es una masa en microgramos (µg), int64.
//
// Por qué µg: una persona de 60 kg reducida 212,5× pesa 60 kg / 212,5³ ≈ 6 mg.
// Con mg no habría resolución para diferenciar tinies entre sí. En µg, 60 kg
// son 6·10¹⁰, muy lejos del límite de int64.
type Mass int64

const (
	Microgram Mass = 1
	Milligram      = 1000 * Microgram
	Gram           = 1000 * Milligram
	Kilogram       = 1000 * Gram
)

func (m Mass) String() string {
	a := m
	if a < 0 {
		a = -a
	}
	switch {
	case a >= Kilogram:
		return fmt.Sprintf("%.3fkg", float64(m)/float64(Kilogram))
	case a >= Gram:
		return fmt.Sprintf("%.3fg", float64(m)/float64(Gram))
	case a >= Milligram:
		return fmt.Sprintf("%.3fmg", float64(m)/float64(Milligram))
	default:
		return fmt.Sprintf("%dµg", int64(m))
	}
}
