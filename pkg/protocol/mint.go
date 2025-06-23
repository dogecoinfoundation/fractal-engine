package protocol

import (
	"bytes"

	"google.golang.org/protobuf/proto"
)

type MintTransaction struct {
	MintID string `json:"mint_id"`
}

func NewMintTransactionEnvelope(hash string, action uint8) MessageEnvelope {
	message := &OnChainMintMessage{
		Version: DEFAULT_VERSION,
		Hash:    hash,
	}

	protoBytes, err := proto.Marshal(message)
	if err != nil {
		return MessageEnvelope{}
	}

	return NewMessageEnvelope(action, DEFAULT_VERSION, protoBytes)
}

func (m *MintTransaction) Serialize() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte(m.MintID))
	return buf.Bytes()
}

func (m *MintTransaction) Deserialize(data []byte) error {
	m.MintID = string(data)
	return nil
}
