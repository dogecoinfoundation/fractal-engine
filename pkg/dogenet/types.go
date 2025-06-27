package dogenet

import "github.com/Dogebox-WG/gossip/dnet"

var ChanFE = dnet.NewTag("FractalEngine")
var TagMint = dnet.NewTag("Mint")
var TagOffer = dnet.NewTag("Offer")
var TagInvoice = dnet.NewTag("Invoice")

type GossipMessage struct {
	Topic string `json:"topic"`
	Data  []byte `json:"data"`
}

type GossipMessageListener func(message GossipMessage)
