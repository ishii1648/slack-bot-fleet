package broker

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
)

// see. https://api.slack.com/authentication/verifying-requests-from-slack
func Auth(ctx context.Context, l *zerolog.Logger, r *http.Request, disableAuthFlag bool) ([]byte, error) {
	verifier, err := slack.NewSecretsVerifier(r.Header, os.Getenv("SLACK_SIGNING_SECRET"))
	if err != nil {
		return nil, err
	}

	bodyReader := io.TeeReader(r.Body, &verifier)
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}

	if disableAuthFlag {
		l.Warn().Msg("disable Verifying requests from Slack")
		return body, nil
	}

	if err := verifier.Ensure(); err != nil {
		return nil, err
	}

	return body, nil
}
