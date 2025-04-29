package api

import (
	"dogecoin.org/fractal-engine/pkg/protocol"
)

type CreateTransferRequestRequest struct {
	protocol.TransferRequest
}

type CreateTransferRequestResponse struct {
	Id string `json:"id"`
}
