package cli

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	FractalEngineHost string   `toml:"fractal_engine_host"`
	FractalEnginePort string   `toml:"fractal_engine_port"`
	KeyLabels         []string `toml:"key_labels"`
	ActiveKey         string   `toml:"active_key"`
}

func SaveConfig(config *Config, path string) error {
	marshalled, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, marshalled, 0644)
}

func LoadConfig(path string) (*Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
