package config

import "github.com/BurntSushi/toml"

type Config struct {
	DogeNetUrl string `toml:"doge_net_url"`
	RpcUrl     string `toml:"rpc_url"`
	DbUrl      string `toml:"db_url"`
}

func LoadConfig(path string) (*Config, error) {
	var config Config

	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
