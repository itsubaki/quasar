package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type FactorizeResponse struct {
	// parameters
	N    int    `json:"N,omitempty"`
	T    int    `json:"t,omitempty"`
	A    int    `json:"a,omitempty"`
	Seed uint64 `json:"seed,omitempty"`

	// results
	P  int    `json:"p,omitempty"`
	Q  int    `json:"q,omitempty"`
	M  string `json:"m,omitempty"`
	SR string `json:"s/r,omitempty"`

	// message
	Message string `json:"message,omitempty"`
}

type RunResponse struct {
	State []State `json:"state"`
}

type State struct {
	Amplitude    Amplitude `json:"amplitude"`
	Probability  float64   `json:"probability"`
	Int          []int     `json:"int"`
	BinaryString []string  `json:"binary_string"`
}

type Amplitude struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
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

func (c *Client) Factorize(ctx context.Context, N, t, a int, seed uint64) (*FactorizeResponse, error) {
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
	var res FactorizeResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &res, nil
}

func (c *Client) Run(ctx context.Context, content string) (*RunResponse, error) {
	// body
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "request.qasm")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}

	if _, err = io.Copy(part, strings.NewReader(content)); err != nil {
		return nil, fmt.Errorf("copy: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer: %w", err)
	}

	// new request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.IdentityToken))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// do
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
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
	var res RunResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &res, nil
}
