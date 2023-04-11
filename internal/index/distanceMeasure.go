package index

type DistanceMeasure interface {
	CalcDistance(v1, v2 []float64) float64
	DirectionPriority(base, target []float64) float64
}

type cosineDistanceMeasure struct{}

func NewCosineDistanceMeasure() DistanceMeasure {
	return &cosineDistanceMeasure{}
}

func (cdm *cosineDistanceMeasure) DirectionPriority(base, target []float64) float64 {
	var ret float64
	for i := range base {
		ret += base[i] * target[i]
	}

	return ret
}

func (cdm *cosineDistanceMeasure) CalcDistance(v1, v2 []float64) float64 {
	var ret float64
	for i := range v1 {
		ret += v1[i] * v2[i]
	}

	return -ret
}
