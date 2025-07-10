package doge

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
)

func WrapOpReturn(data []byte) string {
	opReturnScript := append([]byte{0x6a, byte(len(data))}, data...)
	return hex.EncodeToString(opReturnScript)
}

func GenerateDogecoinKeypair() (privHex string, pubHex string, address string, err error) {
	privKey, err := btcec.NewPrivateKey()
	if err != nil {
		return "", "", "", err
	}

	pubKey := privKey.PubKey()
	pubKeyBytes := pubKey.SerializeCompressed()

	pubKeyHash := btcutil.Hash160(pubKeyBytes)
	address = base58.CheckEncode(pubKeyHash, 0x1E)

	return hex.EncodeToString(privKey.Serialize()), hex.EncodeToString(pubKeyBytes), address, nil
}

func SignPayload(payload []byte, privHex string) (string, error) {
	// Step 1: Decode the private key from hex
	privBytes, err := hex.DecodeString(privHex)
	if err != nil {
		return "", err
	}

	privKey, _ := btcec.PrivKeyFromBytes(privBytes)

	// Step 2: Hash the payload (using SHA256)
	hash := sha256.Sum256(payload)

	// Step 3: Sign the hash
	signature := ecdsa.Sign(privKey, hash[:])

	// Step 4: Encode signature as DER, then to hex string
	sigDER := signature.Serialize()
	sigHex := hex.EncodeToString(sigDER)

	return sigHex, nil
}
