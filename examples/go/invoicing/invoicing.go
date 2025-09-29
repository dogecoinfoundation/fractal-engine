package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"dogecoin.org/fractal-engine/pkg/client"
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"github.com/btcsuite/btcd/chaincfg"
)

func main() {
	// Creates a new Dogecoin keypair
	privHex, pubHex, address, err := doge.GenerateDogecoinKeypair(doge.PrefixRegtest)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Fractal Engine Client
	// NOTE: Change this URL to your instance of the Fractal Engine
	feClient := client.NewTokenisationClient("http://localhost:8080", privHex, pubHex)

	// Example Invoice values
	buyerAddress := "BUYER_ADDRESS"
	mintHash := "MINT_HASH"
	quantity := int32(10)
	price := 20

	// Build Invoice request
	createInvoiceRequest := invoicePayload(address, buyerAddress, mintHash, quantity, price)

	// Sign the payload
	signature, err := doge.SignPayload(createInvoiceRequest.Payload, privHex, pubHex)
	if err != nil {
		log.Fatal(err)
	}

	createInvoiceRequest.Signature = signature
	createInvoiceRequest.PublicKey = pubHex

	// Call Invoice API Endpoint
	invoiceResponse, err := feClient.CreateInvoice(&createInvoiceRequest)
	if err != nil {
		log.Fatal(err)
	}

	// Example UTXO values
	// Usually you would query a wallet or system to retrieve your spendable UTXOs
	invoiceHashFromApiCall := invoiceResponse.Hash
	utxoTxId := "TXID"
	utxoVout := 0
	utxoAmount := 1000
	feeAmount := 1

	envelope := protocol.NewInvoiceTransactionEnvelope(invoiceHashFromApiCall, address, mintHash, quantity, protocol.ACTION_INVOICE)
	encodedTransactionBody := envelope.Serialize()

	inputs := []interface{}{
		map[string]interface{}{
			"txid": utxoTxId,
			"vout": utxoVout,
		},
	}

	outputs := map[string]interface{}{
		"data":  hex.EncodeToString(encodedTransactionBody),
		address: utxoAmount - feeAmount,
	}

	// Initialize Dogecoin Client
	// NOTE: Change this URL to your instance of the Dogecoin Node
	dogeClient := doge.NewRpcClient(&config.Config{
		DogeScheme:   "http",
		DogeHost:     "your_dogecoin_node_host",
		DogePort:     "22555",
		DogeUser:     "test",
		DogePassword: "test",
	})

	// Create Raw Transaction
	rawTx, err := dogeClient.Request("createrawtransaction", []interface{}{inputs, outputs})
	if err != nil {
		log.Fatal(err)
	}

	var rawTxResponse string
	if err := json.Unmarshal(*rawTx, &rawTxResponse); err != nil {
		log.Fatal(err)
	}

	// Sign Raw Transaction
	encodedTx, err := doge.SignRawTransaction(rawTxResponse, privHex, []doge.PrevOutput{
		{
			Address: address,
			Amount:  int64(utxoAmount),
		},
	}, &chaincfg.MainNetParams)
	if err != nil {
		log.Fatal(err)
	}

	// Send Raw Transaction
	res, err := dogeClient.Request("sendrawtransaction", []interface{}{encodedTx})
	if err != nil {
		log.Fatal(err)
	}

	var txid string
	if err := json.Unmarshal(*res, &txid); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction sent: " + txid)
}

func invoicePayload(address string, buyerAddress string, mintHash string, quantity int32, price int) rpc.CreateInvoiceRequest {
	return rpc.CreateInvoiceRequest{
		Payload: rpc.CreateInvoiceRequestPayload{
			PaymentAddress: address,
			BuyerAddress:   buyerAddress,
			MintHash:       mintHash,
			Quantity:       int(quantity),
			Price:          price,
			SellerAddress:  address,
		},
	}
}
