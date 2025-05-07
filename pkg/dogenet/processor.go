package dogenet

import (
	"bufio"
	"context"
	"encoding/hex"
	"io"
	"log"
	"net"
	"net/url"
	"os"

	"code.dogecoin.org/gossip/dnet"
	"code.dogecoin.org/gossip/iden"
	"code.dogecoin.org/governor"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
)

type DogeNetProcessor struct {
	governor.ServiceCtx
	dogeNetUrl string
	store      *store.Store
	sock       net.Conn
	key        dnet.KeyPair
}

var ChanFractal = dnet.NewTag("Fractal")

func NewDogeNetProcessor(dogeNetUrl string, store *store.Store) *DogeNetProcessor {
	return &DogeNetProcessor{dogeNetUrl: dogeNetUrl, store: store}
}

func keyFromEnv() dnet.KeyPair {
	// get the private key from the KEY env-var
	keyHex := os.Getenv("KEY")
	os.Setenv("KEY", "") // don't leave the key in the environment
	if keyHex == "" {
		log.Printf("Missing KEY env-var: identity private key (32 bytes)")
		os.Exit(3)
	}
	keyB, err := hex.DecodeString(keyHex)
	if err != nil {
		log.Printf("Invalid KEY hex in env-var: %v", err)
		os.Exit(3)
	}
	if len(keyB) != 32 {
		log.Printf("Invalid KEY hex in env-var: must be 32 bytes")
		os.Exit(3)
	}
	return dnet.KeyPairFromPrivKey((*[32]byte)(keyB))
}

func (s *DogeNetProcessor) Start() {
	fractalKey := keyFromEnv()
	log.Printf("Fractal PubKey is: %v", hex.EncodeToString(fractalKey.Pub[:]))
	s.key = fractalKey

	dogenetUrl, err := url.Parse(s.dogeNetUrl)
	if err != nil {
		log.Printf("[Fractal] cannot parse dogeNetUrl: %v", err)
		return
	}

	// connect to dogenet service
	sock, err := net.Dial(dogenetUrl.Scheme, dogenetUrl.Host)
	if err != nil {
		log.Printf("[Fractal] cannot connect: %v", err)
		return
	}
	log.Printf("[Fractal] connected to dogenet.")
	// send channel bind request
	bind := dnet.BindMessage{Version: 1, Chan: ChanFractal, PubKey: *fractalKey.Pub}
	_, err = sock.Write(bind.Encode())
	if err != nil {
		log.Printf("[Fractal] cannot send BindMessage: %v", err)
		sock.Close()
		return
	}
	// wait for the return bind request
	reader := bufio.NewReader(sock)
	br_buf := [dnet.BindMessageSize]byte{}
	_, err = io.ReadAtLeast(reader, br_buf[:], len(br_buf))
	if err != nil {
		log.Printf("[Fractal] reading BindMessage reply: %v", err)
		sock.Close()
		return
	}
	if _, ok := dnet.DecodeBindMessage(br_buf[:]); ok {
		// send the node's pubkey to the announce service
		// so it can include the node key in the identity announcement
		// s.announceChanges <- spec.NodePubKeyMsg{PubKey: br.PubKey[:]}
		// TODO
	} else {
		log.Printf("[Fractal] invalid BindMessage reply: %v", err)
		sock.Close()
		return
	}
	log.Printf("[Fractal] completed handshake.")
	// begin sending and listening for messages
	s.sock = sock // for Stop()
	// go s.gossipMyIdentity(sock)
	// go s.gossipRandomIdentities(sock)
	// read messages until reading fails
	for !s.Stopping() {
		msg, err := dnet.ReadMessage(reader)
		if err != nil {
			log.Printf("[Fractal] cannot receive from peer: %v", err)
			sock.Close()
			return
		}
		if msg.Chan != ChanFractal {
			log.Printf("[Fractal] ignored message: [%s][%s]", msg.Chan, msg.Tag)
			continue
		}
		switch msg.Tag {
		case iden.TagIdentity:
			// s.recvIden(msg)
			// TODO
		default:
			log.Printf("[Fractal] unknown message: [%s][%s]", msg.Chan, msg.Tag)
		}
	}
}

func (p *DogeNetProcessor) Process(ctx context.Context, msg *protocol.MessageEnvelope) error {
	return nil
}
