package transport

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tobias-mayer/vector-db/pkg/index"
)

func StartHTTPServer(port int, vectorIndex *index.VectorIndex) {
	address := fmt.Sprintf(":%v", port)
	router := gin.Default()

	registerHandlers(router, vectorIndex)

	err := router.Run(address)
	if err != nil {
		fmt.Println("error starting http server: ", err)
	}
}

type resource struct {
	vectorIndex *index.VectorIndex
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

func registerHandlers(router *gin.Engine, vectorIndex *index.VectorIndex) {
	res := resource{vectorIndex}

	router.POST("/forceRebuild", res.forceRebuild)
	router.POST("/addVec", res.addVec)
	router.POST("/search", res.search)
}

func (r *resource) forceRebuild(_ *gin.Context) {
}

func (r *resource) addVec(c *gin.Context) {
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

func (r *resource) search(c *gin.Context) {
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
	searchResponse := SearchVecResponse{nil, distances}

	c.JSON(http.StatusCreated, searchResponse)
}
