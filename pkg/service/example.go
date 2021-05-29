package service

import (
	"context"

	pb "github.com/ishii1648/slack-bot-fleet/api/services/example"
	sdk "github.com/ishii1648/slack-bot-fleet/pkg/cloud-run-sdk"
	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Example struct {
	logger *zerolog.Logger
	s      *slack.Client
	addr   string
}

func init() {
	registerReactionEventService(serviceName("example"), newExample)
}

func newExample(l *zerolog.Logger, addr string) ReactionEventService {
	return &Example{
		logger: l,
		addr:   addr,
	}
}

func (e *Example) RPC(ctx context.Context, event *slackevents.ReactionAddedEvent, user, reaction, channelID string) error {
	conn, err := sdk.NewTLSConn(e.addr)
	if err != nil {
		return err
	}

	c := pb.NewExampleClient(conn)

	r := &pb.Request{
		Reaction: reaction,
		Item: &pb.EventItem{
			Channel: channelID,
			Ts:      event.Item.Timestamp,
		},
	}

	if _, err := c.Reply(ctx, r); err != nil {
		return err
	}

	return nil
}
