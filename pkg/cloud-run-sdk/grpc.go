package sdk

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	m "cloud.google.com/go/compute/metadata"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// InjectLoggerInterceptor returns a gRPC unary interceptor for injecting zerolog.Logger to the RPC invocation context.
func InjectLoggerInterceptor(rootLogger *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var l zerolog.Logger

		if isCloudRun() {
			l = rootLogger.With().Timestamp().Logger().Hook(sourceLocationHook)
			ctx = l.WithContext(ctx)

			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			values := md.Get("x-cloud-trace-context")
			if len(values) != 1 {
				return handler(ctx, req)
			}

			traceID, _ := traceContextFromHeader(values[0])
			if traceID == "" {
				return handler(ctx, req)
			}
			trace := fmt.Sprintf("projects/%s/traces/%s", ProjectID, traceID)

			l.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("logging.googleapis.com/trace", trace)
			})
		} else {
			l = rootLogger.With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
			ctx = l.WithContext(ctx)
		}

		return handler(ctx, req)
	}
}

func InjectClientAuthInterceptor(idToken string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md := metadata.New(map[string]string{"authorization": fmt.Sprintf("Bearer %s", idToken)})
		ctx = metadata.NewOutgoingContext(ctx, md)

		return nil
	}
}

func NewTLSConn(addr string) (*grpc.ClientConn, error) {
	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	cred := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})

	idToken, err := getIDToken(addr)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(
		addr,
		grpc.WithAuthority(addr),
		grpc.WithTransportCredentials(cred),
		grpc.WithUnaryInterceptor(InjectClientAuthInterceptor(idToken)),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func getIDToken(addr string) (string, error) {
	serviceURL := fmt.Sprintf("https://%s", strings.Split(addr, ":")[0])
	tokenURL := fmt.Sprintf("/instance/service-accounts/default/identity?audience=%s", serviceURL)

	idToken, err := m.Get(tokenURL)
	if err != nil {
		return "", err
	}

	return idToken, nil
}
