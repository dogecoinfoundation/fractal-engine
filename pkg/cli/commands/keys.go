package commands

import (
	"context"
	"log"

	fecli "dogecoin.org/fractal-engine/pkg/cli"
	climodels "dogecoin.org/fractal-engine/pkg/cli/climodels"
	"dogecoin.org/fractal-engine/pkg/cli/keys"
	"dogecoin.org/fractal-engine/pkg/doge"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

var KeysCommand = &cli.Command{
	Name:  "keys",
	Usage: "Manage keys",
	Commands: []*cli.Command{
		{
			Name:   "create",
			Usage:  "Create a new private key, public key and address",
			Action: createKeyAction,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "config-path",
					Usage: "Path to the config file",
					Value: "config.toml",
				},
			},
		},
		{
			Name:   "list",
			Usage:  "List all keys",
			Action: listKeysAction,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "config-path",
					Usage: "Path to the config file",
					Value: "config.toml",
				},
			},
		},
		{
			Name:   "set",
			Usage:  "Set the active key",
			Action: setActiveKeyAction,
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

func setActiveKeyAction(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config-path")

	config, err := fecli.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	items := []list.Item{}
	for _, label := range config.KeyLabels {
		items = append(items, climodels.ListItem(label))
	}

	listModel := climodels.ListModel{
		List: list.New(
			items,
			climodels.ItemDelegate{},
			20,
			10,
		),
	}

	p := tea.NewProgram(listModel)
	res, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	res2, ok := res.(climodels.ListModel)
	if !ok {
		log.Fatal("No item selected")
	}

	selectedItem := res2.List.SelectedItem()
	if selectedItem == nil {
		log.Fatal("No item selected")
	}

	selectedLabel := selectedItem.(climodels.ListItem)

	log.Println("Selected label:", selectedLabel)

	config.ActiveKey = string(selectedLabel)

	err = fecli.SaveConfig(config, configPath)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func listKeysAction(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config-path")

	config, err := fecli.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	store := keys.NewSecureStore()

	var rows []table.Row
	var activeKeyIndex int
	for idx, label := range config.KeyLabels {
		address, err := store.Get(label + "_address")
		if err != nil {
			continue
		}

		publicKey, err := store.Get(label + "_public_key")
		if err != nil {
			continue
		}

		active := ""
		if label == config.ActiveKey {
			active = "*"
			activeKeyIndex = idx
		}

		rows = append(rows, table.Row{active, label, address, publicKey})
	}

	table := climodels.CliTableModel{
		Table: table.New(
			table.WithColumns([]table.Column{
				{Title: "Active", Width: 8},
				{Title: "Name", Width: 10},
				{Title: "Address", Width: 40},
				{Title: "Public Key", Width: 64},
			}),
			table.WithRows(rows),
		),
	}

	table.Table.SetCursor(activeKeyIndex)

	p := tea.NewProgram(table)
	_, err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func createKeyAction(ctx context.Context, cmd *cli.Command) error {
	var label string
	var prefixStr string

	group := huh.NewGroup(
		huh.NewInput().
			Title("What is the label for this key?").
			Value(&label),
		huh.NewSelect[string]().
			Title("What chain is this key for?").
			Options(
				huh.NewOption("Mainnet", "mainnet"),
				huh.NewOption("Testnet", "testnet"),
				huh.NewOption("Regtest", "regtest"),
			).
			Value(&prefixStr),
	)

	form := huh.NewForm(group)
	err := form.Run()
	if err != nil {
		return err
	}

	prefix, err := doge.GetPrefix(prefixStr)
	if err != nil {
		return err
	}

	privHex, pubHex, address, err := doge.GenerateDogecoinKeypair(prefix)
	if err != nil {
		return err
	}

	configPath := cmd.String("config-path")

	config, err := fecli.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	store := keys.NewSecureStore()

	store.Save(label+"_private_key", privHex)
	store.Save(label+"_public_key", pubHex)
	store.Save(label+"_address", address)
	store.Save(label+"_chain", prefixStr)

	config.KeyLabels = append(config.KeyLabels, label)

	if config.ActiveKey == "" {
		config.ActiveKey = label
	}

	err = fecli.SaveConfig(config, configPath)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
