package dogenet

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/Dogebox-WG/gossip/dnet"
)

type NodePubKeyMsg struct {
	PubKey []byte
}

type AddPeer struct {
	Key  string `json:"key"`
	Addr string `json:"addr"`
}

type DogeNetClient struct {
	cfg             *config.Config
	sock            *net.Conn
	feKey           dnet.KeyPair
	announceChanges chan NodePubKeyMsg
	Stopping        bool
}

func (c *DogeNetClient) GossipMint(record store.Mint) error {
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}

	encodedMsg := dnet.EncodeMessageRaw(ChanFE, TagMint, c.feKey, payload)

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
	}
}

func (c *DogeNetClient) AddPeer(addPeer AddPeer) error {
	payload, err := json.Marshal(addPeer)
	if err != nil {
		return err
	}

	resp, err := http.Post("https://httpbin.org/post", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if err != nil {
		return err
	}

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
	c.sock = &sock

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

	fmt.Println(envelope)
}
