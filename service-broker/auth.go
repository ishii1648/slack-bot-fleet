package broker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	pkghttp "net/http"

	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/slack-go/slack"
	"go.opencensus.io/trace"
)

func InjectVerifyingSlackRequest(disableAuth bool) http.Middleware {
	return func(h pkghttp.Handler) pkghttp.Handler {
		return pkghttp.HandlerFunc(func(w pkghttp.ResponseWriter, r *pkghttp.Request) {
			ctx := r.Context()
			logger := zerolog.Ctx(ctx)

			body, err := VerifySlackRequest(ctx, r, disableAuth)
			if err != nil {
				logger.Errorf("invalid request : %v", err)
				return
			}

			h.ServeHTTP(w, r.WithContext(context.WithValue(ctx, "requestBody", body)))
		})
	}
}

// see. https://api.slack.com/authentication/verifying-requests-from-slack
func VerifySlackRequest(ctx context.Context, r *pkghttp.Request, disableAuth bool) ([]byte, error) {
	sc := trace.FromContext(ctx).SpanContext()
	_, span := trace.StartSpanWithRemoteParent(ctx, "VerifySlackRequest", sc)
	defer span.End()
	logger := zerolog.Ctx(ctx)

	projectID, err := util.FetchProjectID()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch projectID : %v", err)
	}

	slackSigningSecret, err := util.FetchSecretLatestVersion(ctx, "slack-signing-secret", projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch slackSigningSecret : %v", err)
	}

	verifier, err := slack.NewSecretsVerifier(r.Header, slackSigningSecret)
	if err != nil {
		return nil, err
	}

	bodyReader := io.TeeReader(r.Body, &verifier)
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}

	if disableAuth {
		logger.Info("skip to verify slack request")
		return body, nil
	}

	if err := verifier.Ensure(); err != nil {
		return nil, err
	}

	return body, nil
}
