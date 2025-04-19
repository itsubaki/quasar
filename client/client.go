package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Response struct {
	// parameters
	N    int    `json:"N"`
	T    int    `json:"t"`
	A    int    `json:"a"`
	Seed uint64 `json:"seed"`

	// results
	P  int    `json:"p,omitempty"`
	Q  int    `json:"q,omitempty"`
	M  string `json:"m"`
	SR string `json:"s/r"`
}

type Client struct {
	BaseURL       string
	IdentityToken string
}

func New(baseURL, identityToken string) *Client {
	return &Client{
		BaseURL:       baseURL,
		IdentityToken: identityToken,
	}
}

func (c *Client) Factorize(ctx context.Context, N, t, a int, seed uint64) (*Response, error) {
	// new request
	reqURL, err := url.JoinPath(c.BaseURL, "shor", fmt.Sprintf("%d", N))
	if err != nil {
		return nil, fmt.Errorf("join path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.IdentityToken))

	// add query parameters
	query := req.URL.Query()
	query.Add("t", fmt.Sprintf("%d", t))
	query.Add("a", fmt.Sprintf("%d", a))
	query.Add("seed", fmt.Sprintf("%d", seed))
	req.URL.RawQuery = query.Encode()

	// do
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	// check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code=%v", resp.StatusCode)
	}

	// read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	// unmarshal
	var res Response
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &res, nil
}
