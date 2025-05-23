package dogenet

import "dogecoin.org/fractal-engine/pkg/config"

type DogeNetClient struct {
	cfg *config.Config
}

func NewDogeNetClient(cfg *config.Config) *DogeNetClient {
	return &DogeNetClient{
		cfg: cfg,
	}
}
