package commands

import (
	"context"
	"fmt"
	"log"
	"time"

	fecli "dogecoin.org/fractal-engine/pkg/cli"
	"dogecoin.org/fractal-engine/pkg/client"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v3"
)

var HealthCommand = &cli.Command{
	Name:   "health",
	Usage:  "Check the health of the fractal engine",
	Action: healthAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "config-path",
			Usage: "Path to the config file",
			Value: "config.toml",
		},
	},
}

func healthAction(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config-path")

	config, err := fecli.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	url := fmt.Sprintf("http://%s:%s", config.FractalEngineHost, config.FractalEnginePort)

	tokenisationClient := client.NewTokenisationClient(url, "", "")

	health, err := tokenisationClient.GetHealth()
	if err != nil {
		log.Fatal(err)
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 1).
		Width(40).Align(lipgloss.Left)

	bold := lipgloss.NewStyle().Bold(true)

	fmt.Println(style.Render("Fractal Engine Health"))
	fmt.Println(style.Render("--------------------------------"))
	fmt.Println(style.Render("Chain: ") + bold.Render(health.Chain))
	fmt.Println(style.Render("Current Block Height: ") + bold.Render(fmt.Sprintf("%d", health.CurrentBlockHeight)))
	fmt.Println(style.Render("Latest Block Height: ") + bold.Render(fmt.Sprintf("%d", health.LatestBlockHeight)))
	fmt.Println(style.Render("Wallets Enabled: ") + bold.Render(fmt.Sprintf("%t", health.WalletsEnabled)))
	fmt.Println(style.Render("Updated At: ") + bold.Render(health.UpdatedAt.Format(time.RFC3339)))

	return nil
}
