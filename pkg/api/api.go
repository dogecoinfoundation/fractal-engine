package api

import (
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/gin-gonic/gin"
)

type APIServer struct {
	store  *store.Store
	router *gin.Engine
	routes *Routes
}

func NewAPIServer(store *store.Store) *APIServer {
	router := gin.Default()

	apiServer := &APIServer{
		store:  store,
		router: router,
	}

	apiServer.routes = NewRoutes(apiServer.router, store)

	return apiServer
}

func (s *APIServer) Start() {
	s.router.Run(":8080")
}
