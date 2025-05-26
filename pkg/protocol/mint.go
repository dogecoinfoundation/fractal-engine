package protocol

type MintTransaction struct {
	TransactionHash string `json:"transaction_hash"`
}

func NewMintTransactionEnvelope(transactionHash string) *MessageEnvelope {
	return NewMessageEnvelope(ACTION_MINT, []byte(transactionHash))
}
