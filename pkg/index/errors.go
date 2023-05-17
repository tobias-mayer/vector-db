package index

import "errors"

var (
	errShapeMismatch = errors.New("not all data points match the specified dimensionality")
	errInvalidIndex  = errors.New("invalid index")
)
