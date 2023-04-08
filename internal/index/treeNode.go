package index

import "github.com/google/uuid"

type treeNode struct {
	nodeId string
	index  *vectorIndex
	// normal vector defining the hyper plane represented by the node
	// splits the search space into two halves represented by the left and right child in the tree
	normalVec []float64

	// if both, left and right are nil, the node represents a leaf node
	left  *treeNode
	right *treeNode

	// if the node is a leaf node, items contains the ids/indices of our data points
	items []int
}

func newTreeNode(index *vectorIndex, normalVec []float64) *treeNode {
	return &treeNode{
		nodeId:    uuid.New().String(),
		index:     index,
		normalVec: normalVec,
		left:      nil,
		right:     nil,
	}
}

func (treeNode *treeNode) build(dataPoints []*DataPoint) {
	if len(dataPoints) > treeNode.index.MaxItemsPerLeafNode {
		// if the current subspace contains more datapoints than MaxItemsPerLeafNode,
		// we need to split it into two new subspaces
		treeNode.buildSubtree(dataPoints)
		return
	}

	// otherwise we have found a leaf node -> left and right stay nil, items are populated with the dp ids
	treeNode.items = make([]int, len(dataPoints))
	for i, dp := range dataPoints {
		treeNode.items[i] = dp.id
	}
}

func (treeNode *treeNode) buildSubtree(dataPoints []*DataPoint) {
	leftDataPoints := []*DataPoint{}
	rightDataPoints := []*DataPoint{}

	for _, dp := range dataPoints {
		// split datapoints into left and right halves based on the metric
		if treeNode.index.CalcDirectionPriority(treeNode.normalVec, dp.embedding) < 0 {
			leftDataPoints = append(leftDataPoints, dp)
		} else {
			rightDataPoints = append(rightDataPoints, dp)
		}
	}

	if len(leftDataPoints) < treeNode.index.MaxItemsPerLeafNode || len(rightDataPoints) < treeNode.index.MaxItemsPerLeafNode {
		treeNode.items = make([]int, len(dataPoints))
		for i, dp := range dataPoints {
			treeNode.items[i] = dp.id
		}
		return
	}

	leftChild := newTreeNode(treeNode.index, treeNode.index.GetSplittingVector(leftDataPoints))
	leftChild.build(leftDataPoints)
	treeNode.left = leftChild

	rightChild := newTreeNode(treeNode.index, treeNode.index.GetSplittingVector(rightDataPoints))
	rightChild.build(rightDataPoints)
	treeNode.right = rightChild

	// treeNode.index.mux.Lock()
	treeNode.index.IdToNodeMapping[leftChild.nodeId] = leftChild
	treeNode.index.IdToNodeMapping[rightChild.nodeId] = rightChild
	// treeNode.index.mux.Unlock()
}
