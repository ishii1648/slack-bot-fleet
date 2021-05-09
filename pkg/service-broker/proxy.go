package broker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	m "cloud.google.com/go/compute/metadata"
	pb "github.com/ishii1648/slack-bot-fleet/api/services/chatbot"
	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func Proxy(ctx context.Context, l *zerolog.Logger, w http.ResponseWriter, body []byte) error {
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		return err
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		var res *slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &res); err != nil {
			return err
		}

		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte(res.Challenge)); err != nil {
			return err
		}
	case slackevents.CallbackEvent:
		innerEvent := eventsAPIEvent.InnerEvent
		switch event := innerEvent.Data.(type) {
		case *slackevents.ReactionAddedEvent:
			api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))
			channel, err := api.GetConversationInfo(event.Item.Channel, false)
			if err != nil {
				return err
			}
			if event.Reaction == "raised_hands" && channel.Name == "development" {
				r := &pb.Request{
					Reaction: event.Reaction,
					Item: &pb.EventItem{
						Channel:      channel.Name,
						MsgTimestamp: event.Item.Timestamp,
					},
				}

				chatbotAddr := os.Getenv("CHATBOT_ADDR")
				if chatbotAddr == "" {
					return errors.New("CHATBOT_ADDR is missing")
				}

				conn, err := newConn(chatbotAddr)
				if err != nil {
					return err
				}

				idToken, err := getIDToken(chatbotAddr)
				if err != nil {
					return err
				}

				c := pb.NewChatbotClient(conn)

				md := metadata.New(map[string]string{"authorization": fmt.Sprintf("Bearer %s", idToken)})
				ctx = metadata.NewOutgoingContext(ctx, md)

				result, err := c.Reply(ctx, r)
				if err != nil {
					return err
				}
				l.Info().Msgf("result : %v", result)
			}
		}
		return nil
	case slackevents.AppRateLimited:
		l.Error().Msg("app's event subscriptions are being rate limited")
		return nil
	default:
		return nil
	}

	return nil
}

func newConn(addr string) (*grpc.ClientConn, error) {
	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	cred := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})

	conn, err := grpc.Dial(addr, grpc.WithAuthority(addr), grpc.WithTransportCredentials(cred))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func getIDToken(addr string) (string, error) {
	serviceURL := fmt.Sprintf("https://%s", strings.Split(addr, ":")[0])
	tokenURL := fmt.Sprintf("/instance/service-accounts/default/identity?audience=%s", serviceURL)

	idToken, err := m.Get(tokenURL)
	if err != nil {
		return "", err
	}

	return idToken, nil
}
