package doge

import "dogecoin.org/fractal-engine/pkg/config"

type DogeClient struct {
	cfg *config.Config
}

func NewDogeClient(cfg *config.Config) *DogeClient {
	return &DogeClient{
		cfg: cfg,
	}
}
