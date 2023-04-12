package testing

import (
	"fmt"
	"math"
	"testing"
)

func AlmostEqual(t *testing.T, a, b, equalityThreshold float64) {
	t.Helper()
	if math.Abs(a-b) > equalityThreshold {
		t.Fatal(fmt.Sprintf("%f-%f=%f is greater than %f", a, b, math.Abs(a-b), equalityThreshold))
	}
}
