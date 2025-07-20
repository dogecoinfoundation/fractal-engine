package commands

import (
	"context"
	"fmt"
	"log"

	fecli "dogecoin.org/fractal-engine/pkg/cli"
	"dogecoin.org/fractal-engine/pkg/cli/keys"
	"dogecoin.org/fractal-engine/pkg/client"
	"github.com/urfave/cli/v3"
)

func getTokenisationClient(ctx context.Context, cmd *cli.Command) (*client.TokenisationClient, error) {
	configPath := cmd.String("config-path")

	config, err := fecli.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	store := keys.NewSecureStore()
	privHex, err := store.Get(config.ActiveKey + "_private_key")
	if err != nil {
		log.Fatal(err)
	}
	pubHex, err := store.Get(config.ActiveKey + "_public_key")
	if err != nil {
		log.Fatal(err)
	}

	url := fmt.Sprintf("http://%s:%s", config.FractalEngineHost, config.FractalEnginePort)

	return client.NewTokenisationClient(url, privHex, pubHex), nil
}
