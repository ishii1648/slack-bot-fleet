package sdk

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// middleware implements http.Handler interface.
type middleware struct {
	rootLogger *zerolog.Logger
	next       http.Handler
}

// InjectLogger returns an HTTP middleware for injecting zerolog.Logger to the request context.
func InjectLogger(rootLogger *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &middleware{rootLogger, next}
	}
}

// ServeHTTP injects zerolog.Logger to the http context and calls the next handler.
func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var l zerolog.Logger

	if isCloudRun() {
		l = m.rootLogger.With().Timestamp().Logger().Hook(sourceLocationHook)
		r = r.WithContext(l.WithContext(r.Context()))

		traceID, _ := traceContextFromHeader(r.Header.Get("X-Cloud-Trace-Context"))
		if traceID == "" {
			m.next.ServeHTTP(w, r)
			return
		}
		trace := fmt.Sprintf("projects/%s/traces/%s", projectID, traceID)

		l.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("logging.googleapis.com/trace", trace)
		})
	} else {
		l = m.rootLogger.With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
		r = r.WithContext(l.WithContext(r.Context()))
	}

	m.next.ServeHTTP(w, r)
}

func RegisterDefaultHTTPServer(fn func(w http.ResponseWriter, r *http.Request) error) *http.Server {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	httpLogger := zerolog.New(os.Stdout)
	middleware := InjectLogger(&httpLogger)
	mux := http.NewServeMux()
	mux.Handle("/", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.Ctx(r.Context())

		if err := fn(w, r); err != nil {
			logger.Error().Msgf("%v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
			return
		}

		w.Write([]byte("ok"))
	})))

	return &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
}
