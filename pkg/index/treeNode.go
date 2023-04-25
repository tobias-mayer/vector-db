package index

import (
	"github.com/google/uuid"
	imath "github.com/tobias-mayer/vector-db/internal/math"
)

type treeNode struct {
	nodeID string
	index  *VectorIndex
	// normal vector defining the hyper plane represented by the node
	// splits the search space into two halves represented by the left and right child in the tree
	normalVec []float64

	// if both, left and right are nil, the node represents a leaf node
	left  *treeNode
	right *treeNode

	// if the node is a leaf node, items contains the ids/indices of our data points
	items []int
}

func newTreeNode(index *VectorIndex, normalVec []float64) *treeNode {
	return &treeNode{
		nodeID:    uuid.New().String(),
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
		treeNode.items[i] = dp.ID
	}
}

func (treeNode *treeNode) buildSubtree(dataPoints []*DataPoint) {
	leftDataPoints := []*DataPoint{}
	rightDataPoints := []*DataPoint{}

	for _, dp := range dataPoints {
		// split datapoints into left and right halves based on the metric
		if imath.VectorDotProduct(treeNode.normalVec, dp.Embedding) < 0 {
			leftDataPoints = append(leftDataPoints, dp)
		} else {
			rightDataPoints = append(rightDataPoints, dp)
		}
	}

	if len(leftDataPoints) < treeNode.index.MaxItemsPerLeafNode || len(rightDataPoints) < treeNode.index.MaxItemsPerLeafNode {
		treeNode.items = make([]int, len(dataPoints))
		for i, dp := range dataPoints {
			treeNode.items[i] = dp.ID
		}

		return
	}

	leftChild := newTreeNode(treeNode.index, treeNode.index.GetNormalVector(leftDataPoints))
	leftChild.build(leftDataPoints)
	treeNode.left = leftChild

	rightChild := newTreeNode(treeNode.index, treeNode.index.GetNormalVector(rightDataPoints))
	rightChild.build(rightDataPoints)
	treeNode.right = rightChild

	treeNode.items = make([]int, 0)
	treeNode.index.IDToNodeMapping[leftChild.nodeID] = leftChild
	treeNode.index.IDToNodeMapping[rightChild.nodeID] = rightChild
}

func (treeNode *treeNode) insert(dpIndex int, dataPoint *DataPoint) {
	// 1. find the corresponding leaf node
	// -> two cases:
	//    - insert into leaf node
	//    - split data into two new leaf nodes

	leaf := treeNode.findLeaf(dataPoint)
	leaf.items = append(leaf.items, dpIndex)

	if len(leaf.items) <= leaf.index.MaxItemsPerLeafNode {
		return
	}

	// leaf.build()
}

func (treeNode *treeNode) findLeaf(dataPoint *DataPoint) *treeNode {
	if len(treeNode.items) > 0 {
		return treeNode
	}

	if imath.VectorDotProduct(treeNode.normalVec, dataPoint.Embedding) < 0 {
		return treeNode.left.findLeaf(dataPoint)
	} else {
		return treeNode.right.findLeaf(dataPoint)
	}
}
