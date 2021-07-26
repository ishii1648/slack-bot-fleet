package example

import (
	"context"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	pb "github.com/ishii1648/slack-bot-fleet/proto/reaction-added-event"
)

type Server struct {
	pb.UnimplementedReactionServer
}

func (s *Server) Run(ctx context.Context, r *pb.Request) (*pb.Result, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debugf("revice request (reaction=%s, user=%s, item={%v})", r.Reaction, r.User, r.Item)

	return &pb.Result{Message: "ok"}, nil
}
