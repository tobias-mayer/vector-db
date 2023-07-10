package index

import (
	"container/heap"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	imath "github.com/tobias-mayer/vector-db/internal/math"
)

const (
	minDataPointsRequired = 2
	DefaultBuckets        = 10.0
)

type DataPoint[T comparable] struct {
	ID        T
	Embedding []float64
}

type SearchResult[T comparable] struct {
	ID       T
	Distance float64
	Vector   []float64
}

func NewDataPoint[T comparable](id T, embedding []float64) *DataPoint[T] {
	return &DataPoint[T]{id, embedding}
}

// T is the type of the identifier used to identify data points
type VectorIndex[T comparable] struct {
	NumberOfRoots        int
	NumberOfDimensions   int
	MaxItemsPerLeafNode  int
	Roots                []*treeNode[T]
	IDToTreeNodeMapping  map[string]*treeNode[T]
	IDToDataPointMapping map[T]*DataPoint[T]
	DataPoints           []*DataPoint[T]
	DistanceMeasure      DistanceMeasure
	Mutex                *sync.Mutex
}

func NewVectorIndex[T comparable](numberOfRoots int, numberOfDimensions int, maxIetmsPerLeafNode int, dataPoints []*DataPoint[T], distanceMeasure DistanceMeasure) (*VectorIndex[T], error) {
	for _, dp := range dataPoints {
		if len(dp.Embedding) != numberOfDimensions {
			return nil, errShapeMismatch
		}
	}

	idToDataPointMapping := make(map[T]*DataPoint[T], len(dataPoints))
	for _, dp := range dataPoints {
		idToDataPointMapping[dp.ID] = dp
	}

	rand.Seed(time.Now().UnixNano())

	return &VectorIndex[T]{
		NumberOfRoots:        numberOfRoots,
		NumberOfDimensions:   numberOfDimensions,
		MaxItemsPerLeafNode:  maxIetmsPerLeafNode,
		Roots:                make([]*treeNode[T], numberOfRoots),
		IDToDataPointMapping: idToDataPointMapping,
		IDToTreeNodeMapping:  map[string]*treeNode[T]{},
		DataPoints:           dataPoints,
		DistanceMeasure:      distanceMeasure,
		Mutex:                &sync.Mutex{},
	}, nil
}

func (vi *VectorIndex[T]) Build() {
	for i := 0; i < vi.NumberOfRoots; i++ {
		normalVec := vi.GetNormalVector(vi.DataPoints)
		rootNode := &treeNode[T]{
			nodeID:    uuid.New().String(),
			index:     vi,
			normalVec: normalVec,
			left:      nil,
			right:     nil,
		}
		vi.Roots[i] = rootNode
		vi.IDToTreeNodeMapping[rootNode.nodeID] = rootNode
	}

	var wg sync.WaitGroup

	wg.Add(vi.NumberOfRoots)

	for _, rootNode := range vi.Roots {
		rootNode := rootNode
		go func() {
			defer wg.Done()
			rootNode.build(vi.DataPoints)
		}()
	}

	wg.Wait()
}

func (vi *VectorIndex[T]) AddDataPoint(dataPoint *DataPoint[T]) error {
	if len(dataPoint.Embedding) != vi.NumberOfDimensions {
		return errShapeMismatch
	}

	vi.DataPoints = append(vi.DataPoints, dataPoint)
	vi.IDToDataPointMapping[dataPoint.ID] = dataPoint

	var wg sync.WaitGroup

	wg.Add(vi.NumberOfRoots)

	for _, rootNode := range vi.Roots {
		rootNode := rootNode
		go func() {
			defer wg.Done()
			rootNode.insert(dataPoint)
		}()
	}

	wg.Wait()

	return nil
}

// nolint: funlen, cyclop
func (vi *VectorIndex[T]) SearchByVector(input []float64, searchNum int, numberOfBuckets float64) (*[]SearchResult[T], error) {
	if len(input) != vi.NumberOfDimensions {
		return nil, errShapeMismatch
	}

	totalBucketSize := int(float64(searchNum) * numberOfBuckets)
	annMap := make(map[T]struct{}, totalBucketSize)
	pq := priorityQueue{}

	// insert root nodes into pq
	for i, r := range vi.Roots {
		pq = append(pq, &queueItem{r.nodeID, i, math.Inf(-1)})
	}

	heap.Init(&pq)

	// search all trees until we found enough data points
	for pq.Len() > 0 && len(annMap) < totalBucketSize {
		q, _ := heap.Pop(&pq).(*queueItem)
		n, ok := vi.IDToTreeNodeMapping[q.value]

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
	idToDist := make(map[T]float64, len(annMap))
	ann := make([]T, 0, len(annMap))

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

	searchResults := make([]SearchResult[T], len(ann))
	for i, id := range ann {
		searchResults[i] = SearchResult[T]{ID: id, Distance: math.Abs(idToDist[id]), Vector: vi.IDToDataPointMapping[id].Embedding}
	}

	return &searchResults, nil
}

func (vi *VectorIndex[T]) SearchByItem() ([]int, error) {
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
func (vi *VectorIndex[T]) GetNormalVector(dataPoints []*DataPoint[T]) []float64 {
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
func (vi *VectorIndex[T]) getRandomCentroids(dataPoints []*DataPoint[T]) ([]float64, []float64) {
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
