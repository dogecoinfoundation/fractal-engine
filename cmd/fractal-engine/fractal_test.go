package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"dogecoin.org/fractal-engine/pkg/client"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/store"
	"dogecoin.org/fractal-engine/pkg/testsupport"
	"github.com/testcontainers/testcontainers-go/network"
	"gotest.tools/assert"
)

var testGroups []*testsupport.TestGroup

func TestMain(m *testing.M) {
	// 🚀 Global setup
	fmt.Println(">>> SETUP: Init resources")

	ctx := context.Background()
	net, err := network.New(ctx, network.WithDriver("bridge"))
	if err != nil {
		panic(err)
	}

	networkName := net.Name

	testGroups = []*testsupport.TestGroup{
		testsupport.NewTestGroup("alpha", networkName, 0, 8086, 44070, 33070),
		testsupport.NewTestGroup("beta", networkName, 1, 8087, 44071, 33071),
		testsupport.NewTestGroup("gamma", networkName, 2, 8088, 44072, 33072),
		testsupport.NewTestGroup("delta", networkName, 3, 8089, 44073, 33073),
	}

	defer func() {
		for _, testGroup := range testGroups {
			testGroup.Stop()
		}
	}()

	for _, testGroup := range testGroups {
		log.Println("Starting test group", testGroup.Name)
		testGroup.Start()
	}

	fmt.Println("Test groups started")

	err = testsupport.ConnectDogeNetPeers(testGroups[0].DogeNetClient, testGroups[1].DogenetContainer, testGroups[1].GossipPort, testGroups[0].LogConsumer, testGroups[1].LogConsumer)
	if err != nil {
		panic(err)
	}

	err = testsupport.ConnectDogeNetPeers(testGroups[1].DogeNetClient, testGroups[2].DogenetContainer, testGroups[2].GossipPort, testGroups[1].LogConsumer, testGroups[2].LogConsumer)
	if err != nil {
		panic(err)
	}

	err = testsupport.ConnectDogeNetPeers(testGroups[2].DogeNetClient, testGroups[3].DogenetContainer, testGroups[3].GossipPort, testGroups[2].LogConsumer, testGroups[3].LogConsumer)
	if err != nil {
		panic(err)
	}

	err = testsupport.ConnectDogeNetPeers(testGroups[3].DogeNetClient, testGroups[0].DogenetContainer, testGroups[0].GossipPort, testGroups[3].LogConsumer, testGroups[0].LogConsumer)
	if err != nil {
		panic(err)
	}

	time.Sleep(15 * time.Second)

	// Run all tests
	code := m.Run()

	// 🧹 Global teardown
	fmt.Println("<<< TEARDOWN: Clean up resources")

	for _, testGroup := range testGroups {
		testGroup.Stop()
	}

	// Exit with the correct status
	os.Exit(code)
}

func TestFractal(t *testing.T) {
	feConfigA := testGroups[0].FeConfig
	feClient := client.NewTokenisationClient("http://" + feConfigA.RpcServerHost + ":" + feConfigA.RpcServerPort)

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

	log.Println("Mint response", mintResponse)
	log.Println("Address book", testGroups[0].AddressBook)
	log.Println("Doge test", testGroups[0].DogeTest)

	err = testsupport.WriteMintToCore(testGroups[0].DogeTest, testGroups[0].AddressBook, &mintResponse)
	if err != nil {
		log.Fatal(err)
	}

	for {
		mints, err := testGroups[0].FeService.Store.GetMints(0, 1)
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
