package broker

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
)

func Run(ctx context.Context, w http.ResponseWriter, r *http.Request, l *zerolog.Logger, disableAuthFlag bool, localForwardFlag bool) error {
	body, err := Auth(ctx, l, r, disableAuthFlag)
	if err != nil {
		return err
	}

	if err := Proxy(ctx, l, w, body, localForwardFlag); err != nil {
		return err
	}

	return nil
}
