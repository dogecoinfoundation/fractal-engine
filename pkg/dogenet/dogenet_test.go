package dogenet_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"dogecoin.org/fractal-engine/pkg/testsupport"
	"github.com/Dogebox-WG/gossip/dnet"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"
)

var dogenetClientA *dogenet.DogeNetClient
var dogenetClientB *dogenet.DogeNetClient
var dogenetA testcontainers.Container
var dogenetB testcontainers.Container

func TestMain(m *testing.M) {
	ctx := context.Background()

	net, err := network.New(ctx, network.WithDriver("bridge"))
	if err != nil {
		panic(err)
	}
	networkName := net.Name

	tokenisationStore, err := store.NewTokenisationStore("sqlite://test.db", config.Config{})
	if err != nil {
		panic(err)
	}

	feKey, err := dnet.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	logConsumerA := &testsupport.StdoutLogConsumer{Name: "alpha"}
	dogenetClientA, dogenetA, err = testsupport.StartDogenetInstance(ctx, feKey, "Dockerfile.dogenet", "alpha", "8085", "44069", "33069", networkName, logConsumerA, tokenisationStore)
	if err != nil {
		panic(err)
	}

	logConsumerB := &testsupport.StdoutLogConsumer{Name: "beta"}
	dogenetClientB, dogenetB, err = testsupport.StartDogenetInstance(ctx, feKey, "Dockerfile.dogenet", "beta", "8086", "44070", "33070", networkName, logConsumerB, tokenisationStore)
	if err != nil {
		panic(err)
	}

	err = testsupport.ConnectDogeNetPeers(dogenetClientA, dogenetB, 33070, logConsumerA, logConsumerB)
	if err != nil {
		panic(err)
	}

	time.Sleep(15 * time.Second)

	m.Run()

	fmt.Println("Cleaning up resources...")

	net.Remove(ctx)
	dogenetA.Terminate(ctx)
	dogenetB.Terminate(ctx)
}

func TestMintMessage(t *testing.T) {
	err := dogenetClientA.GossipMint(store.Mint{
		Id: "1",
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
		assert.Error(t, err, "failed to gossip mint")
	} else {
		fmt.Println("Mint gossiped successfully")
	}

	fmt.Println("Waiting for message...")

	msg := <-dogenetClientB.Messages

	payload := protocol.MessageEnvelope{}
	err = payload.Deserialize(msg.Payload)
	if err != nil {
		panic(err)
	}

	switch payload.Action {
	case protocol.ACTION_MINT:
		mint := protocol.MintMessage{}
		err = proto.Unmarshal(payload.Data, &mint)
		if err != nil {
			assert.Error(t, err, "expected mint message")
		}

		assert.Equal(t, mint.Title, "Test Mint")
		assert.Equal(t, mint.Description, "Test Description")
		assert.Equal(t, mint.FractionCount, int32(100))

	default:
		assert.Error(t, fmt.Errorf("expected mint message"), "expected mint message")
	}
}
