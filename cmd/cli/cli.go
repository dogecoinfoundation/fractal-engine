package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"dogecoin.org/fractal-engine/pkg/cli/commands"
	"github.com/dogeorg/doge"
	"github.com/urfave/cli/v3"
)

func main() {

	pubkeyHash, _ := doge.Base58DecodeCheck("mtscddS96MYf83q14BxGu7AKwscSYrL7S2")
	fmt.Println(hex.EncodeToString(pubkeyHash[1:]))

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
