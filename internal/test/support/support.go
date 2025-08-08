package support

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"code.dogecoin.org/gossip/dnet"
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/service"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/docker/go-connections/nat"
	"github.com/dogecoinfoundation/dogetest/pkg/dogetest"

	bmclient "github.com/dogecoinfoundation/balance-master/pkg/client"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestDB() *store.TokenisationStore {
	// Use the shared database approach for better performance
	return SetupTestDBShared()
}

type TestGroup struct {
	InstanceId             int
	Name                   string
	LogConsumer            *StdoutLogConsumer
	NetworkName            string
	DogeTest               *dogetest.DogeTest
	AddressBook            *dogetest.AddressBook
	FeService              *service.TokenisationService
	FeConfig               *config.Config
	DogeNetClient          *dogenet.DogeNetClient
	BmClient               *bmclient.BalanceMasterClient
	NodeKey                dnet.KeyPair
	DogenetContainer       testcontainers.Container
	BalanceMasterContainer testcontainers.Container
	Running                bool
	DogeCorePort           int
	DnWebPort              int
	DnPort                 int
	DnGossipPort           int
	FeRpcPort              int
	BmPort                 int
	DbUrl                  string // Track DB URL for cleanup
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

	if tg.BalanceMasterContainer != nil {
		err := tg.BalanceMasterContainer.Stop(context.Background(), nil)
		if err != nil {
			log.Println("Error stopping bm container:", err)
		}
	}

	if tg.FeService != nil {
		tg.FeService.Stop()
	}

	// Cleanup SQLite database file
	if tg.DbUrl != "" {
		dbPath := strings.TrimPrefix(tg.DbUrl, "sqlite://")
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			log.Printf("Failed to remove test database %s: %v", dbPath, err)
		}
	}

	tg.Running = false
}

func NewTestGroup(name string, networkName string, instanceId int, dogeCorePort int, feRpcPort int, dnWebPort int, dnPort int, dnGossipPort int, bmPort int) *TestGroup {
	nodeKey, err := dnet.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	dogeTest, err := dogetest.NewDogeTest(dogetest.DogeTestConfig{
		Host:          "0.0.0.0",
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
		BmPort:       bmPort,
	}
}

func (tg *TestGroup) Start(m *testing.M) {
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

	// Use unique database file with timestamp to avoid conflicts
	tokenStore := SetupTestDB()

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

	bmClient, balanceMasterContainer := StartBalanceMasterInstance(tg.NetworkName, tg.BmPort, tg.DogeCorePort, strconv.Itoa(tg.InstanceId))

	feService, feConfig := SetupFractalEngineTestInstance(tg.InstanceId, strconv.Itoa(tg.FeRpcPort), tg.DogeCorePort, dogenetClient, tg.DogeTest, tokenStore)
	tg.FeService = feService
	tg.FeConfig = feConfig

	tg.DogeNetClient = dogenetClient
	tg.DogenetContainer = dogenetContainer
	tg.BalanceMasterContainer = balanceMasterContainer
	tg.BmClient = bmClient

	go feService.Start()

	feService.WaitForRunning()
}

func StartBalanceMasterInstance(networkName string, port int, dogePort int, instanceId string) (*bmclient.BalanceMasterClient, testcontainers.Container) {
	ctx := context.Background()

	logConsumer := &StdoutLogConsumer{Name: "balance-master " + instanceId}

	absPathContext, err := filepath.Abs(filepath.Join(".."))
	if err != nil {
		panic(err)
	}

	cacheBuster := strconv.FormatInt(time.Now().UnixNano(), 10)
	dogePortStr := strconv.Itoa(dogePort)
	dogeHostStr := "dogecoin-" + strconv.Itoa(dogePort)
	rpcServerPortStr := strconv.Itoa(port)

	balanceMasterContainerRequest := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    absPathContext,
			Dockerfile: "Dockerfile.balancemaster",
			KeepImage:  true, // Keep image to avoid rebuilding
			BuildArgs: map[string]*string{
				"CACHE_BUSTER": &cacheBuster,
			},
		},
		Networks: []string{
			networkName,
		},
		Name:         "balance-master-" + instanceId,
		ExposedPorts: []string{strconv.Itoa(port) + "/tcp"},
		Env: map[string]string{
			"DOGE_PORT":       dogePortStr,
			"DOGE_HOST":       dogeHostStr,
			"RPC_SERVER_PORT": rpcServerPortStr,
		},
		WaitingFor: wait.ForLog("Server is ready to handle requests at").WithStartupTimeout(10 * time.Second),
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

	balanceMasterContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: balanceMasterContainerRequest,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	for {
		if balanceMasterContainer.IsRunning() {
			break
		}

		time.Sleep(300 * time.Millisecond)
	}

	natPort, err := nat.NewPort("tcp", strconv.Itoa(port))
	if err != nil {
		panic(err)
	}

	mappedPort, err := balanceMasterContainer.MappedPort(ctx, natPort)
	if err != nil {
		panic(err)
	}

	ip, _ := balanceMasterContainer.Host(ctx)

	bmClient := bmclient.NewBalanceMasterClient(&bmclient.BalanceMasterClientConfig{
		RpcServerHost: ip,
		RpcServerPort: mappedPort.Port(),
	})

	return bmClient, balanceMasterContainer
}

func SetupFractalEngineTestInstance(instanceId int, rpcServerPort string, dogePort int, dogenetClient *dogenet.DogeNetClient, dogeTestInstance *dogetest.DogeTest, tokenStore *store.TokenisationStore) (*service.TokenisationService, *config.Config) {
	// Remove old test files - no longer needed as we use unique names

	mappedPort, err := dogeTestInstance.Container.MappedPort(context.Background(), nat.Port(strconv.Itoa(dogePort)+"/tcp"))
	if err != nil {
		log.Fatal(err)
	}

	feConfig := config.NewConfig()
	feConfig.DogeHost = dogeTestInstance.Host
	// Use the same tokenStore that was passed in, don't create a new DB URL
	// feConfig.DatabaseURL is not needed since we're using the passed tokenStore
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

func ConnectDogeTestPeers(nodeA *dogetest.DogeTest, nodeB *dogetest.DogeTest) error {
	ctx := context.Background()

	// Get node B's container IP (within Docker network)
	nodeBIP, err := nodeB.Container.ContainerIP(ctx)
	if err != nil {
		return fmt.Errorf("failed to get node B IP: %w", err)
	}

	// Use the default regnet P2P port (18444) - internal container port
	nodeBAddress := fmt.Sprintf("%s:18444", nodeBIP)

	log.Printf("Connecting DogeTest peers: Node A -> Node B (%s)", nodeBAddress)

	// Use addnode RPC to connect the nodes
	resp, err := nodeA.Rpc.Request("addnode", []interface{}{nodeBAddress, "add"})
	if err != nil {
		// Check if it's just an empty response (which is normal for addnode)
		if err.Error() != "json-rpc no result or error was returned" {
			return fmt.Errorf("failed to add node B as peer: %w", err)
		}
		log.Printf("AddNode command executed (empty response is normal)")
	} else if resp != nil {
		log.Printf("AddNode response: %s", string(*resp))
	}

	// Wait a moment for the connection to establish
	time.Sleep(3 * time.Second)

	// Verify the connection by checking peer count
	peersResp, err := nodeA.Rpc.Request("getconnectioncount", []interface{}{})
	if err != nil {
		log.Printf("Warning: Could not verify peer connection: %v", err)
	} else {
		var peerCount int
		if err := json.Unmarshal(*peersResp, &peerCount); err == nil {
			log.Printf("Node A peer count: %d", peerCount)
		}
	}

	log.Println("DogeTest peer connection established")
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

	log.Println(content)
}

func StartDogenetInstance(ctx context.Context, feKey dnet.KeyPair, image string, instanceId string, webPort string, port string, gossipPort string, networkName string, logConsumer *StdoutLogConsumer, tokenisationStore *store.TokenisationStore) (*dogenet.DogeNetClient, testcontainers.Container, error) {
	nodeKey, err := dnet.GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}

	absPathContext, err := filepath.Abs(filepath.Join(".."))
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

	envelope := protocol.NewMintTransactionEnvelope(mintResponse.Hash, protocol.ACTION_MINT)
	encodedTransactionBody := envelope.Serialize()

	outputs := map[string]interface{}{
		addressBook.Addresses[0].Address: change,
		"data":                           hex.EncodeToString(encodedTransactionBody),
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

// GenerateDogecoinAddress generates a string that matches the Dogecoin address format
// This creates a valid-looking address for testing purposes (not a real spendable address)
// Dogecoin addresses start with 'D' for mainnet or 'n' for testnet
func GenerateDogecoinAddress(testnet bool) string {
	// Base58 alphabet used in Bitcoin/Dogecoin addresses
	base58Alphabet := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	// Address length is typically 34 characters
	addressLength := 34

	// Start with appropriate prefix
	var address strings.Builder
	if testnet {
		address.WriteString("n")
	} else {
		address.WriteString("D")
	}

	// Generate random characters from base58 alphabet
	for i := 1; i < addressLength; i++ {
		randomBytes := make([]byte, 1)
		_, err := rand.Read(randomBytes)
		if err != nil {
			// Fallback to a deterministic character if random fails
			address.WriteByte(base58Alphabet[i%len(base58Alphabet)])
			continue
		}
		// Use the random byte to pick a character from the alphabet
		index := int(randomBytes[0]) % len(base58Alphabet)
		address.WriteByte(base58Alphabet[index])
	}

	return address.String()
}

// GenerateRandomHash generates a random 64-character hexadecimal hash
// This is useful for testing purposes where a hash-like string is needed
func GenerateRandomHash() string {
	// 32 bytes will give us 64 hex characters
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback to a deterministic hash if random fails
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}
	return hex.EncodeToString(randomBytes)
}
