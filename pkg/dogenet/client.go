package dogenet

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/Dogebox-WG/gossip/dnet"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NodePubKeyMsg struct {
	PubKey []byte
}

type AddPeer struct {
	Key  string `json:"key"`
	Addr string `json:"addr"`
}

type NodeInfo struct {
	Key      string `json:"pubkey"`
	Addr     string `json:"address"`
	Identity string `json:"identity"`
}

type GetNodesResponse []NodeInfo

type GossipClient interface {
	GossipMint(record store.Mint) error
	GossipOffer(record store.Offer) error
	GossipUnconfirmedInvoice(record store.UnconfirmedInvoice) error
	GetNodes() (GetNodesResponse, error)
	AddPeer(addPeer AddPeer) error
	CheckRunning() error
	Start(statusChan chan string) error
	Stop() error
}

type DogeNetClient struct {
	GossipClient
	cfg      *config.Config
	store    *store.TokenisationStore
	sock     net.Conn
	feKey    dnet.KeyPair
	Stopping bool
	Messages chan dnet.Message
}

func convertToStructPBMap(m map[string]interface{}) map[string]*structpb.Value {
	fields := make(map[string]*structpb.Value)
	for k, v := range m {
		fields[k] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: v.(string)}}
	}
	return fields
}

func (c *DogeNetClient) GossipMint(record store.Mint) error {
	mintMessage := protocol.MintMessage{
		Id:              record.Id,
		Title:           record.Title,
		Description:     record.Description,
		FractionCount:   int32(record.FractionCount),
		Tags:            record.Tags,
		TransactionHash: record.TransactionHash.String,
		Metadata:        &structpb.Struct{Fields: convertToStructPBMap(record.Metadata)},
		Hash:            record.Hash,
		Requirements:    &structpb.Struct{Fields: convertToStructPBMap(record.Requirements)},
		LockupOptions:   &structpb.Struct{Fields: convertToStructPBMap(record.LockupOptions)},
		FeedUrl:         record.FeedURL,
		CreatedAt:       timestamppb.New(record.CreatedAt),
	}

	envelope := protocol.MintMessageEnvelope{
		Type:    protocol.ACTION_MINT,
		Version: protocol.DEFAULT_VERSION,
		Payload: &mintMessage,
	}

	data, err := proto.Marshal(&envelope)
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	encodedMsg := dnet.EncodeMessageRaw(ChanFE, TagMint, c.feKey, data)

	err = encodedMsg.Send(c.sock)
	if err != nil {
		return err
	}

	return nil
}

func (c *DogeNetClient) GossipOffer(record store.Offer) error {
	offerMessage := protocol.OfferMessage{
		Id:             record.Id,
		OffererAddress: record.OffererAddress,
		Type:           protocol.OfferType(record.Type),
		Hash:           record.Hash,
		CreatedAt:      timestamppb.New(record.CreatedAt),
		Quantity:       int32(record.Quantity),
		Price:          int32(record.Price),
	}

	envelope := protocol.OfferMessageEnvelope{
		Type:    protocol.ACTION_OFFER,
		Version: protocol.DEFAULT_VERSION,
		Payload: &offerMessage,
	}

	data, err := proto.Marshal(&envelope)
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	encodedMsg := dnet.EncodeMessageRaw(ChanFE, TagOffer, c.feKey, data)

	err = encodedMsg.Send(c.sock)
	if err != nil {
		return err
	}

	return nil
}

func (c *DogeNetClient) GossipUnconfirmedInvoice(record store.UnconfirmedInvoice) error {
	invoiceMessage := protocol.InvoiceMessage{
		Id:             record.Id,
		PaymentAddress: record.PaymentAddress,
		CreatedAt:      timestamppb.New(record.CreatedAt),
	}

	envelope := protocol.InvoiceMessageEnvelope{
		Type:    protocol.ACTION_INVOICE,
		Version: protocol.DEFAULT_VERSION,
		Payload: &invoiceMessage,
	}

	data, err := proto.Marshal(&envelope)
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	encodedMsg := dnet.EncodeMessageRaw(ChanFE, TagInvoice, c.feKey, data)

	err = encodedMsg.Send(c.sock)
	if err != nil {
		return err
	}

	return nil
}

func NewDogeNetClient(cfg *config.Config, store *store.TokenisationStore) *DogeNetClient {
	return &DogeNetClient{
		cfg:      cfg,
		store:    store,
		Stopping: false,
		feKey:    cfg.DogeNetKeyPair,
		Messages: make(chan dnet.Message),
	}
}

func (c *DogeNetClient) GetNodes() (GetNodesResponse, error) {
	resp, err := http.Get("http://" + c.cfg.DogeNetWebAddress + "/peers")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %s", err)
	}

	var nodes GetNodesResponse
	json.Unmarshal(body, &nodes)

	return nodes, nil
}

func (c *DogeNetClient) AddPeer(addPeer AddPeer) error {
	payload, err := json.Marshal(addPeer)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://"+c.cfg.DogeNetWebAddress+"/addpeer", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %s", err)
		}

		fmt.Println(string(body))

		return fmt.Errorf("failed to add peer: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %s", err)
	}

	fmt.Println(string(body))

	return nil
}

func (c *DogeNetClient) CheckRunning() error {
	resp, err := http.Get("http://" + c.cfg.DogeNetWebAddress + "/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *DogeNetClient) StartWithConn(statusChan chan string, conn net.Conn) error {
	c.sock = conn

	return c.Start(statusChan)
}

func (c *DogeNetClient) Start(statusChan chan string) error {
	if c.sock == nil {
		sock, err := net.Dial(c.cfg.DogeNetNetwork, c.cfg.DogeNetAddress)
		if err != nil {
			log.Printf("[FE] cannot connect: %v", err)
			return err
		}
		c.sock = sock
	}

	log.Printf("[FE] connected to dogenet.")
	bind := dnet.BindMessage{Version: 1, Chan: ChanFE, PubKey: *c.feKey.Pub}

	_, err := c.sock.Write(bind.Encode())
	if err != nil {
		log.Printf("[FE] cannot send BindMessage: %v", err)
		c.sock.Close()
		return err
	}

	reader := bufio.NewReader(c.sock)

	br_buf := [dnet.BindMessageSize]byte{}
	_, err = io.ReadAtLeast(reader, br_buf[:], len(br_buf))
	if err != nil {
		log.Printf("[FE] reading BindMessage reply: %v", err)
		c.sock.Close()
		return err
	}

	if _, ok := dnet.DecodeBindMessage(br_buf[:]); ok {
		// send the node's pubkey to the announce service
		// so it can include the node key in the identity announcement
		// TODO
		log.Printf("[FE] Decoded BindMessage reply.")
	} else {
		log.Printf("[FE] invalid BindMessage reply: %v", err)
		c.sock.Close()
		return err
	}
	log.Printf("[FE] completed handshake.")

	// go s.gossipMyIdentity(sock)
	// go s.gossipRandomIdentities(sock)

	if statusChan != nil {
		statusChan <- "Running"
	}

	for !c.Stopping {
		msg, err := dnet.ReadMessage(reader)
		if err != nil {
			log.Printf("[FE] cannot receive from peer: %v", err)
			c.sock.Close()
			return err
		}

		log.Printf("[FE] received message: [%s][%s]", msg.Chan, msg.Tag)

		// write to channel in a goroutine to avoid blocking
		go func() {
			c.Messages <- msg
		}()

		if msg.Chan != ChanFE {
			log.Printf("[FE] ignored message: [%s][%s]", msg.Chan, msg.Tag)
			continue
		}

		switch msg.Tag {
		case TagMint:
			c.recvMint(msg)
		case TagOffer:
			c.recvOffer(msg)
		case TagInvoice:
			c.recvInvoice(msg)
		default:
			log.Printf("[FE] unknown message: [%s][%s]", msg.Chan, msg.Tag)
		}
	}

	return nil
}

func (c *DogeNetClient) Stop() error {
	fmt.Println("Stopping dogenet client")
	c.Stopping = true

	if c.sock != nil {
		c.sock.Close()
	}

	return nil
}

func (c *DogeNetClient) recvOffer(msg dnet.Message) {
	log.Printf("[FE] received offer message")

	envelope := protocol.OfferMessageEnvelope{}
	err := proto.Unmarshal(msg.Payload, &envelope)
	if err != nil {
		log.Println("Error deserializing message envelope:", err)
		return
	}

	if envelope.Type != protocol.ACTION_OFFER {
		log.Printf("[FE] unexpected action: [%s][%s][%d]", msg.Chan, msg.Tag, envelope.Type)
		return
	}

	offer := envelope.Payload

	offerWithoutID := store.OfferWithoutID{
		OffererAddress: offer.OffererAddress,
		Type:           store.OfferType(offer.Type),
		Hash:           offer.Hash,
		Quantity:       int(offer.Quantity),
		Price:          int(offer.Price),
		CreatedAt:      offer.CreatedAt.AsTime(),
	}

	// TODO: check if the offer is valid

	id, err := c.store.SaveOffer(&offerWithoutID)
	if err != nil {
		log.Println("Error saving unconfirmed offer:", err)
		return
	}

	log.Printf("[FE] unconfirmed offer saved: %v", id)
}

func (c *DogeNetClient) recvMint(msg dnet.Message) {
	log.Printf("[FE] received mint message")

	envelope := protocol.MintMessageEnvelope{}
	err := proto.Unmarshal(msg.Payload, &envelope)
	if err != nil {
		log.Println("Error deserializing message envelope:", err)
		return
	}

	if envelope.Type != protocol.ACTION_MINT {
		log.Printf("[FE] unexpected action: [%s][%s][%d]", msg.Chan, msg.Tag, envelope.Type)
		return
	}

	mint := envelope.Payload

	id, err := c.store.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash:            mint.Hash,
		Title:           mint.Title,
		FractionCount:   int(mint.FractionCount),
		Description:     mint.Description,
		Tags:            mint.Tags,
		Metadata:        mint.Metadata.AsMap(),
		TransactionHash: sql.NullString{String: mint.TransactionHash, Valid: true},
		CreatedAt:       mint.CreatedAt.AsTime(),
		Requirements:    mint.Requirements.AsMap(),
		LockupOptions:   mint.LockupOptions.AsMap(),
	})

	if err != nil {
		log.Println("Error saving unconfirmed mint:", err)
		return
	}

	log.Printf("[FE] unconfirmed mint saved: %v", id)
}

func (c *DogeNetClient) recvInvoice(msg dnet.Message) {
	log.Printf("[FE] received invoice message")

	envelope := protocol.InvoiceMessageEnvelope{}
	err := proto.Unmarshal(msg.Payload, &envelope)
	if err != nil {
		log.Println("Error deserializing message envelope:", err)
		return
	}

	if envelope.Type != protocol.ACTION_INVOICE {
		log.Printf("[FE] unexpected action: [%s][%s][%d]", msg.Chan, msg.Tag, envelope.Type)
		return
	}

	invoice := envelope.Payload

	invoiceWithoutID := store.UnconfirmedInvoice{
		PaymentAddress:         invoice.PaymentAddress,
		BuyOfferOffererAddress: invoice.BuyOfferOffererAddress,
		BuyOfferHash:           invoice.BuyOfferHash,
		BuyOfferMintHash:       invoice.BuyOfferMintHash,
		BuyOfferQuantity:       int(invoice.BuyOfferQuantity),
		BuyOfferPrice:          int(invoice.BuyOfferPrice),
		CreatedAt:              invoice.CreatedAt.AsTime(),
		Hash:                   invoice.Hash,
		Id:                     invoice.Id,
	}

	id, err := c.store.SaveUnconfirmedInvoice(&invoiceWithoutID)
	if err != nil {
		log.Println("Error saving unconfirmed invoice:", err)
		return
	}

	log.Printf("[FE] unconfirmed invoice saved: %v", id)
}
