package db

import (
	"github.com/tobias-mayer/vector-db/pkg/index"
	"github.com/tobias-mayer/vector-db/pkg/transport"
)

type VectorDB struct {
	index     *index.VectorIndex
	transport transport.Transport
}

func Run(index *index.VectorIndex, transport transport.Transport) (*VectorDB, error) {
	err := transport.Initialize()
	if err != nil {
		return nil, err
	}

	return &VectorDB{index, transport}, nil
}
