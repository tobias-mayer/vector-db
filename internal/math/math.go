package math

import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}

	return b
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}

	return b
}

func VectorDotProduct(base, target []float64) float64 {
	var ret float64
	for i := range base {
		ret += base[i] * target[i]
	}

	return ret
}
