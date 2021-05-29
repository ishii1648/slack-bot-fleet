package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	sdk "github.com/ishii1648/slack-bot-fleet/pkg/cloud-run-sdk"
	broker "github.com/ishii1648/slack-bot-fleet/service-broker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	disableAuthFlag  = flag.Bool("disable-auth", false, "disable Verifying requests from Slack")
	localForwardFlag = flag.Bool("local-forward", false, "grpc connection to localhost")
)

func main() {
	flag.Parse()

	logger := zerolog.New(os.Stdout)

	srv := sdk.RegisterDefaultHTTPServer(RunWrapper)
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error().Msgf("server closed with error : %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	<-sigCh
	logger.Info().Msg("recive SIGTERM or SIGINT")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Msgf("failed to shutdown HTTP Server : %v", err)
	}

	logger.Info().Msg("HTTP Server shutdowned")
}

func RunWrapper(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.Ctx(r.Context())

	if err := Run(ctx, w, r, logger); err != nil {
		logger.Error().Msgf("%v", err)
		return
	}
}

func Run(ctx context.Context, w http.ResponseWriter, r *http.Request, l *zerolog.Logger) error {
	body, err := broker.Auth(ctx, l, r, *disableAuthFlag)
	if err != nil {
		return err
	}

	if err := broker.Proxy(ctx, l, w, body, *localForwardFlag); err != nil {
		return err
	}

	return nil
}
