package protocol

import (
	"google.golang.org/protobuf/proto"
)

func NewMintTransactionEnvelope(hash string, action uint8) MessageEnvelope {
	message := &OnChainMintMessage{
		Hash: hash,
	}

	protoBytes, err := proto.Marshal(message)
	if err != nil {
		return MessageEnvelope{}
	}

	return NewMessageEnvelope(action, DEFAULT_VERSION, protoBytes)
}
