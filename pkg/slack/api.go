package slack

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Slack struct {
	c *slack.Client
	l *zerolog.Logger
}

func NewSlack(slackBotToken string, l *zerolog.Logger) *Slack {
	return &Slack{
		c: slack.New(slackBotToken),
		l: l,
	}
}

func (s *Slack) FetchRepositoryNameByTs(item *slackevents.Item) (string, error) {
	param := &slack.GetConversationHistoryParameters{
		ChannelID: item.Channel,
		Latest:    item.Timestamp,
		Limit:     1,
		Inclusive: true,
	}

	res, err := s.c.GetConversationHistory(param)
	if err != nil {
		return "", err
	}

	textSlice := strings.Split(res.Messages[0].Text, " ")

	switch len(textSlice) {
	case 1:
		return textSlice[0], nil
	case 2:
		return textSlice[1], nil
	default:
		return "", fmt.Errorf("invalid input : %s", res.Messages[0].Text)
	}
}
