package indexer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/dogeorg/doge/koinu"
)

type IndexerClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewIndexerClient(baseURL string) *IndexerClient {
	return &IndexerClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HealthResponse represents the response from /health endpoint
type HealthResponse struct {
	OK bool `json:"ok"`
}

// BalanceResponse represents the response from /balance endpoint

type BalanceResponse struct {
	Incoming  koinu.Koinu `json:"incoming"`  // takes N confirmations to become Availble
	Available koinu.Koinu `json:"available"` // confirmed balance you can spend
	Outgoing  koinu.Koinu `json:"outgoing"`  // takes N confirmations to become fully Spent
	Current   koinu.Koinu `json:"current"`   // current balance: Incoming + Available
}

// UTXOItem represents a single UTXO
type UTXOItem struct {
	TxID   string      `json:"tx"`     // hex-encoded transaction ID (byte-reversed)
	VOut   uint32      `json:"vout"`   // transaction output number
	Value  koinu.Koinu `json:"value"`  // UTXO value to 8 decimal places, as a decimal string
	Type   string      `json:"type"`   // UTXO type (determines what you need to sign it)
	Script string      `json:"script"` // hex-encoded UTXO locking script (needed to sign the UTXO)
}

// UTXOResponse represents the response from /utxo endpoint
type UTXOResponse struct {
	UTXOs []UTXOItem `json:"utxo"`
}

// GetHealth checks the health status of the indexer service
func (c *IndexerClient) GetHealth() (*HealthResponse, error) {
	url := fmt.Sprintf("%s/health", c.BaseURL)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make health request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}

	return &health, nil
}

// GetBalance retrieves the balance for a given Dogecoin address
func (c *IndexerClient) GetBalance(address string) (*BalanceResponse, error) {
	u, err := url.Parse(fmt.Sprintf("%s/balance", c.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	query := u.Query()
	query.Set("address", address)
	u.RawQuery = query.Encode()

	resp, err := c.HTTPClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to make balance request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("balance request failed with status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read balance response: %w", err)
	}

	fmt.Println("DATA", string(data))

	var balance BalanceResponse
	if err := json.Unmarshal(data, &balance); err != nil {
		return nil, fmt.Errorf("failed to decode balance response: %w", err)
	}

	fmt.Println("balancebalancebalance", balance.Available)

	return &balance, nil
}

// GetUTXO retrieves the UTXOs for a given Dogecoin address
func (c *IndexerClient) GetUTXO(address string) (*UTXOResponse, error) {
	u, err := url.Parse(fmt.Sprintf("%s/utxo", c.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	query := u.Query()
	query.Set("address", address)
	u.RawQuery = query.Encode()

	resp, err := c.HTTPClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to make UTXO request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("UTXO request failed with status: %d", resp.StatusCode)
	}

	var utxos UTXOResponse
	if err := json.NewDecoder(resp.Body).Decode(&utxos); err != nil {
		return nil, fmt.Errorf("failed to decode UTXO response: %w", err)
	}

	return &utxos, nil
}
