package broker

import (
	"context"
	"fmt"
	pkghttp "net/http"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
)

// It's usually a mistake to pass back the concrete type of an error rather than error,
// because it can make it difficult to catch errors,
// but it's the right thing to do here because ServeHTTP is the only place that sees the value and uses its contents.
type AppError struct {
	// http status code for client user
	Code int `json:"code"`
	// error message in HTTP Server
	Message string `json:"message"`
}

func Error(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

func Errorf(code int, format string, a ...interface{}) *AppError {
	return Error(code, fmt.Sprintf(format, a...))
}

func (e *AppError) Error() string {
	return e.Message
}

type AppHandler func(context.Context) (string, *AppError)

func (fn AppHandler) ServeHTTP(w pkghttp.ResponseWriter, r *pkghttp.Request) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx)

	res, err := fn(ctx)
	if err != nil {
		// just display error log because client(slack) can't change request in response of error
		switch {
		case pkghttp.StatusBadRequest >= err.Code:
			logger.Warn(err.Error())
		case pkghttp.StatusInternalServerError >= err.Code:
			logger.Errorf(err.Error())
		}
	}

	if res != "" {
		res = "done"
	}

	if _, err := w.Write([]byte(res)); err != nil {
		logger.Error(err)
	}
}
