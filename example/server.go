package example

import (
	"context"
	"errors"
	"fmt"
	"os"

	pb "github.com/ishii1648/slack-bot-fleet/api/services/example"
	"github.com/ishii1648/slack-bot-fleet/pkg/service"
	"github.com/slack-go/slack"

	// "github.com/ishii1648/slack-bot-fleet/pkg/slack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Server struct {
	l *zerolog.Logger
	s *slack.Client
	pb.UnimplementedExampleServer
}

func (s *Server) Reply(ctx context.Context, r *pb.Request) (*pb.Result, error) {
	s.l = log.Ctx(ctx)
	s.l.Info().Msgf("revice request : %v", r)

	s.s = slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	if err := s.verifyRequest(r.User, r.Item.Channel, r.Reaction); err != nil {
		return nil, fmt.Errorf("failed to verify : %v", err)
	}

	return &pb.Result{Message: "hello"}, nil
}

func (s *Server) verifyRequest(userRealName, channelName, reaction string) error {
	criteriaList, err := service.ParseReactionEventCriteriaListYml("./criteria.yml")
	if err != nil {
		return err
	}

	for _, criteria := range criteriaList {
		if ok := criteria.Match(userRealName, reaction, channelName); ok {
			return nil
		}
	}

	return errors.New("no match service")
}
