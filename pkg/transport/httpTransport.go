package transport

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tobias-mayer/vector-db/pkg/index"
)

type HTTPResource struct {
	port        int
	vectorIndex *index.VectorIndex
}

func NewHTTPServer(port int, vectorIndex *index.VectorIndex) *HTTPResource {
	return &HTTPResource{port, vectorIndex}
}

func (r *HTTPResource) Initialize() error {
	address := fmt.Sprintf(":%v", r.port)
	router := gin.Default()

	router.POST("/forceRebuild", r.forceRebuild)
	router.POST("/addVec", r.addVec)
	router.POST("/search", r.search)

	err := router.Run(address)
	if err != nil {
		fmt.Println("error starting http server: ", err)
	}

	return nil
}

type AddVecRequest struct {
	Vector []float64 `json:"vector" binding:"required"`
}

type SearchVecRequest struct {
	Vector            []float64 `json:"vector" binding:"required"`
	NumberOfNeighbors int       `json:"numberOfNeighbors" binding:"required"`
}

type SearchVecResponse struct {
	Vectors   [][]float64 `json:"vectors" binding:"required"`
	Distances []float64   `json:"distances" binding:"required"`
}

func (r *HTTPResource) forceRebuild(_ *gin.Context) {
}

func (r *HTTPResource) addVec(c *gin.Context) {
	var addReq AddVecRequest
	if err := c.ShouldBindJSON(&addReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if len(addReq.Vector) != r.vectorIndex.NumberOfDimensions {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dimensions"})

		return
	}
}

func (r *HTTPResource) search(c *gin.Context) {
	var searchReq SearchVecRequest
	if err := c.ShouldBindJSON(&searchReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if len(searchReq.Vector) != r.vectorIndex.NumberOfDimensions {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dimensions"})

		return
	}

	res, err := r.vectorIndex.SearchByVector(searchReq.Vector, searchReq.NumberOfNeighbors, index.DefaultBuckets)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	distances := make([]float64, len(*res))
	vectors := make([][]float64, len(*res))

	for i, r := range *res {
		distances[i] = r.Distance
		vectors[i] = r.Vector
	}

	c.JSON(http.StatusCreated, SearchVecResponse{vectors, distances})
}
