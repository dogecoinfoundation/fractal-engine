package dogenet

import (
	"bufio"
	"io"
	"log"
	"net"

	"dogecoin.org/fractal-engine/pkg/config"
	"github.com/Dogebox-WG/gossip/dnet"
)

var ChanFE = dnet.NewTag("FractalEngine")
var TagMint = dnet.NewTag("Mint")

type NodePubKeyMsg struct {
	PubKey []byte
}

type DogeNetClient struct {
	cfg             *config.Config
	sock            net.Conn
	idenKey         dnet.KeyPair
	announceChanges chan NodePubKeyMsg
	Stopping        bool
}

func NewDogeNetClient(cfg *config.Config) *DogeNetClient {
	return &DogeNetClient{
		cfg:      cfg,
		Stopping: false,
	}
}

func (c *DogeNetClient) Start() error {
	sock, err := net.Dial(c.cfg.DogeNetNetwork, c.cfg.DogeNetAddress)
	if err != nil {
		log.Printf("[FE] cannot connect: %v", err)
		return err
	}
	c.sock = sock

	log.Printf("[FE] connected to dogenet.")
	bind := dnet.BindMessage{Version: 1, Chan: ChanFE, PubKey: *c.idenKey.Pub}
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
	if br, ok := dnet.DecodeBindMessage(br_buf[:]); ok {
		// send the node's pubkey to the announce service
		// so it can include the node key in the identity announcement
		c.announceChanges <- NodePubKeyMsg{PubKey: br.PubKey[:]}
	} else {
		log.Printf("[FE] invalid BindMessage reply: %v", err)
		sock.Close()
		return err
	}

	c.sock = sock // for Stop()
	// go s.gossipMyIdentity(sock)
	// go s.gossipRandomIdentities(sock)
	// read messages until reading fails
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
	return c.sock.Close()
}

func (c *DogeNetClient) Gossip() error {
	return nil
}

func (c *DogeNetClient) Listen(topic string, listener GossipMessageListener) error {
	return nil
}

func (c *DogeNetClient) recvMint(msg dnet.Message) {
	// id := iden.DecodeIdentityMsg(msg.Payload)
	// days := (id.Time.Local().Unix() - time.Now().Unix()) / OneUnixDay
	// log.Printf("[Iden] received identity: %v %v %v %v %v signed by: %v (%v days remain)", id.Name, id.Country, id.City, id.Lat, id.Long, hex.EncodeToString(msg.PubKey), days)
	// s.store.SetIdentity(msg.PubKey, msg.Payload, msg.Signature, id.Time.Local().Unix())
}
