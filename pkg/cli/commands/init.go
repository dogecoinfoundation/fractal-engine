package commands

import (
	"context"
	"fmt"
	"log"

	fecli "dogecoin.org/fractal-engine/pkg/cli"
	climodels "dogecoin.org/fractal-engine/pkg/cli/climodels"
	"dogecoin.org/fractal-engine/pkg/client"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

var InitCommand = &cli.Command{
	Name:   "init",
	Usage:  "Creates a fractal engine config file",
	Action: initAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "config-path",
			Usage: "Path to the config file",
			Value: "config.toml",
		},
	},
}

func initAction(ctx context.Context, cmd *cli.Command) error {
	var host string
	var port string

	group := huh.NewGroup(
		huh.NewInput().
			Title("What is the Fractal Engine Host?").
			Value(&host),
		huh.NewInput().
			Title("What is the Fractal Engine Port?").
			Value(&port),
	)

	form := huh.NewForm(group)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	config := fecli.Config{
		FractalEngineHost: host,
		FractalEnginePort: port,
	}

	url := fmt.Sprintf("http://%s:%s", config.FractalEngineHost, config.FractalEnginePort)
	feClient := client.NewTokenisationClient(url, "", "")

	spinner := climodels.NewSpinner()
	p := tea.NewProgram(spinner)
	errorChan := make(chan error, 1)

	go func() {
		_, err = feClient.GetHealth()
		errorChan <- err
		p.Send(climodels.SpinnerDoneMsg{Error: err})
	}()

	_, err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

	err = <-errorChan
	var confirm bool
	if err != nil {
		confirmer := huh.NewConfirm().
			Title("Unable to connect to Fractal Engine, save configuration anyway?").
			Affirmative("Yes!").
			Negative("No.").
			Value(&confirm)

		err = confirmer.Run()
		if err != nil {
			log.Fatal(err)
		}
	}

	if err != nil || confirm {
		err = fecli.SaveConfig(&config, cmd.String("config-path"))
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Config saved to", cmd.String("config-path"))
	} else {
		log.Println("Config not saved")
	}

	return nil
}
