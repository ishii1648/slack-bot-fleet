package broker

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
)

func Reply(l *zerolog.Logger, msg, channelID string) error {
	if msg == "" {
		l.Info().Msg("reply message is none")
		return nil
	}

	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	_, _, err := api.PostMessage(channelID, slack.MsgOptionText(msg, false))
	if err != nil {
		return err
	}

	return nil
}
