package broker

import (
	"context"
	"encoding/json"
	"errors"
	pkghttp "net/http"
	"os"
	"time"

	"github.com/ishii1648/cloud-run-sdk/grpc"
	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	eventItem "github.com/ishii1648/slack-bot-fleet/pkg/event"
	pb "github.com/ishii1648/slack-bot-fleet/proto/reaction-added-event"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"go.opencensus.io/trace"
	pkggrpc "google.golang.org/grpc"
)

func Run(w pkghttp.ResponseWriter, r *pkghttp.Request) *http.Error {
	logger := zerolog.Ctx(r.Context())

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
		user, err := slackClient.GetUserInfoContext(ctx, event.User)
		if err != nil {
			return err
		}

		channel, err := slackClient.GetConversationInfoContext(ctx, event.Item.Channel, false)
		if err != nil {
			return err
		}

		logger.Infof("recieve ReactionAddedEvent(user : %s, channel: %s, reaction: %s)", user.RealName, channel.Name, event.Reaction)

		e := &ReactionAddedEvent{
			userRealName: user.RealName,
			channelName:  channel.Name,
			event:        event,
			logger:       logger,
		}

		if err := e.run(ctx); err != nil {
			return err
		}

		return nil
	}

	return nil
}

type ReactionAddedEvent struct {
	userRealName string
	channelName  string
	event        *slackevents.ReactionAddedEvent
	logger       *zerolog.Logger
}

func (e *ReactionAddedEvent) run(ctx context.Context) error {
	items, err := eventItem.ParseReactionAddedItem("./event.yaml")
	if err != nil {
		return err
	}

	item, req, err := e.getMatchedItem(ctx, items)
	if err != nil {
		return err
	}

	serviceAddr, isLocalhost, err := item.FetchServiceAddr(ctx)
	if err != nil {
		return err
	}
	e.logger.Debugf("get destination service address : %s", serviceAddr)

	if err := e.SendRequest(ctx, req, serviceAddr, isLocalhost); err != nil {
		return err
	}

	return nil
}

func (e *ReactionAddedEvent) getMatchedItem(ctx context.Context, items []eventItem.ReactionAddedItem) (*eventItem.ReactionAddedItem, *pb.Request, error) {
	sc := trace.FromContext(ctx).SpanContext()
	_, span := trace.StartSpanWithRemoteParent(ctx, "ReactionAddedEvent.getMatchedItem", sc)
	defer span.End()

	for _, item := range items {
		if ok := item.Match(e.userRealName, e.event.Reaction, e.channelName); ok {
			req := &pb.Request{
				Reaction: e.event.Reaction,
				User:     e.userRealName,
				Item: &pb.EventItem{
					Channel: e.channelName,
					Ts:      e.event.Item.Timestamp,
				},
			}
			return &item, req, nil
		}
	}

	return nil, nil, errors.New("no matched item")
}

func (e *ReactionAddedEvent) SendRequest(ctx context.Context, req *pb.Request, serviceAddr string, isLocalhost bool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var conn *pkggrpc.ClientConn
	var err error
	if isLocalhost {
		conn, err = pkggrpc.DialContext(ctx, serviceAddr)
	} else {
		conn, err = grpc.NewTLSConn(ctx, serviceAddr, pkggrpc.WithUnaryInterceptor(grpc.TraceIDInterceptor))
	}
	if err != nil {
		return err
	}
	defer conn.Close()

	c := pb.NewReactionClient(conn)

	resp, err := c.Run(ctx, req)
	if err != nil {
		return err
	}
	e.logger.Infof("received message from %s : %s", serviceAddr, resp.Message)

	return nil
}
