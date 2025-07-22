package commands

import (
	"context"
	"log"
	"strconv"

	fecli "dogecoin.org/fractal-engine/pkg/cli"
	climodels "dogecoin.org/fractal-engine/pkg/cli/climodels"
	"dogecoin.org/fractal-engine/pkg/cli/keys"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

var TokensCommand = &cli.Command{
	Name:  "tokens",
	Usage: "Manage tokens",
	Commands: []*cli.Command{
		{
			Name:   "list",
			Usage:  "List tokens",
			Action: listTokensAction,
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

func listTokensAction(ctx context.Context, cmd *cli.Command) error {
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

	tokenisationClient, err := getTokenisationClient(ctx, cmd)
	if err != nil {
		log.Fatal(err)
	}

	var mintHash string

	group := huh.NewGroup(
		huh.NewInput().
			Title("What is the Mint Hash?").
			Value(&mintHash),
	)

	form := huh.NewForm(group)
	err = form.Run()
	if err != nil {
		log.Fatal(err)
	}

	tokenBalances, err := tokenisationClient.GetTokenBalance(address, mintHash)
	if err != nil {
		log.Fatal(err)
	}

	rows := []table.Row{}
	for _, tokenBalance := range tokenBalances {
		rows = append(rows, table.Row{tokenBalance.Address, tokenBalance.MintHash, strconv.Itoa(tokenBalance.Quantity)})
	}

	mintTable := climodels.CliTableModel{
		Table: table.New(
			table.WithColumns([]table.Column{
				{Title: "Address", Width: 20},
				{Title: "Mint Hash", Width: 64},
				{Title: "Quantity", Width: 10},
			}),
		),
	}

	mintTable.Table.SetRows(rows)

	p := tea.NewProgram(mintTable)
	_, err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
