package broker

import (
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/rs/zerolog/log"
	"net/http/httptest"
	"testing"
)

func TestResponseURLVerification(t *testing.T) {
	resprec := httptest.NewRecorder()
	body := `{
		"token": "Jhj5dZrVaK7ZwHHjRyZWjbDl",
		"challenge": "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P",
		"type": "url_verification"
	}`

	logger := zerolog.NewLogger(&log.Logger)

	if err := responseURLVerification(logger, resprec, []byte(body)); err != nil {
		t.Fatal(err)
	}

	if want, got := "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P", resprec.Body.String(); want != got {
		t.Errorf("want challenge %v, got %v", want, got)
	}
}
