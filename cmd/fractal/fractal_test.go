package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"dogecoin.org/chainfollower/pkg/config"
	"dogecoin.org/dogeclient/pkg/dogeclient"
	"dogecoin.org/dogetest/pkg/dogetest"
	"dogecoin.org/fractal-engine/pkg/api"
	"dogecoin.org/fractal-engine/pkg/client"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/server"

	"gotest.tools/v3/assert"
)

var blocks []string
var addressBook *dogetest.AddressBook
var dogeTest *dogetest.DogeTest

func TestMain(m *testing.M) {
	// 🚀 Global setup
	fmt.Println(">>> SETUP: Init resources")

	localDogeTest, err := dogetest.NewDogeTest(dogetest.DogeTestConfig{
		Host:             "localhost",
		InstallationPath: "/Applications/Dogecoin-Qt.app/Contents/MacOS/Dogecoin-Qt",
		ConfigPath:       "/Users/danielw/Library/Application Support/Dogecoin/regtest",
	})
	if err != nil {
		log.Fatal(err)
	}

	dogeTest = localDogeTest

	// err = dogeTest.ClearProcess()
	// if err != nil {
	// 	log.Fatal(err)
	// }

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

	// Run all tests
	code := m.Run()

	// 🧹 Global teardown
	fmt.Println("<<< TEARDOWN: Clean up resources")

	dogeTest.Stop()
	// Exit with the correct status
	os.Exit(code)
}

func TestFractal(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "fractal-test-12345.db")

	os.Remove(tmpFile)

	fmt.Println("tmpFile", tmpFile)

	server := server.NewFractalServer(&config.Config{
		DbUrl:   "sqlite:" + tmpFile,
		RpcUrl:  fmt.Sprintf("http://%s:%d", dogeTest.Host, dogeTest.Port),
		RpcUser: "test",
		RpcPass: "test",
	})

	status := make(chan string)
	go server.Start(status)

	for status := range status {
		if status == "started" {
			break
		}
	}

	err := server.Store.ClearMints()
	if err != nil {
		t.Fatal(err)
	}

	httpClient := &http.Client{}
	client := client.NewFractalEngineClient("http://localhost:8080", httpClient)

	mintRes, err := client.CreateMint(api.CreateMintRequest{
		MintWithoutID: protocol.MintWithoutID{
			Title:         "test",
			FractionCount: 8,
			Description:   "test",
			Tags:          []string{"test"},
			OutputAddress: addressBook.Addresses[0].Address,
			Metadata:      map[string]interface{}{"test": "test"},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	mints, err := client.GetMints(0, 10)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, mints.Total)

	mint := mints.Mints[0]

	assert.Equal(t, mintRes.Id, mint.Id)
	assert.Equal(t, mintRes.Hash, mint.Hash)
	assert.Equal(t, "test", mint.Title)
	assert.Equal(t, 8, mint.FractionCount)
	assert.Equal(t, "test", mint.Description)
	assert.Equal(t, "test", mint.Tags[0])

	dogeClient := dogeclient.NewRpcClient(&dogeclient.Config{
		RpcUrl:  fmt.Sprintf("http://%s:%d", dogeTest.Host, dogeTest.Port),
		RpcUser: "test",
		RpcPass: "test",
	})

	fractalDogeClient := doge.NewFractalDogeClient(dogeClient)

	unspent, err := fractalDogeClient.GetUnspent(addressBook.Addresses[0].Address)
	if err != nil {
		t.Fatal(err)
	}

	txHash, err := fractalDogeClient.CreateMint(&mint, addressBook.Addresses[0].PrivateKey, unspent, addressBook.Addresses[0].Address)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(txHash)

	blocks, err := dogeTest.ConfirmBlocks()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(blocks)

	for status := range status {
		if status == "mint verified" {
			break
		}
	}

	mints2, err := server.Store.GetMints(0, 10, true)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(mints2))

	mint = mints2[0]

	assert.Equal(t, txHash, mint.TransactionHash)
	assert.Equal(t, addressBook.Addresses[1].Address, mint.OutputAddress)
}
