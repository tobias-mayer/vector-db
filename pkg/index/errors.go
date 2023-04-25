package index

import "errors"

var (
	// errInsufficientData = errors.New("cannot build an index with less than two data points").
	errShapeMismatch = errors.New("not all data points match the specified dimensionality")
	errInvalidIndex  = errors.New("invalid index")
)
