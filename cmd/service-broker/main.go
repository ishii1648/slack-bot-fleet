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
	disableAuthFlag = flag.Bool("disable-auth", false, "disable Verifying requests from Slack")
)

func main() {
	flag.Parse()

	logger := zerolog.New(os.Stdout)

	srv := sdk.RegisterDefaultHTTPServer(Run)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error().Msgf("failed to ListenAndServe : %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	<-sigCh
	logger.Info().Msg("recive SIGTERM or SIGINT")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Msgf("failed to shutdown HTTP Server : %v", err)
	}

	logger.Info().Msg("HTTP Server shutdowned")
}

func Run(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	logger := log.Ctx(r.Context())

	body, err := broker.Auth(ctx, logger, r, *disableAuthFlag)
	if err != nil {
		return err
	}

	if err := broker.Proxy(ctx, logger, w, body); err != nil {
		return err
	}

	return nil
}
