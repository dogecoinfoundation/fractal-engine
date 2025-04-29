package doge

import (
	"fmt"
	"sort"
)

const (
	TxOverhead      = 10  // Version (4) + LockTime (4) + Segwit marker and flag (2) in a typical segwit tx
	TxInputOverhead = 41  // OutPoint (36) + Sequence (4) + Script length (1)
	TxInputWitness  = 107 // Typical size for a P2WPKH input witness (signature + pubkey)
	TxOutputBase    = 9   // Value (8) + Script length (1)
	P2PKHSize       = 25  // Size of P2PKH output script
	OpReturnBase    = 9   // Value (8) + Script length (1)
	OpReturnScript  = 2   // OP_RETURN (1) + data length (1)
)

func EstimateTxSizeWithOpReturn(opReturnData string, changeNeeded bool) int {
	size := TxOverhead

	// Will add input sizes later when we know how many we need

	// OP_RETURN output
	opReturnSize := OpReturnBase + OpReturnScript + len(opReturnData)

	// Regular outputs (assuming one change output if needed)
	outputsSize := 0
	if changeNeeded {
		outputsSize = TxOutputBase + P2PKHSize
	}

	return size + opReturnSize + outputsSize
}

func EstimateInputSize() int {
	return TxInputOverhead + TxInputWitness
}

// EstimateFinalTxSize estimates the final transaction size once we know how many inputs
func EstimateFinalTxSize(baseSize int, numInputs int) int {
	return baseSize + (numInputs * EstimateInputSize())
}

func SelectUTXOs(availableUTXOs []UTXO, baseAmount int64, feeRate int64) ([]UTXO, int64, error) {
	// Sort UTXOs by amount (smallest first)
	sort.Slice(availableUTXOs, func(i, j int) bool {
		return availableUTXOs[i].Amount < availableUTXOs[j].Amount
	})

	// Initial size estimate without inputs
	baseTxSize := EstimateTxSizeWithOpReturn("", true) // Just for initial estimate

	// Iteratively try adding inputs until we have enough
	selectedUTXOs := []UTXO{}
	var totalInputAmount int64

	for {
		// Calculate current estimated size with current number of inputs
		currentNumInputs := len(selectedUTXOs)
		estimatedSize := EstimateFinalTxSize(baseTxSize, currentNumInputs)

		// Calculate fee based on estimated size
		estimatedFee := feeRate * int64(estimatedSize)

		// Total amount needed (base amount + fee)
		totalNeeded := baseAmount + estimatedFee

		// Check if we have enough
		if totalInputAmount >= totalNeeded {
			// We have enough, calculate change
			change := totalInputAmount - totalNeeded
			return selectedUTXOs, change, nil
		}

		// We need more inputs
		if currentNumInputs >= len(availableUTXOs) {
			// We've used all available UTXOs and still don't have enough
			return nil, 0, fmt.Errorf("insufficient funds: required %d satoshis but only have %d satoshis available",
				totalNeeded, totalInputAmount)
		}

		// Add another UTXO
		nextUTXO := availableUTXOs[currentNumInputs]
		selectedUTXOs = append(selectedUTXOs, nextUTXO)
		totalInputAmount += int64(nextUTXO.Amount)
	}

}
