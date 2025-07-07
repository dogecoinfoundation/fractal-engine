package test_rpc

import (
	"testing"

	"dogecoin.org/fractal-engine/pkg/rpc"
	"gotest.tools/assert"
)

func TestGetHealth(t *testing.T) {
	tokenisationStore, _, mux, feClient := SetupRpcTest(t)
	rpc.HandleHealthRoutes(tokenisationStore, mux)

	_, err := feClient.GetHealth()
	assert.Error(t, err, "failed to get health: 404 Not Found")

	tokenisationStore.UpsertHealth(100, 200)

	healthResponse, err := feClient.GetHealth()
	assert.NilError(t, err)
	assert.Equal(t, healthResponse.CurrentBlockHeight, int64(100))
	assert.Equal(t, healthResponse.LatestBlockHeight, int64(200))
	assert.Equal(t, healthResponse.UpdatedAt.IsZero(), false)
}
