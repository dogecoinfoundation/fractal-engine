package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/store"
)

type TokenisationClient struct {
	baseUrl    string
	httpClient *http.Client
	privHex    string
	pubHex     string
}

func NewTokenisationClient(baseUrl string, privHex string, pubHex string) *TokenisationClient {
	httpClient := &http.Client{}
	return &TokenisationClient{baseUrl: baseUrl, httpClient: httpClient, privHex: privHex, pubHex: pubHex}
}

func (c *TokenisationClient) CreateInvoice(invoice *rpc.CreateInvoiceRequest) (rpc.CreateInvoiceResponse, error) {
	payloadBytes, err := json.Marshal(invoice.Payload)
	if err != nil {
		return rpc.CreateInvoiceResponse{}, err
	}

	signature, err := doge.SignPayload(payloadBytes, c.privHex)
	if err != nil {
		return rpc.CreateInvoiceResponse{}, err
	}

	invoice.SignedRequest = rpc.SignedRequest{
		PublicKey: c.pubHex,
		Signature: signature,
	}

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

func (c *TokenisationClient) GetInvoices(page int, limit int, mintHash string, offererAddress string) (rpc.GetInvoicesResponse, error) {
	resp, err := c.httpClient.Get(c.baseUrl + fmt.Sprintf("/invoices?page=%d&limit=%d&mint_hash=%s&offerer_address=%s", page, limit, mintHash, offererAddress))
	if err != nil {
		return rpc.GetInvoicesResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rpc.GetInvoicesResponse{}, fmt.Errorf("failed to get invoices: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.GetInvoicesResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.GetInvoicesResponse{}, err
	}

	return result, nil
}

func (c *TokenisationClient) GetHealth() (rpc.GetHealthResponse, error) {
	resp, err := c.httpClient.Get(c.baseUrl + "/health")
	if err != nil {
		return rpc.GetHealthResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rpc.GetHealthResponse{}, fmt.Errorf("failed to get health: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.GetHealthResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.GetHealthResponse{}, err
	}

	return result, nil
}

func (c *TokenisationClient) CreateBuyOffer(offer *rpc.CreateBuyOfferRequest) (rpc.CreateOfferResponse, error) {
	payloadBytes, err := json.Marshal(offer.Payload)
	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	signature, err := doge.SignPayload(payloadBytes, c.privHex)
	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	offer.SignedRequest = rpc.SignedRequest{
		PublicKey: c.pubHex,
		Signature: signature,
	}

	jsonValue, err := json.Marshal(offer)
	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	resp, err := c.httpClient.Post(c.baseUrl+"/buy-offers", "application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return rpc.CreateOfferResponse{}, fmt.Errorf("failed to create buy offer: %s", string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.CreateOfferResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	return result, nil
}

func (c *TokenisationClient) DeleteBuyOffer(offer *rpc.DeleteBuyOfferRequest) (string, error) {
	payloadBytes, err := json.Marshal(offer.Payload)
	if err != nil {
		return "", err
	}

	signature, err := doge.SignPayload(payloadBytes, c.privHex)
	if err != nil {
		return "", err
	}

	offer.SignedRequest = rpc.SignedRequest{
		PublicKey: c.pubHex,
		Signature: signature,
	}

	jsonValue, err := json.Marshal(offer)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(c.baseUrl+"/buy-offers/delete", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to delete buy offer: %s", string(body))
	}

	return "", nil
}

func (c *TokenisationClient) DeleteSellOffer(offer *rpc.DeleteSellOfferRequest) (string, error) {
	payloadBytes, err := json.Marshal(offer.Payload)
	if err != nil {
		return "", err
	}

	signature, err := doge.SignPayload(payloadBytes, c.privHex)
	if err != nil {
		return "", err
	}

	offer.SignedRequest = rpc.SignedRequest{
		PublicKey: c.pubHex,
		Signature: signature,
	}

	jsonValue, err := json.Marshal(offer)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(c.baseUrl+"/sell-offers/delete", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to delete sell offer: %s", string(body))
	}

	return "", nil
}

func (c *TokenisationClient) CreateSellOffer(offer *rpc.CreateSellOfferRequest) (rpc.CreateOfferResponse, error) {
	payloadBytes, err := json.Marshal(offer.Payload)
	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	signature, err := doge.SignPayload(payloadBytes, c.privHex)
	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	offer.SignedRequest = rpc.SignedRequest{
		PublicKey: c.pubHex,
		Signature: signature,
	}

	jsonValue, err := json.Marshal(offer)
	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	resp, err := c.httpClient.Post(c.baseUrl+"/sell-offers", "application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return rpc.CreateOfferResponse{}, fmt.Errorf("failed to create sell offer: %s", string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.CreateOfferResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.CreateOfferResponse{}, err
	}

	return result, nil
}

func (c *TokenisationClient) GetMintByHash(hash string) (rpc.GetMintResponse, error) {
	resp, err := c.httpClient.Get(c.baseUrl + fmt.Sprintf("/mints/%s", hash))
	if err != nil {
		return rpc.GetMintResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rpc.GetMintResponse{}, fmt.Errorf("failed to get mint: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.GetMintResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.GetMintResponse{}, err
	}

	return result, nil
}

func (c *TokenisationClient) GetSellOffersByMintHash(page int, limit int, mintHash string) (rpc.GetSellOffersResponse, error) {
	resp, err := c.httpClient.Get(c.baseUrl + fmt.Sprintf("/sell-offers?page=%d&limit=%d&mint_hash=%s", page, limit, mintHash))
	if err != nil {
		return rpc.GetSellOffersResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rpc.GetSellOffersResponse{}, fmt.Errorf("failed to get offers: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.GetSellOffersResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.GetSellOffersResponse{}, err
	}

	return result, nil
}

func (c *TokenisationClient) GetBuyOffersBySellerAddress(page int, limit int, mintHash string, sellerAddress string) (rpc.GetBuyOffersResponse, error) {
	resp, err := c.httpClient.Get(c.baseUrl + fmt.Sprintf("/buy-offers?page=%d&limit=%d&mint_hash=%s&seller_address=%s", page, limit, mintHash, sellerAddress))
	if err != nil {
		return rpc.GetBuyOffersResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rpc.GetBuyOffersResponse{}, fmt.Errorf("failed to get offers: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.GetBuyOffersResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.GetBuyOffersResponse{}, err
	}

	return result, nil
}

func (c *TokenisationClient) GetBuyOffers(page int, limit int, mintHash string) (rpc.GetBuyOffersResponse, error) {
	resp, err := c.httpClient.Get(c.baseUrl + fmt.Sprintf("/buy-offers?page=%d&limit=%d&mint_hash=%s", page, limit, mintHash))
	if err != nil {
		return rpc.GetBuyOffersResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rpc.GetBuyOffersResponse{}, fmt.Errorf("failed to get offers: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	var result rpc.GetBuyOffersResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return rpc.GetBuyOffersResponse{}, err
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
		body, _ := io.ReadAll(resp.Body)
		return rpc.CreateMintResponse{}, fmt.Errorf("failed to mint token: %s", string(body))
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

func (c *TokenisationClient) GetTokenBalance(address string, mintHash string) ([]store.TokenBalance, error) {
	resp, err := c.httpClient.Get(c.baseUrl + fmt.Sprintf("/token-balances?address=%s&mint_hash=%s", address, mintHash))
	if err != nil {
		return []store.TokenBalance{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []store.TokenBalance{}, fmt.Errorf("failed to get token balance: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)

	var result []store.TokenBalance
	err = json.Unmarshal(body, &result)
	if err != nil {
		return []store.TokenBalance{}, err
	}

	return result, nil
}

func (c *TokenisationClient) GetMints(page int, limit int, publicKey string, includeUnconfirmed bool) (rpc.GetMintsResponse, error) {
	var resp *http.Response
	var err error
	if includeUnconfirmed {
		resp, err = c.httpClient.Get(c.baseUrl + fmt.Sprintf("/mints?page=%d&limit=%d&public_key=%s&include_unconfirmed=true", page, limit, publicKey))
	} else {
		resp, err = c.httpClient.Get(c.baseUrl + fmt.Sprintf("/mints?page=%d&limit=%d&public_key=%s", page, limit, publicKey))
	}
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
