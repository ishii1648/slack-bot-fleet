package slack

import (
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

func InjectVerifyingSlackRequest(disableAuth bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if disableAuth {
				c.Logger().Info("skipping verification")
				return next(c)
			}

			slackSigningSecret, isSet := os.LookupEnv("SLACK_SIGNING_SECRET")
			if !isSet {
				return errors.New("SLACK_SIGNING_SECRET is not set")
			}

			sv, err := slack.NewSecretsVerifier(c.Request().Header, slackSigningSecret)
			if err != nil {
				return err
			}

			body, err := ioutil.ReadAll(c.Request().Body)
			if err != nil {
				return err
			}
			io.WriteString(&sv, string(body))

			c.Set("body", body)

			if err := sv.Ensure(); err != nil {
				return err
			}

			return next(c)
		}
	}
}
