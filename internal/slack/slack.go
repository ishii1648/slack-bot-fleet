package slack

import (
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Slack struct {
	c         *slack.Client
	item      slackevents.Item
	channelID string
}

type Message struct {
	User string `json:"user,omitempty"`
	Text string `json:"text,omitempty"`
}

func NewSlack(botToken string, item slackevents.Item, channelID string) *Slack {
	return &Slack{
		c:         slack.New(botToken),
		item:      item,
		channelID: channelID,
	}
}

func (s *Slack) FetchChannelName() (string, error) {
	channel, err := s.c.GetConversationInfo(s.item.Channel, false)
	if err != nil {
		return "", err
	}
	return channel.Name, nil
}

func (s *Slack) FetchUserNameByID(userID string) (name, realName string, err error) {
	user, err := s.c.GetUserInfo(userID)
	if err != nil {
		return "", "", err
	}

	return user.Name, user.RealName, nil
}

func (s *Slack) FetchMsgByTs(ts string) (*Message, error) {
	param := &slack.GetConversationHistoryParameters{
		ChannelID: s.item.Channel,
		Latest:    ts,
		Limit:     1,
		Inclusive: true,
	}
	res, err := s.c.GetConversationHistory(param)
	if err != nil {
		return nil, err
	}

	m := res.Messages[0]

	return &Message{
		User: m.User,
		Text: m.Text,
	}, nil
}

func (s *Slack) PostMsg(msg string) error {
	_, _, err := s.c.PostMessage(s.channelID, slack.MsgOptionText(msg, false))
	if err != nil {
		return err
	}

	return nil
}
