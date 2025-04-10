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

func (c *FractalEngineClient) CreateMint(mint api.CreateMintRequest) (string, error) {
	url := fmt.Sprintf("%s/mints", c.BaseURL)
	mintBytes, err := json.Marshal(mint)
	if err != nil {
		return "", err
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(mintBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var mintCreateResponse api.CreateMintResponse
	err = json.NewDecoder(resp.Body).Decode(&mintCreateResponse)
	if err != nil {
		return "", err
	}

	return mintCreateResponse.Id, nil
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
