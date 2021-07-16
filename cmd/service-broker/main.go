package main

import (
	"flag"
	pkghttp "net/http"

	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	broker "github.com/ishii1648/slack-bot-fleet/service-broker"
	"github.com/rs/zerolog/log"
)

var (
	disableAuthFlag = flag.Bool("disable-auth", false, "disable Verifying requests from Slack")
	debugFlag       = flag.Bool("debug", false, "debug mode")
)

func main() {
	flag.Parse()

	rootLogger := zerolog.SetDefaultLogger(*debugFlag)

	handler, err := http.BindHandlerWithLogger(&rootLogger, http.AppHandler(Run), broker.InjectVerifyingSlackRequest(*disableAuthFlag))
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	http.StartHTTPServer("/", handler, util.SetupSignalHandler())
}

func Run(w pkghttp.ResponseWriter, r *pkghttp.Request) *http.Error {
	if err := broker.Run(w, r); err != nil {
		return err
	}

	return nil
}
