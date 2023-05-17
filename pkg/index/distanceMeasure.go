package index

import "math"

type DistanceMeasure interface {
	CalcDistance(v1, v2 []float64) float64
}

type cosineDistanceMeasure struct{}

func NewCosineDistanceMeasure() DistanceMeasure {
	return &cosineDistanceMeasure{}
}

func (cdm *cosineDistanceMeasure) CalcDistance(v1, v2 []float64) float64 {
	// calculates the cosine distance between two vectors
	if len(v1) != len(v2) || len(v1) == 0 {
		return 0.0
	}

	dotProduct := 0.0
	magA := 0.0
	magB := 0.0

	for i := 0; i < len(v1); i++ {
		dotProduct += v1[i] * v2[i]
		magA += v1[i] * v1[i]
		magB += v2[i] * v2[i]
	}

	magA = math.Sqrt(magA)
	magB = math.Sqrt(magB)

	if magA == 0 || magB == 0 {
		return 0.0
	}

	return -dotProduct / (magA * magB)
}

type euclideanDistanceMeasure struct{}

func NewEuclideanDistanceMeasure() DistanceMeasure {
	return &euclideanDistanceMeasure{}
}

func (cdm *euclideanDistanceMeasure) CalcDistance(v1, v2 []float64) float64 {
	// calculates the euclidean distance between two vectors
	if len(v1) != len(v2) || len(v1) == 0 {
		return 0.0
	}

	sum := 0.0

	for i := 0; i < len(v1); i++ {
		diff := v1[i] - v2[i]
		sum += diff * diff
	}

	return math.Sqrt(sum)
}
