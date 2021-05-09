package event

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
)

func Auth(ctx context.Context, l *zerolog.Logger, r *http.Request) ([]byte, error) {
	verifier, err := slack.NewSecretsVerifier(r.Header, os.Getenv("SLACK_SIGNING_SECRET"))
	if err != nil {
		return nil, err
	}

	bodyReader := io.TeeReader(r.Body, &verifier)
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}

	if err := verifier.Ensure(); err != nil {
		return nil, err
	}

	return body, nil
}
