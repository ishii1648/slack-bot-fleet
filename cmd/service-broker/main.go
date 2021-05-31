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
	debugFlag        = flag.Bool("debug", false, "debug mode")
)

func main() {
	flag.Parse()

	logger := sdk.SetLogger(zerolog.New(os.Stdout))

	srv := sdk.RegisterDefaultHTTPServer(Run, sdk.InjectLogger(logger, *debugFlag), InjectVerifyingSlackRequest())
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

func InjectVerifyingSlackRequest() sdk.Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := log.Ctx(r.Context())

			if !*disableAuthFlag {
				if err := broker.VerifySlackRequest(logger, r); err != nil {
					logger.Error().Msgf("invalid request : %v", err)
					return
				}
				logger.Debug().Msg("success to verify slack request")
			} else {
				logger.Warn().Msg("skip to verify slack request")
			}

			h.ServeHTTP(w, r)
		})

	}
}

func Run(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.Ctx(r.Context())

	if err := broker.Proxy(ctx, w, r, *localForwardFlag); err != nil {
		logger.Error().Msgf("%v", err)
		return
	}
}
