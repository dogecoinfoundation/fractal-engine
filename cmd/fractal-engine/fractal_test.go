package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		fmt.Println("Received signal, stopping test")
		os.Exit(0)
	}()

	ctx := context.Background()
	net, err := network.New(ctx, network.WithDriver("bridge"))
	if err != nil {
		panic(err)
	}

	networkName := net.Name

	testGroups = []*testsupport.TestGroup{
		testsupport.NewTestGroup("alpha", networkName, 0, 8086, 22555, 33070),
		testsupport.NewTestGroup("beta", networkName, 1, 8087, 22556, 33071),
	}

	for _, testGroup := range testGroups {
		log.Println("Starting test group", testGroup.Name)
		testGroup.Start()
	}

	fmt.Println("Test groups started")

	// time.Sleep(40 * time.Second)

	err = testsupport.ConnectDogeNetPeers(testGroups[0].DogeNetClient, testGroups[1].DogenetContainer, testGroups[1].GossipPort, testGroups[0].LogConsumer, testGroups[1].LogConsumer)
	if err != nil {
		panic(err)
	}

	// time.Sleep(45 * time.Second)

	// Run all tests
	code := m.Run()

	// 🧹 Global teardown
	fmt.Println("<<< TEARDOWN: Clean up resources")

	for _, testGroup := range testGroups {
		go testGroup.Stop()
	}

	for _, testGroup := range testGroups {
		for testGroup.Running {
			time.Sleep(1 * time.Second)
			fmt.Println("Waiting for test group", testGroup.Name, "to stop")
		}
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

	// OG Node
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
		} else {
			fmt.Println("Waiting for mints to be found on OG Node")
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
			assert.Equal(t, mints[0].Title, "Test Mint")
			assert.Equal(t, mints[0].Description, "Test Description")
			assert.Equal(t, mints[0].FractionCount, 100)

			break
		} else {
			fmt.Println("Waiting for mints to be found on 2nd Node")
		}

		time.Sleep(1 * time.Second)
	}
}
