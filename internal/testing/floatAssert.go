package testing

import (
	"math"
	"testing"
)

func AlmostEqual(t *testing.T, a, b, equalityThreshold float64) {
	t.Helper()

	if math.Abs(a-b) > equalityThreshold {
		t.Fatalf("%f-%f=%f is greater than %f", a, b, math.Abs(a-b), equalityThreshold)
	}
}
