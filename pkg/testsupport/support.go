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

type TestDogeConfig struct {
	DogecoindPath string `toml:"dogecoind_path"`
}

type TestConfig struct {
	Doge TestDogeConfig
}

type TestGroup struct {
	InstanceId       int
	Name             string
	LogConsumer      *StdoutLogConsumer
	WebPort          int
	NetworkName      string
	Port             int
	GossipPort       int
	DogeTest         *dogetest.DogeTest
	AddressBook      *dogetest.AddressBook
	FeService        *service.TokenisationService
	FeConfig         *config.Config
	DogeNetClient    *dogenet.DogeNetClient
	NodeKey          dnet.KeyPair
	DogenetContainer testcontainers.Container
}

func (tg *TestGroup) Stop() {
	tg.DogeTest.Stop()
	tg.FeService.Stop()
	tg.DogenetContainer.Stop(context.Background(), nil)
	os.Remove("test" + strconv.Itoa(tg.InstanceId) + ".db")
}

func NewTestGroup(name string, networkName string, instanceId int, installationPath string, webPort int, port int, gossipPort int) *TestGroup {
	nodeKey, err := dnet.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	testConfig := dogetest.DogeTestConfig{
		Host:             "localhost",
		InstallationPath: installationPath,
	}

	dogeTest, err := dogetest.NewDogeTest(testConfig)
	if err != nil {
		panic(err)
	}

	return &TestGroup{
		InstanceId:  instanceId,
		Name:        name,
		NetworkName: networkName,
		DogeTest:    dogeTest,
		NodeKey:     nodeKey,
		WebPort:     webPort,
		Port:        port,
		GossipPort:  gossipPort,
	}
}

func (tg *TestGroup) Start() {
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

	feService, feConfig := SetupFractalEngineTestInstance(tg.InstanceId, strconv.Itoa(tg.WebPort), tg.DogeTest)

	tg.FeService = feService
	tg.FeConfig = feConfig

	logConsumer := &StdoutLogConsumer{Name: tg.Name}
	tg.LogConsumer = logConsumer

	dogenetClient, dogenetContainer, err := StartDogenetInstance(ctx, "Dockerfile.dogenet", strconv.Itoa(tg.InstanceId), strconv.Itoa(tg.WebPort), strconv.Itoa(tg.Port), strconv.Itoa(tg.GossipPort), tg.NetworkName, logConsumer, feService.Store)
	if err != nil {
		panic(err)
	}

	tg.DogeNetClient = dogenetClient
	tg.DogenetContainer = dogenetContainer
}

func SetupDogeTestInstance(testConfig TestConfig) (*dogetest.DogeTest, *dogetest.AddressBook, []string) {
	localDogeTest, err := dogetest.NewDogeTest(dogetest.DogeTestConfig{
		Host:             "localhost",
		InstallationPath: testConfig.Doge.DogecoindPath,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = localDogeTest.Start()
	if err != nil {
		log.Fatal(err)
	}

	addressBook, err := localDogeTest.SetupAddresses([]dogetest.AddressSetup{
		{
			Label:          "test1",
			InitialBalance: 100,
		},
		{
			Label:          "test2",
			InitialBalance: 20,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	blocks, err := localDogeTest.ConfirmBlocks()
	if err != nil {
		log.Fatal(err)
	}

	return localDogeTest, addressBook, blocks
}

func SetupFractalEngineTestInstance(instanceId int, rpcServerPort string, dogeTestInstance *dogetest.DogeTest) (*service.TokenisationService, *config.Config) {
	os.Remove("test" + strconv.Itoa(instanceId) + ".db")

	feConfig := config.NewConfig()
	feConfig.DogeHost = dogeTestInstance.Host
	feConfig.DatabaseURL = "sqlite://test" + strconv.Itoa(instanceId) + ".db"
	feConfig.DogePort = strconv.Itoa(dogeTestInstance.Port)
	feConfig.DogeUser = "test"
	feConfig.DogePassword = "test"
	feConfig.RpcServerPort = rpcServerPort
	// feConfig.PersistFollower = false
	feConfig.MigrationsPath = "../../db/migrations"

	feService := service.NewTokenisationService(feConfig)
	go feService.Start()

	feService.WaitForRunning()

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
}

func StartDogenetInstance(ctx context.Context, image string, instanceId string, webPort string, port string, gossipPort string, networkName string, logConsumer *StdoutLogConsumer, tokenisationStore *store.TokenisationStore) (*dogenet.DogeNetClient, testcontainers.Container, error) {
	nodeKey, err := dnet.GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}

	identKey, err := dnet.GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}

	absPathContext, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return nil, nil, err
	}

	cacheBuster := strconv.FormatInt(time.Now().UnixNano(), 10)

	log.Printf("KEY: %s", hex.EncodeToString(nodeKey.Priv[:]))
	log.Printf("IDENT_KEY: %s", hex.EncodeToString(identKey.Pub[:]))

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
			"IDENT_KEY":    hex.EncodeToString(identKey.Pub[:]),
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

func StartDogecoinInstance(ctx context.Context, image string, networkName string, instanceId string, port string) (testcontainers.Container, error) {
	absPathContext, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return nil, err
	}

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    absPathContext,
			Dockerfile: image,
			KeepImage:  false,
			BuildArgs: map[string]*string{
				"PORT": &port,
			},
		},
		Networks: []string{
			networkName,
		},
		Name:         "dogecoin-" + instanceId,
		ExposedPorts: []string{port + "/tcp"},
		Env: map[string]string{
			"PORT": port,
		},
		WaitingFor: wait.ForLog("init message: Done loading").WithStartupTimeout(10 * time.Second),
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericTmpfsMountSource{},
				Target: "/dogecoin",
			},
		},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Opts: []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
		},
	}

	dogecoinContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	for {
		if dogecoinContainer.IsRunning() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	for {
		handlerPort, err := dogecoinContainer.MappedPort(ctx, nat.Port(port+"/tcp"))
		if err == nil {
			fmt.Printf("Handler port mapped to %s\n", handlerPort.Port())
			break
		}

		fmt.Printf("Waiting for handler port to be mapped... %s\n", err)

		time.Sleep(1 * time.Second)
	}

	ip, _ := dogecoinContainer.Host(ctx)
	mappedPort, _ := dogecoinContainer.MappedPort(ctx, nat.Port(port+"/tcp"))

	fmt.Printf("Dogecoin is running at %s:%s\n", ip, mappedPort.Port())

	return dogecoinContainer, nil
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
