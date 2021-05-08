package main

import (
	"net/http"
	"os"

	"github.com/ishii1648/slack-bot-fleet/pkg/crzerolog"
	event "github.com/ishii1648/slack-bot-fleet/pkg/event-api-server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	rootLogger := zerolog.New(os.Stdout)
	middleware := crzerolog.InjectLogger(&rootLogger)

	http.Handle("/", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.Ctx(r.Context())

		if err := Run(w, r); err != nil {
			logger.Error().Msgf("%v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
			return
		}

		w.Write([]byte("ok"))
	})))

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	log.Fatal().Msg(http.ListenAndServe(":"+port, nil).Error())
}

func Run(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	logger := log.Ctx(r.Context())

	body, err := event.Auth(ctx, logger, r)
	if err != nil {
		return err
	}

	if err := event.Proxy(logger, w, body); err != nil {
		return err
	}

	return nil
}
