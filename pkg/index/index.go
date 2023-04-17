package index

import (
	"container/heap"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	imath "github.com/tobias-mayer/vector-db/internal/math"
)

const (
	minDataPointsRequired = 2
)

type DataPoint struct {
	ID        int
	Embedding []float64
}

type SearchResult struct {
	ID       int
	Distance float64
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
	DistanceMeasure      DistanceMeasure
}

func NewVectorIndex(numberOfRoots int, numberOfDimensions int, maxIetmsPerLeafNode int, dataPoints []*DataPoint, distanceMeasure DistanceMeasure) (*VectorIndex, error) {
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
		DistanceMeasure:      distanceMeasure,
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
func (vi *VectorIndex) SearchByVector(input []float64, searchNum int, numberOfBuckets float64) (*[]SearchResult, error) {
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

		dp := imath.VectorDotProduct(n.normalVec, input)
		heap.Push(&pq, &queueItem{
			value:    n.left.nodeID,
			priority: imath.Max(q.priority, dp),
		})
		heap.Push(&pq, &queueItem{
			value:    n.right.nodeID,
			priority: imath.Max(q.priority, -dp),
		})
	}

	// calculate actual distances
	idToDist := make(map[int]float64, len(annMap))
	ann := make([]int, 0, len(annMap))

	for id := range annMap {
		ann = append(ann, id)
		idToDist[id] = vi.DistanceMeasure.CalcDistance(vi.IDToDataPointMapping[id].Embedding, input)
	}

	// sort the found items by their actual distance
	sort.Slice(ann, func(i, j int) bool {
		return idToDist[ann[i]] < idToDist[ann[j]]
	})

	// return the top n items
	if len(ann) > searchNum {
		ann = ann[:searchNum]
	}

	searchResults := make([]SearchResult, len(ann))
	for i, id := range ann {
		searchResults[i] = SearchResult{ID: id, Distance: math.Abs(idToDist[id])}
	}

	return &searchResults, nil
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

// GetNormalVector calculates the normal vector of a hyperplane that separates
// the two clusters of data points.
// nolint: funlen, gocognit, cyclop, gosec
func (vi *VectorIndex) GetNormalVector(dataPoints []*DataPoint) []float64 {
	lvs := len(dataPoints)
	// Initialize two centroids randomly from the data points.
	c0, c1 := vi.getRandomCentroids(dataPoints)

	// Repeat the two-means clustering algorithm until the two clusters are
	// sufficiently separated or a maximum number of iterations is reached.
	for i := 0; i < cosineMetricsMaxIteration; i++ {
		// Create a map from cluster ID to a slice of vectors assigned to that
		// cluster during clustering.
		clusterToVecs := map[int][][]float64{}

		// Randomly sample a subset of the data points.
		iter := imath.Min(cosineMetricsMaxTargetSample, len(dataPoints))

		// Assign each of the sampled vectors to the cluster with the nearest centroid.
		for i := 0; i < iter; i++ {
			v := dataPoints[rand.Intn(len(dataPoints))].Embedding
			ip0 := vi.DistanceMeasure.CalcDistance(c0, v)
			ip1 := vi.DistanceMeasure.CalcDistance(c1, v)

			if ip0 > ip1 {
				clusterToVecs[0] = append(clusterToVecs[0], v)
			} else {
				clusterToVecs[1] = append(clusterToVecs[1], v)
			}
		}

		// Calculate the ratio of data points assigned to each cluster. If the
		// ratio is below a threshold, the clustering is considered to be
		// sufficiently separated, and the algorithm terminates.
		lc0 := len(clusterToVecs[0])
		lc1 := len(clusterToVecs[1])

		if (float64(lc0)/float64(iter) <= cosineMetricsTwoMeansThreshold) &&
			(float64(lc1)/float64(iter) <= cosineMetricsTwoMeansThreshold) {
			break
		}

		// If one of the clusters has no data points assigned to it, re-initialize
		// the centroids randomly and continue.
		if lc0 == 0 || lc1 == 0 {
			c0, c1 = vi.getRandomCentroids(dataPoints)

			continue
		}

		// Update the centroids based on the data points assigned to each cluster
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

	// Create a new array to hold the resulting normal vector.
	ret := make([]float64, vi.NumberOfDimensions)

	// Calculate the normal vector by subtracting the coordinates of the second centroid from those of the first centroid.
	// Store the resulting value in the corresponding coordinate of the ret slice.
	for d := 0; d < vi.NumberOfDimensions; d++ {
		v := c0[d] - c1[d]
		ret[d] += v
	}

	return ret
}

// nolint: gosec
func (vi *VectorIndex) getRandomCentroids(dataPoints []*DataPoint) ([]float64, []float64) {
	lvs := len(dataPoints)
	k := rand.Intn(lvs)
	l := rand.Intn(lvs - 1)

	if k == l {
		l++
	}

	c0 := dataPoints[k].Embedding
	c1 := dataPoints[l].Embedding

	return c0, c1
}
