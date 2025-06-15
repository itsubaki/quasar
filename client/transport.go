package client

import (
	"fmt"
	"maps"
	"net/http"
)

func NewWithIdentityToken(token string) *http.Client {
	return &http.Client{
		Transport: &HeaderTransport{
			Header: http.Header{
				"Authorization": []string{fmt.Sprintf("Bearer %s", token)},
			},
			RoundTripper: http.DefaultTransport.(*http.Transport).Clone(),
		},
	}
}

type HeaderTransport struct {
	Header       http.Header
	RoundTripper http.RoundTripper
}

func (t *HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	maps.Copy(newReq.Header, t.Header)
	return t.RoundTripper.RoundTrip(newReq)
}
