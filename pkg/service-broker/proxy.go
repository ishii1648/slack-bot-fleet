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
	"github.com/ishii1648/slack-bot-fleet/internal/slack"
	s "github.com/ishii1648/slack-bot-fleet/internal/slack"
	"github.com/rs/zerolog"

	"github.com/slack-go/slack/slackevents"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func Proxy(ctx context.Context, l *zerolog.Logger, w http.ResponseWriter, body []byte) (msg, channelID string, err error) {
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		return "", "", err
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		var res *slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &res); err != nil {
			return "", "", err
		}

		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte(res.Challenge)); err != nil {
			return "", "", err
		}
	case slackevents.CallbackEvent:
		innerEvent := eventsAPIEvent.InnerEvent
		switch event := innerEvent.Data.(type) {
		case *slackevents.ReactionAddedEvent:
			slack := s.NewSlack(os.Getenv("SLACK_BOT_TOKEN"), event.Item, "")

			service := &Service{
				ctx:   ctx,
				l:     l,
				slack: slack,
				item:  event.Item,
			}

			msg, err := service.rpc(event.Reaction, event.User)
			if err != nil {
				return "", event.Item.Channel, err
			}

			return msg, event.Item.Channel, nil
		default:
			return "", "", nil
		}
	case slackevents.AppRateLimited:
		return "", "", errors.New("app's event subscriptions are being rate limited")
	default:
		return "", "", nil
	}

	return "", "", nil
}

type Service struct {
	ctx   context.Context
	l     *zerolog.Logger
	slack *slack.Slack
	item  slackevents.Item
}

func (s *Service) rpc(reaction, eventUserID string) (string, error) {
	channelName, err := s.slack.FetchChannelName()
	if err != nil {
		return "", err
	}

	_, eventUserRealName, err := s.slack.FetchUserNameByID(eventUserID)
	if err != nil {
		return "", err
	}

	var result *pb.Result

	if reaction == "raised_hands" && channelName == "development" && eventUserRealName == "しょーん" {
		result, err = s.rpcWithChatbot(reaction, channelName)
		if err != nil {
			return "", err
		}
	}

	return result.Message, nil
}

func (s *Service) rpcWithChatbot(reaction, channelName string) (*pb.Result, error) {
	msg, err := s.slack.FetchMsgByTs(s.item.Timestamp)
	if err != nil {
		return nil, err
	}

	s.l.Info().Msgf("message attached reaction : %s:%s", msg.Text, msg.User)

	r := &pb.Request{
		Reaction: reaction,
		Item: &pb.EventItem{
			Channel:      channelName,
			MsgTimestamp: s.item.Timestamp,
		},
	}

	chatbotAddr := os.Getenv("CHATBOT_ADDR")
	if chatbotAddr == "" {
		return nil, errors.New("CHATBOT_ADDR is missing")
	}

	conn, err := newConn(chatbotAddr)
	if err != nil {
		return nil, err
	}

	idToken, err := getIDToken(chatbotAddr)
	if err != nil {
		return nil, err
	}

	c := pb.NewChatbotClient(conn)

	md := metadata.New(map[string]string{"authorization": fmt.Sprintf("Bearer %s", idToken)})
	ctx := metadata.NewOutgoingContext(s.ctx, md)

	result, err := c.Reply(ctx, r)
	if err != nil {
		return nil, err
	}

	return result, nil
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
