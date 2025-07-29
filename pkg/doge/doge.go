package doge

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

const (
	PrefixMainnet = 0x1E
	PrefixTestnet = 0x71
	PrefixRegtest = 0x6f
)

type PrevOutput struct {
	Address string // The address of the previous output
	Amount  int64  // The amount in satoshis of the previous output
}

func WrapOpReturn(data []byte) string {
	opReturnScript := append([]byte{0x6a, byte(len(data))}, data...)
	return hex.EncodeToString(opReturnScript)
}

func GetPrefix(prefixStr string) (byte, error) {
	switch prefixStr {
	case "mainnet":
		return PrefixMainnet, nil
	case "testnet":
		return PrefixTestnet, nil
	case "regtest":
		return PrefixRegtest, nil
	}

	return 0, fmt.Errorf("invalid prefix: %s", prefixStr)
}

func GetChainCfg(prefix byte) *chaincfg.Params {
	switch prefix {
	case PrefixMainnet:
		return &chaincfg.MainNetParams
	case PrefixTestnet:
		return &chaincfg.TestNet3Params
	case PrefixRegtest:
		return &chaincfg.RegressionNetParams
	}

	return nil
}

func GenerateDogecoinKeypair(prefix byte) (privHex string, pubHex string, address string, err error) {
	privKey, err := btcec.NewPrivateKey()
	if err != nil {
		return "", "", "", err
	}

	pubKey := privKey.PubKey()
	pubKeyBytes := pubKey.SerializeCompressed()

	pubKeyHash := btcutil.Hash160(pubKeyBytes)
	address = base58.CheckEncode(pubKeyHash, byte(prefix))

	return hex.EncodeToString(privKey.Serialize()), hex.EncodeToString(pubKeyBytes), address, nil
}

func SignRawTransaction(rawTxHex string, privKeyHex string, prevTxOuts []PrevOutput, chainCfg *chaincfg.Params) (string, error) {
	// Decode the raw transaction
	rawTxBytes, err := hex.DecodeString(rawTxHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode raw transaction: %v", err)
	}

	// Deserialize the transaction
	tx := wire.NewMsgTx(wire.TxVersion)
	err = tx.Deserialize(bytes.NewReader(rawTxBytes))
	if err != nil {
		return "", fmt.Errorf("failed to deserialize transaction: %v", err)
	}

	// Convert hex private key to btcec.PrivateKey
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %v", err)
	}

	privKey, pubKey := btcec.PrivKeyFromBytes(privKeyBytes)

	// Sign each input
	for i := range tx.TxIn {
		// Get the previous output info for this input
		if i >= len(prevTxOuts) {
			return "", fmt.Errorf("missing previous output info for input %d", i)
		}

		prevOut := prevTxOuts[i]

		// Create the signature script for the previous output
		// This assumes P2PKH - modify if you're using different script types
		prevAddr, err := btcutil.DecodeAddress(prevOut.Address, chainCfg)
		if err != nil {
			return "", fmt.Errorf("failed to decode address for input %d: %v", i, err)
		}

		prevScript, err := txscript.PayToAddrScript(prevAddr)
		if err != nil {
			return "", fmt.Errorf("failed to create script for input %d: %v", i, err)
		}

		// Create signature hash
		sigHash, err := txscript.CalcSignatureHash(prevScript, txscript.SigHashAll, tx, i)
		if err != nil {
			return "", fmt.Errorf("failed to calculate signature hash for input %d: %v", i, err)
		}

		// Convert to chainhash.Hash for signing
		var sigHashBytes chainhash.Hash
		copy(sigHashBytes[:], sigHash)

		// Create the signature using btcec/v2
		signature := ecdsa.Sign(privKey, sigHashBytes[:])

		// Serialize signature with SIGHASH_ALL byte
		sigBytes := append(signature.Serialize(), byte(txscript.SigHashAll))

		// Build the signature script (scriptSig)
		sigScript, err := txscript.NewScriptBuilder().
			AddData(sigBytes).
			AddData(pubKey.SerializeCompressed()).
			Script()
		if err != nil {
			return "", fmt.Errorf("failed to build signature script for input %d: %v", i, err)
		}

		// Set the signature script on the input
		tx.TxIn[i].SignatureScript = sigScript

		// Verify the signature (optional but recommended)
		vm, err := txscript.NewEngine(
			prevScript,
			tx,
			i,
			txscript.StandardVerifyFlags,
			nil,
			nil,
			prevOut.Amount,
		)
		if err != nil {
			return "", fmt.Errorf("failed to create script engine for input %d: %v", i, err)
		}

		if err := vm.Execute(); err != nil {
			return "", fmt.Errorf("signature verification failed for input %d: %v", i, err)
		}
	}

	// Serialize the signed transaction
	var signedTxBuf bytes.Buffer
	err = tx.Serialize(&signedTxBuf)
	if err != nil {
		return "", fmt.Errorf("failed to serialize signed transaction: %v", err)
	}

	fmt.Printf("Signed transaction hex: %x\n", signedTxBuf.Bytes())

	rawTxHex = hex.EncodeToString(signedTxBuf.Bytes())

	return rawTxHex, nil
}

func HexToDogecoinWIF(hexKey string, compressed bool) (string, error) {
	// Step 1: Decode hex
	privKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", fmt.Errorf("invalid hex key: %v", err)
	}

	// Step 2: Add Dogecoin WIF prefix (0x9e)
	prefix := []byte{0x9e}
	payload := append(prefix, privKeyBytes...)

	// Step 3: Add 0x01 for compressed keys
	if compressed {
		payload = append(payload, 0x01)
	}

	// Step 4: Double SHA256 checksum
	checksum := sha256.Sum256(payload)
	checksum = sha256.Sum256(checksum[:])
	full := append(payload, checksum[:4]...)

	// Step 5: Base58Check encode
	wif := base58.Encode(full)
	return wif, nil
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

func PublicKeyToDogeAddress(pubKeyHex string, prefix byte) (string, error) {
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

	versionedPayload := append([]byte{prefix}, pubKeyHash...)

	firstSHA := sha256.Sum256(versionedPayload)
	secondSHA := sha256.Sum256(firstSHA[:])
	checksum := secondSHA[:4]

	fullPayload := append(versionedPayload, checksum...)

	address := base58.Encode(fullPayload)
	return address, nil
}
