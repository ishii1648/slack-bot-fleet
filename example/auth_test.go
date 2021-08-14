package example

import (
	"bytes"
	// "context"
	pkghttp "net/http"
	"net/http/httptest"
	"testing"
	// "github.com/ishii1648/cloud-run-sdk/http"
	// exampleapi "github.com/ishii1648/slack-bot-fleet/api/example"
)

// func TestVerifyRequestInterceptor(t *testing.T) {
// 	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
// 		return "output", nil
// 	}
// 	var run http.AppHandler = func(ctx context.Context) ([]byte, *http.AppError) {
// 		requestBody, ok := ctx.Value("requestBody").([]byte)
// 		if !ok {
// 			t.Fatal("requestBody is not found")
// 		}
// 		return requestBody, nil
// 	}

// 	tests := []struct {
// 		// req
// 		want bool
// 	}{
// 		{
// 			req: &pb.Request{
// 				Reaction: "ok_hand",
// 				User:     "s_tarou",
// 				Item: &pb.EventItem{
// 					Channel: "development",
// 					Ts:      "1620581892.006600",
// 				},
// 			},
// 			want: false,
// 		},
// 		{
// 			req: &pb.Request{
// 				Reaction: "ok_hand",
// 				User:     "s_ishii",
// 				Item: &pb.EventItem{
// 					Channel: "development",
// 					Ts:      "1620581892.006600",
// 				},
// 			},
// 			want: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		ctx := context.Background()
// 		resp, err := VerifyRequestInterceptor("../tests/routing.yml")(ctx, tt.req, nil, unaryHandler)
// 		if tt.want {
// 			if err != nil {
// 				t.Fatalf("unexpected error: %v", err)
// 			}

// 			respStr, ok := resp.(string)
// 			if !ok {
// 				t.Fatalf("unexpected response type: %T", resp)
// 			}

// 			if respStr != "output" {
// 				t.Errorf("unexpected response: %s", respStr)
// 			}

// 			continue
// 		}

// 		if err != nil && err.Error() != "rpc error: code = InvalidArgument desc = no matched item" {
// 			t.Errorf("unexpected error: %v", err)
// 		}
// 	}
// }

func TestVerifyRequest(t *testing.T) {
	tests := []struct {
		req  []byte
		want bool
	}{
		{
			req: []byte(`{
				"reaction": "ok_hand",
				"user": "m_taro",
				"itemChannel": "sample",
				"itemTimstamp": "1620581892.006600"
			}`),
			want: true,
		},
		{
			req: []byte(`{
				"reaction": "ok_hand",
				"user": "s_ishii",
				"itemChannel": "sample",
				"itemTimstamp": "1620581892.006600"
			}`),
			want: false,
		},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(pkghttp.MethodGet, "http://dummy.url.com", bytes.NewReader(tt.req))
		_, err := verifyRequest("../tests/routing.yml", req)
		if tt.want && err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !tt.want && err == nil {
			t.Errorf("unexpected success")
		}
	}
}
