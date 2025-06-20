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

type Client struct {
	TargetURL  string
	HTTPClient *http.Client
}

func New(targetURL string, client *http.Client) *Client {
	return &Client{
		TargetURL:  targetURL,
		HTTPClient: client,
	}
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code=%v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return body, nil
}

func (c *Client) Factorize(ctx context.Context, N, t, a int, seed uint64) (*FactorizeResponse, error) {
	// new request
	reqURL, err := url.JoinPath(c.TargetURL, "factorize", fmt.Sprintf("%d", N))
	if err != nil {
		return nil, fmt.Errorf("join path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	// add query parameters
	query := req.URL.Query()
	query.Add("t", fmt.Sprintf("%d", t))
	query.Add("a", fmt.Sprintf("%d", a))
	query.Add("seed", fmt.Sprintf("%d", seed))
	req.URL.RawQuery = query.Encode()

	// do
	body, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
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

	// create form file
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TargetURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// do
	body, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	// unmarshal
	var res RunResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &res, nil
}
