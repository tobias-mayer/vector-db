package index

import (
	"fmt"
	"testing"

	"github.com/bmizerany/assert"
	imath "github.com/tobias-mayer/vector-db/internal/math"
	itesting "github.com/tobias-mayer/vector-db/internal/testing"
)

// nolint: dupl
func TestCosineDistance_CalcDistance(t *testing.T) {
	for i, c := range []struct {
		v1, v2 []float64
		exp    float64
		dim    int
	}{
		{
			v1:  []float64{1.2, 0.1},
			v2:  []float64{-1.2, 0.2},
			dim: 2,
			exp: 0.97,
		},
		{
			v1:  []float64{1.2, 0.1, 0, 0, 0, 0, 0, 0, 0, 0},
			v2:  []float64{-1.2, 0.2, 0, 0, 0, 0, 0, 0, 0, 0},
			dim: 10,
			exp: 0.97,
		},
	} {
		c := c

		t.Run(fmt.Sprintf("%d-th case", i), func(t *testing.T) {
			dp := make([]*DataPoint, 2)
			dp[0] = NewDataPoint(0, c.v1)
			dp[1] = NewDataPoint(1, c.v2)

			distanceMeasure := NewCosineDistanceMeasure()
			actual := distanceMeasure.CalcDistance(c.v1, c.v2)
			itesting.AlmostEqual(t, c.exp, actual, 1e-3)
		})
	}
}

// nolint: dupl
func TestEuclidianDistance_CalcDistance(t *testing.T) {
	for i, c := range []struct {
		v1, v2 []float64
		exp    float64
		dim    int
	}{
		{
			v1:  []float64{1.2, 0.1},
			v2:  []float64{-1.2, 0.2},
			dim: 2,
			exp: 2.4,
		},
		{
			v1:  []float64{1.2, 0.1, 0, 0, 0, 0, 0, 0, 0, 0},
			v2:  []float64{-1.2, 0.2, 0, 0, 0, 0, 0, 0, 0, 0},
			dim: 10,
			exp: 2.4,
		},
	} {
		c := c

		t.Run(fmt.Sprintf("%d-th case", i), func(t *testing.T) {
			dp := make([]*DataPoint, 2)
			dp[0] = NewDataPoint(0, c.v1)
			dp[1] = NewDataPoint(1, c.v2)

			distanceMeasure := NewEuclidianDistanceMeasure()
			actual := distanceMeasure.CalcDistance(c.v1, c.v2)
			itesting.AlmostEqual(t, c.exp, actual, 1e-2)
		})
	}
}

// nolint: dupl
func TestCosineDistance_CalcDirectionPriority(t *testing.T) {
	for i, c := range []struct {
		v1, v2 []float64
		exp    float64
		dim    int
	}{
		{
			v1:  []float64{1.2, 0.1},
			v2:  []float64{-1.2, 0.2},
			dim: 2,
			exp: -1.42,
		},
		{
			v1:  []float64{1.2, 0.1, 0, 0, 0, 0, 0, 0, 0, 0},
			v2:  []float64{-1.2, 0.2, 0, 0, 0, 0, 0, 0, 0, 0},
			dim: 10,
			exp: -1.42,
		},
	} {
		c := c

		t.Run(fmt.Sprintf("%d-th case", i), func(t *testing.T) {
			actual := imath.VectorDotProduct(c.v1, c.v2)
			assert.Equal(t, c.exp, actual)
		})
	}
}
