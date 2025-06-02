package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"dogecoin.org/fractal-engine/pkg/client"
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/service"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/dogecoinfoundation/dogetest/pkg/dogetest"
	"gotest.tools/assert"
)

var blocks []string
var addressBook *dogetest.AddressBook
var dogeTest *dogetest.DogeTest
var feService *service.TokenisationService

func TestMain(m *testing.M) {
	// 🚀 Global setup
	fmt.Println(">>> SETUP: Init resources")

	localDogeTest, err := dogetest.NewDogeTest(dogetest.DogeTestConfig{
		Host:             "localhost",
		InstallationPath: "C:\\Program Files\\Dogecoin\\daemon\\dogecoind.exe",
		ConfigPath:       "C:\\Users\\danielw\\code\\doge\\dogetest\\config.json",
	})
	if err != nil {
		log.Fatal(err)
	}

	dogeTest = localDogeTest

	err = dogeTest.Start()
	if err != nil {
		log.Fatal(err)
	}

	addressBook, err = dogeTest.SetupAddresses([]dogetest.AddressSetup{
		{
			Label:          "test1",
			InitialBalance: 100,
		},
		{
			Label:          "test2",
			InitialBalance: 20,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	blocks, err = dogeTest.ConfirmBlocks()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Blocks confirmed:", blocks)

	cfg := config.NewConfig()
	cfg.DogeHost = dogeTest.Host
	cfg.DogePort = strconv.Itoa(dogeTest.Port)
	cfg.DogeUser = "test"
	cfg.DogePassword = "test"
	cfg.PersistFollower = false

	feService = service.NewTokenisationService(cfg)
	go feService.Start()

	feService.WaitForRunning()

	// Run all tests
	code := m.Run()

	// 🧹 Global teardown
	fmt.Println("<<< TEARDOWN: Clean up resources")

	dogeTest.Stop()
	// Exit with the correct status
	os.Exit(code)
}

func TestFractal(t *testing.T) {
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

	unspent, err := dogeTest.Rpc.ListUnspent(addressBook.Addresses[0].Address)
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

	createResp, err := dogeTest.Rpc.Request("createrawtransaction", []interface{}{inputs, outputs})

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

	signResp, err := dogeTest.Rpc.Request("signrawtransaction", []interface{}{hex.EncodeToString(rawTxBytes), prevTxs, privkeys})
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
	sendResp, err := dogeTest.Rpc.Request("sendrawtransaction", []interface{}{signedTx})
	if err != nil {
		log.Fatalf("Error broadcasting transaction: %v", err)
	}

	var txID string
	if err := json.Unmarshal(*sendResp, &txID); err != nil {
		log.Fatalf("Error parsing transaction ID: %v", err)
	}

	fmt.Printf("Transaction sent successfully! TXID: %s\n", txID)

	_, err = dogeTest.ConfirmBlocks()
	if err != nil {
		log.Fatal(err)
	}

	for {

		mints, err := feService.Store.GetMints(0, 1, true)
		if err != nil {
			log.Fatal(err)
		}

		if len(mints) > 0 {
			assert.Equal(t, mints[0].Title, "Test Mint")
			assert.Equal(t, mints[0].Description, "Test Description")
			assert.Equal(t, mints[0].FractionCount, 100)

			break
		}

		time.Sleep(1 * time.Second)
	}
}
