package main

import (
	"fmt"
	"log"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/client"
)

func main() {
	httpClient := &http.Client{}
	feClient := client.NewFractalEngineClient("http://localhost:8080", httpClient)

	// mintRequest := api.NewCreateMintRequest(protocol.MintWithoutID{
	// 	Title:       "Test Mint",
	// 	Description: "Test Description",
	// 	Tags:        []string{"test", "mint"},
	// 	Metadata: map[string]interface{}{
	// 		"test": "test",
	// 	},
	// })

	// id, err := feClient.CreateMint(mintRequest)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(id)
	mints, err := feClient.GetMints(0, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(mints)
}
