package service

import (
	"context"
	"errors"
	"os"

	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type serviceName string

func (s serviceName) isService(criteriaServiceName string) bool {
	return s == serviceName(criteriaServiceName)
}

func Do(ctx context.Context, l *zerolog.Logger, eventsAPIEvent slackevents.EventsAPIEvent, localForwardFlag bool) error {
	s := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	switch event := eventsAPIEvent.InnerEvent.Data.(type) {
	case *slackevents.ReactionAddedEvent:
		l.Debug().Msgf("user : %s, channel: %s, reaction: %s", event.User, event.Item.Channel, event.Reaction)
		return doWithReactionEvent(ctx, s, l, event, localForwardFlag)
	}

	return errors.New("micro service is missing")
}

var reactionEventServiceFactories = make(map[serviceName]reactionEventServiceFactory)

type reactionEventServiceFactory func(l *zerolog.Logger, addr string) ReactionEventService

type ReactionEventService interface {
	RPC(ctx context.Context, event *slackevents.ReactionAddedEvent, user, reaction, channelID string) error
}

func registerReactionEventService(name serviceName, factory reactionEventServiceFactory) {
	reactionEventServiceFactories[name] = factory
}

func doWithReactionEvent(ctx context.Context, s *slack.Client, l *zerolog.Logger, event *slackevents.ReactionAddedEvent, localForwardFlag bool) error {
	criterion, err := ParseReactionEventCriteriaYml("./criteria.yml")
	if err != nil {
		return err
	}

	for _, criteria := range criterion {
		user, err := s.GetUserInfo(event.User)
		if err != nil {
			return err
		}

		channel, err := s.GetConversationInfo(event.Item.Channel, false)
		if err != nil {
			return err
		}

		if ok := criteria.Match(user.RealName, event.Reaction, channel.Name); ok {
			service, err := criteria.GetService(ctx, l, reactionEventServiceFactories, localForwardFlag)
			if err != nil {
				return err
			}

			return service.RPC(ctx, event, user.RealName, event.Reaction, channel.ID)
		}

	}

	return errors.New("criteria is missing")
}
