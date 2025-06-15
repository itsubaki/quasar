package client

import (
	"maps"
	"net/http"
)

type HeaderTransport struct {
	Header       http.Header
	RoundTripper http.RoundTripper
}

func (t *HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	maps.Copy(newReq.Header, t.Header)
	return t.RoundTripper.RoundTrip(newReq)
}
