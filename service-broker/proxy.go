package broker

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	// m "cloud.google.com/go/compute/metadata"
	// pb "github.com/ishii1648/slack-bot-fleet/api/services/example"
	"github.com/ishii1648/slack-bot-fleet/service"
	"github.com/rs/zerolog"
	// "github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	// "google.golang.org/grpc"
	// "google.golang.org/grpc/credentials"
	// "google.golang.org/grpc/metadata"
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
		// We should response before request gRPC,
		// because Slack API resend request within 3 seconds.
		if _, err := w.Write([]byte("ok")); err != nil {
			l.Error().Msgf("failed to Write http.ResponseWriter: %v", err)
		}

		return nil
	case slackevents.AppRateLimited:
		return errors.New("app's event subscriptions are being rate limited")
	default:
		return nil
	}

	return nil
}

func proxyWithEvent(ctx context.Context, l *zerolog.Logger, eventsAPIEvent slackevents.EventsAPIEvent) error {
	switch event := eventsAPIEvent.InnerEvent.Data.(type) {
	case *slackevents.ReactionAddedEvent:
		services, err := service.ParseServicesYml("./service.yml")
		if err != nil {
			return err
		}

		s := NewSlack(os.Getenv("SLACK_BOT_TOKEN"), l)
		channel, err := s.c.GetConversationInfo(event.Item.Channel, false)
		if err != nil {
			return err
		}

		user, err := s.c.GetUserInfo(event.User)
		if err != nil {
			return err
		}

		ms, err := service.GetMicroService(services, channel.Name, event.Reaction, user.RealName)
		if err != nil {
			return err
		}

		if err := ms.Rpc(); err != nil {
			return err
		}

		return nil
	default:
		return nil
	}
}

// type Service struct {
// 	ctx  context.Context
// 	l    *zerolog.Logger
// 	item slackevents.Item
// }

// func (s *Service) proxy(reaction, eventUserID string) error {
// 	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

// 	channel, err := api.GetConversationInfo(s.item.Channel, false)
// 	if err != nil {
// 		return err
// 	}

// 	user, err := api.GetUserInfo(eventUserID)
// 	if err != nil {
// 		return err
// 	}

// 	if reaction == "raised_hands" && channel.Name == "development" && user.RealName == "しょーん" {
// 		if err := s.rpcWithChatbot(reaction, channel.ID); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (s *Service) rpcWithChatbot(reaction, channelID string) error {
// 	// msg, err := s.slack.FetchMsgByTs(s.item.Timestamp)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// s.l.Info().Msgf("message attached reaction : %s:%s", msg.Text, msg.User)

// 	r := &pb.Request{
// 		Reaction: reaction,
// 		Item: &pb.EventItem{
// 			ChannelID:    channelID,
// 			MsgTimestamp: s.item.Timestamp,
// 		},
// 	}

// 	chatbotAddr := os.Getenv("CHATBOT_ADDR")
// 	if chatbotAddr == "" {
// 		return errors.New("CHATBOT_ADDR is missing")
// 	}

// 	conn, err := newConn(chatbotAddr)
// 	if err != nil {
// 		return err
// 	}

// 	idToken, err := getIDToken(chatbotAddr)
// 	if err != nil {
// 		return err
// 	}

// 	c := pb.NewChatbotClient(conn)

// 	md := metadata.New(map[string]string{"authorization": fmt.Sprintf("Bearer %s", idToken)})
// 	ctx := metadata.NewOutgoingContext(s.ctx, md)

// 	result, err := c.Reply(ctx, r)
// 	if err != nil {
// 		return err
// 	}

// 	s.l.Info().Msgf("recieve result from ChatBot.Reply : %v", result)

// 	return nil
// }

// func newConn(addr string) (*grpc.ClientConn, error) {
// 	systemRoots, err := x509.SystemCertPool()
// 	if err != nil {
// 		return nil, err
// 	}

// 	cred := credentials.NewTLS(&tls.Config{
// 		RootCAs: systemRoots,
// 	})

// 	conn, err := grpc.Dial(addr, grpc.WithAuthority(addr), grpc.WithTransportCredentials(cred))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return conn, nil
// }

// func getIDToken(addr string) (string, error) {
// 	serviceURL := fmt.Sprintf("https://%s", strings.Split(addr, ":")[0])
// 	tokenURL := fmt.Sprintf("/instance/service-accounts/default/identity?audience=%s", serviceURL)

// 	idToken, err := m.Get(tokenURL)
// 	if err != nil {
// 		return "", err
// 	}

// 	return idToken, nil
// }
