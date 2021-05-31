package broker

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
)

// see. https://api.slack.com/authentication/verifying-requests-from-slack
func VerifySlackRequest(l *zerolog.Logger, r *http.Request) error {
	verifier, err := slack.NewSecretsVerifier(r.Header, os.Getenv("SLACK_SIGNING_SECRET"))
	if err != nil {
		return err
	}

	bodyReader := io.TeeReader(r.Body, &verifier)
	if _, err := ioutil.ReadAll(bodyReader); err != nil {
		return err
	}

	if err := verifier.Ensure(); err != nil {
		return err
	}

	return nil
}
