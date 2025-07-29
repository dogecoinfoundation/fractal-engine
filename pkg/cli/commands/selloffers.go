package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	fecli "dogecoin.org/fractal-engine/pkg/cli"
	climodels "dogecoin.org/fractal-engine/pkg/cli/climodels"
	"dogecoin.org/fractal-engine/pkg/cli/keys"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

var SellOffersCommand = &cli.Command{
	Name:  "sell-offers",
	Usage: "manage sell offers",
	Commands: []*cli.Command{
		{
			Name:   "create",
			Usage:  "create a sell offer",
			Action: createSellOfferAction,
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
			Usage:  "list sell offers",
			Action: listSellOffersAction,
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

func listSellOffersAction(ctx context.Context, cmd *cli.Command) error {
	tokenisationClient, err := getTokenisationClient(ctx, cmd)
	if err != nil {
		log.Fatal(err)
	}

	var mintHash string

	group := huh.NewGroup(
		huh.NewInput().
			Title("What is the token hash?").
			Value(&mintHash),
	)

	form := huh.NewForm(group)
	err = form.Run()
	if err != nil {
		log.Fatal(err)
	}

	sellOffers, err := tokenisationClient.GetSellOffersByMintHash(1, 10, mintHash)
	if err != nil {
		log.Fatal(err)
	}

	items := []list.Item{}
	for _, sellOffer := range sellOffers.Offers {
		items = append(items, climodels.SimpleListItem{
			Name: "Mint: " + sellOffer.Mint.Title,
			Desc: "Price: " + strconv.Itoa(sellOffer.Offer.Price) + " Qty: " + strconv.Itoa(sellOffer.Offer.Quantity),
		})
	}

	m := climodels.SimpleListModel{List: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.List.Title = "Sell Offers"

	p := tea.NewProgram(m)
	_, err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func createSellOfferAction(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config-path")

	config, err := fecli.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	store := keys.NewSecureStore()
	pubHex, err := store.Get(config.ActiveKey + "_public_key")
	if err != nil {
		log.Fatal(err)
	}

	privHex, err := store.Get(config.ActiveKey + "_private_key")
	if err != nil {
		log.Fatal(err)
	}

	address, err := store.Get(config.ActiveKey + "_address")
	if err != nil {
		log.Fatal(err)
	}

	tokenisationClient, err := getTokenisationClient(ctx, cmd)
	if err != nil {
		log.Fatal(err)
	}

	mints, err := tokenisationClient.GetMints(1, 10, pubHex, false)
	if err != nil {
		log.Fatal(err)
	}

	mintTable := climodels.CliSelectTableModel{
		Table: table.New(
			table.WithColumns([]table.Column{
				{Title: "Hash", Width: 64},
				{Title: "Title", Width: 10},
				{Title: "Description", Width: 10},
				{Title: "Fraction Count", Width: 10},
				{Title: "Created At", Width: 10},
			}),
		),
	}

	rows := []table.Row{}

	for _, mint := range mints.Mints {
		rows = append(rows, table.Row{
			mint.Hash,
			mint.Title,
			mint.Description,
			fmt.Sprintf("%d", mint.FractionCount),
			mint.CreatedAt.Format(time.RFC3339),
		})
	}

	mintTable.Table.SetRows(rows)

	p := tea.NewProgram(mintTable)
	_, err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

	selectedRow := mintTable.Table.SelectedRow()

	if selectedRow == nil {
		log.Fatal("No row selected")
	}

	selectedMintId := selectedRow[0]

	var fractionCount string
	var price string

	group := huh.NewGroup(
		huh.NewInput().
			Title("How many fractions do you want to sell?").
			Value(&fractionCount),
		huh.NewInput().
			Title("What is the price per fraction?").
			Value(&price),
	)

	form := huh.NewForm(group)
	err = form.Run()
	if err != nil {
		log.Fatal(err)
	}

	fractionCountInt, err := strconv.Atoi(fractionCount)
	if err != nil {
		log.Fatal(err)
	}

	priceInt, err := strconv.Atoi(price)
	if err != nil {
		log.Fatal(err)
	}

	createSellOfferRequest := rpc.CreateSellOfferRequest{
		Payload: rpc.CreateSellOfferRequestPayload{
			OffererAddress: address,
			MintHash:       selectedMintId,
			Quantity:       fractionCountInt,
			Price:          priceInt,
		},
	}

	createSellOfferRequest.PublicKey = pubHex

	payloadBytes, err := json.Marshal(createSellOfferRequest.Payload)
	if err != nil {
		log.Fatal(err)
	}

	signature, err := doge.SignPayload(payloadBytes, privHex)
	if err != nil {
		log.Fatal(err)
	}

	createSellOfferRequest.Signature = signature

	createSellOfferResponse, err := tokenisationClient.CreateSellOffer(&createSellOfferRequest)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Sell offer created:", createSellOfferResponse)

	return nil
}
