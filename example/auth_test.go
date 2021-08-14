package example

import (
	"context"
	"testing"

	pb "github.com/ishii1648/slack-bot-fleet/proto/reaction-added-event"
)

func TestVerifyRequestInterceptor(t *testing.T) {
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "output", nil
	}

	tests := []struct {
		req  *pb.Request
		want bool
	}{
		{
			req: &pb.Request{
				Reaction: "ok_hand",
				User:     "s_tarou",
				Item: &pb.EventItem{
					Channel: "development",
					Ts:      "1620581892.006600",
				},
			},
			want: false,
		},
		{
			req: &pb.Request{
				Reaction: "ok_hand",
				User:     "s_ishii",
				Item: &pb.EventItem{
					Channel: "development",
					Ts:      "1620581892.006600",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		ctx := context.Background()
		resp, err := VerifyRequestInterceptor("../tests/routing.yml")(ctx, tt.req, nil, unaryHandler)
		if tt.want {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			respStr, ok := resp.(string)
			if !ok {
				t.Fatalf("unexpected response type: %T", resp)
			}

			if respStr != "output" {
				t.Errorf("unexpected response: %s", respStr)
			}

			continue
		}

		if err != nil && err.Error() != "rpc error: code = InvalidArgument desc = no matched item" {
			t.Errorf("unexpected error: %v", err)
		}
	}
}
