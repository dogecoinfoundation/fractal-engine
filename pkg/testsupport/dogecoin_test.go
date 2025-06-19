package testsupport

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go/network"
	"gotest.tools/assert"
)

func TestDogecoinContainer(t *testing.T) {
	ctx := context.Background()

	net, err := network.New(ctx, network.WithDriver("bridge"))
	if err != nil {
		t.Fatal(err)
	}

	networkName := net.Name

	dogecoinContainer, err := StartDogecoinInstance(ctx, "Dockerfile.dogecoin", networkName, "0", "22555")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, dogecoinContainer.IsRunning(), true)

	mappedPort, err := dogecoinContainer.MappedPort(ctx, nat.Port("22555/tcp"))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	rpcClient := doge.NewRpcClient(&config.Config{
		DogeHost:     "localhost",
		DogeScheme:   "http",
		DogePort:     mappedPort.Port(),
		DogeUser:     "test",
		DogePassword: "test",
	})

	res, err := rpcClient.Request("listunspent", []any{0, 999999999, []string{"DQ6666666666666666666666666666666666666666666666666666666666666666"}})
	if err != nil {
		t.Fatal(err)
	}

	var result []doge.UTXO
	err = json.Unmarshal(*res, &result)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(result), 0)
}
