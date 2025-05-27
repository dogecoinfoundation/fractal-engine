package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/rpc"
)

type TokenisationClient struct {
	baseUrl    string
	httpClient *http.Client
}

func NewTokenisationClient(baseUrl string) *TokenisationClient {
	httpClient := &http.Client{}
	return &TokenisationClient{baseUrl: baseUrl, httpClient: httpClient}
}

func (c *TokenisationClient) Mint(mint *rpc.CreateMintRequest) (rpc.CreateMintResponse, error) {
	jsonValue, err := json.Marshal(mint)
	if err != nil {
		return rpc.CreateMintResponse{}, err
	}

	resp, err := c.httpClient.Post(c.baseUrl+"/mints", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return rpc.CreateMintResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return rpc.CreateMintResponse{}, fmt.Errorf("failed to mint token: %s", string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.CreateMintResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.CreateMintResponse{}, err
	}

	return result, nil
}
