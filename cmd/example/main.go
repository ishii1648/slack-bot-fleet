package main

import (
	"flag"
	"os"

	"github.com/ishii1648/cloud-run-sdk/grpc"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/ishii1648/slack-bot-fleet/example"
	pb "github.com/ishii1648/slack-bot-fleet/proto/reaction-added-event"
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

	s := grpc.NewServer(rootLogger, projectID, example.VerifyRequestInterceptor("./event.yaml"))
	pb.RegisterReactionServer(s.Srv, &example.Server{})

	lis, err := grpc.CreateNetworkListener()
	if err != nil {
		rootLogger.Errorf("failed to create listener : %v", err)
		os.Exit(1)
	}

	s.Start(lis, util.SetupSignalHandler())
}
