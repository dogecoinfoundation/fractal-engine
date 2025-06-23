package testsupport

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/service"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/Dogebox-WG/gossip/dnet"
	"github.com/docker/go-connections/nat"
	"github.com/dogecoinfoundation/dogetest/pkg/dogetest"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestGroup struct {
	InstanceId       int
	Name             string
	LogConsumer      *StdoutLogConsumer
	NetworkName      string
	DogeTest         *dogetest.DogeTest
	AddressBook      *dogetest.AddressBook
	FeService        *service.TokenisationService
	FeConfig         *config.Config
	DogeNetClient    *dogenet.DogeNetClient
	NodeKey          dnet.KeyPair
	DogenetContainer testcontainers.Container
	Running          bool
	DogeCorePort     int
	DnWebPort        int
	DnPort           int
	DnGossipPort     int
	FeRpcPort        int
}

func (tg *TestGroup) Stop() {
	if tg.DogeTest != nil {
		err := tg.DogeTest.Stop()
		if err != nil {
			log.Println("Error stopping doge test:", err)
		}
	}

	if tg.DogenetContainer != nil {
		err := tg.DogenetContainer.Stop(context.Background(), nil)
		if err != nil {
			log.Println("Error stopping dogenet container:", err)
		}
	}

	if tg.FeService != nil {
		tg.FeService.Stop()
	}

	tg.Running = false
}

func NewTestGroup(name string, networkName string, instanceId int, dogeCorePort int, feRpcPort int, dnWebPort int, dnPort int, dnGossipPort int) *TestGroup {
	nodeKey, err := dnet.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	dogeTest, err := dogetest.NewDogeTest(dogetest.DogeTestConfig{
		Host:          "localhost",
		NetworkName:   networkName,
		Port:          dogeCorePort,
		LogContainers: false,
	})
	if err != nil {
		panic(err)
	}

	return &TestGroup{
		InstanceId:   instanceId,
		Name:         name,
		NetworkName:  networkName,
		DogeTest:     dogeTest,
		NodeKey:      nodeKey,
		DogeCorePort: dogeCorePort,
		DnWebPort:    dnWebPort,
		DnPort:       dnPort,
		DnGossipPort: dnGossipPort,
		FeRpcPort:    feRpcPort,
	}
}

func (tg *TestGroup) Start() {
	tg.Running = true
	ctx := context.Background()

	err := tg.DogeTest.Start()
	if err != nil {
		panic(err)
	}

	addressBook, err := tg.DogeTest.SetupAddresses([]dogetest.AddressSetup{
		{
			Label:          "testA" + strconv.Itoa(tg.InstanceId),
			InitialBalance: 100,
		},
		{
			Label:          "testB" + strconv.Itoa(tg.InstanceId),
			InitialBalance: 20,
		},
	})
	if err != nil {
		panic(err)
	}

	tg.AddressBook = addressBook

	_, err = tg.DogeTest.ConfirmBlocks()
	if err != nil {
		panic(err)
	}

	dbUrl := "sqlite://test" + strconv.Itoa(tg.InstanceId) + ".db"

	tokenStore, err := store.NewTokenisationStore(dbUrl, config.Config{
		MigrationsPath: "../../db/migrations",
	})
	if err != nil {
		panic(err)
	}

	nodeKey, err := dnet.GenerateKeyPair()
	if err != nil {
		log.Fatal(err)
	}

	logConsumer := &StdoutLogConsumer{Name: tg.Name}
	tg.LogConsumer = logConsumer

	dogenetClient, dogenetContainer, err := StartDogenetInstance(ctx, nodeKey, "Dockerfile.dogenet", strconv.Itoa(tg.InstanceId), strconv.Itoa(tg.DnWebPort), strconv.Itoa(tg.DnPort), strconv.Itoa(tg.DnGossipPort), tg.NetworkName, logConsumer, tokenStore)
	if err != nil {
		panic(err)
	}

	feService, feConfig := SetupFractalEngineTestInstance(tg.InstanceId, strconv.Itoa(tg.FeRpcPort), tg.DogeCorePort, dogenetClient, tg.DogeTest, tokenStore)
	tg.FeService = feService
	tg.FeConfig = feConfig

	tg.DogeNetClient = dogenetClient
	tg.DogenetContainer = dogenetContainer

	go feService.Start()

	feService.WaitForRunning()
}

func SetupFractalEngineTestInstance(instanceId int, rpcServerPort string, dogePort int, dogenetClient *dogenet.DogeNetClient, dogeTestInstance *dogetest.DogeTest, tokenStore *store.TokenisationStore) (*service.TokenisationService, *config.Config) {
	os.Remove("test" + strconv.Itoa(instanceId) + ".db")

	mappedPort, err := dogeTestInstance.Container.MappedPort(context.Background(), nat.Port(strconv.Itoa(dogePort)+"/tcp"))
	if err != nil {
		log.Fatal(err)
	}

	feConfig := config.NewConfig()
	feConfig.DogeHost = dogeTestInstance.Host
	feConfig.DatabaseURL = "sqlite://test" + strconv.Itoa(instanceId) + ".db"
	feConfig.DogePort = mappedPort.Port()
	feConfig.DogeUser = "test"
	feConfig.DogePassword = "test"
	feConfig.RpcServerPort = rpcServerPort

	// feConfig.PersistFollower = false
	feConfig.MigrationsPath = "../../db/migrations"

	feService := service.NewTokenisationService(feConfig, dogenetClient, tokenStore)

	return feService, feConfig
}

func ConnectDogeNetPeers(fromPeer *dogenet.DogeNetClient, toPeerContainer testcontainers.Container, toPeerPort int, logConsumerA *StdoutLogConsumer, logConsumerB *StdoutLogConsumer) error {
	ctx := context.Background()

	ipB, err := toPeerContainer.ContainerIP(ctx)
	if err != nil {
		return err
	}

	for {
		if logConsumerA.PubKey != "" && logConsumerB.PubKey != "" {
			break
		}

		fmt.Printf("Waiting for pub keys... ")

		time.Sleep(1 * time.Second)
	}

	peerAddressB := ipB + ":" + strconv.Itoa(toPeerPort)

	err = fromPeer.AddPeer(dogenet.AddPeer{
		Key:  logConsumerB.PubKey,
		Addr: peerAddressB,
	})
	if err != nil {
		return err
	}

	fmt.Println("Adding peer...", peerAddressB)

	nodes, err := fromPeer.GetNodes()
	if err != nil {
		return err
	}

	for _, node := range nodes {
		fmt.Println("Node:", node)
	}

	return nil
}

type StdoutLogConsumer struct {
	Name   string
	PubKey string
}

// Accept prints the log to stdout
func (lc *StdoutLogConsumer) Accept(l testcontainers.Log) {
	content := string(l.Content)

	if strings.Contains(content, "Node PubKey is: ") {
		pubKey := strings.Split(content, "Node PubKey is: ")[1]
		lc.PubKey = strings.Trim(pubKey, "\n")
	}

	// log.Println(content)
}

func StartDogenetInstance(ctx context.Context, feKey dnet.KeyPair, image string, instanceId string, webPort string, port string, gossipPort string, networkName string, logConsumer *StdoutLogConsumer, tokenisationStore *store.TokenisationStore) (*dogenet.DogeNetClient, testcontainers.Container, error) {
	nodeKey, err := dnet.GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}

	absPathContext, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return nil, nil, err
	}

	cacheBuster := strconv.FormatInt(time.Now().UnixNano(), 10)

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    absPathContext,
			Dockerfile: image,
			KeepImage:  false,
			BuildArgs: map[string]*string{
				// Bump the value each time to bust cache
				"CACHE_BUSTER": &cacheBuster,
			},
		},
		Networks: []string{
			networkName,
		},
		Name:         "dogenet-" + instanceId,
		ExposedPorts: []string{webPort + "/tcp", port + "/tcp", gossipPort + "/tcp"},
		Env: map[string]string{
			"KEY":          hex.EncodeToString(nodeKey.Priv[:]),
			"IDENT_KEY":    hex.EncodeToString(feKey.Pub[:]),
			"WEB_PORT":     webPort,
			"HANDLER_PORT": port,
			"BIND_PORT":    gossipPort,
			"CACHE_BUSTER": cacheBuster,
		},
		WaitingFor: wait.ForLog("[gossip] listening on").WithStartupTimeout(10 * time.Second),
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericTmpfsMountSource{},
				Target: "/root/storage",
			},
		},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
			Consumers: []testcontainers.LogConsumer{logConsumer},
		},
	}

	dogenetA, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
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

	dogenetClient := dogenet.NewDogeNetClient(dogenetConfig, tokenisationStore)

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
	return dogenetClient, dogenetA, nil
}

func WriteMintToCore(dogeTest *dogetest.DogeTest, addressBook *dogetest.AddressBook, mintResponse *rpc.CreateMintResponse) error {
	unspent, err := dogeTest.Rpc.ListUnspent(addressBook.Addresses[0].Address)
	if err != nil {
		log.Fatal(err)
	}

	if len(unspent) == 0 {
		return fmt.Errorf("no unspent outputs found")
	}

	selectedUTXO := unspent[0]

	inputs := []map[string]interface{}{
		{
			"txid": selectedUTXO.TxID,
			"vout": selectedUTXO.Vout,
		},
	}

	change := selectedUTXO.Amount - 0.5

	outputs := map[string]interface{}{
		addressBook.Addresses[0].Address: change,
		"data":                           mintResponse.EncodedTransactionBody,
	}

	createResp, err := dogeTest.Rpc.Request("createrawtransaction", []interface{}{inputs, outputs})

	if err != nil {
		log.Fatalf("Error creating raw transaction: %v", err)
	}

	var rawTx string

	if err := json.Unmarshal(*createResp, &rawTx); err != nil {
		log.Fatalf("Error parsing raw transaction: %v", err)
	}

	// Step 3: Add OP_RETURN output to the transaction
	rawTxBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		log.Fatalf("Error decoding raw transaction hex: %v", err)
	}

	prevTxs := []map[string]interface{}{
		{

			"txid":         selectedUTXO.TxID,
			"vout":         selectedUTXO.Vout,
			"scriptPubKey": selectedUTXO.ScriptPubKey,
			"amount":       selectedUTXO.Amount,
		},
	}

	// Prepare privkeys (private keys for signing)
	privkeys := []string{addressBook.Addresses[0].PrivateKey}

	signResp, err := dogeTest.Rpc.Request("signrawtransaction", []interface{}{hex.EncodeToString(rawTxBytes), prevTxs, privkeys})
	if err != nil {
		log.Fatalf("Error signing raw transaction: %v", err)
	}

	var signResult map[string]interface{}
	if err := json.Unmarshal(*signResp, &signResult); err != nil {
		log.Fatalf("Error parsing signed transaction: %v", err)
	}

	signedTx, ok := signResult["hex"].(string)
	if !ok {
		log.Fatal("Error retrieving signed transaction hex.")
	}

	// Step 5: Broadcast the signed transaction
	sendResp, err := dogeTest.Rpc.Request("sendrawtransaction", []interface{}{signedTx})
	if err != nil {
		log.Fatalf("Error broadcasting transaction: %v", err)
	}

	var txID string
	if err := json.Unmarshal(*sendResp, &txID); err != nil {
		log.Fatalf("Error parsing transaction ID: %v", err)
	}

	fmt.Printf("Transaction sent successfully! TXID: %s\n", txID)

	time.Sleep(2 * time.Second)

	blockies, err := dogeTest.ConfirmBlocks()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Blockies:", blockies)

	return nil
}
