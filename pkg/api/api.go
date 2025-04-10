package api

import (
	"net/http"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/gin-gonic/gin"
)

type APIServer struct {
	store  *store.Store
	router *gin.Engine
}

func NewAPIServer(store *store.Store) *APIServer {
	router := gin.Default()

	apiServer := &APIServer{
		store:  store,
		router: router,
	}

	apiServer.routes()

	return apiServer
}

func (s *APIServer) Start() {
	s.router.Run(":8080")
}

func (s *APIServer) routes() {
	s.router.POST("/mints", func(c *gin.Context) {
		var mint protocol.Mint
		if err := c.ShouldBindJSON(&mint); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := s.store.SaveMint(&mint)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id})
	})
}
