package main

import (
	"fmt"
	"log"

	"dogecoin.org/fractal-engine/pkg/doge"
)

type MyBody struct {
	Content string `json:"content"`
}

func main() {
	privHex, pubHex, _, err := doge.GenerateDogecoinKeypair(doge.PrefixRegtest)
	if err != nil {
		log.Fatal(err)
	}

	body := MyBody{
		Content: "Hello, World!",
	}

	sig, err := doge.SignPayload(body, privHex, pubHex)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Signature:", sig)
}
