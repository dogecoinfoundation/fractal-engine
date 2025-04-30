package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/gin-gonic/gin"
)

type Routes struct {
	store  *store.Store
	router *gin.Engine
}

func NewRoutes(router *gin.Engine, store *store.Store) *Routes {
	routes := &Routes{
		router: router,
		store:  store,
	}

	routes.registerRoutes()

	return routes
}

func (r *Routes) registerRoutes() {
	r.router.POST("/mints", r.Route_CreateMint)
	r.router.GET("/mints", r.Route_GetMints)
}

func (r *Routes) Route_CreateMint(c *gin.Context) {
	var mintRequest CreateMintRequest
	if err := c.ShouldBindJSON(&mintRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mint, err := protocol.NewMint(protocol.MintWithoutID{
		Title:           mintRequest.Title,
		Description:     mintRequest.Description,
		Tags:            mintRequest.Tags,
		Metadata:        mintRequest.Metadata,
		Hash:            mintRequest.Hash,
		Verified:        mintRequest.Verified,
		TransactionHash: mintRequest.TransactionHash,
		FractionCount:   mintRequest.FractionCount,
		CreatedAt:       time.Now(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := r.store.SaveMint(&mint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mintz, err := r.store.GetMints(0, 10, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(mintz)

	c.JSON(http.StatusOK, gin.H{"id": id, "hash": mint.Hash})
}

func (r *Routes) Route_GetMints(c *gin.Context) {
	// Default values
	limit := 5
	page := 1
	verified := false

	// Parse query params
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "5")); err == nil && l > 0 {
		limit = l
	}
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	if v, err := strconv.ParseBool(c.DefaultQuery("verified", "false")); err == nil {
		verified = v
	}

	// Set our limit to 100
	if limit > 100 {
		limit = 100
	}

	// Calculate pagination
	start := (page - 1) * limit
	end := start + limit

	mints, err := r.store.GetMints(start, end, verified)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	// Clamp the slice range
	if start >= len(mints) {
		c.JSON(http.StatusOK, GetMintsResponse{})
		return
	}
	if end > len(mints) {
		end = len(mints)
	}

	response := GetMintsResponse{
		Mints: mints[start:end],
		Total: len(mints),
		Page:  page,
		Limit: limit,
	}

	c.JSON(http.StatusOK, response)
}
