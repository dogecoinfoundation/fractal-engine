package protocol

import "google.golang.org/protobuf/proto"

func NewPaymentTransactionEnvelope(invoiceHash string, action uint8) MessageEnvelope {
	message := &OnChainPaymentMessage{
		Hash: invoiceHash,
	}

	protoBytes, err := proto.Marshal(message)
	if err != nil {
		return MessageEnvelope{}
	}

	return NewMessageEnvelope(action, DEFAULT_VERSION, protoBytes)
}
