package service_test

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
	"dogecoin.org/fractal-engine/pkg/service"
	"dogecoin.org/fractal-engine/pkg/store"
	"gotest.tools/assert"
)

type rpcResponse struct {
	Id     uint64          `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  any             `json:"error"`
}

func TestTrimmerServiceForOnChainTransactions(t *testing.T) {
	tokenisationStore := test_support.SetupTestDB()

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

	tokenisationStore.SaveOnChainTransaction("0000000000000000000000000000000000000000000000000000000000000000", 45, 1, 1, 1, []byte{}, "0000000000000000000000000000000000000000000000000000000000000000", 100)
	tokenisationStore.SaveOnChainTransaction("0000000000000000000000000000000000000000000000000000000000000000", 30, 1, 1, 1, []byte{}, "0000000000000000000000000000000000000000000000000000000000000000", 100)
	tokenisationStore.SaveOnChainTransaction("0000000000000000000000000000000000000000000000000000000000000000", 85, 1, 1, 1, []byte{}, "0000000000000000000000000000000000000000000000000000000000000000", 100)
	tokenisationStore.SaveOnChainTransaction("0000000000000000000000000000000000000000000000000000000000000000", 86, 1, 1, 1, []byte{}, "0000000000000000000000000000000000000000000000000000000000000000", 100)
	tokenisationStore.SaveOnChainTransaction("0000000000000000000000000000000000000000000000000000000000000000", 87, 1, 1, 1, []byte{}, "0000000000000000000000000000000000000000000000000000000000000000", 100)
	tokenisationStore.SaveOnChainTransaction("0000000000000000000000000000000000000000000000000000000000000000", 100, 1, 1, 1, []byte{}, "0000000000000000000000000000000000000000000000000000000000000000", 100)

	count, err := tokenisationStore.GetOnChainTransactions(0, 100)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 6, len(count))

	tokenisationStore.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash: "0000000000000000000000000000000000000000000000000000000000000000",
	})
	tokenisationStore.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash: "0000000000000000000000000000000000000000000000000000000000000000",
	})
	tokenisationStore.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash: "0000000000000000000000000000000000000000000000000000000000000000",
	})
	tokenisationStore.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash: "0000000000000000000000000000000000000000000000000000000000000000",
	})

	mintCount, err := tokenisationStore.GetUnconfirmedMints(0, 100)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 4, len(mintCount))

	trimmerService := service.NewTrimmerService(14, 2, tokenisationStore, rpcClient)
	go trimmerService.Start()

	time.Sleep(2 * time.Second)

	count, err = tokenisationStore.GetOnChainTransactions(0, 100)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(count))

	mintCount, err = tokenisationStore.GetUnconfirmedMints(0, 100)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(mintCount))

	trimmerService.Stop()
}
