package dogenet

import (
	"database/sql"
	"log"

	"code.dogecoin.org/gossip/dnet"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *DogeNetClient) GossipMint(record store.Mint) error {
	feedUrl := ""
	if record.FeedURL != nil {
		feedUrl = *record.FeedURL
	}

	mintMessage := protocol.MintMessage{
		Id:              record.Id,
		Title:           record.Title,
		Description:     record.Description,
		FractionCount:   int32(record.FractionCount),
		Tags:            record.Tags,
		TransactionHash: record.TransactionHash.String,
		Metadata:        &structpb.Struct{Fields: convertToStructPBMap(record.Metadata)},
		Hash:            record.Hash,
		Requirements:    &structpb.Struct{Fields: convertToStructPBMap(record.Requirements)},
		LockupOptions:   &structpb.Struct{Fields: convertToStructPBMap(record.LockupOptions)},
		FeedUrl:         feedUrl,
		CreatedAt:       timestamppb.New(record.CreatedAt),
	}

	envelope := protocol.MintMessageEnvelope{
		Type:    protocol.ACTION_MINT,
		Version: protocol.DEFAULT_VERSION,
		Payload: &mintMessage,
	}

	data, err := proto.Marshal(&envelope)
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	encodedMsg := dnet.EncodeMessageRaw(ChanFE, TagMint, c.feKey, data)

	err = encodedMsg.Send(c.sock)
	if err != nil {
		return err
	}

	return nil
}

func (c *DogeNetClient) recvMint(msg dnet.Message) {
	log.Printf("[FE] received mint message")

	envelope := protocol.MintMessageEnvelope{}
	err := proto.Unmarshal(msg.Payload, &envelope)
	if err != nil {
		log.Println("Error deserializing message envelope:", err)
		return
	}

	if envelope.Type != protocol.ACTION_MINT {
		log.Printf("[FE] unexpected action: [%s][%s][%d]", msg.Chan, msg.Tag, envelope.Type)
		return
	}

	mint := envelope.Payload

	id, err := c.store.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash:            mint.Hash,
		Title:           mint.Title,
		FractionCount:   int(mint.FractionCount),
		Description:     mint.Description,
		Tags:            mint.Tags,
		Metadata:        mint.Metadata.AsMap(),
		TransactionHash: sql.NullString{String: mint.TransactionHash, Valid: true},
		CreatedAt:       mint.CreatedAt.AsTime(),
		Requirements:    mint.Requirements.AsMap(),
		LockupOptions:   mint.LockupOptions.AsMap(),
	})

	if err != nil {
		log.Println("Error saving unconfirmed mint:", err)
		return
	}

	log.Printf("[FE] unconfirmed mint saved: %v", id)
}
