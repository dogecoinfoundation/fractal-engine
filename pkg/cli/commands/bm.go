package commands

import (
	"context"
	"fmt"
	"log"

	fecli "dogecoin.org/fractal-engine/pkg/cli"
	climodels "dogecoin.org/fractal-engine/pkg/cli/climodels"
	"dogecoin.org/fractal-engine/pkg/cli/keys"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	bmclient "github.com/dogecoinfoundation/balance-master/pkg/client"
	"github.com/urfave/cli/v3"
)

var BmCommand = &cli.Command{
	Name:  "bm",
	Usage: "helper calls for balance master",
	Commands: []*cli.Command{
		{
			Name:   "utxos",
			Usage:  "get utxos for the active key",
			Action: getUtxosAction,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "config-path",
					Usage: "Path to the config file",
					Value: "config.toml",
				},
			},
		},
	},
}

func getUtxosAction(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config-path")

	config, err := fecli.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	bmClient := bmclient.NewBalanceMasterClient(&bmclient.BalanceMasterClientConfig{
		RpcServerHost: config.BalanceMasterHost,
		RpcServerPort: config.BalanceMasterPort,
	})

	secureStore := keys.NewSecureStore()
	address, err := secureStore.Get(config.ActiveKey + "_address")
	if err != nil {
		log.Fatal(err)
	}

	utxos, err := bmClient.GetUtxos(address)
	if err != nil {
		log.Fatal(err)
	}

	rows := []table.Row{}
	for _, utxo := range utxos {
		rows = append(rows, table.Row{
			utxo.Address,
			fmt.Sprintf("%f", utxo.Amount),
			fmt.Sprintf("%d", utxo.VOut),
		})
	}

	table := climodels.CliTableModel{
		Table: table.New(
			table.WithColumns([]table.Column{
				{Title: "Address", Width: 40},
				{Title: "Amount", Width: 10},
				{Title: "VOut", Width: 10},
			}),
			table.WithRows(rows),
		),
	}

	p := tea.NewProgram(table)
	_, err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
