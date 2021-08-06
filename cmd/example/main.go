package main

import (
	"flag"
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/ishii1648/cloud-run-sdk/grpc"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/ishii1648/slack-bot-fleet/example"
	pb "github.com/ishii1648/slack-bot-fleet/proto/reaction-added-event"
	"go.opencensus.io/trace"
)

var (
	debugFlag = flag.Bool("debug", false, "debug mode")
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
			MetricPrefix: "example",
		})
		if err != nil {
			rootLogger.Errorf("failed to NewExporter : %v", err)
			os.Exit(1)
		}
		trace.RegisterExporter(exporter)
	}

	s := grpc.NewServer(rootLogger, projectID, example.VerifyRequestInterceptor("./event.yaml"))
	pb.RegisterReactionServer(s.Srv, &example.Server{})

	lis, err := grpc.CreateNetworkListener()
	if err != nil {
		rootLogger.Errorf("failed to create listener : %v", err)
		os.Exit(1)
	}

	s.Start(lis, util.SetupSignalHandler())
}
