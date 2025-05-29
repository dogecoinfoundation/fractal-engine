package config

type Config struct {
	RpcServerHost   string
	RpcServerPort   string
	DogeNetHost     string
	DogeNetPort     string
	DogeHost        string
	DogeScheme      string
	DogePort        string
	DogeUser        string
	DogePassword    string
	DatabaseURL     string
	PersistFollower bool
}

func NewConfig() *Config {
	return &Config{
		RpcServerHost:   "0.0.0.0",
		RpcServerPort:   "8080",
		DogeNetHost:     "0.0.0.0",
		DogeNetPort:     "22555",
		DogeScheme:      "http",
		DogeHost:        "0.0.0.0",
		DogePort:        "22556",
		DogeUser:        "doge",
		DatabaseURL:     "sqlite://fractal-engine.db",
		PersistFollower: true,
	}
}
