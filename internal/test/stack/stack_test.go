package stack

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	feclient "dogecoin.org/fractal-engine/pkg/client"
	fecfg "dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/rpc"
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

	TrackAddress(&stackConfig)
	TopUp(&stackConfig)
	ConfirmBlocks(&stackConfig)

	return stackConfig
}

func TestSimpleFlow(t *testing.T) {
	stacks := makeStackConfigsAndPeer(2)

	seller := stacks[0]
	buyer := stacks[1]
	mintQty := 100
	sellQty := 20

	mintHash := Mint(seller)
	AssertEqualWithRetry(t, func() interface{} {
		return GetTokenBalance(seller, mintHash)
	}, mintQty, 10, 3*time.Second)
	fmt.Println("Mint confirmed")

	invoiceHash := Invoice(seller, buyer.Address, mintHash, sellQty, 20)
	AssertEqualWithRetry(t, func() interface{} {
		return GetPendingTokenBalance(seller, mintHash)
	}, sellQty, 10, 3*time.Second)
	fmt.Println("Invoice confirmed")

	paymentTrxn := Payment(buyer, seller, invoiceHash, sellQty, 20)
	AssertEqualWithRetry(t, func() interface{} {
		return GetTokenBalance(seller, mintHash)
	}, sellQty, 10, 3*time.Second)
	fmt.Println("Payment confirmed")

	log.Println("Mint: ", mintHash)
	log.Println("Invoice: ", invoiceHash)
	log.Println("Payment Trxn: ", paymentTrxn)

	res, err := buyer.DogeClient.Request("getrawtransaction", []interface{}{paymentTrxn, 2})
	if err != nil {
		panic(err)
	}

	var result doge.RawTxn
	err = json.Unmarshal(*res, &result)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}

func WriteToBlockchain(stackConfig *StackConfig, paymentAddress string, hexBody string, amount float64) string {
	blockChainInfo, err := stackConfig.DogeClient.GetBlockchainInfo()
	if err != nil {
		panic(err)
	}

	chainByte, err := doge.GetPrefix(blockChainInfo.Chain)
	if err != nil {
		log.Fatal(err)
	}
	chainCfg := doge.GetChainCfg(chainByte)

	utxos, err := stackConfig.BalanceMasterClient.GetUtxos(stackConfig.Address)
	if err != nil {
		panic(err)
	}

	if len(utxos) == 0 {
		panic(errors.New("No utxos found"))
	}

	inputs := []interface{}{
		map[string]interface{}{
			"txid": utxos[0].TxID,
			"vout": utxos[0].VOut,
		},
	}

	address := stackConfig.Address

	var outputs map[string]interface{}
	if paymentAddress == address {
		outputs = map[string]interface{}{
			"data":  hexBody,
			address: utxos[0].Amount - amount,
		}
	} else {
		outputs = map[string]interface{}{
			"data":         hexBody,
			paymentAddress: amount,
			address:        utxos[0].Amount - amount,
		}
	}

	rawTx, err := stackConfig.DogeClient.Request("createrawtransaction", []interface{}{inputs, outputs})
	if err != nil {
		panic(err)
	}

	var rawTxResponse string
	if err := json.Unmarshal(*rawTx, &rawTxResponse); err != nil {
		panic(err)
	}

	encodedTx, err := doge.SignRawTransaction(rawTxResponse, stackConfig.PrivKey, []doge.PrevOutput{
		{
			Address: address,
			Amount:  int64(utxos[0].Amount),
		},
	}, chainCfg)

	if err != nil {
		panic(err)
	}

	res, err := stackConfig.DogeClient.Request("sendrawtransaction", []interface{}{encodedTx})
	if err != nil {
		panic(err)
	}

	var txid string

	if err := json.Unmarshal(*res, &txid); err != nil {
		panic(err)
	}

	return txid
}

func GetTokenBalance(stackConfig *StackConfig, mintHash string) int {
	tokens, err := stackConfig.TokenisationClient.GetTokenBalance(stackConfig.Address, mintHash)
	if err != nil {
		panic(err)
	}

	balance := 0
	for _, token := range tokens {
		balance += token.Quantity
	}

	return balance
}

func GetPendingTokenBalance(stackConfig *StackConfig, mintHash string) int {
	tokens, err := stackConfig.TokenisationClient.GetPendingTokenBalance(stackConfig.Address, mintHash)
	if err != nil {
		panic(err)
	}

	balance := 0
	for _, token := range tokens {
		balance += token.Quantity
	}

	return balance
}

func ConfirmBlocks(stackConfig *StackConfig) {
	_, err := stackConfig.DogeClient.Request("generate", []interface{}{10})
	if err != nil {
		panic(err)
	}
}

func TopUp(stackConfig *StackConfig) {
	ctx := context.Background()
	err := stackConfig.TokenisationClient.TopUpBalance(ctx, stackConfig.Address)
	if err != nil {
		panic(err)
	}

	fmt.Println("Topped up address " + stackConfig.Address)
}

func TrackAddress(stackConfig *StackConfig) {
	err := stackConfig.BalanceMasterClient.TrackAddress(stackConfig.Address)
	if err != nil {
		panic(err)
	}

	fmt.Println("Tracking address " + stackConfig.Address)
}

func Payment(buyerConfig *StackConfig, sellerConfig *StackConfig, invoiceHash string, quantity int, price int) string {
	envelope := protocol.NewPaymentTransactionEnvelope(invoiceHash, protocol.ACTION_PAYMENT)
	encodedTransactionBody := envelope.Serialize()

	total := float64(quantity*price + 1)

	txId := WriteToBlockchain(buyerConfig, sellerConfig.Address, hex.EncodeToString(encodedTransactionBody), total)
	ConfirmBlocks(buyerConfig)

	return txId
}

func Invoice(stackConfig *StackConfig, buyerAddress string, mintHash string, quantity int, price int) string {
	invoicePayload := rpc.CreateInvoiceRequestPayload{
		PaymentAddress: stackConfig.Address,
		BuyerAddress:   buyerAddress,
		MintHash:       mintHash,
		Quantity:       quantity,
		Price:          price,
		SellerAddress:  stackConfig.Address,
	}
	mintPayloadBytes, err := json.Marshal(invoicePayload)
	if err != nil {
		panic(err)
	}

	invoiceRequest := rpc.CreateInvoiceRequest{
		Payload: invoicePayload,
	}

	signature, err := doge.SignPayload(mintPayloadBytes, stackConfig.PrivKey)
	if err != nil {
		panic(err)
	}

	invoiceRequest.Signature = signature

	res, err := stackConfig.TokenisationClient.CreateInvoice(&invoiceRequest)
	if err != nil {
		panic(err)
	}

	envelope := protocol.NewInvoiceTransactionEnvelope(res.Hash, stackConfig.Address, mintHash, int32(quantity), protocol.ACTION_INVOICE)
	encodedTransactionBody := envelope.Serialize()

	// just network fees
	WriteToBlockchain(stackConfig, stackConfig.Address, hex.EncodeToString(encodedTransactionBody), float64(1))

	ConfirmBlocks(stackConfig)

	return res.Hash
}

func Mint(stackConfig *StackConfig) string {
	mintPayload := rpc.CreateMintRequestPayload{
		Title:         "Super Lambo",
		FractionCount: 100,
		Description:   "Fast Car",
		ContractOfSale: store.StringInterfaceMap{
			"specifications": map[string]interface{}{
				"model": "Ferrari",
			},
		},
	}
	mintPayloadBytes, err := json.Marshal(mintPayload)
	if err != nil {
		panic(err)
	}

	mintRequest := rpc.CreateMintRequest{
		Payload:   mintPayload,
		Address:   stackConfig.Address,
		PublicKey: stackConfig.PubKey,
	}

	signature, err := doge.SignPayload(mintPayloadBytes, stackConfig.PrivKey)
	if err != nil {
		panic(err)
	}

	mintRequest.Signature = signature

	res, err := stackConfig.TokenisationClient.Mint(&mintRequest)
	if err != nil {
		panic(err)
	}

	envelope := protocol.NewMintTransactionEnvelope(res.Hash, protocol.ACTION_MINT)
	encodedTransactionBody := envelope.Serialize()

	// Only need 1 to cover network fees
	WriteToBlockchain(stackConfig, stackConfig.Address, hex.EncodeToString(encodedTransactionBody), 1)

	ConfirmBlocks(stackConfig)

	return res.Hash
}

func makeStackConfigsAndPeer(stackCount int) []*StackConfig {
	var stacks []*StackConfig
	for i := 0; i < stackCount; i++ {
		newConfig := NewStackConfig(i+1, "regtest")
		stacks = append(stacks, &newConfig)
	}

	for i := 0; i < len(stacks)/2; i += 2 {
		stackA := stacks[i]
		stackB := stacks[(i + 1)]

		// Check for nodes, if doesnt exist, then add peer.
		err := stackA.DogeNetClient.AddPeer(dogenet.AddPeer{
			Key:  stackB.DogeNetPubKey,
			Addr: stackB.DogeNetHost + ":" + strconv.Itoa(stackB.DogeNetBindPort),
		})
		if err != nil {
			panic(err)
		}

		stackA.DogeClient.AddPeer(stackB.DogeHost)
	}

	return stacks
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
