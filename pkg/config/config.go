package config

type Config struct {
	RpcServerHost string
	RpcServerPort string
	DogeNetHost   string
	DogeNetPort   string
	DogeHost      string
	DogePort      string
	DogeUser      string
	DogePassword  string
	DatabaseURL   string
}

func NewConfig() *Config {
	return &Config{
		RpcServerHost: "0.0.0.0",
		RpcServerPort: "8080",
		DogeNetHost:   "0.0.0.0",
		DogeNetPort:   "22555",
		DogeHost:      "0.0.0.0",
		DogePort:      "22556",
		DogeUser:      "doge",
		DatabaseURL:   "postgres://postgres:postgres@localhost:5432/postgres",
	}
}
