package commands

import (
	"context"

	"github.com/urfave/cli/v3"
)

var OffersCommand = &cli.Command{
	Name:  "offers",
	Usage: "manage offers",
	Commands: []*cli.Command{
		{
			Name:   "create-sell-offer",
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
	},
}

func createSellOfferAction(ctx context.Context, cmd *cli.Command) error {

	return nil
}
