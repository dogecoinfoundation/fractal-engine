package protocol

import (
	"bytes"
	"encoding/binary"
)

// Version
// PROTOBUF
// 1.0.0

const (
	FRACTAL_ENGINE_IDENTIFIER = 0xFE0001FE
	DEFAULT_VERSION           = 1
	ACTION_MINT               = 0x01
	ACTION_SELL_OFFER         = 0x02
	ACTION_BUY_OFFER          = 0x03
	ACTION_INVOICE            = 0x04
	ACTION_RECEIPT            = 0x05
)

type MessageEnvelope struct {
	EngineIdentifier uint32
	Action           uint8
	Version          uint8
	Data             []byte
}

func NewMessageEnvelope(action uint8, version uint8, data []byte) *MessageEnvelope {
	return &MessageEnvelope{
		EngineIdentifier: FRACTAL_ENGINE_IDENTIFIER,
		Action:           action,
		Version:          version,
		Data:             data,
	}
}

func (m *MessageEnvelope) IsFractalEngineMessage() bool {
	return m.EngineIdentifier == FRACTAL_ENGINE_IDENTIFIER
}

func (m *MessageEnvelope) Serialize() []byte {
	bufIdentifier := make([]byte, 4)
	binary.BigEndian.PutUint32(bufIdentifier, m.EngineIdentifier)

	buf := new(bytes.Buffer)
	buf.Write(bufIdentifier)
	buf.WriteByte(m.Action)
	buf.Write(m.Data)

	return buf.Bytes()
}

func (m *MessageEnvelope) Deserialize(data []byte) error {
	buf := bytes.NewBuffer(data)

	bufIdentifier := make([]byte, 4)
	buf.Read(bufIdentifier)
	m.EngineIdentifier = binary.BigEndian.Uint32(bufIdentifier)

	action, err := buf.ReadByte()
	if err != nil {
		return err
	}

	m.Action = action

	m.Data = buf.Bytes()
	return nil
}
