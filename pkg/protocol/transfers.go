package protocol

type TransferRequest struct {
	Id              string `json:"id"`
	SenderAddress   string `json:"sender_address"`
	ReceiverAddress string `json:"receiver_address"`
	MintId          string `json:"mint_id"`
	Amount          int64  `json:"amount"`
	PricePerToken   int64  `json:"price_per_token"`
	TransactionHash string `json:"transaction_hash"`
	Approved        bool   `json:"approved"`
	ApprovedAt      string `json:"approved_at"`
	Verified        bool   `json:"verified"`
	CreatedAt       string `json:"created_at"`
}

type Transfer struct {
	Id                string `json:"id"`
	TransferRequestId string `json:"transfer_request_id"`
	TransactionHash   string `json:"transaction_hash"`
	Verified          bool   `json:"verified"`
	CreatedAt         string `json:"created_at"`
}

type Account struct {
	Id        string `json:"id"`
	Address   string `json:"address"`
	Balance   int64  `json:"balance"`
	MintId    string `json:"mint_id"`
	CreatedAt string `json:"created_at"`
}
