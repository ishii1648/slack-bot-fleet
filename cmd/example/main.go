package main

import (
	"context"
	"net"
	"os"

	pb "github.com/ishii1648/slack-bot-fleet/api/services/chatbot"
	// "github.com/ishii1648/slack-bot-fleet/pkg/crzerolog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedChatbotServer
}

func (s *server) Reply(ctx context.Context, r *pb.Request) (*pb.Result, error) {
	l := log.Ctx(ctx)
	l.Info().Msgf("revice request : %v", r)
	return &pb.Result{Message: "hello"}, nil
}

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal().Msgf("Failed to listen: %v", err)
	}

	rootLogger := zerolog.New(os.Stdout)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(crzerolog.InjectLoggerInterceptor(&rootLogger)),
	)
	reflection.Register(s)
	pb.RegisterChatbotServer(s, &server{})
	if err := s.Serve(l); err != nil {
		log.Fatal().Msgf("Failed to serve: %v", err)
	}
}
