package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	pb "github.com/ishii1648/slack-bot-fleet/api/services/example"
	"github.com/ishii1648/slack-bot-fleet/example"
	sdk "github.com/ishii1648/slack-bot-fleet/pkg/cloud-run-sdk"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := sdk.SetLogger(zerolog.New(os.Stdout))

	srv, listener, err := RegisterDefaultGRPCServer(logger)
	if err != nil {
		logger.Fatal().Msgf("failed to RegisterDefaultGRPCServer : %v", err)
	}

	go func() {
		if err := srv.Serve(listener); err != nil {
			logger.Error().Msgf("server closed with error : %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	<-sigCh
	logger.Info().Msg("recive SIGTERM or SIGINT")

	srv.GracefulStop()
	logger.Info().Msg("gRPC Server shutdowned")
}

func RegisterDefaultGRPCServer(logger zerolog.Logger) (*grpc.Server, net.Listener, error) {
	hostAddr := "0.0.0.0"
	if h := os.Getenv("HOST_ADDR"); h != "" {
		hostAddr = h
	}

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", hostAddr, sdk.GetGRPCPort()))
	if err != nil {
		return nil, nil, err
	}
	var zapLogger *zap.Logger
	var customFunc grpc_zap.CodeToLevel

	opts := []grpc_zap.Option{
		grpc_zap.WithLevels(customFunc),
	}
	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
	// grpc_zap.ReplaceGrpcLoggerV2(zapLogger)

	// s := grpc.NewServer(
	// 	grpc.UnaryInterceptor(sdk.LoggerInterceptor(&logger)),
	// )
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.UnaryServerInterceptor(zapLogger, opts...),
		)),
	)
	reflection.Register(s)
	pb.RegisterExampleServer(s, &example.Server{})

	return s, l, nil
}
