package main

import (
	"context"
	"os"

	"dogecoin.org/fractal-engine/pkg/cli/commands"
	"github.com/urfave/cli/v3"
)

func main() {
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
			commands.BmCommand,
			commands.DebugCommand,
		},
	}).Run(context.Background(), os.Args)
}
