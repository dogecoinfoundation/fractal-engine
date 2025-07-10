package dogenet

import "github.com/Dogebox-WG/gossip/dnet"

var ChanFE = dnet.NewTag("FractalEngine")
var TagMint = dnet.NewTag("Mint")
var TagBuyOffer = dnet.NewTag("BuyOffer")
var TagSellOffer = dnet.NewTag("SellOffer")
var TagInvoice = dnet.NewTag("Invoice")
var TagDeleteBuyOffer = dnet.NewTag("DeleteBuyOffer")
var TagDeleteSellOffer = dnet.NewTag("DeleteSellOffer")

type GossipMessage struct {
	Topic string `json:"topic"`
	Data  []byte `json:"data"`
}

type GossipMessageListener func(message GossipMessage)
