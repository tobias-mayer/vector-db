package main

import (
	"fmt"
	"os"

	"github.com/tobias-mayer/vector-db/pkg/index"
)

func main() {
	data := []*index.DataPoint[int]{
		index.NewDataPoint(0, []float64{0.16, 0.9}),
		index.NewDataPoint(1, []float64{0.5, 0.5}),
		index.NewDataPoint(2, []float64{0.014, 0.99}),
		index.NewDataPoint(3, []float64{0.55, 0.48}),
		index.NewDataPoint(4, []float64{0.01, 0.88}),
		index.NewDataPoint(5, []float64{0.59, 0.6}),
		index.NewDataPoint(6, []float64{0.79, 0.57}),
		index.NewDataPoint(7, []float64{0.86, 0.1}),
		index.NewDataPoint(8, []float64{0.009, 0.95}),
		index.NewDataPoint(9, []float64{0.94, 0.01}),
		index.NewDataPoint(10, []float64{0.0, 0.91}),
		index.NewDataPoint(11, []float64{0.84, 0.08}),
		index.NewDataPoint(12, []float64{0.91, 0.12}),
		index.NewDataPoint(13, []float64{0.9, 0.1}),
		index.NewDataPoint(14, []float64{0.81, 0.19}),
		index.NewDataPoint(15, []float64{0.99, 0.2}),
		index.NewDataPoint(16, []float64{0.912, 0.21}),
		index.NewDataPoint(17, []float64{0.92, 0.17}),
		index.NewDataPoint(18, []float64{0.23, 0.81}),
		index.NewDataPoint(19, []float64{0.91, 0.12}),
	}

	index, err := index.NewVectorIndex(1, 2, 2, data, index.NewCosineDistanceMeasure())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	index.Build()

	searchResults, err := index.SearchByVector([]float64{0.1, 0.9}, 5, 10.0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	fmt.Println("The following vectors are the closest neighbors based on cosine similarity:")
	for _, searchResult := range *searchResults {
		fmt.Println(fmt.Sprintf("id: %v, vector: %v, distance: %f", searchResult.ID, data[searchResult.ID].Embedding, searchResult.Distance))
	}
}
