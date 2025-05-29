package main

import (
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/service"
)

func main() {
	cfg := config.NewConfig()

	service := service.NewTokenisationService(cfg)
	service.Start()
}
