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
