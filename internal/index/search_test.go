package index

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/bmizerany/assert"
)

// func TestIndex_GetANNbyVector(t *testing.T) {
// 	for i, c := range []struct {
// 		dim, num, nTree, k int
// 	}{
// 		{dim: 2, num: 1000, nTree: 10, k: 2},
// 		{dim: 10, num: 100, nTree: 5, k: 10},
// 		{dim: 1000, num: 10000, nTree: 5, k: 10},
// 	} {
// 		c := c
// 		t.Run(fmt.Sprintf("%d-th case", i), func(t *testing.T) {
// 			rawItems := make(PointNArray, c.num)
// 			for i := range rawItems {
// 				v := make([]float64, c.dim)

// 				var norm float64
// 				for j := range v {
// 					cof := rand.Float64() - 0.5
// 					v[j] = cof
// 					norm += cof * cof
// 				}

// 				norm = math.Sqrt(norm)
// 				for j := range v {
// 					v[j] /= norm
// 				}

// 				rawItems[i] = v
// 			}

// 			m, err := NewCosineMetric(c.dim)
// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			idx, err := CreateNewIndex(rawItems, c.dim, c.nTree, c.k, m)
// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			key := make(PointN, c.dim)
// 			for i := range key {
// 				key[i] = rand.Float64() - 0.5
// 			}

// 			if _, err = idx.GetANNbyVector(key, 10, 2); err != nil {
// 				t.Fatal(err)
// 			}
// 		})
// 	}
// }

// This unit test is made to verify if our algorithm can correctly find
// the `exact` neighbors. That is done by checking the ratio of exact
// neighbors in the result returned by `getANNbyVector` is less than
// the given threshold.
func TestAnnSearchAccuracy(t *testing.T) {
	for i, c := range []struct {
		k, dim, num, nTree, searchNum int
		threshold, bucketScale        float64
	}{
		{
			k:           2,
			dim:         20,
			num:         10000,
			nTree:       20,
			threshold:   0.90,
			searchNum:   200,
			bucketScale: 20,
		},
		{
			k:           2,
			dim:         20,
			num:         10000,
			nTree:       20,
			threshold:   0.8,
			searchNum:   20,
			bucketScale: 1000,
		},
	} {
		c := c
		t.Run(fmt.Sprintf("%d-th case", i), func(t *testing.T) {
			rawItems := make([]*DataPoint, c.num)
			for i := range rawItems {
				v := make([]float64, c.dim)

				var norm float64
				for j := range v {
					cof := rand.Float64() - 0.5
					v[j] = cof
					norm += cof * cof
				}

				norm = math.Sqrt(norm)
				for j := range v {
					v[j] /= norm
				}

				rawItems[i] = NewDataPoint(i, v)
			}

			idx, err := NewVectorIndex(c.nTree, c.dim, c.k, rawItems)
			if err != nil {
				t.Fatal(err)
			}
			idx.Build()

			// query vector
			query := make([]float64, c.dim)
			query[0] = 0.1

			// exact neighbors
			aDist := map[int64]float64{}
			ids := make([]int64, len(rawItems))
			for i, v := range rawItems {
				ids[i] = int64(i)
				aDist[int64(i)] = idx.CalcDistance(v.Embedding, query)
			}
			sort.Slice(ids, func(i, j int) bool {
				return aDist[ids[i]] < aDist[ids[j]]
			})

			expectedIDsMap := make(map[int]struct{}, c.searchNum)
			for _, id := range ids[:c.searchNum] {
				expectedIDsMap[int(id)] = struct{}{}
			}

			ass, err := idx.SearchByVector(query, c.searchNum, c.bucketScale)
			if err != nil {
				t.Fatal(err)
			}

			var count int
			for _, id := range ass {
				if _, ok := expectedIDsMap[id]; ok {
					count++
				}
			}

			if ratio := float64(count) / float64(c.searchNum); ratio < c.threshold {
				t.Fatalf("Too few exact neighbors found in approximated result: %d / %d = %f", count, c.searchNum, ratio)
			} else {
				t.Logf("ratio of exact neighbors in approximated result: %d / %d = %f", count, c.searchNum, ratio)
			}
		})
	}
}

var (
	dim    = 3
	nTrees = 2
	k      = 10
	nItem  = 100000
)

// func TestNearestGann(t *testing.T) {

// 	rawItems := make(PointR3Array, 0, nItem)
// 	var lat, lon float64
// 	for i := 0; i < nItem; i++ {
// 		lat = float64(rand.Intn(200000)-100000) * 0.0009
// 		lon = float64(rand.Intn(200000)-100000) * 0.0018
// 		pt := NewPointR3(lat, lon)
// 		rawItems = append(rawItems, pt)
// 	}
// 	m, err := NewCosineMetric(dim)
// 	if err != nil {
// 		// err handling
// 		return
// 	}
// 	// create index
// 	idx, err := CreateNewIndex(rawItems, dim, 1, k, m)
// 	if err != nil {
// 		// error handling
// 		return
// 	}
// 	// search
// 	var searchNum = 5
// 	var bucketScale float64 = 10

// 	for i := 0; i < 3; i++ {
// 		lat = float64(rand.Intn(200000)-100000) * 0.0009
// 		lon = float64(rand.Intn(200000)-100000) * 0.0018
// 		pt := NewPointR3(lat, lon)
// 		res, err := idx.GetANNbyVector(pt, searchNum, bucketScale)
// 		log.Println(err, res)
// 	}
// }

// func BenchmarkNearestGann(b *testing.B) {
// 	b.StopTimer()
// 	rawItems := make(PointNArray, 0, nItem)
// 	var lat, lon float64
// 	for i := 0; i < nItem; i++ {
// 		lat = float64(rand.Intn(200000)-100000) * 0.0009
// 		lon = float64(rand.Intn(200000)-100000) * 0.0018
// 		pt := R3OfLocation(lat, lon)
// 		rawItems = append(rawItems, pt[:])
// 	}
// 	m, err := NewCosineMetric(dim)
// 	if err != nil {
// 		// err handling
// 		return
// 	}
// 	// create index
// 	idx, err := CreateNewIndex(rawItems, dim, 1, k, m)
// 	if err != nil {
// 		// error handling
// 		return
// 	}
// 	// search
// 	var searchNum = 5
// 	var bucketScale float64 = 10

// 	b.StartTimer()
// 	for i := 0; i < b.N; i++ {
// 		lat = float64(rand.Intn(200000)-100000) * 0.0009
// 		lon = float64(rand.Intn(200000)-100000) * 0.0018
// 		pt := NewPointR3(lat, lon)
// 		_, _ = idx.GetANNbyVector(pt, searchNum, bucketScale)
// 	}
// }

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
			idx, _ := NewVectorIndex(1, 1, 1, nil)
			actual := idx.DirectionPriority(c.v1, c.v2)
			assert.Equal(t, c.exp, actual)
		})
	}
}

func TestCosineDistance_GetSplittingVector(t *testing.T) {
	for i, c := range []struct {
		dim, num int
	}{
		{
			dim: 5, num: 100,
		},
	} {
		c := c
		t.Run(fmt.Sprintf("%d-th case", i), func(t *testing.T) {
			vs := make([][]float64, c.num)
			for i := 0; i < c.num; i++ {
				v := make([]float64, c.dim)
				for d := 0; d < c.dim; d++ {
					v[d] = rand.Float64()
				}
				vs[i] = v
			}

			dp := make([]*DataPoint, c.num)
			for i := 0; i < c.num; i++ {
				dp[i] = NewDataPoint(i, vs[i])
			}
			idx, _ := NewVectorIndex(1, c.dim, 1, dp)
			idx.GetNormalVector(vs)
		})
	}
}

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
			exp: 1.42,
		},
		{
			v1:  []float64{1.2, 0.1, 0, 0, 0, 0, 0, 0, 0, 0},
			v2:  []float64{-1.2, 0.2, 0, 0, 0, 0, 0, 0, 0, 0},
			dim: 10,
			exp: 1.42,
		},
	} {
		c := c
		t.Run(fmt.Sprintf("%d-th case", i), func(t *testing.T) {
			dp := make([]*DataPoint, 2)
			dp[0] = NewDataPoint(0, c.v1)
			dp[1] = NewDataPoint(1, c.v2)

			idx, _ := NewVectorIndex(1, c.dim, 1, dp)
			actual := idx.CalcDistance(c.v1, c.v2)
			assert.Equal(t, c.exp, actual)
		})
	}
}
