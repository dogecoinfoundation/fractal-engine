package doge

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"dogecoin.org/chainfollower/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/protocol"
)

type UTXO struct {
	TxID          string  `json:"txid"`
	Vout          int     `json:"vout"`
	Amount        float64 `json:"amount"`
	ScriptPubKey  string  `json:"scriptPubKey"`
	RedeemScript  string  `json:"redeemScript,omitempty"`
	Spendable     bool    `json:"spendable"`
	Solvable      bool    `json:"solvable"`
	Desc          string  `json:"desc"`
	Safe          bool    `json:"safe"`
	Confirmations int     `json:"confirmations"`
}

type DogeClient struct {
	rpc *rpc.RpcTransport
}

func NewDogeClient(rpc *rpc.RpcTransport) *DogeClient {
	return &DogeClient{
		rpc: rpc,
	}
}

func (c *DogeClient) GetUnspent(address string) ([]UTXO, error) {
	unspent, err := c.rpc.Request("listunspent", []any{0, 99999999, []string{address}})
	if err != nil {
		return nil, err
	}

	var result []UTXO
	err = json.Unmarshal(*unspent, &result)
	if err != nil {
		return nil, fmt.Errorf("json-rpc unmarshal error: %v | %v", err, string(*unspent))
	}

	return result, nil
}

func (c *DogeClient) CreateMint(mint *protocol.Mint, fromPrivateKey string, unspents []UTXO, toAddress string) (string, error) {
	mintBytes, err := mint.Serialize()
	if err != nil {
		fmt.Println("Error serializing mint:", err)
		return "", err
	}

	message := protocol.NewMessageEnvelope(protocol.ACTION_MINT, mintBytes)
	bytes := message.Serialize()

	trxnId, err := c.createAndSendTransaction(fromPrivateKey, unspents, bytes, mint.FractionCount, toAddress)
	if err != nil {
		fmt.Println("Error creating and sending transaction:", err)
		return "", err
	}

	return trxnId, nil
}

func (c *DogeClient) createAndSendTransaction(privateKey string, uxtos []UTXO, opReturnData []byte, fractions int, outAddress string) (string, error) {

	selectedUTXO := uxtos[0]

	inputs := []map[string]interface{}{
		{
			"txid": selectedUTXO.TxID,
			"vout": selectedUTXO.Vout,
		},
	}

	change := selectedUTXO.Amount - 0.1

	outputs := map[string]interface{}{
		"data":     hex.EncodeToString(opReturnData),
		outAddress: change,
	}

	// Create raw transaction
	createResp, err := c.rpc.Request("createrawtransaction", []interface{}{inputs, outputs})
	if err != nil {
		return "", err
	}

	var rawTx string

	if err := json.Unmarshal(*createResp, &rawTx); err != nil {
		log.Fatalf("Error parsing raw transaction: %v", err)
	}

	// Step 3: Add OP_RETURN output to the transaction
	rawTxBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		log.Fatalf("Error decoding raw transaction hex: %v", err)
	}

	// Step 4: Sign the raw transaction using the private key
	// Prepare prevtxs (previous transaction inputs)
	prevTxs := []map[string]interface{}{
		{

			"txid":         selectedUTXO.TxID,
			"vout":         selectedUTXO.Vout,
			"scriptPubKey": selectedUTXO.ScriptPubKey,
			"amount":       selectedUTXO.Amount,
		},
	}

	// Prepare privkeys (private keys for signing)
	privkeys := []string{privateKey}

	// Sign the transaction
	signResp, err := c.rpc.Request("signrawtransaction", []interface{}{hex.EncodeToString(rawTxBytes), prevTxs, privkeys})
	if err != nil {
		log.Fatalf("Error signing raw transaction: %v", err)
	}

	var signResult map[string]interface{}
	if err := json.Unmarshal(*signResp, &signResult); err != nil {
		log.Fatalf("Error parsing signed transaction: %v", err)
	}

	signedTx, ok := signResult["hex"].(string)
	if !ok {
		log.Fatal("Error retrieving signed transaction hex.")
	}

	// Step 5: Broadcast the signed transaction
	sendResp, err := c.rpc.Request("sendrawtransaction", []interface{}{signedTx})
	if err != nil {
		log.Fatalf("Error broadcasting transaction: %v", err)
	}

	var txID string
	if err := json.Unmarshal(*sendResp, &txID); err != nil {
		log.Fatalf("Error parsing transaction ID: %v", err)
	}

	fmt.Printf("Transaction sent successfully! TXID: %s\n", txID)
	return txID, nil
}
