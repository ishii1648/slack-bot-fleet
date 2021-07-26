package broker

import (
	"bytes"
	"testing"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"net/http/httptest"
)

func TestResponseURLVerification(t *testing.T) {
	resprec := httptest.NewRecorder()
	body := `{
		"token": "Jhj5dZrVaK7ZwHHjRyZWjbDl",
		"challenge": "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P",
		"type": "url_verification"
	}`

	buf := &bytes.Buffer{}
	rootLogger := zerolog.SetLogger(buf, true, false)

	if err := responseURLVerification(rootLogger, resprec, []byte(body)); err != nil {
		t.Fatal(err)
	}

	if want, got := "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P", resprec.Body.String(); want != got {
		t.Errorf("want challenge %v, got %v", want, got)
	}
}
