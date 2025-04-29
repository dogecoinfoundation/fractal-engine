package main

import (
	"fmt"

	"dogecoin.org/fractal-engine/pkg/server"
)

func main() {
	server := server.NewFractalServer(nil)
	status := make(chan string)
	go server.Start(status)

	for {
		select {
		case <-status:
			fmt.Println("Status: ", <-status)
		}
	}
}
