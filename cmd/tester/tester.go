package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"dogecoin.org/fractal-engine/pkg/protocol"
)

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	JsonRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
	ID     string          `json:"id"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

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

func main() {
	if _, err := os.Stat("addresses"); os.IsNotExist(err) {
		fromAddress, fromPrivateKey, toAddress, err := setupAddresses()
		if err != nil {
			fmt.Println("Error setting up addresses:", err)
			return
		}

		_, err = generateToAddress(fromAddress, 100)
		if err != nil {
			fmt.Println("Error generating to address:", err)
			return
		}

		_, err = generateToAddress(toAddress, 100)
		if err != nil {
			fmt.Println("Error generating to address:", err)
			return
		}

		os.WriteFile("addresses", []byte(fromAddress+"\n"+fromPrivateKey+"\n"+toAddress), 0644)
	}

	addresses, err := os.ReadFile("addresses")
	if err != nil {
		fmt.Println("Error reading addresses:", err)
		return
	}

	addressList := strings.Split(string(addresses), "\n")

	fromAddress := addressList[0]
	fromPrivateKey := addressList[1]
	toAddress := addressList[2]

	unspents, err := listUnspent(fromAddress)
	if err != nil {
		fmt.Println("Error getting balance:", err)
		return
	}

	fmt.Println("Unspents:", unspents)

	if len(*unspents) == 0 {
		fmt.Println("No unspent transactions found")
		return
	}

	firstUnspent := (*unspents)[len(*unspents)-1]

	fmt.Println(firstUnspent)

	mint := protocol.Mint{
		Title:         "Test Mint",
		FractionCount: 100,
		Description:   "This is a test mint",
	}

	mintBytes, err := mint.Serialize()
	if err != nil {
		fmt.Println("Error serializing mint:", err)
		return
	}

	message := protocol.NewMessageEnvelope(protocol.ACTION_MINT, mintBytes)
	bytes := message.Serialize()

	trxnId, err := createAndSendTransaction(fromPrivateKey, firstUnspent, bytes, toAddress)
	if err != nil {
		fmt.Println("Error creating and sending transaction:", err)
		return
	}

	fmt.Println("Transaction ID:", trxnId)

}

func setupAddresses() (string, string, string, error) {
	fromAddress, err := getNewAddress()
	if err != nil {
		fmt.Println("Error getting new address:", err)
		return "", "", "", err
	}

	fromPrivateKey, err := dumpPrivateKey(fromAddress)
	if err != nil {
		fmt.Println("Error dumping private key:", err)
		return "", "", "", err
	}

	toAddress, err := getNewAddress()
	if err != nil {
		fmt.Println("Error getting new address:", err)
		return "", "", "", err
	}

	txID, err := generateToAddress(fromAddress, 1)
	if err != nil {
		fmt.Println("Error generating to address:", err)
		return "", "", "", err
	}
	fmt.Println("Transaction ID Add 1 Doge:", txID)

	txID2, err := generateToAddress(toAddress, 1)
	if err != nil {
		fmt.Println("Error generating to address:", err)
		return "", "", "", err
	}
	fmt.Println("Transaction ID Add 1 Doge:", txID2)

	return fromAddress, fromPrivateKey, toAddress, nil
}

func createAndSendTransaction(privateKey string, selectedUTXO UTXO, opReturnData []byte, outAddress string) (string, error) {
	// Step 2: Create raw transaction with OP_RETURN output
	inputs := []map[string]interface{}{
		{
			"txid": selectedUTXO.TxID,
			"vout": selectedUTXO.Vout,
		},
	}

	outputs := map[string]interface{}{
		"data": hex.EncodeToString([]byte(opReturnData)),
	}

	outputs[outAddress] = selectedUTXO.Amount - 1.0

	// Create raw transaction
	createResp, err := callRPC("createrawtransaction", []interface{}{inputs, outputs})
	if err != nil {
		return "", err
	}
	if createResp.Error != nil {
		log.Fatalf("RPC error: %v", createResp.Error.Message)
	}

	var rawTx string
	if err := json.Unmarshal(createResp.Result, &rawTx); err != nil {
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
	signResp, err := callRPC("signrawtransaction", []interface{}{hex.EncodeToString(rawTxBytes), prevTxs, privkeys})
	if err != nil {
		log.Fatalf("Error signing raw transaction: %v", err)
	}
	if signResp.Error != nil {
		log.Fatalf("RPC error: %v", signResp.Error.Message)
	}

	fmt.Println(signResp.Result)

	var signResult map[string]interface{}
	if err := json.Unmarshal(signResp.Result, &signResult); err != nil {
		log.Fatalf("Error parsing signed transaction: %v", err)
	}

	signedTx, ok := signResult["hex"].(string)
	if !ok {
		log.Fatal("Error retrieving signed transaction hex.")
	}

	// Step 5: Broadcast the signed transaction
	sendResp, err := callRPC("sendrawtransaction", []interface{}{signedTx})
	if err != nil {
		log.Fatalf("Error broadcasting transaction: %v", err)
	}
	if sendResp.Error != nil {
		log.Fatalf("RPC error: %v", sendResp.Error.Message)
	}

	var txID string
	if err := json.Unmarshal(sendResp.Result, &txID); err != nil {
		log.Fatalf("Error parsing transaction ID: %v", err)
	}

	fmt.Printf("Transaction sent successfully! TXID: %s\n", txID)
	return txID, nil
}

func listUnspent(address string) (*[]UTXO, error) {
	resp, err := callRPC("listunspent", []interface{}{1, 9999999, []string{address}})
	if err != nil {
		return nil, err
	}

	var utxos []UTXO
	if err := json.Unmarshal(resp.Result, &utxos); err != nil {
		fmt.Println("Error parsing UTXOs:", err)
		return nil, err
	}

	return &utxos, nil
}

func generateToAddress(address string, amount int) ([]string, error) {
	resp, err := callRPC("generatetoaddress", []interface{}{amount, address})
	if err != nil {
		return []string{}, err
	}

	var txID []string
	if err := json.Unmarshal(resp.Result, &txID); err != nil {
		return []string{}, err
	}

	return txID, nil
}

func dumpPrivateKey(address string) (string, error) {
	resp, err := callRPC("dumpprivkey", []interface{}{address})
	if err != nil {
		return "", err
	}

	var privateKey string
	if err := json.Unmarshal(resp.Result, &privateKey); err != nil {
		return "", err
	}

	return privateKey, nil
}

func getNewAddress() (string, error) {
	resp, err := callRPC("getnewaddress", []interface{}{})
	if err != nil {
		return "", err
	}

	var newAddress string
	if err := json.Unmarshal(resp.Result, &newAddress); err != nil {
		return "", err
	}

	return newAddress, nil
}

// callRPC sends a JSON-RPC request to the Dogecoin node
func callRPC(method string, params []interface{}) (*RPCResponse, error) {
	url := "http://127.0.0.1:18332"
	username := "your_username"
	password := "your_password"

	reqBody := RPCRequest{
		JsonRPC: "1.0",
		ID:      "go-client",
		Method:  method,
		Params:  params,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")

	fmt.Println(req)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var rpcResp RPCResponse
	if err := json.NewDecoder(res.Body).Decode(&rpcResp); err != nil {
		return nil, err
	}

	return &rpcResp, nil
}
