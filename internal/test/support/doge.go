package support

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
)

type rpcResponse struct {
	Id     uint64          `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  any             `json:"error"`
}

func NewTestDogeClient(t *testing.T) *doge.RpcClient {
	responseIndex := uint64(0)

	httptestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseIndex = responseIndex + 1

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

			if rpcRequest["method"] == "getblockchaininfo" {
				w.Header().Set("Content-Type", "application/json")
				rpcResponse := rpcResponse{
					Id:     responseIndex,
					Result: json.RawMessage(`{"chain":"test"}`),
				}

				data, err := json.Marshal(rpcResponse)
				if err != nil {
					t.Fatal(err)
				}

				w.Write(data)

				return
			}

			if rpcRequest["method"] == "getwalletinfo" {
				w.Header().Set("Content-Type", "application/json")
				rpcResponse := rpcResponse{
					Id:     responseIndex,
					Result: json.RawMessage("{}"),
				}

				data, err := json.Marshal(rpcResponse)
				if err != nil {
					t.Fatal(err)
				}

				w.Write(data)

				return
			}

			if rpcRequest["method"] == "getbestblockhash" {
				w.Header().Set("Content-Type", "application/json")

				hashStr, err := json.Marshal("0000000000000000000000000000000000000000000000000000000000000000")
				if err != nil {
					t.Fatal(err)
				}

				rpcResponse := rpcResponse{
					Id:     responseIndex,
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
					Confirmations: 7,
					Height:        100,
				})
				if err != nil {
					t.Fatal(err)
				}

				rpcResponse := rpcResponse{
					Id:     responseIndex,
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

	return doge.NewRpcClient(cfg)
}
