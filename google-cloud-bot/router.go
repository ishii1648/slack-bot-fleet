package gcb

import (
	// "net/http"
	"os"

	"github.com/ishii1648/slack-bot-fleet/google-cloud-bot/workflow"
	"github.com/ishii1648/slack-bot-fleet/pkg/slack"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_slack "github.com/slack-go/slack"
)

func Run(disableAuth bool, debug bool) {
	e := echo.New()
	if debug {
		e.Logger.SetLevel(log.DEBUG)
	} else {
		e.Logger.SetLevel(log.INFO)
	}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, error=${error} \n",
	}))
	e.Use(middleware.Recover())
	e.Use(slack.InjectVerifyingSlackRequest(disableAuth))

	slackC := _slack.New(os.Getenv("SLACK_BOT_TOKEN"))
	sudo := workflow.NewSudo(e.Logger, workflow.CommandSudo, slackC)

	slackInteractiveMsg := workflow.NewSlackInteractiveMsg(sudo)

	e.POST("/"+workflow.CommandSudo, sudo.Handle)
	e.POST("/interactive", slackInteractiveMsg.Handle)

	port, isSet := os.LookupEnv("PORT")
	if !isSet {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
