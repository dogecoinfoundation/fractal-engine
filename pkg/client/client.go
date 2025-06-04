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

func (c *TokenisationClient) GetMints(page int, limit int, verified bool) (rpc.GetMintsResponse, error) {
	resp, err := c.httpClient.Get(c.baseUrl + fmt.Sprintf("/mints?page=%d&limit=%d&verified=%t", page, limit, verified))
	if err != nil {
		return rpc.GetMintsResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rpc.GetMintsResponse{}, fmt.Errorf("failed to get mints: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.GetMintsResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.GetMintsResponse{}, err
	}

	return result, nil
}
