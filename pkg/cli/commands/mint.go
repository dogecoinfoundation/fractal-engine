package commands

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	fecli "dogecoin.org/fractal-engine/pkg/cli"
	climodels "dogecoin.org/fractal-engine/pkg/cli/climodels"
	"dogecoin.org/fractal-engine/pkg/cli/keys"
	fecfg "dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	bmclient "github.com/dogecoinfoundation/balance-master/pkg/client"
	"github.com/urfave/cli/v3"
)

var MintCommand = &cli.Command{
	Name:  "mints",
	Usage: "Manage the minting of tokens",
	Commands: []*cli.Command{
		{
			Name:   "create",
			Usage:  "Create a new token",
			Action: mintCreateAction,
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
			Usage:  "List all tokens",
			Action: mintListAction,
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

func mintListAction(ctx context.Context, cmd *cli.Command) error {
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

	tokenisationClient, err := getTokenisationClient(ctx, cmd)
	if err != nil {
		log.Fatal(err)
	}

	mints, err := tokenisationClient.GetMints(1, 10, pubHex, true)
	if err != nil {
		log.Fatal(err)
	}

	mintTable := climodels.CliTableModel{
		Table: table.New(
			table.WithColumns([]table.Column{
				{Title: "Hash", Width: 64},
				{Title: "Title", Width: 20},
				{Title: "Description", Width: 10},
				{Title: "Fraction Count", Width: 10},
				{Title: "Block Height", Width: 10},
				{Title: "Transaction Hash", Width: 10},
				{Title: "Created At", Width: 10},
				{Title: "Confirmed", Width: 10},
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
			fmt.Sprintf("%d", mint.BlockHeight),
			mint.TransactionHash.String,
			mint.CreatedAt.Format(time.RFC3339),
			fmt.Sprintf("%t", mint.TransactionHash.Valid),
		})
	}

	mintTable.Table.SetRows(rows)

	p := tea.NewProgram(mintTable)
	_, err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func mintCreateAction(ctx context.Context, cmd *cli.Command) error {
	var title string
	var fractionCount string
	var description string

	group := huh.NewGroup(
		huh.NewInput().
			Title("What is the Title of the token?").
			Value(&title),
		huh.NewInput().
			Title("What is the Fraction Count?").
			Value(&fractionCount),
		huh.NewInput().
			Title("What is the Description of the token?").
			Value(&description),
	)

	form := huh.NewForm(group)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	fractionCountInt, err := strconv.Atoi(fractionCount)
	if err != nil {
		log.Fatal(err)
	}

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
	pubHex, err := store.Get(config.ActiveKey + "_public_key")
	if err != nil {
		log.Fatal(err)
	}

	privHex, err := store.Get(config.ActiveKey + "_private_key")
	if err != nil {
		log.Fatal(err)
	}

	chain, err := store.Get(config.ActiveKey + "_chain")
	if err != nil {
		log.Fatal(err)
	}

	chainByte, err := doge.GetPrefix(chain)
	if err != nil {
		log.Fatal(err)
	}
	chainCfg := doge.GetChainCfg(chainByte)

	balanceMasterClient := bmclient.NewBalanceMasterClient(&bmclient.BalanceMasterClientConfig{
		RpcServerHost: config.BalanceMasterHost,
		RpcServerPort: config.BalanceMasterPort,
	})

	utxos, err := balanceMasterClient.GetUtxos(address)
	if err != nil {
		log.Fatal(err)
	}

	if len(utxos) == 0 {
		log.Fatal("No utxos found for address", address)
	}

	payload := rpc.CreateMintRequestPayload{
		Title:         title,
		FractionCount: fractionCountInt,
		Description:   description,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	signature, err := doge.SignPayload(payloadBytes, privHex)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("address", address)
	log.Println("pubHex", pubHex)
	log.Println("payload", payload)
	log.Println("signature", signature)

	mintResponse, err := tokenisationClient.Mint(&rpc.CreateMintRequest{
		Address:   address,
		PublicKey: pubHex,
		Payload:   payload,
		Signature: signature,
	})

	if err != nil {
		log.Fatal(err)
	}

	envelope := protocol.NewMintTransactionEnvelope(mintResponse.Hash, protocol.ACTION_MINT)
	encodedTransactionBody := envelope.Serialize()

	inputs := []interface{}{
		map[string]interface{}{
			"txid": utxos[0].TxID,
			"vout": utxos[0].VOut,
		},
	}

	log.Println("encodedTransactionBody", inputs)

	outputs := map[string]interface{}{
		"data":  hex.EncodeToString(encodedTransactionBody),
		address: utxos[0].Amount - 1,
	}

	dogeClient := doge.NewRpcClient(&fecfg.Config{
		DogeScheme:   config.DogeScheme,
		DogeHost:     config.DogeHost,
		DogePort:     config.DogePort,
		DogeUser:     config.DogeUser,
		DogePassword: config.DogePassword,
	})

	rawTx, err := dogeClient.Request("createrawtransaction", []interface{}{inputs, outputs})
	if err != nil {
		log.Fatal(err)
	}

	var rawTxResponse string
	if err := json.Unmarshal(*rawTx, &rawTxResponse); err != nil {
		log.Fatal(err)
	}

	encodedTx, err := doge.SignRawTransaction(rawTxResponse, privHex, []doge.PrevOutput{
		{
			Address: address,
			Amount:  int64(utxos[0].Amount),
		},
	}, chainCfg)

	if err != nil {
		log.Fatal(err)
	}

	res, err := dogeClient.Request("sendrawtransaction", []interface{}{encodedTx})
	if err != nil {
		log.Println("error sending raw transaction", err)
		return err
	}

	var txid string

	if err := json.Unmarshal(*res, &txid); err != nil {
		log.Println("error parsing send raw transaction response", err)
		return err
	}

	return nil
}
