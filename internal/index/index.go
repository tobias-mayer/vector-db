package index

import (
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type DataPoint struct {
	id        int
	embedding []float64
}

type vectorIndex struct {
	NumberOfRoots        int
	NumberOfDimensions   int
	MaxItemsPerLeafNode  int
	Roots                []*treeNode
	IdToNodeMapping      map[string]*treeNode
	IdToDataPointMapping map[int]*DataPoint
	DataPoints           []*DataPoint
}

func NewVectorIndex(numberOfRoots int, numberOfDimensions int, maxIetmsPerLeafNode int, dataPoints []*DataPoint) (*vectorIndex, error) {
	if len(dataPoints) < 2 {
		return nil, errors.New("Cannot build an index with less than two data points")
	}

	for _, dp := range dataPoints {
		if len(dp.embedding) != numberOfDimensions {
			return nil, errors.New("Not all data points match the specified dimensionality")
		}
	}

	idToDataPointMapping := make(map[int]*DataPoint, len(dataPoints))
	for _, dp := range dataPoints {
		idToDataPointMapping[dp.id] = dp
	}

	rand.Seed(time.Now().UnixNano())

	return &vectorIndex{
		NumberOfRoots:        numberOfRoots,
		NumberOfDimensions:   numberOfDimensions,
		MaxItemsPerLeafNode:  maxIetmsPerLeafNode,
		Roots:                make([]*treeNode, numberOfRoots),
		IdToDataPointMapping: idToDataPointMapping,
		IdToNodeMapping:      map[string]*treeNode{},
		DataPoints:           dataPoints,
	}, nil
}

func (index *vectorIndex) Build() {
	vectorSpace := make([][]float64, len(index.DataPoints))
	for i, dp := range index.DataPoints {
		vectorSpace[i] = dp.embedding
	}

	for i := 0; i < index.NumberOfRoots; i++ {
		normalVec := []float64{} // todo: calc normalvector based on two random data points
		rootNode := &treeNode{
			nodeId:    uuid.New().String(),
			index:     index,
			normalVec: normalVec,
			left:      nil,
			right:     nil,
		}
		index.Roots[i] = rootNode
		index.IdToNodeMapping[rootNode.nodeId] = rootNode
	}

	// this should be parallelized
	for _, rootNode := range index.Roots {
		rootNode.build(index.DataPoints)
	}
}

func (index *vectorIndex) SearchByVector() ([]int, error) {
	return nil, nil
}

func (index *vectorIndex) SearchByItem() ([]int, error) {
	return nil, nil
}

const (
	cosineMetricsMaxIteration      = 200
	cosineMetricsMaxTargetSample   = 100
	cosineMetricsTwoMeansThreshold = 0.7
	cosineMetricsCentroidCalcRatio = 0.0001
)

func (vectorIndex *vectorIndex) CalcDistance(v1, v2 []float64) float64 {
	var ret float64
	for i := range v1 {
		ret += v1[i] * v2[i]
	}
	return -ret
}

func (vectorIndex *vectorIndex) GetSplittingVector(dataPoints []*DataPoint) []float64 {
	lvs := len(dataPoints)
	// init centroids
	k := rand.Intn(lvs)
	l := rand.Intn(lvs - 1)
	if k == l {
		l++
	}
	c0 := dataPoints[k].embedding
	c1 := dataPoints[l].embedding

	for i := 0; i < cosineMetricsMaxIteration; i++ {
		clusterToVecs := map[int][][]float64{}

		iter := cosineMetricsMaxTargetSample
		if len(dataPoints) < cosineMetricsMaxTargetSample {
			iter = len(dataPoints)
		}
		for i := 0; i < iter; i++ {
			v := dataPoints[rand.Intn(len(dataPoints))].embedding
			ip0 := vectorIndex.CalcDistance(c0, v)
			ip1 := vectorIndex.CalcDistance(c1, v)
			if ip0 > ip1 {
				clusterToVecs[0] = append(clusterToVecs[0], v)
			} else {
				clusterToVecs[1] = append(clusterToVecs[1], v)
			}
		}

		lc0 := len(clusterToVecs[0])
		lc1 := len(clusterToVecs[1])

		if (float64(lc0)/float64(iter) <= cosineMetricsTwoMeansThreshold) &&
			(float64(lc1)/float64(iter) <= cosineMetricsTwoMeansThreshold) {
			break
		}

		// update centroids
		if lc0 == 0 || lc1 == 0 {
			k := rand.Intn(lvs)
			l := rand.Intn(lvs - 1)
			if k == l {
				l++
			}
			c0 = dataPoints[k].embedding
			c1 = dataPoints[l].embedding
			continue
		}

		c0 = make([]float64, vectorIndex.NumberOfDimensions)
		it0 := int(float64(lvs) * cosineMetricsCentroidCalcRatio)
		for i := 0; i < it0; i++ {
			for d := 0; d < vectorIndex.NumberOfDimensions; d++ {
				c0[d] += clusterToVecs[0][rand.Intn(lc0)][d] / float64(it0)
			}
		}

		c1 = make([]float64, vectorIndex.NumberOfDimensions)
		it1 := int(float64(lvs)*cosineMetricsCentroidCalcRatio + 1)
		for i := 0; i < int(float64(lc1)*cosineMetricsCentroidCalcRatio+1); i++ {
			for d := 0; d < vectorIndex.NumberOfDimensions; d++ {
				c1[d] += clusterToVecs[1][rand.Intn(lc1)][d] / float64(it1)
			}
		}
	}

	ret := make([]float64, vectorIndex.NumberOfDimensions)
	for d := 0; d < vectorIndex.NumberOfDimensions; d++ {
		v := c0[d] - c1[d]
		ret[d] += v
	}
	return ret
}

func (vectorIndex *vectorIndex) CalcDirectionPriority(base, target []float64) float64 {
	var ret float64
	for i := range base {
		ret += base[i] * target[i]
	}
	return ret
}
