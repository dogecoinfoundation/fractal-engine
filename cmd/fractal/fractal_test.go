package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"dogecoin.org/chainfollower/pkg/config"
	"dogecoin.org/dogetest/pkg/dogetest"
	"dogecoin.org/fractal-engine/pkg/api"
	"dogecoin.org/fractal-engine/pkg/client"
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
		InstallationPath: "C:\\Program Files\\Dogecoin\\daemon\\dogecoind.exe",
		ConfigPath:       "C:\\Users\\danielw\\AppData\\Roaming\\Dogecoin\\regtest",
	})
	if err != nil {
		log.Fatal(err)
	}

	dogeTest = localDogeTest

	err = dogeTest.ClearProcess()
	if err != nil {
		log.Fatal(err)
	}

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
	server := server.NewFractalServer(&config.Config{
		DbUrl:   "sqlite://testing.db",
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

	httpClient := &http.Client{}
	client := client.NewFractalEngineClient("http://localhost:8080", httpClient)

	mintRes, err := client.CreateMint(api.CreateMintRequest{
		MintWithoutID: protocol.MintWithoutID{
			Title:         "test",
			FractionCount: 8,
			Description:   "test",
			Tags:          []string{"test"},
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
	assert.Equal(t, []string{"test"}, mint.Tags)
	assert.Equal(t, map[string]interface{}{"test": "test"}, mint.Metadata)

}
