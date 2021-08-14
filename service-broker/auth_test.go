package broker

import (
	"bytes"
	"context"
	pkghttp "net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
)

func TestInjectVerifyingSlackRequest(t *testing.T) {
	buf := &bytes.Buffer{}
	rootLogger := zerolog.SetLogger(buf, false, true)
	resprec := httptest.NewRecorder()

	var run AppHandler = func(ctx context.Context) ([]byte, *AppError) {
		requestBody, ok := ctx.Value("requestBody").([]byte)
		if !ok {
			t.Fatal("requestBody is not found")
		}
		return requestBody, nil
	}

	handler := http.Chain(run, http.InjectLogger(rootLogger, "google-sample-project"), InjectVerifyingSlackRequest(true))

	var requestFunc = func(jsonBody string) *pkghttp.Request {
		r := httptest.NewRequest("POST", "/", strings.NewReader(jsonBody))
		r.Header.Add("X-Slack-Signature", "0123456789abcdef")
		r.Header.Add("X-Slack-Request-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
		r.Header.Add("Content-Type", "application/json")
		return r
	}

	slackEvent := `{"type":"event_callback","event":{"type":"reaction_added","user":"U020VK32D63","item":{"type":"message","channel":"C0213JYV3HC","ts":"1620581892.006600"},"reaction":"ok_hand","item_user":"U020VK32D63","event_ts":"1622040734.000100"}}`
	handler.ServeHTTP(resprec, requestFunc(slackEvent))

	if want, got := slackEvent, resprec.Body.String(); want != got {
		t.Errorf("want slackEvent %v, got %v", want, got)
	}
}
