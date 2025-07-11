package dogenet

import (
	"encoding/json"
	"log"

	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/Dogebox-WG/gossip/dnet"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *DogeNetClient) GossipUnconfirmedInvoice(record store.UnconfirmedInvoice) error {
	invoiceMessage := protocol.InvoiceMessage{
		Id: record.Id,
		Payload: &protocol.InvoicePayload{
			PaymentAddress:         record.PaymentAddress,
			BuyOfferOffererAddress: record.BuyOfferOffererAddress,
			BuyOfferHash:           record.BuyOfferHash,
			BuyOfferMintHash:       record.BuyOfferMintHash,
			BuyOfferQuantity:       int32(record.BuyOfferQuantity),
			BuyOfferPrice:          int32(record.BuyOfferPrice),
			SellOfferAddress:       record.SellOfferAddress,
		},
		CreatedAt: timestamppb.New(record.CreatedAt),
	}

	envelope := protocol.InvoiceMessageEnvelope{
		Type:      protocol.ACTION_INVOICE,
		Version:   protocol.DEFAULT_VERSION,
		Payload:   &invoiceMessage,
		PublicKey: record.PublicKey,
		Signature: record.Signature,
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

	invoicePayload, err := json.Marshal(invoice)
	if err != nil {
		log.Println("Error marshalling invoice:", err)
		return
	}

	err = doge.ValidateSignature(invoicePayload, envelope.PublicKey, envelope.Signature)
	if err != nil {
		log.Println("Error validating signature:", err)
		return
	}

	address, err := doge.PublicKeyToDogeAddress(envelope.PublicKey)
	if err != nil {
		log.Println("Error converting public key to doge address:", err)
		return
	}

	if address != invoice.Payload.PaymentAddress {
		log.Println("Payment address does not match public key")
		return
	}

	invoiceWithoutID := store.UnconfirmedInvoice{
		PaymentAddress:         invoice.Payload.PaymentAddress,
		BuyOfferOffererAddress: invoice.Payload.BuyOfferOffererAddress,
		BuyOfferHash:           invoice.Payload.BuyOfferHash,
		BuyOfferMintHash:       invoice.Payload.BuyOfferMintHash,
		BuyOfferQuantity:       int(invoice.Payload.BuyOfferQuantity),
		BuyOfferPrice:          int(invoice.Payload.BuyOfferPrice),
		CreatedAt:              invoice.CreatedAt.AsTime(),
		Hash:                   invoice.Hash,
		Id:                     invoice.Id,
		PublicKey:              envelope.PublicKey,
		SellOfferAddress:       invoice.Payload.SellOfferAddress,
		Signature:              envelope.Signature,
	}

	id, err := c.store.SaveUnconfirmedInvoice(&invoiceWithoutID)
	if err != nil {
		log.Println("Error saving unconfirmed invoice:", err)
		return
	}

	log.Printf("[FE] unconfirmed invoice saved: %v", id)
}
