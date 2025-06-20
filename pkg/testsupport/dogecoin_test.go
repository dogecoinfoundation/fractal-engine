package testsupport

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"github.com/dogecoinfoundation/dogetest/pkg/dogetest"
	"github.com/testcontainers/testcontainers-go/network"
	"gotest.tools/assert"
)

func TestDogecoinContainer(t *testing.T) {
	ctx := context.Background()

	net, err := network.New(ctx, network.WithDriver("bridge"))
	if err != nil {
		panic(err)
	}

	networkName := net.Name

	dogeTest, err := dogetest.NewDogeTest(dogetest.DogeTestConfig{
		NetworkName: networkName,
		Port:        22555,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = dogeTest.Start()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	mappedPort, err := dogeTest.Container.MappedPort(ctx, "22555/tcp")
	if err != nil {
		t.Fatal(err)
	}

	rpcClient := doge.NewRpcClient(&config.Config{
		DogeHost:     "localhost",
		DogeScheme:   "http",
		DogePort:     mappedPort.Port(),
		DogeUser:     "test",
		DogePassword: "test",
	})

	res, err := rpcClient.Request("getblockchaininfo", []any{})
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]any
	err = json.Unmarshal(*res, &result)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, result["chain"], "regtest")
}
