package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *TokenisationClient) TopUpBalance(ctx context.Context, address string) error {
	resp, err := c.httpClient.Post(c.baseUrl+"/setup-demo-balance?address="+address, "application/json", nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to top up balance: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	var result string
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	return nil
}
