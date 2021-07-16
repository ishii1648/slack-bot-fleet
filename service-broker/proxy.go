package broker

import (
	"context"
	"encoding/json"
	"errors"
	pkghttp "net/http"
	"os"

	"github.com/ishii1648/cloud-run-sdk/grpc"
	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	eventItem "github.com/ishii1648/slack-bot-fleet/pkg/event"
	pb "github.com/ishii1648/slack-bot-fleet/proto/reaction-added-event"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func Run(w pkghttp.ResponseWriter, r *pkghttp.Request) *http.Error {
	body, ok := r.Context().Value("requestBody").([]byte)
	if !ok {
		return &http.Error{
			Error:   errors.New("requestBody not found"),
			Message: "requestBody not found",
			Code:    pkghttp.StatusBadRequest,
		}
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		return &http.Error{
			Error:   err,
			Message: "invalid request",
			Code:    pkghttp.StatusBadRequest,
		}
	}

	logger := zerolog.NewLogger(log.Ctx(r.Context()))

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		if err := responseURLVerification(logger, w, body); err != nil {
			return &http.Error{
				Error:   err,
				Message: "failed to URLVerification",
				Code:    pkghttp.StatusInternalServerError,
			}
		}
	case slackevents.CallbackEvent:
		// We should response before request gRPC,
		// because Slack API resend request within 3 seconds.
		if _, err := w.Write([]byte("ok")); err != nil {
			logger.Errorf("failed to Write http.ResponseWriter: %v", err)
			return nil
		}

		if err := proxy(r.Context(), logger, eventsAPIEvent); err != nil {
			logger.Errorf("failed to proxy: %v", err)
		}

		return nil
	case slackevents.AppRateLimited:
		return &http.Error{
			Error:   errors.New("app's event subscriptions are being rate limited"),
			Message: "app's event subscriptions are being rate limited",
			Code:    pkghttp.StatusBadRequest,
		}
	}

	return nil
}

func responseURLVerification(logger *zerolog.Logger, w pkghttp.ResponseWriter, body []byte) error {
	var res *slackevents.ChallengeResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte(res.Challenge)); err != nil {
		logger.Errorf("failed to Write res.Challange: %v", err)
		return nil
	}

	return nil
}

func proxy(ctx context.Context, logger *zerolog.Logger, eventsAPIEvent slackevents.EventsAPIEvent) error {
	slackBotToken, isSet := os.LookupEnv("SLACK_BOT_TOKEN")
	if !isSet {
		return errors.New("SLACK_BOT_TOKEN is not set")
	}

	slackClient := slack.New(slackBotToken)

	switch event := eventsAPIEvent.InnerEvent.Data.(type) {
	case *slackevents.ReactionAddedEvent:
		logger.Infof("recieve ReactionAddedEvent(user : %s, channel: %s, reaction: %s)", event.User, event.Item.Channel, event.Reaction)

		e := &ReactionAddedEvent{
			event:  event,
			logger: logger,
			slack:  slackClient,
		}

		if err := e.run(ctx); err != nil {
			return err
		}

		return nil
	}

	return nil
}

type ReactionAddedEvent struct {
	event  *slackevents.ReactionAddedEvent
	logger *zerolog.Logger
	slack  *slack.Client
}

func (e *ReactionAddedEvent) run(ctx context.Context) error {
	items, err := eventItem.ParseReactionAddedItem("./event.yaml")
	if err != nil {
		return err
	}

	matchedItem, err := e.getMatchedItem(e.slack, items)
	if err != nil {
		return err
	}

	serviceAddr, isLocalhost, err := matchedItem.GetServiceAddr(ctx)
	if err != nil {
		return err
	}
	e.logger.Debugf("get service address : %s", serviceAddr)

	if err := e.SendRequest(ctx, serviceAddr, isLocalhost); err != nil {
		return err
	}

	return nil
}

func (e *ReactionAddedEvent) getMatchedItem(s *slack.Client, items []eventItem.ReactionAddedItem) (*eventItem.ReactionAddedItem, error) {
	for _, item := range items {
		user, err := s.GetUserInfo(e.event.User)
		if err != nil {
			return nil, err
		}

		channel, err := s.GetConversationInfo(e.event.Item.Channel, false)
		if err != nil {
			return nil, err
		}

		if ok := item.Match(user.RealName, e.event.Reaction, channel.Name); ok {
			return &item, nil
		}
	}

	return nil, errors.New("no matched item")
}

func (e *ReactionAddedEvent) SendRequest(ctx context.Context, serviceAddr string, isLocalhost bool) error {
	r := &pb.Request{
		Reaction: e.event.Reaction,
		User:     e.event.User,
		Item: &pb.EventItem{
			Channel: e.event.Item.Channel,
			Ts:      e.event.Item.Timestamp,
		},
	}

	conn, err := grpc.NewConn(serviceAddr, isLocalhost)
	if err != nil {
		return err
	}

	c := pb.NewReactionClient(conn)

	if _, err := c.Run(ctx, r); err != nil {
		return err
	}

	return nil
}
