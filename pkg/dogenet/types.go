package dogenet

import "code.dogecoin.org/gossip/dnet"

var ChanFE = dnet.NewTag("FractalEngine")
var TagMint = dnet.NewTag("Mint")
var TagBuyOffer = dnet.NewTag("BuyOffer")
var TagSellOffer = dnet.NewTag("SellOffer")
var TagInvoice = dnet.NewTag("Invoice")
var TagInvoiceSignature = dnet.NewTag("InvoiceSignature")
var TagDeleteBuyOffer = dnet.NewTag("DeleteBuyOffer")
var TagDeleteSellOffer = dnet.NewTag("DeleteSellOffer")

type GossipMessage struct {
	Topic string `json:"topic"`
	Data  []byte `json:"data"`
}

type GossipMessageListener func(message GossipMessage)
