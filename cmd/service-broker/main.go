package main

import (
	"net/http"
	"os"

	"github.com/ishii1648/slack-bot-fleet/pkg/crzerolog"
	broker "github.com/ishii1648/slack-bot-fleet/pkg/service-broker"
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

	body, err := broker.Auth(ctx, logger, r)
	if err != nil {
		return err
	}

	msg, channelID, err := broker.Proxy(ctx, logger, w, body)
	if err != nil {
		return err
	}

	if err := broker.Reply(logger, msg, channelID); err != nil {
		return err
	}

	return nil
}
