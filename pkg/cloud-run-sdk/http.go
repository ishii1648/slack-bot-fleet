package sdk

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
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
		trace := fmt.Sprintf("projects/%s/traces/%s", ProjectID, traceID)

		l.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("logging.googleapis.com/trace", trace)
		})
	} else {
		l = m.rootLogger.With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
		r = r.WithContext(l.WithContext(r.Context()))
	}

	m.next.ServeHTTP(w, r)
}

func RegisterDefaultHTTPServer(fn func(w http.ResponseWriter, r *http.Request)) *http.Server {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	hostAddr := "0.0.0.0"
	if h := os.Getenv("HOST_ADDR"); h != "" {
		hostAddr = h
	}

	httpLogger := zerolog.New(os.Stdout)
	middleware := InjectLogger(&httpLogger)
	mux := http.NewServeMux()
	mux.Handle("/", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fn(w, r)
	})))

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%s", hostAddr, port),
		Handler: mux,
	}
}
