package index

import (
	"container/heap"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

type DataPoint struct {
	Id        int
	Embedding []float64
}

func NewDataPoint(id int, embedding []float64) *DataPoint {
	return &DataPoint{id, embedding}
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
		if len(dp.Embedding) != numberOfDimensions {
			return nil, errors.New("Not all data points match the specified dimensionality")
		}
	}

	idToDataPointMapping := make(map[int]*DataPoint, len(dataPoints))
	for _, dp := range dataPoints {
		idToDataPointMapping[dp.Id] = dp
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
	vs := make([][]float64, len(index.IdToDataPointMapping))
	for i, it := range index.DataPoints {
		vs[i] = it.Embedding
	}

	for i := 0; i < index.NumberOfRoots; i++ {
		normalVec := index.GetNormalVector(vs)
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
	fmt.Println()
}

func (index *vectorIndex) SearchByVector(input []float64, searchNum int, bucketScale float64) ([]int, error) {
	if len(input) != index.NumberOfDimensions {
		return nil, errors.New("Input shape does not match index shape")
	}

	bucketSize := int(float64(searchNum) * bucketScale)
	annMap := make(map[int]struct{}, bucketSize)

	pq := priorityQueue{}

	// insert root nodes into pq
	for i, r := range index.Roots {
		n := &queueItem{
			value:    r.nodeId,
			index:    i,
			priority: math.Inf(-1),
		}
		pq = append(pq, n)
	}

	heap.Init(&pq)

	// search all trees until we found enough data points
	for pq.Len() > 0 && len(annMap) < bucketSize {
		q, ok := heap.Pop(&pq).(*queueItem)
		d := q.priority
		n, ok := index.IdToNodeMapping[q.value]
		if !ok {
			return nil, errors.New("invalid index")
		}

		if len(n.items) > 0 {
			for _, id := range n.items {
				annMap[id] = struct{}{}
			}
			continue
		}

		dp := index.DirectionPriority(n.normalVec, input)
		heap.Push(&pq, &queueItem{
			value:    n.left.nodeId,
			priority: max(d, dp),
		})
		heap.Push(&pq, &queueItem{
			value:    n.right.nodeId,
			priority: max(d, -dp),
		})
	}

	// calculate actual distances
	idToDist := make(map[int]float64, len(annMap))
	ann := make([]int, 0, len(annMap))
	for id := range annMap {
		ann = append(ann, id)
		idToDist[id] = index.CalcDistance(index.IdToDataPointMapping[id].Embedding, input)
	}

	// sort the found items by their actual distance
	sort.Slice(ann, func(i, j int) bool {
		return idToDist[ann[i]] < idToDist[ann[j]]
	})

	// return the top n items
	if len(ann) > searchNum {
		ann = ann[:searchNum]
	}
	return ann, nil
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

func (vectorIndex *vectorIndex) GetNormalVector(vs [][]float64) []float64 {
	// lvs := len(dataPoints)
	// // init centroids
	// k := rand.Intn(lvs)
	// l := rand.Intn(lvs - 1)
	// if k == l {
	// 	l++
	// }
	// c0 := dataPoints[k].Embedding
	// c1 := dataPoints[l].Embedding

	// for i := 0; i < cosineMetricsMaxIteration; i++ {
	// 	clusterToVecs := map[int][][]float64{}

	// 	iter := cosineMetricsMaxTargetSample
	// 	if len(dataPoints) < cosineMetricsMaxTargetSample {
	// 		iter = len(dataPoints)
	// 	}
	// 	for i := 0; i < iter; i++ {
	// 		v := dataPoints[rand.Intn(len(dataPoints))].Embedding
	// 		ip0 := vectorIndex.CalcDistance(c0, v)
	// 		ip1 := vectorIndex.CalcDistance(c1, v)
	// 		if ip0 > ip1 {
	// 			clusterToVecs[0] = append(clusterToVecs[0], v)
	// 		} else {
	// 			clusterToVecs[1] = append(clusterToVecs[1], v)
	// 		}
	// 	}

	// 	lc0 := len(clusterToVecs[0])
	// 	lc1 := len(clusterToVecs[1])

	// 	if (float64(lc0)/float64(iter) <= cosineMetricsTwoMeansThreshold) &&
	// 		(float64(lc1)/float64(iter) <= cosineMetricsTwoMeansThreshold) {
	// 		break
	// 	}

	// 	// update centroids
	// 	if lc0 == 0 || lc1 == 0 {
	// 		k := rand.Intn(lvs)
	// 		l := rand.Intn(lvs - 1)
	// 		if k == l {
	// 			l++
	// 		}
	// 		c0 = dataPoints[k].Embedding
	// 		c1 = dataPoints[l].Embedding
	// 		continue
	// 	}

	// 	c0 = make([]float64, vectorIndex.NumberOfDimensions)
	// 	it0 := int(float64(lvs) * cosineMetricsCentroidCalcRatio)
	// 	for i := 0; i < it0; i++ {
	// 		for d := 0; d < vectorIndex.NumberOfDimensions; d++ {
	// 			c0[d] += clusterToVecs[0][rand.Intn(lc0)][d] / float64(it0)
	// 		}
	// 	}

	// 	c1 = make([]float64, vectorIndex.NumberOfDimensions)
	// 	it1 := int(float64(lvs)*cosineMetricsCentroidCalcRatio + 1)
	// 	for i := 0; i < int(float64(lc1)*cosineMetricsCentroidCalcRatio+1); i++ {
	// 		for d := 0; d < vectorIndex.NumberOfDimensions; d++ {
	// 			c1[d] += clusterToVecs[1][rand.Intn(lc1)][d] / float64(it1)
	// 		}
	// 	}
	// }

	// ret := make([]float64, vectorIndex.NumberOfDimensions)
	// for d := 0; d < vectorIndex.NumberOfDimensions; d++ {
	// 	v := c0[d] - c1[d]
	// 	ret[d] += v
	// }
	// return ret

	lvs := len(vs)
	// init centroids
	k := rand.Intn(lvs)
	l := rand.Intn(lvs - 1)
	if k == l {
		l++
	}
	c0 := vs[k]
	c1 := vs[l]

	for i := 0; i < cosineMetricsMaxIteration; i++ {
		clusterToVecs := map[int][][]float64{}

		iter := cosineMetricsMaxTargetSample
		if len(vs) < cosineMetricsMaxTargetSample {
			iter = len(vs)
		}
		for i := 0; i < iter; i++ {
			v := vs[rand.Intn(len(vs))]
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
			c0 = vs[k]
			c1 = vs[l]
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

func (vectorIndex *vectorIndex) DirectionPriority(base, target []float64) float64 {
	var ret float64
	for i := range base {
		ret += base[i] * target[i]
	}
	return ret
}

func max(a, b float64) float64 {
	if a < b {
		return b
	}
	return a
}
