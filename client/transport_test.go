package client_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itsubaki/quasar/client"
)

func TestNewWithIdentityToken(t *testing.T) {
	token := "test"
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != fmt.Sprintf("Bearer %s", token) {
			panic("invalid authorization header")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer s.Close()

	c := client.NewWithIdentityToken(token)
	resp, err := c.Get(s.URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic("invalid status code")
	}
}
