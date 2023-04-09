package index

import (
	"container/heap"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

const (
	minDataPointsRequired = 2
)

type DataPoint struct {
	ID        int
	Embedding []float64
}

func NewDataPoint(id int, embedding []float64) *DataPoint {
	return &DataPoint{id, embedding}
}

type VectorIndex struct {
	NumberOfRoots        int
	NumberOfDimensions   int
	MaxItemsPerLeafNode  int
	Roots                []*treeNode
	IDToNodeMapping      map[string]*treeNode
	IDToDataPointMapping map[int]*DataPoint
	DataPoints           []*DataPoint
}

func NewVectorIndex(numberOfRoots int, numberOfDimensions int, maxIetmsPerLeafNode int, dataPoints []*DataPoint) (*VectorIndex, error) {
	if len(dataPoints) < minDataPointsRequired {
		return nil, errInsufficientData
	}

	for _, dp := range dataPoints {
		if len(dp.Embedding) != numberOfDimensions {
			return nil, errShapeMismatch
		}
	}

	idToDataPointMapping := make(map[int]*DataPoint, len(dataPoints))
	for _, dp := range dataPoints {
		idToDataPointMapping[dp.ID] = dp
	}

	rand.Seed(time.Now().UnixNano())

	return &VectorIndex{
		NumberOfRoots:        numberOfRoots,
		NumberOfDimensions:   numberOfDimensions,
		MaxItemsPerLeafNode:  maxIetmsPerLeafNode,
		Roots:                make([]*treeNode, numberOfRoots),
		IDToDataPointMapping: idToDataPointMapping,
		IDToNodeMapping:      map[string]*treeNode{},
		DataPoints:           dataPoints,
	}, nil
}

func (vi *VectorIndex) Build() {
	for i := 0; i < vi.NumberOfRoots; i++ {
		normalVec := vi.GetNormalVector(vi.DataPoints)
		rootNode := &treeNode{
			nodeID:    uuid.New().String(),
			index:     vi,
			normalVec: normalVec,
			left:      nil,
			right:     nil,
		}
		vi.Roots[i] = rootNode
		vi.IDToNodeMapping[rootNode.nodeID] = rootNode
	}

	// this should be parallelized
	for _, rootNode := range vi.Roots {
		rootNode.build(vi.DataPoints)
	}
}

// nolint: funlen
func (vi *VectorIndex) SearchByVector(input []float64, searchNum int, numberOfBuckets float64) ([]int, error) {
	if len(input) != vi.NumberOfDimensions {
		return nil, errShapeMismatch
	}

	totalBucketSize := int(float64(searchNum) * numberOfBuckets)
	annMap := make(map[int]struct{}, totalBucketSize)
	pq := priorityQueue{}

	// insert root nodes into pq
	for i, r := range vi.Roots {
		pq = append(pq, &queueItem{r.nodeID, i, math.Inf(-1)})
	}

	heap.Init(&pq)

	// search all trees until we found enough data points
	for pq.Len() > 0 && len(annMap) < totalBucketSize {
		q, _ := heap.Pop(&pq).(*queueItem)
		n, ok := vi.IDToNodeMapping[q.value]

		if !ok {
			return nil, errInvalidIndex
		}

		if len(n.items) > 0 {
			for _, id := range n.items {
				annMap[id] = struct{}{}
			}

			continue
		}

		dp := vi.DirectionPriority(n.normalVec, input)
		heap.Push(&pq, &queueItem{
			value:    n.left.nodeID,
			priority: max(q.priority, dp),
		})
		heap.Push(&pq, &queueItem{
			value:    n.right.nodeID,
			priority: max(q.priority, -dp),
		})
	}

	// calculate actual distances
	idToDist := make(map[int]float64, len(annMap))
	ann := make([]int, 0, len(annMap))

	for id := range annMap {
		ann = append(ann, id)
		idToDist[id] = vi.CalcDistance(vi.IDToDataPointMapping[id].Embedding, input)
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

func (vi *VectorIndex) SearchByItem() ([]int, error) {
	return nil, nil
}

const (
	cosineMetricsMaxIteration      = 200
	cosineMetricsMaxTargetSample   = 100
	cosineMetricsTwoMeansThreshold = 0.7
	cosineMetricsCentroidCalcRatio = 0.0001
)

func (vi *VectorIndex) CalcDistance(v1, v2 []float64) float64 {
	var ret float64
	for i := range v1 {
		ret += v1[i] * v2[i]
	}

	return -ret
}

// nolint: funlen, gocognit, cyclop, gosec
func (vi *VectorIndex) GetNormalVector(dataPoints []*DataPoint) []float64 {
	lvs := len(dataPoints)
	// init centroids
	k := rand.Intn(lvs)
	l := rand.Intn(lvs - 1)

	if k == l {
		l++
	}

	c0 := dataPoints[k].Embedding
	c1 := dataPoints[l].Embedding

	for i := 0; i < cosineMetricsMaxIteration; i++ {
		clusterToVecs := map[int][][]float64{}

		iter := cosineMetricsMaxTargetSample
		if len(dataPoints) < cosineMetricsMaxTargetSample {
			iter = len(dataPoints)
		}

		for i := 0; i < iter; i++ {
			v := dataPoints[rand.Intn(len(dataPoints))].Embedding
			ip0 := vi.CalcDistance(c0, v)
			ip1 := vi.CalcDistance(c1, v)

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

			c0 = dataPoints[k].Embedding
			c1 = dataPoints[l].Embedding

			continue
		}

		c0 = make([]float64, vi.NumberOfDimensions)
		it0 := int(float64(lvs) * cosineMetricsCentroidCalcRatio)

		for i := 0; i < it0; i++ {
			for d := 0; d < vi.NumberOfDimensions; d++ {
				c0[d] += clusterToVecs[0][rand.Intn(lc0)][d] / float64(it0)
			}
		}

		c1 = make([]float64, vi.NumberOfDimensions)
		it1 := int(float64(lvs)*cosineMetricsCentroidCalcRatio + 1)

		for i := 0; i < int(float64(lc1)*cosineMetricsCentroidCalcRatio+1); i++ {
			for d := 0; d < vi.NumberOfDimensions; d++ {
				c1[d] += clusterToVecs[1][rand.Intn(lc1)][d] / float64(it1)
			}
		}
	}

	ret := make([]float64, vi.NumberOfDimensions)

	for d := 0; d < vi.NumberOfDimensions; d++ {
		v := c0[d] - c1[d]
		ret[d] += v
	}

	return ret
}

func (vi *VectorIndex) DirectionPriority(base, target []float64) float64 {
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
