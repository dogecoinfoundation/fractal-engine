package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	fecli "dogecoin.org/fractal-engine/pkg/cli"
	climodels "dogecoin.org/fractal-engine/pkg/cli/climodels"
	"dogecoin.org/fractal-engine/pkg/cli/keys"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

var BuyOffersCommand = &cli.Command{
	Name:  "buy-offers",
	Usage: "manage buy offers",
	Commands: []*cli.Command{
		{
			Name:   "create",
			Usage:  "create a buy offer",
			Action: createBuyOfferAction,
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
			Usage:  "list buy offers",
			Action: listBuyOffersAction,
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

func listBuyOffersAction(ctx context.Context, cmd *cli.Command) error {
	tokenisationClient, err := getTokenisationClient(ctx, cmd)
	if err != nil {
		log.Fatal(err)
	}

	configPath := cmd.String("config-path")

	config, err := fecli.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	store := keys.NewSecureStore()
	address, err := store.Get(config.ActiveKey + "_address")
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

	buyOffers, err := tokenisationClient.GetBuyOffersBySellerAddress(1, 10, mintHash, address)
	if err != nil {
		log.Fatal(err)
	}

	items := []list.Item{}
	for _, buyOffer := range buyOffers.Offers {
		items = append(items, climodels.SimpleListItem{
			Name: "Mint: " + buyOffer.Mint.Title,
			Desc: "Price: " + strconv.Itoa(buyOffer.Offer.Price) + " Qty: " + strconv.Itoa(buyOffer.Offer.Quantity),
		})
	}

	m := climodels.SimpleListModel{List: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.List.Title = "Buy Offers"

	p := tea.NewProgram(m)
	_, err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func createBuyOfferAction(ctx context.Context, cmd *cli.Command) error {
	tokenisationClient, err := getTokenisationClient(ctx, cmd)
	if err != nil {
		log.Fatal(err)
	}

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
		items = append(items, climodels.SelectSimpleListItem{
			OfferId: sellOffer.Offer.Id,
			Name:    "Mint: " + sellOffer.Mint.Title + " (Seller: " + sellOffer.Offer.OffererAddress + ")",
			Desc:    "Price: " + strconv.Itoa(sellOffer.Offer.Price) + " Qty: " + strconv.Itoa(sellOffer.Offer.Quantity),
		})
	}

	m := climodels.SelectSimpleListModel{List: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.List.Title = "Sell Offers"

	p := tea.NewProgram(m)
	_, err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

	selectedOfferId := m.List.SelectedItem().(climodels.SelectSimpleListItem)

	var quantity string
	var price string

	group = huh.NewGroup(
		huh.NewInput().
			Title("What is the quantity?").
			Value(&quantity),
		huh.NewInput().
			Title("What is the price?").
			Value(&price),
	)

	form = huh.NewForm(group)
	err = form.Run()
	if err != nil {
		log.Fatal(err)
	}

	quantityInt, err := strconv.Atoi(quantity)
	if err != nil {
		log.Fatal(err)
	}
	priceInt, err := strconv.Atoi(price)
	if err != nil {
		log.Fatal(err)
	}

	var selectedOffer rpc.SellOfferWithMint
	for _, offer := range sellOffers.Offers {
		if offer.Offer.Id == selectedOfferId.OfferId {
			selectedOffer = offer
			break
		}
	}

	log.Println(selectedOffer.Offer.OffererAddress)

	buyOfferRequest := rpc.CreateBuyOfferRequest{
		Payload: rpc.CreateBuyOfferRequestPayload{
			OffererAddress: address,
			SellerAddress:  selectedOffer.Offer.OffererAddress,
			MintHash:       selectedOffer.Offer.MintHash,
			Quantity:       quantityInt,
			Price:          priceInt,
		},
	}

	buyOfferRequest.PublicKey = pubHex

	payloadBytes, err := json.Marshal(buyOfferRequest.Payload)
	if err != nil {
		log.Fatal(err)
	}

	signature, err := doge.SignPayload(payloadBytes, privHex)
	if err != nil {
		log.Fatal(err)
	}

	buyOfferRequest.Signature = signature

	response, err := tokenisationClient.CreateBuyOffer(&buyOfferRequest)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created buy offer: " + response.Id)

	return nil
}
