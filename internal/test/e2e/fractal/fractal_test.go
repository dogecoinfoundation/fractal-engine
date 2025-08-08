package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/client"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"github.com/testcontainers/testcontainers-go/network"
	"gotest.tools/assert"
)

var testGroups []*support.TestGroup

func TestMain(m *testing.M) {
	// 🚀 Global setup
	log.Println(">>> SETUP: Init resources")

	if err := support.InitSharedPostgres(); err != nil {
		panic(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Println("Received signal, stopping test")
		os.Exit(0)
	}()

	ctx := context.Background()
	net, err := network.New(ctx, network.WithDriver("bridge"))
	if err != nil {
		panic(err)
	}

	networkName := net.Name

	testGroups = []*support.TestGroup{
		support.NewTestGroup("alpha", networkName, 0, 20000, 21000, 8086, 22555, 33070, 8890),
		support.NewTestGroup("beta", networkName, 1, 20001, 21001, 8087, 22556, 33071, 8891),
	}

	for _, testGroup := range testGroups {
		log.Println("Starting test group", testGroup.Name)
		testGroup.Start(m)
	}

	log.Println("Test groups started")

	err = support.ConnectDogeNetPeers(testGroups[0].DogeNetClient, testGroups[1].DogenetContainer, testGroups[1].DnGossipPort, testGroups[0].LogConsumer, testGroups[1].LogConsumer)
	if err != nil {
		panic(err)
	}

	// Connect the DogeTest instances for P2P transaction propagation
	err = support.ConnectDogeTestPeers(testGroups[0].DogeTest, testGroups[1].DogeTest)
	if err != nil {
		panic(err)
	}

	// need to wait for the dogenet instance to let the peer be active
	time.Sleep(20 * time.Second)

	// Run all tests
	code := m.Run()

	// 🧹 Global teardown
	log.Println("<<< TEARDOWN: Clean up resources")

	for _, testGroup := range testGroups {
		go testGroup.Stop()
	}

	for _, testGroup := range testGroups {
		for testGroup.Running {
			time.Sleep(1 * time.Second)
			log.Println("Waiting for test group", testGroup.Name, "to stop")
		}
	}

	// Exit with the correct status
	os.Exit(code)
}

/*
*
TestFractal is a test that checks if the fractal engine is working correctly.

1. Mint a token on the OG Node via HTTP API
1a. Send the mint to the dogenet (OG Node) which in turn gossips to the 2nd Node
2. Write the mint to the core (OG Node)
3. Write the mint to the core (2nd Node)
4. Check if the mint is in the store (OG Node)
5. Check if the mint is in the store (2nd Node)
*/
func TestFractal(t *testing.T) {
	feConfigA := testGroups[0].FeConfig

	for _, testGroup := range testGroups {
		err := testGroup.BmClient.TrackAddress(testGroup.AddressBook.Addresses[0].Address)
		if err != nil {
			panic(err)
		}
	}

	privHex, pubHex, _, err := doge.GenerateDogecoinKeypair(doge.PrefixRegtest)
	if err != nil {
		log.Fatal(err)
	}

	feClient := client.NewTokenisationClient("http://"+feConfigA.RpcServerHost+":"+feConfigA.RpcServerPort, privHex, pubHex)

	mintRequestPayload := rpc.CreateMintRequestPayload{
		Title:         "Test Mint",
		FractionCount: 100,
		Description:   "Test Description",
		Tags:          []string{"test", "mint"},
		Metadata: map[string]interface{}{
			"test": "test",
		},
		Requirements:  map[string]interface{}{},
		LockupOptions: map[string]interface{}{},
		FeedURL:       "https://test.com",
	}

	mintRequest := &rpc.CreateMintRequest{
		Address:   testGroups[0].AddressBook.Addresses[0].Address,
		PublicKey: pubHex,
		Payload:   mintRequestPayload,
	}

	requestPayload, err := json.Marshal(mintRequestPayload)
	if err != nil {
		log.Fatal(err)
	}

	signature, err := doge.SignPayload(requestPayload, privHex)
	if err != nil {
		log.Fatal(err)
	}

	mintRequest.Signature = signature

	mintResponse, err := feClient.Mint(mintRequest)

	if err != nil {
		log.Fatal(err)
	}

	// Write mint to core (OG Node only - will propagate to 2nd node via P2P)
	err = support.WriteMintToCore(testGroups[0].DogeTest, testGroups[0].AddressBook, &mintResponse)
	if err != nil {
		log.Fatal(err)
	}

	for _, tg := range testGroups {
		for _, addressvalidation := range tg.AddressBook.Addresses {
			fmt.Printf("%s: %s -> %s\n", tg.Name, addressvalidation.Label, addressvalidation.Address)
		}
	}

	// OG Node
	for {
		mints, err := testGroups[0].FeService.Store.GetMints(0, 1)
		if err != nil {
			log.Fatal(err)
		}

		if len(mints) > 0 {
			// Use the first address from testGroups[0] address book (testA0)
			ownerAddress := testGroups[0].AddressBook.Addresses[0]

			assert.Equal(t, mints[0].Title, "Test Mint")
			assert.Equal(t, mints[0].Description, "Test Description")
			assert.Equal(t, mints[0].FractionCount, 100)
			assert.Equal(t, mints[0].OwnerAddress, ownerAddress.Address)

			break
		} else {
			log.Println("Waiting for mints to be found on OG Node")
		}

		time.Sleep(1 * time.Second)
	}

	_, err = testGroups[1].DogeTest.ConfirmBlocks()
	if err != nil {
		log.Fatal(err)
	}

	// Node should have been gossiped mint + validated from L1
	for {
		mints, err := testGroups[1].FeService.Store.GetMints(0, 1)
		if err != nil {
			log.Fatal(err)
		}

		if len(mints) > 0 {

			assert.Equal(t, mints[0].Title, "Test Mint")
			assert.Equal(t, mints[0].Description, "Test Description")
			assert.Equal(t, mints[0].FractionCount, 100)
			log.Println("ownerAddress.Address", mints[0].OwnerAddress)
			// assert.Equal(t, mints[0].OwnerAddress, ownerAddress.Address)

			break
		} else {
			log.Println("Waiting for mints to be found on 2nd Node")
		}

		time.Sleep(1 * time.Second)
	}

}
