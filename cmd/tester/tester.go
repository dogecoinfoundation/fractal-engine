package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	"dogecoin.org/dogetest/pkg/dogetest"
	"dogecoin.org/fractal-engine/pkg/client"
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/service"
	"dogecoin.org/fractal-engine/pkg/store"
)

func main() {
	dogeTester, err := dogetest.NewDogeTest(dogetest.DogeTestConfig{
		Host:             "localhost",
		InstallationPath: "C:\\Program Files\\Dogecoin\\daemon\\dogecoind.exe",
		ConfigPath:       "C:\\Users\\danielw\\code\\doge\\dogetest\\config.json",
	})

	if err != nil {
		log.Fatal(err)
	}

	err = dogeTester.Start()
	if err != nil {
		log.Fatal(err)
	}

	addressBook, err := dogeTester.SetupAddresses([]dogetest.AddressSetup{
		{
			Label:          "test",
			InitialBalance: 1000,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	_, err = dogeTester.ConfirmBlocks()
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.NewConfig()
	cfg.DogeHost = dogeTester.Host
	cfg.DogePort = strconv.Itoa(dogeTester.Port)
	cfg.DogeUser = "test"
	cfg.DogePassword = "test"
	cfg.PersistFollower = false

	feService := service.NewTokenisationService(cfg)
	go feService.Start()

	feService.WaitForRunning()

	feClient := client.NewTokenisationClient("http://localhost:8080")
	mintResponse, err := feClient.Mint(&rpc.CreateMintRequest{
		MintWithoutID: store.MintWithoutID{
			Title:         "Test Mint",
			FractionCount: 100,
			Description:   "Test Description",
			Tags:          []string{"test", "mint"},
			Metadata: map[string]interface{}{
				"test": "test",
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	unspent, err := dogeTester.Rpc.ListUnspent(addressBook.Addresses[0].Address)
	if err != nil {
		log.Fatal(err)
	}

	selectedUTXO := unspent[0]

	inputs := []map[string]interface{}{
		{
			"txid": selectedUTXO.TxID,
			"vout": selectedUTXO.Vout,
		},
	}

	change := selectedUTXO.Amount - 0.1

	log.Printf("mintResponse: %v", mintResponse.EncodedTransactionBody)

	outputs := map[string]interface{}{
		"data":                           mintResponse.EncodedTransactionBody,
		addressBook.Addresses[0].Address: change,
	}

	createResp, err := dogeTester.Rpc.Request("createrawtransaction", []interface{}{inputs, outputs})

	if err != nil {
		log.Fatalf("Error creating raw transaction: %v", err)
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

	prevTxs := []map[string]interface{}{
		{

			"txid":         selectedUTXO.TxID,
			"vout":         selectedUTXO.Vout,
			"scriptPubKey": selectedUTXO.ScriptPubKey,
			"amount":       selectedUTXO.Amount,
		},
	}

	// Prepare privkeys (private keys for signing)
	privkeys := []string{addressBook.Addresses[0].PrivateKey}

	signResp, err := dogeTester.Rpc.Request("signrawtransaction", []interface{}{hex.EncodeToString(rawTxBytes), prevTxs, privkeys})
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
	sendResp, err := dogeTester.Rpc.Request("sendrawtransaction", []interface{}{signedTx})
	if err != nil {
		log.Fatalf("Error broadcasting transaction: %v", err)
	}

	var txID string
	if err := json.Unmarshal(*sendResp, &txID); err != nil {
		log.Fatalf("Error parsing transaction ID: %v", err)
	}

	fmt.Printf("Transaction sent successfully! TXID: %s\n", txID)

	_, err = dogeTester.ConfirmBlocks()
	if err != nil {
		log.Fatal(err)
	}

	// Wait for interrupt signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan

	feService.Stop()
}
