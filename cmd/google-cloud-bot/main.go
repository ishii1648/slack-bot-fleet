package main

import (
	"flag"

	gcb "github.com/ishii1648/slack-bot-fleet/google-cloud-bot"
)

var (
	disableAuthFlag = flag.Bool("disable-auth", false, "disable verifying requests from Slack")
	debugFlag       = flag.Bool("debug", false, "debug mode")
)

func init() {
	flag.Parse()
}

func main() {
	gcb.Run(*disableAuthFlag, *debugFlag)
}
