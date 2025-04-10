package main

import (
	"fmt"
	"log"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/api"
	"dogecoin.org/fractal-engine/pkg/client"
	"dogecoin.org/fractal-engine/pkg/protocol"
)

func main() {
	httpClient := &http.Client{}
	feClient := client.NewFractalEngineClient("http://localhost:8080", httpClient)

	mintRequest := api.NewCreateMintRequest(protocol.MintWithoutID{
		Title:       "Test Mint",
		Description: "Test Description",
		Tags:        []string{"test", "mint"},
		Metadata: map[string]interface{}{
			"test": "test",
		},
	})

	id, err := feClient.CreateMint(mintRequest)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(id)
}
