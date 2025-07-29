package e2e

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"code.dogecoin.org/gossip/dnet"
	"dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"
)

var dogenetClientA *dogenet.DogeNetClient
var dogenetClientB *dogenet.DogeNetClient
var dogenetA testcontainers.Container
var dogenetB testcontainers.Container
var tokenisationStoreA *store.TokenisationStore
var tokenisationStoreB *store.TokenisationStore

func TestMain(m *testing.M) {
	ctx := context.Background()

	net, err := network.New(ctx, network.WithDriver("bridge"))
	if err != nil {
		panic(err)
	}
	networkName := net.Name

	tokenisationStoreA = support.SetupTestDB()
	tokenisationStoreB = support.SetupTestDB()

	feKey, err := dnet.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	logConsumerA := &support.StdoutLogConsumer{Name: "alpha"}
	dogenetClientA, dogenetA, err = support.StartDogenetInstance(ctx, feKey, "Dockerfile.dogenet", "alpha", "8085", "44069", "33069", networkName, logConsumerA, tokenisationStoreA)
	if err != nil {
		panic(err)
	}

	logConsumerB := &support.StdoutLogConsumer{Name: "beta"}
	dogenetClientB, dogenetB, err = support.StartDogenetInstance(ctx, feKey, "Dockerfile.dogenet", "beta", "8086", "44070", "33070", networkName, logConsumerB, tokenisationStoreB)
	if err != nil {
		panic(err)
	}

	err = support.ConnectDogeNetPeers(dogenetClientA, dogenetB, 33070, logConsumerA, logConsumerB)
	if err != nil {
		panic(err)
	}

	time.Sleep(15 * time.Second)

	m.Run()

	fmt.Println("Cleaning up resources...")

	net.Remove(ctx)
	dogenetA.Terminate(ctx)
	dogenetB.Terminate(ctx)
}

func TestMintMessage(t *testing.T) {
	err := dogenetClientA.GossipMint(store.Mint{
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
	} else {
		fmt.Println("Mint gossiped successfully")
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

func TestOffersMessage(t *testing.T) {
	err := dogenetClientA.GossipMint(store.Mint{
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
	} else {
		fmt.Println("Mint gossiped successfully")
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

	privHex, pubHex, address, err := doge.GenerateDogecoinKeypair(doge.PrefixRegtest)
	if err != nil {
		assert.Error(t, err, "failed to generate dogecoin keypair")
	}

	sellOfferPayload := protocol.SellOfferPayload{
		MintHash:       "1",
		OffererAddress: address,
		Quantity:       100,
		Price:          100,
	}

	payloadBytes, err := protojson.Marshal(&sellOfferPayload)
	if err != nil {
		assert.Error(t, err, "failed to marshal payload")
	}

	signature, err := doge.SignPayload(payloadBytes, privHex)
	if err != nil {
		assert.Error(t, err, "failed to sign payload")
	}

	err = dogenetClientA.GossipSellOffer(store.SellOffer{
		SellOfferWithoutID: store.SellOfferWithoutID{
			MintHash:       sellOfferPayload.MintHash,
			OffererAddress: sellOfferPayload.OffererAddress,
			Quantity:       int(sellOfferPayload.Quantity),
			Price:          int(sellOfferPayload.Price),
			CreatedAt:      time.Now(),
			PublicKey:      pubHex,
			Signature:      signature,
		},
	})

	if err != nil {
		assert.Error(t, err, "failed to gossip sell offer")
	}

	msg = <-dogenetClientB.Messages

	payload = protocol.MessageEnvelope{}
	err = payload.Deserialize(msg.Payload)
	if err != nil {
		panic(err)
	}

	switch payload.Action {
	case protocol.ACTION_SELL_OFFER:
		sellOffer := protocol.SellOfferMessage{}
		err = proto.Unmarshal(payload.Data, &sellOffer)
		if err != nil {
			assert.Error(t, err, "expected sell offer message")
		}

		assert.Equal(t, sellOffer.Payload.MintHash, "1")
		assert.Equal(t, sellOffer.Payload.OffererAddress, address)
		assert.Equal(t, sellOffer.Payload.Quantity, int32(100))
		assert.Equal(t, sellOffer.Payload.Price, int32(100))
	default:
		assert.Error(t, fmt.Errorf("expected sell offer message"), "expected sell offer message")
	}

	time.Sleep(2 * time.Second)

	offers, err := tokenisationStoreB.GetSellOffers(0, 10, "1", address)
	if err != nil {
		assert.Error(t, err, "failed to get sell offers")
	}

	assert.Equal(t, len(offers), 1)
	assert.Equal(t, offers[0].MintHash, "1")
	assert.Equal(t, offers[0].OffererAddress, address)
	assert.Equal(t, offers[0].Quantity, 100)
	assert.Equal(t, offers[0].Price, 100)
	assert.Equal(t, offers[0].PublicKey, pubHex)

	buyOfferPayload := protocol.BuyOfferPayload{
		MintHash:       "MyMintHashOffers123",
		OffererAddress: address,
		SellerAddress:  "selleraddyInvoice",
		Quantity:       100,
		Price:          100,
	}

	log.Println("buyOfferPayload", buyOfferPayload)

	payloadBytes2, err := protojson.Marshal(&buyOfferPayload)
	if err != nil {
		assert.Error(t, err, "failed to marshal payload")
	}

	signature2, err := doge.SignPayload(payloadBytes2, privHex)
	if err != nil {
		assert.Error(t, err, "failed to sign payload")
	}

	err = dogenetClientA.GossipBuyOffer(store.BuyOffer{
		BuyOfferWithoutID: store.BuyOfferWithoutID{
			MintHash:       buyOfferPayload.MintHash,
			OffererAddress: buyOfferPayload.OffererAddress,
			SellerAddress:  buyOfferPayload.SellerAddress,
			Quantity:       int(buyOfferPayload.Quantity),
			Price:          int(buyOfferPayload.Price),
			CreatedAt:      time.Now(),
			PublicKey:      pubHex,
			Signature:      signature2,
		},
	})

	if err != nil {
		assert.Error(t, err, "failed to gossip buy offer")
	}

	msg = <-dogenetClientB.Messages

	payload = protocol.MessageEnvelope{}
	err = payload.Deserialize(msg.Payload)
	if err != nil {
		panic(err)
	}

	switch payload.Action {
	case protocol.ACTION_BUY_OFFER:
		buyOffer := protocol.BuyOfferMessage{}
		err = proto.Unmarshal(payload.Data, &buyOffer)
		if err != nil {
			assert.Error(t, err, "expected buy offer message")
		}

		assert.Equal(t, buyOffer.Payload.MintHash, "MyMintHashOffers123")
		assert.Equal(t, buyOffer.Payload.OffererAddress, address)
		assert.Equal(t, buyOffer.Payload.SellerAddress, "selleraddy")
		assert.Equal(t, buyOffer.Payload.Quantity, int32(100))
		assert.Equal(t, buyOffer.Payload.Price, int32(100))
	default:
		assert.Error(t, fmt.Errorf("expected buy offer message"), "expected buy offer message")
	}

	time.Sleep(2 * time.Second)

	offers2, err := tokenisationStoreB.GetBuyOffersByMintAndSellerAddress(0, 10, "MyMintHashOffers123", "selleraddyInvoice")
	if err != nil {
		assert.Error(t, err, "failed to get buy offers")
	}

	assert.Equal(t, len(offers2), 1)
	assert.Equal(t, offers2[0].MintHash, "MyMintHashOffers123")
	assert.Equal(t, offers2[0].OffererAddress, address)
	assert.Equal(t, offers2[0].SellerAddress, "selleraddyInvoice")
	assert.Equal(t, offers2[0].Quantity, 100)
	assert.Equal(t, offers2[0].Price, 100)
	assert.Equal(t, offers2[0].PublicKey, pubHex)
}

func TestInvoiceMessage(t *testing.T) {
	err := dogenetClientA.GossipMint(store.Mint{
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
	} else {
		fmt.Println("Mint gossiped successfully")
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

	privHex, pubHex, address, err := doge.GenerateDogecoinKeypair(doge.PrefixRegtest)
	if err != nil {
		assert.Error(t, err, "failed to generate dogecoin keypair")
	}

	invoicePayload := protocol.InvoicePayload{
		PaymentAddress:         "paymentaddyzz",
		BuyOfferOffererAddress: "buyofferoffereraddress",
		BuyOfferHash:           "buyofferhash",
		BuyOfferMintHash:       "buyofferminthash",
		BuyOfferQuantity:       100,
		BuyOfferPrice:          100,
		SellOfferAddress:       address,
	}

	payloadBytes, err := protojson.Marshal(&invoicePayload)
	if err != nil {
		assert.Error(t, err, "failed to marshal payload")
	}

	signature, err := doge.SignPayload(payloadBytes, privHex)
	if err != nil {
		assert.Error(t, err, "failed to sign payload")
	}

	err = dogenetClientA.GossipUnconfirmedInvoice(store.UnconfirmedInvoice{
		PaymentAddress:         invoicePayload.PaymentAddress,
		BuyOfferOffererAddress: invoicePayload.BuyOfferOffererAddress,
		BuyOfferHash:           invoicePayload.BuyOfferHash,
		BuyOfferMintHash:       invoicePayload.BuyOfferMintHash,
		BuyOfferQuantity:       int(invoicePayload.BuyOfferQuantity),
		BuyOfferPrice:          int(invoicePayload.BuyOfferPrice),
		SellOfferAddress:       invoicePayload.SellOfferAddress,
		CreatedAt:              time.Now(),
		PublicKey:              pubHex,
		Signature:              signature,
	})

	if err != nil {
		assert.Error(t, err, "failed to gossip invoice")
	}

	msg = <-dogenetClientB.Messages

	payload = protocol.MessageEnvelope{}

	err = payload.Deserialize(msg.Payload)
	if err != nil {
		panic(err)
	}

	switch payload.Action {
	case protocol.ACTION_INVOICE:
		invoice := protocol.InvoiceMessage{}
		err = proto.Unmarshal(payload.Data, &invoice)
		if err != nil {
			assert.Error(t, err, "expected invoice message")
		}

		assert.Equal(t, invoice.Payload.PaymentAddress, "paymentaddyzz")
		assert.Equal(t, invoice.Payload.BuyOfferOffererAddress, "buyofferoffereraddress")
		assert.Equal(t, invoice.Payload.BuyOfferHash, "buyofferhash")
		assert.Equal(t, invoice.Payload.BuyOfferMintHash, "buyofferminthash")
		assert.Equal(t, invoice.Payload.BuyOfferQuantity, int32(100))
		assert.Equal(t, invoice.Payload.BuyOfferPrice, int32(100))
		assert.Equal(t, invoice.Payload.SellOfferAddress, address)
	default:
		assert.Error(t, fmt.Errorf("expected invoice message"), "expected invoice message")
	}

	time.Sleep(2 * time.Second)

}
