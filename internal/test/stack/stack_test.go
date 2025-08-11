package stack

import (
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"testing"

	feclient "dogecoin.org/fractal-engine/pkg/client"
	fecfg "dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	bmclient "github.com/dogecoinfoundation/balance-master/pkg/client"
)

type StackConfig struct {
	InstanceId          int
	BasePort            int
	DogePort            int
	DogeHost            string
	FractalPort         int
	FractalHost         string
	DogeNetPort         int
	DogeNetHost         string
	DogeNetBindPort     int
	DogeNetPubKey       string
	DogeNetWebPort      int
	BalanceMasterPort   int
	BalanceMasterHost   string
	PortgresPort        int
	PostgresHost        string
	DogeNetHandlerPort  int
	PrivKey             string
	PubKey              string
	Address             string
	TokenisationClient  *feclient.TokenisationClient
	BalanceMasterClient *bmclient.BalanceMasterClient
	DogeClient          *doge.RpcClient
	DogeNetClient       dogenet.GossipClient
	TokenisationStore   *store.TokenisationStore
}

func NewStackConfig(instanceId int, chain string) StackConfig {
	prefixByte, err := doge.GetPrefix(chain)
	if err != nil {
		panic(err)
	}

	basePort := 8000 + (instanceId * 100)
	privHex, pubHex, address, err := doge.GenerateDogecoinKeypair(prefixByte)
	if err != nil {
		panic(err)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}

	stackConfig := StackConfig{
		InstanceId:         instanceId,
		BasePort:           basePort,
		DogePort:           basePort + 14556,
		FractalPort:        basePort + 2,
		DogeNetPort:        basePort + 3,
		DogeNetWebPort:     basePort + 4,
		BalanceMasterPort:  basePort + 5,
		PortgresPort:       basePort + 6,
		DogeNetHandlerPort: basePort + 7,
		DogeNetBindPort:    42000 + instanceId,
		Address:            address,
		PrivKey:            privHex,
		PubKey:             pubHex,
	}

	populateStackHosts(&stackConfig, cli)

	stackConfig.TokenisationClient = feclient.NewTokenisationClient("http://"+stackConfig.FractalHost+":"+strconv.Itoa(stackConfig.FractalPort), stackConfig.PrivKey, stackConfig.PubKey)
	stackConfig.BalanceMasterClient = bmclient.NewBalanceMasterClient(&bmclient.BalanceMasterClientConfig{
		RpcServerHost: stackConfig.BalanceMasterHost,
		RpcServerPort: strconv.Itoa(stackConfig.BalanceMasterPort),
	})
	stackConfig.DogeClient = doge.NewRpcClient(&fecfg.Config{
		DogeScheme:   "http",
		DogeHost:     "localhost",
		DogePort:     strconv.Itoa(stackConfig.DogePort),
		DogeUser:     "test",
		DogePassword: "test",
	})

	tokenStore, err := store.NewTokenisationStore("postgres://fractalstore:fractalstore@"+stackConfig.PostgresHost+":"+strconv.Itoa(stackConfig.PortgresPort)+"/fractalstore?sslmode=disable", fecfg.Config{
		MigrationsPath: "../../../db/migrations",
	})

	stackConfig.TokenisationStore = tokenStore
	if err != nil {
		panic(err)
	}

	err = tokenStore.Migrate()
	if err != nil && err.Error() != "no change" {
		panic(err)
	}

	stackConfig.DogeNetClient = dogenet.NewDogeNetClient(&fecfg.Config{
		DogeNetWebAddress: "localhost" + ":" + strconv.Itoa(stackConfig.DogeNetWebPort),
	}, tokenStore)

	return stackConfig
}

func TestStack(t *testing.T) {
	stackCount := 2

	var stacks []*StackConfig
	for i := 0; i < stackCount; i++ {
		newConfig := NewStackConfig(i+1, "regtest")
		stacks = append(stacks, &newConfig)
	}

	for i := 0; i < len(stacks)/2; i += 2 {
		stackA := stacks[i]
		stackB := stacks[(i + 1)]

		fmt.Println("StackA: ", stackA.Address)
		fmt.Println("StackB: ", stackB.Address)

		// Check for nodes, if doesnt exist, then add peer.
		err := stackA.DogeNetClient.AddPeer(dogenet.AddPeer{
			Key:  stackB.DogeNetPubKey,
			Addr: stackB.DogeNetHost + ":" + strconv.Itoa(stackB.DogeNetBindPort),
		})
		if err != nil {
			panic(err)
		}

		err = stackA.DogeClient.AddPeer(stackB.DogeHost)
		if err != nil {
			panic(err)
		}
	}

}

func populateStackHosts(stackConfig *StackConfig, cli *client.Client) {
	ctx := context.Background()
	inspectRes, err := cli.NetworkInspect(ctx, "fractal-shared", network.InspectOptions{})
	if err != nil {
		panic(err)
	}

	instanceId := strconv.Itoa(stackConfig.InstanceId)

	for _, ct := range inspectRes.Containers {
		if ct.Name == "fractalengine-"+instanceId {
			stackConfig.FractalHost = strings.Split(ct.IPv4Address, "/")[0]
		}

		if ct.Name == "dogecoin-"+instanceId {
			stackConfig.DogeHost = strings.Split(ct.IPv4Address, "/")[0]
		}

		if ct.Name == "dogenet-"+instanceId {
			stackConfig.DogeNetHost = strings.Split(ct.IPv4Address, "/")[0]
			res, err := cli.ContainerLogs(ctx, ct.Name, container.LogsOptions{
				ShowStderr: true,
			})
			if err != nil {
				panic(err)
			}

			logBytes, err := io.ReadAll(res)
			if err != nil {
				panic(err)
			}

			logs := string(logBytes)
			re := regexp.MustCompile(`Node PubKey is: ([0-9a-fA-F]+)`)
			matches := re.FindStringSubmatch(logs)
			if len(matches) > 1 {
				stackConfig.DogeNetPubKey = matches[1]
			}
		}

		if ct.Name == "fractalstore-"+instanceId {
			stackConfig.PostgresHost = "localhost" // Connect from outside Docker context
		}

		if ct.Name == "balance-master-"+instanceId {
			stackConfig.BalanceMasterHost = strings.Split(ct.IPv4Address, "/")[0]
		}
	}

}
