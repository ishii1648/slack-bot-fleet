package example

import (
	"context"
	"encoding/json"
	"fmt"
	pkghttp "net/http"

	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	exampleapi "github.com/ishii1648/slack-bot-fleet/api/example"
	"github.com/ishii1648/slack-bot-fleet/pkg/route"
)

func InjectVerifyRequest(eventYamlPath string) http.Middleware {
	return func(h pkghttp.Handler) pkghttp.Handler {
		return pkghttp.HandlerFunc(func(w pkghttp.ResponseWriter, r *pkghttp.Request) {
			ctx := r.Context()
			logger := zerolog.Ctx(ctx)

			body, err := verifyRequest(eventYamlPath, r)
			if err != nil {
				logger.Errorf("failed to verify request: %v", err)
				pkghttp.Error(w, err.Error(), pkghttp.StatusBadRequest)
				return
			}

			h.ServeHTTP(w, r.WithContext(context.WithValue(ctx, "requestBody", body)))
		})
	}
}

func verifyRequest(eventYamlPath string, r *pkghttp.Request) (*exampleapi.RequestBody, error) {
	var body *exampleapi.RequestBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request body: %v", err)
	}

	items, err := route.ParseReactionAddedItem(eventYamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse routing.yml: %v", err)
	}

	var matched bool
	for _, item := range items {
		if ok := item.Match(body.User, body.Reaction, body.ItemChannel); ok {
			matched = true
			break
		}
	}

	if !matched {
		return nil, fmt.Errorf("no matched item (user=%s, reaction=%s, channel=%v)", body.User, body.Reaction, body.ItemChannel)
	}

	return body, nil
}
