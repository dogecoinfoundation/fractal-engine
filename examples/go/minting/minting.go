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
	"dogecoin.org/fractal-engine/pkg/store"
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

	// Build mint request
	createMintRequest := basicMintPayload(address)
	// createMintRequest = mintWithMetadata(address)
	// createMintRequest = mintWithSignatures(address)

	// Sign the payload
	signature, err := doge.SignPayload(createMintRequest.Payload, privHex, pubHex)
	if err != nil {
		log.Fatal(err)
	}

	createMintRequest.Signature = signature
	createMintRequest.PublicKey = pubHex

	// Call Mint API Endpoint
	mintResponse, err := feClient.Mint(&createMintRequest)
	if err != nil {
		log.Fatal(err)
	}

	// Example UTXO values
	// Usually you would query a wallet or system to retrieve your spendable UTXOs
	mintHashFromApiCall := mintResponse.Hash
	utxoTxId := "TXID"
	utxoVout := 0
	utxoAmount := 1000
	feeAmount := 1

	envelope := protocol.NewMintTransactionEnvelope(mintHashFromApiCall, protocol.ACTION_MINT)
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

func basicMintPayload(address string) rpc.CreateMintRequest {
	return rpc.CreateMintRequest{
		Payload: rpc.CreateMintRequestPayload{
			Title:         "Lambo",
			FractionCount: 1000,
			Description:   "Red Lambo Super Car",
			OwnerAddress:  address,
		},
	}
}

func mintWithMetadata(address string) rpc.CreateMintRequest {
	return rpc.CreateMintRequest{
		Payload: rpc.CreateMintRequestPayload{
			Title:         "Lambo",
			FractionCount: 1000,
			Description:   "Red Lambo Super Car",
			OwnerAddress:  address,
			Metadata: store.StringInterfaceMap{
				"vehicle": "car",
				"vin":     "23123213213213",
				"wheels":  6,
			},
		},
	}
}

func mintWithSignatures(address string) rpc.CreateMintRequest {
	return rpc.CreateMintRequest{
		Payload: rpc.CreateMintRequestPayload{
			Title:                    "Lambo",
			FractionCount:            1000,
			Description:              "Red Lambo Super Car",
			OwnerAddress:             address,
			SignatureRequirementType: store.SignatureRequirementType_MIN_SIGNATURES,
			AssetManagers: []store.AssetManager{
				{
					Name:      "Asset Manager",
					PublicKey: "AM_PUBLIC_KEY",
					URL:       "https://example.com/assetManager",
				},
				{
					Name:      "Asset Manager 2",
					PublicKey: "AM_PUBLIC_KEY_2",
					URL:       "https://example.com/assetManager2",
				},
			},
			MinSignatures: 2,
		},
	}
}
