package e2e

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"testing"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"github.com/dogecoinfoundation/dogetest/pkg/dogetest"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go/network"
)

func TestWallet(t *testing.T) {
	ctx := context.Background()

	net, err := network.New(ctx, network.WithDriver("bridge"))
	if err != nil {
		panic(err)
	}

	networkName := net.Name

	dogeTest, err := dogetest.NewDogeTest(dogetest.DogeTestConfig{
		NetworkName: networkName,
		Port:        22557,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = dogeTest.Start()
	if err != nil {
		t.Fatal(err)
	}

	addressBook, err := dogeTest.SetupAddresses([]dogetest.AddressSetup{
		{
			Label:          "test",
			InitialBalance: 10000,
		},
	})

	address := addressBook.Addresses[0]

	mappedPort, err := dogeTest.Container.MappedPort(ctx, "22557/tcp")
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

	inputs := []interface{}{}

	envelope := protocol.NewMintTransactionEnvelope(uuid.New().String(), protocol.ACTION_MINT)
	encodedTransactionBody := envelope.Serialize()

	outputs := map[string]interface{}{
		"data": hex.EncodeToString(encodedTransactionBody),
	}

	res, err := rpcClient.Request("createrawtransaction", []interface{}{inputs, outputs})
	if err != nil {
		t.Fatal(err)
	}

	var rawTx string

	if err := json.Unmarshal(*res, &rawTx); err != nil {
		log.Fatalf("Error parsing raw transaction: %v", err)
	}

	var fundRawTransactionResponse doge.FundRawTransactionResponse

	res, err = rpcClient.Request("fundrawtransaction", []interface{}{rawTx})
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(*res, &fundRawTransactionResponse); err != nil {
		log.Fatalf("Error parsing fund raw transaction response: %v", err)
	}

	privKey, err := rpcClient.DumpPrivKey(address.Address)
	if err != nil {
		t.Fatal(err)
	}

	res, err = rpcClient.Request("signrawtransaction", []interface{}{fundRawTransactionResponse.Hex, []interface{}{}, []interface{}{
		privKey,
	}})
	if err != nil {
		t.Fatal(err)
	}

	var signRawTransactionResponse doge.SignRawTransactionResponse

	if err := json.Unmarshal(*res, &signRawTransactionResponse); err != nil {
		log.Fatalf("Error parsing sign raw transaction response: %v", err)
	}

	log.Println(signRawTransactionResponse)

	res, err = rpcClient.Request("sendrawtransaction", []interface{}{signRawTransactionResponse.Hex})
	if err != nil {
		t.Fatal(err)
	}

	var txid string

	if err := json.Unmarshal(*res, &txid); err != nil {
		log.Fatalf("Error parsing send raw transaction response: %v", err)
	}

	log.Println(txid)
}
