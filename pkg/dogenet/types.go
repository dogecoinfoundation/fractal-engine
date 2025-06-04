package dogenet

import "github.com/Dogebox-WG/gossip/dnet"

var ChanFE = dnet.NewTag("FractalEngine")
var TagMint = dnet.NewTag("Mint")

type GossipMessage struct {
	Topic string `json:"topic"`
	Data  []byte `json:"data"`
}

type GossipMessageListener func(message GossipMessage)
