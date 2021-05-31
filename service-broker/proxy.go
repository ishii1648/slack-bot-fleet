package broker

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/ishii1648/slack-bot-fleet/pkg/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack/slackevents"
)

func Proxy(ctx context.Context, w http.ResponseWriter, r *http.Request, localForwardFlag bool) error {
	logger := log.Ctx(r.Context())

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		return err
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		if err := proxyWithURLVerification(w, body); err != nil {
			return err
		}
	case slackevents.CallbackEvent:
		if err := proxyWithEvent(ctx, logger, w, eventsAPIEvent, localForwardFlag); err != nil {
			return err
		}
		return nil
	case slackevents.AppRateLimited:
		return errors.New("app's event subscriptions are being rate limited")
	}

	return nil
}

func proxyWithURLVerification(w http.ResponseWriter, body []byte) error {
	var res *slackevents.ChallengeResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte(res.Challenge)); err != nil {
		return err
	}

	return nil
}

func proxyWithEvent(ctx context.Context, l *zerolog.Logger, w http.ResponseWriter, eventsAPIEvent slackevents.EventsAPIEvent, localForwardFlag bool) error {
	// We should response before request gRPC,
	// because Slack API resend request within 3 seconds.
	if _, err := w.Write([]byte("ok")); err != nil {
		l.Error().Msgf("failed to Write http.ResponseWriter: %v", err)
	}

	return service.Do(ctx, l, eventsAPIEvent, localForwardFlag)
}
