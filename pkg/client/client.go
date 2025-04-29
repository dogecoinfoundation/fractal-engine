package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/api"
)

type FractalEngineClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewFractalEngineClient(BaseURL string, HTTPClient *http.Client) FractalEngineClient {
	return FractalEngineClient{BaseURL: BaseURL, HTTPClient: HTTPClient}
}

func (c *FractalEngineClient) CreateMint(mint api.CreateMintRequest) (api.CreateMintResponse, error) {
	url := fmt.Sprintf("%s/mints", c.BaseURL)
	mintBytes, err := json.Marshal(mint)
	if err != nil {
		return api.CreateMintResponse{}, err
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(mintBytes))
	if err != nil {
		return api.CreateMintResponse{}, err
	}
	defer resp.Body.Close()

	var mintCreateResponse api.CreateMintResponse
	err = json.NewDecoder(resp.Body).Decode(&mintCreateResponse)
	if err != nil {
		return api.CreateMintResponse{}, err
	}

	return mintCreateResponse, nil
}

func (c *FractalEngineClient) CreateTransferRequest(transferRequest api.CreateTransferRequestRequest) (api.CreateTransferRequestResponse, error) {
	url := fmt.Sprintf("%s/transfer-requests", c.BaseURL)
	transferRequestBytes, err := json.Marshal(transferRequest)
	if err != nil {
		return api.CreateTransferRequestResponse{}, err
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(transferRequestBytes))
	if err != nil {
		return api.CreateTransferRequestResponse{}, err
	}
	defer resp.Body.Close()

	var transferRequestCreateResponse api.CreateTransferRequestResponse
	err = json.NewDecoder(resp.Body).Decode(&transferRequestCreateResponse)
	if err != nil {
		return api.CreateTransferRequestResponse{}, err
	}

	return transferRequestCreateResponse, nil
}

func (c *FractalEngineClient) GetMints(page int, limit int) (api.GetMintsResponse, error) {
	url := fmt.Sprintf("%s/mints?page=%d&limit=%d", c.BaseURL, page, limit)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return api.GetMintsResponse{}, err
	}
	defer resp.Body.Close()

	var getMintsResponse api.GetMintsResponse
	err = json.NewDecoder(resp.Body).Decode(&getMintsResponse)
	if err != nil {
		return api.GetMintsResponse{}, err
	}

	return getMintsResponse, nil
}
