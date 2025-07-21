package commands

import (
	"context"
	"encoding/hex"
	"fmt"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/proto"
)

var DebugCommand = &cli.Command{
	Name:  "debug",
	Usage: "debug commands",
	Commands: []*cli.Command{
		{
			Name:  "decode-op-return",
			Usage: "decode op return",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Required: true,
					Name:     "payload",
					Usage:    "payload",
				},
			},
			Action: decodeOpReturnAction,
		},
	},
}

func decodeOpReturnAction(ctx context.Context, cmd *cli.Command) error {
	payload := cmd.String("payload")

	payloadBytes, err := hex.DecodeString(payload)
	if err != nil {
		return err
	}

	envelope := protocol.MessageEnvelope{}
	err = envelope.Deserialize(payloadBytes)
	if err != nil {
		return err
	}

	switch envelope.Action {
	case protocol.ACTION_MINT:
		mint := protocol.OnChainMintMessage{}
		err = proto.Unmarshal(envelope.Data, &mint)
		if err != nil {
			return err
		}
		fmt.Println("Action: Mint")
		fmt.Println("Hash:", mint.Hash)

	default:
		fmt.Println("Unknown action: ", envelope.Action)
	}

	return nil
}
