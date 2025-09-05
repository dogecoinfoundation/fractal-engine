package doge_test

import (
	"testing"

	"dogecoin.org/fractal-engine/pkg/doge"
)

func TestDogeSigning(t *testing.T) {
	privHex, pubHex, _, err := doge.GenerateDogecoinKeypair(doge.PrefixRegtest)
	if err != nil {
		t.Fatal(err)
	}

	payload := map[string]string{
		"hello": "doge",
	}

	signature, err := doge.SignPayload(payload, privHex, pubHex)
	if err != nil {
		t.Fatal(err)
	}

	err = doge.ValidateSignature(payload, pubHex, signature)
	if err != nil {
		t.Fatal(err)
	}
}
