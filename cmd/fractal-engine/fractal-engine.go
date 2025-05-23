package main

import (
	"dogecoin.org/fractal-engine/pkg/service"
)

func main() {
	service := service.NewTokenisationService()
	service.Start()
}
