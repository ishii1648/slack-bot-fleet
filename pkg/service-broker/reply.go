package broker

import (
	"os"

	s "github.com/ishii1648/slack-bot-fleet/internal/slack"
	"github.com/rs/zerolog"
	"github.com/slack-go/slack/slackevents"
)

func Reply(l *zerolog.Logger, msg, channelID string) error {
	if msg == "" {
		l.Info().Msg("reply message is none")
		return nil
	}

	slack := s.NewSlack(os.Getenv("SLACK_BOT_TOKEN"), slackevents.Item{}, channelID)
	if err := slack.PostMsg(msg); err != nil {
		return err
	}

	return nil
}
