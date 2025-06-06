package dogenet

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/Dogebox-WG/gossip/dnet"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type StdoutLogConsumer struct{}

// Accept prints the log to stdout
func (lc *StdoutLogConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	dogenetClient, dogenetA := StartDogenetInstance(ctx, "Dockerfile.dogenet", "8085/tcp")
	defer func() {
		fmt.Println("Terminating container...")
		_ = dogenetA.Terminate(ctx)
	}()

	dogenetClientB, dogenetB := StartDogenetInstance(ctx, "Dockerfile.dogenet2", "8086/tcp")
	defer func() {
		fmt.Println("Terminating container...")
		_ = dogenetB.Terminate(ctx)
	}()

	err := dogenetClient.AddPeer(AddPeer{
		Key:  string(dogenetClientB.feKey.Pub[:]),
		Addr: dogenetClientB.cfg.DogeNetAddress,
	})

	if err != nil {
		panic(err)
	}

	time.Sleep(2 * time.Second)

	err = dogenetClient.GossipMint(store.Mint{
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
		panic(err)
	}

	time.Sleep(10 * time.Second)
}

func StartDogenetInstance(ctx context.Context, image string, port string) (*DogeNetClient, testcontainers.Container) {
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
		},
		ExposedPorts: []string{port + "/tcp"},
		Env:          map[string]string{},
		WaitingFor:   wait.ForLog("[gossip] listening on").WithStartupTimeout(10 * time.Second),
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericTmpfsMountSource{},
				Target: "/root/storage",
			},
		},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
			Consumers: []testcontainers.LogConsumer{&StdoutLogConsumer{}},
		},
	}

	dogenetA, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	ip, _ := dogenetA.Host(ctx)
	mappedPort, _ := dogenetA.MappedPort(ctx, nat.Port(port+"/tcp"))

	fmt.Printf("Dogenet is running at %s:%s\n", ip, mappedPort.Port())

	dogenetConfig := &config.Config{
		DogeNetNetwork: "tcp",
		DogeNetAddress: ip + ":" + mappedPort.Port(),
		DogeNetKeyPair: nodeKey,
	}

	dogenetClient := NewDogeNetClient(dogenetConfig)

	dogenetStatusChan := make(chan string)
	go dogenetClient.Start(dogenetStatusChan)

	<-dogenetStatusChan

	log.Println("dogenetClient started")
	return dogenetClient, dogenetA
}
