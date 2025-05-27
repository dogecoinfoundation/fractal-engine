package main

import (
	"log"

	"dogecoin.org/fractal-engine/pkg/client"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/service"
	"dogecoin.org/fractal-engine/pkg/store"
)

func main() {
	service := service.NewTokenisationService()
	go service.Start()

	<-service.StatusChan

	feClient := client.NewTokenisationClient("http://localhost:8080")
	result, err := feClient.Mint(&rpc.CreateMintRequest{
		MintWithoutID: store.MintWithoutID{
			Title:         "Test Mint",
			FractionCount: 100,
			Description:   "Test Description",
			Tags:          []string{"test", "mint"},
			Metadata: map[string]interface{}{
				"test": "test",
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println(result)
}
