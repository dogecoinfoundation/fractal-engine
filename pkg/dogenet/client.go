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
	"strings"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/Dogebox-WG/gossip/dnet"
	"google.golang.org/protobuf/proto"
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

type DogeNetClient struct {
	cfg      *config.Config
	sock     *net.Conn
	feKey    dnet.KeyPair
	Stopping bool
	Messages chan dnet.Message
}

func (c *DogeNetClient) GossipMint(record store.Mint) error {
	mintMessage := protocol.MintMessage{
		Id:              record.Id,
		Title:           record.Title,
		Description:     record.Description,
		FractionCount:   int32(record.FractionCount),
		Tags:            strings.Join(record.Tags, ","),
		TransactionHash: record.TransactionHash.String,
		// Metadata:        &structpb.Struct{Fields: record.Metadata},
		Hash: record.Hash,
		// Requirements:    &structpb.Struct{Fields: record.Requirements},
		Resellable: record.Resellable,
		// LockupOptions:   &structpb.Struct{Fields: record.LockupOptions},
		Gossiped:  record.Gossiped,
		Verified:  record.Verified,
		FeedUrl:   record.FeedURL,
		CreatedAt: record.CreatedAt.Format(time.RFC3339),
	}

	data, err := proto.Marshal(&mintMessage)
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	encodedMsg := dnet.EncodeMessageRaw(ChanFE, TagMint, c.feKey, data)

	err = encodedMsg.Send(*c.sock)
	if err != nil {
		return err
	}

	return nil
}

func NewDogeNetClient(cfg *config.Config) *DogeNetClient {
	return &DogeNetClient{
		cfg:      cfg,
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

func (c *DogeNetClient) Start(statusChan chan string) error {
	sock, err := net.Dial(c.cfg.DogeNetNetwork, c.cfg.DogeNetAddress)
	if err != nil {
		log.Printf("[FE] cannot connect: %v", err)
		return err
	}
	c.sock = &sock

	log.Printf("[FE] connected to dogenet.")
	bind := dnet.BindMessage{Version: 1, Chan: ChanFE, PubKey: *c.feKey.Pub}

	_, err = sock.Write(bind.Encode())
	if err != nil {
		log.Printf("[FE] cannot send BindMessage: %v", err)
		sock.Close()
		return err
	}

	reader := bufio.NewReader(sock)

	br_buf := [dnet.BindMessageSize]byte{}
	_, err = io.ReadAtLeast(reader, br_buf[:], len(br_buf))
	if err != nil {
		log.Printf("[FE] reading BindMessage reply: %v", err)
		sock.Close()
		return err
	}
	if _, ok := dnet.DecodeBindMessage(br_buf[:]); ok {
		// send the node's pubkey to the announce service
		// so it can include the node key in the identity announcement
		// TODO
		log.Printf("[FE] Decoded BindMessage reply.")
	} else {
		log.Printf("[FE] invalid BindMessage reply: %v", err)
		sock.Close()
		return err
	}
	log.Printf("[FE] completed handshake.")

	// go s.gossipMyIdentity(sock)
	// go s.gossipRandomIdentities(sock)

	statusChan <- "Running"

	for !c.Stopping {
		msg, err := dnet.ReadMessage(reader)
		if err != nil {
			log.Printf("[FE] cannot receive from peer: %v", err)
			sock.Close()
			return err
		}

		c.Messages <- msg

		if msg.Chan != ChanFE {
			log.Printf("[FE] ignored message: [%s][%s]", msg.Chan, msg.Tag)
			continue
		}
		switch msg.Tag {
		case TagMint:
			c.recvMint(msg)
		default:
			log.Printf("[FE] unknown message: [%s][%s]", msg.Chan, msg.Tag)
		}
	}

	return nil
}

func (c *DogeNetClient) Stop() error {
	c.Stopping = true
	return (*c.sock).Close()
}

func (c *DogeNetClient) Gossip() error {
	return nil
}

func (c *DogeNetClient) Listen(topic string, listener GossipMessageListener) error {
	return nil
}

func (c *DogeNetClient) recvMint(msg dnet.Message) {
	envelope := protocol.MessageEnvelope{}
	err := envelope.Deserialize(msg.Payload)
	if err != nil {
		log.Println("Error deserializing message envelope:", err)
		return
	}

	log.Printf("[FE] %v", msg)
}
