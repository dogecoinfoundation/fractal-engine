package commands

import (
	"context"
	"fmt"
	"log"

	fecli "dogecoin.org/fractal-engine/pkg/cli"
	"dogecoin.org/fractal-engine/pkg/cli/keys"
	"dogecoin.org/fractal-engine/pkg/indexer"
	"github.com/btcsuite/btcutil/base58"
	"github.com/urfave/cli/v3"
)

var IndexerCommand = &cli.Command{
	Name:  "indexer",
	Usage: "Balance management",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "config-path",
			Usage: "Path to the config file",
			Value: "config.toml",
		},
	},
	Commands: []*cli.Command{
		{
			Name:   "health",
			Usage:  "Check the health of the indexer",
			Action: indexerHealthAction,
		},
		{
			Name:   "balance",
			Usage:  "Get the balance of an address",
			Action: indexerBalanceAction,
		},
		{
			Name:   "utxo",
			Usage:  "Get the utxo of an address",
			Action: indexerUtxoAction,
		},
	},
}

func indexerUtxoAction(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config-path")

	config, err := fecli.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	secureStore := keys.NewSecureStore()

	address, err := secureStore.Get(config.ActiveKey + "_address")
	if err != nil {
		log.Fatal(err)
	}

	indexerClient := indexer.NewIndexerClient(config.IndexerURL)

	utxos, err := indexerClient.GetUTXO(address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Indexer utxos:", utxos.UTXOs)

	return nil
}

func indexerBalanceAction(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config-path")

	config, err := fecli.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	secureStore := keys.NewSecureStore()

	address, err := secureStore.Get(config.ActiveKey + "_address")
	if err != nil {
		log.Fatal(err)
	}

	res, _, _ := base58.CheckDecode(address)

	fmt.Printf("Address: %x\n", res)

	indexerClient := indexer.NewIndexerClient(config.IndexerURL)

	balance, err := indexerClient.GetBalance(address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Indexer balance:", balance)

	return nil
}

func indexerHealthAction(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config-path")

	config, err := fecli.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	indexerClient := indexer.NewIndexerClient(config.IndexerURL)

	health, err := indexerClient.GetHealth()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Indexer health:", health.OK)

	return nil
}
