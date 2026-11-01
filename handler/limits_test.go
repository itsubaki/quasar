package handler_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/handler"
	"github.com/itsubaki/quasar/store"
)

func TestCodeSizeLimits(t *testing.T) {
	cases := []struct {
		name   string
		size   int
		errMsg string
	}{
		{
			name: "at limit",
			size: 64 * 1024,
		},
		{
			name: "below limit",
			size: 64*1024 - 1,
		},
		{
			name:   "above limit",
			size:   64*1024 + 1,
			errMsg: "invalid_argument: code size exceeds 65536 bytes",
		},
	}

	svc := &handler.QuasarService{}
	const prefix = "qubit q; //"

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code := prefix + strings.Repeat("x", c.size-len(prefix))
			_, validateErr := svc.Validate(t.Context(), connect.NewRequest(&quasarv1.ValidateRequest{Code: code}))
			_, simulateErr := svc.Simulate(t.Context(), connect.NewRequest(&quasarv1.SimulateRequest{Code: code}))

			for _, err := range []error{validateErr, simulateErr} {
				if err != nil {
					if err.Error() != c.errMsg {
						t.Errorf("got=%v, want=%v", err, c.errMsg)
					}

					continue
				}

				if c.errMsg != "" {
					t.Errorf("got=nil, want=%v", c.errMsg)
				}
			}
		})
	}
}

func TestRequestSizeLimits(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		code       string
		compressed bool
		statusCode int
		errCode    string
	}{
		{
			name:       "Simulate uncompressed",
			method:     "Simulate",
			code:       strings.Repeat("x", 512*1024),
			compressed: false,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Simulate gzip",
			method:     "Simulate",
			code:       strings.Repeat("x", 512*1024),
			compressed: true,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Validate uncompressed",
			method:     "Validate",
			code:       strings.Repeat("x", 512*1024),
			compressed: false,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Validate gzip",
			method:     "Validate",
			code:       strings.Repeat("x", 512*1024),
			compressed: true,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Share uncompressed",
			method:     "Share",
			code:       strings.Repeat("x", 512*1024),
			compressed: false,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Share gzip",
			method:     "Share",
			code:       strings.Repeat("x", 512*1024),
			compressed: true,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Edit uncompressed",
			method:     "Edit",
			code:       strings.Repeat("x", 512*1024),
			compressed: false,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Edit gzip",
			method:     "Edit",
			code:       strings.Repeat("x", 512*1024),
			compressed: true,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			// Preserve the full 64 KiB source allowance even with six-byte JSON escapes.
			name:       "Share escaped source at limit",
			method:     "Share",
			code:       strings.Repeat("\x01", 64*1024),
			statusCode: http.StatusOK,
		},
	}

	h, err := handler.New(1, 1, &store.MemoryStore{})
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{
				"code": c.code,
			})
			if err != nil {
				t.Fatal(err)
			}

			if c.compressed {
				var zipped bytes.Buffer
				w := gzip.NewWriter(&zipped)
				if _, err := w.Write(body); err != nil {
					t.Fatal(err)
				}

				if err := w.Close(); err != nil {
					t.Fatal(err)
				}

				body = zipped.Bytes()
			}

			req := httptest.NewRequest(http.MethodPost, "/quasar.v1.QuasarService/"+c.method, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if c.compressed {
				req.Header.Set("Content-Encoding", "gzip")
			}

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != c.statusCode {
				t.Fatalf("status=%d, want=%d response=%s", rec.Code, c.statusCode, rec.Body.String())
			}

			var result struct {
				Code string `json:"code"`
			}

			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}

			if result.Code != c.errCode {
				t.Errorf("code=%s, want=%s response=%s", result.Code, c.errCode, rec.Body.String())
			}
		})
	}
}
