package protocol

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"google.golang.org/protobuf/proto"
)

type InvoiceTransaction struct {
	InvoiceID string `json:"invoice_id"`
}

func NewInvoiceTransactionEnvelope(hash string, mintHash string, quantity int32, action uint8) MessageEnvelope {
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		fmt.Errorf("Failed to decode hash: %s", err.Error())
		return MessageEnvelope{}
	}

	mintHashBytes, err := hex.DecodeString(mintHash)
	if err != nil {
		fmt.Errorf("Failed to decode hash: %s", err.Error())
		return MessageEnvelope{}
	}

	message := &OnChainInvoiceMessage{
		Version:     DEFAULT_VERSION,
		InvoiceHash: hashBytes,
		MintHash:    mintHashBytes,
		Quantity:    quantity,
	}

	protoBytes, err := proto.Marshal(message)
	if err != nil {
		return MessageEnvelope{}
	}

	return NewMessageEnvelope(action, DEFAULT_VERSION, protoBytes)
}

func (m *InvoiceTransaction) Serialize() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte(m.InvoiceID))
	return buf.Bytes()
}

func (m *InvoiceTransaction) Deserialize(data []byte) error {
	m.InvoiceID = string(data)
	return nil
}
