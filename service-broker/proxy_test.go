package broker

import (
	"testing"
)

func TestResponseURLVerification(t *testing.T) {
	body := []byte(`{
		"token": "Jhj5dZrVaK7ZwHHjRyZWjbDl",
		"challenge": "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P",
		"type": "url_verification"
	}`)

	resChallenge, err := responseURLVerification(body)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P", string(resChallenge); want != got {
		t.Errorf("want challenge %v, got %v", want, got)
	}
}
