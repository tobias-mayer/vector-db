package index

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"testing"
)

// nolint: funlen, gocognit, cyclop, gosec
func TestIndex_SearchByVector(t *testing.T) {
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
			rawItems := make([]*DataPoint[int], c.num)
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

			idx, err := NewVectorIndex(c.nTree, c.dim, c.k, rawItems, NewCosineDistanceMeasure())
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
				aDist[int64(i)] = idx.DistanceMeasure.CalcDistance(v.Embedding, query)
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
			for _, res := range *ass {
				if _, ok := expectedIDsMap[res.ID]; ok {
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

// nolint: funlen, gocognit, cyclop, gosec
func TestIndex_AddAfterBuild(t *testing.T) {
	for i, c := range []struct {
		k, n, dim, num, nTree, searchNum int
		threshold, bucketScale           float64
	}{
		{
			k:           200,
			n:           10,
			dim:         20,
			num:         10000,
			nTree:       20,
			threshold:   0.90,
			searchNum:   200,
			bucketScale: 40,
		},
	} {
		c := c

		t.Run(fmt.Sprintf("%d-th case", i), func(t *testing.T) {
			rawItems := make([]*DataPoint[int], c.num)
			for i := range rawItems {
				v := randVec(c.dim)

				rawItems[i] = NewDataPoint(i, v)
			}

			// we currently need at least two initial data points to build the index
			dataPointsToAdd := make([]*DataPoint[int], c.n)
			for i := range dataPointsToAdd {
				v := randVec(c.dim)
				dataPointsToAdd[i] = NewDataPoint(c.num+i, v)
			}

			idx, err := NewVectorIndex(c.nTree, c.dim, c.k, rawItems, NewCosineDistanceMeasure())
			if err != nil {
				t.Fatal(err)
			}

			idx.Build()

			for i := range dataPointsToAdd {
				err := idx.AddDataPoint(dataPointsToAdd[i])
				if err != nil {
					t.Fatal(err)
				}
			}

			// query vector
			query := make([]float64, c.dim)
			query[0] = 0.1

			// exact neighbors
			aDist := map[int64]float64{}
			ids := make([]int64, len(rawItems))
			for i, v := range rawItems {
				ids[i] = int64(i)
				aDist[int64(i)] = idx.DistanceMeasure.CalcDistance(v.Embedding, query)
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
			for _, res := range *ass {
				if _, ok := expectedIDsMap[res.ID]; ok {
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

// nolint: funlen, gocognit, cyclop, gosec
func TestIndex_SearchByVectoxStringIndexType(t *testing.T) {
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
			rawItems := make([]*DataPoint[string], c.num)
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

				rawItems[i] = NewDataPoint(strconv.Itoa(i), v)
			}

			idx, err := NewVectorIndex[string](c.nTree, c.dim, c.k, rawItems, NewCosineDistanceMeasure())
			if err != nil {
				t.Fatal(err)
			}
			idx.Build()

			// query vector
			query := make([]float64, c.dim)
			query[0] = 0.1

			// exact neighbors
			aDist := map[string]float64{}
			ids := make([]string, len(rawItems))
			for i, v := range rawItems {
				ids[i] = strconv.Itoa(i)
				aDist[strconv.Itoa(i)] = idx.DistanceMeasure.CalcDistance(v.Embedding, query)
			}
			sort.Slice(ids, func(i, j int) bool {
				return aDist[ids[i]] < aDist[ids[j]]
			})

			expectedIDsMap := make(map[string]struct{}, c.searchNum)
			for _, id := range ids[:c.searchNum] {
				expectedIDsMap[id] = struct{}{}
			}

			ass, err := idx.SearchByVector(query, c.searchNum, c.bucketScale)
			if err != nil {
				t.Fatal(err)
			}

			var count int
			for _, res := range *ass {
				if _, ok := expectedIDsMap[res.ID]; ok {
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

// nolint: gosec
func TestIndex_GetSplittingVector(t *testing.T) {
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

			dp := make([]*DataPoint[int], c.num)
			for i := 0; i < c.num; i++ {
				dp[i] = NewDataPoint(i, vs[i])
			}
			idx, _ := NewVectorIndex(1, c.dim, 1, dp, NewCosineDistanceMeasure())
			idx.GetNormalVector(dp)
		})
	}
}

// nolint: gosec
func randVec(dim int) []float64 {
	v := make([]float64, dim)

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

	return v
}
