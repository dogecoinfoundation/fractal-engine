package dogenet

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/Dogebox-WG/gossip/dnet"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"
)

type StdoutLogConsumer struct {
	Name string
}

var pubKeys = map[string]string{}
var dogenetClient *DogeNetClient
var dogenetClientB *DogeNetClient
var dogenetA testcontainers.Container
var dogenetB testcontainers.Container

// Accept prints the log to stdout
func (lc *StdoutLogConsumer) Accept(l testcontainers.Log) {
	content := string(l.Content)

	if strings.Contains(content, "Node PubKey is: ") {
		pubKey := strings.Split(content, "Node PubKey is: ")[1]
		pubKeys[lc.Name] = strings.Trim(pubKey, "\n")
	}
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	net, err := network.New(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := net.Remove(ctx); err != nil {
			log.Printf("failed to remove network: %s", err)
		}
	}()

	networkName := net.Name

	dogenetClient, dogenetA = StartDogenetInstance(ctx, "Dockerfile.dogenet", "alpha", "8085", "44069", "33069", networkName)

	defer func() {
		fmt.Println("Terminating container...")
		_ = dogenetA.Terminate(ctx)
	}()

	dogenetClientB, dogenetB = StartDogenetInstance(ctx, "Dockerfile.dogenet2", "beta", "8086", "44070", "33070", networkName)
	defer func() {
		fmt.Println("Terminating container...")
		_ = dogenetB.Terminate(ctx)
	}()

	time.Sleep(5 * time.Second)

	ipB, err := dogenetB.ContainerIP(ctx)
	if err != nil {
		panic(err)
	}

	for {
		if len(pubKeys) == 2 {
			break
		}

		fmt.Printf("Waiting for pub keys... %v\n", pubKeys)

		time.Sleep(1 * time.Second)
	}

	peerAddressB := ipB + ":" + "33070"

	fmt.Println(pubKeys["beta"])

	err = dogenetClient.AddPeer(AddPeer{
		Key:  pubKeys["beta"],
		Addr: peerAddressB,
	})

	fmt.Println("Adding peer...", peerAddressB)

	if err != nil {
		panic(err)
	}

	time.Sleep(2 * time.Second)

	os.Exit(m.Run())
}

func TestMintMessage(t *testing.T) {
	err := dogenetClient.GossipMint(store.Mint{
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

func StartDogenetInstance(ctx context.Context, image string, instanceId string, webPort string, port string, gossipPort string, networkName string) (*DogeNetClient, testcontainers.Container) {
	nodeKey, err := dnet.GenerateKeyPair()
	if err != nil {
		panic(fmt.Sprintf("cannot generate node keypair: %v", err))
	}

	absPathContext, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		panic(err)
	}

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    absPathContext,
			Dockerfile: image,
			KeepImage:  false,
		},
		Networks: []string{
			networkName,
		},
		Name:         "dogenet-" + instanceId,
		ExposedPorts: []string{webPort + "/tcp", port + "/tcp", gossipPort + "/tcp"},
		Env:          map[string]string{},
		WaitingFor:   wait.ForLog("[gossip] listening on").WithStartupTimeout(10 * time.Second),
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericTmpfsMountSource{},
				Target: "/root/storage",
			},
		},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Opts: []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
			Consumers: []testcontainers.LogConsumer{&StdoutLogConsumer{
				Name: instanceId,
			}},
		},
	}

	dogenetA, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	for {
		if dogenetA.IsRunning() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	for {
		handlerPort, err := dogenetA.MappedPort(ctx, nat.Port(port+"/tcp"))
		if err == nil {
			fmt.Printf("Handler port mapped to %s\n", handlerPort.Port())
			break
		}

		fmt.Printf("Waiting for handler port to be mapped... %s\n", err)

		time.Sleep(1 * time.Second)
	}

	for {
		webPortRes, err := dogenetA.MappedPort(ctx, nat.Port(webPort+"/tcp"))
		if err == nil {
			fmt.Printf("Web port mapped to %s\n", webPortRes.Port())
			break
		}

		fmt.Printf("Waiting for web port to be mapped... %s\n", err)

		time.Sleep(1 * time.Second)
	}

	ip, _ := dogenetA.Host(ctx)
	mappedPortWeb, _ := dogenetA.MappedPort(ctx, nat.Port(webPort+"/tcp"))
	mappedPort, _ := dogenetA.MappedPort(ctx, nat.Port(port+"/tcp"))

	fmt.Printf("Dogenet is running at %s:%s\n", ip, mappedPort.Port())

	dogenetConfig := &config.Config{
		DogeNetNetwork:    "tcp",
		DogeNetAddress:    ip + ":" + mappedPort.Port(),
		DogeNetWebAddress: ip + ":" + mappedPortWeb.Port(),
		DogeNetKeyPair:    nodeKey,
	}

	store, err := store.NewTokenisationStore("sqlite3://test.db", config.Config{})
	if err != nil {
		panic(err)
	}

	dogenetClient := NewDogeNetClient(dogenetConfig, store)

	for {
		err := dogenetClient.CheckRunning()
		if err == nil {
			break
		}

		fmt.Println(err)

		time.Sleep(1 * time.Second)
	}

	fmt.Println("Dogenet is running")

	dogenetStatusChan := make(chan string)
	go dogenetClient.Start(dogenetStatusChan)

	<-dogenetStatusChan

	log.Println("dogenetClient started")
	return dogenetClient, dogenetA
}
