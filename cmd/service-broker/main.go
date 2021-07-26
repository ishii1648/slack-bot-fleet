package main

import (
	"flag"
	pkghttp "net/http"
	"os"

	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	broker "github.com/ishii1648/slack-bot-fleet/service-broker"
)

var (
	disableAuthFlag = flag.Bool("disable-auth", false, "disable Verifying requests from Slack")
	debugFlag       = flag.Bool("debug", false, "debug mode")
)

func main() {
	flag.Parse()

	rootLogger := zerolog.SetDefaultLogger(*debugFlag)

	projectID, err := util.FetchProjectID()
	if err != nil {
		rootLogger.Errorf("failed to fetch project ID : %v", err)
		os.Exit(1)
	}

	server := http.NewServerWithLogger(rootLogger, projectID)

	server.HandleWithRoot(http.AppHandler(Run), broker.InjectVerifyingSlackRequest(*disableAuthFlag))
	server.Start(util.SetupSignalHandler())
}

func Run(w pkghttp.ResponseWriter, r *pkghttp.Request) *http.Error {
	if err := broker.Run(w, r); err != nil {
		return err
	}

	return nil
}
