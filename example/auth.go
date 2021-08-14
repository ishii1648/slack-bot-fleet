package example

import (
	"context"
	"encoding/json"
	"fmt"
	// "io/ioutil"
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

			var body *exampleapi.RequestBody
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				errMsg := fmt.Sprintf("failed to unmarshal request body: %v", err)
				logger.Error(errMsg)
				pkghttp.Error(w, errMsg, pkghttp.StatusBadRequest)
				return
			}

			items, err := route.ParseReactionAddedItem(eventYamlPath)
			if err != nil {
				errMsg := fmt.Sprintf("failed to parse routing.yml: %v", err)
				logger.Error(errMsg)
				pkghttp.Error(w, errMsg, pkghttp.StatusInternalServerError)
				return
			}

			var matched bool
			for _, item := range items {
				if ok := item.Match(body.User, body.Reaction, body.ItemChannel); ok {
					matched = true
				}
			}

			if !matched {
				errMsg := fmt.Sprintf("no matched item (user=%s, reaction=%s, channel={%v}, )", body.User, body.Reaction, body.ItemChannel)
				logger.Error(errMsg)
				pkghttp.Error(w, errMsg, pkghttp.StatusBadRequest)
				return
			}

			h.ServeHTTP(w, r.WithContext(context.WithValue(ctx, "requestBody", body)))
		})
	}
}
