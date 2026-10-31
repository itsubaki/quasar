package client

import (
	"fmt"
	"maps"
	"net/http"
)

func NewWithIdentityToken(token string) *http.Client {
	return &http.Client{
		// The transport adds credentials to every request, including redirects.
		// Return redirects to the caller so credentials never follow a new origin.
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
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
