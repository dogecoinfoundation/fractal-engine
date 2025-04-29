package protocol

import (
	"bytes"
	"encoding/binary"
)

const (
	FRACTAL_ENGINE_IDENTIFIER = 0xFE0001FE
	ACTION_MINT               = 0x01
	ACTION_TRANSFER_REQUEST   = 0x02
)

type MessageEnvelope struct {
	EngineIdentifier uint32
	Action           uint8
	Data             []byte
}

func NewMessageEnvelope(action uint8, data []byte) *MessageEnvelope {
	return &MessageEnvelope{
		EngineIdentifier: FRACTAL_ENGINE_IDENTIFIER,
		Action:           action,
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

	var err error
	bufIdentifier := make([]byte, 4)
	buf.Read(bufIdentifier)
	m.EngineIdentifier = binary.BigEndian.Uint32(bufIdentifier)

	m.Action, err = buf.ReadByte()
	if err != nil {
		return err
	}

	m.Data = buf.Bytes()
	return nil
}
