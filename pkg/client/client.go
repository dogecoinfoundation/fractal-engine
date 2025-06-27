package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/store"
)

type TokenisationClient struct {
	baseUrl    string
	httpClient *http.Client
}

func NewTokenisationClient(baseUrl string) *TokenisationClient {
	httpClient := &http.Client{}
	return &TokenisationClient{baseUrl: baseUrl, httpClient: httpClient}
}

func (c *TokenisationClient) CreateInvoice(invoice *rpc.CreateInvoiceRequest) (rpc.CreateInvoiceResponse, error) {
	jsonValue, err := json.Marshal(invoice)
	if err != nil {
		return rpc.CreateInvoiceResponse{}, err
	}

	resp, err := c.httpClient.Post(c.baseUrl+"/invoices", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return rpc.CreateInvoiceResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return rpc.CreateInvoiceResponse{}, fmt.Errorf("failed to create invoice: %s", string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.CreateInvoiceResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.CreateInvoiceResponse{}, err
	}

	return result, nil
}

func (c *TokenisationClient) Offer(offer *rpc.CreateOfferRequest) (rpc.CreateOfferResponse, error) {
	jsonValue, err := json.Marshal(offer)
	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	resp, err := c.httpClient.Post(c.baseUrl+"/offers", "application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return rpc.CreateOfferResponse{}, fmt.Errorf("failed to create offer: %s", string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.CreateOfferResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	return result, nil
}

func (c *TokenisationClient) GetOffers(page int, limit int, mintHash string, offerType store.OfferType) (rpc.GetOffersResponse, error) {
	resp, err := c.httpClient.Get(c.baseUrl + fmt.Sprintf("/offers?page=%d&limit=%d&mint_hash=%s&type=%d", page, limit, mintHash, offerType))
	if err != nil {
		return rpc.GetOffersResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rpc.GetOffersResponse{}, fmt.Errorf("failed to get offers: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.GetOffersResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.GetOffersResponse{}, err
	}

	return result, nil
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

func (c *TokenisationClient) GetMints(page int, limit int) (rpc.GetMintsResponse, error) {
	resp, err := c.httpClient.Get(c.baseUrl + fmt.Sprintf("/mints?page=%d&limit=%d", page, limit))
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
