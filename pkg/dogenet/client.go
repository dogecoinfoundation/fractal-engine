package dogenet

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/config"

	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/Dogebox-WG/gossip/dnet"
	"google.golang.org/protobuf/types/known/structpb"
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
	GossipBuyOffer(record store.BuyOffer) error
	GossipSellOffer(record store.SellOffer) error
	GossipDeleteBuyOffer(hash string, publicKey string, signature string) error
	GossipDeleteSellOffer(hash string, publicKey string, signature string) error
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
	Running  bool
}

func convertToStructPBMap(m map[string]interface{}) map[string]*structpb.Value {
	fields := make(map[string]*structpb.Value)
	for k, v := range m {
		fields[k] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: v.(string)}}
	}
	return fields
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
	if c.Running {
		statusChan <- "Running"
		log.Println("Dogenet client already running")
		return nil
	}

	c.Running = true

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
		case TagBuyOffer:
			c.recvBuyOffer(msg)
		case TagSellOffer:
			c.recvSellOffer(msg)
		case TagInvoice:
			c.recvInvoice(msg)
		case TagDeleteBuyOffer:
			c.recvDeleteBuyOffer(msg)
		case TagDeleteSellOffer:
			c.recvDeleteSellOffer(msg)
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
