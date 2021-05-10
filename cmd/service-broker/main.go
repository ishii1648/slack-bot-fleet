package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ishii1648/slack-bot-fleet/pkg/crzerolog"
	broker "github.com/ishii1648/slack-bot-fleet/pkg/service-broker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	disableAuthFlag = flag.Bool("disable-auth", false, "disable Verifying requests from Slack")
)

func main() {
	flag.Parse()

	logger := zerolog.New(os.Stdout)

	srv := RegisterHTTPServer()
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error().Msgf("failed to ListenAndServe : %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)

	<-sigCh
	logger.Info().Msg("recive SIGTERM")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Msgf("failed to shutdown HTTP Server : %v", err)
	}

	logger.Info().Msg("HTTP Server shutdowned")
}

func RegisterHTTPServer() *http.Server {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	httpLogger := zerolog.New(os.Stdout)
	middleware := crzerolog.InjectLogger(&httpLogger)
	mux := http.NewServeMux()
	mux.Handle("/", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.Ctx(r.Context())

		if err := Run(w, r); err != nil {
			logger.Error().Msgf("%v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
			return
		}

		w.Write([]byte("ok"))
	})))

	return &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
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
