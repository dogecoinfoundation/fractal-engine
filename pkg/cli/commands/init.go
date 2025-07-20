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
	host := "localhost"
	port := "8891"
	balanceMasterHost := "localhost"
	balanceMasterPort := "8899"
	dogeScheme := "http"
	dogeHost := "localhost"
	dogePort := "22556"
	dogeUser := "test"
	dogePassword := "test"

	group := huh.NewGroup(
		huh.NewInput().
			Title("What is the Fractal Engine Host?").
			Value(&host),
		huh.NewInput().
			Title("What is the Fractal Engine Port?").
			Value(&port),
		huh.NewInput().
			Title("What is the Balance Master Host?").
			Value(&balanceMasterHost),
		huh.NewInput().
			Title("What is the Balance Master Port?").
			Value(&balanceMasterPort),
		huh.NewInput().
			Title("What is the Dogecoin Scheme?").
			Value(&dogeScheme),
		huh.NewInput().
			Title("What is the Dogecoin Host?").
			Value(&dogeHost),
		huh.NewInput().
			Title("What is the Dogecoin Port?").
			Value(&dogePort),
		huh.NewInput().
			Title("What is the Dogecoin User?").
			Value(&dogeUser),
		huh.NewInput().
			Title("What is the Dogecoin Password?").
			Value(&dogePassword),
	)

	form := huh.NewForm(group)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	config := fecli.Config{
		FractalEngineHost: host,
		FractalEnginePort: port,
		BalanceMasterHost: balanceMasterHost,
		BalanceMasterPort: balanceMasterPort,
		DogeScheme:        dogeScheme,
		DogeHost:          dogeHost,
		DogePort:          dogePort,
		DogeUser:          dogeUser,
		DogePassword:      dogePassword,
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

	if err == nil || confirm {
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
