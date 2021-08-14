package example

import (
	"context"
	pkghttp "net/http"

	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	exampleapi "github.com/ishii1648/slack-bot-fleet/api/example"
)

func Run(ctx context.Context) ([]byte, *http.AppError) {
	body, ok := ctx.Value("requestBody").(*exampleapi.RequestBody)
	if !ok {
		return nil, http.Error(pkghttp.StatusBadRequest, "requestBody not found")
	}

	logger := zerolog.Ctx(ctx)
	logger.Infof("revice request (user=%s, reaction=%s, channel=%v)", body.User, body.Reaction, body.ItemChannel)

	return []byte("ok"), nil
}
