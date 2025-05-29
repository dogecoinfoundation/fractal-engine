package protocol

import "bytes"

type MintTransaction struct {
	MintID string `json:"mint_id"`
}

func NewMintTransactionEnvelope(mintId string) *MessageEnvelope {
	return NewMessageEnvelope(ACTION_MINT, []byte(mintId))
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
