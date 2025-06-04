package config

type Config struct {
	RpcServerHost   string
	RpcServerPort   string
	DogeNetNetwork  string
	DogeNetAddress  string
	DogeHost        string
	DogeScheme      string
	DogePort        string
	DogeUser        string
	DogePassword    string
	DatabaseURL     string
	PersistFollower bool
	MigrationsPath  string
}

func NewConfig() *Config {
	return &Config{
		RpcServerHost:   "0.0.0.0",
		RpcServerPort:   "8891",
		DogeNetNetwork:  "tcp",
		DogeNetAddress:  "0.0.0.0:8085",
		DogeScheme:      "http",
		DogeHost:        "dogecoin",
		DogePort:        "22555",
		DogeUser:        "test",
		DogePassword:    "test",
		DatabaseURL:     "memory://fractal-engine.db",
		PersistFollower: true,
		MigrationsPath:  "db/migrations",
	}
}
