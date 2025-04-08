package doge

import (
	"encoding/hex"
	"strings"

	"dogecoin.org/chainfollower/pkg/types"
)

func ParseOpReturnData(vout types.RawTxnVOut) []byte {
	asm := vout.ScriptPubKey.Asm
	parts := strings.Split(asm, " ")
	if len(parts) > 0 {
		op := parts[0]
		if op == "OP_RETURN" {
			bytes, err := hex.DecodeString(parts[1])
			if err != nil {
				return nil
			}

			return bytes
		}
	}
	return nil
}
