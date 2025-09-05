package doge

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/btcutil/base58"
	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/gowebpki/jcs"
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

func isCompressedPubHex(pubHex string) bool {
	h := strings.ToLower(strings.TrimSpace(pubHex))
	return strings.HasPrefix(h, "02") || strings.HasPrefix(h, "03")
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

func SignPayload(payload interface{}, privHex, pubHex string) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	digest, err := CanonicalHash(b) // [32]byte — SAME function used by verifier
	if err != nil {
		return "", err
	}

	skBytes, err := hex.DecodeString(strings.TrimSpace(privHex))
	if err != nil {
		return "", err
	}
	sk := secp.PrivKeyFromBytes(skBytes)

	compressed := strings.HasPrefix(strings.ToLower(strings.TrimSpace(pubHex)), "02") ||
		strings.HasPrefix(strings.ToLower(strings.TrimSpace(pubHex)), "03")

	sig := ecdsa.SignCompact(sk, digest[:], compressed) // 65 bytes
	return base64.StdEncoding.EncodeToString(sig), nil  // <- base64, not hex
}

func CanonicalHash(jsonBytes []byte) ([32]byte, error) {
	canon, err := jcs.Transform(jsonBytes) // returns canonical JSON bytes
	if err != nil {
		return [32]byte{}, err
	}

	return sha256.Sum256(canon), nil
}

func ValidateSignature(payload interface{}, publicKey string, signature string) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	hashBytes, err := CanonicalHash(payloadBytes)
	if err != nil {
		return err
	}

	err = VerifyDogecoinCompactSigFromHexHash(hex.EncodeToString(hashBytes[:]), signature, publicKey)
	if err != nil {
		log.Println(err)
		return errors.New("signature verification failed")
	}

	return nil
}

func dogeMessageHash(msg []byte) [32]byte {
	// "\x19Dogecoin Signed Message:\n" + varint(len(msg)) + msg, then double-SHA256
	prefix := []byte("\x19Dogecoin Signed Message:\n")
	var buf bytes.Buffer
	buf.Write(prefix)
	// varint (little-endian payload after the tag)
	n := uint64(len(msg))
	switch {
	case n < 0xfd:
		buf.WriteByte(byte(n))
	case n <= 0xffff:
		buf.Write([]byte{0xfd, byte(n), byte(n >> 8)})
	case n <= 0xffffffff:
		buf.Write([]byte{0xfe, byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24)})
	default:
		buf.Write([]byte{
			0xff,
			byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24),
			byte(n >> 32), byte(n >> 40), byte(n >> 48), byte(n >> 56),
		})
	}
	buf.Write(msg)

	h1 := sha256.Sum256(buf.Bytes())
	return sha256.Sum256(h1[:])
}

// VerifyDogecoinCompactSigFromHexHash expects:
//   - hexHash: hex-encoded 32-byte digest
//   - sigB64:  base64 of a 65-byte Bitcoin-style compact signature
//   - pubHex:  hex pubkey (02/03.. compressed or 04.. uncompressed)
func VerifyDogecoinCompactSigFromHexHash(hexHash, sigB64, pubHex string) error {
	// 1) Decode inputs
	h, err := hex.DecodeString(strings.TrimSpace(hexHash))
	if err != nil || len(h) != 32 {
		return fmt.Errorf("bad hash: %v (len=%d)", err, len(h))
	}
	sig, err := base64.StdEncoding.DecodeString(strings.TrimSpace(sigB64))
	if err != nil {
		return fmt.Errorf("bad signature base64: %v", err)
	}
	if len(sig) != 65 {
		return fmt.Errorf("bad compact sig len: %d (want 65)", len(sig))
	}
	expBytes, err := hex.DecodeString(strings.TrimSpace(pubHex))
	if err != nil {
		return fmt.Errorf("bad pubHex: %v", err)
	}
	wantPub, err := secp.ParsePubKey(expBytes)
	if err != nil {
		return fmt.Errorf("parse pubHex: %v", err)
	}

	// 2) Inspect header (Bitcoin-style: 27..34)
	hdr := sig[0]
	if hdr < 27 || hdr > 34 {
		return fmt.Errorf("unexpected compact header: %d", hdr)
	}

	compressed := ((hdr - 27) & 4) != 0

	// Optional sanity: header’s compressed bit should match pubHex form
	isExpCompressed := expBytes[0] == 0x02 || expBytes[0] == 0x03
	if compressed != isExpCompressed {
		return fmt.Errorf("compressed-bit mismatch: sig says %v, pubHex compressed=%v", compressed, isExpCompressed)
	}

	// 3) Recover and compare the EC point (format-agnostic)
	recPub, _, err := ecdsa.RecoverCompact(sig, h)
	if err != nil {
		return fmt.Errorf("recover failed: %v", err)
	}

	if !bytes.Equal(recPub.SerializeCompressed(), wantPub.SerializeCompressed()) {
		return fmt.Errorf("recovered pubkey mismatch; got %x want %x",
			recPub.SerializeCompressed(), wantPub.SerializeCompressed())
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
