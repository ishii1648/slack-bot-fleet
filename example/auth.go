package example

import (
	"context"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/slack-bot-fleet/pkg/event"
	pb "github.com/ishii1648/slack-bot-fleet/proto/reaction-added-event"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func VerifyRequestInterceptor(eventYamlPath string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger := zerolog.Ctx(ctx)

		r, ok := req.(*pb.Request)
		if !ok {
			logger.Error("failed to cast request")
			return nil, status.Error(codes.InvalidArgument, "failed to cast request")
		}

		items, err := event.ParseReactionAddedItem(eventYamlPath)
		if err != nil {
			logger.Errorf("failed parse event.yaml : %v", err)
			return nil, status.Error(codes.Internal, "failed parse event.yaml")
		}

		var matched bool
		for _, item := range items {
			if ok := item.Match(r.User, r.Reaction, r.Item.Channel); ok {
				matched = true
			}
		}

		if !matched {
			logger.Errorf("no matched item (reaction=%s, user=%s, item={%v})", r.Reaction, r.User, r.Item)
			return nil, status.Error(codes.InvalidArgument, "no matched item")
		}

		return handler(ctx, req)
	}
}
