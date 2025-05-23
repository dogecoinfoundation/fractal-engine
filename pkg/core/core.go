package core

import (
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/dogenet"
)

type TokenisationCore struct {
	DogeNetClient *dogenet.DogeNetClient
	DogeClient    *doge.DogeClient
	Config        *config.Config
}

func NewTokenisationCore(cfg *config.Config) *TokenisationCore {
	return &TokenisationCore{
		DogeNetClient: dogenet.NewDogeNetClient(),
		DogeClient:    doge.NewDogeClient(),
		Config:        cfg,
	}
}
