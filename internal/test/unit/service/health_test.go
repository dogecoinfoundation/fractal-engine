package test_service

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	test_support "dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/health"
	"gotest.tools/assert"
)

func TestHealth(t *testing.T) {
	tokenisationStore := test_support.SetupTestDB(t)

	tokenisationStore.UpsertChainPosition(50, "0000000000000000000000000000000000000000000000000000000000000000", false)

	httptestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}

			var rpcRequest map[string]any
			err = json.Unmarshal(body, &rpcRequest)
			if err != nil {
				t.Fatal(err)
			}

			if rpcRequest["method"] == "getbestblockhash" {
				w.Header().Set("Content-Type", "application/json")

				hashStr, err := json.Marshal("0000000000000000000000000000000000000000000000000000000000000000")
				if err != nil {
					t.Fatal(err)
				}

				rpcResponse := rpcResponse{
					Id:     1,
					Result: json.RawMessage(hashStr),
					Error:  nil,
				}

				data, err := json.Marshal(rpcResponse)
				if err != nil {
					t.Fatal(err)
				}

				w.Write(data)

				return
			}

			if rpcRequest["method"] == "getblockheader" {
				w.Header().Set("Content-Type", "application/json")

				hashStr, err := json.Marshal(doge.BlockHeader{
					Confirmations: 1,
					Height:        100,
				})
				if err != nil {
					t.Fatal(err)
				}

				rpcResponse := rpcResponse{
					Id:     2,
					Result: json.RawMessage(hashStr),
					Error:  nil,
				}

				data, err := json.Marshal(rpcResponse)
				if err != nil {
					t.Fatal(err)
				}

				w.Write(data)

				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}))

	parsedUrl, err := url.Parse(httptestServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.NewConfig()
	cfg.DogeHost = parsedUrl.Hostname()
	cfg.DogeScheme = parsedUrl.Scheme
	cfg.DogePort = parsedUrl.Port()
	cfg.DogeUser = "test"
	cfg.DogePassword = "test"

	rpcClient := doge.NewRpcClient(cfg)

	healthService := health.NewHealthService(rpcClient, tokenisationStore)
	go healthService.Start()

	time.Sleep(1 * time.Second)

	currentBlockHeight, latestBlockHeight, updatedAt, err := tokenisationStore.GetHealth()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(50), currentBlockHeight)
	assert.Equal(t, int64(100), latestBlockHeight)
	assert.Equal(t, updatedAt.IsZero(), false)
}
