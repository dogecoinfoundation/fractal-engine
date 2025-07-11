package doge

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
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

func ValidateSignature(payload []byte, publicKey string, signature string) error {
	// 2. Hash message
	hash := sha256.Sum256(payload)

	// 3. Decode public key
	pubKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return errors.New("invalid public key format")
	}

	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return errors.New("failed to parse public key")
	}

	// 4. Decode signature
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return errors.New("invalid signature encoding")
	}

	sig, err := ecdsa.ParseDERSignature(sigBytes)
	if err != nil {
		return errors.New("failed to parse DER signature")
	}

	// 5. Verify signature
	if !sig.Verify(hash[:], pubKey) {
		return errors.New("signature verification failed")
	}

	return nil
}

func PublicKeyToDogeAddress(pubKeyHex string) (string, error) {
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid public key hex: %v", err)
	}

	sha256Hasher := sha256.New()
	sha256Hasher.Write(pubKeyBytes)
	shaHashed := sha256Hasher.Sum(nil)

	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(shaHashed)
	pubKeyHash := ripemd160Hasher.Sum(nil)

	versionedPayload := append([]byte{0x1E}, pubKeyHash...)

	firstSHA := sha256.Sum256(versionedPayload)
	secondSHA := sha256.Sum256(firstSHA[:])
	checksum := secondSHA[:4]

	fullPayload := append(versionedPayload, checksum...)

	address := base58.Encode(fullPayload)
	return address, nil
}
