package main

import (
	"flag"
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	broker "github.com/ishii1648/slack-bot-fleet/service-broker"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
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

	if util.IsCloudRun() {
		exporter, err := stackdriver.NewExporter(stackdriver.Options{
			MetricPrefix: "service-broker",
		})
		if err != nil {
			rootLogger.Errorf("failed to NewExporter : %v", err)
			os.Exit(1)
		}
		trace.RegisterExporter(exporter)
	}

	server := http.NewServerWithLogger(rootLogger, projectID)

	httpHandler := &ochttp.Handler{
		Handler:     http.Chain(http.AppHandler(broker.Run), http.InjectLogger(rootLogger, projectID), broker.InjectVerifyingSlackRequest(*disableAuthFlag)),
		Propagation: &propagation.HTTPFormat{},
	}

	server.Handle("/", httpHandler)
	server.Start(util.SetupSignalHandler())
}
