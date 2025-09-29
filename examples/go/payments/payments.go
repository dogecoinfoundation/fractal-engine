package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/dogeorg/doge/koinu"
)

func main() {
	// Creates a new Dogecoin keypair
	privHex, _, address, err := doge.GenerateDogecoinKeypair(doge.PrefixRegtest)
	if err != nil {
		log.Fatal(err)
	}

	// Example Invoice values
	invoiceHash := "INVOICE_HASH"
	invoiceQuantity := 25
	invoicePrice := 10
	utxoTxId := "TXID"
	utxoVout := 0

	utxoAmount, err := koinu.ParseKoinu("1000")
	if err != nil {
		log.Fatal(err)
	}

	feeAmount, err := koinu.ParseKoinu("1")
	if err != nil {
		log.Fatal(err)
	}

	envelope := protocol.NewPaymentTransactionEnvelope(invoiceHash, protocol.ACTION_PAYMENT)
	encodedTransactionBody := envelope.Serialize()

	inputs := []interface{}{
		map[string]interface{}{
			"txid": utxoTxId,
			"vout": utxoVout,
		},
	}

	dogeUtxoValue := utxoAmount
	buyOfferValue := koinu.Koinu(invoiceQuantity * invoicePrice)

	change := dogeUtxoValue - buyOfferValue - feeAmount
	sellerAddress := "PAYMENT_ADDRESS_OF_SELLER"

	outputs := map[string]interface{}{
		"data": hex.EncodeToString(encodedTransactionBody),
	}

	if address == sellerAddress {
		outputs[address] = change
	} else {
		outputs[address] = change
		outputs[sellerAddress] = buyOfferValue
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
