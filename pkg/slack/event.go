package slack

import (
// "github.com/slack-go/slack"
// "github.com/slack-go/slack/slackevents"
)

type Event interface {
}

type ReactionAddedEvent struct {
	userRealName string
	channelName  string
	projectID    string
	locationID   string
}
