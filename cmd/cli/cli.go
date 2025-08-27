package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"dogecoin.org/fractal-engine/pkg/cli/commands"
	"github.com/urfave/cli/v3"
)

func main() {

	rawUrl := "postgres://fractal_engine:iQqnR,TcgYHzlQ.Yute0Ym-M53@databasestack-fractaldbe0e3850a-z0lqu2lpdngx.c4zeeksug1mc.us-east-1.rds.amazonaws.com:5432/fractal"
	parsedURL, err := url.Parse(rawUrl)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}

	fmt.Println("parsedURL", parsedURL)

	(&cli.Command{
		Name:      "Fractal Engine CLI",
		Usage:     "fecli",
		UsageText: "fecli [command]",
		Commands: []*cli.Command{
			commands.InitCommand,
			commands.KeysCommand,
			commands.HealthCommand,
			commands.MintCommand,
			commands.DemoCommand,
			commands.IndexerCommand,
			commands.DebugCommand,
			commands.SellOffersCommand,
			commands.BuyOffersCommand,
			commands.InvoiceCommand,
			commands.PaymentsCommand,
			commands.TokensCommand,
		},
	}).Run(context.Background(), os.Args)
}
