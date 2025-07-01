package e2e

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/client"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"github.com/testcontainers/testcontainers-go/network"
	"gotest.tools/assert"
)

var testGroups []*support.TestGroup

func TestMain(m *testing.M) {
	// 🚀 Global setup
	log.Println(">>> SETUP: Init resources")

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
		support.NewTestGroup("alpha", networkName, 0, 20000, 21000, 8086, 22555, 33070),
		support.NewTestGroup("beta", networkName, 1, 20001, 21001, 8087, 22556, 33071),
	}

	for _, testGroup := range testGroups {
		log.Println("Starting test group", testGroup.Name)
		testGroup.Start()
	}

	log.Println("Test groups started")

	err = support.ConnectDogeNetPeers(testGroups[0].DogeNetClient, testGroups[1].DogenetContainer, testGroups[1].DnGossipPort, testGroups[0].LogConsumer, testGroups[1].LogConsumer)
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
	feClient := client.NewTokenisationClient("http://" + feConfigA.RpcServerHost + ":" + feConfigA.RpcServerPort)

	mintResponse, err := feClient.Mint(&rpc.CreateMintRequest{
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
		OwnerAddress:  "testA0",
	})

	if err != nil {
		log.Fatal(err)
	}

	// Write mint to core (OG Node)
	err = support.WriteMintToCore(testGroups[0].DogeTest, testGroups[0].AddressBook, &mintResponse)
	if err != nil {
		log.Fatal(err)
	}

	// Write mint to core (2nd node)
	err = support.WriteMintToCore(testGroups[1].DogeTest, testGroups[1].AddressBook, &mintResponse)
	if err != nil {
		log.Fatal(err)
	}

	// OG Node
	for {
		mints, err := testGroups[0].FeService.Store.GetMints(0, 1)
		if err != nil {
			log.Fatal(err)
		}

		if len(mints) > 0 {
			ownerAddress, err := testGroups[0].AddressBook.GetAddress("testA0")
			if err != nil {
				log.Fatal(err)
			}

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

	// Node should have been gossiped mint + validated from L1
	for {
		mints, err := testGroups[1].FeService.Store.GetMints(0, 1)
		if err != nil {
			log.Fatal(err)
		}

		if len(mints) > 0 {
			ownerAddress, err := testGroups[1].AddressBook.GetAddress("testA1")
			if err != nil {
				log.Fatal(err)
			}

			assert.Equal(t, mints[0].Title, "Test Mint")
			assert.Equal(t, mints[0].Description, "Test Description")
			assert.Equal(t, mints[0].FractionCount, 100)
			assert.Equal(t, mints[0].OwnerAddress, ownerAddress.Address)

			break
		} else {
			log.Println("Waiting for mints to be found on 2nd Node")
		}

		time.Sleep(1 * time.Second)
	}

}
